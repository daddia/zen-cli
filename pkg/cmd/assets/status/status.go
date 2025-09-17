package status

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

// StatusOptions contains options for the status command
type StatusOptions struct {
	IO           *iostreams.IOStreams
	AssetClient  func() (assets.AssetClientInterface, error)
	OutputFormat string
}

// StatusInfo represents the status information to display
type StatusInfo struct {
	Authentication AuthenticationInfo `json:"authentication" yaml:"authentication"`
	Cache          CacheInfo          `json:"cache" yaml:"cache"`
	Repository     RepositoryInfo     `json:"repository" yaml:"repository"`
}

type AuthenticationInfo struct {
	Provider      string    `json:"provider,omitempty" yaml:"provider,omitempty"`
	Authenticated bool      `json:"authenticated" yaml:"authenticated"`
	Username      string    `json:"username,omitempty" yaml:"username,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	LastValidated time.Time `json:"last_validated,omitempty" yaml:"last_validated,omitempty"`
	Status        string    `json:"status" yaml:"status"`
}

type CacheInfo struct {
	Status     string    `json:"status" yaml:"status"`
	SizeMB     float64   `json:"size_mb" yaml:"size_mb"`
	HitRatio   float64   `json:"hit_ratio" yaml:"hit_ratio"`
	LastSync   time.Time `json:"last_sync" yaml:"last_sync"`
	AssetCount int       `json:"asset_count" yaml:"asset_count"`
}

type RepositoryInfo struct {
	URL       string    `json:"url,omitempty" yaml:"url,omitempty"`
	Branch    string    `json:"branch,omitempty" yaml:"branch,omitempty"`
	LastSync  time.Time `json:"last_sync" yaml:"last_sync"`
	Status    string    `json:"status" yaml:"status"`
	Available bool      `json:"available" yaml:"available"`
}

// NewCmdAssetsStatus creates the assets status command
func NewCmdAssetsStatus(f *cmdutil.Factory) *cobra.Command {
	opts := &StatusOptions{
		IO:          f.IOStreams,
		AssetClient: f.AssetClient,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication and cache status",
		Long: `Display the current status of asset authentication, cache, and repository.

This command shows:
- Authentication status for configured Git providers
- Local cache status including size and hit ratio
- Repository synchronization status
- Asset availability (online/offline mode)

The status information helps troubleshoot authentication issues,
monitor cache performance, and understand asset availability.`,
		Example: `  # Show status in default text format
  zen assets status

  # Show status in JSON format
  zen assets status --output json

  # Show status in YAML format
  zen assets status --output yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get output format from persistent flag
			opts.OutputFormat, _ = cmd.Flags().GetString("output")
			return statusRun(opts)
		},
	}

	return cmd
}

func statusRun(opts *StatusOptions) error {
	ctx := context.Background()

	// Get asset client
	client, err := opts.AssetClient()
	if err != nil {
		return errors.Wrap(err, "failed to get asset client")
	}
	defer client.Close()

	// Gather status information
	status, err := gatherStatusInfo(ctx, client)
	if err != nil {
		return errors.Wrap(err, "failed to gather status information")
	}

	// Display status based on output format
	switch opts.OutputFormat {
	case "json":
		return displayStatusJSON(opts, status)
	case "yaml":
		return displayStatusYAML(opts, status)
	default:
		return displayStatusText(opts, status)
	}
}

func gatherStatusInfo(ctx context.Context, client assets.AssetClientInterface) (*StatusInfo, error) {
	status := &StatusInfo{}

	// Get cache information
	cacheInfo, err := client.GetCacheInfo(ctx)
	if err != nil {
		// Don't fail completely, just mark cache as unavailable
		status.Cache = CacheInfo{
			Status: "unavailable",
		}
	} else {
		status.Cache = CacheInfo{
			Status:     "healthy",
			SizeMB:     float64(cacheInfo.TotalSize) / (1024 * 1024),
			HitRatio:   cacheInfo.CacheHitRatio,
			LastSync:   cacheInfo.LastSync,
			AssetCount: cacheInfo.AssetCount,
		}
	}

	// Authentication status - this would need access to auth provider
	// For now, provide placeholder information
	status.Authentication = AuthenticationInfo{
		Provider:      "github", // This would come from configuration
		Authenticated: false,    // This would come from auth provider
		Status:        "unknown",
	}

	// Repository status - this would need access to repository info
	status.Repository = RepositoryInfo{
		URL:       "https://github.com/example/assets.git", // From config
		Branch:    "main",                                  // From config
		LastSync:  status.Cache.LastSync,
		Status:    "unknown",
		Available: false,
	}

	return status, nil
}

func displayStatusJSON(opts *StatusOptions, status *StatusInfo) error {
	encoder := json.NewEncoder(opts.IO.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(status)
}

func displayStatusYAML(opts *StatusOptions, status *StatusInfo) error {
	encoder := yaml.NewEncoder(opts.IO.Out)
	defer encoder.Close()
	return encoder.Encode(status)
}

func displayStatusText(opts *StatusOptions, status *StatusInfo) error {
	cs := internal.NewColorScheme(opts.IO)

	// Header
	fmt.Fprintf(opts.IO.Out, "%s Asset Status\n\n", cs.Bold("📦"))

	// Authentication section
	fmt.Fprintf(opts.IO.Out, "%s Authentication\n", cs.Bold("🔐"))
	if status.Authentication.Authenticated {
		fmt.Fprintf(opts.IO.Out, "  Status: %s Authenticated\n", cs.Green("✓"))
		if status.Authentication.Provider != "" {
			fmt.Fprintf(opts.IO.Out, "  Provider: %s\n", status.Authentication.Provider)
		}
		if status.Authentication.Username != "" {
			fmt.Fprintf(opts.IO.Out, "  Username: %s\n", status.Authentication.Username)
		}
		if !status.Authentication.ExpiresAt.IsZero() {
			fmt.Fprintf(opts.IO.Out, "  Expires: %s\n", formatTime(status.Authentication.ExpiresAt))
		}
	} else {
		fmt.Fprintf(opts.IO.Out, "  Status: %s Not authenticated\n", cs.Red("✗"))
		fmt.Fprintf(opts.IO.Out, "  %s Run 'zen assets auth <provider>' to authenticate\n", cs.Gray("→"))
	}
	fmt.Fprintln(opts.IO.Out)

	// Cache section
	fmt.Fprintf(opts.IO.Out, "%s Cache\n", cs.Bold("💾"))
	switch status.Cache.Status {
	case "healthy":
		fmt.Fprintf(opts.IO.Out, "  Status: %s Healthy\n", cs.Green("✓"))
		fmt.Fprintf(opts.IO.Out, "  Size: %.1f MB\n", status.Cache.SizeMB)
		if status.Cache.AssetCount > 0 {
			fmt.Fprintf(opts.IO.Out, "  Assets: %d cached\n", status.Cache.AssetCount)
		}
		if status.Cache.HitRatio > 0 {
			fmt.Fprintf(opts.IO.Out, "  Hit ratio: %.1f%%\n", status.Cache.HitRatio*100)
		}
		if !status.Cache.LastSync.IsZero() {
			fmt.Fprintf(opts.IO.Out, "  Last sync: %s\n", formatTime(status.Cache.LastSync))
		}
	case "unavailable":
		fmt.Fprintf(opts.IO.Out, "  Status: %s Unavailable\n", cs.Red("✗"))
		fmt.Fprintf(opts.IO.Out, "  %s Cache may need to be initialized\n", cs.Gray("→"))
	default:
		fmt.Fprintf(opts.IO.Out, "  Status: %s Unknown\n", cs.Yellow("?"))
	}
	fmt.Fprintln(opts.IO.Out)

	// Repository section
	fmt.Fprintf(opts.IO.Out, "%s Repository\n", cs.Bold("📡"))
	if status.Repository.Available {
		fmt.Fprintf(opts.IO.Out, "  Status: %s Connected\n", cs.Green("✓"))
		if status.Repository.URL != "" {
			fmt.Fprintf(opts.IO.Out, "  URL: %s\n", status.Repository.URL)
		}
		if status.Repository.Branch != "" {
			fmt.Fprintf(opts.IO.Out, "  Branch: %s\n", status.Repository.Branch)
		}
		if !status.Repository.LastSync.IsZero() {
			fmt.Fprintf(opts.IO.Out, "  Last sync: %s\n", formatTime(status.Repository.LastSync))
		}
	} else {
		fmt.Fprintf(opts.IO.Out, "  Status: %s Offline\n", cs.Yellow("⚠"))
		fmt.Fprintf(opts.IO.Out, "  %s Using cached assets only\n", cs.Gray("→"))
		if !status.Authentication.Authenticated {
			fmt.Fprintf(opts.IO.Out, "  %s Authentication required for repository access\n", cs.Gray("→"))
		}
	}

	// Footer with helpful commands
	if opts.IO.IsStdoutTTY() {
		fmt.Fprintln(opts.IO.Out)
		fmt.Fprintf(opts.IO.Out, "%s Quick actions:\n", cs.Gray("💡"))
		if !status.Authentication.Authenticated {
			fmt.Fprintf(opts.IO.Out, "  %s zen assets auth github    # Authenticate with GitHub\n", cs.Gray("→"))
		}
		if !status.Repository.Available {
			fmt.Fprintf(opts.IO.Out, "  %s zen assets sync           # Synchronize repository\n", cs.Gray("→"))
		}
		fmt.Fprintf(opts.IO.Out, "  %s zen assets list           # List available assets\n", cs.Gray("→"))
	}

	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02 15:04")
	}
}
