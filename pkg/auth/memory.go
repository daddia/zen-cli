package auth

import (
	"context"
	"sync"

	"github.com/daddia/zen/internal/logging"
)

// MemoryStorage implements CredentialStorage using in-memory storage
// This is primarily for testing and should not be used in production
type MemoryStorage struct {
	config      Config
	logger      logging.Logger
	mu          sync.RWMutex
	credentials map[string]*Credential
}

// NewMemoryStorage creates a new memory-based credential storage
func NewMemoryStorage(config Config, logger logging.Logger) (*MemoryStorage, error) {
	return &MemoryStorage{
		config:      config,
		logger:      logger,
		credentials: make(map[string]*Credential),
	}, nil
}

// Store saves credentials for the specified provider
func (m *MemoryStorage) Store(ctx context.Context, provider string, credential *Credential) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy to avoid external modifications
	stored := *credential
	m.credentials[provider] = &stored

	m.logger.Debug("stored credential in memory", "provider", provider)
	return nil
}

// Retrieve gets credentials for the specified provider
func (m *MemoryStorage) Retrieve(ctx context.Context, provider string) (*Credential, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	credential, exists := m.credentials[provider]
	if !exists {
		return nil, NewAuthError(
			ErrorCodeCredentialNotFound,
			"credential not found in memory storage",
			provider,
		)
	}

	// Return a copy to avoid external modifications
	result := *credential
	return &result, nil
}

// Delete removes credentials for the specified provider
func (m *MemoryStorage) Delete(ctx context.Context, provider string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.credentials, provider)
	m.logger.Debug("deleted credential from memory", "provider", provider)
	return nil
}

// List returns all stored provider names
func (m *MemoryStorage) List(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.credentials))
	for provider := range m.credentials {
		providers = append(providers, provider)
	}

	return providers, nil
}

// Clear removes all stored credentials
func (m *MemoryStorage) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.credentials = make(map[string]*Credential)
	m.logger.Debug("cleared all credentials from memory")
	return nil
}

// Close closes the storage and releases resources
func (m *MemoryStorage) Close() error {
	// Nothing to close for memory storage
	return nil
}
