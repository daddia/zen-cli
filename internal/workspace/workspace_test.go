package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNew(t *testing.T) {
	logger := logging.NewBasic()
	manager := New("/test/root", "zen.yaml", logger)

	assert.Equal(t, "/test/root", manager.Root())
	assert.Equal(t, "/test/root/.zen/config.yaml", filepath.ToSlash(manager.ConfigFile()))
	assert.Equal(t, "/test/root/.zen", filepath.ToSlash(manager.ZenDirectory()))
}

func TestDetectProject_Empty(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	info := manager.DetectProject()
	assert.Equal(t, ProjectTypeUnknown, info.Type)
	assert.Equal(t, filepath.Base(tempDir), info.Name)
}

func TestDetectProject_Git(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create git config with GitHub remote
	gitConfig := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://github.com/user/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`

	configPath := filepath.Join(gitDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(gitConfig), 0644))

	info := manager.DetectProject()
	assert.Contains(t, []ProjectType{ProjectTypeGit, ProjectTypeUnknown}, info.Type)
	assert.Equal(t, "github", info.Metadata["git_provider"])
}

func TestDetectProject_NodeJS(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create package.json
	packageJSON := map[string]interface{}{
		"name":        "test-project",
		"version":     "1.0.0",
		"description": "A test project",
		"dependencies": map[string]string{
			"react": "^18.0.0",
		},
		"devDependencies": map[string]string{
			"typescript": "^4.0.0",
		},
	}

	data, err := json.Marshal(packageJSON)
	require.NoError(t, err)

	packagePath := filepath.Join(tempDir, "package.json")
	require.NoError(t, os.WriteFile(packagePath, data, 0644))

	info := manager.DetectProject()
	assert.Equal(t, ProjectTypeNodeJS, info.Type)
	assert.Equal(t, "test-project", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "A test project", info.Description)
	assert.Equal(t, "typescript", info.Language)
	assert.Equal(t, "react", info.Framework)
}

func TestDetectProject_Go(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create go.mod
	goMod := `module github.com/user/test-project

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)`

	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte(goMod), 0644))

	info := manager.DetectProject()
	assert.Equal(t, ProjectTypeGo, info.Type)
	assert.Equal(t, "test-project", info.Name)
	assert.Equal(t, "1.21", info.Version)
	assert.Equal(t, "go", info.Language)
	assert.Equal(t, "github.com/user/test-project", info.Metadata["go_module"])
	assert.Equal(t, "1.21", info.Metadata["go_version"])
}

func TestDetectProject_Python(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create pyproject.toml
	pyprojectToml := `[tool.poetry]
name = "test-project"
version = "0.1.0"
description = "A test Python project"

[tool.poetry.dependencies]
python = "^3.9"`

	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	require.NoError(t, os.WriteFile(pyprojectPath, []byte(pyprojectToml), 0644))

	info := manager.DetectProject()
	assert.Equal(t, ProjectTypePython, info.Type)
	assert.Equal(t, "test-project", info.Name)
	assert.Equal(t, "0.1.0", info.Version)
	assert.Equal(t, "A test Python project", info.Description)
	assert.Equal(t, "python", info.Language)
	assert.Equal(t, "pyproject.toml", info.Metadata["python_project_file"])
}

func TestDetectProject_Rust(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create Cargo.toml
	cargoToml := `[package]
name = "test-project"
version = "0.1.0"
description = "A test Rust project"
edition = "2021"`

	cargoPath := filepath.Join(tempDir, "Cargo.toml")
	require.NoError(t, os.WriteFile(cargoPath, []byte(cargoToml), 0644))

	info := manager.DetectProject()
	assert.Equal(t, ProjectTypeRust, info.Type)
	assert.Equal(t, "test-project", info.Name)
	assert.Equal(t, "0.1.0", info.Version)
	assert.Equal(t, "A test Rust project", info.Description)
	assert.Equal(t, "rust", info.Language)
}

func TestDetectProject_Java_Maven(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create pom.xml
	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>
</project>`

	pomPath := filepath.Join(tempDir, "pom.xml")
	require.NoError(t, os.WriteFile(pomPath, []byte(pomXML), 0644))

	info := manager.DetectProject()
	assert.Equal(t, ProjectTypeJava, info.Type)
	assert.Equal(t, "java", info.Language)
	assert.Equal(t, "maven", info.Framework)
	assert.Equal(t, "maven", info.Metadata["build_tool"])
}

func TestInitialize_NewWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "", logger)

	err := manager.Initialize(false)
	require.NoError(t, err)

	// Check .zen directory was created with simplified structure
	zenDir := filepath.Join(tempDir, ".zen")
	assert.DirExists(t, zenDir)

	// Check essential directories exist
	assert.DirExists(t, filepath.Join(zenDir, "library"))
	assert.DirExists(t, filepath.Join(zenDir, "cache"))
	assert.DirExists(t, filepath.Join(zenDir, "logs"))
	assert.DirExists(t, filepath.Join(zenDir, "work"))
	assert.DirExists(t, filepath.Join(zenDir, "metadata"))

	// Check that old directories are not created
	assert.NoDirExists(t, filepath.Join(zenDir, "tasks"))     // Now under work/
	assert.NoDirExists(t, filepath.Join(zenDir, "templates")) // Not created by default
	assert.NoDirExists(t, filepath.Join(zenDir, "scripts"))   // Not created by default

	// Check config file was created
	configFile := manager.ConfigFile()
	assert.FileExists(t, configFile)

	// Parse and validate config
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)

	var config WorkspaceConfig
	err = yaml.Unmarshal(data, &config)
	require.NoError(t, err)

	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, tempDir, config.Workspace.Root)
	assert.Equal(t, filepath.Base(tempDir), config.Workspace.Name)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
	assert.Equal(t, "text", config.CLI.OutputFormat)
	assert.False(t, config.CLI.NoColor)
}

func TestInitialize_ExistingWorkspace_NoForce(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "", logger)

	// Create .zen directory and existing config file
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	configFile := manager.ConfigFile()
	require.NoError(t, os.WriteFile(configFile, []byte("existing"), 0644))

	err := manager.Initialize(false)
	require.NoError(t, err) // Should succeed (idempotent behavior)

	// Check that config file was reinitialized with new content
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version: \"1.0\"") // New config should be written
}

func TestInitialize_ExistingWorkspace_WithForce(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "", logger)

	// Create .zen directory first
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))

	// Create existing config file
	configFile := manager.ConfigFile()
	existingConfig := "version: 0.9\nold: config"
	require.NoError(t, os.WriteFile(configFile, []byte(existingConfig), 0644))

	err := manager.Initialize(true)
	require.NoError(t, err)

	// Check backup was created
	backupsDir := filepath.Join(zenDir, "backups")
	entries, err := os.ReadDir(backupsDir)
	require.NoError(t, err)

	var backupFound bool
	for _, entry := range entries {
		if entry.Name() != ".gitkeep" {
			backupFound = true
			// Check backup content
			backupPath := filepath.Join(backupsDir, entry.Name())
			backupData, err := os.ReadFile(backupPath)
			require.NoError(t, err)
			assert.Equal(t, existingConfig, string(backupData))
		}
	}
	assert.True(t, backupFound, "Backup file should be created")

	// Check new config file was created
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version: \"1.0\"")
}

func TestInitialize_WithGitIgnore(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	configFile := filepath.Join(tempDir, "zen.yaml")
	manager := New(tempDir, configFile, logger)

	// Create existing .gitignore
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	existingContent := "*.log\nnode_modules/\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(existingContent), 0644))

	err := manager.Initialize(false)
	require.NoError(t, err)

	// Check .gitignore was updated
	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, existingContent)
	assert.Contains(t, content, ".zen/")
	assert.Contains(t, content, "# Zen CLI workspace directory")
}

func TestInitialize_GitIgnoreAlreadyHasZen(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	configFile := filepath.Join(tempDir, "zen.yaml")
	manager := New(tempDir, configFile, logger)

	// Create .gitignore that already has .zen
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	existingContent := "*.log\n.zen/\nnode_modules/\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(existingContent), 0644))

	err := manager.Initialize(false)
	require.NoError(t, err)

	// Check .gitignore was not modified
	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(data))
}

func TestStatus_NotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "", logger)

	status, err := manager.Status()
	require.NoError(t, err)

	assert.False(t, status.Initialized)
	assert.Equal(t, tempDir, status.Root)
	assert.Equal(t, manager.ConfigFile(), status.ConfigPath)
	assert.Equal(t, filepath.Join(tempDir, ".zen"), status.ZenDirectory)
	assert.Equal(t, ProjectTypeUnknown, status.Project.Type)
}

func TestStatus_Initialized(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "", logger)

	// Initialize workspace
	require.NoError(t, manager.Initialize(false))

	status, err := manager.Status()
	require.NoError(t, err)

	assert.True(t, status.Initialized)
	assert.Equal(t, tempDir, status.Root)
	assert.Equal(t, manager.ConfigFile(), status.ConfigPath)
	assert.Equal(t, filepath.Join(tempDir, ".zen"), status.ZenDirectory)
}

func TestStatus_WithProjectDetection(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	configFile := filepath.Join(tempDir, "zen.yaml")
	manager := New(tempDir, configFile, logger)

	// Create a Go project
	goMod := `module github.com/test/project

go 1.21`
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte(goMod), 0644))

	status, err := manager.Status()
	require.NoError(t, err)

	assert.Equal(t, ProjectTypeGo, status.Project.Type)
	assert.Equal(t, "project", status.Project.Name)
	assert.Equal(t, "go", status.Project.Language)
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	projectInfo := ProjectInfo{
		Type:        ProjectTypeGo,
		Name:        "test-project",
		Description: "A test project",
		Language:    "go",
		Version:     "1.21",
	}

	config := manager.createDefaultConfig(projectInfo)

	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, tempDir, config.Workspace.Root)
	assert.Equal(t, "test-project", config.Workspace.Name)
	assert.Equal(t, "A test project", config.Workspace.Description)
	assert.Equal(t, projectInfo, config.Project)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
	assert.Equal(t, "text", config.CLI.OutputFormat)
	assert.False(t, config.CLI.NoColor)

	// Check timestamps are set
	assert.False(t, config.Workspace.CreatedAt.IsZero())
	assert.False(t, config.Workspace.UpdatedAt.IsZero())
	assert.True(t, config.Workspace.CreatedAt.Equal(config.Workspace.UpdatedAt))
}

func TestProjectTypePriority(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create both Git and Go project files
	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	goMod := `module github.com/test/project
go 1.21`
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte(goMod), 0644))

	info := manager.DetectProject()

	// Go should take priority over Git
	assert.Equal(t, ProjectTypeGo, info.Type)
	assert.Equal(t, "go", info.Language)
	assert.Equal(t, "project", info.Name)
}

// Benchmark tests
func BenchmarkDetectProject(b *testing.B) {
	tempDir := b.TempDir()
	logger := logging.NewBasic()
	manager := New(tempDir, "zen.yaml", logger)

	// Create a complex project with multiple indicators
	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(b, os.MkdirAll(gitDir, 0755))

	packageJSON := map[string]interface{}{
		"name":    "test-project",
		"version": "1.0.0",
		"dependencies": map[string]string{
			"react": "^18.0.0",
		},
	}
	data, err := json.Marshal(packageJSON)
	require.NoError(b, err)
	packagePath := filepath.Join(tempDir, "package.json")
	require.NoError(b, os.WriteFile(packagePath, data, 0644))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.DetectProject()
	}
}

func BenchmarkInitialize(b *testing.B) {
	logger := logging.NewBasic()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tempDir := b.TempDir()
		configFile := filepath.Join(tempDir, "zen.yaml")
		manager := New(tempDir, configFile, logger)
		b.StartTimer()

		err := manager.Initialize(false)
		require.NoError(b, err)
	}
}
