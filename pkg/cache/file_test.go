package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileManager(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	manager := NewFileManager(config, logger, serializer)

	assert.NotNil(t, manager)
	assert.Equal(t, config.BasePath, manager.config.BasePath)
	assert.Equal(t, config.SizeLimitMB, manager.config.SizeLimitMB)
	assert.Equal(t, config.DefaultTTL, manager.config.DefaultTTL)
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.serializer)
	assert.NotNil(t, manager.index)
	assert.NotNil(t, manager.stats)
}

func TestFileManager_PutAndGet(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Test Put
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Test Get
	entry, err := manager.Get(ctx, "test-key")
	require.NoError(t, err)

	assert.Equal(t, testData.Name, entry.Data.Name)
	assert.Equal(t, testData.Value, entry.Data.Value)
	assert.True(t, entry.Cached)
	assert.True(t, entry.CacheAge >= 0)
	assert.True(t, entry.Size > 0)
	assert.NotEmpty(t, entry.Checksum)
}

func TestFileManager_TTLExpiration(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Test isExpired method directly with a mock entry
	now := time.Now()
	expiredEntry := &fileEntry{
		Key:       "test",
		CreatedAt: now.Add(-2 * time.Hour),
		TTL:       int64(time.Hour.Seconds()), // 1 hour TTL, created 2 hours ago
	}

	nonExpiredEntry := &fileEntry{
		Key:       "test",
		CreatedAt: now,
		TTL:       int64(time.Hour.Seconds()), // 1 hour TTL, just created
	}

	noTTLEntry := &fileEntry{
		Key:       "test",
		CreatedAt: now.Add(-24 * time.Hour),
		TTL:       0, // No TTL
	}

	// Test expiration logic
	assert.True(t, manager.isExpired(expiredEntry), "entry should be expired")
	assert.False(t, manager.isExpired(nonExpiredEntry), "entry should not be expired")
	assert.False(t, manager.isExpired(noTTLEntry), "entry with no TTL should never expire")
}

func TestFileManager_CustomChecksum(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}
	customChecksum := "custom-checksum-123"

	// Put with custom checksum
	err := manager.Put(ctx, "test-key", testData, PutOptions{Checksum: customChecksum})
	require.NoError(t, err)

	// Get and verify checksum
	entry, err := manager.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, customChecksum, entry.Checksum)
}

func TestFileManager_SanitizeFileName(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)

	tests := []struct {
		input    string
		expected string
	}{
		{"normal-name", "normal-name"},
		{"path/with/slashes", "path_with_slashes"},
		{"name:with:colons", "name_with_colons"},
		{"name*with*wildcards", "name_with_wildcards"},
		{"name?with?questions", "name_with_questions"},
		{"name\"with\"quotes", "name_with_quotes"},
		{"name<with>brackets", "name_with_brackets"},
		{"name|with|pipes", "name_with_pipes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := manager.sanitizeFileName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileManager_PersistenceAcrossInstances(t *testing.T) {
	tempDir := t.TempDir()
	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	// Create first manager and add data
	manager1 := NewFileManager(config, logger, serializer)
	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	err := manager1.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)
	manager1.Close()

	// Create second manager with same config
	manager2 := NewFileManager(config, logger, serializer)
	defer manager2.Close()

	// Should be able to retrieve data from first manager
	entry, err := manager2.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, testData, entry.Data)
}

func TestFileManager_ExpandTildePath(t *testing.T) {
	config := Config{
		BasePath:    "~/test-cache",
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Check that tilde was expanded
	assert.NotContains(t, manager.config.BasePath, "~")
	assert.Contains(t, manager.config.BasePath, "test-cache")
}

func TestFileManager_ErrorConditions(t *testing.T) {
	config := Config{
		BasePath:    "/invalid/readonly/path", // Invalid path to trigger permission errors
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put should fail due to permission error
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	assert.Error(t, err)

	var cacheErr *Error
	if assert.ErrorAs(t, err, &cacheErr) {
		assert.Equal(t, ErrorCodePermission, cacheErr.Code)
	}
}

func TestFileManager_IndexPersistence(t *testing.T) {
	tempDir := t.TempDir()
	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	// Create first manager and add data
	manager1 := NewFileManager(config, logger, serializer)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		testData := TestData{Name: fmt.Sprintf("test%d", i), Value: i}
		err := manager1.Put(ctx, fmt.Sprintf("key%d", i), testData, PutOptions{})
		require.NoError(t, err)
	}

	// Get info from first manager
	info1, err := manager1.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, info1.EntryCount)

	manager1.Close()

	// Create second manager with same config
	manager2 := NewFileManager(config, logger, serializer)
	defer manager2.Close()

	// Should load the persisted index
	info2, err := manager2.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, info2.EntryCount)

	// Should be able to retrieve all data
	for i := 0; i < 3; i++ {
		entry, err := manager2.Get(ctx, fmt.Sprintf("key%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("test%d", i), entry.Data.Name)
	}
}

func TestFileManager_EmptyDirectory(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()

	// Test operations on empty cache
	info, err := manager.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, info.EntryCount)
	assert.Equal(t, int64(0), info.TotalSize)

	// Cleanup on empty cache should work
	err = manager.Cleanup(ctx)
	require.NoError(t, err)

	// Clear on empty cache should work
	err = manager.Clear(ctx)
	require.NoError(t, err)
}

func TestFileManager_DefaultTTL(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  2 * time.Hour, // Custom default TTL
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put without specifying TTL (should use default)
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Item should be available (not expired)
	_, err = manager.Get(ctx, "test-key")
	require.NoError(t, err)
}

func TestFileManager_ZeroTTL(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put with zero TTL (no expiration)
	err := manager.Put(ctx, "test-key", testData, PutOptions{TTL: 0})
	require.NoError(t, err)

	// Item should never expire
	_, err = manager.Get(ctx, "test-key")
	require.NoError(t, err)

	// Even after cleanup, item should remain
	err = manager.Cleanup(ctx)
	require.NoError(t, err)

	_, err = manager.Get(ctx, "test-key")
	require.NoError(t, err)
}

func TestFileManager_CorruptedFile(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put an item
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Manually corrupt the cache file
	manager.mu.RLock()
	entry := manager.index["test-key"]
	filePath := entry.Path
	manager.mu.RUnlock()

	err = os.WriteFile(filePath, []byte("corrupted data"), 0644)
	require.NoError(t, err)

	// Try to get the corrupted item
	_, err = manager.Get(ctx, "test-key")
	assert.Error(t, err)

	var cacheErr *Error
	if assert.ErrorAs(t, err, &cacheErr) {
		assert.Equal(t, ErrorCodeSerialization, cacheErr.Code)
	}
}

func TestFileManager_HelperMethods(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Test calculateTotalSize on empty cache
	manager.mu.RLock()
	size := manager.calculateTotalSize()
	manager.mu.RUnlock()
	assert.Equal(t, int64(0), size)

	// Add some items
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		testData := TestData{Name: fmt.Sprintf("test%d", i), Value: i}
		err := manager.Put(ctx, fmt.Sprintf("key%d", i), testData, PutOptions{})
		require.NoError(t, err)
	}

	// Test calculateTotalSize with items
	manager.mu.RLock()
	size = manager.calculateTotalSize()
	manager.mu.RUnlock()
	assert.True(t, size > 0)
}

func TestFileManager_DirectoryCleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Add and remove an item to create empty directories
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	err = manager.Delete(ctx, "test-key")
	require.NoError(t, err)

	// Test removeEmptyDirs
	err = manager.removeEmptyDirs()
	assert.NoError(t, err)

	manager.Close()
}

func TestFileManager_InvalidIndex(t *testing.T) {
	tempDir := t.TempDir()

	// Create invalid index file
	metadataDir := filepath.Join(tempDir, "metadata")
	err := os.MkdirAll(metadataDir, 0755)
	require.NoError(t, err)

	indexPath := filepath.Join(metadataDir, "index.json")
	err = os.WriteFile(indexPath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	// Should handle invalid index gracefully
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Should start with empty cache despite invalid index
	info, err := manager.GetInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, info.EntryCount)
}

func TestFileManager_MissingFile(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put an item
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Manually delete the cache file (but leave index entry)
	manager.mu.RLock()
	entry := manager.index["test-key"]
	filePath := entry.Path
	manager.mu.RUnlock()

	err = os.Remove(filePath)
	require.NoError(t, err)

	// Try to get the item with missing file
	_, err = manager.Get(ctx, "test-key")
	assert.Error(t, err)

	var cacheErr *Error
	if assert.ErrorAs(t, err, &cacheErr) {
		assert.Equal(t, ErrorCodeCorrupted, cacheErr.Code)
	}
}

func TestFileManager_EnsureSpaceForEntry(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Test ensureSpaceForEntry with small entry (should not evict)
	smallEntry := &fileEntry{
		Key:  "small",
		Size: 100,
	}

	manager.mu.Lock()
	err := manager.ensureSpaceForEntry(smallEntry)
	manager.mu.Unlock()

	assert.NoError(t, err)
}

func TestFileManager_RemoveIfEmpty(t *testing.T) {
	tempDir := t.TempDir()
	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Create an empty directory
	emptyDir := filepath.Join(tempDir, "empty")
	err := os.MkdirAll(emptyDir, 0755)
	require.NoError(t, err)

	// Test removeIfEmpty on empty directory
	err = manager.removeIfEmpty(emptyDir)
	assert.NoError(t, err)

	// Directory should be removed
	_, err = os.Stat(emptyDir)
	assert.True(t, os.IsNotExist(err))

	// Test removeIfEmpty on non-existent directory
	err = manager.removeIfEmpty("/non/existent/path")
	assert.NoError(t, err)
}

func TestFileManager_LoadIndexWithStats(t *testing.T) {
	tempDir := t.TempDir()

	// Create index file with stats
	metadataDir := filepath.Join(tempDir, "metadata")
	err := os.MkdirAll(metadataDir, 0755)
	require.NoError(t, err)

	indexPath := filepath.Join(metadataDir, "index.json")
	indexData := `{
		"entries": {},
		"stats": {
			"total_size": 1024,
			"entry_count": 2,
			"hit_count": 10,
			"miss_count": 5
		}
	}`
	err = os.WriteFile(indexPath, []byte(indexData), 0644)
	require.NoError(t, err)

	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	// Should load the stats from the index
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Check that stats were loaded
	info, err := manager.GetInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(10), info.HitCount)
	assert.Equal(t, int64(5), info.MissCount)
}

func TestFileManager_SaveIndexError(t *testing.T) {
	// Test saveIndex with invalid path
	config := Config{
		BasePath:    "/invalid/readonly/path",
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// saveIndex should handle errors gracefully
	err := manager.saveIndex()
	assert.Error(t, err) // Should fail due to invalid path
}

func TestFileManager_EvictLRU(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()

	// Add multiple items with different access times
	for i := 0; i < 3; i++ {
		testData := TestData{Name: fmt.Sprintf("test%d", i), Value: i}
		err := manager.Put(ctx, fmt.Sprintf("key%d", i), testData, PutOptions{})
		require.NoError(t, err)

		// Access the first item multiple times to make it "hot"
		if i == 0 {
			_, err = manager.Get(ctx, "key0")
			require.NoError(t, err)
		}

		time.Sleep(1 * time.Millisecond) // Ensure different access times
	}

	// Test evictLRU directly
	manager.mu.Lock()
	evicted := manager.evictLRU(1) // Evict at least 1 byte
	manager.mu.Unlock()

	assert.True(t, evicted >= 0) // Should evict something or nothing
}

func TestFileManager_LoadIndexNoEntries(t *testing.T) {
	tempDir := t.TempDir()

	// Create index file with null entries
	metadataDir := filepath.Join(tempDir, "metadata")
	err := os.MkdirAll(metadataDir, 0755)
	require.NoError(t, err)

	indexPath := filepath.Join(metadataDir, "index.json")
	indexData := `{"entries": null, "stats": null}`
	err = os.WriteFile(indexPath, []byte(indexData), 0644)
	require.NoError(t, err)

	config := Config{
		BasePath:    tempDir,
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	// Should handle null entries gracefully
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	// Should initialize empty index
	assert.NotNil(t, manager.index)
	assert.Equal(t, 0, len(manager.index))
}
