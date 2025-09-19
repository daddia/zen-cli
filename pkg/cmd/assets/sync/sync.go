package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmd/assets/internal"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// SyncOptions contains options for the sync command
type SyncOptions struct {
	IO           *iostreams.IOStreams
	AssetClient  func() (assets.AssetClientInterface, error)
	OutputFormat string
	Force        bool
	Branch       string
	Timeout      int
}

// NewCmdAssetsSync creates the assets sync command
func NewCmdAssetsSync(f *cmdutil.Factory) *cobra.Command {
	opts := &SyncOptions{
		IO:          f.IOStreams,
		AssetClient: f.AssetClient,
		Branch:      "main",
		Timeout:     60, // 1 minute default for metadata-only sync
	}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize assets with remote repository",
		Long: `Synchronize the local asset metadata with the remote repository.

This command fetches the latest asset manifest (catalog) from the configured
Git repository and updates the local metadata cache. It does NOT download
actual asset content - assets are downloaded on-demand when requested.

What gets synchronized:
- Asset manifest (manifest.yaml) - lightweight metadata only
- Asset descriptions, categories, tags, and checksums
- Asset availability and version information

Actual asset content is downloaded only when you use 'zen assets get <name>'.
This keeps sync operations fast and minimizes network/disk usage.

The sync operation requires authentication with the Git provider.
Use 'zen assets auth' to configure authentication first.`,
		Example: `  # Synchronize asset metadata (manifest only)
  zen assets sync

  # Force refresh of cached metadata
  zen assets sync --force

  # Sync from a specific branch
  zen assets sync --branch develop

  # Sync with custom timeout
  zen assets sync --timeout 60

  # Output sync results as JSON
  zen assets sync --output json

  # After sync, list available assets
  zen assets list

  # Download specific asset content on-demand
  zen assets get technical-spec`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get output format from persistent flag
			opts.OutputFormat, _ = cmd.Flags().GetString("output")
			return syncRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force refresh of cached metadata")
	cmd.Flags().StringVar(&opts.Branch, "branch", "main", "Branch to synchronize")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 60, "Timeout in seconds for sync operation")

	return cmd
}

func syncRun(opts *SyncOptions) error {
	ctx := context.Background()

	// Create timeout context
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)
		defer cancel()
	}

	// Get asset client
	client, err := opts.AssetClient()
	if err != nil {
		return errors.Wrap(err, "failed to get asset client")
	}
	defer client.Close()

	// Show sync start message (unless JSON/YAML output)
	if opts.OutputFormat == "text" || opts.OutputFormat == "" {
		if opts.IO.IsStdoutTTY() {
			cs := internal.NewColorScheme(opts.IO)
			fmt.Fprintf(opts.IO.Out, "%s Synchronizing assets repository...\n", cs.Bold("Syncing"))

			// Show progress indicators
			go showProgressIndicators(ctx, opts.IO)
		} else {
			fmt.Fprintln(opts.IO.Out, "Synchronizing assets repository...")
		}
	}

	// Perform synchronization
	syncRequest := assets.SyncRequest{
		Force:  opts.Force,
		Branch: opts.Branch,
	}

	result, err := client.SyncRepository(ctx, syncRequest)
	if err != nil {
		// Check for specific error types
		if assetErr, ok := err.(*assets.AssetClientError); ok {
			switch assetErr.Code {
			case assets.ErrorCodeAuthenticationFailed:
				return fmt.Errorf("authentication failed. Run 'zen assets auth <provider>' to authenticate")
			case assets.ErrorCodeNetworkError:
				return fmt.Errorf("network error during sync: %v", assetErr.Message)
			case assets.ErrorCodeRepositoryError:
				return fmt.Errorf("repository error: %v", assetErr.Message)
			}
		}
		return errors.Wrap(err, "sync operation failed")
	}

	// Display results based on output format
	switch opts.OutputFormat {
	case "json":
		return displaySyncJSON(opts, result)
	case "yaml":
		return displaySyncYAML(opts, result)
	default:
		return displaySyncText(opts, result)
	}
}

func showProgressIndicators(ctx context.Context, io *iostreams.IOStreams) {
	if !io.IsStdoutTTY() {
		return
	}

	cs := internal.NewColorScheme(io)
	steps := []string{
		"Authenticating with Git provider",
		"Fetching repository updates",
		"Updating local cache",
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	stepIndex := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if stepIndex < len(steps) {
				fmt.Fprintf(io.Out, "%s %s\n", cs.Green("✓"), steps[stepIndex])
				stepIndex++
			} else {
				return
			}
		}
	}
}

func displaySyncJSON(opts *SyncOptions, result *assets.SyncResult) error {
	encoder := json.NewEncoder(opts.IO.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func displaySyncYAML(opts *SyncOptions, result *assets.SyncResult) error {
	encoder := yaml.NewEncoder(opts.IO.Out)
	defer encoder.Close()
	return encoder.Encode(result)
}

func displaySyncText(opts *SyncOptions, result *assets.SyncResult) error {
	cs := internal.NewColorScheme(opts.IO)

	// Clear progress indicators if we showed them
	if opts.IO.IsStdoutTTY() {
		fmt.Fprint(opts.IO.Out, "\033[3A\033[K\033[K\033[K") // Clear last 3 lines
	}

	// Status-based output
	switch result.Status {
	case "success":
		fmt.Fprintf(opts.IO.Out, "%s Sync completed successfully\n", cs.Green("✓"))
	case "partial":
		fmt.Fprintf(opts.IO.Out, "%s Sync completed with warnings\n", cs.Yellow("!"))
		if result.Error != "" {
			fmt.Fprintf(opts.IO.Out, "%s Warning: %s\n", cs.Yellow("!"), result.Error)
		}
	case "error":
		fmt.Fprintf(opts.IO.Out, "%s Sync failed\n", cs.Red("✗"))
		if result.Error != "" {
			fmt.Fprintf(opts.IO.Out, "%s Error: %s\n", cs.Red("✗"), result.Error)
		}
		return fmt.Errorf("sync operation failed")
	default:
		fmt.Fprintf(opts.IO.Out, "%s Sync status unknown\n", cs.Yellow("?"))
	}

	// Show statistics
	if result.AssetsAdded > 0 || result.AssetsUpdated > 0 || result.AssetsRemoved > 0 {
		fmt.Fprintln(opts.IO.Out)
		fmt.Fprintf(opts.IO.Out, "%s Changes:\n", cs.Bold("Changes"))

		if result.AssetsAdded > 0 {
			fmt.Fprintf(opts.IO.Out, "  %s Added: %s assets\n",
				cs.Green("+"), cs.Bold(fmt.Sprintf("%d", result.AssetsAdded)))
		}

		if result.AssetsUpdated > 0 {
			fmt.Fprintf(opts.IO.Out, "  %s Updated: %s assets\n",
				cs.Blue("~"), cs.Bold(fmt.Sprintf("%d", result.AssetsUpdated)))
		}

		if result.AssetsRemoved > 0 {
			fmt.Fprintf(opts.IO.Out, "  %s Removed: %s assets\n",
				cs.Red("-"), cs.Bold(fmt.Sprintf("%d", result.AssetsRemoved)))
		}
	}

	// Show cache and timing information
	fmt.Fprintln(opts.IO.Out)
	fmt.Fprintf(opts.IO.Out, "%s Summary:\n", cs.Bold("Summary"))

	if result.CacheSizeMB > 0 {
		fmt.Fprintf(opts.IO.Out, "  Cache size: %.1f MB\n", result.CacheSizeMB)
	}

	if result.DurationMS > 0 {
		duration := time.Duration(result.DurationMS) * time.Millisecond
		fmt.Fprintf(opts.IO.Out, "  Duration: %s\n", formatDuration(duration))
	}

	if !result.LastSync.IsZero() {
		fmt.Fprintf(opts.IO.Out, "  Last sync: %s\n", result.LastSync.Format("2006-01-02 15:04:05"))
	}

	// Helpful next steps
	if opts.IO.IsStdoutTTY() && result.Status == "success" {
		fmt.Fprintln(opts.IO.Out)
		fmt.Fprintf(opts.IO.Out, "%s Next steps:\n", cs.Gray("Tip:"))
		fmt.Fprintf(opts.IO.Out, "  %s zen assets list           # List available assets\n", cs.Gray("→"))
		fmt.Fprintf(opts.IO.Out, "  %s zen assets info <name>    # Get asset information\n", cs.Gray("→"))
		fmt.Fprintf(opts.IO.Out, "  %s zen assets status         # Check overall status\n", cs.Gray("→"))
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
