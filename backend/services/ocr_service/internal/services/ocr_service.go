package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ocr_service/internal/events"
	"ocr_service/internal/extractor"
	"ocr_service/internal/handler"
	"ocr_service/internal/models"
)

type OCRService struct {
	consumer *events.Consumer
	producer *events.Producer
}

func NewOCRService(consumer *events.Consumer, producer *events.Producer) *OCRService {
	return &OCRService{consumer: consumer, producer: producer}
}

func (s *OCRService) Run(ctx context.Context) {
	log.Println("OCR Service: starting consumer loop")
	for {
		select {
		case <-ctx.Done():
			log.Println("OCR Service: context done, exiting consumer loop")
			return
		default:
			msg, err := s.consumer.ReadMessage(ctx)
			if err != nil {
				log.Printf("consumer read error: %v", err)
				time.Sleep(time.Second)
				continue
			}
			log.Printf("OCR Service: got message on topic %s", msg.Topic)
			var evt models.FileUploadedEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("invalid file_uploaded event: %v", err)
				continue
			}

			// Set initial status
			handler.SetFileStatus(evt.FileID, "processing", 0, "OCR processing started")

			log.Printf("Processing file: %s (ID: %s)", evt.FileName, evt.FileID)
			chunks, err := extractor.ProcessFile(evt.FilePath)
			if err != nil {
				log.Printf("error processing file %s: %v", evt.FilePath, err)

				// Update status to failed
				handler.SetFileStatus(evt.FileID, "failed", 0, err.Error())

				// Publish error event
				errorEvt := map[string]interface{}{
					"event":     "OCR_FAILED",
					"file_id":   evt.FileID,
					"error":     err.Error(),
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				}
				b, _ := json.Marshal(errorEvt)
				s.producer.Publish("ocr_failed", b)
				continue
			}

			if len(chunks) == 0 {
				log.Printf("No text extracted from file %s", evt.FileID)
				handler.SetFileStatus(evt.FileID, "completed", 0, "No text extracted")
				continue
			}

			// Update status - OCR completed, waiting for embedding
			handler.SetFileStatus(evt.FileID, "embedding", len(chunks), "OCR completed, embedding in progress")

			ocrEvt := models.OCRProcessedEvent{
				Event:       "OCR_PROCESSED",
				FileID:      evt.FileID,
				FileName:    evt.FileName,
				ContentType: evt.ContentType,
				Chunks:      chunks,
				Language:    "tur",
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			}
			b, _ := json.Marshal(ocrEvt)
			if err := s.producer.Publish("ocr_processed", b); err != nil {
				log.Printf("failed to publish ocr_processed: %v", err)
				handler.SetFileStatus(evt.FileID, "failed", len(chunks), "Failed to publish OCR result")
				continue
			}
			log.Printf("✓ Published OCR_PROCESSED for file %s with %d chunks", evt.FileID, len(chunks))
		}
	}
}
