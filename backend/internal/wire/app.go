package wire

import (
	"incidex/internal/config"
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App holds all handlers and middleware required for the server.
type App struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client

	// Handlers
	AuthHandler          *handler.AuthHandler
	PasswordResetHandler *handler.PasswordResetHandler
	TagHandler           *handler.TagHandler
	IncidentHandler      *handler.IncidentHandler
	UserHandler          *handler.UserHandler
	StatsHandler         *handler.StatsHandler
	ActivityHandler      *handler.IncidentActivityHandler
	ExportHandler        *handler.ExportHandler
	AttachmentHandler    *handler.AttachmentHandler
	NotificationHandler  *handler.NotificationHandler
	PostMortemHandler    *handler.PostMortemHandler
	ActionItemHandler    *handler.ActionItemHandler
	AuditLogHandler      *handler.AuditLogHandler
	ReportHandler        *handler.ReportHandler
	HealthHandler        *handler.HealthHandler

	// Middleware
	JWTMiddleware    *middleware.JWTMiddleware
	AuditMiddleware  *middleware.AuditMiddleware
	LoginRateLimiter *middleware.RateLimitMiddleware
	APIRateLimiter   *middleware.RateLimitMiddleware
}
