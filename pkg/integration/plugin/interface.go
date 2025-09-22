package plugin

import (
	"context"
	"time"
)

// IntegrationPluginInterface defines the standardized plugin contract for external system integrations
type IntegrationPluginInterface interface {
	// Plugin Identity and Lifecycle
	Name() string
	Version() string
	Description() string

	// Lifecycle Management
	Initialize(ctx context.Context, config *PluginConfig) error
	Validate(ctx context.Context) error
	HealthCheck(ctx context.Context) (*PluginHealth, error)
	Shutdown(ctx context.Context) error

	// Core Operations
	FetchTask(ctx context.Context, externalID string, opts *FetchOptions) (*TaskData, error)
	CreateTask(ctx context.Context, taskData *TaskData, opts *CreateOptions) (*TaskData, error)
	UpdateTask(ctx context.Context, externalID string, taskData *TaskData, opts *UpdateOptions) (*TaskData, error)
	DeleteTask(ctx context.Context, externalID string, opts *DeleteOptions) error
	SearchTasks(ctx context.Context, query *SearchQuery, opts *SearchOptions) ([]*TaskData, error)

	// Synchronization
	SyncTask(ctx context.Context, taskID string, opts *SyncOptions) (*SyncResult, error)
	GetSyncMetadata(ctx context.Context, taskID string) (*SyncMetadata, error)

	// Data Mapping
	MapToZen(ctx context.Context, externalData interface{}) (*TaskData, error)
	MapToExternal(ctx context.Context, zenData *TaskData) (interface{}, error)
	GetFieldMapping() *FieldMappingConfig

	// Authentication and Configuration
	GetAuthConfig() *AuthConfig
	GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error)
	SupportsOperation(operation OperationType) bool
}

// LifecycleInterface defines plugin lifecycle methods
type LifecycleInterface interface {
	Initialize(ctx context.Context, config *PluginConfig) error
	Validate(ctx context.Context) error
	HealthCheck(ctx context.Context) (*PluginHealth, error)
	Shutdown(ctx context.Context) error
}

// HealthCheckInterface defines health check capabilities
type HealthCheckInterface interface {
	HealthCheck(ctx context.Context) (*PluginHealth, error)
	GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error)
}

// PluginConfig contains plugin configuration
type PluginConfig struct {
	// Plugin identification
	Name    string `json:"name" yaml:"name" validate:"required"`
	Version string `json:"version" yaml:"version" validate:"required,semver"`
	Enabled bool   `json:"enabled" yaml:"enabled"`

	// Connection settings
	BaseURL    string            `json:"base_url" yaml:"base_url" validate:"required,url"`
	Timeout    time.Duration     `json:"timeout" yaml:"timeout"`
	MaxRetries int               `json:"max_retries" yaml:"max_retries"`
	Headers    map[string]string `json:"headers" yaml:"headers"`

	// Authentication
	Auth *AuthConfig `json:"auth" yaml:"auth" validate:"required"`

	// Rate limiting
	RateLimit *RateLimitConfig `json:"rate_limit" yaml:"rate_limit"`

	// Field mapping
	FieldMapping *FieldMappingConfig `json:"field_mapping" yaml:"field_mapping"`

	// Caching
	Cache *CacheConfig `json:"cache" yaml:"cache"`

	// Plugin-specific settings
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// TaskData represents standardized task data
type TaskData struct {
	// Core task fields
	ID          string `json:"id" yaml:"id" validate:"required"`
	ExternalID  string `json:"external_id" yaml:"external_id"`
	Title       string `json:"title" yaml:"title" validate:"required"`
	Description string `json:"description" yaml:"description"`
	Status      string `json:"status" yaml:"status" validate:"required"`
	Priority    string `json:"priority" yaml:"priority"`
	Type        string `json:"type" yaml:"type"`

	// Ownership and team
	Owner    string `json:"owner" yaml:"owner"`
	Assignee string `json:"assignee" yaml:"assignee"`
	Team     string `json:"team" yaml:"team"`

	// Timestamps
	Created time.Time  `json:"created" yaml:"created"`
	Updated time.Time  `json:"updated" yaml:"updated"`
	DueDate *time.Time `json:"due_date,omitempty" yaml:"due_date,omitempty"`

	// Organization
	Labels     []string `json:"labels" yaml:"labels"`
	Tags       []string `json:"tags" yaml:"tags"`
	Components []string `json:"components" yaml:"components"`

	// External system specific
	ExternalURL string                 `json:"external_url" yaml:"external_url"`
	RawData     map[string]interface{} `json:"raw_data,omitempty" yaml:"raw_data,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
	Version  int64                  `json:"version" yaml:"version"`
	Checksum string                 `json:"checksum" yaml:"checksum"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Type           AuthType          `json:"type" yaml:"type" validate:"required"`
	CredentialsRef string            `json:"credentials_ref" yaml:"credentials_ref"`
	TokenStorage   string            `json:"token_storage" yaml:"token_storage"`
	RefreshToken   bool              `json:"refresh_token" yaml:"refresh_token"`
	TokenExpiry    time.Duration     `json:"token_expiry" yaml:"token_expiry"`
	CustomFields   map[string]string `json:"custom_fields,omitempty" yaml:"custom_fields,omitempty"`
}

// AuthType represents authentication types
type AuthType string

const (
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeCustom AuthType = "custom"
)

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute" yaml:"requests_per_minute"`
	BurstSize         int           `json:"burst_size" yaml:"burst_size"`
	BackoffStrategy   string        `json:"backoff_strategy" yaml:"backoff_strategy"`
	MaxRetries        int           `json:"max_retries" yaml:"max_retries"`
	BaseDelay         time.Duration `json:"base_delay" yaml:"base_delay"`
}

// CacheConfig contains caching configuration
type CacheConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	TTL            time.Duration `json:"ttl" yaml:"ttl"`
	MaxSize        int           `json:"max_size" yaml:"max_size"`
	EvictionPolicy string        `json:"eviction_policy" yaml:"eviction_policy"`
}

// FieldMappingConfig represents field mapping configuration
type FieldMappingConfig struct {
	Mappings   []FieldMapping    `json:"mappings" yaml:"mappings"`
	Transforms []FieldTransform  `json:"transforms" yaml:"transforms"`
	Validation []FieldValidation `json:"validation" yaml:"validation"`
}

// FieldMapping represents a field mapping between systems
type FieldMapping struct {
	ZenField      string        `json:"zen_field" yaml:"zen_field" validate:"required"`
	ExternalField string        `json:"external_field" yaml:"external_field" validate:"required"`
	Direction     SyncDirection `json:"direction" yaml:"direction"`
	Required      bool          `json:"required" yaml:"required"`
	DefaultValue  interface{}   `json:"default_value,omitempty" yaml:"default_value,omitempty"`
}

// FieldTransform represents a field transformation
type FieldTransform struct {
	Field     string                 `json:"field" yaml:"field" validate:"required"`
	Type      TransformType          `json:"type" yaml:"type" validate:"required"`
	Config    map[string]interface{} `json:"config" yaml:"config"`
	Direction SyncDirection          `json:"direction" yaml:"direction"`
}

// FieldValidation represents field validation rules
type FieldValidation struct {
	Field string                 `json:"field" yaml:"field" validate:"required"`
	Rules map[string]interface{} `json:"rules" yaml:"rules"`
}

// TransformType represents transformation types
type TransformType string

const (
	TransformTypeMap      TransformType = "map"
	TransformTypeFormat   TransformType = "format"
	TransformTypeTemplate TransformType = "template"
	TransformTypeCustom   TransformType = "custom"
)

// SyncDirection represents sync direction
type SyncDirection string

const (
	SyncDirectionPull          SyncDirection = "pull"
	SyncDirectionPush          SyncDirection = "push"
	SyncDirectionBidirectional SyncDirection = "bidirectional"
)

// PluginHealth represents plugin health status
type PluginHealth struct {
	Provider     string        `json:"provider"`
	Healthy      bool          `json:"healthy"`
	LastChecked  time.Time     `json:"last_checked"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount   int           `json:"error_count"`
	LastError    string        `json:"last_error,omitempty"`
}

// RateLimitInfo contains rate limiting information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
}

// Operation options and results

// FetchOptions contains options for fetching tasks
type FetchOptions struct {
	IncludeRaw bool          `json:"include_raw"`
	Fields     []string      `json:"fields"`
	Timeout    time.Duration `json:"timeout"`
}

// CreateOptions contains options for creating tasks
type CreateOptions struct {
	SyncBack       bool `json:"sync_back"`
	ValidateFields bool `json:"validate_fields"`
}

// UpdateOptions contains options for updating tasks
type UpdateOptions struct {
	SyncBack       bool `json:"sync_back"`
	ValidateFields bool `json:"validate_fields"`
}

// DeleteOptions contains options for deleting tasks
type DeleteOptions struct {
	Force bool `json:"force"`
}

// SearchOptions contains options for searching tasks
type SearchOptions struct {
	MaxResults int           `json:"max_results"`
	StartAt    int           `json:"start_at"`
	Timeout    time.Duration `json:"timeout"`
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query   string                 `json:"query"`
	Filters map[string]interface{} `json:"filters"`
}

// SyncOptions contains options for synchronization
type SyncOptions struct {
	Direction        SyncDirection    `json:"direction"`
	ConflictStrategy ConflictStrategy `json:"conflict_strategy"`
	DryRun           bool             `json:"dry_run"`
	ForceSync        bool             `json:"force_sync"`
	Timeout          time.Duration    `json:"timeout"`
}

// ConflictStrategy represents conflict resolution strategy
type ConflictStrategy string

const (
	ConflictStrategyLocalWins    ConflictStrategy = "local_wins"
	ConflictStrategyRemoteWins   ConflictStrategy = "remote_wins"
	ConflictStrategyManualReview ConflictStrategy = "manual_review"
	ConflictStrategyTimestamp    ConflictStrategy = "timestamp"
)

// SyncResult represents sync operation result
type SyncResult struct {
	Success       bool          `json:"success"`
	TaskID        string        `json:"task_id"`
	ExternalID    string        `json:"external_id"`
	Direction     SyncDirection `json:"direction"`
	ChangedFields []string      `json:"changed_fields"`
	Duration      time.Duration `json:"duration"`
	Timestamp     time.Time     `json:"timestamp"`
	Error         string        `json:"error,omitempty"`
	ErrorCode     string        `json:"error_code,omitempty"`
	Retryable     bool          `json:"retryable,omitempty"`
}

// SyncMetadata represents sync metadata
type SyncMetadata struct {
	TaskID           string                 `json:"task_id"`
	ExternalID       string                 `json:"external_id"`
	LastSyncTime     time.Time              `json:"last_sync_time"`
	SyncDirection    SyncDirection          `json:"sync_direction"`
	ConflictStrategy ConflictStrategy       `json:"conflict_strategy"`
	Metadata         map[string]interface{} `json:"metadata"`
	Version          int64                  `json:"version"`
	Status           string                 `json:"status"`
}

// OperationType represents operation types
type OperationType string

const (
	OperationTypeFetch  OperationType = "fetch"
	OperationTypeCreate OperationType = "create"
	OperationTypeUpdate OperationType = "update"
	OperationTypeDelete OperationType = "delete"
	OperationTypeSearch OperationType = "search"
	OperationTypeSync   OperationType = "sync"
)
