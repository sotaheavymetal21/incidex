package domain

import (
	"context"
	"time"
)

// Severity はインシデントの重要度を表します
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Status はインシデントの現在のステータスを表します
type Status string

const (
	StatusOpen          Status = "open"
	StatusInvestigating Status = "investigating"
	StatusResolved      Status = "resolved"
	StatusClosed        Status = "closed"
)

// Incident はインシデントエンティティを表します
type Incident struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:500;not null;index" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Severity    Severity  `gorm:"size:20;not null;index" json:"severity"`
	Status      Status    `gorm:"size:20;not null;default:'open';index" json:"status"`
	ImpactScope string    `gorm:"size:500" json:"impact_scope"`
	DetectedAt  time.Time `gorm:"not null;index" json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	AssigneeID  *uint     `gorm:"index" json:"assignee_id"`
	CreatorID   uint      `gorm:"not null;index" json:"creator_id"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// SLA フィールド
	SLATargetResolutionHours int        `gorm:"default:0" json:"sla_target_resolution_hours"` // SLA目標解決時間（時間単位）
	SLADeadline              *time.Time `gorm:"index" json:"sla_deadline"`                     // SLA期限
	SLAViolated              bool       `gorm:"default:false;index" json:"sla_violated"`       // SLA違反フラグ

	// リレーション
	Assignee   *User       `gorm:"foreignKey:AssigneeID" json:"assignee"`
	Assignees  []User      `gorm:"many2many:incident_assignees;" json:"assignees,omitempty"`
	Creator    *User       `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Tags       []Tag       `gorm:"many2many:incident_tags" json:"tags,omitempty"`
	PostMortem *PostMortem `gorm:"foreignKey:IncidentID" json:"post_mortem,omitempty"`
}

// IncidentFilters はインシデントのフィルタリングオプションを表します
type IncidentFilters struct {
	Severity     string
	Status       string
	TagIDs       []uint
	Search       string
	SortBy       string
	Order        string
	AssignedToID *uint // 担当者IDでフィルタリング
}

// Pagination はページネーションパラメータを表します
type Pagination struct {
	Page  int
	Limit int
}

// PaginationResult はページネーションのメタデータを表します
type PaginationResult struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// IncidentRepository はインシデントデータアクセスのインターフェースを定義します
// GetDefaultSLAHours は重要度に基づいてデフォルトのSLA解決時間を返します
func GetDefaultSLAHours(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 4 // Critical インシデントは4時間
	case SeverityHigh:
		return 24 // High は24時間
	case SeverityMedium:
		return 72 // Medium は72時間
	case SeverityLow:
		return 168 // Low は168時間（1週間）
	default:
		return 72 // デフォルトは72時間
	}
}

// CalculateSLADeadline は検知時刻と目標時間に基づいてSLA期限を計算します
func (i *Incident) CalculateSLADeadline() *time.Time {
	if i.SLATargetResolutionHours <= 0 {
		return nil
	}
	deadline := i.DetectedAt.Add(time.Duration(i.SLATargetResolutionHours) * time.Hour)
	return &deadline
}

// CheckSLAViolation はインシデントがSLAに違反しているかを確認します
func (i *Incident) CheckSLAViolation() bool {
	if i.SLADeadline == nil {
		return false
	}

	// 解決済みの場合、期限後に解決されたかを確認
	if i.ResolvedAt != nil {
		return i.ResolvedAt.After(*i.SLADeadline)
	}

	// 未解決の場合、現在時刻が期限を過ぎているかを確認
	return time.Now().After(*i.SLADeadline)
}

// GetResolutionTime はインシデント解決にかかった時間を返します（MTTR計算用）
func (i *Incident) GetResolutionTime() *time.Duration {
	if i.ResolvedAt == nil {
		return nil
	}
	duration := i.ResolvedAt.Sub(i.DetectedAt)
	return &duration
}

// IsOpen はインシデントが未解決または未クローズの場合に true を返します
func (i *Incident) IsOpen() bool {
	return i.Status == StatusOpen || i.Status == StatusInvestigating
}

// SLAMetrics はSLAパフォーマンスメトリクスを表します
type SLAMetrics struct {
	TotalIncidents      int64   `json:"total_incidents"`
	ResolvedIncidents   int64   `json:"resolved_incidents"`
	SLAViolatedCount    int64   `json:"sla_violated_count"`
	SLAComplianceRate   float64 `json:"sla_compliance_rate"`    // SLA内で解決されたインシデントの割合
	AverageMTTR         float64 `json:"average_mttr"`            // 平均解決時間（時間単位）
	MedianMTTR          float64 `json:"median_mttr"`             // 解決時間の中央値（時間単位）
	CurrentlyOverdue    int64   `json:"currently_overdue"`       // SLA期限を過ぎた未解決インシデント数
}

type IncidentRepository interface {
	Create(ctx context.Context, incident *Incident) error
	FindAll(ctx context.Context, filters IncidentFilters, pagination Pagination) ([]*Incident, *PaginationResult, error)
	FindByID(ctx context.Context, id uint) (*Incident, error)
	Update(ctx context.Context, incident *Incident) error
	UpdateAssignee(ctx context.Context, incidentID uint, assigneeID *uint) error
	Delete(ctx context.Context, id uint) error

	// 統計メソッド
	Count(count *int64) error
	CountBySeverity(severity Severity, count *int64) error
	CountByStatus(status Status, count *int64) error
	FindRecent(limit int) ([]*Incident, error)
	GetAllIncidents() ([]*Incident, error)

	// SLA メソッド
	CountSLAViolated(count *int64) error
	GetSLAMetrics() (*SLAMetrics, error)
}
