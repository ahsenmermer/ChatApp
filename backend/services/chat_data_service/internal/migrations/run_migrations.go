package migrations

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func RunMigrations(conn clickhouse.Conn) error {
	file := filepath.Join("internal", "migrations", "001_create_tables.sql")
	sql, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return conn.Exec(ctx, string(sql))
}
