package create

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/filesystem"
	"github.com/daddia/zen/pkg/iostreams"
	zentemplate "github.com/daddia/zen/pkg/template"
	"github.com/daddia/zen/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidTaskTypes(t *testing.T) {
	expected := []string{"story", "bug", "epic", "spike", "task"}
	actual := ValidTaskTypes()
	assert.Equal(t, expected, actual)
}

func TestIsValidTaskType(t *testing.T) {
	tests := []struct {
		taskType string
		expected bool
	}{
		{"story", true},
		{"bug", true},
		{"epic", true},
		{"spike", true},
		{"task", true},
		{"invalid", false},
		{"", false},
		{"STORY", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			result := IsValidTaskType(tt.taskType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTaskID(t *testing.T) {
	tests := []struct {
		name        string
		taskID      string
		expectError bool
		errorCode   types.ErrorCode
	}{
		{
			name:        "valid task ID",
			taskID:      "PROJ-123",
			expectError: false,
		},
		{
			name:        "valid alphanumeric ID",
			taskID:      "ABC123",
			expectError: false,
		},
		{
			name:        "empty task ID",
			taskID:      "",
			expectError: true,
			errorCode:   types.ErrorCodeInvalidInput,
		},
		{
			name:        "too short task ID",
			taskID:      "AB",
			expectError: true,
			errorCode:   types.ErrorCodeInvalidInput,
		},
		{
			name:        "task ID with spaces",
			taskID:      "PROJ 123",
			expectError: true,
			errorCode:   types.ErrorCodeInvalidInput,
		},
		{
			name:        "task ID with invalid characters",
			taskID:      "PROJ/123",
			expectError: true,
			errorCode:   types.ErrorCodeInvalidInput,
		},
		{
			name:        "task ID with colon",
			taskID:      "PROJ:123",
			expectError: true,
			errorCode:   types.ErrorCodeInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTaskID(tt.taskID)
			if tt.expectError {
				require.Error(t, err)
				var typedErr *types.Error
				require.ErrorAs(t, err, &typedErr)
				assert.Equal(t, tt.errorCode, typedErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateTaskDirectories(t *testing.T) {
	tempDir := t.TempDir()
	taskDir := filepath.Join(tempDir, "PROJ-123")

	// Use the shared filesystem utilities
	fsManager := filesystem.New(logging.NewBasic())
	err := fsManager.CreateTaskDirectory(taskDir)
	require.NoError(t, err)

	// Check main task directory exists
	assert.DirExists(t, taskDir)

	// Check only essential subdirectories exist (simplified structure)
	expectedDirs := []string{
		".zenflow",
		"metadata",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(taskDir, dir)
		assert.DirExists(t, dirPath, "Directory %s should exist", dir)
	}

	// Check that work-type directories are NOT created by default
	workTypeDirs := []string{
		"research",
		"spikes",
		"design",
		"execution",
		"outcomes",
	}

	for _, dir := range workTypeDirs {
		dirPath := filepath.Join(taskDir, dir)
		assert.NoDirExists(t, dirPath, "Directory %s should NOT exist by default", dir)
	}
}

func TestBuildTemplateVariables(t *testing.T) {
	opts := &CreateOptions{
		TaskID:   "PROJ-123",
		TaskType: "story",
		Title:    "Test Story",
		Owner:    "testuser",
		Team:     "testteam",
		Priority: "P1",
	}

	variables := buildTemplateVariables(opts)

	// Test basic task information
	assert.Equal(t, "PROJ-123", variables["TASK_ID"])
	assert.Equal(t, "Test Story", variables["TASK_TITLE"])
	assert.Equal(t, "story", variables["TASK_TYPE"])
	assert.Equal(t, "proposed", variables["TASK_STATUS"])
	assert.Equal(t, "P1", variables["PRIORITY"])

	// Test ownership
	assert.Equal(t, "testuser", variables["OWNER_NAME"])
	assert.Equal(t, "testuser@company.com", variables["OWNER_EMAIL"])
	assert.Equal(t, "testuser", variables["GITHUB_USERNAME"])
	assert.Equal(t, "testteam", variables["TEAM_NAME"])

	// Test workflow stages
	workflowStages, ok := variables["WORKFLOW_STAGES"].([]map[string]interface{})
	require.True(t, ok, "WORKFLOW_STAGES should be a slice of maps")
	assert.Len(t, workflowStages, 7, "Should have 7 workflow stages")

	// Test first stage
	stage1 := workflowStages[0]
	assert.Equal(t, "01-align", stage1["id"])
	assert.Equal(t, "Align", stage1["name"])
	assert.Equal(t, "not_started", stage1["status"])
	assert.Equal(t, 0, stage1["progress"])

	// Test current stage
	assert.Equal(t, "01-align", variables["CURRENT_STAGE"])

	// Test defaults
	assert.Equal(t, "M", variables["SIZE"])
	assert.Equal(t, 3, variables["STORY_POINTS"])
	assert.Equal(t, "i2d", variables["STREAM_TYPE"])
}

func TestBuildTemplateVariablesWithDefaults(t *testing.T) {
	opts := &CreateOptions{
		TaskID:   "PROJ-123",
		TaskType: "bug",
		// Title, Owner, Team, Priority not set - should use defaults
	}

	// Set environment variable for owner default
	originalUser := os.Getenv("USER")
	os.Setenv("USER", "envuser")
	defer func() {
		if originalUser != "" {
			os.Setenv("USER", originalUser)
		} else {
			os.Unsetenv("USER")
		}
	}()

	variables := buildTemplateVariables(opts)

	// Test defaults
	assert.Equal(t, "New bug task", variables["TASK_TITLE"])
	assert.Equal(t, "envuser", variables["OWNER_NAME"])
	assert.Equal(t, "default", variables["TEAM_NAME"])
	assert.Equal(t, "P2", variables["PRIORITY"])
}

func TestGenerateFallbackIndexFile(t *testing.T) {
	tempDir := t.TempDir()

	variables := map[string]interface{}{
		"TASK_ID":          "PROJ-123",
		"TASK_TITLE":       "Test Story",
		"CURRENT_STAGE":    "01-align",
		"OWNER_NAME":       "testuser",
		"TEAM_NAME":        "testteam",
		"TASK_TYPE":        "story",
		"last_updated":     "January 1, 2025",
		"next_review_date": "January 8, 2025",
	}

	err := generateFallbackIndexFile(tempDir, variables)
	require.NoError(t, err)

	// Check file was created
	indexPath := filepath.Join(tempDir, "index.md")
	assert.FileExists(t, indexPath)

	// Read and verify content
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# PROJ-123: Test Story")
	assert.Contains(t, contentStr, "**Owner:** testuser")
	assert.Contains(t, contentStr, "**Team:** testteam")
	assert.Contains(t, contentStr, "**Type:** story")
	assert.Contains(t, contentStr, "### Current Stage: Align")
	assert.Contains(t, contentStr, "*Last updated: January 1, 2025")
	assert.Contains(t, contentStr, "Next review: January 8, 2025*")
}

func TestGenerateFallbackManifestFile(t *testing.T) {
	tempDir := t.TempDir()

	variables := map[string]interface{}{
		"TASK_ID":         "PROJ-123",
		"TASK_TITLE":      "Test Story",
		"TASK_TYPE":       "story",
		"TASK_STATUS":     "proposed",
		"PRIORITY":        "P1",
		"SIZE":            "M",
		"STORY_POINTS":    3,
		"OWNER_NAME":      "testuser",
		"OWNER_EMAIL":     "testuser@company.com",
		"GITHUB_USERNAME": "testuser",
		"TEAM_NAME":       "testteam",
		"STREAM_TYPE":     "i2d",
		"CREATED_DATE":    "2025-01-01",
		"TARGET_DATE":     "2025-01-15",
		"LAST_UPDATED":    "2025-01-01 10:00:00",
		"CURRENT_STAGE":   "01-align",
	}

	err := generateFallbackManifestFile(tempDir, variables)
	require.NoError(t, err)

	// Check file was created
	manifestPath := filepath.Join(tempDir, "manifest.yaml")
	assert.FileExists(t, manifestPath)

	// Read and verify content
	content, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, `id: "PROJ-123"`)
	assert.Contains(t, contentStr, `title: "Test Story"`)
	assert.Contains(t, contentStr, `type: "story"`)
	assert.Contains(t, contentStr, `priority: "P1"`)
	assert.Contains(t, contentStr, `name: "testuser"`)
	assert.Contains(t, contentStr, `current_stage: "01-align"`)
}

func TestGenerateFallbackTaskrcFile(t *testing.T) {
	tempDir := t.TempDir()

	variables := map[string]interface{}{
		"TASK_ID":    "PROJ-123",
		"TASK_TYPE":  "story",
		"OWNER_NAME": "testuser",
	}

	err := generateFallbackTaskrcFile(tempDir, variables)
	require.NoError(t, err)

	// Check file was created
	taskrcPath := filepath.Join(tempDir, ".taskrc.yaml")
	assert.FileExists(t, taskrcPath)

	// Read and verify content
	content, err := os.ReadFile(taskrcPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, `id: "PROJ-123"`)
	assert.Contains(t, contentStr, `type: "story"`)
	assert.Contains(t, contentStr, `- "testuser"`)
}

func TestCreateRun_WorkspaceNotInitialized(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactoryWithWorkspace(streams, false, false)

	opts := &CreateOptions{
		IO:               streams,
		WorkspaceManager: factory.WorkspaceManager,
		TemplateEngine:   factory.TemplateEngine,
		TaskID:           "PROJ-123",
		TaskType:         "story",
	}

	err := createRun(opts)
	require.Error(t, err)

	var typedErr *types.Error
	require.ErrorAs(t, err, &typedErr)
	assert.Equal(t, types.ErrorCodeWorkspaceNotInit, typedErr.Code)
	assert.Contains(t, typedErr.Message, "workspace not initialized")
}

func TestCreateRun_TaskAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactoryWithWorkspace(streams, true, false)

	// Create existing task directory
	taskDir := filepath.Join(tempDir, ".zen", "work", "tasks", "PROJ-123")
	err := os.MkdirAll(taskDir, 0755)
	require.NoError(t, err)

	// Mock workspace manager to return our temp directory
	factory.WorkspaceManager = func() (cmdutil.WorkspaceManager, error) {
		return &mockWorkspaceManager{
			root:        tempDir,
			initialized: true,
		}, nil
	}

	opts := &CreateOptions{
		IO:               streams,
		WorkspaceManager: factory.WorkspaceManager,
		TemplateEngine:   factory.TemplateEngine,
		TaskID:           "PROJ-123",
		TaskType:         "story",
	}

	err = createRun(opts)
	require.Error(t, err)

	var typedErr *types.Error
	require.ErrorAs(t, err, &typedErr)
	assert.Equal(t, types.ErrorCodeAlreadyExists, typedErr.Code)
	assert.Contains(t, typedErr.Message, "task 'PROJ-123' already exists")
}

func TestCreateRun_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactoryWithWorkspace(streams, true, false)

	// Mock workspace manager to return our temp directory
	factory.WorkspaceManager = func() (cmdutil.WorkspaceManager, error) {
		return &mockWorkspaceManager{
			root:        tempDir,
			initialized: true,
		}, nil
	}

	opts := &CreateOptions{
		IO:               streams,
		WorkspaceManager: factory.WorkspaceManager,
		TemplateEngine:   factory.TemplateEngine,
		TaskID:           "PROJ-123",
		TaskType:         "story",
		DryRun:           true,
	}

	err := createRun(opts)
	require.NoError(t, err)

	// Check output contains dry run information
	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, "Task creation plan for PROJ-123")
	assert.Contains(t, output, "Type: story")
	assert.Contains(t, output, "Run without --dry-run to create the task")

	// Check no actual directories were created
	taskDir := filepath.Join(tempDir, ".zen", "work", "tasks", "PROJ-123")
	assert.NoDirExists(t, taskDir)
}

func TestCreateRun_Success(t *testing.T) {
	tempDir := t.TempDir()
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactoryWithWorkspace(streams, true, false)

	// Mock workspace manager to return our temp directory
	factory.WorkspaceManager = func() (cmdutil.WorkspaceManager, error) {
		return &mockWorkspaceManager{
			root:        tempDir,
			initialized: true,
		}, nil
	}

	// Mock template engine to return errors so fallback functions are used
	factory.TemplateEngine = func() (cmdutil.TemplateEngineInterface, error) {
		return &mockTemplateEngine{}, nil
	}

	opts := &CreateOptions{
		IO:               streams,
		WorkspaceManager: factory.WorkspaceManager,
		TemplateEngine:   factory.TemplateEngine,
		TaskID:           "PROJ-123",
		TaskType:         "story",
		Title:            "Test Story",
		Owner:            "testuser",
		Team:             "testteam",
		Priority:         "P1",
	}

	err := createRun(opts)
	require.NoError(t, err)

	// Check success output
	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, "Created task PROJ-123")
	assert.Contains(t, output, "Type: story")
	assert.Contains(t, output, "Next steps:")

	// Check task directory structure was created
	taskDir := filepath.Join(tempDir, ".zen", "work", "tasks", "PROJ-123")
	assert.DirExists(t, taskDir)

	// Check only essential directories are created
	expectedDirs := []string{
		".zenflow", "metadata",
	}
	for _, dir := range expectedDirs {
		assert.DirExists(t, filepath.Join(taskDir, dir))
	}

	// Check that work-type directories are NOT created by default
	workTypeDirs := []string{
		"research", "spikes", "design", "execution", "outcomes",
	}
	for _, dir := range workTypeDirs {
		assert.NoDirExists(t, filepath.Join(taskDir, dir), "Work-type directory %s should not be created by default", dir)
	}

	// Check files were created
	assert.FileExists(t, filepath.Join(taskDir, "index.md"))
	assert.FileExists(t, filepath.Join(taskDir, "manifest.yaml"))
	assert.FileExists(t, filepath.Join(taskDir, ".taskrc.yaml"))

	// Verify file contents
	indexContent, err := os.ReadFile(filepath.Join(taskDir, "index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(indexContent), "PROJ-123: Test Story")
	assert.Contains(t, string(indexContent), "testuser")
	assert.Contains(t, string(indexContent), "testteam")

	manifestContent, err := os.ReadFile(filepath.Join(taskDir, "manifest.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(manifestContent), `id: "PROJ-123"`)
	assert.Contains(t, string(manifestContent), `title: "Test Story"`)

	taskrcContent, err := os.ReadFile(filepath.Join(taskDir, ".taskrc.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(taskrcContent), `id: "PROJ-123"`)
	assert.Contains(t, string(taskrcContent), `type: "story"`)
}

// mockWorkspaceManager is a simple mock for testing
type mockWorkspaceManager struct {
	root        string
	initialized bool
}

func (m *mockWorkspaceManager) Root() string {
	return m.root
}

func (m *mockWorkspaceManager) ConfigFile() string {
	return filepath.Join(m.root, ".zen", "config.yaml")
}

func (m *mockWorkspaceManager) Initialize() error {
	m.initialized = true
	return nil
}

func (m *mockWorkspaceManager) InitializeWithForce(force bool) error {
	m.initialized = true
	return nil
}

func (m *mockWorkspaceManager) Status() (cmdutil.WorkspaceStatus, error) {
	return cmdutil.WorkspaceStatus{
		Initialized: m.initialized,
		ConfigPath:  m.ConfigFile(),
		Root:        m.root,
		Project: cmdutil.ProjectInfo{
			Type:     "test",
			Name:     "test-project",
			Language: "go",
		},
	}, nil
}

// mockTemplateEngine returns errors for LoadTemplate to force fallback usage
type mockTemplateEngine struct{}

func (e *mockTemplateEngine) LoadTemplate(ctx context.Context, name string) (*zentemplate.Template, error) {
	return nil, fmt.Errorf("template %s not found", name)
}

func (e *mockTemplateEngine) RenderTemplate(ctx context.Context, tmpl *zentemplate.Template, variables map[string]interface{}) (string, error) {
	return "", fmt.Errorf("render not supported")
}

func (e *mockTemplateEngine) ListTemplates(ctx context.Context, filter zentemplate.TemplateFilter) (*zentemplate.TemplateList, error) {
	return &zentemplate.TemplateList{}, nil
}

func (e *mockTemplateEngine) ValidateVariables(ctx context.Context, tmpl *zentemplate.Template, variables map[string]interface{}) error {
	return nil
}

func (e *mockTemplateEngine) CompileTemplate(ctx context.Context, name, content string, metadata *zentemplate.TemplateMetadata) (*zentemplate.Template, error) {
	return nil, fmt.Errorf("compile not supported")
}

func (e *mockTemplateEngine) GetFunctions() template.FuncMap {
	return template.FuncMap{}
}

func TestGenerateTaskFiles_TemplateSuccess(t *testing.T) {
	tempDir := t.TempDir()
	taskDir := filepath.Join(tempDir, "PROJ-123")

	// Create task directory
	err := os.MkdirAll(taskDir, 0755)
	require.NoError(t, err)

	// Mock template engine that succeeds
	templateEngine := &successTemplateEngine{}

	opts := &CreateOptions{
		TaskID:   "PROJ-123",
		TaskType: "story",
		Title:    "Test Story",
		Owner:    "testuser",
		TemplateEngine: func() (cmdutil.TemplateEngineInterface, error) {
			return templateEngine, nil
		},
	}

	ctx := context.Background()
	err = generateTaskFiles(ctx, opts, taskDir)
	require.NoError(t, err)

	// Check that files were created
	assert.FileExists(t, filepath.Join(taskDir, "index.md"))
	assert.FileExists(t, filepath.Join(taskDir, "manifest.yaml"))
	assert.FileExists(t, filepath.Join(taskDir, ".taskrc.yaml"))
}

func TestGenerateTaskFiles_TemplateEngineError(t *testing.T) {
	tempDir := t.TempDir()
	taskDir := filepath.Join(tempDir, "PROJ-123")

	opts := &CreateOptions{
		TaskID:   "PROJ-123",
		TaskType: "story",
		TemplateEngine: func() (cmdutil.TemplateEngineInterface, error) {
			return nil, fmt.Errorf("template engine failed")
		},
	}

	ctx := context.Background()
	err := generateTaskFiles(ctx, opts, taskDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get template engine")
}

func TestBuildTemplateVariables_EdgeCases(t *testing.T) {
	// Test with minimal options
	opts := &CreateOptions{
		TaskID:   "MIN-123",
		TaskType: "task",
		// No title, owner, team, priority - should use defaults
	}

	// Clear environment variables
	originalUser := os.Getenv("USER")
	os.Unsetenv("USER")
	defer func() {
		if originalUser != "" {
			os.Setenv("USER", originalUser)
		}
	}()

	variables := buildTemplateVariables(opts)

	// Test defaults when no environment
	assert.Equal(t, "MIN-123", variables["TASK_ID"])
	assert.Equal(t, "New task task", variables["TASK_TITLE"])
	assert.Equal(t, "unknown", variables["OWNER_NAME"])
	assert.Equal(t, "default", variables["TEAM_NAME"])
	assert.Equal(t, "P2", variables["PRIORITY"])
}

func TestValidateTaskID_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		taskID      string
		expectError bool
	}{
		{
			name:        "exactly 3 characters",
			taskID:      "ABC",
			expectError: false,
		},
		{
			name:        "with numbers",
			taskID:      "PROJ-123-SUB",
			expectError: false,
		},
		{
			name:        "with underscores",
			taskID:      "PROJ_123",
			expectError: false,
		},
		{
			name:        "unicode characters",
			taskID:      "PROJ-ñ",
			expectError: false,
		},
		{
			name:        "question mark",
			taskID:      "PROJ?123",
			expectError: true,
		},
		{
			name:        "pipe character",
			taskID:      "PROJ|123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTaskID(tt.taskID)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// successTemplateEngine simulates successful template operations
type successTemplateEngine struct{}

func (e *successTemplateEngine) LoadTemplate(ctx context.Context, name string) (*zentemplate.Template, error) {
	return &zentemplate.Template{
		Name:    name,
		Content: fmt.Sprintf("# Template %s\n{{.TASK_ID}}: {{.TASK_TITLE}}", name),
	}, nil
}

func (e *successTemplateEngine) RenderTemplate(ctx context.Context, tmpl *zentemplate.Template, variables map[string]interface{}) (string, error) {
	return fmt.Sprintf("# Template %s\n%s: %s", tmpl.Name, variables["TASK_ID"], variables["TASK_TITLE"]), nil
}

func (e *successTemplateEngine) ListTemplates(ctx context.Context, filter zentemplate.TemplateFilter) (*zentemplate.TemplateList, error) {
	return &zentemplate.TemplateList{}, nil
}

func (e *successTemplateEngine) ValidateVariables(ctx context.Context, tmpl *zentemplate.Template, variables map[string]interface{}) error {
	return nil
}

func (e *successTemplateEngine) CompileTemplate(ctx context.Context, name, content string, metadata *zentemplate.TemplateMetadata) (*zentemplate.Template, error) {
	return &zentemplate.Template{Name: name, Content: content}, nil
}

func (e *successTemplateEngine) GetFunctions() template.FuncMap {
	return template.FuncMap{}
}

func TestNewCmdTaskCreate(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdTaskCreate(factory)

	require.NotNil(t, cmd)
	assert.Equal(t, "create <task-id>", cmd.Use)
	assert.Equal(t, "Create a new task with structured workflow", cmd.Short)
	assert.Contains(t, cmd.Long, "Create a new task with structured workflow directories")
	assert.Contains(t, cmd.Example, "zen task create USER-123")

	// Check flags
	typeFlag := cmd.Flags().Lookup("type")
	require.NotNil(t, typeFlag)
	assert.Equal(t, "", typeFlag.DefValue)

	titleFlag := cmd.Flags().Lookup("title")
	require.NotNil(t, titleFlag)

	ownerFlag := cmd.Flags().Lookup("owner")
	require.NotNil(t, ownerFlag)

	teamFlag := cmd.Flags().Lookup("team")
	require.NotNil(t, teamFlag)

	priorityFlag := cmd.Flags().Lookup("priority")
	require.NotNil(t, priorityFlag)
	assert.Equal(t, "P2", priorityFlag.DefValue)
}

func TestNewCmdTaskCreate_MissingTypeFlag(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdTaskCreate(factory)
	cmd.SetArgs([]string{"PROJ-123"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \"type\" not set")
}

func TestNewCmdTaskCreate_InvalidType(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdTaskCreate(factory)
	cmd.SetArgs([]string{"PROJ-123", "--type", "invalid"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task type 'invalid'")
}
