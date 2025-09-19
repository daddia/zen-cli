package auth

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStorage(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	logger := logging.NewBasic()

	// Act
	storage, err := NewMemoryStorage(config, logger)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, storage)
	assert.NotNil(t, storage.credentials)
	assert.Equal(t, config, storage.config)
	assert.Equal(t, logger, storage.logger)

	// Cleanup
	storage.Close()
}

func TestMemoryStorage_Store(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
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

	// Verify credential was stored
	storage.mu.RLock()
	stored, exists := storage.credentials[provider]
	storage.mu.RUnlock()

	assert.True(t, exists)
	assert.NotNil(t, stored)
	assert.Equal(t, credential.Provider, stored.Provider)
	assert.Equal(t, credential.Token, stored.Token)
	assert.Equal(t, credential.Type, stored.Type)
}

func TestMemoryStorage_Store_MakesDeepCopy(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	original := &Credential{
		Provider: provider,
		Token:    "original-token",
		Type:     "token",
	}

	// Act
	err = storage.Store(ctx, provider, original)
	require.NoError(t, err)

	// Modify original after storing
	original.Token = "modified-token"

	// Assert - stored credential should not be affected
	retrieved, err := storage.Retrieve(ctx, provider)
	require.NoError(t, err)
	assert.Equal(t, "original-token", retrieved.Token)
}

func TestMemoryStorage_Retrieve(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	expected := &Credential{
		Provider:  provider,
		Token:     "test-token",
		Type:      "token",
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
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

func TestMemoryStorage_Retrieve_NotFound(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
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

func TestMemoryStorage_Retrieve_ReturnsDeepCopy(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "github"
	original := &Credential{
		Provider: provider,
		Token:    "original-token",
		Type:     "token",
	}

	err = storage.Store(ctx, provider, original)
	require.NoError(t, err)

	// Act
	retrieved, err := storage.Retrieve(ctx, provider)
	require.NoError(t, err)

	// Modify retrieved credential
	retrieved.Token = "modified-token"

	// Assert - original stored credential should not be affected
	retrieved2, err := storage.Retrieve(ctx, provider)
	require.NoError(t, err)
	assert.Equal(t, "original-token", retrieved2.Token)
}

func TestMemoryStorage_Delete(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
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

	// Act
	err = storage.Delete(ctx, provider)

	// Assert
	require.NoError(t, err)

	// Verify credential was deleted
	_, err = storage.Retrieve(ctx, provider)
	assert.Error(t, err)
	assert.True(t, IsAuthError(err))
	assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
}

func TestMemoryStorage_Delete_NonExistent(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	provider := "nonexistent"

	// Act
	err = storage.Delete(ctx, provider)

	// Assert - deleting non-existent credential should not error
	assert.NoError(t, err)
}

func TestMemoryStorage_List(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
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

func TestMemoryStorage_List_Empty(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Act
	listed, err := storage.List(ctx)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, listed)
}

func TestMemoryStorage_Clear(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
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

	// Verify individual retrievals fail
	for _, provider := range providers {
		_, err = storage.Retrieve(ctx, provider)
		assert.Error(t, err)
		assert.True(t, IsAuthError(err))
		assert.Equal(t, ErrorCodeCredentialNotFound, GetErrorCode(err))
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)

	// Act
	err = storage.Close()

	// Assert
	assert.NoError(t, err)
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
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

func TestMemoryStorage_InterfaceCompliance(t *testing.T) {
	// Arrange
	storage, err := NewMemoryStorage(DefaultConfig(), logging.NewBasic())
	require.NoError(t, err)
	defer storage.Close()

	// Assert - verify interface compliance
	var _ CredentialStorage = storage
}
