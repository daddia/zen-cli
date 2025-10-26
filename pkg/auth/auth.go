package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/go-viper/mapstructure/v2"
)

// Manager represents the authentication interface for all providers
type Manager interface {
	// Authenticate authenticates with the specified provider
	Authenticate(ctx context.Context, provider string) error

	// GetCredentials returns credentials for the specified provider
	GetCredentials(provider string) (string, error)

	// ValidateCredentials validates stored credentials for the provider
	ValidateCredentials(ctx context.Context, provider string) error

	// RefreshCredentials refreshes expired credentials if possible
	RefreshCredentials(ctx context.Context, provider string) error

	// IsAuthenticated checks if credentials are available and valid for the provider
	IsAuthenticated(ctx context.Context, provider string) bool

	// ListProviders returns all configured providers
	ListProviders() []string

	// DeleteCredentials removes stored credentials for the provider
	DeleteCredentials(provider string) error

	// GetProviderInfo returns information about the provider
	GetProviderInfo(provider string) (*ProviderInfo, error)
}

// ProviderInfo contains information about an authentication provider
type ProviderInfo struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Description  string            `json:"description"`
	Instructions []string          `json:"instructions"`
	EnvVars      []string          `json:"env_vars"`
	ConfigKeys   []string          `json:"config_keys"`
	Scopes       []string          `json:"scopes,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Credential represents stored authentication credentials
type Credential struct {
	Provider      string            `json:"provider"`
	Token         string            `json:"token"`
	Type          string            `json:"type"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	Scopes        []string          `json:"scopes,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	LastUsed      time.Time         `json:"last_used"`
	LastValidated *time.Time        `json:"last_validated,omitempty"`
}

// IsExpired checks if the credential is expired
func (c *Credential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsValid checks if the credential is valid (not expired and has token)
func (c *Credential) IsValid() bool {
	return c.Token != "" && !c.IsExpired()
}

// Config represents authentication configuration
type Config struct {
	// Storage configuration
	StorageType   string `yaml:"storage_type" json:"storage_type"`     // keychain, file, memory
	StoragePath   string `yaml:"storage_path" json:"storage_path"`     // For file storage
	EncryptionKey string `yaml:"encryption_key" json:"encryption_key"` // For file encryption

	// Validation configuration
	ValidationTimeout time.Duration `yaml:"validation_timeout" json:"validation_timeout"`
	CacheTimeout      time.Duration `yaml:"cache_timeout" json:"cache_timeout"`

	// Provider configuration
	Providers map[string]ProviderConfig `yaml:"providers" json:"providers"`
}

// ProviderConfig represents provider-specific configuration
type ProviderConfig struct {
	Type       string            `yaml:"type" json:"type"`
	BaseURL    string            `yaml:"base_url" json:"base_url"`
	Scopes     []string          `yaml:"scopes" json:"scopes"`
	EnvVars    []string          `yaml:"env_vars" json:"env_vars"`
	ConfigKeys []string          `yaml:"config_keys" json:"config_keys"`
	Metadata   map[string]string `yaml:"metadata" json:"metadata"`
}

// DefaultConfig returns default authentication configuration
func DefaultConfig() Config {
	return Config{
		StorageType:       "keychain",
		ValidationTimeout: 10 * time.Second,
		CacheTimeout:      1 * time.Hour,
		Providers: map[string]ProviderConfig{
			"github": {
				Type:       "token",
				BaseURL:    "https://api.github.com",
				Scopes:     []string{"repo"},
				EnvVars:    []string{"ZEN_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"},
				ConfigKeys: []string{"github.token"},
			},
			"gitlab": {
				Type:       "token",
				BaseURL:    "https://gitlab.com/api/v4",
				Scopes:     []string{"read_repository"},
				EnvVars:    []string{"ZEN_GITLAB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN"},
				ConfigKeys: []string{"gitlab.token"},
			},
			"jira": {
				Type:       "basic",
				BaseURL:    "", // Will be configured per instance
				Scopes:     []string{"read", "write"},
				EnvVars:    []string{"ZEN_JIRA_TOKEN", "JIRA_TOKEN", "ZEN_JIRA_EMAIL", "JIRA_EMAIL"},
				ConfigKeys: []string{"jira.token", "jira.email", "jira.server_url"},
				Metadata: map[string]string{
					"auth_method": "basic", // Jira Cloud uses email + API token
					"token_type":  "api_token",
				},
			},
		},
	}
}

// Implement config.Configurable interface

// Validate validates the authentication configuration
func (c Config) Validate() error {
	validStorageTypes := []string{"keychain", "file", "memory"}
	validType := false
	for _, t := range validStorageTypes {
		if c.StorageType == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid storage_type: %s (must be one of: keychain, file, memory)", c.StorageType)
	}

	if c.StorageType == "file" && c.StoragePath == "" {
		return fmt.Errorf("storage_path is required when storage_type is 'file'")
	}

	if c.ValidationTimeout <= 0 {
		return fmt.Errorf("validation_timeout must be positive")
	}

	if c.CacheTimeout <= 0 {
		return fmt.Errorf("cache_timeout must be positive")
	}

	return nil
}

// Defaults returns a new Config with default values
func (c Config) Defaults() config.Configurable {
	return DefaultConfig()
}

// ConfigParser implements config.ConfigParser[Config] interface
type ConfigParser struct{}

// Parse converts raw configuration data to Config
func (p ConfigParser) Parse(raw map[string]interface{}) (Config, error) {
	var cfg Config

	// Use mapstructure to decode the raw map into our config struct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return cfg, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return cfg, fmt.Errorf("failed to decode auth config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for auth
func (p ConfigParser) Section() string {
	return "auth"
}

// NewManager creates a new authentication manager with the specified configuration
func NewManager(config Config, logger logging.Logger, storage CredentialStorage) Manager {
	return NewTokenManager(config, logger, storage)
}
