package domain

import (
	"context"
	"time"
)

// Tag represents a tag entity for categorizing incidents.
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	Color     string    `gorm:"default:'#808080'" json:"color"` // Hex color code
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagRepository defines the interface for tag data access.
type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	FindAll(ctx context.Context) ([]*Tag, error)
	FindByID(ctx context.Context, id uint) (*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uint) error
}
