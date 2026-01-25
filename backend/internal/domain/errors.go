package domain

import (
	"fmt"
	"net/http"
)

// ErrorCode は特定の error タイプを表します
type ErrorCode string

const (
	// クライアント error (4xx)
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"

	// サーバー error (5xx)
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPI   ErrorCode = "EXTERNAL_API_ERROR"
)

// DomainError はユーザーフレンドリーなメッセージを持つドメインレベルの error を表します
type DomainError struct {
	Code       ErrorCode
	Message    string
	StatusCode int
	Details    map[string]interface{}
	Err        error // ロギング用の元 error
}

// Error は error インターフェースを実装します
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap は元の error を返します
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError は新しい DomainError を作成します
func NewDomainError(code ErrorCode, statusCode int, message string) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetails は error に詳細情報を追加します
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	e.Details[key] = value
	return e
}

// WithError は元の error をラップします
func (e *DomainError) WithError(err error) *DomainError {
	e.Err = err
	return e
}

// ヘルパー関数

// ErrNotFound は not found error を作成します
func ErrNotFound(resource string) *DomainError {
	return NewDomainError(
		ErrCodeNotFound,
		http.StatusNotFound,
		fmt.Sprintf("%s not found", resource),
	)
}

// ErrUnauthorized は unauthorized error を作成します
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

// ErrForbidden は forbidden error を作成します
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

// ErrValidation は バリデーション error を作成します
func ErrValidation(message string) *DomainError {
	return NewDomainError(
		ErrCodeValidation,
		http.StatusBadRequest,
		message,
	)
}

// ErrConflict は conflict error を作成します
func ErrConflict(message string) *DomainError {
	return NewDomainError(
		ErrCodeConflict,
		http.StatusConflict,
		message,
	)
}

// ErrBadRequest は bad request error を作成します
func ErrBadRequest(message string) *DomainError {
	return NewDomainError(
		ErrCodeBadRequest,
		http.StatusBadRequest,
		message,
	)
}

// ErrInternal は internal server error を作成します
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

// ErrDatabase は データベース error を作成します
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

// ErrExternalAPI は外部 API error を作成します
func ErrExternalAPI(service string, err error) *DomainError {
	return NewDomainError(
		ErrCodeExternalAPI,
		http.StatusBadGateway,
		fmt.Sprintf("External service '%s' error", service),
	).WithError(err)
}

// IsDomainError は error が DomainError かどうかを確認します
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

// AsDomainError は可能であれば error を DomainError に変換します
func AsDomainError(err error) (*DomainError, bool) {
	if err == nil {
		return nil, false
	}
	domainErr, ok := err.(*DomainError)
	return domainErr, ok
}
