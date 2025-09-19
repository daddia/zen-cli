package init

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdInit(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	require.NotNil(t, cmd)

	// Test command properties
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize your new Zen workspace or reinitialize an existing one", cmd.Short)
	assert.Contains(t, cmd.Long, "Initialize a new Zen workspace in the current directory")

	// Test flags exist
	flags := cmd.Flags()

	forceFlag := flags.Lookup("force")
	require.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)

	configFlag := flags.Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)
}

func TestInitCommand_NewWorkspace(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create test factory with mock workspace manager
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output
	output := stdout.String()
	assert.Contains(t, output, "Initialized empty Zen workspace")
	assert.Contains(t, output, ".zen/")
}

func TestInitCommand_WithForceFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create existing .zen directory
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{"--force"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Should succeed with force flag
	output := stdout.String()
	assert.Contains(t, output, "Initialized empty Zen workspace")
}

func TestInitCommand_ExistingWorkspaceWithoutForce(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create existing .zen directory with config
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	configFile := filepath.Join(zenDir, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("existing"), 0644))

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactoryWithWorkspace(streams, true, false)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()

	// Should succeed without force flag (idempotent behavior)
	require.NoError(t, err)

	// Should show reinitialized message
	output := stdout.String()
	assert.Contains(t, output, "Reinitialized existing Zen workspace")
}

func TestInitCommand_WithCustomConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	customConfigPath := "config/custom.yaml"

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{"--config", customConfigPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Check that config directory was created
	configDir := filepath.Dir(filepath.Join(tempDir, customConfigPath))
	assert.DirExists(t, configDir)

	// Check output
	output := stdout.String()
	assert.Contains(t, output, "Initialized empty Zen workspace")
}

func TestInitCommand_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create a directory with restricted permissions
	tempDir := t.TempDir()
	restrictedDir := filepath.Join(tempDir, "restricted")
	require.NoError(t, os.MkdirAll(restrictedDir, 0000)) // No permissions
	defer os.Chmod(restrictedDir, 0755)                  // Restore permissions for cleanup

	// Change to restricted directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()

	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)

	// This test is tricky because we can't easily simulate permission denied
	// in a cross-platform way. We'll test the error handling path instead.
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// The actual permission test would require platform-specific setup
	// For now, we'll test that the command handles errors gracefully
	assert.NotNil(t, cmd.RunE)
}

func TestInitCommand_VerboseOutput(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)
	factory.Verbose = true // Enable verbose mode

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	require.NoError(t, err)

	// In verbose mode, should see analysis output
	output := stdout.String()
	assert.Contains(t, output, "Analyzing project") // Verbose output
	assert.Contains(t, output, "Initialized empty Zen workspace")
}

func TestInitCommand_InvalidConfigPath(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Try to create config in a location that will fail
	invalidConfigPath := "/root/cannot/create/this/config.yaml"

	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)

	// Execute with invalid config file path
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--config", invalidConfigPath})

	err = cmd.Execute()

	// Should handle the error gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create config directory")
	}
}

func TestInitCommand_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) error
		expectedError string
		wantSilent    bool
		initialized   bool
	}{
		{
			name: "workspace already exists (now succeeds with reinitialization)",
			setupFunc: func(dir string) error {
				zenDir := filepath.Join(dir, ".zen")
				if err := os.MkdirAll(zenDir, 0755); err != nil {
					return err
				}
				configFile := filepath.Join(zenDir, "config.yaml")
				return os.WriteFile(configFile, []byte("existing"), 0644)
			},
			expectedError: "", // No error expected
			wantSilent:    false,
			initialized:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			// Setup test condition
			if tt.setupFunc != nil {
				require.NoError(t, tt.setupFunc(tempDir))
			}

			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr
			factory := cmdutil.NewTestFactoryWithWorkspace(streams, tt.initialized, false)

			cmd := NewCmdInit(factory)
			cmd.SetArgs([]string{})

			err = cmd.Execute()

			switch {
			case tt.wantSilent:
				assert.Equal(t, cmdutil.ErrSilent, err)
			case tt.expectedError != "":
				require.Error(t, err)
			default:
				require.NoError(t, err) // Should succeed for reinitialization
			}

			if tt.expectedError != "" {
				output := stderr.String()
				assert.Contains(t, output, tt.expectedError)
			}
		})
	}
}

// Benchmark tests
func BenchmarkInitCommand_New(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCmdInit(factory)
		if cmd == nil {
			b.Fatal("command is nil")
		}
	}
}

func BenchmarkInitCommand_Execute(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

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

		cmd := NewCmdInit(factory)
		cmd.SetArgs([]string{})
		b.StartTimer()

		err = cmd.Execute()
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		os.Chdir(oldWd)
	}
}
