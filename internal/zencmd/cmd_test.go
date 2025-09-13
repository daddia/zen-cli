package zencmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

// Mock IOStreams for testing
type mockIOStreams struct{}

func (m *mockIOStreams) FormatError(text string) string {
	return "✗ " + text
}

func TestMain_Success(t *testing.T) {
	// This test would require more complex mocking to actually test Main()
	// For now, we'll test the individual components
	assert.True(t, true) // Placeholder
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode cmdutil.ExitCode
	}{
		{
			name:         "silent error",
			err:          cmdutil.ErrSilent,
			expectedCode: cmdutil.ExitError,
		},
		{
			name:         "pending error",
			err:          cmdutil.ErrPending,
			expectedCode: cmdutil.ExitError,
		},
		{
			name:         "user cancellation",
			err:          errors.New("cancelled"),
			expectedCode: cmdutil.ExitCancel,
		},
		{
			name:         "no results error",
			err:          cmdutil.NoResultsError{Message: "no items found"},
			expectedCode: cmdutil.ExitOK,
		},
		{
			name:         "flag error",
			err:          &cmdutil.FlagError{Err: errors.New("invalid flag")},
			expectedCode: cmdutil.ExitError,
		},
		{
			name:         "general error",
			err:          errors.New("something went wrong"),
			expectedCode: cmdutil.ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock factory with basic IOStreams
			f := &cmdutil.Factory{
				IOStreams: iostreams.Test(),
			}

			code := handleError(tt.err, f)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "simple error",
			err:      errors.New("test error"),
			expected: "✗ test error",
		},
		{
			name:     "error with suggestion",
			err:      errors.New("command not found"),
			expected: "✗ command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			mockStreams := &mockIOStreams{}

			printError(buf, tt.err, mockStreams)

			if tt.expected == "" {
				assert.Empty(t, buf.String())
			} else {
				assert.Contains(t, buf.String(), tt.expected)
			}
		})
	}
}

func TestGetErrorSuggestion(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "unknown command",
			err:      errors.New("unknown command \"foo\" for \"zen\""),
			expected: "Use 'zen --help' to see available commands",
		},
		{
			name:     "workspace not initialized",
			err:      errors.New("workspace not initialized"),
			expected: "Use 'zen --help' for usage information or check the documentation at https://zen.dev/docs",
		},
		{
			name:     "config not found",
			err:      errors.New("config file not found"),
			expected: "Run 'zen config' to check your configuration or 'zen init' to initialize a workspace",
		},
		{
			name:     "permission denied",
			err:      errors.New("permission denied"),
			expected: "Check file permissions or try running with appropriate privileges",
		},
		{
			name:     "generic error",
			err:      errors.New("some other error"),
			expected: "Use 'zen --help' for usage information or check the documentation at https://zen.dev/docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorSuggestion(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUserCancellation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "cancelled error",
			err:      errors.New("cancelled"),
			expected: true,
		},
		{
			name:     "interrupted error",
			err:      errors.New("interrupted"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("something else"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdutil.IsUserCancellation(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
