package main

import (
	"context"
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/cache"
	"incidex/internal/infrastructure/notification"
	"incidex/internal/infrastructure/persistence"
	"incidex/internal/infrastructure/storage"
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"
	"incidex/internal/interface/http/router"
	"incidex/internal/pkg/logger"
	"incidex/internal/usecase"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Initialize Logger
	env := logger.GetEnv()
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Incidex server", zap.String("environment", env))

	cfg := config.Load()

	// Initialize Database
	dbConn := db.Connect(cfg.DatabaseURL)

	// Database migrations are managed by goose
	// Run 'make migrate-up' (local) or 'make migrate-docker-up' (Docker) to apply migrations
	log.Println("INFO: Database migrations are managed by goose.")
	log.Println("Run 'make migrate-up' or 'make migrate-docker-up' to apply database migrations.")

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

	// Initialize Redis Cache
	redisClient := db.ConnectRedis(cfg.RedisURL)
	cacheRepo := cache.NewRedisCache(redisClient)

	// Dependency Injection
	// Auth
	userRepo := persistence.NewUserRepository(dbConn)

	// Create initial admin user if configured and no users exist
	createInitialAdminIfNeeded(dbConn, userRepo, cfg)

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

	// Incidents
	incidentRepo := persistence.NewIncidentRepository(dbConn)

	// Users
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	// Stats
	statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
	statsHandler := handler.NewStatsHandler(statsUsecase)

	// Activity handler
	activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, notificationService)
	activityHandler := handler.NewIncidentActivityHandler(activityUsecase)

	// Attachments
	attachmentRepo := persistence.NewAttachmentRepository(dbConn)
	attachmentUsecase := usecase.NewAttachmentUsecase(attachmentRepo, incidentRepo, minioStorage)
	attachmentHandler := handler.NewAttachmentHandler(attachmentUsecase)

	// Templates
	templateRepo := persistence.NewIncidentTemplateRepository(dbConn)
	templateUsecase := usecase.NewIncidentTemplateUsecase(templateRepo, tagRepo, incidentRepo, userRepo)
	templateHandler := handler.NewIncidentTemplateHandler(templateUsecase)

	// Post-mortems
	postMortemRepo := persistence.NewPostMortemRepository(dbConn)
	postMortemUsecase := usecase.NewPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)
	postMortemHandler := handler.NewPostMortemHandler(postMortemUsecase)

	// Initialize IncidentUsecase after PostMortemRepo is available
	incidentUsecase := usecase.NewIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, notificationService, cacheRepo)
	incidentHandler := handler.NewIncidentHandler(incidentUsecase)

	// Export
	exportHandler := handler.NewExportHandler(incidentUsecase)

	// Action items
	actionItemRepo := persistence.NewActionItemRepository(dbConn)
	actionItemUsecase := usecase.NewActionItemUsecase(actionItemRepo, postMortemRepo)
	actionItemHandler := handler.NewActionItemHandler(actionItemUsecase)

	// Audit logs
	auditLogRepo := persistence.NewAuditLogRepository(dbConn)
	auditLogUsecase := usecase.NewAuditLogUsecase(auditLogRepo)
	auditLogHandler := handler.NewAuditLogHandler(auditLogUsecase)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogRepo, userRepo)

	// Reports
	reportRepo := persistence.NewReportRepository(dbConn)
	reportUsecase := usecase.NewReportUsecase(reportRepo)
	reportHandler := handler.NewReportHandler(reportUsecase, incidentUsecase)

	r := gin.Default()

	// Audit log middleware (before CORS)
	r.Use(auditMiddleware.Log())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
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
	router.RegisterRoutes(r, authHandler, jwtMiddleware, tagHandler, incidentHandler, userHandler, statsHandler, activityHandler, exportHandler, attachmentHandler, notificationHandler, templateHandler, postMortemHandler, actionItemHandler, auditLogHandler, reportHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createInitialAdminIfNeeded creates an initial admin user if:
// 1. INITIAL_ADMIN_* environment variables are set
// 2. No users exist in the database
func createInitialAdminIfNeeded(dbConn *gorm.DB, userRepo domain.UserRepository, cfg *config.Config) {
	ctx := context.Background()

	// Check if initial admin configuration is provided
	if cfg.InitialAdminEmail == "" || cfg.InitialAdminPassword == "" || cfg.InitialAdminName == "" {
		log.Println("INFO: Initial admin user not configured (INITIAL_ADMIN_* environment variables not set)")
		return
	}

	// Check if any users already exist
	var userCount int64
	if err := dbConn.Model(&domain.User{}).Count(&userCount).Error; err != nil {
		log.Printf("WARNING: Failed to count users: %v", err)
		return
	}

	if userCount > 0 {
		log.Printf("INFO: Users already exist (%d users found), skipping initial admin creation", userCount)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.InitialAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash initial admin password: %v", err)
		return
	}

	// Create initial admin user
	adminUser := &domain.User{
		Email:        cfg.InitialAdminEmail,
		PasswordHash: string(hashedPassword),
		Name:         cfg.InitialAdminName,
		Role:         domain.RoleAdmin,
		IsActive:     true,
	}

	if err := userRepo.Create(ctx, adminUser); err != nil {
		log.Printf("ERROR: Failed to create initial admin user: %v", err)
		return
	}

	log.Printf("SUCCESS: Initial admin user created successfully (email: %s, name: %s)", adminUser.Email, adminUser.Name)
	log.Println("IMPORTANT: Please change the admin password immediately after first login!")
}
