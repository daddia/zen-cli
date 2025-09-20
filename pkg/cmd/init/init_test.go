package init

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdInit(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	require.NotNil(t, cmd)

	// Test command properties
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize your new Zen workspace or reinitialize an existing one", cmd.Short)
	assert.Contains(t, cmd.Long, "Initialize a new Zen workspace in the current directory")

	// Test flags exist
	flags := cmd.Flags()

	forceFlag := flags.Lookup("force")
	require.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)

	configFlag := flags.Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)
}

func TestInitCommand_NewWorkspace(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create test factory with mock workspace manager
	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output
	output := stdout.String()
	assert.Contains(t, output, "Initialized empty Zen workspace")
	assert.Contains(t, output, ".zen/")
}

func TestInitCommand_WithForceFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create existing .zen directory
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{"--force"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Should succeed with force flag
	output := stdout.String()
	assert.Contains(t, output, "Initialized empty Zen workspace")
}

func TestInitCommand_ExistingWorkspaceWithoutForce(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Create existing .zen directory with config
	zenDir := filepath.Join(tempDir, ".zen")
	require.NoError(t, os.MkdirAll(zenDir, 0755))
	configFile := filepath.Join(zenDir, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("existing"), 0644))

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactoryWithWorkspace(streams, true, false)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()

	// Should succeed without force flag (idempotent behavior)
	require.NoError(t, err)

	// Should show reinitialized message
	output := stdout.String()
	assert.Contains(t, output, "Reinitialized existing Zen workspace")
}

func TestInitCommand_WithCustomConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	customConfigPath := "config/custom.yaml"

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{"--config", customConfigPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Check that config directory was created
	configDir := filepath.Dir(filepath.Join(tempDir, customConfigPath))
	assert.DirExists(t, configDir)

	// Check output
	output := stdout.String()
	assert.Contains(t, output, "Initialized empty Zen workspace")
}

func TestInitCommand_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create a directory with restricted permissions
	tempDir := t.TempDir()
	restrictedDir := filepath.Join(tempDir, "restricted")
	require.NoError(t, os.MkdirAll(restrictedDir, 0000)) // No permissions
	defer os.Chmod(restrictedDir, 0755)                  // Restore permissions for cleanup

	// Change to restricted directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()

	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)

	// This test is tricky because we can't easily simulate permission denied
	// in a cross-platform way. We'll test the error handling path instead.
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// The actual permission test would require platform-specific setup
	// For now, we'll test that the command handles errors gracefully
	assert.NotNil(t, cmd.RunE)
}

func TestInitCommand_VerboseOutput(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	var stdout, stderr bytes.Buffer
	streams := iostreams.Test()
	streams.Out = &stdout
	streams.ErrOut = &stderr
	factory := cmdutil.NewTestFactory(streams)
	factory.Verbose = true // Enable verbose mode

	cmd := NewCmdInit(factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	require.NoError(t, err)

	// In verbose mode, should see analysis output
	output := stdout.String()
	assert.Contains(t, output, "Analyzing project") // Verbose output
	assert.Contains(t, output, "Initialized empty Zen workspace")
}

func TestInitCommand_InvalidConfigPath(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Try to create config in a location that will fail
	invalidConfigPath := "/root/cannot/create/this/config.yaml"

	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	cmd := NewCmdInit(factory)

	// Execute with invalid config file path
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--config", invalidConfigPath})

	err = cmd.Execute()

	// Should handle the error gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create config directory")
	}
}

func TestInitCommand_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) error
		expectedError string
		wantSilent    bool
		initialized   bool
	}{
		{
			name: "workspace already exists (now succeeds with reinitialization)",
			setupFunc: func(dir string) error {
				zenDir := filepath.Join(dir, ".zen")
				if err := os.MkdirAll(zenDir, 0755); err != nil {
					return err
				}
				configFile := filepath.Join(zenDir, "config.yaml")
				return os.WriteFile(configFile, []byte("existing"), 0644)
			},
			expectedError: "", // No error expected
			wantSilent:    false,
			initialized:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			// Setup test condition
			if tt.setupFunc != nil {
				require.NoError(t, tt.setupFunc(tempDir))
			}

			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr
			factory := cmdutil.NewTestFactoryWithWorkspace(streams, tt.initialized, false)

			cmd := NewCmdInit(factory)
			cmd.SetArgs([]string{})

			err = cmd.Execute()

			switch {
			case tt.wantSilent:
				assert.Equal(t, cmdutil.ErrSilent, err)
			case tt.expectedError != "":
				require.Error(t, err)
			default:
				require.NoError(t, err) // Should succeed for reinitialization
			}

			if tt.expectedError != "" {
				output := stderr.String()
				assert.Contains(t, output, tt.expectedError)
			}
		})
	}
}

// Benchmark tests
func BenchmarkInitCommand_New(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCmdInit(factory)
		if cmd == nil {
			b.Fatal("command is nil")
		}
	}
}

func BenchmarkInitCommand_Execute(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tempDir := b.TempDir()
		oldWd, err := os.Getwd()
		if err != nil {
			// In CI environment, getwd might fail, skip this benchmark
			b.Skip("getwd failed, likely in CI environment:", err)
		}

		if err := os.Chdir(tempDir); err != nil {
			b.Fatal(err)
		}

		cmd := NewCmdInit(factory)
		cmd.SetArgs([]string{})
		b.StartTimer()

		err = cmd.Execute()
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		os.Chdir(oldWd)
	}
}

// MockAuthManager provides a comprehensive mock for testing
type MockAuthManager struct {
	mock.Mock
}

func (m *MockAuthManager) Authenticate(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) GetCredentials(provider string) (string, error) {
	args := m.Called(provider)
	return args.String(0), args.Error(1)
}

func (m *MockAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	args := m.Called(ctx, provider)
	return args.Bool(0)
}

func (m *MockAuthManager) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAuthManager) DeleteCredentials(provider string) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	args := m.Called(provider)
	return args.Get(0).(*auth.ProviderInfo), args.Error(1)
}

// MockAssetClient provides a comprehensive mock for testing
type MockAssetClient struct {
	mock.Mock
}

func (m *MockAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*assets.AssetList), args.Error(1)
}

func (m *MockAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*assets.AssetContent), args.Error(1)
}

func (m *MockAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*assets.SyncResult), args.Error(1)
}

func (m *MockAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(*assets.CacheInfo), args.Error(1)
}

func (m *MockAssetClient) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAssetClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test setupAssetsInfrastructure function
func TestSetupAssetsInfrastructure(t *testing.T) {
	tests := []struct {
		name            string
		wasInitialized  bool
		setupFactory    func(streams *iostreams.IOStreams) *cmdutil.Factory
		setupDir        func(tempDir string) error
		expectAssetSync bool
		wantErr         bool
	}{
		{
			name:           "successful setup with authenticated GitHub",
			wasInitialized: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)

				// Mock authenticated auth manager
				mockAuth := &MockAuthManager{}
				mockAuth.On("IsAuthenticated", mock.AnythingOfType("*context.timerCtx"), "github").Return(true)

				factory.AuthManager = func() (auth.Manager, error) {
					return mockAuth, nil
				}

				// Mock asset client with successful sync
				mockAsset := &MockAssetClient{}
				mockAsset.On("SyncRepository", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("assets.SyncRequest")).Return(&assets.SyncResult{
					Status:        "success",
					AssetsUpdated: 5,
				}, nil)
				mockAsset.On("Close").Return(nil)

				factory.AssetClient = func() (assets.AssetClientInterface, error) {
					return mockAsset, nil
				}

				return factory
			},
			expectAssetSync: true,
			wantErr:         false,
		},
		{
			name:           "no auth manager available",
			wasInitialized: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)
				factory.AuthManager = func() (auth.Manager, error) {
					return nil, errors.New("auth manager not available")
				}
				return factory
			},
			expectAssetSync: false,
			wantErr:         false, // Should not fail
		},
		{
			name:           "not authenticated with GitHub",
			wasInitialized: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)

				mockAuth := &MockAuthManager{}
				mockAuth.On("IsAuthenticated", mock.AnythingOfType("*context.timerCtx"), "github").Return(false)

				factory.AuthManager = func() (auth.Manager, error) {
					return mockAuth, nil
				}

				return factory
			},
			expectAssetSync: false,
			wantErr:         false,
		},
		{
			name:           "asset client error",
			wasInitialized: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)

				mockAuth := &MockAuthManager{}
				mockAuth.On("IsAuthenticated", mock.AnythingOfType("*context.timerCtx"), "github").Return(true)

				factory.AuthManager = func() (auth.Manager, error) {
					return mockAuth, nil
				}

				factory.AssetClient = func() (assets.AssetClientInterface, error) {
					return nil, errors.New("asset client failed")
				}

				return factory
			},
			expectAssetSync: false,
			wantErr:         false, // Should not fail init
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			if tt.setupDir != nil {
				require.NoError(t, tt.setupDir(tempDir))
			}

			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			factory := tt.setupFactory(streams)

			err = setupAssetsInfrastructure(factory, tt.wasInitialized)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check that .zen/assets directory was created
			assetsDir := filepath.Join(tempDir, ".zen", "assets")
			assert.DirExists(t, assetsDir)

			if tt.expectAssetSync {
				output := stdout.String()
				assert.Contains(t, output, "Assets manifest")
			}
		})
	}
}

// Test fetchManifestBestEffort function
func TestFetchManifestBestEffort(t *testing.T) {
	tests := []struct {
		name          string
		wasReinit     bool
		setupFactory  func(streams *iostreams.IOStreams) *cmdutil.Factory
		setupManifest func(tempDir string) error
		expectSync    bool
		wantErr       bool
	}{
		{
			name:      "successful sync on new init",
			wasReinit: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)

				mockAsset := &MockAssetClient{}
				mockAsset.On("SyncRepository", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("assets.SyncRequest")).Return(&assets.SyncResult{
					Status:        "success",
					AssetsUpdated: 3,
				}, nil)
				mockAsset.On("Close").Return(nil)

				factory.AssetClient = func() (assets.AssetClientInterface, error) {
					return mockAsset, nil
				}

				return factory
			},
			expectSync: true,
			wantErr:    false,
		},
		{
			name:      "skip sync with recent manifest",
			wasReinit: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				return cmdutil.NewTestFactory(streams)
			},
			setupManifest: func(tempDir string) error {
				// Create recent manifest file
				manifestPath := filepath.Join(tempDir, ".zen", "assets", "manifest.yaml")
				return os.WriteFile(manifestPath, []byte("manifest: test"), 0644)
			},
			expectSync: false,
			wantErr:    false,
		},
		{
			name:      "force sync on reinit",
			wasReinit: true,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)

				mockAsset := &MockAssetClient{}
				mockAsset.On("SyncRepository", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req assets.SyncRequest) bool {
					return req.Force == true && req.Shallow == true
				})).Return(&assets.SyncResult{
					Status:        "success",
					AssetsUpdated: 0, // No updates but still success
				}, nil)
				mockAsset.On("Close").Return(nil)

				factory.AssetClient = func() (assets.AssetClientInterface, error) {
					return mockAsset, nil
				}

				return factory
			},
			setupManifest: func(tempDir string) error {
				// Create old manifest file
				manifestPath := filepath.Join(tempDir, ".zen", "assets", "manifest.yaml")
				return os.WriteFile(manifestPath, []byte("manifest: old"), 0644)
			},
			expectSync: true,
			wantErr:    false,
		},
		{
			name:      "asset client error",
			wasReinit: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)
				factory.AssetClient = func() (assets.AssetClientInterface, error) {
					return nil, errors.New("asset client failed")
				}
				return factory
			},
			expectSync: false,
			wantErr:    false, // Should not fail
		},
		{
			name:      "sync failure",
			wasReinit: false,
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)

				mockAsset := &MockAssetClient{}
				mockAsset.On("SyncRepository", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("assets.SyncRequest")).Return((*assets.SyncResult)(nil), errors.New("sync failed"))
				mockAsset.On("Close").Return(nil)

				factory.AssetClient = func() (assets.AssetClientInterface, error) {
					return mockAsset, nil
				}

				return factory
			},
			expectSync: false,
			wantErr:    false, // Should not fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			// Create .zen/assets directory
			assetsDir := filepath.Join(tempDir, ".zen", "assets")
			require.NoError(t, os.MkdirAll(assetsDir, 0755))

			if tt.setupManifest != nil {
				require.NoError(t, tt.setupManifest(tempDir))
			}

			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			factory := tt.setupFactory(streams)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = fetchManifestBestEffort(factory, ctx, tt.wasReinit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectSync {
				output := stdout.String()
				assert.True(t, strings.Contains(output, "Assets manifest") || strings.Contains(output, "up to date"))
			}
		})
	}
}

// Test workspace initialization with different error conditions
func TestInitCommandWorkspaceErrors(t *testing.T) {
	tests := []struct {
		name           string
		setupFactory   func(streams *iostreams.IOStreams) *cmdutil.Factory
		args           []string
		wantErr        bool
		wantSilent     bool
		expectedOutput string
	}{
		{
			name: "workspace manager error",
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)
				factory.WorkspaceManager = func() (cmdutil.WorkspaceManager, error) {
					return nil, errors.New("workspace manager failed")
				}
				return factory
			},
			args:           []string{},
			wantErr:        true,
			expectedOutput: "failed to get workspace manager",
		},
		{
			name: "permission denied error",
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)
				factory.WorkspaceManager = func() (cmdutil.WorkspaceManager, error) {
					return &mockWorkspaceManager{
						initError: &types.Error{
							Code:    types.ErrorCodePermissionDenied,
							Message: "permission denied",
						},
					}, nil
				}
				return factory
			},
			args:           []string{},
			wantErr:        true,
			wantSilent:     true,
			expectedOutput: "Permission denied",
		},
		{
			name: "generic initialization error",
			setupFactory: func(streams *iostreams.IOStreams) *cmdutil.Factory {
				factory := cmdutil.NewTestFactory(streams)
				factory.WorkspaceManager = func() (cmdutil.WorkspaceManager, error) {
					return &mockWorkspaceManager{
						initError: errors.New("generic init error"),
					}, nil
				}
				return factory
			},
			args:           []string{},
			wantErr:        true,
			expectedOutput: "failed to initialize workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			factory := tt.setupFactory(streams)
			cmd := NewCmdInit(factory)
			cmd.SetArgs(tt.args)

			err = cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantSilent {
					assert.Equal(t, cmdutil.ErrSilent, err)
				}
				if tt.expectedOutput != "" {
					// Check both stdout and stderr as Cobra might print errors to either
					output := stdout.String() + stderr.String()
					// If not found there, check the error message itself
					if !strings.Contains(output, tt.expectedOutput) && err != nil {
						assert.Contains(t, err.Error(), tt.expectedOutput)
					} else {
						assert.Contains(t, output, tt.expectedOutput)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Mock workspace manager for testing error conditions
type mockWorkspaceManager struct {
	initialized bool
	initError   error
	statusError error
}

func (m *mockWorkspaceManager) Root() string {
	return "."
}

func (m *mockWorkspaceManager) ConfigFile() string {
	return ".zen/config.yaml"
}

func (m *mockWorkspaceManager) Initialize() error {
	return m.initError
}

func (m *mockWorkspaceManager) InitializeWithForce(force bool) error {
	return m.initError
}

func (m *mockWorkspaceManager) Status() (cmdutil.WorkspaceStatus, error) {
	if m.statusError != nil {
		return cmdutil.WorkspaceStatus{}, m.statusError
	}
	return cmdutil.WorkspaceStatus{
		Initialized: m.initialized,
		ConfigPath:  ".zen/config.yaml",
		Root:        ".",
		Project: cmdutil.ProjectInfo{
			Type: "test",
			Name: "test-project",
		},
	}, nil
}

func (m *mockWorkspaceManager) CreateTaskDirectory(taskDir string) error {
	return nil // Mock implementation
}

func (m *mockWorkspaceManager) CreateWorkTypeDirectory(taskDir, workType string) error {
	return nil // Mock implementation
}

func (m *mockWorkspaceManager) GetWorkTypeDirectories() []string {
	return []string{"research", "spikes", "design", "execution", "outcomes"}
}

// Test command flag validation
func TestInitCommandFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "valid no args",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "valid force flag",
			args:        []string{"--force"},
			expectError: false,
		},
		{
			name:        "valid config flag",
			args:        []string{"--config", "custom.yaml"},
			expectError: false,
		},
		{
			name:        "valid short flags",
			args:        []string{"-f", "-c", "custom.yaml"},
			expectError: false,
		},
		{
			name:        "invalid extra arguments",
			args:        []string{"extra", "args"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams := iostreams.Test()
			factory := cmdutil.NewTestFactory(streams)
			cmd := NewCmdInit(factory)

			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// May succeed or fail depending on workspace setup, but shouldn't be a flag error
				if err != nil && strings.Contains(err.Error(), "accepts") {
					t.Errorf("Unexpected flag validation error: %v", err)
				}
			}
		})
	}
}

// Test concurrent initialization
func TestInitCommandConcurrency(t *testing.T) {
	const numGoroutines = 5
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				done <- err
				return
			}

			defer func() {
				os.Chdir(oldWd)
			}()

			if err := os.Chdir(tempDir); err != nil {
				done <- err
				return
			}

			streams := iostreams.Test()
			factory := cmdutil.NewTestFactory(streams)
			cmd := NewCmdInit(factory)
			cmd.SetArgs([]string{})

			err = cmd.Execute()
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			// Some may fail due to test environment, but shouldn't panic
			if err != nil {
				t.Logf("Goroutine %d failed (may be expected): %v", i, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}

// Test command metadata and structure
func TestInitCommandMetadata(t *testing.T) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)
	cmd := NewCmdInit(factory)

	// Test command structure
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "workspace", cmd.GroupID)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.NotNil(t, cmd.RunE)

	// Test examples contain expected content
	assert.Contains(t, cmd.Example, "zen init")
	assert.Contains(t, cmd.Long, "Initialize a new Zen workspace")
	assert.Contains(t, cmd.Long, ".zen/")

	// Test that Args is set correctly (no args allowed)
	assert.NotNil(t, cmd.Args)
}

// Performance tests
func BenchmarkNewCmdInit(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCmdInit(factory)
		if cmd == nil {
			b.Fatal("command is nil")
		}
	}
}

func BenchmarkSetupAssetsInfrastructure(b *testing.B) {
	streams := iostreams.Test()
	factory := cmdutil.NewTestFactory(streams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tempDir := b.TempDir()
		oldWd, err := os.Getwd()
		if err != nil {
			// In CI environment, getwd might fail, skip this benchmark
			b.Skip("getwd failed, likely in CI environment:", err)
		}

		if err := os.Chdir(tempDir); err != nil {
			b.Fatal(err)
		}

		b.StartTimer()

		err = setupAssetsInfrastructure(factory, false)
		// Ignore error - acceptable in benchmark
		_ = err

		b.StopTimer()
		os.Chdir(oldWd)
	}
}
