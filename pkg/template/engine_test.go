package template

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAssetClient implements AssetClientInterface for testing
type MockAssetClient struct {
	mock.Mock
}

func (m *MockAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*assets.AssetList), args.Error(1)
}

func (m *MockAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*assets.AssetContent), args.Error(1)
}

func (m *MockAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*assets.SyncResult), args.Error(1)
}

func (m *MockAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(*assets.CacheInfo), args.Error(1)
}

func (m *MockAssetClient) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAssetClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewEngine(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		CacheEnabled:  true,
		CacheTTL:      30 * time.Minute,
		CacheSize:     100,
		StrictMode:    false,
		WorkspaceRoot: "/test/workspace",
	}

	engine := NewEngine(logger, mockAssetClient, config)

	assert.NotNil(t, engine)
	// Config should match except for default delimiters which are set by NewEngine
	assert.Equal(t, config.CacheEnabled, engine.config.CacheEnabled)
	assert.Equal(t, config.CacheTTL, engine.config.CacheTTL)
	assert.Equal(t, config.CacheSize, engine.config.CacheSize)
	assert.Equal(t, config.StrictMode, engine.config.StrictMode)
	assert.Equal(t, config.WorkspaceRoot, engine.config.WorkspaceRoot)
	assert.NotNil(t, engine.loader)
	assert.NotNil(t, engine.validator)
	assert.NotNil(t, engine.functions)
	assert.NotNil(t, engine.cache)
}

func TestEngine_CompileTemplate(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		CacheEnabled:  false, // Disable cache for this test
		WorkspaceRoot: "/test/workspace",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := NewEngine(logger, mockAssetClient, config)

	tests := []struct {
		name     string
		template string
		metadata *TemplateMetadata
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.name}}!",
			metadata: &TemplateMetadata{
				Name:        "test-template",
				Description: "Test template",
			},
			wantErr: false,
		},
		{
			name:     "template with zen functions",
			template: "Task ID: {{taskID \"TEST\"}}",
			metadata: &TemplateMetadata{
				Name: "zen-template",
			},
			wantErr: false,
		},
		{
			name:     "invalid template syntax",
			template: "Hello {{.name",
			metadata: &TemplateMetadata{
				Name: "invalid-template",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tmpl, err := engine.CompileTemplate(ctx, tt.name, tt.template, tt.metadata)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tmpl)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)
				assert.Equal(t, tt.name, tmpl.Name)
				assert.Equal(t, tt.template, tmpl.Content)
				assert.NotNil(t, tmpl.Compiled)
				assert.NotEmpty(t, tmpl.Checksum)
			}
		})
	}
}

func TestEngine_RenderTemplate(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		CacheEnabled:  false,
		StrictMode:    false,
		WorkspaceRoot: "/test/workspace",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := NewEngine(logger, mockAssetClient, config)

	tests := []struct {
		name      string
		template  string
		variables map[string]interface{}
		expected  string
		wantErr   bool
	}{
		{
			name:     "simple variable substitution",
			template: "Hello {{.name}}!",
			variables: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:      "zen function usage",
			template:  "Today is {{today}}",
			variables: map[string]interface{}{},
			expected:  "Today is " + time.Now().Format("2006-01-02"),
			wantErr:   false,
		},
		{
			name:      "string functions",
			template:  "{{upper \"hello world\"}}",
			variables: map[string]interface{}{},
			expected:  "HELLO WORLD",
			wantErr:   false,
		},
		{
			name:     "conditional rendering",
			template: "{{if .show}}Visible{{else}}Hidden{{end}}",
			variables: map[string]interface{}{
				"show": true,
			},
			expected: "Visible",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			metadata := &TemplateMetadata{
				Name: tt.name,
			}

			tmpl, err := engine.CompileTemplate(ctx, tt.name, tt.template, metadata)
			require.NoError(t, err)

			result, err := engine.RenderTemplate(ctx, tmpl, tt.variables)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEngine_LoadTemplate(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		CacheEnabled:  true,
		CacheTTL:      30 * time.Minute,
		CacheSize:     100,
		WorkspaceRoot: "/test/workspace",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := NewEngine(logger, mockAssetClient, config)

	templateContent := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "test-template",
			Type:        assets.AssetTypeTemplate,
			Description: "Test template",
			Variables: []assets.Variable{
				{
					Name:        "name",
					Description: "Name variable",
					Type:        "string",
					Required:    true,
				},
			},
		},
		Content:  "Hello {{.name}}!",
		Checksum: "sha256:test",
		Cached:   false,
	}

	mockAssetClient.On("GetAsset", mock.Anything, "test-template", mock.Anything).
		Return(templateContent, nil)

	ctx := context.Background()
	tmpl, err := engine.LoadTemplate(ctx, "test-template")

	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.Equal(t, "test-template", tmpl.Name)
	assert.Equal(t, "Hello {{.name}}!", tmpl.Content)
	assert.NotNil(t, tmpl.Compiled)
	assert.Len(t, tmpl.Variables, 1)
	assert.Equal(t, "name", tmpl.Variables[0].Name)

	mockAssetClient.AssertExpectations(t)
}

func TestEngine_LoadTemplateWithCache(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		CacheEnabled:  true,
		CacheTTL:      30 * time.Minute,
		CacheSize:     100,
		WorkspaceRoot: "/test/workspace",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := NewEngine(logger, mockAssetClient, config)

	templateContent := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "cached-template",
			Type:        assets.AssetTypeTemplate,
			Description: "Cached template",
		},
		Content:  "Cached content",
		Checksum: "sha256:cached",
		Cached:   false,
	}

	// First call should load from asset client
	mockAssetClient.On("GetAsset", mock.Anything, "cached-template", mock.Anything).
		Return(templateContent, nil).Once()

	ctx := context.Background()

	// First load
	tmpl1, err1 := engine.LoadTemplate(ctx, "cached-template")
	assert.NoError(t, err1)
	assert.NotNil(t, tmpl1)

	// Second load should use cache (no additional asset client call)
	tmpl2, err2 := engine.LoadTemplate(ctx, "cached-template")
	assert.NoError(t, err2)
	assert.NotNil(t, tmpl2)
	assert.Equal(t, tmpl1.Name, tmpl2.Name)
	assert.Equal(t, tmpl1.Content, tmpl2.Content)

	mockAssetClient.AssertExpectations(t)
}

func TestEngine_ListTemplates(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		WorkspaceRoot: "/test/workspace",
	}

	engine := NewEngine(logger, mockAssetClient, config)

	assetList := &assets.AssetList{
		Assets: []assets.AssetMetadata{
			{
				Name:        "template1",
				Type:        assets.AssetTypeTemplate,
				Description: "First template",
				Category:    "test",
				Tags:        []string{"tag1", "tag2"},
			},
			{
				Name:        "template2",
				Type:        assets.AssetTypeTemplate,
				Description: "Second template",
				Category:    "test",
				Tags:        []string{"tag2", "tag3"},
			},
		},
		Total:   2,
		HasMore: false,
	}

	mockAssetClient.On("ListAssets", mock.Anything, mock.MatchedBy(func(filter assets.AssetFilter) bool {
		return filter.Type == assets.AssetTypeTemplate && filter.Category == "test"
	})).Return(assetList, nil)

	ctx := context.Background()
	filter := TemplateFilter{
		Category: "test",
	}

	result, err := engine.ListTemplates(ctx, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Templates, 2)
	assert.Equal(t, 2, result.Total)
	assert.False(t, result.HasMore)
	assert.Equal(t, "template1", result.Templates[0].Name)
	assert.Equal(t, "template2", result.Templates[1].Name)

	mockAssetClient.AssertExpectations(t)
}

func TestEngine_ValidateVariables(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		WorkspaceRoot: "/test/workspace",
	}

	engine := NewEngine(logger, mockAssetClient, config)

	tmpl := &Template{
		Name: "test-template",
		Variables: []VariableSpec{
			{
				Name:        "required_var",
				Description: "Required variable",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "optional_var",
				Description: "Optional variable",
				Type:        "int",
				Required:    false,
				Default:     42,
			},
		},
	}

	tests := []struct {
		name      string
		variables map[string]interface{}
		wantErr   bool
	}{
		{
			name: "valid variables",
			variables: map[string]interface{}{
				"required_var": "test value",
				"optional_var": 123,
			},
			wantErr: false,
		},
		{
			name: "missing required variable",
			variables: map[string]interface{}{
				"optional_var": 123,
			},
			wantErr: true,
		},
		{
			name: "wrong type",
			variables: map[string]interface{}{
				"required_var": "test value",
				"optional_var": "not an int",
			},
			wantErr: true,
		},
		{
			name: "required variable only",
			variables: map[string]interface{}{
				"required_var": "test value",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := engine.ValidateVariables(ctx, tmpl, tt.variables)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEngine_GetFunctions(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}
	config := EngineConfig{
		WorkspaceRoot: "/test/workspace",
	}

	engine := NewEngine(logger, mockAssetClient, config)

	functions := engine.GetFunctions()

	assert.NotNil(t, functions)

	// Test some expected functions
	expectedFunctions := []string{
		"upper", "lower", "trim", // Standard functions
		"taskID", "today", "now", // Zen functions
		"camelCase", "snakeCase", // String manipulation
		"zenflowStages", "stageNumber", // Workflow functions
		"workspacePath", "joinPath", // Path functions
	}

	for _, funcName := range expectedFunctions {
		assert.Contains(t, functions, funcName, "Function %s should be available", funcName)
	}
}

func TestEngineConfig_Defaults(t *testing.T) {
	config := EngineConfig{}

	// Test that zero values work as expected
	assert.False(t, config.CacheEnabled)
	assert.Equal(t, time.Duration(0), config.CacheTTL)
	assert.Equal(t, 0, config.CacheSize)
	assert.False(t, config.StrictMode)
	assert.False(t, config.EnableAI)
	assert.Empty(t, config.WorkspaceRoot)
}

func TestTemplateEngineError(t *testing.T) {
	err := &TemplateEngineError{
		Code:    ErrorCodeTemplateNotFound,
		Message: "Template not found",
		Details: "Additional details",
	}

	assert.Equal(t, "Template not found", err.Error())
	assert.Equal(t, ErrorCodeTemplateNotFound, err.Code)
	assert.Equal(t, "Additional details", err.Details)
}
