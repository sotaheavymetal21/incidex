package main

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/persistence"
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"
	"incidex/internal/interface/http/router"
	"incidex/internal/usecase"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize Database
	dbConn := db.Connect(cfg.DatabaseURL)

	// Auto Migration
	if err := dbConn.AutoMigrate(&domain.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Dependency Injection
	userRepo := persistence.NewUserRepository(dbConn)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret, 24*time.Hour)
	authHandler := handler.NewAuthHandler(authUsecase)
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	r := gin.Default()

	// Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Incidex API is running",
		})
	})

	// Register Routes
	router.RegisterRoutes(r, authHandler, jwtMiddleware)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
