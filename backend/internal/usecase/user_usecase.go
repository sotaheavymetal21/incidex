package usecase

import (
	"context"
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
		return nil, domain.ErrNotFound("user")
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrNotFound("user")
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
	// Validate password strength
	if err := domain.ValidatePasswordStrength(password); err != nil {
		return nil, domain.ErrValidation(err.Error())
	}

	// Check if email already exists
	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrConflict("Email already exists")
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
		return nil, domain.ErrNotFound("user")
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrValidation("Cannot update deleted user")
	}

	// Check if email is being changed to one that already exists
	if user.Email != email {
		existingUser, err := u.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, domain.ErrConflict("Email already exists")
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
		return domain.ErrNotFound("user")
	}
	if user.DeletedAt != nil {
		return domain.ErrValidation("Cannot update password for deleted user")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrUnauthorized("Invalid old password")
	}

	// Validate new password strength
	if err := domain.ValidatePasswordStrength(newPassword); err != nil {
		return err
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
	// Validate password strength
	if err := domain.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrNotFound("user")
	}
	if user.DeletedAt != nil {
		return domain.ErrValidation("Cannot reset password for deleted user")
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
		return domain.ErrNotFound("user")
	}
	if user.DeletedAt != nil {
		return domain.ErrValidation("User already deleted")
	}

	return u.userRepo.Delete(ctx, id)
}

func (u *userUsecase) ToggleActive(ctx context.Context, currentUserID uint, id uint, isActive bool) error {
	// Prevent users from deactivating themselves
	if currentUserID == id && !isActive {
		return domain.ErrValidation("Cannot deactivate your own account")
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrNotFound("user")
	}
	if user.DeletedAt != nil {
		return domain.ErrValidation("Cannot toggle active status of deleted user")
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
			return domain.ErrValidation("Cannot deactivate the last active admin user")
		}
	}

	return u.userRepo.ToggleActive(ctx, id, isActive)
}
