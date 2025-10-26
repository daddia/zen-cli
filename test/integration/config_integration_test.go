package integration_test

import (
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStandardConfigInterfaces tests the new standard configuration interfaces
func TestStandardConfigInterfaces(t *testing.T) {
	// Load default configuration
	cfg := config.LoadDefaults()
	require.NotNil(t, cfg)

	t.Run("Assets Config", func(t *testing.T) {
		// Test assets configuration using standard API
		assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
		require.NoError(t, err)

		// Verify it implements Configurable interface
		assert.NoError(t, assetConfig.Validate())

		// Verify defaults
		defaults := assetConfig.Defaults()
		assert.NotNil(t, defaults)

		// Verify parser section
		parser := assets.ConfigParser{}
		assert.Equal(t, "assets", parser.Section())
	})

	t.Run("Template Config", func(t *testing.T) {
		// Test template configuration using standard API
		templateConfig, err := config.GetConfig(cfg, template.ConfigParser{})
		require.NoError(t, err)

		// Verify it implements Configurable interface
		assert.NoError(t, templateConfig.Validate())

		// Verify defaults
		defaults := templateConfig.Defaults()
		assert.NotNil(t, defaults)

		// Verify parser section
		parser := template.ConfigParser{}
		assert.Equal(t, "templates", parser.Section())
	})

	t.Run("Auth Config", func(t *testing.T) {
		// Test auth configuration using standard API
		authConfig, err := config.GetConfig(cfg, auth.ConfigParser{})
		require.NoError(t, err)

		// Verify it implements Configurable interface
		assert.NoError(t, authConfig.Validate())

		// Verify defaults
		defaults := authConfig.Defaults()
		assert.NotNil(t, defaults)

		// Verify parser section
		parser := auth.ConfigParser{}
		assert.Equal(t, "auth", parser.Section())
	})

	t.Run("Cache Config", func(t *testing.T) {
		// Test cache configuration using standard API
		cacheConfig, err := config.GetConfig(cfg, cache.ConfigParser{})
		require.NoError(t, err)

		// Verify it implements Configurable interface
		assert.NoError(t, cacheConfig.Validate())

		// Verify defaults
		defaults := cacheConfig.Defaults()
		assert.NotNil(t, defaults)

		// Verify parser section
		parser := cache.ConfigParser{}
		assert.Equal(t, "cache", parser.Section())
	})
}

// TestConfigSetAndGet tests setting and getting configuration values
func TestConfigSetAndGet(t *testing.T) {
	cfg := config.LoadDefaults()
	require.NotNil(t, cfg)

	t.Run("Set and Get Asset Config", func(t *testing.T) {
		// Create a custom asset config
		customConfig := assets.AssetConfig{
			RepositoryURL:          "https://github.com/test/repo.git",
			Branch:                 "test-branch",
			CachePath:              "/tmp/test-cache",
			CacheSizeMB:            200,
			AuthProvider:           "github",
			SyncTimeoutSeconds:     60,
			MaxConcurrentOps:       5,
			IntegrityChecksEnabled: false,
			PrefetchEnabled:        false,
		}

		// Validate the custom config
		require.NoError(t, customConfig.Validate())

		// Set the config using standard API
		err := config.SetConfig(cfg, assets.ConfigParser{}, customConfig)
		require.NoError(t, err)

		// Get the config back using standard API
		retrievedConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
		require.NoError(t, err)

		// Verify the values match
		assert.Equal(t, customConfig.RepositoryURL, retrievedConfig.RepositoryURL)
		assert.Equal(t, customConfig.Branch, retrievedConfig.Branch)
		assert.Equal(t, customConfig.CachePath, retrievedConfig.CachePath)
		assert.Equal(t, customConfig.CacheSizeMB, retrievedConfig.CacheSizeMB)
	})
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	t.Run("Invalid Asset Config", func(t *testing.T) {
		invalidConfig := assets.AssetConfig{
			// Missing required fields
			RepositoryURL: "", // Required field
			Branch:        "",
		}

		// Should fail validation
		err := invalidConfig.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository_url")
	})

	t.Run("Invalid Auth Config", func(t *testing.T) {
		invalidConfig := auth.Config{
			StorageType:       "invalid_type", // Invalid storage type
			ValidationTimeout: -1,             // Invalid timeout
		}

		// Should fail validation
		err := invalidConfig.Validate()
		assert.Error(t, err)
	})

	t.Run("Invalid Cache Config", func(t *testing.T) {
		invalidConfig := cache.Config{
			BasePath:    "", // Required field
			SizeLimitMB: -1, // Invalid size
		}

		// Should fail validation
		err := invalidConfig.Validate()
		assert.Error(t, err)
	})
}
