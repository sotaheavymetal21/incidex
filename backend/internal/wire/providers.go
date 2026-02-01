package wire

import (
	"incidex/internal/config"
	"incidex/internal/db"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/cache"
	"incidex/internal/infrastructure/notification"
	"incidex/internal/infrastructure/persistence"
	"incidex/internal/infrastructure/storage"
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"
	"incidex/internal/pkg/logger"
	"incidex/internal/usecase"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Config value types for Wire injection
type JWTSecret string
type FrontendURL string
type JWTExpiry time.Duration
type IsProduction bool

// ProvideJWTSecret extracts JWT secret from config.
func ProvideJWTSecret(cfg *config.Config) JWTSecret {
	return JWTSecret(cfg.JWTSecret)
}

// ProvideFrontendURL extracts frontend URL from config.
func ProvideFrontendURL(cfg *config.Config) FrontendURL {
	return FrontendURL(cfg.FrontendURL)
}

// ProvideJWTExpiry provides the JWT token expiry duration.
func ProvideJWTExpiry() JWTExpiry {
	return JWTExpiry(1 * time.Hour)
}

// ProvideIsProduction determines if running in production mode.
func ProvideIsProduction(cfg *config.Config) IsProduction {
	return IsProduction(cfg.AppEnv == "production" || cfg.AppEnv == "prod")
}

// ConfigSet provides config-derived values.
var ConfigSet = wire.NewSet(
	ProvideJWTSecret,
	ProvideFrontendURL,
	ProvideJWTExpiry,
	ProvideIsProduction,
)

// ProvideDB initializes the database connection.
func ProvideDB(cfg *config.Config) *gorm.DB {
	return db.Connect(cfg.DatabaseURL, logger.Log, cfg.DBLogLevel)
}

// ProvideRedis initializes the Redis client.
func ProvideRedis(cfg *config.Config) *redis.Client {
	return db.ConnectRedis(cfg.RedisURL)
}

// ProvideMinIOStorage initializes MinIO storage.
func ProvideMinIOStorage(cfg *config.Config, isProduction IsProduction) (*storage.MinIOStorage, error) {
	useSSL := bool(isProduction)
	return storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		storage.DefaultBucketName,
		useSSL,
	)
}

// ProvideCacheRepository creates Redis cache repository.
func ProvideCacheRepository(client *redis.Client) domain.CacheRepository {
	return cache.NewRedisCache(client)
}

// InfrastructureSet provides infrastructure components.
var InfrastructureSet = wire.NewSet(
	ProvideDB,
	ProvideRedis,
	ProvideMinIOStorage,
	ProvideCacheRepository,
)

// ProvideUserRepository creates user repository.
func ProvideUserRepository(db *gorm.DB) domain.UserRepository {
	return persistence.NewUserRepository(db)
}

// ProvideRefreshTokenRepository creates refresh token repository.
func ProvideRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return persistence.NewRefreshTokenRepository(db)
}

// ProvidePasswordResetTokenRepository creates password reset token repository.
func ProvidePasswordResetTokenRepository(db *gorm.DB) domain.PasswordResetTokenRepository {
	return persistence.NewPasswordResetTokenRepository(db)
}

// ProvideTagRepository creates tag repository.
func ProvideTagRepository(db *gorm.DB) domain.TagRepository {
	return persistence.NewTagRepository(db)
}

// ProvideIncidentRepository creates incident repository.
func ProvideIncidentRepository(db *gorm.DB) domain.IncidentRepository {
	return persistence.NewIncidentRepository(db)
}

// ProvideIncidentActivityRepository creates incident activity repository.
func ProvideIncidentActivityRepository(db *gorm.DB) domain.IncidentActivityRepository {
	return persistence.NewIncidentActivityRepository(db)
}

// ProvideNotificationSettingRepository creates notification setting repository.
func ProvideNotificationSettingRepository(db *gorm.DB) domain.NotificationSettingRepository {
	return persistence.NewNotificationSettingRepository(db)
}

// ProvideAttachmentRepository creates attachment repository.
func ProvideAttachmentRepository(db *gorm.DB) domain.AttachmentRepository {
	return persistence.NewAttachmentRepository(db)
}

// ProvidePostMortemRepository creates post-mortem repository.
func ProvidePostMortemRepository(db *gorm.DB) domain.PostMortemRepository {
	return persistence.NewPostMortemRepository(db)
}

// ProvideActionItemRepository creates action item repository.
func ProvideActionItemRepository(db *gorm.DB) domain.ActionItemRepository {
	return persistence.NewActionItemRepository(db)
}

// ProvideAuditLogRepository creates audit log repository.
func ProvideAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return persistence.NewAuditLogRepository(db)
}

// ProvideReportRepository creates report repository.
func ProvideReportRepository(db *gorm.DB) domain.ReportRepository {
	return persistence.NewReportRepository(db)
}

// RepositorySet provides all repositories.
var RepositorySet = wire.NewSet(
	ProvideUserRepository,
	ProvideRefreshTokenRepository,
	ProvidePasswordResetTokenRepository,
	ProvideTagRepository,
	ProvideIncidentRepository,
	ProvideIncidentActivityRepository,
	ProvideNotificationSettingRepository,
	ProvideAttachmentRepository,
	ProvidePostMortemRepository,
	ProvideActionItemRepository,
	ProvideAuditLogRepository,
	ProvideReportRepository,
)

// ProvideEmailService creates email service.
func ProvideEmailService(frontendURL FrontendURL) *notification.EmailService {
	return notification.NewEmailService(string(frontendURL))
}

// ProvideNotificationService creates notification service.
func ProvideNotificationService(
	settingRepo domain.NotificationSettingRepository,
	userRepo domain.UserRepository,
	frontendURL FrontendURL,
) *notification.NotificationService {
	return notification.NewNotificationService(settingRepo, userRepo, string(frontendURL))
}

// ServiceSet provides services.
var ServiceSet = wire.NewSet(
	ProvideEmailService,
	ProvideNotificationService,
)

// ProvideAuthUsecase creates auth usecase.
func ProvideAuthUsecase(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	jwtSecret JWTSecret,
	jwtExpiry JWTExpiry,
) usecase.AuthUsecase {
	return usecase.NewAuthUsecase(userRepo, refreshTokenRepo, string(jwtSecret), time.Duration(jwtExpiry))
}

// ProvidePasswordResetUsecase creates password reset usecase.
func ProvidePasswordResetUsecase(
	userRepo domain.UserRepository,
	passwordResetTokenRepo domain.PasswordResetTokenRepository,
	emailService *notification.EmailService,
	frontendURL FrontendURL,
) usecase.PasswordResetUsecase {
	return usecase.NewPasswordResetUsecase(userRepo, passwordResetTokenRepo, emailService, string(frontendURL))
}

// ProvideTagUsecase creates tag usecase.
func ProvideTagUsecase(tagRepo domain.TagRepository) usecase.TagUsecase {
	return usecase.NewTagUsecase(tagRepo)
}

// ProvideUserUsecase creates user usecase.
func ProvideUserUsecase(userRepo domain.UserRepository) usecase.UserUsecase {
	return usecase.NewUserUsecase(userRepo)
}

// ProvideStatsUsecase creates stats usecase.
func ProvideStatsUsecase(
	incidentRepo domain.IncidentRepository,
	cacheRepo domain.CacheRepository,
) *usecase.StatsUsecase {
	return usecase.NewStatsUsecase(incidentRepo, cacheRepo)
}

// ProvideIncidentActivityUsecase creates incident activity usecase.
func ProvideIncidentActivityUsecase(
	activityRepo domain.IncidentActivityRepository,
	incidentRepo domain.IncidentRepository,
	userRepo domain.UserRepository,
	notificationService *notification.NotificationService,
) *usecase.IncidentActivityUsecase {
	return usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, notificationService)
}

// ProvideAttachmentUsecase creates attachment usecase.
func ProvideAttachmentUsecase(
	attachmentRepo domain.AttachmentRepository,
	incidentRepo domain.IncidentRepository,
	minioStorage *storage.MinIOStorage,
) usecase.AttachmentUsecase {
	return usecase.NewAttachmentUsecase(attachmentRepo, incidentRepo, minioStorage)
}

// ProvidePostMortemUsecase creates post-mortem usecase.
func ProvidePostMortemUsecase(
	postMortemRepo domain.PostMortemRepository,
	incidentRepo domain.IncidentRepository,
	activityRepo domain.IncidentActivityRepository,
	userRepo domain.UserRepository,
) usecase.PostMortemUsecase {
	return usecase.NewPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)
}

// ProvideIncidentUsecase creates incident usecase.
func ProvideIncidentUsecase(
	incidentRepo domain.IncidentRepository,
	tagRepo domain.TagRepository,
	userRepo domain.UserRepository,
	activityRepo domain.IncidentActivityRepository,
	postMortemRepo domain.PostMortemRepository,
	notificationService *notification.NotificationService,
	cacheRepo domain.CacheRepository,
) usecase.IncidentUsecase {
	return usecase.NewIncidentUsecase(
		incidentRepo,
		tagRepo,
		userRepo,
		activityRepo,
		postMortemRepo,
		notificationService,
		cacheRepo,
	)
}

// ProvideActionItemUsecase creates action item usecase.
func ProvideActionItemUsecase(
	actionItemRepo domain.ActionItemRepository,
	postMortemRepo domain.PostMortemRepository,
) usecase.ActionItemUsecase {
	return usecase.NewActionItemUsecase(actionItemRepo, postMortemRepo)
}

// ProvideAuditLogUsecase creates audit log usecase.
func ProvideAuditLogUsecase(auditLogRepo domain.AuditLogRepository) usecase.AuditLogUsecase {
	return usecase.NewAuditLogUsecase(auditLogRepo)
}

// ProvideReportUsecase creates report usecase.
func ProvideReportUsecase(reportRepo domain.ReportRepository) usecase.ReportUsecase {
	return usecase.NewReportUsecase(reportRepo)
}

// ProvideNotificationUsecase creates notification usecase.
func ProvideNotificationUsecase(notificationRepo domain.NotificationSettingRepository) *usecase.NotificationUsecase {
	return usecase.NewNotificationUsecase(notificationRepo)
}

// UsecaseSet provides all usecases.
var UsecaseSet = wire.NewSet(
	ProvideAuthUsecase,
	ProvidePasswordResetUsecase,
	ProvideTagUsecase,
	ProvideUserUsecase,
	ProvideStatsUsecase,
	ProvideIncidentActivityUsecase,
	ProvideAttachmentUsecase,
	ProvidePostMortemUsecase,
	ProvideIncidentUsecase,
	ProvideActionItemUsecase,
	ProvideAuditLogUsecase,
	ProvideReportUsecase,
	ProvideNotificationUsecase,
)

// ProvideAuthHandler creates auth handler.
func ProvideAuthHandler(authUsecase usecase.AuthUsecase, isProduction IsProduction) *handler.AuthHandler {
	return handler.NewAuthHandler(authUsecase, bool(isProduction))
}

// ProvidePasswordResetHandler creates password reset handler.
func ProvidePasswordResetHandler(passwordResetUsecase usecase.PasswordResetUsecase) *handler.PasswordResetHandler {
	return handler.NewPasswordResetHandler(passwordResetUsecase)
}

// ProvideTagHandler creates tag handler.
func ProvideTagHandler(tagUsecase usecase.TagUsecase) *handler.TagHandler {
	return handler.NewTagHandler(tagUsecase)
}

// ProvideIncidentHandler creates incident handler.
func ProvideIncidentHandler(incidentUsecase usecase.IncidentUsecase) *handler.IncidentHandler {
	return handler.NewIncidentHandler(incidentUsecase)
}

// ProvideUserHandler creates user handler.
func ProvideUserHandler(userUsecase usecase.UserUsecase) *handler.UserHandler {
	return handler.NewUserHandler(userUsecase)
}

// ProvideStatsHandler creates stats handler.
func ProvideStatsHandler(statsUsecase *usecase.StatsUsecase) *handler.StatsHandler {
	return handler.NewStatsHandler(statsUsecase)
}

// ProvideActivityHandler creates incident activity handler.
func ProvideActivityHandler(activityUsecase *usecase.IncidentActivityUsecase) *handler.IncidentActivityHandler {
	return handler.NewIncidentActivityHandler(activityUsecase)
}

// ProvideExportHandler creates export handler.
func ProvideExportHandler(incidentUsecase usecase.IncidentUsecase) *handler.ExportHandler {
	return handler.NewExportHandler(incidentUsecase)
}

// ProvideAttachmentHandler creates attachment handler.
func ProvideAttachmentHandler(attachmentUsecase usecase.AttachmentUsecase) *handler.AttachmentHandler {
	return handler.NewAttachmentHandler(attachmentUsecase)
}

// ProvideNotificationHandler creates notification handler.
func ProvideNotificationHandler(notificationUsecase *usecase.NotificationUsecase) *handler.NotificationHandler {
	return handler.NewNotificationHandler(notificationUsecase)
}

// ProvidePostMortemHandler creates post-mortem handler.
func ProvidePostMortemHandler(postMortemUsecase usecase.PostMortemUsecase) *handler.PostMortemHandler {
	return handler.NewPostMortemHandler(postMortemUsecase)
}

// ProvideActionItemHandler creates action item handler.
func ProvideActionItemHandler(actionItemUsecase usecase.ActionItemUsecase) *handler.ActionItemHandler {
	return handler.NewActionItemHandler(actionItemUsecase)
}

// ProvideAuditLogHandler creates audit log handler.
func ProvideAuditLogHandler(auditLogUsecase usecase.AuditLogUsecase) *handler.AuditLogHandler {
	return handler.NewAuditLogHandler(auditLogUsecase)
}

// ProvideReportHandler creates report handler.
func ProvideReportHandler(reportUsecase usecase.ReportUsecase) *handler.ReportHandler {
	return handler.NewReportHandler(reportUsecase)
}

// ProvideHealthHandler creates health handler.
func ProvideHealthHandler(db *gorm.DB) *handler.HealthHandler {
	return handler.NewHealthHandler(db)
}

// HandlerSet provides all handlers.
var HandlerSet = wire.NewSet(
	ProvideAuthHandler,
	ProvidePasswordResetHandler,
	ProvideTagHandler,
	ProvideIncidentHandler,
	ProvideUserHandler,
	ProvideStatsHandler,
	ProvideActivityHandler,
	ProvideExportHandler,
	ProvideAttachmentHandler,
	ProvideNotificationHandler,
	ProvidePostMortemHandler,
	ProvideActionItemHandler,
	ProvideAuditLogHandler,
	ProvideReportHandler,
	ProvideHealthHandler,
)

// ProvideJWTMiddleware creates JWT middleware.
func ProvideJWTMiddleware(jwtSecret JWTSecret) *middleware.JWTMiddleware {
	return middleware.NewJWTMiddleware(string(jwtSecret))
}

// ProvideAuditMiddleware creates audit middleware.
func ProvideAuditMiddleware(
	auditLogRepo domain.AuditLogRepository,
	userRepo domain.UserRepository,
) *middleware.AuditMiddleware {
	return middleware.NewAuditMiddleware(auditLogRepo, userRepo)
}

// LoginRateLimiter is a typed rate limiter for login endpoints.
type LoginRateLimiter *middleware.RateLimitMiddleware

// APIRateLimiter is a typed rate limiter for API endpoints.
type APIRateLimiter *middleware.RateLimitMiddleware

// ProvideLoginRateLimiter creates rate limiter for login endpoints.
func ProvideLoginRateLimiter(client *redis.Client) LoginRateLimiter {
	return middleware.NewRateLimitMiddleware(client, 5, 1*time.Minute)
}

// ProvideAPIRateLimiter creates rate limiter for API endpoints.
func ProvideAPIRateLimiter(client *redis.Client) APIRateLimiter {
	return middleware.NewRateLimitMiddleware(client, 100, 1*time.Minute)
}

// MiddlewareSet provides all middleware.
var MiddlewareSet = wire.NewSet(
	ProvideJWTMiddleware,
	ProvideAuditMiddleware,
	ProvideLoginRateLimiter,
	ProvideAPIRateLimiter,
)

// ProvideApp creates the App struct with all dependencies.
func ProvideApp(
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	authHandler *handler.AuthHandler,
	passwordResetHandler *handler.PasswordResetHandler,
	tagHandler *handler.TagHandler,
	incidentHandler *handler.IncidentHandler,
	userHandler *handler.UserHandler,
	statsHandler *handler.StatsHandler,
	activityHandler *handler.IncidentActivityHandler,
	exportHandler *handler.ExportHandler,
	attachmentHandler *handler.AttachmentHandler,
	notificationHandler *handler.NotificationHandler,
	postMortemHandler *handler.PostMortemHandler,
	actionItemHandler *handler.ActionItemHandler,
	auditLogHandler *handler.AuditLogHandler,
	reportHandler *handler.ReportHandler,
	healthHandler *handler.HealthHandler,
	jwtMiddleware *middleware.JWTMiddleware,
	auditMiddleware *middleware.AuditMiddleware,
	loginRateLimiter LoginRateLimiter,
	apiRateLimiter APIRateLimiter,
) *App {
	return &App{
		Config:               cfg,
		DB:                   db,
		Redis:                redisClient,
		AuthHandler:          authHandler,
		PasswordResetHandler: passwordResetHandler,
		TagHandler:           tagHandler,
		IncidentHandler:      incidentHandler,
		UserHandler:          userHandler,
		StatsHandler:         statsHandler,
		ActivityHandler:      activityHandler,
		ExportHandler:        exportHandler,
		AttachmentHandler:    attachmentHandler,
		NotificationHandler:  notificationHandler,
		PostMortemHandler:    postMortemHandler,
		ActionItemHandler:    actionItemHandler,
		AuditLogHandler:      auditLogHandler,
		ReportHandler:        reportHandler,
		HealthHandler:        healthHandler,
		JWTMiddleware:        jwtMiddleware,
		AuditMiddleware:      auditMiddleware,
		LoginRateLimiter:     (*middleware.RateLimitMiddleware)(loginRateLimiter),
		APIRateLimiter:       (*middleware.RateLimitMiddleware)(apiRateLimiter),
	}
}

// AllProviders combines all provider sets.
var AllProviders = wire.NewSet(
	ConfigSet,
	InfrastructureSet,
	RepositorySet,
	ServiceSet,
	UsecaseSet,
	HandlerSet,
	MiddlewareSet,
	ProvideApp,
)
