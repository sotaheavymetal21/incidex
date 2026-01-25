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
	"github.com/stretchr/testify/require"
)

func TestAuditLogUsecase_GetAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns audit logs", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		expectedLogs := []*domain.AuditLog{
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 1 }),
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 2 }),
		}

		filters := domain.AuditLogFilters{Page: 1, Limit: 10}
		auditLogRepo.On("FindAll", ctx, filters).Return(expectedLogs, int64(2), nil)

		logs, total, err := usecase.GetAll(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, logs, 2)
		assert.Equal(t, int64(2), total)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no logs exist", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		filters := domain.AuditLogFilters{Page: 1, Limit: 10}
		auditLogRepo.On("FindAll", ctx, filters).Return([]*domain.AuditLog{}, int64(0), nil)

		logs, total, err := usecase.GetAll(ctx, filters)

		require.NoError(t, err)
		assert.Empty(t, logs)
		assert.Equal(t, int64(0), total)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("filters by user ID", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		userID := uint(5)
		expectedLogs := []*domain.AuditLog{
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.UserID = &userID }),
		}

		filters := domain.AuditLogFilters{UserID: &userID, Page: 1, Limit: 10}
		auditLogRepo.On("FindAll", ctx, filters).Return(expectedLogs, int64(1), nil)

		logs, total, err := usecase.GetAll(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, int64(1), total)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("filters by action type", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		action := domain.AuditActionCreate
		expectedLogs := []*domain.AuditLog{
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.Action = action }),
		}

		filters := domain.AuditLogFilters{Action: &action, Page: 1, Limit: 10}
		auditLogRepo.On("FindAll", ctx, filters).Return(expectedLogs, int64(1), nil)

		logs, total, err := usecase.GetAll(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, int64(1), total)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("filters by resource type", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		resourceType := "incident"
		expectedLogs := []*domain.AuditLog{
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ResourceType = resourceType }),
		}

		filters := domain.AuditLogFilters{ResourceType: &resourceType, Page: 1, Limit: 10}
		auditLogRepo.On("FindAll", ctx, filters).Return(expectedLogs, int64(1), nil)

		logs, total, err := usecase.GetAll(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, int64(1), total)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("filters by date range", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		expectedLogs := []*domain.AuditLog{
			testutil.NewTestAuditLog(),
		}

		filters := domain.AuditLogFilters{
			StartDate: &startDate,
			EndDate:   &endDate,
			Page:      1,
			Limit:     10,
		}
		auditLogRepo.On("FindAll", ctx, filters).Return(expectedLogs, int64(1), nil)

		logs, total, err := usecase.GetAll(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, int64(1), total)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		filters := domain.AuditLogFilters{Page: 1, Limit: 10}
		auditLogRepo.On("FindAll", ctx, filters).Return(nil, int64(0), errors.New("database error"))

		logs, total, err := usecase.GetAll(ctx, filters)

		require.Error(t, err)
		assert.Nil(t, logs)
		assert.Equal(t, int64(0), total)

		auditLogRepo.AssertExpectations(t)
	})
}

func TestAuditLogUsecase_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns audit log by ID", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		expectedLog := testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 1 })
		auditLogRepo.On("FindByID", ctx, uint(1)).Return(expectedLog, nil)

		log, err := usecase.GetByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, log)
		assert.Equal(t, uint(1), log.ID)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("returns nil when log not found", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		auditLogRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		log, err := usecase.GetByID(ctx, 999)

		require.NoError(t, err)
		assert.Nil(t, log)

		auditLogRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		auditLogRepo := mocks.NewMockAuditLogRepository()
		usecase := NewAuditLogUsecase(auditLogRepo)

		auditLogRepo.On("FindByID", ctx, uint(1)).Return(nil, errors.New("database error"))

		log, err := usecase.GetByID(ctx, 1)

		require.Error(t, err)
		assert.Nil(t, log)

		auditLogRepo.AssertExpectations(t)
	})
}
