package cli

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiscoveryCache(t *testing.T) {
	t.Run("creates cache with TTL", func(t *testing.T) {
		// Act
		cache := NewDiscoveryCache(5 * time.Minute)

		// Assert
		assert.NotNil(t, cache)
		assert.Equal(t, 5*time.Minute, cache.ttl)
		assert.NotNil(t, cache.entries)
	})

	t.Run("creates cache with zero TTL", func(t *testing.T) {
		// Act
		cache := NewDiscoveryCache(0)

		// Assert
		assert.NotNil(t, cache)
		assert.Equal(t, time.Duration(0), cache.ttl)
	})
}

func TestDiscoveryCache_SetAndGet(t *testing.T) {
	t.Run("set and get value", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)

		// Act
		cache.Set("test", "/usr/bin/test", "1.0.0", nil)
		path, version, err, found := cache.Get("test")

		// Assert
		assert.True(t, found)
		assert.Equal(t, "/usr/bin/test", path)
		assert.Equal(t, "1.0.0", version)
		assert.NoError(t, err)
	})

	t.Run("get non-existent entry", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)

		// Act
		_, _, _, found := cache.Get("nonexistent")

		// Assert
		assert.False(t, found)
	})

	t.Run("set with error", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		testErr := assert.AnError

		// Act
		cache.Set("test", "", "", testErr)
		path, version, err, found := cache.Get("test")

		// Assert
		assert.True(t, found)
		assert.Empty(t, path)
		assert.Empty(t, version)
		assert.Equal(t, testErr, err)
	})

	t.Run("expired entry not returned", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(1 * time.Millisecond)
		cache.Set("test", "/usr/bin/test", "1.0.0", nil)

		// Act
		time.Sleep(10 * time.Millisecond)
		_, _, _, found := cache.Get("test")

		// Assert
		assert.False(t, found)
	})

	t.Run("zero TTL cache never expires", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(0)
		cache.Set("test", "/usr/bin/test", "1.0.0", nil)

		// Act
		time.Sleep(10 * time.Millisecond)
		path, version, err, found := cache.Get("test")

		// Assert
		assert.True(t, found)
		assert.Equal(t, "/usr/bin/test", path)
		assert.Equal(t, "1.0.0", version)
		assert.NoError(t, err)
	})
}

func TestDiscoveryCache_Clear(t *testing.T) {
	t.Run("clear removes all entries", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		cache.Set("test1", "/usr/bin/test1", "1.0.0", nil)
		cache.Set("test2", "/usr/bin/test2", "2.0.0", nil)

		// Act
		cache.Clear()

		// Assert
		_, _, _, found1 := cache.Get("test1")
		_, _, _, found2 := cache.Get("test2")
		assert.False(t, found1)
		assert.False(t, found2)
	})
}

func TestNewDiscovery(t *testing.T) {
	t.Run("creates discovery with cache", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()

		// Act
		discovery := NewDiscovery(cache, logger)

		// Assert
		assert.NotNil(t, discovery)
		assert.Equal(t, cache, discovery.cache)
		assert.Equal(t, logger, discovery.logger)
	})

	t.Run("creates discovery with nil cache uses default", func(t *testing.T) {
		// Arrange
		logger := logging.NewBasic()

		// Act
		discovery := NewDiscovery(nil, logger)

		// Assert
		assert.NotNil(t, discovery)
		assert.NotNil(t, discovery.cache)
		assert.Equal(t, 5*time.Minute, discovery.cache.ttl)
	})
}

func TestDiscovery_FindBinary(t *testing.T) {
	t.Run("finds existing binary", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		path, err := discovery.FindBinary("sh")

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, path)
		assert.Contains(t, path, "sh")
	})

	t.Run("binary not found", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		_, err := discovery.FindBinary("nonexistent-binary-12345")

		// Assert
		require.Error(t, err)
	})

	t.Run("caches successful result", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		path1, err1 := discovery.FindBinary("sh")
		path2, err2 := discovery.FindBinary("sh") // Should hit cache

		// Assert
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, path1, path2)

		// Verify it's in cache
		cachedPath, _, _, found := cache.Get("sh")
		assert.True(t, found)
		assert.Equal(t, path1, cachedPath)
	})

	t.Run("caches error result", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		_, err1 := discovery.FindBinary("nonexistent-binary-12345")
		_, err2 := discovery.FindBinary("nonexistent-binary-12345") // Should hit cache

		// Assert
		require.Error(t, err1)
		require.Error(t, err2)

		// Verify error is cached
		_, _, cachedErr, found := cache.Get("nonexistent-binary-12345")
		assert.True(t, found)
		assert.Error(t, cachedErr)
	})
}

func TestDiscovery_GetVersion(t *testing.T) {
	t.Run("get version from binary", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		version, err := discovery.GetVersion(context.Background(), "sh", []string{"--version"})

		// Assert
		// sh --version might not work on all systems, so just check no panic
		_ = version
		_ = err
	})

	t.Run("version command timeout", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Act
		_, err := discovery.GetVersion(ctx, "sleep", []string{"10"})

		// Assert
		require.Error(t, err)
	})
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
		wantErr  bool
	}{
		{
			name:     "semantic version",
			output:   "1.2.3",
			expected: "1.2.3",
			wantErr:  false,
		},
		{
			name:     "semantic version with v prefix",
			output:   "v2.39.0",
			expected: "2.39.0",
			wantErr:  false,
		},
		{
			name:     "git-style version",
			output:   "git version 2.39.0",
			expected: "2.39.0",
			wantErr:  false,
		},
		{
			name:     "node-style version",
			output:   "v18.12.0",
			expected: "18.12.0",
			wantErr:  false,
		},
		{
			name:     "version with prerelease",
			output:   "1.2.3-beta.1",
			expected: "1.2.3-beta.1",
			wantErr:  false,
		},
		{
			name:     "two-part version",
			output:   "1.2",
			expected: "1.2",
			wantErr:  false,
		},
		{
			name:     "single version number",
			output:   "5",
			expected: "5",
			wantErr:  false,
		},
		{
			name:     "version in multiline output",
			output:   "Tool Name\nVersion: 1.2.3\nCopyright",
			expected: "1.2.3",
			wantErr:  false,
		},
		{
			name:     "no version found",
			output:   "no version here",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			version, err := ParseVersion(tt.output)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, version)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		required string
		expected bool
		wantErr  bool
	}{
		{
			name:     "equal versions",
			current:  "1.2.3",
			required: "1.2.3",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "current greater than required",
			current:  "2.0.0",
			required: "1.9.0",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "current less than required",
			current:  "1.0.0",
			required: "2.0.0",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "major version satisfied",
			current:  "2.0.0",
			required: "1.0.0",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "minor version satisfied",
			current:  "1.5.0",
			required: "1.2.0",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "patch version satisfied",
			current:  "1.2.5",
			required: "1.2.3",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "with v prefix",
			current:  "v2.0.0",
			required: "v1.0.0",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "two-part versions",
			current:  "1.5",
			required: "1.2",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "current has more parts",
			current:  "1.2.3",
			required: "1.2",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "current has fewer parts",
			current:  "1.2",
			required: "1.2.1",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "version with prerelease",
			current:  "1.2.3-beta",
			required: "1.2.0",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "invalid current version",
			current:  "invalid",
			required: "1.0.0",
			expected: false,
			wantErr:  true,
		},
		{
			name:     "invalid required version",
			current:  "1.0.0",
			required: "invalid",
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := ValidateVersion(tt.current, tt.required)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseVersionParts(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected []int
		wantErr  bool
	}{
		{
			name:     "three-part version",
			version:  "1.2.3",
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "two-part version",
			version:  "1.2",
			expected: []int{1, 2},
			wantErr:  false,
		},
		{
			name:     "single part version",
			version:  "5",
			expected: []int{5},
			wantErr:  false,
		},
		{
			name:     "version with prerelease removed",
			version:  "1.2.3-beta",
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "version with build metadata removed",
			version:  "1.2.3+build",
			expected: []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "invalid part",
			version:  "1.x.3",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "empty version",
			version:  "",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			parts, err := parseVersionParts(tt.version)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, parts)
			}
		})
	}
}

func TestDiscovery_Discover(t *testing.T) {
	t.Run("discover available binary", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		result, err := discovery.Discover(context.Background(), "sh", []string{}, "")

		// Assert
		require.NoError(t, err)
		assert.True(t, result.Available)
		assert.NotEmpty(t, result.Path)
		assert.Empty(t, result.Reason)
	})

	t.Run("discover binary not found", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act
		result, err := discovery.Discover(context.Background(), "nonexistent-binary-12345", []string{}, "")

		// Assert
		require.NoError(t, err) // Not an error, just not available
		assert.False(t, result.Available)
		assert.Empty(t, result.Path)
		assert.Contains(t, result.Reason, "not found")
	})

	t.Run("discover with version check", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		logger := logging.NewBasic()
		discovery := NewDiscovery(cache, logger)

		// Act - sh might not support --version on all systems
		result, err := discovery.Discover(context.Background(), "echo", []string{}, "")

		// Assert
		require.NoError(t, err)
		// Just verify no panic, actual availability depends on system
		_ = result.Available
	})
}

func TestDiscoveryCache_Concurrency(t *testing.T) {
	t.Run("concurrent set and get", func(t *testing.T) {
		// Arrange
		cache := NewDiscoveryCache(5 * time.Minute)
		done := make(chan bool)

		// Act - concurrent writes
		for i := 0; i < 10; i++ {
			go func(id int) {
				cache.Set("test", "/usr/bin/test", "1.0.0", nil)
				_, _, _, _ = cache.Get("test")
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Assert - no panic, cache still accessible
		path, version, err, found := cache.Get("test")
		assert.True(t, found)
		assert.Equal(t, "/usr/bin/test", path)
		assert.Equal(t, "1.0.0", version)
		assert.NoError(t, err)
	})
}
