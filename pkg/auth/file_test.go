package auth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config with temp directory",
			config: Config{
				StoragePath: filepath.Join(t.TempDir(), "auth"),
			},
			expectError: false,
		},
		{
			name: "config with tilde path",
			config: Config{
				StoragePath: "~/.zen/auth-test",
			},
			expectError: false,
		},
		{
			name: "config with encryption key",
			config: Config{
				StoragePath:   filepath.Join(t.TempDir(), "auth"),
				EncryptionKey: "test-encryption-key",
			},
			expectError: false,
		},
		{
			name: "empty storage path defaults",
			config: Config{
				StoragePath: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := logging.NewBasic()

			// Act
			storage, err := NewFileStorage(tt.config, logger)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				assert.NotEmpty(t, storage.storePath)

				// Verify directory was created
				_, err := os.Stat(storage.storePath)
				assert.NoError(t, err)
			}

			// Cleanup
			if storage != nil {
				storage.Close()
				if tt.config.StoragePath != "" && !filepath.IsAbs(tt.config.StoragePath) {
					// Only clean up temp directories
					if strings.HasPrefix(storage.storePath, os.TempDir()) {
						os.RemoveAll(storage.storePath)
					}
				}
			}
		})
	}
}

func TestFileStorage_Store_Unencrypted(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	credential := &Credential{
		Provider:  provider,
		Token:     "test-token",
		Type:      "token",
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	// Act
	err = storage.Store(ctx, provider, credential)

	// Assert
	require.NoError(t, err)

	// Verify file was created
	filename := storage.getCredentialFilename(provider)
	_, err = os.Stat(filename)
	assert.NoError(t, err)

	// Verify file permissions are restrictive
	fileInfo, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())
}

func TestFileStorage_Store_Encrypted(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{
		StoragePath:   filepath.Join(tempDir, "auth"),
		EncryptionKey: "test-encryption-key-32-characters",
	}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	credential := &Credential{
		Provider: provider,
		Token:    "test-token",
		Type:     "token",
	}

	// Act
	err = storage.Store(ctx, provider, credential)

	// Assert
	require.NoError(t, err)

	// Verify file was created and is encrypted (not readable as plain JSON)
	filename := storage.getCredentialFilename(provider)
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Encrypted data should not contain the plain token
	assert.NotContains(t, string(data), "test-token")
}

func TestFileStorage_Retrieve_Unencrypted(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	expected := &Credential{
		Provider:  provider,
		Token:     "test-token",
		Type:      "token",
		CreatedAt: time.Now().Truncate(time.Second), // Truncate for comparison
		LastUsed:  time.Now().Truncate(time.Second),
	}

	err = storage.Store(ctx, provider, expected)
	require.NoError(t, err)

	// Act
	retrieved, err := storage.Retrieve(ctx, provider)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expected.Provider, retrieved.Provider)
	assert.Equal(t, expected.Token, retrieved.Token)
	assert.Equal(t, expected.Type, retrieved.Type)
}

func TestFileStorage_Retrieve_Encrypted(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{
		StoragePath:   filepath.Join(tempDir, "auth"),
		EncryptionKey: "test-encryption-key-32-characters",
	}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	expected := &Credential{
		Provider: provider,
		Token:    "test-token",
		Type:     "token",
	}

	err = storage.Store(ctx, provider, expected)
	require.NoError(t, err)

	// Act
	retrieved, err := storage.Retrieve(ctx, provider)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expected.Provider, retrieved.Provider)
	assert.Equal(t, expected.Token, retrieved.Token)
	assert.Equal(t, expected.Type, retrieved.Type)
}

func TestFileStorage_Retrieve_NotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "nonexistent"

	// Act
	credential, err := storage.Retrieve(ctx, provider)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, credential)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
}

func TestFileStorage_Delete(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	credential := &Credential{
		Provider: provider,
		Token:    "test-token",
		Type:     "token",
	}

	err = storage.Store(ctx, provider, credential)
	require.NoError(t, err)

	// Verify file exists
	filename := storage.getCredentialFilename(provider)
	_, err = os.Stat(filename)
	require.NoError(t, err)

	// Act
	err = storage.Delete(ctx, provider)

	// Assert
	require.NoError(t, err)

	// Verify file was deleted
	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err))

	// Verify retrieve fails
	_, err = storage.Retrieve(ctx, provider)
	assert.Error(t, err)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
}

func TestFileStorage_Delete_NonExistent(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "nonexistent"

	// Act
	err = storage.Delete(ctx, provider)

	// Assert - deleting non-existent credential should not error
	assert.NoError(t, err)
}

func TestFileStorage_List(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	providers := []string{"github", "gitlab", "custom"}

	// Store multiple credentials
	for _, provider := range providers {
		credential := &Credential{
			Provider: provider,
			Token:    "token-" + provider,
			Type:     "token",
		}
		err = storage.Store(ctx, provider, credential)
		require.NoError(t, err)
	}

	// Act
	listed, err := storage.List(ctx)

	// Assert
	require.NoError(t, err)
	assert.ElementsMatch(t, providers, listed)
}

func TestFileStorage_List_Empty(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Act
	listed, err := storage.List(ctx)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, listed)
}

func TestFileStorage_Clear(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	providers := []string{"github", "gitlab", "custom"}

	// Store multiple credentials
	for _, provider := range providers {
		credential := &Credential{
			Provider: provider,
			Token:    "token-" + provider,
			Type:     "token",
		}
		err = storage.Store(ctx, provider, credential)
		require.NoError(t, err)
	}

	// Act
	err = storage.Clear(ctx)

	// Assert
	require.NoError(t, err)

	// Verify all credentials were cleared
	listed, err := storage.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, listed)
}

func TestFileStorage_GetCredentialFilename(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{
			name:     "simple provider",
			provider: "github",
			expected: "github.cred",
		},
		{
			name:     "provider with slash",
			provider: "github/enterprise",
			expected: "github_enterprise.cred",
		},
		{
			name:     "provider with backslash",
			provider: "github\\enterprise",
			expected: "github_enterprise.cred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			filename := storage.getCredentialFilename(tt.provider)

			// Assert
			assert.Equal(t, filepath.Join(storage.storePath, tt.expected), filename)
		})
	}
}

func TestFileStorage_EncryptDecrypt(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{
		StoragePath:   filepath.Join(tempDir, "auth"),
		EncryptionKey: "test-encryption-key-32-characters",
	}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	originalData := []byte("test data to encrypt")

	// Act - encrypt
	encrypted, err := storage.encrypt(originalData)
	require.NoError(t, err)

	// Assert - encrypted data should be different
	assert.NotEqual(t, originalData, encrypted)
	assert.Greater(t, len(encrypted), len(originalData)) // Should be larger due to nonce

	// Act - decrypt
	decrypted, err := storage.decrypt(encrypted)
	require.NoError(t, err)

	// Assert - decrypted should match original
	assert.Equal(t, originalData, decrypted)
}

func TestFileStorage_EncryptDecrypt_InvalidData(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{
		StoragePath:   filepath.Join(tempDir, "auth"),
		EncryptionKey: "test-encryption-key-32-characters",
	}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "too short data",
			data: []byte("short"),
		},
		{
			name: "invalid encrypted data",
			data: []byte("this is not encrypted data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := storage.decrypt(tt.data)

			// Assert
			assert.Error(t, err)
		})
	}
}

func TestFileStorage_Close(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)

	// Act
	err = storage.Close()

	// Assert
	assert.NoError(t, err)
}

func TestFileStorage_ConcurrentAccess(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"

	// Act - concurrent operations
	done := make(chan bool, 3)

	// Concurrent store
	go func() {
		credential := &Credential{
			Provider: provider,
			Token:    "token-1",
			Type:     "token",
		}
		err := storage.Store(ctx, provider, credential)
		assert.NoError(t, err)
		done <- true
	}()

	// Concurrent retrieve (may fail initially)
	go func() {
		_, _ = storage.Retrieve(ctx, provider)
		// Error is acceptable here due to timing
		done <- true
	}()

	// Concurrent list
	go func() {
		_, err := storage.List(ctx)
		assert.NoError(t, err)
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Assert - final state should be consistent
	retrieved, err := storage.Retrieve(ctx, provider)
	require.NoError(t, err)
	assert.Equal(t, "token-1", retrieved.Token)
}

func TestFileStorage_InterfaceCompliance(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	config := Config{StoragePath: filepath.Join(tempDir, "auth")}
	storage, err := NewFileStorage(config, logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	// Assert - verify interface compliance
	var _ CredentialStorage = storage
}
