package types

import (
	"encoding/json"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// General error codes
	ErrorCodeUnknown          ErrorCode = "UNKNOWN"
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrorCodeTimeout          ErrorCode = "TIMEOUT"

	// Configuration error codes
	ErrorCodeInvalidConfig  ErrorCode = "INVALID_CONFIG"
	ErrorCodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"

	// Workspace error codes
	ErrorCodeWorkspaceNotInit ErrorCode = "WORKSPACE_NOT_INITIALIZED"
	ErrorCodeInvalidWorkspace ErrorCode = "INVALID_WORKSPACE"

	// Asset error codes
	ErrorCodeAssetNotFound        ErrorCode = "ASSET_NOT_FOUND"
	ErrorCodeAuthenticationFailed ErrorCode = "AUTHENTICATION_FAILED"
	ErrorCodeNetworkError         ErrorCode = "NETWORK_ERROR"
	ErrorCodeCacheError           ErrorCode = "CACHE_ERROR"
	ErrorCodeIntegrityError       ErrorCode = "INTEGRITY_ERROR"
	ErrorCodeRateLimited          ErrorCode = "RATE_LIMITED"
	ErrorCodeRepositoryError      ErrorCode = "REPOSITORY_ERROR"
)

// Error represents a standardized error response
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error implements the error interface
func (e Error) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Metadata represents common metadata fields
type Metadata struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Status represents the status of an entity
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Result represents a generic result with status and data
type Result struct {
	Status  Status          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Priority represents task or item priority
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// OutputFormat represents supported output formats
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)
