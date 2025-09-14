package get

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRun(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		config         *config.Config
		expectedOutput string
		expectedError  string
	}{
		{
			name: "get log_level",
			key:  "log_level",
			config: &config.Config{
				LogLevel: "debug",
			},
			expectedOutput: "debug\n",
		},
		{
			name: "get cli.verbose",
			key:  "cli.verbose",
			config: &config.Config{
				CLI: config.CLIConfig{
					Verbose: true,
				},
			},
			expectedOutput: "true\n",
		},
		{
			name: "get workspace.root",
			key:  "workspace.root",
			config: &config.Config{
				Workspace: config.WorkspaceConfig{
					Root: "/custom/workspace",
				},
			},
			expectedOutput: "/custom/workspace\n",
		},
		{
			name:          "invalid key",
			key:           "invalid.key",
			config:        &config.Config{},
			expectedError: "unknown configuration key: invalid.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios := iostreams.Test()
			out := ios.Out.(*bytes.Buffer)

			opts := &GetOptions{
				IO:  ios,
				Key: tt.key,
				Config: func() (*config.Config, error) {
					return tt.config, nil
				},
			}

			err := getRun(opts)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, out.String())
			}
		})
	}
}

func TestGetRun_ConfigError(t *testing.T) {
	ios := iostreams.Test()

	opts := &GetOptions{
		IO:  ios,
		Key: "log_level",
		Config: func() (*config.Config, error) {
			return nil, assert.AnError
		},
	}

	err := getRun(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}
