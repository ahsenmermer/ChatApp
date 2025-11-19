package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	KafkaBrokers             []string
	KafkaTopicUserRegistered string

	ServicePort    string
	LogLevel       string
	MigrationsPath string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	cfg.PostgresHost = getenv("POSTGRES_HOST", "localhost")
	cfg.PostgresPort = getenvInt("POSTGRES_PORT", 5432)
	cfg.PostgresUser = getenv("POSTGRES_USER", "postgres")
	cfg.PostgresPassword = getenv("POSTGRES_PASSWORD", "postgres")
	cfg.PostgresDB = getenv("POSTGRES_DB", "subscription_db")

	cfg.KafkaBrokers = parseCSV(getenv("KAFKA_BROKERS", "localhost:9092"))
	cfg.KafkaTopicUserRegistered = getenv("KAFKA_TOPIC_USER_REGISTERED", "user_registered")

	cfg.ServicePort = getenv("SERVICE_PORT", "8081")
	cfg.LogLevel = getenv("LOG_LEVEL", "info")
	cfg.MigrationsPath = getenv("MIGRATIONS_PATH", "internal/migrations")

	if cfg.PostgresHost == "" || cfg.PostgresUser == "" || cfg.PostgresDB == "" {
		return nil, fmt.Errorf("postgres config incomplete")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka brokers not configured")
	}

	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func parseCSV(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
