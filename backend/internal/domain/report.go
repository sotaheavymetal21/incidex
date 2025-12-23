package domain

import "time"

// MonthlyReport represents a comprehensive monthly incident report
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

// ReportPeriod defines the time period for the report
type ReportPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Month     int       `json:"month"`
	Year      int       `json:"year"`
}

// IncidentSummary provides high-level statistics
type IncidentSummary struct {
	TotalIncidents    int `json:"total_incidents"`
	NewIncidents      int `json:"new_incidents"`
	ResolvedIncidents int `json:"resolved_incidents"`
	OpenIncidents     int `json:"open_incidents"`
	CriticalIncidents int `json:"critical_incidents"`
}

// DailyIncidentCount tracks incidents per day
type DailyIncidentCount struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// TagStatistic shows tag usage statistics
type TagStatistic struct {
	TagID   uint   `json:"tag_id"`
	TagName string `json:"tag_name"`
	Count   int    `json:"count"`
}

// PerformanceMetrics tracks performance indicators
type PerformanceMetrics struct {
	AverageResolutionTime float64 `json:"average_resolution_time_hours"`
	MedianResolutionTime  float64 `json:"median_resolution_time_hours"`
	SLAComplianceRate     float64 `json:"sla_compliance_rate"`
	MeanTimeToAcknowledge float64 `json:"mean_time_to_acknowledge_hours"`
}

// PeriodComparison compares current period with previous period
type PeriodComparison struct {
	PreviousPeriod    ReportPeriod `json:"previous_period"`
	TotalIncidentsChange    int    `json:"total_incidents_change"`
	TotalIncidentsChangePercent float64 `json:"total_incidents_change_percent"`
	ResolvedIncidentsChange int    `json:"resolved_incidents_change"`
	ResolvedIncidentsChangePercent float64 `json:"resolved_incidents_change_percent"`
}

// ReportRepository defines operations for generating reports
type ReportRepository interface {
	GetMonthlyReport(startDate, endDate time.Time) (*MonthlyReport, error)
	GetIncidentCountByDay(startDate, endDate time.Time) ([]DailyIncidentCount, error)
	GetTopTags(startDate, endDate time.Time, limit int) ([]TagStatistic, error)
}
