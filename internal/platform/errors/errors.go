package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a unique error code
type ErrorCode string

const (
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	StatusCode int       `json:"-"`
	Err        error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with AppError context
func Wrap(err error, code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common error constructors
func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(ErrCodeNotFound, message, http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

func InternalError(message string) *AppError {
	return New(ErrCodeInternalError, message, http.StatusInternalServerError)
}

func ServiceUnavailable(message string) *AppError {
	return New(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

func ValidationError(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}
