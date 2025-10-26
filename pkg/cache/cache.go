package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/go-viper/mapstructure/v2"
)

// Manager represents a generic cache interface
type Manager[T any] interface {
	// Get retrieves an item from cache
	Get(ctx context.Context, key string) (*Entry[T], error)

	// Put stores an item in cache
	Put(ctx context.Context, key string, data T, opts PutOptions) error

	// Delete removes an item from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all cached items
	Clear(ctx context.Context) error

	// GetInfo returns cache information
	GetInfo(ctx context.Context) (*Info, error)

	// Cleanup performs cache maintenance
	Cleanup(ctx context.Context) error

	// Close cleans up cache resources
	Close() error
}

// Entry represents a cached item with metadata
type Entry[T any] struct {
	Data     T      `json:"data"`
	Checksum string `json:"checksum"`
	Cached   bool   `json:"cached"`
	CacheAge int64  `json:"cache_age"` // Age in seconds
	Size     int64  `json:"size"`      // Size in bytes
}

// PutOptions represents options for storing items in cache
type PutOptions struct {
	TTL      time.Duration `json:"ttl,omitempty"`      // Time to live (0 = use default)
	Checksum string        `json:"checksum,omitempty"` // Content checksum for integrity
}

// Info represents cache status information
type Info struct {
	TotalSize     int64     `json:"total_size"`      // Total cache size in bytes
	EntryCount    int       `json:"entry_count"`     // Number of cached entries
	LastCleanup   time.Time `json:"last_cleanup"`    // Last cleanup time
	HitCount      int64     `json:"hit_count"`       // Cache hit count
	MissCount     int64     `json:"miss_count"`      // Cache miss count
	CacheHitRatio float64   `json:"cache_hit_ratio"` // Hit ratio (0.0 - 1.0)
}

// Config represents cache configuration
type Config struct {
	BasePath          string        `yaml:"base_path" json:"base_path"`
	SizeLimitMB       int64         `yaml:"size_limit_mb" json:"size_limit_mb"`
	DefaultTTL        time.Duration `yaml:"default_ttl" json:"default_ttl"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	EnableCompression bool          `yaml:"enable_compression" json:"enable_compression"`
}

// DefaultConfig returns default cache configuration
func DefaultConfig() Config {
	return Config{
		BasePath:          "~/.zen/cache",
		SizeLimitMB:       100,
		DefaultTTL:        24 * time.Hour,
		CleanupInterval:   1 * time.Hour,
		EnableCompression: false,
	}
}

// Implement config.Configurable interface

// Validate validates the cache configuration
func (c Config) Validate() error {
	if c.BasePath == "" {
		return fmt.Errorf("base_path is required")
	}
	if c.SizeLimitMB <= 0 {
		return fmt.Errorf("size_limit_mb must be positive")
	}
	if c.DefaultTTL <= 0 {
		return fmt.Errorf("default_ttl must be positive")
	}
	if c.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup_interval must be positive")
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
	var cfg Config

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
		return cfg, fmt.Errorf("failed to decode cache config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for cache
func (p ConfigParser) Section() string {
	return "cache"
}

// Error represents cache-specific errors
type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

// ErrorCode represents specific cache error codes
type ErrorCode string

const (
	ErrorCodeNotFound      ErrorCode = "not_found"
	ErrorCodeInvalidKey    ErrorCode = "invalid_key"
	ErrorCodeStorageFull   ErrorCode = "storage_full"
	ErrorCodeCorrupted     ErrorCode = "corrupted"
	ErrorCodePermission    ErrorCode = "permission"
	ErrorCodeSerialization ErrorCode = "serialization"
)

// NewManager creates a new cache manager with the specified configuration
func NewManager[T any](config Config, logger logging.Logger, serializer Serializer[T]) Manager[T] {
	return NewFileManager(config, logger, serializer)
}
