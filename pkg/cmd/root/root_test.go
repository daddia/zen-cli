package root

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRoot(t *testing.T) {
	// Create test factory
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Create root command
	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)
	require.NotNil(t, cmd)

	// Test basic command properties
	assert.Equal(t, "zen", cmd.Use)
	assert.Equal(t, "AI-Powered Productivity Suite", cmd.Short)
	assert.Contains(t, cmd.Long, "Zen CLI - AI-Powered Productivity Suite")
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)

	// Test that subcommands are added
	subcommands := cmd.Commands()
	assert.True(t, len(subcommands) > 0)

	// Check for expected subcommands
	commandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		commandNames[i] = subcmd.Name()
	}

	expectedCommands := []string{"version", "init", "config", "status", "completion"}
	for _, expected := range expectedCommands {
		assert.Contains(t, commandNames, expected, "Expected command %s not found", expected)
	}
}

func TestRootCommandFlags(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)

	// Test persistent flags exist
	flags := cmd.PersistentFlags()

	// Check verbose flag
	verboseFlag := flags.Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
	assert.Equal(t, "false", verboseFlag.DefValue)

	// Check no-color flag
	noColorFlag := flags.Lookup("no-color")
	require.NotNil(t, noColorFlag)
	assert.Equal(t, "false", noColorFlag.DefValue)

	// Check output flag
	outputFlag := flags.Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "o", outputFlag.Shorthand)
	assert.Equal(t, "text", outputFlag.DefValue)

	// Check config flag
	configFlag := flags.Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)

	// Check dry-run flag
	dryRunFlag := flags.Lookup("dry-run")
	require.NotNil(t, dryRunFlag)
	assert.Equal(t, "false", dryRunFlag.DefValue)
}

func TestRootCommandPersistentPreRunE(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		setup   func(*cmdutil.Factory)
	}{
		{
			name:    "valid execution",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "with verbose flag",
			args:    []string{"--verbose"},
			wantErr: false,
		},
		{
			name:    "with no-color flag",
			args:    []string{"--no-color"},
			wantErr: false,
		},
		{
			name:    "with output format",
			args:    []string{"--output", "json"},
			wantErr: false,
		},
		{
			name:    "with dry-run",
			args:    []string{"--dry-run"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams := iostreams.Test()
			factory := cmdutil.NewTestFactory(streams)

			if tt.setup != nil {
				tt.setup(factory)
			}

			cmd, err := NewCmdRoot(factory)
			require.NoError(t, err)

			// Set args and execute persistent pre-run
			cmd.SetArgs(tt.args)
			err = cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			if cmd.PersistentPreRunE != nil {
				err = cmd.PersistentPreRunE(cmd, []string{})
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestRootCommandHelp(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err = cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "AI-Powered Productivity Suite")
	assert.Contains(t, output, "zen init")
	assert.Contains(t, output, "zen config")
	assert.Contains(t, output, "zen status")
	assert.Contains(t, output, "Additional Commands:")
}

func TestRootCommandGroups(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)

	// Check that command groups are added
	groups := cmd.Groups()
	assert.True(t, len(groups) > 0)

	// Check for expected groups
	groupIDs := make([]string, len(groups))
	for i, group := range groups {
		groupIDs[i] = group.ID
	}

	expectedGroups := []string{"core", "product", "engineering"}
	for _, expected := range expectedGroups {
		assert.Contains(t, groupIDs, expected, "Expected group %s not found", expected)
	}
}

func TestPlaceholderCommands(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)

	// Test placeholder commands exist
	placeholderCommands := []string{"workflow", "product", "integrations", "agents"}

	for _, cmdName := range placeholderCommands {
		subcmd, _, err := cmd.Find([]string{cmdName})
		require.NoError(t, err, "Placeholder command %s not found", cmdName)
		assert.Equal(t, cmdName, subcmd.Name())

		// Test execution of placeholder command
		var buf bytes.Buffer
		subcmd.SetOut(&buf)
		subcmd.SetArgs([]string{})

		// Check if RunE exists before calling it
		if subcmd.RunE != nil {
			err = subcmd.RunE(subcmd, []string{})
			require.NoError(t, err)
		} else if subcmd.Run != nil {
			subcmd.Run(subcmd, []string{})
		}

		output := buf.String()
		assert.Contains(t, output, "planned for future implementation")
		assert.Contains(t, output, cmdName)
	}
}

func TestCompletionCommand(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)

	// Find completion command
	completionCmd, _, err := cmd.Find([]string{"completion"})
	require.NoError(t, err)
	assert.Equal(t, "completion", completionCmd.Name())

	// Test valid shells
	validShells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range validShells {
		t.Run("completion_"+shell, func(t *testing.T) {
			var buf bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &buf

			// Update factory IOStreams for this test
			testFactory := cmdutil.NewTestFactory(streams)

			// Create a fresh completion command with the test factory
			testRootCmd, err := NewCmdRoot(testFactory)
			require.NoError(t, err)

			testCompletionCmd, _, err := testRootCmd.Find([]string{"completion"})
			require.NoError(t, err)

			testCompletionCmd.SetArgs([]string{shell})

			err = testCompletionCmd.RunE(testCompletionCmd, []string{shell})
			require.NoError(t, err)

			output := buf.String()
			assert.NotEmpty(t, output, "Completion script should not be empty")
		})
	}

	// Test invalid shell
	t.Run("completion_invalid", func(t *testing.T) {
		var buf bytes.Buffer
		completionCmd.SetOut(&buf)
		completionCmd.SetArgs([]string{"invalid"})

		err := completionCmd.RunE(completionCmd, []string{"invalid"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported shell")
	})
}

func TestRootCommandFlagBinding(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	require.NoError(t, err)

	// Test that flags are properly bound
	tests := []struct {
		name     string
		flag     string
		value    string
		expected interface{}
	}{
		{
			name:     "verbose flag",
			flag:     "verbose",
			value:    "true",
			expected: true,
		},
		{
			name:     "no-color flag",
			flag:     "no-color",
			value:    "true",
			expected: true,
		},
		{
			name:     "output flag",
			flag:     "output",
			value:    "json",
			expected: "json",
		},
		{
			name:     "config flag",
			flag:     "config",
			value:    "/path/to/config",
			expected: "/path/to/config",
		},
		{
			name:     "dry-run flag",
			flag:     "dry-run",
			value:    "true",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"--" + tt.flag, tt.value}
			if tt.flag == "dry-run" || tt.flag == "verbose" || tt.flag == "no-color" {
				// Boolean flags don't need values
				args = []string{"--" + tt.flag}
			}

			cmd.SetArgs(args)
			err := cmd.ParseFlags(args)
			require.NoError(t, err)

			flag := cmd.PersistentFlags().Lookup(tt.flag)
			require.NotNil(t, flag)

			if flag.Value.Type() == "bool" {
				actual, err := cmd.PersistentFlags().GetBool(tt.flag)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			} else {
				actual, err := cmd.PersistentFlags().GetString(tt.flag)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewCmdRoot(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd, err := NewCmdRoot(factory)
		if err != nil {
			b.Fatal(err)
		}
		if cmd == nil {
			b.Fatal("command is nil")
		}
	}
}

func BenchmarkRootCommandHelp(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd, err := NewCmdRoot(factory)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}
