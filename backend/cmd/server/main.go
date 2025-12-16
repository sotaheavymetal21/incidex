package main

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize Database
	// We are not using the db instance yet, but establishing connection to verify it works
	_ = db.Connect(cfg.DatabaseURL)

	r := gin.Default()

	// Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Incidex API is running",
		})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
