package zencmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/daddia/zen/pkg/cmd/factory"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/stretchr/testify/assert"
)

func TestGetErrorSuggestion(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "config not found error",
			err:      errors.New("config file not found"),
			expected: "Run 'zen config' to check your configuration or 'zen init' to initialize a workspace",
		},
		{
			name:     "config invalid error",
			err:      errors.New("invalid config syntax"),
			expected: "Check your configuration file syntax with 'zen config validate'",
		},
		{
			name:     "workspace not found error",
			err:      errors.New("workspace not found"),
			expected: "Run 'zen init' to initialize a new workspace in this directory",
		},
		{
			name:     "unknown flag error",
			err:      errors.New("unknown flag --invalid"),
			expected: "Use 'zen --help' to see available flags and options",
		},
		{
			name:     "unknown command error",
			err:      errors.New("unknown command 'invalid'"),
			expected: "Use 'zen --help' to see available commands",
		},
		{
			name:     "permission error",
			err:      errors.New("permission denied"),
			expected: "Check file permissions or try running with appropriate privileges",
		},
		{
			name:     "network error",
			err:      errors.New("network connection failed"),
			expected: "Check your internet connection and try again",
		},
		{
			name:     "timeout error",
			err:      errors.New("operation timeout"),
			expected: "The operation timed out. Try again or check network connectivity",
		},
		{
			name:     "authentication error",
			err:      errors.New("authentication failed"),
			expected: "Check your credentials or run authentication setup",
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			expected: "Use 'zen --help' for usage information or check the documentation at https://zen.dev/docs",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorSuggestion(tt.err)
			assert.Equal(t, tt.expected, result)
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
			name:     "error with suggestion",
			err:      errors.New("config not found"),
			expected: "Error: config not found\n\nRun 'zen config' to check your configuration or 'zen init' to initialize a workspace\n",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			printError(buf, tt.err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestHandleError(t *testing.T) {
	f := factory.New()

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
			err:          cmdutil.NoResultsError{Message: "no results"},
			expectedCode: cmdutil.ExitOK,
		},
		{
			name:         "flag error",
			err:          &cmdutil.FlagError{Err: errors.New("invalid flag")},
			expectedCode: cmdutil.ExitError,
		},
		{
			name:         "generic error",
			err:          errors.New("something went wrong"),
			expectedCode: cmdutil.ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handleError(tt.err, f)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
