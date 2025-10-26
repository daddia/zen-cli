package task

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
				Source:     "jira",
				Sync:       "daily",
				ProjectKey: "PROJ",
			},
			wantError: false,
		},
		{
			name: "invalid source",
			config: Config{
				Source:     "invalid",
				Sync:       "manual",
				ProjectKey: "",
			},
			wantError: true,
			errorMsg:  "invalid source",
		},
		{
			name: "invalid sync",
			config: Config{
				Source:     "local",
				Sync:       "invalid",
				ProjectKey: "",
			},
			wantError: true,
			errorMsg:  "invalid sync",
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

	// Cast to task.Config to check values
	taskDefaults, ok := defaults.(Config)
	require.True(t, ok)

	assert.Equal(t, "local", taskDefaults.Source)
	assert.Equal(t, "manual", taskDefaults.Sync)
	assert.Equal(t, "", taskDefaults.ProjectKey)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "local", config.Source)
	assert.Equal(t, "manual", config.Sync)
	assert.Equal(t, "", config.ProjectKey)

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
				Source:     "local",
				Sync:       "manual",
				ProjectKey: "",
			},
		},
		{
			name: "with values",
			raw: map[string]interface{}{
				"source":      "jira",
				"sync":        "daily",
				"project_key": "PROJ-123",
			},
			expected: Config{
				Source:     "jira",
				Sync:       "daily",
				ProjectKey: "PROJ-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.Parse(tt.raw)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.Source, config.Source)
			assert.Equal(t, tt.expected.Sync, config.Sync)
			assert.Equal(t, tt.expected.ProjectKey, config.ProjectKey)
		})
	}
}

func TestConfigParser_Section(t *testing.T) {
	parser := ConfigParser{}
	assert.Equal(t, "task", parser.Section())
}
