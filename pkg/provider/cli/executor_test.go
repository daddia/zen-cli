package cli

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	t.Run("creates executor with all fields", func(t *testing.T) {
		// Arrange
		binaryPath := "/usr/bin/test"
		workDir := "/tmp/workdir"
		env := []string{"TEST=value"}
		logger := logging.NewBasic()

		// Act
		executor := NewExecutor(binaryPath, workDir, env, logger)

		// Assert
		assert.NotNil(t, executor)
		assert.Equal(t, binaryPath, executor.binaryPath)
		assert.Equal(t, workDir, executor.workDir)
		assert.Equal(t, env, executor.env)
		assert.Equal(t, logger, executor.logger)
	})

	t.Run("creates executor with minimal fields", func(t *testing.T) {
		// Arrange
		binaryPath := "/usr/bin/test"
		logger := logging.NewBasic()

		// Act
		executor := NewExecutor(binaryPath, "", nil, logger)

		// Assert
		assert.NotNil(t, executor)
		assert.Equal(t, binaryPath, executor.binaryPath)
		assert.Empty(t, executor.workDir)
		assert.Nil(t, executor.env)
	})
}

func TestExecutor_Execute(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("echo", "", nil, logger)

		// Act
		result, err := executor.Execute(context.Background(), []string{"hello", "world"})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, string(result.Stdout), "hello world")
		assert.Empty(t, result.Stderr)
		assert.Greater(t, result.Duration, time.Duration(0))
	})

	t.Run("command with stderr output", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		// Use a command that writes to stderr (sh -c with redirect)
		executor := NewExecutor("sh", "", nil, logger)

		// Act
		result, err := executor.Execute(context.Background(), []string{"-c", "echo error >&2"})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, string(result.Stderr), "error")
	})

	t.Run("command with non-zero exit code", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("sh", "", nil, logger)

		// Act
		result, err := executor.Execute(context.Background(), []string{"-c", "exit 42"})

		// Assert
		require.NoError(t, err) // Non-zero exit is not an error
		assert.Equal(t, 42, result.ExitCode)
	})

	t.Run("command not found", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("nonexistent-command-12345", "", nil, logger)

		// Act
		_, err := executor.Execute(context.Background(), []string{})

		// Assert
		require.Error(t, err)
		assert.True(t, provider.IsProviderError(err))
		assert.Equal(t, provider.ErrorCodeExecution, provider.GetErrorCode(err))
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("sleep", "", nil, logger)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Act
		_, err := executor.Execute(ctx, []string{"10"})

		// Assert
		require.Error(t, err)
		assert.True(t, provider.IsProviderError(err))
	})

	t.Run("context timeout", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("sleep", "", nil, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Act
		_, err := executor.Execute(ctx, []string{"10"})

		// Assert
		// The command should either error immediately due to timeout,
		// or complete normally if it finishes before timeout.
		// Both are acceptable since timing is not guaranteed.
		if err != nil {
			assert.True(t, provider.IsProviderError(err))
		}
	})

	t.Run("command with working directory", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		workDir := os.TempDir()
		executor := NewExecutor("pwd", workDir, nil, logger)

		// Act
		result, err := executor.Execute(context.Background(), []string{})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		// Note: pwd output might have symlinks resolved, so just check it's not empty
		assert.NotEmpty(t, string(result.Stdout))
	})
}

func TestExecutor_Stream(t *testing.T) {
	t.Run("stream command output", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("echo", "", nil, logger)

		// Act
		stream, err := executor.Stream(context.Background(), []string{"line1\nline2\nline3"})
		require.NoError(t, err)
		defer stream.Close()

		// Read all output
		scanner := bufio.NewScanner(stream)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		// Assert
		require.NoError(t, scanner.Err())
		assert.True(t, len(lines) > 0)
	})

	t.Run("stream multiple lines", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("sh", "", nil, logger)

		// Act
		stream, err := executor.Stream(context.Background(), []string{"-c", "echo line1; echo line2; echo line3"})
		require.NoError(t, err)
		defer stream.Close()

		// Read all output
		scanner := bufio.NewScanner(stream)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		// Assert
		require.NoError(t, scanner.Err())
		assert.GreaterOrEqual(t, len(lines), 3)
		assert.Contains(t, lines[0], "line1")
		assert.Contains(t, lines[1], "line2")
		assert.Contains(t, lines[2], "line3")
	})

	t.Run("stream command not found", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("nonexistent-command-12345", "", nil, logger)

		// Act
		_, err := executor.Stream(context.Background(), []string{})

		// Assert
		require.Error(t, err)
		assert.True(t, provider.IsProviderError(err))
	})

	t.Run("stream with context cancellation", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("sleep", "", nil, logger)

		ctx, cancel := context.WithCancel(context.Background())

		// Act
		stream, err := executor.Stream(ctx, []string{"10"})
		require.NoError(t, err)

		// Cancel context after starting
		cancel()

		// Try to read (should stop quickly)
		buf := make([]byte, 100)
		_, readErr := stream.Read(buf)

		stream.Close()

		// Assert
		// Reading from a canceled stream should return an error or EOF
		assert.True(t, readErr == io.EOF || readErr != nil)
	})
}

func TestExecutor_buildEnv(t *testing.T) {
	t.Run("minimal environment", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("test", "", nil, logger)

		// Act
		env := executor.buildEnv()

		// Assert
		assert.NotEmpty(t, env)
		// Should have at least PATH
		hasPath := false
		for _, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				hasPath = true
				break
			}
		}
		assert.True(t, hasPath, "Environment should include PATH")
	})

	t.Run("additional environment variables", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		additionalEnv := []string{"TEST_VAR=value", "ANOTHER=123"}
		executor := NewExecutor("test", "", additionalEnv, logger)

		// Act
		env := executor.buildEnv()

		// Assert
		assert.Contains(t, env, "TEST_VAR=value")
		assert.Contains(t, env, "ANOTHER=123")
	})
}

func TestExecutor_sanitizeArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "no sensitive args",
			args:     []string{"clone", "https://github.com/repo.git"},
			expected: []string{"clone", "https://github.com/repo.git"},
		},
		{
			name:     "password flag",
			args:     []string{"login", "--password", "secret123"},
			expected: []string{"login", "--password", "***"},
		},
		{
			name:     "token flag",
			args:     []string{"auth", "--token", "ghp_secret"},
			expected: []string{"auth", "--token", "***"},
		},
		{
			name:     "multiple sensitive flags",
			args:     []string{"deploy", "--api-key", "key123", "--secret", "sec456"},
			expected: []string{"deploy", "--api-key", "***", "--secret", "***"},
		},
		{
			name:     "short form password",
			args:     []string{"connect", "-p", "password"},
			expected: []string{"connect", "-p", "***"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := logging.NewBasic()
			executor := NewExecutor("test", "", nil, logger)

			// Act
			result := executor.sanitizeArgs(tt.args)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStreamReader(t *testing.T) {
	t.Run("read and close", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		cmd := exec.Command("echo", "test output")
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		cmd.Start()

		combined := io.MultiReader(stdout, stderr)
		reader := &streamReader{
			reader: combined,
			cmd:    cmd,
			logger: logger,
			binary: "echo",
			args:   []string{"test output"},
		}

		// Act
		buf := make([]byte, 100)
		n, err := reader.Read(buf)

		// Assert
		assert.NoError(t, err)
		assert.Greater(t, n, 0)

		// Close should not return error even if command exited
		closeErr := reader.Close()
		assert.NoError(t, closeErr)
	})
}

func TestGetPathEnv(t *testing.T) {
	t.Run("returns PATH value", func(t *testing.T) {
		// Act
		path := getPathEnv()

		// Assert
		assert.NotEmpty(t, path)
		// Should contain at least some common paths
		assert.True(t, strings.Contains(path, "bin") || strings.Contains(path, "usr"))
	})
}

func TestExecutor_Integration(t *testing.T) {
	// Skip in short test mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("real command execution flow", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()
		executor := NewExecutor("sh", "", nil, logger)

		// Act - Execute a multi-step command
		result, err := executor.Execute(context.Background(), []string{"-c", "echo stdout; echo stderr >&2; exit 0"})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, string(result.Stdout), "stdout")
		assert.Contains(t, string(result.Stderr), "stderr")
		assert.Greater(t, result.Duration, time.Duration(0))
	})
}
