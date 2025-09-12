package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"github.com/spf13/cobra"
)

// Execute runs the CLI with the given context, configuration, and logger
func Execute(ctx context.Context, cfg *config.Config, logger logging.Logger) error {
	rootCmd := newRootCommand(cfg, logger)

	// Add context to command
	rootCmd.SetContext(ctx)

	return rootCmd.Execute()
}

// newRootCommand creates the root command
func newRootCommand(cfg *config.Config, logger logging.Logger) *cobra.Command {
	var verbose bool
	var noColor bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "zen",
		Short: "AI-Powered Product Lifecycle Productivity Platform",
		Long: `Zen is a unified CLI that revolutionizes productivity across the entire product lifecycle.

By orchestrating intelligent workflows for both product management and engineering,
Zen eliminates context switching, automates repetitive tasks, and ensures consistent
quality delivery from ideation to production.

Features:
• Product Management Excellence - Market research, strategy, and roadmap planning
• Engineering Workflow Automation - 12-stage development workflow automation
• AI-First Intelligence - Multi-provider LLM support with context-aware automation
• Comprehensive Integrations - Product tools, engineering platforms, and communication

For more information, visit: https://github.com/jonathandaddia/zen`,
		Version: getVersion(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Update configuration with flag values
			if cmd.Flags().Changed("verbose") {
				cfg.CLI.Verbose = verbose
				if verbose {
					cfg.LogLevel = "debug"
				}
			}
			if cmd.Flags().Changed("no-color") {
				cfg.CLI.NoColor = noColor
			}
			if cmd.Flags().Changed("output") {
				cfg.CLI.OutputFormat = outputFormat
			}

			// Apply NO_COLOR environment variable
			if os.Getenv("NO_COLOR") != "" {
				cfg.CLI.NoColor = true
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", cfg.CLI.Verbose,
		"Enable verbose output")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", cfg.CLI.NoColor,
		"Disable colored output")
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", cfg.CLI.OutputFormat,
		"Output format (text, json, yaml)")

	// Add subcommands
	cmd.AddCommand(newVersionCommand(cfg, logger))
	cmd.AddCommand(newInitCommand(cfg, logger))
	cmd.AddCommand(newConfigCommand(cfg, logger))
	cmd.AddCommand(newStatusCommand(cfg, logger))

	// Add placeholder commands for future development
	cmd.AddCommand(newPlaceholderCommand("workflow", "Manage engineering workflows", cfg, logger))
	cmd.AddCommand(newPlaceholderCommand("product", "Product management commands", cfg, logger))
	cmd.AddCommand(newPlaceholderCommand("integrations", "Manage external integrations", cfg, logger))
	cmd.AddCommand(newPlaceholderCommand("templates", "Template management", cfg, logger))
	cmd.AddCommand(newPlaceholderCommand("agents", "AI agent management", cfg, logger))

	return cmd
}

// newPlaceholderCommand creates a placeholder command for future implementation
func newPlaceholderCommand(name, description string, cfg *config.Config, logger logging.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: description,
		Long:  fmt.Sprintf("%s\n\nThis command is planned for future implementation.", description),
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("command not yet implemented",
				"command", name,
				"description", description)
			fmt.Fprintf(cmd.OutOrStdout(), "📋 Command '%s' is planned for future implementation.\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "💡 Description: %s\n", description)
			fmt.Fprintln(cmd.OutOrStdout(), "\n🚀 This will be available in upcoming releases!")
		},
	}
}

// getVersion returns version information
func getVersion() string {
	// This will be replaced with actual version info at build time
	return "dev"
}
