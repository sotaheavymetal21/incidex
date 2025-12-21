package domain

import (
	"context"
	"time"
)

// ActionItem represents an action item from a post-mortem.
type ActionItem struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	PostMortemID uint         `gorm:"not null;index" json:"post_mortem_id"`
	Title        string       `gorm:"size:500;not null" json:"title"`
	Description  string       `gorm:"type:text" json:"description"`
	AssigneeID   *uint        `gorm:"index" json:"assignee_id"`
	Priority     Priority     `gorm:"size:20;not null;default:'medium';index" json:"priority"`
	Status       ActionStatus `gorm:"size:20;not null;default:'pending';index" json:"status"`
	DueDate      *time.Time   `json:"due_date"`
	RelatedLinks string       `gorm:"type:text" json:"related_links"` // JSON array of links
	CreatedAt    time.Time    `gorm:"index" json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	CompletedAt  *time.Time   `json:"completed_at"`

	// Relations
	PostMortem *PostMortem `gorm:"foreignKey:PostMortemID" json:"-"`
	Assignee   *User       `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}

// Priority represents the priority level of an action item
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// ActionStatus represents the status of an action item
type ActionStatus string

const (
	ActionStatusPending    ActionStatus = "pending"
	ActionStatusInProgress ActionStatus = "in_progress"
	ActionStatusCompleted  ActionStatus = "completed"
)

// ActionItemRepository defines the interface for action item data access.
type ActionItemRepository interface {
	Create(ctx context.Context, item *ActionItem) error
	FindByID(ctx context.Context, id uint) (*ActionItem, error)
	FindByPostMortemID(ctx context.Context, postMortemID uint) ([]*ActionItem, error)
	Update(ctx context.Context, item *ActionItem) error
	Delete(ctx context.Context, id uint) error
	FindAll(ctx context.Context, filters ActionItemFilters, pagination Pagination) ([]*ActionItem, *PaginationResult, error)
}

// ActionItemFilters represents filtering options for action items.
type ActionItemFilters struct {
	Status     string
	Priority   string
	AssigneeID uint
	Search     string
	SortBy     string
	Order      string
}
