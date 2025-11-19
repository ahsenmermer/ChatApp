package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"chat_service/internal/config"
	"chat_service/internal/router"

	"github.com/IBM/sarama"
)

// Kafka bağlantısı için retry
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

	// Router setup
	app := router.SetupRouter(cfg)

	// Kafka bağlantısı hazır olana kadar bekle
	if err := waitForKafka(cfg.KafkaBrokers, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// Kafka consumer başlat
	go startKafkaConsumer(cfg)

	log.Printf("Starting Chat Service on port %s...", cfg.Port)
	log.Fatal(app.Listen("0.0.0.0:" + cfg.Port))
}

func startKafkaConsumer(cfg *config.Config) {
	configKafka := sarama.NewConfig()
	configKafka.Consumer.Return.Errors = true
	configKafka.Version = sarama.V2_8_0_0

	// Kafka consumer başlat
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

	log.Printf("🎧 Chat service Kafka consumer started for topic: %s", cfg.KafkaTopicChatMessages)

	for msg := range consumer.Messages() {
		log.Printf("📥 Kafka message received: %s", string(msg.Value))

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
			log.Printf("💡 Quota changed event for user %s", event.UserID)
			// memory veya cache güncelleme yapılabilir
		default:
			// Eğer chat_completed gibi event'ler gelirse burada işleyebilirsin
			log.Printf("ℹ️ Unknown event type: %s ; conversation_id=%s", event.Type, event.ConversationID)
		}
	}
}
