package domain

import (
	"context"
	"time"
)

// Tag はインシデントを分類するためのタグエンティティを表します
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	Color     string    `gorm:"default:'#808080'" json:"color"` // 16進数カラーコード
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagRepository はタグデータアクセスのインターフェースを定義します
type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	FindAll(ctx context.Context) ([]*Tag, error)
	FindByID(ctx context.Context, id uint) (*Tag, error)
	FindByIDs(ctx context.Context, ids []uint) ([]Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uint) error
}
