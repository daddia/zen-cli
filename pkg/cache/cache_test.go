package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "~/.zen/cache", config.BasePath)
	assert.Equal(t, int64(100), config.SizeLimitMB)
	assert.Equal(t, 24*time.Hour, config.DefaultTTL)
	assert.Equal(t, 1*time.Hour, config.CleanupInterval)
	assert.False(t, config.EnableCompression)
}

func TestNewManager(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[string]()

	manager := NewManager(config, logger, serializer)

	assert.NotNil(t, manager)
}

func TestCacheError(t *testing.T) {
	err := &Error{
		Code:    ErrorCodeNotFound,
		Message: "test error",
		Details: "test details",
	}

	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, ErrorCodeNotFound, err.Code)
	assert.Equal(t, "test details", err.Details)
}

// Test data structures for testing
type TestData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func createTestManager(t *testing.T) Manager[TestData] {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	return NewManager(config, logger, serializer)
}

func TestManager_PutAndGet(t *testing.T) {
	manager := createTestManager(t)
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

func TestManager_GetNotFound(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()

	// Try to get non-existent key
	entry, err := manager.Get(ctx, "nonexistent")

	assert.Nil(t, entry)
	assert.Error(t, err)

	var cacheErr *Error
	assert.ErrorAs(t, err, &cacheErr)
	assert.Equal(t, ErrorCodeNotFound, cacheErr.Code)
	assert.Contains(t, cacheErr.Message, "not found in cache")
}

func TestManager_PutEmptyKey(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	err := manager.Put(ctx, "", testData, PutOptions{})

	assert.Error(t, err)
	var cacheErr *Error
	assert.ErrorAs(t, err, &cacheErr)
	assert.Equal(t, ErrorCodeInvalidKey, cacheErr.Code)
	assert.Contains(t, cacheErr.Message, "cannot be empty")
}

func TestManager_Delete(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put an item first
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Verify it exists
	_, err = manager.Get(ctx, "test-key")
	require.NoError(t, err)

	// Delete it
	err = manager.Delete(ctx, "test-key")
	require.NoError(t, err)

	// Verify it's gone
	_, err = manager.Get(ctx, "test-key")
	assert.Error(t, err)
}

func TestManager_DeleteNotFound(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()

	// Delete non-existent key should not error
	err := manager.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestManager_Clear(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()

	// Put multiple items
	for i := 0; i < 3; i++ {
		testData := TestData{Name: "test", Value: i}
		err := manager.Put(ctx, fmt.Sprintf("key%d", i), testData, PutOptions{})
		require.NoError(t, err)
	}

	// Clear all
	err := manager.Clear(ctx)
	require.NoError(t, err)

	// Verify all are gone
	for i := 0; i < 3; i++ {
		_, err := manager.Get(ctx, fmt.Sprintf("key%d", i))
		assert.Error(t, err)
	}
}

func TestManager_GetInfo(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()

	// Initially empty
	info, err := manager.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.TotalSize)
	assert.Equal(t, 0, info.EntryCount)
	assert.Equal(t, float64(0), info.CacheHitRatio)

	// Add an item
	testData := TestData{Name: "test", Value: 42}
	err = manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Check info after adding
	info, err = manager.GetInfo(ctx)
	require.NoError(t, err)
	assert.True(t, info.TotalSize > 0)
	assert.Equal(t, 1, info.EntryCount)

	// Test cache hit ratio
	_, err = manager.Get(ctx, "test-key") // Hit
	require.NoError(t, err)

	_, err = manager.Get(ctx, "missing") // Miss
	assert.Error(t, err)

	info, err = manager.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), info.HitCount)
	assert.Equal(t, int64(1), info.MissCount)
	assert.Equal(t, float64(0.5), info.CacheHitRatio)
}

func TestManager_Cleanup(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Add a normal item
	err := manager.Put(ctx, "test-key", testData, PutOptions{})
	require.NoError(t, err)

	// Run cleanup (should not remove non-expired items)
	err = manager.Cleanup(ctx)
	require.NoError(t, err)

	// Item should still be available
	_, err = manager.Get(ctx, "test-key")
	require.NoError(t, err)
}

func TestManager_WithCustomTTL(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}

	// Put with custom TTL
	customTTL := 2 * time.Hour
	err := manager.Put(ctx, "test-key", testData, PutOptions{TTL: customTTL})
	require.NoError(t, err)

	// Get and verify
	entry, err := manager.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, testData, entry.Data)
}

func TestManager_WithChecksum(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()
	testData := TestData{Name: "test", Value: 42}
	customChecksum := "custom-checksum-123"

	// Put with custom checksum
	err := manager.Put(ctx, "test-key", testData, PutOptions{Checksum: customChecksum})
	require.NoError(t, err)

	// Get and verify checksum is preserved
	entry, err := manager.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, customChecksum, entry.Checksum)
}

func TestManager_StringData(t *testing.T) {
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	logger := logging.NewBasic()
	serializer := NewStringSerializer()
	manager := NewFileManager(config, logger, serializer)
	defer manager.Close()

	ctx := context.Background()
	testString := "Hello, World!"

	// Test Put and Get with string data
	err := manager.Put(ctx, "string-key", testString, PutOptions{})
	require.NoError(t, err)

	entry, err := manager.Get(ctx, "string-key")
	require.NoError(t, err)
	assert.Equal(t, testString, entry.Data)
}

func TestManager_Close(t *testing.T) {
	manager := createTestManager(t)

	err := manager.Close()
	assert.NoError(t, err)
}

func TestManager_MultipleOperations(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	ctx := context.Background()

	// Test multiple puts and gets
	for i := 0; i < 5; i++ {
		testData := TestData{Name: fmt.Sprintf("test%d", i), Value: i}
		err := manager.Put(ctx, fmt.Sprintf("key%d", i), testData, PutOptions{})
		require.NoError(t, err)
	}

	// Verify all items
	for i := 0; i < 5; i++ {
		entry, err := manager.Get(ctx, fmt.Sprintf("key%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("test%d", i), entry.Data.Name)
		assert.Equal(t, i, entry.Data.Value)
	}

	// Test info
	info, err := manager.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, info.EntryCount)
	assert.True(t, info.TotalSize > 0)
}

func TestManager_ErrorCodes(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrorCodeNotFound, "not_found"},
		{ErrorCodeInvalidKey, "invalid_key"},
		{ErrorCodeStorageFull, "storage_full"},
		{ErrorCodeCorrupted, "corrupted"},
		{ErrorCodePermission, "permission"},
		{ErrorCodeSerialization, "serialization"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.code))
		})
	}
}

func TestManager_ConfigValidation(t *testing.T) {
	logger := logging.NewBasic()
	serializer := NewJSONSerializer[TestData]()

	// Test with zero size limit
	config := Config{
		BasePath:    t.TempDir(),
		SizeLimitMB: 0,
		DefaultTTL:  time.Hour,
	}

	manager := NewManager(config, logger, serializer)
	assert.NotNil(t, manager)
	manager.Close()

	// Test with very large TTL
	config.DefaultTTL = 365 * 24 * time.Hour
	manager = NewManager(config, logger, serializer)
	assert.NotNil(t, manager)
	manager.Close()
}

func TestPutOptions(t *testing.T) {
	opts := PutOptions{
		TTL:      time.Hour,
		Checksum: "test-checksum",
	}

	assert.Equal(t, time.Hour, opts.TTL)
	assert.Equal(t, "test-checksum", opts.Checksum)
}

func TestEntry(t *testing.T) {
	entry := Entry[string]{
		Data:     "test data",
		Checksum: "test-checksum",
		Cached:   true,
		CacheAge: 3600,
		Size:     100,
	}

	assert.Equal(t, "test data", entry.Data)
	assert.Equal(t, "test-checksum", entry.Checksum)
	assert.True(t, entry.Cached)
	assert.Equal(t, int64(3600), entry.CacheAge)
	assert.Equal(t, int64(100), entry.Size)
}

func TestInfo(t *testing.T) {
	info := Info{
		TotalSize:     1024,
		EntryCount:    5,
		HitCount:      10,
		MissCount:     2,
		CacheHitRatio: 0.83,
	}

	assert.Equal(t, int64(1024), info.TotalSize)
	assert.Equal(t, 5, info.EntryCount)
	assert.Equal(t, int64(10), info.HitCount)
	assert.Equal(t, int64(2), info.MissCount)
	assert.Equal(t, 0.83, info.CacheHitRatio)
}
