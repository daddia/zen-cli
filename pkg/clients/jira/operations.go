package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Core Operations Implementation

// FetchTask fetches a task from Jira
func (p *Plugin) FetchTask(ctx context.Context, externalID string, opts *FetchOptions) (*PluginTaskData, error) {
	if opts == nil {
		opts = &FetchOptions{}
	}

	p.logger.Info("fetching task from Jira plugin", "external_id", externalID, "base_url", p.config.BaseURL)

	// Build request URL
	endpoint := p.buildJiraURL(fmt.Sprintf("rest/api/3/issue/%s", externalID))
	p.logger.Info("making API request", "url", endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if err := p.addAuthentication(req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, p.handleHTTPError(resp)
	}

	// Parse response
	var jiraIssue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&jiraIssue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to standard task data
	taskData := p.convertJiraIssueToTaskData(&jiraIssue)

	if opts.IncludeRaw {
		// Store raw Jira data
		rawData, _ := json.Marshal(jiraIssue)
		var rawMap map[string]interface{}
		json.Unmarshal(rawData, &rawMap)
		taskData.RawData = rawMap
	}

	p.logger.Debug("successfully fetched task", "task_id", taskData.ID, "title", taskData.Title)

	return taskData, nil
}

// CreateTask creates a new task in Jira
func (p *Plugin) CreateTask(ctx context.Context, taskData *PluginTaskData, opts *CreateOptions) (*PluginTaskData, error) {
	p.logger.Debug("creating task in Jira", "title", taskData.Title)

	// Convert to Jira format
	createRequest := p.convertTaskDataToJiraCreate(taskData)

	// Marshal request
	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build request
	endpoint := p.buildJiraURL("rest/api/3/issue")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	if err := p.addAuthentication(req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, p.handleHTTPError(resp)
	}

	// Parse response
	var createResponse JiraCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Fetch the created issue to get complete data
	createdTask, err := p.FetchTask(ctx, createResponse.Key, &FetchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created task: %w", err)
	}

	p.logger.Info("successfully created task in Jira",
		"task_id", createdTask.ID,
		"external_id", createdTask.ExternalID,
		"title", createdTask.Title)

	return createdTask, nil
}

// UpdateTask updates an existing task in Jira
func (p *Plugin) UpdateTask(ctx context.Context, externalID string, taskData *PluginTaskData, opts *UpdateOptions) (*PluginTaskData, error) {
	p.logger.Debug("updating task in Jira", "external_id", externalID, "title", taskData.Title)

	// Convert to Jira update format
	updateRequest := p.convertTaskDataToJiraUpdate(taskData)

	// Marshal request
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build request
	endpoint := p.buildJiraURL(fmt.Sprintf("rest/api/3/issue/%s", externalID))

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	if err := p.addAuthentication(req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return nil, p.handleHTTPError(resp)
	}

	// Fetch updated issue to get complete data
	updatedTask, err := p.FetchTask(ctx, externalID, &FetchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated task: %w", err)
	}

	p.logger.Info("successfully updated task in Jira",
		"task_id", updatedTask.ID,
		"external_id", updatedTask.ExternalID,
		"title", updatedTask.Title)

	return updatedTask, nil
}

// DeleteTask deletes a task in Jira
func (p *Plugin) DeleteTask(ctx context.Context, externalID string, opts *DeleteOptions) error {
	p.logger.Debug("deleting task in Jira", "external_id", externalID)

	// Build request
	endpoint := p.buildJiraURL(fmt.Sprintf("rest/api/3/issue/%s", externalID))

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if err := p.addAuthentication(req); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return p.handleHTTPError(resp)
	}

	p.logger.Info("successfully deleted task in Jira", "external_id", externalID)

	return nil
}

// SearchTasks searches for tasks in Jira
func (p *Plugin) SearchTasks(ctx context.Context, query *SearchQuery, opts *SearchOptions) ([]*PluginTaskData, error) {
	if opts == nil {
		opts = &SearchOptions{
			MaxResults: 50,
			StartAt:    0,
		}
	}

	p.logger.Debug("searching tasks in Jira", "jql", query.JQL)

	// Build JQL query
	jqlQuery := p.buildJQLQuery(query)

	// Build request URL with parameters
	endpoint := p.buildJiraURL("rest/api/3/search")

	params := url.Values{}
	params.Set("jql", jqlQuery)
	params.Set("maxResults", strconv.Itoa(opts.MaxResults))
	params.Set("startAt", strconv.Itoa(opts.StartAt))

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if err := p.addAuthentication(req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, p.handleHTTPError(resp)
	}

	// Parse response
	var searchResponse JiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to standard task data
	tasks := make([]*PluginTaskData, len(searchResponse.Issues))
	for i, issue := range searchResponse.Issues {
		tasks[i] = p.convertJiraIssueToTaskData(&issue)
	}

	p.logger.Debug("successfully searched tasks", "found", len(tasks))

	return tasks, nil
}

// Synchronization Operations

// SyncTask synchronizes a task between Zen and Jira
func (p *Plugin) SyncTask(ctx context.Context, taskID string, opts *SyncOptions) (*SyncResult, error) {
	start := time.Now()

	result := &SyncResult{
		TaskID:        taskID,
		Direction:     opts.Direction,
		Timestamp:     time.Now(),
		ChangedFields: []string{},
	}

	p.logger.Debug("syncing task", "task_id", taskID, "direction", opts.Direction)

	// Implementation would depend on the sync direction and strategy
	// This is a placeholder for the sync logic
	switch opts.Direction {
	case SyncDirectionPull:
		err := p.syncPullFromJira(ctx, taskID, opts, result)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result, err
		}
	case SyncDirectionPush:
		err := p.syncPushToJira(ctx, taskID, opts, result)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result, err
		}
	case SyncDirectionBidirectional:
		err := p.syncBidirectional(ctx, taskID, opts, result)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result, err
		}
	default:
		err := fmt.Errorf("unsupported sync direction: %s", opts.Direction)
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.Duration = time.Since(start)

	return result, nil
}

// GetSyncMetadata retrieves sync metadata for a task
func (p *Plugin) GetSyncMetadata(ctx context.Context, taskID string) (*SyncMetadata, error) {
	// This would typically load from a metadata store
	// For now, return a placeholder
	return &SyncMetadata{
		TaskID:           taskID,
		LastSyncTime:     time.Now(),
		SyncDirection:    SyncDirectionBidirectional,
		ConflictStrategy: ConflictStrategyTimestamp,
		Metadata:         make(map[string]interface{}),
		Version:          1,
		Status:           "active",
	}, nil
}

// Helper methods for sync operations

func (p *Plugin) syncPullFromJira(ctx context.Context, taskID string, opts *SyncOptions, result *SyncResult) error {
	// Implementation for pulling data from Jira to Zen
	p.logger.Debug("pulling task data from Jira", "task_id", taskID)
	// TODO: Implement pull logic
	return nil
}

func (p *Plugin) syncPushToJira(ctx context.Context, taskID string, opts *SyncOptions, result *SyncResult) error {
	// Implementation for pushing data from Zen to Jira
	p.logger.Debug("pushing task data to Jira", "task_id", taskID)
	// TODO: Implement push logic
	return nil
}

func (p *Plugin) syncBidirectional(ctx context.Context, taskID string, opts *SyncOptions, result *SyncResult) error {
	// Implementation for bidirectional sync with conflict resolution
	p.logger.Debug("bidirectional sync for task", "task_id", taskID)
	// TODO: Implement bidirectional sync logic
	return nil
}

// buildJQLQuery builds a JQL query from search parameters
func (p *Plugin) buildJQLQuery(query *SearchQuery) string {
	if query.JQL != "" {
		return query.JQL
	}

	// Build JQL from filters
	var conditions []string

	// Always include project filter
	conditions = append(conditions, fmt.Sprintf("project = %s", p.config.ProjectKey))

	for key, value := range query.Filters {
		switch key {
		case "status":
			if status, ok := value.(string); ok {
				conditions = append(conditions, fmt.Sprintf("status = \"%s\"", status))
			}
		case "assignee":
			if assignee, ok := value.(string); ok {
				conditions = append(conditions, fmt.Sprintf("assignee = \"%s\"", assignee))
			}
		case "priority":
			if priority, ok := value.(string); ok {
				conditions = append(conditions, fmt.Sprintf("priority = \"%s\"", priority))
			}
		case "type":
			if issueType, ok := value.(string); ok {
				conditions = append(conditions, fmt.Sprintf("issuetype = \"%s\"", issueType))
			}
		}
	}

	return strings.Join(conditions, " AND ")
}
