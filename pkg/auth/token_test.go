package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)

	// Act
	manager := NewTokenManager(config, logger, storage)

	// Assert
	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, storage, manager.storage)
	assert.NotNil(t, manager.cache)
}

func TestTokenManager_GetCredentials_FromEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		envVar   string
		token    string
	}{
		{
			name:     "github token",
			provider: "github",
			envVar:   "GITHUB_TOKEN",
			token:    "ghp_test_token",
		},
		{
			name:     "github token alternative",
			provider: "github",
			envVar:   "GH_TOKEN",
			token:    "ghp_alt_token",
		},
		{
			name:     "zen github token",
			provider: "github",
			envVar:   "ZEN_GITHUB_TOKEN",
			token:    "ghp_zen_token",
		},
		{
			name:     "gitlab token",
			provider: "gitlab",
			envVar:   "GITLAB_TOKEN",
			token:    "glpat_test_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			logger := logging.NewBasic()
			storage, err := NewMemoryStorage(config, logger)
			require.NoError(t, err)
			manager := NewTokenManager(config, logger, storage)

			// Set environment variable
			os.Setenv(tt.envVar, tt.token)
			defer os.Unsetenv(tt.envVar)

			// Act
			token, err := manager.GetCredentials(tt.provider)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.token, token)
		})
	}
}

func TestTokenManager_GetCredentials_FromStorage(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	ctx := context.Background()
	provider := "github"
	expectedToken := "stored_token"
	credential := &Credential{
		Provider: provider,
		Token:    expectedToken,
		Type:     "token",
	}

	err = storage.Store(ctx, provider, credential)
	require.NoError(t, err)

	// Act
	token, err := manager.GetCredentials(provider)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestTokenManager_GetCredentials_NotFound(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	provider := "nonexistent"

	// Ensure no environment variables are set
	envVars := []string{"NONEXISTENT_TOKEN", "ZEN_NONEXISTENT_TOKEN"}
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Act
	token, err := manager.GetCredentials(provider)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
}

func TestTokenManager_GetCredentials_CacheHit(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	provider := "github"
	expectedToken := "cached_token"

	// Pre-populate cache
	manager.cache[provider] = &Credential{
		Provider: provider,
		Token:    expectedToken,
		Type:     "token",
	}

	// Act
	token, err := manager.GetCredentials(provider)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestTokenManager_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*TokenManager)
		provider string
		expected bool
	}{
		{
			name: "authenticated with valid cached credential",
			setup: func(tm *TokenManager) {
				now := time.Now()
				tm.cache["github"] = &Credential{
					Provider:      "github",
					Token:         "valid_token",
					Type:          "token",
					LastValidated: &now,
				}
			},
			provider: "github",
			expected: true,
		},
		{
			name: "not authenticated - no credential",
			setup: func(tm *TokenManager) {
				// No setup - no credential available
			},
			provider: "github",
			expected: false,
		},
		{
			name: "authenticated from environment",
			setup: func(tm *TokenManager) {
				os.Setenv("GITHUB_TOKEN", "env_token")
			},
			provider: "github",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			logger := logging.NewBasic()
			storage, err := NewMemoryStorage(config, logger)
			require.NoError(t, err)
			manager := NewTokenManager(config, logger, storage)

			tt.setup(manager)
			defer os.Unsetenv("GITHUB_TOKEN")

			ctx := context.Background()

			// Act
			result := manager.IsAuthenticated(ctx, tt.provider)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenManager_ListProviders(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	// Act
	providers := manager.ListProviders()

	// Assert
	assert.Contains(t, providers, "github")
	assert.Contains(t, providers, "gitlab")
	assert.Len(t, providers, len(config.Providers))
}

func TestTokenManager_DeleteCredentials(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	ctx := context.Background()
	provider := "github"
	credential := &Credential{
		Provider: provider,
		Token:    "test_token",
		Type:     "token",
	}

	// Store credential
	err = storage.Store(ctx, provider, credential)
	require.NoError(t, err)

	// Cache credential
	manager.cache[provider] = credential

	// Act
	err = manager.DeleteCredentials(provider)

	// Assert
	require.NoError(t, err)

	// Verify credential was removed from cache
	_, exists := manager.cache[provider]
	assert.False(t, exists)

	// Verify credential was removed from storage
	_, err = storage.Retrieve(ctx, provider)
	assert.Error(t, err)
}

func TestTokenManager_GetProviderInfo(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{
			name:     "github provider",
			provider: "github",
			expected: true,
		},
		{
			name:     "gitlab provider",
			provider: "gitlab",
			expected: true,
		},
		{
			name:     "unsupported provider",
			provider: "unsupported",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			logger := logging.NewBasic()
			storage, err := NewMemoryStorage(config, logger)
			require.NoError(t, err)
			manager := NewTokenManager(config, logger, storage)

			// Act
			info, err := manager.GetProviderInfo(tt.provider)

			// Assert
			if tt.expected {
				require.NoError(t, err)
				assert.Equal(t, tt.provider, info.Name)
				assert.Equal(t, "token", info.Type)
				assert.NotEmpty(t, info.Description)
				assert.NotEmpty(t, info.Instructions)
				assert.NotEmpty(t, info.EnvVars)
			} else {
				assert.Error(t, err)
				assert.Nil(t, info)
				assert.True(t, IsAuthError(err))
				assert.Equal(t, ErrorCodeProviderNotSupported, GetErrorCode(err))
			}
		})
	}
}

func TestTokenManager_RefreshCredentials(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	ctx := context.Background()
	provider := "github"

	// Act
	err = manager.RefreshCredentials(ctx, provider)

	// Assert
	assert.Error(t, err)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeAuthenticationFailed, GetErrorCode(err))
	assert.Contains(t, err.Error(), "refresh not supported")
}

func TestTokenManager_ValidateGitHubToken(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		responseBody    string
		expectedError   bool
		expectedErrCode ErrorCode
	}{
		{
			name:          "valid token",
			statusCode:    http.StatusOK,
			responseBody:  `{"login":"testuser"}`,
			expectedError: false,
		},
		{
			name:            "invalid token",
			statusCode:      http.StatusUnauthorized,
			expectedError:   true,
			expectedErrCode: ErrorCodeInvalidCredentials,
		},
		{
			name:            "insufficient scopes",
			statusCode:      http.StatusForbidden,
			expectedError:   true,
			expectedErrCode: ErrorCodeInsufficientScopes,
		},
		{
			name:            "rate limited",
			statusCode:      http.StatusTooManyRequests,
			expectedError:   true,
			expectedErrCode: ErrorCodeRateLimited,
		},
		{
			name:            "server error",
			statusCode:      http.StatusInternalServerError,
			expectedError:   true,
			expectedErrCode: ErrorCodeNetworkError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/user", r.URL.Path)
				assert.Contains(t, r.Header.Get("Authorization"), "token")
				assert.Equal(t, "zen-cli/1.0", r.Header.Get("User-Agent"))

				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			config := DefaultConfig()
			config.ValidationTimeout = 1 * time.Second
			logger := logging.NewBasic()
			storage, err := NewMemoryStorage(config, logger)
			require.NoError(t, err)
			manager := NewTokenManager(config, logger, storage)

			ctx := context.Background()
			token := "test_token"

			// Act
			err = manager.validateGitHubToken(ctx, token, server.URL)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.True(t, IsAuthError(err))
				assert.Equal(t, tt.expectedErrCode, GetErrorCode(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenManager_ValidateGitLabToken(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		responseBody    string
		expectedError   bool
		expectedErrCode ErrorCode
	}{
		{
			name:          "valid token",
			statusCode:    http.StatusOK,
			responseBody:  `{"username":"testuser"}`,
			expectedError: false,
		},
		{
			name:            "invalid token",
			statusCode:      http.StatusUnauthorized,
			expectedError:   true,
			expectedErrCode: ErrorCodeInvalidCredentials,
		},
		{
			name:            "insufficient scopes",
			statusCode:      http.StatusForbidden,
			expectedError:   true,
			expectedErrCode: ErrorCodeInsufficientScopes,
		},
		{
			name:            "rate limited",
			statusCode:      http.StatusTooManyRequests,
			expectedError:   true,
			expectedErrCode: ErrorCodeRateLimited,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/user", r.URL.Path)
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer")
				assert.Equal(t, "zen-cli/1.0", r.Header.Get("User-Agent"))

				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			config := DefaultConfig()
			config.ValidationTimeout = 1 * time.Second
			logger := logging.NewBasic()
			storage, err := NewMemoryStorage(config, logger)
			require.NoError(t, err)
			manager := NewTokenManager(config, logger, storage)

			ctx := context.Background()
			token := "test_token"

			// Act
			err = manager.validateGitLabToken(ctx, token, server.URL)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.True(t, IsAuthError(err))
				assert.Equal(t, tt.expectedErrCode, GetErrorCode(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenManager_Authenticate_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"testuser"}`))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Providers["github"] = ProviderConfig{
		Type:    "token",
		BaseURL: server.URL,
		EnvVars: []string{"GITHUB_TOKEN"},
	}
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	os.Setenv("GITHUB_TOKEN", "test_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	ctx := context.Background()
	provider := "github"

	// Act
	err = manager.Authenticate(ctx, provider)

	// Assert
	require.NoError(t, err)

	// Verify credential was cached
	cached, exists := manager.cache[provider]
	assert.True(t, exists)
	assert.Equal(t, "test_token", cached.Token)
	assert.NotNil(t, cached.LastValidated)
}

func TestTokenManager_Authenticate_ValidationFailure(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Providers["github"] = ProviderConfig{
		Type:    "token",
		BaseURL: server.URL,
		EnvVars: []string{"GITHUB_TOKEN"},
	}
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	os.Setenv("GITHUB_TOKEN", "invalid_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	ctx := context.Background()
	provider := "github"

	// Act
	err = manager.Authenticate(ctx, provider)

	// Assert
	assert.Error(t, err)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeInvalidCredentials, GetErrorCode(err))
}

func TestTokenManager_GetProviderInstructions(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()
	storage, err := NewMemoryStorage(config, logger)
	require.NoError(t, err)
	manager := NewTokenManager(config, logger, storage)

	tests := []struct {
		name     string
		provider string
		expected []string
	}{
		{
			name:     "github instructions",
			provider: "github",
			expected: []string{
				"1. Go to https://github.com/settings/tokens",
				"2. Click 'Generate new token (classic)'",
			},
		},
		{
			name:     "gitlab instructions",
			provider: "gitlab",
			expected: []string{
				"1. Go to your GitLab project settings",
				"2. Navigate to Access Tokens",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			instructions := manager.getProviderInstructions(tt.provider)

			// Assert
			for _, expected := range tt.expected {
				assert.Contains(t, instructions, expected)
			}
		})
	}
}
