package usecase

import (
	"context"
	"errors"
	"incidex/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
	CreateUser(ctx context.Context, email, password, name string, role domain.Role) (*domain.User, error)
	Update(ctx context.Context, id uint, name, email string, role domain.Role) (*domain.User, error)
	UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error
	AdminResetPassword(ctx context.Context, id uint, newPassword string) error
	Delete(ctx context.Context, id uint) error
	ToggleActive(ctx context.Context, currentUserID uint, id uint, isActive bool) error
}

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return nil, errors.New("user has been deleted")
	}
	return user, nil
}

func (u *userUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := u.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	// Filter out deleted users
	activeUsers := make([]*domain.User, 0)
	for _, user := range users {
		if user.DeletedAt == nil {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers, nil
}

func (u *userUsecase) CreateUser(ctx context.Context, email, password, name string, role domain.Role) (*domain.User, error) {
	// Validate password length
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if email already exists
	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         role,
		IsActive:     true,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) Update(ctx context.Context, id uint, name, email string, role domain.Role) (*domain.User, error) {
	// Find existing user
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return nil, errors.New("cannot update deleted user")
	}

	// Check if email is being changed to one that already exists
	if user.Email != email {
		existingUser, err := u.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already exists")
		}
	}

	// Update user fields
	user.Name = name
	user.Email = email
	user.Role = role

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return errors.New("cannot update password for deleted user")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update only the password field to avoid unique constraint violations
	return u.userRepo.UpdatePassword(ctx, id, string(hashedPassword))
}

func (u *userUsecase) AdminResetPassword(ctx context.Context, id uint, newPassword string) error {
	// Validate password length
	if len(newPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return errors.New("cannot reset password for deleted user")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update only the password field to avoid unique constraint violations
	return u.userRepo.UpdatePassword(ctx, id, string(hashedPassword))
}

func (u *userUsecase) Delete(ctx context.Context, id uint) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return errors.New("user already deleted")
	}

	return u.userRepo.Delete(ctx, id)
}

func (u *userUsecase) ToggleActive(ctx context.Context, currentUserID uint, id uint, isActive bool) error {
	// Prevent users from deactivating themselves
	if currentUserID == id && !isActive {
		return errors.New("cannot deactivate your own account")
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return errors.New("cannot toggle active status of deleted user")
	}

	// Prevent deactivating the last active admin
	if !isActive && user.Role == domain.RoleAdmin && user.IsActive {
		activeAdmins, err := u.userRepo.FindAll(ctx)
		if err != nil {
			return err
		}
		activeAdminCount := 0
		for _, admin := range activeAdmins {
			if admin.Role == domain.RoleAdmin && admin.IsActive && admin.DeletedAt == nil {
				activeAdminCount++
			}
		}
		if activeAdminCount <= 1 {
			return errors.New("cannot deactivate the last active admin user")
		}
	}

	return u.userRepo.ToggleActive(ctx, id, isActive)
}
