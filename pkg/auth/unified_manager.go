package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
)

// UnifiedAuthManager combines token and basic auth capabilities
type UnifiedAuthManager struct {
	tokenManager *TokenManager
	basicManager *BasicAuthManager
	config       *config.Config
	logger       logging.Logger
}

// NewUnifiedAuthManager creates a new unified auth manager
func NewUnifiedAuthManager(authConfig Config, cfg *config.Config, logger logging.Logger, storage CredentialStorage) *UnifiedAuthManager {
	return &UnifiedAuthManager{
		tokenManager: NewTokenManager(authConfig, logger, storage),
		basicManager: NewBasicAuthManager(authConfig, logger, storage),
		config:       cfg,
		logger:       logger,
	}
}

// Authenticate authenticates with the specified provider
func (u *UnifiedAuthManager) Authenticate(ctx context.Context, provider string) error {
	// Check if provider is configured
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if !exists {
		// Fall back to token manager for non-integration providers
		return u.tokenManager.Authenticate(ctx, provider)
	}

	// If Basic Auth credentials are in config, we're already authenticated
	if providerConfig.Email != "" && providerConfig.APIKey != "" {
		u.logger.Debug("using Basic Auth credentials from config", "provider", provider)
		return nil
	}

	// Otherwise use token authentication
	return u.tokenManager.Authenticate(ctx, provider)
}

// GetCredentials returns credentials for the specified provider
func (u *UnifiedAuthManager) GetCredentials(provider string) (string, error) {
	// Check if provider needs Basic Auth
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if exists && providerConfig.Type == "basic" {
		// Get Basic Auth credentials
		cred, err := u.basicManager.GetBasicAuthFromConfig(u.config, provider)
		if err != nil {
			return "", err
		}
		// Return as Basic Auth header value
		auth := cred.Email + ":" + cred.APIKey
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)), nil
	}

	// Fall back to token manager
	return u.tokenManager.GetCredentials(provider)
}

// GetAuthorizationHeader returns the Authorization header value for a provider
func (u *UnifiedAuthManager) GetAuthorizationHeader(provider string) (string, error) {
	// Check provider configuration
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if !exists {
		// Try token-based auth
		token, err := u.tokenManager.GetCredentials(provider)
		if err != nil {
			return "", err
		}
		return "Bearer " + token, nil
	}

	// Handle different auth types
	switch providerConfig.Type {
	case "basic":
		if providerConfig.Email != "" && providerConfig.APIKey != "" {
			auth := providerConfig.Email + ":" + providerConfig.APIKey
			return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)), nil
		}
		// Try to get from storage
		cred, err := u.basicManager.GetBasicAuthCredentials(provider)
		if err != nil {
			return "", err
		}
		auth := cred.Email + ":" + cred.APIKey
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)), nil

	case "token", "bearer":
		if providerConfig.APIKey != "" {
			return "Bearer " + providerConfig.APIKey, nil
		}
		// Try to get from token manager
		token, err := u.tokenManager.GetCredentials(provider)
		if err != nil {
			return "", err
		}
		return "Bearer " + token, nil

	case "oauth2":
		// OAuth2 would be handled here
		token, err := u.tokenManager.GetCredentials(provider)
		if err != nil {
			return "", err
		}
		return "Bearer " + token, nil

	default:
		// Default to token-based auth
		token, err := u.tokenManager.GetCredentials(provider)
		if err != nil {
			return "", err
		}
		if strings.Contains(token, "Basic ") || strings.Contains(token, "Bearer ") {
			return token, nil
		}
		return "Bearer " + token, nil
	}
}

// ValidateCredentials validates stored credentials for the provider
func (u *UnifiedAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	// Check if provider needs Basic Auth
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if exists && providerConfig.Type == "basic" {
		// For Basic Auth, check if we have the required fields
		if providerConfig.Email == "" || providerConfig.APIKey == "" {
			// Try to get from storage
			_, err := u.basicManager.GetBasicAuthCredentials(provider)
			return err
		}
		return nil
	}

	// Fall back to token manager
	return u.tokenManager.ValidateCredentials(ctx, provider)
}

// RefreshCredentials refreshes expired credentials if possible
func (u *UnifiedAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	// Basic Auth doesn't support refresh
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if exists && providerConfig.Type == "basic" {
		return fmt.Errorf("credential refresh not supported for Basic Auth")
	}

	// Fall back to token manager
	return u.tokenManager.RefreshCredentials(ctx, provider)
}

// IsAuthenticated checks if credentials are available and valid for the provider
func (u *UnifiedAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	// Check if provider needs Basic Auth
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if exists && providerConfig.Type == "basic" {
		// Check if credentials are in config
		if providerConfig.Email != "" && providerConfig.APIKey != "" {
			return true
		}
		// Try to get from storage
		_, err := u.basicManager.GetBasicAuthCredentials(provider)
		return err == nil
	}

	// Fall back to token manager
	return u.tokenManager.IsAuthenticated(ctx, provider)
}

// ListProviders returns all configured providers
func (u *UnifiedAuthManager) ListProviders() []string {
	providers := u.tokenManager.ListProviders()

	// Add integration providers
	for provider := range u.config.Integrations.Providers {
		found := false
		for _, p := range providers {
			if p == provider {
				found = true
				break
			}
		}
		if !found {
			providers = append(providers, provider)
		}
	}

	return providers
}

// DeleteCredentials removes stored credentials for the provider
func (u *UnifiedAuthManager) DeleteCredentials(provider string) error {
	// Try both managers
	err1 := u.tokenManager.DeleteCredentials(provider)
	err2 := u.basicManager.storage.Delete(context.Background(), provider)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to delete credentials: %v, %v", err1, err2)
	}

	return nil
}

// GetProviderInfo returns information about the provider
func (u *UnifiedAuthManager) GetProviderInfo(provider string) (*ProviderInfo, error) {
	// Check if it's an integration provider
	providerConfig, exists := u.config.Integrations.Providers[provider]
	if exists {
		instructions := []string{}
		envVars := []string{}
		configKeys := []string{}

		if providerConfig.Type == "basic" {
			instructions = append(instructions, "Configure email and api_key in config.yaml")
			configKeys = append(configKeys, "email", "api_key")
		} else {
			instructions = append(instructions, "Configure API token in config.yaml or environment")
			envVars = append(envVars, fmt.Sprintf("ZEN_%s_TOKEN", strings.ToUpper(provider)))
			configKeys = append(configKeys, "api_key", "token")
		}

		return &ProviderInfo{
			Name:         provider,
			Type:         providerConfig.Type,
			Description:  fmt.Sprintf("%s integration", provider),
			Instructions: instructions,
			EnvVars:      envVars,
			ConfigKeys:   configKeys,
			Metadata: map[string]string{
				"base_url":    providerConfig.URL,
				"project_key": providerConfig.ProjectKey,
			},
		}, nil
	}

	// Fall back to token manager
	return u.tokenManager.GetProviderInfo(provider)
}
