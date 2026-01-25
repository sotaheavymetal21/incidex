package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// RefreshToken は JWT 認証用のリフレッシュ token を表します
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

// RefreshTokenRepository はリフレッシュ token データ操作のインターフェースを定義します
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeByToken(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID uint) error
	DeleteExpired(ctx context.Context) error
}

// IsExpired はリフレッシュ token が期限切れかを確認します
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked はリフレッシュ token が無効化されているかを確認します
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid はリフレッシュ token が有効か（期限切れでなく無効化されていないか）を確認します
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}
