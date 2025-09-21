package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daddia/zen/pkg/clients/jira"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
)

// Operations provides task-related operations that can be used across commands
type Operations struct {
	factory    *cmdutil.Factory
	jiraClient *jira.Client
	io         *iostreams.IOStreams
}

// NewOperations creates a new task operations instance
func NewOperations(factory *cmdutil.Factory) *Operations {
	ops := &Operations{
		factory: factory,
		io:      factory.IOStreams,
	}

	// Initialize Jira client if integration is available
	if factory.IntegrationManager != nil {
		if integrationManager, err := factory.IntegrationManager(); err == nil {
			ops.jiraClient = jira.NewClient(integrationManager, factory.Logger)
		}
	}

	return ops
}

// FetchFromJira fetches task details from Jira
func (ops *Operations) FetchFromJira(ctx context.Context, taskID string) (*jira.TaskData, error) {
	if ops.jiraClient == nil {
		return nil, fmt.Errorf("jira client not available")
	}

	if !ops.jiraClient.IsConfigured() {
		return nil, fmt.Errorf("jira integration not configured")
	}

	fmt.Fprintf(ops.io.Out, "%s Fetching task details from Jira...\n",
		ops.io.ColorInfo("ℹ"))

	// Fetch task data
	fetchOpts := jira.FetchTaskOptions{
		TaskID:     taskID,
		IncludeRaw: true,
		Timeout:    30 * time.Second,
	}

	taskData, err := ops.jiraClient.FetchTask(ctx, fetchOpts)
	if err != nil {
		return nil, err
	}

	// Display fetched information
	fmt.Fprintf(ops.io.Out, "%s Fetched Jira issue: %s\n",
		ops.io.FormatSuccess("✓"), taskData.ID)
	fmt.Fprintf(ops.io.Out, "  %s Title: %s\n",
		ops.io.ColorNeutral("→"), taskData.Title)
	fmt.Fprintf(ops.io.Out, "  %s Type: %s\n",
		ops.io.ColorNeutral("→"), taskData.Type)
	fmt.Fprintf(ops.io.Out, "  %s Status: %s\n",
		ops.io.ColorNeutral("→"), taskData.Status)
	fmt.Fprintf(ops.io.Out, "  %s Priority: %s\n",
		ops.io.ColorNeutral("→"), taskData.Priority)
	if taskData.Assignee != "" {
		fmt.Fprintf(ops.io.Out, "  %s Assignee: %s\n",
			ops.io.ColorNeutral("→"), taskData.Assignee)
	}

	return taskData, nil
}

// CreateJiraMetadata creates the jira.json metadata file
func (ops *Operations) CreateJiraMetadata(taskDir string, taskData *jira.TaskData) error {
	if ops.jiraClient == nil {
		return fmt.Errorf("jira client not available")
	}

	if err := ops.jiraClient.CreateMetadataFile(taskDir, taskData); err != nil {
		return err
	}

	fmt.Fprintf(ops.io.Out, "%s Created Jira metadata: %s\n",
		ops.io.FormatSuccess("✓"), "metadata/jira.json")

	return nil
}

// ValidateJiraConnection tests the Jira connection
func (ops *Operations) ValidateJiraConnection(ctx context.Context) error {
	if ops.jiraClient == nil {
		return fmt.Errorf("jira client not available")
	}

	return ops.jiraClient.ValidateConnection(ctx)
}

// GetJiraMetadata reads Jira metadata from a task directory
func (ops *Operations) GetJiraMetadata(taskDir string) (*jira.MetadataInfo, error) {
	if ops.jiraClient == nil {
		return nil, fmt.Errorf("jira client not available")
	}

	return ops.jiraClient.GetMetadataInfo(taskDir)
}

// CreateJiraMetadataFromTask creates Jira metadata for a task that was fetched from Jira
func (ops *Operations) CreateJiraMetadataFromTask(taskDir, taskID string) error {
	if ops.jiraClient == nil {
		return fmt.Errorf("jira client not available")
	}

	// This would typically use cached data from the fetch operation
	// For now, just create a basic metadata structure
	metadataDir := filepath.Join(taskDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Create basic Jira metadata structure
	jiraMetadata := map[string]interface{}{
		"external_system": "jira",
		"external_id":     taskID,
		"fetched_at":      time.Now().Format(time.RFC3339),
		"sync_enabled":    true,
		"last_sync":       time.Now().Format(time.RFC3339),
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

	fmt.Fprintf(ops.io.Out, "%s Created Jira metadata: %s\n",
		ops.io.FormatSuccess("✓"), "metadata/jira.json")

	return nil
}
