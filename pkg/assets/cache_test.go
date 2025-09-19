package assets

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileCacheManager(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()

	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)

	assert.NotNil(t, cache)
	assert.Equal(t, tempDir, cache.basePath)
	assert.Equal(t, int64(100), cache.sizeLimitMB)
	assert.Equal(t, time.Hour, cache.defaultTTL)
	assert.NotNil(t, cache.logger)
	assert.NotNil(t, cache.index)
}

func TestFileCacheManager_Put_And_Get(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Create test content
	content := &AssetContent{
		Metadata: AssetMetadata{
			Name:        "test-asset",
			Type:        AssetTypeTemplate,
			Description: "Test asset",
		},
		Content:  "# Test Content",
		Checksum: "sha256:test123",
	}

	// Test Put
	err := cache.Put(ctx, "test-key", content)
	require.NoError(t, err)

	// Test Get
	retrieved, err := cache.Get(ctx, "test-key")
	require.NoError(t, err)

	assert.Equal(t, content.Content, retrieved.Content)
	assert.Equal(t, content.Checksum, retrieved.Checksum)
	assert.True(t, retrieved.Cached)
	assert.True(t, retrieved.CacheAge >= 0)
}

func TestFileCacheManager_Get_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Try to get non-existent key
	content, err := cache.Get(ctx, "nonexistent")

	assert.Nil(t, content)
	assert.Error(t, err)

	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeCacheError, assetErr.Code)
	assert.Contains(t, assetErr.Message, "not found in cache")
}

func TestFileCacheManager_Put_EmptyKey(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	content := &AssetContent{
		Content: "test",
	}

	err := cache.Put(ctx, "", content)

	assert.Error(t, err)
	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeCacheError, assetErr.Code)
	assert.Contains(t, assetErr.Message, "cannot be empty")
}

func TestFileCacheManager_Delete(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Put an item first
	content := &AssetContent{
		Content:  "test content",
		Checksum: "test123",
	}

	err := cache.Put(ctx, "test-key", content)
	require.NoError(t, err)

	// Verify it exists
	_, err = cache.Get(ctx, "test-key")
	require.NoError(t, err)

	// Delete it
	err = cache.Delete(ctx, "test-key")
	require.NoError(t, err)

	// Verify it's gone
	_, err = cache.Get(ctx, "test-key")
	assert.Error(t, err)
}

func TestFileCacheManager_Delete_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Delete non-existent key should not error
	err := cache.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestFileCacheManager_Clear(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Put multiple items
	for i := 0; i < 3; i++ {
		content := &AssetContent{
			Content:  fmt.Sprintf("content %d", i),
			Checksum: fmt.Sprintf("checksum%d", i),
		}
		err := cache.Put(ctx, fmt.Sprintf("key%d", i), content)
		require.NoError(t, err)
	}

	// Clear all
	err := cache.Clear(ctx)
	require.NoError(t, err)

	// Verify all are gone
	for i := 0; i < 3; i++ {
		_, err := cache.Get(ctx, fmt.Sprintf("key%d", i))
		assert.Error(t, err)
	}
}

func TestFileCacheManager_GetInfo(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Initially empty
	info, err := cache.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.TotalSize)
	assert.Equal(t, 0, info.AssetCount)

	// Add an item
	content := &AssetContent{
		Content:  "test content",
		Checksum: "test123",
	}

	err = cache.Put(ctx, "test-key", content)
	require.NoError(t, err)

	// Check info after adding
	info, err = cache.GetInfo(ctx)
	require.NoError(t, err)
	assert.True(t, info.TotalSize > 0)
	assert.Equal(t, 1, info.AssetCount)
}

func TestFileCacheManager_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)
	ctx := context.Background()

	// Add an item
	content := &AssetContent{
		Content:  "test content",
		Checksum: "test123",
	}

	err := cache.Put(ctx, "test-key", content)
	require.NoError(t, err)

	// Run cleanup (should not remove non-expired items)
	err = cache.Cleanup(ctx)
	require.NoError(t, err)

	// Verify item still exists
	_, err = cache.Get(ctx, "test-key")
	assert.NoError(t, err)
}

func TestFileCacheManager_SanitizeFileName(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewFileCacheManager(tempDir, 100, time.Hour, logger)

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
			result := cache.sanitizeFileName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
