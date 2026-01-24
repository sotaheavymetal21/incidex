package domain

import (
	"errors"
	"slices"
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
	MinLength:      12,
	RequireUpper:   true,
	RequireLower:   true,
	RequireNumber:  true,
	RequireSpecial: true,
}

// Validate checks if a password meets the policy requirements
func (p *PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return ErrValidation("Password must be at least 12 characters long")
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
func IsCommonPassword(password string) bool {
	// Top 100 most common passwords
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "111111", "iloveyou", "master", "sunshine",
		"ashley", "bailey", "passw0rd", "shadow", "123123",
		"654321", "superman", "qazwsx", "michael", "football",
		"welcome", "jesus", "ninja", "mustang", "password1",
		"123456789", "adobe123", "admin", "1234567890", "photoshop",
		"1234", "12345", "password123", "pussy", "hunter",
		"harley", "computer", "mickey", "tigger", "qwerty123",
		"charlie", "donald", "pepper", "ginger", "liverpool",
		"buster", "dallas", "access", "love", "flower",
		"batman", "startrek", "killer", "soccer", "ranger",
		"sunshine", "jordan", "asshole", "master", "andrea",
		"fuckme", "pepper", "whatever", "hockey", "corvette",
		"maggie", "george", "bigdog", "cheese", "matthew",
		"112233", "samsung", "compaq", "boston", "taylor",
		"yellow", "jessica", "summer", "sparky", "test",
		"scooter", "forever", "maverick", "cookie", "peanut",
		"morgan", "falcon", "cowboy", "ferrari", "knight",
		"amazon", "google", "twitter", "facebook", "linkedin",
		"welcome1", "password!", "admin123", "root", "toor",
	}

	return slices.Contains(commonPasswords, password)
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
