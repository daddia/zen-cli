package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"github.com/spf13/cobra"
)

// newInitCommand creates the init command
func newInitCommand(cfg *config.Config, logger logging.Logger) *cobra.Command {
	var force bool
	var workspaceDir string

	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new Zen workspace",
		Long: `Initialize a new Zen workspace with default configuration.

This command creates a zen.yaml configuration file and sets up the basic
workspace structure for managing your product lifecycle workflows.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine workspace directory
			if len(args) > 0 {
				workspaceDir = args[0]
			} else {
				var err error
				workspaceDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}

			// Create directory if it doesn't exist
			if err := os.MkdirAll(workspaceDir, 0755); err != nil {
				return fmt.Errorf("failed to create workspace directory: %w", err)
			}

			configPath := filepath.Join(workspaceDir, "zen.yaml")

			// Check if config file already exists
			if !force {
				if _, err := os.Stat(configPath); err == nil {
					return fmt.Errorf("zen.yaml already exists in %s (use --force to overwrite)", workspaceDir)
				}
			}

			// Create default configuration
			configContent := `# Zen CLI Configuration
# This file configures your Zen workspace for AI-powered product lifecycle management

# Logging configuration
log_level: info
log_format: text

# CLI settings
cli:
  no_color: false
  verbose: false
  output_format: text

# Workspace settings
workspace:
  root: .
  config_file: zen.yaml

# Development settings (optional)
development:
  debug: false
  profile: false

# Future configuration sections will be added here as features are implemented:
# - AI agents configuration
# - Integration settings
# - Workflow templates
# - Quality gates
`

			// Write configuration file
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to write configuration file: %w", err)
			}

			logger.Info("workspace initialized",
				"directory", workspaceDir,
				"config_file", configPath)

			fmt.Printf("✅ Zen workspace initialized in %s\n", workspaceDir)
			fmt.Printf("📝 Configuration file created: %s\n", configPath)
			fmt.Println("\n🚀 Next steps:")
			fmt.Println("   zen status    - Check workspace status")
			fmt.Println("   zen --help    - Explore available commands")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")

	return cmd
}
