package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateTempWorkspace creates a temporary workspace with basic structure
func CreateTempWorkspace(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()

	// Create basic workspace structure
	zenDir := filepath.Join(tempDir, ".zen")
	err := os.MkdirAll(zenDir, 0755)
	require.NoError(t, err, "failed to create .zen directory")

	// Create basic config file
	configContent := `version: "1.0"
workspace:
  root: .
  name: test-workspace
project:
  type: unknown
  name: test-project
logging:
  level: info
  format: text
cli:
  output_format: text
  no_color: false
`
	configPath := filepath.Join(zenDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to create config file")

	return tempDir
}

// CreateProjectFiles creates common project files for testing
func CreateProjectFiles(t *testing.T, dir string, projectType string) {
	t.Helper()

	switch projectType {
	case "go":
		goMod := `module github.com/test/project

go 1.25

require (
	github.com/spf13/cobra v1.8.0
)`
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
		require.NoError(t, err, "failed to create go.mod")

	case "node":
		packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test Node.js project",
  "dependencies": {
    "react": "^18.0.0"
  }
}`
		err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err, "failed to create package.json")

	case "python":
		pyprojectToml := `[tool.poetry]
name = "test-project"
version = "0.1.0"
description = "Test Python project"

[tool.poetry.dependencies]
python = "^3.9"`
		err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(pyprojectToml), 0644)
		require.NoError(t, err, "failed to create pyproject.toml")

	case "git":
		gitDir := filepath.Join(dir, ".git")
		err := os.MkdirAll(gitDir, 0755)
		require.NoError(t, err, "failed to create .git directory")

		gitConfig := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
[remote "origin"]
	url = https://github.com/user/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`
		err = os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644)
		require.NoError(t, err, "failed to create git config")
	}
}

// CleanupWorkspace removes workspace files (for tests that don't use t.TempDir)
func CleanupWorkspace(t *testing.T, dir string) {
	t.Helper()

	if dir == "" || dir == "/" || dir == "." {
		t.Fatal("refusing to cleanup root or current directory")
	}

	err := os.RemoveAll(dir)
	if err != nil {
		t.Logf("Warning: failed to cleanup workspace %s: %v", dir, err)
	}
}
