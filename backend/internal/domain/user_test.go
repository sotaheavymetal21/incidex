package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid name",
			input:   "田中太郎",
			wantErr: false,
		},
		{
			name:    "valid English name",
			input:   "John Doe",
			wantErr: false,
		},
		{
			name:    "valid name with spaces",
			input:   "  田中 太郎  ",
			wantErr: false,
		},
		{
			name:        "empty name",
			input:       "",
			wantErr:     true,
			errContains: "必須",
		},
		{
			name:        "whitespace only name",
			input:       "   ",
			wantErr:     true,
			errContains: "必須",
		},
		{
			name:        "name too long - 51 ASCII characters",
			input:       strings.Repeat("a", 51),
			wantErr:     true,
			errContains: "50文字以内",
		},
		{
			name:    "name exactly 50 ASCII characters",
			input:   strings.Repeat("a", 50),
			wantErr: false,
		},
		{
			name:    "Japanese name within limit",
			input:   "田中太郎",
			wantErr: false,
		},
		{
			name:        "name with emoji",
			input:       "Test😀User",
			wantErr:     true,
			errContains: "絵文字",
		},
		{
			name:        "name with control character",
			input:       "Test\x00User",
			wantErr:     true,
			errContains: "不正な文字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid email",
			input:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			input:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			input:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			input:   "user.name@example.com",
			wantErr: false,
		},
		{
			name:        "empty email",
			input:       "",
			wantErr:     true,
			errContains: "必須",
		},
		{
			name:        "invalid email - no @",
			input:       "userexample.com",
			wantErr:     true,
			errContains: "有効なメールアドレス",
		},
		{
			name:        "invalid email - no domain",
			input:       "user@",
			wantErr:     true,
			errContains: "有効なメールアドレス",
		},
		{
			name:        "invalid email - no TLD",
			input:       "user@example",
			wantErr:     true,
			errContains: "有効なメールアドレス",
		},
		{
			name:        "invalid email - spaces",
			input:       "user @example.com",
			wantErr:     true,
			errContains: "有効なメールアドレス",
		},
		{
			name:        "email too long",
			input:       strings.Repeat("a", 250) + "@example.com",
			wantErr:     true,
			errContains: "254文字以内",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateEmail(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmployeeNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid employee number",
			input:   "EMP-001",
			wantErr: false,
		},
		{
			name:    "valid numeric employee number",
			input:   "12345",
			wantErr: false,
		},
		{
			name:    "valid alphanumeric",
			input:   "ABC123",
			wantErr: false,
		},
		{
			name:    "empty employee number is valid (optional)",
			input:   "",
			wantErr: false,
		},
		{
			name:        "employee number too long",
			input:       strings.Repeat("A", 21),
			wantErr:     true,
			errContains: "20文字以内",
		},
		{
			name:    "employee number exactly 20 chars",
			input:   strings.Repeat("A", 20),
			wantErr: false,
		},
		{
			name:        "invalid characters - spaces",
			input:       "EMP 001",
			wantErr:     true,
			errContains: "英数字とハイフン",
		},
		{
			name:        "invalid characters - special",
			input:       "EMP#001",
			wantErr:     true,
			errContains: "英数字とハイフン",
		},
		{
			name:        "invalid characters - Japanese",
			input:       "社員001",
			wantErr:     true,
			errContains: "英数字とハイフン",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateEmployeeNumber(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDepartment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid department",
			input:   "Engineering",
			wantErr: false,
		},
		{
			name:    "valid Japanese department",
			input:   "開発部",
			wantErr: false,
		},
		{
			name:    "empty department is valid (optional)",
			input:   "",
			wantErr: false,
		},
		{
			name:        "department too long",
			input:       strings.Repeat("a", 51),
			wantErr:     true,
			errContains: "50文字以内",
		},
		{
			name:    "department exactly 50 chars",
			input:   strings.Repeat("a", 50),
			wantErr: false,
		},
		{
			name:        "department with emoji",
			input:       "開発部😀",
			wantErr:     true,
			errContains: "絵文字",
		},
		{
			name:        "department with control character",
			input:       "開発部\x00",
			wantErr:     true,
			errContains: "不正な文字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDepartment(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUserInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userName       string
		email          string
		employeeNumber string
		department     string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "all valid fields",
			userName:       "田中太郎",
			email:          "tanaka@example.com",
			employeeNumber: "EMP-001",
			department:     "開発部",
			wantErr:        false,
		},
		{
			name:           "valid with optional fields empty",
			userName:       "田中太郎",
			email:          "tanaka@example.com",
			employeeNumber: "",
			department:     "",
			wantErr:        false,
		},
		{
			name:           "invalid name fails first",
			userName:       "",
			email:          "tanaka@example.com",
			employeeNumber: "EMP-001",
			department:     "開発部",
			wantErr:        true,
			errContains:    "名前",
		},
		{
			name:           "invalid email fails",
			userName:       "田中太郎",
			email:          "invalid-email",
			employeeNumber: "EMP-001",
			department:     "開発部",
			wantErr:        true,
			errContains:    "メールアドレス",
		},
		{
			name:           "invalid employee number fails",
			userName:       "田中太郎",
			email:          "tanaka@example.com",
			employeeNumber: "EMP 001",
			department:     "開発部",
			wantErr:        true,
			errContains:    "社員番号",
		},
		{
			name:           "invalid department fails",
			userName:       "田中太郎",
			email:          "tanaka@example.com",
			employeeNumber: "EMP-001",
			department:     "開発部😀",
			wantErr:        true,
			errContains:    "所属部署",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateUserInput(tt.userName, tt.email, tt.employeeNumber, tt.department)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainsEmoji(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "string with emoji face",
			input: "Hello 😀",
			want:  true,
		},
		{
			name:  "string with emoji sun",
			input: "☀️ Sunny",
			want:  true,
		},
		{
			name:  "normal text",
			input: "Hello World",
			want:  false,
		},
		{
			name:  "Japanese text",
			input: "こんにちは",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsEmoji(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainsControlChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "string with null character",
			input: "Hello\x00World",
			want:  true,
		},
		{
			name:  "string with bell character",
			input: "Hello\x07World",
			want:  true,
		},
		{
			name:  "string with newline (allowed)",
			input: "Hello\nWorld",
			want:  false,
		},
		{
			name:  "string with tab (allowed)",
			input: "Hello\tWorld",
			want:  false,
		},
		{
			name:  "string with carriage return (allowed)",
			input: "Hello\rWorld",
			want:  false,
		},
		{
			name:  "normal text",
			input: "Hello World",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsControlChars(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoleConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Role("admin"), RoleAdmin)
	assert.Equal(t, Role("editor"), RoleEditor)
	assert.Equal(t, Role("viewer"), RoleViewer)
}
