package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/integration"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/clients"
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

func TestProvider_GetTaskData(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/issue/PROJ-123")
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		// Mock Jira issue response
		issue := JiraIssue{
			ID:  "10001",
			Key: "PROJ-123",
			Fields: struct {
				Summary     string    `json:"summary"`
				Description string    `json:"description"`
				Created     time.Time `json:"created"`
				Updated     time.Time `json:"updated"`
				Status      struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				} `json:"status"`
				Priority struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				} `json:"priority"`
				Assignee struct {
					DisplayName  string `json:"displayName"`
					EmailAddress string `json:"emailAddress"`
					AccountID    string `json:"accountId"`
				} `json:"assignee"`
				IssueType struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				} `json:"issuetype"`
				Project struct {
					Key  string `json:"key"`
					Name string `json:"name"`
					ID   string `json:"id"`
				} `json:"project"`
			}{
				Summary:     "Test Issue",
				Description: "Test issue description",
				Created:     time.Now().Add(-24 * time.Hour),
				Updated:     time.Now().Add(-1 * time.Hour),
				Status: struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				}{Name: "In Progress", ID: "3"},
				Priority: struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				}{Name: "High", ID: "2"},
				Assignee: struct {
					DisplayName  string `json:"displayName"`
					EmailAddress string `json:"emailAddress"`
					AccountID    string `json:"accountId"`
				}{DisplayName: "John Doe", EmailAddress: "john@example.com"},
				IssueType: struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				}{Name: "Task", ID: "10001"},
				Project: struct {
					Key  string `json:"key"`
					Name string `json:"name"`
					ID   string `json:"id"`
				}{Key: "PROJ", Name: "Test Project", ID: "10000"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	// Create provider
	config := &config.IntegrationProviderConfig{
		ServerURL:      server.URL,
		ProjectKey:     "PROJ",
		AuthType:       "basic",
		CredentialsRef: "jira_creds",
	}

	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	authManager.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil) // base64 "user:pass"

	provider := NewProvider(config, logger, authManager)

	// Test GetTaskData
	taskData, err := provider.GetTaskData(context.Background(), "PROJ-123")
	require.NoError(t, err)

	assert.Equal(t, "PROJ-123", taskData.ID)
	assert.Equal(t, "Test Issue", taskData.Title)
	assert.Equal(t, "Test issue description", taskData.Description)
	assert.Equal(t, "In Progress", taskData.Status)
	assert.Equal(t, "High", taskData.Priority)
	assert.Equal(t, "John Doe", taskData.Assignee)
	assert.Contains(t, taskData.Fields, "issue_type")
	assert.Equal(t, "Task", taskData.Fields["issue_type"])

	authManager.AssertExpectations(t)
}

func TestProvider_CreateTask(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/issue")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var createReq JiraCreateIssueRequest
		err := json.NewDecoder(r.Body).Decode(&createReq)
		require.NoError(t, err)

		assert.Equal(t, "PROJ", createReq.Fields.Project.Key)
		assert.Equal(t, "New Task", createReq.Fields.Summary)
		assert.Equal(t, "Task description", createReq.Fields.Description)
		assert.Equal(t, "Task", createReq.Fields.IssueType.Name)

		// Mock create response
		response := map[string]interface{}{
			"id":   "10002",
			"key":  "PROJ-124",
			"self": r.URL.String() + "/rest/api/3/issue/10002",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider
	config := &config.IntegrationProviderConfig{
		ServerURL:      server.URL,
		ProjectKey:     "PROJ",
		AuthType:       "token",
		CredentialsRef: "jira_token",
	}

	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	authManager.On("GetCredentials", "jira_token").Return("test-token", nil)

	provider := NewProvider(config, logger, authManager)

	// Test CreateTask
	zenData := &integration.ZenTaskData{
		ID:          "task-new",
		Title:       "New Task",
		Description: "Task description",
		Status:      "todo",
		Priority:    "high",
	}

	externalData, err := provider.CreateTask(context.Background(), zenData)
	require.NoError(t, err)

	assert.Equal(t, "PROJ-124", externalData.ID)
	assert.Equal(t, "New Task", externalData.Title)
	assert.Equal(t, "Task description", externalData.Description)
	assert.Contains(t, externalData.Fields, "jira_id")
	assert.Equal(t, "10002", externalData.Fields["jira_id"])

	authManager.AssertExpectations(t)
}

func TestProvider_UpdateTask(t *testing.T) {
	// Create mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && strings.Contains(r.URL.Path, "/rest/api/3/issue/PROJ-123") {
			// Handle update request
			var updateReq JiraUpdateIssueRequest
			err := json.NewDecoder(r.Body).Decode(&updateReq)
			require.NoError(t, err)

			assert.Equal(t, "Updated Task", updateReq.Fields["summary"])
			assert.Equal(t, "Updated description", updateReq.Fields["description"])

			w.WriteHeader(http.StatusNoContent)
		} else if r.Method == "GET" && strings.Contains(r.URL.Path, "/rest/api/3/issue/PROJ-123") {
			// Handle get request (called after update)
			issue := JiraIssue{
				Key: "PROJ-123",
				Fields: struct {
					Summary     string    `json:"summary"`
					Description string    `json:"description"`
					Created     time.Time `json:"created"`
					Updated     time.Time `json:"updated"`
					Status      struct {
						Name string `json:"name"`
						ID   string `json:"id"`
					} `json:"status"`
					Priority struct {
						Name string `json:"name"`
						ID   string `json:"id"`
					} `json:"priority"`
					Assignee struct {
						DisplayName  string `json:"displayName"`
						EmailAddress string `json:"emailAddress"`
						AccountID    string `json:"accountId"`
					} `json:"assignee"`
					IssueType struct {
						Name string `json:"name"`
						ID   string `json:"id"`
					} `json:"issuetype"`
					Project struct {
						Key  string `json:"key"`
						Name string `json:"name"`
						ID   string `json:"id"`
					} `json:"project"`
				}{
					Summary:     "Updated Task",
					Description: "Updated description",
					Updated:     time.Now(),
					Status: struct {
						Name string `json:"name"`
						ID   string `json:"id"`
					}{Name: "In Progress"},
					Priority: struct {
						Name string `json:"name"`
						ID   string `json:"id"`
					}{Name: "Medium"},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(issue)
		}
	}))
	defer server.Close()

	// Create provider
	config := &config.IntegrationProviderConfig{
		ServerURL:      server.URL,
		ProjectKey:     "PROJ",
		AuthType:       "basic",
		CredentialsRef: "jira_creds",
	}

	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	authManager.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

	provider := NewProvider(config, logger, authManager)

	// Test UpdateTask
	zenData := &integration.ZenTaskData{
		ID:          "task-123",
		Title:       "Updated Task",
		Description: "Updated description",
		Status:      "in_progress",
		Priority:    "medium",
	}

	externalData, err := provider.UpdateTask(context.Background(), "PROJ-123", zenData)
	require.NoError(t, err)

	assert.Equal(t, "PROJ-123", externalData.ID)
	assert.Equal(t, "Updated Task", externalData.Title)
	assert.Equal(t, "Updated description", externalData.Description)

	authManager.AssertExpectations(t)
}

func TestProvider_ValidateConnection(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
	}{
		{
			name: "successful validation",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Path, "/rest/api/3/serverInfo")
				w.Header().Set("Content-Type", "application/json")
				response := map[string]interface{}{
					"version":     "8.20.0",
					"buildNumber": 820000,
				}
				json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
		},
		{
			name: "authentication failure",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"errorMessages":["Authentication failed"]}`))
			},
			expectedError: true,
		},
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"errorMessages":["Internal server error"]}`))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create provider
			config := &config.IntegrationProviderConfig{
				ServerURL:      server.URL,
				ProjectKey:     "PROJ",
				AuthType:       "basic",
				CredentialsRef: "jira_creds",
			}

			logger := logging.NewBasic()
			authManager := &mockAuthManager{}
			authManager.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

			provider := NewProvider(config, logger, authManager)

			// Test connection validation
			err := provider.ValidateConnection(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			authManager.AssertExpectations(t)
		})
	}
}

func TestProvider_HealthCheck(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"version": "8.20.0",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider
	config := &config.IntegrationProviderConfig{
		ServerURL:      server.URL,
		ProjectKey:     "PROJ",
		AuthType:       "basic",
		CredentialsRef: "jira_creds",
	}

	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	authManager.On("GetCredentials", "jira_creds").Return("dXNlcjpwYXNz", nil)

	provider := NewProvider(config, logger, authManager)

	// Test health check
	health, err := provider.HealthCheck(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "jira", health.Provider)
	assert.True(t, health.Healthy)
	assert.Greater(t, health.ResponseTime, time.Duration(0))
	assert.NotNil(t, health.RateLimitInfo)

	authManager.AssertExpectations(t)
}

func TestProvider_SearchTasks(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/search")
		// JQL parameter should be present but may be URL encoded

		// Mock search response
		response := JiraSearchResponse{
			Issues: []JiraIssue{
				{
					Key: "PROJ-123",
					Fields: struct {
						Summary     string    `json:"summary"`
						Description string    `json:"description"`
						Created     time.Time `json:"created"`
						Updated     time.Time `json:"updated"`
						Status      struct {
							Name string `json:"name"`
							ID   string `json:"id"`
						} `json:"status"`
						Priority struct {
							Name string `json:"name"`
							ID   string `json:"id"`
						} `json:"priority"`
						Assignee struct {
							DisplayName  string `json:"displayName"`
							EmailAddress string `json:"emailAddress"`
							AccountID    string `json:"accountId"`
						} `json:"assignee"`
						IssueType struct {
							Name string `json:"name"`
							ID   string `json:"id"`
						} `json:"issuetype"`
						Project struct {
							Key  string `json:"key"`
							Name string `json:"name"`
							ID   string `json:"id"`
						} `json:"project"`
					}{
						Summary: "Search Result 1",
						Status: struct {
							Name string `json:"name"`
							ID   string `json:"id"`
						}{Name: "To Do"},
					},
				},
			},
			Total:      1,
			MaxResults: 50,
			StartAt:    0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider
	config := &config.IntegrationProviderConfig{
		ServerURL:      server.URL,
		ProjectKey:     "PROJ",
		AuthType:       "token",
		CredentialsRef: "jira_token",
	}

	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	authManager.On("GetCredentials", "jira_token").Return("test-token", nil)

	provider := NewProvider(config, logger, authManager)

	// Test search
	query := map[string]interface{}{
		"status": "To Do",
	}

	tasks, err := provider.SearchTasks(context.Background(), query)
	require.NoError(t, err)

	assert.Len(t, tasks, 1)
	assert.Equal(t, "PROJ-123", tasks[0].ID)
	assert.Equal(t, "Search Result 1", tasks[0].Title)
	assert.Equal(t, "To Do", tasks[0].Status)

	authManager.AssertExpectations(t)
}

func TestProvider_DataMapping(t *testing.T) {
	config := &config.IntegrationProviderConfig{
		ServerURL:  "https://test.atlassian.net",
		ProjectKey: "PROJ",
	}

	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	provider := NewProvider(config, logger, authManager)

	// Test MapToZen
	externalData := &integration.ExternalTaskData{
		ID:          "PROJ-123",
		Title:       "External Task",
		Description: "External description",
		Status:      "In Progress",
		Priority:    "High",
		Assignee:    "John Doe",
		Created:     time.Now().Add(-24 * time.Hour),
		Updated:     time.Now().Add(-1 * time.Hour),
		Fields: map[string]interface{}{
			"issue_type": "Bug",
			"project":    "PROJ",
		},
	}

	zenData, err := provider.MapToZen(externalData)
	require.NoError(t, err)

	assert.Equal(t, "PROJ-123", zenData.ID)
	assert.Equal(t, "External Task", zenData.Title)
	assert.Equal(t, "External description", zenData.Description)
	assert.Equal(t, "in_progress", zenData.Status) // Mapped from "In Progress"
	assert.Equal(t, "high", zenData.Priority)      // Mapped from "High"
	assert.Equal(t, "John Doe", zenData.Owner)
	assert.Equal(t, "jira", zenData.Metadata["external_system"])

	// Test MapToExternal
	zenData2 := &integration.ZenTaskData{
		ID:          "task-456",
		Title:       "Zen Task",
		Description: "Zen description",
		Status:      "completed",
		Priority:    "medium",
		Owner:       "Jane Doe",
	}

	externalData2, err := provider.MapToExternal(zenData2)
	require.NoError(t, err)

	assert.Equal(t, "task-456", externalData2.ID)
	assert.Equal(t, "Zen Task", externalData2.Title)
	assert.Equal(t, "Zen description", externalData2.Description)
	assert.Equal(t, "Done", externalData2.Status)     // Mapped from "completed"
	assert.Equal(t, "Medium", externalData2.Priority) // Mapped from "medium"
	assert.Equal(t, "Jane Doe", externalData2.Assignee)
	assert.Equal(t, "PROJ", externalData2.Fields["project_key"])
}

func TestProvider_StatusMapping(t *testing.T) {
	config := &config.IntegrationProviderConfig{}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	provider := NewProvider(config, logger, authManager)

	// Test Jira to Zen status mapping
	tests := []struct {
		jiraStatus string
		zenStatus  string
	}{
		{"To Do", "todo"},
		{"In Progress", "in_progress"},
		{"Done", "completed"},
		{"Blocked", "blocked"},
		{"Custom Status", "custom_status"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("jira_%s_to_zen_%s", tt.jiraStatus, tt.zenStatus), func(t *testing.T) {
			result := provider.mapJiraStatusToZen(tt.jiraStatus)
			assert.Equal(t, tt.zenStatus, result)
		})
	}

	// Test Zen to Jira status mapping
	zenToJiraTests := []struct {
		zenStatus  string
		jiraStatus string
	}{
		{"todo", "To Do"},
		{"in_progress", "In Progress"},
		{"completed", "Done"},
		{"blocked", "Blocked"},
		{"custom", "custom"},
	}

	for _, tt := range zenToJiraTests {
		t.Run(fmt.Sprintf("zen_%s_to_jira_%s", tt.zenStatus, tt.jiraStatus), func(t *testing.T) {
			result := provider.mapZenStatusToJira(tt.zenStatus)
			assert.Equal(t, tt.jiraStatus, result)
		})
	}
}

func TestProvider_ErrorHandling(t *testing.T) {
	// Test different HTTP error scenarios
	tests := []struct {
		name              string
		statusCode        int
		expectedCode      string
		expectedRetryable bool
	}{
		{"bad request", 400, "INVALID_REQUEST", false},
		{"unauthorized", 401, "AUTHENTICATION_FAILED", false},
		{"forbidden", 403, "AUTHENTICATION_FAILED", false},
		{"not found", 404, "NOT_FOUND", false},
		{"rate limited", 429, "RATE_LIMITED", true},
		{"server error", 500, "INTERNAL_ERROR", true},
		{"bad gateway", 502, "INTERNAL_ERROR", true},
		{"service unavailable", 503, "INTERNAL_ERROR", true},
		{"gateway timeout", 504, "INTERNAL_ERROR", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.IntegrationProviderConfig{}
			logger := logging.NewBasic()
			authManager := &mockAuthManager{}
			provider := NewProvider(config, logger, authManager)

			err := provider.handleAPIError(tt.statusCode, []byte("error response"))
			require.Error(t, err)

			clientErr, ok := err.(*clients.ClientError)
			require.True(t, ok)

			assert.Equal(t, tt.expectedCode, clientErr.Code)
			assert.Equal(t, tt.expectedRetryable, clientErr.Retryable)
			assert.Equal(t, tt.statusCode, clientErr.StatusCode)
		})
	}
}

func TestProvider_JQLQueryBuilder(t *testing.T) {
	config := &config.IntegrationProviderConfig{
		ProjectKey: "PROJ",
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	provider := NewProvider(config, logger, authManager)

	tests := []struct {
		name     string
		query    map[string]interface{}
		expected string
	}{
		{
			name:     "empty query",
			query:    map[string]interface{}{},
			expected: "project = PROJ",
		},
		{
			name: "status filter",
			query: map[string]interface{}{
				"status": "In Progress",
			},
			expected: "project = PROJ AND status = \"In Progress\"",
		},
		{
			name: "multiple filters",
			query: map[string]interface{}{
				"status":   "To Do",
				"assignee": "john.doe",
				"priority": "High",
			},
			expected: "project = PROJ AND status = \"To Do\" AND assignee = \"john.doe\" AND priority = \"High\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.buildJQLQuery(tt.query)

			// Check that all expected parts are present
			assert.Contains(t, result, "project = PROJ")

			if status, ok := tt.query["status"].(string); ok {
				assert.Contains(t, result, fmt.Sprintf("status = \"%s\"", status))
			}

			if assignee, ok := tt.query["assignee"].(string); ok {
				assert.Contains(t, result, fmt.Sprintf("assignee = \"%s\"", assignee))
			}

			if priority, ok := tt.query["priority"].(string); ok {
				assert.Contains(t, result, fmt.Sprintf("priority = \"%s\"", priority))
			}
		})
	}
}

func TestProvider_InterfaceCompliance(t *testing.T) {
	// Verify that Provider implements IntegrationProvider interface
	config := &config.IntegrationProviderConfig{}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}

	var _ integration.IntegrationProvider = NewProvider(config, logger, authManager)

	provider := NewProvider(config, logger, authManager)

	// Test basic interface methods
	assert.Equal(t, "jira", provider.Name())
	assert.False(t, provider.SupportsRealtime())
	assert.Empty(t, provider.GetWebhookURL())

	fieldMapping := provider.GetFieldMapping()
	assert.NotEmpty(t, fieldMapping)
	assert.Contains(t, fieldMapping, "title")
	assert.Contains(t, fieldMapping, "status")
}
