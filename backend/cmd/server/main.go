package main

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/ai"
	"incidex/internal/infrastructure/notification"
	"incidex/internal/infrastructure/persistence"
	"incidex/internal/infrastructure/storage"
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
	if err := dbConn.AutoMigrate(&domain.User{}, &domain.Tag{}, &domain.Incident{}, &domain.IncidentActivity{}, &domain.Attachment{}, &domain.NotificationSetting{}, &domain.IncidentTemplate{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize MinIO Storage
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		storage.DefaultBucketName,
		false, // useSSL = false for local development
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
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

	// Incident Activities
	activityRepo := persistence.NewIncidentActivityRepository(dbConn)

	// Notifications
	notificationRepo := persistence.NewNotificationSettingRepository(dbConn)
	notificationService := notification.NewNotificationService(notificationRepo, userRepo)
	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)

	// AI Service
	aiService := ai.NewOpenAIService()

	// Incidents
	incidentRepo := persistence.NewIncidentRepository(dbConn)
	incidentUsecase := usecase.NewIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, notificationService, aiService)
	incidentHandler := handler.NewIncidentHandler(incidentUsecase)

	// Users
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	// Stats
	statsUsecase := usecase.NewStatsUsecase(incidentRepo)
	statsHandler := handler.NewStatsHandler(statsUsecase)

	// Activity handler
	activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, notificationService)
	activityHandler := handler.NewIncidentActivityHandler(activityUsecase)

	// Export
	exportHandler := handler.NewExportHandler(incidentUsecase)

	// Attachments
	attachmentRepo := persistence.NewAttachmentRepository(dbConn)
	attachmentUsecase := usecase.NewAttachmentUsecase(attachmentRepo, incidentRepo, minioStorage)
	attachmentHandler := handler.NewAttachmentHandler(attachmentUsecase)

	// Templates
	templateRepo := persistence.NewIncidentTemplateRepository(dbConn)
	templateUsecase := usecase.NewIncidentTemplateUsecase(templateRepo, tagRepo, incidentRepo, userRepo)
	templateHandler := handler.NewIncidentTemplateHandler(templateUsecase)

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
	router.RegisterRoutes(r, authHandler, jwtMiddleware, tagHandler, incidentHandler, userHandler, statsHandler, activityHandler, exportHandler, attachmentHandler, notificationHandler, templateHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
