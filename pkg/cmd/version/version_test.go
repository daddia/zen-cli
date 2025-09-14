package version

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewCmdVersion(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	require.NotNil(t, cmd)

	// Test command properties
	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Display version information", cmd.Short)
	assert.Contains(t, cmd.Long, "Display the version, build information, and platform details for Zen CLI")
}

func TestVersionCommand_TextOutput(t *testing.T) {
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "zen version")
	assert.Contains(t, output, "Build:")
	assert.Contains(t, output, "Commit:")
	assert.Contains(t, output, "Built:")
	assert.Contains(t, output, "Go:")
	assert.Contains(t, output, "Platform:")
}

func TestVersionCommand_JSONOutput(t *testing.T) {
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	factory := cmdutil.NewTestFactory(streams)

	// Create root command to set up persistent flags
	rootCmd, err := createMockRootCommand(factory)
	require.NoError(t, err)

	versionCmd := NewCmdVersion(factory)
	rootCmd.AddCommand(versionCmd)

	rootCmd.SetArgs([]string{"--output", "json", "version"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()

	// Parse JSON to validate structure
	var versionInfo map[string]interface{}
	err = json.Unmarshal([]byte(output), &versionInfo)
	require.NoError(t, err)

	// Check required fields
	assert.Contains(t, versionInfo, "version")
	assert.Contains(t, versionInfo, "git_commit")
	assert.Contains(t, versionInfo, "build_date")
	assert.Contains(t, versionInfo, "go_version")
	assert.Contains(t, versionInfo, "platform")
}

func TestVersionCommand_YAMLOutput(t *testing.T) {
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	factory := cmdutil.NewTestFactory(streams)

	// Create root command to set up persistent flags
	rootCmd, err := createMockRootCommand(factory)
	require.NoError(t, err)

	versionCmd := NewCmdVersion(factory)
	rootCmd.AddCommand(versionCmd)

	rootCmd.SetArgs([]string{"--output", "yaml", "version"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()

	// Parse YAML to validate structure
	var versionInfo map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &versionInfo)
	require.NoError(t, err)

	// Check required fields
	assert.Contains(t, versionInfo, "version")
	assert.Contains(t, versionInfo, "git_commit")
	assert.Contains(t, versionInfo, "build_date")
	assert.Contains(t, versionInfo, "go_version")
	assert.Contains(t, versionInfo, "platform")
}

func TestVersionCommand_InvalidOutputFormat(t *testing.T) {
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	factory := cmdutil.NewTestFactory(streams)

	// Create root command to set up persistent flags
	rootCmd, err := createMockRootCommand(factory)
	require.NoError(t, err)

	versionCmd := NewCmdVersion(factory)
	rootCmd.AddCommand(versionCmd)

	rootCmd.SetArgs([]string{"--output", "invalid", "version"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Should fall back to text output
	output := stdout.String()
	assert.Contains(t, output, "zen version")
}

func TestVersionInfo_Structure(t *testing.T) {
	// Since getVersionInfo is not exported, we'll test through the command
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()

	// Check that version output contains expected fields
	assert.Contains(t, output, "zen version")
	assert.Contains(t, output, "Build:")
	assert.Contains(t, output, "Commit:")
	assert.Contains(t, output, "Built:")
	assert.Contains(t, output, "Go:")
	assert.Contains(t, output, "Platform:")
}

func TestVersionCommand_ColorOutput(t *testing.T) {
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.SetColorEnabled(true)
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "zen version")
	// Note: Color codes would be present in actual output but hard to test
}

func TestVersionCommand_NoColorOutput(t *testing.T) {
	var stdout bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.SetColorEnabled(false)
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "zen version")
	// Should not contain ANSI color codes
	assert.NotContains(t, output, "\033[")
}

func TestVersionCommand_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Help output typically goes to stdout for --help
	output := stdout.String()
	if output == "" {
		output = stderr.String() // Fallback to stderr
	}
	assert.Contains(t, output, "Display the version, build information, and platform details for Zen CLI")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "version")
}

func TestVersionCommand_NoArgs(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)

	// Test that command accepts no arguments
	assert.Equal(t, "version", cmd.Use)

	// Execute with no args should work
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestVersionCommand_WithExtraArgs(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)

	// Test with extra arguments - should still work
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"extra", "args"})

	err := cmd.Execute()
	// May error due to args validation, but should handle gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "unknown command") // Cobra's actual message
	}
}

// Benchmark tests
func BenchmarkVersionCommand_New(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCmdVersion(factory)
		if cmd == nil {
			b.Fatal("command is nil")
		}
	}
}

func BenchmarkVersionCommand_Execute(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdVersion(factory)
	cmd.SetArgs([]string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		err := cmd.Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVersionCommand_JSONOutput(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	rootCmd, err := createMockRootCommand(factory)
	if err != nil {
		b.Fatal(err)
	}

	versionCmd := NewCmdVersion(factory)
	rootCmd.AddCommand(versionCmd)
	rootCmd.SetArgs([]string{"--output", "json", "version"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var stdout bytes.Buffer
		rootCmd.SetOut(&stdout)

		err := rootCmd.Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetVersionInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout
		factory := cmdutil.NewTestFactory(streams)

		cmd := NewCmdVersion(factory)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err != nil {
			b.Fatal(err)
		}

		output := stdout.String()
		if !strings.HasPrefix(output, "zen version") {
			b.Fatalf("version output is invalid, got: %q", output)
		}
	}
}

// Helper function to create a mock root command for testing output formats
func createMockRootCommand(factory *cmdutil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "zen",
		Short: "Test root command",
	}

	// Add output flag like the real root command
	var outputFormat string
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format")

	return cmd, nil
}
