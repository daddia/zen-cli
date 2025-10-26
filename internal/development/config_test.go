package development

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
	}{
		{
			name: "valid config - debug enabled",
			config: Config{
				Debug:   true,
				Profile: false,
			},
			wantError: false,
		},
		{
			name: "valid config - all disabled",
			config: Config{
				Debug:   false,
				Profile: false,
			},
			wantError: false,
		},
		{
			name: "valid config - all enabled",
			config: Config{
				Debug:   true,
				Profile: true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
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
	
	// Cast to development.Config to check values
	devDefaults, ok := defaults.(Config)
	require.True(t, ok)
	
	assert.Equal(t, false, devDefaults.Debug)
	assert.Equal(t, false, devDefaults.Profile)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, false, config.Profile)
	
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
				Debug:   false,
				Profile: false,
			},
		},
		{
			name: "with values",
			raw: map[string]interface{}{
				"debug":   true,
				"profile": true,
			},
			expected: Config{
				Debug:   true,
				Profile: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.Parse(tt.raw)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected.Debug, config.Debug)
			assert.Equal(t, tt.expected.Profile, config.Profile)
		})
	}
}

func TestConfigParser_Section(t *testing.T) {
	parser := ConfigParser{}
	assert.Equal(t, "development", parser.Section())
}
