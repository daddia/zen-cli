package assets

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/errors"
)

// TokenAuthProvider implements AuthProvider using token-based authentication
type TokenAuthProvider struct {
	logger logging.Logger
	tokens map[string]string // provider -> token
}

// NewTokenAuthProvider creates a new token-based authentication provider
func NewTokenAuthProvider(logger logging.Logger) *TokenAuthProvider {
	return &TokenAuthProvider{
		logger: logger,
		tokens: make(map[string]string),
	}
}

// Authenticate authenticates with the Git provider
func (a *TokenAuthProvider) Authenticate(ctx context.Context, provider string) error {
	a.logger.Debug("authenticating with provider", "provider", provider)

	token, err := a.GetCredentials(provider)
	if err != nil {
		return err
	}

	if token == "" {
		return &AssetClientError{
			Code:    ErrorCodeAuthenticationFailed,
			Message: fmt.Sprintf("no authentication token found for provider '%s'", provider),
			Details: a.getTokenInstructions(provider),
		}
	}

	// Validate token by making a test API call
	if err := a.ValidateCredentials(ctx, provider); err != nil {
		return err
	}

	a.logger.Debug("authentication successful", "provider", provider)
	return nil
}

// GetCredentials returns credentials for the specified provider
func (a *TokenAuthProvider) GetCredentials(provider string) (string, error) {
	// Check if token is already cached
	if token, exists := a.tokens[provider]; exists && token != "" {
		return token, nil
	}

	// Try to load from environment variables
	token := a.getTokenFromEnv(provider)
	if token != "" {
		a.tokens[provider] = token
		return token, nil
	}

	// Try to load from configuration file
	token, err := a.getTokenFromConfig(provider)
	if err != nil {
		return "", errors.Wrap(err, "failed to load token from config")
	}

	if token != "" {
		a.tokens[provider] = token
		return token, nil
	}

	return "", &AssetClientError{
		Code:    ErrorCodeAuthenticationFailed,
		Message: fmt.Sprintf("no authentication token found for provider '%s'", provider),
		Details: a.getTokenInstructions(provider),
	}
}

// ValidateCredentials validates stored credentials
func (a *TokenAuthProvider) ValidateCredentials(ctx context.Context, provider string) error {
	token, err := a.GetCredentials(provider)
	if err != nil {
		return err
	}

	switch provider {
	case "github":
		return a.validateGitHubToken(ctx, token)
	case "gitlab":
		return a.validateGitLabToken(ctx, token)
	default:
		return &AssetClientError{
			Code:    ErrorCodeConfigurationError,
			Message: fmt.Sprintf("unsupported provider '%s'", provider),
		}
	}
}

// RefreshCredentials refreshes expired credentials if possible
func (a *TokenAuthProvider) RefreshCredentials(ctx context.Context, provider string) error {
	// For token-based auth, we can't refresh automatically
	// User needs to provide a new token
	return &AssetClientError{
		Code:    ErrorCodeAuthenticationFailed,
		Message: fmt.Sprintf("token refresh not supported for provider '%s'", provider),
		Details: a.getTokenInstructions(provider),
	}
}

// Private helper methods

func (a *TokenAuthProvider) getTokenFromEnv(provider string) string {
	envVars := a.getEnvVarNames(provider)

	for _, envVar := range envVars {
		if token := os.Getenv(envVar); token != "" {
			a.logger.Debug("token loaded from environment", "provider", provider, "env_var", envVar)
			return token
		}
	}

	return ""
}

func (a *TokenAuthProvider) getTokenFromConfig(provider string) (string, error) {
	// This would integrate with the config system
	// For now, return empty to use environment variables
	return "", nil
}

func (a *TokenAuthProvider) getEnvVarNames(provider string) []string {
	switch provider {
	case "github":
		return []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN"}
	case "gitlab":
		return []string{"GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"}
	default:
		return []string{fmt.Sprintf("ZEN_%s_TOKEN", strings.ToUpper(provider))}
	}
}

func (a *TokenAuthProvider) validateGitHubToken(ctx context.Context, token string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make a test API call to validate the token
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return errors.Wrap(err, "failed to create validation request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("User-Agent", "zen-cli/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return &AssetClientError{
			Code:    ErrorCodeNetworkError,
			Message: "failed to validate GitHub token",
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		a.logger.Debug("GitHub token validation successful")
		return nil
	case http.StatusUnauthorized:
		return &AssetClientError{
			Code:    ErrorCodeAuthenticationFailed,
			Message: "GitHub token is invalid or expired",
			Details: "Please check your token and ensure it has the required scopes",
		}
	case http.StatusForbidden:
		return &AssetClientError{
			Code:    ErrorCodeAuthenticationFailed,
			Message: "GitHub token lacks required permissions",
			Details: "Token needs 'repo' scope for private repositories",
		}
	case http.StatusTooManyRequests:
		return &AssetClientError{
			Code:       ErrorCodeRateLimited,
			Message:    "GitHub API rate limit exceeded",
			RetryAfter: 3600, // 1 hour
		}
	default:
		return &AssetClientError{
			Code:    ErrorCodeNetworkError,
			Message: fmt.Sprintf("unexpected response from GitHub API: %d", resp.StatusCode),
		}
	}
}

func (a *TokenAuthProvider) validateGitLabToken(ctx context.Context, token string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make a test API call to validate the token
	req, err := http.NewRequestWithContext(ctx, "GET", "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return errors.Wrap(err, "failed to create validation request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "zen-cli/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return &AssetClientError{
			Code:    ErrorCodeNetworkError,
			Message: "failed to validate GitLab token",
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		a.logger.Debug("GitLab token validation successful")
		return nil
	case http.StatusUnauthorized:
		return &AssetClientError{
			Code:    ErrorCodeAuthenticationFailed,
			Message: "GitLab token is invalid or expired",
			Details: "Please check your token and ensure it has the required scopes",
		}
	case http.StatusForbidden:
		return &AssetClientError{
			Code:    ErrorCodeAuthenticationFailed,
			Message: "GitLab token lacks required permissions",
			Details: "Token needs 'read_repository' scope for private repositories",
		}
	case http.StatusTooManyRequests:
		return &AssetClientError{
			Code:       ErrorCodeRateLimited,
			Message:    "GitLab API rate limit exceeded",
			RetryAfter: 3600, // 1 hour
		}
	default:
		return &AssetClientError{
			Code:    ErrorCodeNetworkError,
			Message: fmt.Sprintf("unexpected response from GitLab API: %d", resp.StatusCode),
		}
	}
}

func (a *TokenAuthProvider) getTokenInstructions(provider string) interface{} {
	switch provider {
	case "github":
		return map[string]interface{}{
			"message": "GitHub Personal Access Token required",
			"instructions": []string{
				"1. Go to https://github.com/settings/tokens",
				"2. Generate a new token with 'repo' scope",
				"3. Set the token in environment variable: export GITHUB_TOKEN=your_token",
				"4. Or use: zen config set github.token your_token",
			},
			"env_vars": []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN"},
		}
	case "gitlab":
		return map[string]interface{}{
			"message": "GitLab Project Access Token required",
			"instructions": []string{
				"1. Go to your GitLab project settings",
				"2. Create a Project Access Token with 'read_repository' scope",
				"3. Set the token in environment variable: export GITLAB_TOKEN=your_token",
				"4. Or use: zen config set gitlab.token your_token",
			},
			"env_vars": []string{"GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"},
		}
	default:
		return map[string]interface{}{
			"message":  fmt.Sprintf("Authentication token required for provider '%s'", provider),
			"env_vars": []string{fmt.Sprintf("ZEN_%s_TOKEN", strings.ToUpper(provider))},
		}
	}
}
