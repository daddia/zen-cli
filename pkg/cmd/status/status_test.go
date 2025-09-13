package status

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/daddia/zen/pkg/cmd/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock IOStreams for testing
type mockIOStreams struct{}

func (m *mockIOStreams) FormatSectionHeader(text string) string {
	return text + "\n" + "=============="
}

func (m *mockIOStreams) FormatBold(text string) string {
	return text
}

func (m *mockIOStreams) FormatBoolStatus(value bool, trueText, falseText string) string {
	if value {
		return "✓ " + trueText
	}
	return "✗ " + falseText
}

func (m *mockIOStreams) Indent(text string, level int) string {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	return indent + text
}

func TestNewCmdStatus(t *testing.T) {
	f := factory.New()
	cmd := NewCmdStatus(f)

	assert.Equal(t, "status", cmd.Use)
	assert.Contains(t, cmd.Short, "status")
	assert.NotEmpty(t, cmd.Long)
}

func TestDisplayTextStatus(t *testing.T) {
	status := Status{
		Workspace: WorkspaceStatus{
			Initialized: true,
			Path:        "/home/user/project",
			ConfigFile:  "zen.yaml",
		},
		Configuration: ConfigStatus{
			Loaded:   true,
			Source:   "zen.yaml",
			LogLevel: "info",
		},
		System: SystemStatus{
			OS:           "linux",
			Architecture: "amd64",
			GoVersion:    "go1.21.0",
			NumCPU:       8,
		},
		Integrations: IntegrationStatus{
			Available: []string{"jira", "confluence", "git", "slack"},
			Active:    []string{"git"},
		},
	}

	buf := &bytes.Buffer{}
	mockStreams := &mockIOStreams{}
	err := displayTextStatus(buf, status, mockStreams)
	assert.NoError(t, err)

	output := buf.String()

	// Check for main sections
	assert.Contains(t, output, "Zen CLI Status")
	assert.Contains(t, output, "==============")
	assert.Contains(t, output, "Workspace:")
	assert.Contains(t, output, "Configuration:")
	assert.Contains(t, output, "System:")
	assert.Contains(t, output, "Integrations:")

	// Check for specific values
	assert.Contains(t, output, "✓ Ready")
	assert.Contains(t, output, "/home/user/project")
	assert.Contains(t, output, "zen.yaml")
	assert.Contains(t, output, "✓ Loaded")
	assert.Contains(t, output, "linux")
	assert.Contains(t, output, "amd64")
	assert.Contains(t, output, "go1.21.0")
	assert.Contains(t, output, "8")
}

func TestDisplayTextStatus_NotInitialized(t *testing.T) {
	status := Status{
		Workspace: WorkspaceStatus{
			Initialized: false,
			Path:        "/home/user/project",
			ConfigFile:  "",
		},
		Configuration: ConfigStatus{
			Loaded:   false,
			Source:   "none",
			LogLevel: "unknown",
		},
		System: SystemStatus{
			OS:           "darwin",
			Architecture: "arm64",
			GoVersion:    "go1.21.0",
			NumCPU:       4,
		},
		Integrations: IntegrationStatus{
			Available: []string{},
			Active:    []string{},
		},
	}

	buf := &bytes.Buffer{}
	mockStreams := &mockIOStreams{}
	err := displayTextStatus(buf, status, mockStreams)
	assert.NoError(t, err)

	output := buf.String()

	// Check for failure status
	assert.Contains(t, output, "✗ Not Initialized")
	assert.Contains(t, output, "✗ Not Loaded")
	assert.Contains(t, output, "none")
	assert.Contains(t, output, "unknown")
}

func TestGetConfigSource(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() func()
		expected string
	}{
		{
			name: "nil config",
			setup: func() func() {
				return func() {}
			},
			expected: "none",
		},
		{
			name: "non-nil config",
			setup: func() func() {
				return func() {}
			},
			expected: "defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			var cfg interface{}
			if tt.name != "nil config" {
				cfg = map[string]interface{}{}
			}

			result := getConfigSource(cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusCommand_Integration(t *testing.T) {
	f := factory.New()
	cmd := NewCmdStatus(f)

	// Test that the command can be created and basic properties are set
	require.NotNil(t, cmd)
	assert.Equal(t, "status", cmd.Name())
	assert.True(t, cmd.Runnable())

	// Test that it accepts no arguments (cobra.NoArgs)
	assert.NotNil(t, cmd.Args)
}

func TestStatusJSON(t *testing.T) {
	status := Status{
		Workspace: WorkspaceStatus{
			Initialized: true,
			Path:        "/test",
			ConfigFile:  "zen.yaml",
		},
		Configuration: ConfigStatus{
			Loaded:   true,
			Source:   "zen.yaml",
			LogLevel: "info",
		},
		System: SystemStatus{
			OS:           "linux",
			Architecture: "amd64",
			GoVersion:    "go1.21.0",
			NumCPU:       4,
		},
		Integrations: IntegrationStatus{
			Available: []string{"git"},
			Active:    []string{},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(status)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"initialized":true`)
	assert.Contains(t, string(data), `"path":"/test"`)
	assert.Contains(t, string(data), `"loaded":true`)
	assert.Contains(t, string(data), `"os":"linux"`)
}
