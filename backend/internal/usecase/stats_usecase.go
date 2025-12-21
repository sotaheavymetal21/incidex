package usecase

import (
	"incidex/internal/domain"
	"time"
)

type StatsUsecase struct {
	incidentRepo domain.IncidentRepository
}

func NewStatsUsecase(incidentRepo domain.IncidentRepository) *StatsUsecase {
	return &StatsUsecase{
		incidentRepo: incidentRepo,
	}
}

type DashboardStats struct {
	TotalIncidents    int64                      `json:"total_incidents"`
	BySeverity        map[string]int64           `json:"by_severity"`
	ByStatus          map[string]int64           `json:"by_status"`
	RecentIncidents   []*domain.Incident         `json:"recent_incidents"`
	TrendData         []TrendDataPoint           `json:"trend_data"`
}

type TrendDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (u *StatsUsecase) GetDashboardStats(period string) (*DashboardStats, error) {
	// 総件数
	var totalCount int64
	if err := u.incidentRepo.Count(&totalCount); err != nil {
		return nil, err
	}

	// 深刻度別集計
	bySeverity := make(map[string]int64)
	severities := []domain.Severity{
		domain.SeverityCritical,
		domain.SeverityHigh,
		domain.SeverityMedium,
		domain.SeverityLow,
	}
	for _, severity := range severities {
		var count int64
		if err := u.incidentRepo.CountBySeverity(severity, &count); err != nil {
			return nil, err
		}
		bySeverity[string(severity)] = count
	}

	// ステータス別集計
	byStatus := make(map[string]int64)
	statuses := []domain.Status{
		domain.StatusOpen,
		domain.StatusInvestigating,
		domain.StatusResolved,
		domain.StatusClosed,
	}
	for _, status := range statuses {
		var count int64
		if err := u.incidentRepo.CountByStatus(status, &count); err != nil {
			return nil, err
		}
		byStatus[string(status)] = count
	}

	// 最近のインシデント（直近10件）
	recentIncidents, err := u.incidentRepo.FindRecent(10)
	if err != nil {
		return nil, err
	}

	// トレンドデータの生成
	trendData, err := u.generateTrendData(period)
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		TotalIncidents:  totalCount,
		BySeverity:      bySeverity,
		ByStatus:        byStatus,
		RecentIncidents: recentIncidents,
		TrendData:       trendData,
	}, nil
}

func (u *StatsUsecase) generateTrendData(period string) ([]TrendDataPoint, error) {
	now := time.Now()
	var days int
	var dateFormat string

	switch period {
	case "daily":
		days = 30
		dateFormat = "2006-01-02"
	case "weekly":
		days = 84 // 12週間
		dateFormat = "2006-01-02"
	case "monthly":
		days = 365 // 12ヶ月
		dateFormat = "2006-01"
	default:
		days = 30
		dateFormat = "2006-01-02"
	}

	trendData := make([]TrendDataPoint, 0)
	dataMap := make(map[string]int64)

	// 全インシデントを取得
	allIncidents, err := u.incidentRepo.GetAllIncidents()
	if err != nil {
		return nil, err
	}

	// 期間内のインシデントを日付ごとに集計
	cutoffDate := now.AddDate(0, 0, -days)
	for _, incident := range allIncidents {
		if incident.DetectedAt.After(cutoffDate) {
			dateKey := incident.DetectedAt.Format(dateFormat)
			dataMap[dateKey]++
		}
	}

	// データポイントを生成（期間内の全日付）
	if period == "weekly" {
		// 週次集計
		for i := 0; i < 12; i++ {
			weekStart := now.AddDate(0, 0, -7*(11-i))
			weekEnd := weekStart.AddDate(0, 0, 7)
			weekLabel := weekStart.Format("01/02")

			var weekCount int64
			for _, incident := range allIncidents {
				if incident.DetectedAt.After(weekStart) && incident.DetectedAt.Before(weekEnd) {
					weekCount++
				}
			}
			trendData = append(trendData, TrendDataPoint{
				Date:  weekLabel,
				Count: weekCount,
			})
		}
	} else if period == "monthly" {
		// 月次集計
		for i := 0; i < 12; i++ {
			month := now.AddDate(0, -(11 - i), 0)
			monthKey := month.Format("2006-01")
			trendData = append(trendData, TrendDataPoint{
				Date:  month.Format("01月"),
				Count: dataMap[monthKey],
			})
		}
	} else {
		// 日次集計
		for i := 0; i < days; i++ {
			date := now.AddDate(0, 0, -(days-1-i))
			dateKey := date.Format(dateFormat)
			trendData = append(trendData, TrendDataPoint{
				Date:  date.Format("01/02"),
				Count: dataMap[dateKey],
			})
		}
	}

	return trendData, nil
}

// GetSLAMetrics returns SLA performance metrics
func (u *StatsUsecase) GetSLAMetrics() (*domain.SLAMetrics, error) {
	return u.incidentRepo.GetSLAMetrics()
}
