package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/integration"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/clients"
)

// JiraTime handles Jira's timestamp format which uses +1000 instead of +10:00
type JiraTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Jira timestamps
func (jt *JiraTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	str := strings.Trim(string(data), `"`)
	jt.Time = parseJiraTime(str)
	return nil
}

// parseJiraTime parses Jira timestamp formats
func parseJiraTime(str string) time.Time {
	// Try parsing with different Jira timestamp formats
	formats := []string{
		"2006-01-02T15:04:05.000-0700", // Jira format: 2025-09-21T22:24:41.965+1000
		"2006-01-02T15:04:05-0700",     // Without milliseconds
		time.RFC3339,                   // Standard format
		time.RFC3339Nano,               // With nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t
		}
	}

	// Return zero time if parsing fails
	return time.Time{}
}

// Provider implements the IntegrationProvider interface for Jira
type Provider struct {
	config        *config.IntegrationProviderConfig
	logger        logging.Logger
	auth          auth.Manager
	httpClient    *http.Client
	fieldMappings map[string]string
	baseURL       string
	projectKey    string
}

// JiraIssue represents a Jira issue structure
type JiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary     string   `json:"summary"`
		Description string   `json:"description"`
		Created     JiraTime `json:"created"`
		Updated     JiraTime `json:"updated"`
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
			Name string `json:"name,omitempty"`
		} `json:"priority,omitempty"`
		Assignee struct {
			AccountID string `json:"accountId,omitempty"`
		} `json:"assignee,omitempty"`
	} `json:"fields"`
}

// JiraUpdateIssueRequest represents the request structure for updating a Jira issue
type JiraUpdateIssueRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

// JiraSearchResponse represents the response from Jira search API
type JiraSearchResponse struct {
	Issues     []JiraIssue `json:"issues"`
	Total      int         `json:"total"`
	MaxResults int         `json:"maxResults"`
	StartAt    int         `json:"startAt"`
}

// NewProvider creates a new Jira integration provider
func NewProvider(
	config *config.IntegrationProviderConfig,
	logger logging.Logger,
	authManager auth.Manager,
) *Provider {
	return &Provider{
		config:     config,
		logger:     logger,
		auth:       authManager,
		baseURL:    config.URL,
		projectKey: config.ProjectKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 5,
			},
		},
		fieldMappings: config.FieldMapping,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "jira"
}

// GetTaskData retrieves task data from Jira
func (p *Provider) GetTaskData(ctx context.Context, externalID string) (*integration.ExternalTaskData, error) {
	p.logger.Debug("getting Jira issue", "issue_key", externalID)

	// Build API URL
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", p.baseURL, externalID)
	p.logger.Debug("making API request", "url", url)

	// Make API request
	resp, err := p.makeAPIRequest(ctx, "GET", url, nil)
	if err != nil {
		p.logger.Debug("API request failed", "error", err)
		return nil, fmt.Errorf("failed to get Jira issue: %w", err)
	}
	p.logger.Debug("API request successful", "response_size", len(resp))

	// Parse response using flexible approach to handle Jira's time format
	p.logger.Debug("attempting to parse Jira response", "response_length", len(resp))
	var rawIssue map[string]interface{}
	if err := json.Unmarshal(resp, &rawIssue); err != nil {
		snippet := string(resp)
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		p.logger.Debug("json unmarshal failed", "error", err, "response_snippet", snippet)
		return nil, fmt.Errorf("failed to parse Jira issue response: %w", err)
	}
	p.logger.Debug("successfully parsed raw issue")

	// Extract fields manually with flexible time parsing
	var issue JiraIssue
	if key, ok := rawIssue["key"].(string); ok {
		issue.Key = key
	}
	if id, ok := rawIssue["id"].(string); ok {
		issue.ID = id
	}
	if self, ok := rawIssue["self"].(string); ok {
		issue.Self = self
	}

	if fields, ok := rawIssue["fields"].(map[string]interface{}); ok {
		if summary, ok := fields["summary"].(string); ok {
			issue.Fields.Summary = summary
		}
		if desc, ok := fields["description"].(string); ok {
			issue.Fields.Description = desc
		}

		// Parse timestamps flexibly
		if created, ok := fields["created"].(string); ok {
			issue.Fields.Created = JiraTime{Time: parseJiraTime(created)}
		}
		if updated, ok := fields["updated"].(string); ok {
			issue.Fields.Updated = JiraTime{Time: parseJiraTime(updated)}
		}

		// Parse nested objects safely
		if status, ok := fields["status"].(map[string]interface{}); ok {
			if name, ok := status["name"].(string); ok {
				issue.Fields.Status.Name = name
			}
		}
		if priority, ok := fields["priority"].(map[string]interface{}); ok {
			if name, ok := priority["name"].(string); ok {
				issue.Fields.Priority.Name = name
			}
		}
		if assignee, ok := fields["assignee"].(map[string]interface{}); ok {
			if displayName, ok := assignee["displayName"].(string); ok {
				issue.Fields.Assignee.DisplayName = displayName
			}
		}
		if issueType, ok := fields["issuetype"].(map[string]interface{}); ok {
			if name, ok := issueType["name"].(string); ok {
				issue.Fields.IssueType.Name = name
			}
		}
		if project, ok := fields["project"].(map[string]interface{}); ok {
			if key, ok := project["key"].(string); ok {
				issue.Fields.Project.Key = key
			}
		}
	}

	// Convert to external task data
	taskData := &integration.ExternalTaskData{
		ID:          issue.Key,
		Title:       issue.Fields.Summary,
		Description: issue.Fields.Description,
		Status:      issue.Fields.Status.Name,
		Priority:    issue.Fields.Priority.Name,
		Assignee:    issue.Fields.Assignee.DisplayName,
		Created:     issue.Fields.Created.Time,
		Updated:     issue.Fields.Updated.Time,
		Fields: map[string]interface{}{
			"issue_type": issue.Fields.IssueType.Name,
			"project":    issue.Fields.Project.Key,
			"self":       issue.Self,
		},
	}

	p.logger.Debug("retrieved Jira issue", "key", issue.Key, "summary", issue.Fields.Summary)

	return taskData, nil
}

// CreateTask creates a new task in Jira
func (p *Provider) CreateTask(ctx context.Context, taskData *integration.ZenTaskData) (*integration.ExternalTaskData, error) {
	p.logger.Debug("creating Jira issue", "title", taskData.Title)

	// Build create request
	createReq := JiraCreateIssueRequest{}
	createReq.Fields.Project.Key = p.projectKey
	createReq.Fields.Summary = taskData.Title
	createReq.Fields.Description = taskData.Description
	createReq.Fields.IssueType.Name = "Task" // Default issue type

	if taskData.Priority != "" {
		createReq.Fields.Priority.Name = taskData.Priority
	}

	// Convert to JSON
	reqBody, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create request: %w", err)
	}

	// Make API request
	url := fmt.Sprintf("%s/rest/api/3/issue", p.baseURL)
	resp, err := p.makeAPIRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira issue: %w", err)
	}

	// Parse response
	var createResp struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Self string `json:"self"`
	}
	if err := json.Unmarshal(resp, &createResp); err != nil {
		return nil, fmt.Errorf("failed to parse create response: %w", err)
	}

	// Return created task data
	externalData := &integration.ExternalTaskData{
		ID:          createResp.Key,
		Title:       taskData.Title,
		Description: taskData.Description,
		Status:      "To Do", // Default status
		Priority:    taskData.Priority,
		Created:     time.Now(),
		Updated:     time.Now(),
		Fields: map[string]interface{}{
			"jira_id": createResp.ID,
			"self":    createResp.Self,
			"project": p.projectKey,
		},
	}

	p.logger.Info("created Jira issue", "key", createResp.Key, "title", taskData.Title)

	return externalData, nil
}

// UpdateTask updates an existing task in Jira
func (p *Provider) UpdateTask(ctx context.Context, externalID string, taskData *integration.ZenTaskData) (*integration.ExternalTaskData, error) {
	p.logger.Debug("updating Jira issue", "issue_key", externalID, "title", taskData.Title)

	// Build update request
	updateReq := JiraUpdateIssueRequest{
		Fields: make(map[string]interface{}),
	}

	// Map Zen fields to Jira fields
	if taskData.Title != "" {
		updateReq.Fields["summary"] = taskData.Title
	}
	if taskData.Description != "" {
		updateReq.Fields["description"] = taskData.Description
	}
	if taskData.Priority != "" {
		updateReq.Fields["priority"] = map[string]string{"name": taskData.Priority}
	}

	// Convert to JSON
	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update request: %w", err)
	}

	// Make API request
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", p.baseURL, externalID)
	_, err = p.makeAPIRequest(ctx, "PUT", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update Jira issue: %w", err)
	}

	// Get updated issue data
	updatedData, err := p.GetTaskData(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated issue data: %w", err)
	}

	p.logger.Info("updated Jira issue", "key", externalID, "title", taskData.Title)

	return updatedData, nil
}

// SearchTasks searches for tasks in Jira
func (p *Provider) SearchTasks(ctx context.Context, query map[string]interface{}) ([]*integration.ExternalTaskData, error) {
	p.logger.Debug("searching Jira issues", "query", query)

	// Build JQL query
	jql := p.buildJQLQuery(query)

	// Build search URL with proper encoding
	searchURL := fmt.Sprintf("%s/rest/api/3/search?jql=%s&maxResults=50", p.baseURL, url.QueryEscape(jql))

	// Make API request
	resp, err := p.makeAPIRequest(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search Jira issues: %w", err)
	}

	// Parse response
	var searchResp JiraSearchResponse
	if err := json.Unmarshal(resp, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert to external task data
	tasks := make([]*integration.ExternalTaskData, 0, len(searchResp.Issues))
	for _, issue := range searchResp.Issues {
		taskData := &integration.ExternalTaskData{
			ID:          issue.Key,
			Title:       issue.Fields.Summary,
			Description: issue.Fields.Description,
			Status:      issue.Fields.Status.Name,
			Priority:    issue.Fields.Priority.Name,
			Assignee:    issue.Fields.Assignee.DisplayName,
			Created:     issue.Fields.Created.Time,
			Updated:     issue.Fields.Updated.Time,
			Fields: map[string]interface{}{
				"issue_type": issue.Fields.IssueType.Name,
				"project":    issue.Fields.Project.Key,
				"self":       issue.Self,
			},
		}
		tasks = append(tasks, taskData)
	}

	p.logger.Debug("found Jira issues", "count", len(tasks), "total", searchResp.Total)

	return tasks, nil
}

// ValidateConnection tests the connection to Jira
func (p *Provider) ValidateConnection(ctx context.Context) error {
	p.logger.Debug("validating Jira connection")

	// Test connection by getting server info
	url := fmt.Sprintf("%s/rest/api/3/serverInfo", p.baseURL)

	_, err := p.makeAPIRequest(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("jira connection validation failed: %w", err)
	}

	p.logger.Debug("Jira connection validated successfully")

	return nil
}

// HealthCheck performs a health check on the Jira provider
func (p *Provider) HealthCheck(ctx context.Context) (*integration.ProviderHealth, error) {
	start := time.Now()

	health := &integration.ProviderHealth{
		Provider:    p.Name(),
		LastChecked: time.Now(),
	}

	// Test basic connectivity
	err := p.ValidateConnection(ctx)
	responseTime := time.Since(start)

	health.ResponseTime = responseTime

	if err != nil {
		health.Healthy = false
		health.LastError = err.Error()
		health.ErrorCount++
	} else {
		health.Healthy = true
		health.ErrorCount = 0
	}

	// Get rate limit info if available
	if rateLimitInfo, err := p.GetRateLimitInfo(ctx); err == nil {
		health.RateLimitInfo = rateLimitInfo
	}

	return health, nil
}

// GetRateLimitInfo returns current rate limit information
func (p *Provider) GetRateLimitInfo(ctx context.Context) (*integration.RateLimitInfo, error) {
	// Jira doesn't provide rate limit info in headers by default
	// This would need to be implemented based on Jira's specific rate limiting
	return &integration.RateLimitInfo{
		Limit:     1000,
		Remaining: 950, // Placeholder
		ResetTime: time.Now().Add(1 * time.Hour),
	}, nil
}

// SupportsRealtime returns true if the provider supports real-time updates
func (p *Provider) SupportsRealtime() bool {
	return false // Jira webhooks would be implemented separately
}

// GetWebhookURL returns the webhook URL for real-time updates
func (p *Provider) GetWebhookURL() string {
	return "" // Not implemented yet
}

// GetFieldMapping returns the field mapping configuration
func (p *Provider) GetFieldMapping() map[string]string {
	if p.fieldMappings != nil {
		return p.fieldMappings
	}

	// Return default field mappings
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

// MapToZen converts Jira issue data to Zen format
func (p *Provider) MapToZen(external *integration.ExternalTaskData) (*integration.ZenTaskData, error) {
	if external == nil {
		return nil, fmt.Errorf("external task data cannot be nil")
	}

	zenData := &integration.ZenTaskData{
		ID:          external.ID,
		Title:       external.Title,
		Description: external.Description,
		Status:      p.mapJiraStatusToZen(external.Status),
		Priority:    p.mapJiraPriorityToZen(external.Priority),
		Owner:       external.Assignee,
		Created:     external.Created,
		Updated:     external.Updated,
		Metadata: map[string]interface{}{
			"external_system": "jira",
			"external_id":     external.ID,
			"jira_fields":     external.Fields,
		},
	}

	p.logger.Debug("mapped Jira data to Zen format", "jira_key", external.ID, "zen_id", zenData.ID)

	return zenData, nil
}

// MapToExternal converts Zen task data to Jira format
func (p *Provider) MapToExternal(zen *integration.ZenTaskData) (*integration.ExternalTaskData, error) {
	if zen == nil {
		return nil, fmt.Errorf("zen task data cannot be nil")
	}

	externalData := &integration.ExternalTaskData{
		ID:          zen.ID,
		Title:       zen.Title,
		Description: zen.Description,
		Status:      p.mapZenStatusToJira(zen.Status),
		Priority:    p.mapZenPriorityToJira(zen.Priority),
		Assignee:    zen.Owner,
		Created:     zen.Created,
		Updated:     zen.Updated,
		Fields: map[string]interface{}{
			"project_key": p.projectKey,
			"issue_type":  "Task",
		},
	}

	// Copy metadata
	if zen.Metadata != nil {
		if jiraFields, ok := zen.Metadata["jira_fields"].(map[string]interface{}); ok {
			for k, v := range jiraFields {
				externalData.Fields[k] = v
			}
		}
	}

	p.logger.Debug("mapped Zen data to Jira format", "zen_id", zen.ID, "jira_key", externalData.ID)

	return externalData, nil
}

// makeAPIRequest makes an authenticated API request to Jira
func (p *Provider) makeAPIRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	// Create request
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Zen-CLI/1.0")

	// Add authentication
	if err := p.addAuthentication(req); err != nil {
		return nil, fmt.Errorf("failed to add authentication: %w", err)
	}

	// Make request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, &clients.ClientError{
			Code:      clients.ErrorCodeConnectionFailed,
			Message:   fmt.Sprintf("HTTP request failed: %v", err),
			Retryable: true,
		}
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return nil, p.handleAPIError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// addAuthentication adds authentication to the request
func (p *Provider) addAuthentication(req *http.Request) error {
	// Debug logging
	p.logger.Debug("authentication config",
		"email", p.config.Email,
		"has_api_key", p.config.APIKey != "",
		"credentials_ref", p.config.Credentials,
		"auth_type", p.config.Type)

	// Use credentials directly from config if available
	if p.config.Email != "" && p.config.APIKey != "" {
		p.logger.Debug("using direct config credentials")
		return p.setBasicAuth(req, p.config.Email, p.config.APIKey)
	}

	// Fall back to auth manager if credentials reference is provided
	if p.config.Credentials == "" {
		return fmt.Errorf("no credentials configured - email: '%s', api_key: '%s', credentials: '%s'",
			p.config.Email,
			func() string {
				if p.config.APIKey != "" {
					return "[REDACTED]"
				} else {
					return ""
				}
			}(),
			p.config.Credentials)
	}

	credentials, err := p.auth.GetCredentials(p.config.Credentials)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	switch p.config.Type {
	case "basic":
		// Assume credentials are in "username:password" format
		req.Header.Set("Authorization", "Basic "+credentials)
	case "token":
		// For Jira, use basic auth with email:token
		if p.config.Email != "" {
			return p.setBasicAuth(req, p.config.Email, credentials)
		}
		// Fallback to bearer token
		req.Header.Set("Authorization", "Bearer "+credentials)
	case "oauth2":
		// Assume credentials are an OAuth2 access token
		req.Header.Set("Authorization", "Bearer "+credentials)
	default:
		return fmt.Errorf("unsupported auth type: %s", p.config.Type)
	}

	return nil
}

// setBasicAuth sets basic authentication header for Jira
func (p *Provider) setBasicAuth(req *http.Request, email, token string) error {
	auth := email + ":" + token
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)
	return nil
}

// handleAPIError handles Jira API errors
func (p *Provider) handleAPIError(statusCode int, body []byte) error {
	var errorCode string
	var retryable bool

	switch statusCode {
	case 400:
		errorCode = clients.ErrorCodeInvalidRequest
		retryable = false
	case 401:
		errorCode = clients.ErrorCodeAuthenticationFailed
		retryable = false
	case 403:
		errorCode = clients.ErrorCodeAuthenticationFailed
		retryable = false
	case 404:
		errorCode = clients.ErrorCodeNotFound
		retryable = false
	case 429:
		errorCode = clients.ErrorCodeRateLimited
		retryable = true
	case 500, 502, 503, 504:
		errorCode = clients.ErrorCodeInternalError
		retryable = true
	default:
		errorCode = clients.ErrorCodeUnknown
		retryable = false
	}

	return &clients.ClientError{
		Code:       errorCode,
		Message:    fmt.Sprintf("Jira API error: %d", statusCode),
		StatusCode: statusCode,
		Details: map[string]interface{}{
			"response_body": string(body),
		},
		Retryable: retryable,
	}
}

// buildJQLQuery builds a JQL query from search parameters
func (p *Provider) buildJQLQuery(query map[string]interface{}) string {
	var conditions []string

	// Add project filter
	conditions = append(conditions, fmt.Sprintf("project = %s", p.projectKey))

	// Add other conditions based on query parameters
	if status, ok := query["status"].(string); ok && status != "" {
		conditions = append(conditions, fmt.Sprintf("status = \"%s\"", status))
	}

	if assignee, ok := query["assignee"].(string); ok && assignee != "" {
		conditions = append(conditions, fmt.Sprintf("assignee = \"%s\"", assignee))
	}

	if priority, ok := query["priority"].(string); ok && priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = \"%s\"", priority))
	}

	return strings.Join(conditions, " AND ")
}

// Status mapping functions

func (p *Provider) mapJiraStatusToZen(jiraStatus string) string {
	statusMap := map[string]string{
		"To Do":       "todo",
		"In Progress": "in_progress",
		"Done":        "completed",
		"Blocked":     "blocked",
		"Canceled":    "canceled",
	}

	if zenStatus, ok := statusMap[jiraStatus]; ok {
		return zenStatus
	}

	return strings.ToLower(strings.ReplaceAll(jiraStatus, " ", "_"))
}

func (p *Provider) mapZenStatusToJira(zenStatus string) string {
	statusMap := map[string]string{
		"todo":        "To Do",
		"in_progress": "In Progress",
		"completed":   "Done",
		"blocked":     "Blocked",
		"canceled":    "Canceled",
	}

	if jiraStatus, ok := statusMap[zenStatus]; ok {
		return jiraStatus
	}

	return zenStatus
}

func (p *Provider) mapJiraPriorityToZen(jiraPriority string) string {
	priorityMap := map[string]string{
		"Highest": "critical",
		"High":    "high",
		"Medium":  "medium",
		"Low":     "low",
		"Lowest":  "low",
	}

	if zenPriority, ok := priorityMap[jiraPriority]; ok {
		return zenPriority
	}

	return strings.ToLower(jiraPriority)
}

func (p *Provider) mapZenPriorityToJira(zenPriority string) string {
	priorityMap := map[string]string{
		"critical": "Highest",
		"high":     "High",
		"medium":   "Medium",
		"low":      "Low",
	}

	if jiraPriority, ok := priorityMap[zenPriority]; ok {
		return jiraPriority
	}

	return zenPriority
}
