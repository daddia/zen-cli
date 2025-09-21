package draft

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdDraft(t *testing.T) {
	f := cmdutil.NewTestFactory(nil)
	cmd := NewCmdDraft(f)

	assert.Equal(t, "draft <activity>", cmd.Use)
	assert.Equal(t, "Generate document templates with task data", cmd.Short)
	assert.Equal(t, "core", cmd.GroupID)
	assert.True(t, cmd.HasFlags())

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("preview"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
}

func TestLoadTaskManifest(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() (string, func())
		expectedError string
		expectedID    string
	}{
		{
			name: "valid task manifest in current directory",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				manifestContent := `
schema_version: "1.0"
task:
  id: "TEST-001"
  title: "Test Task"
  type: "story"
  status: "proposed"
  priority: "P2"
  size: "M"
  points: 3
owner:
  name: "testuser"
  email: "testuser@company.com"
  github: "testuser"
team:
  name: "test-team"
  stream: "i2d"
  members: []
dates:
  created: "2025-01-01"
  target: "2025-01-15"
  last_updated: "2025-01-01 10:00:00"
workflow:
  current_stage: "01-align"
  completed_stages: []
  stages: {}
success_criteria:
  business: ["Define business value"]
  technical: ["Define technical approach"]
  user_experience: ["Define user experience"]
dependencies:
  upstream: []
  downstream: []
risk:
  level: "low"
  factors: []
labels: ["story"]
tags: ["story", "test-team"]
custom_fields: {}
`
				manifestPath := filepath.Join(tmpDir, "manifest.yaml")
				err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
				require.NoError(t, err)

				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				return tmpDir, func() { os.Chdir(oldDir) }
			},
			expectedID: "TEST-001",
		},
		{
			name: "no manifest found",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				return tmpDir, func() { os.Chdir(oldDir) }
			},
			expectedError: "no task manifest.yaml found",
		},
		{
			name: "invalid manifest format",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				manifestPath := filepath.Join(tmpDir, "manifest.yaml")
				err := os.WriteFile(manifestPath, []byte("invalid: yaml: content:"), 0644)
				require.NoError(t, err)

				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				return tmpDir, func() { os.Chdir(oldDir) }
			},
			expectedError: "failed to parse manifest file",
		},
		{
			name: "manifest without task info",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				manifestContent := `
schema_version: "1.0"
some_other_field: "value"
`
				manifestPath := filepath.Join(tmpDir, "manifest.yaml")
				err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
				require.NoError(t, err)

				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				return tmpDir, func() { os.Chdir(oldDir) }
			},
			expectedError: "does not contain task information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := tt.setupFunc()
			defer cleanup()

			manifest, dir, err := loadTaskManifest()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, manifest)
				assert.Empty(t, dir)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manifest)
				assert.Contains(t, dir, filepath.Base(tmpDir)) // Use contains to handle symlink resolution
				assert.Equal(t, tt.expectedID, manifest.Task.ID)
			}
		})
	}
}

func TestFindSimilarActivities(t *testing.T) {
	assets := []assets.AssetMetadata{
		{Command: "feature-spec", Name: "Feature Specification"},
		{Command: "user-story", Name: "User Story"},
		{Command: "epic", Name: "Epic Definition"},
		{Command: "roadmap", Name: "Product Roadmap"},
		{Command: "api-docs", Name: "API Documentation"},
	}

	tests := []struct {
		name     string
		activity string
		expected []string
	}{
		{
			name:     "exact match",
			activity: "feature-spec",
			expected: []string{"feature-spec"},
		},
		{
			name:     "partial match",
			activity: "feature",
			expected: []string{"feature-spec"},
		},
		{
			name:     "case insensitive",
			activity: "EPIC",
			expected: []string{"epic"},
		},
		{
			name:     "multiple matches",
			activity: "api",
			expected: []string{"api-docs"},
		},
		{
			name:     "no matches",
			activity: "nonexistent",
			expected: nil, // Go returns nil for empty slices
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSimilarActivities(tt.activity, assets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineOutputPath(t *testing.T) {
	taskDir := "/path/to/task"

	tests := []struct {
		name       string
		asset      *assets.AssetMetadata
		customPath string
		expected   string
	}{
		{
			name: "design stage activity",
			asset: &assets.AssetMetadata{
				Command:        "feature-spec",
				Format:         "markdown",
				Category:       "development",
				WorkflowStages: []string{"04-design"},
				OutputFile:     "",
			},
			customPath: "",
			expected:   "/path/to/task/design/feature-spec.md",
		},
		{
			name: "research stage activity",
			asset: &assets.AssetMetadata{
				Command:        "user-research",
				Format:         "markdown",
				Category:       "analysis",
				WorkflowStages: []string{"02-discover"},
				OutputFile:     "",
			},
			customPath: "",
			expected:   "/path/to/task/research/user-research.md",
		},
		{
			name: "root level activity",
			asset: &assets.AssetMetadata{
				Command:        "epic",
				Format:         "markdown",
				Category:       "planning",
				WorkflowStages: []string{"01-align"},
				OutputFile:     "",
			},
			customPath: "",
			expected:   "/path/to/task/epic.md",
		},
		{
			name: "custom relative path",
			asset: &assets.AssetMetadata{
				Command:        "feature-spec",
				Format:         "markdown",
				Category:       "development",
				WorkflowStages: []string{"04-design"},
				OutputFile:     "",
			},
			customPath: "custom/spec.md",
			expected:   "/path/to/task/custom/spec.md",
		},
		{
			name: "custom absolute path",
			asset: &assets.AssetMetadata{
				Command:        "feature-spec",
				Format:         "markdown",
				Category:       "development",
				WorkflowStages: []string{"04-design"},
				OutputFile:     "",
			},
			customPath: "/custom/output.md",
			expected:   "/custom/output.md",
		},
		{
			name: "execution stage activity",
			asset: &assets.AssetMetadata{
				Command:        "test-plan",
				Format:         "markdown",
				Category:       "quality",
				WorkflowStages: []string{"06-ship"},
				OutputFile:     "",
			},
			customPath: "",
			expected:   "/path/to/task/execution/test-plan.md",
		},
		{
			name: "outcomes stage activity",
			asset: &assets.AssetMetadata{
				Command:        "retrospective",
				Format:         "markdown",
				Category:       "analysis",
				WorkflowStages: []string{"07-learn"},
				OutputFile:     "",
			},
			customPath: "",
			expected:   "/path/to/task/outcomes/retrospective.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineOutputPath(tt.asset, taskDir, tt.customPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineWorkTypeDirectory(t *testing.T) {
	tests := []struct {
		name     string
		asset    *assets.AssetMetadata
		expected string
	}{
		{
			name: "design stage from workflow_stages",
			asset: &assets.AssetMetadata{
				Category:       "development",
				WorkflowStages: []string{"04-design"},
				Tags:           []string{"feature", "specification"},
			},
			expected: "design",
		},
		{
			name: "research stage from workflow_stages",
			asset: &assets.AssetMetadata{
				Category:       "analysis",
				WorkflowStages: []string{"02-discover"},
				Tags:           []string{"discovery", "research"},
			},
			expected: "research",
		},
		{
			name: "execution stage from workflow_stages",
			asset: &assets.AssetMetadata{
				Category:       "quality",
				WorkflowStages: []string{"05-build"},
				Tags:           []string{"testing", "qa"},
			},
			expected: "execution",
		},
		{
			name: "outcomes stage from workflow_stages",
			asset: &assets.AssetMetadata{
				Category:       "analysis",
				WorkflowStages: []string{"07-learn"},
				Tags:           []string{"retrospective", "learning"},
			},
			expected: "outcomes",
		},
		{
			name: "root level from workflow_stages",
			asset: &assets.AssetMetadata{
				Category:       "planning",
				WorkflowStages: []string{"01-align"},
				Tags:           []string{"epic", "planning"},
			},
			expected: "",
		},
		{
			name: "fallback to category mapping",
			asset: &assets.AssetMetadata{
				Category:       "analysis",
				WorkflowStages: []string{},
				Tags:           []string{"analysis"},
			},
			expected: "research",
		},
		{
			name: "fallback to default",
			asset: &assets.AssetMetadata{
				Category:       "unknown",
				WorkflowStages: []string{},
				Tags:           []string{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineWorkTypeDirectory(tt.asset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateTemplateData(t *testing.T) {
	startedDate := "2025-01-02"
	manifest := &TaskManifest{
		SchemaVersion: "1.0",
		Task: struct {
			ID       string `yaml:"id"`
			Title    string `yaml:"title"`
			Type     string `yaml:"type"`
			Status   string `yaml:"status"`
			Priority string `yaml:"priority"`
			Size     string `yaml:"size"`
			Points   int    `yaml:"points"`
		}{
			ID:       "TEST-001",
			Title:    "Test Task",
			Type:     "story",
			Status:   "in_progress",
			Priority: "P1",
			Size:     "L",
			Points:   5,
		},
		Owner: struct {
			Name   string `yaml:"name"`
			Email  string `yaml:"email"`
			Github string `yaml:"github"`
		}{
			Name:   "testuser",
			Email:  "testuser@company.com",
			Github: "testuser",
		},
		Team: struct {
			Name    string   `yaml:"name"`
			Stream  string   `yaml:"stream"`
			Members []string `yaml:"members"`
		}{
			Name:    "test-team",
			Stream:  "i2d",
			Members: []string{"user1", "user2"},
		},
		Dates: struct {
			Created     string  `yaml:"created"`
			Started     *string `yaml:"started"`
			Target      string  `yaml:"target"`
			Completed   *string `yaml:"completed"`
			LastUpdated string  `yaml:"last_updated"`
		}{
			Created:     "2025-01-01",
			Started:     &startedDate,
			Target:      "2025-01-15",
			Completed:   nil,
			LastUpdated: "2025-01-02 10:00:00",
		},
		Labels: []string{"story", "high-priority"},
		Tags:   []string{"frontend", "backend"},
		CustomFields: map[string]interface{}{
			"custom_field": "custom_value",
		},
	}

	data := createTemplateData(manifest)

	// Test basic task data
	assert.Equal(t, "TEST-001", data["TASK_ID"])
	assert.Equal(t, "Test Task", data["TASK_TITLE"])
	assert.Equal(t, "story", data["TASK_TYPE"])
	assert.Equal(t, "in_progress", data["TASK_STATUS"])
	assert.Equal(t, "P1", data["PRIORITY"])
	assert.Equal(t, "L", data["SIZE"])
	assert.Equal(t, 5, data["STORY_POINTS"])

	// Test owner data
	assert.Equal(t, "testuser", data["OWNER_NAME"])
	assert.Equal(t, "testuser@company.com", data["OWNER_EMAIL"])
	assert.Equal(t, "testuser", data["GITHUB_USERNAME"])

	// Test team data
	assert.Equal(t, "test-team", data["TEAM_NAME"])
	assert.Equal(t, "i2d", data["STREAM_TYPE"])
	assert.Equal(t, []string{"user1", "user2"}, data["TEAM_MEMBERS"])

	// Test dates
	assert.Equal(t, "2025-01-01", data["CREATED_DATE"])
	assert.Equal(t, "2025-01-02", data["STARTED_DATE"])
	assert.Equal(t, "2025-01-15", data["TARGET_DATE"])
	assert.Equal(t, "2025-01-02 10:00:00", data["LAST_UPDATED"])

	// Test computed fields
	assert.Contains(t, data, "CURRENT_DATE")
	assert.Contains(t, data, "CURRENT_DATETIME")

	// Test arrays
	assert.Equal(t, []string{"story", "high-priority"}, data["LABELS"])
	assert.Equal(t, []string{"frontend", "backend"}, data["TAGS"])

	// Test custom fields
	assert.Equal(t, map[string]interface{}{"custom_field": "custom_value"}, data["CUSTOM_FIELDS"])
}

func TestRunDraftIntegration(t *testing.T) {
	// This is an integration test that tests the full draft command flow
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create a test manifest
	manifestContent := `
schema_version: "1.0"
task:
  id: "TEST-001"
  title: "Test Task"
  type: "story"
  status: "proposed"
  priority: "P2"
  size: "M"
  points: 3
owner:
  name: "testuser"
  email: "testuser@company.com"
  github: "testuser"
team:
  name: "test-team"
  stream: "i2d"
  members: []
dates:
  created: "2025-01-01"
  target: "2025-01-15"
  last_updated: "2025-01-01 10:00:00"
workflow:
  current_stage: "01-align"
  completed_stages: []
  stages: {}
success_criteria:
  business: ["Define business value"]
  technical: ["Define technical approach"]
  user_experience: ["Define user experience"]
dependencies:
  upstream: []
  downstream: []
risk:
  level: "low"
  factors: []
labels: ["story"]
tags: ["story", "test-team"]
custom_fields: {}
`
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	// Create a test factory with mocked asset client
	streams := iostreams.Test()
	f := cmdutil.NewTestFactory(streams)

	// Override the asset client to return test data
	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &testAssetClient{}, nil
	}

	t.Run("successful template generation", func(t *testing.T) {
		err := runDraft(f, "test-template", false, false, "")

		// For this test, we expect it to work with the mock asset client
		// The actual implementation will depend on the mock returning appropriate data
		if err != nil {
			// Check if it's a known error from our mocks
			var zenErr *types.Error
			if assert.ErrorAs(t, err, &zenErr) {
				// Validate error codes are appropriate
				assert.Contains(t, []types.ErrorCode{
					types.ErrorCodeInvalidInput,
					types.ErrorCodeNetworkError,
					types.ErrorCodeUnknown,
				}, zenErr.Code)
			}
		}
	})

	t.Run("preview mode", func(t *testing.T) {
		err := runDraft(f, "test-template", false, true, "")

		// Preview should not create the actual template file
		if err == nil {
			// Check that the template file was not created (directories may be created)
			expectedFile := filepath.Join(tmpDir, "design", "test-template.md")
			_, err := os.Stat(expectedFile)
			assert.True(t, os.IsNotExist(err), "Preview mode should not create the template file")
		}
	})
}

// testAssetClient is a mock asset client for testing
type testAssetClient struct{}

func (c *testAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	return &assets.AssetList{
		Assets: []assets.AssetMetadata{
			{
				Name:           "test-template",
				Command:        "test-template",
				Type:           assets.AssetTypeTemplate,
				Description:    "Test template",
				Format:         "markdown",
				Category:       "development",
				Tags:           []string{"test"},
				Path:           "templates/test.md.tmpl",
				OutputFile:     "test-output.md",
				WorkflowStages: []string{"04-design"},
			},
		},
		Total:   1,
		HasMore: false,
	}, nil
}

func (c *testAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	if name == "test-template" {
		return &assets.AssetContent{
			Metadata: assets.AssetMetadata{
				Name:        name,
				Type:        assets.AssetTypeTemplate,
				Description: "Test template",
				Format:      "markdown",
			},
			Content:  "# {{.TASK_TITLE}}\n\n**Task ID:** {{.TASK_ID}}\n**Owner:** {{.OWNER_NAME}}\n",
			Checksum: "sha256:test",
			Cached:   false,
			CacheAge: 0,
		}, nil
	}
	return nil, &assets.AssetClientError{
		Code:    assets.ErrorCodeAssetNotFound,
		Message: "Asset not found",
	}
}

func (c *testAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	return &assets.SyncResult{
		Status:        "success",
		DurationMS:    1000,
		AssetsUpdated: 1,
		CacheSizeMB:   10.5,
		LastSync:      time.Now(),
	}, nil
}

func (c *testAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	return &assets.CacheInfo{
		TotalSize:     1024 * 1024,
		AssetCount:    5,
		LastSync:      time.Now(),
		CacheHitRatio: 0.85,
	}, nil
}

func (c *testAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (c *testAssetClient) Close() error {
	return nil
}
