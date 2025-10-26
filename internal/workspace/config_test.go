package workspace

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
				Root:    "/test/root",
				ZenPath: ".zen",
			},
			wantError: false,
		},
		{
			name: "empty root",
			config: Config{
				Root:    "",
				ZenPath: ".zen",
			},
			wantError: true,
			errorMsg:  "root is required",
		},
		{
			name: "empty zen_path",
			config: Config{
				Root:    "/test/root",
				ZenPath: "",
			},
			wantError: true,
			errorMsg:  "zen_path is required",
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
	
	// Cast to workspace.Config to check values
	workspaceDefaults, ok := defaults.(Config)
	require.True(t, ok)
	
	assert.Equal(t, ".", workspaceDefaults.Root)
	assert.Equal(t, ".zen", workspaceDefaults.ZenPath)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, ".", config.Root)
	assert.Equal(t, ".zen", config.ZenPath)
	
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
				Root:    ".",
				ZenPath: ".zen",
			},
		},
		{
			name: "with values",
			raw: map[string]interface{}{
				"root":     "/custom/root",
				"zen_path": ".custom",
			},
			expected: Config{
				Root:    "/custom/root",
				ZenPath: ".custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.Parse(tt.raw)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected.Root, config.Root)
			assert.Equal(t, tt.expected.ZenPath, config.ZenPath)
		})
	}
}

func TestConfigParser_Section(t *testing.T) {
	parser := ConfigParser{}
	assert.Equal(t, "workspace", parser.Section())
}
