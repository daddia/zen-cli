//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestE2E_CriticalUserJourney tests the critical path: init → config → status
func TestE2E_CriticalUserJourney(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	// Create workspace for this test
	workspaceDir := env.CreateTestWorkspace(t, "critical-journey-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	// Step 1: Initialize workspace
	t.Run("init_workspace", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "init")
		result.RequireSuccess(t, "init command should succeed")

		assert.Contains(t, result.Stdout, "Initialized empty Zen workspace")
		assert.Contains(t, result.Stdout, ".zen/")
	})

	// Step 2: Check status after initialization
	t.Run("status_after_init", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "status")
		result.RequireSuccess(t, "status command should succeed")

		assert.Contains(t, result.Stdout, "Zen CLI Status")
		assert.Contains(t, result.Stdout, "Ready")
	})

	// Step 3: Configure workspace
	t.Run("config_management", func(t *testing.T) {
		// List current configuration
		result := env.RunZenCommand(t, workspaceDir, "config", "list")
		result.RequireSuccess(t, "config list should succeed")

		assert.Contains(t, result.Stdout, "log_level")
		assert.Contains(t, result.Stdout, "info")

		// Set configuration value
		result = env.RunZenCommand(t, workspaceDir, "config", "set", "log_level", "debug")
		result.RequireSuccess(t, "config set should succeed")

		assert.Contains(t, result.Stdout, "debug")

		// Get configuration value to verify
		result = env.RunZenCommand(t, workspaceDir, "config", "get", "log_level")
		result.RequireSuccess(t, "config get should succeed")

		assert.Contains(t, result.Stdout, "debug")
	})

	// Step 4: Final status check
	t.Run("final_status_check", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "status")
		result.RequireSuccess(t, "final status check should succeed")

		assert.Contains(t, result.Stdout, "Zen CLI Status")
		assert.Contains(t, result.Stdout, "Ready")
		assert.Contains(t, result.Stdout, "debug") // Should show updated log level
	})
}

// TestE2E_WorkspaceForceInit tests force initialization
func TestE2E_WorkspaceForceInit(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	workspaceDir := env.CreateTestWorkspace(t, "force-init-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	// Initialize workspace first time
	result := env.RunZenCommand(t, workspaceDir, "init")
	result.RequireSuccess(t, "initial init should succeed")
	assert.Contains(t, result.Stdout, "Initialized empty Zen workspace")

	// Try to initialize again without force (should be idempotent)
	result = env.RunZenCommand(t, workspaceDir, "init")
	result.RequireSuccess(t, "second init should be idempotent")

	// Initialize with force flag (should succeed)
	result = env.RunZenCommand(t, workspaceDir, "init", "--force")
	result.RequireSuccess(t, "force init should succeed")
	assert.Contains(t, result.Stdout, "Reinitialized")
}

// TestE2E_OutputFormats tests different output formats
func TestE2E_OutputFormats(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	workspaceDir := env.CreateTestWorkspace(t, "output-formats-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	// Initialize workspace
	result := env.RunZenCommand(t, workspaceDir, "init")
	result.RequireSuccess(t, "init should succeed")

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
			result := env.RunZenCommand(t, workspaceDir, tt.args...)
			result.RequireSuccess(t, "command %v should succeed", tt.args)

			assert.Contains(t, result.Stdout, tt.expectedText)
		})
	}
}

// TestE2E_GlobalFlags tests global flags functionality
func TestE2E_GlobalFlags(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	workspaceDir := env.CreateTestWorkspace(t, "global-flags-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

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
			result := env.RunZenCommand(t, workspaceDir, tt.args...)

			if tt.name == "help_flag" {
				// Help should exit with 0 and show help text
				result.RequireSuccess(t, "help command should succeed")
				assert.Contains(t, result.Stdout, "Zen. The unified control plane for product & engineering.")
			} else {
				result.RequireSuccess(t, "command with %s should succeed", tt.name)
				assert.NotEmpty(t, result.Stdout, "should produce output")
			}
		})
	}
}

// TestE2E_ErrorScenarios tests error handling
func TestE2E_ErrorScenarios(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	workspaceDir := env.CreateTestWorkspace(t, "error-scenarios-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := env.RunZenCommand(t, workspaceDir, tt.args...)

			if tt.expectError {
				result.RequireError(t, "expected error for %s", tt.name)
				if tt.expectedError != "" {
					errorOutput := result.Stderr + result.Stdout
					assert.Contains(t, strings.ToLower(errorOutput),
						strings.ToLower(tt.expectedError))
				}
			} else {
				result.RequireSuccess(t, "unexpected error for %s", tt.name)
			}
		})
	}
}

// Note: Project type detection and cross-platform compatibility tests
// are covered in the new workspace_test.go file with better structure
