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

// SyncStatus represents the current status of a sync record
type SyncStatus string

const (
	SyncStatusActive   SyncStatus = "active"
	SyncStatusPaused   SyncStatus = "paused"
	SyncStatusError    SyncStatus = "error"
	SyncStatusConflict SyncStatus = "conflict"
	SyncStatusDisabled SyncStatus = "disabled"
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
	CreatedAt        time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" yaml:"updated_at"`
	Version          int64                  `json:"version" yaml:"version"`
	Status           SyncStatus             `json:"status" yaml:"status"`
	ErrorCount       int                    `json:"error_count" yaml:"error_count"`
	LastError        string                 `json:"last_error,omitempty" yaml:"last_error,omitempty"`
	RetryAfter       *time.Time             `json:"retry_after,omitempty" yaml:"retry_after,omitempty"`
	DataHash         string                 `json:"data_hash,omitempty" yaml:"data_hash,omitempty"`
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
	Conflicts     []FieldConflict        `json:"conflicts"`
	Error         string                 `json:"error,omitempty"`
	ErrorCode     string                 `json:"error_code,omitempty"`
	Retryable     bool                   `json:"retryable,omitempty"`
	Duration      time.Duration          `json:"duration"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SyncOptions represents options for synchronization operations
type SyncOptions struct {
	Direction        SyncDirection    `json:"direction"`
	ConflictStrategy ConflictStrategy `json:"conflict_strategy"`
	DryRun           bool             `json:"dry_run"`
	ForceSync        bool             `json:"force_sync"`
	Timeout          time.Duration    `json:"timeout,omitempty"`
	RetryCount       int              `json:"retry_count,omitempty"`
	BatchSize        int              `json:"batch_size,omitempty"`
	Parallel         bool             `json:"parallel,omitempty"`
	UserID           string           `json:"user_id,omitempty"`
	CorrelationID    string           `json:"correlation_id,omitempty"`
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

	// HealthCheck performs a health check on the provider
	HealthCheck(ctx context.Context) (*ProviderHealth, error)

	// GetRateLimitInfo returns current rate limit information
	GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error)

	// SupportsRealtime returns true if the provider supports real-time updates
	SupportsRealtime() bool

	// GetWebhookURL returns the webhook URL for real-time updates (if supported)
	GetWebhookURL() string
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

// IntegrationError represents an integration-specific error
type IntegrationError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Provider  string                 `json:"provider,omitempty"`
	TaskID    string                 `json:"task_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Retryable bool                   `json:"retryable"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

func (e IntegrationError) Error() string {
	return e.Message
}

// Error codes for integration operations
const (
	ErrCodePluginNotFound   = "PLUGIN_NOT_FOUND"
	ErrCodePluginLoadFailed = "PLUGIN_LOAD_FAILED"
	ErrCodeAuthFailed       = "AUTH_FAILED"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeNetworkError     = "NETWORK_ERROR"
	ErrCodeSyncConflict     = "SYNC_CONFLICT"
	ErrCodeInvalidData      = "INVALID_DATA"
	ErrCodeConfigError      = "CONFIG_ERROR"
	ErrCodeProviderError    = "PROVIDER_ERROR"
	ErrCodeTimeoutError     = "TIMEOUT_ERROR"
)

// FieldConflict represents a conflict between local and external field values
type FieldConflict struct {
	Field             string      `json:"field"`
	ZenValue          interface{} `json:"zen_value"`
	ExternalValue     interface{} `json:"external_value"`
	ZenTimestamp      time.Time   `json:"zen_timestamp"`
	ExternalTimestamp time.Time   `json:"external_timestamp"`
	Resolution        string      `json:"resolution,omitempty"`
}

// ConflictRecord represents a conflict that requires manual resolution
type ConflictRecord struct {
	ID         string          `json:"id"`
	TaskID     string          `json:"task_id"`
	Conflicts  []FieldConflict `json:"conflicts"`
	CreatedAt  time.Time       `json:"created_at"`
	Status     ConflictStatus  `json:"status"`
	ResolvedBy string          `json:"resolved_by,omitempty"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
}

// ConflictStatus represents the status of a conflict resolution
type ConflictStatus string

const (
	ConflictStatusPending  ConflictStatus = "pending"
	ConflictStatusResolved ConflictStatus = "resolved"
	ConflictStatusIgnored  ConflictStatus = "ignored"
)

// ProviderHealth represents the health status of a provider
type ProviderHealth struct {
	Provider      string         `json:"provider"`
	Healthy       bool           `json:"healthy"`
	LastChecked   time.Time      `json:"last_checked"`
	ResponseTime  time.Duration  `json:"response_time"`
	ErrorCount    int            `json:"error_count"`
	LastError     string         `json:"last_error,omitempty"`
	RateLimitInfo *RateLimitInfo `json:"rate_limit_info,omitempty"`
}

// RateLimitInfo contains rate limiting information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
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

	// GetProviderHealth returns health status for a provider
	GetProviderHealth(ctx context.Context, provider string) (*ProviderHealth, error)

	// GetAllProviderHealth returns health status for all providers
	GetAllProviderHealth(ctx context.Context) (map[string]*ProviderHealth, error)
}
