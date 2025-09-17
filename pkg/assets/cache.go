package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/errors"
)

// FileCacheManager implements CacheManager using file system storage
type FileCacheManager struct {
	basePath    string
	sizeLimitMB int64
	defaultTTL  time.Duration
	logger      logging.Logger

	mu    sync.RWMutex
	index map[string]*cacheEntry
}

type cacheEntry struct {
	Key        string        `json:"key"`
	Path       string        `json:"path"`
	Size       int64         `json:"size"`
	Checksum   string        `json:"checksum"`
	CreatedAt  time.Time     `json:"created_at"`
	AccessedAt time.Time     `json:"accessed_at"`
	TTL        int64         `json:"ttl"`
	Metadata   AssetMetadata `json:"metadata"`
}

type cacheIndex struct {
	Entries map[string]*cacheEntry `json:"entries"`
	Stats   cacheStats             `json:"stats"`
}

type cacheStats struct {
	TotalSize   int64     `json:"total_size"`
	EntryCount  int       `json:"entry_count"`
	LastCleanup time.Time `json:"last_cleanup"`
	HitCount    int64     `json:"hit_count"`
	MissCount   int64     `json:"miss_count"`
}

// NewFileCacheManager creates a new file-based cache manager
func NewFileCacheManager(basePath string, sizeLimitMB int64, defaultTTL time.Duration, logger logging.Logger) *FileCacheManager {
	manager := &FileCacheManager{
		basePath:    basePath,
		sizeLimitMB: sizeLimitMB,
		defaultTTL:  defaultTTL,
		logger:      logger,
		index:       make(map[string]*cacheEntry),
	}

	// Load existing index
	if err := manager.loadIndex(); err != nil {
		logger.Warn("failed to load cache index, starting fresh", "error", err)
	}

	return manager
}

// Get retrieves an asset from cache
func (c *FileCacheManager) Get(ctx context.Context, key string) (*AssetContent, error) {
	c.mu.RLock()
	entry, exists := c.index[key]
	c.mu.RUnlock()

	if !exists {
		return nil, &AssetClientError{
			Code:    ErrorCodeCacheError,
			Message: fmt.Sprintf("asset '%s' not found in cache", key),
		}
	}

	// Check TTL
	if c.isExpired(entry) {
		c.logger.Debug("cache entry expired", "key", key)
		c.Delete(ctx, key) // Clean up expired entry
		return nil, &AssetClientError{
			Code:    ErrorCodeCacheError,
			Message: fmt.Sprintf("cached asset '%s' has expired", key),
		}
	}

	// Read content from file
	content, err := os.ReadFile(entry.Path)
	if err != nil {
		c.logger.Warn("failed to read cached file", "key", key, "path", entry.Path, "error", err)
		c.Delete(ctx, key) // Clean up invalid entry
		return nil, &AssetClientError{
			Code:    ErrorCodeCacheError,
			Message: fmt.Sprintf("failed to read cached asset '%s'", key),
			Details: err.Error(),
		}
	}

	// Update access time
	c.mu.Lock()
	entry.AccessedAt = time.Now()
	c.mu.Unlock()

	// Calculate cache age
	cacheAge := time.Since(entry.CreatedAt).Seconds()

	result := &AssetContent{
		Metadata: entry.Metadata,
		Content:  string(content),
		Checksum: entry.Checksum,
		Cached:   true,
		CacheAge: int64(cacheAge),
	}

	c.logger.Debug("asset retrieved from cache", "key", key, "size", len(content))
	return result, nil
}

// Put stores an asset in cache
func (c *FileCacheManager) Put(ctx context.Context, key string, content *AssetContent) error {
	if key == "" {
		return &AssetClientError{
			Code:    ErrorCodeCacheError,
			Message: "cache key cannot be empty",
		}
	}

	// Ensure cache directory exists
	contentDir := filepath.Join(c.basePath, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create cache directory")
	}

	// Generate file path
	fileName := c.sanitizeFileName(key) + ".cache"
	filePath := filepath.Join(contentDir, fileName)

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content.Content), 0644); err != nil {
		return errors.Wrap(err, "failed to write cache file")
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to stat cache file")
	}

	// Create cache entry
	entry := &cacheEntry{
		Key:        key,
		Path:       filePath,
		Size:       fileInfo.Size(),
		Checksum:   content.Checksum,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		TTL:        int64(c.defaultTTL.Seconds()),
		Metadata:   content.Metadata,
	}

	// Check if we need to make space
	c.mu.Lock()
	if err := c.ensureSpaceForEntry(entry); err != nil {
		c.mu.Unlock()
		os.Remove(filePath) // Clean up the file we just created
		return errors.Wrap(err, "failed to ensure cache space")
	}

	// Add to index
	c.index[key] = entry
	c.mu.Unlock()

	// Save index
	if err := c.saveIndex(); err != nil {
		c.logger.Warn("failed to save cache index", "error", err)
	}

	c.logger.Debug("asset cached", "key", key, "size", entry.Size, "path", filePath)
	return nil
}

// Delete removes an asset from cache
func (c *FileCacheManager) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.index[key]
	if !exists {
		return nil // Already deleted
	}

	// Remove file
	if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
		c.logger.Warn("failed to remove cache file", "key", key, "path", entry.Path, "error", err)
	}

	// Remove from index
	delete(c.index, key)

	c.logger.Debug("asset removed from cache", "key", key)
	return nil
}

// Clear removes all cached assets
func (c *FileCacheManager) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all files
	for key, entry := range c.index {
		if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
			c.logger.Warn("failed to remove cache file during clear", "key", key, "error", err)
		}
	}

	// Clear index
	c.index = make(map[string]*cacheEntry)

	// Remove cache directory if empty
	if err := c.removeEmptyDirs(); err != nil {
		c.logger.Warn("failed to clean up cache directories", "error", err)
	}

	// Save empty index
	if err := c.saveIndex(); err != nil {
		c.logger.Warn("failed to save empty cache index", "error", err)
	}

	c.logger.Info("cache cleared")
	return nil
}

// GetInfo returns cache information
func (c *FileCacheManager) GetInfo(ctx context.Context) (*CacheInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalSize int64
	for _, entry := range c.index {
		totalSize += entry.Size
	}

	var lastSync time.Time
	for _, entry := range c.index {
		if entry.CreatedAt.After(lastSync) {
			lastSync = entry.CreatedAt
		}
	}

	return &CacheInfo{
		TotalSize:     totalSize,
		AssetCount:    len(c.index),
		LastSync:      lastSync,
		CacheHitRatio: 0, // This will be calculated by the client
	}, nil
}

// Cleanup performs cache maintenance
func (c *FileCacheManager) Cleanup(ctx context.Context) error {
	c.logger.Debug("starting cache cleanup")

	c.mu.Lock()
	defer c.mu.Unlock()

	var removed int
	var freedBytes int64

	// Remove expired entries
	for key, entry := range c.index {
		if c.isExpired(entry) {
			if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
				c.logger.Warn("failed to remove expired cache file", "key", key, "error", err)
			}
			freedBytes += entry.Size
			delete(c.index, key)
			removed++
		}
	}

	// Check size limit and evict if necessary
	totalSize := c.calculateTotalSize()
	sizeLimitBytes := c.sizeLimitMB * 1024 * 1024

	if totalSize > sizeLimitBytes {
		evicted := c.evictLRU(totalSize - sizeLimitBytes)
		removed += evicted
		freedBytes += totalSize - c.calculateTotalSize()
	}

	// Save updated index
	if err := c.saveIndex(); err != nil {
		c.logger.Warn("failed to save index after cleanup", "error", err)
	}

	c.logger.Debug("cache cleanup completed", "removed", removed, "freed_bytes", freedBytes)
	return nil
}

// Private helper methods

func (c *FileCacheManager) loadIndex() error {
	indexPath := filepath.Join(c.basePath, "metadata", "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing index, start fresh
		}
		return err
	}

	var index cacheIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	c.index = index.Entries
	if c.index == nil {
		c.index = make(map[string]*cacheEntry)
	}

	return nil
}

func (c *FileCacheManager) saveIndex() error {
	metadataDir := filepath.Join(c.basePath, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return err
	}

	indexPath := filepath.Join(metadataDir, "index.json")

	index := cacheIndex{
		Entries: c.index,
		Stats: cacheStats{
			TotalSize:   c.calculateTotalSize(),
			EntryCount:  len(c.index),
			LastCleanup: time.Now(),
		},
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}

func (c *FileCacheManager) isExpired(entry *cacheEntry) bool {
	if entry.TTL <= 0 {
		return false // No expiration
	}

	expirationTime := entry.CreatedAt.Add(time.Duration(entry.TTL) * time.Second)
	return time.Now().After(expirationTime)
}

func (c *FileCacheManager) ensureSpaceForEntry(newEntry *cacheEntry) error {
	totalSize := c.calculateTotalSize()
	sizeLimitBytes := c.sizeLimitMB * 1024 * 1024

	if totalSize+newEntry.Size <= sizeLimitBytes {
		return nil // Enough space
	}

	// Need to evict entries
	spaceNeeded := (totalSize + newEntry.Size) - sizeLimitBytes
	evicted := c.evictLRU(spaceNeeded)

	c.logger.Debug("evicted entries to make space", "evicted", evicted, "space_needed", spaceNeeded)
	return nil
}

func (c *FileCacheManager) evictLRU(spaceNeeded int64) int {
	// Sort entries by access time (oldest first)
	type entryWithKey struct {
		key   string
		entry *cacheEntry
	}

	var entries []entryWithKey
	for key, entry := range c.index {
		entries = append(entries, entryWithKey{key: key, entry: entry})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].entry.AccessedAt.Before(entries[j].entry.AccessedAt)
	})

	var evicted int
	var freedSpace int64

	for _, item := range entries {
		if freedSpace >= spaceNeeded {
			break
		}

		// Remove file
		if err := os.Remove(item.entry.Path); err != nil && !os.IsNotExist(err) {
			c.logger.Warn("failed to remove file during eviction", "key", item.key, "error", err)
		}

		freedSpace += item.entry.Size
		delete(c.index, item.key)
		evicted++
	}

	return evicted
}

func (c *FileCacheManager) calculateTotalSize() int64 {
	var total int64
	for _, entry := range c.index {
		total += entry.Size
	}
	return total
}

func (c *FileCacheManager) sanitizeFileName(name string) string {
	// Replace problematic characters with underscores
	sanitized := name
	problematicChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range problematicChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	return filepath.Base(sanitized) // Remove path separators
}

func (c *FileCacheManager) removeEmptyDirs() error {
	dirs := []string{
		filepath.Join(c.basePath, "content"),
		filepath.Join(c.basePath, "metadata"),
		c.basePath,
	}

	for _, dir := range dirs {
		if err := c.removeIfEmpty(dir); err != nil {
			return err
		}
	}

	return nil
}

func (c *FileCacheManager) removeIfEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		return os.Remove(dir)
	}

	return nil
}
