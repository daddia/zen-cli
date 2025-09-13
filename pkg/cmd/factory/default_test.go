package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	f := New()

	assert.NotNil(t, f)
	assert.Equal(t, "zen", f.ExecutableName)
	assert.Equal(t, "dev", f.AppVersion)
	assert.NotNil(t, f.IOStreams)
	assert.NotNil(t, f.Logger)
	assert.NotNil(t, f.Config)
	assert.NotNil(t, f.WorkspaceManager)
	assert.NotNil(t, f.AgentManager)
}

func TestConfigFunc(t *testing.T) {
	configFn := configFunc()

	// First call loads config
	cfg1, err1 := configFn()
	// Config may or may not exist, so we don't assert on error

	// Second call returns cached result
	cfg2, err2 := configFn()

	if err1 == nil {
		assert.Equal(t, cfg1, cfg2)
	}
	assert.Equal(t, err1, err2)
}

func TestWorkspaceManager(t *testing.T) {
	f := New()

	wm, err := f.WorkspaceManager()
	require.NoError(t, err)
	require.NotNil(t, wm)

	// Test workspace manager methods
	assert.NotEmpty(t, wm.Root())
	assert.NotEmpty(t, wm.ConfigFile())

	status, err := wm.Status()
	assert.NoError(t, err)
	// Workspace may not be initialized in test environment
	assert.NotEmpty(t, status.Root)
}

func TestAgentManager(t *testing.T) {
	f := New()

	am, err := f.AgentManager()
	require.NoError(t, err)
	require.NotNil(t, am)

	// Test agent manager methods
	agents, err := am.List()
	assert.NoError(t, err)
	assert.NotNil(t, agents)

	result, err := am.Execute("test", nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
