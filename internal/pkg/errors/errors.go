package errors

import (
	"fmt"
	"net/http"
)

// AppError represents application-specific errors
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error constructors
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, message, nil)
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message, nil)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, message, nil)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message, nil)
}

func InternalServerError(message string, err error) *AppError {
	return New(http.StatusInternalServerError, message, err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if err != nil {
		if e, ok := err.(*AppError); ok {
			return e, true
		}
	}
	return appErr, false
}
