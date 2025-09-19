package auth

import (
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredential_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		credential *Credential
		expected   bool
	}{
		{
			name: "no expiration time",
			credential: &Credential{
				Token:     "test-token",
				ExpiresAt: nil,
			},
			expected: false,
		},
		{
			name: "not expired",
			credential: func() *Credential {
				future := time.Now().Add(1 * time.Hour)
				return &Credential{
					Token:     "test-token",
					ExpiresAt: &future,
				}
			}(),
			expected: false,
		},
		{
			name: "expired",
			credential: func() *Credential {
				past := time.Now().Add(-1 * time.Hour)
				return &Credential{
					Token:     "test-token",
					ExpiresAt: &past,
				}
			}(),
			expected: true,
		},
		{
			name: "future expiration",
			credential: func() *Credential {
				future := time.Now().Add(1 * time.Hour)
				return &Credential{
					Token:     "test-token",
					ExpiresAt: &future,
				}
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.credential.IsExpired()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCredential_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		credential *Credential
		expected   bool
	}{
		{
			name: "valid credential with no expiration",
			credential: &Credential{
				Token:     "test-token",
				ExpiresAt: nil,
			},
			expected: true,
		},
		{
			name: "empty token",
			credential: &Credential{
				Token:     "",
				ExpiresAt: nil,
			},
			expected: false,
		},
		{
			name: "expired credential",
			credential: func() *Credential {
				past := time.Now().Add(-1 * time.Hour)
				return &Credential{
					Token:     "test-token",
					ExpiresAt: &past,
				}
			}(),
			expected: false,
		},
		{
			name: "valid credential with future expiration",
			credential: func() *Credential {
				future := time.Now().Add(1 * time.Hour)
				return &Credential{
					Token:     "test-token",
					ExpiresAt: &future,
				}
			}(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.credential.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	// Act
	config := DefaultConfig()

	// Assert
	assert.Equal(t, "keychain", config.StorageType)
	assert.Equal(t, 10*time.Second, config.ValidationTimeout)
	assert.Equal(t, 1*time.Hour, config.CacheTimeout)

	// Verify GitHub provider config
	require.Contains(t, config.Providers, "github")
	githubConfig := config.Providers["github"]
	assert.Equal(t, "token", githubConfig.Type)
	assert.Equal(t, "https://api.github.com", githubConfig.BaseURL)
	assert.Contains(t, githubConfig.EnvVars, "GITHUB_TOKEN")
	assert.Contains(t, githubConfig.EnvVars, "GH_TOKEN")
	assert.Contains(t, githubConfig.EnvVars, "ZEN_GITHUB_TOKEN")
	assert.Contains(t, githubConfig.Scopes, "repo")

	// Verify GitLab provider config
	require.Contains(t, config.Providers, "gitlab")
	gitlabConfig := config.Providers["gitlab"]
	assert.Equal(t, "token", gitlabConfig.Type)
	assert.Equal(t, "https://gitlab.com/api/v4", gitlabConfig.BaseURL)
	assert.Contains(t, gitlabConfig.EnvVars, "GITLAB_TOKEN")
	assert.Contains(t, gitlabConfig.EnvVars, "GL_TOKEN")
	assert.Contains(t, gitlabConfig.EnvVars, "ZEN_GITLAB_TOKEN")
	assert.Contains(t, gitlabConfig.Scopes, "read_repository")
}

func TestNewManager(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)

	// Act
	manager := NewManager(config, logger, storage)

	// Assert
	assert.NotNil(t, manager)

	// Verify it's a TokenManager
	tokenManager, ok := manager.(*TokenManager)
	assert.True(t, ok)
	assert.NotNil(t, tokenManager)
}

func TestProviderInfo(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected *ProviderInfo
	}{
		{
			name:     "github provider info",
			provider: "github",
			expected: &ProviderInfo{
				Name:        "github",
				Type:        "token",
				Description: "GitHub authentication",
				EnvVars:     []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN"},
			},
		},
		{
			name:     "gitlab provider info",
			provider: "gitlab",
			expected: &ProviderInfo{
				Name:        "gitlab",
				Type:        "token",
				Description: "GitLab authentication",
				EnvVars:     []string{"GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			providerConfig := config.Providers[tt.provider]

			// Act & Assert
			assert.Equal(t, tt.expected.Type, providerConfig.Type)
			assert.ElementsMatch(t, tt.expected.EnvVars, providerConfig.EnvVars)
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		isValid bool
	}{
		{
			name:    "default config is valid",
			config:  DefaultConfig(),
			isValid: true,
		},
		{
			name: "custom valid config",
			config: Config{
				StorageType:       "file",
				ValidationTimeout: 5 * time.Second,
				CacheTimeout:      30 * time.Minute,
				Providers: map[string]ProviderConfig{
					"custom": {
						Type:    "token",
						BaseURL: "https://api.custom.com",
						EnvVars: []string{"CUSTOM_TOKEN"},
					},
				},
			},
			isValid: true,
		},
		{
			name: "empty config",
			config: Config{
				Providers: make(map[string]ProviderConfig),
			},
			isValid: true, // Empty config should be valid with defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act & Assert - basic structure validation
			if tt.isValid {
				assert.NotNil(t, tt.config.Providers)
				assert.GreaterOrEqual(t, tt.config.ValidationTimeout, time.Duration(0))
				assert.GreaterOrEqual(t, tt.config.CacheTimeout, time.Duration(0))
			}
		})
	}
}

func TestCredential_Metadata(t *testing.T) {
	// Arrange
	now := time.Now()
	expiry := now.Add(1 * time.Hour)

	credential := &Credential{
		Provider:      "github",
		Token:         "test-token",
		Type:          "token",
		ExpiresAt:     &expiry,
		Scopes:        []string{"repo", "read:user"},
		Metadata:      map[string]string{"user": "testuser"},
		CreatedAt:     now,
		LastUsed:      now,
		LastValidated: &now,
	}

	// Act & Assert
	assert.Equal(t, "github", credential.Provider)
	assert.Equal(t, "test-token", credential.Token)
	assert.Equal(t, "token", credential.Type)
	assert.Equal(t, &expiry, credential.ExpiresAt)
	assert.ElementsMatch(t, []string{"repo", "read:user"}, credential.Scopes)
	assert.Equal(t, "testuser", credential.Metadata["user"])
	assert.Equal(t, now, credential.CreatedAt)
	assert.Equal(t, now, credential.LastUsed)
	assert.Equal(t, &now, credential.LastValidated)
	assert.True(t, credential.IsValid())
	assert.False(t, credential.IsExpired())
}
