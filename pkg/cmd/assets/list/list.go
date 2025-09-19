package list

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmd/assets/internal"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ListOptions contains options for the list command
type ListOptions struct {
	IO           *iostreams.IOStreams
	AssetClient  func() (assets.AssetClientInterface, error)
	OutputFormat string
	Type         string
	Category     string
	Tags         []string
	Limit        int
	Offset       int
}

// NewCmdAssetsList creates the assets list command
func NewCmdAssetsList(f *cmdutil.Factory) *cobra.Command {
	opts := &ListOptions{
		IO:          f.IOStreams,
		AssetClient: f.AssetClient,
		Limit:       50, // Default limit
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available assets",
		Long: `List available assets with optional filtering.

Assets can be filtered by type, category, and tags. Results are paginated
to handle large asset repositories efficiently.

Asset Types:
- template: Reusable content templates with variables
- prompt: AI prompts for various tasks
- mcp: Model Context Protocol definitions
- schema: JSON/YAML schemas for validation

The list command works offline using cached asset metadata. Use 'zen assets sync'
to update the cache with the latest assets from the repository.`,
		Example: `  # List all assets
  zen assets list

  # List only templates
  zen assets list --type template

  # List assets in a specific category
  zen assets list --category documentation

  # List assets with specific tags
  zen assets list --tags ai,technical

  # Combine filters
  zen assets list --type prompt --category planning --tags sprint

  # Limit results and use pagination
  zen assets list --limit 10 --offset 20

  # Output as JSON
  zen assets list --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get output format from persistent flag
			opts.OutputFormat, _ = cmd.Flags().GetString("output")
			return listRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by asset type (template, prompt, mcp, schema)")
	cmd.Flags().StringVar(&opts.Category, "category", "", "Filter by category")
	cmd.Flags().StringSliceVar(&opts.Tags, "tags", nil, "Filter by tags (comma-separated)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of results")
	cmd.Flags().IntVar(&opts.Offset, "offset", 0, "Number of results to skip")

	return cmd
}

func listRun(opts *ListOptions) error {
	ctx := context.Background()

	// Get asset client
	client, err := opts.AssetClient()
	if err != nil {
		return errors.Wrap(err, "failed to get asset client")
	}
	defer client.Close()

	// Build filter from options
	filter := assets.AssetFilter{
		Category: opts.Category,
		Tags:     opts.Tags,
		Limit:    opts.Limit,
		Offset:   opts.Offset,
	}

	// Parse asset type if provided
	if opts.Type != "" {
		assetType, err := parseAssetType(opts.Type)
		if err != nil {
			return err
		}
		filter.Type = assetType
	}

	// List assets
	assetList, err := client.ListAssets(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "failed to list assets")
	}

	// Display results based on output format
	switch opts.OutputFormat {
	case "json":
		return displayListJSON(opts, assetList)
	case "yaml":
		return displayListYAML(opts, assetList)
	default:
		return displayListText(opts, assetList, filter)
	}
}

func parseAssetType(typeStr string) (assets.AssetType, error) {
	switch strings.ToLower(typeStr) {
	case "template":
		return assets.AssetTypeTemplate, nil
	case "prompt":
		return assets.AssetTypePrompt, nil
	case "mcp":
		return assets.AssetTypeMCP, nil
	case "schema":
		return assets.AssetTypeSchema, nil
	default:
		return "", fmt.Errorf("invalid asset type '%s'. Valid types: template, prompt, mcp, schema", typeStr)
	}
}

func displayListJSON(opts *ListOptions, assetList *assets.AssetList) error {
	encoder := json.NewEncoder(opts.IO.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(assetList)
}

func displayListYAML(opts *ListOptions, assetList *assets.AssetList) error {
	encoder := yaml.NewEncoder(opts.IO.Out)
	defer encoder.Close()
	return encoder.Encode(assetList)
}

func displayListText(opts *ListOptions, assetList *assets.AssetList, filter assets.AssetFilter) error {
	cs := internal.NewColorScheme(opts.IO)

	if len(assetList.Assets) == 0 {
		fmt.Fprintf(opts.IO.Out, "%s No assets found", cs.Gray("Info:"))
		if hasFilters(filter) {
			fmt.Fprintf(opts.IO.Out, " matching the specified criteria")
		}
		fmt.Fprintln(opts.IO.Out, ".")

		if opts.IO.IsStdoutTTY() {
			fmt.Fprintln(opts.IO.Out)
			if !hasFilters(filter) {
				fmt.Fprintf(opts.IO.Out, "%s Try running 'zen assets sync' to synchronize with the repository.\n", cs.Gray("Tip:"))
			} else {
				fmt.Fprintf(opts.IO.Out, "%s Try adjusting your filters or run 'zen assets list' to see all assets.\n", cs.Gray("Tip:"))
			}
		}
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(opts.IO.Out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		cs.Bold("NAME"),
		cs.Bold("TYPE"),
		cs.Bold("CATEGORY"),
		cs.Bold("DESCRIPTION"))

	// Assets
	for _, asset := range assetList.Assets {
		description := asset.Description
		if len(description) > 60 {
			description = description[:57] + "..."
		}

		// Color code asset types
		var typeColor string
		switch asset.Type {
		case assets.AssetTypeTemplate:
			typeColor = cs.Blue(string(asset.Type))
		case assets.AssetTypePrompt:
			typeColor = cs.Green(string(asset.Type))
		case assets.AssetTypeMCP:
			typeColor = cs.Yellow(string(asset.Type))
		case assets.AssetTypeSchema:
			typeColor = cs.Magenta(string(asset.Type))
		default:
			typeColor = string(asset.Type)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			asset.Name,
			typeColor,
			asset.Category,
			description)
	}

	w.Flush()

	// Summary
	fmt.Fprintln(opts.IO.Out)
	if assetList.HasMore {
		fmt.Fprintf(opts.IO.Out, "Total: %s assets (showing %s of %s)",
			cs.Bold(fmt.Sprintf("%d", assetList.Total)),
			cs.Bold(fmt.Sprintf("%d", len(assetList.Assets))),
			cs.Bold(fmt.Sprintf("%d", assetList.Total)))

		if opts.IO.IsStdoutTTY() {
			nextOffset := filter.Offset + filter.Limit
			if nextOffset < assetList.Total {
				fmt.Fprintf(opts.IO.Out, "\n%s Use --offset %d to see more results",
					cs.Gray("Tip:"), nextOffset)
			}
		}
	} else {
		fmt.Fprintf(opts.IO.Out, "Total: %s assets",
			cs.Bold(fmt.Sprintf("%d", assetList.Total)))
	}
	fmt.Fprintln(opts.IO.Out)

	// Show active filters
	if hasFilters(filter) && opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out, "\n%s Active filters:", cs.Gray("Filters:"))
		if filter.Type != "" {
			fmt.Fprintf(opts.IO.Out, " type=%s", filter.Type)
		}
		if filter.Category != "" {
			fmt.Fprintf(opts.IO.Out, " category=%s", filter.Category)
		}
		if len(filter.Tags) > 0 {
			fmt.Fprintf(opts.IO.Out, " tags=%s", strings.Join(filter.Tags, ","))
		}
		fmt.Fprintln(opts.IO.Out)
	}

	return nil
}

func hasFilters(filter assets.AssetFilter) bool {
	return filter.Type != "" || filter.Category != "" || len(filter.Tags) > 0
}
