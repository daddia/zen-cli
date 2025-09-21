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
version: "1.0.0"
activities:
  "Technical Spec":
    name: "Technical Spec"
    command: "tech-spec"
    description: "Technical specification template"
    format: "markdown"
    category: "documentation"
    workflow_stages: ["04-design"]
    tags: ["technical", "specification"]
    use_cases:
      - "Create technical specifications"
    assets:
      prompt: "technical-spec.md.prompt.tmpl"
      output:
        - "technical-spec.md.tmpl"
    variables:
      - name: "FEATURE_NAME"
        description: "Name of the feature"
        required: true
        type: "string"
      - name: "VERSION"
        description: "Version number"
        required: false
        type: "string"
        default: "1.0"

  "Code Review":
    name: "Code Review"
    command: "code-review"
    description: "Code review prompt"
    format: "markdown"
    category: "quality"
    workflow_stages: ["05-build"]
    tags: ["code-review", "quality"]
    use_cases:
      - "Review code quality"
    assets:
      prompt: "code-review.md.prompt.tmpl"
      output:
        - "code-review.md.tmpl"
`

	assets, err := parser.Parse(ctx, []byte(manifestYAML))
	require.NoError(t, err)

	// Should have 2 activities total
	assert.Len(t, assets, 2)

	// Check technical spec activity
	techSpec := findAssetByName(assets, "Technical Spec")
	require.NotNil(t, techSpec)
	assert.Equal(t, AssetTypeTemplate, techSpec.Type)
	assert.Equal(t, "tech-spec", techSpec.Command)
	assert.Equal(t, "Technical specification template", techSpec.Description)
	assert.Equal(t, "documentation", techSpec.Category)
	assert.Equal(t, []string{"technical", "specification"}, techSpec.Tags)
	assert.Equal(t, "technical-spec.md.tmpl", techSpec.OutputFile)
	assert.Len(t, techSpec.Variables, 2)

	// Check first variable
	var1 := techSpec.Variables[0]
	assert.Equal(t, "FEATURE_NAME", var1.Name)
	assert.Equal(t, "Name of the feature", var1.Description)
	assert.True(t, var1.Required)
	assert.Equal(t, "string", var1.Type)

	// Check second variable
	var2 := techSpec.Variables[1]
	assert.Equal(t, "VERSION", var2.Name)
	assert.False(t, var2.Required)
	assert.Equal(t, "1.0", var2.Default)

	// Check code review activity
	codeReview := findAssetByName(assets, "Code Review")
	require.NotNil(t, codeReview)
	assert.Equal(t, AssetTypeTemplate, codeReview.Type)
	assert.Equal(t, "code-review", codeReview.Command)
	assert.Equal(t, "code-review.md.tmpl", codeReview.OutputFile)
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
			name: "missing activity name",
			manifest: `
schema_version: "1.0"
version: "1.0.0"
activities:
  "Test Activity":
    command: "test"
    description: "Test without name"
`,
			expectedErr: "missing required field: name",
		},
		{
			name: "missing activity description",
			manifest: `
schema_version: "1.0"
version: "1.0.0"
activities:
  "Test Activity":
    name: "Test Activity"
    command: "test"
`,
			expectedErr: "missing required field: description",
		},
		{
			name: "duplicate activity names",
			manifest: `
schema_version: "1.0"
version: "1.0.0"
activities:
  "Duplicate Activity":
    name: "duplicate"
    command: "dup1"
    description: "First activity"
  "Another Duplicate":
    name: "duplicate"
    command: "dup2"
    description: "Second activity"
`,
			expectedErr: "duplicate activity name: duplicate",
		},
		{
			name: "variable missing name",
			manifest: `
schema_version: "1.0"
version: "1.0.0"
activities:
  "Test Activity":
    name: "Test Activity"
    command: "test"
    description: "Test activity"
    variables:
      - description: "Variable without name"
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
		name        string
		activity    manifestAsset
		activityKey string
		wantErr     bool
	}{
		{
			name: "valid activity",
			activity: manifestAsset{
				Name:        "Test Activity",
				Command:     "test-command",
				Description: "Test activity description",
				Format:      "markdown",
				Category:    "development",
				Assets: struct {
					Prompt string   `yaml:"prompt,omitempty"`
					Output []string `yaml:"output,omitempty"`
				}{
					Output: []string{"test-output.md.tmpl"},
				},
			},
			activityKey: "Test Activity",
			wantErr:     false,
		},
		{
			name: "activity missing name",
			activity: manifestAsset{
				Command:     "test-command",
				Description: "Test",
			},
			activityKey: "Test",
			wantErr:     false, // convertManifestActivity doesn't validate required fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.convertManifestActivity(tt.activity, tt.activityKey)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.activity.Name, result.Name)
				assert.Equal(t, tt.activity.Command, result.Command)
				assert.Equal(t, tt.activity.Description, result.Description)
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
