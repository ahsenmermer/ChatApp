package config

import "os"

type Config struct {
	Port         string
	KafkaBrokers []string
	KafkaGroup   string
	QdrantURL    string
	XenovaURL    string
}

func LoadConfig() *Config {
	port := os.Getenv("EMBEDDING_SERVICE_PORT")
	if port == "" {
		port = "8400"
	}
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	group := os.Getenv("KAFKA_GROUP")
	if group == "" {
		group = "embedding-service-group"
	}
	qdrant := os.Getenv("QDRANT_URL")
	if qdrant == "" {
		qdrant = "http://localhost:6333"
	}
	xenova := os.Getenv("XENOVA_URL")
	if xenova == "" {
		xenova = "http://localhost:3000"
	}
	return &Config{
		Port:         port,
		KafkaBrokers: []string{brokersEnv},
		KafkaGroup:   group,
		QdrantURL:    qdrant,
		XenovaURL:    xenova,
	}
}
