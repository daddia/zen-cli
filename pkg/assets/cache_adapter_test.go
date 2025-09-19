package assets

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAssetCacheManager(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()

	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.cache)
}

func TestAssetCacheManager_PutAndGet(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)
	defer cache.Close()

	ctx := context.Background()

	// Create test asset content
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

func TestAssetCacheManager_GetNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)
	defer cache.Close()

	ctx := context.Background()

	// Try to get non-existent asset
	content, err := cache.Get(ctx, "nonexistent")

	assert.Nil(t, content)
	assert.Error(t, err)

	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeAssetNotFound, assetErr.Code)
}

func TestAssetCacheManager_PutNilContent(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)
	defer cache.Close()

	ctx := context.Background()

	err := cache.Put(ctx, "test-key", nil)

	assert.Error(t, err)
	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeCacheError, assetErr.Code)
	assert.Contains(t, assetErr.Message, "cannot be nil")
}

func TestAssetCacheManager_Delete(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)
	defer cache.Close()

	ctx := context.Background()

	// Create and put test content
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

func TestAssetCacheManager_Clear(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)
	defer cache.Close()

	ctx := context.Background()

	// Put multiple assets
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

func TestAssetCacheManager_GetInfo(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	cache := NewAssetCacheManager(tempDir, 100, time.Hour, logger)
	defer cache.Close()

	ctx := context.Background()

	// Initially empty
	info, err := cache.GetInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.TotalSize)
	assert.Equal(t, 0, info.AssetCount)

	// Add an asset
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

func TestAssetContentSerializer(t *testing.T) {
	serializer := NewAssetContentSerializer()

	content := AssetContent{
		Metadata: AssetMetadata{
			Name: "test",
			Type: AssetTypeTemplate,
		},
		Content:  "# Test Content",
		Checksum: "sha256:test",
	}

	// Test Serialize
	data, err := serializer.Serialize(content)
	require.NoError(t, err)
	assert.Equal(t, []byte("# Test Content"), data)

	// Test Deserialize
	result, err := serializer.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, "# Test Content", result.Content)

	// Test ContentType
	assert.Equal(t, "text/plain", serializer.ContentType())
}

func TestConvertCacheErrorCode(t *testing.T) {
	tests := []struct {
		input    cache.ErrorCode
		expected AssetErrorCode
	}{
		{cache.ErrorCodeNotFound, ErrorCodeAssetNotFound},
		{cache.ErrorCodeInvalidKey, ErrorCodeCacheError},
		{cache.ErrorCodeStorageFull, ErrorCodeCacheError},
		{cache.ErrorCodeCorrupted, ErrorCodeIntegrityError},
		{cache.ErrorCodePermission, ErrorCodeCacheError},
		{cache.ErrorCodeSerialization, ErrorCodeCacheError},
		{"unknown", ErrorCodeCacheError},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := convertCacheErrorCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
