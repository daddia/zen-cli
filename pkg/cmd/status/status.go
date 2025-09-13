package status

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Status represents the current system status
type Status struct {
	Workspace     WorkspaceStatus   `json:"workspace" yaml:"workspace"`
	Configuration ConfigStatus      `json:"configuration" yaml:"configuration"`
	System        SystemStatus      `json:"system" yaml:"system"`
	Integrations  IntegrationStatus `json:"integrations" yaml:"integrations"`
}

// WorkspaceStatus represents workspace information
type WorkspaceStatus struct {
	Initialized bool   `json:"initialized" yaml:"initialized"`
	Path        string `json:"path" yaml:"path"`
	ConfigFile  string `json:"config_file" yaml:"config_file"`
}

// ConfigStatus represents configuration status
type ConfigStatus struct {
	Loaded   bool   `json:"loaded" yaml:"loaded"`
	Source   string `json:"source" yaml:"source"`
	LogLevel string `json:"log_level" yaml:"log_level"`
}

// SystemStatus represents system information
type SystemStatus struct {
	OS           string `json:"os" yaml:"os"`
	Architecture string `json:"architecture" yaml:"architecture"`
	GoVersion    string `json:"go_version" yaml:"go_version"`
	NumCPU       int    `json:"num_cpu" yaml:"num_cpu"`
}

// IntegrationStatus represents integration status
type IntegrationStatus struct {
	Available []string `json:"available" yaml:"available"`
	Active    []string `json:"active" yaml:"active"`
}

// NewCmdStatus creates the status command
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Display workspace and system status",
		Long: `Display comprehensive status information about your Zen workspace,
configuration, system environment, and available integrations.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get configuration
			cfg, configErr := f.Config()

			// Get workspace manager
			var wsStatus cmdutil.WorkspaceStatus
			if ws, err := f.WorkspaceManager(); err == nil {
				wsStatus, _ = ws.Status()
			}

			// Build status
			status := Status{
				Workspace: WorkspaceStatus{
					Initialized: wsStatus.Initialized,
					Path:        wsStatus.Root,
					ConfigFile:  wsStatus.ConfigPath,
				},
				Configuration: ConfigStatus{
					Loaded: configErr == nil,
					Source: getConfigSource(cfg),
					LogLevel: func() string {
						if cfg != nil {
							return cfg.LogLevel
						}
						return "unknown"
					}(),
				},
				System: SystemStatus{
					OS:           runtime.GOOS,
					Architecture: runtime.GOARCH,
					GoVersion:    runtime.Version(),
					NumCPU:       runtime.NumCPU(),
				},
				Integrations: IntegrationStatus{
					Available: []string{"jira", "confluence", "git", "slack"},
					Active:    []string{},
				},
			}

			// Get output format
			outputFormat := "text"
			if cmd.Parent() != nil && cmd.Parent().PersistentFlags().Changed("output") {
				if val, err := cmd.Parent().PersistentFlags().GetString("output"); err == nil {
					outputFormat = val
				}
			}

			switch outputFormat {
			case "json":
				encoder := json.NewEncoder(f.IOStreams.Out)
				encoder.SetIndent("", "  ")
				return encoder.Encode(status)

			case "yaml":
				encoder := yaml.NewEncoder(f.IOStreams.Out)
				defer encoder.Close()
				return encoder.Encode(status)

			default:
				return displayTextStatus(f.IOStreams.Out, status)
			}
		},
	}

	return cmd
}

// getConfigSource determines where configuration was loaded from
func getConfigSource(cfg interface{}) string {
	if cfg == nil {
		return "none"
	}

	// Check for config file
	if _, err := os.Stat("zen.yaml"); err == nil {
		return "zen.yaml"
	}
	if _, err := os.Stat(".zen.yaml"); err == nil {
		return ".zen.yaml"
	}

	// Check home directory
	if home, err := os.UserHomeDir(); err == nil {
		if _, err := os.Stat(home + "/.zen/config.yaml"); err == nil {
			return "~/.zen/config.yaml"
		}
	}

	return "defaults"
}

// displayTextStatus displays status in human-readable text format
func displayTextStatus(out interface{ Write([]byte) (int, error) }, status Status) error {
	fmt.Fprintln(out, "Zen CLI Status")
	fmt.Fprintln(out, "==============")
	fmt.Fprintln(out)

	// Workspace status
	fmt.Fprintln(out, "📁 Workspace:")
	fmt.Fprintf(out, "   Status:      %s\n", getStatusIcon(status.Workspace.Initialized))
	fmt.Fprintf(out, "   Path:        %s\n", status.Workspace.Path)
	fmt.Fprintf(out, "   Config File: %s\n", status.Workspace.ConfigFile)
	fmt.Fprintln(out)

	// Configuration status
	fmt.Fprintln(out, "⚙️  Configuration:")
	fmt.Fprintf(out, "   Status:    %s\n", getStatusIcon(status.Configuration.Loaded))
	fmt.Fprintf(out, "   Source:    %s\n", status.Configuration.Source)
	fmt.Fprintf(out, "   Log Level: %s\n", status.Configuration.LogLevel)
	fmt.Fprintln(out)

	// System information
	fmt.Fprintln(out, "💻 System:")
	fmt.Fprintf(out, "   OS:           %s\n", status.System.OS)
	fmt.Fprintf(out, "   Architecture: %s\n", status.System.Architecture)
	fmt.Fprintf(out, "   Go Version:   %s\n", status.System.GoVersion)
	fmt.Fprintf(out, "   CPU Cores:    %d\n", status.System.NumCPU)
	fmt.Fprintln(out)

	// Integrations
	fmt.Fprintln(out, "🔌 Integrations:")
	fmt.Fprintf(out, "   Available: %v\n", status.Integrations.Available)
	fmt.Fprintf(out, "   Active:    %v\n", status.Integrations.Active)

	return nil
}

// getStatusIcon returns an icon based on status
func getStatusIcon(ok bool) string {
	if ok {
		return "✅ Ready"
	}
	return "❌ Not Ready"
}
