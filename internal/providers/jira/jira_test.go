package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/integration"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthManager implements auth.Manager for testing
type MockAuthManager struct {
	mock.Mock
}

func (m *MockAuthManager) Authenticate(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) GetCredentials(provider string) (string, error) {
	args := m.Called(provider)
	return args.String(0), args.Error(1)
}

func (m *MockAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	args := m.Called(ctx, provider)
	return args.Bool(0)
}

func (m *MockAuthManager) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAuthManager) DeleteCredentials(provider string) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	args := m.Called(provider)
	return args.Get(0).(*auth.ProviderInfo), args.Error(1)
}

func TestNewJiraProvider(t *testing.T) {
	config := &config.IntegrationProviderConfig{
		ServerURL:  "https://test.atlassian.net",
		ProjectKey: "TEST",
		AuthType:   "basic",
	}
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}

	provider := NewJiraProvider(config, logger, authManager)

	assert.NotNil(t, provider)
	assert.Equal(t, "jira", provider.Name())
	assert.Equal(t, config, provider.config)
}

func TestJiraProvider_Name(t *testing.T) {
	provider := createTestJiraProvider(t)

	assert.Equal(t, "jira", provider.Name())
}

func TestJiraProvider_GetFieldMapping(t *testing.T) {
	provider := createTestJiraProvider(t)

	mapping := provider.GetFieldMapping()

	assert.NotNil(t, mapping)
	assert.Contains(t, mapping, "task_id")
	assert.Contains(t, mapping, "title")
	assert.Equal(t, "key", mapping["task_id"])
	assert.Equal(t, "summary", mapping["title"])
}

func TestJiraProvider_MapToZen(t *testing.T) {
	provider := createTestJiraProvider(t)

	external := &integration.ExternalTaskData{
		ID:          "PROJ-123",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "In Progress",
		Priority:    "High",
		Assignee:    "John Doe",
		Created:     time.Now(),
		Updated:     time.Now(),
		Fields: map[string]interface{}{
			"key": "PROJ-123",
		},
	}

	zen, err := provider.MapToZen(external)

	assert.NoError(t, err)
	assert.NotNil(t, zen)
	assert.Equal(t, "PROJ-123", zen.ID)
	assert.Equal(t, "Test Task", zen.Title)
	assert.Equal(t, "Test Description", zen.Description)
	assert.Equal(t, "in_progress", zen.Status)
	assert.Equal(t, "P1", zen.Priority)
	assert.Equal(t, "John Doe", zen.Owner)
	assert.Contains(t, zen.Metadata, "external_system")
	assert.Equal(t, "jira", zen.Metadata["external_system"])
}

func TestJiraProvider_MapToExternal(t *testing.T) {
	provider := createTestJiraProvider(t)

	zen := &integration.ZenTaskData{
		ID:          "TASK-1",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "in_progress",
		Priority:    "P1",
		Owner:       "John Doe",
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	external, err := provider.MapToExternal(zen)

	assert.NoError(t, err)
	assert.NotNil(t, external)
	assert.Equal(t, "TASK-1", external.ID)
	assert.Equal(t, "Test Task", external.Title)
	assert.Equal(t, "Test Description", external.Description)
	assert.Equal(t, "In Progress", external.Status)
	assert.Equal(t, "High", external.Priority)
	assert.Equal(t, "John Doe", external.Assignee)
}

func TestJiraProvider_ValidateConnection(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/myself" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name": "test-user"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &config.IntegrationProviderConfig{
		ServerURL:  server.URL,
		ProjectKey: "TEST",
		AuthType:   "basic",
	}
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}
	authManager.On("GetCredentials", "jira").Return("test@example.com:token", nil)

	provider := NewJiraProvider(config, logger, authManager)

	err := provider.ValidateConnection(context.Background())

	assert.NoError(t, err)
	authManager.AssertExpectations(t)
}

func TestJiraProvider_ValidateConnection_Failure(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	config := &config.IntegrationProviderConfig{
		ServerURL:  server.URL,
		ProjectKey: "TEST",
		AuthType:   "basic",
	}
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}
	authManager.On("GetCredentials", "jira").Return("test@example.com:token", nil)

	provider := NewJiraProvider(config, logger, authManager)

	err := provider.ValidateConnection(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	authManager.AssertExpectations(t)
}

func TestJiraProvider_GetTaskData(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/issue/PROJ-123" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "10001",
				"key": "PROJ-123",
				"self": "https://test.atlassian.net/rest/api/3/issue/10001",
				"fields": {
					"summary": "Test Issue",
					"description": "Test Description",
					"created": "2023-01-01T00:00:00.000Z",
					"updated": "2023-01-01T00:00:00.000Z",
					"status": {
						"name": "In Progress"
					},
					"priority": {
						"name": "High"
					},
					"assignee": {
						"displayName": "John Doe"
					},
					"issuetype": {
						"name": "Task"
					}
				}
			}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &config.IntegrationProviderConfig{
		ServerURL:  server.URL,
		ProjectKey: "TEST",
		AuthType:   "basic",
	}
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}
	authManager.On("GetCredentials", "jira").Return("test@example.com:token", nil)

	provider := NewJiraProvider(config, logger, authManager)

	taskData, err := provider.GetTaskData(context.Background(), "PROJ-123")

	assert.NoError(t, err)
	assert.NotNil(t, taskData)
	assert.Equal(t, "PROJ-123", taskData.ID)
	assert.Equal(t, "Test Issue", taskData.Title)
	assert.Equal(t, "Test Description", taskData.Description)
	assert.Equal(t, "In Progress", taskData.Status)
	assert.Equal(t, "High", taskData.Priority)
	assert.Equal(t, "John Doe", taskData.Assignee)
	authManager.AssertExpectations(t)
}

func TestJiraProvider_mapStatus(t *testing.T) {
	provider := createTestJiraProvider(t)

	tests := []struct {
		jiraStatus string
		expected   string
	}{
		{"To Do", "not_started"},
		{"In Progress", "in_progress"},
		{"Done", "completed"},
		{"Closed", "completed"},
		{"Blocked", "blocked"},
		{"Unknown Status", "unknown_status"},
	}

	for _, tt := range tests {
		t.Run(tt.jiraStatus, func(t *testing.T) {
			result := provider.mapStatus(tt.jiraStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJiraProvider_mapPriority(t *testing.T) {
	provider := createTestJiraProvider(t)

	tests := []struct {
		jiraPriority string
		expected     string
	}{
		{"Highest", "P0"},
		{"High", "P1"},
		{"Medium", "P2"},
		{"Low", "P3"},
		{"Lowest", "P3"},
		{"Unknown", "P2"},
	}

	for _, tt := range tests {
		t.Run(tt.jiraPriority, func(t *testing.T) {
			result := provider.mapPriority(tt.jiraPriority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func createTestJiraProvider(t *testing.T) *JiraProvider {
	config := &config.IntegrationProviderConfig{
		ServerURL:  "https://test.atlassian.net",
		ProjectKey: "TEST",
		AuthType:   "basic",
		FieldMapping: map[string]string{
			"task_id":     "key",
			"title":       "summary",
			"description": "description",
			"status":      "status.name",
			"priority":    "priority.name",
			"assignee":    "assignee.displayName",
		},
	}
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}

	return NewJiraProvider(config, logger, authManager)
}
