package migrations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"auth_service/internal/database"
)

// Run executes all SQL files in the migrations folder
func Run(migrationsPath string) error {
	db := database.DB
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Get all .sql files from the provided path
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	// Sort files to ensure they run in order (001, 002, etc.)
	sort.Strings(files)

	if len(files) == 0 {
		log.Printf("⚠️  No migration files found in %s\n", migrationsPath)
		return nil
	}

	log.Printf("📂 Found %d migration file(s) in %s\n", len(files), migrationsPath)

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		sqlStatements := string(sqlBytes)
		if strings.TrimSpace(sqlStatements) == "" {
			log.Printf("⚠️  %s is empty, skipping\n", filepath.Base(file))
			continue
		}

		log.Printf("🚀 Running migration: %s\n", filepath.Base(file))
		if _, err := db.Exec(sqlStatements); err != nil {
			return fmt.Errorf("failed to execute %s: %w", filepath.Base(file), err)
		}
	}

	log.Println("✅ All migrations applied successfully!")
	return nil
}
