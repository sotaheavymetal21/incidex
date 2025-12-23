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

	// SLA Fields
	SLATargetResolutionHours int        `gorm:"default:0" json:"sla_target_resolution_hours"` // SLA目標解決時間（時間単位）
	SLADeadline              *time.Time `gorm:"index" json:"sla_deadline"`                     // SLA期限
	SLAViolated              bool       `gorm:"default:false;index" json:"sla_violated"`       // SLA違反フラグ

	// Relations
	Assignee   *User       `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Assignees  []User      `gorm:"many2many:incident_assignees;" json:"assignees,omitempty"`
	Creator    *User       `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Tags       []Tag       `gorm:"many2many:incident_tags" json:"tags,omitempty"`
	PostMortem *PostMortem `gorm:"foreignKey:IncidentID" json:"post_mortem,omitempty"`
}

// IncidentFilters represents filtering options for incidents.
type IncidentFilters struct {
	Severity     string
	Status       string
	TagIDs       []uint
	Search       string
	SortBy       string
	Order        string
	AssignedToID *uint  // Filter by assignee ID
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
// GetDefaultSLAHours returns the default SLA resolution hours based on severity
func GetDefaultSLAHours(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 4 // 4 hours for critical incidents
	case SeverityHigh:
		return 24 // 24 hours for high severity
	case SeverityMedium:
		return 72 // 72 hours for medium severity
	case SeverityLow:
		return 168 // 168 hours (1 week) for low severity
	default:
		return 72 // Default to 72 hours
	}
}

// CalculateSLADeadline calculates the SLA deadline based on detected time and target hours
func (i *Incident) CalculateSLADeadline() *time.Time {
	if i.SLATargetResolutionHours <= 0 {
		return nil
	}
	deadline := i.DetectedAt.Add(time.Duration(i.SLATargetResolutionHours) * time.Hour)
	return &deadline
}

// CheckSLAViolation checks if the incident has violated its SLA
func (i *Incident) CheckSLAViolation() bool {
	if i.SLADeadline == nil {
		return false
	}

	// If resolved, check if it was resolved after the deadline
	if i.ResolvedAt != nil {
		return i.ResolvedAt.After(*i.SLADeadline)
	}

	// If not resolved, check if current time is after the deadline
	return time.Now().After(*i.SLADeadline)
}

// GetResolutionTime returns the time taken to resolve the incident (for MTTR calculation)
func (i *Incident) GetResolutionTime() *time.Duration {
	if i.ResolvedAt == nil {
		return nil
	}
	duration := i.ResolvedAt.Sub(i.DetectedAt)
	return &duration
}

// IsOpen returns true if the incident is not resolved or closed
func (i *Incident) IsOpen() bool {
	return i.Status == StatusOpen || i.Status == StatusInvestigating
}

// SLAMetrics represents SLA performance metrics
type SLAMetrics struct {
	TotalIncidents      int64   `json:"total_incidents"`
	ResolvedIncidents   int64   `json:"resolved_incidents"`
	SLAViolatedCount    int64   `json:"sla_violated_count"`
	SLAComplianceRate   float64 `json:"sla_compliance_rate"`    // Percentage of incidents resolved within SLA
	AverageMTTR         float64 `json:"average_mttr"`            // Average Mean Time To Resolve (in hours)
	MedianMTTR          float64 `json:"median_mttr"`             // Median resolution time (in hours)
	CurrentlyOverdue    int64   `json:"currently_overdue"`       // Number of open incidents past their SLA deadline
}

type IncidentRepository interface {
	Create(ctx context.Context, incident *Incident) error
	FindAll(ctx context.Context, filters IncidentFilters, pagination Pagination) ([]*Incident, *PaginationResult, error)
	FindByID(ctx context.Context, id uint) (*Incident, error)
	Update(ctx context.Context, incident *Incident) error
	Delete(ctx context.Context, id uint) error

	// Stats methods
	Count(count *int64) error
	CountBySeverity(severity Severity, count *int64) error
	CountByStatus(status Status, count *int64) error
	FindRecent(limit int) ([]*Incident, error)
	GetAllIncidents() ([]*Incident, error)

	// SLA methods
	CountSLAViolated(count *int64) error
	GetSLAMetrics() (*SLAMetrics, error)
}
