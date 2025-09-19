package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		authErr  *Error
		expected string
	}{
		{
			name: "error with provider",
			authErr: &Error{
				Code:     ErrorCodeAuthenticationFailed,
				Message:  "authentication failed",
				Provider: "github",
			},
			expected: "[github:authentication_failed] authentication failed",
		},
		{
			name: "error without provider",
			authErr: &Error{
				Code:    ErrorCodeStorageError,
				Message: "storage error",
			},
			expected: "[storage_error] storage error",
		},
		{
			name: "error with details",
			authErr: &Error{
				Code:     ErrorCodeNetworkError,
				Message:  "network error",
				Provider: "gitlab",
				Details:  "connection timeout",
			},
			expected: "[gitlab:network_error] network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.authErr.Error()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "auth error",
			err: &Error{
				Code:    ErrorCodeAuthenticationFailed,
				Message: "test error",
			},
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := IsAuthError(tt.err)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name: "auth error",
			err: &Error{
				Code:    ErrorCodeInvalidCredentials,
				Message: "invalid credentials",
			},
			expected: ErrorCodeInvalidCredentials,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: "",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := GetErrorCode(tt.err)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewAuthError(t *testing.T) {
	// Arrange
	code := ErrorCodeAuthenticationFailed
	message := "authentication failed"
	provider := "github"

	// Act
	err := NewAuthError(code, message, provider)

	// Assert
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, provider, err.Provider)
	assert.Nil(t, err.Details)
	assert.Equal(t, 0, err.RetryAfter)
}

func TestNewAuthErrorWithDetails(t *testing.T) {
	// Arrange
	code := ErrorCodeNetworkError
	message := "network error"
	provider := "gitlab"
	details := map[string]string{"reason": "timeout"}

	// Act
	err := NewAuthErrorWithDetails(code, message, provider, details)

	// Assert
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, details, err.Details)
	assert.Equal(t, 0, err.RetryAfter)
}

func TestNewRateLimitError(t *testing.T) {
	// Arrange
	provider := "github"
	retryAfter := 3600

	// Act
	err := NewRateLimitError(provider, retryAfter)

	// Assert
	assert.Equal(t, ErrorCodeRateLimited, err.Code)
	assert.Contains(t, err.Message, "API rate limit exceeded")
	assert.Contains(t, err.Message, provider)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, retryAfter, err.RetryAfter)
}

func TestNewStorageError(t *testing.T) {
	// Arrange
	message := "storage operation failed"
	details := "disk full"

	// Act
	err := NewStorageError(message, details)

	// Assert
	assert.Equal(t, ErrorCodeStorageError, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, details, err.Details)
	assert.Empty(t, err.Provider)
	assert.Equal(t, 0, err.RetryAfter)
}

func TestNewNetworkError(t *testing.T) {
	// Arrange
	provider := "gitlab"
	details := "connection timeout"

	// Act
	err := NewNetworkError(provider, details)

	// Assert
	assert.Equal(t, ErrorCodeNetworkError, err.Code)
	assert.Contains(t, err.Message, "Network error accessing")
	assert.Contains(t, err.Message, provider)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, details, err.Details)
}

func TestErrorCode_Constants(t *testing.T) {
	// Arrange & Act & Assert - verify all error codes are defined
	assert.Equal(t, ErrorCode("authentication_failed"), ErrorCodeAuthenticationFailed)
	assert.Equal(t, ErrorCode("invalid_credentials"), ErrorCodeInvalidCredentials)
	assert.Equal(t, ErrorCode("credential_not_found"), ErrorCodeCredentialNotFound)
	assert.Equal(t, ErrorCode("provider_not_supported"), ErrorCodeProviderNotSupported)
	assert.Equal(t, ErrorCode("storage_error"), ErrorCodeStorageError)
	assert.Equal(t, ErrorCode("network_error"), ErrorCodeNetworkError)
	assert.Equal(t, ErrorCode("configuration_error"), ErrorCodeConfigurationError)
	assert.Equal(t, ErrorCode("rate_limited"), ErrorCodeRateLimited)
	assert.Equal(t, ErrorCode("token_expired"), ErrorCodeTokenExpired)
	assert.Equal(t, ErrorCode("insufficient_scopes"), ErrorCodeInsufficientScopes)
}

func TestError_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Error
		verify func(t *testing.T, err *Error)
	}{
		{
			name: "rate limit error with all fields",
			setup: func() *Error {
				return &Error{
					Code:       ErrorCodeRateLimited,
					Message:    "GitHub API rate limit exceeded",
					Provider:   "github",
					Details:    map[string]interface{}{"limit": 5000, "remaining": 0},
					RetryAfter: 3600,
				}
			},
			verify: func(t *testing.T, err *Error) {
				assert.Equal(t, ErrorCodeRateLimited, err.Code)
				assert.Contains(t, err.Message, "rate limit")
				assert.Equal(t, "github", err.Provider)
				assert.NotNil(t, err.Details)
				assert.Equal(t, 3600, err.RetryAfter)
				assert.True(t, IsAuthError(err))
				assert.Equal(t, ErrorCodeRateLimited, GetErrorCode(err))
			},
		},
		{
			name: "storage error with nested details",
			setup: func() *Error {
				return NewStorageError("failed to access keychain", map[string]interface{}{
					"platform": "darwin",
					"error":    "security: SecItemCopyMatching: The specified item could not be found in the keychain.",
				})
			},
			verify: func(t *testing.T, err *Error) {
				assert.Equal(t, ErrorCodeStorageError, err.Code)
				assert.Contains(t, err.Message, "keychain")
				assert.NotNil(t, err.Details)
				details, ok := err.Details.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "darwin", details["platform"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.setup()

			// Assert
			tt.verify(t, err)
		})
	}
}
