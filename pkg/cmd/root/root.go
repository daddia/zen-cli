package root

import (
	"fmt"

	"github.com/daddia/zen/pkg/cmd/assets"
	"github.com/daddia/zen/pkg/cmd/config"
	"github.com/daddia/zen/pkg/cmd/factory"
	cmdinit "github.com/daddia/zen/pkg/cmd/init"
	"github.com/daddia/zen/pkg/cmd/status"
	"github.com/daddia/zen/pkg/cmd/version"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdRoot creates the root command
func NewCmdRoot(f *cmdutil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:           "zen",
		Short:         "AI-Powered Productivity Suite",
		Long:          "Zen. The unified control plane for product & engineering.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	var verbose bool
	var noColor bool
	var outputFormat string
	var configFile string
	var dryRun bool

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, yaml)")
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without making changes")

	// Apply flag values and reload configuration with command context
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Update factory with flag values
		if cmd.Flags().Changed("verbose") {
			f.Verbose = verbose
		}
		if cmd.Flags().Changed("config") {
			f.ConfigFile = configFile
		}
		if cmd.Flags().Changed("dry-run") {
			f.DryRun = dryRun
		}

		// Reload configuration with command context to ensure flag binding
		f.Config = factory.ConfigWithCommand(cmd)

		// Get updated configuration
		cfg, err := f.Config()
		if err != nil {
			f.Logger.Error("failed to load configuration", "error", err)
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Apply configuration-based updates
		if cfg.CLI.Verbose || verbose {
			f.Logger = f.Logger.WithLevel("debug")
		}
		if cfg.CLI.NoColor || noColor {
			f.IOStreams.SetColorEnabled(false)
		}
		if dryRun {
			f.Logger.Info("dry-run mode enabled - no changes will be made")
		}

		// Log configuration sources for debugging
		if cfg.CLI.Verbose {
			sources := cfg.GetLoadedSources()
			if len(sources) > 0 {
				f.Logger.Debug("configuration loaded from sources", "sources", sources)
			}
			if configFile := cfg.GetConfigFile(); configFile != "" {
				f.Logger.Debug("using configuration file", "file", configFile)
			}
		}

		return nil
	}

	// Add command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "start",
		Title: "Zen commands to get you flowing:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "workspace",
		Title: "start a zenspace:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "assets",
		Title: "aceess Zen assets library:",
	})

	// Add subcommands
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(cmdinit.NewCmdInit(f))
	cmd.AddCommand(config.NewCmdConfig(f))
	cmd.AddCommand(status.NewCmdStatus(f))
	cmd.AddCommand(assets.NewCmdAssets(f))

	// Add shell completion command
	cmd.AddCommand(newCompletionCommand(f))

	return cmd, nil
}

// Root exposes the root command for documentation generators and external tools
// This creates a default factory and returns a fully configured root command
func Root() (*cobra.Command, error) {
	f := factory.New()
	return NewCmdRoot(f)
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

// newCompletionCommand creates the shell completion command
func newCompletionCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for Zen CLI.

The completion script for each shell will be different. Please refer to your shell's
documentation on how to install completion scripts.

Examples:
  # Generate bash completion script
  zen completion bash > /usr/local/etc/bash_completion.d/zen

  # Generate zsh completion script
  zen completion zsh > "${fpath[1]}/_zen"

  # Generate fish completion script
  zen completion fish > ~/.config/fish/completions/zen.fish

  # Generate PowerShell completion script
  zen completion powershell > zen.ps1`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(f.IOStreams.Out)
			case "zsh":
				return cmd.Root().GenZshCompletion(f.IOStreams.Out)
			case "fish":
				return cmd.Root().GenFishCompletion(f.IOStreams.Out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(f.IOStreams.Out)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}
