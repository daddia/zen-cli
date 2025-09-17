package info

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmd/assets/internal"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// InfoOptions contains options for the info command
type InfoOptions struct {
	IO              *iostreams.IOStreams
	AssetClient     func() (assets.AssetClientInterface, error)
	OutputFormat    string
	AssetName       string
	IncludeContent  bool
	VerifyIntegrity bool
}

// NewCmdAssetsInfo creates the assets info command
func NewCmdAssetsInfo(f *cmdutil.Factory) *cobra.Command {
	opts := &InfoOptions{
		IO:              f.IOStreams,
		AssetClient:     f.AssetClient,
		VerifyIntegrity: true,
	}

	cmd := &cobra.Command{
		Use:   "info <asset-name>",
		Short: "Show detailed information about an asset",
		Long: `Display detailed information about a specific asset.

This command shows comprehensive metadata about an asset including:
- Basic information (name, type, description)
- Categorization (category, tags)
- Template variables (for template assets)
- File information (size, checksum, last updated)
- Cache status and integrity

The asset content can optionally be included in the output using
the --include-content flag.`,
		Example: `  # Show asset information
  zen assets info technical-spec

  # Include the asset content in output
  zen assets info user-story --include-content

  # Output as JSON with content
  zen assets info technical-spec --output json --include-content

  # Skip integrity verification for faster response
  zen assets info large-template --no-verify`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.AssetName = args[0]
			// Get output format from persistent flag
			opts.OutputFormat, _ = cmd.Flags().GetString("output")
			return infoRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.IncludeContent, "include-content", false, "Include asset content in output")
	cmd.Flags().BoolVar(&opts.VerifyIntegrity, "verify", true, "Verify asset integrity")

	return cmd
}

func infoRun(opts *InfoOptions) error {
	ctx := context.Background()

	// Get asset client
	client, err := opts.AssetClient()
	if err != nil {
		return errors.Wrap(err, "failed to get asset client")
	}
	defer client.Close()

	// Get asset information
	getOpts := assets.GetAssetOptions{
		IncludeMetadata: true,
		VerifyIntegrity: opts.VerifyIntegrity,
		UseCache:        true,
	}

	assetContent, err := client.GetAsset(ctx, opts.AssetName, getOpts)
	if err != nil {
		// Check if it's an asset not found error
		if assetErr, ok := err.(*assets.AssetClientError); ok {
			switch assetErr.Code {
			case assets.ErrorCodeAssetNotFound:
				return fmt.Errorf("asset '%s' not found. Use 'zen assets list' to see available assets", opts.AssetName)
			case assets.ErrorCodeIntegrityError:
				return fmt.Errorf("asset integrity verification failed: %v", assetErr.Message)
			}
		}
		return errors.Wrap(err, "failed to get asset information")
	}

	// Display information based on output format
	switch opts.OutputFormat {
	case "json":
		return displayInfoJSON(opts, assetContent)
	case "yaml":
		return displayInfoYAML(opts, assetContent)
	default:
		return displayInfoText(opts, assetContent)
	}
}

func displayInfoJSON(opts *InfoOptions, content *assets.AssetContent) error {
	output := content
	if !opts.IncludeContent {
		// Create a copy without content to avoid modifying original
		output = &assets.AssetContent{
			Metadata: content.Metadata,
			Checksum: content.Checksum,
			Cached:   content.Cached,
			CacheAge: content.CacheAge,
		}
	}

	encoder := json.NewEncoder(opts.IO.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func displayInfoYAML(opts *InfoOptions, content *assets.AssetContent) error {
	output := content
	if !opts.IncludeContent {
		// Create a copy without content
		output = &assets.AssetContent{
			Metadata: content.Metadata,
			Checksum: content.Checksum,
			Cached:   content.Cached,
			CacheAge: content.CacheAge,
		}
	}

	encoder := yaml.NewEncoder(opts.IO.Out)
	defer encoder.Close()
	return encoder.Encode(output)
}

func displayInfoText(opts *InfoOptions, content *assets.AssetContent) error {
	cs := internal.NewColorScheme(opts.IO)
	meta := content.Metadata

	// Header
	fmt.Fprintf(opts.IO.Out, "%s %s\n\n", cs.Bold("📄"), cs.Bold(meta.Name))

	// Basic information
	fmt.Fprintf(opts.IO.Out, "%s Basic Information\n", cs.Bold("ℹ️"))
	fmt.Fprintf(opts.IO.Out, "  Name: %s\n", meta.Name)

	// Color code asset type
	var typeDisplay string
	switch meta.Type {
	case assets.AssetTypeTemplate:
		typeDisplay = cs.Blue(string(meta.Type))
	case assets.AssetTypePrompt:
		typeDisplay = cs.Green(string(meta.Type))
	case assets.AssetTypeMCP:
		typeDisplay = cs.Yellow(string(meta.Type))
	case assets.AssetTypeSchema:
		typeDisplay = cs.Magenta(string(meta.Type))
	default:
		typeDisplay = string(meta.Type)
	}
	fmt.Fprintf(opts.IO.Out, "  Type: %s\n", typeDisplay)

	if meta.Category != "" {
		fmt.Fprintf(opts.IO.Out, "  Category: %s\n", meta.Category)
	}

	if meta.Description != "" {
		fmt.Fprintf(opts.IO.Out, "  Description: %s\n", meta.Description)
	}
	fmt.Fprintln(opts.IO.Out)

	// Tags
	if len(meta.Tags) > 0 {
		fmt.Fprintf(opts.IO.Out, "%s Tags\n", cs.Bold("🏷️"))
		fmt.Fprintf(opts.IO.Out, "  %s\n", strings.Join(meta.Tags, ", "))
		fmt.Fprintln(opts.IO.Out)
	}

	// Variables (for templates)
	if len(meta.Variables) > 0 {
		fmt.Fprintf(opts.IO.Out, "%s Template Variables\n", cs.Bold("🔧"))
		for _, variable := range meta.Variables {
			required := ""
			if variable.Required {
				required = cs.Red(" (required)")
			}

			fmt.Fprintf(opts.IO.Out, "  %s%s: %s",
				cs.Bold(variable.Name),
				required,
				variable.Description)

			if variable.Type != "" {
				fmt.Fprintf(opts.IO.Out, " [%s]", cs.Gray(variable.Type))
			}

			if variable.Default != "" {
				fmt.Fprintf(opts.IO.Out, " (default: %s)", cs.Gray(variable.Default))
			}

			fmt.Fprintln(opts.IO.Out)
		}
		fmt.Fprintln(opts.IO.Out)
	}

	// File information
	fmt.Fprintf(opts.IO.Out, "%s File Information\n", cs.Bold("📁"))
	fmt.Fprintf(opts.IO.Out, "  Path: %s\n", meta.Path)

	// Calculate content size
	contentSize := len(content.Content)
	fmt.Fprintf(opts.IO.Out, "  Size: %s\n", formatFileSize(contentSize))

	if meta.Checksum != "" {
		// Show truncated checksum for readability
		checksum := meta.Checksum
		if len(checksum) > 20 {
			checksum = checksum[:16] + "..."
		}
		fmt.Fprintf(opts.IO.Out, "  Checksum: %s\n", cs.Gray(checksum))
	}

	if !meta.UpdatedAt.IsZero() {
		fmt.Fprintf(opts.IO.Out, "  Last Updated: %s\n", formatTime(meta.UpdatedAt))
	}
	fmt.Fprintln(opts.IO.Out)

	// Cache information
	fmt.Fprintf(opts.IO.Out, "%s Cache Status\n", cs.Bold("💾"))
	if content.Cached {
		fmt.Fprintf(opts.IO.Out, "  Status: %s Cached\n", cs.Green("✓"))
		if content.CacheAge > 0 {
			cacheAge := time.Duration(content.CacheAge) * time.Second
			fmt.Fprintf(opts.IO.Out, "  Age: %s\n", formatDuration(cacheAge))
		}
	} else {
		fmt.Fprintf(opts.IO.Out, "  Status: %s Not cached\n", cs.Yellow("⚠"))
	}

	// Integrity status
	if opts.VerifyIntegrity {
		fmt.Fprintf(opts.IO.Out, "  Integrity: %s Verified\n", cs.Green("✓"))
	}
	fmt.Fprintln(opts.IO.Out)

	// Content preview or full content
	if opts.IncludeContent {
		fmt.Fprintf(opts.IO.Out, "%s Content\n", cs.Bold("📝"))
		fmt.Fprintf(opts.IO.Out, "%s\n", content.Content)
	} else if opts.IO.IsStdoutTTY() && contentSize > 0 {
		// Show content preview
		preview := content.Content
		const previewLength = 200
		if len(preview) > previewLength {
			preview = preview[:previewLength] + "..."
		}

		fmt.Fprintf(opts.IO.Out, "%s Content Preview\n", cs.Bold("👁️"))
		fmt.Fprintf(opts.IO.Out, "%s\n", cs.Gray(preview))
		fmt.Fprintln(opts.IO.Out)
		fmt.Fprintf(opts.IO.Out, "%s Use --include-content to see the full content\n", cs.Gray("💡"))
	}

	return nil
}

func formatFileSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	days := d.Hours() / 24
	return fmt.Sprintf("%.1fd", days)
}
