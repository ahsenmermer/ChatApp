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
	"embedding_service/internal/router" // 🆕 Router import edildi
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

// waitForQdrant waits for Qdrant to become healthy
func waitForQdrant(url string, retries int) error {
	client := &http.Client{Timeout: 3 * time.Second}

	for i := 0; i < retries; i++ {
		// Qdrant'ın root endpoint'i (/) kullanılmalı - /health endpoint'i yok
		resp, err := client.Get(url + "/")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			log.Println("✅ Qdrant is ready!")
			return nil
		}

		if err != nil {
			log.Printf("⏳ Qdrant not ready yet (error: %v), retrying... (%d/%d)", err, i+1, retries)
		} else {
			log.Printf("⏳ Qdrant not ready yet (status: %d), retrying... (%d/%d)", resp.StatusCode, i+1, retries)
			resp.Body.Close()
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("qdrant not available after %d retries", retries)
}

func main() {
	cfg := config.LoadConfig()

	// Wait for dependencies
	log.Println("⏳ Waiting for dependencies...")

	// 1) Wait for Kafka
	if err := waitForKafka(cfg.KafkaBrokers, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// 2) Wait for Qdrant
	if err := waitForQdrant(cfg.QdrantURL, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// 3) Wait for Xenova
	if err := waitForXenova(cfg.XenovaURL, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// Kafka producer
	producer, err := events.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	// Kafka consumer
	consumer, err := events.NewConsumer(cfg.KafkaBrokers, "ocr_processed", cfg.KafkaGroup)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Qdrant repo
	qrepo := repository.NewQdrantRepo(cfg.QdrantURL)

	// Xenova client
	xcli := services.NewXenovaClient(cfg.XenovaURL)

	// Embedding service
	embSvc := services.NewEmbeddingService(consumer, producer, qrepo, xcli)

	// Search service
	searchSvc := services.NewSearchService(qrepo, xcli)

	// Run consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	go embSvc.Run(ctx)

	// 🆕 Setup router
	handler := router.SetupRouter(searchSvc)

	// HTTP server with search endpoints
	go func() {
		addr := ":" + cfg.Port
		log.Printf("✅ Embedding Service HTTP running on %s", addr)
		log.Printf("📍 Search endpoint: POST http://localhost:%s/api/search", cfg.Port)
		log.Printf("📍 File search: POST http://localhost:%s/api/search/file?file_id=xxx", cfg.Port)

		// 🆕 Router kullanılıyor (embSvc.RunHTTP yerine)
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	log.Printf("✅ Embedding Service running on :%s", cfg.Port)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("🛑 Shutting down Embedding service...")
	cancel()
	time.Sleep(1 * time.Second)
}
