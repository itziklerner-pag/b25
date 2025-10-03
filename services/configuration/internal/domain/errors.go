package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a configuration is not found
	ErrNotFound = errors.New("configuration not found")

	// ErrDuplicateKey is returned when a configuration with the same key already exists
	ErrDuplicateKey = errors.New("configuration key already exists")

	// ErrInvalidVersion is returned when an invalid version is requested
	ErrInvalidVersion = errors.New("invalid configuration version")

	// ErrInvalidFormat is returned when the configuration format is invalid
	ErrInvalidFormat = errors.New("invalid configuration format")

	// ErrValidationFailed is returned when validation fails
	ErrValidationFailed = errors.New("configuration validation failed")

	// ErrDatabaseError is returned when a database operation fails
	ErrDatabaseError = errors.New("database error")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
