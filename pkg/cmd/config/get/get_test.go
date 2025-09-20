package get

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRun(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		config         *config.Config
		expectedOutput string
		expectedError  string
	}{
		{
			name: "get log_level",
			key:  "log_level",
			config: &config.Config{
				LogLevel: "debug",
			},
			expectedOutput: "debug\n",
		},
		{
			name: "get cli.verbose",
			key:  "cli.verbose",
			config: &config.Config{
				CLI: config.CLIConfig{
					Verbose: true,
				},
			},
			expectedOutput: "true\n",
		},
		{
			name: "get workspace.root",
			key:  "workspace.root",
			config: &config.Config{
				Workspace: config.WorkspaceConfig{
					Root: "/custom/workspace",
				},
			},
			expectedOutput: "/custom/workspace\n",
		},
		{
			name:          "invalid key",
			key:           "invalid.key",
			config:        &config.Config{},
			expectedError: "unknown configuration key: invalid.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios := iostreams.Test()
			out := ios.Out.(*bytes.Buffer)

			opts := &GetOptions{
				IO:  ios,
				Key: tt.key,
				Config: func() (*config.Config, error) {
					return tt.config, nil
				},
			}

			err := getRun(opts)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, out.String())
			}
		})
	}
}

func TestGetRun_ConfigError(t *testing.T) {
	ios := iostreams.Test()

	opts := &GetOptions{
		IO:  ios,
		Key: "log_level",
		Config: func() (*config.Config, error) {
			return nil, assert.AnError
		},
	}

	err := getRun(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestNewCmdConfigGet(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdConfigGet(factory, nil)

	require.NotNil(t, cmd)
	assert.Equal(t, "get <key>", cmd.Use)
	assert.Equal(t, "Print the value of a given configuration key", cmd.Short)
	assert.Contains(t, cmd.Long, "Print the value of a configuration key")
	assert.Contains(t, cmd.Example, "zen config get log_level")

	// Test that it requires exactly one argument
	assert.NotNil(t, cmd.Args)
}

func TestNewCmdConfigGet_WithRunFunc(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Custom run function for testing
	var capturedOpts *GetOptions
	runFunc := func(opts *GetOptions) error {
		capturedOpts = opts
		return nil
	}

	cmd := NewCmdConfigGet(factory, runFunc)
	cmd.SetArgs([]string{"log_level"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify the options were passed correctly
	require.NotNil(t, capturedOpts)
	assert.Equal(t, "log_level", capturedOpts.Key)
	assert.NotNil(t, capturedOpts.IO)
	assert.NotNil(t, capturedOpts.Config)
}

func TestNewCmdConfigGet_InvalidArgs(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdConfigGet(factory, nil)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")

	// Test with too many arguments
	cmd.SetArgs([]string{"key1", "key2"})
	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 2")
}
