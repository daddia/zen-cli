package template

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAssetLoader(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}

	loader := NewAssetLoader(mockAssetClient, logger)

	assert.NotNil(t, loader)
	assert.Equal(t, mockAssetClient, loader.assetClient)
	assert.Equal(t, logger, loader.logger)
}

func TestAssetLoader_LoadByName(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	expectedContent := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "test-template",
			Type:        assets.AssetTypeTemplate,
			Description: "Test template",
		},
		Content:  "Hello {{.name}}!",
		Checksum: "sha256:test",
		Cached:   false,
	}

	mockAssetClient.On("GetAsset", mock.Anything, "test-template", mock.MatchedBy(func(opts assets.GetAssetOptions) bool {
		return opts.IncludeMetadata && opts.VerifyIntegrity && opts.UseCache
	})).Return(expectedContent, nil)

	ctx := context.Background()
	result, err := loader.LoadByName(ctx, "test-template")

	assert.NoError(t, err)
	assert.Equal(t, expectedContent, result)
	mockAssetClient.AssertExpectations(t)
}

func TestAssetLoader_LoadByName_Error(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	expectedError := &assets.AssetClientError{
		Code:    assets.ErrorCodeAssetNotFound,
		Message: "Asset not found",
	}

	mockAssetClient.On("GetAsset", mock.Anything, "missing-template", mock.Anything).
		Return((*assets.AssetContent)(nil), expectedError)

	ctx := context.Background()
	result, err := loader.LoadByName(ctx, "missing-template")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to load template 'missing-template'")
	mockAssetClient.AssertExpectations(t)
}

func TestAssetLoader_LoadByCategory(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	assetList := &assets.AssetList{
		Assets: []assets.AssetMetadata{
			{
				Name:        "template1",
				Type:        assets.AssetTypeTemplate,
				Category:    "test",
				Description: "First template",
			},
			{
				Name:        "template2",
				Type:        assets.AssetTypeTemplate,
				Category:    "test",
				Description: "Second template",
			},
		},
		Total:   2,
		HasMore: false,
	}

	content1 := &assets.AssetContent{
		Metadata: assetList.Assets[0],
		Content:  "Content 1",
		Checksum: "sha256:content1",
	}

	content2 := &assets.AssetContent{
		Metadata: assetList.Assets[1],
		Content:  "Content 2",
		Checksum: "sha256:content2",
	}

	mockAssetClient.On("ListAssets", mock.Anything, mock.MatchedBy(func(filter assets.AssetFilter) bool {
		return filter.Type == assets.AssetTypeTemplate && filter.Category == "test"
	})).Return(assetList, nil)

	mockAssetClient.On("GetAsset", mock.Anything, "template1", mock.Anything).Return(content1, nil)
	mockAssetClient.On("GetAsset", mock.Anything, "template2", mock.Anything).Return(content2, nil)

	ctx := context.Background()
	results, err := loader.LoadByCategory(ctx, "test")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Content 1", results[0].Content)
	assert.Equal(t, "Content 2", results[1].Content)
	mockAssetClient.AssertExpectations(t)
}

func TestAssetLoader_LoadByCategory_WithErrors(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	assetList := &assets.AssetList{
		Assets: []assets.AssetMetadata{
			{Name: "template1", Type: assets.AssetTypeTemplate, Category: "test"},
			{Name: "template2", Type: assets.AssetTypeTemplate, Category: "test"},
		},
		Total:   2,
		HasMore: false,
	}

	content1 := &assets.AssetContent{
		Metadata: assetList.Assets[0],
		Content:  "Content 1",
	}

	mockAssetClient.On("ListAssets", mock.Anything, mock.Anything).Return(assetList, nil)
	mockAssetClient.On("GetAsset", mock.Anything, "template1", mock.Anything).Return(content1, nil)
	mockAssetClient.On("GetAsset", mock.Anything, "template2", mock.Anything).
		Return((*assets.AssetContent)(nil), assert.AnError)

	ctx := context.Background()
	results, err := loader.LoadByCategory(ctx, "test")

	// Should succeed but only return the successful template
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Content 1", results[0].Content)
	mockAssetClient.AssertExpectations(t)
}

func TestAssetLoader_ListAvailable(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	expectedList := &assets.AssetList{
		Assets: []assets.AssetMetadata{
			{Name: "template1", Type: assets.AssetTypeTemplate},
			{Name: "template2", Type: assets.AssetTypeTemplate},
		},
		Total:   2,
		HasMore: false,
	}

	mockAssetClient.On("ListAssets", mock.Anything, mock.MatchedBy(func(filter assets.AssetFilter) bool {
		return filter.Type == assets.AssetTypeTemplate && filter.Category == "test"
	})).Return(expectedList, nil)

	ctx := context.Background()
	filter := assets.AssetFilter{
		Category: "test",
		Tags:     []string{"tag1"},
	}

	result, err := loader.ListAvailable(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedList, result)
	mockAssetClient.AssertExpectations(t)
}

func TestAssetLoader_GetMetadata_YAMLFrontmatter(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	content := `---
name: test-template
description: A test template
category: testing
version: 1.0.0
author: Test Author
tags:
  - test
  - example
variables:
  - name: user_name
    description: User's name
    type: string
    required: true
---
# Template Content
Hello {{.user_name}}!`

	ctx := context.Background()
	metadata, err := loader.GetMetadata(ctx, content)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "test-template", metadata.Name)
	assert.Equal(t, "A test template", metadata.Description)
	assert.Equal(t, "testing", metadata.Category)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "Test Author", metadata.Author)
	assert.Equal(t, []string{"test", "example"}, metadata.Tags)
	assert.Len(t, metadata.Variables, 1)
	assert.Equal(t, "user_name", metadata.Variables[0].Name)
	assert.Equal(t, "string", metadata.Variables[0].Type)
	assert.True(t, metadata.Variables[0].Required)
}

func TestAssetLoader_GetMetadata_Comments(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	content := `# @name: comment-template
# @description: Template with comment metadata
# @category: comments
# @version: 2.0.0
# @author: Comment Author
# @tags: comment, metadata
# @created: 2025-09-19
# @updated: 2025-09-19
# @variable: name:string:true:User's name
# @variable: age:int:false:25:User's age

# Template Content
Hello {{.name}}, you are {{.age}} years old!`

	ctx := context.Background()
	metadata, err := loader.GetMetadata(ctx, content)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "comment-template", metadata.Name)
	assert.Equal(t, "Template with comment metadata", metadata.Description)
	assert.Equal(t, "comments", metadata.Category)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Equal(t, "Comment Author", metadata.Author)
	assert.Equal(t, []string{"comment", "metadata"}, metadata.Tags)

	expectedDate := time.Date(2025, 9, 19, 0, 0, 0, 0, time.UTC)
	assert.True(t, metadata.CreatedAt.Equal(expectedDate))
	assert.True(t, metadata.UpdatedAt.Equal(expectedDate))

	assert.Len(t, metadata.Variables, 2)
	assert.Equal(t, "name", metadata.Variables[0].Name)
	assert.Equal(t, "string", metadata.Variables[0].Type)
	assert.True(t, metadata.Variables[0].Required)
	assert.Equal(t, "User's name", metadata.Variables[0].Description)

	assert.Equal(t, "age", metadata.Variables[1].Name)
	assert.Equal(t, "int", metadata.Variables[1].Type)
	assert.False(t, metadata.Variables[1].Required)
	assert.Equal(t, "25", metadata.Variables[1].Default)
	assert.Equal(t, "User's age", metadata.Variables[1].Description)
}

func TestAssetLoader_GetMetadata_NoMetadata(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	content := `# Just a simple template
Hello {{.name}}!`

	ctx := context.Background()
	metadata, err := loader.GetMetadata(ctx, content)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	// Should return empty metadata structure
	assert.Empty(t, metadata.Name)
	assert.Empty(t, metadata.Description)
	assert.Empty(t, metadata.Category)
	assert.Len(t, metadata.Variables, 0)
}

func TestAssetLoader_ParseVariableSpec(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	tests := []struct {
		name     string
		spec     string
		expected *VariableSpec
	}{
		{
			name: "basic required variable",
			spec: "name:string:true:User's name",
			expected: &VariableSpec{
				Name:        "name",
				Type:        "string",
				Required:    true,
				Description: "User's name",
			},
		},
		{
			name: "optional variable with default",
			spec: "age:int:false:25:User's age",
			expected: &VariableSpec{
				Name:        "age",
				Type:        "int",
				Required:    false,
				Default:     "25",
				Description: "User's age",
			},
		},
		{
			name: "variable with description only",
			spec: "url:string:true:The URL to connect to",
			expected: &VariableSpec{
				Name:        "url",
				Type:        "string",
				Required:    true,
				Description: "The URL to connect to",
			},
		},
		{
			name:     "invalid spec - too few parts",
			spec:     "name:string",
			expected: nil,
		},
		{
			name:     "empty spec",
			spec:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.parseVariableSpec(tt.spec)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.Required, result.Required)
				assert.Equal(t, tt.expected.Default, result.Default)
				assert.Equal(t, tt.expected.Description, result.Description)
			}
		})
	}
}

func TestAssetLoader_ExtractFrontmatter(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	loader := NewAssetLoader(mockAssetClient, logger)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "valid frontmatter",
			content: `---
name: test
description: A test
---
Content here`,
			expected: "name: test\ndescription: A test",
		},
		{
			name: "no frontmatter",
			content: `# Just content
Hello world`,
			expected: "",
		},
		{
			name: "frontmatter with extra whitespace",
			content: `---
name: test
---
Content`,
			expected: "name: test",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.extractFrontmatter(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewMetadataExtractor(t *testing.T) {
	logger := logging.NewBasic()
	extractor := NewMetadataExtractor(logger)

	assert.NotNil(t, extractor)
	assert.Equal(t, logger, extractor.logger)
}

func TestMetadataExtractor_ExtractFromAsset(t *testing.T) {
	logger := logging.NewBasic()
	extractor := NewMetadataExtractor(logger)

	asset := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "test-asset",
			Type:        assets.AssetTypeTemplate,
			Description: "Test asset description",
			Category:    "test",
			Tags:        []string{"tag1", "tag2"},
			Variables: []assets.Variable{
				{
					Name:        "test_var",
					Description: "Test variable",
					Type:        "string",
					Required:    true,
				},
			},
			UpdatedAt: time.Date(2025, 9, 19, 12, 0, 0, 0, time.UTC),
		},
		Content: "Template content",
	}

	metadata := extractor.ExtractFromAsset(asset)

	assert.NotNil(t, metadata)
	assert.Equal(t, "test-asset", metadata.Name)
	assert.Equal(t, "Test asset description", metadata.Description)
	assert.Equal(t, "test", metadata.Category)
	assert.Equal(t, []string{"tag1", "tag2"}, metadata.Tags)
	assert.Len(t, metadata.Variables, 1)
	assert.Equal(t, "test_var", metadata.Variables[0].Name)
	assert.Equal(t, "string", metadata.Variables[0].Type)
	assert.True(t, metadata.Variables[0].Required)
}

func TestMetadataExtractor_MergeMetadata(t *testing.T) {
	logger := logging.NewBasic()
	extractor := NewMetadataExtractor(logger)

	base := &TemplateMetadata{
		Name:        "base-template",
		Description: "Base description",
		Category:    "base",
		Version:     "1.0.0",
		Tags:        []string{"base"},
		Variables: []VariableSpec{
			{Name: "base_var", Type: "string"},
		},
	}

	override := &TemplateMetadata{
		Name:        "override-template",
		Description: "Override description",
		Author:      "Override Author",
		Version:     "2.0.0",
		Tags:        []string{"override", "new"},
		Variables: []VariableSpec{
			{Name: "override_var", Type: "int"},
		},
	}

	merged := extractor.MergeMetadata(base, override)

	assert.NotNil(t, merged)
	assert.Equal(t, "override-template", merged.Name) // Override takes precedence
	assert.Equal(t, "Override description", merged.Description)
	assert.Equal(t, "base", merged.Category) // Base value kept (override is empty)
	assert.Equal(t, "2.0.0", merged.Version)
	assert.Equal(t, "Override Author", merged.Author)
	assert.Equal(t, []string{"override", "new"}, merged.Tags)
	assert.Len(t, merged.Variables, 1)
	assert.Equal(t, "override_var", merged.Variables[0].Name)
}
