package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/sarama"

	"embedding_service/internal/config"
	"embedding_service/internal/events"
	"embedding_service/internal/repository"
	"embedding_service/internal/services"
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

// waitForXenova waits for Xenova HTTP server to become healthy
func waitForXenova(url string, retries int) error {
	client := &http.Client{Timeout: 3 * time.Second}

	for i := 0; i < retries; i++ {
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			log.Println("✅ Xenova is ready!")
			return nil
		}

		log.Printf("⏳ Xenova not ready yet, retrying... (%d/%d)", i+1, retries)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("xenova not available after %d retries", retries)
}

func main() {
	cfg := config.LoadConfig()

	// Wait for Kafka to be ready
	if err := waitForKafka(cfg.KafkaBrokers, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// 2) Wait Xenova
	if err := waitForXenova(cfg.XenovaURL, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// Kafka
	producer, err := events.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to new producer: %v", err)
	}
	defer producer.Close()

	consumer, err := events.NewConsumer(cfg.KafkaBrokers, "ocr_processed", cfg.KafkaGroup)
	if err != nil {
		log.Fatalf("failed to new consumer: %v", err)
	}
	defer consumer.Close()

	// Qdrant repo
	qrepo := repository.NewQdrantRepo(cfg.QdrantURL)

	// Xenova client
	xcli := services.NewXenovaClient(cfg.XenovaURL)

	embSvc := services.NewEmbeddingService(consumer, producer, qrepo, xcli)

	// run consumer
	ctx, cancel := context.WithCancel(context.Background())
	go embSvc.Run(ctx)

	// Simple HTTP health server
	go func() {
		if err := embSvc.RunHTTP(":" + cfg.Port); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	log.Printf("✅ Embedding Service running on :%s", cfg.Port)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down Embedding service...")
	cancel()
	time.Sleep(1 * time.Second)
}
