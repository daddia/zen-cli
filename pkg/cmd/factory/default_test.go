package factory

import (
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	factory := New()
	require.NotNil(t, factory)
	
	// Test that factory has all required components
	assert.NotNil(t, factory.Config)
	assert.NotNil(t, factory.IOStreams)
	assert.NotNil(t, factory.Logger)
	assert.NotNil(t, factory.WorkspaceManager)
	assert.NotNil(t, factory.AuthManager)
	assert.NotNil(t, factory.AssetClient)
	assert.NotNil(t, factory.TemplateEngine)
}

func TestConfigFunc(t *testing.T) {
	configFn := configFunc()
	require.NotNil(t, configFn)
	
	cfg, err := configFn()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	
	// Test core config fields
	assert.NotEmpty(t, cfg.LogLevel)
	assert.NotEmpty(t, cfg.LogFormat)
}

func TestAssetClientFunctionality(t *testing.T) {
	streams := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: streams,
		Config:    configFunc(),
		Logger:    logging.NewBasic(),
	}
	f.AuthManager = authFunc(f)

	client, err := f.AssetClient()
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestStandardConfigIntegration(t *testing.T) {
	// Test that factory uses standard config APIs
	cfg := config.LoadDefaults()
	require.NotNil(t, cfg)
	
	// Test assets config
	assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
	require.NoError(t, err)
	assert.NotEmpty(t, assetConfig.RepositoryURL)
	assert.NotEmpty(t, assetConfig.Branch)
	
	// Test auth config
	authConfig, err := config.GetConfig(cfg, auth.ConfigParser{})
	require.NoError(t, err)
	assert.NotEmpty(t, authConfig.StorageType)
}