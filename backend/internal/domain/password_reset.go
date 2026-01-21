package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	// PasswordResetTokenExpiration is the duration for which a password reset token is valid
	PasswordResetTokenExpiration = 1 * time.Hour
	// PasswordResetTokenLength is the length of the token in bytes (will be hex encoded to 64 characters)
	PasswordResetTokenLength = 32
)

// PasswordResetToken represents a password reset request
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null;size:64" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the table name for GORM
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired checks if the token has expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired and not used)
func (t *PasswordResetToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// GenerateToken generates a cryptographically secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, PasswordResetTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// PasswordResetTokenRepository defines the interface for password reset token operations
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, id uint) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uint) error
}
