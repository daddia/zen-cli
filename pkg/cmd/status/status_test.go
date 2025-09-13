package status

import (
	"bytes"
	"encoding/json"
	"runtime"
	"testing"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/jonathandaddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdStatus(t *testing.T) {
	f := &cmdutil.Factory{
		IOStreams: iostreams.Test(),
		Config: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
		WorkspaceManager: func() (cmdutil.WorkspaceManager, error) {
			return &mockWorkspaceManager{}, nil
		},
	}

	cmd := NewCmdStatus(f)

	assert.Equal(t, "status", cmd.Use)
	assert.Equal(t, "Display workspace and system status", cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestStatusOutput(t *testing.T) {
	f := &cmdutil.Factory{
		IOStreams: iostreams.Test(),
		Config: func() (*config.Config, error) {
			return &config.Config{
				LogLevel: "debug",
			}, nil
		},
		WorkspaceManager: func() (cmdutil.WorkspaceManager, error) {
			return &mockWorkspaceManager{
				initialized: true,
				root:        "/test/workspace",
				configFile:  "zen.yaml",
			}, nil
		},
	}

	t.Run("text output", func(t *testing.T) {
		out := &bytes.Buffer{}
		f.IOStreams = &iostreams.IOStreams{
			Out: out,
		}

		cmd := NewCmdStatus(f)
		err := cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "Zen CLI Status")
		assert.Contains(t, output, "Workspace:")
		assert.Contains(t, output, "Configuration:")
		assert.Contains(t, output, "System:")
		assert.Contains(t, output, "Integrations:")
		assert.Contains(t, output, runtime.GOOS)
		assert.Contains(t, output, runtime.Version())
	})

	t.Run("json output", func(t *testing.T) {
		out := &bytes.Buffer{}
		f.IOStreams = &iostreams.IOStreams{
			Out: out,
		}

		// Create root command with output flag
		rootCmd := &cobra.Command{Use: "test"}
		rootCmd.PersistentFlags().String("output", "text", "")

		cmd := NewCmdStatus(f)
		rootCmd.AddCommand(cmd)

		// Set args to trigger json output
		rootCmd.SetArgs([]string{"status", "--output", "json"})
		rootCmd.PersistentFlags().Set("output", "json")

		err := rootCmd.Execute()
		require.NoError(t, err)

		var status Status
		err = json.Unmarshal(out.Bytes(), &status)
		require.NoError(t, err)

		assert.True(t, status.Workspace.Initialized)
		assert.Equal(t, "/test/workspace", status.Workspace.Path)
		assert.True(t, status.Configuration.Loaded)
		assert.Equal(t, "debug", status.Configuration.LogLevel)
		assert.Equal(t, runtime.GOOS, status.System.OS)
	})
}

func TestGetStatusIcon(t *testing.T) {
	assert.Equal(t, "✅ Ready", getStatusIcon(true))
	assert.Equal(t, "❌ Not Ready", getStatusIcon(false))
}

func TestDisplayTextStatus(t *testing.T) {
	buf := &bytes.Buffer{}
	status := Status{
		Workspace: WorkspaceStatus{
			Initialized: true,
			Path:        "/home/user/project",
			ConfigFile:  "zen.yaml",
		},
		Configuration: ConfigStatus{
			Loaded:   true,
			Source:   "zen.yaml",
			LogLevel: "info",
		},
		System: SystemStatus{
			OS:           "linux",
			Architecture: "amd64",
			GoVersion:    "go1.21",
			NumCPU:       8,
		},
		Integrations: IntegrationStatus{
			Available: []string{"jira", "git"},
			Active:    []string{"git"},
		},
	}

	err := displayTextStatus(buf, status)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Zen CLI Status")
	assert.Contains(t, output, "✅ Ready")
	assert.Contains(t, output, "/home/user/project")
	assert.Contains(t, output, "linux")
	assert.Contains(t, output, "8")
}

// mockWorkspaceManager is a mock implementation for testing
type mockWorkspaceManager struct {
	initialized bool
	root        string
	configFile  string
}

func (m *mockWorkspaceManager) Root() string {
	return m.root
}

func (m *mockWorkspaceManager) ConfigFile() string {
	return m.configFile
}

func (m *mockWorkspaceManager) Initialize() error {
	return nil
}

func (m *mockWorkspaceManager) Status() (cmdutil.WorkspaceStatus, error) {
	return cmdutil.WorkspaceStatus{
		Initialized: m.initialized,
		ConfigPath:  m.configFile,
		Root:        m.root,
	}, nil
}
