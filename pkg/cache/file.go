package cache

import (
	"context"
	"crypto/sha256"
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

// FileManager implements Manager using file system storage
type FileManager[T any] struct {
	config     Config
	logger     logging.Logger
	serializer Serializer[T]

	mu    sync.RWMutex
	index map[string]*fileEntry
	stats *cacheStats
}

type fileEntry struct {
	Key        string    `json:"key"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	Checksum   string    `json:"checksum"`
	CreatedAt  time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`
	TTL        int64     `json:"ttl"` // TTL in seconds (0 = no expiration)
}

type fileIndex struct {
	Entries map[string]*fileEntry `json:"entries"`
	Stats   *cacheStats           `json:"stats"`
}

type cacheStats struct {
	TotalSize   int64     `json:"total_size"`
	EntryCount  int       `json:"entry_count"`
	LastCleanup time.Time `json:"last_cleanup"`
	HitCount    int64     `json:"hit_count"`
	MissCount   int64     `json:"miss_count"`
}

// NewFileManager creates a new file-based cache manager
func NewFileManager[T any](config Config, logger logging.Logger, serializer Serializer[T]) *FileManager[T] {
	// Expand tilde in base path
	basePath := config.BasePath
	if strings.HasPrefix(basePath, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			basePath = filepath.Join(home, basePath[2:])
		}
	}

	manager := &FileManager[T]{
		config:     config,
		logger:     logger,
		serializer: serializer,
		index:      make(map[string]*fileEntry),
		stats: &cacheStats{
			LastCleanup: time.Now(),
		},
	}

	// Update config with expanded path
	manager.config.BasePath = basePath

	// Load existing index
	if err := manager.loadIndex(); err != nil {
		logger.Warn("failed to load cache index, starting fresh", "error", err)
	}

	return manager
}

// Get retrieves an item from cache
func (c *FileManager[T]) Get(ctx context.Context, key string) (*Entry[T], error) {
	c.mu.RLock()
	entry, exists := c.index[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.stats.MissCount++
		c.mu.Unlock()

		return nil, &Error{
			Code:    ErrorCodeNotFound,
			Message: fmt.Sprintf("key '%s' not found in cache", key),
		}
	}

	// Check TTL
	if c.isExpired(entry) {
		c.logger.Debug("cache entry expired", "key", key)
		c.Delete(ctx, key) // Clean up expired entry

		c.mu.Lock()
		c.stats.MissCount++
		c.mu.Unlock()

		return nil, &Error{
			Code:    ErrorCodeNotFound,
			Message: fmt.Sprintf("cached item '%s' has expired", key),
		}
	}

	// Read content from file
	data, err := os.ReadFile(entry.Path)
	if err != nil {
		c.logger.Warn("failed to read cached file", "key", key, "path", entry.Path, "error", err)
		c.Delete(ctx, key) // Clean up invalid entry

		c.mu.Lock()
		c.stats.MissCount++
		c.mu.Unlock()

		return nil, &Error{
			Code:    ErrorCodeCorrupted,
			Message: fmt.Sprintf("failed to read cached item '%s'", key),
			Details: err.Error(),
		}
	}

	// Deserialize data
	deserializedData, err := c.serializer.Deserialize(data)
	if err != nil {
		c.logger.Warn("failed to deserialize cached data", "key", key, "error", err)
		c.Delete(ctx, key) // Clean up corrupted entry

		c.mu.Lock()
		c.stats.MissCount++
		c.mu.Unlock()

		return nil, &Error{
			Code:    ErrorCodeSerialization,
			Message: fmt.Sprintf("failed to deserialize cached item '%s'", key),
			Details: err.Error(),
		}
	}

	// Update access time and hit count
	c.mu.Lock()
	entry.AccessedAt = time.Now()
	c.stats.HitCount++
	c.mu.Unlock()

	// Calculate cache age
	cacheAge := time.Since(entry.CreatedAt).Seconds()

	result := &Entry[T]{
		Data:     deserializedData,
		Checksum: entry.Checksum,
		Cached:   true,
		CacheAge: int64(cacheAge),
		Size:     entry.Size,
	}

	c.logger.Debug("item retrieved from cache", "key", key, "size", len(data))
	return result, nil
}

// Put stores an item in cache
func (c *FileManager[T]) Put(ctx context.Context, key string, data T, opts PutOptions) error {
	if key == "" {
		return &Error{
			Code:    ErrorCodeInvalidKey,
			Message: "cache key cannot be empty",
		}
	}

	// Serialize data
	serializedData, err := c.serializer.Serialize(data)
	if err != nil {
		return &Error{
			Code:    ErrorCodeSerialization,
			Message: fmt.Sprintf("failed to serialize data for key '%s'", key),
			Details: err.Error(),
		}
	}

	// Calculate checksum if not provided
	checksum := opts.Checksum
	if checksum == "" {
		checksum = fmt.Sprintf("sha256:%x", sha256.Sum256(serializedData))
	}

	// Ensure cache directory exists
	contentDir := filepath.Join(c.config.BasePath, "content")
	if err := os.MkdirAll(contentDir, 0750); err != nil {
		return &Error{
			Code:    ErrorCodePermission,
			Message: "failed to create cache directory",
			Details: err.Error(),
		}
	}

	// Generate file path
	fileName := c.sanitizeFileName(key) + ".cache"
	filePath := filepath.Join(contentDir, fileName)

	// Write content to file
	if err := os.WriteFile(filePath, serializedData, 0600); err != nil {
		return &Error{
			Code:    ErrorCodePermission,
			Message: fmt.Sprintf("failed to write cache file for key '%s'", key),
			Details: err.Error(),
		}
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to stat cache file")
	}

	// Determine TTL
	ttl := opts.TTL
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	// Create cache entry
	entry := &fileEntry{
		Key:        key,
		Path:       filePath,
		Size:       fileInfo.Size(),
		Checksum:   checksum,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		TTL:        int64(ttl.Seconds()),
	}

	// Check if we need to make space
	c.mu.Lock()
	if err := c.ensureSpaceForEntry(entry); err != nil {
		c.mu.Unlock()
		os.Remove(filePath) // Clean up the file we just created
		return &Error{
			Code:    ErrorCodeStorageFull,
			Message: fmt.Sprintf("failed to ensure cache space for key '%s'", key),
			Details: err.Error(),
		}
	}

	// Add to index
	c.index[key] = entry
	c.mu.Unlock()

	// Save index (outside of lock to avoid holding it too long)
	if err := c.saveIndex(); err != nil {
		c.logger.Warn("failed to save cache index", "error", err)
	}

	c.logger.Debug("item cached", "key", key, "size", entry.Size, "path", filePath)
	return nil
}

// Delete removes an item from cache
func (c *FileManager[T]) Delete(ctx context.Context, key string) error {
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

	c.logger.Debug("item removed from cache", "key", key)
	return nil
}

// Clear removes all cached items
func (c *FileManager[T]) Clear(ctx context.Context) error {
	// Get list of files to remove while holding lock
	c.mu.Lock()
	var filesToRemove []string
	for _, entry := range c.index {
		filesToRemove = append(filesToRemove, entry.Path)
	}

	// Clear index and reset stats
	c.index = make(map[string]*fileEntry)
	c.stats = &cacheStats{
		LastCleanup: time.Now(),
	}
	c.mu.Unlock()

	// Remove all files (outside lock)
	for _, filePath := range filesToRemove {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			c.logger.Warn("failed to remove cache file during clear", "path", filePath, "error", err)
		}
	}

	// Remove cache directory if empty (outside lock)
	if err := c.removeEmptyDirs(); err != nil {
		c.logger.Warn("failed to clean up cache directories", "error", err)
	}

	// Save empty index (outside lock to avoid deadlock)
	if err := c.saveIndex(); err != nil {
		c.logger.Warn("failed to save empty cache index", "error", err)
	}

	c.logger.Info("cache cleared")
	return nil
}

// GetInfo returns cache information
func (c *FileManager[T]) GetInfo(ctx context.Context) (*Info, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalSize int64
	for _, entry := range c.index {
		totalSize += entry.Size
	}

	// Calculate hit ratio
	var hitRatio float64
	totalRequests := c.stats.HitCount + c.stats.MissCount
	if totalRequests > 0 {
		hitRatio = float64(c.stats.HitCount) / float64(totalRequests)
	}

	return &Info{
		TotalSize:     totalSize,
		EntryCount:    len(c.index),
		LastCleanup:   c.stats.LastCleanup,
		HitCount:      c.stats.HitCount,
		MissCount:     c.stats.MissCount,
		CacheHitRatio: hitRatio,
	}, nil
}

// Cleanup performs cache maintenance
func (c *FileManager[T]) Cleanup(ctx context.Context) error {
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
	sizeLimitBytes := c.config.SizeLimitMB * 1024 * 1024

	if totalSize > sizeLimitBytes {
		evicted := c.evictLRU(totalSize - sizeLimitBytes)
		removed += evicted
		freedBytes += totalSize - c.calculateTotalSize()
	}

	// Update stats
	c.stats.LastCleanup = time.Now()

	// Note: Index will be saved on next Put operation or Close()
	// Avoiding saveIndex here to prevent potential deadlocks

	c.logger.Debug("cache cleanup completed", "removed", removed, "freed_bytes", freedBytes)
	return nil
}

// Close cleans up cache resources
func (c *FileManager[T]) Close() error {
	c.logger.Debug("closing cache manager")

	// Save final index
	if err := c.saveIndex(); err != nil {
		c.logger.Warn("failed to save index on close", "error", err)
	}

	c.logger.Debug("cache manager closed")
	return nil
}

// Private helper methods

func (c *FileManager[T]) loadIndex() error {
	indexPath := filepath.Join(c.config.BasePath, "metadata", "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing index, start fresh
		}
		return err
	}

	var index fileIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	c.index = index.Entries
	if c.index == nil {
		c.index = make(map[string]*fileEntry)
	}

	if index.Stats != nil {
		c.stats = index.Stats
	} else {
		c.stats = &cacheStats{LastCleanup: time.Now()}
	}

	return nil
}

func (c *FileManager[T]) saveIndex() error {
	metadataDir := filepath.Join(c.config.BasePath, "metadata")
	if err := os.MkdirAll(metadataDir, 0750); err != nil {
		return err
	}

	indexPath := filepath.Join(metadataDir, "index.json")

	// Create a copy of the index and stats under lock
	c.mu.RLock()
	indexCopy := make(map[string]*fileEntry, len(c.index))
	for k, v := range c.index {
		// Create a copy of the entry to avoid race conditions
		entryCopy := *v
		indexCopy[k] = &entryCopy
	}

	statsCopy := *c.stats
	statsCopy.TotalSize = c.calculateTotalSize()
	statsCopy.EntryCount = len(c.index)
	c.mu.RUnlock()

	index := fileIndex{
		Entries: indexCopy,
		Stats:   &statsCopy,
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0600)
}

func (c *FileManager[T]) isExpired(entry *fileEntry) bool {
	if entry.TTL <= 0 {
		return false // No expiration
	}

	expirationTime := entry.CreatedAt.Add(time.Duration(entry.TTL) * time.Second)
	return time.Now().After(expirationTime)
}

func (c *FileManager[T]) ensureSpaceForEntry(newEntry *fileEntry) error {
	totalSize := c.calculateTotalSize()
	sizeLimitBytes := c.config.SizeLimitMB * 1024 * 1024

	if totalSize+newEntry.Size <= sizeLimitBytes {
		return nil // Enough space
	}

	// Need to evict entries
	spaceNeeded := (totalSize + newEntry.Size) - sizeLimitBytes
	evicted := c.evictLRU(spaceNeeded)

	c.logger.Debug("evicted entries to make space", "evicted", evicted, "space_needed", spaceNeeded)
	return nil
}

func (c *FileManager[T]) evictLRU(spaceNeeded int64) int {
	// Sort entries by access time (oldest first)
	type entryWithKey struct {
		key   string
		entry *fileEntry
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

func (c *FileManager[T]) calculateTotalSize() int64 {
	var total int64
	for _, entry := range c.index {
		total += entry.Size
	}
	return total
}

func (c *FileManager[T]) sanitizeFileName(name string) string {
	// Replace problematic characters with underscores
	sanitized := name
	problematicChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range problematicChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	return filepath.Base(sanitized) // Remove path separators
}

func (c *FileManager[T]) removeEmptyDirs() error {
	dirs := []string{
		filepath.Join(c.config.BasePath, "content"),
		filepath.Join(c.config.BasePath, "metadata"),
		c.config.BasePath,
	}

	for _, dir := range dirs {
		if err := c.removeIfEmpty(dir); err != nil {
			return err
		}
	}

	return nil
}

func (c *FileManager[T]) removeIfEmpty(dir string) error {
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
