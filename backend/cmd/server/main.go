package main

import (
	"context"
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"incidex/internal/interface/http/middleware"
	"incidex/internal/interface/http/router"
	"incidex/internal/pkg/logger"
	"incidex/internal/wire"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	env := logger.GetEnv()
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Incidex server", zap.String("environment", env))

	cfg := config.Load()

	// Run migrations if AUTO_MIGRATE is enabled
	runMigrationsIfEnabled(cfg)

	// Initialize app with Wire DI
	app, err := wire.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Create initial admin user if needed
	createInitialAdminIfNeeded(app.DB, cfg)

	// Setup router
	r := setupRouter(app, cfg)

	// Start server with graceful shutdown
	startServer(r, cfg.Port)
}

// runMigrationsIfEnabled runs database migrations if AUTO_MIGRATE is enabled.
func runMigrationsIfEnabled(cfg *config.Config) {
	if cfg.AutoMigrate {
		log.Println("INFO: AUTO_MIGRATE is enabled. Running database migrations...")
		if err := db.RunMigrations(cfg.MigrationsDir, cfg.DatabaseURL); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		log.Println("SUCCESS: Database migrations completed successfully")
	} else {
		log.Println("INFO: AUTO_MIGRATE is disabled. Database migrations are managed manually.")
	}
}

// setupRouter configures the Gin router with all middleware and routes.
func setupRouter(app *wire.App, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Request ID middleware
	r.Use(middleware.RequestID())

	// Security headers middleware
	r.Use(middleware.SecurityHeaders())

	// Audit log middleware
	r.Use(app.AuditMiddleware.Log())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register routes
	router.RegisterRoutes(
		r,
		app.AuthHandler,
		app.JWTMiddleware,
		app.TagHandler,
		app.IncidentHandler,
		app.UserHandler,
		app.StatsHandler,
		app.ActivityHandler,
		app.ExportHandler,
		app.AttachmentHandler,
		app.NotificationHandler,
		app.PostMortemHandler,
		app.ActionItemHandler,
		app.AuditLogHandler,
		app.ReportHandler,
		app.HealthHandler,
		app.PasswordResetHandler,
		app.LoginRateLimiter,
		app.APIRateLimiter,
	)

	return r
}

// startServer starts the HTTP server with graceful shutdown support.
func startServer(handler http.Handler, port string) {
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server asynchronously
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited gracefully")
}

// createInitialAdminIfNeeded creates the initial admin user if:
// 1. INITIAL_ADMIN_* environment variables are set
// 2. No users exist in the database
func createInitialAdminIfNeeded(dbConn *gorm.DB, cfg *config.Config) {
	// Check if initial admin configuration is provided
	if cfg.InitialAdminEmail == "" || cfg.InitialAdminPassword == "" || cfg.InitialAdminName == "" {
		log.Println("INFO: Initial admin user not configured (INITIAL_ADMIN_* environment variables not set)")
		return
	}

	// Check if users already exist
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

	// Create initial admin user directly via GORM
	adminUser := &domain.User{
		Email:        cfg.InitialAdminEmail,
		PasswordHash: string(hashedPassword),
		Name:         cfg.InitialAdminName,
		Role:         domain.RoleAdmin,
		IsActive:     true,
	}

	if err := dbConn.Create(adminUser).Error; err != nil {
		log.Printf("ERROR: Failed to create initial admin user: %v", err)
		return
	}

	log.Printf("SUCCESS: Initial admin user created successfully (email: %s, name: %s)", adminUser.Email, adminUser.Name)
	log.Println("IMPORTANT: Please change the admin password immediately after first login!")
}
