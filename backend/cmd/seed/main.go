package main

import (
	"fmt"
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"log"
	"os"
	"strings"
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
	password, err := db.Seed(dbConn)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("\n✅ Database seeding completed successfully!")

	// Display password and user information if password was generated
	if password != "" {
		separator := strings.Repeat("=", 50)
		fmt.Println("\n" + separator)
		fmt.Println("📋 Test User Accounts")
		fmt.Println(separator)
		fmt.Println("All test users use the same password:")
		fmt.Printf("  Password: %s\n", password)
		fmt.Println("\nTest Users:")
		fmt.Println("  - admin@example.com (Admin)")
		fmt.Println("  - editor1@example.com (Editor)")
		fmt.Println("  - editor2@example.com (Editor)")
		fmt.Println("  - viewer1@example.com (Viewer)")
		fmt.Println("  - viewer2@example.com (Viewer)")
		fmt.Println("\n💡 Tip: Set TEST_USER_PASSWORD environment variable to use a custom password.")
		fmt.Println(separator)
	}

	os.Exit(0)
}
