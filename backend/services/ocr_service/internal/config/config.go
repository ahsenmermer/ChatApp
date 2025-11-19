package config

import "os"

type Config struct {
	Port string
}

// LoadConfig config değerlerini environment variable veya default ile yükler
func LoadConfig() *Config {
	port := os.Getenv("OCR_SERVICE_PORT")
	if port == "" {
		port = "8090" // default port
	}

	return &Config{
		Port: port,
	}
}
