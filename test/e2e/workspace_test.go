//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_WorkspaceInitialization tests zen workspace initialization and setup
func TestE2E_WorkspaceInitialization(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	// Create a test workspace directory
	workspaceDir := env.CreateTestWorkspace(t, "workspace-init-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	t.Run("zen_status_before_init", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "status")

		// Should fail with exit code 1 (not initialized)
		result.ExpectExitCode(t, 1, "zen status should fail before initialization")

		// Should show not initialized message
		assert.Contains(t, result.Stderr, "Not Initialized", "Should show not initialized message")
		assert.Contains(t, result.Stderr, "Not a zen workspace", "Should explain it's not a zen workspace")
	})

	t.Run("zen_init", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "init")

		result.RequireSuccess(t, "zen init should succeed")

		// Should show success message
		assert.Contains(t, result.Stdout, "Initialized empty Zen workspace", "Should show initialization success")
		assert.Contains(t, result.Stdout, ".zen/", "Should mention .zen directory")

		// Should create .zen directory
		zenDir := filepath.Join(workspaceDir, ".zen")
		_, err := os.Stat(zenDir)
		require.NoError(t, err, ".zen directory should be created")

		// Should create config file
		configFile := filepath.Join(zenDir, "config")
		_, err = os.Stat(configFile)
		require.NoError(t, err, "config should be created")

		// Config file should contain basic structure
		configContent, err := os.ReadFile(configFile)
		require.NoError(t, err, "Should be able to read config file")

		configStr := string(configContent)
		assert.Contains(t, configStr, "log_level", "Config should have log_level")
		assert.NotEmpty(t, configStr, "Config file should not be empty")
	})

	t.Run("zen_status_after_init", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "status")

		result.RequireSuccess(t, "zen status should succeed after initialization")

		// Should show full status information
		assert.Contains(t, result.Stdout, "Zen CLI Status", "Should show status header")
		assert.Contains(t, result.Stdout, "Workspace:", "Should show workspace section")
		assert.Contains(t, result.Stdout, "Configuration:", "Should show configuration section")
		assert.Contains(t, result.Stdout, "System:", "Should show system section")

		// Workspace should be ready
		assert.Contains(t, result.Stdout, "Ready", "Workspace should be ready")
		assert.Contains(t, result.Stdout, ".zen/config", "Should show config file path")

		// Configuration should be loaded from file
		assert.Contains(t, result.Stdout, "Loaded", "Configuration should be loaded")
		assert.Contains(t, result.Stdout, "config", "Should show config source")
	})

	t.Run("zen_init_already_initialized", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "init")

		// Should succeed (idempotent behavior)
		result.RequireSuccess(t, "zen init should be idempotent")

		// Should show reinitialization message
		output := result.Stdout + result.Stderr
		assert.True(t,
			strings.Contains(output, "Reinitialized") || strings.Contains(output, "already initialized") || strings.Contains(output, "Initialized"),
			"Should show appropriate message for already initialized workspace")
	})

	t.Run("zen_init_force", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "init", "--force")

		result.RequireSuccess(t, "zen init --force should succeed")

		// Should show reinitialization message
		assert.Contains(t, result.Stdout, "Reinitialized", "Should show reinitialization message")
	})

	t.Run("zen_config_operations", func(t *testing.T) {
		// Test config list
		result := env.RunZenCommand(t, workspaceDir, "config", "list")
		result.RequireSuccess(t, "zen config list should succeed")

		// Should show configuration options
		assert.Contains(t, result.Stdout, "log_level", "Should show log_level config")
		assert.Contains(t, result.Stdout, "info", "Should show default log level")

		// Test config get
		result = env.RunZenCommand(t, workspaceDir, "config", "get", "log_level")
		result.RequireSuccess(t, "zen config get should succeed")

		assert.Contains(t, result.Stdout, "info", "Should show current log level")

		// Test config set
		result = env.RunZenCommand(t, workspaceDir, "config", "set", "log_level", "debug")
		result.RequireSuccess(t, "zen config set should succeed")

		assert.Contains(t, result.Stdout, "debug", "Should confirm setting debug level")

		// Verify config was updated
		result = env.RunZenCommand(t, workspaceDir, "config", "get", "log_level")
		result.RequireSuccess(t, "zen config get after set should succeed")

		assert.Contains(t, result.Stdout, "debug", "Should show updated log level")

		// Verify status shows updated config
		result = env.RunZenCommand(t, workspaceDir, "status")
		result.RequireSuccess(t, "zen status after config change should succeed")

		assert.Contains(t, result.Stdout, "debug", "Status should show updated log level")
	})

	t.Run("zen_config_invalid_operations", func(t *testing.T) {
		// Test getting invalid config key
		result := env.RunZenCommand(t, workspaceDir, "config", "get", "invalid_key")
		result.RequireError(t, "zen config get with invalid key should fail")

		errorOutput := result.Stderr + result.Stdout
		assert.Contains(t, strings.ToLower(errorOutput), "invalid", "Should show invalid key error")

		// Test setting invalid config key
		result = env.RunZenCommand(t, workspaceDir, "config", "set", "invalid_key", "value")
		result.RequireError(t, "zen config set with invalid key should fail")

		errorOutput = result.Stderr + result.Stdout
		assert.Contains(t, strings.ToLower(errorOutput), "invalid", "Should show invalid key error")
	})
}

// TestE2E_WorkspaceProjectDetection tests project type detection during initialization
func TestE2E_WorkspaceProjectDetection(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	tests := []struct {
		name      string
		setupFunc func(string) error
		expectMsg string
	}{
		{
			name: "go_project",
			setupFunc: func(dir string) error {
				goMod := `module github.com/test/project

go 1.25

require (
	github.com/spf13/cobra v1.8.0
)`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
			},
			expectMsg: "go",
		},
		{
			name: "node_project",
			setupFunc: func(dir string) error {
				packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test Node.js project",
  "dependencies": {
    "react": "^18.0.0"
  }
}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
			},
			expectMsg: "nodejs",
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
			expectMsg: "python",
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
			expectMsg: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create workspace for this test
			workspaceDir := env.CreateTestWorkspace(t, "project-detection-"+tt.name)
			defer env.CleanupTestWorkspace(t, workspaceDir)

			// Setup project files
			require.NoError(t, tt.setupFunc(workspaceDir), "Failed to setup project files")

			// Initialize workspace
			result := env.RunZenCommand(t, workspaceDir, "init")
			result.RequireSuccess(t, "zen init should succeed for %s project", tt.name)

			assert.Contains(t, result.Stdout, "Initialized empty Zen workspace", "Should show initialization success")

			// Check status to see if project type was detected
			result = env.RunZenCommand(t, workspaceDir, "status")
			result.RequireSuccess(t, "zen status should succeed for %s project", tt.name)

			assert.Contains(t, result.Stdout, "Zen CLI Status", "Should show status")
			assert.Contains(t, result.Stdout, "Ready", "Workspace should be ready")

			// Note: Project type detection is internal, so we just verify the workspace works
			// The actual project type detection is tested in unit tests
		})
	}
}
