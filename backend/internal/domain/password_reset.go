package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	// PasswordResetTokenExpiration はパスワードリセット token の有効期間です
	PasswordResetTokenExpiration = 1 * time.Hour
	// PasswordResetTokenLength は token のバイト長です（16進数エンコードで64文字になります）
	PasswordResetTokenLength = 32
)

// PasswordResetToken はパスワードリセット request を表します
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null;size:64" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName は GORM 用のテーブル名を返します
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired は token が期限切れかを確認します
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed は token が使用済みかを確認します
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid は token が有効か（期限切れでなく未使用か）を確認します
func (t *PasswordResetToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// GenerateToken は暗号学的に安全なランダム token を生成します
func GenerateToken() (string, error) {
	bytes := make([]byte, PasswordResetTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// PasswordResetTokenRepository はパスワードリセット token 操作のインターフェースを定義します
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, id uint) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uint) error
}
