package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/pkg/errors"
)

// TokenManager implements Manager using token-based authentication
type TokenManager struct {
	config  Config
	logger  logging.Logger
	storage CredentialStorage

	// In-memory cache for performance
	mu    sync.RWMutex
	cache map[string]*Credential
}

// NewTokenManager creates a new token-based authentication manager
func NewTokenManager(config Config, logger logging.Logger, storage CredentialStorage) *TokenManager {
	return &TokenManager{
		config:  config,
		logger:  logger,
		storage: storage,
		cache:   make(map[string]*Credential),
	}
}

// Authenticate authenticates with the specified provider
func (t *TokenManager) Authenticate(ctx context.Context, provider string) error {
	t.logger.Debug("authenticating with provider", "provider", provider)

	// Get or create credential
	credential, err := t.getOrCreateCredential(ctx, provider)
	if err != nil {
		return err
	}

	if credential.Token == "" {
		return NewAuthError(
			ErrorCodeAuthenticationFailed,
			fmt.Sprintf("no authentication token found for provider '%s'", provider),
			provider,
		)
	}

	// Validate token by making a test API call
	if err := t.validateToken(ctx, provider, credential.Token); err != nil {
		return err
	}

	// Update last validated time and store
	credential.LastValidated = &time.Time{}
	*credential.LastValidated = time.Now()
	credential.LastUsed = time.Now()

	if err := t.storage.Store(ctx, provider, credential); err != nil {
		t.logger.Warn("failed to update credential storage", "provider", provider, "error", err)
	}

	// Update cache
	t.mu.Lock()
	t.cache[provider] = credential
	t.mu.Unlock()

	t.logger.Debug("authentication successful", "provider", provider)
	return nil
}

// GetCredentials returns credentials for the specified provider
func (t *TokenManager) GetCredentials(provider string) (string, error) {
	// Check cache first
	t.mu.RLock()
	if cached, exists := t.cache[provider]; exists && cached.IsValid() {
		t.mu.RUnlock()
		return cached.Token, nil
	}
	t.mu.RUnlock()

	// Try to load from storage
	ctx := context.Background()
	credential, err := t.storage.Retrieve(ctx, provider)
	if err == nil && credential != nil && credential.IsValid() {
		// Update cache
		t.mu.Lock()
		t.cache[provider] = credential
		t.mu.Unlock()
		return credential.Token, nil
	}

	// Try environment variables
	if token := t.getTokenFromEnv(provider); token != "" {
		// Create and cache credential
		credential := &Credential{
			Provider:  provider,
			Token:     token,
			Type:      "token",
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}

		t.mu.Lock()
		t.cache[provider] = credential
		t.mu.Unlock()

		// Store for future use
		if err := t.storage.Store(ctx, provider, credential); err != nil {
			t.logger.Warn("failed to store credential from environment", "provider", provider, "error", err)
		}

		return token, nil
	}

	return "", NewAuthError(
		ErrorCodeCredentialNotFound,
		fmt.Sprintf("no authentication token found for provider '%s'", provider),
		provider,
	)
}

// ValidateCredentials validates stored credentials for the provider
func (t *TokenManager) ValidateCredentials(ctx context.Context, provider string) error {
	token, err := t.GetCredentials(provider)
	if err != nil {
		return err
	}

	return t.validateToken(ctx, provider, token)
}

// RefreshCredentials refreshes expired credentials if possible
func (t *TokenManager) RefreshCredentials(ctx context.Context, provider string) error {
	// For token-based auth, we can't refresh automatically
	// User needs to provide a new token
	return NewAuthError(
		ErrorCodeAuthenticationFailed,
		fmt.Sprintf("token refresh not supported for provider '%s'", provider),
		provider,
	)
}

// IsAuthenticated checks if credentials are available and valid for the provider
func (t *TokenManager) IsAuthenticated(ctx context.Context, provider string) bool {
	// Quick check from cache
	t.mu.RLock()
	if cached, exists := t.cache[provider]; exists && cached.IsValid() {
		// Check if recently validated (within cache timeout)
		if cached.LastValidated != nil && time.Since(*cached.LastValidated) < t.config.CacheTimeout {
			t.mu.RUnlock()
			return true
		}
	}
	t.mu.RUnlock()

	// Try to get credentials (this will check storage and env vars)
	token, err := t.GetCredentials(provider)
	if err != nil || token == "" {
		return false
	}

	// For a quick check, we don't validate against the API
	// Full validation happens in ValidateCredentials
	return true
}

// ListProviders returns all configured providers
func (t *TokenManager) ListProviders() []string {
	providers := make([]string, 0, len(t.config.Providers))
	for provider := range t.config.Providers {
		providers = append(providers, provider)
	}
	return providers
}

// DeleteCredentials removes stored credentials for the provider
func (t *TokenManager) DeleteCredentials(provider string) error {
	ctx := context.Background()

	// Remove from cache
	t.mu.Lock()
	delete(t.cache, provider)
	t.mu.Unlock()

	// Remove from storage
	return t.storage.Delete(ctx, provider)
}

// GetProviderInfo returns information about the provider
func (t *TokenManager) GetProviderInfo(provider string) (*ProviderInfo, error) {
	config, exists := t.config.Providers[provider]
	if !exists {
		return nil, NewAuthError(
			ErrorCodeProviderNotSupported,
			fmt.Sprintf("provider '%s' is not supported", provider),
			provider,
		)
	}

	return &ProviderInfo{
		Name:         provider,
		Type:         config.Type,
		Description:  t.getProviderDescription(provider),
		Instructions: t.getProviderInstructions(provider),
		EnvVars:      config.EnvVars,
		ConfigKeys:   config.ConfigKeys,
		Scopes:       config.Scopes,
		Metadata:     config.Metadata,
	}, nil
}

// Private helper methods

func (t *TokenManager) getOrCreateCredential(ctx context.Context, provider string) (*Credential, error) {
	// Try to get from storage first
	credential, err := t.storage.Retrieve(ctx, provider)
	if err == nil && credential != nil {
		return credential, nil
	}

	// Try environment variables
	if token := t.getTokenFromEnv(provider); token != "" {
		return &Credential{
			Provider:  provider,
			Token:     token,
			Type:      "token",
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}, nil
	}

	return nil, NewAuthError(
		ErrorCodeCredentialNotFound,
		fmt.Sprintf("no authentication token found for provider '%s'", provider),
		provider,
	)
}

func (t *TokenManager) getTokenFromEnv(provider string) string {
	config, exists := t.config.Providers[provider]
	if !exists {
		return ""
	}

	for _, envVar := range config.EnvVars {
		if token := os.Getenv(envVar); token != "" {
			t.logger.Debug("token loaded from environment", "provider", provider, "env_var", envVar)
			return token
		}
	}

	return ""
}

func (t *TokenManager) validateToken(ctx context.Context, provider, token string) error {
	config, exists := t.config.Providers[provider]
	if !exists {
		return NewAuthError(
			ErrorCodeProviderNotSupported,
			fmt.Sprintf("provider '%s' is not supported", provider),
			provider,
		)
	}

	switch provider {
	case "github":
		return t.validateGitHubToken(ctx, token, config.BaseURL)
	case "gitlab":
		return t.validateGitLabToken(ctx, token, config.BaseURL)
	default:
		return NewAuthError(
			ErrorCodeProviderNotSupported,
			fmt.Sprintf("validation not implemented for provider '%s'", provider),
			provider,
		)
	}
}

func (t *TokenManager) validateGitHubToken(ctx context.Context, token, baseURL string) error {
	client := &http.Client{
		Timeout: t.config.ValidationTimeout,
	}

	url := baseURL + "/user"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create validation request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("User-Agent", "zen-cli/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return NewNetworkError("github", err.Error())
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		t.logger.Debug("GitHub token validation successful")
		return nil
	case http.StatusUnauthorized:
		return NewAuthError(
			ErrorCodeInvalidCredentials,
			"GitHub token is invalid or expired",
			"github",
		)
	case http.StatusForbidden:
		return NewAuthError(
			ErrorCodeInsufficientScopes,
			"GitHub token lacks required permissions",
			"github",
		)
	case http.StatusTooManyRequests:
		return NewRateLimitError("github", 3600)
	default:
		return NewNetworkError("github", fmt.Sprintf("unexpected response: %d", resp.StatusCode))
	}
}

func (t *TokenManager) validateGitLabToken(ctx context.Context, token, baseURL string) error {
	client := &http.Client{
		Timeout: t.config.ValidationTimeout,
	}

	url := baseURL + "/user"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create validation request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "zen-cli/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return NewNetworkError("gitlab", err.Error())
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		t.logger.Debug("GitLab token validation successful")
		return nil
	case http.StatusUnauthorized:
		return NewAuthError(
			ErrorCodeInvalidCredentials,
			"GitLab token is invalid or expired",
			"gitlab",
		)
	case http.StatusForbidden:
		return NewAuthError(
			ErrorCodeInsufficientScopes,
			"GitLab token lacks required permissions",
			"gitlab",
		)
	case http.StatusTooManyRequests:
		return NewRateLimitError("gitlab", 3600)
	default:
		return NewNetworkError("gitlab", fmt.Sprintf("unexpected response: %d", resp.StatusCode))
	}
}

func (t *TokenManager) getProviderDescription(provider string) string {
	switch provider {
	case "github":
		return "GitHub Personal Access Token authentication"
	case "gitlab":
		return "GitLab Project Access Token authentication"
	default:
		return fmt.Sprintf("Token-based authentication for %s", provider)
	}
}

func (t *TokenManager) getProviderInstructions(provider string) []string {
	switch provider {
	case "github":
		return []string{
			"1. Go to https://github.com/settings/tokens",
			"2. Click 'Generate new token (classic)'",
			"3. Select 'repo' scope for private repositories",
			"4. Copy the generated token",
			"5. Set environment variable: export ZEN_GITHUB_TOKEN=your_token",
			"6. Or use: zen config set github.token your_token",
		}
	case "gitlab":
		return []string{
			"1. Go to your GitLab project settings",
			"2. Navigate to Access Tokens",
			"3. Create a Project Access Token with 'read_repository' scope",
			"4. Copy the generated token",
			"5. Set environment variable: export ZEN_GITLAB_TOKEN=your_token",
			"6. Or use: zen config set gitlab.token your_token",
		}
	default:
		return []string{
			fmt.Sprintf("1. Obtain a token for %s", provider),
			fmt.Sprintf("2. Set environment variable: export %s_TOKEN=your_token", strings.ToUpper(provider)),
		}
	}
}
