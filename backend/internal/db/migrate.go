package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq" // PostgreSQL ドライバ
	"github.com/pressly/goose/v3"
)

// RunMigrations は goose を使用してデータベースマイグレーションを実行します
// migrationsDir: マイグレーションディレクトリへのパス（例: "./migrations"）
// databaseURL: PostgreSQL 接続文字列
func RunMigrations(migrationsDir string, databaseURL string) error {
	log.Println("INFO: Starting database migrations...")

	// マイグレーション用のデータベース接続を開きます
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	// データベース接続を検証します
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// マイグレーションディレクトリが存在するか確認します
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
	}

	// 絶対パスを取得します
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	log.Printf("INFO: Migrations directory: %s", absPath)

	// goose の方言を設定します
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// マイグレーションを実行します
	if err := goose.Up(db, absPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// 現在のバージョンを取得します
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	log.Printf("SUCCESS: Database migrations completed. Current version: %d", version)
	return nil
}

// GetMigrationStatus は現在のマイグレーション状況を返します
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
