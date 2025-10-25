package cmdutil

import (
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/iostreams"
	zentemplate "github.com/daddia/zen/pkg/template"
	"github.com/daddia/zen/pkg/types"
)

// Factory provides a set of dependencies for commands
type Factory struct {
	AppVersion     string
	ExecutableName string

	IOStreams *iostreams.IOStreams
	Logger    logging.Logger

	Config             func() (*config.Config, error)
	WorkspaceManager   func() (WorkspaceManager, error)
	AgentManager       func() (AgentManager, error)
	AuthManager        func() (auth.Manager, error)
	AssetClient        func() (assets.AssetClientInterface, error)
	Cache              func(basePath string) cache.Manager[string]
	TemplateEngine     func() (TemplateEngineInterface, error)
	IntegrationManager func() (IntegrationManagerInterface, error)

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
	ZenDirectory() string
	Initialize() error
	InitializeWithForce(force bool) error
	Status() (WorkspaceStatus, error)
	CreateTaskDirectory(taskDir string) error
	CreateWorkTypeDirectory(taskDir, workType string) error
	GetWorkTypeDirectories() []string
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

// TemplateEngineInterface is an alias for zentemplate.TemplateEngine
type TemplateEngineInterface = zentemplate.TemplateEngine

// IntegrationManagerInterface defines the interface for integration management
// This is a minimal interface that avoids circular dependencies
type IntegrationManagerInterface interface {
	// IsConfigured checks if integration is properly configured
	IsConfigured() bool

	// GetTaskSystem returns the configured task system of record
	GetTaskSystem() string

	// IsSyncEnabled returns true if sync is enabled
	IsSyncEnabled() bool
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
		AuthManager: func() (auth.Manager, error) {
			return &testAuthManager{}, nil
		},
		AssetClient: func() (assets.AssetClientInterface, error) {
			return &testAssetClient{}, nil
		},
		Cache: func(basePath string) cache.Manager[string] {
			config := cache.Config{
				BasePath:    basePath,
				SizeLimitMB: 10,
				DefaultTTL:  time.Hour,
			}
			serializer := cache.NewStringSerializer()
			return cache.NewManager(config, logging.NewBasic(), serializer)
		},
		TemplateEngine: func() (TemplateEngineInterface, error) {
			return &testTemplateEngine{}, nil
		},
		IntegrationManager: func() (IntegrationManagerInterface, error) {
			return &testIntegrationManager{}, nil
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
	return ".zen/config"
}

func (m *testWorkspaceManager) ZenDirectory() string {
	return ".zen"
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
		ConfigPath:  ".zen/config",
		Root:        ".",
		Project: ProjectInfo{
			Type:     "test",
			Name:     "test-project",
			Language: "go",
		},
	}, nil
}

func (m *testWorkspaceManager) CreateTaskDirectory(taskDir string) error {
	if m.shouldError {
		return fmt.Errorf("test error creating task directory")
	}
	return nil
}

func (m *testWorkspaceManager) CreateWorkTypeDirectory(taskDir, workType string) error {
	if m.shouldError {
		return fmt.Errorf("test error creating work type directory")
	}
	return nil
}

func (m *testWorkspaceManager) GetWorkTypeDirectories() []string {
	return []string{"research", "spikes", "design", "execution", "outcomes"}
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

// testAuthManager is a mock auth manager for testing
type testAuthManager struct{}

func (a *testAuthManager) Authenticate(ctx context.Context, provider string) error {
	return nil
}

func (a *testAuthManager) GetCredentials(provider string) (string, error) {
	return "test-token", nil
}

func (a *testAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	return nil
}

func (a *testAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	return nil
}

func (a *testAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	return true
}

func (a *testAuthManager) ListProviders() []string {
	return []string{"github", "gitlab"}
}

func (a *testAuthManager) DeleteCredentials(provider string) error {
	return nil
}

func (a *testAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	switch provider {
	case "github":
		return &auth.ProviderInfo{
			Name:        provider,
			Type:        "token",
			Description: "Test provider",
			EnvVars:     []string{"ZEN_GITHUB_TOKEN", "GITHUB_TOKEN"},
		}, nil
	case "gitlab":
		return &auth.ProviderInfo{
			Name:        provider,
			Type:        "token",
			Description: "Test provider",
			EnvVars:     []string{"ZEN_GITLAB_TOKEN", "GITLAB_TOKEN"},
		}, nil
	default:
		return nil, fmt.Errorf("provider '%s' is not supported", provider)
	}
}

// testTemplateEngine is a mock template engine for testing
type testTemplateEngine struct{}

func (e *testTemplateEngine) LoadTemplate(ctx context.Context, name string) (*zentemplate.Template, error) {
	return &zentemplate.Template{
		Name:    name,
		Content: "# Test Template\nHello {{.name}}!",
	}, nil
}

func (e *testTemplateEngine) RenderTemplate(ctx context.Context, tmpl *zentemplate.Template, variables map[string]interface{}) (string, error) {
	return "# Test Template\nHello World!", nil
}

func (e *testTemplateEngine) ListTemplates(ctx context.Context, filter zentemplate.TemplateFilter) (*zentemplate.TemplateList, error) {
	return &zentemplate.TemplateList{
		Templates: []zentemplate.TemplateMetadata{
			{
				Name:        "test-template",
				Description: "Test template",
				Category:    "test",
			},
		},
		Total:   1,
		HasMore: false,
	}, nil
}

func (e *testTemplateEngine) ValidateVariables(ctx context.Context, tmpl *zentemplate.Template, variables map[string]interface{}) error {
	return nil
}

func (e *testTemplateEngine) CompileTemplate(ctx context.Context, name, content string, metadata *zentemplate.TemplateMetadata) (*zentemplate.Template, error) {
	return &zentemplate.Template{
		Name:     name,
		Content:  content,
		Metadata: metadata,
	}, nil
}

func (e *testTemplateEngine) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"upper": func(s string) string { return s },
	}
}

// testIntegrationManager is a mock integration manager for testing
type testIntegrationManager struct{}

func (i *testIntegrationManager) IsConfigured() bool {
	return false // Default to not configured for tests
}

func (i *testIntegrationManager) GetTaskSystem() string {
	return ""
}

func (i *testIntegrationManager) IsSyncEnabled() bool {
	return false
}
