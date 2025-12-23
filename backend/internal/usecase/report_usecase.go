package usecase

import (
	"context"
	"incidex/internal/domain"
	"time"
)

type ReportUsecase interface {
	GetMonthlyReport(ctx context.Context, year, month int) (*domain.MonthlyReport, error)
	GetCustomReport(ctx context.Context, startDate, endDate time.Time) (*domain.MonthlyReport, error)
}

type reportUsecase struct {
	reportRepo domain.ReportRepository
}

func NewReportUsecase(reportRepo domain.ReportRepository) ReportUsecase {
	return &reportUsecase{
		reportRepo: reportRepo,
	}
}

func (u *reportUsecase) GetMonthlyReport(ctx context.Context, year, month int) (*domain.MonthlyReport, error) {
	// Calculate start and end dates for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second) // Last second of the month

	return u.reportRepo.GetMonthlyReport(startDate, endDate)
}

func (u *reportUsecase) GetCustomReport(ctx context.Context, startDate, endDate time.Time) (*domain.MonthlyReport, error) {
	return u.reportRepo.GetMonthlyReport(startDate, endDate)
}
