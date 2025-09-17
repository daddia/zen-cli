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
	SchemaVersion string `yaml:"schema_version"`
	Generated     string `yaml:"generated"`
	Assets        struct {
		Templates []manifestAsset `yaml:"templates"`
		Prompts   []manifestAsset `yaml:"prompts"`
		MCP       []manifestAsset `yaml:"mcp"`
		Schemas   []manifestAsset `yaml:"schemas"`
	} `yaml:"assets"`
}

// manifestAsset represents an asset entry in the manifest
type manifestAsset struct {
	Name        string             `yaml:"name"`
	Template    string             `yaml:"template,omitempty"`
	Prompt      string             `yaml:"prompt,omitempty"`
	Description string             `yaml:"description"`
	Format      string             `yaml:"format"`
	Category    string             `yaml:"category"`
	Tags        []string           `yaml:"tags"`
	Variables   []manifestVariable `yaml:"variables,omitempty"`
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

	// Process templates
	for _, template := range manifest.Assets.Templates {
		asset, err := p.convertManifestAsset(template, AssetTypeTemplate)
		if err != nil {
			p.logger.Warn("failed to convert template asset", "name", template.Name, "error", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Process prompts
	for _, prompt := range manifest.Assets.Prompts {
		asset, err := p.convertManifestAsset(prompt, AssetTypePrompt)
		if err != nil {
			p.logger.Warn("failed to convert prompt asset", "name", prompt.Name, "error", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Process MCP files
	for _, mcp := range manifest.Assets.MCP {
		asset, err := p.convertManifestAsset(mcp, AssetTypeMCP)
		if err != nil {
			p.logger.Warn("failed to convert MCP asset", "name", mcp.Name, "error", err)
			continue
		}
		assets = append(assets, asset)
	}

	// Process schemas
	for _, schema := range manifest.Assets.Schemas {
		asset, err := p.convertManifestAsset(schema, AssetTypeSchema)
		if err != nil {
			p.logger.Warn("failed to convert schema asset", "name", schema.Name, "error", err)
			continue
		}
		assets = append(assets, asset)
	}

	p.logger.Info("manifest parsed successfully", "assets", len(assets))
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

	// Validate asset entries
	allAssets := []manifestAsset{}
	allAssets = append(allAssets, manifest.Assets.Templates...)
	allAssets = append(allAssets, manifest.Assets.Prompts...)
	allAssets = append(allAssets, manifest.Assets.MCP...)
	allAssets = append(allAssets, manifest.Assets.Schemas...)

	names := make(map[string]bool)
	for _, asset := range allAssets {
		// Check required fields
		if asset.Name == "" {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: "asset missing required field: name",
			}
		}

		if asset.Description == "" {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: fmt.Sprintf("asset '%s' missing required field: description", asset.Name),
			}
		}

		// Check for duplicate names
		if names[asset.Name] {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: fmt.Sprintf("duplicate asset name: %s", asset.Name),
			}
		}
		names[asset.Name] = true

		// Validate variables
		for _, variable := range asset.Variables {
			if variable.Name == "" {
				return &AssetClientError{
					Code:    ErrorCodeConfigurationError,
					Message: fmt.Sprintf("asset '%s' has variable missing name", asset.Name),
				}
			}
		}
	}

	p.logger.Debug("manifest validation successful")
	return nil
}

// Private helper methods

func (p *YAMLManifestParser) convertManifestAsset(manifestAsset manifestAsset, assetType AssetType) (AssetMetadata, error) {
	// Determine the file path based on asset type
	var path string
	switch assetType {
	case AssetTypeTemplate:
		if manifestAsset.Template == "" {
			return AssetMetadata{}, fmt.Errorf("template asset '%s' missing template field", manifestAsset.Name)
		}
		path = "templates/" + manifestAsset.Template
	case AssetTypePrompt:
		if manifestAsset.Prompt == "" {
			return AssetMetadata{}, fmt.Errorf("prompt asset '%s' missing prompt field", manifestAsset.Name)
		}
		path = "prompts/" + manifestAsset.Prompt
	case AssetTypeMCP:
		// For MCP files, use name as filename
		path = "mcp/" + manifestAsset.Name
	case AssetTypeSchema:
		// For schema files, use name as filename
		path = "schemas/" + manifestAsset.Name
	default:
		return AssetMetadata{}, fmt.Errorf("unsupported asset type: %s", assetType)
	}

	// Convert variables
	var variables []Variable
	for _, v := range manifestAsset.Variables {
		variables = append(variables, Variable{
			Name:        v.Name,
			Description: v.Description,
			Required:    v.Required,
			Type:        v.Type,
			Default:     v.Default,
		})
	}

	return AssetMetadata{
		Name:        manifestAsset.Name,
		Type:        assetType,
		Description: manifestAsset.Description,
		Format:      manifestAsset.Format,
		Category:    manifestAsset.Category,
		Tags:        manifestAsset.Tags,
		Variables:   variables,
		Path:        path,
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
