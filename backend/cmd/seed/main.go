package main

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/pkg/logger"
	"log"
)

func main() {
	// ロガーを初期化します
	env := logger.GetEnv()
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg := config.Load()

	// ロガーを使用してデータベースに接続します
	dbConn := db.Connect(cfg.DatabaseURL, logger.Log, cfg.DBLogLevel)

	_, err := db.Seed(dbConn)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
}
