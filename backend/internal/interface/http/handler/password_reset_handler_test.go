package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPasswordResetUsecase は usecase.PasswordResetUsecase のモック実装です
type MockPasswordResetUsecase struct {
	mock.Mock
}

func NewMockPasswordResetUsecase() *MockPasswordResetUsecase {
	return &MockPasswordResetUsecase{}
}

func (m *MockPasswordResetUsecase) RequestPasswordReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockPasswordResetUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

func (m *MockPasswordResetUsecase) ValidateToken(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func TestPasswordResetHandler_RequestPasswordReset(t *testing.T) {
	t.Parallel()

	t.Run("successfully requests password reset", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		mockUsecase.On("RequestPasswordReset", mock.Anything, "user@example.com").Return(nil)

		reqBody := RequestPasswordResetRequest{
			Email: "user@example.com",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset-request", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RequestPasswordReset(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "パスワードリセット")
	})

	t.Run("returns success even when email not found (prevents email enumeration)", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		// Even with a NotFound error, we should return 200 OK to prevent email enumeration
		mockUsecase.On("RequestPasswordReset", mock.Anything, "nonexistent@example.com").
			Return(nil) // The usecase handles the error silently

		reqBody := RequestPasswordResetRequest{
			Email: "nonexistent@example.com",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset-request", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RequestPasswordReset(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("fails with invalid email format", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		reqBody := RequestPasswordResetRequest{
			Email: "invalid-email",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset-request", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RequestPasswordReset(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUsecase.AssertNotCalled(t, "RequestPasswordReset")
	})

	t.Run("fails with missing email", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		reqBody := map[string]string{}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset-request", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RequestPasswordReset(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPasswordResetHandler_ResetPassword(t *testing.T) {
	t.Parallel()

	t.Run("successfully resets password", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		// Password must be at least 12 characters with various requirements
		mockUsecase.On("ResetPassword", mock.Anything, "valid-token", "NewPassword123!").Return(nil)

		reqBody := ResetPasswordRequest{
			Token:       "valid-token",
			NewPassword: "NewPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ResetPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "パスワードが正常にリセット")
	})

	t.Run("fails with invalid token", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		mockUsecase.On("ResetPassword", mock.Anything, "invalid-token", "NewPassword123!").
			Return(domain.ErrValidation("invalid or expired token"))

		reqBody := ResetPasswordRequest{
			Token:       "invalid-token",
			NewPassword: "NewPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails with missing token", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		reqBody := map[string]string{
			"new_password": "NewPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails with short password", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		reqBody := ResetPasswordRequest{
			Token:       "valid-token",
			NewPassword: "short",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/password/reset", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPasswordResetHandler_ValidateToken(t *testing.T) {
	t.Parallel()

	t.Run("returns valid true for valid token", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		mockUsecase.On("ValidateToken", mock.Anything, "valid-token").Return(true, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/password/validate-token?token=valid-token", nil)

		handler.ValidateToken(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["valid"].(bool))
	})

	t.Run("returns valid false for invalid token", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		mockUsecase.On("ValidateToken", mock.Anything, "invalid-token").Return(false, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/password/validate-token?token=invalid-token", nil)

		handler.ValidateToken(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response["valid"].(bool))
	})

	t.Run("fails with missing token", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockPasswordResetUsecase()
		handler := NewPasswordResetHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/password/validate-token", nil)

		handler.ValidateToken(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
