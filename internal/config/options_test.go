package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCurrentValue(t *testing.T) {
	config := &Config{
		LogLevel:  "debug",
		LogFormat: "json",
		CLI: CLIConfig{
			NoColor:      true,
			Verbose:      false,
			OutputFormat: "yaml",
		},
		Workspace: WorkspaceConfig{
			Root:       "/custom/path",
			ConfigFile: "custom.yaml",
		},
		Development: DevelopmentConfig{
			Debug:   true,
			Profile: false,
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"log_level", "debug"},
		{"log_format", "json"},
		{"cli.no_color", "true"},
		{"cli.verbose", "false"},
		{"cli.output_format", "yaml"},
		{"workspace.root", "/custom/path"},
		{"workspace.config_file", "custom.yaml"},
		{"development.debug", "true"},
		{"development.profile", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			opt, found := FindOption(tt.key)
			require.True(t, found, "Option %s should exist", tt.key)

			result := opt.GetCurrentValue(config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentValue_DefaultValues(t *testing.T) {
	// Empty config should return default values
	config := &Config{}

	tests := []struct {
		key      string
		expected string
	}{
		{"log_level", "info"},
		{"log_format", "text"},
		{"cli.no_color", "false"},
		{"cli.verbose", "false"},
		{"cli.output_format", "text"},
		{"workspace.root", "."},
		{"workspace.config_file", "zen.yaml"},
		{"development.debug", "false"},
		{"development.profile", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.key+"_default", func(t *testing.T) {
			opt, found := FindOption(tt.key)
			require.True(t, found, "Option %s should exist", tt.key)

			result := opt.GetCurrentValue(config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindOption(t *testing.T) {
	tests := []struct {
		key   string
		found bool
	}{
		{"log_level", true},
		{"cli.verbose", true},
		{"workspace.root", true},
		{"development.debug", true},
		{"nonexistent.key", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			opt, found := FindOption(tt.key)
			assert.Equal(t, tt.found, found)

			if tt.found {
				assert.NotNil(t, opt)
				assert.Equal(t, tt.key, opt.Key)
			} else {
				assert.Nil(t, opt)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		key     string
		wantErr bool
	}{
		{"log_level", false},
		{"cli.verbose", false},
		{"workspace.root", false},
		{"development.debug", false},
		{"nonexistent.key", true},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := ValidateKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown configuration key")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		wantErr bool
	}{
		// Valid values
		{"log_level", "debug", false},
		{"log_level", "info", false},
		{"log_format", "text", false},
		{"log_format", "json", false},
		{"cli.no_color", "true", false},
		{"cli.no_color", "false", false},
		{"cli.output_format", "yaml", false},
		{"workspace.root", "/any/path", false}, // No allowed values restriction

		// Invalid values for restricted fields
		{"log_level", "invalid", true},
		{"log_format", "xml", true},
		{"cli.no_color", "maybe", true},
		{"cli.output_format", "csv", true},

		// Invalid keys
		{"nonexistent.key", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.key+"_"+tt.value, func(t *testing.T) {
			err := ValidateValue(tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvalidValueError(t *testing.T) {
	err := &InvalidValueError{
		Key:         "log_level",
		Value:       "invalid",
		ValidValues: []string{"debug", "info", "warn"},
	}

	expectedMsg := `invalid value "invalid" for key "log_level" (valid values: debug, info, warn)`
	assert.Equal(t, expectedMsg, err.Error())
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "Simple"},
		{"snake_case", "SnakeCase"},
		{"multiple_words_here", "MultipleWordsHere"},
		{"single", "Single"},
		{"", ""},
		{"already_Pascal", "AlreadyPascal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValueFromConfig_InvalidPaths(t *testing.T) {
	config := &Config{
		LogLevel: "info",
	}

	// Test with invalid option that doesn't exist
	opt := ConfigOption{
		Key:          "nonexistent.invalid",
		DefaultValue: "default",
		Type:         "string",
	}

	result := opt.getValueFromConfig(config)
	assert.Empty(t, result)
}

func TestGetValueFromConfig_ComplexTypes(t *testing.T) {
	config := &Config{
		LogLevel:  "debug",
		LogFormat: "json",
		CLI: CLIConfig{
			NoColor:      true,
			Verbose:      false,
			OutputFormat: "yaml",
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"log_level", "debug"},
		{"log_format", "json"},
		{"cli.no_color", "true"},
		{"cli.verbose", "false"},
		{"cli.output_format", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			opt := ConfigOption{
				Key:  tt.key,
				Type: "string",
			}

			result := opt.getValueFromConfig(config)
			assert.Equal(t, tt.expected, result)
		})
	}
}
