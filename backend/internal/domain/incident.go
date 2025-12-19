package domain

import (
	"context"
	"time"
)

// Severity represents the severity level of an incident.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Status represents the current status of an incident.
type Status string

const (
	StatusOpen          Status = "open"
	StatusInvestigating Status = "investigating"
	StatusResolved      Status = "resolved"
	StatusClosed        Status = "closed"
)

// Incident represents an incident entity.
type Incident struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:500;not null;index" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Summary     string    `gorm:"size:300" json:"summary"`
	Severity    Severity  `gorm:"size:20;not null;index" json:"severity"`
	Status      Status    `gorm:"size:20;not null;default:'open';index" json:"status"`
	ImpactScope string    `gorm:"size:500" json:"impact_scope"`
	DetectedAt  time.Time `gorm:"not null;index" json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	AssigneeID  *uint     `gorm:"index" json:"assignee_id"`
	CreatorID   uint      `gorm:"not null;index" json:"creator_id"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Assignee *User `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Creator  *User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Tags     []Tag `gorm:"many2many:incident_tags" json:"tags,omitempty"`
}

// IncidentFilters represents filtering options for incidents.
type IncidentFilters struct {
	Severity string
	Status   string
	TagIDs   []uint
	Search   string
	SortBy   string
	Order    string
}

// Pagination represents pagination parameters.
type Pagination struct {
	Page  int
	Limit int
}

// PaginationResult represents pagination metadata.
type PaginationResult struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// IncidentRepository defines the interface for incident data access.
type IncidentRepository interface {
	Create(ctx context.Context, incident *Incident) error
	FindAll(ctx context.Context, filters IncidentFilters, pagination Pagination) ([]*Incident, *PaginationResult, error)
	FindByID(ctx context.Context, id uint) (*Incident, error)
	Update(ctx context.Context, incident *Incident) error
	Delete(ctx context.Context, id uint) error
}
