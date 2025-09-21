package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/daddia/zen/internal/logging"
	"gopkg.in/yaml.v3"
)

// YAMLManifestParser implements ManifestParser for YAML manifest files
type YAMLManifestParser struct {
	logger logging.Logger
}

// NewYAMLManifestParser creates a new YAML manifest parser
func NewYAMLManifestParser(logger logging.Logger) *YAMLManifestParser {
	return &YAMLManifestParser{
		logger: logger,
	}
}

// manifestFile represents the structure of the manifest.yaml file
type manifestFile struct {
	SchemaVersion string                   `yaml:"schema_version"`
	Generated     string                   `yaml:"generated"`
	Version       string                   `yaml:"version"`
	Activities    map[string]manifestAsset `yaml:"activities"`
}

// manifestAsset represents an activity entry in the manifest
type manifestAsset struct {
	Name           string   `yaml:"name"`
	Command        string   `yaml:"command"`
	Description    string   `yaml:"description"`
	Format         string   `yaml:"format"`
	Category       string   `yaml:"category"`
	WorkflowStages []string `yaml:"workflow_stages"`
	Tags           []string `yaml:"tags"`
	UseCases       []string `yaml:"use_cases"`
	Assets         struct {
		Prompt string   `yaml:"prompt,omitempty"`
		Output []string `yaml:"output,omitempty"`
	} `yaml:"assets"`
	Variables []manifestVariable `yaml:"variables,omitempty"`
	Version   string             `yaml:"version,omitempty"`
}

// manifestVariable represents a variable in the manifest
type manifestVariable struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Type        string `yaml:"type"`
	Default     string `yaml:"default,omitempty"`
}

// Parse parses the manifest file and returns asset metadata
func (p *YAMLManifestParser) Parse(ctx context.Context, content []byte) ([]AssetMetadata, error) {
	p.logger.Debug("parsing manifest", "size", len(content))

	var manifest manifestFile
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, &AssetClientError{
			Code:    ErrorCodeConfigurationError,
			Message: "failed to parse manifest YAML",
			Details: err.Error(),
		}
	}

	// Validate schema version
	if manifest.SchemaVersion == "" {
		return nil, &AssetClientError{
			Code:    ErrorCodeConfigurationError,
			Message: "manifest missing schema_version",
		}
	}

	var assets []AssetMetadata

	// Process activities
	for activityKey, activity := range manifest.Activities {
		asset, err := p.convertManifestActivity(activity, activityKey)
		if err != nil {
			p.logger.Warn("failed to convert activity", "name", activity.Name, "key", activityKey, "error", err)
			continue
		}
		assets = append(assets, asset)
	}

	p.logger.Debug("manifest parsed successfully", "assets", len(assets))
	return assets, nil
}

// Validate validates the manifest structure
func (p *YAMLManifestParser) Validate(ctx context.Context, content []byte) error {
	p.logger.Debug("validating manifest")

	var manifest manifestFile
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return &AssetClientError{
			Code:    ErrorCodeConfigurationError,
			Message: "invalid manifest YAML syntax",
			Details: err.Error(),
		}
	}

	// Validate required fields
	if manifest.SchemaVersion == "" {
		return &AssetClientError{
			Code:    ErrorCodeConfigurationError,
			Message: "manifest missing required field: schema_version",
		}
	}

	// Validate activity entries
	names := make(map[string]bool)
	for _, activity := range manifest.Activities {
		// Check required fields
		if activity.Name == "" {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: "activity missing required field: name",
			}
		}

		if activity.Description == "" {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: fmt.Sprintf("activity '%s' missing required field: description", activity.Name),
			}
		}

		if activity.Command == "" {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: fmt.Sprintf("activity '%s' missing required field: command", activity.Name),
			}
		}

		// Check for duplicate names
		if names[activity.Name] {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: fmt.Sprintf("duplicate activity name: %s", activity.Name),
			}
		}
		names[activity.Name] = true

		// Validate variables
		for _, variable := range activity.Variables {
			if variable.Name == "" {
				return &AssetClientError{
					Code:    ErrorCodeConfigurationError,
					Message: fmt.Sprintf("activity '%s' has variable missing name", activity.Name),
				}
			}
		}
	}

	p.logger.Debug("manifest validation successful")
	return nil
}

// Private helper methods

func (p *YAMLManifestParser) convertManifestActivity(activity manifestAsset, activityKey string) (AssetMetadata, error) {
	// Convert variables
	var variables []Variable
	for _, v := range activity.Variables {
		variables = append(variables, Variable(v))
	}

	// Determine the primary output file (first one if multiple)
	var outputFile string
	if len(activity.Assets.Output) > 0 {
		outputFile = activity.Assets.Output[0]
	}

	// Determine the asset type based on format
	var assetType AssetType
	switch activity.Format {
	case "template":
		assetType = AssetTypeTemplate
	case "markdown":
		assetType = AssetTypeTemplate // Treat markdown as template
	case "yaml", "yml":
		assetType = AssetTypeTemplate // Treat YAML as template
	case "code":
		assetType = AssetTypeTemplate // Treat code as template
	default:
		assetType = AssetTypeTemplate // Default to template
	}

	return AssetMetadata{
		Name:        activity.Name,
		Type:        assetType,
		Description: activity.Description,
		Format:      activity.Format,
		Category:    activity.Category,
		Tags:        activity.Tags,
		Variables:   variables,
		Path:        outputFile, // Use the output file as the path
		Command:     activity.Command,
		OutputFile:  outputFile,
		UpdatedAt:   time.Now(), // This would ideally come from Git
	}, nil
}

// ManifestValidationError represents manifest validation errors
type ManifestValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

func (e ManifestValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("manifest validation error at line %d, field '%s': %s", e.Line, e.Field, e.Message)
	}
	return fmt.Sprintf("manifest validation error in field '%s': %s", e.Field, e.Message)
}
