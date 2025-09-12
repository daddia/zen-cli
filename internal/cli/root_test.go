package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
)

func TestExecute(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "text",
		CLI: config.CLIConfig{
			NoColor:      false,
			Verbose:      false,
			OutputFormat: "text",
		},
		Workspace: config.WorkspaceConfig{
			Root:       ".",
			ConfigFile: "zen.yaml",
		},
		Development: config.DevelopmentConfig{
			Debug:   false,
			Profile: false,
		},
	}

	logger := logging.NewBasic()
	ctx := context.Background()

	// Test that execute doesn't panic and returns without error for help
	err := Execute(ctx, cfg, logger)

	// The command should complete successfully (help is shown by default)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestNewRootCommand(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "text",
		CLI: config.CLIConfig{
			NoColor:      false,
			Verbose:      false,
			OutputFormat: "text",
		},
		Workspace: config.WorkspaceConfig{
			Root:       ".",
			ConfigFile: "zen.yaml",
		},
		Development: config.DevelopmentConfig{
			Debug:   false,
			Profile: false,
		},
	}

	logger := logging.NewBasic()

	cmd := newRootCommand(cfg, logger)

	// Test basic command properties
	if cmd.Use != "zen" {
		t.Errorf("Use = %q, want %q", cmd.Use, "zen")
	}

	if !strings.Contains(cmd.Short, "AI-Powered Product Lifecycle") {
		t.Errorf("Short description doesn't contain expected text")
	}

	if !strings.Contains(cmd.Long, "Zen is a unified CLI") {
		t.Errorf("Long description doesn't contain expected text")
	}

	// Test that subcommands are added
	expectedCommands := []string{"version", "init", "config", "status", "workflow", "product", "integrations", "templates", "agents"}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expectedCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q not found", expectedCmd)
		}
	}
}

func TestRootCommandHelp(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "text",
		CLI: config.CLIConfig{
			NoColor:      false,
			Verbose:      false,
			OutputFormat: "text",
		},
		Workspace: config.WorkspaceConfig{
			Root:       ".",
			ConfigFile: "zen.yaml",
		},
		Development: config.DevelopmentConfig{
			Debug:   false,
			Profile: false,
		},
	}

	logger := logging.NewBasic()
	cmd := newRootCommand(cfg, logger)

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set help flag and execute
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help command error = %v", err)
	}

	output := buf.String()

	// Check that help contains expected content
	expectedContent := []string{
		"zen",
		"unified CLI that revolutionizes productivity",
		"Usage:",
		"Available Commands:",
		"Flags:",
		"--help",
		"--verbose",
		"--no-color",
		"--output",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing expected content: %q", expected)
		}
	}
}

func TestRootCommandVersion(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "text",
		CLI: config.CLIConfig{
			NoColor:      false,
			Verbose:      false,
			OutputFormat: "text",
		},
		Workspace: config.WorkspaceConfig{
			Root:       ".",
			ConfigFile: "zen.yaml",
		},
		Development: config.DevelopmentConfig{
			Debug:   false,
			Profile: false,
		},
	}

	logger := logging.NewBasic()
	cmd := newRootCommand(cfg, logger)

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set version flag and execute
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Version command error = %v", err)
	}

	output := buf.String()

	// Check that version output is present
	if !strings.Contains(output, "zen version") {
		t.Errorf("Version output missing expected content")
	}
}

func TestNewPlaceholderCommand(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.NewBasic()

	cmd := newPlaceholderCommand("test", "Test description", cfg, logger)

	if cmd.Use != "test" {
		t.Errorf("Use = %q, want %q", cmd.Use, "test")
	}

	if cmd.Short != "Test description" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Test description")
	}

	// Test running the placeholder command
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "planned for future implementation") {
		t.Errorf("Placeholder command output missing expected content. Got: %q", output)
	}
}

func TestGetVersion(t *testing.T) {
	version := getVersion()

	// Should return "dev" by default (no build-time version info in tests)
	if version != "dev" {
		t.Errorf("getVersion() = %q, want %q", version, "dev")
	}
}
