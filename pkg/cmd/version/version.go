package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// BuildInfo contains build information
type BuildInfo struct {
	Version   string `json:"version" yaml:"version"`
	GitCommit string `json:"git_commit,omitempty" yaml:"git_commit,omitempty"`
	BuildDate string `json:"build_date,omitempty" yaml:"build_date,omitempty"`
	GoVersion string `json:"go_version" yaml:"go_version"`
	Platform  string `json:"platform" yaml:"platform"`
}

// NewCmdVersion creates the version command
func NewCmdVersion(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long: `Display the version, build information, and platform details for Zen CLI.

This command shows comprehensive version information including the release version,
build details, Git commit hash, build date, Go version used for compilation,
and target platform.`,
		Example: `  # Display version information
  zen version
  
  # Output as JSON for scripting
  zen version --output json
  
  # Output as YAML
  zen version --output yaml
  
  # Check version in scripts
  zen version --output json | jq -r '.version'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := BuildInfo{
				Version:   f.AppVersion,
				GitCommit: f.BuildInfo["commit"],
				BuildDate: f.BuildInfo["build_time"],
				GoVersion: runtime.Version(),
				Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}

			// Check parent command flags for output format
			if cmd.Parent() != nil && cmd.Parent().PersistentFlags().Changed("output") {
				if val, err := cmd.Parent().PersistentFlags().GetString("output"); err == nil {
					outputFormat = val
				}
			}

			switch outputFormat {
			case "json":
				encoder := json.NewEncoder(f.IOStreams.Out)
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)

			case "yaml":
				encoder := yaml.NewEncoder(f.IOStreams.Out)
				defer encoder.Close()
				return encoder.Encode(info)

			default:
				return displayTextVersion(f.IOStreams.Out, info)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text",
		"Output format (text, json, yaml)")

	return cmd
}

// displayTextVersion displays version in human-readable text format
func displayTextVersion(out interface{ Write([]byte) (int, error) }, info BuildInfo) error {
	// Comprehensive format with all build information
	fmt.Fprintf(out, "zen version %s\n", info.Version)
	fmt.Fprintf(out, "Build: %s\n", info.BuildDate)
	fmt.Fprintf(out, "Commit: %s\n", info.GitCommit)
	fmt.Fprintf(out, "Built: %s\n", info.BuildDate)
	fmt.Fprintf(out, "Go: %s\n", info.GoVersion)
	fmt.Fprintf(out, "Platform: %s\n", info.Platform)
	return nil
}
