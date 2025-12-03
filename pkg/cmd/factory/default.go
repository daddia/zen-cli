package factory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/cli"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/template"
	"github.com/spf13/cobra"
)

// New creates a new factory with all dependencies configured
func New() *cmdutil.Factory {
	f := &cmdutil.Factory{
		AppVersion:     getVersion(),
		ExecutableName: "zen",
		BuildInfo:      GetBuildInfo(),
	}

	// Build dependency chain (order matters)
	f.Config = configFunc()                   // No dependencies
	f.IOStreams = ioStreams(f)                // Depends on Config
	f.Logger = loggerFunc(f)                  // Depends on Config
	f.WorkspaceManager = workspaceFunc(f)     // Depends on Config, Logger
	f.AgentManager = agentFunc(f)             // Depends on Config, Logger
	f.AuthManager = authFunc(f)               // Depends on Config, Logger
	f.AssetClient = assetClientFunc(f)        // Depends on Config, Logger, AuthManager
	f.Cache = cacheFunc(f)                    // Depends on Logger
	f.TemplateEngine = templateEngineFunc(f)  // Depends on Config, Logger, AssetClient
	f.IntegrationManager = integrationFunc(f) // Depends on Config, Logger, AuthManager, Cache

	return f
}

func configFunc() func() (*config.Config, error) {
	var cachedConfig *config.Config
	var configError error

	return func() (*config.Config, error) {
		if cachedConfig != nil || configError != nil {
			return cachedConfig, configError
		}
		cachedConfig, configError = config.Load()
		return cachedConfig, configError
	}
}

// ConfigWithCommand creates a config function that binds to the provided command
func ConfigWithCommand(cmd *cobra.Command) func() (*config.Config, error) {
	var cachedConfig *config.Config
	var configError error

	return func() (*config.Config, error) {
		if cachedConfig != nil || configError != nil {
			return cachedConfig, configError
		}
		cachedConfig, configError = config.LoadWithCommand(cmd)
		return cachedConfig, configError
	}
}

func ioStreams(f *cmdutil.Factory) *iostreams.IOStreams {
	io := iostreams.System()

	cfg, err := f.Config()
	if err == nil {
		// Get CLI configuration using standard API
		cliConfig, err := config.GetConfig(cfg, cli.ConfigParser{})
		if err == nil {
			// Apply configuration settings
			if cliConfig.NoColor {
				io.SetColorEnabled(false)
			}
		}

		// Check for NO_COLOR environment variable
		if os.Getenv("NO_COLOR") != "" {
			io.SetColorEnabled(false)
		}

		// Check for ZEN_PROMPT_DISABLED
		if os.Getenv("ZEN_PROMPT_DISABLED") != "" {
			io.SetNeverPrompt(true)
		}
	}

	return io
}

func loggerFunc(f *cmdutil.Factory) logging.Logger {
	cfg, err := f.Config()
	if err != nil {
		return logging.NewBasic()
	}

	return logging.New(cfg.Core.LogLevel, cfg.Core.LogFormat)
}

func workspaceFunc(f *cmdutil.Factory) func() (cmdutil.WorkspaceManager, error) {
	return func() (cmdutil.WorkspaceManager, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}

		// Get workspace configuration using standard API
		workspaceConfig, err := config.GetConfig(cfg, workspace.ConfigParser{})
		if err != nil {
			return nil, fmt.Errorf("failed to get workspace config: %w", err)
		}

		// Create workspace manager with typed config
		manager := workspace.New(workspaceConfig, f.Logger)

		return &workspaceManager{
			manager: manager,
			logger:  f.Logger,
		}, nil
	}
}

func agentFunc(f *cmdutil.Factory) func() (cmdutil.AgentManager, error) {
	return func() (cmdutil.AgentManager, error) {
		// Placeholder for future agent manager implementation
		return &agentManager{
			logger: f.Logger,
		}, nil
	}
}

func authFunc(f *cmdutil.Factory) func() (auth.Manager, error) {
	var cachedAuth auth.Manager
	var authError error

	return func() (auth.Manager, error) {
		if cachedAuth != nil || authError != nil {
			return cachedAuth, authError
		}

		cfg, err := f.Config()
		if err != nil {
			authError = err
			return nil, authError
		}

		logger := f.Logger

		// Get auth configuration using standard API
		authConfig, err := config.GetConfig(cfg, auth.ConfigParser{})
		if err != nil {
			authError = err
			return nil, authError
		}

		// Create storage backend
		storage, err := auth.NewStorage(authConfig.StorageType, authConfig, logger)
		if err != nil {
			authError = err
			return nil, authError
		}

		// Create unified auth manager that handles both token and basic auth
		cachedAuth = auth.NewUnifiedAuthManager(authConfig, cfg, logger, storage)

		return cachedAuth, nil
	}
}

// Build information variables set at build time via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func getVersion() string {
	return version
}

// GetBuildInfo returns structured build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    version,
		"commit":     commit,
		"build_time": buildTime,
	}
}

// workspaceManager implements cmdutil.WorkspaceManager
type workspaceManager struct {
	manager *workspace.Manager
	logger  logging.Logger
}

func (w *workspaceManager) Root() string {
	return w.manager.Root()
}

func (w *workspaceManager) ConfigFile() string {
	return w.manager.ConfigFile()
}

func (w *workspaceManager) ZenDirectory() string {
	return w.manager.ZenDirectory()
}

func (w *workspaceManager) Initialize() error {
	return w.manager.Initialize(false)
}

func (w *workspaceManager) InitializeWithForce(force bool) error {
	return w.manager.Initialize(force)
}

func (w *workspaceManager) Status() (cmdutil.WorkspaceStatus, error) {
	status, err := w.manager.Status()
	if err != nil {
		return cmdutil.WorkspaceStatus{}, err
	}

	return cmdutil.WorkspaceStatus{
		Initialized: status.Initialized,
		ConfigPath:  status.ConfigPath,
		Root:        status.Root,
		Project: cmdutil.ProjectInfo{
			Type: string(status.Project.Type),
			Name: status.Project.Name,
		},
	}, nil
}

func (w *workspaceManager) CreateTaskDirectory(taskDir string) error {
	return w.manager.CreateTaskDirectory(taskDir)
}

func (w *workspaceManager) CreateWorkTypeDirectory(taskDir, workType string) error {
	return w.manager.CreateWorkTypeDirectory(taskDir, workType)
}

func (w *workspaceManager) GetWorkTypeDirectories() []string {
	return w.manager.GetWorkTypeDirectories()
}

// agentManager implements cmdutil.AgentManager
type agentManager struct {
	logger logging.Logger
}

func (a *agentManager) List() ([]string, error) {
	// Placeholder implementation
	return []string{}, nil
}

func (a *agentManager) Execute(name string, input interface{}) (interface{}, error) {
	// Placeholder implementation
	return nil, nil
}

// integrationManager implements cmdutil.IntegrationManagerInterface
type integrationManager struct {
	logger logging.Logger
}

func (i *integrationManager) SyncTasks(ctx context.Context) error {
	// Placeholder implementation
	return nil
}

func (i *integrationManager) GetProviders() []string {
	// Placeholder implementation
	return []string{}
}

func (i *integrationManager) GetTaskSystem() string {
	// Placeholder implementation
	return ""
}

func (i *integrationManager) IsConfigured() bool {
	// Placeholder implementation
	return false
}

func (i *integrationManager) IsSyncEnabled() bool {
	// Placeholder implementation
	return false
}

func assetClientFunc(f *cmdutil.Factory) func() (assets.AssetClientInterface, error) {
	var cachedClient assets.AssetClientInterface
	var clientError error

	return func() (assets.AssetClientInterface, error) {
		if cachedClient != nil || clientError != nil {
			return cachedClient, clientError
		}

		cfg, err := f.Config()
		if err != nil {
			clientError = err
			return nil, clientError
		}

		logger := f.Logger

		// Get asset configuration using standard API
		assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
		if err != nil {
			clientError = err
			return nil, clientError
		}

		// Get shared auth manager
		authManager, err := f.AuthManager()
		if err != nil {
			clientError = err
			return nil, clientError
		}

		// Create adapter for assets interface compatibility
		authProvider := assets.NewAuthProviderAdapter(authManager)

		// Set up cache path
		cachePath := assetConfig.CachePath
		if strings.HasPrefix(cachePath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				clientError = err
				return nil, clientError
			}
			cachePath = filepath.Join(home, cachePath[2:])
		}

		cache := assets.NewAssetCacheManager(
			cachePath,
			assetConfig.CacheSizeMB,
			assetConfig.DefaultTTL,
			logger,
		)

		// Use HTTP client for individual file fetching (no repository cloning needed)
		httpClient := assets.NewHTTPManifestClient(logger, authProvider, assetConfig.AuthProvider)

		parser := assets.NewYAMLManifestParser(logger)

		// Create client with HTTP-based file fetching
		cachedClient = assets.NewClientWithHTTP(assetConfig, logger, authProvider, cache, httpClient, parser)

		return cachedClient, nil
	}
}

func templateEngineFunc(f *cmdutil.Factory) func() (cmdutil.TemplateEngineInterface, error) {
	var cachedEngine cmdutil.TemplateEngineInterface
	var engineError error

	return func() (cmdutil.TemplateEngineInterface, error) {
		if cachedEngine != nil || engineError != nil {
			return cachedEngine, engineError
		}

		cfg, err := f.Config()
		if err != nil {
			engineError = err
			return nil, engineError
		}

		logger := f.Logger

		// Get asset client for template loading
		assetClient, err := f.AssetClient()
		if err != nil {
			engineError = err
			return nil, engineError
		}

		// Get template engine configuration using standard API
		templateConfig, err := config.GetConfig(cfg, template.ConfigParser{})
		if err != nil {
			engineError = err
			return nil, engineError
		}

		// Create template engine
		cachedEngine = template.NewEngine(logger, assetClient, templateConfig)

		return cachedEngine, nil
	}
}

func cacheFunc(f *cmdutil.Factory) func(basePath string) cache.Manager[string] {
	return func(basePath string) cache.Manager[string] {
		config := cache.Config{
			BasePath:    basePath,
			SizeLimitMB: 50, // Default 50MB for general purpose cache
			DefaultTTL:  24 * time.Hour,
		}
		serializer := cache.NewStringSerializer()
		return cache.NewManager(config, f.Logger, serializer)
	}
}

func integrationFunc(f *cmdutil.Factory) func() (cmdutil.IntegrationManagerInterface, error) {
	var cachedIntegration cmdutil.IntegrationManagerInterface
	var integrationError error

	return func() (cmdutil.IntegrationManagerInterface, error) {
		if cachedIntegration != nil || integrationError != nil {
			return cachedIntegration, integrationError
		}

		logger := f.Logger

		// Create integration manager (placeholder implementation)
		// TODO: Implement proper integration manager once integration types are refactored
		cachedIntegration = &integrationManager{
			logger: logger,
		}

		return cachedIntegration, nil
	}
}
