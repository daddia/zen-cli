package status

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
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
		Use:     "status",
		Short:   "Display workspace and system status",
		GroupID: "workspace",
		Long: `Display comprehensive status information about your Zen workspace,
configuration, system environment, and available integrations.

This command provides a detailed overview of the current state of your Zen installation
and workspace, helping you troubleshoot issues and understand your environment.`,
		Example: `  # Display status overview
  zen status

  # Output status as JSON for scripting
  zen status --output json

  # Output status as YAML
  zen status --output yaml

  # Check status with verbose output
  zen status --verbose`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get workspace manager and check if we're in a zen workspace
			ws, err := f.WorkspaceManager()
			if err != nil {
				return fmt.Errorf("failed to get workspace manager: %w", err)
			}

			wsStatus, err := ws.Status()
			if err != nil {
				return fmt.Errorf("failed to get workspace status: %w", err)
			}

			// If not in a zen workspace, return git-like error message
			if !wsStatus.Initialized {
				fmt.Fprintf(f.IOStreams.ErrOut, "%s\n",
					f.IOStreams.FormatError("Not Initialized: Not a zen workspace (or any of the parent directories): .zen"))
				return cmdutil.ErrSilent
			}

			// Get configuration
			cfg, configErr := f.Config()

			// Build status
			status := Status{
				Workspace: WorkspaceStatus{
					Initialized: wsStatus.Initialized,
					Path:        wsStatus.Root,
					ConfigFile:  wsStatus.ConfigPath,
				},
				Configuration: ConfigStatus{
					Loaded: configErr == nil && isRealConfig(cfg),
					Source: getConfigSource(cfg),
					LogLevel: func() string {
						if cfg != nil {
							return cfg.Core.LogLevel
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
				return displayTextStatus(f.IOStreams.Out, status, f.IOStreams)
			}
		},
	}

	return cmd
}

// getConfigSource determines where configuration was loaded from
func getConfigSource(cfg *config.Config) string {
	if cfg == nil {
		return "none"
	}

	// Use the actual config file that was loaded
	if configFile := cfg.GetConfigFile(); configFile != "" {
		return filepath.Base(configFile)
	}

	return "defaults"
}

// isRealConfig determines if the config is from a real source vs just defaults
func isRealConfig(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}

	// Check if config was loaded from any real source (not just defaults)
	sources := cfg.GetLoadedSources()
	for _, source := range sources {
		if source != "defaults" {
			return true
		}
	}
	return false
}

// displayTextStatus displays status in human-readable text format following design guide
func displayTextStatus(out interface{ Write([]byte) (int, error) }, status Status, iostreams interface {
	FormatSectionHeader(string) string
	FormatBoolStatus(bool, string, string) string
	FormatBold(string) string
	Indent(string, int) string
}) error {
	// Main header following design guide typography
	fmt.Fprintln(out, iostreams.FormatSectionHeader("Zen CLI Status"))
	fmt.Fprintln(out)

	// Workspace status with proper formatting and indentation
	fmt.Fprintln(out, iostreams.FormatBold("Workspace:"))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Status:      %s\n",
		iostreams.FormatBoolStatus(status.Workspace.Initialized, "Ready", "Not Initialized")), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Path:        %s\n", status.Workspace.Path), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Config File: %s\n", status.Workspace.ConfigFile), 1))
	fmt.Fprintln(out)

	// Configuration status
	fmt.Fprintln(out, iostreams.FormatBold("Configuration:"))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Status:    %s\n",
		iostreams.FormatBoolStatus(status.Configuration.Loaded, "Loaded", "Not Loaded")), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Source:    %s\n", status.Configuration.Source), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Log Level: %s\n", status.Configuration.LogLevel), 1))
	fmt.Fprintln(out)

	// System information
	fmt.Fprintln(out, iostreams.FormatBold("System:"))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("OS:           %s\n", status.System.OS), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Architecture: %s\n", status.System.Architecture), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Go Version:   %s\n", status.System.GoVersion), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("CPU Cores:    %d\n", status.System.NumCPU), 1))
	fmt.Fprintln(out)

	// Integrations
	fmt.Fprintln(out, iostreams.FormatBold("Integrations:"))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Available: %v\n", status.Integrations.Available), 1))
	fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Active:    %v\n", status.Integrations.Active), 1))

	return nil
}
