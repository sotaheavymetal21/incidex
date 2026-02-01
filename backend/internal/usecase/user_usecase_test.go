package usecase

import (
	"context"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func createTestUserUsecase(userRepo *mocks.MockUserRepository) UserUsecase {
	return NewUserUsecase(userRepo)
}

func TestUserUsecase_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns user when found", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		expectedUser := testutil.NewTestUser()
		userRepo.On("FindByID", ctx, uint(1)).Return(expectedUser, nil)

		user, err := usecase.GetByID(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)

		userRepo.AssertExpectations(t)
	})

	t.Run("returns not found when user does not exist", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		user, err := usecase.GetByID(ctx, 999)

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("returns not found for deleted user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		deletedAt := time.Now()
		deletedUser := testutil.NewTestUser(func(u *domain.User) {
			u.DeletedAt = &deletedAt
		})
		userRepo.On("FindByID", ctx, uint(1)).Return(deletedUser, nil)

		user, err := usecase.GetByID(ctx, 1)

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("returns error when database fails", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(1)).Return(nil, domain.ErrDatabase("db error", nil))

		user, err := usecase.GetByID(ctx, 1)

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

func TestUserUsecase_GetAllUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns all active users", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		deletedAt := time.Now()
		users := []*domain.User{
			testutil.NewTestUser(func(u *domain.User) { u.ID = 1 }),
			testutil.NewTestUser(func(u *domain.User) { u.ID = 2 }),
			testutil.NewTestUser(func(u *domain.User) { u.ID = 3; u.DeletedAt = &deletedAt }), // Deleted
		}
		userRepo.On("FindAll", ctx).Return(users, nil)

		result, err := usecase.GetAllUsers(ctx)

		require.NoError(t, err)
		assert.Len(t, result, 2) // Only active users
		assert.Equal(t, uint(1), result[0].ID)
		assert.Equal(t, uint(2), result[1].ID)
	})

	t.Run("returns error when database fails", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindAll", ctx).Return(nil, domain.ErrDatabase("db error", nil))

		result, err := usecase.GetAllUsers(ctx)

		require.Error(t, err)
		assert.Nil(t, result)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

func TestUserUsecase_CreateUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := usecase.CreateUser(ctx, "newuser@example.com", "StrongPass123!", "New User", domain.RoleViewer, "EMP-001", "Engineering")

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "New User", user.Name)
		assert.Equal(t, "newuser@example.com", user.Email)
		assert.Equal(t, domain.RoleViewer, user.Role)
		assert.True(t, user.IsActive)

		userRepo.AssertExpectations(t)
	})

	t.Run("fails when email already exists", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser()
		userRepo.On("FindByEmail", ctx, "existing@example.com").Return(existingUser, nil)

		user, err := usecase.CreateUser(ctx, "existing@example.com", "StrongPass123!", "New User", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeConflict, domainErr.Code)
	})

	t.Run("fails with weak password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, nil)

		user, err := usecase.CreateUser(ctx, "newuser@example.com", "weak", "New User", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("fails with invalid email", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		user, err := usecase.CreateUser(ctx, "invalid-email", "StrongPass123!", "New User", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		user, err := usecase.CreateUser(ctx, "new@example.com", "StrongPass123!", "", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails when database returns error on create", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(domain.ErrDatabase("db error", nil))

		user, err := usecase.CreateUser(ctx, "new@example.com", "StrongPass123!", "New User", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

func TestUserUsecase_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser(func(u *domain.User) { u.ID = 1 })
		userRepo.On("FindByID", ctx, uint(1)).Return(existingUser, nil)
		userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := usecase.Update(ctx, 1, "Updated Name", "test@example.com", domain.RoleEditor, "", "")

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", user.Name)
		assert.Equal(t, domain.RoleEditor, user.Role)
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		user, err := usecase.Update(ctx, 999, "Updated Name", "test@example.com", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails when email already taken by another user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 1
			u.Email = "old@example.com"
		})
		otherUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 2
			u.Email = "taken@example.com"
		})

		userRepo.On("FindByID", ctx, uint(1)).Return(existingUser, nil)
		userRepo.On("FindByEmail", ctx, "taken@example.com").Return(otherUser, nil)

		user, err := usecase.Update(ctx, 1, "Updated Name", "taken@example.com", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeConflict, domainErr.Code)
	})

	t.Run("fails with invalid email", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser(func(u *domain.User) { u.ID = 1 })
		userRepo.On("FindByID", ctx, uint(1)).Return(existingUser, nil)

		user, err := usecase.Update(ctx, 1, "Updated Name", "invalid-email", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser(func(u *domain.User) { u.ID = 1 })
		userRepo.On("FindByID", ctx, uint(1)).Return(existingUser, nil)

		user, err := usecase.Update(ctx, 1, "", "test@example.com", domain.RoleViewer, "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("allows same email for same user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 1
			u.Email = "same@example.com"
		})

		userRepo.On("FindByID", ctx, uint(1)).Return(existingUser, nil)
		userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := usecase.Update(ctx, 1, "Updated Name", "same@example.com", domain.RoleEditor, "", "")

		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("allows email change when not taken", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		existingUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 1
			u.Email = "old@example.com"
		})

		userRepo.On("FindByID", ctx, uint(1)).Return(existingUser, nil)
		userRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, nil) // Not taken
		userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := usecase.Update(ctx, 1, "Updated Name", "new@example.com", domain.RoleEditor, "", "")

		require.NoError(t, err)
		assert.NotNil(t, user)
	})
}

func TestUserUsecase_UpdatePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful password update", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("OldPassword123!"), bcrypt.DefaultCost)
		user := testutil.NewTestUser(func(u *domain.User) {
			u.PasswordHash = string(hashedPassword)
		})

		userRepo.On("FindByID", ctx, uint(1)).Return(user, nil)
		userRepo.On("UpdatePassword", ctx, uint(1), mock.AnythingOfType("string")).Return(nil)

		err := usecase.UpdatePassword(ctx, 1, "OldPassword123!", "NewStrongPass123!")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("fails with incorrect old password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), bcrypt.DefaultCost)
		user := testutil.NewTestUser(func(u *domain.User) {
			u.PasswordHash = string(hashedPassword)
		})

		userRepo.On("FindByID", ctx, uint(1)).Return(user, nil)

		err := usecase.UpdatePassword(ctx, 1, "WrongPassword123!", "NewStrongPass123!")

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeUnauthorized, domainErr.Code)
	})

	t.Run("fails with weak new password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("OldPassword123!"), bcrypt.DefaultCost)
		user := testutil.NewTestUser(func(u *domain.User) {
			u.PasswordHash = string(hashedPassword)
		})

		userRepo.On("FindByID", ctx, uint(1)).Return(user, nil)

		err := usecase.UpdatePassword(ctx, 1, "OldPassword123!", "weak")

		require.Error(t, err)
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		err := usecase.UpdatePassword(ctx, 999, "OldPassword123!", "NewStrongPass123!")

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails when database returns error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(1)).Return(nil, domain.ErrDatabase("db error", nil))

		err := usecase.UpdatePassword(ctx, 1, "OldPassword123!", "NewStrongPass123!")

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

func TestUserUsecase_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		user := testutil.NewTestUser()
		userRepo.On("FindByID", ctx, uint(1)).Return(user, nil)
		userRepo.On("Delete", ctx, uint(1)).Return(nil)

		err := usecase.Delete(ctx, 1)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		err := usecase.Delete(ctx, 999)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails when user already deleted", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		deletedAt := time.Now()
		user := testutil.NewTestUser(func(u *domain.User) {
			u.DeletedAt = &deletedAt
		})
		userRepo.On("FindByID", ctx, uint(1)).Return(user, nil)

		err := usecase.Delete(ctx, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})
}

func TestUserUsecase_ToggleActive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful activation", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 2
			u.IsActive = false
		})
		userRepo.On("FindByID", ctx, uint(2)).Return(user, nil)
		userRepo.On("ToggleActive", ctx, uint(2), true).Return(nil)

		err := usecase.ToggleActive(ctx, 1, 2, true)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("cannot deactivate own account", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		err := usecase.ToggleActive(ctx, 1, 1, false) // Same user ID

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
		assert.Contains(t, domainErr.Message, "own account")
	})

	t.Run("cannot deactivate last admin", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		adminUser := testutil.NewTestAdmin(func(u *domain.User) {
			u.ID = 2
		})
		allUsers := []*domain.User{adminUser} // Only one admin

		userRepo.On("FindByID", ctx, uint(2)).Return(adminUser, nil)
		userRepo.On("FindAll", ctx).Return(allUsers, nil)

		err := usecase.ToggleActive(ctx, 1, 2, false)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
		assert.Contains(t, domainErr.Message, "last active admin")
	})

	t.Run("can deactivate admin when another admin exists", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		adminToDeactivate := testutil.NewTestAdmin(func(u *domain.User) {
			u.ID = 2
		})
		anotherAdmin := testutil.NewTestAdmin(func(u *domain.User) {
			u.ID = 3
		})
		allUsers := []*domain.User{adminToDeactivate, anotherAdmin}

		userRepo.On("FindByID", ctx, uint(2)).Return(adminToDeactivate, nil)
		userRepo.On("FindAll", ctx).Return(allUsers, nil)
		userRepo.On("ToggleActive", ctx, uint(2), false).Return(nil)

		err := usecase.ToggleActive(ctx, 1, 2, false)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		err := usecase.ToggleActive(ctx, 1, 999, true)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})
}

func TestUserUsecase_AdminResetPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully resets password for active user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 5
			u.DeletedAt = nil
		})

		userRepo.On("FindByID", ctx, uint(5)).Return(user, nil)
		userRepo.On("UpdatePassword", ctx, uint(5), mock.AnythingOfType("string")).Return(nil)

		err := usecase.AdminResetPassword(ctx, 5, "NewStrongPass123!")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("fails with weak password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		err := usecase.AdminResetPassword(ctx, 5, "weak")

		require.Error(t, err)
		userRepo.AssertNotCalled(t, "FindByID")
		userRepo.AssertNotCalled(t, "UpdatePassword")
	})

	t.Run("fails when user not found", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		err := usecase.AdminResetPassword(ctx, 999, "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails for deleted user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		usecase := createTestUserUsecase(userRepo)

		deletedAt := time.Now()
		user := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 5
			u.DeletedAt = &deletedAt
		})

		userRepo.On("FindByID", ctx, uint(5)).Return(user, nil)

		err := usecase.AdminResetPassword(ctx, 5, "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)

		userRepo.AssertNotCalled(t, "UpdatePassword")
	})
}
