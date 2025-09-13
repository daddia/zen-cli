package root

import (
	"testing"

	"github.com/jonathandaddia/zen/pkg/cmd/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRoot(t *testing.T) {
	f := factory.New()

	cmd, err := NewCmdRoot(f)
	require.NoError(t, err)
	require.NotNil(t, cmd)

	// Check basic properties
	assert.Equal(t, "zen", cmd.Use)
	assert.Contains(t, cmd.Short, "AI-Powered")
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)

	// Check flags
	assert.NotNil(t, cmd.PersistentFlags().Lookup("verbose"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("no-color"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("output"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("dry-run"))

	// Check command groups
	assert.NotEmpty(t, cmd.Groups())

	// Check subcommands exist
	subcommands := cmd.Commands()
	assert.Greater(t, len(subcommands), 0)

	// Check for specific commands
	hasVersion := false
	hasInit := false
	hasConfig := false
	hasStatus := false
	hasCompletion := false

	for _, subcmd := range subcommands {
		switch subcmd.Name() {
		case "version":
			hasVersion = true
		case "init":
			hasInit = true
		case "config":
			hasConfig = true
		case "status":
			hasStatus = true
		case "completion":
			hasCompletion = true
		}
	}

	assert.True(t, hasVersion, "should have version command")
	assert.True(t, hasInit, "should have init command")
	assert.True(t, hasConfig, "should have config command")
	assert.True(t, hasStatus, "should have status command")
	assert.True(t, hasCompletion, "should have completion command")
}

func TestGlobalFlags(t *testing.T) {
	tests := []struct {
		name string
		flag string
	}{
		{"verbose flag", "verbose"},
		{"no-color flag", "no-color"},
		{"output flag", "output"},
		{"config flag", "config"},
		{"dry-run flag", "dry-run"},
	}

	f := factory.New()
	cmd, err := NewCmdRoot(f)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.PersistentFlags().Lookup(tt.flag)
			assert.NotNil(t, flag, "flag %s should exist", tt.flag)
		})
	}
}

func TestConfigFlag(t *testing.T) {
	f := factory.New()
	cmd, err := NewCmdRoot(f)
	require.NoError(t, err)

	cmd.SetArgs([]string{"--config", "/path/to/config.yaml", "version"})
	_ = cmd.Execute()

	assert.Equal(t, "/path/to/config.yaml", f.ConfigFile)
}

func TestDryRunFlag(t *testing.T) {
	f := factory.New()
	cmd, err := NewCmdRoot(f)
	require.NoError(t, err)

	cmd.SetArgs([]string{"--dry-run", "version"})
	_ = cmd.Execute()

	assert.True(t, f.DryRun)
}

func TestCompletionCommand(t *testing.T) {
	f := factory.New()

	cmd := newCompletionCommand(f)

	assert.Equal(t, "completion [bash|zsh|fish|powershell]", cmd.Use)
	assert.Equal(t, "Generate shell completion scripts", cmd.Short)
	assert.Contains(t, cmd.Long, "Generate shell completion scripts for Zen CLI")
	assert.NotNil(t, cmd.RunE)

	// Test valid args
	assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, cmd.ValidArgs)
}

func TestPlaceholderCommands(t *testing.T) {
	f := factory.New()

	cmd := newPlaceholderCommand("test", "Test description", f)

	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Test description", cmd.Short)
	assert.Contains(t, cmd.Long, "planned for future implementation")
	assert.NotNil(t, cmd.Run)
}
