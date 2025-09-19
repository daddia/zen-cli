package template

import (
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
)

// MemoryCache implements CacheManager using in-memory storage
type MemoryCache struct {
	logger  logging.Logger
	maxSize int
	ttl     time.Duration

	mu    sync.RWMutex
	items map[string]*cacheItem
	stats CacheStats
}

// cacheItem represents a cached template with metadata
type cacheItem struct {
	template    *Template
	createdAt   time.Time
	accessedAt  time.Time
	accessCount int64
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int, ttl time.Duration, logger logging.Logger) *MemoryCache {
	if maxSize <= 0 {
		maxSize = 100 // Default cache size
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute // Default TTL
	}

	cache := &MemoryCache{
		logger:  logger,
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*cacheItem),
		stats:   CacheStats{},
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a compiled template from cache
func (c *MemoryCache) Get(key string) (*Template, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if item has expired
	if c.isExpired(item) {
		c.stats.Misses++
		// Remove expired item (will be done by cleanup goroutine)
		return nil, false
	}

	// Update access statistics (with proper locking)
	c.mu.RUnlock()
	c.mu.Lock()
	item.accessedAt = time.Now()
	item.accessCount++
	c.stats.Hits++
	c.mu.Unlock()
	c.mu.RLock()

	c.logger.Debug("template cache hit", "key", key, "access_count", item.accessCount)
	return item.template, true
}

// Set stores a compiled template in cache
func (c *MemoryCache) Set(key string, tmpl *Template) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items to make room
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}

	// Create cache item
	item := &cacheItem{
		template:    tmpl,
		createdAt:   time.Now(),
		accessedAt:  time.Now(),
		accessCount: 0,
	}

	c.items[key] = item
	c.stats.Size = len(c.items)

	c.logger.Debug("template cached", "key", key, "cache_size", c.stats.Size)
	return nil
}

// Delete removes a template from cache
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		delete(c.items, key)
		c.stats.Size = len(c.items)
		c.logger.Debug("template removed from cache", "key", key, "cache_size", c.stats.Size)
	}

	return nil
}

// Clear clears all cached templates
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.stats.Size = 0
	c.stats.Hits = 0
	c.stats.Misses = 0

	c.logger.Debug("template cache cleared")
	return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.items)

	// Calculate hit ratio
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(total)
	}

	return stats
}

// isExpired checks if a cache item has expired
func (c *MemoryCache) isExpired(item *cacheItem) bool {
	return time.Since(item.createdAt) > c.ttl
}

// evictLRU evicts the least recently used item
func (c *MemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	// Find the least recently accessed item
	for key, item := range c.items {
		if oldestKey == "" || item.accessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.accessedAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		c.logger.Debug("evicted LRU template from cache", "key", oldestKey)
	}
}

// cleanupExpired runs periodically to remove expired items
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		var expiredKeys []string
		for key, item := range c.items {
			if c.isExpired(item) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(c.items, key)
		}

		if len(expiredKeys) > 0 {
			c.stats.Size = len(c.items)
			c.logger.Debug("cleaned up expired templates", "count", len(expiredKeys), "cache_size", c.stats.Size)
		}

		c.mu.Unlock()
	}
}

// NoOpCache implements CacheManager with no-op operations (disabled caching)
type NoOpCache struct {
	logger logging.Logger
}

// NewNoOpCache creates a cache that doesn't actually cache anything
func NewNoOpCache(logger logging.Logger) *NoOpCache {
	return &NoOpCache{
		logger: logger,
	}
}

// Get always returns cache miss
func (c *NoOpCache) Get(key string) (*Template, bool) {
	return nil, false
}

// Set does nothing
func (c *NoOpCache) Set(key string, tmpl *Template) error {
	return nil
}

// Delete does nothing
func (c *NoOpCache) Delete(key string) error {
	return nil
}

// Clear does nothing
func (c *NoOpCache) Clear() error {
	return nil
}

// Stats returns empty statistics
func (c *NoOpCache) Stats() CacheStats {
	return CacheStats{
		Size:     0,
		Hits:     0,
		Misses:   0,
		HitRatio: 0.0,
	}
}
