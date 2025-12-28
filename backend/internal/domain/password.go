package domain

import (
	"errors"
	"unicode"
)

// PasswordPolicy defines password validation rules
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordPolicy is the default password policy
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	RequireUpper:   true,
	RequireLower:   true,
	RequireNumber:  true,
	RequireSpecial: false, // Optional for now
}

// Validate checks if a password meets the policy requirements
func (p *PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return ErrValidation("Password must be at least 8 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return ErrValidation("Password must contain at least one uppercase letter")
	}
	if p.RequireLower && !hasLower {
		return ErrValidation("Password must contain at least one lowercase letter")
	}
	if p.RequireNumber && !hasNumber {
		return ErrValidation("Password must contain at least one number")
	}
	if p.RequireSpecial && !hasSpecial {
		return ErrValidation("Password must contain at least one special character")
	}

	return nil
}

// ValidatePassword validates a password using the default policy
func ValidatePassword(password string) error {
	return DefaultPasswordPolicy.Validate(password)
}

// IsCommonPassword checks if the password is in a list of common passwords
// For now, just check for extremely weak passwords
func IsCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "12345678", "password123", "qwerty", "abc123",
		"password1", "12345", "1234567890", "letmein", "welcome",
	}

	for _, common := range commonPasswords {
		if password == common {
			return true
		}
	}
	return false
}

// ValidatePasswordStrength validates password strength including common password check
func ValidatePasswordStrength(password string) error {
	// Check common passwords first
	if IsCommonPassword(password) {
		return errors.New("password is too common, please choose a stronger password")
	}

	// Validate against policy
	return ValidatePassword(password)
}
