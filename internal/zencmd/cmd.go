package zencmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jonathandaddia/zen/pkg/cmd/factory"
	"github.com/jonathandaddia/zen/pkg/cmd/root"
	"github.com/jonathandaddia/zen/pkg/cmdutil"
)

// Main is the main entry point for the Zen CLI
func Main() cmdutil.ExitCode {
	// Setup graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create factory
	cmdFactory := factory.New()
	stderr := cmdFactory.IOStreams.ErrOut

	// Create root command
	rootCmd, err := root.NewCmdRoot(cmdFactory)
	if err != nil {
		fmt.Fprintf(stderr, "failed to create root command: %s\n", err)
		return cmdutil.ExitError
	}

	// Set context
	rootCmd.SetContext(ctx)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		return handleError(err, cmdFactory)
	}

	return cmdutil.ExitOK
}

func handleError(err error, f *cmdutil.Factory) cmdutil.ExitCode {
	stderr := f.IOStreams.ErrOut

	// Check for specific error types
	if err == cmdutil.SilentError {
		return cmdutil.ExitError
	}

	if err == cmdutil.PendingError {
		return cmdutil.ExitError
	}

	if cmdutil.IsUserCancellation(err) {
		fmt.Fprint(stderr, "\n")
		return cmdutil.ExitCancel
	}

	var noResultsError cmdutil.NoResultsError
	if errors.As(err, &noResultsError) {
		if f.IOStreams.IsStdoutTTY() {
			fmt.Fprintln(stderr, noResultsError.Error())
		}
		return cmdutil.ExitOK
	}

	// Print the error
	printError(stderr, err)

	// Check for flag errors
	var flagError *cmdutil.FlagError
	if errors.As(err, &flagError) {
		return cmdutil.ExitError
	}

	return cmdutil.ExitError
}

func printError(out io.Writer, err error) {
	if err == nil {
		return
	}

	fmt.Fprintln(out, err)

	// Add helpful suggestions based on error type
	if msg := getErrorSuggestion(err); msg != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, msg)
	}
}

func getErrorSuggestion(err error) string {
	if err == nil {
		return ""
	}

	// Add error-specific suggestions here
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "config"):
		return "Try running 'zen config' to check your configuration"
	case strings.Contains(errMsg, "workspace"):
		return "Try running 'zen init' to initialize your workspace"
	case strings.Contains(errMsg, "permission"):
		return "Check file permissions and try again"
	default:
		return ""
	}
}
