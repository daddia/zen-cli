package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/integration"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
)

// JiraProvider implements the IntegrationProvider interface for Jira
type JiraProvider struct {
	config        *config.IntegrationProviderConfig
	logger        logging.Logger
	auth          auth.Manager
	httpClient    *http.Client
	fieldMappings map[string]string
}

// JiraIssue represents a Jira issue structure
type JiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary     string    `json:"summary"`
		Description string    `json:"description"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Assignee struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"assignee"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

// JiraCreateIssueRequest represents the request structure for creating a Jira issue
type JiraCreateIssueRequest struct {
	Fields struct {
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority,omitempty"`
	} `json:"fields"`
}

// JiraUpdateIssueRequest represents the request structure for updating a Jira issue
type JiraUpdateIssueRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

// NewJiraProvider creates a new Jira integration provider
func NewJiraProvider(
	config *config.IntegrationProviderConfig,
	logger logging.Logger,
	authManager auth.Manager,
) *JiraProvider {
	return &JiraProvider{
		config: config,
		logger: logger,
		auth:   authManager,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		fieldMappings: config.FieldMapping,
	}
}

// Name returns the provider name
func (j *JiraProvider) Name() string {
	return "jira"
}

// GetTaskData retrieves task data from Jira
func (j *JiraProvider) GetTaskData(ctx context.Context, externalID string) (*integration.ExternalTaskData, error) {
	j.logger.Debug("getting Jira task data", "issue_key", externalID)

	url := fmt.Sprintf("%s/rest/api/3/issue/%s", j.config.ServerURL, externalID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := j.setAuthHeaders(req); err != nil {
		return nil, fmt.Errorf("failed to set auth headers: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira API returned status %d", resp.StatusCode)
	}

	var jiraIssue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&jiraIssue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Jira issue to external task data
	taskData := &integration.ExternalTaskData{
		ID:          jiraIssue.Key,
		Title:       jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Status:      jiraIssue.Fields.Status.Name,
		Priority:    jiraIssue.Fields.Priority.Name,
		Assignee:    jiraIssue.Fields.Assignee.DisplayName,
		Created:     jiraIssue.Fields.Created,
		Updated:     jiraIssue.Fields.Updated,
		Fields: map[string]interface{}{
			"key":        jiraIssue.Key,
			"id":         jiraIssue.ID,
			"self":       jiraIssue.Self,
			"issue_type": jiraIssue.Fields.IssueType.Name,
		},
	}

	j.logger.Debug("retrieved Jira task data", "issue_key", externalID, "title", taskData.Title)

	return taskData, nil
}

// CreateTask creates a new task in Jira
func (j *JiraProvider) CreateTask(ctx context.Context, taskData *integration.ZenTaskData) (*integration.ExternalTaskData, error) {
	j.logger.Debug("creating Jira task", "title", taskData.Title)

	url := fmt.Sprintf("%s/rest/api/3/issue", j.config.ServerURL)

	// Create the request payload
	createReq := JiraCreateIssueRequest{}
	createReq.Fields.Project.Key = j.config.ProjectKey
	createReq.Fields.Summary = taskData.Title
	createReq.Fields.Description = taskData.Description
	createReq.Fields.IssueType.Name = "Task" // Default to Task type

	// Set priority if provided
	if taskData.Priority != "" {
		createReq.Fields.Priority.Name = taskData.Priority
	}

	payload, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if err := j.setAuthHeaders(req); err != nil {
		return nil, fmt.Errorf("failed to set auth headers: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("jira API returned status %d", resp.StatusCode)
	}

	var createdIssue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&createdIssue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Get the full issue data
	return j.GetTaskData(ctx, createdIssue.Key)
}

// UpdateTask updates an existing task in Jira
func (j *JiraProvider) UpdateTask(ctx context.Context, externalID string, taskData *integration.ZenTaskData) (*integration.ExternalTaskData, error) {
	j.logger.Debug("updating Jira task", "issue_key", externalID, "title", taskData.Title)

	url := fmt.Sprintf("%s/rest/api/3/issue/%s", j.config.ServerURL, externalID)

	// Create the update payload
	updateReq := JiraUpdateIssueRequest{
		Fields: make(map[string]interface{}),
	}

	if taskData.Title != "" {
		updateReq.Fields["summary"] = taskData.Title
	}
	if taskData.Description != "" {
		updateReq.Fields["description"] = taskData.Description
	}
	if taskData.Priority != "" {
		updateReq.Fields["priority"] = map[string]string{"name": taskData.Priority}
	}

	payload, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if err := j.setAuthHeaders(req); err != nil {
		return nil, fmt.Errorf("failed to set auth headers: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("jira API returned status %d", resp.StatusCode)
	}

	// Get the updated issue data
	return j.GetTaskData(ctx, externalID)
}

// SearchTasks searches for tasks in Jira
func (j *JiraProvider) SearchTasks(ctx context.Context, query map[string]interface{}) ([]*integration.ExternalTaskData, error) {
	j.logger.Debug("searching Jira tasks", "query", query)

	// Build JQL query
	jql := fmt.Sprintf("project = %s", j.config.ProjectKey)
	if title, ok := query["title"].(string); ok && title != "" {
		jql += fmt.Sprintf(" AND summary ~ \"%s\"", title)
	}

	url := fmt.Sprintf("%s/rest/api/3/search?jql=%s", j.config.ServerURL, jql)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := j.setAuthHeaders(req); err != nil {
		return nil, fmt.Errorf("failed to set auth headers: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira API returned status %d", resp.StatusCode)
	}

	var searchResult struct {
		Issues []JiraIssue `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert issues to external task data
	tasks := make([]*integration.ExternalTaskData, 0, len(searchResult.Issues))
	for _, issue := range searchResult.Issues {
		taskData := &integration.ExternalTaskData{
			ID:          issue.Key,
			Title:       issue.Fields.Summary,
			Description: issue.Fields.Description,
			Status:      issue.Fields.Status.Name,
			Priority:    issue.Fields.Priority.Name,
			Assignee:    issue.Fields.Assignee.DisplayName,
			Created:     issue.Fields.Created,
			Updated:     issue.Fields.Updated,
		}
		tasks = append(tasks, taskData)
	}

	j.logger.Debug("found Jira tasks", "count", len(tasks))

	return tasks, nil
}

// ValidateConnection tests the connection to Jira
func (j *JiraProvider) ValidateConnection(ctx context.Context) error {
	j.logger.Debug("validating Jira connection")

	url := fmt.Sprintf("%s/rest/api/3/myself", j.config.ServerURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := j.setAuthHeaders(req); err != nil {
		return fmt.Errorf("failed to set auth headers: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jira connection validation failed with status %d", resp.StatusCode)
	}

	j.logger.Debug("Jira connection validated successfully")

	return nil
}

// GetFieldMapping returns the field mapping configuration
func (j *JiraProvider) GetFieldMapping() map[string]string {
	if j.fieldMappings != nil {
		return j.fieldMappings
	}

	// Return default mappings if none configured
	return map[string]string{
		"task_id":     "key",
		"title":       "summary",
		"description": "description",
		"status":      "status.name",
		"priority":    "priority.name",
		"assignee":    "assignee.displayName",
		"created":     "created",
		"updated":     "updated",
	}
}

// MapToZen converts external task data to Zen format
func (j *JiraProvider) MapToZen(external *integration.ExternalTaskData) (*integration.ZenTaskData, error) {
	if external == nil {
		return nil, fmt.Errorf("external data cannot be nil")
	}

	zenData := &integration.ZenTaskData{
		ID:          external.ID,
		Title:       external.Title,
		Description: external.Description,
		Status:      j.mapStatus(external.Status),
		Priority:    j.mapPriority(external.Priority),
		Owner:       external.Assignee,
		Created:     external.Created,
		Updated:     external.Updated,
		Metadata: map[string]interface{}{
			"external_system": "jira",
			"external_id":     external.ID,
			"jira_fields":     external.Fields,
		},
	}

	return zenData, nil
}

// MapToExternal converts Zen task data to external format
func (j *JiraProvider) MapToExternal(zen *integration.ZenTaskData) (*integration.ExternalTaskData, error) {
	if zen == nil {
		return nil, fmt.Errorf("zen data cannot be nil")
	}

	external := &integration.ExternalTaskData{
		ID:          zen.ID,
		Title:       zen.Title,
		Description: zen.Description,
		Status:      j.mapStatusToJira(zen.Status),
		Priority:    j.mapPriorityToJira(zen.Priority),
		Assignee:    zen.Owner,
		Created:     zen.Created,
		Updated:     zen.Updated,
		Fields:      make(map[string]interface{}),
	}

	// Copy metadata
	if zen.Metadata != nil {
		for k, v := range zen.Metadata {
			external.Fields[k] = v
		}
	}

	return external, nil
}

// setAuthHeaders sets the authentication headers for Jira API requests
func (j *JiraProvider) setAuthHeaders(req *http.Request) error {
	// Get credentials from auth manager
	token, err := j.auth.GetCredentials("jira")
	if err != nil {
		return fmt.Errorf("failed to get Jira credentials: %w", err)
	}

	// Jira Cloud uses Basic Auth with email + API token
	// The token should be in format "email:api_token"
	if strings.Contains(token, ":") {
		parts := strings.SplitN(token, ":", 2)
		req.SetBasicAuth(parts[0], parts[1])
	} else {
		// Fallback: assume token is just the API token and get email from config
		email := j.config.Settings["email"]
		if emailStr, ok := email.(string); ok {
			req.SetBasicAuth(emailStr, token)
		} else {
			return fmt.Errorf("email not configured for Jira authentication")
		}
	}

	return nil
}

// mapStatus maps Jira status to Zen status
func (j *JiraProvider) mapStatus(jiraStatus string) string {
	statusMap := map[string]string{
		"To Do":       "not_started",
		"In Progress": "in_progress",
		"Done":        "completed",
		"Closed":      "completed",
		"Blocked":     "blocked",
	}

	if zenStatus, ok := statusMap[jiraStatus]; ok {
		return zenStatus
	}

	// Default mapping
	return strings.ToLower(strings.ReplaceAll(jiraStatus, " ", "_"))
}

// mapPriority maps Jira priority to Zen priority
func (j *JiraProvider) mapPriority(jiraPriority string) string {
	priorityMap := map[string]string{
		"Highest": "P0",
		"High":    "P1",
		"Medium":  "P2",
		"Low":     "P3",
		"Lowest":  "P3",
	}

	if zenPriority, ok := priorityMap[jiraPriority]; ok {
		return zenPriority
	}

	return "P2" // Default to medium priority
}

// mapStatusToJira maps Zen status to Jira status
func (j *JiraProvider) mapStatusToJira(zenStatus string) string {
	statusMap := map[string]string{
		"not_started": "To Do",
		"in_progress": "In Progress",
		"completed":   "Done",
		"blocked":     "Blocked",
		"canceled":    "Closed",
	}

	if jiraStatus, ok := statusMap[zenStatus]; ok {
		return jiraStatus
	}

	return "To Do" // Default status
}

// mapPriorityToJira maps Zen priority to Jira priority
func (j *JiraProvider) mapPriorityToJira(zenPriority string) string {
	priorityMap := map[string]string{
		"P0": "Highest",
		"P1": "High",
		"P2": "Medium",
		"P3": "Low",
	}

	if jiraPriority, ok := priorityMap[zenPriority]; ok {
		return jiraPriority
	}

	return "Medium" // Default priority
}
