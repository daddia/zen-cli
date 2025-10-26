package init

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/fs"
	"github.com/daddia/zen/pkg/types"
	"github.com/spf13/cobra"
)

// NewCmdInit creates the init command
func NewCmdInit(f *cmdutil.Factory) *cobra.Command {
	var force bool
	var configFile string

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize your new Zen workspace or reinitialize an existing one",
		GroupID: "workspace",
		Long: `Initialize a new Zen workspace in the current directory.

This command creates a .zen/ directory structure and a zen.yaml configuration file
with default settings based on your project type. It automatically detects common
project types like Git repositories, Node.js, Go, Python, Rust, and Java projects.

Running 'zen init' in an existing workspace is safe and will reinitialize the workspace
without errors, similar to 'git init' behavior.

The .zen/ directory contains:
  - Configuration files
  - Library directory with manifest cache
  - Cache directory
  - Log directory
  - Templates directory (for future use)
  - Backups directory

If GitHub authentication is configured, zen init will automatically:
  - Set up the library infrastructure
  - Download the latest library manifest (if needed)
  - Make library available for immediate use`,
		Example: `  # Initialize in current directory (safe to run multiple times)
  zen init

  # Reinitialize existing workspace (safe operation like git init)
  zen init

  # Force reinitialize with backup of existing configuration
  zen init --force

  # Initialize with custom config file location
  zen init --config ./config/zen.yaml

  # Initialize with verbose output to see project detection
  zen init --verbose`,
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
					if err := os.MkdirAll(configDir, 0750); err != nil {
						return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
					}
				}
			}

			// Show project detection results (only in verbose mode)
			if f.Verbose {
				fmt.Fprintf(f.IOStreams.Out, "Analyzing project in %s...\n", cwd)
			}

			// Check if workspace already exists before initialization
			status, err := ws.Status()
			var wasInitialized bool
			if err == nil {
				wasInitialized = status.Initialized
			}

			// Initialize workspace with force flag
			if err := ws.InitializeWithForce(force); err != nil {
				// Handle typed errors
				if zenErr, ok := err.(*types.Error); ok {
					switch zenErr.Code {
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

			// Get current working directory for output
			cwd, err2 := os.Getwd()
			if err2 != nil {
				if status, err := ws.Status(); err == nil {
					cwd = status.Root
				} else {
					cwd, _ = os.Getwd() // fallback
				}
			}

			// Success message - match git's professional format with reinitialize behavior
			if wasInitialized {
				fmt.Fprintf(f.IOStreams.Out, "Reinitialized existing Zen workspace in %s/.zen/\n", cwd)
			} else {
				fmt.Fprintf(f.IOStreams.Out, "Initialized empty Zen workspace in %s/.zen/\n", cwd)
			}

			// Create initial config file if it doesn't exist
			if err := createInitialConfig(f); err != nil {
				// Don't fail init if config creation fails - just warn
				fmt.Fprintf(f.IOStreams.ErrOut, "! Warning: Failed to create initial config: %v\n", err)
			}

			// Set up library infrastructure
			if err := setupLibraryInfrastructure(f, wasInitialized); err != nil {
				// Don't fail init if library setup fails - just warn
				fmt.Fprintf(f.IOStreams.ErrOut, "! Warning: Failed to set up library infrastructure: %v\n", err)
				fmt.Fprintf(f.IOStreams.ErrOut, "  You can set up library later with 'zen assets sync'\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration and create backup")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (default: zen.yaml)")

	return cmd
}

// createInitialConfig creates an initial config file through the config module
func createInitialConfig(f *cmdutil.Factory) error {
	// Load current config (will use defaults if no file exists)
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// If no config file was loaded, create one with default values
	if cfg.GetConfigFile() == "" {
		// Set a basic initial value to create the file
		if err := cfg.SetValue("log_level", "info"); err != nil {
			return fmt.Errorf("failed to create initial config: %w", err)
		}
	}

	return nil
}

// setupLibraryInfrastructure sets up the library infrastructure during workspace initialization
func setupLibraryInfrastructure(f *cmdutil.Factory, wasInitialized bool) error {
	// 1. Create .zen/library directory
	libraryDir := filepath.Join(".zen", "library")
	if err := os.MkdirAll(libraryDir, 0750); err != nil {
		return fmt.Errorf("failed to create library directory: %w", err)
	}

	// 2. Check if GitHub authentication is available
	authManager, err := f.AuthManager()
	if err != nil {
		// No auth manager available - skip manifest fetch
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if authenticated with GitHub
	if !authManager.IsAuthenticated(ctx, "github") {
		// No GitHub authentication - skip manifest fetch
		// This is normal for first-time users
		return nil
	}

	// 3. Try to fetch manifest if authenticated (best effort)
	return fetchManifestBestEffort(f, ctx, wasInitialized)
}

// fetchManifestBestEffort attempts to fetch the assets manifest without failing init
func fetchManifestBestEffort(f *cmdutil.Factory, ctx context.Context, wasReinit bool) error {
	// Get asset client
	assetClient, err := f.AssetClient()
	if err != nil {
		// Asset client not available - skip
		return nil
	}
	defer assetClient.Close()

	// Create filesystem manager for file existence checks
	fsManager := fs.New(f.Logger)

	// Check if manifest already exists and is recent (< 24 hours)
	manifestPath := filepath.Join(".zen", "library", "manifest.yaml")
	if !wasReinit {
		if stat, err := os.Stat(manifestPath); err == nil {
			age := time.Since(stat.ModTime())
			if age < 24*time.Hour {
				// Manifest is recent, skip fetch
				return nil
			}
		}
	}

	// Try to sync manifest (best effort)
	syncReq := assets.SyncRequest{
		Force:   wasReinit, // Force refresh on reinit
		Shallow: true,      // Only manifest, not full content
		Branch:  "main",
	}

	result, err := assetClient.SyncRepository(ctx, syncReq)
	if err != nil {
		// Sync failed - don't fail init, just skip
		return nil
	}

	// Show success message only if sync worked AND manifest file actually exists
	if result.Status == "success" {
		// Verify the manifest file actually exists on disk using filesystem manager
		if fsManager.FileExists(manifestPath) {
			if result.AssetsUpdated > 0 {
				fmt.Fprintf(f.IOStreams.Out, "✓ Zen library synchronized (%d assets available)\n", result.AssetsUpdated)
			} else {
				fmt.Fprintf(f.IOStreams.Out, "✓ Zen library is up to date\n")
			}
		}
		// If manifest doesn't exist, don't show any success message
	}

	return nil
}
