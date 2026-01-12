package main

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"log"
)

func main() {
	cfg := config.Load()
	dbConn := db.Connect(cfg.DatabaseURL)

	_, err := db.Seed(dbConn)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
}
