package migrations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"subscription_service/internal/database"
)

// RunMigrations executes all SQL files in the given path
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

		if _, err := database.DB.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec migration %s: %w", file, err)
		}

		log.Printf("🚀 Migration applied: %s", file)
	}

	log.Println("✅ All migrations applied successfully")
	return nil
}
