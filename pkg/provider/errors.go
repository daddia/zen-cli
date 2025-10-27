package provider

import (
	"fmt"
)

// ErrorCode represents specific provider error codes
type ErrorCode string

const (
	// ErrorCodeNotFound indicates the provider binary or API endpoint was not found
	ErrorCodeNotFound ErrorCode = "not_found"

	// ErrorCodeVersionMismatch indicates the provider version doesn't meet requirements
	ErrorCodeVersionMismatch ErrorCode = "version_mismatch"

	// ErrorCodeExecution indicates the provider operation failed during execution
	ErrorCodeExecution ErrorCode = "execution_failed"

	// ErrorCodeTimeout indicates the operation exceeded its deadline
	ErrorCodeTimeout ErrorCode = "timeout"

	// ErrorCodeCanceled indicates the operation was canceled by the caller
	ErrorCodeCanceled ErrorCode = "canceled"

	// ErrorCodeParse indicates the provider output could not be parsed
	ErrorCodeParse ErrorCode = "parse_failed"

	// ErrorCodeInvalidOp indicates the operation is not supported by the provider
	ErrorCodeInvalidOp ErrorCode = "invalid_operation"
)

// Error represents a provider error with structured information
//
// Error provides rich context about provider failures including the error code,
// human-readable message, provider name, operation that failed, and the underlying
// error if any.
type Error struct {
	// Code is the specific error code for programmatic handling
	Code ErrorCode

	// Message is the human-readable error message
	Message string

	// Provider is the name of the provider that generated the error
	Provider string

	// Op is the operation that was being executed when the error occurred
	// Example: "git.clone", "pull_request.list"
	Op string

	// Err is the underlying error, if any
	Err error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Provider != "" && e.Op != "" {
		if e.Err != nil {
			return fmt.Sprintf("[%s:%s:%s] %s: %v", e.Provider, e.Op, e.Code, e.Message, e.Err)
		}
		return fmt.Sprintf("[%s:%s:%s] %s", e.Provider, e.Op, e.Code, e.Message)
	}
	if e.Provider != "" {
		if e.Err != nil {
			return fmt.Sprintf("[%s:%s] %s: %v", e.Provider, e.Code, e.Message, e.Err)
		}
		return fmt.Sprintf("[%s:%s] %s", e.Provider, e.Code, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error chain support
func (e *Error) Unwrap() error {
	return e.Err
}

// IsProviderError checks if an error is a provider error
func IsProviderError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// GetErrorCode extracts the error code from a provider error
// Returns empty string if the error is not a provider error
func GetErrorCode(err error) ErrorCode {
	if provErr, ok := err.(*Error); ok {
		return provErr.Code
	}
	return ""
}

// NewError creates a new provider error
func NewError(code ErrorCode, message, provider string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Provider: provider,
	}
}

// NewErrorWithOp creates a new provider error with operation context
func NewErrorWithOp(code ErrorCode, message, provider, op string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Provider: provider,
		Op:       op,
	}
}

// WrapError wraps an underlying error with provider context
func WrapError(code ErrorCode, message, provider string, err error) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Provider: provider,
		Err:      err,
	}
}

// WrapErrorWithOp wraps an underlying error with provider and operation context
func WrapErrorWithOp(code ErrorCode, message, provider, op string, err error) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Provider: provider,
		Op:       op,
		Err:      err,
	}
}

// ErrNotFound creates a not found error
func ErrNotFound(provider, resource string) *Error {
	return NewError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), provider)
}

// ErrVersionMismatch creates a version mismatch error
func ErrVersionMismatch(provider, current, required string) *Error {
	return NewError(ErrorCodeVersionMismatch,
		fmt.Sprintf("version %s does not meet requirement %s", current, required),
		provider)
}

// ErrExecution creates an execution error
func ErrExecution(provider, op, message string, err error) *Error {
	return WrapErrorWithOp(ErrorCodeExecution, message, provider, op, err)
}

// ErrTimeout creates a timeout error
func ErrTimeout(provider, op string) *Error {
	return NewErrorWithOp(ErrorCodeTimeout, "operation exceeded deadline", provider, op)
}

// ErrCanceled creates a canceled error
func ErrCanceled(provider, op string) *Error {
	return NewErrorWithOp(ErrorCodeCanceled, "operation was canceled", provider, op)
}

// ErrParse creates a parse error
func ErrParse(provider, op, message string, err error) *Error {
	return WrapErrorWithOp(ErrorCodeParse, message, provider, op, err)
}

// ErrInvalidOp creates an invalid operation error
func ErrInvalidOp(provider, op string) *Error {
	return NewErrorWithOp(ErrorCodeInvalidOp,
		fmt.Sprintf("operation '%s' is not supported", op),
		provider, op)
}
