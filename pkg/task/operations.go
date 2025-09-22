package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/integration/factory"
	"github.com/daddia/zen/pkg/integration/orchestrator"
	"github.com/daddia/zen/pkg/integration/plugin"
	"github.com/daddia/zen/pkg/iostreams"
)

// Operations provides task-related operations that can be used across commands
type Operations struct {
	factory       *cmdutil.Factory
	io            *iostreams.IOStreams
	clientFactory factory.ClientFactoryInterface
	orchestrator  orchestrator.OperationOrchestratorInterface
}

// TaskData represents generic task data from any external source
type TaskData struct {
	ID          string                 `json:"id"`
	ExternalID  string                 `json:"external_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority"`
	Type        string                 `json:"type"`
	Assignee    string                 `json:"assignee"`
	Owner       string                 `json:"owner"`
	Team        string                 `json:"team"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	Labels      []string               `json:"labels"`
	Components  []string               `json:"components"`
	ExternalURL string                 `json:"external_url"`
	Source      string                 `json:"source"` // The source system (jira, github, linear, etc.)
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MetadataInfo contains information about external source metadata files
type MetadataInfo struct {
	FilePath    string    `json:"file_path"`
	ExternalID  string    `json:"external_id"`
	Source      string    `json:"source"`
	LastSync    time.Time `json:"last_sync"`
	SyncEnabled bool      `json:"sync_enabled"`
}

// NewOperations creates a new task operations instance
func NewOperations(factory *cmdutil.Factory) *Operations {
	ops := &Operations{
		factory: factory,
		io:      factory.IOStreams,
	}

	// Initialize integration components if available
	if err := ops.initializeIntegrationComponents(); err != nil {
		factory.Logger.Debug("failed to initialize integration components", "error", err)
		// Continue without integration - it's optional
	}

	return ops
}

// initializeIntegrationComponents initializes the integration framework components
func (ops *Operations) initializeIntegrationComponents() error {
	// Get configuration
	config, err := ops.factory.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check if integrations are configured
	if config.Integrations.TaskSystem == "" || config.Integrations.TaskSystem == "none" {
		return fmt.Errorf("no task system configured")
	}

	// Get auth manager
	authMgr, err := ops.factory.AuthManager()
	if err != nil {
		return fmt.Errorf("failed to get auth manager: %w", err)
	}

	// Create client factory
	ops.clientFactory = factory.NewClientFactory(
		ops.factory.Logger,
		config,
		authMgr,
		nil, // Plugin registry would be injected here
	)

	// Create operation orchestrator
	ops.orchestrator = orchestrator.NewOperationOrchestrator(ops.factory.Logger)

	ops.factory.Logger.Debug("integration components initialized successfully")

	return nil
}

// FetchFromSource fetches task details from any configured external source system
func (ops *Operations) FetchFromSource(ctx context.Context, taskID string, source string) (*TaskData, error) {
	// Removed verbose fetching message to follow Zen design guidelines

	// Check if integration components are available
	if ops.clientFactory == nil {
		return ops.fetchFromSourceFallback(ctx, taskID, source)
	}

	// Get or create plugin instance
	pluginInstance, err := ops.clientFactory.CreatePlugin(ctx, source)
	if err != nil {
		ops.factory.Logger.Warn("failed to create plugin, falling back to legacy method", "source", source, "error", err)
		return ops.fetchFromSourceFallback(ctx, taskID, source)
	}

	// Create operation for orchestrator
	operation := &orchestrator.Operation{
		ID:         fmt.Sprintf("fetch-%s-%d", taskID, time.Now().UnixNano()),
		PluginName: source,
		Type:       orchestrator.OperationTypeFetch,
		ExternalID: taskID,
		Options: map[string]interface{}{
			"include_raw": true,
			"timeout":     30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	// Execute operation through orchestrator
	result, err := ops.orchestrator.ExecuteOperation(ctx, operation)
	if err != nil {
		ops.factory.Logger.Warn("orchestrated fetch failed, falling back to direct call", "error", err)
		return ops.fetchFromSourceDirect(ctx, taskID, source, pluginInstance)
	}

	// Convert result to TaskData
	if !result.Success {
		return nil, fmt.Errorf("fetch operation failed: %s", result.Error)
	}

	// Extract task data from result
	if pluginTaskData, ok := result.Data.(*plugin.TaskData); ok {
		return ops.convertPluginTaskDataToTaskData(pluginTaskData), nil
	}

	// Fallback if data format is unexpected
	return ops.fetchFromSourceDirect(ctx, taskID, source, pluginInstance)
}

// fetchFromSourceDirect fetches directly from plugin (bypass orchestrator)
func (ops *Operations) fetchFromSourceDirect(ctx context.Context, taskID string, source string, pluginInstance plugin.IntegrationPluginInterface) (*TaskData, error) {
	// Fetch task data using plugin
	pluginTaskData, err := pluginInstance.FetchTask(ctx, taskID, &plugin.FetchOptions{
		IncludeRaw: true,
		Timeout:    30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task from %s: %w", source, err)
	}

	// Convert to generic task data
	taskData := ops.convertPluginTaskDataToTaskData(pluginTaskData)

	// Display fetched information
	ops.displayFetchedTaskInfo(taskData)

	return taskData, nil
}

// fetchFromSourceFallback uses the legacy integration manager as fallback
func (ops *Operations) fetchFromSourceFallback(ctx context.Context, taskID string, source string) (*TaskData, error) {
	// Get integration manager
	integrationMgr, err := ops.factory.IntegrationManager()
	if err != nil {
		return nil, fmt.Errorf("integration manager not available: %w", err)
	}

	// Check if integration is configured
	if !integrationMgr.IsConfigured() {
		return nil, fmt.Errorf("external integrations not configured")
	}

	// Get the configured task system
	taskSystem := integrationMgr.GetTaskSystem()
	if taskSystem != source {
		return nil, fmt.Errorf("requested source '%s' does not match configured task system '%s'", source, taskSystem)
	}

	// Create a simple task data structure with basic information
	taskData := &TaskData{
		ID:          taskID,
		ExternalID:  taskID,
		Title:       fmt.Sprintf("Task from %s: %s", source, taskID),
		Description: fmt.Sprintf("Task imported from %s", source),
		Status:      "proposed",
		Priority:    "P2",
		Type:        "task",
		Source:      source,
		Created:     time.Now(),
		Updated:     time.Now(),
		ExternalURL: ops.buildExternalURL(source, taskID),
		Metadata: map[string]interface{}{
			"external_system": source,
			"imported_at":     time.Now().Format(time.RFC3339),
			"fallback_mode":   true,
		},
	}

	// Display fetched information
	ops.displayFetchedTaskInfo(taskData)

	return taskData, nil
}

// convertPluginTaskDataToTaskData converts plugin task data to generic task data
func (ops *Operations) convertPluginTaskDataToTaskData(pluginData *plugin.TaskData) *TaskData {
	return &TaskData{
		ID:          pluginData.ID,
		ExternalID:  pluginData.ExternalID,
		Title:       pluginData.Title,
		Description: pluginData.Description,
		Status:      pluginData.Status,
		Priority:    pluginData.Priority,
		Type:        pluginData.Type,
		Assignee:    pluginData.Assignee,
		Owner:       pluginData.Owner,
		Team:        pluginData.Team,
		Created:     pluginData.Created,
		Updated:     pluginData.Updated,
		DueDate:     pluginData.DueDate,
		Labels:      pluginData.Labels,
		Components:  pluginData.Components,
		ExternalURL: pluginData.ExternalURL,
		Source:      ops.extractSourceFromMetadata(pluginData.Metadata),
		RawData:     pluginData.RawData,
		Metadata:    pluginData.Metadata,
	}
}

// displayFetchedTaskInfo displays information about the fetched task
func (ops *Operations) displayFetchedTaskInfo(taskData *TaskData) {
	fmt.Fprintf(ops.io.Out, "%s\n",
		ops.io.FormatSuccess(fmt.Sprintf("Task data fetch for %s success", taskData.Source)))
}

// extractSourceFromMetadata extracts the source system from metadata
func (ops *Operations) extractSourceFromMetadata(metadata map[string]interface{}) string {
	if source, ok := metadata["external_system"].(string); ok {
		return source
	}
	return "unknown"
}

// CreateSourceMetadata creates metadata file for any external source system
func (ops *Operations) CreateSourceMetadata(ctx context.Context, taskDir string, taskData *TaskData) error {
	fmt.Fprintf(ops.io.Out, "%s Creating %s metadata...\n",
		ops.io.ColorInfo("ℹ"), taskData.Source)

	// Create metadata directory
	metadataDir := filepath.Join(taskDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Create standardized metadata structure
	metadata := map[string]interface{}{
		"external_system": taskData.Source,
		"external_id":     taskData.ExternalID,
		"fetched_at":      time.Now().Format(time.RFC3339),
		"sync_enabled":    true,
		"last_sync":       time.Now().Format(time.RFC3339),
		"task_data": map[string]interface{}{
			"id":           taskData.ID,
			"title":        taskData.Title,
			"type":         taskData.Type,
			"status":       taskData.Status,
			"priority":     taskData.Priority,
			"assignee":     taskData.Assignee,
			"owner":        taskData.Owner,
			"team":         taskData.Team,
			"description":  taskData.Description,
			"created":      taskData.Created.Format(time.RFC3339),
			"updated":      taskData.Updated.Format(time.RFC3339),
			"external_url": taskData.ExternalURL,
			"labels":       taskData.Labels,
			"components":   taskData.Components,
		},
		"metadata": taskData.Metadata,
	}

	// Include raw data if available
	if taskData.RawData != nil {
		metadata["raw_data"] = taskData.RawData
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s metadata: %w", taskData.Source, err)
	}

	// Write to source-specific metadata file
	metadataFilePath := filepath.Join(metadataDir, fmt.Sprintf("%s.json", taskData.Source))
	if err := os.WriteFile(metadataFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write %s metadata file: %w", taskData.Source, err)
	}

	fmt.Fprintf(ops.io.Out, "%s Created %s metadata: %s\n",
		ops.io.FormatSuccess("✓"), taskData.Source, fmt.Sprintf("metadata/%s.json", taskData.Source))

	return nil
}

// ValidateSourceConnection tests the connection to any configured external source system
func (ops *Operations) ValidateSourceConnection(ctx context.Context, source string) error {
	// Get integration manager
	integrationMgr, err := ops.factory.IntegrationManager()
	if err != nil {
		return fmt.Errorf("integration manager not available: %w", err)
	}

	// Check if integration is configured
	if !integrationMgr.IsConfigured() {
		return fmt.Errorf("external integrations not configured")
	}

	// For now, just validate that the source matches the configured task system
	taskSystem := integrationMgr.GetTaskSystem()
	if taskSystem != source {
		return fmt.Errorf("source '%s' does not match configured task system '%s'", source, taskSystem)
	}

	// Connection validation would be handled by the specific provider
	// This is a placeholder until the full integration service is available
	return nil
}

// GetSourceMetadata reads metadata from a task directory for any source system
func (ops *Operations) GetSourceMetadata(taskDir string, source string) (*MetadataInfo, error) {
	metadataFilePath := filepath.Join(taskDir, "metadata", fmt.Sprintf("%s.json", source))

	// Check if file exists
	if _, err := os.Stat(metadataFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no %s metadata found for task", source)
	}

	// Read and parse metadata file
	data, err := os.ReadFile(metadataFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s metadata: %w", source, err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse %s metadata: %w", source, err)
	}

	// Extract metadata info
	info := &MetadataInfo{
		FilePath:    metadataFilePath,
		Source:      source,
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

// Helper methods

// buildExternalURL builds the external URL for a task
func (ops *Operations) buildExternalURL(source string, taskID string) string {
	// Get configuration to build proper URLs
	config, err := ops.factory.Config()
	if err != nil {
		return ""
	}

	if len(config.Integrations.Providers) == 0 {
		return ""
	}

	providerConfig, exists := config.Integrations.Providers[source]
	if !exists {
		return ""
	}

	// Build URL based on source system
	switch source {
	case "jira":
		return fmt.Sprintf("%s/browse/%s", providerConfig.URL, taskID)
	case "github":
		return fmt.Sprintf("%s/issues/%s", providerConfig.URL, taskID)
	case "linear":
		return fmt.Sprintf("%s/issue/%s", providerConfig.URL, taskID)
	default:
		return providerConfig.URL
	}
}
