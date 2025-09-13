package version

import (
	"bytes"
	"encoding/json"
	"runtime"
	"testing"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewCmdVersion(t *testing.T) {
	f := &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:  iostreams.Test(),
		BuildInfo: map[string]string{
			"version":    "1.0.0",
			"commit":     "abc123",
			"build_time": "2024-01-01T00:00:00Z",
		},
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
				assert.Equal(t, "zen version 1.0.0\n", output)
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
				BuildInfo: map[string]string{
					"version":    "1.0.0",
					"commit":     "abc123",
					"build_time": "2024-01-01T00:00:00Z",
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
	assert.Equal(t, "zen version 1.2.3\n", output)
}
