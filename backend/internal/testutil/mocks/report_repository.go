package mocks

import (
	"incidex/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockReportRepository は domain.ReportRepository のモック実装です
type MockReportRepository struct {
	mock.Mock
}

func NewMockReportRepository() *MockReportRepository {
	return &MockReportRepository{}
}

func (m *MockReportRepository) GetMonthlyReport(startDate, endDate time.Time) (*domain.MonthlyReport, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MonthlyReport), args.Error(1)
}

func (m *MockReportRepository) GetIncidentCountByDay(startDate, endDate time.Time) ([]domain.DailyIncidentCount, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DailyIncidentCount), args.Error(1)
}

func (m *MockReportRepository) GetTopTags(startDate, endDate time.Time, limit int) ([]domain.TagStatistic, error) {
	args := m.Called(startDate, endDate, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TagStatistic), args.Error(1)
}
