package config

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmd/factory"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock IOStreams for testing
type mockIOStreams struct{}

func (m *mockIOStreams) FormatSectionHeader(text string) string {
	return text + "\n" + "================"
}

func (m *mockIOStreams) FormatBold(text string) string {
	return text
}

func (m *mockIOStreams) Indent(text string, level int) string {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	return indent + text
}

func TestNewCmdConfig(t *testing.T) {
	f := factory.New()
	cmd := NewCmdConfig(f)

	assert.Equal(t, "config <command>", cmd.Use)
	assert.Contains(t, cmd.Short, "Manage configuration")
	assert.NotEmpty(t, cmd.Long)
}

func TestDisplayTextConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected []string
	}{
		{
			name: "basic config",
			config: map[string]interface{}{
				"LogLevel":  "info",
				"LogFormat": "text",
				"CLI": map[string]interface{}{
					"NoColor":      false,
					"Verbose":      false,
					"OutputFormat": "text",
				},
			},
			expected: []string{
				"Zen Configuration",
				"================",
				"Logging:",
				"  Level:  info",
				"  Format: text",
				"CLI:",
				"  No Color:      false",
				"  Verbose:       false",
				"  Output Format: text",
			},
		},
		{
			name: "workspace config",
			config: map[string]interface{}{
				"LogLevel": "debug",
				"Workspace": map[string]interface{}{
					"Root":       "/home/user/project",
					"ConfigFile": "zen.yaml",
				},
			},
			expected: []string{
				"Zen Configuration",
				"================",
				"Logging:",
				"  Level:  debug",
				"Workspace:",
				"  Root:        /home/user/project",
				"  Config File: zen.yaml",
			},
		},
		{
			name: "development config",
			config: map[string]interface{}{
				"LogLevel": "trace",
				"Development": map[string]interface{}{
					"Debug":   true,
					"Profile": false,
				},
			},
			expected: []string{
				"Zen Configuration",
				"================",
				"Logging:",
				"  Level:  trace",
				"Development:",
				"  Debug:   true",
				"  Profile: false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			mockStreams := &mockIOStreams{}
			err := displayTextConfig(buf, tt.config, mockStreams)
			assert.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestDisplayTextConfig_JSONMarshaling(t *testing.T) {
	// Test with a struct that needs JSON marshaling
	type TestConfig struct {
		LogLevel string `json:"LogLevel"`
		CLI      struct {
			NoColor bool `json:"NoColor"`
		} `json:"CLI"`
	}

	config := TestConfig{
		LogLevel: "warn",
	}
	config.CLI.NoColor = true

	buf := &bytes.Buffer{}
	mockStreams := &mockIOStreams{}
	err := displayTextConfig(buf, config, mockStreams)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Zen Configuration")
	assert.Contains(t, output, "Level:  warn")
	assert.Contains(t, output, "No Color:      true")
}

func TestDisplayTextConfig_InvalidJSON(t *testing.T) {
	// Test with something that can't be marshaled to JSON
	invalidConfig := make(chan int)

	buf := &bytes.Buffer{}
	mockStreams := &mockIOStreams{}
	err := displayTextConfig(buf, invalidConfig, mockStreams)
	assert.Error(t, err)
}

func TestConfigCommand_Integration(t *testing.T) {
	f := factory.New()
	cmd := NewCmdConfig(f)

	// Test that the command can be created and basic properties are set
	require.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Name())
	assert.True(t, cmd.Runnable())

	// Test flags
	assert.NotNil(t, cmd.Flags())
}

func TestDisplayCurrentConfig(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdConfig(factory)
	cmd.SetArgs([]string{})

	// Test the default behavior (no subcommand)
	err := cmd.Execute()
	require.NoError(t, err)

	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, "Zen Configuration")
}

func TestDisplayCurrentConfig_ConfigError(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Mock factory to return config error
	factory.Config = func() (*config.Config, error) {
		return nil, assert.AnError
	}

	cmd := NewCmdConfig(factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestDisplayCurrentConfig_JSONOutput(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Create a root command with output flag and proper groups
	rootCmd := &cobra.Command{Use: "zen"}
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	// Add the workspace group that config command expects
	rootCmd.AddGroup(&cobra.Group{
		ID:    "workspace",
		Title: "Workspace commands:",
	})

	cmd := NewCmdConfig(factory)
	rootCmd.AddCommand(cmd)

	// Set the output flag to JSON
	rootCmd.SetArgs([]string{"config", "--output", "json"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, `"LogLevel"`)
	assert.Contains(t, output, `"LogFormat"`)
}

func TestDisplayCurrentConfig_YAMLOutput(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Create a root command with output flag and proper groups
	rootCmd := &cobra.Command{Use: "zen"}
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	// Add the workspace group that config command expects
	rootCmd.AddGroup(&cobra.Group{
		ID:    "workspace",
		Title: "Workspace commands:",
	})

	cmd := NewCmdConfig(factory)
	rootCmd.AddCommand(cmd)

	// Set the output flag to YAML
	rootCmd.SetArgs([]string{"config", "--output", "yaml"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, "loglevel:")
	assert.Contains(t, output, "logformat:")
}

func TestDisplayCurrentConfig_InvalidOutputFormat(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	// Create a root command with output flag and proper groups
	rootCmd := &cobra.Command{Use: "zen"}
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	// Add the workspace group that config command expects
	rootCmd.AddGroup(&cobra.Group{
		ID:    "workspace",
		Title: "Workspace commands:",
	})

	cmd := NewCmdConfig(factory)
	rootCmd.AddCommand(cmd)

	// Set invalid output format - should fall back to text
	rootCmd.SetArgs([]string{"config", "--output", "xml"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, "Zen Configuration")
}
