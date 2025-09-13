package root

import (
	"fmt"

	"github.com/jonathandaddia/zen/pkg/cmd/config"
	cmdinit "github.com/jonathandaddia/zen/pkg/cmd/init"
	"github.com/jonathandaddia/zen/pkg/cmd/status"
	"github.com/jonathandaddia/zen/pkg/cmd/version"
	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdRoot creates the root command
func NewCmdRoot(f *cmdutil.Factory) (*cobra.Command, error) {
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
• Comprehensive Integrations - Product tools, engineering platforms, and communication`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	var verbose bool
	var noColor bool
	var outputFormat string

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, yaml)")

	// Apply flag values
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("verbose") {
			// Update logger level if verbose
			if verbose {
				f.Logger = f.Logger.WithLevel("debug")
			}
		}
		if cmd.Flags().Changed("no-color") {
			f.IOStreams.SetColorEnabled(!noColor)
		}
		return nil
	}

	// Add command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "product",
		Title: "Product management commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "engineering",
		Title: "Engineering workflow commands",
	})

	// Add subcommands
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(cmdinit.NewCmdInit(f))
	cmd.AddCommand(config.NewCmdConfig(f))
	cmd.AddCommand(status.NewCmdStatus(f))

	// Add placeholder commands for future development
	cmd.AddCommand(newPlaceholderCommand("workflow", "Manage engineering workflows", f))
	cmd.AddCommand(newPlaceholderCommand("product", "Product management commands", f))
	cmd.AddCommand(newPlaceholderCommand("integrations", "Manage external integrations", f))
	cmd.AddCommand(newPlaceholderCommand("templates", "Template management", f))
	cmd.AddCommand(newPlaceholderCommand("agents", "AI agent management", f))

	return cmd, nil
}

// newPlaceholderCommand creates a placeholder command for future implementation
func newPlaceholderCommand(name, description string, f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: description,
		Long:  fmt.Sprintf("%s\n\nThis command is planned for future implementation.", description),
		Run: func(cmd *cobra.Command, args []string) {
			f.Logger.Info("command not yet implemented",
				"command", name,
				"description", description)
			fmt.Fprintf(cmd.OutOrStdout(), "📋 Command '%s' is planned for future implementation.\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "💡 Description: %s\n", description)
			fmt.Fprintln(cmd.OutOrStdout(), "\n🚀 This will be available in upcoming releases!")
		},
	}
}
