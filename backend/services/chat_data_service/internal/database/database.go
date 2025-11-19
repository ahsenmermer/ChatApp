package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"chat_data_service/internal/config"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// ConnectWithRetry → ClickHouse bağlantısı kurulana kadar bekler (retry mekanizmalı)
func ConnectWithRetry(cfg *config.Config, retries int, delaySeconds int) (clickhouse.Conn, error) {
	var conn clickhouse.Conn
	var err error

	for i := 1; i <= retries; i++ {
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{cfg.ClickHouseHost},
			Auth: clickhouse.Auth{
				Database: cfg.ClickHouseDB,
				Username: cfg.ClickHouseUser,
				Password: cfg.ClickHousePassword,
			},
			Debug:       true,
			DialTimeout: 5 * time.Second,
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
		})

		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if pingErr := conn.Ping(ctx); pingErr == nil {
				log.Printf("✅ ClickHouse bağlantısı başarılı (deneme %d/%d)", i, retries)
				return conn, nil
			} else {
				log.Printf("⚠️ ClickHouse henüz hazır değil (deneme %d/%d): %v", i, retries, pingErr)
			}
		} else {
			log.Printf("⚠️ ClickHouse bağlantı hatası (deneme %d/%d): %v", i, retries, err)
		}

		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}

	return nil, fmt.Errorf("ClickHouse bağlantısı %d denemede başarısız oldu: %v", retries, err)
}
