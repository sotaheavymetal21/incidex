package domain

import (
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
	Create(tag *Tag) error
	FindAll() ([]*Tag, error)
	FindByID(id uint) (*Tag, error)
	Update(tag *Tag) error
	Delete(id uint) error
}
