package clients

import (
	"context"
	"time"
)

// Client represents a generic external service client
type Client interface {
	// Name returns the client name
	Name() string

	// ValidateConnection tests the connection to the external service
	ValidateConnection(ctx context.Context) error

	// Close cleans up client resources
	Close() error
}

// HTTPConfig contains common HTTP client configuration
type HTTPConfig struct {
	BaseURL    string            `json:"base_url" yaml:"base_url"`
	Timeout    time.Duration     `json:"timeout" yaml:"timeout"`
	Retries    int               `json:"retries" yaml:"retries"`
	UserAgent  string            `json:"user_agent" yaml:"user_agent"`
	Headers    map[string]string `json:"headers" yaml:"headers"`
	RateLimits RateLimitConfig   `json:"rate_limits" yaml:"rate_limits"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute" yaml:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour" yaml:"requests_per_hour"`
	BurstSize         int `json:"burst_size" yaml:"burst_size"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Type           string            `json:"type" yaml:"type"`                       // basic, oauth2, token, api_key
	CredentialsRef string            `json:"credentials_ref" yaml:"credentials_ref"` // Reference to auth system
	Headers        map[string]string `json:"headers" yaml:"headers"`                 // Additional auth headers
	TokenEndpoint  string            `json:"token_endpoint" yaml:"token_endpoint"`   // OAuth2 token endpoint
	Scopes         []string          `json:"scopes" yaml:"scopes"`                   // OAuth2 scopes
}

// ClientError represents a client-specific error
type ClientError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Retryable  bool                   `json:"retryable"`
}

func (e ClientError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrorCodeConnectionFailed     = "CONNECTION_FAILED"
	ErrorCodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	ErrorCodeRateLimited          = "RATE_LIMITED"
	ErrorCodeNotFound             = "NOT_FOUND"
	ErrorCodeInvalidRequest       = "INVALID_REQUEST"
	ErrorCodeInternalError        = "INTERNAL_ERROR"
	ErrorCodeTimeout              = "TIMEOUT"
	ErrorCodeUnknown              = "UNKNOWN"
)

// HealthStatus represents the health status of a client
type HealthStatus struct {
	Healthy     bool                   `json:"healthy"`
	LastChecked time.Time              `json:"last_checked"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ClientMetrics contains client performance metrics
type ClientMetrics struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastRequestTime time.Time     `json:"last_request_time"`
	LastErrorTime   time.Time     `json:"last_error_time,omitempty"`
}
