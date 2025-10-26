package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/go-viper/mapstructure/v2"
)

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeTemplate AssetType = "template"
	AssetTypePrompt   AssetType = "prompt"
	AssetTypeMCP      AssetType = "mcp"
	AssetTypeSchema   AssetType = "schema"
)

// AssetMetadata contains metadata about an asset/activity
type AssetMetadata struct {
	Name           string     `yaml:"name" json:"name"`
	Type           AssetType  `yaml:"type" json:"type"`
	Description    string     `yaml:"description" json:"description"`
	Format         string     `yaml:"format" json:"format"`
	Category       string     `yaml:"category" json:"category"`
	Tags           []string   `yaml:"tags" json:"tags"`
	Variables      []Variable `yaml:"variables" json:"variables"`
	Checksum       string     `yaml:"checksum" json:"checksum"`
	Path           string     `yaml:"path" json:"path"`
	Command        string     `yaml:"command" json:"command"`                 // CLI command for the activity
	OutputFile     string     `yaml:"output_file" json:"output_file"`         // Primary output file
	WorkflowStages []string   `yaml:"workflow_stages" json:"workflow_stages"` // Zenflow stages this activity belongs to
	UpdatedAt      time.Time  `yaml:"updated_at" json:"updated_at"`
}

// Variable represents a template variable
type Variable struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Required    bool   `yaml:"required" json:"required"`
	Type        string `yaml:"type" json:"type"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
}

// AssetContent represents asset content with metadata
type AssetContent struct {
	Metadata AssetMetadata `json:"metadata"`
	Content  string        `json:"content"`
	Checksum string        `json:"checksum"`
	Cached   bool          `json:"cached"`
	CacheAge int64         `json:"cache_age"`
}

// AssetFilter represents filtering options for asset queries
type AssetFilter struct {
	Type     AssetType `json:"type,omitempty"`
	Category string    `json:"category,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Limit    int       `json:"limit,omitempty"`
	Offset   int       `json:"offset,omitempty"`
}

// AssetList represents a paginated list of assets
type AssetList struct {
	Assets  []AssetMetadata `json:"assets"`
	Total   int             `json:"total"`
	HasMore bool            `json:"has_more"`
}

// SyncRequest represents a repository synchronization request
type SyncRequest struct {
	Force   bool   `json:"force"`
	Shallow bool   `json:"shallow"`
	Branch  string `json:"branch"`
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	Status        string    `json:"status"`
	DurationMS    int64     `json:"duration_ms"`
	AssetsUpdated int       `json:"assets_updated"`
	AssetsAdded   int       `json:"assets_added"`
	AssetsRemoved int       `json:"assets_removed"`
	CacheSizeMB   float64   `json:"cache_size_mb"`
	LastSync      time.Time `json:"last_sync"`
	Error         string    `json:"error,omitempty"`
}

// CacheInfo represents cache status information
type CacheInfo struct {
	TotalSize     int64     `json:"total_size"`
	AssetCount    int       `json:"asset_count"`
	LastSync      time.Time `json:"last_sync"`
	CacheHitRatio float64   `json:"cache_hit_ratio"`
}

// GetAssetOptions represents options for asset retrieval
type GetAssetOptions struct {
	IncludeMetadata bool `json:"include_metadata"`
	VerifyIntegrity bool `json:"verify_integrity"`
	UseCache        bool `json:"use_cache"`
}

// AssetClientError represents asset client specific errors
type AssetClientError struct {
	Code       AssetErrorCode `json:"code"`
	Message    string         `json:"message"`
	Details    interface{}    `json:"details,omitempty"`
	RetryAfter int            `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e AssetClientError) Error() string {
	return e.Message
}

// AssetErrorCode represents specific asset client error codes
type AssetErrorCode string

const (
	ErrorCodeAssetNotFound        AssetErrorCode = "asset_not_found"
	ErrorCodeAuthenticationFailed AssetErrorCode = "authentication_failed"
	ErrorCodeNetworkError         AssetErrorCode = "network_error"
	ErrorCodeCacheError           AssetErrorCode = "cache_error"
	ErrorCodeIntegrityError       AssetErrorCode = "integrity_error"
	ErrorCodeRateLimited          AssetErrorCode = "rate_limited"
	ErrorCodeRepositoryError      AssetErrorCode = "repository_error"
	ErrorCodeConfigurationError   AssetErrorCode = "configuration_error"
)

// AssetClientInterface defines the interface for asset operations
type AssetClientInterface interface {
	// ListAssets retrieves assets matching the provided filter
	ListAssets(ctx context.Context, filter AssetFilter) (*AssetList, error)

	// GetAsset retrieves a specific asset by name
	GetAsset(ctx context.Context, name string, opts GetAssetOptions) (*AssetContent, error)

	// SyncRepository synchronizes with the remote repository
	SyncRepository(ctx context.Context, req SyncRequest) (*SyncResult, error)

	// GetCacheInfo returns current cache status
	GetCacheInfo(ctx context.Context) (*CacheInfo, error)

	// ClearCache removes all cached assets
	ClearCache(ctx context.Context) error

	// Close cleans up client resources
	Close() error
}

// AuthProvider represents authentication provider interface
type AuthProvider interface {
	// Authenticate authenticates with the Git provider
	Authenticate(ctx context.Context, provider string) error

	// GetCredentials returns credentials for the specified provider
	GetCredentials(provider string) (string, error)

	// ValidateCredentials validates stored credentials
	ValidateCredentials(ctx context.Context, provider string) error

	// RefreshCredentials refreshes expired credentials if possible
	RefreshCredentials(ctx context.Context, provider string) error
}

// CacheManager represents cache management interface
type CacheManager interface {
	// Get retrieves an asset from cache
	Get(ctx context.Context, key string) (*AssetContent, error)

	// Put stores an asset in cache
	Put(ctx context.Context, key string, content *AssetContent) error

	// Delete removes an asset from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all cached assets
	Clear(ctx context.Context) error

	// GetInfo returns cache information
	GetInfo(ctx context.Context) (*CacheInfo, error)

	// Cleanup performs cache maintenance
	Cleanup(ctx context.Context) error
}

// ManifestParser represents manifest parsing interface
type ManifestParser interface {
	// Parse parses the manifest file and returns asset metadata
	Parse(ctx context.Context, content []byte) ([]AssetMetadata, error)

	// Validate validates the manifest structure
	Validate(ctx context.Context, content []byte) error
}

// Config represents asset client configuration
type Config struct {
	// Repository configuration
	RepositoryURL string `yaml:"repository_url" json:"repository_url" mapstructure:"repository_url"`
	Branch        string `yaml:"branch" json:"branch" mapstructure:"branch"`

	// Cache configuration
	CachePath   string        `yaml:"cache_path" json:"cache_path" mapstructure:"cache_path"`
	CacheSizeMB int64         `yaml:"cache_size_mb" json:"cache_size_mb" mapstructure:"cache_size_mb"`
	DefaultTTL  time.Duration `yaml:"default_ttl" json:"default_ttl" mapstructure:"default_ttl"`

	// Authentication configuration
	AuthProvider string `yaml:"auth_provider" json:"auth_provider" mapstructure:"auth_provider"`

	// Performance configuration
	SyncTimeoutSeconds int `yaml:"sync_timeout_seconds" json:"sync_timeout_seconds" mapstructure:"sync_timeout_seconds"`
	MaxConcurrentOps   int `yaml:"max_concurrent_ops" json:"max_concurrent_ops" mapstructure:"max_concurrent_ops"`

	// Feature flags
	IntegrityChecksEnabled bool `yaml:"integrity_checks_enabled" json:"integrity_checks_enabled" mapstructure:"integrity_checks_enabled"`
	PrefetchEnabled        bool `yaml:"prefetch_enabled" json:"prefetch_enabled" mapstructure:"prefetch_enabled"`
}

// DefaultConfig returns default asset client configuration
func DefaultConfig() Config {
	return Config{
		RepositoryURL:          "https://github.com/daddia/zen-assets.git", // Default official repository
		Branch:                 "main",
		CachePath:              "~/.zen/library",
		CacheSizeMB:            100,
		DefaultTTL:             24 * time.Hour,
		AuthProvider:           "github",
		SyncTimeoutSeconds:     30,
		MaxConcurrentOps:       3,
		IntegrityChecksEnabled: true,
		PrefetchEnabled:        true,
	}
}

// Implement config.Configurable interface

// Validate validates the asset configuration
func (c Config) Validate() error {
	if c.RepositoryURL == "" {
		return fmt.Errorf("repository_url is required")
	}
	if c.Branch == "" {
		return fmt.Errorf("branch is required")
	}
	if c.CachePath == "" {
		return fmt.Errorf("cache_path is required")
	}
	if c.CacheSizeMB <= 0 {
		return fmt.Errorf("cache_size_mb must be positive")
	}
	if c.SyncTimeoutSeconds <= 0 {
		return fmt.Errorf("sync_timeout_seconds must be positive")
	}
	if c.MaxConcurrentOps <= 0 {
		return fmt.Errorf("max_concurrent_ops must be positive")
	}
	return nil
}

// Defaults returns a new Config with default values
func (c Config) Defaults() config.Configurable {
	return DefaultConfig()
}

// ConfigParser implements config.ConfigParser[Config] interface
type ConfigParser struct{}

// Parse converts raw configuration data to Config
func (p ConfigParser) Parse(raw map[string]interface{}) (Config, error) {
	// Start with defaults to ensure all fields are properly initialized
	cfg := DefaultConfig()

	// If raw data is empty, return defaults
	if len(raw) == 0 {
		return cfg, nil
	}

	// Use mapstructure to decode the raw map into our config struct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return cfg, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return cfg, fmt.Errorf("failed to decode asset config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for assets
func (p ConfigParser) Section() string {
	return "assets"
}
