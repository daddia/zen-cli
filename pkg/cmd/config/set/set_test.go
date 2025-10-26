package set

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetRun_CoreConfig(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError bool
	}{
		{
			name:      "valid log_level",
			key:       "log_level",
			value:     "debug",
			wantError: false,
		},
		{
			name:      "valid log_format",
			key:       "log_format",
			value:     "json",
			wantError: false,
		},
		{
			name:      "invalid log_level",
			key:       "log_level",
			value:     "invalid",
			wantError: true,
		},
		{
			name:      "invalid key",
			key:       "invalid.key",
			value:     "value",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test streams
			streams := iostreams.Test()
			
			// Create options
			opts := &SetOptions{
				IO: streams,
				Config: func() (*config.Config, error) {
					return config.LoadDefaults(), nil
				},
				Key:   tt.key,
				Value: tt.value,
			}

			// Run the command
			err := setRun(opts)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				
				// Check output contains success message
				output := streams.Out.(*bytes.Buffer).String()
				assert.Contains(t, output, "✓")
			}
		})
	}
}

func TestParseConfigKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantComp  string
		wantField string
		wantError bool
	}{
		{
			name:      "core config key",
			key:       "log_level",
			wantComp:  "core",
			wantField: "log_level",
			wantError: false,
		},
		{
			name:      "component config key",
			key:       "assets.repository_url",
			wantComp:  "assets",
			wantField: "repository_url",
			wantError: false,
		},
		{
			name:      "invalid key",
			key:       "invalid",
			wantComp:  "",
			wantField: "",
			wantError: true,
		},
		{
			name:      "empty key",
			key:       "",
			wantComp:  "",
			wantField: "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, field, err := parseConfigKey(tt.key)
			
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantComp, comp)
				assert.Equal(t, tt.wantField, field)
			}
		})
	}
}

func TestNewCmdConfigSet(t *testing.T) {
	streams := iostreams.Test()
	factory := &cmdutil.Factory{
		IOStreams: streams,
		Config: func() (*config.Config, error) {
			return config.LoadDefaults(), nil
		},
	}

	cmd := NewCmdConfigSet(factory, nil)

	require.NotNil(t, cmd)
	assert.Equal(t, "set <key> <value>", cmd.Use)
	assert.Equal(t, "Update configuration with a value for the given key", cmd.Short)
	assert.Contains(t, cmd.Long, "Set a configuration value")
}