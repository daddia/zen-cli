//go:build integration

package integration_test

import (
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/development"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/cli"
	"github.com/daddia/zen/pkg/task"
	"github.com/daddia/zen/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigSystemIntegration tests the complete config system integration
func TestConfigSystemIntegration(t *testing.T) {
	// Load configuration
	cfg := config.LoadDefaults()
	require.NotNil(t, cfg)

	t.Run("All Components Have Working Config", func(t *testing.T) {
		// Test all component configs can be loaded and validated
		components := []struct {
			name   string
			parser interface{}
		}{
			{"assets", assets.ConfigParser{}},
			{"auth", auth.ConfigParser{}},
			{"cache", cache.ConfigParser{}},
			{"cli", cli.ConfigParser{}},
			{"development", development.ConfigParser{}},
			{"task", task.ConfigParser{}},
			{"templates", template.ConfigParser{}},
			{"workspace", workspace.ConfigParser{}},
		}

		for _, comp := range components {
			t.Run(comp.name, func(t *testing.T) {
				switch parser := comp.parser.(type) {
				case assets.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case auth.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case cache.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case cli.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case development.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case task.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case template.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				case workspace.ConfigParser:
					config, err := config.GetConfig(cfg, parser)
					require.NoError(t, err)
					require.NoError(t, config.Validate())
					assert.NotNil(t, config.Defaults())
				}
			})
		}
	})

	t.Run("Config Set and Get Round Trip", func(t *testing.T) {
		// Test setting and getting component configs
		t.Run("Assets Config", func(t *testing.T) {
			// Get current config
			assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
			require.NoError(t, err)

			// Modify a field
			assetConfig.Branch = "test-branch"

			// Set the config back
			err = config.SetConfig(cfg, assets.ConfigParser{}, assetConfig)
			require.NoError(t, err)

			// Get the config again and verify
			updatedConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
			require.NoError(t, err)
			assert.Equal(t, "test-branch", updatedConfig.Branch)
		})
	})
}

// TestConfigArchitectureCompliance tests that the architecture rules are followed
func TestConfigArchitectureCompliance(t *testing.T) {
	t.Run("Only Core Config Contains File Operations", func(t *testing.T) {
		// This test would ideally use static analysis to verify
		// that only internal/config package imports viper
		// For now, we test that the APIs work correctly

		cfg := config.LoadDefaults()
		require.NotNil(t, cfg)

		// Verify config file operations work
		assert.NotEmpty(t, cfg.GetLoadedSources())
	})

	t.Run("All Components Implement Standard Interfaces", func(t *testing.T) {
		// Test that all parsers implement the required methods
		parsers := []interface{}{
			assets.ConfigParser{},
			auth.ConfigParser{},
			cache.ConfigParser{},
			cli.ConfigParser{},
			development.ConfigParser{},
			task.ConfigParser{},
			template.ConfigParser{},
			workspace.ConfigParser{},
		}

		for _, parser := range parsers {
			// Each parser should have Section() method
			switch p := parser.(type) {
			case assets.ConfigParser:
				assert.Equal(t, "assets", p.Section())
			case auth.ConfigParser:
				assert.Equal(t, "auth", p.Section())
			case cache.ConfigParser:
				assert.Equal(t, "cache", p.Section())
			case cli.ConfigParser:
				assert.Equal(t, "cli", p.Section())
			case development.ConfigParser:
				assert.Equal(t, "development", p.Section())
			case task.ConfigParser:
				assert.Equal(t, "task", p.Section())
			case template.ConfigParser:
				assert.Equal(t, "templates", p.Section())
			case workspace.ConfigParser:
				assert.Equal(t, "workspace", p.Section())
			}
		}
	})
}

// TestConfigPerformance tests that performance requirements are met
func TestConfigPerformance(t *testing.T) {
	t.Run("Config Loading Performance", func(t *testing.T) {
		// Test that config loading is fast enough
		for i := 0; i < 100; i++ {
			cfg, err := config.Load()
			require.NoError(t, err)
			require.NotNil(t, cfg)
		}
		// If this completes without timeout, performance is acceptable
	})

	t.Run("Component Config Parsing Performance", func(t *testing.T) {
		cfg := config.LoadDefaults()
		require.NotNil(t, cfg)

		// Test parsing performance for each component
		for i := 0; i < 100; i++ {
			_, err := config.GetConfig(cfg, assets.ConfigParser{})
			require.NoError(t, err)

			_, err = config.GetConfig(cfg, workspace.ConfigParser{})
			require.NoError(t, err)

			_, err = config.GetConfig(cfg, cli.ConfigParser{})
			require.NoError(t, err)
		}
		// If this completes without timeout, performance is acceptable
	})
}
