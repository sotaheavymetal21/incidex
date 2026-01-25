package domain

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type User struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Email          string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string     `gorm:"not null" json:"-"`
	Name           string     `gorm:"not null" json:"name"`
	EmployeeNumber *string    `gorm:"uniqueIndex" json:"employee_number,omitempty"`
	Department     *string    `json:"department,omitempty"`
	Role           Role       `gorm:"not null;default:'viewer'" json:"role"`
	IsActive       bool       `gorm:"default:true;not null" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, id uint, passwordHash string) error
	Delete(ctx context.Context, id uint) error
	ToggleActive(ctx context.Context, id uint, isActive bool) error
}

// ユーザーバリデーション定数
const (
	MaxNameLength           = 50
	MaxEmailLength          = 254
	MaxEmployeeNumberLength = 20
	MaxDepartmentLength     = 50
)

// バリデーション用の正規表現
var (
	employeeNumberRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]*$`)
	emailRegex          = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// containsEmoji は文字列に絵文字が含まれているかを確認します
func containsEmoji(s string) bool {
	for _, r := range s {
		if r > 0x1F000 || // ほとんどの絵文字はこの範囲
			(r >= 0x2600 && r <= 0x27BF) || // その他のシンボル
			(r >= 0x1F300 && r <= 0x1F9FF) { // 各種絵文字ブロック
			return true
		}
	}
	return false
}

// containsControlChars は文字列に制御文字が含まれているかを確認します
func containsControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return true
		}
	}
	return false
}

// ValidateName はユーザー名をバリデーションします
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrValidation("名前は必須です")
	}
	if len(name) > MaxNameLength {
		return ErrValidation("名前は50文字以内で入力してください")
	}
	if containsEmoji(name) {
		return ErrValidation("名前に絵文字は使用できません")
	}
	if containsControlChars(name) {
		return ErrValidation("名前に不正な文字が含まれています")
	}
	return nil
}

// ValidateEmail はメールアドレスをバリデーションします
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrValidation("メールアドレスは必須です")
	}
	if len(email) > MaxEmailLength {
		return ErrValidation("メールアドレスは254文字以内で入力してください")
	}
	if !emailRegex.MatchString(email) {
		return ErrValidation("有効なメールアドレスを入力してください")
	}
	return nil
}

// ValidateEmployeeNumber は社員番号をバリデーションします
func ValidateEmployeeNumber(employeeNumber string) error {
	if employeeNumber == "" {
		return nil // オプションフィールド
	}
	if len(employeeNumber) > MaxEmployeeNumberLength {
		return ErrValidation("社員番号は20文字以内で入力してください")
	}
	if !employeeNumberRegex.MatchString(employeeNumber) {
		return ErrValidation("社員番号は英数字とハイフンのみ使用できます")
	}
	return nil
}

// ValidateDepartment は部署名をバリデーションします
func ValidateDepartment(department string) error {
	if department == "" {
		return nil // オプションフィールド
	}
	if len(department) > MaxDepartmentLength {
		return ErrValidation("所属部署は50文字以内で入力してください")
	}
	if containsEmoji(department) {
		return ErrValidation("所属部署に絵文字は使用できません")
	}
	if containsControlChars(department) {
		return ErrValidation("所属部署に不正な文字が含まれています")
	}
	return nil
}

// ValidateUserInput はすべてのユーザー入力フィールドをバリデーションします
func ValidateUserInput(name, email, employeeNumber, department string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if err := ValidateEmployeeNumber(employeeNumber); err != nil {
		return err
	}
	if err := ValidateDepartment(department); err != nil {
		return err
	}
	return nil
}
