package migrations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"auth_service/internal/database"
)

// Run executes all SQL files in the migrations folder
func Run() error {
	db := database.DB
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	migrationsDir := "internal/migrations"

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		sqlStatements := string(sqlBytes)
		if strings.TrimSpace(sqlStatements) == "" {
			log.Printf("⚠️  %s is empty, skipping\n", file)
			continue
		}

		log.Printf("🚀 Running migration: %s\n", file)
		if _, err := db.Exec(sqlStatements); err != nil {
			return fmt.Errorf("failed to execute %s: %w", file, err)
		}
	}

	log.Println("✅ All migrations applied successfully!")
	return nil
}
