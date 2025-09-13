package cmdutil

import (
	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/iostreams"
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
