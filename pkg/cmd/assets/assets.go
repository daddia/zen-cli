package assets

import (
	"github.com/daddia/zen/pkg/cmd/assets/auth"
	"github.com/daddia/zen/pkg/cmd/assets/info"
	"github.com/daddia/zen/pkg/cmd/assets/list"
	"github.com/daddia/zen/pkg/cmd/assets/status"
	"github.com/daddia/zen/pkg/cmd/assets/sync"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdAssets creates the assets command with subcommands
func NewCmdAssets(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets <command>",
		Short: "Manage assets and templates",
		Long: `Manage assets and templates for Zen CLI.

Assets include templates, prompts, and other reusable content stored in
Git repositories. This command provides authentication, discovery, and
synchronization capabilities for asset management.

Authentication:
  Assets are stored in Git repositories that may require authentication.
  Use 'zen assets auth' to configure authentication with GitHub or GitLab.

Discovery:
  List available assets with filtering and search capabilities.
  Get detailed information about specific assets including metadata.

Synchronization:
  Keep your local asset cache synchronized with remote repositories.
  Assets are cached locally for fast access and offline usage.`,
		Example: `  # Configure authentication with GitHub
  zen assets auth github

  # Check authentication and cache status
  zen assets status

  # List all available assets
  zen assets list

  # List only template assets
  zen assets list --type template

  # Get detailed information about a specific asset
  zen assets info technical-spec

  # Synchronize with remote repository
  zen assets sync

  # Force a full synchronization
  zen assets sync --force`,
		GroupID: "assets",
	}

	// Add subcommands
	cmd.AddCommand(auth.NewCmdAssetsAuth(f))
	cmd.AddCommand(status.NewCmdAssetsStatus(f))
	cmd.AddCommand(list.NewCmdAssetsList(f))
	cmd.AddCommand(info.NewCmdAssetsInfo(f))
	cmd.AddCommand(sync.NewCmdAssetsSync(f))

	return cmd
}
