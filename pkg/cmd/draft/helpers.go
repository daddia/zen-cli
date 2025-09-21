package draft

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/processor"
	"gopkg.in/yaml.v3"
)

// loadTaskManifest loads the task manifest from the current directory or parent directories
func loadTaskManifest() (*TaskManifest, string, error) {
	// Look for manifest.yaml in current directory and parent directories
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for a task manifest
	dir := currentDir
	for {
		manifestPath := filepath.Join(dir, "manifest.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			// Found manifest.yaml, check if it's a task manifest
			content, err := os.ReadFile(manifestPath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read manifest file: %w", err)
			}

			var manifest TaskManifest
			if err := yaml.Unmarshal(content, &manifest); err != nil {
				return nil, "", fmt.Errorf("failed to parse manifest file: %w", err)
			}

			// Validate it's a task manifest
			if manifest.Task.ID == "" {
				return nil, "", fmt.Errorf("manifest.yaml found but does not contain task information")
			}

			return &manifest, dir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached root directory
			break
		}
		dir = parentDir
	}

	return nil, "", fmt.Errorf("no task manifest.yaml found in current directory or parent directories")
}

// findSimilarActivities finds activities with similar names for suggestions
func findSimilarActivities(activity string, assets []assets.AssetMetadata) []string {
	var suggestions []string
	activity = strings.ToLower(activity)

	for _, asset := range assets {
		command := strings.ToLower(asset.Command)
		name := strings.ToLower(asset.Name)

		// Simple similarity check - contains substring or starts with
		if strings.Contains(command, activity) || strings.Contains(activity, command) ||
			strings.Contains(name, activity) || strings.Contains(activity, name) {
			suggestions = append(suggestions, asset.Command)
			if len(suggestions) >= 3 { // Limit suggestions
				break
			}
		}
	}

	return suggestions
}

// WorkTypeMapping maps workflow stages and categories to task work-type directories
var WorkTypeMapping = map[string]string{
	// Stage-based mapping (primary)
	"01-align":      "",          // Root level for alignment documents
	"02-discover":   "research",  // Discovery and investigation
	"03-prioritize": "",          // Root level for prioritization
	"04-design":     "design",    // Specifications and planning
	"05-build":      "execution", // Implementation and testing
	"06-ship":       "execution", // Deployment and release
	"07-learn":      "outcomes",  // Learning and retrospectives

	// Category-based mapping (fallback)
	"analysis":        "research",
	"planning":        "",       // Root level
	"development":     "design", // Default to design unless stage overrides
	"quality":         "execution",
	"operations":      "execution",
	"documentation":   "execution",
	"task-management": "", // Root level
}

// determineOutputPath determines the output file path based on activity and options
func determineOutputPath(asset *assets.AssetMetadata, taskDir, customPath string) string {
	if customPath != "" {
		if filepath.IsAbs(customPath) {
			return customPath
		}
		return filepath.Join(taskDir, customPath)
	}

	// Determine work-type directory based on workflow stages and category
	workTypeDir := determineWorkTypeDirectory(asset)

	// Determine output filename based on asset format and name
	var filename string
	if asset.OutputFile != "" {
		filename = asset.OutputFile
	} else {
		// Generate filename from command name and format
		switch asset.Format {
		case "markdown":
			filename = asset.Command + ".md"
		case "yaml":
			filename = asset.Command + ".yaml"
		case "json":
			filename = asset.Command + ".json"
		default:
			filename = asset.Command + ".md" // Default to markdown
		}
	}

	// Combine task directory, work-type directory, and filename
	if workTypeDir == "" {
		return filepath.Join(taskDir, filename)
	}
	return filepath.Join(taskDir, workTypeDir, filename)
}

// determineWorkTypeDirectory determines the work-type directory based on asset metadata
func determineWorkTypeDirectory(asset *assets.AssetMetadata) string {
	// First priority: Use workflow stages from asset metadata
	if len(asset.WorkflowStages) > 0 {
		// Use the first workflow stage as the primary indicator
		primaryStage := asset.WorkflowStages[0]
		if workType, exists := WorkTypeMapping[primaryStage]; exists {
			return workType
		}
	}

	// Second priority: Check tags for workflow stage information
	if len(asset.Tags) > 0 {
		for _, tag := range asset.Tags {
			if workType, exists := WorkTypeMapping[tag]; exists {
				return workType
			}
		}
	}

	// Third priority: Category-based mapping
	if workType, exists := WorkTypeMapping[asset.Category]; exists {
		return workType
	}

	// Final fallback: Default mapping based on category patterns
	switch asset.Category {
	case "analysis":
		return "research"
	case "planning":
		return "" // Root level
	case "development":
		return "design"
	case "quality":
		return "execution"
	case "operations":
		return "execution"
	case "documentation":
		return "execution"
	default:
		return "" // Root level as final fallback
	}
}

// processTemplate processes the template content with task manifest data
func processTemplate(templateContent *assets.AssetContent, taskManifest *TaskManifest, asset *assets.AssetMetadata) (string, error) {
	// Create template data from task manifest
	templateData := createTemplateData(taskManifest)

	// Parse the Go template
	tmpl, err := template.New(asset.Name).Parse(templateContent.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Determine output type from asset format
	var outputType processor.OutputType
	switch asset.Format {
	case "markdown":
		outputType = processor.MarkdownOutput
	case "yaml":
		outputType = processor.YAMLOutput
	case "json":
		outputType = processor.JSONOutput
	default:
		outputType = processor.MarkdownOutput // Default to markdown
	}

	// Create processing context
	ctx := processor.ProcessingContext{
		TemplateInfo: processor.TemplateInfo{
			FilePath:   asset.Path,
			OutputType: outputType,
			FileName:   asset.Name,
			Extension:  filepath.Ext(asset.Name),
		},
		Template: tmpl,
		Data:     templateData,
		Options: processor.ProcessingOptions{
			ValidateOutput:  true,
			StrictMode:      false,
			TimeoutDuration: 30 * time.Second,
			MaxMemoryMB:     50,
		},
	}

	// Create processor factory and get appropriate processor
	factory := processor.NewFactory()
	proc, err := factory.CreateProcessor(outputType)
	if err != nil {
		return "", fmt.Errorf("failed to create processor: %w", err)
	}

	// Process the template
	result, err := proc.Process(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to process template: %w", err)
	}

	return result, nil
}

// createTemplateData creates template data map from task manifest
func createTemplateData(manifest *TaskManifest) map[string]interface{} {
	data := map[string]interface{}{
		// Task information
		"TASK_ID":      manifest.Task.ID,
		"TASK_TITLE":   manifest.Task.Title,
		"TASK_TYPE":    manifest.Task.Type,
		"TASK_STATUS":  manifest.Task.Status,
		"PRIORITY":     manifest.Task.Priority,
		"SIZE":         manifest.Task.Size,
		"STORY_POINTS": manifest.Task.Points,

		// Owner information
		"OWNER_NAME":      manifest.Owner.Name,
		"OWNER_EMAIL":     manifest.Owner.Email,
		"GITHUB_USERNAME": manifest.Owner.Github,

		// Team information
		"TEAM_NAME":    manifest.Team.Name,
		"STREAM_TYPE":  manifest.Team.Stream,
		"TEAM_MEMBERS": manifest.Team.Members,

		// Dates
		"CREATED_DATE": manifest.Dates.Created,
		"TARGET_DATE":  manifest.Dates.Target,
		"LAST_UPDATED": manifest.Dates.LastUpdated,

		// Workflow
		"CURRENT_STAGE":    manifest.Workflow.CurrentStage,
		"COMPLETED_STAGES": manifest.Workflow.CompletedStages,
		"WORKFLOW_STAGES":  manifest.Workflow.Stages,

		// Success criteria
		"BUSINESS_CRITERIA":  manifest.SuccessCriteria.Business,
		"TECHNICAL_CRITERIA": manifest.SuccessCriteria.Technical,
		"UX_CRITERIA":        manifest.SuccessCriteria.UserExperience,

		// Dependencies
		"UPSTREAM_DEPS":   manifest.Dependencies.Upstream,
		"DOWNSTREAM_DEPS": manifest.Dependencies.Downstream,

		// Risk
		"RISK_LEVEL":   manifest.Risk.Level,
		"RISK_FACTORS": manifest.Risk.Factors,

		// Labels and tags
		"LABELS": manifest.Labels,
		"TAGS":   manifest.Tags,

		// Custom fields
		"CUSTOM_FIELDS": manifest.CustomFields,

		// Additional computed fields
		"CURRENT_DATE":     time.Now().Format("2006-01-02"),
		"CURRENT_DATETIME": time.Now().Format("2006-01-02 15:04:05"),
	}

	// Add optional date fields if they exist
	if manifest.Dates.Started != nil {
		data["STARTED_DATE"] = *manifest.Dates.Started
	}
	if manifest.Dates.Completed != nil {
		data["COMPLETED_DATE"] = *manifest.Dates.Completed
	}

	return data
}
