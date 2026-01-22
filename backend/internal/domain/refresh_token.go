package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a refresh token for JWT authentication
type RefreshToken struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	RevokedAt *time.Time     `gorm:"index" json:"revoked_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// RefreshTokenRepository defines the interface for refresh token data operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeByToken(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID uint) error
	DeleteExpired(ctx context.Context) error
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the refresh token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}
