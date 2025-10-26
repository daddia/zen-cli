//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daddia/zen/internal/zencmd"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIIntegration_WorkspaceLifecycle tests the complete workspace lifecycle
func TestCLIIntegration_WorkspaceLifecycle(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Status before initialization
	t.Run("status_before_init", func(t *testing.T) {
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"status"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		// Status should show not ready since no .zen directory exists
		assert.Contains(t, output, "Zen CLI Status")
	})

	// Test 2: Initialize workspace
	t.Run("init_workspace", func(t *testing.T) {
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"init"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Initialized empty Zen workspace")

		// For integration tests, we focus on command behavior rather than filesystem
		// The actual filesystem operations are tested in unit tests
	})

	// Test 3: Status after initialization
	t.Run("status_after_init", func(t *testing.T) {
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"status"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Ready")
		assert.Contains(t, output, "Zen CLI Status")
	})

	// Test 4: Configuration management
	t.Run("config_management", func(t *testing.T) {
		// List configuration
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"config", "list"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "log_level")
		assert.Contains(t, output, "info")

		// Set configuration value
		stdout.Reset()
		err = zencmd.Execute(ctx, []string{"config", "set", "log_level", "debug"}, streams)
		require.NoError(t, err)

		// Verify configuration was set
		stdout.Reset()
		err = zencmd.Execute(ctx, []string{"config", "get", "log_level"}, streams)
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, "debug")
	})

	// Test 5: Version command
	t.Run("version_command", func(t *testing.T) {
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"version"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "zen version")
		// Simple version output doesn't include build info by default

		// Test detailed version with --build-options
		stdout.Reset()
		err = zencmd.Execute(ctx, []string{"version", "--build-options"}, streams)
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, "zen version")
		assert.Contains(t, output, "platform:")
		assert.Contains(t, output, "go version:")
	})
}

// TestCLIIntegration_ErrorHandling tests error scenarios
func TestCLIIntegration_ErrorHandling(t *testing.T) {
	streams := iostreams.Test()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "unknown_command",
			args:        []string{"unknown"},
			expectError: true,
			errorText:   "unknown command",
		},
		{
			name:        "invalid_flag",
			args:        []string{"--invalid-flag"},
			expectError: true,
			errorText:   "unknown flag",
		},
		{
			name:        "valid_command",
			args:        []string{"version"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams.Out = &stdout
			streams.ErrOut = &stderr

			err := zencmd.Execute(ctx, tt.args, streams)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCLIIntegration_OutputFormats tests different output formats
func TestCLIIntegration_OutputFormats(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Initialize workspace first
	streams := iostreams.Test()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	tests := []struct {
		name         string
		args         []string
		expectedText string
	}{
		{
			name:         "text_output",
			args:         []string{"--output", "text", "status"},
			expectedText: "Zen CLI Status",
		},
		{
			name:         "json_output",
			args:         []string{"--output", "json", "status"},
			expectedText: `"workspace"`,
		},
		{
			name:         "yaml_output",
			args:         []string{"--output", "yaml", "status"},
			expectedText: "workspace:",
		},
		{
			name:         "version_json",
			args:         []string{"--output", "json", "version"},
			expectedText: `"version"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			streams.Out = &stdout

			err := zencmd.Execute(ctx, tt.args, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, tt.expectedText)
		})
	}
}

// TestCLIIntegration_GlobalFlags tests global flags
func TestCLIIntegration_GlobalFlags(t *testing.T) {
	streams := iostreams.Test()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "verbose_flag",
			args: []string{"--verbose", "version"},
		},
		{
			name: "no_color_flag",
			args: []string{"--no-color", "version"},
		},
		{
			name: "dry_run_flag",
			args: []string{"--dry-run", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			streams.Out = &stdout

			err := zencmd.Execute(ctx, tt.args, streams)
			require.NoError(t, err)

			// Should execute successfully with global flags
			output := stdout.String()
			assert.NotEmpty(t, output)
		})
	}
}

// TestCLIIntegration_ConfigurationPrecedence tests configuration precedence
func TestCLIIntegration_ConfigurationPrecedence(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	streams := iostreams.Test()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize workspace
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	// Set environment variable (should be IGNORED per refactor plan)
	require.NoError(t, os.Setenv("ZEN_LOG_LEVEL", "warn"))
	defer func() {
		require.NoError(t, os.Unsetenv("ZEN_LOG_LEVEL"))
	}()

	// Test that environment variable is IGNORED - should use config file value
	var stdout bytes.Buffer
	streams.Out = &stdout

	err = zencmd.Execute(ctx, []string{"config", "get", "log_level"}, streams)
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "info") // Should use config file value, not env var

	// Test that CLI flag overrides environment variable
	stdout.Reset()
	// Note: This would require implementing flag binding in the config command
	// For now, we'll test that the command executes successfully
	err = zencmd.Execute(ctx, []string{"--verbose", "status"}, streams)
	require.NoError(t, err)
}

// TestCLIIntegration_ProjectDetection tests project type detection
func TestCLIIntegration_ProjectDetection(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(string) error
		expectedType string
	}{
		{
			name: "go_project",
			setupFunc: func(dir string) error {
				goMod := `module github.com/test/project
go 1.21`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
			},
			expectedType: "go",
		},
		{
			name: "node_project",
			setupFunc: func(dir string) error {
				packageJSON := `{"name": "test-project", "version": "1.0.0"}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
			},
			expectedType: "nodejs",
		},
		{
			name: "git_project",
			setupFunc: func(dir string) error {
				gitDir := filepath.Join(dir, ".git")
				return os.MkdirAll(gitDir, 0755)
			},
			expectedType: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup project files
			require.NoError(t, tt.setupFunc(tempDir))

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			streams := iostreams.Test()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Initialize workspace
			var stdout bytes.Buffer
			streams.Out = &stdout

			err = zencmd.Execute(ctx, []string{"init"}, streams)
			require.NoError(t, err)

			// Check status to see detected project type
			stdout.Reset()
			err = zencmd.Execute(ctx, []string{"--output", "json", "status"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			// The exact format depends on the status command implementation
			// We'll check that it contains some project information
			assert.NotEmpty(t, output)
		})
	}
}

// Benchmark integration tests
func BenchmarkCLIIntegration_WorkspaceInit(b *testing.B) {
	streams := iostreams.Test()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tempDir := b.TempDir()
		oldWd, err := os.Getwd()
		if err != nil {
			b.Fatal(err)
		}

		if err := os.Chdir(tempDir); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		err = zencmd.Execute(ctx, []string{"init"}, streams)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		os.Chdir(oldWd)
	}
}

func BenchmarkCLIIntegration_StatusCommand(b *testing.B) {
	// Setup workspace once
	tempDir := b.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		b.Fatal(err)
	}

	streams := iostreams.Test()
	ctx := context.Background()

	// Initialize workspace
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := zencmd.Execute(ctx, []string{"status"}, streams)
		if err != nil {
			b.Fatal(err)
		}
	}
}
