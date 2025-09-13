package cmdutil

import (
	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"github.com/jonathandaddia/zen/pkg/iostreams"
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

	// Global flag values
	ConfigFile string
	DryRun     bool

	// Build information
	BuildInfo map[string]string
}

// WorkspaceManager defines the interface for workspace operations
type WorkspaceManager interface {
	Root() string
	ConfigFile() string
	Initialize() error
	Status() (WorkspaceStatus, error)
}

// WorkspaceStatus represents the current workspace state
type WorkspaceStatus struct {
	Initialized bool
	ConfigPath  string
	Root        string
}

// AgentManager defines the interface for AI agent operations
type AgentManager interface {
	List() ([]string, error)
	Execute(name string, input interface{}) (interface{}, error)
}
