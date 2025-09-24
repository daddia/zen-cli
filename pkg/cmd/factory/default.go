package factory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/integration"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/internal/providers/jira"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/plugin"
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
		// Apply configuration settings
		if cfg.CLI.NoColor {
			io.SetColorEnabled(false)
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

	return logging.New(cfg.LogLevel, cfg.LogFormat)
}

func workspaceFunc(f *cmdutil.Factory) func() (cmdutil.WorkspaceManager, error) {
	return func() (cmdutil.WorkspaceManager, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		return &workspaceManager{
			root:       cfg.Workspace.Root,
			configFile: cfg.Workspace.ConfigFile,
			logger:     f.Logger,
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

		// Get auth configuration from main config
		authConfig := getAuthConfig(cfg, f)

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
	root       string
	configFile string
	logger     logging.Logger
	manager    *workspace.Manager
}

func (w *workspaceManager) Root() string {
	return w.root
}

func (w *workspaceManager) ConfigFile() string {
	return w.configFile
}

func (w *workspaceManager) ZenDirectory() string {
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}
	return w.manager.ZenDirectory()
}

func (w *workspaceManager) Initialize() error {
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}
	return w.manager.Initialize(false)
}

func (w *workspaceManager) InitializeWithForce(force bool) error {
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}
	return w.manager.Initialize(force)
}

func (w *workspaceManager) Status() (cmdutil.WorkspaceStatus, error) {
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}

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
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}
	return w.manager.CreateTaskDirectory(taskDir)
}

func (w *workspaceManager) CreateWorkTypeDirectory(taskDir, workType string) error {
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}
	return w.manager.CreateWorkTypeDirectory(taskDir, workType)
}

func (w *workspaceManager) GetWorkTypeDirectories() []string {
	if w.manager == nil {
		w.manager = workspace.New(w.root, w.configFile, w.logger)
	}
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

		// Get asset configuration from main config
		assetConfig := getAssetConfig(cfg)

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

		// Get template engine configuration
		templateConfig := getTemplateEngineConfig(cfg)

		// Create template engine
		cachedEngine = template.NewEngine(logger, assetClient, templateConfig)

		return cachedEngine, nil
	}
}

func getTemplateEngineConfig(cfg *config.Config) template.EngineConfig {
	// Default configuration
	config := template.EngineConfig{
		CacheEnabled:  true,
		CacheTTL:      30 * time.Minute,
		CacheSize:     100,
		StrictMode:    false,
		EnableAI:      false,
		WorkspaceRoot: cfg.Workspace.Root,
	}

	config.DefaultDelims.Left = "{{"
	config.DefaultDelims.Right = "}}"

	// Override with config values if present
	if cfg.Templates.CacheEnabled != nil {
		config.CacheEnabled = *cfg.Templates.CacheEnabled
	}
	if cfg.Templates.CacheTTL != "" {
		if duration, err := time.ParseDuration(cfg.Templates.CacheTTL); err == nil {
			config.CacheTTL = duration
		}
	}
	if cfg.Templates.CacheSize > 0 {
		config.CacheSize = cfg.Templates.CacheSize
	}
	if cfg.Templates.StrictMode != nil {
		config.StrictMode = *cfg.Templates.StrictMode
	}
	if cfg.Templates.EnableAI != nil {
		config.EnableAI = *cfg.Templates.EnableAI
	}
	if cfg.Templates.LeftDelim != "" {
		config.DefaultDelims.Left = cfg.Templates.LeftDelim
	}
	if cfg.Templates.RightDelim != "" {
		config.DefaultDelims.Right = cfg.Templates.RightDelim
	}

	return config
}

func getAssetConfig(cfg *config.Config) assets.AssetConfig {
	// Start with defaults
	config := assets.DefaultAssetConfig()

	// Override with values from main configuration
	if cfg.Assets.RepositoryURL != "" {
		config.RepositoryURL = cfg.Assets.RepositoryURL
	}

	if cfg.Assets.Branch != "" {
		config.Branch = cfg.Assets.Branch
	}

	if cfg.Assets.AuthProvider != "" {
		config.AuthProvider = cfg.Assets.AuthProvider
	}

	if cfg.Assets.CachePath != "" {
		// Expand tilde in cache path
		cachePath := cfg.Assets.CachePath
		if strings.HasPrefix(cachePath, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				cachePath = filepath.Join(home, cachePath[2:])
			}
		}
		config.CachePath = cachePath
	}

	if cfg.Assets.CacheSizeMB > 0 {
		config.CacheSizeMB = cfg.Assets.CacheSizeMB
	}

	if cfg.Assets.SyncTimeoutSeconds > 0 {
		config.SyncTimeoutSeconds = cfg.Assets.SyncTimeoutSeconds
	}

	config.IntegrityChecksEnabled = cfg.Assets.IntegrityChecksEnabled
	config.PrefetchEnabled = cfg.Assets.PrefetchEnabled

	// Environment variable overrides (for testing and advanced users)
	if repoURL := os.Getenv("ZEN_ASSET_REPOSITORY_URL"); repoURL != "" {
		config.RepositoryURL = repoURL
	}

	if provider := os.Getenv("ZEN_AUTH_PROVIDER"); provider != "" {
		config.AuthProvider = provider
	}

	if branch := os.Getenv("ZEN_ASSET_BRANCH"); branch != "" {
		config.Branch = branch
	}

	return config
}

func getAuthConfig(cfg *config.Config, f *cmdutil.Factory) auth.Config {
	// Start with defaults
	config := auth.DefaultConfig()

	// Override with values from main configuration
	if cfg.Assets.AuthProvider != "" {
		// Map asset auth provider to auth storage type
		switch cfg.Assets.AuthProvider {
		case "github", "gitlab":
			config.StorageType = "keychain" // Prefer keychain for security
		default:
			config.StorageType = "file"
		}
	}

	// Environment variable overrides (for testing and advanced users)
	if storageType := os.Getenv("ZEN_AUTH_STORAGE_TYPE"); storageType != "" {
		config.StorageType = storageType
	}

	if storagePath := os.Getenv("ZEN_AUTH_STORAGE_PATH"); storagePath != "" {
		config.StoragePath = storagePath
	}

	if encryptionKey := os.Getenv("ZEN_AUTH_ENCRYPTION_KEY"); encryptionKey != "" {
		config.EncryptionKey = encryptionKey
	}

	// Set default storage path if not specified
	if config.StoragePath == "" {
		// Use project .zen directory for project-specific credentials
		if wm, err := f.WorkspaceManager(); err == nil {
			config.StoragePath = filepath.Join(wm.ZenDirectory(), "auth")
		} else {
			// Fallback to home directory if workspace is not available (for testing/edge cases)
			config.StoragePath = "~/.zen/auth"
		}
	}

	return config
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

		cfg, err := f.Config()
		if err != nil {
			integrationError = err
			return nil, integrationError
		}

		logger := f.Logger

		authManager, err := f.AuthManager()
		if err != nil {
			integrationError = err
			return nil, integrationError
		}

		// Create cache for sync records
		cacheConfig := cache.Config{
			BasePath:    filepath.Join(os.TempDir(), "zen", "integration"),
			SizeLimitMB: 10, // 10MB for sync records
			DefaultTTL:  24 * time.Hour,
		}
		serializer := cache.NewJSONSerializer[*integration.TaskSyncRecord]()
		syncCache := cache.NewManager(cacheConfig, logger, serializer)

		// Create integration service
		integrationService := integration.NewService(cfg, logger, authManager, syncCache)

		// Initialize plugin runtime (using simple runtime for now)
		pluginRuntime := plugin.NewSimpleRuntime(logger, authManager)

		// Create plugin registry
		pluginRegistry := plugin.NewRegistry(logger, cfg.Integrations.PluginDirectories, pluginRuntime)

		// Discover plugins
		if err := pluginRegistry.DiscoverPlugins(context.Background()); err != nil {
			logger.Warn("failed to discover plugins", "error", err)
		}

		// Register Jira provider if configured
		if cfg.Work.Tasks.Source == "jira" {
			if providerConfig, ok := cfg.Integrations.Providers["jira"]; ok {
				jiraProvider := jira.NewProvider(&providerConfig, logger, authManager)
				if err := integrationService.RegisterProvider(jiraProvider); err != nil {
					logger.Warn("failed to register Jira provider", "error", err)
				}
			}
		}

		cachedIntegration = integrationService
		return cachedIntegration, nil
	}
}
