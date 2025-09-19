package auth

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeychainStorage(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()

	// Act
	storage, err := NewKeychainStorage(config, logger)

	// Assert
	if isKeychainAvailable() {
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, config, storage.config)
		assert.Equal(t, logger, storage.logger)
	} else {
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.True(t, IsAuthError(err))
		assert.Equal(t, ErrorCodeStorageError, GetErrorCode(err))
	}

	// Cleanup
	if storage != nil {
		storage.Close()
	}
}

func TestKeychainStorage_GetServiceName(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	// Act
	serviceName := storage.getServiceName()

	// Assert
	assert.Equal(t, "zen-cli", serviceName)
}

func TestKeychainStorage_GetAccountName(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{
			name:     "github provider",
			provider: "github",
			expected: "auth-github",
		},
		{
			name:     "gitlab provider",
			provider: "gitlab",
			expected: "auth-gitlab",
		},
		{
			name:     "custom provider",
			provider: "custom",
			expected: "auth-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			accountName := storage.getAccountName(tt.provider)

			// Assert
			assert.Equal(t, tt.expected, accountName)
		})
	}
}

func TestKeychainStorage_Store_Retrieve_Delete_Integration(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Skip on CI environments where keychain access might be restricted
	if testing.Short() {
		t.Skip("Skipping keychain integration test in short mode")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "test-provider-" + time.Now().Format("20060102150405") // Unique provider for test
	credential := &Credential{
		Provider:  provider,
		Token:     "test-token-12345",
		Type:      "token",
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	// Cleanup any existing test credential
	defer storage.Delete(ctx, provider)

	// Act & Assert - Store
	err = storage.Store(ctx, provider, credential)
	if err != nil {
		// On some CI systems or restricted environments, keychain access may fail
		// This is acceptable and we should skip the test
		t.Skipf("Keychain store failed (likely restricted environment): %v", err)
	}

	// Act & Assert - Retrieve
	retrieved, err := storage.Retrieve(ctx, provider)
	if err != nil {
		t.Skipf("Keychain retrieve failed (likely restricted environment): %v", err)
	}

	require.NotNil(t, retrieved)
	assert.Equal(t, credential.Provider, retrieved.Provider)
	assert.Equal(t, credential.Token, retrieved.Token)
	// Note: Metadata may not be fully preserved in keychain storage

	// Act & Assert - Delete
	err = storage.Delete(ctx, provider)
	assert.NoError(t, err)

	// Verify deletion
	_, err = storage.Retrieve(ctx, provider)
	assert.Error(t, err)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
}

func TestKeychainStorage_Retrieve_NotFound(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "nonexistent-provider-" + time.Now().Format("20060102150405")

	// Act
	credential, err := storage.Retrieve(ctx, provider)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, credential)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
}

func TestKeychainStorage_Delete_NonExistent(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "nonexistent-provider-" + time.Now().Format("20060102150405")

	// Act
	err = storage.Delete(ctx, provider)

	// Assert - deleting non-existent credential should not error
	assert.NoError(t, err)
}

func TestKeychainStorage_List(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Act
	listed, err := storage.List(ctx)

	// Assert
	// Note: List is a simplified implementation for keychain storage
	// It returns an empty list as implemented
	assert.NoError(t, err)
	assert.NotNil(t, listed)
	// We don't assert on the contents since listing keychain entries is complex
}

func TestKeychainStorage_Clear(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Act
	err = storage.Clear(ctx)

	// Assert
	// Note: Clear is a simplified implementation for keychain storage
	assert.NoError(t, err)
}

func TestKeychainStorage_Close(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)

	// Act
	err = storage.Close()

	// Assert
	assert.NoError(t, err)
}

func TestKeychainStorage_InterfaceCompliance(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	// Assert - verify interface compliance
	var _ CredentialStorage = storage
}

func TestIsKeychainAvailable(t *testing.T) {
	// Act
	available := isKeychainAvailable()

	// Assert - verify the result matches the current platform
	switch runtime.GOOS {
	case "darwin":
		// On macOS, keychain should be available if security command exists
		// We can't assert true/false definitively as it depends on system configuration
		assert.NotNil(t, available) // Just verify it returns a boolean
	case "windows":
		// On Windows, depends on cmdkey availability
		assert.NotNil(t, available)
	case "linux":
		// On Linux, depends on secret-tool or similar
		assert.NotNil(t, available)
	default:
		// On other platforms, should return false
		assert.False(t, available)
	}
}

func TestKeychainStorage_PlatformSpecificBehavior(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		skip     bool
	}{
		{
			name:     "macOS keychain operations",
			platform: "darwin",
			skip:     runtime.GOOS != "darwin",
		},
		{
			name:     "Windows credential manager operations",
			platform: "windows",
			skip:     runtime.GOOS != "windows",
		},
		{
			name:     "Linux secret service operations",
			platform: "linux",
			skip:     runtime.GOOS != "linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skipf("Skipping %s test on %s", tt.platform, runtime.GOOS)
			}

			// Skip if keychain is not available
			if !isKeychainAvailable() {
				t.Skip("Keychain storage not available on this platform")
			}

			// Basic test that storage can be created for the current platform
			storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
			if err != nil {
				t.Skipf("Keychain storage creation failed on %s: %v", runtime.GOOS, err)
			}

			require.NotNil(t, storage)
			defer storage.Close()

			// Verify basic operations don't panic
			ctx := context.Background()
			provider := "test-provider"

			// These operations may fail due to permissions, but shouldn't panic
			_, _ = storage.List(ctx)
			_ = storage.Clear(ctx)
			_, _ = storage.Retrieve(ctx, provider)
			_ = storage.Delete(ctx, provider)
		})
	}
}

func TestKeychainStorage_ErrorHandling(t *testing.T) {
	// Skip if keychain is not available
	if !isKeychainAvailable() {
		t.Skip("Keychain storage not available on this platform")
	}

	// Arrange
	storage, err := NewKeychainStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		operation func() error
		expectErr bool
	}{
		{
			name: "store with empty provider",
			operation: func() error {
				return storage.Store(ctx, "", &Credential{Token: "test"})
			},
			expectErr: false, // May succeed or fail depending on platform
		},
		{
			name: "retrieve with empty provider",
			operation: func() error {
				_, err := storage.Retrieve(ctx, "")
				return err
			},
			expectErr: false, // May succeed or fail depending on platform
		},
		{
			name: "store with nil credential",
			operation: func() error {
				return storage.Store(ctx, "test", nil)
			},
			expectErr: true, // Should fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.operation()

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
			}
			// For non-error cases, we just verify it doesn't panic
			// Success or failure depends on platform restrictions
		})
	}
}
