package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/daddia/zen/internal/logging"
)

// KeychainStorage implements CredentialStorage using OS native credential storage
type KeychainStorage struct {
	config Config
	logger logging.Logger
}

// NewKeychainStorage creates a new keychain-based credential storage
func NewKeychainStorage(config Config, logger logging.Logger) (*KeychainStorage, error) {
	if !isKeychainAvailable() {
		return nil, NewStorageError("keychain storage not available on this platform", nil)
	}

	return &KeychainStorage{
		config: config,
		logger: logger,
	}, nil
}

// Store saves credentials for the specified provider
func (k *KeychainStorage) Store(ctx context.Context, provider string, credential *Credential) error {
	if credential == nil {
		return NewStorageError("credential cannot be nil", nil)
	}

	// Serialize credential (excluding the token which is stored separately)
	metadata := *credential
	metadata.Token = "" // Don't store token in metadata

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return NewStorageError("failed to serialize credential metadata", err.Error())
	}

	service := k.getServiceName()
	account := k.getAccountName(provider)

	switch runtime.GOOS {
	case "darwin":
		return k.storeMacOS(service, account, credential.Token, string(metadataJSON))
	case "windows":
		return k.storeWindows(service, account, credential.Token, string(metadataJSON))
	case "linux":
		return k.storeLinux(service, account, credential.Token, string(metadataJSON))
	default:
		return NewStorageError("keychain storage not supported on this platform", runtime.GOOS)
	}
}

// Retrieve gets credentials for the specified provider
func (k *KeychainStorage) Retrieve(ctx context.Context, provider string) (*Credential, error) {
	service := k.getServiceName()
	account := k.getAccountName(provider)

	var token, metadata string
	var err error

	switch runtime.GOOS {
	case "darwin":
		token, metadata, err = k.retrieveMacOS(service, account)
	case "windows":
		token, metadata, err = k.retrieveWindows(service, account)
	case "linux":
		token, metadata, err = k.retrieveLinux(service, account)
	default:
		return nil, NewStorageError("keychain storage not supported on this platform", runtime.GOOS)
	}

	if err != nil {
		return nil, err
	}

	// Deserialize metadata
	var credential Credential
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &credential); err != nil {
			k.logger.Warn("failed to deserialize credential metadata, using token only", "provider", provider, "error", err)
			// Fallback: create minimal credential with just the token
			credential = Credential{
				Provider: provider,
				Type:     "token",
			}
		}
	}

	// Set the token
	credential.Token = token
	credential.Provider = provider

	return &credential, nil
}

// Delete removes credentials for the specified provider
func (k *KeychainStorage) Delete(ctx context.Context, provider string) error {
	service := k.getServiceName()
	account := k.getAccountName(provider)

	switch runtime.GOOS {
	case "darwin":
		return k.deleteMacOS(service, account)
	case "windows":
		return k.deleteWindows(service, account)
	case "linux":
		return k.deleteLinux(service, account)
	default:
		return NewStorageError("keychain storage not supported on this platform", runtime.GOOS)
	}
}

// List returns all stored provider names
func (k *KeychainStorage) List(ctx context.Context) ([]string, error) {
	// This is a simplified implementation
	// In practice, listing keychain entries is complex and platform-specific
	return []string{}, nil
}

// Clear removes all stored credentials
func (k *KeychainStorage) Clear(ctx context.Context) error {
	// This is a simplified implementation
	// Would need to list all entries and delete them individually
	return nil
}

// Close closes the storage and releases resources
func (k *KeychainStorage) Close() error {
	return nil
}

// Private helper methods

func (k *KeychainStorage) getServiceName() string {
	return "zen-cli"
}

func (k *KeychainStorage) getAccountName(provider string) string {
	return fmt.Sprintf("auth-%s", provider)
}

// macOS keychain operations
func (k *KeychainStorage) storeMacOS(service, account, password, metadata string) error {
	// Delete existing entry first (security add-generic-password fails if entry exists)
	k.deleteMacOS(service, account)

	args := []string{
		"add-generic-password",
		"-s", service,
		"-a", account,
		"-w", password,
	}

	if metadata != "" {
		args = append(args, "-j", metadata)
	}

	cmd := exec.Command("security", args...)
	if err := cmd.Run(); err != nil {
		return NewStorageError("failed to store credential in macOS keychain", err.Error())
	}

	k.logger.Debug("stored credential in macOS keychain", "service", service, "account", account)
	return nil
}

func (k *KeychainStorage) retrieveMacOS(service, account string) (string, string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", "", NewAuthError(
			ErrorCodeCredentialNotFound,
			"credential not found in macOS keychain",
			strings.TrimPrefix(account, "auth-"),
		)
	}

	password := strings.TrimSpace(string(output))

	// Try to get metadata (this is simplified - would need more complex parsing)
	metadataCmd := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-g")
	metadataOutput, _ := metadataCmd.CombinedOutput()

	// Extract metadata from output (simplified)
	metadata := ""
	lines := strings.Split(string(metadataOutput), "\n")
	for _, line := range lines {
		if strings.Contains(line, "comments:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				metadata = strings.TrimSpace(strings.Trim(parts[1], `"`))
			}
		}
	}

	return password, metadata, nil
}

func (k *KeychainStorage) deleteMacOS(service, account string) error {
	cmd := exec.Command("security", "delete-generic-password", "-s", service, "-a", account)
	_ = cmd.Run() // Ignore error - entry might not exist
	return nil
}

// Windows credential manager operations (simplified)
func (k *KeychainStorage) storeWindows(service, account, password, metadata string) error {
	target := fmt.Sprintf("%s/%s", service, account)

	// Use cmdkey to store credential
	cmd := exec.Command("cmdkey", "/generic:"+target, "/user:"+account, "/pass:"+password)
	if err := cmd.Run(); err != nil {
		return NewStorageError("failed to store credential in Windows credential manager", err.Error())
	}

	k.logger.Debug("stored credential in Windows credential manager", "target", target)
	return nil
}

func (k *KeychainStorage) retrieveWindows(service, account string) (string, string, error) {
	// This is a simplified implementation
	// Windows credential manager access from Go is complex
	return "", "", NewStorageError("Windows keychain retrieval not fully implemented", nil)
}

func (k *KeychainStorage) deleteWindows(service, account string) error {
	target := fmt.Sprintf("%s/%s", service, account)
	cmd := exec.Command("cmdkey", "/delete:"+target)
	_ = cmd.Run() // Ignore error - entry might not exist
	return nil
}

// Linux secret service operations (simplified)
func (k *KeychainStorage) storeLinux(service, account, password, metadata string) error {
	// This would require integration with libsecret or similar
	// For now, fall back to file storage
	return NewStorageError("Linux keychain storage not fully implemented", nil)
}

func (k *KeychainStorage) retrieveLinux(service, account string) (string, string, error) {
	return "", "", NewStorageError("Linux keychain retrieval not fully implemented", nil)
}

func (k *KeychainStorage) deleteLinux(service, account string) error {
	return NewStorageError("Linux keychain deletion not fully implemented", nil)
}

// isKeychainAvailable checks if keychain storage is available on the current platform
func isKeychainAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		// Check if security command is available
		_, err := exec.LookPath("security")
		return err == nil
	case "windows":
		// Check if cmdkey is available
		_, err := exec.LookPath("cmdkey")
		return err == nil
	case "linux":
		// Check if secret-tool or similar is available
		_, err := exec.LookPath("secret-tool")
		return err == nil
	default:
		return false
	}
}
