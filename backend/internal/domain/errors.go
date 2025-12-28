package domain

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"

	// Server errors (5xx)
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPI   ErrorCode = "EXTERNAL_API_ERROR"
)

// DomainError represents a domain-level error with user-friendly messaging
type DomainError struct {
	Code       ErrorCode
	Message    string
	StatusCode int
	Details    map[string]interface{}
	Err        error // Original error for logging
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new DomainError
func NewDomainError(code ErrorCode, statusCode int, message string) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	e.Details[key] = value
	return e
}

// WithError wraps an underlying error
func (e *DomainError) WithError(err error) *DomainError {
	e.Err = err
	return e
}

// Helper functions for common error types

// ErrNotFound creates a not found error
func ErrNotFound(resource string) *DomainError {
	return NewDomainError(
		ErrCodeNotFound,
		http.StatusNotFound,
		fmt.Sprintf("%s not found", resource),
	)
}

// ErrUnauthorized creates an unauthorized error
func ErrUnauthorized(message string) *DomainError {
	if message == "" {
		message = "Unauthorized access"
	}
	return NewDomainError(
		ErrCodeUnauthorized,
		http.StatusUnauthorized,
		message,
	)
}

// ErrForbidden creates a forbidden error
func ErrForbidden(message string) *DomainError {
	if message == "" {
		message = "Access forbidden"
	}
	return NewDomainError(
		ErrCodeForbidden,
		http.StatusForbidden,
		message,
	)
}

// ErrValidation creates a validation error
func ErrValidation(message string) *DomainError {
	return NewDomainError(
		ErrCodeValidation,
		http.StatusBadRequest,
		message,
	)
}

// ErrConflict creates a conflict error
func ErrConflict(message string) *DomainError {
	return NewDomainError(
		ErrCodeConflict,
		http.StatusConflict,
		message,
	)
}

// ErrBadRequest creates a bad request error
func ErrBadRequest(message string) *DomainError {
	return NewDomainError(
		ErrCodeBadRequest,
		http.StatusBadRequest,
		message,
	)
}

// ErrInternal creates an internal server error
func ErrInternal(message string, err error) *DomainError {
	if message == "" {
		message = "An internal error occurred"
	}
	return NewDomainError(
		ErrCodeInternal,
		http.StatusInternalServerError,
		message,
	).WithError(err)
}

// ErrDatabase creates a database error
func ErrDatabase(message string, err error) *DomainError {
	if message == "" {
		message = "A database error occurred"
	}
	return NewDomainError(
		ErrCodeDatabaseError,
		http.StatusInternalServerError,
		message,
	).WithError(err)
}

// ErrExternalAPI creates an external API error
func ErrExternalAPI(service string, err error) *DomainError {
	return NewDomainError(
		ErrCodeExternalAPI,
		http.StatusBadGateway,
		fmt.Sprintf("External service '%s' error", service),
	).WithError(err)
}

// IsDomainError checks if an error is a DomainError
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

// AsDomainError converts an error to DomainError if possible
func AsDomainError(err error) (*DomainError, bool) {
	if err == nil {
		return nil, false
	}
	domainErr, ok := err.(*DomainError)
	return domainErr, ok
}
