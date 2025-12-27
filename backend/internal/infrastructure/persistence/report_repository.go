package persistence

import (
	"incidex/internal/domain"
	"time"

	"gorm.io/gorm"
)

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) domain.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetMonthlyReport(startDate, endDate time.Time) (*domain.MonthlyReport, error) {
	report := &domain.MonthlyReport{
		Period: domain.ReportPeriod{
			StartDate: startDate,
			EndDate:   endDate,
			Month:     int(startDate.Month()),
			Year:      startDate.Year(),
		},
	}

	// Get summary statistics
	summary, err := r.getIncidentSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.Summary = *summary

	// Get severity breakdown
	severityBreakdown, err := r.getSeverityBreakdown(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.SeverityBreakdown = severityBreakdown

	// Get status breakdown
	statusBreakdown, err := r.getStatusBreakdown(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.StatusBreakdown = statusBreakdown

	// Get daily trend
	dailyTrend, err := r.GetIncidentCountByDay(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.DailyTrend = dailyTrend

	// Get top tags
	topTags, err := r.GetTopTags(startDate, endDate, 10)
	if err != nil {
		return nil, err
	}
	report.TopTags = topTags

	// Get performance metrics
	metrics, err := r.getPerformanceMetrics(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.PerformanceMetrics = *metrics

	// Get comparison with previous period
	comparison, err := r.getPeriodComparison(startDate, endDate)
	if err == nil {
		report.Comparison = comparison
	}

	return report, nil
}

func (r *reportRepository) getIncidentSummary(startDate, endDate time.Time) (*domain.IncidentSummary, error) {
	var summary domain.IncidentSummary

	// Total incidents created in period
	err := r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&[]int64{int64(summary.TotalIncidents)}[0]).Error
	if err != nil {
		return nil, err
	}

	var totalCount int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	summary.TotalIncidents = int(totalCount)
	summary.NewIncidents = int(totalCount) // Same as total for the period

	// Resolved incidents in period
	var resolvedCount int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("status = ?", domain.StatusResolved).
		Count(&resolvedCount).Error
	if err != nil {
		return nil, err
	}
	summary.ResolvedIncidents = int(resolvedCount)

	// Open incidents (created in period and still open)
	var openCount int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("status IN ?", []domain.Status{
			domain.StatusOpen,
			domain.StatusInvestigating,
		}).
		Count(&openCount).Error
	if err != nil {
		return nil, err
	}
	summary.OpenIncidents = int(openCount)

	// Critical incidents
	var criticalCount int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("severity = ?", domain.SeverityCritical).
		Count(&criticalCount).Error
	if err != nil {
		return nil, err
	}
	summary.CriticalIncidents = int(criticalCount)

	return &summary, nil
}

func (r *reportRepository) getSeverityBreakdown(startDate, endDate time.Time) (map[string]int, error) {
	type SeverityCount struct {
		Severity string
		Count    int
	}

	var results []SeverityCount
	err := r.db.Model(&domain.Incident{}).
		Select("severity, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("severity").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	breakdown := make(map[string]int)
	for _, result := range results {
		breakdown[result.Severity] = result.Count
	}

	return breakdown, nil
}

func (r *reportRepository) getStatusBreakdown(startDate, endDate time.Time) (map[string]int, error) {
	type StatusCount struct {
		Status string
		Count  int
	}

	var results []StatusCount
	err := r.db.Model(&domain.Incident{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	breakdown := make(map[string]int)
	for _, result := range results {
		breakdown[result.Status] = result.Count
	}

	return breakdown, nil
}

func (r *reportRepository) GetIncidentCountByDay(startDate, endDate time.Time) ([]domain.DailyIncidentCount, error) {
	type DayCount struct {
		Date  time.Time
		Count int
	}

	var results []DayCount
	err := r.db.Model(&domain.Incident{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	dailyCounts := make([]domain.DailyIncidentCount, 0, len(results))
	for _, result := range results {
		dailyCounts = append(dailyCounts, domain.DailyIncidentCount{
			Date:  result.Date,
			Count: result.Count,
		})
	}

	return dailyCounts, nil
}

func (r *reportRepository) GetTopTags(startDate, endDate time.Time, limit int) ([]domain.TagStatistic, error) {
	type TagCount struct {
		TagID   uint
		TagName string
		Count   int
	}

	var results []TagCount
	err := r.db.Table("incidents").
		Select("tags.id as tag_id, tags.name as tag_name, COUNT(*) as count").
		Joins("JOIN incident_tags ON incidents.id = incident_tags.incident_id").
		Joins("JOIN tags ON incident_tags.tag_id = tags.id").
		Where("incidents.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("tags.id, tags.name").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	tagStats := make([]domain.TagStatistic, 0, len(results))
	for _, result := range results {
		tagStats = append(tagStats, domain.TagStatistic{
			TagID:   result.TagID,
			TagName: result.TagName,
			Count:   result.Count,
		})
	}

	return tagStats, nil
}

func (r *reportRepository) getPerformanceMetrics(startDate, endDate time.Time) (*domain.PerformanceMetrics, error) {
	metrics := &domain.PerformanceMetrics{}

	// Get resolved incidents
	var resolvedIncidents []domain.Incident
	err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("status = ?", domain.StatusResolved).
		Where("resolved_at IS NOT NULL").
		Find(&resolvedIncidents).Error
	if err != nil {
		return nil, err
	}

	if len(resolvedIncidents) > 0 {
		// Calculate average resolution time
		var totalHours float64
		var count int

		for _, incident := range resolvedIncidents {
			if incident.ResolvedAt != nil {
				// Use DetectedAt instead of CreatedAt for accurate resolution time
				hours := incident.ResolvedAt.Sub(incident.DetectedAt).Hours()
				// Only include positive values
				if hours >= 0 {
					totalHours += hours
					count++
				}
			}
		}

		if count > 0 {
			metrics.AverageResolutionTime = totalHours / float64(count)
		}
	}

	return metrics, nil
}

func (r *reportRepository) getPeriodComparison(startDate, endDate time.Time) (*domain.PeriodComparison, error) {
	// Calculate previous period (same duration)
	duration := endDate.Sub(startDate)
	prevStartDate := startDate.Add(-duration)
	prevEndDate := startDate

	// Get current period totals
	var currentTotal int64
	err := r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&currentTotal).Error
	if err != nil {
		return nil, err
	}

	var currentResolved int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("status = ?", domain.StatusResolved).
		Count(&currentResolved).Error
	if err != nil {
		return nil, err
	}

	// Get previous period totals
	var previousTotal int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", prevStartDate, prevEndDate).
		Count(&previousTotal).Error
	if err != nil {
		return nil, err
	}

	var previousResolved int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", prevStartDate, prevEndDate).
		Where("status = ?", domain.StatusResolved).
		Count(&previousResolved).Error
	if err != nil {
		return nil, err
	}

	comparison := &domain.PeriodComparison{
		PreviousPeriod: domain.ReportPeriod{
			StartDate: prevStartDate,
			EndDate:   prevEndDate,
			Month:     int(prevStartDate.Month()),
			Year:      prevStartDate.Year(),
		},
		TotalIncidentsChange: int(currentTotal - previousTotal),
		ResolvedIncidentsChange: int(currentResolved - previousResolved),
	}

	// Calculate percentage changes
	if previousTotal > 0 {
		comparison.TotalIncidentsChangePercent = float64(currentTotal-previousTotal) / float64(previousTotal) * 100
	} else if currentTotal > 0 {
		// If previous period had no data but current has, show 100% increase
		comparison.TotalIncidentsChangePercent = 100.0
	}

	if previousResolved > 0 {
		comparison.ResolvedIncidentsChangePercent = float64(currentResolved-previousResolved) / float64(previousResolved) * 100
	} else if currentResolved > 0 {
		// If previous period had no data but current has, show 100% increase
		comparison.ResolvedIncidentsChangePercent = 100.0
	}

	return comparison, nil
}
