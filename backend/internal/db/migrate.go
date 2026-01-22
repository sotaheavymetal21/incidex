package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

// RunMigrations executes database migrations using goose
// migrationsDir: path to the migrations directory (e.g., "./migrations")
// databaseURL: PostgreSQL connection string
func RunMigrations(migrationsDir string, databaseURL string) error {
	log.Println("INFO: Starting database migrations...")

	// Open database connection for migrations
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
	}

	// Get absolute path
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	log.Printf("INFO: Migrations directory: %s", absPath)

	// Set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db, absPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get current version
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	log.Printf("SUCCESS: Database migrations completed. Current version: %d", version)
	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(migrationsDir string, databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	return goose.Status(db, absPath)
}
