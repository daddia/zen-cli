package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
)

// BasicAuthManager handles Basic Authentication with username/email and password/token
type BasicAuthManager struct {
	config  Config
	logger  logging.Logger
	storage CredentialStorage
	cache   map[string]*BasicAuthCredential
}

// BasicAuthCredential represents credentials for Basic Authentication
type BasicAuthCredential struct {
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	Username  string    `json:"username,omitempty"`
	APIKey    string    `json:"api_key"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// NewBasicAuthManager creates a new Basic Auth manager
func NewBasicAuthManager(config Config, logger logging.Logger, storage CredentialStorage) *BasicAuthManager {
	return &BasicAuthManager{
		config:  config,
		logger:  logger,
		storage: storage,
		cache:   make(map[string]*BasicAuthCredential),
	}
}

// GetBasicAuthCredentials retrieves Basic Auth credentials for a provider
func (b *BasicAuthManager) GetBasicAuthCredentials(provider string) (*BasicAuthCredential, error) {
	// Check cache first
	if cached, exists := b.cache[provider]; exists {
		return cached, nil
	}

	// Try to load from storage
	ctx := context.Background()
	credential, err := b.storage.Retrieve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	// Convert generic credential to BasicAuthCredential
	basicCred := &BasicAuthCredential{}
	if credential.Metadata != nil {
		if data, err := json.Marshal(credential.Metadata); err == nil {
			if err := json.Unmarshal(data, basicCred); err == nil {
				b.cache[provider] = basicCred
				return basicCred, nil
			}
		}
	}

	return nil, fmt.Errorf("no basic auth credentials found for provider '%s'", provider)
}

// StoreBasicAuthCredentials stores Basic Auth credentials for a provider
func (b *BasicAuthManager) StoreBasicAuthCredentials(provider string, cred *BasicAuthCredential) error {
	ctx := context.Background()

	// Convert to generic credential for storage
	metadata := make(map[string]interface{})
	if data, err := json.Marshal(cred); err == nil {
		if err := json.Unmarshal(data, &metadata); err != nil {
			return fmt.Errorf("failed to convert credentials: %w", err)
		}
	}

	// Convert metadata to string map
	strMetadata := make(map[string]string)
	for k, v := range metadata {
		strMetadata[k] = fmt.Sprintf("%v", v)
	}

	credential := &Credential{
		Provider:  provider,
		Token:     b.generateBasicAuthToken(cred.Email, cred.APIKey),
		Type:      "basic",
		CreatedAt: cred.CreatedAt,
		LastUsed:  cred.LastUsed,
		ExpiresAt: &cred.ExpiresAt,
		Metadata:  strMetadata,
	}

	// Store in storage backend
	if err := b.storage.Store(ctx, provider, credential); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	// Update cache
	b.cache[provider] = cred

	return nil
}

// GetBasicAuthFromConfig retrieves Basic Auth credentials from config
func (b *BasicAuthManager) GetBasicAuthFromConfig(cfg *config.Config, provider string) (*BasicAuthCredential, error) {
	// Check if provider is configured
	providerConfig, exists := cfg.Integrations.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not configured", provider)
	}

	// Check if credentials are directly in config
	if providerConfig.Email != "" && providerConfig.APIKey != "" {
		cred := &BasicAuthCredential{
			Provider:  provider,
			Email:     providerConfig.Email,
			APIKey:    providerConfig.APIKey,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}

		// Cache for future use
		b.cache[provider] = cred

		return cred, nil
	}

	// Try to get from storage using credentials reference
	if providerConfig.Credentials != "" {
		return b.GetBasicAuthCredentials(providerConfig.Credentials)
	}

	return nil, fmt.Errorf("no credentials found for provider '%s'", provider)
}

// generateBasicAuthToken generates a Basic Auth token from email and API key
func (b *BasicAuthManager) generateBasicAuthToken(email, apiKey string) string {
	auth := email + ":" + apiKey
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// GetAuthHeader returns the Authorization header value for Basic Auth
func (b *BasicAuthManager) GetAuthHeader(provider string, cfg *config.Config) (string, error) {
	// Try to get from config first
	cred, err := b.GetBasicAuthFromConfig(cfg, provider)
	if err != nil {
		// Fallback to stored credentials
		cred, err = b.GetBasicAuthCredentials(provider)
		if err != nil {
			return "", err
		}
	}

	// Generate Basic Auth header
	return "Basic " + b.generateBasicAuthToken(cred.Email, cred.APIKey), nil
}
