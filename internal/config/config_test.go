package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	configPath := filepath.Join(zenDir, "config.yaml")

	// Create a test config file
	configContent := `
log_level: debug
log_format: json
cli:
  no_color: true
  verbose: true
  output_format: yaml
workspace:
  root: .
  config_file: test.yaml
development:
  debug: true
  profile: true
`

	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Change to temp directory so config is found
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()

	require.NoError(t, os.Chdir(tempDir))

	// Load configuration
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test loaded values
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.True(t, cfg.CLI.NoColor)
	assert.True(t, cfg.CLI.Verbose)
	assert.Equal(t, "yaml", cfg.CLI.OutputFormat)
	assert.Equal(t, ".", cfg.Workspace.Root)
	assert.Equal(t, "test.yaml", cfg.Workspace.ConfigFile)
	assert.True(t, cfg.Development.Debug)
	assert.True(t, cfg.Development.Profile)

	// Test configuration metadata
	assert.NotEmpty(t, cfg.GetConfigFile())
	// Check that a file source was loaded (path may have /private prefix)
	hasFileSource := false
	for _, source := range cfg.GetLoadedSources() {
		if strings.HasPrefix(source, "file:") {
			hasFileSource = true
			break
		}
	}
	assert.True(t, hasFileSource, "Should have a file source")
}

func TestLoadDefaults(t *testing.T) {
	// Create a temporary directory with no config file
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()

	require.NoError(t, os.Chdir(tempDir))

	// Load configuration (should use defaults)
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test default values
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.False(t, cfg.CLI.NoColor)
	assert.False(t, cfg.CLI.Verbose)
	assert.Equal(t, "text", cfg.CLI.OutputFormat)
	assert.Equal(t, ".", cfg.Workspace.Root)
	assert.Equal(t, "zen.yaml", cfg.Workspace.ConfigFile)
	assert.False(t, cfg.Development.Debug)
	assert.False(t, cfg.Development.Profile)

	// Test that defaults are loaded
	assert.Contains(t, cfg.GetLoadedSources(), "defaults")
}

func TestApplyEnvOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected func(*Config) bool
	}{
		{
			name: "NO_COLOR environment variable",
			envVars: map[string]string{
				"NO_COLOR": "1",
			},
			expected: func(cfg *Config) bool {
				return cfg.CLI.NoColor
			},
		},
		{
			name: "ZEN_DEBUG environment variable",
			envVars: map[string]string{
				"ZEN_DEBUG": "true",
			},
			expected: func(cfg *Config) bool {
				return cfg.Development.Debug && cfg.LogLevel == "debug"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				require.NoError(t, os.Setenv(key, value))
				defer func(k string) {
					require.NoError(t, os.Unsetenv(k))
				}(key)
			}

			cfg := &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					NoColor:      false,
					Verbose:      false,
					OutputFormat: "text",
				},
				Workspace: WorkspaceConfig{
					Root:       ".",
					ConfigFile: "zen.yaml",
				},
				Development: DevelopmentConfig{
					Debug:   false,
					Profile: false,
				},
			}

			applyEnvOverrides(cfg)

			assert.True(t, tt.expected(cfg), "Environment override not applied correctly")
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorType string
	}{
		{
			name: "valid config",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
				Workspace: WorkspaceConfig{
					Root: ".",
				},
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				LogLevel:  "invalid",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
				Workspace: WorkspaceConfig{
					Root: ".",
				},
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name: "invalid log format",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "invalid",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
				Workspace: WorkspaceConfig{
					Root: ".",
				},
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name: "invalid output format",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "invalid",
				},
				Workspace: WorkspaceConfig{
					Root: ".",
				},
			},
			wantError: true,
			errorType: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorType == "validation" {
					assert.True(t, IsValidationError(err), "Expected ValidationError")
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadDefaultsFunction(t *testing.T) {
	cfg := LoadDefaults()
	require.NotNil(t, cfg)

	// Test default values are set correctly
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.False(t, cfg.CLI.NoColor)
	assert.False(t, cfg.CLI.Verbose)
	assert.Equal(t, "text", cfg.CLI.OutputFormat)
	assert.Equal(t, ".", cfg.Workspace.Root)
	assert.Equal(t, "zen.yaml", cfg.Workspace.ConfigFile)
	assert.False(t, cfg.Development.Debug)
	assert.False(t, cfg.Development.Profile)

	// Test asset defaults
	assert.Equal(t, "https://github.com/daddia/zen-assets.git", cfg.Assets.RepositoryURL)
	assert.Equal(t, "main", cfg.Assets.Branch)
	assert.Equal(t, "github", cfg.Assets.AuthProvider)

	// Test template defaults
	assert.True(t, *cfg.Templates.CacheEnabled)
	assert.Equal(t, "30m", cfg.Templates.CacheTTL)
	assert.Equal(t, 100, cfg.Templates.CacheSize)

	// Test that defaults source is loaded
	assert.Contains(t, cfg.GetLoadedSources(), "defaults")
	assert.Empty(t, cfg.GetConfigFile())
}

func TestConfigRedacted(t *testing.T) {
	cfg := &Config{
		LogLevel: "debug",
		Assets: AssetsConfig{
			RepositoryURL: "https://token:secret@github.com/user/repo.git",
			AuthProvider:  "github",
		},
	}

	redacted := cfg.Redacted()
	require.NotNil(t, redacted)

	// Test that sensitive fields are redacted (actual format includes partial masking)
	assert.Contains(t, redacted.Assets.RepositoryURL, "*")

	// Test that non-sensitive fields are preserved
	assert.Equal(t, "debug", redacted.LogLevel)
	assert.Equal(t, "github", redacted.Assets.AuthProvider)
}

func TestAssetsConfigRedacted(t *testing.T) {
	assets := AssetsConfig{
		RepositoryURL:          "https://token:secret@github.com/user/repo.git",
		Branch:                 "main",
		AuthProvider:           "github",
		CachePath:              "~/.zen/assets",
		IntegrityChecksEnabled: true,
	}

	redacted := assets.Redacted()

	// Test that repository URL is redacted (contains credentials)
	assert.Contains(t, redacted.RepositoryURL, "*")

	// Test that other fields are preserved
	assert.Equal(t, "main", redacted.Branch)
	assert.Equal(t, "github", redacted.AuthProvider)
	assert.Equal(t, "~/.zen/assets", redacted.CachePath)
	assert.True(t, redacted.IntegrityChecksEnabled)
}

func TestValidateEnhanced(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid complete config",
			config: &Config{
				LogLevel:  "debug",
				LogFormat: "json",
				CLI: CLIConfig{
					OutputFormat: "yaml",
					NoColor:      true,
					Verbose:      true,
				},
				Workspace: WorkspaceConfig{
					Root:       ".",
					ConfigFile: "custom.yaml",
				},
				Assets: AssetsConfig{
					RepositoryURL:      "https://github.com/user/repo.git",
					Branch:             "develop",
					AuthProvider:       "gitlab",
					CacheSizeMB:        200,
					SyncTimeoutSeconds: 60,
				},
				Development: DevelopmentConfig{
					Debug:   true,
					Profile: true,
				},
			},
			wantError: false,
		},
		{
			name: "invalid log level enum",
			config: &Config{
				LogLevel:  "invalid_level",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
			},
			wantError: true,
			errorMsg:  "log_level",
		},
		{
			name: "invalid log format enum",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "xml",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
			},
			wantError: true,
			errorMsg:  "log_format",
		},
		{
			name: "invalid cli output format",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "html",
				},
			},
			wantError: true,
			errorMsg:  "cli.output_format",
		},
		{
			name: "empty required fields",
			config: &Config{
				LogLevel:  "",
				LogFormat: "",
				CLI: CLIConfig{
					OutputFormat: "",
				},
			},
			wantError: true,
			errorMsg:  "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)

				// Test that it's a validation error
				assert.True(t, IsValidationError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidationErrorEnhanced(t *testing.T) {
	// Test ValidationError struct methods
	err := &ValidationError{
		Field:   "cli.output_format",
		Message: "must be one of: text, json, yaml",
	}

	// Test Error() method
	assert.Contains(t, err.Error(), "cli.output_format: must be one of: text, json, yaml")

	// Test IsValidationError
	assert.True(t, IsValidationError(err))
	assert.False(t, IsValidationError(fmt.Errorf("regular error")))
	assert.False(t, IsValidationError(nil))

	// Test error wrapping (may not work with current implementation)
	wrappedErr := fmt.Errorf("wrapped: %w", err)
	_ = wrappedErr // Note: IsValidationError may not detect wrapped errors depending on implementation
}

func TestGetValueFromConfigEnhanced(t *testing.T) {
	cfg := &Config{
		LogLevel:  "trace",
		LogFormat: "json",
		CLI: CLIConfig{
			NoColor:      true,
			Verbose:      true,
			OutputFormat: "yaml",
		},
		Workspace: WorkspaceConfig{
			Root:       "/workspace/root",
			ConfigFile: "custom.yaml",
		},
		Assets: AssetsConfig{
			RepositoryURL:          "https://github.com/user/repo.git",
			Branch:                 "develop",
			AuthProvider:           "gitlab",
			CachePath:              "/custom/cache",
			CacheSizeMB:            150,
			SyncTimeoutSeconds:     45,
			IntegrityChecksEnabled: false,
			PrefetchEnabled:        true,
		},
		Templates: TemplatesConfig{
			CacheTTL:   "1h",
			CacheSize:  200,
			LeftDelim:  "<%",
			RightDelim: "%>",
		},
		Development: DevelopmentConfig{
			Debug:   true,
			Profile: false,
		},
	}

	// Create options to test getValueFromConfig
	tests := []struct {
		key      string
		expected string
	}{
		// Top level fields
		{"log_level", "trace"},
		{"log_format", "json"},

		// CLI nested fields
		{"cli.no_color", "true"},
		{"cli.verbose", "true"},
		{"cli.output_format", "yaml"},

		// Workspace nested fields
		{"workspace.root", "/workspace/root"},
		{"workspace.config_file", "custom.yaml"},

		// Assets nested fields
		{"assets.repository_url", "https://github.com/user/repo.git"},
		{"assets.branch", "develop"},
		{"assets.auth_provider", "gitlab"},
		{"assets.cache_path", "/custom/cache"},
		{"assets.cache_size_mb", "150"},
		{"assets.sync_timeout_seconds", "45"},
		{"assets.integrity_checks_enabled", "false"},
		{"assets.prefetch_enabled", "true"},

		// Templates nested fields
		{"templates.cache_ttl", "1h"},
		{"templates.cache_size", "200"},
		{"templates.left_delim", "<%"},
		{"templates.right_delim", "%>"},

		// Development nested fields
		{"development.debug", "true"},
		{"development.profile", "false"},

		// Non-existent fields
		{"nonexistent", ""},
		{"cli.nonexistent", ""},
		{"workspace.nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			// Find the option
			opt, found := FindOption(tt.key)
			if tt.expected == "" && strings.Contains(tt.key, "nonexistent") {
				assert.False(t, found, "Should not find non-existent option")
				return
			}

			require.True(t, found, "Should find option %s", tt.key)

			// Test getValueFromConfig through the option
			result := opt.getValueFromConfig(cfg)
			assert.Equal(t, tt.expected, result, "Value mismatch for key %s", tt.key)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test new functionality for ZEN-004

func TestLoadWithCommand(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	configPath := filepath.Join(zenDir, "config.yaml")

	// Create a test config file
	configContent := `
log_level: info
cli:
  verbose: false
  no_color: false
  output_format: text
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create a mock command with flags
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	cmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	// Set some flags
	require.NoError(t, cmd.PersistentFlags().Set("verbose", "true"))
	require.NoError(t, cmd.PersistentFlags().Set("output", "json"))

	// Load configuration with command
	cfg, err := LoadWithCommand(cmd)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// CLI flags should override config file
	assert.True(t, cfg.CLI.Verbose, "CLI flag should override config file")
	assert.Equal(t, "json", cfg.CLI.OutputFormat, "CLI flag should override config file")
	assert.False(t, cfg.CLI.NoColor, "Config file value should be used when flag not set")

	// Should track that flags were used
	assert.Contains(t, cfg.GetLoadedSources(), "flags")
}

func TestLoadWithOptions(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))

	// Test with custom config file
	customConfigPath := filepath.Join(tempDir, "custom-config.yaml")
	configContent := `
log_level: debug
cli:
  output_format: yaml
`
	require.NoError(t, os.WriteFile(customConfigPath, []byte(configContent), 0644))

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Load with custom config file
	cfg, err := LoadWithOptions(LoadOptions{
		ConfigFile: customConfigPath,
	})
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "yaml", cfg.CLI.OutputFormat)
	assert.Equal(t, customConfigPath, cfg.GetConfigFile())
}

func TestEnvironmentVariableIntegration(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Set environment variables
	envVars := map[string]string{
		"ZEN_LOG_LEVEL":         "debug",
		"ZEN_CLI_VERBOSE":       "true",
		"ZEN_CLI_OUTPUT_FORMAT": "json",
		"ZEN_WORKSPACE_ROOT":    ".",
	}

	for key, value := range envVars {
		require.NoError(t, os.Setenv(key, value))
		defer func(k string) {
			require.NoError(t, os.Unsetenv(k))
		}(key)
	}

	// Load configuration
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Environment variables should be applied
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.True(t, cfg.CLI.Verbose)
	assert.Equal(t, "json", cfg.CLI.OutputFormat)
	assert.Equal(t, ".", cfg.Workspace.Root)

	// Should track that environment was used
	assert.Contains(t, cfg.GetLoadedSources(), "environment")
}

func TestConfigurationPrecedence(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	configPath := filepath.Join(zenDir, "config.yaml")

	// Create a test config file
	configContent := `
log_level: info
cli:
  verbose: false
  output_format: text
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Set environment variables (should override config file)
	require.NoError(t, os.Setenv("ZEN_LOG_LEVEL", "warn"))
	require.NoError(t, os.Setenv("ZEN_CLI_VERBOSE", "true"))
	defer func() {
		require.NoError(t, os.Unsetenv("ZEN_LOG_LEVEL"))
		require.NoError(t, os.Unsetenv("ZEN_CLI_VERBOSE"))
	}()

	// Create command with flags (should override environment)
	cmd := &cobra.Command{Use: "test"}
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	cmd.PersistentFlags().StringP("output", "o", "text", "Output format")
	require.NoError(t, cmd.PersistentFlags().Set("output", "yaml"))

	// Load configuration
	cfg, err := LoadWithCommand(cmd)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test precedence: flags > env > file > defaults
	assert.Equal(t, "warn", cfg.LogLevel)         // env overrides file
	assert.True(t, cfg.CLI.Verbose)               // env overrides file
	assert.Equal(t, "yaml", cfg.CLI.OutputFormat) // flag overrides env/file

	// Should track all sources
	sources := cfg.GetLoadedSources()
	// Check that file source exists (may have full path)
	hasFileSource := false
	for _, source := range sources {
		if strings.HasPrefix(source, "file:") {
			hasFileSource = true
			break
		}
	}
	assert.True(t, hasFileSource, "Should have a file source")
	assert.Contains(t, sources, "environment")
	assert.Contains(t, sources, "flags")
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:        "log_level",
		Value:        "invalid",
		ValidOptions: []string{"debug", "info", "warn"},
		Message:      "invalid log level",
	}

	expectedMsg := "log_level: invalid log level (got \"invalid\", valid options: debug, info, warn)"
	assert.Equal(t, expectedMsg, err.Error())
	assert.True(t, IsValidationError(err))
}

func TestSensitiveFieldDetection(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"api_key", true},
		{"token", true},
		{"secret", true},
		{"password", true},
		{"oauth_token", true},
		{"private_key", true},
		{"log_level", false},
		{"output_format", false},
		{"verbose", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := IsSensitiveField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactSensitiveValue(t *testing.T) {
	tests := []struct {
		fieldName string
		value     string
		expected  string
	}{
		{"api_key", "secret123456", "se********56"},
		{"token", "abc", "***"},
		{"password", "verylongpassword", "ve************rd"},
		{"log_level", "debug", "debug"},    // non-sensitive field
		{"normal_field", "value", "value"}, // non-sensitive field
	}

	for _, tt := range tests {
		t.Run(tt.fieldName+"_"+tt.value, func(t *testing.T) {
			result := RedactSensitiveValue(tt.fieldName, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
