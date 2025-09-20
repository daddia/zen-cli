package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/daddia/zen/internal/logging"
	"github.com/pkg/errors"
)

// FileStorage implements CredentialStorage using encrypted file storage
type FileStorage struct {
	config    Config
	logger    logging.Logger
	mu        sync.RWMutex
	storePath string
	gcm       cipher.AEAD
}

// NewFileStorage creates a new file-based credential storage
func NewFileStorage(config Config, logger logging.Logger) (*FileStorage, error) {
	// Use the provided storage path (factory should set this to project .zen/auth)
	storePath := config.StoragePath
	if storePath == "" {
		return nil, NewStorageError("storage path not configured", "auth storage path must be provided")
	}

	if strings.HasPrefix(storePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, NewStorageError("failed to get user home directory", err.Error())
		}
		storePath = filepath.Join(home, storePath[2:])
	}

	// Ensure directory exists
	if err := os.MkdirAll(storePath, 0700); err != nil {
		return nil, NewStorageError("failed to create auth storage directory", err.Error())
	}

	// Set up encryption
	var gcm cipher.AEAD
	if config.EncryptionKey != "" {
		key := sha256.Sum256([]byte(config.EncryptionKey))
		block, err := aes.NewCipher(key[:])
		if err != nil {
			return nil, NewStorageError("failed to create cipher", err.Error())
		}

		gcm, err = cipher.NewGCM(block)
		if err != nil {
			return nil, NewStorageError("failed to create GCM cipher", err.Error())
		}
	}

	return &FileStorage{
		config:    config,
		logger:    logger,
		storePath: storePath,
		gcm:       gcm,
	}, nil
}

// Store saves credentials for the specified provider
func (f *FileStorage) Store(ctx context.Context, provider string, credential *Credential) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Serialize credential
	data, err := json.Marshal(credential)
	if err != nil {
		return NewStorageError("failed to serialize credential", err.Error())
	}

	// Encrypt if encryption is enabled
	if f.gcm != nil {
		data, err = f.encrypt(data)
		if err != nil {
			return NewStorageError("failed to encrypt credential", err.Error())
		}
	}

	// Write to file
	filename := f.getCredentialFilename(provider)
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return NewStorageError("failed to write credential file", err.Error())
	}

	f.logger.Debug("stored credential in file", "provider", provider, "file", filename)
	return nil
}

// Retrieve gets credentials for the specified provider
func (f *FileStorage) Retrieve(ctx context.Context, provider string) (*Credential, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	filename := f.getCredentialFilename(provider)

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, NewAuthError(
			ErrorCodeCredentialNotFound,
			"credential file not found",
			provider,
		)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, NewStorageError("failed to read credential file", err.Error())
	}

	// Decrypt if encryption is enabled
	if f.gcm != nil {
		data, err = f.decrypt(data)
		if err != nil {
			return nil, NewStorageError("failed to decrypt credential", err.Error())
		}
	}

	// Deserialize credential
	var credential Credential
	if err := json.Unmarshal(data, &credential); err != nil {
		return nil, NewStorageError("failed to deserialize credential", err.Error())
	}

	return &credential, nil
}

// Delete removes credentials for the specified provider
func (f *FileStorage) Delete(ctx context.Context, provider string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	filename := f.getCredentialFilename(provider)

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return NewStorageError("failed to delete credential file", err.Error())
	}

	f.logger.Debug("deleted credential file", "provider", provider, "file", filename)
	return nil
}

// List returns all stored provider names
func (f *FileStorage) List(ctx context.Context) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	entries, err := os.ReadDir(f.storePath)
	if err != nil {
		return nil, NewStorageError("failed to list credential files", err.Error())
	}

	var providers []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".cred") {
			provider := strings.TrimSuffix(name, ".cred")
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

// Clear removes all stored credentials
func (f *FileStorage) Clear(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := os.ReadDir(f.storePath)
	if err != nil {
		return NewStorageError("failed to list credential files", err.Error())
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".cred") {
			filename := filepath.Join(f.storePath, name)
			if err := os.Remove(filename); err != nil {
				f.logger.Warn("failed to delete credential file", "file", filename, "error", err)
			}
		}
	}

	f.logger.Debug("cleared all credential files")
	return nil
}

// Close closes the storage and releases resources
func (f *FileStorage) Close() error {
	// Nothing to close for file storage
	return nil
}

// Private helper methods

func (f *FileStorage) getCredentialFilename(provider string) string {
	// Sanitize provider name for filename
	sanitized := strings.ReplaceAll(provider, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	return filepath.Join(f.storePath, sanitized+".cred")
}

func (f *FileStorage) encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, f.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to generate nonce")
	}

	ciphertext := f.gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (f *FileStorage) decrypt(data []byte) ([]byte, error) {
	nonceSize := f.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := f.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt")
	}

	return plaintext, nil
}
