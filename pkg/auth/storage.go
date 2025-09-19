package auth

import (
	"context"

	"github.com/daddia/zen/internal/logging"
)

// CredentialStorage represents the interface for storing and retrieving credentials
type CredentialStorage interface {
	// Store saves credentials for the specified provider
	Store(ctx context.Context, provider string, credential *Credential) error

	// Retrieve gets credentials for the specified provider
	Retrieve(ctx context.Context, provider string) (*Credential, error)

	// Delete removes credentials for the specified provider
	Delete(ctx context.Context, provider string) error

	// List returns all stored provider names
	List(ctx context.Context) ([]string, error)

	// Clear removes all stored credentials
	Clear(ctx context.Context) error

	// Close closes the storage and releases resources
	Close() error
}

// StorageInfo represents information about the storage backend
type StorageInfo struct {
	Type        string `json:"type"`
	Available   bool   `json:"available"`
	Secure      bool   `json:"secure"`
	Description string `json:"description"`
}

// GetStorageInfo returns information about available storage backends
func GetStorageInfo() map[string]StorageInfo {
	return map[string]StorageInfo{
		"keychain": {
			Type:        "keychain",
			Available:   isKeychainAvailable(),
			Secure:      true,
			Description: "OS native credential storage (Keychain/Credential Manager/Secret Service)",
		},
		"file": {
			Type:        "file",
			Available:   true,
			Secure:      true,
			Description: "Encrypted file-based credential storage",
		},
		"memory": {
			Type:        "memory",
			Available:   true,
			Secure:      false,
			Description: "In-memory credential storage (for testing)",
		},
	}
}

// NewStorage creates a new credential storage backend
func NewStorage(storageType string, config Config, logger logging.Logger) (CredentialStorage, error) {
	switch storageType {
	case "keychain":
		return NewKeychainStorage(config, logger)
	case "file":
		return NewFileStorage(config, logger)
	case "memory":
		return NewMemoryStorage(config, logger)
	default:
		// Try keychain first, fall back to file
		if isKeychainAvailable() {
			return NewKeychainStorage(config, logger)
		}
		return NewFileStorage(config, logger)
	}
}
