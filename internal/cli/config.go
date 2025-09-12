package cli

import (
	"encoding/json"
	"fmt"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newConfigCommand creates the config command
func newConfigCommand(cfg *config.Config, logger logging.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display current configuration",
		Long: `Display the current Zen configuration settings.

This shows the effective configuration after loading from files,
environment variables, and command-line flags.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.CLI.OutputFormat {
			case "json":
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(cfg)

			case "yaml":
				encoder := yaml.NewEncoder(cmd.OutOrStdout())
				defer encoder.Close()
				return encoder.Encode(cfg)

			default:
				return displayTextConfig(cmd, cfg)
			}
		},
	}

	return cmd
}

// displayTextConfig displays configuration in human-readable text format
func displayTextConfig(cmd *cobra.Command, cfg *config.Config) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Zen Configuration")
	fmt.Fprintln(cmd.OutOrStdout(), "=================")
	fmt.Fprintln(cmd.OutOrStdout())

	// Logging configuration
	fmt.Fprintln(cmd.OutOrStdout(), "Logging:")
	fmt.Fprintf(cmd.OutOrStdout(), "  Level:  %s\n", cfg.LogLevel)
	fmt.Fprintf(cmd.OutOrStdout(), "  Format: %s\n", cfg.LogFormat)
	fmt.Fprintln(cmd.OutOrStdout())

	// CLI configuration
	fmt.Fprintln(cmd.OutOrStdout(), "CLI:")
	fmt.Fprintf(cmd.OutOrStdout(), "  No Color:      %t\n", cfg.CLI.NoColor)
	fmt.Fprintf(cmd.OutOrStdout(), "  Verbose:       %t\n", cfg.CLI.Verbose)
	fmt.Fprintf(cmd.OutOrStdout(), "  Output Format: %s\n", cfg.CLI.OutputFormat)
	fmt.Fprintln(cmd.OutOrStdout())

	// Workspace configuration
	fmt.Fprintln(cmd.OutOrStdout(), "Workspace:")
	fmt.Fprintf(cmd.OutOrStdout(), "  Root:        %s\n", cfg.Workspace.Root)
	fmt.Fprintf(cmd.OutOrStdout(), "  Config File: %s\n", cfg.Workspace.ConfigFile)
	fmt.Fprintln(cmd.OutOrStdout())

	// Development configuration
	fmt.Fprintln(cmd.OutOrStdout(), "Development:")
	fmt.Fprintf(cmd.OutOrStdout(), "  Debug:   %t\n", cfg.Development.Debug)
	fmt.Fprintf(cmd.OutOrStdout(), "  Profile: %t\n", cfg.Development.Profile)

	return nil
}
