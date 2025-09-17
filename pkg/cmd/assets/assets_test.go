package assets

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdAssets(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no subcommand shows help",
			args: []string{},
			want: "Manage assets and templates for Zen CLI",
		},
		{
			name: "help flag shows help",
			args: []string{"--help"},
			want: "Manage assets and templates for Zen CLI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io := iostreams.Test()
			stdout := io.Out
			f := cmdutil.NewTestFactory(io)

			cmd := NewCmdAssets(f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(stdout)
			cmd.SetErr(stdout)

			err := cmd.Execute()

			// Both cases should show help and exit with success
			assert.NoError(t, err)

			output := stdout.(*bytes.Buffer).String()
			assert.Contains(t, output, tt.want)
		})
	}
}

func TestCmdAssetsHasSubcommands(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssets(f)

	// Check that all expected subcommands are present
	expectedSubcommands := []string{"auth", "status", "list", "info", "sync"}

	for _, expectedCmd := range expectedSubcommands {
		subCmd, _, err := cmd.Find([]string{expectedCmd})
		require.NoError(t, err, "subcommand %s should exist", expectedCmd)
		assert.Equal(t, expectedCmd, subCmd.Name(), "subcommand name should match")
	}
}

func TestCmdAssetsMetadata(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssets(f)

	// Test command metadata
	assert.Equal(t, "assets <command>", cmd.Use)
	assert.Equal(t, "Manage assets and templates", cmd.Short)
	assert.Contains(t, cmd.Long, "Manage assets and templates for Zen CLI")
	assert.Contains(t, cmd.Example, "zen assets auth github")
	assert.Equal(t, "core", cmd.GroupID)
}

func TestCmdAssetsValidation(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssets(f)
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(stdout)
	cmd.SetErr(stdout)

	err := cmd.Execute()

	// Should fail with unknown subcommand
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}
