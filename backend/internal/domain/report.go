package domain

import "time"

// MonthlyReport は包括的な月次インシデントレポートを表します
type MonthlyReport struct {
	Period           ReportPeriod            `json:"period"`
	Summary          IncidentSummary         `json:"summary"`
	SeverityBreakdown map[string]int         `json:"severity_breakdown"`
	StatusBreakdown  map[string]int          `json:"status_breakdown"`
	DailyTrend       []DailyIncidentCount    `json:"daily_trend"`
	TopTags          []TagStatistic          `json:"top_tags"`
	PerformanceMetrics PerformanceMetrics    `json:"performance_metrics"`
	Comparison       *PeriodComparison       `json:"comparison,omitempty"`
}

// ReportPeriod はレポートの期間を定義します
type ReportPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Month     int       `json:"month"`
	Year      int       `json:"year"`
}

// IncidentSummary は高レベルの統計情報を提供します
type IncidentSummary struct {
	TotalIncidents    int `json:"total_incidents"`
	NewIncidents      int `json:"new_incidents"`
	ResolvedIncidents int `json:"resolved_incidents"`
	OpenIncidents     int `json:"open_incidents"`
	CriticalIncidents int `json:"critical_incidents"`
}

// DailyIncidentCount は日ごとのインシデント数を追跡します
type DailyIncidentCount struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// TagStatistic はタグ使用統計を表示します
type TagStatistic struct {
	TagID   uint   `json:"tag_id"`
	TagName string `json:"tag_name"`
	Count   int    `json:"count"`
}

// PerformanceMetrics はパフォーマンス指標を追跡します
type PerformanceMetrics struct {
	AverageResolutionTime float64 `json:"average_resolution_time_hours"`
}

// PeriodComparison は現在の期間と前の期間を比較します
type PeriodComparison struct {
	PreviousPeriod    ReportPeriod `json:"previous_period"`
	TotalIncidentsChange    int    `json:"total_incidents_change"`
	TotalIncidentsChangePercent float64 `json:"total_incidents_change_percent"`
	ResolvedIncidentsChange int    `json:"resolved_incidents_change"`
	ResolvedIncidentsChangePercent float64 `json:"resolved_incidents_change_percent"`
}

// ReportRepository はレポート生成のための操作を定義します
type ReportRepository interface {
	GetMonthlyReport(startDate, endDate time.Time) (*MonthlyReport, error)
	GetIncidentCountByDay(startDate, endDate time.Time) ([]DailyIncidentCount, error)
	GetTopTags(startDate, endDate time.Time, limit int) ([]TagStatistic, error)
}
