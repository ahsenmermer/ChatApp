package config

import "os"

type Config struct {
	Port               string
	ClickHouseHost     string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string
	KafkaBrokers       string
	KafkaTopic         string
}

func Load() (*Config, error) {
	return &Config{
		Port:               getEnv("PORT", "8082"),
		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "clickhouse:9000"),
		ClickHouseDB:       getEnv("CLICKHOUSE_DB", "chat_data"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "kafka:9092"),
		KafkaTopic:         getEnv("KAFKA_TOPIC", "chat_messages"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
