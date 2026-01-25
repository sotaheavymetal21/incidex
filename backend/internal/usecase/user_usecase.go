package usecase

import (
	"context"
	"incidex/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
	CreateUser(ctx context.Context, email, password, name string, role domain.Role, employeeNumber, department string) (*domain.User, error)
	Update(ctx context.Context, id uint, name, email string, role domain.Role, employeeNumber, department string) (*domain.User, error)
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
	// 削除済みユーザーを除外
	activeUsers := make([]*domain.User, 0)
	for _, user := range users {
		if user.DeletedAt == nil {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers, nil
}

func (u *userUsecase) CreateUser(ctx context.Context, email, password, name string, role domain.Role, employeeNumber, department string) (*domain.User, error) {
	// ユーザー入力をバリデーション
	if err := domain.ValidateUserInput(name, email, employeeNumber, department); err != nil {
		return nil, err
	}

	// パスワード強度をバリデーション
	if err := domain.ValidatePasswordStrength(password); err != nil {
		return nil, domain.ErrValidation(err.Error())
	}

	// メールアドレスが既に存在するかチェック
	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrConflict("Email already exists")
	}

	// パスワードをhash化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// ユーザーを作成
	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         role,
		IsActive:     true,
	}
	// 空でない場合のみオプションフィールドを設定
	if employeeNumber != "" {
		user.EmployeeNumber = &employeeNumber
	}
	if department != "" {
		user.Department = &department
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) Update(ctx context.Context, id uint, name, email string, role domain.Role, employeeNumber, department string) (*domain.User, error) {
	// ユーザー入力をバリデーション
	if err := domain.ValidateUserInput(name, email, employeeNumber, department); err != nil {
		return nil, err
	}

	// 既存のユーザーを検索
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

	// メールアドレスが既に存在するものに変更されていないかチェック
	if user.Email != email {
		existingUser, err := u.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, domain.ErrConflict("Email already exists")
		}
	}

	// ユーザーフィールドを更新
	user.Name = name
	user.Email = email
	user.Role = role
	// 空でない場合のみオプションフィールドを設定、それ以外はnil
	if employeeNumber != "" {
		user.EmployeeNumber = &employeeNumber
	} else {
		user.EmployeeNumber = nil
	}
	if department != "" {
		user.Department = &department
	} else {
		user.Department = nil
	}

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

	// 古いパスワードを検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrUnauthorized("Invalid old password")
	}

	// 新しいパスワード強度をバリデーション
	if err := domain.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 新しいパスワードをhash化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 一意制約違反を避けるため、パスワードフィールドのみを更新
	return u.userRepo.UpdatePassword(ctx, id, string(hashedPassword))
}

func (u *userUsecase) AdminResetPassword(ctx context.Context, id uint, newPassword string) error {
	// パスワード強度をバリデーション
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

	// 新しいパスワードをhash化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 一意制約違反を避けるため、パスワードフィールドのみを更新
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
	// ユーザーが自分自身を無効化するのを防止
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

	// 最後のアクティブな管理者を無効化するのを防止
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
