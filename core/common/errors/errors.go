// common/errors/errors.go
package errors

import (
	"fmt"
	"net/http"
)

type AppErrorType string

const (
	ErrorTypeValidation   AppErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     AppErrorType = "NOT_FOUND"
	ErrorTypeDatabase     AppErrorType = "DATABASE_ERROR"
	ErrorTypeUnauthorized AppErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    AppErrorType = "FORBIDDEN"
	ErrorTypeConflict     AppErrorType = "CONFLICT"
	ErrorTypeInternal     AppErrorType = "INTERNAL_ERROR"
)

type AppError struct {
	Type       AppErrorType `json:"type"`
	Message    string       `json:"message"`
	StatusCode int          `json:"-"`
	Err        error        `json:"-"` // Error original para logs
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Constructores de errores comunes
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

func NewDatabaseError(err error) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Message:    "Database operation failed",
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewInternalError(err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    "An internal error occurred",
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}
