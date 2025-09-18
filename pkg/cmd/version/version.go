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
	var buildOptions bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long: `Display the version information for Zen CLI.

By default, shows just the version number. Use --build-options to see detailed
build information including Git commit, build date, Go version, and platform.`,
		Example: `  # Display simple version
  zen version

  # Display detailed build information
  zen version --build-options

  # Output as JSON for scripting
  zen version --build-options --output json

  # Output as YAML
  zen version --build-options --output yaml

  # Check version in scripts
  zen version --build-options --output json | jq -r '.version'`,
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

			// Simple version output (like git version)
			if !buildOptions && outputFormat == "text" {
				fmt.Fprintf(f.IOStreams.Out, "zen version %s\n", info.Version)
				return nil
			}

			// Detailed output for --build-options or non-text formats
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
				return displayDetailedVersion(f.IOStreams.Out, info)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text",
		"Output format (text, json, yaml)")
	cmd.Flags().BoolVar(&buildOptions, "build-options", false,
		"Show detailed build information")

	return cmd
}

// displayDetailedVersion displays detailed version and build information
func displayDetailedVersion(out interface{ Write([]byte) (int, error) }, info BuildInfo) error {
	// Format similar to git --build-options
	fmt.Fprintf(out, "zen version %s\n", info.Version)
	fmt.Fprintf(out, "platform: %s\n", info.Platform)
	if info.GitCommit != "" {
		fmt.Fprintf(out, "built from commit: %s\n", info.GitCommit)
	}
	if info.BuildDate != "" {
		fmt.Fprintf(out, "build date: %s\n", info.BuildDate)
	}
	fmt.Fprintf(out, "go version: %s\n", info.GoVersion)
	return nil
}
