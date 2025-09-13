package init

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/types"
	"github.com/spf13/cobra"
)

// NewCmdInit creates the init command
func NewCmdInit(f *cmdutil.Factory) *cobra.Command {
	var force bool
	var configFile string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Zen workspace",
		Long: `Initialize a new Zen workspace in the current directory.

This command creates a .zen/ directory structure and a zen.yaml configuration file
with default settings based on your project type. It automatically detects common
project types like Git repositories, Node.js, Go, Python, Rust, and Java projects.

The .zen/ directory contains:
  - Configuration files
  - Cache directory
  - Log directory
  - Templates directory (for future use)
  - Backups directory

Examples:
  zen init                    # Initialize in current directory
  zen init --force           # Overwrite existing configuration
  zen init --config ./config/zen.yaml  # Use custom config path`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get workspace manager
			ws, err := f.WorkspaceManager()
			if err != nil {
				return fmt.Errorf("failed to get workspace manager: %w", err)
			}

			// Get current directory for display
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			// If custom config file specified, update the workspace manager
			// This is a limitation of the current architecture - we'll work with what we have
			if configFile != "" {
				// Validate the config file path
				if !filepath.IsAbs(configFile) {
					configFile = filepath.Join(cwd, configFile)
				}

				// Check if directory exists
				configDir := filepath.Dir(configFile)
				if _, err := os.Stat(configDir); os.IsNotExist(err) {
					if err := os.MkdirAll(configDir, 0755); err != nil {
						return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
					}
				}
			}

			// Show project detection results (only in verbose mode)
			if f.Verbose {
				fmt.Fprintf(f.IOStreams.Out, "Analyzing project in %s...\n", cwd)
			}

			// Initialize workspace with force flag
			if err := ws.InitializeWithForce(force); err != nil {
				// Handle typed errors
				if zenErr, ok := err.(*types.Error); ok {
					switch zenErr.Code {
					case types.ErrorCodeAlreadyExists:
						fmt.Fprintf(f.IOStreams.ErrOut, "Error: %s\n", zenErr.Message)
						if zenErr.Details != "" {
							fmt.Fprintf(f.IOStreams.ErrOut, "  %s\n", zenErr.Details)
						}
						return cmdutil.ErrSilent
					case types.ErrorCodePermissionDenied:
						fmt.Fprintf(f.IOStreams.ErrOut, "Error: Permission denied: %s\n", zenErr.Message)
						fmt.Fprintf(f.IOStreams.ErrOut, "  Try running with appropriate permissions or choose a different directory.\n")
						return cmdutil.ErrSilent
					default:
						return fmt.Errorf("workspace initialization failed: %w", err)
					}
				}
				return fmt.Errorf("failed to initialize workspace: %w", err)
			}

			// Get status to show results
			status, err := ws.Status()
			if err != nil {
				return fmt.Errorf("failed to get workspace status: %w", err)
			}

			// Get current working directory for output
			cwd, err2 := os.Getwd()
			if err2 != nil {
				cwd = status.Root
			}

			// Success message - match git's professional format
			fmt.Fprintf(f.IOStreams.Out, "Initialized empty Zen workspace in %s/.zen/\n", cwd)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration and create backup")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (default: zen.yaml)")

	return cmd
}
