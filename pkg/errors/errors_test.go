package errors

import (
	"errors"
	"testing"

	"github.com/daddia/zen/pkg/types"
)

func TestNew(t *testing.T) {
	err := New("test error")
	if err == nil {
		t.Error("New() returned nil")
	}

	if err.Error() != "test error" {
		t.Errorf("New() error message = %q, want %q", err.Error(), "test error")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "wrapped")

	if wrappedErr == nil {
		t.Error("Wrap() returned nil")
	}

	expected := "wrapped: original error"
	if wrappedErr.Error() != expected {
		t.Errorf("Wrap() error message = %q, want %q", wrappedErr.Error(), expected)
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrapf(originalErr, "wrapped with %s", "context")

	if wrappedErr == nil {
		t.Error("Wrapf() returned nil")
	}

	expected := "wrapped with context: original error"
	if wrappedErr.Error() != expected {
		t.Errorf("Wrapf() error message = %q, want %q", wrappedErr.Error(), expected)
	}
}

func TestNewWithCode(t *testing.T) {
	err := NewWithCode(types.ErrorCodeInvalidInput, "invalid input")

	if err == nil {
		t.Error("NewWithCode() returned nil")
	}

	zenErr, ok := err.(*types.Error)
	if !ok {
		t.Error("NewWithCode() did not return *types.Error")
	}

	if zenErr.Code != types.ErrorCodeInvalidInput {
		t.Errorf("Code = %q, want %q", zenErr.Code, types.ErrorCodeInvalidInput)
	}

	if zenErr.Message != "invalid input" {
		t.Errorf("Message = %q, want %q", zenErr.Message, "invalid input")
	}
}

func TestNewWithCodef(t *testing.T) {
	err := NewWithCodef(types.ErrorCodeNotFound, "user %d not found", 123)

	if err == nil {
		t.Error("NewWithCodef() returned nil")
	}

	zenErr, ok := err.(*types.Error)
	if !ok {
		t.Error("NewWithCodef() did not return *types.Error")
	}

	if zenErr.Code != types.ErrorCodeNotFound {
		t.Errorf("Code = %q, want %q", zenErr.Code, types.ErrorCodeNotFound)
	}

	expected := "user 123 not found"
	if zenErr.Message != expected {
		t.Errorf("Message = %q, want %q", zenErr.Message, expected)
	}
}

func TestNewWithDetails(t *testing.T) {
	err := NewWithDetails(types.ErrorCodeInvalidConfig, "invalid config", "missing field 'name'")

	if err == nil {
		t.Error("NewWithDetails() returned nil")
	}

	zenErr, ok := err.(*types.Error)
	if !ok {
		t.Error("NewWithDetails() did not return *types.Error")
	}

	if zenErr.Code != types.ErrorCodeInvalidConfig {
		t.Errorf("Code = %q, want %q", zenErr.Code, types.ErrorCodeInvalidConfig)
	}

	if zenErr.Message != "invalid config" {
		t.Errorf("Message = %q, want %q", zenErr.Message, "invalid config")
	}

	if zenErr.Details != "missing field 'name'" {
		t.Errorf("Details = %q, want %q", zenErr.Details, "missing field 'name'")
	}
}

func TestIsCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     types.ErrorCode
		expected bool
	}{
		{
			name:     "matching code",
			err:      NewWithCode(types.ErrorCodeInvalidInput, "test"),
			code:     types.ErrorCodeInvalidInput,
			expected: true,
		},
		{
			name:     "non-matching code",
			err:      NewWithCode(types.ErrorCodeInvalidInput, "test"),
			code:     types.ErrorCodeNotFound,
			expected: false,
		},
		{
			name:     "non-zen error",
			err:      errors.New("regular error"),
			code:     types.ErrorCodeInvalidInput,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			code:     types.ErrorCodeInvalidInput,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCode(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("IsCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected types.ErrorCode
	}{
		{
			name:     "zen error",
			err:      NewWithCode(types.ErrorCodeInvalidInput, "test"),
			expected: types.ErrorCodeInvalidInput,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: types.ErrorCodeUnknown,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: types.ErrorCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCode(tt.err)
			if result != tt.expected {
				t.Errorf("GetCode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() error
		expectedCode types.ErrorCode
	}{
		{
			name:         "ErrInvalidInput",
			constructor:  func() error { return ErrInvalidInput("test input") },
			expectedCode: types.ErrorCodeInvalidInput,
		},
		{
			name:         "ErrNotFound",
			constructor:  func() error { return ErrNotFound("user") },
			expectedCode: types.ErrorCodeNotFound,
		},
		{
			name:         "ErrAlreadyExists",
			constructor:  func() error { return ErrAlreadyExists("user") },
			expectedCode: types.ErrorCodeAlreadyExists,
		},
		{
			name:         "ErrInvalidConfig",
			constructor:  func() error { return ErrInvalidConfig("missing field") },
			expectedCode: types.ErrorCodeInvalidConfig,
		},
		{
			name:         "ErrConfigNotFound",
			constructor:  func() error { return ErrConfigNotFound("/path/to/config") },
			expectedCode: types.ErrorCodeConfigNotFound,
		},
		{
			name:         "ErrWorkspaceNotInitialized",
			constructor:  ErrWorkspaceNotInitialized,
			expectedCode: types.ErrorCodeWorkspaceNotInit,
		},
		{
			name:         "ErrInvalidWorkspace",
			constructor:  func() error { return ErrInvalidWorkspace("invalid structure") },
			expectedCode: types.ErrorCodeInvalidWorkspace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err == nil {
				t.Error("Constructor returned nil")
				return
			}

			code := GetCode(err)
			if code != tt.expectedCode {
				t.Errorf("Error code = %q, want %q", code, tt.expectedCode)
			}

			if err.Error() == "" {
				t.Error("Error message is empty")
			}
		})
	}
}

func TestErrNotFoundMessage(t *testing.T) {
	err := ErrNotFound("user")
	expected := "user not found"

	if err.Error() != expected {
		t.Errorf("ErrNotFound() message = %q, want %q", err.Error(), expected)
	}
}

func TestErrAlreadyExistsMessage(t *testing.T) {
	err := ErrAlreadyExists("user")
	expected := "user already exists"

	if err.Error() != expected {
		t.Errorf("ErrAlreadyExists() message = %q, want %q", err.Error(), expected)
	}
}

func TestErrInvalidConfigWithDetails(t *testing.T) {
	details := "missing required field 'name'"
	err := ErrInvalidConfig(details)

	zenErr, ok := err.(*types.Error)
	if !ok {
		t.Error("ErrInvalidConfig() did not return *types.Error")
		return
	}

	if zenErr.Details != details {
		t.Errorf("Details = %q, want %q", zenErr.Details, details)
	}

	expectedMessage := "invalid configuration: " + details
	if err.Error() != expectedMessage {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMessage)
	}
}

func TestErrConfigNotFoundWithPath(t *testing.T) {
	path := "/path/to/config.yaml"
	err := ErrConfigNotFound(path)

	zenErr, ok := err.(*types.Error)
	if !ok {
		t.Error("ErrConfigNotFound() did not return *types.Error")
		return
	}

	if zenErr.Details != path {
		t.Errorf("Details = %q, want %q", zenErr.Details, path)
	}
}
