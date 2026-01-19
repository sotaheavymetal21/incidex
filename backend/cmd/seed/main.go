package main

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/pkg/logger"
	"log"
)

func main() {
	// Initialize Logger
	env := logger.GetEnv()
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg := config.Load()
	
	// Determine if production environment
	isProduction := cfg.AppEnv == "production" || cfg.AppEnv == "prod"
	
	// Connect to database with logger
	dbConn := db.Connect(cfg.DatabaseURL, logger.Log, isProduction)

	_, err := db.Seed(dbConn)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
}
