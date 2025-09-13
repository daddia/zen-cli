package zencmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/jonathandaddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected cmdutil.ExitCode
	}{
		{
			name:     "silent error",
			err:      cmdutil.SilentError,
			expected: cmdutil.ExitError,
		},
		{
			name:     "pending error",
			err:      cmdutil.PendingError,
			expected: cmdutil.ExitError,
		},
		{
			name:     "user cancellation",
			err:      errors.New("cancelled"),
			expected: cmdutil.ExitCancel,
		},
		{
			name:     "no results error",
			err:      cmdutil.NoResultsError{Message: "no items"},
			expected: cmdutil.ExitOK,
		},
		{
			name:     "flag error",
			err:      &cmdutil.FlagError{Err: errors.New("bad flag")},
			expected: cmdutil.ExitError,
		},
		{
			name:     "general error",
			err:      errors.New("something went wrong"),
			expected: cmdutil.ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{
				IOStreams: &iostreams.IOStreams{
					ErrOut: &bytes.Buffer{},
				},
			}
			result := handleError(tt.err, f)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedOutput string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedOutput: "",
		},
		{
			name:           "simple error",
			err:            errors.New("test error"),
			expectedOutput: "test error\n",
		},
		{
			name:           "config error with suggestion",
			err:            errors.New("config not found"),
			expectedOutput: "config not found\n\nTry running 'zen config' to check your configuration\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			printError(buf, tt.err)
			assert.Equal(t, tt.expectedOutput, buf.String())
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
			name:     "config error",
			err:      errors.New("config not found"),
			expected: "Try running 'zen config' to check your configuration",
		},
		{
			name:     "workspace error",
			err:      errors.New("workspace not initialized"),
			expected: "Try running 'zen init' to initialize your workspace",
		},
		{
			name:     "permission error",
			err:      errors.New("permission denied"),
			expected: "Check file permissions and try again",
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown error"),
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
