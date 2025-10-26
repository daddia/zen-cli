package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg := LoadDefaults()
	require.NotNil(t, cfg)

	// Test core defaults
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)

	// Test that defaults source is loaded
	assert.Contains(t, cfg.GetLoadedSources(), "defaults")
	assert.Empty(t, cfg.GetConfigFile())
}

func TestLoad(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test core configuration
	assert.NotEmpty(t, cfg.LogLevel)
	assert.NotEmpty(t, cfg.LogFormat)
}

func TestValidateCore(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "text",
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				LogLevel:  "invalid",
				LogFormat: "text",
			},
			wantError: true,
			errorMsg:  "invalid log_level",
		},
		{
			name: "invalid log format",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "invalid",
			},
			wantError: true,
			errorMsg:  "invalid log_format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCore(tt.config)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigRedacted(t *testing.T) {
	cfg := &Config{
		LogLevel:  "debug",
		LogFormat: "json",
	}

	redacted := cfg.Redacted()
	require.NotNil(t, redacted)

	// Test that core fields are preserved
	assert.Equal(t, "debug", redacted.LogLevel)
	assert.Equal(t, "json", redacted.LogFormat)
}
