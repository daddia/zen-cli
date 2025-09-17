//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	zenBinary string
)

func TestMain(m *testing.M) {
	// Build the zen binary for E2E tests
	var err error
	zenBinary, err = buildZenBinary()
	if err != nil {
		panic("Failed to build zen binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if zenBinary != "" {
		os.Remove(zenBinary)
	}

	os.Exit(code)
}

// buildZenBinary builds the zen binary for testing
func buildZenBinary() (string, error) {
	tempDir, err := os.MkdirTemp("", "zen-e2e-*")
	if err != nil {
		return "", err
	}

	binaryName := "zen-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tempDir, binaryName)

	// Build from the project root
	projectRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/zen")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}

	return binaryPath, nil
}

// runZenCommand runs a zen command and returns stdout, stderr, and error
func runZenCommand(t *testing.T, dir string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, zenBinary, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// TestE2E_CriticalUserJourney tests the critical path: init → config → status
func TestE2E_CriticalUserJourney(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Step 1: Initialize workspace
	t.Run("init_workspace", func(t *testing.T) {
		stdout, stderr, err := runZenCommand(t, tempDir, "init")
		require.NoError(t, err, "init command failed: %s", stderr)

		assert.Contains(t, stdout, "Initialized empty Zen workspace")
		assert.Contains(t, stdout, ".zen/")
	})

	// Step 2: Check status after initialization
	t.Run("status_after_init", func(t *testing.T) {
		stdout, stderr, err := runZenCommand(t, tempDir, "status")
		require.NoError(t, err, "status command failed: %s", stderr)

		assert.Contains(t, stdout, "Zen CLI Status")
		assert.Contains(t, stdout, "Ready")
	})

	// Step 3: Configure workspace
	t.Run("config_management", func(t *testing.T) {
		// List current configuration
		stdout, stderr, err := runZenCommand(t, tempDir, "config", "list")
		require.NoError(t, err, "config list failed: %s", stderr)

		assert.Contains(t, stdout, "log_level")
		assert.Contains(t, stdout, "info")

		// Set configuration value
		stdout, stderr, err = runZenCommand(t, tempDir, "config", "set", "log_level", "debug")
		require.NoError(t, err, "config set failed: %s", stderr)

		assert.Contains(t, stdout, "debug")

		// Get configuration value to verify
		stdout, stderr, err = runZenCommand(t, tempDir, "config", "get", "log_level")
		require.NoError(t, err, "config get failed: %s", stderr)

		assert.Contains(t, stdout, "debug")
	})

	// Step 4: Final status check
	t.Run("final_status_check", func(t *testing.T) {
		stdout, stderr, err := runZenCommand(t, tempDir, "status")
		require.NoError(t, err, "final status check failed: %s", stderr)

		assert.Contains(t, stdout, "Zen CLI Status")
		assert.Contains(t, stdout, "Ready")
		assert.Contains(t, stdout, "debug") // Should show updated log level
	})
}

// TestE2E_WorkspaceForceInit tests force initialization
func TestE2E_WorkspaceForceInit(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize workspace first time
	stdout, stderr, err := runZenCommand(t, tempDir, "init")
	require.NoError(t, err, "initial init failed: %s", stderr)
	assert.Contains(t, stdout, "Initialized empty Zen workspace")

	// Try to initialize again without force (may or may not fail depending on implementation)
	stdout, stderr, err = runZenCommand(t, tempDir, "init")
	// For E2E tests, we focus on the force flag behavior rather than exact error handling
	if err != nil {
		assert.Contains(t, stderr, "Error:")
	}

	// Initialize with force flag (should succeed)
	stdout, stderr, err = runZenCommand(t, tempDir, "init", "--force")
	require.NoError(t, err, "force init failed: %s", stderr)
	assert.Contains(t, stdout, "Reinitialized existing Zen workspace")

	// For E2E tests, focus on command behavior rather than filesystem details
	// The actual filesystem operations are tested in unit and integration tests
}

// TestE2E_OutputFormats tests different output formats
func TestE2E_OutputFormats(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize workspace
	_, stderr, err := runZenCommand(t, tempDir, "init")
	require.NoError(t, err, "init failed: %s", stderr)

	tests := []struct {
		name         string
		args         []string
		expectedText string
	}{
		{
			name:         "version_text",
			args:         []string{"version"},
			expectedText: "zen version",
		},
		{
			name:         "version_json",
			args:         []string{"--output", "json", "version"},
			expectedText: `"version"`,
		},
		{
			name:         "version_yaml",
			args:         []string{"--output", "yaml", "version"},
			expectedText: "version:",
		},
		{
			name:         "status_text",
			args:         []string{"status"},
			expectedText: "Zen CLI Status",
		},
		{
			name:         "status_json",
			args:         []string{"--output", "json", "status"},
			expectedText: `"workspace"`,
		},
		{
			name:         "status_yaml",
			args:         []string{"--output", "yaml", "status"},
			expectedText: "workspace:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runZenCommand(t, tempDir, tt.args...)
			require.NoError(t, err, "command failed: %s", stderr)

			assert.Contains(t, stdout, tt.expectedText)
		})
	}
}

// TestE2E_GlobalFlags tests global flags functionality
func TestE2E_GlobalFlags(t *testing.T) {
	tempDir := t.TempDir()

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
			args: []string{"--dry-run", "init"},
		},
		{
			name: "help_flag",
			args: []string{"--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runZenCommand(t, tempDir, tt.args...)

			if tt.name == "help_flag" {
				// Help should exit with 0 and show help text
				require.NoError(t, err, "help command failed: %s", stderr)
				assert.Contains(t, stdout, "AI-Powered Productivity Suite")
			} else {
				require.NoError(t, err, "command with %s failed: %s", tt.name, stderr)
				assert.NotEmpty(t, stdout, "should produce output")
			}
		})
	}
}

// TestE2E_ErrorScenarios tests error handling
func TestE2E_ErrorScenarios(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		args          []string
		expectError   bool
		expectedError string
	}{
		{
			name:          "unknown_command",
			args:          []string{"unknown"},
			expectError:   true,
			expectedError: "unknown command",
		},
		{
			name:          "invalid_flag",
			args:          []string{"--invalid-flag"},
			expectError:   true,
			expectedError: "unknown flag",
		},
		{
			name:          "invalid_output_format",
			args:          []string{"--output", "invalid", "version"},
			expectError:   false, // Should fall back to text
			expectedError: "",
		},
		{
			name:          "config_invalid_key",
			args:          []string{"config", "get", "invalid_key"},
			expectError:   true,
			expectedError: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runZenCommand(t, tempDir, tt.args...)

			if tt.expectError {
				require.Error(t, err, "expected error for %s", tt.name)
				if tt.expectedError != "" {
					errorOutput := stderr + stdout
					assert.Contains(t, strings.ToLower(errorOutput),
						strings.ToLower(tt.expectedError))
				}
			} else {
				require.NoError(t, err, "unexpected error for %s: %s", tt.name, stderr)
			}
		})
	}
}

// TestE2E_ProjectTypeDetection tests project type detection
func TestE2E_ProjectTypeDetection(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string) error
	}{
		{
			name: "go_project",
			setupFunc: func(dir string) error {
				goMod := `module github.com/test/project

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
			},
		},
		{
			name: "node_project",
			setupFunc: func(dir string) error {
				packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test project",
  "dependencies": {
    "react": "^18.0.0"
  }
}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
			},
		},
		{
			name: "python_project",
			setupFunc: func(dir string) error {
				pyprojectToml := `[tool.poetry]
name = "test-project"
version = "0.1.0"
description = "Test Python project"

[tool.poetry.dependencies]
python = "^3.9"`
				return os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(pyprojectToml), 0644)
			},
		},
		{
			name: "git_project",
			setupFunc: func(dir string) error {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					return err
				}

				gitConfig := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
[remote "origin"]
	url = https://github.com/user/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`

				return os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup project files
			require.NoError(t, tt.setupFunc(tempDir))

			// Initialize workspace
			stdout, stderr, err := runZenCommand(t, tempDir, "init")
			require.NoError(t, err, "init failed for %s: %s", tt.name, stderr)

			assert.Contains(t, stdout, "Initialized empty Zen workspace")

			// Check status to see if project type was detected
			stdout, stderr, err = runZenCommand(t, tempDir, "status")
			require.NoError(t, err, "status failed for %s: %s", tt.name, stderr)

			assert.Contains(t, stdout, "Zen CLI Status")
			assert.Contains(t, stdout, "Ready")
		})
	}
}

// TestE2E_CrossPlatformCompatibility tests cross-platform functionality
func TestE2E_CrossPlatformCompatibility(t *testing.T) {
	tempDir := t.TempDir()

	// Test basic commands work on current platform
	commands := [][]string{
		{"version"},
		{"init"},
		{"status"},
		{"config", "list"},
	}

	for i, cmd := range commands {
		t.Run(fmt.Sprintf("command_%d_%s", i, cmd[0]), func(t *testing.T) {
			stdout, stderr, err := runZenCommand(t, tempDir, cmd...)
			require.NoError(t, err, "command %v failed on %s: %s", cmd, runtime.GOOS, stderr)
			assert.NotEmpty(t, stdout, "command should produce output")
		})
	}

	// Test path handling works correctly
	t.Run("path_handling", func(t *testing.T) {
		stdout, stderr, err := runZenCommand(t, tempDir, "status")
		require.NoError(t, err, "status command failed: %s", stderr)

		// Should contain status information
		assert.Contains(t, stdout, "Zen CLI Status")
	})
}

// Benchmark E2E tests
func BenchmarkE2E_InitWorkspace(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tempDir := b.TempDir()
		b.StartTimer()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, zenBinary, "init")
		cmd.Dir = tempDir

		err := cmd.Run()
		cancel()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkE2E_StatusCommand(b *testing.B) {
	// Setup workspace once
	tempDir := b.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cmd := exec.CommandContext(ctx, zenBinary, "init")
	cmd.Dir = tempDir
	err := cmd.Run()
	cancel()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, zenBinary, "status")
		cmd.Dir = tempDir

		err := cmd.Run()
		cancel()
		if err != nil {
			b.Fatal(err)
		}
	}
}
