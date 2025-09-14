package factory

import (
	"os"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
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
