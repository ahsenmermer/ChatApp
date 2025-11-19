package config

import (
	"log"
	"os"
)

// Config: Chat service config struct
type Config struct {
	Port                   string
	OpenRouterKey          string
	KafkaBrokers           []string
	KafkaTopicChatMessages string
	AuthServiceURL         string
	SubscriptionServiceURL string
}

// LoadConfig: Çevresel değişkenleri okuyup Config struct'ını döner
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

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8000"
	}

	subscriptionURL := os.Getenv("SUBSCRIPTION_SERVICE_URL")
	if subscriptionURL == "" {
		subscriptionURL = "http://localhost:8081"
	}

	return &Config{
		Port:                   port,
		OpenRouterKey:          openRouterKey,
		KafkaBrokers:           kafkaBrokers,
		KafkaTopicChatMessages: kafkaTopic,
		AuthServiceURL:         authURL,
		SubscriptionServiceURL: subscriptionURL,
	}
}
