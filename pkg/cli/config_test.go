package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				NoColor:      false,
				Verbose:      true,
				OutputFormat: "json",
			},
			wantError: false,
		},
		{
			name: "invalid output format",
			config: Config{
				NoColor:      false,
				Verbose:      false,
				OutputFormat: "invalid",
			},
			wantError: true,
			errorMsg:  "invalid output_format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	config := Config{}
	defaults := config.Defaults()
	
	require.NotNil(t, defaults)
	
	// Cast to cli.Config to check values
	cliDefaults, ok := defaults.(Config)
	require.True(t, ok)
	
	assert.Equal(t, false, cliDefaults.NoColor)
	assert.Equal(t, false, cliDefaults.Verbose)
	assert.Equal(t, "text", cliDefaults.OutputFormat)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, false, config.NoColor)
	assert.Equal(t, false, config.Verbose)
	assert.Equal(t, "text", config.OutputFormat)
	
	// Validate defaults
	require.NoError(t, config.Validate())
}

func TestConfigParser_Parse(t *testing.T) {
	parser := ConfigParser{}
	
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected Config
	}{
		{
			name: "empty raw data",
			raw:  map[string]interface{}{},
			expected: Config{
				NoColor:      false,
				Verbose:      false,
				OutputFormat: "text",
			},
		},
		{
			name: "with values",
			raw: map[string]interface{}{
				"no_color":      true,
				"verbose":       true,
				"output_format": "json",
			},
			expected: Config{
				NoColor:      true,
				Verbose:      true,
				OutputFormat: "json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.Parse(tt.raw)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected.NoColor, config.NoColor)
			assert.Equal(t, tt.expected.Verbose, config.Verbose)
			assert.Equal(t, tt.expected.OutputFormat, config.OutputFormat)
		})
	}
}

func TestConfigParser_Section(t *testing.T) {
	parser := ConfigParser{}
	assert.Equal(t, "cli", parser.Section())
}
