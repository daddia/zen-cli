package task

import (
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdTask(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdTask(factory)

	require.NotNil(t, cmd)
	assert.Equal(t, "task <command>", cmd.Use)
	assert.Equal(t, "Manage tasks and workflow", cmd.Short)
	assert.Contains(t, cmd.Long, "seven-stage Zenflow")
	assert.Contains(t, cmd.Long, "Manage tasks and workflow")
	assert.Equal(t, "core", cmd.GroupID)

	// Check that it has subcommands
	assert.True(t, cmd.HasSubCommands())

	// Check for create subcommand
	createCmd, _, err := cmd.Find([]string{"create"})
	require.NoError(t, err)
	assert.Equal(t, "create", createCmd.Name())

	// Sync subcommand is not implemented yet
	_, _, err = cmd.Find([]string{"sync"})
	assert.Error(t, err) // Should error because sync command doesn't exist

	// Main task command doesn't have flags - they're on subcommands
	// Check that main command has no flags
	assert.False(t, cmd.Flags().HasFlags())
}

func TestTaskCommandHelp(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdTask(factory)

	// Test that help can be displayed without error
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()

	// The help command should not return an error
	assert.NoError(t, err)
}
