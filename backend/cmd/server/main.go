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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize Database
	dbConn := db.Connect(cfg.DatabaseURL)

	// Auto Migration
	if err := dbConn.AutoMigrate(&domain.User{}, &domain.Tag{}, &domain.Incident{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Dependency Injection
	// Auth
	userRepo := persistence.NewUserRepository(dbConn)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret, 24*time.Hour)
	authHandler := handler.NewAuthHandler(authUsecase)
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	// Tags
	tagRepo := persistence.NewTagRepository(dbConn)
	tagUsecase := usecase.NewTagUsecase(tagRepo)
	tagHandler := handler.NewTagHandler(tagUsecase)

	// Incidents
	incidentRepo := persistence.NewIncidentRepository(dbConn)
	incidentUsecase := usecase.NewIncidentUsecase(incidentRepo, tagRepo, userRepo)
	incidentHandler := handler.NewIncidentHandler(incidentUsecase)

	// Users
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Incidex API is running",
		})
	})

	// Register Routes
	router.RegisterRoutes(r, authHandler, jwtMiddleware, tagHandler, incidentHandler, userHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
