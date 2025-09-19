package assets

import (
	"context"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/cache"
)

// AssetCacheManager implements CacheManager using the generic cache package
type AssetCacheManager struct {
	cache cache.Manager[AssetContent]
}

// AssetContentSerializer implements cache.Serializer for AssetContent
type AssetContentSerializer struct{}

// NewAssetContentSerializer creates a new AssetContent serializer
func NewAssetContentSerializer() *AssetContentSerializer {
	return &AssetContentSerializer{}
}

// Serialize converts AssetContent to bytes (using the Content field directly)
func (s *AssetContentSerializer) Serialize(data AssetContent) ([]byte, error) {
	// For assets, we store the raw content directly, not JSON
	// Metadata is stored separately in the cache index
	return []byte(data.Content), nil
}

// Deserialize converts bytes back to AssetContent
func (s *AssetContentSerializer) Deserialize(data []byte) (AssetContent, error) {
	// Return basic AssetContent with the raw content
	// Metadata will be populated by the cache adapter
	return AssetContent{
		Content: string(data),
	}, nil
}

// ContentType returns the content type for asset content
func (s *AssetContentSerializer) ContentType() string {
	return "text/plain" // Assets are typically text-based
}

// NewAssetCacheManager creates a new asset cache manager using the generic cache
// The cache is session-based with a default TTL matching CLI session duration
func NewAssetCacheManager(basePath string, sizeLimitMB int64, defaultTTL time.Duration, logger logging.Logger) *AssetCacheManager {
	// Use session-based TTL if not specified
	if defaultTTL == 0 {
		// Default to 1 hour for CLI session cache
		// This ensures cached assets are available for the duration of typical workflows
		defaultTTL = 1 * time.Hour
	}

	config := cache.Config{
		BasePath:    basePath,
		SizeLimitMB: sizeLimitMB,
		DefaultTTL:  defaultTTL,
	}

	serializer := NewAssetContentSerializer()
	genericCache := cache.NewManager(config, logger, serializer)

	return &AssetCacheManager{
		cache: genericCache,
	}
}

// Get retrieves an asset from cache
func (a *AssetCacheManager) Get(ctx context.Context, key string) (*AssetContent, error) {
	entry, err := a.cache.Get(ctx, key)
	if err != nil {
		// Convert generic cache error to asset error
		if cacheErr, ok := err.(*cache.Error); ok {
			return nil, &AssetClientError{
				Code:    convertCacheErrorCode(cacheErr.Code),
				Message: cacheErr.Message,
				Details: cacheErr.Details,
			}
		}
		return nil, err
	}

	// Populate the full AssetContent structure
	result := &entry.Data
	result.Cached = entry.Cached
	result.CacheAge = entry.CacheAge

	// If checksum wasn't set during serialization, use the cache checksum
	if result.Checksum == "" {
		result.Checksum = entry.Checksum
	}

	return result, nil
}

// Put stores an asset in cache
func (a *AssetCacheManager) Put(ctx context.Context, key string, content *AssetContent) error {
	if content == nil {
		return &AssetClientError{
			Code:    ErrorCodeCacheError,
			Message: "content cannot be nil",
		}
	}

	opts := cache.PutOptions{
		Checksum: content.Checksum,
	}

	err := a.cache.Put(ctx, key, *content, opts)
	if err != nil {
		// Convert generic cache error to asset error
		if cacheErr, ok := err.(*cache.Error); ok {
			return &AssetClientError{
				Code:    convertCacheErrorCode(cacheErr.Code),
				Message: cacheErr.Message,
				Details: cacheErr.Details,
			}
		}
		return err
	}

	return nil
}

// Delete removes an asset from cache
func (a *AssetCacheManager) Delete(ctx context.Context, key string) error {
	return a.cache.Delete(ctx, key)
}

// Clear removes all cached assets
func (a *AssetCacheManager) Clear(ctx context.Context) error {
	return a.cache.Clear(ctx)
}

// GetInfo returns cache information
func (a *AssetCacheManager) GetInfo(ctx context.Context) (*CacheInfo, error) {
	info, err := a.cache.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Convert generic cache info to asset cache info
	return &CacheInfo{
		TotalSize:     info.TotalSize,
		AssetCount:    info.EntryCount,
		LastSync:      time.Time{}, // This will be set by the asset client
		CacheHitRatio: info.CacheHitRatio,
	}, nil
}

// Cleanup performs cache maintenance
func (a *AssetCacheManager) Cleanup(ctx context.Context) error {
	return a.cache.Cleanup(ctx)
}

// Close cleans up cache resources
func (a *AssetCacheManager) Close() error {
	return a.cache.Close()
}

// convertCacheErrorCode converts generic cache error codes to asset error codes
func convertCacheErrorCode(code cache.ErrorCode) AssetErrorCode {
	switch code {
	case cache.ErrorCodeNotFound:
		return ErrorCodeAssetNotFound
	case cache.ErrorCodeInvalidKey:
		return ErrorCodeCacheError
	case cache.ErrorCodeStorageFull:
		return ErrorCodeCacheError
	case cache.ErrorCodeCorrupted:
		return ErrorCodeIntegrityError
	case cache.ErrorCodePermission:
		return ErrorCodeCacheError
	case cache.ErrorCodeSerialization:
		return ErrorCodeCacheError
	default:
		return ErrorCodeCacheError
	}
}
