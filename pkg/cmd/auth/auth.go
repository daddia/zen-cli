package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// AuthOptions contains options for the auth command
type AuthOptions struct {
	IO          *iostreams.IOStreams
	AuthManager func() (auth.Manager, error)
	Provider    string
	TokenFile   string
	Token       string
	Validate    bool
	List        bool
	Delete      bool
}

// NewCmdAuth creates the main auth command
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	opts := &AuthOptions{
		IO:          f.IOStreams,
		AuthManager: f.AuthManager,
		Validate:    true,
	}

	cmd := &cobra.Command{
		Use:   "auth [provider]",
		Short: "Authenticate with Git providers",
		Long: `Authenticate with Git providers for secure access to repositories and services.

Supported providers:
- github: GitHub Personal Access Token authentication
- gitlab: GitLab Project Access Token authentication

Authentication tokens are stored securely using your operating system's
credential manager (Keychain on macOS, Credential Manager on Windows,
Secret Service on Linux).

The auth command provides a centralized authentication system used by all
Zen CLI components that require Git provider access, including asset management,
repository operations, and future integrations.`,
		Example: heredoc.Doc(`
			# Authenticate with GitHub (interactive)
			zen auth github

			# Authenticate with GitHub using a token
			zen auth github --token ghp_your_token_here

			# Authenticate with GitHub using a token file
			zen auth github --token-file ~/.tokens/github

			# Authenticate with GitLab
			zen auth gitlab --token glpat_your_token_here

			# List all authenticated providers
			zen auth --list

			# Validate existing credentials
			zen auth github --validate

			# Delete stored credentials
			zen auth github --delete

			# Use environment variable (Zen standard)
			ZEN_GITHUB_TOKEN=ghp_token zen auth github
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Provider = strings.ToLower(args[0])
			}

			return authRun(opts)
		},
		GroupID: "core",
	}

	cmd.Flags().StringVar(&opts.TokenFile, "token-file", "", "Path to file containing authentication token")
	cmd.Flags().StringVar(&opts.Token, "token", "", "Authentication token (use environment variable for better security)")
	cmd.Flags().BoolVar(&opts.Validate, "validate", true, "Validate token after authentication")
	cmd.Flags().BoolVar(&opts.List, "list", false, "List all authenticated providers")
	cmd.Flags().BoolVar(&opts.Delete, "delete", false, "Delete stored credentials for the provider")

	return cmd
}

func authRun(opts *AuthOptions) error {
	ctx := context.Background()

	// Get auth manager
	authManager, err := opts.AuthManager()
	if err != nil {
		return errors.Wrap(err, "failed to get authentication manager")
	}

	// Handle list operation
	if opts.List {
		return listProviders(opts, authManager)
	}

	// Validate provider is specified for non-list operations
	if opts.Provider == "" {
		return fmt.Errorf("provider is required. Supported providers: %s",
			strings.Join(authManager.ListProviders(), ", "))
	}

	// Validate provider is supported
	_, err = authManager.GetProviderInfo(opts.Provider)
	if err != nil {
		return errors.Wrap(err, "invalid provider")
	}

	// Handle delete operation
	if opts.Delete {
		return deleteCredentials(opts, authManager)
	}

	// Handle authentication
	return authenticateProvider(ctx, opts, authManager)
}

func listProviders(opts *AuthOptions, authManager auth.Manager) error {
	providers := authManager.ListProviders()

	fmt.Fprintf(opts.IO.Out, "Configured authentication providers:\n\n")

	for _, provider := range providers {
		// Check authentication status
		isAuth := authManager.IsAuthenticated(context.Background(), provider)
		status := "✗ Not authenticated"
		if isAuth {
			status = "✓ Authenticated"
		}

		// Get provider info
		info, err := authManager.GetProviderInfo(provider)
		if err == nil {
			fmt.Fprintf(opts.IO.Out, "  %s: %s\n", provider, status)
			fmt.Fprintf(opts.IO.Out, "    Description: %s\n", info.Description)
			fmt.Fprintf(opts.IO.Out, "    Environment variables: %s\n", strings.Join(info.EnvVars, ", "))
		} else {
			fmt.Fprintf(opts.IO.Out, "  %s: %s\n", provider, status)
		}
		fmt.Fprintln(opts.IO.Out)
	}

	return nil
}

func deleteCredentials(opts *AuthOptions, authManager auth.Manager) error {
	err := authManager.DeleteCredentials(opts.Provider)
	if err != nil {
		return errors.Wrap(err, "failed to delete credentials")
	}

	fmt.Fprintf(opts.IO.Out, "✓ Deleted credentials for %s\n", opts.Provider)
	return nil
}

func authenticateProvider(ctx context.Context, opts *AuthOptions, authManager auth.Manager) error {
	// Get authentication token
	token, err := getAuthToken(opts, authManager)
	if err != nil {
		return errors.Wrap(err, "failed to get authentication token")
	}

	// Store the token by creating a credential and storing it
	// This is a bridge until we can refactor the auth manager interface
	if token != "" {
		// For now, we'll set the environment variable so the auth manager can pick it up
		envVar := getEnvVarName(opts.Provider)
		if envVar != "" {
			// Temporarily set the environment variable for the auth manager to use
			originalValue := getEnvVar(envVar)
			setEnvVar(envVar, token)
			defer func() {
				if originalValue != "" {
					setEnvVar(envVar, originalValue)
				} else {
					unsetEnvVar(envVar)
				}
			}()
		}
	}

	// Authenticate with the provider
	if err := authManager.Authenticate(ctx, opts.Provider); err != nil {
		return errors.Wrap(err, "authentication failed")
	}

	// Success message
	fmt.Fprintf(opts.IO.Out, "✓ Successfully authenticated with %s\n",
		cases.Title(language.English).String(opts.Provider))

	// Show additional info
	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out, "Authentication token stored securely in system credential manager.\n")
		fmt.Fprintf(opts.IO.Out, "Use 'zen auth --list' to check authentication status.\n")
	}

	return nil
}

// Helper functions (similar to assets auth but updated for main command)

func getAuthToken(opts *AuthOptions, authManager auth.Manager) (string, error) {
	// Priority: explicit token > token file > environment variable > interactive prompt

	if opts.Token != "" {
		return opts.Token, nil
	}

	if opts.TokenFile != "" {
		return readTokenFromFile(opts.TokenFile)
	}

	// Check if already available via environment
	if token, err := authManager.GetCredentials(opts.Provider); err == nil && token != "" {
		return token, nil
	}

	// Interactive prompt
	return promptForToken(opts)
}

func readTokenFromFile(filepath string) (string, error) {
	// Implementation would read token from file
	return "", fmt.Errorf("token file reading not yet implemented")
}

func promptForToken(opts *AuthOptions) (string, error) {
	if !opts.IO.CanPrompt() {
		return "", fmt.Errorf("authentication token required but prompting is disabled")
	}

	// Get provider info for instructions
	authManager, err := opts.AuthManager()
	if err != nil {
		return "", err
	}

	info, err := authManager.GetProviderInfo(opts.Provider)
	if err != nil {
		return "", err
	}

	// Show instructions
	fmt.Fprintf(opts.IO.Out, "⚠ Authentication required for %s\n\n", cases.Title(language.English).String(opts.Provider))
	fmt.Fprintf(opts.IO.Out, "Instructions: %s:\n", info.Description)
	for _, instruction := range info.Instructions {
		fmt.Fprintf(opts.IO.Out, "%s\n", instruction)
	}
	fmt.Fprintln(opts.IO.Out)

	// Prompt for token
	fmt.Fprintf(opts.IO.Out, "Enter your %s token: ", cases.Title(language.English).String(opts.Provider))

	var token string
	_, err = fmt.Fscanln(opts.IO.In, &token)
	if err != nil {
		return "", errors.Wrap(err, "failed to read token")
	}

	return token, nil
}

func getEnvVarName(provider string) string {
	switch provider {
	case "github":
		return "ZEN_GITHUB_TOKEN"
	case "gitlab":
		return "ZEN_GITLAB_TOKEN"
	default:
		return fmt.Sprintf("ZEN_%s_TOKEN", strings.ToUpper(provider))
	}
}

func getEnvVar(name string) string {
	return os.Getenv(name)
}

func setEnvVar(name, value string) {
	os.Setenv(name, value)
}

func unsetEnvVar(name string) {
	os.Unsetenv(name)
}
