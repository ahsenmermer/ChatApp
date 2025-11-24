package config

import "os"

type Config struct {
	Port         string
	KafkaBrokers []string
	KafkaGroup   string
}

func LoadConfig() *Config {
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
