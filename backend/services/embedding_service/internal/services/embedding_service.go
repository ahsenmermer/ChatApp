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

// cleanPayload nil değerleri ve uyumsuz tipleri temizler
func cleanPayload(p map[string]interface{}) map[string]interface{} {
	clean := make(map[string]interface{})
	for k, v := range p {
		switch v := v.(type) {
		case nil:
			continue
		case string, int, float64, bool, map[string]interface{}:
			clean[k] = v
		default:
			// bilinmeyen tipleri string olarak kaydet
			clean[k] = fmt.Sprintf("%v", v)
		}
	}
	return clean
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
		log.Printf("Failed to update status: %v", err)
		return
	}
	defer resp.Body.Close()
}

func (s *EmbeddingService) Run(ctx context.Context) {
	log.Println("Embedding service: start consumer loop")

	for {
		select {
		case <-ctx.Done():
			log.Println("Embedding service: context done")
			return
		default:
			msg, err := s.consumer.ReadMessage(ctx)
			if err != nil {
				log.Printf("consumer read error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var evt models.OCRProcessedEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("invalid ocr_processed event: %v", err)
				continue
			}

			log.Printf("Processing OCR_PROCESSED for file %s (%s) with %d chunks", evt.FileID, evt.FileName, len(evt.Chunks))

			if len(evt.Chunks) == 0 {
				s.updateFileStatus(evt.FileID, "completed", 0, "No chunks to embed")
				continue
			}

			// Embed first chunk to determine dimension
			firstVec, err := s.xenova.Embed(evt.Chunks[0].Text)
			if err != nil {
				log.Printf("xenova embed failed for first chunk: %v", err)
				s.updateFileStatus(evt.FileID, "failed", 0, fmt.Sprintf("Embedding failed: %v", err))
				s.publishError(evt.FileID, err)
				continue
			}

			dimension := len(firstVec)
			if err := s.qrepo.EnsureCollection("documents", dimension); err != nil {
				log.Printf("failed to ensure collection: %v", err)
				s.updateFileStatus(evt.FileID, "failed", 0, fmt.Sprintf("Collection setup failed: %v", err))
				s.publishError(evt.FileID, err)
				continue
			}

			var stored []models.StoredChunkInfo

			for i, ch := range evt.Chunks {
				vec, err := s.xenova.Embed(ch.Text)
				if err != nil {
					log.Printf("xenova embed failed for chunk %s: %v", ch.ChunkID, err)
					continue
				}

				payload := cleanPayload(map[string]interface{}{
					"file_id":      evt.FileID,
					"file_name":    evt.FileName,
					"content_type": evt.ContentType,
					"chunk_id":     ch.ChunkID,
					"page":         ch.Page,
					"text":         ch.Text,
					"metadata":     ch.Metadata,
				})

				if _, err := s.qrepo.UpsertVector("documents", ch.ChunkID, vec, payload); err != nil {
					log.Printf("qdrant upsert failed for chunk %s: %v", ch.ChunkID, err)
					continue
				}

				stored = append(stored, models.StoredChunkInfo{
					ChunkID:  ch.ChunkID,
					VectorID: ch.ChunkID,
					Page:     ch.Page,
				})

				if i == 0 {
					log.Printf("First chunk embedded successfully")
				}
			}

			s.updateFileStatus(evt.FileID, "completed", len(stored), fmt.Sprintf("Successfully embedded %d chunks", len(stored)))

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
				log.Printf("failed to publish embedding_stored: %v", err)
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

func (s *EmbeddingService) RunHTTP(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	log.Printf("Embedding service HTTP listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
