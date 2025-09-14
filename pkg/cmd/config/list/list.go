package list

import (
	"fmt"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
)

// ListOptions contains options for the list command
type ListOptions struct {
	IO     *iostreams.IOStreams
	Config func() (*config.Config, error)
}

// NewCmdConfigList creates the config list command
func NewCmdConfigList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Print a list of configuration keys and values",
		Aliases: []string{"ls"},
		Long: `List all configuration keys and their current values.

This shows the effective configuration after loading from files,
environment variables, and command-line flags.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return listRun(opts)
		},
	}

	return cmd
}

func listRun(opts *ListOptions) error {
	// Load configuration
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// List all configuration options
	for _, option := range config.Options {
		value := option.GetCurrentValue(cfg)

		// Redact sensitive values
		if config.IsSensitiveField(option.Key) {
			value = config.RedactSensitiveValue(option.Key, value)
		}

		fmt.Fprintf(opts.IO.Out, "%s=%s\n", option.Key, value)
	}

	return nil
}
