package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRun(t *testing.T) {
	ios := iostreams.Test()
	out := ios.Out.(*bytes.Buffer)

	testConfig := &config.Config{
		LogLevel:  "debug",
		LogFormat: "json",
		CLI: config.CLIConfig{
			NoColor:      true,
			Verbose:      false,
			OutputFormat: "yaml",
		},
		Workspace: config.WorkspaceConfig{
			Root:       "/test/workspace",
			ConfigFile: "test.yaml",
		},
		Development: config.DevelopmentConfig{
			Debug:   true,
			Profile: false,
		},
	}

	opts := &ListOptions{
		IO: ios,
		Config: func() (*config.Config, error) {
			return testConfig, nil
		},
	}

	err := listRun(opts)
	require.NoError(t, err)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Verify we have output for all configuration options
	assert.True(t, len(lines) >= len(config.Options), "Should have at least as many lines as config options")

	// Check for expected key-value pairs
	expectedPairs := map[string]string{
		"log_level":             "debug",
		"log_format":            "json",
		"cli.no_color":          "true",
		"cli.verbose":           "false",
		"cli.output_format":     "yaml",
		"workspace.root":        "/test/workspace",
		"workspace.config_file": "test.yaml",
		"development.debug":     "true",
		"development.profile":   "false",
	}

	for expectedKey, expectedValue := range expectedPairs {
		expectedLine := expectedKey + "=" + expectedValue
		assert.Contains(t, output, expectedLine, "Output should contain %s", expectedLine)
	}
}

func TestListRun_ConfigError(t *testing.T) {
	ios := iostreams.Test()

	opts := &ListOptions{
		IO: ios,
		Config: func() (*config.Config, error) {
			return nil, assert.AnError
		},
	}

	err := listRun(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestNewCmdConfigList(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdConfigList(factory, nil)

	require.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "Print a list of configuration keys and values", cmd.Short)
	assert.Contains(t, cmd.Long, "List all configuration keys")
	assert.Contains(t, cmd.Long, "effective configuration")

	// Test that it accepts no arguments
	assert.NotNil(t, cmd.Args)
}

func TestNewCmdConfigList_WithRunFunc(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Custom run function for testing
	var capturedOpts *ListOptions
	runFunc := func(opts *ListOptions) error {
		capturedOpts = opts
		return nil
	}

	cmd := NewCmdConfigList(factory, runFunc)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify the options were passed correctly
	require.NotNil(t, capturedOpts)
	assert.NotNil(t, capturedOpts.IO)
	assert.NotNil(t, capturedOpts.Config)
}
