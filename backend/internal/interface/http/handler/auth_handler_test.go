package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	// テスト環境では Gin をテストモードに設定
	gin.SetMode(gin.TestMode)
}

// MockAuthUsecase は usecase.AuthUsecase のモック実装です
type MockAuthUsecase struct {
	mock.Mock
}

func NewMockAuthUsecase() *MockAuthUsecase {
	return &MockAuthUsecase{}
}

func (m *MockAuthUsecase) Register(ctx context.Context, name, email, password, employeeNumber, department string) (*domain.User, error) {
	args := m.Called(ctx, name, email, password, employeeNumber, department)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthUsecase) Login(ctx context.Context, email, password string) (*usecase.AuthResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.AuthResponse), args.Error(1)
}

func (m *MockAuthUsecase) RefreshAccessToken(ctx context.Context, refreshToken string) (*usecase.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.AuthResponse), args.Error(1)
}

func (m *MockAuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthUsecase) GetUserFromToken(ctx context.Context, token string) (*domain.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	t.Parallel()

	t.Run("successfully registers new user", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.Name = "John Doe"
			u.Email = "john@example.com"
		})

		authUsecase.On("Register", mock.Anything, "John Doe", "john@example.com", "StrongPass123!", "EMP-001", "Engineering").
			Return(user, nil)

		reqBody := RegisterRequest{
			Name:           "John Doe",
			Email:          "john@example.com",
			Password:       "StrongPass123!",
			EmployeeNumber: "EMP-001",
			Department:     "Engineering",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "user")
		authUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid email", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		reqBody := RegisterRequest{
			Name:           "John Doe",
			Email:          "invalid-email",
			Password:       "StrongPass123!",
			EmployeeNumber: "EMP-001",
			Department:     "Engineering",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		authUsecase.AssertNotCalled(t, "Register")
	})

	t.Run("fails with weak password", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		reqBody := RegisterRequest{
			Name:           "John Doe",
			Email:          "john@example.com",
			Password:       "weak",
			EmployeeNumber: "EMP-001",
			Department:     "Engineering",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		authUsecase.AssertNotCalled(t, "Register")
	})

	t.Run("fails with missing required fields", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		reqBody := map[string]string{
			"email":    "john@example.com",
			"password": "StrongPass123!",
			// name, employee_number, department が欠けている
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		authUsecase.AssertNotCalled(t, "Register")
	})

	t.Run("fails when email already exists", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		authUsecase.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrConflict("Email already exists"))

		reqBody := RegisterRequest{
			Name:           "John Doe",
			Email:          "existing@example.com",
			Password:       "StrongPass123!",
			EmployeeNumber: "EMP-001",
			Department:     "Engineering",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		authUsecase.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	t.Run("successfully logs in and sets cookie", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false) // isProduction = false

		user := testutil.NewTestUser(func(u *domain.User) {
			u.Email = "user@example.com"
		})

		authResponse := &usecase.AuthResponse{
			User:         user,
			AccessToken:  "access-token-12345",
			RefreshToken: "refresh-token-67890",
		}

		authUsecase.On("Login", mock.Anything, "user@example.com", "Password123!").
			Return(authResponse, nil)

		reqBody := LoginRequest{
			Email:    "user@example.com",
			Password: "Password123!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "user")
		assert.Equal(t, "access-token-12345", response["access_token"])

		// Cookie が設定されているか確認
		cookies := w.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
				break
			}
		}

		require.NotNil(t, refreshCookie)
		assert.Equal(t, "refresh-token-67890", refreshCookie.Value)
		assert.True(t, refreshCookie.HttpOnly)
		assert.Equal(t, "/", refreshCookie.Path)
		assert.False(t, refreshCookie.Secure) // isProduction = false

		authUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid credentials", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		authUsecase.On("Login", mock.Anything, "user@example.com", "WrongPassword").
			Return(nil, domain.ErrUnauthorized("Invalid email or password"))

		reqBody := LoginRequest{
			Email:    "user@example.com",
			Password: "WrongPassword",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		authUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing email", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		reqBody := map[string]string{
			"password": "Password123!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		authUsecase.AssertNotCalled(t, "Login")
	})

	t.Run("sets secure cookie in production mode", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, true) // isProduction = true

		user := testutil.NewTestUser()
		authResponse := &usecase.AuthResponse{
			User:         user,
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}

		authUsecase.On("Login", mock.Anything, mock.Anything, mock.Anything).
			Return(authResponse, nil)

		reqBody := LoginRequest{
			Email:    "user@example.com",
			Password: "Password123!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
				break
			}
		}

		require.NotNil(t, refreshCookie)
		assert.True(t, refreshCookie.Secure) // isProduction = true
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Parallel()

	t.Run("successfully refreshes token from cookie", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		user := testutil.NewTestUser()
		authResponse := &usecase.AuthResponse{
			User:         user,
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
		}

		authUsecase.On("RefreshAccessToken", mock.Anything, "old-refresh-token").
			Return(authResponse, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "old-refresh-token",
		})

		handler.Refresh(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Equal(t, "new-access-token", response["access_token"])

		// 新しい refresh token が cookie に設定されていることを確認
		cookies := w.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
				break
			}
		}

		require.NotNil(t, refreshCookie)
		assert.Equal(t, "new-refresh-token", refreshCookie.Value)

		authUsecase.AssertExpectations(t)
	})

	t.Run("successfully refreshes token from request body", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		user := testutil.NewTestUser()
		authResponse := &usecase.AuthResponse{
			User:         user,
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
		}

		authUsecase.On("RefreshAccessToken", mock.Anything, "body-refresh-token").
			Return(authResponse, nil)

		reqBody := RefreshRequest{
			RefreshToken: "body-refresh-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Refresh(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		authUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing refresh token", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)

		handler.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "Refresh token required")
		authUsecase.AssertNotCalled(t, "RefreshAccessToken")
	})

	t.Run("fails with invalid refresh token", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		authUsecase.On("RefreshAccessToken", mock.Anything, "invalid-token").
			Return(nil, domain.ErrUnauthorized("Invalid refresh token"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "invalid-token",
		})

		handler.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		authUsecase.AssertExpectations(t)
	})

	t.Run("prefers cookie over request body", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		user := testutil.NewTestUser()
		authResponse := &usecase.AuthResponse{
			User:         user,
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
		}

		// cookie からのトークンが使用されるべき
		authUsecase.On("RefreshAccessToken", mock.Anything, "cookie-token").
			Return(authResponse, nil)

		reqBody := RefreshRequest{
			RefreshToken: "body-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "cookie-token",
		})

		handler.Refresh(c)

		assert.Equal(t, http.StatusOK, w.Code)
		authUsecase.AssertExpectations(t)
		authUsecase.AssertNotCalled(t, "RefreshAccessToken", mock.Anything, "body-token")
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	t.Run("successfully logs out and clears cookie", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		authUsecase.On("Logout", mock.Anything, "refresh-token-to-revoke").
			Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "refresh-token-to-revoke",
		})

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")

		// Cookie がクリアされているか確認（MaxAge = -1）
		cookies := w.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
				break
			}
		}

		require.NotNil(t, refreshCookie)
		assert.Equal(t, "", refreshCookie.Value)
		assert.Equal(t, -1, refreshCookie.MaxAge)

		authUsecase.AssertExpectations(t)
	})

	t.Run("succeeds even without refresh token cookie", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")

		// Logout usecase は呼ばれないはず
		authUsecase.AssertNotCalled(t, "Logout")
	})

	t.Run("succeeds even when usecase logout fails", func(t *testing.T) {
		t.Parallel()

		authUsecase := NewMockAuthUsecase()
		handler := NewAuthHandler(authUsecase, false)

		authUsecase.On("Logout", mock.Anything, "refresh-token").
			Return(errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "refresh-token",
		})

		handler.Logout(c)

		// usecase がエラーを返してもログアウトは成功する
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")

		authUsecase.AssertExpectations(t)
	})
}
