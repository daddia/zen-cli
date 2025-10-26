package list

import (
	"bytes"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRun(t *testing.T) {
	// Create test streams
	streams := iostreams.Test()

	// Create options
	opts := &ListOptions{
		IO: streams,
		Config: func() (*config.Config, error) {
			return config.LoadDefaults(), nil
		},
	}

	// Run the command
	err := listRun(opts)
	require.NoError(t, err)

	// Check that output contains core config
	output := streams.Out.(*bytes.Buffer).String()
	assert.Contains(t, output, "[core]")
	assert.Contains(t, output, "log_level")
	assert.Contains(t, output, "log_format")

	// Check that output contains component configs
	assert.Contains(t, output, "[assets]")
	assert.Contains(t, output, "[workspace]")
}

func TestNewCmdConfigList(t *testing.T) {
	streams := iostreams.Test()
	factory := &cmdutil.Factory{
		IOStreams: streams,
		Config: func() (*config.Config, error) {
			return config.LoadDefaults(), nil
		},
	}

	cmd := NewCmdConfigList(factory, nil)

	require.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "Print a list of configuration keys and values", cmd.Short)
	assert.Contains(t, cmd.Long, "List all configuration keys")
	assert.Contains(t, cmd.Aliases, "ls")
}
