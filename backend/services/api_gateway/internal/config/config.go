package config

import "os"

type Config struct {
	AuthServiceURL         string
	SubscriptionServiceURL string
	ChatServiceURL         string
	ChatDataServiceURL     string
	OCRServiceURL          string
	EmbeddingServiceURL    string
	Port                   string
}

// Load reads from env and returns Config (fallbacks provided)
func Load() Config {
	return Config{
		AuthServiceURL:         getEnv("AUTH_SERVICE_URL", "http://auth_service:8080"),
		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://subscription_service:8081"),
		ChatServiceURL:         getEnv("CHAT_SERVICE_URL", "http://chat_service:8082"),
		ChatDataServiceURL:     getEnv("CHAT_DATA_SERVICE_URL", "http://chat_data_service:8083"),
		OCRServiceURL:          getEnv("OCR_SERVICE_URL", "http://ocr_service:8090"),
		EmbeddingServiceURL:    getEnv("EMBEDDING_SERVICE_URL", "http://embedding_service:8400"),
		Port:                   getEnv("GATEWAY_PORT", "8085"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
