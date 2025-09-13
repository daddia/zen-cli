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
		}
	}

	assert.True(t, hasVersion, "should have version command")
	assert.True(t, hasInit, "should have init command")
	assert.True(t, hasConfig, "should have config command")
	assert.True(t, hasStatus, "should have status command")
}

func TestPlaceholderCommands(t *testing.T) {
	f := factory.New()

	cmd := newPlaceholderCommand("test", "Test description", f)

	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Test description", cmd.Short)
	assert.Contains(t, cmd.Long, "planned for future implementation")
	assert.NotNil(t, cmd.Run)
}
