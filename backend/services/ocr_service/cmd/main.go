package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"

	"ocr_service/internal/config"
	"ocr_service/internal/events"
	"ocr_service/internal/handler"
	"ocr_service/internal/router"
	"ocr_service/internal/services"
)

// waitForKafka retries connecting to Kafka before giving up
func waitForKafka(brokers []string, retries int) error {
	for i := 0; i < retries; i++ {
		config := sarama.NewConfig()
		client, err := sarama.NewClient(brokers, config)
		if err == nil {
			client.Close()
			log.Println("✅ Kafka is ready!")
			return nil
		}
		log.Printf("⏳ Kafka not ready yet, retrying... (%d/%d)", i+1, retries)
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("kafka not available after %d retries", retries)
}

func main() {
	// 🆕 .env yükle
	_ = godotenv.Load("internal/config/.env")

	cfg := config.LoadConfig()

	// Wait for Kafka to be ready
	if err := waitForKafka(cfg.KafkaBrokers, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// Kafka
	producer, err := events.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	consumer, err := events.NewConsumer(cfg.KafkaBrokers, "file_uploaded", cfg.KafkaGroup)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Services
	ocrSvc := services.NewOCRService(consumer, producer)

	// Handler
	uploadHandler := handler.NewUploadHandler(producer)

	// Kafka consumer background loop
	ctx, cancel := context.WithCancel(context.Background())
	go ocrSvc.Run(ctx)

	// HTTP server
	go func() {
		r := router.SetupRouter(uploadHandler, producer)
		addr := ":" + cfg.Port
		log.Printf("✅ OCR Service running on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("🛑 Shutting down OCR service...")
	cancel()
	time.Sleep(time.Second)
}
