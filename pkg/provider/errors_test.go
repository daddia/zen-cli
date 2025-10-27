package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCode_Constants(t *testing.T) {
	// Verify all error code constants are properly defined
	assert.Equal(t, ErrorCode("not_found"), ErrorCodeNotFound)
	assert.Equal(t, ErrorCode("version_mismatch"), ErrorCodeVersionMismatch)
	assert.Equal(t, ErrorCode("execution_failed"), ErrorCodeExecution)
	assert.Equal(t, ErrorCode("timeout"), ErrorCodeTimeout)
	assert.Equal(t, ErrorCode("canceled"), ErrorCodeCanceled)
	assert.Equal(t, ErrorCode("parse_failed"), ErrorCodeParse)
	assert.Equal(t, ErrorCode("invalid_operation"), ErrorCodeInvalidOp)
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		provErr  *Error
		expected string
	}{
		{
			name: "error with provider and operation",
			provErr: &Error{
				Code:     ErrorCodeExecution,
				Message:  "command failed",
				Provider: "git",
				Op:       "git.clone",
			},
			expected: "[git:git.clone:execution_failed] command failed",
		},
		{
			name: "error with provider only",
			provErr: &Error{
				Code:     ErrorCodeNotFound,
				Message:  "binary not found",
				Provider: "terraform",
			},
			expected: "[terraform:not_found] binary not found",
		},
		{
			name: "error without context",
			provErr: &Error{
				Code:    ErrorCodeTimeout,
				Message: "operation timed out",
			},
			expected: "[timeout] operation timed out",
		},
		{
			name: "error with underlying error and full context",
			provErr: &Error{
				Code:     ErrorCodeParse,
				Message:  "failed to parse output",
				Provider: "jira",
				Op:       "task.get",
				Err:      errors.New("invalid JSON"),
			},
			expected: "[jira:task.get:parse_failed] failed to parse output: invalid JSON",
		},
		{
			name: "error with underlying error and provider only",
			provErr: &Error{
				Code:     ErrorCodeExecution,
				Message:  "execution failed",
				Provider: "github",
				Err:      errors.New("network error"),
			},
			expected: "[github:execution_failed] execution failed: network error",
		},
		{
			name: "error with underlying error and no context",
			provErr: &Error{
				Code:    ErrorCodeCanceled,
				Message: "operation canceled",
				Err:     errors.New("context deadline exceeded"),
			},
			expected: "[canceled] operation canceled: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.provErr.Error()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	tests := []struct {
		name         string
		provErr      *Error
		expectUnwrap bool
	}{
		{
			name: "error with underlying error",
			provErr: &Error{
				Code:     ErrorCodeExecution,
				Message:  "command failed",
				Provider: "git",
				Err:      errors.New("exit code 1"),
			},
			expectUnwrap: true,
		},
		{
			name: "error without underlying error",
			provErr: &Error{
				Code:     ErrorCodeNotFound,
				Message:  "binary not found",
				Provider: "terraform",
			},
			expectUnwrap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			unwrapped := tt.provErr.Unwrap()

			// Assert
			if tt.expectUnwrap {
				assert.NotNil(t, unwrapped)
				assert.Equal(t, tt.provErr.Err, unwrapped)
			} else {
				assert.Nil(t, unwrapped)
			}
		})
	}
}

func TestIsProviderError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "provider error",
			err: &Error{
				Code:     ErrorCodeNotFound,
				Message:  "not found",
				Provider: "git",
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
		{
			name:     "wrapped standard error",
			err:      fmt.Errorf("wrapped: %w", errors.New("inner error")),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := IsProviderError(tt.err)

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
			name: "provider error with code",
			err: &Error{
				Code:     ErrorCodeExecution,
				Message:  "execution failed",
				Provider: "git",
			},
			expected: ErrorCodeExecution,
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

func TestNewError(t *testing.T) {
	// Arrange
	code := ErrorCodeNotFound
	message := "binary not found"
	provider := "terraform"

	// Act
	err := NewError(code, message, provider)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, provider, err.Provider)
	assert.Empty(t, err.Op)
	assert.Nil(t, err.Err)
}

func TestNewErrorWithOp(t *testing.T) {
	// Arrange
	code := ErrorCodeInvalidOp
	message := "operation not supported"
	provider := "git"
	op := "git.rebase"

	// Act
	err := NewErrorWithOp(code, message, provider, op)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Nil(t, err.Err)
}

func TestWrapError(t *testing.T) {
	// Arrange
	code := ErrorCodeExecution
	message := "command failed"
	provider := "git"
	underlyingErr := errors.New("exit code 128")

	// Act
	err := WrapError(code, message, provider, underlyingErr)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, provider, err.Provider)
	assert.Empty(t, err.Op)
	assert.Equal(t, underlyingErr, err.Err)
	assert.Equal(t, underlyingErr, err.Unwrap())
}

func TestWrapErrorWithOp(t *testing.T) {
	// Arrange
	code := ErrorCodeParse
	message := "failed to parse output"
	provider := "jira"
	op := "task.get"
	underlyingErr := errors.New("invalid JSON")

	// Act
	err := WrapErrorWithOp(code, message, provider, op, underlyingErr)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Equal(t, underlyingErr, err.Err)
	assert.True(t, errors.Is(err, underlyingErr))
}

func TestErrNotFound(t *testing.T) {
	// Arrange
	provider := "kubectl"
	resource := "binary"

	// Act
	err := ErrNotFound(provider, resource)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeNotFound, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Contains(t, err.Message, resource)
	assert.Contains(t, err.Message, "not found")
}

func TestErrVersionMismatch(t *testing.T) {
	// Arrange
	provider := "terraform"
	current := "1.2.0"
	required := ">=1.3.0"

	// Act
	err := ErrVersionMismatch(provider, current, required)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeVersionMismatch, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Contains(t, err.Message, current)
	assert.Contains(t, err.Message, required)
}

func TestErrExecution(t *testing.T) {
	// Arrange
	provider := "git"
	op := "git.clone"
	message := "clone failed"
	underlyingErr := errors.New("repository not found")

	// Act
	err := ErrExecution(provider, op, message, underlyingErr)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeExecution, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, underlyingErr, err.Err)
}

func TestErrTimeout(t *testing.T) {
	// Arrange
	provider := "jira"
	op := "task.search"

	// Act
	err := ErrTimeout(provider, op)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeTimeout, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Contains(t, err.Message, "deadline")
}

func TestErrCanceled(t *testing.T) {
	// Arrange
	provider := "github"
	op := "pull_request.list"

	// Act
	err := ErrCanceled(provider, op)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeCanceled, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Contains(t, err.Message, "canceled")
}

func TestErrParse(t *testing.T) {
	// Arrange
	provider := "jira"
	op := "task.get"
	message := "invalid JSON response"
	underlyingErr := errors.New("unexpected character")

	// Act
	err := ErrParse(provider, op, message, underlyingErr)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeParse, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, underlyingErr, err.Err)
}

func TestErrInvalidOp(t *testing.T) {
	// Arrange
	provider := "git"
	op := "git.unsupported"

	// Act
	err := ErrInvalidOp(provider, op)

	// Assert
	require.NotNil(t, err)
	assert.Equal(t, ErrorCodeInvalidOp, err.Code)
	assert.Equal(t, provider, err.Provider)
	assert.Equal(t, op, err.Op)
	assert.Contains(t, err.Message, op)
	assert.Contains(t, err.Message, "not supported")
}

func TestError_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Error
		verify func(t *testing.T, err *Error)
	}{
		{
			name: "execution error with full context",
			setup: func() *Error {
				return ErrExecution("git", "git.push", "push rejected", errors.New("non-fast-forward"))
			},
			verify: func(t *testing.T, err *Error) {
				assert.Equal(t, ErrorCodeExecution, err.Code)
				assert.Equal(t, "git", err.Provider)
				assert.Equal(t, "git.push", err.Op)
				assert.Contains(t, err.Message, "push rejected")
				assert.NotNil(t, err.Err)
				assert.True(t, IsProviderError(err))
				assert.Equal(t, ErrorCodeExecution, GetErrorCode(err))
			},
		},
		{
			name: "nested error chain",
			setup: func() *Error {
				innerErr := errors.New("connection refused")
				wrappedErr := fmt.Errorf("network error: %w", innerErr)
				return WrapErrorWithOp(ErrorCodeExecution, "API call failed", "github", "repo.list", wrappedErr)
			},
			verify: func(t *testing.T, err *Error) {
				assert.NotNil(t, err.Err)
				assert.Contains(t, err.Error(), "API call failed")
				assert.Contains(t, err.Error(), "network error")
			},
		},
		{
			name: "version mismatch with specific versions",
			setup: func() *Error {
				return ErrVersionMismatch("terraform", "1.2.9", ">=1.3.0")
			},
			verify: func(t *testing.T, err *Error) {
				assert.Equal(t, ErrorCodeVersionMismatch, err.Code)
				assert.Contains(t, err.Message, "1.2.9")
				assert.Contains(t, err.Message, ">=1.3.0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.setup()

			// Assert
			require.NotNil(t, err)
			tt.verify(t, err)
		})
	}
}

func TestError_ErrorChaining(t *testing.T) {
	t.Run("errors.Is works with wrapped errors", func(t *testing.T) {
		// Arrange
		innerErr := errors.New("connection timeout")
		provErr := WrapError(ErrorCodeExecution, "execution failed", "jira", innerErr)

		// Act & Assert
		assert.True(t, errors.Is(provErr, innerErr))
	})

	t.Run("errors.As works with provider errors", func(t *testing.T) {
		// Arrange
		var targetErr *Error
		provErr := ErrNotFound("git", "binary")

		// Act
		found := errors.As(provErr, &targetErr)

		// Assert
		assert.True(t, found)
		assert.Equal(t, ErrorCodeNotFound, targetErr.Code)
		assert.Equal(t, "git", targetErr.Provider)
	})

	t.Run("wrapped provider error in chain", func(t *testing.T) {
		// Arrange
		provErr := ErrTimeout("github", "api.request")
		wrappedErr := fmt.Errorf("operation failed: %w", provErr)

		// Act
		var targetErr *Error
		found := errors.As(wrappedErr, &targetErr)

		// Assert
		assert.True(t, found)
		assert.Equal(t, ErrorCodeTimeout, targetErr.Code)
	})
}
