package draft

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdDraft(f *cmdutil.Factory) *cobra.Command {
	var force bool
	var preview bool
	var outputPath string

	cmd := &cobra.Command{
		Use:   "draft <activity>",
		Short: "Generate document templates with task data",
		Long: `Generate document templates populated with task manifest data.

This command fetches Go templates from the zen-assets repository and processes
them using data from the current task's manifest.yaml file.

Examples:
  # Generate a feature specification
  zen draft feature-spec

  # Preview template before generating
  zen draft user-story --preview

  # Force overwrite existing file
  zen draft epic --force

  # Generate to custom path
  zen draft roadmap --output ./custom/path/`,
		Args:    cobra.ExactArgs(1),
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDraft(f, args[0], force, preview, outputPath)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&preview, "preview", false, "Preview template without generating")
	cmd.Flags().StringVar(&outputPath, "output", "", "Custom output directory")

	return cmd
}

// TaskManifest represents the structure of a task manifest.yaml file
type TaskManifest struct {
	SchemaVersion string `yaml:"schema_version"`
	Task          struct {
		ID       string `yaml:"id"`
		Title    string `yaml:"title"`
		Type     string `yaml:"type"`
		Status   string `yaml:"status"`
		Priority string `yaml:"priority"`
		Size     string `yaml:"size"`
		Points   int    `yaml:"points"`
	} `yaml:"task"`
	Owner struct {
		Name   string `yaml:"name"`
		Email  string `yaml:"email"`
		Github string `yaml:"github"`
	} `yaml:"owner"`
	Team struct {
		Name    string   `yaml:"name"`
		Stream  string   `yaml:"stream"`
		Members []string `yaml:"members"`
	} `yaml:"team"`
	Dates struct {
		Created     string  `yaml:"created"`
		Started     *string `yaml:"started"`
		Target      string  `yaml:"target"`
		Completed   *string `yaml:"completed"`
		LastUpdated string  `yaml:"last_updated"`
	} `yaml:"dates"`
	Workflow struct {
		CurrentStage    string                   `yaml:"current_stage"`
		CompletedStages []string                 `yaml:"completed_stages"`
		Stages          map[string]WorkflowStage `yaml:"stages"`
	} `yaml:"workflow"`
	SuccessCriteria struct {
		Business       []string `yaml:"business"`
		Technical      []string `yaml:"technical"`
		UserExperience []string `yaml:"user_experience"`
	} `yaml:"success_criteria"`
	Dependencies struct {
		Upstream   []string `yaml:"upstream"`
		Downstream []string `yaml:"downstream"`
	} `yaml:"dependencies"`
	Risk struct {
		Level   string   `yaml:"level"`
		Factors []string `yaml:"factors"`
	} `yaml:"risk"`
	Labels       []string               `yaml:"labels"`
	Tags         []string               `yaml:"tags"`
	CustomFields map[string]interface{} `yaml:"custom_fields"`
}

type WorkflowStage struct {
	Name      string   `yaml:"name"`
	Status    string   `yaml:"status"`
	Progress  int      `yaml:"progress"`
	Started   *string  `yaml:"started"`
	Completed *string  `yaml:"completed"`
	Artifacts []string `yaml:"artifacts"`
}

func runDraft(f *cmdutil.Factory, activity string, force, preview bool, outputPath string) error {
	ctx := context.Background()

	// Get logger and IO streams
	logger := f.Logger
	io := f.IOStreams

	logger.Debug("starting draft command", "activity", activity, "force", force, "preview", preview)

	// Step 1: Validate we're in a task directory and load task manifest
	taskManifest, taskDir, err := loadTaskManifest()
	if err != nil {
		return &types.Error{
			Code:    types.ErrorCodeInvalidInput,
			Message: fmt.Sprintf("Failed to load task manifest: %v", err),
		}
	}

	logger.Debug("loaded task manifest", "task_id", taskManifest.Task.ID, "task_dir", taskDir)

	// Step 2: Get asset client and validate activity exists
	assetClient, err := f.AssetClient()
	if err != nil {
		return &types.Error{
			Code:    types.ErrorCodeInvalidConfig,
			Message: fmt.Sprintf("Failed to initialize asset client: %v", err),
		}
	}

	// List assets to find the activity
	assetList, err := assetClient.ListAssets(ctx, assets.AssetFilter{
		Type: assets.AssetTypeTemplate,
	})
	if err != nil {
		return &types.Error{
			Code:    types.ErrorCodeNetworkError,
			Message: fmt.Sprintf("Failed to list assets: %v", err),
		}
	}

	// Find the activity in the asset list
	var activityAsset *assets.AssetMetadata
	for _, asset := range assetList.Assets {
		if asset.Command == activity {
			activityAsset = &asset
			break
		}
	}

	if activityAsset == nil {
		// Show similar activities as suggestions
		suggestions := findSimilarActivities(activity, assetList.Assets)
		errorMsg := fmt.Sprintf("Unknown activity '%s'", activity)
		if len(suggestions) > 0 {
			errorMsg += fmt.Sprintf(". Did you mean: %s?", strings.Join(suggestions, ", "))
		}
		return &types.Error{
			Code:    types.ErrorCodeInvalidInput,
			Message: errorMsg,
		}
	}

	logger.Debug("found activity asset", "name", activityAsset.Name, "format", activityAsset.Format)

	// Step 3: Fetch the template content
	templateContent, err := assetClient.GetAsset(ctx, activityAsset.Name, assets.GetAssetOptions{
		IncludeMetadata: true,
		UseCache:        true,
	})
	if err != nil {
		return &types.Error{
			Code:    types.ErrorCodeAssetNotFound,
			Message: fmt.Sprintf("Failed to fetch template: %v", err),
		}
	}

	// Step 4: Determine output file path
	outputFile := determineOutputPath(activityAsset, taskDir, outputPath)

	// Check if file exists and handle conflicts
	if !force && !preview {
		if _, err := os.Stat(outputFile); err == nil {
			return &types.Error{
				Code:    types.ErrorCodeAlreadyExists,
				Message: fmt.Sprintf("File %s already exists. Use --force to overwrite or --preview to see content", filepath.Base(outputFile)),
			}
		}
	}

	// Step 5: Process template with task data
	if !io.ColorEnabled() {
		fmt.Fprintf(io.Out, "- Fetching template for %s...", activity)
	} else {
		fmt.Fprintf(io.Out, "%s Fetching template for %s...", "✓", activity)
	}

	processedContent, err := processTemplate(templateContent, taskManifest, activityAsset)
	if err != nil {
		if !io.ColorEnabled() {
			fmt.Fprintf(io.Out, " ✗\n")
		} else {
			fmt.Fprintf(io.Out, " %s\n", "✗")
		}
		return &types.Error{
			Code:    types.ErrorCodeUnknown,
			Message: fmt.Sprintf("Failed to process template: %v", err),
		}
	}

	if !io.ColorEnabled() {
		fmt.Fprintf(io.Out, " ✓ Processing with task data...")
	} else {
		fmt.Fprintf(io.Out, " %s Processing with task data...", "✓")
	}

	// Step 6: Preview or write file
	if preview {
		if !io.ColorEnabled() {
			fmt.Fprintf(io.Out, " ✓\n\n--- Preview of %s ---\n", filepath.Base(outputFile))
		} else {
			fmt.Fprintf(io.Out, " %s\n\n--- Preview of %s ---\n", "✓", filepath.Base(outputFile))
		}
		fmt.Fprint(io.Out, processedContent)
		fmt.Fprintf(io.Out, "\n--- End Preview ---\n")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		if !io.ColorEnabled() {
			fmt.Fprintf(io.Out, " ✗\n")
		} else {
			fmt.Fprintf(io.Out, " %s\n", "✗")
		}
		return &types.Error{
			Code:    types.ErrorCodePermissionDenied,
			Message: fmt.Sprintf("Failed to create output directory: %v", err),
		}
	}

	// Write the processed content
	if err := os.WriteFile(outputFile, []byte(processedContent), 0644); err != nil {
		if !io.ColorEnabled() {
			fmt.Fprintf(io.Out, " ✗\n")
		} else {
			fmt.Fprintf(io.Out, " %s\n", "✗")
		}
		return &types.Error{
			Code:    types.ErrorCodePermissionDenied,
			Message: fmt.Sprintf("Failed to write file: %v", err),
		}
	}

	if !io.ColorEnabled() {
		fmt.Fprintf(io.Out, " ✓\n✓ Generated %s with task %s data in %s\n",
			filepath.Base(outputFile), taskManifest.Task.ID, filepath.Dir(outputFile))
	} else {
		fmt.Fprintf(io.Out, " %s\n%s Generated %s with task %s data in %s\n",
			"✓",
			"✓",
			filepath.Base(outputFile),
			taskManifest.Task.ID,
			filepath.Dir(outputFile))
	}

	return nil
}
