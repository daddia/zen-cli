package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// WorkspaceStatus represents the status of a Zen workspace
type WorkspaceStatus struct {
	Initialized   bool   `json:"initialized" yaml:"initialized"`
	ConfigFile    string `json:"config_file" yaml:"config_file"`
	ConfigExists  bool   `json:"config_exists" yaml:"config_exists"`
	WorkspaceRoot string `json:"workspace_root" yaml:"workspace_root"`
	ValidConfig   bool   `json:"valid_config" yaml:"valid_config"`
	ConfigError   string `json:"config_error,omitempty" yaml:"config_error,omitempty"`
}

// newStatusCommand creates the status command
func newStatusCommand(cfg *config.Config, logger logging.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show workspace status",
		Long: `Show the status of the current Zen workspace.

This command displays information about the workspace configuration,
initialization status, and any potential issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			status := checkWorkspaceStatus(cfg)

			switch cfg.CLI.OutputFormat {
			case "json":
				return displayStatusJSON(cmd, status)
			case "yaml":
				return displayStatusYAML(cmd, status)
			default:
				return displayStatusText(cmd, status)
			}
		},
	}

	return cmd
}

// checkWorkspaceStatus checks the current workspace status
func checkWorkspaceStatus(cfg *config.Config) WorkspaceStatus {
	status := WorkspaceStatus{
		WorkspaceRoot: cfg.Workspace.Root,
		ConfigFile:    cfg.Workspace.ConfigFile,
	}

	// Check if config file exists
	configPath := filepath.Join(status.WorkspaceRoot, status.ConfigFile)
	if _, err := os.Stat(configPath); err == nil {
		status.ConfigExists = true
		status.Initialized = true

		// Try to load and validate config
		if _, err := config.Load(); err == nil {
			status.ValidConfig = true
		} else {
			status.ValidConfig = false
			status.ConfigError = err.Error()
		}
	} else {
		status.ConfigExists = false
		status.Initialized = false
	}

	return status
}

// displayStatusText displays status in human-readable text format
func displayStatusText(cmd *cobra.Command, status WorkspaceStatus) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Zen Workspace Status")
	fmt.Fprintln(cmd.OutOrStdout(), "====================")
	fmt.Fprintln(cmd.OutOrStdout())

	// Workspace root
	fmt.Fprintf(cmd.OutOrStdout(), "📁 Workspace Root: %s\n", status.WorkspaceRoot)
	fmt.Fprintf(cmd.OutOrStdout(), "📝 Config File:    %s\n", status.ConfigFile)
	fmt.Fprintln(cmd.OutOrStdout())

	// Initialization status
	if status.Initialized {
		fmt.Fprintln(cmd.OutOrStdout(), "✅ Status: Initialized")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "❌ Status: Not initialized")
		fmt.Fprintln(cmd.OutOrStdout(), "   Run 'zen init' to initialize workspace")
		return nil
	}

	// Configuration status
	if status.ValidConfig {
		fmt.Fprintln(cmd.OutOrStdout(), "✅ Configuration: Valid")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "⚠️  Configuration: Invalid")
		if status.ConfigError != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   Error: %s\n", status.ConfigError)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "💡 Next steps:")
	fmt.Fprintln(cmd.OutOrStdout(), "   zen config    - View current configuration")
	fmt.Fprintln(cmd.OutOrStdout(), "   zen --help    - Explore available commands")

	return nil
}

// displayStatusJSON displays status in JSON format
func displayStatusJSON(cmd *cobra.Command, status WorkspaceStatus) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(status)
}

// displayStatusYAML displays status in YAML format
func displayStatusYAML(cmd *cobra.Command, status WorkspaceStatus) error {
	encoder := yaml.NewEncoder(cmd.OutOrStdout())
	defer encoder.Close()
	return encoder.Encode(status)
}
