package auth

import (
	"fmt"
)

// ErrorCode represents specific authentication error codes
type ErrorCode string

const (
	ErrorCodeAuthenticationFailed ErrorCode = "authentication_failed"
	ErrorCodeInvalidCredentials   ErrorCode = "invalid_credentials"
	ErrorCodeCredentialNotFound   ErrorCode = "credential_not_found"
	ErrorCodeProviderNotSupported ErrorCode = "provider_not_supported"
	ErrorCodeStorageError         ErrorCode = "storage_error"
	ErrorCodeNetworkError         ErrorCode = "network_error"
	ErrorCodeConfigurationError   ErrorCode = "configuration_error"
	ErrorCodeRateLimited          ErrorCode = "rate_limited"
	ErrorCodeTokenExpired         ErrorCode = "token_expired"
	ErrorCodeInsufficientScopes   ErrorCode = "insufficient_scopes"
)

// Error represents an authentication error with structured information
type Error struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Provider   string      `json:"provider,omitempty"`
	RetryAfter int         `json:"retry_after,omitempty"` // Seconds to wait before retry
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Provider != "" {
		return fmt.Sprintf("[%s:%s] %s", e.Provider, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// GetErrorCode extracts the error code from an authentication error
func GetErrorCode(err error) ErrorCode {
	if authErr, ok := err.(*Error); ok {
		return authErr.Code
	}
	return ""
}

// NewAuthError creates a new authentication error
func NewAuthError(code ErrorCode, message string, provider string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Provider: provider,
	}
}

// NewAuthErrorWithDetails creates a new authentication error with details
func NewAuthErrorWithDetails(code ErrorCode, message string, provider string, details interface{}) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Provider: provider,
		Details:  details,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(provider string, retryAfter int) *Error {
	return &Error{
		Code:       ErrorCodeRateLimited,
		Message:    fmt.Sprintf("API rate limit exceeded for %s", provider),
		Provider:   provider,
		RetryAfter: retryAfter,
	}
}

// NewStorageError creates a new storage error
func NewStorageError(message string, details interface{}) *Error {
	return &Error{
		Code:    ErrorCodeStorageError,
		Message: message,
		Details: details,
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(provider string, details interface{}) *Error {
	return &Error{
		Code:     ErrorCodeNetworkError,
		Message:  fmt.Sprintf("Network error accessing %s API", provider),
		Provider: provider,
		Details:  details,
	}
}
