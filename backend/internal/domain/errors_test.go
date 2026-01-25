package domain

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *DomainError
		contains []string
	}{
		{
			name: "error without underlying error",
			err: &DomainError{
				Code:    ErrCodeNotFound,
				Message: "User not found",
			},
			contains: []string{"NOT_FOUND", "User not found"},
		},
		{
			name: "error with underlying error",
			err: &DomainError{
				Code:    ErrCodeInternal,
				Message: "Database error",
				Err:     errors.New("connection refused"),
			},
			contains: []string{"INTERNAL_ERROR", "Database error", "connection refused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errStr := tt.err.Error()
			for _, contain := range tt.contains {
				assert.Contains(t, errStr, contain)
			}
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("returns underlying error when present", func(t *testing.T) {
		t.Parallel()
		underlying := errors.New("underlying error")
		domainErr := &DomainError{
			Code:    ErrCodeInternal,
			Message: "Something went wrong",
			Err:     underlying,
		}

		unwrapped := domainErr.Unwrap()
		assert.Equal(t, underlying, unwrapped)
	})

	t.Run("returns nil when no underlying error", func(t *testing.T) {
		t.Parallel()
		domainErr := &DomainError{
			Code:    ErrCodeNotFound,
			Message: "Not found",
		}

		unwrapped := domainErr.Unwrap()
		assert.Nil(t, unwrapped)
	})
}

func TestNewDomainError(t *testing.T) {
	t.Parallel()

	err := NewDomainError(ErrCodeValidation, http.StatusBadRequest, "Invalid input")

	assert.Equal(t, ErrCodeValidation, err.Code)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, "Invalid input", err.Message)
	assert.NotNil(t, err.Details)
	assert.Nil(t, err.Err)
}

func TestDomainError_WithDetails(t *testing.T) {
	t.Parallel()

	err := NewDomainError(ErrCodeValidation, http.StatusBadRequest, "Invalid input").
		WithDetails("field", "email").
		WithDetails("reason", "format")

	assert.Equal(t, "email", err.Details["field"])
	assert.Equal(t, "format", err.Details["reason"])
}

func TestDomainError_WithError(t *testing.T) {
	t.Parallel()

	underlying := errors.New("db connection failed")
	err := NewDomainError(ErrCodeDatabaseError, http.StatusInternalServerError, "Database error").
		WithError(underlying)

	assert.Equal(t, underlying, err.Err)
	assert.True(t, errors.Is(err, underlying))
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()

	err := ErrNotFound("user")

	assert.Equal(t, ErrCodeNotFound, err.Code)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Contains(t, err.Message, "user")
	assert.Contains(t, err.Message, "not found")
}

func TestErrUnauthorized(t *testing.T) {
	t.Parallel()

	t.Run("with custom message", func(t *testing.T) {
		t.Parallel()
		err := ErrUnauthorized("Invalid token")

		assert.Equal(t, ErrCodeUnauthorized, err.Code)
		assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
		assert.Equal(t, "Invalid token", err.Message)
	})

	t.Run("with empty message uses default", func(t *testing.T) {
		t.Parallel()
		err := ErrUnauthorized("")

		assert.Equal(t, ErrCodeUnauthorized, err.Code)
		assert.Equal(t, "Unauthorized access", err.Message)
	})
}

func TestErrForbidden(t *testing.T) {
	t.Parallel()

	t.Run("with custom message", func(t *testing.T) {
		t.Parallel()
		err := ErrForbidden("Insufficient permissions")

		assert.Equal(t, ErrCodeForbidden, err.Code)
		assert.Equal(t, http.StatusForbidden, err.StatusCode)
		assert.Equal(t, "Insufficient permissions", err.Message)
	})

	t.Run("with empty message uses default", func(t *testing.T) {
		t.Parallel()
		err := ErrForbidden("")

		assert.Equal(t, "Access forbidden", err.Message)
	})
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	err := ErrValidation("Email is required")

	assert.Equal(t, ErrCodeValidation, err.Code)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, "Email is required", err.Message)
}

func TestErrConflict(t *testing.T) {
	t.Parallel()

	err := ErrConflict("Email already exists")

	assert.Equal(t, ErrCodeConflict, err.Code)
	assert.Equal(t, http.StatusConflict, err.StatusCode)
	assert.Equal(t, "Email already exists", err.Message)
}

func TestErrBadRequest(t *testing.T) {
	t.Parallel()

	err := ErrBadRequest("Invalid request body")

	assert.Equal(t, ErrCodeBadRequest, err.Code)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, "Invalid request body", err.Message)
}

func TestErrInternal(t *testing.T) {
	t.Parallel()

	t.Run("with custom message", func(t *testing.T) {
		t.Parallel()
		underlying := errors.New("unexpected error")
		err := ErrInternal("Something went wrong", underlying)

		assert.Equal(t, ErrCodeInternal, err.Code)
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
		assert.Equal(t, "Something went wrong", err.Message)
		assert.Equal(t, underlying, err.Err)
	})

	t.Run("with empty message uses default", func(t *testing.T) {
		t.Parallel()
		err := ErrInternal("", nil)

		assert.Equal(t, "An internal error occurred", err.Message)
	})
}

func TestErrDatabase(t *testing.T) {
	t.Parallel()

	t.Run("with custom message", func(t *testing.T) {
		t.Parallel()
		underlying := errors.New("connection refused")
		err := ErrDatabase("Failed to connect", underlying)

		assert.Equal(t, ErrCodeDatabaseError, err.Code)
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
		assert.Equal(t, "Failed to connect", err.Message)
		assert.Equal(t, underlying, err.Err)
	})

	t.Run("with empty message uses default", func(t *testing.T) {
		t.Parallel()
		err := ErrDatabase("", nil)

		assert.Equal(t, "A database error occurred", err.Message)
	})
}

func TestErrExternalAPI(t *testing.T) {
	t.Parallel()

	underlying := errors.New("timeout")
	err := ErrExternalAPI("Slack", underlying)

	assert.Equal(t, ErrCodeExternalAPI, err.Code)
	assert.Equal(t, http.StatusBadGateway, err.StatusCode)
	assert.Contains(t, err.Message, "Slack")
	assert.Equal(t, underlying, err.Err)
}

func TestIsDomainError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "domain error returns true",
			err:  ErrNotFound("user"),
			want: true,
		},
		{
			name: "standard error returns false",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "nil returns false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsDomainError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAsDomainError(t *testing.T) {
	t.Parallel()

	t.Run("converts domain error", func(t *testing.T) {
		t.Parallel()
		domainErr := ErrNotFound("user")

		result, ok := AsDomainError(domainErr)

		require.True(t, ok)
		assert.Equal(t, domainErr, result)
	})

	t.Run("returns false for standard error", func(t *testing.T) {
		t.Parallel()
		standardErr := errors.New("standard error")

		result, ok := AsDomainError(standardErr)

		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("returns false for nil", func(t *testing.T) {
		t.Parallel()

		result, ok := AsDomainError(nil)

		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestErrorCodes(t *testing.T) {
	t.Parallel()

	// error コードが期待通りであることを確認
	assert.Equal(t, ErrorCode("NOT_FOUND"), ErrCodeNotFound)
	assert.Equal(t, ErrorCode("UNAUTHORIZED"), ErrCodeUnauthorized)
	assert.Equal(t, ErrorCode("FORBIDDEN"), ErrCodeForbidden)
	assert.Equal(t, ErrorCode("VALIDATION_ERROR"), ErrCodeValidation)
	assert.Equal(t, ErrorCode("CONFLICT"), ErrCodeConflict)
	assert.Equal(t, ErrorCode("BAD_REQUEST"), ErrCodeBadRequest)
	assert.Equal(t, ErrorCode("INTERNAL_ERROR"), ErrCodeInternal)
	assert.Equal(t, ErrorCode("DATABASE_ERROR"), ErrCodeDatabaseError)
	assert.Equal(t, ErrorCode("EXTERNAL_API_ERROR"), ErrCodeExternalAPI)
}
