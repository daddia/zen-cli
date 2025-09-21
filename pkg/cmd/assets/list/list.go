package list

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
		Long: `List available activities with optional filtering.

This command reads from the local manifest file (.zen/library/manifest.yaml) for fast,
offline activity discovery. The manifest contains metadata about all available activities
without storing the actual content locally.

Activities can be filtered by category and tags. Results are paginated
to handle large activity repositories efficiently.

Each activity represents a workflow step with associated templates and prompts
for generating documentation, code, or configurations.

The list command works offline using the local manifest. Use 'zen assets sync'
to update the manifest with the latest activities from the repository.`,
		Example: `  # List all activities
  zen assets list

  # List activities in a specific category
  zen assets list --category development

  # List activities with specific tags
  zen assets list --tags api,design

  # Combine filters
  zen assets list --category planning --tags strategy

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

	cmd.Flags().StringVar(&opts.Category, "category", "", "Filter by category")
	cmd.Flags().StringSliceVar(&opts.Tags, "tags", nil, "Filter by tags (comma-separated)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of results")
	cmd.Flags().IntVar(&opts.Offset, "offset", 0, "Number of results to skip")

	return cmd
}

func listRun(opts *ListOptions) error {
	// Validate input parameters
	if opts.Limit < 0 {
		return fmt.Errorf("invalid argument: limit cannot be negative")
	}
	if opts.Offset < 0 {
		return fmt.Errorf("invalid argument: offset cannot be negative")
	}

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

	// Sort assets alphabetically by name
	sort.Slice(assetList.Assets, func(i, j int) bool {
		return assetList.Assets[i].Name < assetList.Assets[j].Name
	})

	// Create table writer
	w := tabwriter.NewWriter(opts.IO.Out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header - new format: name | command | description | output format
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		cs.Bold("Name"),
		cs.Bold("Command"),
		cs.Bold("Description"),
		cs.Bold("Output Format"))

	// Activities
	for _, asset := range assetList.Assets {
		description := asset.Description
		if len(description) > 60 {
			description = description[:57] + "..."
		}

		// Format command with backticks for CLI commands
		command := fmt.Sprintf("`%s`", asset.Command)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			asset.Name,
			cs.Blue(command),
			description,
			asset.Format)
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
	return filter.Category != "" || len(filter.Tags) > 0
}
