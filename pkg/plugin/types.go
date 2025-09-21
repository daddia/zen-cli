package plugin

import (
	"fmt"
	"time"
)

// Manifest represents a plugin manifest file
type Manifest struct {
	SchemaVersion     string              `yaml:"schema_version" json:"schema_version"`
	Plugin            PluginMetadata      `yaml:"plugin" json:"plugin"`
	Capabilities      []string            `yaml:"capabilities" json:"capabilities"`
	Runtime           RuntimeConfig       `yaml:"runtime" json:"runtime"`
	APIRequirements   []string            `yaml:"api_requirements" json:"api_requirements"`
	Security          SecurityConfig      `yaml:"security" json:"security"`
	ConfigurationSchema map[string]ConfigField `yaml:"configuration_schema" json:"configuration_schema"`
}

// PluginMetadata contains basic plugin information
type PluginMetadata struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
	Author      string `yaml:"author" json:"author"`
	Homepage    string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	Repository  string `yaml:"repository,omitempty" json:"repository,omitempty"`
	License     string `yaml:"license,omitempty" json:"license,omitempty"`
	Keywords    []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`
}

// RuntimeConfig contains plugin runtime configuration
type RuntimeConfig struct {
	WASMFile         string        `yaml:"wasm_file" json:"wasm_file"`
	MemoryLimit      string        `yaml:"memory_limit" json:"memory_limit"`
	ExecutionTimeout time.Duration `yaml:"execution_timeout" json:"execution_timeout"`
	CPULimit         string        `yaml:"cpu_limit,omitempty" json:"cpu_limit,omitempty"`
	MaxInstances     int           `yaml:"max_instances,omitempty" json:"max_instances,omitempty"`
}

// SecurityConfig contains plugin security configuration
type SecurityConfig struct {
	Permissions []string `yaml:"permissions" json:"permissions"`
	Signature   string   `yaml:"signature,omitempty" json:"signature,omitempty"`
	Checksum    string   `yaml:"checksum" json:"checksum"`
	TrustedKeys []string `yaml:"trusted_keys,omitempty" json:"trusted_keys,omitempty"`
	Sandbox     SandboxConfig `yaml:"sandbox,omitempty" json:"sandbox,omitempty"`
}

// SandboxConfig contains sandbox configuration
type SandboxConfig struct {
	NetworkAccess   bool     `yaml:"network_access" json:"network_access"`
	FileSystemAccess bool    `yaml:"filesystem_access" json:"filesystem_access"`
	AllowedHosts    []string `yaml:"allowed_hosts,omitempty" json:"allowed_hosts,omitempty"`
	AllowedPaths    []string `yaml:"allowed_paths,omitempty" json:"allowed_paths,omitempty"`
}

// ConfigField represents a configuration field definition
type ConfigField struct {
	Type        string      `yaml:"type" json:"type"`
	Required    bool        `yaml:"required" json:"required"`
	Description string      `yaml:"description" json:"description"`
	Default     interface{} `yaml:"default,omitempty" json:"default,omitempty"`
	Validation  string      `yaml:"validation,omitempty" json:"validation,omitempty"`
	Options     []string    `yaml:"options,omitempty" json:"options,omitempty"`
	Sensitive   bool        `yaml:"sensitive,omitempty" json:"sensitive,omitempty"`
}

// PluginError represents a plugin-specific error
type PluginError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Plugin    string                 `json:"plugin,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

func (e PluginError) Error() string {
	return e.Message
}

// Plugin error codes
const (
	ErrCodePluginNotFound      = "PLUGIN_NOT_FOUND"
	ErrCodePluginLoadFailed    = "PLUGIN_LOAD_FAILED"
	ErrCodePluginExecutionFailed = "PLUGIN_EXECUTION_FAILED"
	ErrCodeInvalidManifest     = "INVALID_MANIFEST"
	ErrCodeSecurityViolation   = "SECURITY_VIOLATION"
	ErrCodeResourceExhausted   = "RESOURCE_EXHAUSTED"
	ErrCodeTimeoutError        = "TIMEOUT_ERROR"
	ErrCodeInvalidArguments    = "INVALID_ARGUMENTS"
	ErrCodeRuntimeError        = "RUNTIME_ERROR"
)

// HostAPI represents the host API interface available to plugins
type HostAPI interface {
	// HTTP operations
	HTTPRequest(method, url string, headers map[string]string, body []byte) ([]byte, error)
	
	// Configuration access
	GetConfig(key string) (string, error)
	GetCredentials(credentialRef string) (string, error)
	
	// Logging operations
	Log(level string, message string) error
	
	// Task operations (if permitted)
	GetTask(taskID string) ([]byte, error)
	UpdateTask(taskID string, data []byte) error
	
	// Validation operations
	ValidateConfig(configJSON string, schemaName string) error
}

// PluginCapability represents a plugin capability
type PluginCapability string

const (
	CapabilityTaskSync       PluginCapability = "task_sync"
	CapabilityFieldMapping   PluginCapability = "field_mapping"
	CapabilityWebhookSupport PluginCapability = "webhook_support"
	CapabilityRealTimeSync   PluginCapability = "real_time_sync"
	CapabilityBatchOperations PluginCapability = "batch_operations"
	CapabilityDataValidation PluginCapability = "data_validation"
	CapabilityCustomFields   PluginCapability = "custom_fields"
)

// PluginPermission represents a plugin permission
type PluginPermission string

const (
	PermissionNetworkHTTPOutbound PluginPermission = "network.http.outbound"
	PermissionNetworkHTTPInbound  PluginPermission = "network.http.inbound"
	PermissionConfigRead          PluginPermission = "config.read"
	PermissionConfigWrite         PluginPermission = "config.write"
	PermissionTaskRead            PluginPermission = "task.read"
	PermissionTaskWrite           PluginPermission = "task.write"
	PermissionTaskReadWrite       PluginPermission = "task.read_write"
	PermissionCredentialRead      PluginPermission = "credential.read"
	PermissionLogging             PluginPermission = "logging"
	PermissionFileSystemRead      PluginPermission = "filesystem.read"
	PermissionFileSystemWrite     PluginPermission = "filesystem.write"
)

// ExecutionContext contains context for plugin execution
type ExecutionContext struct {
	PluginName    string                 `json:"plugin_name"`
	Function      string                 `json:"function"`
	RequestID     string                 `json:"request_id"`
	UserID        string                 `json:"user_id,omitempty"`
	Timeout       time.Duration          `json:"timeout"`
	Permissions   []PluginPermission     `json:"permissions"`
	Configuration map[string]interface{} `json:"configuration"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionResult contains the result of plugin execution
type ExecutionResult struct {
	Success   bool                   `json:"success"`
	Data      []byte                 `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PluginMetrics contains plugin performance metrics
type PluginMetrics struct {
	LoadTime          time.Duration `json:"load_time"`
	ExecutionCount    int64         `json:"execution_count"`
	SuccessCount      int64         `json:"success_count"`
	ErrorCount        int64         `json:"error_count"`
	AverageExecTime   time.Duration `json:"average_exec_time"`
	LastExecutionTime time.Time     `json:"last_execution_time"`
	MemoryUsage       int64         `json:"memory_usage"`
	CPUUsage          float64       `json:"cpu_usage"`
}

// PluginHealth represents the health status of a plugin
type PluginHealth struct {
	Plugin      string        `json:"plugin"`
	Healthy     bool          `json:"healthy"`
	Status      PluginStatus  `json:"status"`
	LastChecked time.Time     `json:"last_checked"`
	Uptime      time.Duration `json:"uptime"`
	Version     string        `json:"version"`
	Metrics     *PluginMetrics `json:"metrics,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// PluginEvent represents an event in the plugin lifecycle
type PluginEvent struct {
	Type      PluginEventType        `json:"type"`
	Plugin    string                 `json:"plugin"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// PluginEventType represents the type of plugin event
type PluginEventType string

const (
	EventTypePluginDiscovered PluginEventType = "plugin_discovered"
	EventTypePluginLoaded     PluginEventType = "plugin_loaded"
	EventTypePluginUnloaded   PluginEventType = "plugin_unloaded"
	EventTypePluginError      PluginEventType = "plugin_error"
	EventTypePluginExecution  PluginEventType = "plugin_execution"
	EventTypePluginHealthCheck PluginEventType = "plugin_health_check"
)

// PluginConfig represents runtime configuration for a plugin
type PluginConfig struct {
	Name          string                 `json:"name"`
	Enabled       bool                   `json:"enabled"`
	Configuration map[string]interface{} `json:"configuration"`
	Permissions   []PluginPermission     `json:"permissions"`
	ResourceLimits ResourceLimits        `json:"resource_limits"`
}

// ResourceLimits represents resource limits for plugin execution
type ResourceLimits struct {
	MaxMemoryMB      int           `json:"max_memory_mb"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	MaxCPUPercent    float64       `json:"max_cpu_percent"`
	MaxInstances     int           `json:"max_instances"`
}

// DefaultResourceLimits returns default resource limits
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxMemoryMB:      10,
		MaxExecutionTime: 30 * time.Second,
		MaxCPUPercent:    50.0,
		MaxInstances:     1,
	}
}

// ValidatePermissions checks if the requested permissions are valid
func ValidatePermissions(permissions []string) error {
	validPermissions := map[string]bool{
		string(PermissionNetworkHTTPOutbound): true,
		string(PermissionNetworkHTTPInbound):  true,
		string(PermissionConfigRead):          true,
		string(PermissionConfigWrite):         true,
		string(PermissionTaskRead):            true,
		string(PermissionTaskWrite):           true,
		string(PermissionTaskReadWrite):       true,
		string(PermissionCredentialRead):      true,
		string(PermissionLogging):             true,
		string(PermissionFileSystemRead):      true,
		string(PermissionFileSystemWrite):     true,
	}
	
	for _, permission := range permissions {
		if !validPermissions[permission] {
			return &PluginError{
				Code:      ErrCodeSecurityViolation,
				Message:   fmt.Sprintf("invalid permission: %s", permission),
				Timestamp: time.Now(),
			}
		}
	}
	
	return nil
}

// ValidateCapabilities checks if the requested capabilities are valid
func ValidateCapabilities(capabilities []string) error {
	validCapabilities := map[string]bool{
		string(CapabilityTaskSync):        true,
		string(CapabilityFieldMapping):    true,
		string(CapabilityWebhookSupport):  true,
		string(CapabilityRealTimeSync):    true,
		string(CapabilityBatchOperations): true,
		string(CapabilityDataValidation):  true,
		string(CapabilityCustomFields):    true,
	}
	
	for _, capability := range capabilities {
		if !validCapabilities[capability] {
			return &PluginError{
				Code:      ErrCodeInvalidManifest,
				Message:   fmt.Sprintf("invalid capability: %s", capability),
				Timestamp: time.Now(),
			}
		}
	}
	
	return nil
}
