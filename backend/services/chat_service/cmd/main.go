package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"chat_service/internal/config"
	"chat_service/internal/models"
	"chat_service/internal/router"
	"chat_service/internal/services"

	"github.com/IBM/sarama"
)

var chatSvc *services.ChatService

func waitForKafka(brokers []string, retries int) error {
	for i := 0; i < retries; i++ {
		client, err := sarama.NewClient(brokers, sarama.NewConfig())
		if err == nil {
			client.Close()
			return nil
		}
		log.Printf("Kafka not ready yet, retrying... (%d/%d)", i+1, retries)
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("kafka not available after %d retries", retries)
}

func main() {
	cfg := config.LoadConfig()

	chatSvc = services.NewChatService(cfg)

	app := router.SetupRouter(chatSvc)

	if err := waitForKafka(cfg.KafkaBrokers, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	go startChatConsumer(cfg)
	go startEmbeddingConsumer(cfg)

	log.Printf("Starting Chat Service on port %s...", cfg.Port)
	log.Fatal(app.Listen("0.0.0.0:" + cfg.Port))
}

func startChatConsumer(cfg *config.Config) {
	configKafka := sarama.NewConfig()
	configKafka.Consumer.Return.Errors = true
	configKafka.Version = sarama.V2_8_0_0

	master, err := sarama.NewConsumer(cfg.KafkaBrokers, configKafka)
	if err != nil {
		log.Fatalf("failed to start Kafka consumer: %v", err)
	}
	defer master.Close()

	consumer, err := master.ConsumePartition(cfg.KafkaTopicChatMessages, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("failed to consume partition: %v", err)
	}
	defer consumer.Close()

	log.Printf("🎧 Chat consumer started for topic: %s", cfg.KafkaTopicChatMessages)

	for msg := range consumer.Messages() {
		var event struct {
			Type           string `json:"type"`
			UserID         string `json:"user_id"`
			Message        string `json:"message,omitempty"`
			ConversationID string `json:"conversation_id,omitempty"`
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("⚠️ Failed to unmarshal event: %v", err)
			continue
		}

		switch event.Type {
		case "quota_changed":
			log.Printf("💡 Quota changed for user %s", event.UserID)
		default:
			log.Printf("ℹ️ Event: %s ; conversation=%s", event.Type, event.ConversationID)
		}
	}
}

func startEmbeddingConsumer(cfg *config.Config) {
	configKafka := sarama.NewConfig()
	configKafka.Consumer.Return.Errors = true
	configKafka.Version = sarama.V2_8_0_0

	master, err := sarama.NewConsumer(cfg.KafkaBrokers, configKafka)
	if err != nil {
		log.Fatalf("failed to start embedding consumer: %v", err)
	}
	defer master.Close()

	topic := "embedding_stored"

	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("failed to consume embedding partition: %v", err)
	}
	defer consumer.Close()

	log.Printf("🎧 Embedding consumer started for topic: %s (reading from oldest)", topic)

	for msg := range consumer.Messages() {
		log.Printf("📥 Raw Kafka message (offset=%d): %s", msg.Offset, string(msg.Value))

		var evt models.EmbeddingStoredEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			log.Printf("⚠️ Failed to unmarshal embedding event: %v", err)
			continue
		}

		if evt.Event == "EMBEDDING_STORED" {
			log.Printf("📥 EMBEDDING_STORED event received: file_id=%s, chunks=%d",
				evt.FileID, evt.TotalChunks)

			// FileTracker'ı güncelle
			chatSvc.GetFileTracker().UpdateStatus(
				evt.FileID,
				evt.FileName,
				"ready",
				evt.TotalChunks,
			)

			log.Printf("✅ File ready: %s (%s) with %d chunks",
				evt.FileID, evt.FileName, evt.TotalChunks)

			// ✅ YENİ: User bilgisi için 5 saniye bekle (retry)
			var fileInfo *models.FileStatus
			maxRetries := 5
			for i := 0; i < maxRetries; i++ {
				fileInfo = chatSvc.GetFileTracker().GetFileInfo(evt.FileID)
				if fileInfo != nil && fileInfo.UserID != "" {
					break // Bulundu!
				}
				log.Printf("⏳ Waiting for user info... (attempt %d/%d)", i+1, maxRetries)
				time.Sleep(1 * time.Second)
			}

			if fileInfo != nil && fileInfo.UserID != "" {
				// Kafka'ya file_attached eventi gönder
				err := chatSvc.GetKafkaProducer().PublishFileAttached(
					fileInfo.UserID,
					evt.FileName,
					evt.FileID,
					fileInfo.ConversationID,
				)
				if err != nil {
					log.Printf("⚠️ Failed to publish file_attached event: %v", err)
				} else {
					log.Printf("✅ File attachment message sent to chat history")
				}
			} else {
				log.Printf("⚠️ No user info found for file %s after waiting, skipping chat history entry", evt.FileID)
			}
		} else {
			log.Printf("⚠️ Unknown event type: '%s'", evt.Event)
		}
	}
}
