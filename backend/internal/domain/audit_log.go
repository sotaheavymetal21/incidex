package domain

import (
	"context"
	"time"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionRead   AuditAction = "read"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
)

type AuditLog struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	UserID       *uint       `json:"user_id"`
	UserName     string      `json:"user_name"`
	UserEmail    string      `json:"user_email"`
	Action       AuditAction `gorm:"not null" json:"action"`
	ResourceType string      `json:"resource_type"` // e.g., "incident", "user", "tag"
	ResourceID   *uint       `json:"resource_id"`
	Method       string      `json:"method"` // HTTP method: GET, POST, PUT, DELETE
	Path         string      `json:"path"`   // Request path
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
	StatusCode   int         `json:"status_code"`
	Details      string      `gorm:"type:text" json:"details"` // Additional details in JSON format
	CreatedAt    time.Time   `json:"created_at"`
}

type AuditLogFilters struct {
	UserID       *uint
	Action       *AuditAction
	ResourceType *string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	Limit        int
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	FindAll(ctx context.Context, filters AuditLogFilters) ([]*AuditLog, int64, error)
	FindByID(ctx context.Context, id uint) (*AuditLog, error)
}
