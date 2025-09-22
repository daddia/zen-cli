package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/integration/plugin"
)

// AuthAbstraction provides unified authentication interface for plugins
type AuthAbstraction struct {
	logger    logging.Logger
	authMgr   auth.Manager
	providers map[string]AuthProviderInterface
	tokens    map[string]*TokenInfo
	mu        sync.RWMutex
}

// AuthProviderInterface defines authentication provider interface
type AuthProviderInterface interface {
	// Name returns the provider name
	Name() string

	// Authenticate performs authentication and returns tokens
	Authenticate(ctx context.Context, config *plugin.AuthConfig) (*TokenInfo, error)

	// RefreshToken refreshes an expired token
	RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error)

	// ValidateToken validates a token
	ValidateToken(ctx context.Context, token string) error

	// AddAuthToRequest adds authentication to an HTTP request
	AddAuthToRequest(req *http.Request, tokenInfo *TokenInfo) error

	// GetTokenExpiry returns when the token expires
	GetTokenExpiry(tokenInfo *TokenInfo) time.Time

	// SupportsRefresh returns true if the provider supports token refresh
	SupportsRefresh() bool
}

// CredentialManagerInterface defines credential management interface
type CredentialManagerInterface interface {
	// StoreCredentials stores credentials securely
	StoreCredentials(provider string, credentials *CredentialData) error

	// GetCredentials retrieves stored credentials
	GetCredentials(provider string) (*CredentialData, error)

	// DeleteCredentials deletes stored credentials
	DeleteCredentials(provider string) error

	// ListProviders returns all providers with stored credentials
	ListProviders() []string

	// RotateCredentials rotates credentials for a provider
	RotateCredentials(ctx context.Context, provider string) error
}

// TokenRefreshInterface defines token refresh interface
type TokenRefreshInterface interface {
	// RefreshToken refreshes an expired token
	RefreshToken(ctx context.Context, provider string) (*TokenInfo, error)

	// ScheduleRefresh schedules automatic token refresh
	ScheduleRefresh(provider string, refreshBefore time.Duration) error

	// CancelRefresh cancels scheduled token refresh
	CancelRefresh(provider string) error

	// IsRefreshScheduled checks if refresh is scheduled for a provider
	IsRefreshScheduled(provider string) bool
}

// TokenInfo contains token information
type TokenInfo struct {
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                 `json:"refresh_token,omitempty"`
	TokenType    string                 `json:"token_type"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Scopes       []string               `json:"scopes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CredentialData contains credential information
type CredentialData struct {
	Type         plugin.AuthType   `json:"type"`
	Username     string            `json:"username,omitempty"`
	Password     string            `json:"password,omitempty"`
	APIKey       string            `json:"api_key,omitempty"`
	ClientID     string            `json:"client_id,omitempty"`
	ClientSecret string            `json:"client_secret,omitempty"`
	TokenInfo    *TokenInfo        `json:"token_info,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewAuthAbstraction creates a new authentication abstraction layer
func NewAuthAbstraction(logger logging.Logger, authMgr auth.Manager) *AuthAbstraction {
	aa := &AuthAbstraction{
		logger:    logger,
		authMgr:   authMgr,
		providers: make(map[string]AuthProviderInterface),
		tokens:    make(map[string]*TokenInfo),
	}

	// Register built-in auth providers
	aa.registerBuiltinProviders()

	return aa
}

// registerBuiltinProviders registers built-in authentication providers
func (aa *AuthAbstraction) registerBuiltinProviders() {
	aa.providers["oauth2"] = &OAuth2Provider{logger: aa.logger}
	aa.providers["basic"] = &BasicAuthProvider{logger: aa.logger}
	aa.providers["api_key"] = &APIKeyProvider{logger: aa.logger}
	aa.providers["bearer"] = &BearerTokenProvider{logger: aa.logger}
}

// GetAuthProvider returns an authentication provider by type
func (aa *AuthAbstraction) GetAuthProvider(authType plugin.AuthType) (AuthProviderInterface, error) {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	provider, exists := aa.providers[string(authType)]
	if !exists {
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	return provider, nil
}

// AuthenticatePlugin authenticates a plugin using its auth configuration
func (aa *AuthAbstraction) AuthenticatePlugin(ctx context.Context, pluginName string, authConfig *plugin.AuthConfig) (*TokenInfo, error) {
	aa.logger.Debug("authenticating plugin", "plugin", pluginName, "auth_type", authConfig.Type)

	provider, err := aa.GetAuthProvider(authConfig.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth provider: %w", err)
	}

	tokenInfo, err := provider.Authenticate(ctx, authConfig)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Store token for future use
	aa.mu.Lock()
	aa.tokens[pluginName] = tokenInfo
	aa.mu.Unlock()

	aa.logger.Info("plugin authenticated successfully", "plugin", pluginName, "expires_at", tokenInfo.ExpiresAt)

	return tokenInfo, nil
}

// GetPluginToken returns the stored token for a plugin
func (aa *AuthAbstraction) GetPluginToken(pluginName string) (*TokenInfo, error) {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	tokenInfo, exists := aa.tokens[pluginName]
	if !exists {
		return nil, fmt.Errorf("no token found for plugin: %s", pluginName)
	}

	// Check if token is expired
	if time.Now().After(tokenInfo.ExpiresAt) {
		return nil, fmt.Errorf("token expired for plugin: %s", pluginName)
	}

	return tokenInfo, nil
}

// AddAuthToRequest adds authentication to an HTTP request for a plugin
func (aa *AuthAbstraction) AddAuthToRequest(req *http.Request, pluginName string, authConfig *plugin.AuthConfig) error {
	provider, err := aa.GetAuthProvider(authConfig.Type)
	if err != nil {
		return fmt.Errorf("failed to get auth provider: %w", err)
	}

	tokenInfo, err := aa.GetPluginToken(pluginName)
	if err != nil {
		// Try to authenticate if no valid token
		tokenInfo, err = aa.AuthenticatePlugin(req.Context(), pluginName, authConfig)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	return provider.AddAuthToRequest(req, tokenInfo)
}

// RefreshPluginToken refreshes a plugin's token
func (aa *AuthAbstraction) RefreshPluginToken(ctx context.Context, pluginName string, authConfig *plugin.AuthConfig) (*TokenInfo, error) {
	provider, err := aa.GetAuthProvider(authConfig.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth provider: %w", err)
	}

	if !provider.SupportsRefresh() {
		return nil, fmt.Errorf("provider does not support token refresh: %s", authConfig.Type)
	}

	aa.mu.RLock()
	currentToken, exists := aa.tokens[pluginName]
	aa.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no token to refresh for plugin: %s", pluginName)
	}

	newToken, err := provider.RefreshToken(ctx, currentToken)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Store new token
	aa.mu.Lock()
	aa.tokens[pluginName] = newToken
	aa.mu.Unlock()

	aa.logger.Info("token refreshed successfully", "plugin", pluginName, "expires_at", newToken.ExpiresAt)

	return newToken, nil
}

// Built-in authentication providers

// OAuth2Provider implements OAuth 2.0 authentication
type OAuth2Provider struct {
	logger logging.Logger
}

func (op *OAuth2Provider) Name() string { return "oauth2" }

func (op *OAuth2Provider) Authenticate(ctx context.Context, config *plugin.AuthConfig) (*TokenInfo, error) {
	// Placeholder implementation for OAuth 2.0 flow
	// In production, this would implement the full OAuth 2.0 flow
	return &TokenInfo{
		AccessToken: "oauth2_access_token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil
}

func (op *OAuth2Provider) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	// Placeholder implementation for token refresh
	return &TokenInfo{
		AccessToken: "refreshed_oauth2_token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil
}

func (op *OAuth2Provider) ValidateToken(ctx context.Context, token string) error {
	// Placeholder implementation
	return nil
}

func (op *OAuth2Provider) AddAuthToRequest(req *http.Request, tokenInfo *TokenInfo) error {
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenInfo.TokenType, tokenInfo.AccessToken))
	return nil
}

func (op *OAuth2Provider) GetTokenExpiry(tokenInfo *TokenInfo) time.Time {
	return tokenInfo.ExpiresAt
}

func (op *OAuth2Provider) SupportsRefresh() bool {
	return true
}

// BasicAuthProvider implements basic authentication
type BasicAuthProvider struct {
	logger logging.Logger
}

func (bp *BasicAuthProvider) Name() string { return "basic" }

func (bp *BasicAuthProvider) Authenticate(ctx context.Context, config *plugin.AuthConfig) (*TokenInfo, error) {
	// For basic auth, we just validate that credentials are available
	username := config.CustomFields["username"]
	password := config.CustomFields["password"]
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password required for basic auth")
	}

	return &TokenInfo{
		AccessToken: username + ":" + password,
		TokenType:   "Basic",
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Basic auth doesn't expire, but set a long expiry
	}, nil
}

func (bp *BasicAuthProvider) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	return tokenInfo, nil // Basic auth doesn't need refresh
}

func (bp *BasicAuthProvider) ValidateToken(ctx context.Context, token string) error {
	return nil // Basic auth is always valid if credentials are correct
}

func (bp *BasicAuthProvider) AddAuthToRequest(req *http.Request, tokenInfo *TokenInfo) error {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", tokenInfo.AccessToken))
	return nil
}

func (bp *BasicAuthProvider) GetTokenExpiry(tokenInfo *TokenInfo) time.Time {
	return tokenInfo.ExpiresAt
}

func (bp *BasicAuthProvider) SupportsRefresh() bool {
	return false
}

// APIKeyProvider implements API key authentication
type APIKeyProvider struct {
	logger logging.Logger
}

func (ap *APIKeyProvider) Name() string { return "api_key" }

func (ap *APIKeyProvider) Authenticate(ctx context.Context, config *plugin.AuthConfig) (*TokenInfo, error) {
	if len(config.CustomFields) == 0 || config.CustomFields["api_key"] == "" {
		return nil, fmt.Errorf("API key required")
	}

	return &TokenInfo{
		AccessToken: config.CustomFields["api_key"],
		TokenType:   "ApiKey",
		ExpiresAt:   time.Now().Add(24 * time.Hour), // API keys typically don't expire
	}, nil
}

func (ap *APIKeyProvider) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	return tokenInfo, nil // API keys don't need refresh
}

func (ap *APIKeyProvider) ValidateToken(ctx context.Context, token string) error {
	return nil // API key validation would be done by the external service
}

func (ap *APIKeyProvider) AddAuthToRequest(req *http.Request, tokenInfo *TokenInfo) error {
	// API key can be added in different ways depending on the service
	// Common patterns: header, query parameter, custom header
	req.Header.Set("X-API-Key", tokenInfo.AccessToken)
	return nil
}

func (ap *APIKeyProvider) GetTokenExpiry(tokenInfo *TokenInfo) time.Time {
	return tokenInfo.ExpiresAt
}

func (ap *APIKeyProvider) SupportsRefresh() bool {
	return false
}

// BearerTokenProvider implements bearer token authentication
type BearerTokenProvider struct {
	logger logging.Logger
}

func (btp *BearerTokenProvider) Name() string { return "bearer" }

func (btp *BearerTokenProvider) Authenticate(ctx context.Context, config *plugin.AuthConfig) (*TokenInfo, error) {
	if len(config.CustomFields) == 0 || config.CustomFields["token"] == "" {
		return nil, fmt.Errorf("bearer token required")
	}

	return &TokenInfo{
		AccessToken: config.CustomFields["token"],
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(time.Hour), // Default 1 hour expiry
	}, nil
}

func (btp *BearerTokenProvider) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	return tokenInfo, nil // Bearer tokens typically don't refresh
}

func (btp *BearerTokenProvider) ValidateToken(ctx context.Context, token string) error {
	return nil
}

func (btp *BearerTokenProvider) AddAuthToRequest(req *http.Request, tokenInfo *TokenInfo) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenInfo.AccessToken))
	return nil
}

func (btp *BearerTokenProvider) GetTokenExpiry(tokenInfo *TokenInfo) time.Time {
	return tokenInfo.ExpiresAt
}

func (btp *BearerTokenProvider) SupportsRefresh() bool {
	return false
}

// RegisterAuthProvider registers a custom authentication provider
func (aa *AuthAbstraction) RegisterAuthProvider(authType string, provider AuthProviderInterface) error {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	aa.providers[authType] = provider

	aa.logger.Info("registered auth provider", "type", authType, "name", provider.Name())

	return nil
}

// GetSupportedAuthTypes returns all supported authentication types
func (aa *AuthAbstraction) GetSupportedAuthTypes() []string {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	types := make([]string, 0, len(aa.providers))
	for authType := range aa.providers {
		types = append(types, authType)
	}

	return types
}

// ValidateAuthConfig validates authentication configuration
func (aa *AuthAbstraction) ValidateAuthConfig(config *plugin.AuthConfig) error {
	if config.Type == "" {
		return fmt.Errorf("authentication type is required")
	}

	provider, err := aa.GetAuthProvider(config.Type)
	if err != nil {
		return err
	}

	// Provider-specific validation
	switch config.Type {
	case plugin.AuthTypeOAuth2:
		if config.CredentialsRef == "" {
			return fmt.Errorf("credentials reference required for OAuth2")
		}

	case plugin.AuthTypeBasic:
		if len(config.CustomFields) == 0 || config.CustomFields["username"] == "" || config.CustomFields["password"] == "" {
			return fmt.Errorf("username and password required for basic auth")
		}

	case plugin.AuthTypeAPIKey:
		if len(config.CustomFields) == 0 || config.CustomFields["api_key"] == "" {
			return fmt.Errorf("API key required")
		}

	case plugin.AuthTypeBearer:
		if len(config.CustomFields) == 0 || config.CustomFields["token"] == "" {
			return fmt.Errorf("bearer token required")
		}
	}

	aa.logger.Debug("auth config validated", "type", config.Type, "provider", provider.Name())

	return nil
}

// IsTokenExpired checks if a token is expired or will expire soon
func (aa *AuthAbstraction) IsTokenExpired(pluginName string, gracePeriod time.Duration) bool {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	tokenInfo, exists := aa.tokens[pluginName]
	if !exists {
		return true // No token means expired
	}

	return time.Now().Add(gracePeriod).After(tokenInfo.ExpiresAt)
}

// ClearPluginToken clears the stored token for a plugin
func (aa *AuthAbstraction) ClearPluginToken(pluginName string) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	delete(aa.tokens, pluginName)

	aa.logger.Debug("cleared token for plugin", "plugin", pluginName)
}

// GetTokenMetrics returns token metrics for monitoring
func (aa *AuthAbstraction) GetTokenMetrics() map[string]TokenMetrics {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	metrics := make(map[string]TokenMetrics)
	for plugin, token := range aa.tokens {
		metrics[plugin] = TokenMetrics{
			Plugin:       plugin,
			Type:         token.TokenType,
			ExpiresAt:    token.ExpiresAt,
			IsExpired:    time.Now().After(token.ExpiresAt),
			TimeToExpiry: time.Until(token.ExpiresAt),
		}
	}

	return metrics
}

// TokenMetrics contains token metrics
type TokenMetrics struct {
	Plugin       string        `json:"plugin"`
	Type         string        `json:"type"`
	ExpiresAt    time.Time     `json:"expires_at"`
	IsExpired    bool          `json:"is_expired"`
	TimeToExpiry time.Duration `json:"time_to_expiry"`
}
