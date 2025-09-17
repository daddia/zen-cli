package cmdutil

import (
	"context"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/types"
)

// Factory provides a set of dependencies for commands
type Factory struct {
	AppVersion     string
	ExecutableName string

	IOStreams *iostreams.IOStreams
	Logger    logging.Logger

	Config           func() (*config.Config, error)
	WorkspaceManager func() (WorkspaceManager, error)
	AgentManager     func() (AgentManager, error)
	AssetClient      func() (assets.AssetClientInterface, error)

	// Global flag values
	ConfigFile string
	DryRun     bool
	Verbose    bool

	// Build information
	BuildInfo map[string]string
}

// WorkspaceManager defines the interface for workspace operations
type WorkspaceManager interface {
	Root() string
	ConfigFile() string
	Initialize() error
	InitializeWithForce(force bool) error
	Status() (WorkspaceStatus, error)
}

// WorkspaceStatus represents the current workspace state
type WorkspaceStatus struct {
	Initialized bool
	ConfigPath  string
	Root        string
	Project     ProjectInfo
}

// ProjectInfo contains detected project information
type ProjectInfo struct {
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Language    string            `json:"language,omitempty"`
	Framework   string            `json:"framework,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// AgentManager defines the interface for AI agent operations
type AgentManager interface {
	List() ([]string, error)
	Execute(name string, input interface{}) (interface{}, error)
}

// NewTestFactory creates a factory for testing
func NewTestFactory(streams *iostreams.IOStreams) *Factory {
	if streams == nil {
		streams = iostreams.Test()
	}

	return &Factory{
		AppVersion:     "dev",
		ExecutableName: "zen-test",
		IOStreams:      streams,
		Logger:         logging.NewBasic(),
		Config: func() (*config.Config, error) {
			return config.LoadDefaults(), nil
		},
		WorkspaceManager: func() (WorkspaceManager, error) {
			return &testWorkspaceManager{initialized: false, shouldError: false}, nil
		},
		AgentManager: func() (AgentManager, error) {
			return &testAgentManager{}, nil
		},
		AssetClient: func() (assets.AssetClientInterface, error) {
			return &testAssetClient{}, nil
		},
		BuildInfo: map[string]string{
			"version":    "dev",
			"commit":     "test-commit",
			"build_time": "2024-01-01T00:00:00Z",
		},
	}
}

// NewTestFactoryWithWorkspace creates a factory with a specific workspace state
func NewTestFactoryWithWorkspace(streams *iostreams.IOStreams, initialized bool, shouldError bool) *Factory {
	factory := NewTestFactory(streams)
	factory.WorkspaceManager = func() (WorkspaceManager, error) {
		return &testWorkspaceManager{initialized: initialized, shouldError: shouldError}, nil
	}
	return factory
}

// testWorkspaceManager is a mock workspace manager for testing
type testWorkspaceManager struct {
	initialized bool
	shouldError bool
}

func (m *testWorkspaceManager) Root() string {
	return "."
}

func (m *testWorkspaceManager) ConfigFile() string {
	return ".zen/config.yaml"
}

func (m *testWorkspaceManager) Initialize() error {
	if m.shouldError {
		return &types.Error{
			Code:    types.ErrorCodeAlreadyExists,
			Message: "workspace already exists",
		}
	}
	m.initialized = true
	return nil
}

func (m *testWorkspaceManager) InitializeWithForce(force bool) error {
	if m.shouldError && !force && m.initialized {
		return &types.Error{
			Code:    types.ErrorCodeAlreadyExists,
			Message: "workspace already exists",
		}
	}
	m.initialized = true
	return nil
}

func (m *testWorkspaceManager) Status() (WorkspaceStatus, error) {
	return WorkspaceStatus{
		Initialized: m.initialized,
		ConfigPath:  ".zen/config.yaml",
		Root:        ".",
		Project: ProjectInfo{
			Type:     "test",
			Name:     "test-project",
			Language: "go",
		},
	}, nil
}

// testAgentManager is a mock agent manager for testing
type testAgentManager struct{}

func (m *testAgentManager) List() ([]string, error) {
	return []string{"test-agent"}, nil
}

func (m *testAgentManager) Execute(name string, input interface{}) (interface{}, error) {
	return "test-output", nil
}

// testAssetClient is a mock asset client for testing
type testAssetClient struct{}

func (c *testAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	return &assets.AssetList{
		Assets: []assets.AssetMetadata{
			{
				Name:        "test-template",
				Type:        assets.AssetTypeTemplate,
				Description: "Test template",
				Format:      "markdown",
				Category:    "test",
				Tags:        []string{"test"},
				Path:        "templates/test.md.template",
			},
		},
		Total:   1,
		HasMore: false,
	}, nil
}

func (c *testAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	return &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        name,
			Type:        assets.AssetTypeTemplate,
			Description: "Test asset",
		},
		Content:  "# Test Content",
		Checksum: "sha256:test",
		Cached:   false,
		CacheAge: 0,
	}, nil
}

func (c *testAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	return &assets.SyncResult{
		Status:        "success",
		DurationMS:    1000,
		AssetsUpdated: 1,
		CacheSizeMB:   10.5,
		LastSync:      time.Now(),
	}, nil
}

func (c *testAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	return &assets.CacheInfo{
		TotalSize:     1024 * 1024, // 1MB
		AssetCount:    5,
		LastSync:      time.Now(),
		CacheHitRatio: 0.85,
	}, nil
}

func (c *testAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (c *testAssetClient) Close() error {
	return nil
}
