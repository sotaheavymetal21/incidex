package domain

import (
	"context"
	"time"
)

// ActionItem はポストモーテムからのアクションアイテムを表します
type ActionItem struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	PostMortemID uint         `gorm:"not null;index" json:"post_mortem_id"`
	Title        string       `gorm:"size:500;not null" json:"title"`
	Description  string       `gorm:"type:text" json:"description"`
	AssigneeID   *uint        `gorm:"index" json:"assignee_id"`
	Priority     Priority     `gorm:"size:20;not null;default:'medium';index" json:"priority"`
	Status       ActionStatus `gorm:"size:20;not null;default:'pending';index" json:"status"`
	DueDate      *time.Time   `json:"due_date"`
	RelatedLinks string       `gorm:"type:jsonb;default:'[]'" json:"related_links"` // リンクの JSON 配列
	CreatedAt    time.Time    `gorm:"index" json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	CompletedAt  *time.Time   `json:"completed_at"`

	// リレーション
	PostMortem *PostMortem `gorm:"foreignKey:PostMortemID" json:"-"`
	Assignee   *User       `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}

// Priority はアクションアイテムの優先度を表します
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// ActionStatus はアクションアイテムのステータスを表します
type ActionStatus string

const (
	ActionStatusPending    ActionStatus = "pending"
	ActionStatusInProgress ActionStatus = "in_progress"
	ActionStatusCompleted  ActionStatus = "completed"
)

// ActionItemRepository はアクションアイテムデータアクセスのインターフェースを定義します
type ActionItemRepository interface {
	Create(ctx context.Context, item *ActionItem) error
	FindByID(ctx context.Context, id uint) (*ActionItem, error)
	FindByPostMortemID(ctx context.Context, postMortemID uint) ([]*ActionItem, error)
	Update(ctx context.Context, item *ActionItem) error
	Delete(ctx context.Context, id uint) error
	FindAll(ctx context.Context, filters ActionItemFilters, pagination Pagination) ([]*ActionItem, *PaginationResult, error)
}

// ActionItemFilters はアクションアイテムのフィルタリングオプションを表します
type ActionItemFilters struct {
	Status     string
	Priority   string
	AssigneeID uint
	Search     string
	SortBy     string
	Order      string
}
