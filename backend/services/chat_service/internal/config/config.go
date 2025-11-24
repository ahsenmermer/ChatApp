package config

import (
	"log"
	"os"
)

type Config struct {
	Port                   string
	OpenRouterKey          string
	KafkaBrokers           []string
	KafkaTopicChatMessages string
	KafkaTopicEmbedding    string // YENİ
	AuthServiceURL         string
	SubscriptionServiceURL string
	QdrantURL              string // YENİ
	XenovaURL              string // YENİ
}

func LoadConfig() *Config {
	port := os.Getenv("CHAT_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	openRouterKey := os.Getenv("OPENROUTER_KEY")
	if openRouterKey == "" {
		log.Fatal("OPENROUTER_KEY environment variable is required")
	}

	kafkaBrokersEnv := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersEnv == "" {
		kafkaBrokersEnv = "localhost:9092"
	}
	kafkaBrokers := []string{kafkaBrokersEnv}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "chat_messages"
	}

	kafkaTopicEmbedding := os.Getenv("KAFKA_TOPIC_EMBEDDING")
	if kafkaTopicEmbedding == "" {
		kafkaTopicEmbedding = "embedding_stored"
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8000"
	}

	subscriptionURL := os.Getenv("SUBSCRIPTION_SERVICE_URL")
	if subscriptionURL == "" {
		subscriptionURL = "http://localhost:8081"
	}

	qdrantURL := os.Getenv("QDRANT_URL")
	if qdrantURL == "" {
		qdrantURL = "http://localhost:6333"
	}

	xenovaURL := os.Getenv("XENOVA_URL")
	if xenovaURL == "" {
		xenovaURL = "http://localhost:3000"
	}

	return &Config{
		Port:                   port,
		OpenRouterKey:          openRouterKey,
		KafkaBrokers:           kafkaBrokers,
		KafkaTopicChatMessages: kafkaTopic,
		KafkaTopicEmbedding:    kafkaTopicEmbedding,
		AuthServiceURL:         authURL,
		SubscriptionServiceURL: subscriptionURL,
		QdrantURL:              qdrantURL,
		XenovaURL:              xenovaURL,
	}
}
