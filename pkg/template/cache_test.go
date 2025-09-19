package template

import (
	"fmt"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCache(t *testing.T) {
	logger := logging.NewBasic()

	// Test with valid parameters
	cache := NewMemoryCache(50, 15*time.Minute, logger)
	assert.NotNil(t, cache)
	assert.Equal(t, 50, cache.maxSize)
	assert.Equal(t, 15*time.Minute, cache.ttl)

	// Test with zero/negative parameters (should use defaults)
	cache = NewMemoryCache(0, 0, logger)
	assert.Equal(t, 100, cache.maxSize)        // Default
	assert.Equal(t, 30*time.Minute, cache.ttl) // Default
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(10, 30*time.Minute, logger)

	tmpl := &Template{
		Name:    "test-template",
		Content: "Hello {{.name}}!",
	}

	// Test Set
	err := cache.Set("test-key", tmpl)
	assert.NoError(t, err)

	// Test Get
	retrieved, found := cache.Get("test-key")
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Equal(t, tmpl.Name, retrieved.Name)
	assert.Equal(t, tmpl.Content, retrieved.Content)

	// Test Get non-existent key
	retrieved, found = cache.Get("non-existent")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestMemoryCache_Delete(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(10, 30*time.Minute, logger)

	tmpl := &Template{
		Name:    "test-template",
		Content: "Hello {{.name}}!",
	}

	// Set and verify
	cache.Set("test-key", tmpl)
	_, found := cache.Get("test-key")
	assert.True(t, found)

	// Delete and verify
	err := cache.Delete("test-key")
	assert.NoError(t, err)

	_, found = cache.Get("test-key")
	assert.False(t, found)

	// Delete non-existent key (should not error)
	err = cache.Delete("non-existent")
	assert.NoError(t, err)
}

func TestMemoryCache_Clear(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(10, 30*time.Minute, logger)

	// Add multiple items
	for i := 0; i < 5; i++ {
		tmpl := &Template{
			Name:    fmt.Sprintf("template-%d", i),
			Content: "test content",
		}
		cache.Set(fmt.Sprintf("key-%d", i), tmpl)
	}

	// Verify items exist
	stats := cache.Stats()
	assert.Equal(t, 5, stats.Size)

	// Clear cache
	err := cache.Clear()
	assert.NoError(t, err)

	// Verify cache is empty
	stats = cache.Stats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
}

func TestMemoryCache_Stats(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(10, 30*time.Minute, logger)

	tmpl := &Template{
		Name:    "test-template",
		Content: "test content",
	}

	// Initial stats
	stats := cache.Stats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 0.0, stats.HitRatio)

	// Add item and get stats
	cache.Set("test-key", tmpl)
	stats = cache.Stats()
	assert.Equal(t, 1, stats.Size)

	// Test cache hit
	cache.Get("test-key")
	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 1.0, stats.HitRatio)

	// Test cache miss
	cache.Get("non-existent")
	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 0.5, stats.HitRatio)
}

func TestMemoryCache_LRUEviction(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(3, 30*time.Minute, logger) // Small cache for testing eviction

	// Fill cache to capacity
	for i := 0; i < 3; i++ {
		tmpl := &Template{
			Name:    fmt.Sprintf("template-%d", i),
			Content: "test content",
		}
		cache.Set(fmt.Sprintf("key-%d", i), tmpl)
	}

	// Verify cache is full
	stats := cache.Stats()
	assert.Equal(t, 3, stats.Size)

	// Access key-1 to make it more recently used
	cache.Get("key-1")

	// Add one more item, should evict least recently used (key-0)
	tmpl := &Template{
		Name:    "template-new",
		Content: "new content",
	}
	cache.Set("key-new", tmpl)

	// Verify cache size is still 3
	stats = cache.Stats()
	assert.Equal(t, 3, stats.Size)

	// Verify key-0 was evicted
	_, found := cache.Get("key-0")
	assert.False(t, found)

	// Verify key-1 and key-2 are still there
	_, found = cache.Get("key-1")
	assert.True(t, found)
	_, found = cache.Get("key-2")
	assert.True(t, found)
	_, found = cache.Get("key-new")
	assert.True(t, found)
}

func TestMemoryCache_TTLExpiration(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(10, 50*time.Millisecond, logger) // Very short TTL

	tmpl := &Template{
		Name:    "test-template",
		Content: "test content",
	}

	// Set item
	cache.Set("test-key", tmpl)

	// Should be available immediately
	retrieved, found := cache.Get("test-key")
	assert.True(t, found)
	assert.NotNil(t, retrieved)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	retrieved, found = cache.Get("test-key")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestMemoryCache_AccessCount(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(10, 30*time.Minute, logger)

	tmpl := &Template{
		Name:    "test-template",
		Content: "test content",
	}

	cache.Set("test-key", tmpl)

	// Access multiple times
	for i := 0; i < 5; i++ {
		cache.Get("test-key")
	}

	// Check internal state (access count should be tracked)
	cache.mu.RLock()
	item, exists := cache.items["test-key"]
	cache.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, int64(5), item.accessCount)
}

func TestNewNoOpCache(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewNoOpCache(logger)

	assert.NotNil(t, cache)
	assert.Equal(t, logger, cache.logger)
}

func TestNoOpCache_Operations(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewNoOpCache(logger)

	tmpl := &Template{
		Name:    "test-template",
		Content: "test content",
	}

	// All operations should work but not actually cache anything
	err := cache.Set("test-key", tmpl)
	assert.NoError(t, err)

	retrieved, found := cache.Get("test-key")
	assert.False(t, found)
	assert.Nil(t, retrieved)

	err = cache.Delete("test-key")
	assert.NoError(t, err)

	err = cache.Clear()
	assert.NoError(t, err)

	stats := cache.Stats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 0.0, stats.HitRatio)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	logger := logging.NewBasic()
	cache := NewMemoryCache(100, 30*time.Minute, logger)

	tmpl := &Template{
		Name:    "test-template",
		Content: "test content",
	}

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.Set(fmt.Sprintf("key-%d", id), tmpl)
			done <- true
		}(i)
	}

	// Wait for all writes to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.Get(fmt.Sprintf("key-%d", id))
			done <- true
		}(i)
	}

	// Wait for all reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify cache state
	stats := cache.Stats()
	assert.Equal(t, 10, stats.Size)
}
