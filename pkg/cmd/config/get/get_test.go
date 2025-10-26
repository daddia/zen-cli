package get

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRun_CoreConfig(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantError bool
	}{
		{
			name:      "valid log_level",
			key:       "log_level",
			wantError: false,
		},
		{
			name:      "valid log_format",
			key:       "log_format",
			wantError: false,
		},
		{
			name:      "invalid key",
			key:       "invalid.key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test streams
			streams := iostreams.Test()

			// Create options
			opts := &GetOptions{
				IO: streams,
				Config: func() (*config.Config, error) {
					return config.LoadDefaults(), nil
				},
				Key: tt.key,
			}

			// Run the command
			err := getRun(opts)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Check that output was written
				output := streams.Out.(*bytes.Buffer).String()
				assert.NotEmpty(t, output)
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

func TestNewCmdConfigGet(t *testing.T) {
	streams := iostreams.Test()
	factory := &cmdutil.Factory{
		IOStreams: streams,
		Config: func() (*config.Config, error) {
			return config.LoadDefaults(), nil
		},
	}

	cmd := NewCmdConfigGet(factory, nil)

	require.NotNil(t, cmd)
	assert.Equal(t, "get <key>", cmd.Use)
	assert.Equal(t, "Print the value of a given configuration key", cmd.Short)
	assert.Contains(t, cmd.Long, "Print the value of a configuration key")
}
