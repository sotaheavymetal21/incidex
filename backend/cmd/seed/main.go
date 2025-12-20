package main

import (
	"fmt"
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"log"
	"os"
)

func main() {
	log.Println("Incidex Database Seeder")
	log.Println("=======================")

	cfg := config.Load()

	// Connect to database
	dbConn := db.Connect(cfg.DatabaseURL)

	// Run migrations first
	log.Println("Running migrations...")
	if err := dbConn.AutoMigrate(&domain.User{}, &domain.Tag{}, &domain.Incident{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Run seed
	if err := db.Seed(dbConn); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("\n✅ Database seeding completed successfully!")

	os.Exit(0)
}
