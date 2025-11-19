package config

import "os"

type Config struct {
	Port      string
	QdrantURL string
	ModelPath string
}

func LoadConfig() *Config {
	port := os.Getenv("EMBEDDING_SERVICE_PORT")
	if port == "" {
		port = "8400"
	}

	qdrant := os.Getenv("QDRANT_URL")
	if qdrant == "" {
		qdrant = "http://localhost:6333"
	}

	model := os.Getenv("XENOVA_MODEL_PATH")
	if model == "" {
		model = "./internal/embedding/model"
	}

	return &Config{
		Port:      port,
		QdrantURL: qdrant,
		ModelPath: model,
	}
}
