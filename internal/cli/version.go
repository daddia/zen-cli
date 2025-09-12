package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// VersionInfo contains version and build information
type VersionInfo struct {
	Version   string `json:"version" yaml:"version"`
	Commit    string `json:"commit" yaml:"commit"`
	BuildTime string `json:"build_time" yaml:"build_time"`
	GoVersion string `json:"go_version" yaml:"go_version"`
	Platform  string `json:"platform" yaml:"platform"`
}

// Version information set at build time via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// newVersionCommand creates the version command
func newVersionCommand(cfg *config.Config, logger logging.Logger) *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display version, build, and runtime information for the Zen CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			versionInfo := VersionInfo{
				Version:   version,
				Commit:    commit,
				BuildTime: buildTime,
				GoVersion: runtime.Version(),
				Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}

			if short {
				fmt.Fprintln(cmd.OutOrStdout(), versionInfo.Version)
				return nil
			}

			switch cfg.CLI.OutputFormat {
			case "json":
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(versionInfo)

			case "yaml":
				encoder := yaml.NewEncoder(cmd.OutOrStdout())
				defer encoder.Close()
				return encoder.Encode(versionInfo)

			default:
				fmt.Fprintf(cmd.OutOrStdout(), "zen version %s\n", versionInfo.Version)
				fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", versionInfo.Commit)
				fmt.Fprintf(cmd.OutOrStdout(), "  built: %s\n", versionInfo.BuildTime)
				fmt.Fprintf(cmd.OutOrStdout(), "  go version: %s\n", versionInfo.GoVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  platform: %s\n", versionInfo.Platform)
				return nil
			}
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "Show only version number")

	return cmd
}
