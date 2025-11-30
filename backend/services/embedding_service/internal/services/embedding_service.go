package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"embedding_service/internal/events"
	"embedding_service/internal/models"
	"embedding_service/internal/repository"
)

type EmbeddingService struct {
	consumer   *events.Consumer
	producer   *events.Producer
	qrepo      *repository.QdrantRepo
	xenova     *XenovaClient
	ocrBaseURL string
}

func NewEmbeddingService(consumer *events.Consumer, producer *events.Producer, qrepo *repository.QdrantRepo, x *XenovaClient) *EmbeddingService {
	ocrURL := os.Getenv("OCR_SERVICE_URL")
	if ocrURL == "" {
		ocrURL = "http://ocr_service:8090"
	}
	return &EmbeddingService{
		consumer:   consumer,
		producer:   producer,
		qrepo:      qrepo,
		xenova:     x,
		ocrBaseURL: ocrURL,
	}
}

func (s *EmbeddingService) updateFileStatus(fileID, status string, totalChunks int, message string) {
	url := fmt.Sprintf("%s/internal/status", s.ocrBaseURL)
	payload := map[string]interface{}{
		"file_id":      fileID,
		"status":       status,
		"total_chunks": totalChunks,
		"message":      message,
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ Failed to update status: %v", err)
		return
	}
	defer resp.Body.Close()
}

func (s *EmbeddingService) Run(ctx context.Context) {
	log.Println("🚀 Embedding service: start consumer loop")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Embedding service: context done")
			return
		default:
			msg, err := s.consumer.ReadMessage(ctx)
			if err != nil {
				log.Printf("❌ consumer read error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var evt models.OCRProcessedEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("❌ invalid ocr_processed event: %v", err)
				continue
			}

			log.Printf("📥 Processing OCR_PROCESSED for file %s (%s) with %d chunks", evt.FileID, evt.FileName, len(evt.Chunks))

			if len(evt.Chunks) == 0 {
				log.Printf("⚠️ No chunks to embed for file %s", evt.FileID)
				s.updateFileStatus(evt.FileID, "completed", 0, "No chunks to embed")
				continue
			}

			// İlk chunk'ı embed et ve dimension al
			var firstVec []float32
			maxRetries := 3
			for i := 0; i < maxRetries; i++ {
				firstVec, err = s.xenova.Embed(evt.Chunks[0].Text)
				if err == nil {
					break
				}
				log.Printf("⚠️ Xenova embed retry %d/%d failed: %v", i+1, maxRetries, err)
				time.Sleep(time.Second * time.Duration(i+1))
			}

			if err != nil {
				log.Printf("❌ xenova embed failed after retries: %v", err)
				s.updateFileStatus(evt.FileID, "failed", 0, fmt.Sprintf("Embedding failed: %v", err))
				s.publishError(evt.FileID, err)
				continue
			}

			dimension := len(firstVec)
			log.Printf("📊 Embedding dimension: %d", dimension)

			// Collection'ı oluştur/kontrol et
			if err := s.qrepo.EnsureCollection("documents", dimension); err != nil {
				log.Printf("❌ failed to ensure collection: %v", err)
				s.updateFileStatus(evt.FileID, "failed", 0, fmt.Sprintf("Collection setup failed: %v", err))
				s.publishError(evt.FileID, err)
				continue
			}

			var stored []models.StoredChunkInfo
			successCount := 0

			// Tüm chunk'ları işle
			for i, ch := range evt.Chunks {
				// Embedding oluştur (retry ile)
				var vec []float32
				for retry := 0; retry < maxRetries; retry++ {
					vec, err = s.xenova.Embed(ch.Text)
					if err == nil {
						break
					}
					log.Printf("⚠️ Embed retry %d/%d for chunk %d: %v", retry+1, maxRetries, i, err)
					time.Sleep(time.Millisecond * 500)
				}

				if err != nil {
					log.Printf("❌ xenova embed failed for chunk %s after retries: %v", ch.ChunkID, err)
					continue
				}

				// 🔥 KRİTİK: Payload'u sadece Qdrant uyumlu tiplerle oluştur
				payload := map[string]interface{}{
					"file_id":      evt.FileID,
					"file_name":    evt.FileName,
					"content_type": evt.ContentType,
					"chunk_id":     ch.ChunkID,
					"text":         ch.Text,
				}

				// Sadece 0'dan büyük değerleri ekle
				if ch.Page > 0 {
					payload["page"] = float64(ch.Page) // 🔥 int -> float64
				}
				if ch.StartOffset > 0 {
					payload["start_offset"] = float64(ch.StartOffset) // 🔥 int -> float64
				}
				if ch.EndOffset > 0 {
					payload["end_offset"] = float64(ch.EndOffset) // 🔥 int -> float64
				}

				// 🔥 Metadata'yı flatten et - Sadece primitive tipler
				if ch.Metadata != nil {
					for k, v := range ch.Metadata {
						if v == nil {
							continue // nil değerleri atla
						}

						switch val := v.(type) {
						case string:
							if val != "" {
								payload["meta_"+k] = val
							}
						case int:
							payload["meta_"+k] = float64(val)
						case int64:
							payload["meta_"+k] = float64(val)
						case int32:
							payload["meta_"+k] = float64(val)
						case float64:
							payload["meta_"+k] = val
						case float32:
							payload["meta_"+k] = float64(val)
						case bool:
							payload["meta_"+k] = val
						default:
							// Bilinmeyen tipleri string'e çevir
							strVal := fmt.Sprintf("%v", val)
							if strVal != "" && strVal != "<nil>" {
								payload["meta_"+k] = strVal
							}
						}
					}
				}

				// 🔥 DEBUG: İlk chunk'ın payload'unu logla
				if i == 0 {
					payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
					log.Printf("📦 First chunk payload:\n%s", string(payloadJSON))
				}

				// Qdrant'a kaydet
				if _, err := s.qrepo.UpsertVector("documents", ch.ChunkID, vec, payload); err != nil {
					log.Printf("❌ qdrant upsert failed for chunk %s: %v", ch.ChunkID, err)
					continue
				}

				stored = append(stored, models.StoredChunkInfo{
					ChunkID:  ch.ChunkID,
					VectorID: ch.ChunkID,
					Page:     ch.Page,
				})

				successCount++

				if i == 0 {
					log.Printf("✅ First chunk embedded successfully")
				}

				// Progress log
				if (i+1)%5 == 0 || i == len(evt.Chunks)-1 {
					log.Printf("📈 Progress: %d/%d chunks embedded", i+1, len(evt.Chunks))
				}
			}

			log.Printf("✅ Total embedded: %d/%d chunks for file %s", successCount, len(evt.Chunks), evt.FileID)
			s.updateFileStatus(evt.FileID, "completed", len(stored), fmt.Sprintf("Successfully embedded %d/%d chunks", successCount, len(evt.Chunks)))

			// EMBEDDING_STORED event yayınla
			out := models.EmbeddingStoredEvent{
				Event:       "EMBEDDING_STORED",
				FileID:      evt.FileID,
				FileName:    evt.FileName,
				ContentType: evt.ContentType,
				Chunks:      stored,
				TotalChunks: len(stored),
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			}

			b, _ := json.Marshal(out)
			if err := s.producer.Publish("embedding_stored", b); err != nil {
				log.Printf("❌ failed to publish embedding_stored: %v", err)
			} else {
				log.Printf("✅ Published EMBEDDING_STORED event for file %s", evt.FileID)
			}
		}
	}
}

func (s *EmbeddingService) publishError(fileID string, err error) {
	errorEvt := map[string]interface{}{
		"event":     "EMBEDDING_FAILED",
		"file_id":   fileID,
		"error":     err.Error(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(errorEvt)
	s.producer.Publish("embedding_failed", b)
}
