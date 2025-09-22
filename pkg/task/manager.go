package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/integration/factory"
	"github.com/daddia/zen/pkg/integration/orchestrator"
	"github.com/daddia/zen/pkg/integration/plugin"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/templates"
)

// Manager provides comprehensive task management functionality
type Manager struct {
	factory       *cmdutil.Factory
	logger        logging.Logger
	io            *iostreams.IOStreams
	clientFactory factory.ClientFactoryInterface
	orchestrator  orchestrator.OperationOrchestratorInterface
}

// ManagerInterface defines the task manager interface
type ManagerInterface interface {
	// Task CRUD operations
	CreateTask(ctx context.Context, request *CreateTaskRequest) (*Task, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	UpdateTask(ctx context.Context, taskID string, updates *TaskUpdates) (*Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error)

	// External source synchronization
	SyncTask(ctx context.Context, taskID string, opts *SyncOptions) (*SyncResult, error)
	SyncAllTasks(ctx context.Context, opts *SyncOptions) ([]*SyncResult, error)
	PullFromSource(ctx context.Context, taskID string, source string) (*Task, error)
	PushToSource(ctx context.Context, taskID string, source string) (*SyncResult, error)

	// Task lifecycle management
	ProgressTask(ctx context.Context, taskID string, stage string) error
	GetTaskProgress(ctx context.Context, taskID string) (*TaskProgress, error)

	// Metadata management
	GetTaskMetadata(ctx context.Context, taskID string) (*TaskMetadata, error)
	UpdateTaskMetadata(ctx context.Context, taskID string, metadata *TaskMetadata) error

	// Source integration
	GetTaskSources(ctx context.Context, taskID string) ([]string, error)
	AddTaskSource(ctx context.Context, taskID string, source string, externalID string) error
	RemoveTaskSource(ctx context.Context, taskID string, source string) error
}

// Task represents a complete task with all its data and metadata
type Task struct {
	// Core task information
	ID          string `json:"id" yaml:"id"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
	Status      string `json:"status" yaml:"status"`
	Priority    string `json:"priority" yaml:"priority"`

	// Ownership and team
	Owner string `json:"owner" yaml:"owner"`
	Team  string `json:"team" yaml:"team"`

	// Timestamps
	Created time.Time  `json:"created" yaml:"created"`
	Updated time.Time  `json:"updated" yaml:"updated"`
	DueDate *time.Time `json:"due_date,omitempty" yaml:"due_date,omitempty"`

	// Organization
	Labels []string `json:"labels" yaml:"labels"`
	Tags   []string `json:"tags" yaml:"tags"`

	// Workflow
	CurrentStage string `json:"current_stage" yaml:"current_stage"`
	Progress     int    `json:"progress" yaml:"progress"`

	// External sources
	Sources map[string]*TaskSource `json:"sources" yaml:"sources"`

	// File paths
	WorkspacePath string `json:"workspace_path" yaml:"workspace_path"`
	IndexPath     string `json:"index_path" yaml:"index_path"`
	ManifestPath  string `json:"manifest_path" yaml:"manifest_path"`
	MetadataPath  string `json:"metadata_path" yaml:"metadata_path"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// TaskSource represents an external source for a task
type TaskSource struct {
	System        string                 `json:"system" yaml:"system"`
	ExternalID    string                 `json:"external_id" yaml:"external_id"`
	ExternalURL   string                 `json:"external_url" yaml:"external_url"`
	LastSync      time.Time              `json:"last_sync" yaml:"last_sync"`
	SyncEnabled   bool                   `json:"sync_enabled" yaml:"sync_enabled"`
	SyncDirection string                 `json:"sync_direction" yaml:"sync_direction"`
	Metadata      map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// CreateTaskRequest contains parameters for creating a new task
type CreateTaskRequest struct {
	ID       string `json:"id" validate:"required"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Owner    string `json:"owner"`
	Team     string `json:"team"`
	Priority string `json:"priority"`

	// External source integration
	FromSource string `json:"from_source,omitempty"`
	ExternalID string `json:"external_id,omitempty"`

	// Template options
	TemplateVars map[string]interface{} `json:"template_vars,omitempty"`

	// Creation options
	DryRun bool `json:"dry_run"`
}

// TaskUpdates contains fields that can be updated
type TaskUpdates struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	Owner       *string    `json:"owner,omitempty"`
	Team        *string    `json:"team,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Labels      []string   `json:"labels,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

// TaskFilter contains filtering options for listing tasks
type TaskFilter struct {
	Type    string   `json:"type,omitempty"`
	Status  string   `json:"status,omitempty"`
	Owner   string   `json:"owner,omitempty"`
	Team    string   `json:"team,omitempty"`
	Labels  []string `json:"labels,omitempty"`
	Sources []string `json:"sources,omitempty"`
	Stage   string   `json:"stage,omitempty"`
}

// SyncOptions contains options for synchronization operations
type SyncOptions struct {
	Direction        SyncDirection    `json:"direction"`
	ConflictStrategy ConflictStrategy `json:"conflict_strategy"`
	DryRun           bool             `json:"dry_run"`
	Force            bool             `json:"force"`
	Sources          []string         `json:"sources,omitempty"` // Specific sources to sync
}

// SyncDirection represents sync direction
type SyncDirection string

const (
	SyncDirectionPull          SyncDirection = "pull"
	SyncDirectionPush          SyncDirection = "push"
	SyncDirectionBidirectional SyncDirection = "bidirectional"
)

// ConflictStrategy represents conflict resolution strategy
type ConflictStrategy string

const (
	ConflictStrategyLocalWins    ConflictStrategy = "local_wins"
	ConflictStrategyRemoteWins   ConflictStrategy = "remote_wins"
	ConflictStrategyManualReview ConflictStrategy = "manual_review"
	ConflictStrategyTimestamp    ConflictStrategy = "timestamp"
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	TaskID        string        `json:"task_id"`
	Source        string        `json:"source"`
	Success       bool          `json:"success"`
	Direction     SyncDirection `json:"direction"`
	ChangedFields []string      `json:"changed_fields"`
	Conflicts     []Conflict    `json:"conflicts,omitempty"`
	Error         string        `json:"error,omitempty"`
	Duration      time.Duration `json:"duration"`
	Timestamp     time.Time     `json:"timestamp"`
}

// Conflict represents a data conflict between local and remote
type Conflict struct {
	Field       string      `json:"field"`
	LocalValue  interface{} `json:"local_value"`
	RemoteValue interface{} `json:"remote_value"`
	LocalTime   time.Time   `json:"local_time"`
	RemoteTime  time.Time   `json:"remote_time"`
	Resolution  string      `json:"resolution,omitempty"`
}

// TaskProgress represents task workflow progress
type TaskProgress struct {
	CurrentStage    string                 `json:"current_stage"`
	StageNumber     int                    `json:"stage_number"`
	TotalStages     int                    `json:"total_stages"`
	Progress        int                    `json:"progress"` // 0-100
	CompletedStages []string               `json:"completed_stages"`
	Artifacts       []string               `json:"artifacts"`
	QualityGates    []QualityGate          `json:"quality_gates"`
	Blockers        []string               `json:"blockers"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// QualityGate represents a quality gate for stage progression
type QualityGate struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Required    bool       `json:"required"`
	Status      string     `json:"status"` // passed, failed, pending
	CheckedAt   *time.Time `json:"checked_at,omitempty"`
	CheckedBy   string     `json:"checked_by,omitempty"`
}

// TaskMetadata represents task metadata and configuration
type TaskMetadata struct {
	Sources        map[string]*TaskSource `json:"sources"`
	Configuration  map[string]interface{} `json:"configuration"`
	WorkflowConfig *WorkflowConfig        `json:"workflow_config"`
	SyncSettings   *SyncSettings          `json:"sync_settings"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	LastModified   time.Time              `json:"last_modified"`
	Version        int64                  `json:"version"`
}

// WorkflowConfig contains workflow-specific configuration
type WorkflowConfig struct {
	Stages              []WorkflowStage `json:"stages"`
	AutoProgress        bool            `json:"auto_progress"`
	QualityGatesEnabled bool            `json:"quality_gates_enabled"`
	NotificationEvents  []string        `json:"notification_events"`
}

// WorkflowStage represents a workflow stage
type WorkflowStage struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Order        int           `json:"order"`
	QualityGates []QualityGate `json:"quality_gates"`
	Artifacts    []string      `json:"artifacts"`
}

// SyncSettings contains synchronization settings
type SyncSettings struct {
	Enabled          bool                  `json:"enabled"`
	Frequency        string                `json:"frequency"` // manual, hourly, daily
	Direction        SyncDirection         `json:"direction"`
	ConflictStrategy ConflictStrategy      `json:"conflict_strategy"`
	Sources          map[string]SourceSync `json:"sources"`
	LastSync         time.Time             `json:"last_sync"`
}

// SourceSync contains source-specific sync settings
type SourceSync struct {
	Enabled       bool              `json:"enabled"`
	Direction     SyncDirection     `json:"direction"`
	FieldMappings map[string]string `json:"field_mappings"`
	LastSync      time.Time         `json:"last_sync"`
	ErrorCount    int               `json:"error_count"`
	LastError     string            `json:"last_error,omitempty"`
}

// NewManager creates a new task manager
func NewManager(factory *cmdutil.Factory) *Manager {
	if factory == nil {
		return nil
	}

	mgr := &Manager{
		factory: factory,
		logger:  factory.Logger,
		io:      factory.IOStreams,
	}

	// Initialize integration components if available
	if err := mgr.initializeIntegrationComponents(); err != nil {
		factory.Logger.Debug("failed to initialize integration components", "error", err)
		// Continue without integration - it's optional
	}

	return mgr
}

// initializeIntegrationComponents initializes the integration framework components
func (m *Manager) initializeIntegrationComponents() error {
	// Get configuration
	config, err := m.factory.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check if integrations are configured
	if config.Integrations.TaskSystem == "" || config.Integrations.TaskSystem == "none" {
		return fmt.Errorf("no task system configured")
	}

	// Get auth manager
	authMgr, err := m.factory.AuthManager()
	if err != nil {
		return fmt.Errorf("failed to get auth manager: %w", err)
	}

	// Create client factory
	m.clientFactory = factory.NewClientFactory(
		m.logger,
		config,
		authMgr,
		nil, // Plugin registry would be injected here
	)

	// Create operation orchestrator
	m.orchestrator = orchestrator.NewOperationOrchestrator(m.logger)

	m.logger.Debug("integration components initialized successfully")

	return nil
}

// CreateTask creates a new task with optional external source data
func (m *Manager) CreateTask(ctx context.Context, request *CreateTaskRequest) (*Task, error) {
	m.logger.Debug("creating task", "id", request.ID, "from_source", request.FromSource)

	// Validate request
	if err := m.validateCreateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid create request: %w", err)
	}

	// Check if task already exists
	if m.taskExists(request.ID) {
		return nil, fmt.Errorf("task already exists: %s", request.ID)
	}

	// Fetch data from external source if specified
	var sourceData *TaskData
	if request.FromSource != "" {
		var err error
		sourceData, err = m.fetchFromSource(ctx, request.ID, request.FromSource)
		if err != nil {
			// Log warning but continue with local task creation
			m.logger.Warn("failed to fetch from external source, creating local task",
				"task_id", request.ID,
				"source", request.FromSource,
				"error", err)
			
			// Print warning to user
			fmt.Fprintf(m.io.ErrOut, "%s Failed to fetch from %s: %v\n",
				m.io.FormatWarning("⚠"), request.FromSource, err)
			fmt.Fprintf(m.io.ErrOut, "%s Creating local task without external data\n",
				m.io.ColorInfo("ℹ"))
			
			// Continue with local task creation
			sourceData = nil
		}

		// Override request fields with source data (External SoR) if fetch was successful
		if sourceData != nil {
			if sourceData.Title != "" {
				request.Title = sourceData.Title
			}
			if sourceData.Type != "" {
				request.Type = sourceData.Type
			}
			if sourceData.Owner != "" {
				request.Owner = sourceData.Owner
			}
			if sourceData.Team != "" {
				request.Team = sourceData.Team
			}
			if sourceData.Priority != "" {
				request.Priority = sourceData.Priority
			}
		}
	}

	// Create task structure
	task := &Task{
		ID:           request.ID,
		Title:        request.Title,
		Type:         request.Type,
		Status:       "proposed",
		Priority:     request.Priority,
		Owner:        request.Owner,
		Team:         request.Team,
		Created:      time.Now(),
		Updated:      time.Now(),
		CurrentStage: "01-align",
		Progress:     0,
		Sources:      make(map[string]*TaskSource),
		Metadata:     make(map[string]interface{}),
	}

	// Set defaults
	if task.Type == "" {
		task.Type = "story"
	}
	if task.Priority == "" {
		task.Priority = "P2"
	}
	if task.Owner == "" {
		if user := os.Getenv("USER"); user != "" {
			task.Owner = user
		} else {
			task.Owner = "unknown"
		}
	}
	if task.Team == "" {
		task.Team = "default"
	}

	// Add external source if data was fetched
	if sourceData != nil {
		task.Sources[request.FromSource] = &TaskSource{
			System:        request.FromSource,
			ExternalID:    sourceData.ExternalID,
			ExternalURL:   sourceData.ExternalURL,
			LastSync:      time.Now(),
			SyncEnabled:   true,
			SyncDirection: string(SyncDirectionBidirectional),
			Metadata:      sourceData.Metadata,
		}
	}

	// Create task directory structure
	if err := m.createTaskStructure(ctx, task, request, sourceData); err != nil {
		return nil, fmt.Errorf("failed to create task structure: %w", err)
	}

	m.logger.Info("task created successfully", "id", task.ID, "type", task.Type, "source", request.FromSource)

	return task, nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	m.logger.Debug("getting task", "id", taskID)

	// Check if task exists
	if !m.taskExists(taskID) {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Load task from manifest
	task, err := m.loadTaskFromManifest(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to load task: %w", err)
	}

	return task, nil
}

// PullFromSource pulls latest data from external source
func (m *Manager) PullFromSource(ctx context.Context, taskID string, source string) (*Task, error) {
	m.logger.Debug("pulling task from source", "task_id", taskID, "source", source)

	// Get current task
	task, err := m.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Check if task has this source
	taskSource, exists := task.Sources[source]
	if !exists {
		return nil, fmt.Errorf("task %s is not linked to source %s", taskID, source)
	}

	// Fetch latest data from source
	sourceData, err := m.fetchFromSource(ctx, taskSource.ExternalID, source)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", source, err)
	}

	// Update task with source data
	if err := m.updateTaskFromSourceData(ctx, task, sourceData, source); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Update source metadata
	taskSource.LastSync = time.Now()
	task.Sources[source] = taskSource

	// Save updated task
	if err := m.saveTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	m.logger.Info("task pulled from source successfully", "task_id", taskID, "source", source)

	return task, nil
}

// PushToSource pushes task data to external source
func (m *Manager) PushToSource(ctx context.Context, taskID string, source string) (*SyncResult, error) {
	m.logger.Debug("pushing task to source", "task_id", taskID, "source", source)

	// Get current task
	task, err := m.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Check if task has this source
	taskSource, exists := task.Sources[source]
	if !exists {
		return nil, fmt.Errorf("task %s is not linked to source %s", taskID, source)
	}

	// Convert task to plugin format
	pluginTaskData := m.convertTaskToPluginData(task)

	// Get plugin instance
	pluginInstance, err := m.getOrCreatePlugin(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}

	// Update external task
	start := time.Now()
	updatedData, err := pluginInstance.UpdateTask(ctx, taskSource.ExternalID, pluginTaskData, &plugin.UpdateOptions{
		SyncBack:       true,
		ValidateFields: true,
	})

	result := &SyncResult{
		TaskID:    taskID,
		Source:    source,
		Direction: SyncDirectionPush,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.ChangedFields = m.detectChangedFields(pluginTaskData, updatedData)

	// Update source metadata
	taskSource.LastSync = time.Now()
	task.Sources[source] = taskSource

	// Save updated task
	if err := m.saveTask(ctx, task); err != nil {
		m.logger.Warn("failed to save task after push", "task_id", taskID, "error", err)
	}

	m.logger.Info("task pushed to source successfully", "task_id", taskID, "source", source)

	return result, nil
}

// SyncTask synchronizes a task with all its external sources
func (m *Manager) SyncTask(ctx context.Context, taskID string, opts *SyncOptions) (*SyncResult, error) {
	m.logger.Debug("syncing task", "task_id", taskID, "direction", opts.Direction)

	// Get current task
	task, err := m.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Determine sources to sync
	sourcesToSync := opts.Sources
	if len(sourcesToSync) == 0 {
		// Sync all sources
		for source := range task.Sources {
			sourcesToSync = append(sourcesToSync, source)
		}
	}

	// For now, sync with the first source (single source sync)
	// In the future, this would handle multi-source sync with conflict resolution
	if len(sourcesToSync) == 0 {
		return nil, fmt.Errorf("no sources to sync for task: %s", taskID)
	}

	source := sourcesToSync[0]

	switch opts.Direction {
	case SyncDirectionPull:
		_, err := m.PullFromSource(ctx, taskID, source)
		if err != nil {
			return &SyncResult{
				TaskID:    taskID,
				Source:    source,
				Success:   false,
				Direction: opts.Direction,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}, err
		}

	case SyncDirectionPush:
		return m.PushToSource(ctx, taskID, source)

	case SyncDirectionBidirectional:
		// First pull, then push (simple bidirectional sync)
		if _, err := m.PullFromSource(ctx, taskID, source); err != nil {
			return &SyncResult{
				TaskID:    taskID,
				Source:    source,
				Success:   false,
				Direction: opts.Direction,
				Error:     fmt.Sprintf("pull failed: %v", err),
				Timestamp: time.Now(),
			}, err
		}

		return m.PushToSource(ctx, taskID, source)
	}

	return &SyncResult{
		TaskID:    taskID,
		Source:    source,
		Success:   true,
		Direction: opts.Direction,
		Timestamp: time.Now(),
	}, nil
}

// ListTasks returns a list of tasks matching the given filter
func (m *Manager) ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error) {
	// TODO: Implement proper task listing from workspace
	// For now, return empty list to satisfy interface
	return []*Task{}, nil
}

// SyncAllTasks synchronizes all tasks with their external sources
func (m *Manager) SyncAllTasks(ctx context.Context, opts *SyncOptions) ([]*SyncResult, error) {
	// List all tasks
	tasks, err := m.ListTasks(ctx, &TaskFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	var results []*SyncResult
	for _, task := range tasks {
		// Skip tasks without external sources
		if len(task.Sources) == 0 {
			continue
		}

		// Sync each task
		result, err := m.SyncTask(ctx, task.ID, opts)
		if err != nil {
			// Log error but continue with other tasks
			m.logger.Warn("failed to sync task", "task_id", task.ID, "error", err)
			result = &SyncResult{
				TaskID:    task.ID,
				Success:   false,
				Direction: opts.Direction,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// Helper methods

// taskExists checks if a task exists in the workspace
func (m *Manager) taskExists(taskID string) bool {
	ws, err := m.factory.WorkspaceManager()
	if err != nil {
		return false
	}

	status, err := ws.Status()
	if err != nil {
		return false
	}

	taskDir := filepath.Join(status.Root, ".zen", "work", "tasks", taskID)
	_, err = os.Stat(taskDir)
	return err == nil
}

// validateCreateRequest validates a create task request
func (m *Manager) validateCreateRequest(request *CreateTaskRequest) error {
	if request.ID == "" {
		return fmt.Errorf("task ID is required")
	}

	// Validate task ID format
	if len(request.ID) < 3 {
		return fmt.Errorf("task ID must be at least 3 characters long")
	}

	if strings.ContainsAny(request.ID, " /\\:*?\"<>|") {
		return fmt.Errorf("task ID contains invalid characters")
	}

	return nil
}

// fetchFromSource fetches task data from external source
func (m *Manager) fetchFromSource(ctx context.Context, taskID string, source string) (*TaskData, error) {
	// Use the existing operations for now
	// In the future, this would use the integration framework directly
	ops := NewOperations(m.factory)
	return ops.FetchFromSource(ctx, taskID, source)
}

// getOrCreatePlugin gets or creates a plugin instance
func (m *Manager) getOrCreatePlugin(ctx context.Context, source string) (plugin.IntegrationPluginInterface, error) {
	if m.clientFactory == nil {
		return nil, fmt.Errorf("integration not available")
	}

	return m.clientFactory.CreatePlugin(ctx, source)
}

// convertTaskToPluginData converts a task to plugin task data format
func (m *Manager) convertTaskToPluginData(task *Task) *plugin.TaskData {
	return &plugin.TaskData{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		Type:        task.Type,
		Owner:       task.Owner,
		Assignee:    task.Owner,
		Team:        task.Team,
		Created:     task.Created,
		Updated:     task.Updated,
		DueDate:     task.DueDate,
		Labels:      task.Labels,
		Tags:        task.Tags,
		Metadata:    task.Metadata,
	}
}

// detectChangedFields detects which fields changed between local and remote
func (m *Manager) detectChangedFields(local, remote *plugin.TaskData) []string {
	var changed []string

	if local.Title != remote.Title {
		changed = append(changed, "title")
	}
	if local.Status != remote.Status {
		changed = append(changed, "status")
	}
	if local.Priority != remote.Priority {
		changed = append(changed, "priority")
	}
	if local.Assignee != remote.Assignee {
		changed = append(changed, "assignee")
	}

	return changed
}

// updateTaskFromSourceData updates task with data from external source
func (m *Manager) updateTaskFromSourceData(ctx context.Context, task *Task, sourceData *TaskData, source string) error {
	// Update task fields with source data
	task.Title = sourceData.Title
	task.Status = sourceData.Status
	task.Priority = sourceData.Priority
	task.Updated = time.Now()

	if sourceData.Owner != "" {
		task.Owner = sourceData.Owner
	}
	if sourceData.Team != "" {
		task.Team = sourceData.Team
	}
	if len(sourceData.Labels) > 0 {
		task.Labels = sourceData.Labels
	}

	return nil
}

// createTaskStructure creates the complete task directory structure
func (m *Manager) createTaskStructure(ctx context.Context, task *Task, request *CreateTaskRequest, sourceData *TaskData) error {
	// Get workspace manager
	ws, err := m.factory.WorkspaceManager()
	if err != nil {
		return fmt.Errorf("failed to get workspace manager: %w", err)
	}

	status, err := ws.Status()
	if err != nil {
		return fmt.Errorf("failed to get workspace status: %w", err)
	}

	// Create task directory
	taskDir := filepath.Join(status.Root, ".zen", "work", "tasks", task.ID)
	if err := ws.CreateTaskDirectory(taskDir); err != nil {
		return fmt.Errorf("failed to create task directories: %w", err)
	}

	// Set task paths
	task.WorkspacePath = taskDir
	task.IndexPath = filepath.Join(taskDir, "index.md")
	task.ManifestPath = filepath.Join(taskDir, "manifest.yaml")
	task.MetadataPath = filepath.Join(taskDir, "metadata")

	// Generate task files from templates
	if err := m.generateTaskFiles(ctx, task, request, sourceData); err != nil {
		return fmt.Errorf("failed to generate task files: %w", err)
	}

	// Save source metadata if available
	if sourceData != nil {
		if err := m.saveSourceMetadata(taskDir, sourceData, request.FromSource); err != nil {
			m.logger.Warn("failed to save source metadata", "error", err)
		}
	}

	return nil
}

// generateTaskFiles generates task files from templates with source data
func (m *Manager) generateTaskFiles(ctx context.Context, task *Task, request *CreateTaskRequest, sourceData *TaskData) error {
	// Create template loader
	templateLoader := templates.NewLocalTemplateLoader()

	// Build template variables with source data sync
	variables := m.buildTemplateVariables(task, request, sourceData)

	// Generate files
	files := map[string]string{
		"index.md":      "index.md",
		"manifest.yaml": "manifest.yaml",
		"taskrc.yaml":   ".taskrc.yaml", // Template is taskrc.yaml, output is .taskrc.yaml
	}

	for templateName, fileName := range files {
		if err := m.generateFileFromTemplate(templateLoader, templateName, task.WorkspacePath, fileName, variables); err != nil {
			return fmt.Errorf("failed to generate %s: %w", fileName, err)
		}
	}

	return nil
}

// buildTemplateVariables builds comprehensive template variables with source data
func (m *Manager) buildTemplateVariables(task *Task, request *CreateTaskRequest, sourceData *TaskData) map[string]interface{} {
	now := time.Now()

	// Base template variables
	variables := map[string]interface{}{
		// Core task information
		"TASK_ID":     task.ID,
		"TASK_TITLE":  task.Title,
		"TASK_TYPE":   task.Type,
		"TASK_STATUS": task.Status,
		"PRIORITY":    task.Priority,

		// Ownership and team
		"OWNER_NAME":      task.Owner,
		"OWNER_EMAIL":     fmt.Sprintf("%s@company.com", strings.ReplaceAll(task.Owner, " ", ".")),
		"GITHUB_USERNAME": strings.ToLower(strings.ReplaceAll(task.Owner, " ", "")),
		"TEAM_NAME":       task.Team,

		// Dates
		"CREATED_DATE": now.Format("2006-01-02"),
		"LAST_UPDATED": now.Format("2006-01-02 15:04:05"),
		"TARGET_DATE":  now.AddDate(0, 0, 14).Format("2006-01-02"),

		// Workflow
		"CURRENT_STAGE":          task.CurrentStage,
		"current_stage_name":     "Align",
		"stage_number":           "1",
		"current_stage_progress": "0",

		// Integration flags (default to false)
		"JIRA_INTEGRATION":   false,
		"GITHUB_INTEGRATION": false,
		"LINEAR_INTEGRATION": false,
		"SYNC_ENABLED":       false,
		"EXTERNAL_SYSTEM":    "",

		// Labels and organization
		"LABELS": []string{task.Type},
		"TAGS":   []string{task.Type, task.Team},
	}

	// Sync data from external source if available
	if sourceData != nil {
		m.syncDataToTemplateVariables(variables, sourceData, request.FromSource)
	}

	// Add custom template variables
	if request.TemplateVars != nil {
		for key, value := range request.TemplateVars {
			variables[key] = value
		}
	}

	return variables
}

// syncDataToTemplateVariables syncs external source data to template variables
func (m *Manager) syncDataToTemplateVariables(variables map[string]interface{}, sourceData *TaskData, source string) {
	// Set integration flags
	variables[fmt.Sprintf("%s_INTEGRATION", strings.ToUpper(source))] = true
	variables["EXTERNAL_SYSTEM"] = source
	variables["SYNC_ENABLED"] = true

	// Sync core fields (External System of Record overrides)
	if sourceData.Title != "" {
		variables["TASK_TITLE"] = sourceData.Title
	}
	if sourceData.Type != "" {
		variables["TASK_TYPE"] = sourceData.Type
	}
	if sourceData.Status != "" {
		variables["TASK_STATUS"] = m.mapExternalStatusToZenStatus(sourceData.Status, source)
	}
	if sourceData.Priority != "" {
		variables["PRIORITY"] = m.mapExternalPriorityToZenPriority(sourceData.Priority, source)
	}
	if sourceData.Assignee != "" {
		variables["OWNER_NAME"] = sourceData.Assignee
		variables["OWNER_EMAIL"] = m.extractEmailFromAssignee(sourceData.Assignee)
		variables["GITHUB_USERNAME"] = m.extractUsernameFromAssignee(sourceData.Assignee)
	}
	if sourceData.Team != "" {
		variables["TEAM_NAME"] = sourceData.Team
	}
	if sourceData.ExternalURL != "" {
		variables[fmt.Sprintf("%s_URL", strings.ToUpper(source))] = sourceData.ExternalURL
	}

	// Sync timestamps
	if !sourceData.Created.IsZero() {
		variables["CREATED_DATE"] = sourceData.Created.Format("2006-01-02")
		variables[fmt.Sprintf("%s_CREATED_RAW", strings.ToUpper(source))] = sourceData.Created.Format(time.RFC3339)
	}
	if !sourceData.Updated.IsZero() {
		variables["LAST_UPDATED"] = sourceData.Updated.Format("2006-01-02 15:04:05")
		variables[fmt.Sprintf("%s_UPDATED_RAW", strings.ToUpper(source))] = sourceData.Updated.Format(time.RFC3339)
	}

	// Sync labels and components
	if len(sourceData.Labels) > 0 {
		variables["LABELS"] = sourceData.Labels
	}
	if len(sourceData.Components) > 0 {
		variables[fmt.Sprintf("%s_COMPONENTS", strings.ToUpper(source))] = sourceData.Components
	}

	// Add external system specific data
	variables[fmt.Sprintf("%s_EXTERNAL_ID", strings.ToUpper(source))] = sourceData.ExternalID

	// Extract rich data from raw response
	if sourceData.RawData != nil {
		m.syncSourceSpecificData(variables, sourceData.RawData, source)
	}
}

// syncSourceSpecificData extracts source-specific rich data
func (m *Manager) syncSourceSpecificData(variables map[string]interface{}, rawData map[string]interface{}, source string) {
	switch source {
	case "jira":
		m.syncJiraSpecificData(variables, rawData)
	case "github":
		m.syncGitHubSpecificData(variables, rawData)
	case "linear":
		m.syncLinearSpecificData(variables, rawData)
	}
}

// Data mapping helper methods (moved from create command)

func (m *Manager) mapExternalStatusToZenStatus(externalStatus, source string) string {
	switch source {
	case "jira":
		switch strings.ToLower(externalStatus) {
		case "to do", "open", "new":
			return "proposed"
		case "in progress", "doing":
			return "in_progress"
		case "done", "closed", "resolved":
			return "completed"
		case "blocked":
			return "blocked"
		default:
			return "proposed"
		}
	case "github":
		switch strings.ToLower(externalStatus) {
		case "open":
			return "proposed"
		case "closed":
			return "completed"
		default:
			return "proposed"
		}
	default:
		return strings.ToLower(strings.ReplaceAll(externalStatus, " ", "_"))
	}
}

func (m *Manager) mapExternalPriorityToZenPriority(externalPriority, source string) string {
	switch source {
	case "jira":
		switch strings.ToLower(externalPriority) {
		case "highest", "critical":
			return "P0"
		case "high":
			return "P1"
		case "medium":
			return "P2"
		case "low", "lowest":
			return "P3"
		default:
			return "P2"
		}
	default:
		return "P2"
	}
}

func (m *Manager) extractEmailFromAssignee(assignee string) string {
	if strings.Contains(assignee, "@") {
		return assignee
	}
	username := strings.ToLower(strings.ReplaceAll(assignee, " ", "."))
	return fmt.Sprintf("%s@company.com", username)
}

func (m *Manager) extractUsernameFromAssignee(assignee string) string {
	return strings.ToLower(strings.ReplaceAll(assignee, " ", ""))
}
