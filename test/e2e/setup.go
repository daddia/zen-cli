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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	// TestDirName is the name of the test directory created outside zen-cli
	TestDirName = "zen-test"
)

// CommandResult holds the result of running a command
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// TestEnvironment holds the test environment setup
type TestEnvironment struct {
	ZenBinary   string
	TestRootDir string
	ProjectRoot string
}

// SetupTestEnvironment creates the test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// Get the project root (zen-cli directory)
	projectRoot, err := getProjectRoot()
	require.NoError(t, err, "Failed to get project root")

	// Create zen-test directory outside of zen-cli
	parentDir := filepath.Dir(projectRoot)
	testRootDir := filepath.Join(parentDir, TestDirName)

	// Clean up any existing test directory
	if err := os.RemoveAll(testRootDir); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to clean existing test directory: %v", err)
	}

	// Create fresh test directory
	err = os.MkdirAll(testRootDir, 0755)
	require.NoError(t, err, "Failed to create test root directory")

	// Build zen binary
	zenBinary, err := buildZenBinary(projectRoot)
	require.NoError(t, err, "Failed to build zen binary")

	env := &TestEnvironment{
		ZenBinary:   zenBinary,
		TestRootDir: testRootDir,
		ProjectRoot: projectRoot,
	}

	t.Logf("Test environment setup complete:")
	t.Logf("  Project Root: %s", env.ProjectRoot)
	t.Logf("  Test Root: %s", env.TestRootDir)
	t.Logf("  Zen Binary: %s", env.ZenBinary)

	return env
}

// TeardownTestEnvironment cleans up the test environment
func TeardownTestEnvironment(t *testing.T, env *TestEnvironment) {
	if env == nil {
		return
	}

	// Clean up zen binary
	if env.ZenBinary != "" {
		if err := os.Remove(env.ZenBinary); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to remove zen binary: %v", err)
		}
	}

	// Clean up test directory
	if env.TestRootDir != "" {
		if err := os.RemoveAll(env.TestRootDir); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to remove test directory: %v", err)
		}
	}

	t.Logf("Test environment cleanup complete")
}

// CreateTestWorkspace creates a new test workspace directory
func (env *TestEnvironment) CreateTestWorkspace(t *testing.T, name string) string {
	workspaceDir := filepath.Join(env.TestRootDir, name)
	err := os.MkdirAll(workspaceDir, 0755)
	require.NoError(t, err, "Failed to create test workspace: %s", name)
	return workspaceDir
}

// CleanupTestWorkspace removes a test workspace directory
func (env *TestEnvironment) CleanupTestWorkspace(t *testing.T, workspaceDir string) {
	if err := os.RemoveAll(workspaceDir); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to cleanup workspace %s: %v", workspaceDir, err)
	}
}

// getProjectRoot finds the zen-cli project root directory
func getProjectRoot() (string, error) {
	// Start from current directory and walk up to find go.mod
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := currentDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root directory
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (go.mod not found)")
}

// buildZenBinary builds the zen binary for testing
func buildZenBinary(projectRoot string) (string, error) {
	tempDir, err := os.MkdirTemp("", "zen-e2e-binary-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	binaryName := "zen-e2e"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tempDir, binaryName)

	// Build from the project root
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/zen")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}

	return binaryPath, nil
}

// RunZenCommand runs a zen command in the specified directory
func (env *TestEnvironment) RunZenCommand(t *testing.T, workDir string, args ...string) *CommandResult {
	return env.RunZenCommandWithTimeout(t, workDir, 30*time.Second, args...)
}

// RunZenCommandWithTimeout runs a zen command with a custom timeout
func (env *TestEnvironment) RunZenCommandWithTimeout(t *testing.T, workDir string, timeout time.Duration, args ...string) *CommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, env.ZenBinary, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Error:    err,
	}

	t.Logf("Command: zen %v", args)
	t.Logf("Exit Code: %d", result.ExitCode)
	if result.Stdout != "" {
		t.Logf("Stdout: %s", result.Stdout)
	}
	if result.Stderr != "" {
		t.Logf("Stderr: %s", result.Stderr)
	}

	return result
}

// RequireSuccess asserts that the command succeeded (exit code 0)
func (r *CommandResult) RequireSuccess(t *testing.T, msgAndArgs ...interface{}) {
	require.Equal(t, 0, r.ExitCode, "Command should succeed. Stderr: %s", r.Stderr)
	require.NoError(t, r.Error, msgAndArgs...)
}

// RequireError asserts that the command failed (non-zero exit code)
func (r *CommandResult) RequireError(t *testing.T, msgAndArgs ...interface{}) {
	require.NotEqual(t, 0, r.ExitCode, "Command should fail. Stdout: %s", r.Stdout)
}

// ExpectExitCode asserts the specific exit code
func (r *CommandResult) ExpectExitCode(t *testing.T, expectedCode int, msgAndArgs ...interface{}) {
	require.Equal(t, expectedCode, r.ExitCode, msgAndArgs...)
}
