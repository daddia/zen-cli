package assets

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/errors"
	"github.com/daddia/zen/pkg/git"
)

// Client implements AssetClientInterface
type Client struct {
	config AssetConfig
	logger logging.Logger
	auth   AuthProvider
	cache  CacheManager
	git    git.Repository
	parser ManifestParser

	// Internal state
	mu           sync.RWMutex
	lastSync     time.Time
	manifestData []AssetMetadata

	// Performance metrics
	metrics struct {
		cacheHits   int64
		cacheMisses int64
		syncCount   int64
		errorCount  int64
	}
}

// NewClient creates a new asset client
func NewClient(config AssetConfig, logger logging.Logger, auth AuthProvider, cache CacheManager, gitRepo git.Repository, parser ManifestParser) *Client {
	return &Client{
		config: config,
		logger: logger,
		auth:   auth,
		cache:  cache,
		git:    gitRepo,
		parser: parser,
	}
}

// ListAssets retrieves assets matching the provided filter
func (c *Client) ListAssets(ctx context.Context, filter AssetFilter) (*AssetList, error) {
	c.logger.Debug("listing assets", "filter", filter)

	// Ensure we have current manifest data
	if err := c.ensureManifestLoaded(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to load manifest")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Apply filtering
	filtered := c.filterAssets(c.manifestData, filter)

	// Apply pagination
	total := len(filtered)
	start := filter.Offset
	if start > total {
		start = total
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	end := start + limit
	if end > total {
		end = total
	}

	result := &AssetList{
		Assets:  filtered[start:end],
		Total:   total,
		HasMore: end < total,
	}

	c.logger.Debug("assets listed", "total", total, "returned", len(result.Assets))
	return result, nil
}

// GetAsset retrieves a specific asset by name
func (c *Client) GetAsset(ctx context.Context, name string, opts GetAssetOptions) (*AssetContent, error) {
	c.logger.Debug("getting asset", "name", name, "options", opts)

	if name == "" {
		return nil, &AssetClientError{
			Code:    ErrorCodeAssetNotFound,
			Message: "asset name cannot be empty",
		}
	}

	// Try cache first if enabled
	if opts.UseCache {
		if content, err := c.cache.Get(ctx, name); err == nil {
			c.mu.Lock()
			c.metrics.cacheHits++
			c.mu.Unlock()

			if opts.VerifyIntegrity {
				if err := c.verifyIntegrity(content); err != nil {
					c.logger.Warn("cache integrity check failed", "asset", name, "error", err)
					// Continue to load from repository
				} else {
					c.logger.Debug("asset served from cache", "name", name)
					return content, nil
				}
			} else {
				c.logger.Debug("asset served from cache", "name", name)
				return content, nil
			}
		}

		c.mu.Lock()
		c.metrics.cacheMisses++
		c.mu.Unlock()
	}

	// Load from repository
	content, err := c.loadAssetFromRepository(ctx, name, opts)
	if err != nil {
		c.mu.Lock()
		c.metrics.errorCount++
		c.mu.Unlock()
		return nil, err
	}

	// Cache the result
	if opts.UseCache {
		if err := c.cache.Put(ctx, name, content); err != nil {
			c.logger.Warn("failed to cache asset", "name", name, "error", err)
		}
	}

	c.logger.Debug("asset loaded from repository", "name", name)
	return content, nil
}

// SyncRepository synchronizes with the remote repository
func (c *Client) SyncRepository(ctx context.Context, req SyncRequest) (*SyncResult, error) {
	c.logger.Info("starting repository sync", "request", req)

	startTime := time.Now()
	result := &SyncResult{
		Status: "success",
	}

	// Try to authenticate (optional for public repositories)
	if err := c.auth.Authenticate(ctx, c.config.AuthProvider); err != nil {
		// Check if this is just a missing token (could be public repo)
		if assetErr, ok := err.(*AssetClientError); ok && assetErr.Code == ErrorCodeAuthenticationFailed {
			c.logger.Warn("no authentication token found, attempting anonymous access", "provider", c.config.AuthProvider)
			// Continue without authentication - will work for public repositories
		} else {
			// Other authentication errors should fail
			c.mu.Lock()
			c.metrics.errorCount++
			c.mu.Unlock()

			result.Status = "error"
			result.Error = fmt.Sprintf("authentication failed: %v", err)
			return result, &AssetClientError{
				Code:    ErrorCodeAuthenticationFailed,
				Message: "failed to authenticate with Git provider",
				Details: err.Error(),
			}
		}
	}

	// Set up timeout context
	syncCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.SyncTimeoutSeconds)*time.Second)
	defer cancel()

	// Load and parse manifest
	manifestContent, err := c.git.GetFile(syncCtx, "manifest.yaml")
	if err != nil {
		result.Status = "partial"
		result.Error = fmt.Sprintf("failed to load manifest: %v", err)
	} else {
		// Parse manifest
		newManifest, err := c.parser.Parse(syncCtx, manifestContent)
		if err != nil {
			result.Status = "partial"
			result.Error = fmt.Sprintf("failed to parse manifest: %v", err)
		} else {
			// Save manifest to .zen/assets/manifest.yaml
			if err := c.saveManifestToDisk(manifestContent); err != nil {
				c.logger.Warn("failed to save manifest to disk", "error", err)
				// Don't fail the sync, just warn
			}

			// Update manifest data and calculate changes
			c.mu.Lock()
			oldCount := len(c.manifestData)
			c.manifestData = newManifest
			c.lastSync = time.Now()
			c.metrics.syncCount++
			c.mu.Unlock()

			// Calculate changes (simplified)
			newCount := len(newManifest)
			switch {
			case newCount > oldCount:
				result.AssetsAdded = newCount - oldCount
			case newCount < oldCount:
				result.AssetsRemoved = oldCount - newCount
			default:
				result.AssetsUpdated = newCount // Assume all updated if same count
			}
		}
	}

	// Get cache info
	if cacheInfo, err := c.cache.GetInfo(ctx); err == nil {
		result.CacheSizeMB = float64(cacheInfo.TotalSize) / (1024 * 1024)
	}

	result.DurationMS = time.Since(startTime).Milliseconds()
	result.LastSync = c.lastSync

	c.logger.Info("repository sync completed", "result", result)
	return result, nil
}

// GetCacheInfo returns current cache status
func (c *Client) GetCacheInfo(ctx context.Context) (*CacheInfo, error) {
	c.logger.Debug("getting cache info")

	info, err := c.cache.GetInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache info")
	}

	// Calculate hit ratio
	c.mu.RLock()
	totalRequests := c.metrics.cacheHits + c.metrics.cacheMisses
	if totalRequests > 0 {
		info.CacheHitRatio = float64(c.metrics.cacheHits) / float64(totalRequests)
	}
	c.mu.RUnlock()

	return info, nil
}

// ClearCache removes all cached assets
func (c *Client) ClearCache(ctx context.Context) error {
	c.logger.Info("clearing asset cache")

	if err := c.cache.Clear(ctx); err != nil {
		return errors.Wrap(err, "failed to clear cache")
	}

	c.logger.Info("asset cache cleared")
	return nil
}

// Close cleans up client resources
func (c *Client) Close() error {
	c.logger.Debug("closing asset client")

	// Perform any cleanup operations
	c.mu.Lock()
	c.manifestData = nil
	c.mu.Unlock()

	c.logger.Debug("asset client closed")
	return nil
}

// Private helper methods

func (c *Client) ensureManifestLoaded(ctx context.Context) error {
	c.mu.RLock()
	hasManifest := len(c.manifestData) > 0
	c.mu.RUnlock()

	if !hasManifest {
		// Try to load manifest from cache/repository
		manifestContent, err := c.git.GetFile(ctx, "manifest.yaml")
		if err != nil {
			return errors.Wrap(err, "failed to load manifest file")
		}

		manifest, err := c.parser.Parse(ctx, manifestContent)
		if err != nil {
			return errors.Wrap(err, "failed to parse manifest")
		}

		c.mu.Lock()
		c.manifestData = manifest
		c.mu.Unlock()
	}

	return nil
}

func (c *Client) filterAssets(assets []AssetMetadata, filter AssetFilter) []AssetMetadata {
	var filtered []AssetMetadata

	for _, asset := range assets {
		// Type filter
		if filter.Type != "" && asset.Type != filter.Type {
			continue
		}

		// Category filter
		if filter.Category != "" && asset.Category != filter.Category {
			continue
		}

		// Tags filter (AND operation)
		if len(filter.Tags) > 0 {
			hasAllTags := true
			for _, filterTag := range filter.Tags {
				found := false
				for _, assetTag := range asset.Tags {
					if strings.EqualFold(assetTag, filterTag) {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		filtered = append(filtered, asset)
	}

	return filtered
}

func (c *Client) loadAssetFromRepository(ctx context.Context, name string, opts GetAssetOptions) (*AssetContent, error) {
	// Find asset metadata
	c.mu.RLock()
	var metadata *AssetMetadata
	for i := range c.manifestData {
		if c.manifestData[i].Name == name {
			metadata = &c.manifestData[i]
			break
		}
	}
	c.mu.RUnlock()

	if metadata == nil {
		return nil, &AssetClientError{
			Code:    ErrorCodeAssetNotFound,
			Message: fmt.Sprintf("asset '%s' not found", name),
		}
	}

	// Load content from repository
	content, err := c.git.GetFile(ctx, metadata.Path)
	if err != nil {
		return nil, &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: fmt.Sprintf("failed to load asset content: %v", err),
		}
	}

	// Calculate checksum
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256(content))

	// Verify checksum if requested
	if opts.VerifyIntegrity && metadata.Checksum != "" {
		if checksum != metadata.Checksum {
			return nil, &AssetClientError{
				Code:    ErrorCodeIntegrityError,
				Message: fmt.Sprintf("asset integrity check failed for '%s'", name),
				Details: map[string]string{
					"expected": metadata.Checksum,
					"actual":   checksum,
				},
			}
		}
	}

	result := &AssetContent{
		Content:  string(content),
		Checksum: checksum,
		Cached:   false,
		CacheAge: 0,
	}

	if opts.IncludeMetadata {
		result.Metadata = *metadata
	}

	return result, nil
}

func (c *Client) verifyIntegrity(content *AssetContent) error {
	if content.Checksum == "" {
		return nil // No checksum to verify
	}

	calculated := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(content.Content)))
	if calculated != content.Checksum {
		return &AssetClientError{
			Code:    ErrorCodeIntegrityError,
			Message: "asset integrity verification failed",
		}
	}

	return nil
}

// saveManifestToDisk saves the manifest content to .zen/assets/manifest.yaml
func (c *Client) saveManifestToDisk(manifestContent []byte) error {
	// Ensure the .zen/assets directory exists
	assetsDir := c.config.CachePath
	if strings.HasPrefix(assetsDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "failed to get user home directory")
		}
		assetsDir = filepath.Join(home, assetsDir[2:])
	}

	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create assets directory")
	}

	// Write manifest to file
	manifestPath := filepath.Join(assetsDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, manifestContent, 0644); err != nil {
		return errors.Wrap(err, "failed to write manifest file")
	}

	c.logger.Debug("manifest saved to disk", "path", manifestPath)
	return nil
}

// sanitizeURL removes sensitive information from URLs for logging

// GetMetrics returns client performance metrics (for monitoring/debugging)
func (c *Client) GetMetrics() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"cache_hits":   c.metrics.cacheHits,
		"cache_misses": c.metrics.cacheMisses,
		"sync_count":   c.metrics.syncCount,
		"error_count":  c.metrics.errorCount,
		"last_sync":    c.lastSync,
	}
}
