package assets

import (
	"context"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewYAMLManifestParser(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)

	assert.NotNil(t, parser)
	assert.Equal(t, logger, parser.logger)
}

func TestYAMLManifestParser_Parse_ValidManifest(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)
	ctx := context.Background()

	manifestYAML := `
schema_version: "1.0"
generated: "2024-01-01T00:00:00Z"
assets:
  templates:
    - name: technical-spec
      template: technical-spec.md.template
      description: Technical specification template
      format: markdown
      category: documentation
      tags:
        - documentation
        - technical
      variables:
        - name: FEATURE_NAME
          description: Name of the feature
          required: true
          type: string
        - name: VERSION
          description: Version number
          required: false
          type: string
          default: "1.0"
  prompts:
    - name: code-review
      prompt: code-review.md.prompt
      description: Code review prompt
      format: markdown
      category: quality
      tags:
        - code-review
        - quality
  mcp:
    - name: test-mcp
      description: Test MCP definition
      format: json
      category: integration
      tags:
        - mcp
  schemas:
    - name: test-schema
      description: Test schema
      format: json
      category: validation
      tags:
        - schema
`

	assets, err := parser.Parse(ctx, []byte(manifestYAML))
	require.NoError(t, err)

	// Should have 4 assets total
	assert.Len(t, assets, 4)

	// Check template asset
	template := findAssetByName(assets, "technical-spec")
	require.NotNil(t, template)
	assert.Equal(t, AssetTypeTemplate, template.Type)
	assert.Equal(t, "Technical specification template", template.Description)
	assert.Equal(t, "documentation", template.Category)
	assert.Equal(t, []string{"documentation", "technical"}, template.Tags)
	assert.Equal(t, "templates/technical-spec.md.template", template.Path)
	assert.Len(t, template.Variables, 2)

	// Check first variable
	var1 := template.Variables[0]
	assert.Equal(t, "FEATURE_NAME", var1.Name)
	assert.Equal(t, "Name of the feature", var1.Description)
	assert.True(t, var1.Required)
	assert.Equal(t, "string", var1.Type)

	// Check second variable
	var2 := template.Variables[1]
	assert.Equal(t, "VERSION", var2.Name)
	assert.False(t, var2.Required)
	assert.Equal(t, "1.0", var2.Default)

	// Check prompt asset
	prompt := findAssetByName(assets, "code-review")
	require.NotNil(t, prompt)
	assert.Equal(t, AssetTypePrompt, prompt.Type)
	assert.Equal(t, "prompts/code-review.md.prompt", prompt.Path)

	// Check MCP asset
	mcp := findAssetByName(assets, "test-mcp")
	require.NotNil(t, mcp)
	assert.Equal(t, AssetTypeMCP, mcp.Type)
	assert.Equal(t, "mcp/test-mcp", mcp.Path)

	// Check schema asset
	schema := findAssetByName(assets, "test-schema")
	require.NotNil(t, schema)
	assert.Equal(t, AssetTypeSchema, schema.Type)
	assert.Equal(t, "schemas/test-schema", schema.Path)
}

func TestYAMLManifestParser_Parse_InvalidYAML(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)
	ctx := context.Background()

	invalidYAML := `
schema_version: "1.0"
assets:
  templates:
    - name: test
      invalid: yaml: syntax
`

	assets, err := parser.Parse(ctx, []byte(invalidYAML))

	assert.Nil(t, assets)
	assert.Error(t, err)

	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeConfigurationError, assetErr.Code)
	assert.Contains(t, assetErr.Message, "failed to parse manifest YAML")
}

func TestYAMLManifestParser_Parse_MissingSchemaVersion(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)
	ctx := context.Background()

	manifestYAML := `
assets:
  templates:
    - name: test-template
      template: test.md.template
      description: Test template
`

	assets, err := parser.Parse(ctx, []byte(manifestYAML))

	assert.Nil(t, assets)
	assert.Error(t, err)

	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeConfigurationError, assetErr.Code)
	assert.Contains(t, assetErr.Message, "missing schema_version")
}

func TestYAMLManifestParser_Validate_Success(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)
	ctx := context.Background()

	validManifest := `
schema_version: "1.0"
assets:
  templates:
    - name: test-template
      template: test.md.template
      description: Test template
      format: markdown
      category: test
`

	err := parser.Validate(ctx, []byte(validManifest))
	assert.NoError(t, err)
}

func TestYAMLManifestParser_Validate_InvalidYAML(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)
	ctx := context.Background()

	invalidYAML := `
schema_version: "1.0"
assets:
  templates:
    - name: test
      invalid: yaml: [syntax
`

	err := parser.Validate(ctx, []byte(invalidYAML))

	assert.Error(t, err)
	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeConfigurationError, assetErr.Code)
	assert.Contains(t, assetErr.Message, "invalid manifest YAML syntax")
}

func TestYAMLManifestParser_Validate_MissingRequiredFields(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)
	ctx := context.Background()

	tests := []struct {
		name        string
		manifest    string
		expectedErr string
	}{
		{
			name: "missing schema version",
			manifest: `
assets:
  templates:
    - name: test
      description: Test
`,
			expectedErr: "missing required field: schema_version",
		},
		{
			name: "missing asset name",
			manifest: `
schema_version: "1.0"
assets:
  templates:
    - description: Test without name
`,
			expectedErr: "missing required field: name",
		},
		{
			name: "missing asset description",
			manifest: `
schema_version: "1.0"
assets:
  templates:
    - name: test-template
`,
			expectedErr: "missing required field: description",
		},
		{
			name: "duplicate asset names",
			manifest: `
schema_version: "1.0"
assets:
  templates:
    - name: duplicate
      description: First asset
    - name: duplicate
      description: Second asset
`,
			expectedErr: "duplicate asset name: duplicate",
		},
		{
			name: "variable missing name",
			manifest: `
schema_version: "1.0"
assets:
  templates:
    - name: test-template
      description: Test template
      variables:
        - description: Variable without name
`,
			expectedErr: "has variable missing name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.Validate(ctx, []byte(tt.manifest))

			assert.Error(t, err)
			var assetErr *AssetClientError
			assert.ErrorAs(t, err, &assetErr)
			assert.Equal(t, ErrorCodeConfigurationError, assetErr.Code)
			assert.Contains(t, assetErr.Message, tt.expectedErr)
		})
	}
}

func TestYAMLManifestParser_ConvertManifestAsset_MissingPath(t *testing.T) {
	logger := logging.NewBasic()
	parser := NewYAMLManifestParser(logger)

	tests := []struct {
		name      string
		asset     manifestAsset
		assetType AssetType
		wantErr   bool
	}{
		{
			name: "template missing template field",
			asset: manifestAsset{
				Name:        "test",
				Description: "Test",
			},
			assetType: AssetTypeTemplate,
			wantErr:   true,
		},
		{
			name: "prompt missing prompt field",
			asset: manifestAsset{
				Name:        "test",
				Description: "Test",
			},
			assetType: AssetTypePrompt,
			wantErr:   true,
		},
		{
			name: "valid template",
			asset: manifestAsset{
				Name:        "test",
				Template:    "test.md.template",
				Description: "Test",
			},
			assetType: AssetTypeTemplate,
			wantErr:   false,
		},
		{
			name: "valid prompt",
			asset: manifestAsset{
				Name:        "test",
				Prompt:      "test.md.prompt",
				Description: "Test",
			},
			assetType: AssetTypePrompt,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.convertManifestAsset(tt.asset, tt.assetType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.asset.Name, result.Name)
				assert.Equal(t, tt.assetType, result.Type)
			}
		})
	}
}

func TestManifestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      ManifestValidationError
		expected string
	}{
		{
			name: "error with line number",
			err: ManifestValidationError{
				Field:   "name",
				Message: "missing required field",
				Line:    10,
			},
			expected: "manifest validation error at line 10, field 'name': missing required field",
		},
		{
			name: "error without line number",
			err: ManifestValidationError{
				Field:   "schema_version",
				Message: "invalid version",
			},
			expected: "manifest validation error in field 'schema_version': invalid version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to find asset by name in test
func findAssetByName(assets []AssetMetadata, name string) *AssetMetadata {
	for i := range assets {
		if assets[i].Name == name {
			return &assets[i]
		}
	}
	return nil
}
