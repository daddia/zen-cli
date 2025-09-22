package jira

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
)

// Plugin implements the standardized integration plugin interface for Jira
type Plugin struct {
	config     *PluginConfig
	logger     logging.Logger
	httpClient *http.Client
	authMgr    auth.Manager
}

// PluginConfig contains Jira plugin configuration
type PluginConfig struct {
	Name       string                 `json:"name" yaml:"name" validate:"required"`
	Version    string                 `json:"version" yaml:"version" validate:"required,semver"`
	Enabled    bool                   `json:"enabled" yaml:"enabled"`
	BaseURL    string                 `json:"base_url" yaml:"base_url" validate:"required,url"`
	ProjectKey string                 `json:"project_key" yaml:"project_key" validate:"required"`
	Timeout    time.Duration          `json:"timeout" yaml:"timeout"`
	MaxRetries int                    `json:"max_retries" yaml:"max_retries"`
	Headers    map[string]string      `json:"headers" yaml:"headers"`
	Auth       *AuthConfig            `json:"auth" yaml:"auth" validate:"required"`
	RateLimit  *RateLimitConfig       `json:"rate_limit" yaml:"rate_limit"`
	Cache      *CacheConfig           `json:"cache" yaml:"cache"`
	Settings   map[string]interface{} `json:"settings" yaml:"settings"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Type           AuthType      `json:"type" yaml:"type" validate:"required"`
	CredentialsRef string        `json:"credentials_ref" yaml:"credentials_ref"`
	TokenStorage   string        `json:"token_storage" yaml:"token_storage"`
	RefreshToken   bool          `json:"refresh_token" yaml:"refresh_token"`
	TokenExpiry    time.Duration `json:"token_expiry" yaml:"token_expiry"`
}

type AuthType string

const (
	AuthTypeBasic  AuthType = "basic"
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeToken  AuthType = "token"
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

// PluginTaskData represents standardized task data for the plugin
type PluginTaskData struct {
	ID          string                 `json:"id" yaml:"id" validate:"required"`
	ExternalID  string                 `json:"external_id" yaml:"external_id"`
	Title       string                 `json:"title" yaml:"title" validate:"required"`
	Description string                 `json:"description" yaml:"description"`
	Status      string                 `json:"status" yaml:"status" validate:"required"`
	Priority    string                 `json:"priority" yaml:"priority"`
	Type        string                 `json:"type" yaml:"type"`
	Owner       string                 `json:"owner" yaml:"owner"`
	Assignee    string                 `json:"assignee" yaml:"assignee"`
	Team        string                 `json:"team" yaml:"team"`
	Created     time.Time              `json:"created" yaml:"created"`
	Updated     time.Time              `json:"updated" yaml:"updated"`
	DueDate     *time.Time             `json:"due_date,omitempty" yaml:"due_date,omitempty"`
	Labels      []string               `json:"labels" yaml:"labels"`
	Tags        []string               `json:"tags" yaml:"tags"`
	Components  []string               `json:"components" yaml:"components"`
	ExternalURL string                 `json:"external_url" yaml:"external_url"`
	RawData     map[string]interface{} `json:"raw_data,omitempty" yaml:"raw_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata" yaml:"metadata"`
	Version     int64                  `json:"version" yaml:"version"`
	Checksum    string                 `json:"checksum" yaml:"checksum"`
}

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
	JQL     string                 `json:"jql"`
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

// NewPlugin creates a new Jira plugin instance
func NewPlugin(config *PluginConfig, logger logging.Logger, authMgr auth.Manager) *Plugin {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Plugin{
		config:  config,
		logger:  logger,
		authMgr: authMgr,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Plugin Identity and Lifecycle Methods

// Name returns the plugin name
func (p *Plugin) Name() string {
	return p.config.Name
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return p.config.Version
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Jira integration plugin for task management"
}

// Initialize initializes the plugin with configuration
func (p *Plugin) Initialize(ctx context.Context, config *PluginConfig) error {
	p.config = config

	// Validate configuration
	if err := p.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Test connection
	if err := p.ValidateConnection(ctx); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}

	p.logger.Info("Jira plugin initialized successfully",
		"base_url", p.config.BaseURL,
		"project_key", p.config.ProjectKey)

	return nil
}

// Validate validates the plugin configuration and connectivity
func (p *Plugin) Validate(ctx context.Context) error {
	if err := p.validateConfig(); err != nil {
		return err
	}

	return p.ValidateConnection(ctx)
}

// HealthCheck performs a health check on the plugin
func (p *Plugin) HealthCheck(ctx context.Context) (*PluginHealth, error) {
	start := time.Now()

	health := &PluginHealth{
		Provider:    p.Name(),
		LastChecked: time.Now(),
	}

	// Test connection to Jira
	if err := p.ValidateConnection(ctx); err != nil {
		health.Healthy = false
		health.LastError = err.Error()
		health.ErrorCount = 1
	} else {
		health.Healthy = true
	}

	health.ResponseTime = time.Since(start)

	return health, nil
}

// Shutdown performs cleanup when the plugin is being shut down
func (p *Plugin) Shutdown(ctx context.Context) error {
	p.logger.Info("shutting down Jira plugin")

	// Close HTTP client connections
	if transport, ok := p.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	return nil
}

// ValidateConnection tests the connection to Jira
func (p *Plugin) ValidateConnection(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/rest/api/3/serverInfo", strings.TrimRight(p.config.BaseURL, "/"))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if err := p.addAuthentication(req); err != nil {
		return fmt.Errorf("failed to add authentication: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetRateLimitInfo returns current rate limit information
func (p *Plugin) GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error) {
	// Jira includes rate limit info in response headers
	// This would be populated from actual API responses
	return &RateLimitInfo{
		Limit:     300, // Default Jira Cloud limit
		Remaining: 300, // Would be updated from actual responses
		ResetTime: time.Now().Add(time.Minute),
	}, nil
}

// SupportsOperation returns true if the plugin supports the given operation
func (p *Plugin) SupportsOperation(operation OperationType) bool {
	supportedOps := map[OperationType]bool{
		OperationTypeFetch:  true,
		OperationTypeCreate: true,
		OperationTypeUpdate: true,
		OperationTypeDelete: true,
		OperationTypeSearch: true,
		OperationTypeSync:   true,
	}

	return supportedOps[operation]
}

// Helper methods

func (p *Plugin) validateConfig() error {
	if p.config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	if p.config.ProjectKey == "" {
		return fmt.Errorf("project_key is required")
	}

	if p.config.Auth == nil {
		return fmt.Errorf("auth configuration is required")
	}

	return nil
}

func (p *Plugin) addAuthentication(req *http.Request) error {
	if p.config.Auth == nil {
		return fmt.Errorf("no authentication configuration")
	}

	switch p.config.Auth.Type {
	case AuthTypeBasic:
		credentials, err := p.authMgr.GetCredentials(p.config.Auth.CredentialsRef)
		if err != nil {
			return fmt.Errorf("failed to get credentials: %w", err)
		}
		req.Header.Set("Authorization", "Basic "+credentials)

	case AuthTypeToken:
		token, err := p.authMgr.GetCredentials(p.config.Auth.CredentialsRef)
		if err != nil {
			return fmt.Errorf("failed to get token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

	case AuthTypeOAuth2:
		token, err := p.authMgr.GetCredentials(p.config.Auth.CredentialsRef)
		if err != nil {
			return fmt.Errorf("failed to get OAuth token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

	default:
		return fmt.Errorf("unsupported auth type: %s", p.config.Auth.Type)
	}

	return nil
}

func (p *Plugin) buildJiraURL(path string) string {
	baseURL := strings.TrimRight(p.config.BaseURL, "/")
	path = strings.TrimLeft(path, "/")
	return fmt.Sprintf("%s/%s", baseURL, path)
}

func (p *Plugin) handleHTTPError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("authentication failed: %s", resp.Status)
	case http.StatusNotFound:
		return fmt.Errorf("resource not found: %s", resp.Status)
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded: %s", resp.Status)
	case http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return fmt.Errorf("server error: %s", resp.Status)
	default:
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}
}
