package factory

import (
	"os"
	"path/filepath"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
)

// New creates a new factory with all dependencies configured
func New() *cmdutil.Factory {
	f := &cmdutil.Factory{
		AppVersion:     getVersion(),
		ExecutableName: "zen",
		BuildInfo:      GetBuildInfo(),
	}

	// Build dependency chain (order matters)
	f.Config = configFunc()               // No dependencies
	f.IOStreams = ioStreams(f)            // Depends on Config
	f.Logger = loggerFunc(f)              // Depends on Config
	f.WorkspaceManager = workspaceFunc(f) // Depends on Config, Logger
	f.AgentManager = agentFunc(f)         // Depends on Config, Logger

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
			// If config loading fails, use current directory and default config file
			cwd, cwdErr := os.Getwd()
			if cwdErr != nil {
				return nil, cwdErr
			}
			return &workspaceManagerAdapter{
				manager: workspace.New(cwd, "", f.Logger), // Empty string will default to .zen/config.yaml
			}, nil
		}

		configFile := cfg.Workspace.ConfigFile
		if configFile == "" {
			configFile = filepath.Join(".zen", "config.yaml")
		}

		return &workspaceManagerAdapter{
			manager: workspace.New(cfg.Workspace.Root, configFile, f.Logger),
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

// workspaceManagerAdapter adapts the internal workspace.Manager to cmdutil.WorkspaceManager
type workspaceManagerAdapter struct {
	manager *workspace.Manager
}

func (w *workspaceManagerAdapter) Root() string {
	return w.manager.Root()
}

func (w *workspaceManagerAdapter) ConfigFile() string {
	return w.manager.ConfigFile()
}

func (w *workspaceManagerAdapter) Initialize() error {
	return w.manager.Initialize(false)
}

func (w *workspaceManagerAdapter) InitializeWithForce(force bool) error {
	return w.manager.Initialize(force)
}

func (w *workspaceManagerAdapter) Status() (cmdutil.WorkspaceStatus, error) {
	status, err := w.manager.Status()
	if err != nil {
		return cmdutil.WorkspaceStatus{}, err
	}

	return cmdutil.WorkspaceStatus{
		Initialized: status.Initialized,
		ConfigPath:  status.ConfigPath,
		Root:        status.Root,
		Project: cmdutil.ProjectInfo{
			Type:        string(status.Project.Type),
			Name:        status.Project.Name,
			Description: status.Project.Description,
			Version:     status.Project.Version,
			Language:    status.Project.Language,
			Framework:   status.Project.Framework,
			Metadata:    status.Project.Metadata,
		},
	}, nil
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
