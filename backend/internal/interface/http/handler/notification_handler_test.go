package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil/mocks"
	"incidex/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotificationHandler_GetMyNotificationSetting(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns notification setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		setting := &domain.NotificationSetting{
			ID:                      1,
			UserID:                  1,
			EmailEnabled:            true,
			SlackEnabled:            false,
			NotifyOnIncidentCreated: true,
			NotifyOnAssigned:        true,
		}

		notificationRepo.On("GetByUserID", uint(1)).Return(setting, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings", nil)
		c.Set("userID", uint(1))

		handler.GetMyNotificationSetting(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.NotificationSetting
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, uint(1), response.UserID)
		assert.True(t, response.EmailEnabled)
	})

	t.Run("returns default settings when not found", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, domain.ErrNotFound("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings", nil)
		c.Set("userID", uint(1))

		handler.GetMyNotificationSetting(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.NotificationSetting
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, uint(1), response.UserID)
		// Default settings should have email enabled
		assert.True(t, response.EmailEnabled)
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings", nil)
		// No userID set

		handler.GetMyNotificationSetting(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		notificationRepo.AssertNotCalled(t, "GetByUserID")
	})
}

func TestNotificationHandler_UpdateMyNotificationSetting(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates notification setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		existingSetting := &domain.NotificationSetting{
			ID:           1,
			UserID:       1,
			EmailEnabled: true,
		}

		updatedSetting := &domain.NotificationSetting{
			ID:           1,
			UserID:       1,
			EmailEnabled: false,
			SlackEnabled: true,
		}

		notificationRepo.On("GetByUserID", uint(1)).Return(existingSetting, nil).Once()
		notificationRepo.On("Update", mock.Anything).Return(nil)
		notificationRepo.On("GetByUserID", uint(1)).Return(updatedSetting, nil).Once()

		reqBody := domain.NotificationSetting{
			EmailEnabled: false,
			SlackEnabled: true,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/notifications/settings", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.UpdateMyNotificationSetting(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("creates new setting if not exists", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		newSetting := &domain.NotificationSetting{
			ID:           1,
			UserID:       1,
			EmailEnabled: true,
		}

		notificationRepo.On("GetByUserID", uint(1)).Return(nil, domain.ErrNotFound("not found")).Once()
		notificationRepo.On("Create", mock.Anything).Return(nil)
		notificationRepo.On("GetByUserID", uint(1)).Return(newSetting, nil).Once()

		reqBody := domain.NotificationSetting{
			EmailEnabled: true,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/notifications/settings", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.UpdateMyNotificationSetting(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		reqBody := domain.NotificationSetting{EmailEnabled: true}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/notifications/settings", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		// No userID set

		handler.UpdateMyNotificationSetting(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails with invalid request body", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/notifications/settings", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.UpdateMyNotificationSetting(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNotificationHandler_GetUserNotificationSetting(t *testing.T) {
	t.Parallel()

	t.Run("admin can get user notification setting", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		setting := &domain.NotificationSetting{
			ID:           1,
			UserID:       5,
			EmailEnabled: true,
		}

		notificationRepo.On("GetByUserID", uint(5)).Return(setting, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings/5", nil)
		c.Params = gin.Params{{Key: "id", Value: "5"}}
		c.Set("role", domain.RoleAdmin)

		handler.GetUserNotificationSetting(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.NotificationSetting
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, uint(5), response.UserID)
	})

	t.Run("editor cannot access user settings", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings/5", nil)
		c.Params = gin.Params{{Key: "id", Value: "5"}}
		c.Set("role", domain.RoleEditor)

		handler.GetUserNotificationSetting(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		notificationRepo.AssertNotCalled(t, "GetByUserID")
	})

	t.Run("fails without role", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings/5", nil)
		c.Params = gin.Params{{Key: "id", Value: "5"}}
		// No role set

		handler.GetUserNotificationSetting(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("fails with invalid user ID", func(t *testing.T) {
		t.Parallel()

		notificationRepo := mocks.NewMockNotificationSettingRepository()
		notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
		handler := NewNotificationHandler(notificationUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/notifications/settings/invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}
		c.Set("role", domain.RoleAdmin)

		handler.GetUserNotificationSetting(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
