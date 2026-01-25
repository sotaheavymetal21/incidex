package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReportUsecase_GetMonthlyReport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns monthly report", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		expectedReport := testutil.NewTestMonthlyReport()

		// The usecase calculates start and end dates from year and month
		reportRepo.On("GetMonthlyReport", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(expectedReport, nil)

		report, err := usecase.GetMonthlyReport(ctx, 2024, 1)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 10, report.Summary.TotalIncidents)

		reportRepo.AssertExpectations(t)
	})

	t.Run("calculates correct date range for January", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		expectedReport := testutil.NewTestMonthlyReport()

		expectedStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedEndDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

		reportRepo.On("GetMonthlyReport", expectedStartDate, expectedEndDate).Return(expectedReport, nil)

		report, err := usecase.GetMonthlyReport(ctx, 2024, 1)

		require.NoError(t, err)
		assert.NotNil(t, report)

		reportRepo.AssertExpectations(t)
	})

	t.Run("calculates correct date range for December", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		expectedReport := testutil.NewTestMonthlyReport()

		expectedStartDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		expectedEndDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

		reportRepo.On("GetMonthlyReport", expectedStartDate, expectedEndDate).Return(expectedReport, nil)

		report, err := usecase.GetMonthlyReport(ctx, 2024, 12)

		require.NoError(t, err)
		assert.NotNil(t, report)

		reportRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		reportRepo.On("GetMonthlyReport", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(nil, errors.New("database error"))

		report, err := usecase.GetMonthlyReport(ctx, 2024, 1)

		require.Error(t, err)
		assert.Nil(t, report)

		reportRepo.AssertExpectations(t)
	})
}

func TestReportUsecase_GetCustomReport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns custom report", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		expectedReport := testutil.NewTestMonthlyReport()

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

		reportRepo.On("GetMonthlyReport", startDate, endDate).Return(expectedReport, nil)

		report, err := usecase.GetCustomReport(ctx, startDate, endDate)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 10, report.Summary.TotalIncidents)

		reportRepo.AssertExpectations(t)
	})

	t.Run("supports multi-month date range", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		expectedReport := testutil.NewTestMonthlyReport(func(r *domain.MonthlyReport) {
			r.Summary.TotalIncidents = 50
		})

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

		reportRepo.On("GetMonthlyReport", startDate, endDate).Return(expectedReport, nil)

		report, err := usecase.GetCustomReport(ctx, startDate, endDate)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 50, report.Summary.TotalIncidents)

		reportRepo.AssertExpectations(t)
	})

	t.Run("handles single day report", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		expectedReport := testutil.NewTestMonthlyReport(func(r *domain.MonthlyReport) {
			r.Summary.TotalIncidents = 2
		})

		date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 6, 15, 23, 59, 59, 0, time.UTC)

		reportRepo.On("GetMonthlyReport", date, endDate).Return(expectedReport, nil)

		report, err := usecase.GetCustomReport(ctx, date, endDate)

		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 2, report.Summary.TotalIncidents)

		reportRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		reportRepo := mocks.NewMockReportRepository()
		usecase := NewReportUsecase(reportRepo)

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

		reportRepo.On("GetMonthlyReport", startDate, endDate).Return(nil, errors.New("database error"))

		report, err := usecase.GetCustomReport(ctx, startDate, endDate)

		require.Error(t, err)
		assert.Nil(t, report)

		reportRepo.AssertExpectations(t)
	})
}
