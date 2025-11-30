package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"

	"subscription_service/internal/config"
	"subscription_service/internal/database"
	"subscription_service/internal/handler"
	"subscription_service/internal/migrations"
	"subscription_service/internal/repository"
	"subscription_service/internal/router"
	"subscription_service/internal/services"
)

// waitForKafka retries connecting to Kafka before giving up
func waitForKafka(brokers []string, retries int) error {
	for i := 0; i < retries; i++ {
		config := sarama.NewConfig()
		client, err := sarama.NewClient(brokers, config)
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
	// Load environment variables
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Printf("Config loaded: pg=%s:%d db=%s kafka=%v port=%s",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB, cfg.KafkaBrokers, cfg.ServicePort)

	// Connect to Postgres
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations (migrations.Run artık path parametresi alıyor)
	if err := migrations.Run(cfg.MigrationsPath); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Wait for Kafka to be ready
	if err := waitForKafka(cfg.KafkaBrokers, 12); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// Initialize repository
	subRepo := repository.NewPostgresSubscriptionRepository(database.DB)

	// Initialize service (Kafka consumer dahil)
	subService := services.NewUserSubscriptionService(
		subRepo,
		cfg.KafkaBrokers,
		cfg.KafkaTopicUserRegistered,
	)

	// Start Kafka consumer to listen for "user_registered" events
	if err := subService.StartKafkaConsumer(); err != nil {
		log.Fatalf("❌ Failed to start Kafka consumer: %v", err)
	}

	// Initialize HTTP handler
	subHandler := handler.NewSubscriptionHandler(subService)

	// Setup routes
	mux := http.NewServeMux()
	router.SetupSubscriptionRoutes(mux, subHandler)

	log.Printf("✅ Subscription Service running on :%s", cfg.ServicePort)
	if err := http.ListenAndServe(":"+cfg.ServicePort, mux); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
