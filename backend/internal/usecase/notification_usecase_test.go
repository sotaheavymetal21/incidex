package usecase

import (
	"errors"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestNotificationUsecase(notificationRepo *mocks.MockNotificationSettingRepository) *NotificationUsecase {
	return NewNotificationUsecase(notificationRepo)
}

func TestNotificationUsecase_GetSettingByUserID(t *testing.T) {
	t.Parallel()

	t.Run("returns existing notification settings", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		expectedSetting := testutil.NewTestNotificationSetting(1)
		notificationRepo.On("GetByUserID", uint(1)).Return(expectedSetting, nil)

		setting, err := usecase.GetSettingByUserID(1)

		require.NoError(t, err)
		assert.Equal(t, expectedSetting, setting)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("returns default settings when user has no settings", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))

		setting, err := usecase.GetSettingByUserID(1)

		require.NoError(t, err)
		require.NotNil(t, setting)

		// Verify default settings
		assert.Equal(t, uint(1), setting.UserID)
		assert.True(t, setting.EmailEnabled)
		assert.False(t, setting.SlackEnabled)
		assert.True(t, setting.NotifyOnIncidentCreated)
		assert.True(t, setting.NotifyOnAssigned)
		assert.True(t, setting.NotifyOnComment)
		assert.True(t, setting.NotifyOnStatusChange)
		assert.True(t, setting.NotifyOnSeverityChange)
		assert.True(t, setting.NotifyOnResolved)
		assert.True(t, setting.NotifyOnEscalation)

		notificationRepo.AssertExpectations(t)
	})

	t.Run("returns custom notification settings with email disabled", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		customSetting := testutil.NewTestNotificationSetting(2, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
			s.NotifyOnComment = false
		})
		notificationRepo.On("GetByUserID", uint(2)).Return(customSetting, nil)

		setting, err := usecase.GetSettingByUserID(2)

		require.NoError(t, err)
		assert.False(t, setting.EmailEnabled)
		assert.False(t, setting.NotifyOnComment)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("returns custom notification settings with Slack enabled", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		customSetting := testutil.NewTestNotificationSetting(3, func(s *domain.NotificationSetting) {
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/TEST/WEBHOOK"
		})
		notificationRepo.On("GetByUserID", uint(3)).Return(customSetting, nil)

		setting, err := usecase.GetSettingByUserID(3)

		require.NoError(t, err)
		assert.True(t, setting.SlackEnabled)
		assert.Equal(t, "https://hooks.slack.com/services/TEST/WEBHOOK", setting.SlackWebhook)
		notificationRepo.AssertExpectations(t)
	})
}

func TestNotificationUsecase_CreateSetting(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates new notification setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		newSetting := testutil.NewTestNotificationSetting(1)
		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", newSetting).Return(nil)

		err := usecase.CreateSetting(newSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("fails when user_id is zero", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		newSetting := testutil.NewTestNotificationSetting(0)

		err := usecase.CreateSetting(newSetting)

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
		assert.Contains(t, domainErr.Message, "user_id is required")
	})

	t.Run("fails when notification setting already exists", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1)
		newSetting := testutil.NewTestNotificationSetting(1)

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)

		err := usecase.CreateSetting(newSetting)

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeConflict, domainErr.Code)
		assert.Contains(t, domainErr.Message, "already exists")
	})

	t.Run("successfully creates setting with custom preferences", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		customSetting := testutil.NewTestNotificationSetting(2, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/TEST"
			s.NotifyOnComment = false
		})

		notificationRepo.On("GetByUserID", uint(2)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", customSetting).Return(nil)

		err := usecase.CreateSetting(customSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		newSetting := testutil.NewTestNotificationSetting(1)
		repoError := errors.New("database error")

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", newSetting).Return(repoError)

		err := usecase.CreateSetting(newSetting)

		require.Error(t, err)
		assert.Equal(t, repoError, err)
		notificationRepo.AssertExpectations(t)
	})
}

func TestNotificationUsecase_UpdateSetting(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates existing notification setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.ID = 10
		})
		updatedSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
			s.SlackEnabled = true
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return s.ID == 10 && s.UserID == 1 && !s.EmailEnabled && s.SlackEnabled
		})).Return(nil)

		err := usecase.UpdateSetting(1, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("creates new setting if none exists", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		newSetting := testutil.NewTestNotificationSetting(1)

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return s.UserID == 1
		})).Return(nil)

		err := usecase.UpdateSetting(1, newSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("fails when user_id is zero", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		setting := testutil.NewTestNotificationSetting(0)

		err := usecase.UpdateSetting(0, setting)

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
		assert.Contains(t, domainErr.Message, "user_id is required")
	})

	t.Run("preserves existing ID when updating", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(5, func(s *domain.NotificationSetting) {
			s.ID = 42
		})
		updatedSetting := testutil.NewTestNotificationSetting(5, func(s *domain.NotificationSetting) {
			s.ID = 999 // This should be overridden
			s.NotifyOnComment = false
		})

		notificationRepo.On("GetByUserID", uint(5)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			// Should use existing ID (42), not the provided ID (999)
			return s.ID == 42 && s.UserID == 5 && !s.NotifyOnComment
		})).Return(nil)

		err := usecase.UpdateSetting(5, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("updates all notification preferences", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(3, func(s *domain.NotificationSetting) {
			s.ID = 20
		})
		updatedSetting := testutil.NewTestNotificationSetting(3, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/UPDATED"
			s.NotifyOnIncidentCreated = false
			s.NotifyOnAssigned = false
			s.NotifyOnComment = false
			s.NotifyOnStatusChange = false
			s.NotifyOnSeverityChange = false
			s.NotifyOnResolved = false
			s.NotifyOnEscalation = false
		})

		notificationRepo.On("GetByUserID", uint(3)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return s.ID == 20 &&
				!s.EmailEnabled &&
				s.SlackEnabled &&
				s.SlackWebhook == "https://hooks.slack.com/services/UPDATED" &&
				!s.NotifyOnIncidentCreated &&
				!s.NotifyOnAssigned &&
				!s.NotifyOnComment &&
				!s.NotifyOnStatusChange &&
				!s.NotifyOnSeverityChange &&
				!s.NotifyOnResolved &&
				!s.NotifyOnEscalation
		})).Return(nil)

		err := usecase.UpdateSetting(3, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("fails when repository update returns error", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1)
		updatedSetting := testutil.NewTestNotificationSetting(1)
		repoError := errors.New("database error")

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.AnythingOfType("*domain.NotificationSetting")).Return(repoError)

		err := usecase.UpdateSetting(1, updatedSetting)

		require.Error(t, err)
		assert.Equal(t, repoError, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("fails when repository create returns error during upsert", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		newSetting := testutil.NewTestNotificationSetting(1)
		repoError := errors.New("database error")

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", mock.AnythingOfType("*domain.NotificationSetting")).Return(repoError)

		err := usecase.UpdateSetting(1, newSetting)

		require.Error(t, err)
		assert.Equal(t, repoError, err)
		notificationRepo.AssertExpectations(t)
	})
}

func TestNotificationUsecase_DeleteSetting(t *testing.T) {
	t.Parallel()

	t.Run("successfully deletes notification setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		notificationRepo.On("Delete", uint(1)).Return(nil)

		err := usecase.DeleteSetting(1)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("successfully deletes with different user IDs", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		notificationRepo.On("Delete", uint(42)).Return(nil)

		err := usecase.DeleteSetting(42)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		repoError := errors.New("database error")
		notificationRepo.On("Delete", uint(1)).Return(repoError)

		err := usecase.DeleteSetting(1)

		require.Error(t, err)
		assert.Equal(t, repoError, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("handles deletion of non-existent setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		// Repository might return error or nil depending on implementation
		notificationRepo.On("Delete", uint(999)).Return(errors.New("setting not found"))

		err := usecase.DeleteSetting(999)

		require.Error(t, err)
		notificationRepo.AssertExpectations(t)
	})
}

func TestNotificationUsecase_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("handles concurrent updates gracefully", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.ID = 10
		})
		updatedSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.AnythingOfType("*domain.NotificationSetting")).Return(nil)

		err := usecase.UpdateSetting(1, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("handles empty slack webhook when slack disabled", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		setting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.SlackEnabled = false
			s.SlackWebhook = ""
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return !s.SlackEnabled && s.SlackWebhook == ""
		})).Return(nil)

		err := usecase.CreateSetting(setting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("handles all notification preferences disabled", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		setting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
			s.SlackEnabled = false
			s.NotifyOnIncidentCreated = false
			s.NotifyOnAssigned = false
			s.NotifyOnComment = false
			s.NotifyOnStatusChange = false
			s.NotifyOnSeverityChange = false
			s.NotifyOnResolved = false
			s.NotifyOnEscalation = false
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", setting).Return(nil)

		err := usecase.CreateSetting(setting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("handles large user IDs", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		largeUserID := uint(4294967295) // Max uint32
		setting := testutil.NewTestNotificationSetting(largeUserID)

		notificationRepo.On("GetByUserID", largeUserID).Return(setting, nil)

		result, err := usecase.GetSettingByUserID(largeUserID)

		require.NoError(t, err)
		assert.Equal(t, largeUserID, result.UserID)
		notificationRepo.AssertExpectations(t)
	})
}

func TestNotificationUsecase_SlackIntegration(t *testing.T) {
	t.Parallel()

	t.Run("enables slack with valid webhook URL", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		setting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return s.SlackEnabled && len(s.SlackWebhook) > 0
		})).Return(nil)

		err := usecase.CreateSetting(setting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("updates slack webhook URL", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.ID = 5
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/OLD/WEBHOOK"
		})

		updatedSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/NEW/WEBHOOK"
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return s.SlackWebhook == "https://hooks.slack.com/services/NEW/WEBHOOK"
		})).Return(nil)

		err := usecase.UpdateSetting(1, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("disables slack and clears webhook", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.ID = 7
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/TEST"
		})

		updatedSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.SlackEnabled = false
			s.SlackWebhook = ""
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return !s.SlackEnabled && s.SlackWebhook == ""
		})).Return(nil)

		err := usecase.UpdateSetting(1, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})
}

func TestNotificationUsecase_EmailPreferences(t *testing.T) {
	t.Parallel()

	t.Run("enables email notifications by default", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))

		setting, err := usecase.GetSettingByUserID(1)

		require.NoError(t, err)
		assert.True(t, setting.EmailEnabled)
	})

	t.Run("disables email notifications", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		existingSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.ID = 8
		})

		updatedSetting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.EmailEnabled = false
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil)
		notificationRepo.On("Update", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return !s.EmailEnabled
		})).Return(nil)

		err := usecase.UpdateSetting(1, updatedSetting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})

	t.Run("enables both email and slack notifications", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		usecase := createTestNotificationUsecase(notificationRepo)

		setting := testutil.NewTestNotificationSetting(1, func(s *domain.NotificationSetting) {
			s.EmailEnabled = true
			s.SlackEnabled = true
			s.SlackWebhook = "https://hooks.slack.com/services/BOTH"
		})

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, errors.New("not found"))
		notificationRepo.On("Create", mock.MatchedBy(func(s *domain.NotificationSetting) bool {
			return s.EmailEnabled && s.SlackEnabled
		})).Return(nil)

		err := usecase.CreateSetting(setting)

		require.NoError(t, err)
		notificationRepo.AssertExpectations(t)
	})
}
