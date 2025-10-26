package cmdutil

import (
	"context"
	"errors"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	t.Run("creates factory with basic fields", func(t *testing.T) {
		f := &Factory{
			AppVersion:     "1.0.0",
			ExecutableName: "zen",
			IOStreams:      iostreams.Test(),
			Logger:         logging.NewBasic(),
		}

		assert.Equal(t, "1.0.0", f.AppVersion)
		assert.Equal(t, "zen", f.ExecutableName)
		assert.NotNil(t, f.IOStreams)
		assert.NotNil(t, f.Logger)
	})

	t.Run("config function returns config", func(t *testing.T) {
		expectedConfig := &config.Config{
			LogLevel: "info",
		}

		f := &Factory{
			Config: func() (*config.Config, error) {
				return expectedConfig, nil
			},
		}

		cfg, err := f.Config()
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig, cfg)
	})

	t.Run("config function returns error", func(t *testing.T) {
		expectedErr := errors.New("config error")

		f := &Factory{
			Config: func() (*config.Config, error) {
				return nil, expectedErr
			},
		}

		cfg, err := f.Config()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNewTestFactory(t *testing.T) {
	t.Run("with nil streams", func(t *testing.T) {
		f := NewTestFactory(nil)

		assert.NotNil(t, f)
		assert.Equal(t, "dev", f.AppVersion)
		assert.Equal(t, "zen-test", f.ExecutableName)
		assert.NotNil(t, f.IOStreams)
		assert.NotNil(t, f.Logger)
		assert.NotNil(t, f.Config)
		assert.NotNil(t, f.WorkspaceManager)
		assert.NotNil(t, f.AgentManager)
		assert.NotNil(t, f.AssetClient)
		assert.NotNil(t, f.BuildInfo)
	})

	t.Run("with provided streams", func(t *testing.T) {
		streams := iostreams.Test()
		f := NewTestFactory(streams)

		assert.Equal(t, streams, f.IOStreams)
	})
}

func TestNewTestFactoryWithWorkspace(t *testing.T) {
	streams := iostreams.Test()

	t.Run("initialized workspace", func(t *testing.T) {
		f := NewTestFactoryWithWorkspace(streams, true, false)

		ws, err := f.WorkspaceManager()
		require.NoError(t, err)

		status, err := ws.Status()
		require.NoError(t, err)
		assert.True(t, status.Initialized)
	})

	t.Run("uninitialized workspace", func(t *testing.T) {
		f := NewTestFactoryWithWorkspace(streams, false, false)

		ws, err := f.WorkspaceManager()
		require.NoError(t, err)

		status, err := ws.Status()
		require.NoError(t, err)
		assert.False(t, status.Initialized)
	})

	t.Run("workspace with error", func(t *testing.T) {
		f := NewTestFactoryWithWorkspace(streams, false, true)

		ws, err := f.WorkspaceManager()
		require.NoError(t, err)

		err = ws.Initialize()
		assert.Error(t, err)
	})
}

func TestTestWorkspaceManager(t *testing.T) {
	wm := &testWorkspaceManager{initialized: false, shouldError: false}

	assert.Equal(t, ".", wm.Root())
	assert.Equal(t, ".zen/config", wm.ConfigFile())

	// Test initialization
	err := wm.Initialize()
	assert.NoError(t, err)
	assert.True(t, wm.initialized)

	// Test status
	status, err := wm.Status()
	assert.NoError(t, err)
	assert.True(t, status.Initialized)
	assert.Equal(t, "test-project", status.Project.Name)
	assert.Equal(t, "go", status.Project.Language)
}

func TestTestWorkspaceManager_WithErrors(t *testing.T) {
	wm := &testWorkspaceManager{initialized: true, shouldError: true}

	// Test initialize with force
	err := wm.InitializeWithForce(true)
	assert.NoError(t, err) // Force should work

	// Test initialize without force on initialized workspace with error flag
	err = wm.InitializeWithForce(false)
	assert.Error(t, err)
}

func TestTestAgentManager(t *testing.T) {
	am := &testAgentManager{}

	agents, err := am.List()
	assert.NoError(t, err)
	assert.Equal(t, []string{"test-agent"}, agents)

	result, err := am.Execute("test-agent", "input")
	assert.NoError(t, err)
	assert.Equal(t, "test-output", result)
}

func TestTestAssetClient(t *testing.T) {
	client := &testAssetClient{}
	ctx := context.Background()

	// Test ListAssets
	list, err := client.ListAssets(ctx, assets.AssetFilter{})
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Assets, 1)
	assert.Equal(t, "test-template", list.Assets[0].Name)

	// Test GetAsset
	content, err := client.GetAsset(ctx, "test", assets.GetAssetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, "test", content.Metadata.Name)

	// Test SyncRepository
	result, err := client.SyncRepository(ctx, assets.SyncRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// Test GetCacheInfo
	cacheInfo, err := client.GetCacheInfo(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, cacheInfo)
	assert.Equal(t, int64(1024*1024), cacheInfo.TotalSize)

	// Test ClearCache
	err = client.ClearCache(ctx)
	assert.NoError(t, err)

	// Test Close
	err = client.Close()
	assert.NoError(t, err)
}
