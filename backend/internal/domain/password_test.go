package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordPolicy_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid password with all requirements",
			password: "ValidPass123!",
			wantErr:  false,
		},
		{
			name:     "valid password exactly 12 characters",
			password: "Ab1!defghijk",
			wantErr:  false,
		},
		{
			name:        "too short - 11 characters",
			password:    "Ab1!defghij",
			wantErr:     true,
			errContains: "at least 12 characters",
		},
		{
			name:        "missing uppercase letter",
			password:    "validpass123!",
			wantErr:     true,
			errContains: "uppercase letter",
		},
		{
			name:        "missing lowercase letter",
			password:    "VALIDPASS123!",
			wantErr:     true,
			errContains: "lowercase letter",
		},
		{
			name:        "missing number",
			password:    "ValidPassword!",
			wantErr:     true,
			errContains: "number",
		},
		{
			name:        "missing special character",
			password:    "ValidPass1234",
			wantErr:     true,
			errContains: "special character",
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     true,
			errContains: "at least 12 characters",
		},
		{
			name:     "password with unicode special chars",
			password: "ValidPass123@#",
			wantErr:  false,
		},
		{
			name:     "password with symbols",
			password: "ValidPass123$",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := DefaultPasswordPolicy.Validate(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	t.Run("uses default policy", func(t *testing.T) {
		t.Parallel()
		// 強いパスワードで成功するはず
		err := ValidatePassword("StrongPass123!")
		assert.NoError(t, err)

		// 弱いパスワードで失敗するはず
		err = ValidatePassword("weak")
		assert.Error(t, err)
	})
}

func TestIsCommonPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "password is common",
			password: "password",
			want:     true,
		},
		{
			name:     "password123 is common",
			password: "password123",
			want:     true,
		},
		{
			name:     "qwerty is common",
			password: "qwerty",
			want:     true,
		},
		{
			name:     "123456 is common",
			password: "123456",
			want:     true,
		},
		{
			name:     "admin is common",
			password: "admin",
			want:     true,
		},
		{
			name:     "strong unique password is not common",
			password: "xK9#mP2$vL7@",
			want:     false,
		},
		{
			name:     "random string is not common",
			password: "MyUnique$ecretP@ss",
			want:     false,
		},
		{
			name:     "empty string is not in common list",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsCommonPassword(tt.password)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "strong unique password passes",
			password: "StrongPass123!",
			wantErr:  false,
		},
		{
			name:        "common password fails",
			password:    "password",
			wantErr:     true,
			errContains: "too common",
		},
		{
			name:        "common password123 fails",
			password:    "password123",
			wantErr:     true,
			errContains: "too common",
		},
		{
			name:        "weak password fails policy check",
			password:    "Weak1!",
			wantErr:     true,
			errContains: "at least 12 characters",
		},
		{
			name:     "complex password passes all checks",
			password: "MyC0mpl3x#P@ssword!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCustomPasswordPolicy(t *testing.T) {
	t.Parallel()

	t.Run("custom policy with lower requirements", func(t *testing.T) {
		t.Parallel()
		policy := &PasswordPolicy{
			MinLength:      8,
			RequireUpper:   false,
			RequireLower:   true,
			RequireNumber:  true,
			RequireSpecial: false,
		}

		// 大文字と特殊文字なしで成功するはず
		err := policy.Validate("simple12")
		assert.NoError(t, err)

		// 最小長さは依然として必要
		err = policy.Validate("simp1")
		assert.Error(t, err)
	})
}
