package factory

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	f := New()

	assert.NotNil(t, f)
	assert.Equal(t, "zen", f.ExecutableName)
	assert.Equal(t, "dev", f.AppVersion)
	assert.NotNil(t, f.IOStreams)
	assert.NotNil(t, f.Logger)
	assert.NotNil(t, f.Config)
	assert.NotNil(t, f.WorkspaceManager)
	assert.NotNil(t, f.AgentManager)
}

func TestConfigFunc(t *testing.T) {
	configFn := configFunc()

	// First call loads config
	cfg1, err1 := configFn()
	// Config may or may not exist, so we don't assert on error

	// Second call returns cached result
	cfg2, err2 := configFn()

	if err1 == nil {
		assert.Equal(t, cfg1, cfg2)
	}
	assert.Equal(t, err1, err2)
}

func TestWorkspaceManager(t *testing.T) {
	f := New()

	wm, err := f.WorkspaceManager()
	require.NoError(t, err)
	require.NotNil(t, wm)

	// Test workspace manager methods
	assert.NotEmpty(t, wm.Root())
	assert.NotEmpty(t, wm.ConfigFile())

	status, err := wm.Status()
	assert.NoError(t, err)
	// Workspace may not be initialized in test environment
	assert.NotEmpty(t, status.Root)
}

func TestAgentManager(t *testing.T) {
	f := New()

	am, err := f.AgentManager()
	require.NoError(t, err)
	require.NotNil(t, am)

	// Test agent manager methods
	agents, err := am.List()
	assert.NoError(t, err)
	assert.NotNil(t, agents)

	result, err := am.Execute("test", nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// Test dependency injection and caching
func TestFactoryDependencyCaching(t *testing.T) {
	f := New()

	// Test config caching
	cfg1, err1 := f.Config()
	cfg2, err2 := f.Config()

	// Should return the same instance (cached)
	if err1 == nil && err2 == nil {
		assert.Same(t, cfg1, cfg2, "Config should be cached")
	}
	assert.Equal(t, err1, err2, "Errors should be consistent")

	// Test workspace manager caching behavior
	wm1, err1 := f.WorkspaceManager()
	wm2, err2 := f.WorkspaceManager()

	// Each call creates a new instance but with same configuration
	if err1 == nil && err2 == nil {
		assert.NotSame(t, wm1, wm2, "WorkspaceManager creates new instances")
		assert.Equal(t, wm1.Root(), wm2.Root(), "But they should have same root")
	}
}

// Test IOStreams configuration
func TestIOStreamsConfiguration(t *testing.T) {
	// Test with NO_COLOR environment variable
	t.Run("NO_COLOR environment variable", func(t *testing.T) {
		oldValue := os.Getenv("NO_COLOR")
		defer func() {
			if oldValue == "" {
				os.Unsetenv("NO_COLOR")
			} else {
				os.Setenv("NO_COLOR", oldValue)
			}
		}()

		os.Setenv("NO_COLOR", "1")
		f := New()
		io := f.IOStreams

		assert.False(t, io.ColorEnabled(), "Color should be disabled when NO_COLOR is set")
	})

	// Test with ZEN_PROMPT_DISABLED
	t.Run("ZEN_PROMPT_DISABLED environment variable", func(t *testing.T) {
		oldValue := os.Getenv("ZEN_PROMPT_DISABLED")
		defer func() {
			if oldValue == "" {
				os.Unsetenv("ZEN_PROMPT_DISABLED")
			} else {
				os.Setenv("ZEN_PROMPT_DISABLED", oldValue)
			}
		}()

		os.Setenv("ZEN_PROMPT_DISABLED", "1")
		f := New()
		io := f.IOStreams

		assert.False(t, io.CanPrompt(), "Prompting should be disabled when ZEN_PROMPT_DISABLED is set")
	})
}

// Test logger configuration
func TestLoggerConfiguration(t *testing.T) {
	f := New()

	logger := f.Logger
	assert.NotNil(t, logger)

	// Test that logger is properly configured
	// This is a basic test since the logger interface is minimal
	assert.NotPanics(t, func() {
		// Logger should not panic on basic operations
		logger.Debug("test debug message")
		logger.Info("test info message")
	})
}

// Test workspace manager functionality
func TestWorkspaceManagerFunctionality(t *testing.T) {
	f := New()

	wm, err := f.WorkspaceManager()
	require.NoError(t, err)
	require.NotNil(t, wm)

	// Test workspace manager interface
	assert.NotEmpty(t, wm.Root())
	assert.NotEmpty(t, wm.ConfigFile())

	// Test status method
	status, err := wm.Status()
	assert.NoError(t, err)
	assert.NotEmpty(t, status.Root)
}

// Test auth manager functionality
func TestAuthManagerFunctionality(t *testing.T) {
	f := New()

	am, err := f.AuthManager()
	// May fail if auth dependencies are not available in test environment
	if err != nil {
		t.Skipf("Auth manager not available in test environment: %v", err)
		return
	}

	require.NotNil(t, am)

	// Test basic auth manager interface
	providers := am.ListProviders()
	assert.NotNil(t, providers)

	// Test provider info
	if len(providers) > 0 {
		info, err := am.GetProviderInfo(providers[0])
		assert.NoError(t, err)
		assert.NotNil(t, info)
	}
}

// Test asset client functionality
func TestAssetClientFunctionality(t *testing.T) {
	// Use test factory which provides mocked asset client
	f := cmdutil.NewTestFactory(nil)

	client, err := f.AssetClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx := context.Background()

	// Test cache info (should work with mock)
	cacheInfo, err := client.GetCacheInfo(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, cacheInfo)

	// Test list assets (should work with mock)
	assetList, err := client.ListAssets(ctx, assets.AssetFilter{})
	assert.NoError(t, err)
	assert.NotNil(t, assetList)
	assert.Equal(t, 1, assetList.Total) // Mock returns 1 asset
}

// Test cache functionality
func TestCacheManagerFunctionality(t *testing.T) {
	f := New()

	tempDir := t.TempDir()
	cacheManager := f.Cache(tempDir)
	require.NotNil(t, cacheManager)

	// Test basic cache operations
	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test put and get
	err := cacheManager.Put(ctx, key, value, cache.PutOptions{TTL: time.Hour})
	assert.NoError(t, err)

	entry, err := cacheManager.Get(ctx, key)
	assert.NoError(t, err)
	if entry != nil {
		assert.Equal(t, value, entry.Data)
	} else {
		t.Error("Expected cache entry but got nil")
	}

	// Test delete
	err = cacheManager.Delete(ctx, key)
	assert.NoError(t, err)

	// After deletion, Get should return an error (key not found)
	entry, err = cacheManager.Get(ctx, key)
	assert.Error(t, err, "Should get error for deleted key")
	assert.Nil(t, entry, "Entry should be nil after deletion")
}

// Test ConfigWithCommand function
func TestConfigWithCommand(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	configFunc := ConfigWithCommand(cmd)
	require.NotNil(t, configFunc)

	// Test that it returns a config
	cfg1, err1 := configFunc()
	cfg2, err2 := configFunc()

	// Should cache the result
	if err1 == nil && err2 == nil {
		assert.Same(t, cfg1, cfg2)
	}
	assert.Equal(t, err1, err2)
}

// Test build information
func TestBuildInfo(t *testing.T) {
	f := New()

	buildInfo := f.BuildInfo
	require.NotNil(t, buildInfo)

	// Test that required fields are present
	assert.Contains(t, buildInfo, "version")
	assert.Contains(t, buildInfo, "commit")
	assert.Contains(t, buildInfo, "build_time")

	// Test GetBuildInfo function
	globalBuildInfo := GetBuildInfo()
	assert.Equal(t, buildInfo, globalBuildInfo)
}

// Test version information
func TestVersionInfo(t *testing.T) {
	f := New()

	assert.Equal(t, "dev", f.AppVersion)
	assert.Equal(t, "zen", f.ExecutableName)

	// Test getVersion function
	version := getVersion()
	assert.Equal(t, f.AppVersion, version)
}

// Test asset configuration
func TestGetAssetConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *config.Config
		setupEnv    func()
		cleanupEnv  func()
		validate    func(t *testing.T, cfg assets.AssetConfig)
	}{
		{
			name: "default configuration",
			setupConfig: func() *config.Config {
				return config.LoadDefaults()
			},
			setupEnv:   func() {},
			cleanupEnv: func() {},
			validate: func(t *testing.T, cfg assets.AssetConfig) {
				assert.NotEmpty(t, cfg.RepositoryURL)
				assert.NotEmpty(t, cfg.Branch)
				assert.NotEmpty(t, cfg.CachePath)
				assert.Greater(t, cfg.CacheSizeMB, int64(0))
			},
		},
		{
			name: "environment variable overrides",
			setupConfig: func() *config.Config {
				return config.LoadDefaults()
			},
			setupEnv: func() {
				os.Setenv("ZEN_ASSET_REPOSITORY_URL", "https://example.com/test-repo")
				os.Setenv("ZEN_AUTH_PROVIDER", "test-provider")
				os.Setenv("ZEN_ASSET_BRANCH", "test-branch")
			},
			cleanupEnv: func() {
				os.Unsetenv("ZEN_ASSET_REPOSITORY_URL")
				os.Unsetenv("ZEN_AUTH_PROVIDER")
				os.Unsetenv("ZEN_ASSET_BRANCH")
			},
			validate: func(t *testing.T, cfg assets.AssetConfig) {
				assert.Equal(t, "https://example.com/test-repo", cfg.RepositoryURL)
				assert.Equal(t, "test-provider", cfg.AuthProvider)
				assert.Equal(t, "test-branch", cfg.Branch)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			cfg := tt.setupConfig()
			assetConfig := getAssetConfig(cfg)

			tt.validate(t, assetConfig)
		})
	}
}

// Test auth configuration
func TestGetAuthConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *config.Config
		setupEnv    func()
		cleanupEnv  func()
		validate    func(t *testing.T, cfg auth.Config)
	}{
		{
			name: "default configuration",
			setupConfig: func() *config.Config {
				return config.LoadDefaults()
			},
			setupEnv:   func() {},
			cleanupEnv: func() {},
			validate: func(t *testing.T, cfg auth.Config) {
				assert.NotEmpty(t, cfg.StorageType)
				assert.NotEmpty(t, cfg.StoragePath)
			},
		},
		{
			name: "github provider configuration",
			setupConfig: func() *config.Config {
				cfg := config.LoadDefaults()
				cfg.Assets.AuthProvider = "github"
				return cfg
			},
			setupEnv: func() {
				// Clear any existing environment variable that might override
				os.Unsetenv("ZEN_AUTH_STORAGE_TYPE")
			},
			cleanupEnv: func() {},
			validate: func(t *testing.T, cfg auth.Config) {
				assert.Equal(t, "keychain", cfg.StorageType)
			},
		},
		{
			name: "environment variable overrides",
			setupConfig: func() *config.Config {
				return config.LoadDefaults()
			},
			setupEnv: func() {
				os.Setenv("ZEN_AUTH_STORAGE_TYPE", "file")
				os.Setenv("ZEN_AUTH_STORAGE_PATH", "/tmp/test-auth")
				os.Setenv("ZEN_AUTH_ENCRYPTION_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("ZEN_AUTH_STORAGE_TYPE")
				os.Unsetenv("ZEN_AUTH_STORAGE_PATH")
				os.Unsetenv("ZEN_AUTH_ENCRYPTION_KEY")
			},
			validate: func(t *testing.T, cfg auth.Config) {
				assert.Equal(t, "file", cfg.StorageType)
				assert.Equal(t, "/tmp/test-auth", cfg.StoragePath)
				assert.Equal(t, "test-key", cfg.EncryptionKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			cfg := tt.setupConfig()
			authConfig := getAuthConfig(cfg)

			tt.validate(t, authConfig)
		})
	}
}

// Test workspace manager implementation
func TestWorkspaceManagerImplementation(t *testing.T) {
	f := New()

	wm, err := f.WorkspaceManager()
	require.NoError(t, err)

	// Cast to concrete type to test implementation details
	impl, ok := wm.(*workspaceManager)
	require.True(t, ok, "WorkspaceManager should be of type *workspaceManager")

	// Test that manager is lazily initialized
	assert.Nil(t, impl.manager, "Manager should be nil initially")

	// Test initialization
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	err = wm.Initialize()
	// May fail in test environment, but should not panic
	if err == nil {
		assert.NotNil(t, impl.manager, "Manager should be initialized after Initialize() call")
	}
}

// Test agent manager implementation
func TestAgentManagerImplementation(t *testing.T) {
	f := New()

	am, err := f.AgentManager()
	require.NoError(t, err)

	// Cast to concrete type to test implementation details
	impl, ok := am.(*agentManager)
	require.True(t, ok, "AgentManager should be of type *agentManager")

	// Test placeholder implementation
	agents, err := impl.List()
	assert.NoError(t, err)
	assert.Empty(t, agents, "Placeholder should return empty list")

	result, err := impl.Execute("test", "input")
	assert.NoError(t, err)
	assert.Nil(t, result, "Placeholder should return nil")
}

// Test error handling in factory
func TestFactoryErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		setupFactory func() *cmdutil.Factory
		testFunc     func(t *testing.T, f *cmdutil.Factory)
	}{
		{
			name: "config error propagation",
			setupFactory: func() *cmdutil.Factory {
				f := &cmdutil.Factory{
					Config: func() (*config.Config, error) {
						return nil, errors.New("config error")
					},
				}
				return f
			},
			testFunc: func(t *testing.T, f *cmdutil.Factory) {
				_, err := f.Config()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "config error")
			},
		},
		{
			name: "workspace manager with config error",
			setupFactory: func() *cmdutil.Factory {
				f := New()
				f.Config = func() (*config.Config, error) {
					return nil, errors.New("config error")
				}
				return f
			},
			testFunc: func(t *testing.T, f *cmdutil.Factory) {
				_, err := f.WorkspaceManager()
				assert.Error(t, err)
			},
		},
		{
			name: "auth manager with config error",
			setupFactory: func() *cmdutil.Factory {
				f := New()
				f.Config = func() (*config.Config, error) {
					return nil, errors.New("config error")
				}
				return f
			},
			testFunc: func(t *testing.T, f *cmdutil.Factory) {
				_, err := f.AuthManager()
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.setupFactory()
			tt.testFunc(t, f)
		})
	}
}

// Test concurrent access to factory
func TestFactoryConcurrency(t *testing.T) {
	f := New()
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	// Test concurrent access to cached dependencies
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					done <- errors.New("panic occurred")
				}
			}()

			// Access various factory methods concurrently
			_, err1 := f.Config()
			_, err2 := f.WorkspaceManager()
			_, err3 := f.AgentManager()

			// Some may fail in test environment, but shouldn't panic
			if err1 != nil || err2 != nil || err3 != nil {
				done <- nil // Expected in test environment
			} else {
				done <- nil
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			if err != nil && strings.Contains(err.Error(), "panic") {
				t.Errorf("Goroutine %d panicked: %v", i, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}

// Test cache path expansion
func TestCachePathExpansion(t *testing.T) {
	// Test home directory expansion in asset config
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)

	cfg := &config.Config{}
	cfg.Assets.CachePath = "~/.zen/cache"

	assetConfig := getAssetConfig(cfg)

	expectedPath := filepath.Join(tempHome, ".zen", "cache")
	assert.Equal(t, expectedPath, assetConfig.CachePath)
}

// Benchmark tests
func BenchmarkFactoryNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := New()
		if f == nil {
			b.Fatal("factory is nil")
		}
	}
}

func BenchmarkFactoryConfigAccess(b *testing.B) {
	f := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := f.Config()
		// Ignore error - expected in some test environments
		_ = err
	}
}

func BenchmarkFactoryWorkspaceManagerAccess(b *testing.B) {
	f := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := f.WorkspaceManager()
		// Ignore error - expected in some test environments
		_ = err
	}
}

func BenchmarkCacheOperations(b *testing.B) {
	f := New()
	tempDir := b.TempDir()
	cacheManager := f.Cache(tempDir)

	b.ResetTimer()

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		key := "benchmark-key"
		value := "benchmark-value"

		cacheManager.Put(ctx, key, value, cache.PutOptions{TTL: time.Hour})
		cacheManager.Get(ctx, key)
		cacheManager.Delete(ctx, key)
	}
}
