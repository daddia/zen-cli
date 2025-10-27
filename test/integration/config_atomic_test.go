//go:build integration

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigAtomicUpdates tests that configuration updates are atomic
func TestConfigAtomicUpdates(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to temp directory
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	// Create .zen directory
	zenDir := filepath.Join(tempDir, ".zen")
	err = os.MkdirAll(zenDir, 0700)
	require.NoError(t, err)

	t.Run("Atomic Config File Updates", func(t *testing.T) {
		// Load default configuration
		cfg := config.LoadDefaults()
		require.NotNil(t, cfg)

		// Get initial asset config
		assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
		require.NoError(t, err)

		// Modify the config
		assetConfig.Branch = "atomic-test-branch"
		assetConfig.RepositoryURL = "https://github.com/test/atomic-repo.git"

		// Set the config (this should be atomic)
		err = config.SetConfig(cfg, assets.ConfigParser{}, assetConfig)
		require.NoError(t, err)

		// Verify the config file exists and contains the expected data
		configPath := filepath.Join(zenDir, "config")
		assert.FileExists(t, configPath)

		// Verify no temporary files are left behind
		tempFiles, err := filepath.Glob(filepath.Join(zenDir, "config.*.tmp"))
		require.NoError(t, err)
		assert.Empty(t, tempFiles, "No temporary files should be left behind")

		// Verify the configuration can be read back correctly
		newCfg, err := config.Load()
		require.NoError(t, err)

		newAssetConfig, err := config.GetConfig(newCfg, assets.ConfigParser{})
		require.NoError(t, err)

		assert.Equal(t, "atomic-test-branch", newAssetConfig.Branch)
		assert.Equal(t, "https://github.com/test/atomic-repo.git", newAssetConfig.RepositoryURL)
	})

	t.Run("Config File Permissions", func(t *testing.T) {
		configPath := filepath.Join(zenDir, "config")

		// Check that config file has secure permissions
		info, err := os.Stat(configPath)
		require.NoError(t, err)

		// Config file should be readable/writable by owner only (0600)
		mode := info.Mode()
		assert.Equal(t, os.FileMode(0600), mode.Perm(), "Config file should have 0600 permissions")
	})

	t.Run("Concurrent Config Updates", func(t *testing.T) {
		// This test ensures that concurrent updates don't corrupt the config
		cfg := config.LoadDefaults()
		require.NotNil(t, cfg)

		// Run multiple concurrent updates
		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Get current config
				assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
				if err != nil {
					t.Errorf("Failed to get config in goroutine %d: %v", id, err)
					return
				}

				// Modify config with unique values
				assetConfig.Branch = fmt.Sprintf("concurrent-branch-%d", id)

				// Small delay to increase chance of race condition
				time.Sleep(time.Millisecond * 10)

				// Set config
				err = config.SetConfig(cfg, assets.ConfigParser{}, assetConfig)
				if err != nil {
					t.Errorf("Failed to set config in goroutine %d: %v", id, err)
					return
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 3; i++ {
			<-done
		}

		// Verify config file is still valid and readable
		finalCfg, err := config.Load()
		require.NoError(t, err)

		finalAssetConfig, err := config.GetConfig(finalCfg, assets.ConfigParser{})
		require.NoError(t, err)

		// Should have one of the concurrent branch names
		assert.Contains(t, finalAssetConfig.Branch, "concurrent-branch-")

		// Verify no temporary files are left behind
		tempFiles, err := filepath.Glob(filepath.Join(zenDir, "config.*.tmp"))
		require.NoError(t, err)
		assert.Empty(t, tempFiles, "No temporary files should be left behind after concurrent updates")
	})
}
