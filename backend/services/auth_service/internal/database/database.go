package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"auth_service/internal/config"
)

var DB *sqlx.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Println("✅ Database connected successfully")
	return nil
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("⚠️ Error closing DB: %v", err)
		} else {
			log.Println("🧹 Database connection closed")
		}
	}
}

// RunMigrations executes all SQL files
func RunMigrations(path string) error {
	files, err := filepath.Glob(fmt.Sprintf("%s/*.sql", path))
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read file %s: %w", file, err)
		}

		if _, err := DB.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec migration %s: %w", file, err)
		}

		log.Printf("🚀 Migration applied: %s", file)
	}

	log.Println("✅ All migrations applied successfully")
	return nil
}
