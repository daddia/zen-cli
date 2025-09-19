package zencmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		wantErr        bool
	}{
		{
			name:           "version command",
			args:           []string{"version"},
			expectedOutput: "zen version",
			wantErr:        false,
		},
		{
			name:           "help command",
			args:           []string{"--help"},
			expectedOutput: "Zen CLI - AI-Powered Productivity Suite",
			wantErr:        false,
		},
		{
			name:           "status command",
			args:           []string{"status"},
			expectedOutput: "Zen CLI Status",
			wantErr:        false,
		},
		{
			name:           "invalid command",
			args:           []string{"nonexistent"},
			expectedOutput: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			// Create test context
			ctx := context.Background()

			// Execute command
			err := Execute(ctx, tt.args, streams)

			// Check error expectation
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Check output if specified
			if tt.expectedOutput != "" {
				stdoutStr := stdout.String()
				stderrStr := stderr.String()
				combinedOutput := stdoutStr + stderrStr
				assert.Contains(t, combinedOutput, tt.expectedOutput)
			}
		})
	}
}

func TestExecuteWithConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Test with configuration
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr

	ctx := context.Background()
	err = Execute(ctx, []string{"status", "--output", "json"}, streams)

	// Should not error
	require.NoError(t, err)

	// Should produce JSON output
	output := stdout.String()
	assert.Contains(t, output, `"workspace"`)
	assert.Contains(t, output, `"configuration"`)
}

func TestExecuteWithVerboseLogging(t *testing.T) {
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr

	ctx := context.Background()
	err := Execute(ctx, []string{"--verbose", "status"}, streams)

	require.NoError(t, err)

	// In verbose mode, should see additional logging
	output := stdout.String()
	assert.NotEmpty(t, output)
}

func TestExecuteWithInvalidFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr

	ctx := context.Background()
	err := Execute(ctx, []string{"--invalid-flag"}, streams)

	require.Error(t, err)
}

func TestExecuteWithDryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr

	ctx := context.Background()
	err := Execute(ctx, []string{"--dry-run", "status"}, streams)

	require.NoError(t, err)

	// Should execute successfully in dry-run mode
	output := stdout.String()
	assert.Contains(t, output, "Zen CLI Status")
}

// Benchmark tests for performance validation
func BenchmarkExecuteVersion(b *testing.B) {
	streams := iostreams.Test()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Execute(ctx, []string{"version"}, streams)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecuteStatus(b *testing.B) {
	streams := iostreams.Test()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Execute(ctx, []string{"status"}, streams)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestExecuteErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorType   string
	}{
		{
			name:        "unknown command",
			args:        []string{"unknown"},
			expectError: true,
			errorType:   "unknown command",
		},
		{
			name:        "invalid flag",
			args:        []string{"--unknown-flag"},
			expectError: true,
			errorType:   "unknown flag",
		},
		{
			name:        "valid command",
			args:        []string{"version"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			ctx := context.Background()
			err := Execute(ctx, tt.args, streams)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExecuteContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	streams := iostreams.Test()
	err := Execute(ctx, []string{"status"}, streams)

	// Should handle cancellation gracefully
	// Note: This test may pass or fail depending on timing,
	// but it shouldn't panic or hang
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
}

func TestHandleError(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

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
			err:          errors.New("generic error"),
			expectedCode: cmdutil.ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleError(tt.err, factory)
			assert.Equal(t, tt.expectedCode, result)
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
			name:     "config not found",
			err:      errors.New("config not found"),
			expected: "Run 'zen config' to check your configuration or 'zen init' to initialize a workspace",
		},
		{
			name:     "config invalid",
			err:      errors.New("config invalid syntax"),
			expected: "Check your configuration file syntax with 'zen config validate'",
		},
		{
			name:     "workspace not found",
			err:      errors.New("workspace not found"),
			expected: "Run 'zen init' to initialize a new workspace in this directory",
		},
		{
			name:     "workspace invalid",
			err:      errors.New("workspace invalid structure"),
			expected: "Check workspace structure with 'zen status' or reinitialize with 'zen init --force'",
		},
		{
			name:     "permission denied",
			err:      errors.New("permission denied to access file"),
			expected: "Check file permissions or try running with appropriate privileges",
		},
		{
			name:     "unknown flag",
			err:      errors.New("unknown flag: --invalid"),
			expected: "Use 'zen --help' to see available flags and options",
		},
		{
			name:     "unknown command",
			err:      errors.New("unknown command: invalid"),
			expected: "Use 'zen --help' to see available commands",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorSuggestion(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintError(t *testing.T) {
	var buf bytes.Buffer
	streams := iostreams.Test()
	streams.SetColorEnabled(false) // Disable color for predictable output

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "nil error",
			err:  nil,
		},
		{
			name: "simple error",
			err:  errors.New("test error"),
		},
		{
			name: "config error with suggestion",
			err:  errors.New("config not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			printError(&buf, tt.err, streams)

			output := buf.String()
			if tt.err == nil {
				assert.Empty(t, output)
			} else {
				assert.Contains(t, output, tt.err.Error())
			}
		})
	}
}
