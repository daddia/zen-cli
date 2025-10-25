//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestE2E_CoreCommands tests core zen commands that should work without initialization
func TestE2E_CoreCommands(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer TeardownTestEnvironment(t, env)

	// Create a test workspace directory (but don't initialize it as zen workspace)
	workspaceDir := env.CreateTestWorkspace(t, "core-commands-test")
	defer env.CleanupTestWorkspace(t, workspaceDir)

	t.Run("zen_status_not_initialized", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "status")

		// Should fail with exit code 1 (not initialized)
		result.ExpectExitCode(t, 1, "zen status should fail when not in zen workspace")

		// Should show git-like error message
		assert.Contains(t, result.Stderr, "Not Initialized", "Should show not initialized message")
		assert.Contains(t, result.Stderr, "Not a zen workspace", "Should explain it's not a zen workspace")
		assert.Contains(t, result.Stderr, ".zen", "Should mention .zen directory")

		// Should use proper error formatting with ✗ symbol
		assert.Contains(t, result.Stderr, "✗", "Should use error symbol")
	})

	t.Run("zen_help", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "help")

		result.RequireSuccess(t, "zen help should succeed")

		// Should show main help text
		assert.Contains(t, result.Stdout, "Zen. The unified control plane for product & engineering.", "Should show main description")
		assert.Contains(t, result.Stdout, "Usage:", "Should show usage information")
		assert.Contains(t, result.Stdout, "Commands:", "Should show commands section")

		// Should include core commands
		assert.Contains(t, result.Stdout, "init", "Should list init command")
		assert.Contains(t, result.Stdout, "status", "Should list status command")
		assert.Contains(t, result.Stdout, "config", "Should list config command")
		assert.Contains(t, result.Stdout, "version", "Should list version command")
	})

	t.Run("zen_help_flag", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "--help")

		result.RequireSuccess(t, "zen --help should succeed")

		// Should show same content as help command
		assert.Contains(t, result.Stdout, "Zen. The unified control plane for product & engineering.", "Should show main description")
		assert.Contains(t, result.Stdout, "Usage:", "Should show usage information")
	})

	t.Run("zen_version", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "version")

		result.RequireSuccess(t, "zen version should succeed")

		// Should show version information
		assert.Contains(t, result.Stdout, "zen version", "Should show zen version")

		// Should include build information (even if "unknown" in dev builds)
		output := strings.ToLower(result.Stdout)
		assert.True(t,
			strings.Contains(output, "commit") || strings.Contains(output, "build") || strings.Contains(output, "date") || strings.Contains(output, "dev"),
			"Should show build information (commit, build, date, or dev)")
	})

	t.Run("zen_version_json", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "--output", "json", "version")

		result.RequireSuccess(t, "zen version --output json should succeed")

		// Should be valid JSON with version field
		assert.Contains(t, result.Stdout, `"version"`, "Should contain version field in JSON")
		assert.Contains(t, result.Stdout, `{`, "Should be JSON format")
		assert.Contains(t, result.Stdout, `}`, "Should be JSON format")
	})

	t.Run("zen_version_yaml", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "--output", "yaml", "version")

		result.RequireSuccess(t, "zen version --output yaml should succeed")

		// Should be valid YAML with version field
		assert.Contains(t, result.Stdout, "version:", "Should contain version field in YAML")
	})

	t.Run("zen_global_flags", func(t *testing.T) {
		// Test --verbose flag
		result := env.RunZenCommand(t, workspaceDir, "--verbose", "version")
		result.RequireSuccess(t, "zen --verbose version should succeed")
		assert.Contains(t, result.Stdout, "zen version", "Should show version with verbose flag")

		// Test --no-color flag
		result = env.RunZenCommand(t, workspaceDir, "--no-color", "version")
		result.RequireSuccess(t, "zen --no-color version should succeed")
		assert.Contains(t, result.Stdout, "zen version", "Should show version with no-color flag")
	})

	t.Run("zen_invalid_command", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "nonexistent-command")

		result.RequireError(t, "zen with invalid command should fail")

		// Should show error message
		errorOutput := result.Stderr + result.Stdout
		assert.Contains(t, strings.ToLower(errorOutput), "unknown command", "Should show unknown command error")
	})

	t.Run("zen_invalid_flag", func(t *testing.T) {
		result := env.RunZenCommand(t, workspaceDir, "--invalid-flag")

		result.RequireError(t, "zen with invalid flag should fail")

		// Should show error message
		errorOutput := result.Stderr + result.Stdout
		assert.Contains(t, strings.ToLower(errorOutput), "unknown flag", "Should show unknown flag error")
	})
}
