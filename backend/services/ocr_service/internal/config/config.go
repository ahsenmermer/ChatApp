package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	KafkaBrokers []string
	KafkaGroup   string
}

func LoadConfig() *Config {
	// 🆕 .env dosyasını yükle
	_ = godotenv.Load("internal/config/.env")

	port := os.Getenv("OCR_SERVICE_PORT")
	if port == "" {
		port = "8090"
	}
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "kafka:9092"
	}
	group := os.Getenv("KAFKA_GROUP")
	if group == "" {
		group = "ocr-service-group"
	}

	return &Config{
		Port:         port,
		KafkaBrokers: []string{brokersEnv},
		KafkaGroup:   group,
	}
}
