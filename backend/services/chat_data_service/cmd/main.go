package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"chat_data_service/internal/config"
	"chat_data_service/internal/database"
	"chat_data_service/internal/migrations"
	"chat_data_service/internal/models"
	"chat_data_service/internal/repository"
	"chat_data_service/internal/router"
	"chat_data_service/internal/services"

	"github.com/IBM/sarama"
)

func waitForKafka(brokers []string, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		client, err := sarama.NewClient(brokers, sarama.NewConfig())
		if err == nil {
			client.Close()
			log.Printf("✅ Kafka erişilebilir")
			return nil
		}
		log.Printf("Kafka hazır değil, yeniden dene... (%d/%d)", i+1, retries)
		time.Sleep(delay)
	}
	return fmt.Errorf("Kafka %d deneme sonunda erişilemedi", retries)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config yüklenemedi: %v", err)
	}

	conn, err := database.ConnectWithRetry(cfg, 10, 5)
	if err != nil {
		log.Fatalf("❌ ClickHouse bağlantısı başarısız: %v", err)
	}
	defer conn.Close()

	if err := migrations.RunMigrations(conn); err != nil {
		log.Fatalf("❌ Migration hatası: %v", err)
	}

	repo := repository.NewChatRepository(conn)
	service := services.NewChatDataService(repo)
	r := router.SetupRouter(service)

	if err := waitForKafka(strings.Split(cfg.KafkaBrokers, ","), 12, 5*time.Second); err != nil {
		log.Fatalf("❌ %v", err)
	}

	go startKafkaConsumer(cfg, service)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Chat Data Service %s portunda çalışıyor...", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Sunucu başlatılamadı: %v", err)
	}
}

func startKafkaConsumer(cfg *config.Config, service *services.ChatDataService) {
	log.Printf("🎧 Kafka consumer başlatılıyor (topic=%s, broker=%s)...", cfg.KafkaTopic, cfg.KafkaBrokers)

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Consumer.Return.Errors = true
	kafkaCfg.Version = sarama.V2_8_0_0

	consumer, err := sarama.NewConsumer(strings.Split(cfg.KafkaBrokers, ","), kafkaCfg)
	if err != nil {
		log.Fatalf("❌ Kafka consumer oluşturulamadı: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(cfg.KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("❌ Kafka partition dinlenemedi: %v", err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		log.Printf("📥 Kafka mesajı alındı: %s", string(msg.Value))

		var rawEvent map[string]interface{}
		if err := json.Unmarshal(msg.Value, &rawEvent); err != nil {
			log.Printf("❌ Invalid JSON: %v", err)
			continue
		}

		eventType, _ := rawEvent["event_type"].(string)

		switch eventType {
		case "chat_completed":
			var event struct {
				EventType      string `json:"event_type"`
				UserID         string `json:"user_id"`
				Message        string `json:"message"`
				Response       string `json:"response"`
				ConversationID string `json:"conversation_id"`
				Timestamp      string `json:"timestamp"`
			}

			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("⚠️ JSON çözümlenemedi: %v", err)
				continue
			}

			ts := time.Now().UTC()
			if event.Timestamp != "" {
				t, err := time.Parse(time.RFC3339, event.Timestamp)
				if err == nil {
					ts = t
				} else {
					log.Printf("⚠️ Timestamp parse hatası: %v", err)
				}
			}

			chMsg := &models.ChatMessage{
				UserID:         event.UserID,
				UserMessage:    event.Message,
				AIResponse:     event.Response,
				ConversationID: event.ConversationID,
				Timestamp:      ts,
			}

			if err := service.SaveMessage(chMsg); err != nil {
				log.Printf("❌ ClickHouse kaydedilemedi: %v", err)
			} else {
				log.Printf("✅ Mesaj ClickHouse'a kaydedildi: user=%s, user_msg=%s, ai_msg=%s",
					event.UserID, chMsg.UserMessage, chMsg.AIResponse)
			}

		case "file_attached": // ✅ YENİ
			var fileEvent struct {
				EventType      string `json:"event_type"`
				UserID         string `json:"user_id"`
				Message        string `json:"message"`
				ConversationID string `json:"conversation_id"`
				FileID         string `json:"file_id"`
				FileName       string `json:"file_name"`
				Timestamp      string `json:"timestamp"`
			}

			if err := json.Unmarshal(msg.Value, &fileEvent); err != nil {
				log.Printf("❌ Failed to parse file_attached event: %v", err)
				continue
			}

			// Timestamp parse
			timestamp := time.Now().UTC()
			if fileEvent.Timestamp != "" {
				t, err := time.Parse(time.RFC3339, fileEvent.Timestamp)
				if err == nil {
					timestamp = t
				}
			}

			// ClickHouse'a kaydet (AI response boş)
			fileMsg := &models.ChatMessage{
				UserID:         fileEvent.UserID,
				UserMessage:    fileEvent.Message, // "📎 öneri.pdf"
				AIResponse:     "",                // AI response yok
				ConversationID: fileEvent.ConversationID,
				Timestamp:      timestamp,
			}

			if err := service.SaveMessage(fileMsg); err != nil {
				log.Printf("❌ File attachment message save error: %v", err)
				continue
			}

			log.Printf("✅ File attachment message saved: user=%s, file=%s, conversation=%s",
				fileEvent.UserID, fileEvent.FileName, fileEvent.ConversationID)

		default:
			log.Printf("⚠️ Unknown event type: %s", eventType)
		}
	}
}
