package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daddia/zen/internal/integration"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/cmdutil"
)

// Client provides Jira integration functionality for CLI commands
type Client struct {
	integrationManager cmdutil.IntegrationManagerInterface
	logger             logging.Logger
}

// NewClient creates a new Jira client
func NewClient(integrationManager cmdutil.IntegrationManagerInterface, logger logging.Logger) *Client {
	return &Client{
		integrationManager: integrationManager,
		logger:             logger,
	}
}

// TaskData represents Jira task data for CLI operations
type TaskData struct {
	ID          string                        `json:"id"`
	Title       string                        `json:"title"`
	Type        string                        `json:"type"`
	Status      string                        `json:"status"`
	Priority    string                        `json:"priority"`
	Assignee    string                        `json:"assignee"`
	Description string                        `json:"description"`
	Created     time.Time                     `json:"created"`
	Updated     time.Time                     `json:"updated"`
	RawData     *integration.ExternalTaskData `json:"raw_data"`
}

// FetchTaskOptions contains options for fetching tasks from Jira
type FetchTaskOptions struct {
	TaskID     string
	IncludeRaw bool // Include raw Jira data
	Timeout    time.Duration
}

// SyncTaskOptions contains options for syncing tasks with Jira
type SyncTaskOptions struct {
	TaskID    string
	Direction string // pull, push, bidirectional
	DryRun    bool
	Force     bool
	Timeout   time.Duration
}

// MetadataInfo contains information about Jira metadata files
type MetadataInfo struct {
	FilePath    string    `json:"file_path"`
	ExternalID  string    `json:"external_id"`
	LastSync    time.Time `json:"last_sync"`
	SyncEnabled bool      `json:"sync_enabled"`
}

// IsConfigured checks if Jira integration is properly configured
func (c *Client) IsConfigured() bool {
	return c.integrationManager.IsConfigured() &&
		c.integrationManager.GetTaskSystem() == "jira"
}

// FetchTask fetches a task from Jira and returns the data
func (c *Client) FetchTask(ctx context.Context, opts FetchTaskOptions) (*TaskData, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("jira integration not configured")
	}

	// Get the integration service to access provider methods
	if service, ok := c.integrationManager.(*integration.Service); ok {
		provider, err := service.GetProvider("jira")
		if err != nil {
			return nil, fmt.Errorf("failed to get Jira provider: %w", err)
		}

		// Validate connection
		if err := provider.ValidateConnection(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to Jira: %w", err)
		}

		// Fetch external data
		externalData, err := provider.GetTaskData(ctx, opts.TaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task %s from Jira: %w", opts.TaskID, err)
		}

		// Map to Zen format
		zenData, err := provider.MapToZen(externalData)
		if err != nil {
			return nil, fmt.Errorf("failed to map Jira data: %w", err)
		}

		// Create task data
		taskData := &TaskData{
			ID:          zenData.ID,
			Title:       zenData.Title,
			Type:        c.mapJiraIssueTypeToZen(externalData),
			Status:      zenData.Status,
			Priority:    zenData.Priority,
			Assignee:    zenData.Owner,
			Description: zenData.Description,
			Created:     zenData.Created,
			Updated:     zenData.Updated,
		}

		if opts.IncludeRaw {
			taskData.RawData = externalData
		}

		return taskData, nil
	}

	return nil, fmt.Errorf("integration manager does not support provider access")
}

// CreateMetadataFile creates a jira.json metadata file in the task directory
func (c *Client) CreateMetadataFile(taskDir string, taskData *TaskData) error {
	// Create metadata directory
	metadataDir := filepath.Join(taskDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Create Jira metadata structure
	jiraMetadata := map[string]interface{}{
		"external_system": "jira",
		"external_id":     taskData.ID,
		"fetched_at":      time.Now().Format(time.RFC3339),
		"sync_enabled":    true,
		"last_sync":       time.Now().Format(time.RFC3339),
		"task_data": map[string]interface{}{
			"title":       taskData.Title,
			"type":        taskData.Type,
			"status":      taskData.Status,
			"priority":    taskData.Priority,
			"assignee":    taskData.Assignee,
			"description": taskData.Description,
			"created":     taskData.Created.Format(time.RFC3339),
			"updated":     taskData.Updated.Format(time.RFC3339),
		},
	}

	// Include raw data if available
	if taskData.RawData != nil {
		jiraMetadata["raw_data"] = taskData.RawData
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(jiraMetadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Jira metadata: %w", err)
	}

	// Write to jira.json file
	jiraFilePath := filepath.Join(metadataDir, "jira.json")
	if err := os.WriteFile(jiraFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write Jira metadata file: %w", err)
	}

	c.logger.Debug("created Jira metadata file", "path", jiraFilePath)

	return nil
}

// GetMetadataInfo reads Jira metadata from a task directory
func (c *Client) GetMetadataInfo(taskDir string) (*MetadataInfo, error) {
	jiraFilePath := filepath.Join(taskDir, "metadata", "jira.json")

	// Check if file exists
	if _, err := os.Stat(jiraFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no Jira metadata found for task")
	}

	// Read and parse metadata file
	data, err := os.ReadFile(jiraFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Jira metadata: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse Jira metadata: %w", err)
	}

	// Extract metadata info
	info := &MetadataInfo{
		FilePath:    jiraFilePath,
		SyncEnabled: true,
	}

	if externalID, ok := metadata["external_id"].(string); ok {
		info.ExternalID = externalID
	}

	if lastSyncStr, ok := metadata["last_sync"].(string); ok {
		if lastSync, err := time.Parse(time.RFC3339, lastSyncStr); err == nil {
			info.LastSync = lastSync
		}
	}

	if syncEnabled, ok := metadata["sync_enabled"].(bool); ok {
		info.SyncEnabled = syncEnabled
	}

	return info, nil
}

// mapJiraIssueTypeToZen maps Jira issue types to Zen task types
func (c *Client) mapJiraIssueTypeToZen(externalData *integration.ExternalTaskData) string {
	jiraIssueType, ok := externalData.Fields["issue_type"].(string)
	if !ok {
		return "story" // Default
	}

	issueTypeMap := map[string]string{
		"Story":      "story",
		"User Story": "story",
		"Bug":        "bug",
		"Defect":     "bug",
		"Epic":       "epic",
		"Initiative": "epic",
		"Spike":      "spike",
		"Research":   "spike",
		"Task":       "task",
		"Sub-task":   "task",
		"Subtask":    "task",
	}

	if zenType, ok := issueTypeMap[jiraIssueType]; ok {
		return zenType
	}

	// Default to story for unknown types
	return "story"
}

// ValidateConnection tests the Jira connection
func (c *Client) ValidateConnection(ctx context.Context) error {
	if !c.IsConfigured() {
		return fmt.Errorf("jira integration not configured")
	}

	if service, ok := c.integrationManager.(*integration.Service); ok {
		provider, err := service.GetProvider("jira")
		if err != nil {
			return fmt.Errorf("failed to get Jira provider: %w", err)
		}

		return provider.ValidateConnection(ctx)
	}

	return fmt.Errorf("integration manager does not support provider access")
}
