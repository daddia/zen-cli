package version

import (
	"bytes"
	"encoding/json"
	"runtime"
	"testing"

	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/jonathandaddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewCmdVersion(t *testing.T) {
	f := &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:  iostreams.Test(),
	}

	cmd := NewCmdVersion(f)

	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Display version information", cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestVersionOutput(t *testing.T) {
	tests := []struct {
		name         string
		outputFormat string
		checkFunc    func(t *testing.T, output string)
	}{
		{
			name:         "text output",
			outputFormat: "text",
			checkFunc: func(t *testing.T, output string) {
				assert.Contains(t, output, "Zen CLI")
				assert.Contains(t, output, "1.0.0")
				assert.Contains(t, output, "Go:")
				assert.Contains(t, output, "OS:")
			},
		},
		{
			name:         "json output",
			outputFormat: "json",
			checkFunc: func(t *testing.T, output string) {
				var info BuildInfo
				err := json.Unmarshal([]byte(output), &info)
				require.NoError(t, err)
				assert.Equal(t, "1.0.0", info.Version)
				assert.Equal(t, runtime.Version(), info.GoVersion)
			},
		},
		{
			name:         "yaml output",
			outputFormat: "yaml",
			checkFunc: func(t *testing.T, output string) {
				var info BuildInfo
				err := yaml.Unmarshal([]byte(output), &info)
				require.NoError(t, err)
				assert.Equal(t, "1.0.0", info.Version)
				assert.Equal(t, runtime.Version(), info.GoVersion)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			f := &cmdutil.Factory{
				AppVersion: "1.0.0",
				IOStreams: &iostreams.IOStreams{
					Out: out,
				},
			}

			cmd := NewCmdVersion(f)
			cmd.SetArgs([]string{"--output", tt.outputFormat})

			err := cmd.Execute()
			require.NoError(t, err)

			tt.checkFunc(t, out.String())
		})
	}
}

func TestDisplayTextVersion(t *testing.T) {
	buf := &bytes.Buffer{}
	info := BuildInfo{
		Version:   "1.2.3",
		GitCommit: "abc123",
		BuildDate: "2024-01-01",
		GoVersion: "go1.21",
		Platform:  "linux/amd64",
	}

	err := displayTextVersion(buf, info)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Zen CLI 1.2.3")
	assert.Contains(t, output, "Commit: abc123")
	assert.Contains(t, output, "Built:  2024-01-01")
	assert.Contains(t, output, "Go:     go1.21")
	assert.Contains(t, output, "OS:     linux/amd64")
}
