package assets

import (
	"context"
	"time"
)

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeTemplate AssetType = "template"
	AssetTypePrompt   AssetType = "prompt"
	AssetTypeMCP      AssetType = "mcp"
	AssetTypeSchema   AssetType = "schema"
)

// AssetMetadata contains metadata about an asset
type AssetMetadata struct {
	Name        string     `yaml:"name" json:"name"`
	Type        AssetType  `yaml:"type" json:"type"`
	Description string     `yaml:"description" json:"description"`
	Format      string     `yaml:"format" json:"format"`
	Category    string     `yaml:"category" json:"category"`
	Tags        []string   `yaml:"tags" json:"tags"`
	Variables   []Variable `yaml:"variables" json:"variables"`
	Checksum    string     `yaml:"checksum" json:"checksum"`
	Path        string     `yaml:"path" json:"path"`
	UpdatedAt   time.Time  `yaml:"updated_at" json:"updated_at"`
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

// GitRepository represents Git repository operations interface
type GitRepository interface {
	// Clone clones the repository to local cache
	Clone(ctx context.Context, url, branch string, shallow bool) error

	// Pull updates the local repository
	Pull(ctx context.Context) error

	// GetFile retrieves a file from the repository
	GetFile(ctx context.Context, path string) ([]byte, error)

	// ListFiles lists all files in the repository
	ListFiles(ctx context.Context, pattern string) ([]string, error)

	// GetLastCommit returns the last commit hash
	GetLastCommit(ctx context.Context) (string, error)

	// IsClean returns true if the repository has no uncommitted changes
	IsClean(ctx context.Context) (bool, error)
}

// ManifestParser represents manifest parsing interface
type ManifestParser interface {
	// Parse parses the manifest file and returns asset metadata
	Parse(ctx context.Context, content []byte) ([]AssetMetadata, error)

	// Validate validates the manifest structure
	Validate(ctx context.Context, content []byte) error
}

// AssetConfig represents asset client configuration
type AssetConfig struct {
	// Repository configuration
	RepositoryURL string `yaml:"repository_url" json:"repository_url"`
	Branch        string `yaml:"branch" json:"branch"`

	// Cache configuration
	CachePath   string        `yaml:"cache_path" json:"cache_path"`
	CacheSizeMB int64         `yaml:"cache_size_mb" json:"cache_size_mb"`
	DefaultTTL  time.Duration `yaml:"default_ttl" json:"default_ttl"`

	// Authentication configuration
	AuthProvider string `yaml:"auth_provider" json:"auth_provider"`

	// Performance configuration
	SyncTimeoutSeconds int `yaml:"sync_timeout_seconds" json:"sync_timeout_seconds"`
	MaxConcurrentOps   int `yaml:"max_concurrent_ops" json:"max_concurrent_ops"`

	// Feature flags
	IntegrityChecksEnabled bool `yaml:"integrity_checks_enabled" json:"integrity_checks_enabled"`
	PrefetchEnabled        bool `yaml:"prefetch_enabled" json:"prefetch_enabled"`
}

// DefaultAssetConfig returns default asset client configuration
func DefaultAssetConfig() AssetConfig {
	return AssetConfig{
		Branch:                 "main",
		CachePath:              "~/.zen/cache/assets",
		CacheSizeMB:            100,
		DefaultTTL:             24 * time.Hour,
		AuthProvider:           "github",
		SyncTimeoutSeconds:     30,
		MaxConcurrentOps:       3,
		IntegrityChecksEnabled: true,
		PrefetchEnabled:        true,
	}
}
