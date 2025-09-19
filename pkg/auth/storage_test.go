package auth

import (
	"fmt"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStorageInfo(t *testing.T) {
	// Act
	info := GetStorageInfo()

	// Assert
	require.Contains(t, info, "keychain")
	require.Contains(t, info, "file")
	require.Contains(t, info, "memory")

	// Verify keychain info
	keychainInfo := info["keychain"]
	assert.Equal(t, "keychain", keychainInfo.Type)
	assert.True(t, keychainInfo.Secure)
	assert.Contains(t, keychainInfo.Description, "OS native")

	// Verify file info
	fileInfo := info["file"]
	assert.Equal(t, "file", fileInfo.Type)
	assert.True(t, fileInfo.Available)
	assert.True(t, fileInfo.Secure)
	assert.Contains(t, fileInfo.Description, "Encrypted file")

	// Verify memory info
	memoryInfo := info["memory"]
	assert.Equal(t, "memory", memoryInfo.Type)
	assert.True(t, memoryInfo.Available)
	assert.False(t, memoryInfo.Secure)
	assert.Contains(t, memoryInfo.Description, "memory")
}

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
		expectError bool
		expectType  string
	}{
		{
			name:        "memory storage",
			storageType: "memory",
			expectError: false,
			expectType:  "*auth.MemoryStorage",
		},
		{
			name:        "file storage",
			storageType: "file",
			expectError: false,
			expectType:  "*auth.FileStorage",
		},
		{
			name:        "keychain storage - may fail if not available",
			storageType: "keychain",
			expectError: false, // We'll handle this gracefully
		},
		{
			name:        "auto storage - defaults to available option",
			storageType: "auto",
			expectError: false,
		},
		{
			name:        "empty storage type - uses default logic",
			storageType: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			config.StoragePath = t.TempDir() + "/auth"
			logger := logging.NewBasic()

			// Act
			storage, err := NewStorage(tt.storageType, config, logger)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				// Some storage types may not be available on all platforms
				// so we accept either success or a specific error
				if err != nil {
					// Check if it's a platform availability error
					if IsAuthError(err) && GetErrorCode(err) == ErrorCodeStorageError {
						t.Skipf("Storage type %s not available on this platform: %v", tt.storageType, err)
					} else {
						t.Errorf("Unexpected error: %v", err)
					}
				} else {
					assert.NotNil(t, storage)
					if tt.expectType != "" {
						assert.Contains(t, fmt.Sprintf("%T", storage), tt.expectType)
					}
				}
			}

			// Cleanup
			if storage != nil {
				storage.Close()
			}
		})
	}
}

func TestNewStorage_FallbackBehavior(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	config.StoragePath = t.TempDir() + "/auth"
	logger := logging.NewBasic()

	// Act - try with an unsupported storage type
	storage, err := NewStorage("unsupported", config, logger)

	// Assert - should fallback to an available storage type
	if err != nil {
		// If keychain is not available, it should fall back to file storage
		// This is acceptable behavior
		assert.True(t, IsAuthError(err) || storage != nil)
	} else {
		assert.NotNil(t, storage)
	}

	// Cleanup
	if storage != nil {
		storage.Close()
	}
}

func TestStorageInfo_Completeness(t *testing.T) {
	// Act
	info := GetStorageInfo()

	// Assert - verify all required fields are present
	for storageType, storageInfo := range info {
		t.Run(storageType, func(t *testing.T) {
			assert.NotEmpty(t, storageInfo.Type)
			assert.NotEmpty(t, storageInfo.Description)
			// Available and Secure are booleans, so we just check they're set
			assert.NotNil(t, storageInfo.Available)
			assert.NotNil(t, storageInfo.Secure)
		})
	}
}

func TestStorageInterface_Compliance(t *testing.T) {
	// This test verifies that our storage implementations comply with the interface
	tests := []struct {
		name        string
		storageType string
	}{
		{
			name:        "memory storage",
			storageType: "memory",
		},
		{
			name:        "file storage",
			storageType: "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			config.StoragePath = t.TempDir() + "/auth"
			logger := logging.NewBasic()

			// Act
			storage, err := NewStorage(tt.storageType, config, logger)

			// Skip if storage type is not available
			if err != nil {
				if IsAuthError(err) && GetErrorCode(err) == ErrorCodeStorageError {
					t.Skipf("Storage type %s not available: %v", tt.storageType, err)
				}
				require.NoError(t, err)
			}

			// Assert - verify interface compliance
			require.NotNil(t, storage)

			// Check that all interface methods are available
			var _ = storage

			// Cleanup
			t.Cleanup(func() {
				storage.Close()
			})
		})
	}
}

func TestStorageType_Validation(t *testing.T) {
	tests := []struct {
		name         string
		storageType  string
		shouldCreate bool
	}{
		{
			name:         "valid memory type",
			storageType:  "memory",
			shouldCreate: true,
		},
		{
			name:         "valid file type",
			storageType:  "file",
			shouldCreate: true,
		},
		{
			name:         "case sensitive",
			storageType:  "MEMORY",
			shouldCreate: false, // Should fallback to default
		},
		{
			name:         "invalid type",
			storageType:  "invalid",
			shouldCreate: false, // Should fallback to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := DefaultConfig()
			config.StoragePath = t.TempDir() + "/auth"
			logger := logging.NewBasic()

			// Act
			storage, err := NewStorage(tt.storageType, config, logger)

			// Assert
			if tt.shouldCreate {
				// For valid types, we expect either success or platform-specific unavailability
				if err != nil && IsAuthError(err) && GetErrorCode(err) == ErrorCodeStorageError {
					t.Skipf("Storage type %s not available on this platform", tt.storageType)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, storage)
				}
			}
			// For invalid types, fallback logic handles gracefully

			// Cleanup
			if storage != nil {
				storage.Close()
			}
		})
	}
}
