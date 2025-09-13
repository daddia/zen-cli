package config

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdConfig(t *testing.T) {
	f := &cmdutil.Factory{
		IOStreams: iostreams.Test(),
		Config: func() (*config.Config, error) {
			return &config.Config{
				LogLevel:  "info",
				LogFormat: "text",
			}, nil
		},
	}

	cmd := NewCmdConfig(f)

	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, "Display current configuration", cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestConfigOutput(t *testing.T) {
	testConfig := &config.Config{
		LogLevel:  "debug",
		LogFormat: "json",
		CLI: config.CLIConfig{
			NoColor:      true,
			Verbose:      true,
			OutputFormat: "yaml",
		},
		Workspace: config.WorkspaceConfig{
			Root:       "/test/workspace",
			ConfigFile: "zen.yaml",
		},
		Development: config.DevelopmentConfig{
			Debug:   true,
			Profile: false,
		},
	}

	t.Run("text output", func(t *testing.T) {
		out := &bytes.Buffer{}
		f := &cmdutil.Factory{
			IOStreams: &iostreams.IOStreams{
				Out: out,
			},
			Config: func() (*config.Config, error) {
				return testConfig, nil
			},
		}

		cmd := NewCmdConfig(f)
		err := cmd.Execute()
		require.NoError(t, err)

		output := out.String()
		assert.Contains(t, output, "Zen Configuration")
		assert.Contains(t, output, "Logging:")
		assert.Contains(t, output, "debug")
	})

	t.Run("json output", func(t *testing.T) {
		out := &bytes.Buffer{}
		f := &cmdutil.Factory{
			IOStreams: &iostreams.IOStreams{
				Out: out,
			},
			Config: func() (*config.Config, error) {
				return testConfig, nil
			},
		}

		// Create root command with output flag
		rootCmd := &cobra.Command{Use: "test"}
		rootCmd.PersistentFlags().String("output", "text", "")

		cmd := NewCmdConfig(f)
		rootCmd.AddCommand(cmd)

		// Set args to trigger json output
		rootCmd.SetArgs([]string{"config", "--output", "json"})
		rootCmd.PersistentFlags().Set("output", "json")

		err := rootCmd.Execute()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(out.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "debug", result["LogLevel"])
	})
}

func TestDisplayTextConfig(t *testing.T) {
	buf := &bytes.Buffer{}

	cfg := map[string]interface{}{
		"LogLevel":  "info",
		"LogFormat": "text",
		"CLI": map[string]interface{}{
			"NoColor":      false,
			"Verbose":      true,
			"OutputFormat": "text",
		},
		"Workspace": map[string]interface{}{
			"Root":       "/home/user",
			"ConfigFile": "zen.yaml",
		},
		"Development": map[string]interface{}{
			"Debug":   false,
			"Profile": false,
		},
	}

	err := displayTextConfig(buf, cfg)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Zen Configuration")
	assert.Contains(t, output, "Logging:")
	assert.Contains(t, output, "Level:  info")
	assert.Contains(t, output, "CLI:")
	assert.Contains(t, output, "Workspace:")
	assert.Contains(t, output, "Development:")
}
