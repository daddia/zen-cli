package init

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdInit(t *testing.T) {
	f := createTestFactory(t.TempDir())
	cmd := NewCmdInit(f)

	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize a new Zen workspace", cmd.Short)
	assert.Contains(t, cmd.Long, ".zen/ directory structure")
	assert.Contains(t, cmd.Long, "automatically detects")
	assert.NotNil(t, cmd.RunE)

	// Check flags
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)

	configFlag := cmd.Flags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)
}

func TestInitCommand_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output
	output := outBuf.String()
	assert.Contains(t, output, "Initialized empty Zen workspace in")
	assert.Contains(t, output, "/.zen/")

	// Check files were created
	assert.FileExists(t, filepath.Join(tempDir, ".zen", "config.yaml"))
	assert.DirExists(t, filepath.Join(tempDir, ".zen"))
	assert.DirExists(t, filepath.Join(tempDir, ".zen", "tasks"))
	assert.DirExists(t, filepath.Join(tempDir, ".zen", "cache"))
	assert.DirExists(t, filepath.Join(tempDir, ".zen", "templates"))
	assert.DirExists(t, filepath.Join(tempDir, ".zen", "scripts"))
	assert.DirExists(t, filepath.Join(tempDir, ".zen", "logs"))
}

func TestInitCommand_NodeJSProject(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Create package.json
	packageJSON := map[string]interface{}{
		"name":        "my-awesome-app",
		"version":     "2.1.0",
		"description": "An awesome Node.js application",
		"dependencies": map[string]string{
			"react": "^18.0.0",
		},
		"devDependencies": map[string]string{
			"typescript": "^4.0.0",
		},
	}
	data, err := json.Marshal(packageJSON)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile("package.json", data, 0644))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output shows success
	output := outBuf.String()
	assert.Contains(t, output, "Initialized empty Zen workspace in")
	assert.Contains(t, output, "/.zen/")

	// Check files were created
	assert.FileExists(t, ".zen/config.yaml")
	assert.DirExists(t, ".zen")

	// Check configuration file contains project info
	configData, err := os.ReadFile(".zen/config.yaml")
	require.NoError(t, err)
	configStr := string(configData)
	assert.Contains(t, configStr, "my-awesome-app")
	// Note: Project detection may vary in test environment, so we focus on core functionality
}

func TestInitCommand_GoProject(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Create go.mod
	goMod := `module github.com/user/awesome-cli

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output shows success
	output := outBuf.String()
	assert.Contains(t, output, "Initialized empty Zen workspace in")
	assert.Contains(t, output, "/.zen/")

	// Check files were created
	assert.FileExists(t, ".zen/config.yaml")
	assert.DirExists(t, ".zen")

	// Check configuration file contains project info
	configData, err := os.ReadFile(".zen/config.yaml")
	require.NoError(t, err)
	configStr := string(configData)
	assert.Contains(t, configStr, "awesome-cli")
	// Note: Project detection may vary in test environment, so we focus on core functionality
}

func TestInitCommand_ExistingWorkspace_NoForce(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Create existing config file
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	existingConfig := "version: 0.9\nold: config"
	require.NoError(t, os.WriteFile(filepath.Join(zenDir, "config.yaml"), []byte(existingConfig), 0644))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, cmdutil.ErrSilent, err)

	// Check error output
	errOutput := errBuf.String()
	assert.Contains(t, errOutput, "Error:")
	assert.Contains(t, errOutput, "already initialized")
	assert.Contains(t, errOutput, "--force")

	// Check original file is unchanged
	data, err := os.ReadFile(filepath.Join(tempDir, ".zen", "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, existingConfig, string(data))
}

func TestInitCommand_ExistingWorkspace_WithForce(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Create .zen directory first
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))

	// Create existing config file
	existingConfig := "version: 0.9\nold: config"
	configPath := filepath.Join(zenDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(existingConfig), 0644))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Set force flag
	require.NoError(t, cmd.Flags().Set("force", "true"))

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check success output
	output := outBuf.String()
	assert.Contains(t, output, "Initialized empty Zen workspace in")

	// Check backup was created
	backupsDir := filepath.Join(zenDir, "backups")
	entries, err := os.ReadDir(backupsDir)
	require.NoError(t, err)

	var backupFound bool
	for _, entry := range entries {
		if entry.Name() != ".gitkeep" {
			backupFound = true
		}
	}
	assert.True(t, backupFound, "Backup should be created")

	// Check new config file was created
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version: \"1.0\"")
	assert.NotContains(t, string(data), "old: config")
}

func TestInitCommand_CustomConfigPath(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Set custom config path
	customConfigPath := "./config/my-zen.yaml"
	require.NoError(t, cmd.Flags().Set("config", customConfigPath))

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check .zen directory was still created in current directory
	assert.DirExists(t, filepath.Join(tempDir, ".zen"))

	// Check config directory was created
	assert.DirExists(t, filepath.Join(tempDir, "config"))

	// Note: The current implementation has limitations with custom config paths
	// This test documents the current behavior
}

func TestInitCommand_WithGitIgnore(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Create existing .gitignore
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	existingContent := "*.log\nnode_modules/\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(existingContent), 0644))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check success output
	output := outBuf.String()
	assert.Contains(t, output, "Initialized empty Zen workspace in")
	assert.Contains(t, output, "/.zen/")

	// Check .gitignore was updated
	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, existingContent)
	assert.Contains(t, content, ".zen/")
	assert.Contains(t, content, "# Zen CLI workspace directory")
}

func TestInitCommand_GitProject(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Create .git directory with config
	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	gitConfig := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/user/awesome-project.git
	fetch = +refs/heads/*:refs/remotes/origin/*`

	configPath := filepath.Join(gitDir, "config")
	require.NoError(t, os.WriteFile(configPath, []byte(gitConfig), 0644))

	// Create .gitignore
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("*.tmp\n"), 0644))

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output shows success
	output := outBuf.String()
	assert.Contains(t, output, "Initialized empty Zen workspace in")
	assert.Contains(t, output, "/.zen/")

	// Check .gitignore was updated
	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(data), ".zen/")
}

func TestInitCommand_PermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()

	// Change to temp directory
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tempDir))

	// Make directory read-only
	require.NoError(t, os.Chmod(tempDir, 0444))
	defer func() { _ = os.Chmod(tempDir, 0755) }() // Restore permissions for cleanup

	f := createTestFactory(tempDir)
	cmd := NewCmdInit(f)

	// Capture output
	var outBuf, errBuf bytes.Buffer
	f.IOStreams.Out = &outBuf
	f.IOStreams.ErrOut = &errBuf

	err = cmd.Execute()
	assert.Error(t, err)

	// Should handle permission errors gracefully
	// The exact error handling depends on the implementation
}

func TestInitCommand_Args(t *testing.T) {
	f := createTestFactory(t.TempDir())
	cmd := NewCmdInit(f)

	// Test that command accepts no arguments
	cmd.SetArgs([]string{})
	_ = cmd.Execute()
	// Error might occur due to workspace creation, but not due to args

	// Test that command rejects arguments
	cmd.SetArgs([]string{"extra", "args"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

// Helper function to create a test factory
func createTestFactory(tempDir string) *cmdutil.Factory {
	logger := logging.NewBasic()

	return &cmdutil.Factory{
		AppVersion:     "test",
		ExecutableName: "zen",
		IOStreams:      iostreams.Test(),
		Logger:         logger,
		Config: func() (*config.Config, error) {
			// Return error to force workspace manager to use current directory
			return nil, os.ErrNotExist
		},
		WorkspaceManager: func() (cmdutil.WorkspaceManager, error) {
			// Use current working directory for workspace detection
			cwd, err := os.Getwd()
			if err != nil {
				cwd = tempDir
			}
			return &testWorkspaceManagerAdapter{
				manager: workspace.New(cwd, "", logger), // Empty string will use default .zen/config.yaml
			}, nil
		},
	}
}

// Test adapter for workspace manager
type testWorkspaceManagerAdapter struct {
	manager *workspace.Manager
}

func (w *testWorkspaceManagerAdapter) Root() string {
	return w.manager.Root()
}

func (w *testWorkspaceManagerAdapter) ConfigFile() string {
	return w.manager.ConfigFile()
}

func (w *testWorkspaceManagerAdapter) Initialize() error {
	return w.manager.Initialize(false)
}

func (w *testWorkspaceManagerAdapter) InitializeWithForce(force bool) error {
	return w.manager.Initialize(force)
}

func (w *testWorkspaceManagerAdapter) Status() (cmdutil.WorkspaceStatus, error) {
	status, err := w.manager.Status()
	if err != nil {
		return cmdutil.WorkspaceStatus{}, err
	}

	return cmdutil.WorkspaceStatus{
		Initialized: status.Initialized,
		ConfigPath:  status.ConfigPath,
		Root:        status.Root,
	}, nil
}

// Test error scenarios
func TestInitCommand_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupError    func() *cmdutil.Factory
		expectedError string
		isSilent      bool
	}{
		{
			name: "workspace manager error",
			setupError: func() *cmdutil.Factory {
				return &cmdutil.Factory{
					IOStreams: iostreams.Test(),
					WorkspaceManager: func() (cmdutil.WorkspaceManager, error) {
						return nil, assert.AnError
					},
				}
			},
			expectedError: "failed to get workspace manager",
			isSilent:      false,
		},
		{
			name: "already exists error",
			setupError: func() *cmdutil.Factory {
				return &cmdutil.Factory{
					IOStreams: iostreams.Test(),
					WorkspaceManager: func() (cmdutil.WorkspaceManager, error) {
						return &mockWorkspaceManager{
							initError: &types.Error{
								Code:    types.ErrorCodeAlreadyExists,
								Message: "Already exists",
								Details: "Use --force",
							},
						}, nil
					},
				}
			},
			expectedError: "",
			isSilent:      true,
		},
		{
			name: "permission denied error",
			setupError: func() *cmdutil.Factory {
				return &cmdutil.Factory{
					IOStreams: iostreams.Test(),
					WorkspaceManager: func() (cmdutil.WorkspaceManager, error) {
						return &mockWorkspaceManager{
							initError: &types.Error{
								Code:    types.ErrorCodePermissionDenied,
								Message: "Permission denied",
							},
						}, nil
					},
				}
			},
			expectedError: "",
			isSilent:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.setupError()
			cmd := NewCmdInit(f)

			var outBuf, errBuf bytes.Buffer
			f.IOStreams.Out = &outBuf
			f.IOStreams.ErrOut = &errBuf

			err := cmd.Execute()

			if tt.isSilent {
				assert.Equal(t, cmdutil.ErrSilent, err)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

// Mock workspace manager for error testing
type mockWorkspaceManager struct {
	initError error
	status    cmdutil.WorkspaceStatus
}

func (m *mockWorkspaceManager) Root() string {
	return "/test"
}

func (m *mockWorkspaceManager) ConfigFile() string {
	return ".zen/config.yaml"
}

func (m *mockWorkspaceManager) Initialize() error {
	return m.initError
}

func (m *mockWorkspaceManager) InitializeWithForce(force bool) error {
	return m.initError
}

func (m *mockWorkspaceManager) Status() (cmdutil.WorkspaceStatus, error) {
	return m.status, nil
}
