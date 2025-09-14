package get

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
)

// GetOptions contains options for the get command
type GetOptions struct {
	IO     *iostreams.IOStreams
	Config func() (*config.Config, error)
	Key    string
}

// NewCmdConfigGet creates the config get command
func NewCmdConfigGet(f *cmdutil.Factory, runF func(*GetOptions) error) *cobra.Command {
	opts := &GetOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value of a given configuration key",
		Long: `Print the value of a configuration key.

Configuration keys use dot notation to access nested values:
- log_level
- log_format
- cli.no_color
- cli.verbose
- cli.output_format
- workspace.root
- workspace.config_file
- development.debug
- development.profile`,
		Example: heredoc.Doc(`
			$ zen config get log_level
			$ zen config get cli.output_format
			$ zen config get workspace.root
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]

			if runF != nil {
				return runF(opts)
			}

			return getRun(opts)
		},
	}

	return cmd
}

func getRun(opts *GetOptions) error {
	// Validate the key exists
	if err := config.ValidateKey(opts.Key); err != nil {
		return fmt.Errorf("invalid configuration key: %w", err)
	}

	// Load configuration
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find the configuration option
	opt, found := config.FindOption(opts.Key)
	if !found {
		return &nonExistentKeyError{key: opts.Key}
	}

	// Get current value
	value := opt.GetCurrentValue(cfg)
	if value == "" {
		return &nonExistentKeyError{key: opts.Key}
	}

	// Redact sensitive values
	if config.IsSensitiveField(opts.Key) {
		value = config.RedactSensitiveValue(opts.Key, value)
	}

	fmt.Fprintf(opts.IO.Out, "%s\n", value)
	return nil
}

type nonExistentKeyError struct {
	key string
}

func (e nonExistentKeyError) Error() string {
	return fmt.Sprintf("could not find configuration key %q", e.key)
}
