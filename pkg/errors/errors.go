package errors

import (
	"fmt"

	"github.com/daddia/zen/pkg/types"
	"github.com/pkg/errors"
)

// New creates a new error with the given message
func New(message string) error {
	return errors.New(message)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

// Wrapf wraps an error with formatted additional context
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

// NewWithCode creates a new error with a specific error code
func NewWithCode(code types.ErrorCode, message string) error {
	return &types.Error{
		Code:    code,
		Message: message,
	}
}

// NewWithCodef creates a new error with a specific error code and formatted message
func NewWithCodef(code types.ErrorCode, format string, args ...interface{}) error {
	return &types.Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewWithDetails creates a new error with code, message, and details
func NewWithDetails(code types.ErrorCode, message, details string) error {
	return &types.Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// IsCode checks if an error has a specific error code
func IsCode(err error, code types.ErrorCode) bool {
	if zenErr, ok := err.(*types.Error); ok {
		return zenErr.Code == code
	}
	return false
}

// GetCode extracts the error code from an error, returns ErrorCodeUnknown if not found
func GetCode(err error) types.ErrorCode {
	if zenErr, ok := err.(*types.Error); ok {
		return zenErr.Code
	}
	return types.ErrorCodeUnknown
}

// Common error constructors

// ErrInvalidInput creates an invalid input error
func ErrInvalidInput(message string) error {
	return NewWithCode(types.ErrorCodeInvalidInput, message)
}

// ErrNotFound creates a not found error
func ErrNotFound(resource string) error {
	return NewWithCodef(types.ErrorCodeNotFound, "%s not found", resource)
}

// ErrAlreadyExists creates an already exists error
func ErrAlreadyExists(resource string) error {
	return NewWithCodef(types.ErrorCodeAlreadyExists, "%s already exists", resource)
}

// ErrInvalidConfig creates an invalid configuration error
func ErrInvalidConfig(details string) error {
	return NewWithDetails(types.ErrorCodeInvalidConfig, "invalid configuration", details)
}

// ErrConfigNotFound creates a configuration not found error
func ErrConfigNotFound(path string) error {
	return NewWithDetails(types.ErrorCodeConfigNotFound, "configuration file not found", path)
}

// ErrWorkspaceNotInitialized creates a workspace not initialized error
func ErrWorkspaceNotInitialized() error {
	return NewWithCode(types.ErrorCodeWorkspaceNotInit, "workspace not initialized")
}

// ErrInvalidWorkspace creates an invalid workspace error
func ErrInvalidWorkspace(details string) error {
	return NewWithDetails(types.ErrorCodeInvalidWorkspace, "invalid workspace", details)
}
