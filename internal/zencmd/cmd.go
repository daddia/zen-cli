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

	// Format error message with emoji for better visibility
	fmt.Fprintf(out, "❌ Error: %s\n", err.Error())

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
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "config") && strings.Contains(errMsg, "not found"):
		return "💡 Suggestion: Run 'zen config' to check your configuration or 'zen init' to initialize a workspace"
	case strings.Contains(errMsg, "config") && strings.Contains(errMsg, "invalid"):
		return "💡 Suggestion: Check your configuration file syntax with 'zen config validate'"
	case strings.Contains(errMsg, "workspace") && strings.Contains(errMsg, "not found"):
		return "💡 Suggestion: Run 'zen init' to initialize a new workspace in this directory"
	case strings.Contains(errMsg, "workspace") && strings.Contains(errMsg, "invalid"):
		return "💡 Suggestion: Check workspace structure with 'zen status' or reinitialize with 'zen init --force'"
	case strings.Contains(errMsg, "permission"):
		return "💡 Suggestion: Check file permissions or try running with appropriate privileges"
	case strings.Contains(errMsg, "unknown flag"):
		return "💡 Suggestion: Use 'zen --help' to see available flags and options"
	case strings.Contains(errMsg, "unknown command"):
		return "💡 Suggestion: Use 'zen --help' to see available commands"
	case strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection"):
		return "💡 Suggestion: Check your internet connection and try again"
	case strings.Contains(errMsg, "timeout"):
		return "💡 Suggestion: The operation timed out. Try again or check network connectivity"
	case strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "auth"):
		return "💡 Suggestion: Check your credentials or run authentication setup"
	default:
		return "💡 Suggestion: Use 'zen --help' for usage information or check the documentation at https://zen.dev/docs"
	}
}
