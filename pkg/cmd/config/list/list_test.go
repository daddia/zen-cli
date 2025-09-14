package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/daddia/zen/internal/config"
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
