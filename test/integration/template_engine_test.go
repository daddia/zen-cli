package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAssetClient implements AssetClientInterface for integration testing
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

// TestTemplateEngine_GoTemplateSyntaxIntegration demonstrates the Template Engine
// working with proper Go template syntax (not Handlebars)
func TestTemplateEngine_GoTemplateSyntaxIntegration(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}

	config := template.EngineConfig{
		CacheEnabled:  true,
		CacheTTL:      30 * time.Minute,
		CacheSize:     100,
		StrictMode:    false,
		WorkspaceRoot: "/test/workspace",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := template.NewEngine(logger, mockAssetClient, config)

	// Test Go template with variables, conditionals, and loops
	goTemplateContent := `# Task Manifest
schema_version: "1.0"

# Basic task information
task:
  id: "{{.TASK_ID}}"
  title: "{{.TASK_TITLE}}"
  type: "{{.TASK_TYPE}}"
  status: "{{.TASK_STATUS}}"
  priority: "{{.PRIORITY}}"
  points: {{.STORY_POINTS}}

# Owner information
owner:
  name: "{{.OWNER_NAME}}"
  email: "{{.OWNER_EMAIL}}"
  github: "{{.GITHUB_USERNAME}}"

# Team members using Go template range
team:
  name: "{{.TEAM_NAME}}"
  members:
{{range .TEAM_MEMBERS}}    - name: "{{.name}}"
      role: "{{.role}}"
      github: "{{.github}}"
{{end}}

# Workflow stages with conditional logic
workflow:
  current_stage: "{{.CURRENT_STAGE}}"
  completed_stages:
{{range .COMPLETED_STAGES}}    - "{{.}}"
{{end}}

  # Stage progress using Go template conditionals
{{if .SHOW_PROGRESS}}  progress:
{{range .WORKFLOW_STAGES}}    {{.id}}:
      status: "{{.status}}"
      progress: {{.progress}}
{{if .started_date}}      started: "{{.started_date}}"{{end}}
{{if .completed_date}}      completed: "{{.completed_date}}"{{end}}
{{end}}{{end}}

# Generated metadata using Zen functions
generated:
  task_id: "{{taskID "ZEN"}}"
  created_date: "{{today}}"
  workspace_path: "{{workspacePath "tasks"}}"
  stage_name: "{{stageName .CURRENT_STAGE}}"
  stage_number: {{stageNumber .CURRENT_STAGE}}
  next_stage: "{{nextStage .CURRENT_STAGE}}"

# String manipulation examples
naming:
  camel_case: "{{camelCase .TASK_TITLE}}"
  snake_case: "{{snakeCase .TASK_TITLE}}"
  kebab_case: "{{kebabCase .TASK_TITLE}}"

# Conditional rendering
{{if .INCLUDE_DEBUG}}debug:
  enabled: true
  level: "{{.DEBUG_LEVEL}}"
{{end}}`

	_ = &template.TemplateMetadata{
		Name:        "go-template-test",
		Description: "Go template syntax test",
		Category:    "test",
		Variables: []template.VariableSpec{
			{Name: "TASK_ID", Type: "string", Required: true},
			{Name: "TASK_TITLE", Type: "string", Required: true},
			{Name: "TASK_TYPE", Type: "string", Required: true, Default: "story"},
			{Name: "TASK_STATUS", Type: "string", Required: true, Default: "in_progress"},
			{Name: "PRIORITY", Type: "string", Required: true, Default: "P2"},
			{Name: "STORY_POINTS", Type: "int", Required: false, Default: 3},
			{Name: "OWNER_NAME", Type: "string", Required: true},
			{Name: "OWNER_EMAIL", Type: "string", Required: true},
			{Name: "GITHUB_USERNAME", Type: "string", Required: true},
			{Name: "TEAM_NAME", Type: "string", Required: true},
			{Name: "TEAM_MEMBERS", Type: "array", Required: false},
			{Name: "CURRENT_STAGE", Type: "string", Required: true},
			{Name: "COMPLETED_STAGES", Type: "array", Required: false},
			{Name: "SHOW_PROGRESS", Type: "bool", Required: false, Default: true},
			{Name: "WORKFLOW_STAGES", Type: "array", Required: false},
			{Name: "INCLUDE_DEBUG", Type: "bool", Required: false, Default: false},
			{Name: "DEBUG_LEVEL", Type: "string", Required: false},
		},
	}

	assetContent := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "go-template-test",
			Type:        assets.AssetTypeTemplate,
			Description: "Go template syntax test",
			Category:    "test",
		},
		Content:  goTemplateContent,
		Checksum: "sha256:gotemplate",
		Cached:   false,
	}

	mockAssetClient.On("GetAsset", mock.Anything, "go-template-test", mock.Anything).
		Return(assetContent, nil)

	ctx := context.Background()

	// Load and compile the template
	tmpl, err := engine.LoadTemplate(ctx, "go-template-test")
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	// Test variables with Go template data structures
	variables := map[string]interface{}{
		"TASK_ID":         "ZEN-001",
		"TASK_TITLE":      "Implement Go Template Engine",
		"TASK_TYPE":       "feature",
		"TASK_STATUS":     "in_progress",
		"PRIORITY":        "P1",
		"STORY_POINTS":    5,
		"OWNER_NAME":      "John Doe",
		"OWNER_EMAIL":     "john.doe@company.com",
		"GITHUB_USERNAME": "johndoe",
		"TEAM_NAME":       "Platform Team",
		"TEAM_MEMBERS": []map[string]interface{}{
			{"name": "Alice Smith", "role": "Developer", "github": "alice"},
			{"name": "Bob Johnson", "role": "QA", "github": "bob"},
		},
		"CURRENT_STAGE":    "04-design",
		"COMPLETED_STAGES": []string{"01-align", "02-discover", "03-prioritize"},
		"SHOW_PROGRESS":    true,
		"WORKFLOW_STAGES": []map[string]interface{}{
			{"id": "01-align", "status": "completed", "progress": 100, "completed_date": "2025-09-15"},
			{"id": "02-discover", "status": "completed", "progress": 100, "completed_date": "2025-09-16"},
			{"id": "03-prioritize", "status": "completed", "progress": 100, "completed_date": "2025-09-17"},
			{"id": "04-design", "status": "in_progress", "progress": 60, "started_date": "2025-09-18"},
		},
		"INCLUDE_DEBUG": true,
		"DEBUG_LEVEL":   "debug",
	}

	// Render the template
	result, err := engine.RenderTemplate(ctx, tmpl, variables)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify Go template features work correctly
	assert.Contains(t, result, `id: "ZEN-001"`)
	assert.Contains(t, result, `title: "Implement Go Template Engine"`)
	assert.Contains(t, result, `points: 5`)

	// Verify range loops work (Go template syntax)
	assert.Contains(t, result, `- name: "Alice Smith"`)
	assert.Contains(t, result, `  role: "Developer"`)
	assert.Contains(t, result, `- name: "Bob Johnson"`)
	assert.Contains(t, result, `  role: "QA"`)

	// Verify completed stages array
	assert.Contains(t, result, `- "01-align"`)
	assert.Contains(t, result, `- "02-discover"`)
	assert.Contains(t, result, `- "03-prioritize"`)

	// Verify conditional rendering
	assert.Contains(t, result, `debug:`)
	assert.Contains(t, result, `enabled: true`)
	assert.Contains(t, result, `level: "debug"`)

	// Verify Zen functions work
	assert.Contains(t, result, `task_id: "ZEN-`)                                      // taskID function
	assert.Contains(t, result, `created_date: "`+time.Now().Format("2006-01-02")+`"`) // today function
	assert.Contains(t, result, `workspace_path: "/test/workspace/tasks"`)             // workspacePath function
	assert.Contains(t, result, `stage_name: "Design"`)                                // stageName function
	assert.Contains(t, result, `stage_number: 4`)                                     // stageNumber function
	assert.Contains(t, result, `next_stage: "05-build"`)                              // nextStage function

	// Verify string manipulation functions
	assert.Contains(t, result, `camel_case: "implementGoTemplateEngine"`)    // camelCase function
	assert.Contains(t, result, `snake_case: "implement_go_template_engine"`) // snakeCase function
	assert.Contains(t, result, `kebab_case: "implement-go-template-engine"`) // kebabCase function

	mockAssetClient.AssertExpectations(t)
}

// TestTemplateEngine_GoTemplateAdvancedFeatures tests more advanced Go template features
func TestTemplateEngine_GoTemplateAdvancedFeatures(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}

	config := template.EngineConfig{
		CacheEnabled:  false,
		StrictMode:    false,
		WorkspaceRoot: "/test",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := template.NewEngine(logger, mockAssetClient, config)

	// Template with advanced Go template features
	advancedTemplate := `# Advanced Go Template Features

# With statement and pipeline
{{with .user}}User: {{.name}} ({{.email}}){{end}}

# Complex conditionals with comparisons
{{if gt .score 90}}Grade: A
{{else if gt .score 80}}Grade: B
{{else if gt .score 70}}Grade: C
{{else}}Grade: F{{end}}

# Range with index
Items:
{{range $i, $item := .items}}{{add $i 1}}. {{$item}}
{{end}}

# Range with key-value pairs
Config:
{{range $key, $value := .config}}{{$key}}: {{$value}}
{{end}}

# Nested templates and functions
{{define "header"}}=== {{.title}} ==={{end}}
{{template "header" .}}

# String manipulation with pipes
Original: {{.text}}
Upper: {{.text | upper}}
Trimmed: {{.text | trim | upper}}

# Mathematical operations
Sum: {{add .a .b}}
Product: {{mul .a .b}}
Comparison: {{if eq .a .b}}Equal{{else}}Not equal{{end}}

# Complex data structures
{{range .projects}}Project: {{.name}}
  Status: {{.status}}
  Team:
{{range .team}}    - {{.name}} ({{.role}})
{{end}}
{{end}}`

	metadata := &template.TemplateMetadata{
		Name:        "advanced-template",
		Description: "Advanced Go template features",
	}

	ctx := context.Background()
	tmpl, err := engine.CompileTemplate(ctx, "advanced-template", advancedTemplate, metadata)
	require.NoError(t, err)

	variables := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		"score": 85,
		"items": []string{"apple", "banana", "cherry"},
		"config": map[string]interface{}{
			"debug":   true,
			"timeout": 30,
			"retries": 3,
		},
		"title": "Project Status",
		"text":  "  Hello World  ",
		"a":     10,
		"b":     5,
		"projects": []map[string]interface{}{
			{
				"name":   "Project Alpha",
				"status": "active",
				"team": []map[string]interface{}{
					{"name": "Alice", "role": "Lead"},
					{"name": "Bob", "role": "Developer"},
				},
			},
			{
				"name":   "Project Beta",
				"status": "planning",
				"team": []map[string]interface{}{
					{"name": "Charlie", "role": "Architect"},
				},
			},
		},
	}

	result, err := engine.RenderTemplate(ctx, tmpl, variables)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify advanced features work
	assert.Contains(t, result, "User: John Doe (john@example.com)")
	assert.Contains(t, result, "Grade: B") // 85 > 80
	assert.Contains(t, result, "1. apple")
	assert.Contains(t, result, "2. banana")
	assert.Contains(t, result, "3. cherry")
	assert.Contains(t, result, "debug: true")
	assert.Contains(t, result, "timeout: 30")
	assert.Contains(t, result, "=== Project Status ===")
	assert.Contains(t, result, "Upper:   HELLO WORLD  ") // upper function applied to original text
	assert.Contains(t, result, "Trimmed: HELLO WORLD")
	assert.Contains(t, result, "Sum: 15")
	assert.Contains(t, result, "Product: 50")
	assert.Contains(t, result, "Not equal")
	assert.Contains(t, result, "Project: Project Alpha")
	assert.Contains(t, result, "- Alice (Lead)")
	assert.Contains(t, result, "- Bob (Developer)")
	assert.Contains(t, result, "Project: Project Beta")
	assert.Contains(t, result, "- Charlie (Architect)")
}

// TestTemplateEngine_ErrorHandling tests error handling with Go templates
func TestTemplateEngine_ErrorHandling(t *testing.T) {
	logger := logging.NewBasic()
	mockAssetClient := &MockAssetClient{}

	config := template.EngineConfig{
		CacheEnabled:  false,
		StrictMode:    true, // Strict mode for error testing
		WorkspaceRoot: "/test",
	}
	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	engine := template.NewEngine(logger, mockAssetClient, config)

	tests := []struct {
		name         string
		template     string
		variables    map[string]interface{}
		expectError  bool
		errorMessage string
	}{
		{
			name:         "missing variable in strict mode",
			template:     "Hello {{.missing_var}}!",
			variables:    map[string]interface{}{},
			expectError:  true,
			errorMessage: "failed to render template",
		},
		{
			name:     "invalid function call",
			template: "{{invalidFunction .name}}",
			variables: map[string]interface{}{
				"name": "test",
			},
			expectError:  true,
			errorMessage: "failed to compile template",
		},
		{
			name:     "type error in function",
			template: "{{add .name .age}}", // Can't add string and int
			variables: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expectError: false, // Our add function handles this gracefully
		},
		{
			name:     "valid template",
			template: "Hello {{.name}}, you are {{.age}} years old!",
			variables: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			metadata := &template.TemplateMetadata{
				Name: tt.name,
			}

			tmpl, err := engine.CompileTemplate(ctx, tt.name, tt.template, metadata)

			// Check for compilation errors
			if tt.errorMessage == "failed to compile template" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				return
			}

			require.NoError(t, err)

			result, err := engine.RenderTemplate(ctx, tmpl, tt.variables)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

// TestTemplateEngine_FactoryIntegration tests integration with the Factory pattern
func TestTemplateEngine_FactoryIntegration(t *testing.T) {
	// This test would normally use the actual factory, but since we're in integration tests,
	// we'll test the factory integration separately when the full system is available
	t.Skip("Factory integration test - requires full system setup")
}
