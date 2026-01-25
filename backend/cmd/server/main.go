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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// ロガーを初期化します
	env := logger.GetEnv()
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Incidex server", zap.String("environment", env))

	cfg := config.Load()

	// セキュアなロギングでデータベースを初期化します
	isProduction := cfg.AppEnv == "production" || cfg.AppEnv == "prod"
	dbConn := db.Connect(cfg.DatabaseURL, logger.Log, cfg.DBLogLevel)

	// AUTO_MIGRATE が有効な場合はデータベースマイグレーションを実行します
	if cfg.AutoMigrate {
		log.Println("INFO: AUTO_MIGRATE is enabled. Running database migrations...")
		if err := db.RunMigrations(cfg.MigrationsDir, cfg.DatabaseURL); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		log.Println("SUCCESS: Database migrations completed successfully")
	} else {
		log.Println("INFO: AUTO_MIGRATE is disabled. Database migrations are managed manually.")
		log.Println("To enable auto-migration, set AUTO_MIGRATE=true environment variable.")
		log.Println("To run migrations manually: 'make migrate-up' (local) or 'make migrate-docker-up' (Docker)")
	}

	// MinIO ストレージを初期化します
	// 本番環境では SSL を使用し、開発環境ではローカルテスト用に無効化します
	useSSL := isProduction
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		storage.DefaultBucketName,
		useSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
	}

	// Redis キャッシュを初期化します
	redisClient := db.ConnectRedis(cfg.RedisURL)
	cacheRepo := cache.NewRedisCache(redisClient)

	// 依存性注入
	// 認証
	userRepo := persistence.NewUserRepository(dbConn)
	refreshTokenRepo := persistence.NewRefreshTokenRepository(dbConn)

	// 設定されている場合、かつユーザーが存在しない場合に初期管理者ユーザーを作成します
	createInitialAdminIfNeeded(dbConn, userRepo, cfg)

	// JWT token の有効期限: access token は1時間（よりセキュア）
	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, cfg.JWTSecret, 1*time.Hour)
	authHandler := handler.NewAuthHandler(authUsecase, isProduction)
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	// パスワードリセット
	emailService := notification.NewEmailService(cfg.FrontendURL)
	passwordResetTokenRepo := persistence.NewPasswordResetTokenRepository(dbConn)
	passwordResetUsecase := usecase.NewPasswordResetUsecase(userRepo, passwordResetTokenRepo, emailService, cfg.FrontendURL)
	passwordResetHandler := handler.NewPasswordResetHandler(passwordResetUsecase)

	// タグ
	tagRepo := persistence.NewTagRepository(dbConn)
	tagUsecase := usecase.NewTagUsecase(tagRepo)
	tagHandler := handler.NewTagHandler(tagUsecase)

	// インシデントアクティビティ
	activityRepo := persistence.NewIncidentActivityRepository(dbConn)

	// 通知
	notificationRepo := persistence.NewNotificationSettingRepository(dbConn)
	notificationService := notification.NewNotificationService(notificationRepo, userRepo, cfg.FrontendURL)
	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)

	// インシデント
	incidentRepo := persistence.NewIncidentRepository(dbConn)

	// ユーザー
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	// 統計
	statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
	statsHandler := handler.NewStatsHandler(statsUsecase)

	// アクティビティ handler
	activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, notificationService)
	activityHandler := handler.NewIncidentActivityHandler(activityUsecase)

	// 添付ファイル
	attachmentRepo := persistence.NewAttachmentRepository(dbConn)
	attachmentUsecase := usecase.NewAttachmentUsecase(attachmentRepo, incidentRepo, minioStorage)
	attachmentHandler := handler.NewAttachmentHandler(attachmentUsecase)

	// ポストモーテム
	postMortemRepo := persistence.NewPostMortemRepository(dbConn)
	postMortemUsecase := usecase.NewPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)
	postMortemHandler := handler.NewPostMortemHandler(postMortemUsecase)

	// PostMortemRepo が利用可能になった後に IncidentUsecase を初期化します
	incidentUsecase := usecase.NewIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, notificationService, cacheRepo)
	incidentHandler := handler.NewIncidentHandler(incidentUsecase)

	// エクスポート
	exportHandler := handler.NewExportHandler(incidentUsecase)

	// アクションアイテム
	actionItemRepo := persistence.NewActionItemRepository(dbConn)
	actionItemUsecase := usecase.NewActionItemUsecase(actionItemRepo, postMortemRepo)
	actionItemHandler := handler.NewActionItemHandler(actionItemUsecase)

	// 監査ログ
	auditLogRepo := persistence.NewAuditLogRepository(dbConn)
	auditLogUsecase := usecase.NewAuditLogUsecase(auditLogRepo)
	auditLogHandler := handler.NewAuditLogHandler(auditLogUsecase)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogRepo, userRepo)

	// レポート
	reportRepo := persistence.NewReportRepository(dbConn)
	reportUsecase := usecase.NewReportUsecase(reportRepo)
	reportHandler := handler.NewReportHandler(reportUsecase)

	// ヘルスチェック
	healthHandler := handler.NewHealthHandler(dbConn)

	// レートリミッター
	// ログインレート制限: 1分あたり5リクエスト
	loginRateLimiter := middleware.NewRateLimitMiddleware(redisClient, 5, 1*time.Minute)
	// 一般 API レート制限: 1分あたり100リクエスト
	apiRateLimiter := middleware.NewRateLimitMiddleware(redisClient, 100, 1*time.Minute)

	r := gin.Default()

	// セキュリティヘッダー middleware（最初に適用）
	r.Use(middleware.SecurityHeaders())

	// 監査ログ middleware
	r.Use(auditMiddleware.Log())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ルートを登録します
	router.RegisterRoutes(r, authHandler, jwtMiddleware, tagHandler, incidentHandler, userHandler, statsHandler, activityHandler, exportHandler, attachmentHandler, notificationHandler, postMortemHandler, actionItemHandler, auditLogHandler, reportHandler, healthHandler, passwordResetHandler, loginRateLimiter, apiRateLimiter)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createInitialAdminIfNeeded は以下の条件で初期管理者ユーザーを作成します:
// 1. INITIAL_ADMIN_* 環境変数が設定されている
// 2. データベースにユーザーが存在しない
func createInitialAdminIfNeeded(dbConn *gorm.DB, userRepo domain.UserRepository, cfg *config.Config) {
	ctx := context.Background()

	// 初期管理者設定が提供されているか確認します
	if cfg.InitialAdminEmail == "" || cfg.InitialAdminPassword == "" || cfg.InitialAdminName == "" {
		log.Println("INFO: Initial admin user not configured (INITIAL_ADMIN_* environment variables not set)")
		return
	}

	// ユーザーが既に存在するか確認します
	var userCount int64
	if err := dbConn.Model(&domain.User{}).Count(&userCount).Error; err != nil {
		log.Printf("WARNING: Failed to count users: %v", err)
		return
	}

	if userCount > 0 {
		log.Printf("INFO: Users already exist (%d users found), skipping initial admin creation", userCount)
		return
	}

	// パスワードをハッシュ化します
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.InitialAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash initial admin password: %v", err)
		return
	}

	// 初期管理者ユーザーを作成します
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
