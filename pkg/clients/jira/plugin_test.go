package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock auth manager for testing
type mockAuthManager struct {
	mock.Mock
}

func (m *mockAuthManager) GetCredentials(provider string) (string, error) {
	args := m.Called(provider)
	return args.String(0), args.Error(1)
}

func (m *mockAuthManager) Authenticate(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *mockAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *mockAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *mockAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	args := m.Called(ctx, provider)
	return args.Bool(0)
}

func (m *mockAuthManager) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockAuthManager) DeleteCredentials(provider string) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *mockAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	args := m.Called(provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.ProviderInfo), args.Error(1)
}

func createTestPlugin() (*Plugin, *mockAuthManager) {
	config := &PluginConfig{
		Name:       "jira",
		Version:    "1.0.0",
		Enabled:    true,
		BaseURL:    "https://test.atlassian.net",
		ProjectKey: "TEST",
		Timeout:    30 * time.Second,
		Auth: &AuthConfig{
			Type:           AuthTypeBasic,
			CredentialsRef: "jira_creds",
		},
	}

	logger := logging.NewBasic()
	authMgr := &mockAuthManager{}

	plugin := NewPlugin(config, logger, authMgr)

	return plugin, authMgr
}

func createMockJiraIssue() JiraIssue {
	return JiraIssue{
		ID:   "10001",
		Key:  "TEST-123",
		Self: "https://test.atlassian.net/rest/api/3/issue/10001",
		Fields: JiraFields{
			Summary:     "Test Issue",
			Description: "Test issue description",
			Created:     time.Now().Add(-24 * time.Hour),
			Updated:     time.Now().Add(-1 * time.Hour),
			Status: JiraStatus{
				ID:   "3",
				Name: "In Progress",
			},
			Priority: JiraPriority{
				ID:   "2",
				Name: "High",
			},
			Assignee: JiraUser{
				AccountID:    "12345",
				DisplayName:  "John Doe",
				EmailAddress: "john@example.com",
			},
			IssueType: JiraIssueType{
				ID:   "10001",
				Name: "Task",
			},
			Project: JiraProject{
				ID:   "10000",
				Key:  "TEST",
				Name: "Test Project",
			},
			Labels: []string{"backend", "api"},
		},
	}
}

func TestPlugin_Lifecycle(t *testing.T) {
	plugin, authMgr := createTestPlugin()

	// Test plugin identity
	assert.Equal(t, "jira", plugin.Name())
	assert.Equal(t, "1.0.0", plugin.Version())
	assert.NotEmpty(t, plugin.Description())

	// Test operation support
	assert.True(t, plugin.SupportsOperation(OperationTypeFetch))
	assert.True(t, plugin.SupportsOperation(OperationTypeCreate))
	assert.True(t, plugin.SupportsOperation(OperationTypeUpdate))
	assert.True(t, plugin.SupportsOperation(OperationTypeDelete))
	assert.True(t, plugin.SupportsOperation(OperationTypeSearch))
	assert.True(t, plugin.SupportsOperation(OperationTypeSync))

	authMgr.AssertExpectations(t)
}

func TestPlugin_FetchTask(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/issue/TEST-123")
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Contains(t, r.Header.Get("Authorization"), "Basic")

		issue := createMockJiraIssue()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	plugin, authMgr := createTestPlugin()
	plugin.config.BaseURL = server.URL

	authMgr.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil) // base64 "user:pass"

	// Test FetchTask
	taskData, err := plugin.FetchTask(context.Background(), "TEST-123", &FetchOptions{})
	require.NoError(t, err)

	assert.Equal(t, "TEST-123", taskData.ID)
	assert.Equal(t, "TEST-123", taskData.ExternalID)
	assert.Equal(t, "Test Issue", taskData.Title)
	assert.Equal(t, "Test issue description", taskData.Description)
	assert.Equal(t, "in_progress", taskData.Status)
	assert.Equal(t, "P1", taskData.Priority)
	assert.Equal(t, "task", taskData.Type)
	assert.Equal(t, "John Doe", taskData.Assignee)
	assert.Contains(t, taskData.ExternalURL, "browse/TEST-123")
	assert.Equal(t, "jira", taskData.Metadata["external_system"])
	assert.Equal(t, "TEST", taskData.Metadata["project_key"])

	authMgr.AssertExpectations(t)
}

func TestPlugin_FetchTask_WithRawData(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issue := createMockJiraIssue()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	plugin, authMgr := createTestPlugin()
	plugin.config.BaseURL = server.URL

	authMgr.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

	// Test FetchTask with raw data
	taskData, err := plugin.FetchTask(context.Background(), "TEST-123", &FetchOptions{
		IncludeRaw: true,
	})
	require.NoError(t, err)

	assert.NotNil(t, taskData.RawData)
	assert.Contains(t, taskData.RawData, "key")
	assert.Contains(t, taskData.RawData, "fields")

	authMgr.AssertExpectations(t)
}

func TestPlugin_CreateTask(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/rest/api/3/issue") {
			// Handle create request
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("Authorization"), "Basic")

			var createReq JiraCreateRequest
			err := json.NewDecoder(r.Body).Decode(&createReq)
			require.NoError(t, err)

			assert.Equal(t, "TEST", createReq.Fields.Project.Key)
			assert.Equal(t, "New Task", createReq.Fields.Summary)
			assert.Equal(t, "Task description", createReq.Fields.Description)
			assert.Equal(t, "Task", createReq.Fields.IssueType.Name)

			response := JiraCreateResponse{
				ID:   "10002",
				Key:  "TEST-124",
				Self: r.URL.Scheme + "://" + r.URL.Host + "/rest/api/3/issue/10002",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "GET" && strings.Contains(r.URL.Path, "/rest/api/3/issue/TEST-124") {
			// Handle fetch request (called after create)
			issue := createMockJiraIssue()
			issue.Key = "TEST-124"
			issue.Fields.Summary = "New Task"
			issue.Fields.Description = "Task description"

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(issue)
		}
	}))
	defer server.Close()

	plugin, authMgr := createTestPlugin()
	plugin.config.BaseURL = server.URL

	authMgr.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil).Times(2)

	// Test CreateTask
	taskData := &PluginTaskData{
		Title:       "New Task",
		Description: "Task description",
		Status:      "proposed",
		Priority:    "P2",
		Type:        "task",
	}

	createdTask, err := plugin.CreateTask(context.Background(), taskData, &CreateOptions{})
	require.NoError(t, err)

	assert.Equal(t, "TEST-124", createdTask.ID)
	assert.Equal(t, "TEST-124", createdTask.ExternalID)
	assert.Equal(t, "New Task", createdTask.Title)
	assert.Equal(t, "Task description", createdTask.Description)

	authMgr.AssertExpectations(t)
}

func TestPlugin_SearchTasks(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/search")
		assert.Contains(t, r.URL.RawQuery, "jql=")
		assert.Contains(t, r.URL.RawQuery, "project")

		response := JiraSearchResponse{
			Issues:     []JiraIssue{createMockJiraIssue()},
			Total:      1,
			MaxResults: 50,
			StartAt:    0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	plugin, authMgr := createTestPlugin()
	plugin.config.BaseURL = server.URL

	authMgr.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

	// Test SearchTasks
	query := &SearchQuery{
		Filters: map[string]interface{}{
			"status": "In Progress",
		},
	}

	tasks, err := plugin.SearchTasks(context.Background(), query, &SearchOptions{})
	require.NoError(t, err)

	assert.Len(t, tasks, 1)
	assert.Equal(t, "TEST-123", tasks[0].ID)
	assert.Equal(t, "Test Issue", tasks[0].Title)
	assert.Equal(t, "in_progress", tasks[0].Status)

	authMgr.AssertExpectations(t)
}

func TestPlugin_HealthCheck(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/rest/api/3/serverInfo")

		serverInfo := JiraServerInfo{
			Version:     "8.20.0",
			BuildNumber: 820000,
			ServerTime:  time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serverInfo)
	}))
	defer server.Close()

	plugin, authMgr := createTestPlugin()
	plugin.config.BaseURL = server.URL

	authMgr.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

	// Test health check
	health, err := plugin.HealthCheck(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "jira", health.Provider)
	assert.True(t, health.Healthy)
	assert.Greater(t, health.ResponseTime, time.Duration(0))
	assert.Empty(t, health.LastError)

	authMgr.AssertExpectations(t)
}

func TestPlugin_DataMapping(t *testing.T) {
	plugin, _ := createTestPlugin()

	// Test MapToZen
	jiraIssue := createMockJiraIssue()
	taskData, err := plugin.MapToZen(context.Background(), &jiraIssue)
	require.NoError(t, err)

	assert.Equal(t, "TEST-123", taskData.ID)
	assert.Equal(t, "Test Issue", taskData.Title)
	assert.Equal(t, "Test issue description", taskData.Description)
	assert.Equal(t, "in_progress", taskData.Status)
	assert.Equal(t, "P1", taskData.Priority)
	assert.Equal(t, "task", taskData.Type)
	assert.Equal(t, "John Doe", taskData.Assignee)
	assert.Equal(t, []string{"backend", "api"}, taskData.Labels)

	// Test MapToExternal
	zenTaskData := &PluginTaskData{
		ID:          "task-456",
		Title:       "Zen Task",
		Description: "Zen description",
		Status:      "completed",
		Priority:    "P2",
		Type:        "story",
		Labels:      []string{"frontend", "ui"},
	}

	externalData, err := plugin.MapToExternal(context.Background(), zenTaskData)
	require.NoError(t, err)

	createReq, ok := externalData.(*JiraCreateRequest)
	require.True(t, ok)

	assert.Equal(t, "TEST", createReq.Fields.Project.Key)
	assert.Equal(t, "Zen Task", createReq.Fields.Summary)
	assert.Equal(t, "Zen description", createReq.Fields.Description)
	assert.Equal(t, "Story", createReq.Fields.IssueType.Name)
	assert.Equal(t, "Medium", createReq.Fields.Priority.Name)
	assert.Equal(t, []string{"frontend", "ui"}, createReq.Fields.Labels)
}

func TestPlugin_StatusMapping(t *testing.T) {
	plugin, _ := createTestPlugin()

	// Test Jira to Zen status mapping
	tests := []struct {
		jiraStatus string
		zenStatus  string
	}{
		{"To Do", "proposed"},
		{"In Progress", "in_progress"},
		{"Done", "completed"},
		{"Blocked", "blocked"},
		{"Custom Status", "custom_status"},
	}

	for _, tt := range tests {
		t.Run(tt.jiraStatus, func(t *testing.T) {
			result := plugin.mapJiraStatusToZen(tt.jiraStatus)
			assert.Equal(t, tt.zenStatus, result)
		})
	}

	// Test Zen to Jira status mapping
	zenToJiraTests := []struct {
		zenStatus  string
		jiraStatus string
	}{
		{"proposed", "To Do"},
		{"in_progress", "In Progress"},
		{"completed", "Done"},
		{"blocked", "Blocked"},
	}

	for _, tt := range zenToJiraTests {
		t.Run(tt.zenStatus, func(t *testing.T) {
			result := plugin.mapZenStatusToJira(tt.zenStatus)
			assert.Equal(t, tt.jiraStatus, result)
		})
	}
}

func TestPlugin_PriorityMapping(t *testing.T) {
	plugin, _ := createTestPlugin()

	// Test Jira to Zen priority mapping
	tests := []struct {
		jiraPriority string
		zenPriority  string
	}{
		{"Highest", "P0"},
		{"Critical", "P0"},
		{"High", "P1"},
		{"Medium", "P2"},
		{"Low", "P3"},
		{"Lowest", "P3"},
		{"Unknown", "P2"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.jiraPriority, func(t *testing.T) {
			result := plugin.mapJiraPriorityToZen(tt.jiraPriority)
			assert.Equal(t, tt.zenPriority, result)
		})
	}
}

func TestPlugin_TypeMapping(t *testing.T) {
	plugin, _ := createTestPlugin()

	// Test Jira to Zen type mapping
	tests := []struct {
		jiraType string
		zenType  string
	}{
		{"Story", "story"},
		{"User Story", "story"},
		{"Bug", "bug"},
		{"Defect", "bug"},
		{"Epic", "epic"},
		{"Task", "task"},
		{"Spike", "spike"},
		{"Unknown Type", "task"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.jiraType, func(t *testing.T) {
			result := plugin.mapJiraTypeToZen(tt.jiraType)
			assert.Equal(t, tt.zenType, result)
		})
	}
}

func TestPlugin_ErrorHandling(t *testing.T) {
	plugin, authMgr := createTestPlugin()

	tests := []struct {
		name        string
		statusCode  int
		wantErr     bool
		errContains string
	}{
		{"unauthorized", 401, true, "authentication failed"},
		{"forbidden", 403, true, "authentication failed"},
		{"not found", 404, true, "resource not found"},
		{"rate limited", 429, true, "rate limit exceeded"},
		{"server error", 500, true, "server error"},
		{"bad gateway", 502, true, "server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns the test status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"errorMessages":["Test error"]}`))
			}))
			defer server.Close()

			plugin.config.BaseURL = server.URL
			authMgr.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

			_, err := plugin.FetchTask(context.Background(), "TEST-123", &FetchOptions{})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	authMgr.AssertExpectations(t)
}

func TestPlugin_Configuration(t *testing.T) {
	tests := []struct {
		name    string
		config  *PluginConfig
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &PluginConfig{
				Name:       "jira",
				Version:    "1.0.0",
				BaseURL:    "https://test.atlassian.net",
				ProjectKey: "TEST",
				Auth: &AuthConfig{
					Type:           AuthTypeBasic,
					CredentialsRef: "jira_creds",
				},
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: &PluginConfig{
				Name:       "jira",
				Version:    "1.0.0",
				ProjectKey: "TEST",
				Auth: &AuthConfig{
					Type:           AuthTypeBasic,
					CredentialsRef: "jira_creds",
				},
			},
			wantErr: true,
		},
		{
			name: "missing project key",
			config: &PluginConfig{
				Name:    "jira",
				Version: "1.0.0",
				BaseURL: "https://test.atlassian.net",
				Auth: &AuthConfig{
					Type:           AuthTypeBasic,
					CredentialsRef: "jira_creds",
				},
			},
			wantErr: true,
		},
		{
			name: "missing auth config",
			config: &PluginConfig{
				Name:       "jira",
				Version:    "1.0.0",
				BaseURL:    "https://test.atlassian.net",
				ProjectKey: "TEST",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewBasic()
			authMgr := &mockAuthManager{}

			plugin := NewPlugin(tt.config, logger, authMgr)
			err := plugin.validateConfig()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlugin_GetFieldMapping(t *testing.T) {
	plugin, _ := createTestPlugin()

	mapping := plugin.GetFieldMapping()
	require.NotNil(t, mapping)

	// Verify essential mappings exist
	mappingMap := make(map[string]PluginFieldMapping)
	for _, mapping := range mapping.Mappings {
		mappingMap[mapping.ZenField] = mapping
	}

	// Check required mappings
	assert.Contains(t, mappingMap, "id")
	assert.Contains(t, mappingMap, "title")
	assert.Contains(t, mappingMap, "status")
	assert.Contains(t, mappingMap, "type")

	// Check required fields are marked as required
	assert.True(t, mappingMap["id"].Required)
	assert.True(t, mappingMap["title"].Required)
	assert.True(t, mappingMap["status"].Required)
	assert.True(t, mappingMap["type"].Required)
}
