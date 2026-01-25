package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserUsecase は usecase.UserUsecase のモック実装です
type MockUserUsecase struct {
	mock.Mock
}

func NewMockUserUsecase() *MockUserUsecase {
	return &MockUserUsecase{}
}

func (m *MockUserUsecase) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserUsecase) CreateUser(ctx context.Context, email, password, name string, role domain.Role, employeeNumber, department string) (*domain.User, error) {
	args := m.Called(ctx, email, password, name, role, employeeNumber, department)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUsecase) Update(ctx context.Context, id uint, name, email string, role domain.Role, employeeNumber, department string) (*domain.User, error) {
	args := m.Called(ctx, id, name, email, role, employeeNumber, department)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUsecase) UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	args := m.Called(ctx, id, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserUsecase) AdminResetPassword(ctx context.Context, id uint, newPassword string) error {
	args := m.Called(ctx, id, newPassword)
	return args.Error(0)
}

func (m *MockUserUsecase) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserUsecase) ToggleActive(ctx context.Context, currentUserID uint, id uint, isActive bool) error {
	args := m.Called(ctx, currentUserID, id, isActive)
	return args.Error(0)
}

func TestUserHandler_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("successfully retrieves all users", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		users := []*domain.User{
			testutil.NewTestUser(),
			testutil.NewTestAdmin(),
		}

		userUsecase.On("GetAllUsers", mock.Anything).Return(users, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*domain.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response, 2)
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("GetAllUsers", mock.Anything).
			Return(nil, domain.ErrInternal("Database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		userUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("successfully retrieves user by ID", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 123
			u.Name = "Test User"
		})

		userUsecase.On("GetByID", mock.Anything, uint(123)).Return(user, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Test User", response["name"])
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID format", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "invalid user ID")
		userUsecase.AssertNotCalled(t, "GetByID")
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("GetByID", mock.Anything, uint(999)).
			Return(nil, domain.ErrNotFound("User"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		userUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates user", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.Name = "New User"
			u.Email = "new@example.com"
			u.Role = domain.RoleEditor
		})

		userUsecase.On("CreateUser",
			mock.Anything,
			"new@example.com",
			"Password123!",
			"New User",
			domain.RoleEditor,
			"EMP-001",
			"Engineering",
		).Return(user, nil)

		reqBody := CreateUserRequest{
			Email:          "new@example.com",
			Password:       "Password123!",
			Name:           "New User",
			Role:           domain.RoleEditor,
			EmployeeNumber: "EMP-001",
			Department:     "Engineering",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "name")
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid email", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		reqBody := CreateUserRequest{
			Email:    "invalid-email",
			Password: "Password123!",
			Name:     "New User",
			Role:     domain.RoleEditor,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		userUsecase.AssertNotCalled(t, "CreateUser")
	})

	t.Run("fails with invalid role", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		reqBody := map[string]any{
			"email":    "new@example.com",
			"password": "Password123!",
			"name":     "New User",
			"role":     "invalid_role",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		userUsecase.AssertNotCalled(t, "CreateUser")
	})

	t.Run("fails when email already exists", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("CreateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrConflict("Email already exists"))

		reqBody := CreateUserRequest{
			Email:    "existing@example.com",
			Password: "Password123!",
			Name:     "New User",
			Role:     domain.RoleEditor,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		userUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates user", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		updatedUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 123
			u.Name = "Updated User"
			u.Email = "updated@example.com"
		})

		userUsecase.On("Update",
			mock.Anything,
			uint(123),
			"Updated User",
			"updated@example.com",
			domain.RoleEditor,
			"EMP-001",
			"Engineering",
		).Return(updatedUser, nil)

		reqBody := UpdateUserRequest{
			Name:           "Updated User",
			Email:          "updated@example.com",
			Role:           domain.RoleEditor,
			EmployeeNumber: "EMP-001",
			Department:     "Engineering",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "name")
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		reqBody := UpdateUserRequest{
			Name:  "Updated User",
			Email: "updated@example.com",
			Role:  domain.RoleEditor,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/invalid", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "invalid user ID")
		userUsecase.AssertNotCalled(t, "Update")
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("Update", mock.Anything, uint(999), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrNotFound("User"))

		reqBody := UpdateUserRequest{
			Name:  "Updated User",
			Email: "updated@example.com",
			Role:  domain.RoleEditor,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/999", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		userUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("successfully deletes user", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("Delete", mock.Anything, uint(123)).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "deleted")
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/invalid", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		userUsecase.AssertNotCalled(t, "Delete")
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("Delete", mock.Anything, uint(999)).
			Return(domain.ErrNotFound("User"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		userUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_ToggleActive(t *testing.T) {
	t.Parallel()

	t.Run("successfully toggles user active status", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("ToggleActive", mock.Anything, uint(1), uint(123), false).Return(nil)

		reqBody := ToggleActiveRequest{
			IsActive: false,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/123/toggle-active", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))

		handler.ToggleActive(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "updated")
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing user ID in context", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		reqBody := ToggleActiveRequest{
			IsActive: false,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/123/toggle-active", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		// userID を設定しない

		handler.ToggleActive(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		userUsecase.AssertNotCalled(t, "ToggleActive")
	})

	t.Run("fails when trying to deactivate self", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("ToggleActive", mock.Anything, uint(123), uint(123), false).
			Return(domain.ErrBadRequest("Cannot deactivate your own account"))

		reqBody := ToggleActiveRequest{
			IsActive: false,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/123/toggle-active", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(123))

		handler.ToggleActive(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		userUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_UpdatePassword(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates password", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("UpdatePassword", mock.Anything, uint(123), "OldPassword123!", "NewPassword456!").
			Return(nil)

		reqBody := UpdatePasswordRequest{
			OldPassword: "OldPassword123!",
			NewPassword: "NewPassword456!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/123/password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.UpdatePassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "password updated")
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with incorrect old password", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("UpdatePassword", mock.Anything, uint(123), "WrongPassword!", "NewPassword456!").
			Return(domain.ErrUnauthorized("Incorrect password"))

		reqBody := UpdatePasswordRequest{
			OldPassword: "WrongPassword!",
			NewPassword: "NewPassword456!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/123/password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.UpdatePassword(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with weak new password", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		reqBody := UpdatePasswordRequest{
			OldPassword: "OldPassword123!",
			NewPassword: "weak",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/123/password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.UpdatePassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		userUsecase.AssertNotCalled(t, "UpdatePassword")
	})
}

func TestUserHandler_AdminResetPassword(t *testing.T) {
	t.Parallel()

	t.Run("successfully resets password", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("AdminResetPassword", mock.Anything, uint(123), "NewPassword456!").
			Return(nil)

		reqBody := AdminResetPasswordRequest{
			NewPassword: "NewPassword456!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/123/reset-password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.AdminResetPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "reset")
		userUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		reqBody := AdminResetPasswordRequest{
			NewPassword: "NewPassword456!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/invalid/reset-password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

		handler.AdminResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		userUsecase.AssertNotCalled(t, "AdminResetPassword")
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userUsecase := NewMockUserUsecase()
		handler := NewUserHandler(userUsecase)

		userUsecase.On("AdminResetPassword", mock.Anything, uint(999), "NewPassword456!").
			Return(domain.ErrNotFound("User"))

		reqBody := AdminResetPasswordRequest{
			NewPassword: "NewPassword456!",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/999/reset-password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.AdminResetPassword(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		userUsecase.AssertExpectations(t)
	})
}
