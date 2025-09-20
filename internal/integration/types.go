package integration

import (
	"context"
	"time"
)

// SyncDirection represents the direction of task synchronization
type SyncDirection string

const (
	SyncDirectionPull          SyncDirection = "pull"
	SyncDirectionPush          SyncDirection = "push"
	SyncDirectionBidirectional SyncDirection = "bidirectional"
)

// ConflictStrategy represents how to handle sync conflicts
type ConflictStrategy string

const (
	ConflictStrategyLocalWins    ConflictStrategy = "local_wins"
	ConflictStrategyRemoteWins   ConflictStrategy = "remote_wins"
	ConflictStrategyManualReview ConflictStrategy = "manual_review"
	ConflictStrategyTimestamp    ConflictStrategy = "timestamp"
)

// TaskSyncRecord represents a synchronization relationship between a Zen task and external system
type TaskSyncRecord struct {
	TaskID           string                 `json:"task_id" yaml:"task_id"`
	ExternalID       string                 `json:"external_id" yaml:"external_id"`
	ExternalSystem   string                 `json:"external_system" yaml:"external_system"`
	LastSyncTime     time.Time              `json:"last_sync_time" yaml:"last_sync_time"`
	SyncDirection    SyncDirection          `json:"sync_direction" yaml:"sync_direction"`
	FieldMappings    map[string]string      `json:"field_mappings" yaml:"field_mappings"`
	ConflictStrategy ConflictStrategy       `json:"conflict_strategy" yaml:"conflict_strategy"`
	Metadata         map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// ExternalTaskData represents task data from an external system
type ExternalTaskData struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority"`
	Assignee    string                 `json:"assignee"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Fields      map[string]interface{} `json:"fields"`
}

// ZenTaskData represents task data in Zen format
type ZenTaskData struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority"`
	Owner       string                 `json:"owner"`
	Team        string                 `json:"team"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	Success       bool                   `json:"success"`
	TaskID        string                 `json:"task_id"`
	ExternalID    string                 `json:"external_id"`
	Direction     SyncDirection          `json:"direction"`
	ChangedFields []string               `json:"changed_fields"`
	Conflicts     []string               `json:"conflicts"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SyncOptions represents options for synchronization operations
type SyncOptions struct {
	Direction        SyncDirection    `json:"direction"`
	ConflictStrategy ConflictStrategy `json:"conflict_strategy"`
	DryRun           bool             `json:"dry_run"`
	ForceSync        bool             `json:"force_sync"`
}

// IntegrationProvider represents an external system integration provider
type IntegrationProvider interface {
	// Name returns the provider name (e.g., "jira", "github")
	Name() string

	// GetTaskData retrieves task data from the external system
	GetTaskData(ctx context.Context, externalID string) (*ExternalTaskData, error)

	// CreateTask creates a new task in the external system
	CreateTask(ctx context.Context, taskData *ZenTaskData) (*ExternalTaskData, error)

	// UpdateTask updates an existing task in the external system
	UpdateTask(ctx context.Context, externalID string, taskData *ZenTaskData) (*ExternalTaskData, error)

	// SearchTasks searches for tasks in the external system
	SearchTasks(ctx context.Context, query map[string]interface{}) ([]*ExternalTaskData, error)

	// ValidateConnection tests the connection to the external system
	ValidateConnection(ctx context.Context) error

	// GetFieldMapping returns the field mapping configuration
	GetFieldMapping() map[string]string

	// MapToZen converts external task data to Zen format
	MapToZen(external *ExternalTaskData) (*ZenTaskData, error)

	// MapToExternal converts Zen task data to external format
	MapToExternal(zen *ZenTaskData) (*ExternalTaskData, error)
}

// TaskSyncInterface defines the interface for task synchronization operations
type TaskSyncInterface interface {
	// SyncTask synchronizes a single task
	SyncTask(ctx context.Context, taskID string, opts SyncOptions) (*SyncResult, error)

	// SyncAllTasks synchronizes all tasks configured for sync
	SyncAllTasks(ctx context.Context, opts SyncOptions) ([]*SyncResult, error)

	// GetSyncRecord retrieves the sync record for a task
	GetSyncRecord(ctx context.Context, taskID string) (*TaskSyncRecord, error)

	// CreateSyncRecord creates a new sync record
	CreateSyncRecord(ctx context.Context, record *TaskSyncRecord) error

	// UpdateSyncRecord updates an existing sync record
	UpdateSyncRecord(ctx context.Context, record *TaskSyncRecord) error

	// DeleteSyncRecord deletes a sync record
	DeleteSyncRecord(ctx context.Context, taskID string) error

	// ListSyncRecords lists all sync records
	ListSyncRecords(ctx context.Context) ([]*TaskSyncRecord, error)
}

// DataMapperInterface defines the interface for data mapping operations
type DataMapperInterface interface {
	// MapFields maps fields between external and Zen formats
	MapFields(source map[string]interface{}, mapping map[string]string) (map[string]interface{}, error)

	// ValidateMapping validates a field mapping configuration
	ValidateMapping(mapping map[string]string) error

	// GetDefaultMapping returns the default field mapping for a provider
	GetDefaultMapping(provider string) map[string]string
}

// IntegrationManagerInterface defines the main integration management interface
type IntegrationManagerInterface interface {
	// GetProvider returns a provider by name
	GetProvider(name string) (IntegrationProvider, error)

	// RegisterProvider registers a new integration provider
	RegisterProvider(provider IntegrationProvider) error

	// ListProviders returns all registered providers
	ListProviders() []string

	// IsConfigured checks if integration is properly configured
	IsConfigured() bool

	// GetTaskSystem returns the configured task system of record
	GetTaskSystem() string

	// IsSyncEnabled returns true if sync is enabled
	IsSyncEnabled() bool
}
