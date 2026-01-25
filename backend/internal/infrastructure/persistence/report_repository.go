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

	// サマリー統計を取得します
	summary, err := r.getIncidentSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.Summary = *summary

	// 重大度別の内訳を取得します
	severityBreakdown, err := r.getSeverityBreakdown(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.SeverityBreakdown = severityBreakdown

	// ステータス別の内訳を取得します
	statusBreakdown, err := r.getStatusBreakdown(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.StatusBreakdown = statusBreakdown

	// 日別トレンドを取得します
	dailyTrend, err := r.GetIncidentCountByDay(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.DailyTrend = dailyTrend

	// 上位タグを取得します
	topTags, err := r.GetTopTags(startDate, endDate, 10)
	if err != nil {
		return nil, err
	}
	report.TopTags = topTags

	// パフォーマンスメトリクスを取得します
	metrics, err := r.getPerformanceMetrics(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.PerformanceMetrics = *metrics

	// 前期間との比較を取得します
	comparison, err := r.getPeriodComparison(startDate, endDate)
	if err == nil {
		report.Comparison = comparison
	}

	return report, nil
}

func (r *reportRepository) getIncidentSummary(startDate, endDate time.Time) (*domain.IncidentSummary, error) {
	var summary domain.IncidentSummary

	// 期間中に作成された総インシデント数
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
	summary.NewIncidents = int(totalCount) // 期間中は合計と同じです

	// 期間中に解決されたインシデント
	var resolvedCount int64
	err = r.db.Model(&domain.Incident{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("status = ?", domain.StatusResolved).
		Count(&resolvedCount).Error
	if err != nil {
		return nil, err
	}
	summary.ResolvedIncidents = int(resolvedCount)

	// オープン中のインシデント（期間中に作成され、まだオープン状態のもの）
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

	// クリティカルインシデント
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

	// 解決済みインシデントを取得します
	var resolvedIncidents []domain.Incident
	err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("status = ?", domain.StatusResolved).
		Where("resolved_at IS NOT NULL").
		Find(&resolvedIncidents).Error
	if err != nil {
		return nil, err
	}

	if len(resolvedIncidents) > 0 {
		// 平均解決時間を計算します
		var totalHours float64
		var count int

		for _, incident := range resolvedIncidents {
			if incident.ResolvedAt != nil {
				// 正確な解決時間のためにCreatedAtではなくDetectedAtを使用します
				hours := incident.ResolvedAt.Sub(incident.DetectedAt).Hours()
				// 正の値のみを含めます
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
	// 前期間を計算します（同じ期間の長さ）
	duration := endDate.Sub(startDate)
	prevStartDate := startDate.Add(-duration)
	prevEndDate := startDate

	// 現在期間の合計を取得します
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

	// 前期間の合計を取得します
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

	// パーセンテージ変化を計算します
	if previousTotal > 0 {
		comparison.TotalIncidentsChangePercent = float64(currentTotal-previousTotal) / float64(previousTotal) * 100
	} else if currentTotal > 0 {
		// 前期間にデータがなく現在期間にデータがある場合、100%増加として表示します
		comparison.TotalIncidentsChangePercent = 100.0
	}

	if previousResolved > 0 {
		comparison.ResolvedIncidentsChangePercent = float64(currentResolved-previousResolved) / float64(previousResolved) * 100
	} else if currentResolved > 0 {
		// 前期間にデータがなく現在期間にデータがある場合、100%増加として表示します
		comparison.ResolvedIncidentsChangePercent = 100.0
	}

	return comparison, nil
}
