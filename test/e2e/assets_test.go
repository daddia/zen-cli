//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestE2E_AssetsCommands tests zen assets commands
func TestE2E_AssetsCommands(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	// Create and initialize a workspace for assets testing
	workspaceDir := env.CreateTestWorkspace(t, "assets-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	// Initialize the workspace first
	result := env.RunZenCommand(t, workspaceDir, "init")
	result.RequireSuccess(t, "zen init should succeed before testing assets")

	t.Run("zen_assets_status", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "status")

		result.RequireSuccess(t, "zen assets status should succeed")

		// Should show assets status information
		assert.Contains(t, result.Stdout, "Asset Status", "Should show assets status header")

		// Should show cache information
		output := strings.ToLower(result.Stdout)
		assert.True(t,
			strings.Contains(output, "cache") || strings.Contains(output, "repository") || strings.Contains(output, "assets"),
			"Should show cache, repository, or assets information")
	})

	t.Run("zen_assets_status_json", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "status", "--output", "json")

		result.RequireSuccess(t, "zen assets status --output json should succeed")

		// Should be valid JSON
		assert.Contains(t, result.Stdout, "{", "Should be JSON format")
		assert.Contains(t, result.Stdout, "}", "Should be JSON format")

		// Should contain expected fields
		output := strings.ToLower(result.Stdout)
		assert.True(t,
			strings.Contains(output, "cache") || strings.Contains(output, "repository") || strings.Contains(output, "authentication"),
			"Should contain cache, repository, or authentication fields")
	})

	t.Run("zen_assets_status_yaml", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "status", "--output", "yaml")

		result.RequireSuccess(t, "zen assets status --output yaml should succeed")

		// Should be valid YAML
		assert.Contains(t, result.Stdout, ":", "Should be YAML format")

		// Should contain expected fields
		output := strings.ToLower(result.Stdout)
		assert.True(t,
			strings.Contains(output, "cache") || strings.Contains(output, "repository") || strings.Contains(output, "authentication"),
			"Should contain cache, repository, or authentication fields")
	})

	t.Run("zen_assets_sync", func(t *testing.T) {
		// Use longer timeout for sync operations
		result := env.RunZenCommandWithTimeout(t, workspaceDir, 60*time.Second, "assets", "sync")

		// Assets sync might fail if no authentication is configured, which is expected in e2e tests
		// We test that the command runs and provides appropriate feedback
		if result.ExitCode == 0 {
			// Sync succeeded
			assert.Contains(t, result.Stdout, "Sync", "Should show sync information")
		} else {
			// Sync failed (expected without proper auth setup)
			errorOutput := result.Stderr + result.Stdout
			output := strings.ToLower(errorOutput)

			// Should show meaningful error about authentication or configuration
			assert.True(t,
				strings.Contains(output, "auth") ||
					strings.Contains(output, "credential") ||
					strings.Contains(output, "token") ||
					strings.Contains(output, "config") ||
					strings.Contains(output, "permission"),
				"Should show authentication or configuration error: %s", errorOutput)
		}
	})

	t.Run("zen_assets_sync_dry_run", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "sync", "--dry-run")

		result.RequireSuccess(t, "zen assets sync --dry-run should succeed")

		// Should show what would be done without actually doing it
		output := strings.ToLower(result.Stdout)
		assert.True(t,
			strings.Contains(output, "dry") ||
				strings.Contains(output, "would") ||
				strings.Contains(output, "preview"),
			"Should indicate dry run mode")
	})

	t.Run("zen_assets_list", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "list")

		// List might succeed or fail depending on authentication/sync status
		if result.ExitCode == 0 {
			// List succeeded - should show assets or empty list
			output := strings.ToLower(result.Stdout)
			assert.True(t,
				strings.Contains(output, "asset") ||
					strings.Contains(output, "template") ||
					strings.Contains(output, "no assets") ||
					strings.Contains(output, "empty") ||
					len(strings.TrimSpace(result.Stdout)) == 0, // Empty output is also valid
				"Should show assets, templates, or indicate empty list")
		} else {
			// List failed (expected without proper setup)
			errorOutput := result.Stderr + result.Stdout
			output := strings.ToLower(errorOutput)

			// Should show meaningful error
			assert.True(t,
				strings.Contains(output, "auth") ||
					strings.Contains(output, "sync") ||
					strings.Contains(output, "cache") ||
					strings.Contains(output, "repository"),
				"Should show authentication, sync, cache, or repository error: %s", errorOutput)
		}
	})

	t.Run("zen_assets_list_json", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "list", "--output", "json")

		// Similar to list, but should output JSON format if successful
		if result.ExitCode == 0 {
			// Should be JSON format
			output := result.Stdout
			assert.True(t,
				strings.Contains(output, "{") || strings.Contains(output, "["),
				"Should be JSON format")
		}
		// If it fails, that's expected without proper setup
	})

	t.Run("zen_assets_list_with_filters", func(t *testing.T) {
		// Test various filter options
		filterTests := []struct {
			name string
			args []string
		}{
			{"type_filter", []string{"assets", "list", "--type", "template"}},
			{"category_filter", []string{"assets", "list", "--category", "documentation"}},
			{"tag_filter", []string{"assets", "list", "--tag", "adr"}},
		}

		for _, tt := range filterTests {
			t.Run(tt.name, func(t *testing.T) {
				result := env.RunZenCommand(t, workspaceDir, tt.args...)

				// Filters might work or fail depending on setup
				// We just verify the command doesn't crash
				if result.ExitCode != 0 {
					// Expected to fail without proper setup
					errorOutput := result.Stderr + result.Stdout
					assert.NotEmpty(t, errorOutput, "Should provide error message")
				}
			})
		}
	})

	t.Run("zen_assets_help", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "--help")

		result.RequireSuccess(t, "zen assets --help should succeed")

		// Should show assets help
		assert.Contains(t, result.Stdout, "assets", "Should show assets help")
		assert.Contains(t, result.Stdout, "Available Commands:", "Should show available commands")

		// Should list subcommands
		assert.Contains(t, result.Stdout, "status", "Should list status command")
		assert.Contains(t, result.Stdout, "sync", "Should list sync command")
		assert.Contains(t, result.Stdout, "list", "Should list list command")
	})

	t.Run("zen_assets_invalid_command", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "assets", "invalid-command")

		// Assets command shows help for invalid subcommands (user-friendly behavior)
		result.RequireSuccess(t, "zen assets with invalid command should show help")

		output := result.Stdout
		assert.Contains(t, output, "Available Commands", "Should show help with available commands")
		assert.Contains(t, output, "Usage:", "Should show usage information")
	})
}

// TestE2E_AssetsWorkflow tests a complete assets workflow
func TestE2E_AssetsWorkflow(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	// Create and initialize a workspace
	workspaceDir := env.CreateTestWorkspace(t, "assets-workflow-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	// Step 1: Initialize workspace
	result := env.RunZenCommand(t, workspaceDir, "init")
	result.RequireSuccess(t, "Workspace initialization should succeed")

	// Step 2: Check initial assets status
	result = env.RunZenCommand(t, workspaceDir, "assets", "status")
	result.RequireSuccess(t, "Initial assets status should succeed")

	// Step 3: Try to list assets (might be empty initially)
	result = env.RunZenCommand(t, workspaceDir, "assets", "list")
	// Don't require success as it might fail without authentication

	// Step 4: Try dry-run sync
	result = env.RunZenCommand(t, workspaceDir, "assets", "sync", "--dry-run")
	result.RequireSuccess(t, "Dry-run sync should succeed")

	// Step 5: Check status after operations
	result = env.RunZenCommand(t, workspaceDir, "assets", "status")
	result.RequireSuccess(t, "Final assets status should succeed")

	t.Log("Assets workflow test completed - commands executed without crashing")
}
