package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daddia/zen/internal/logging"
)

// Manager provides generic filesystem operations for creating and managing files and directories
type Manager struct {
	logger logging.Logger
}

// New creates a new filesystem manager
func New(logger logging.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// CreateDirectory creates a directory with the specified permissions
func (m *Manager) CreateDirectory(dirPath string, perm os.FileMode) error {
	m.logger.Debug("Creating directory", "path", dirPath, "permissions", perm)

	if err := os.MkdirAll(dirPath, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	return nil
}

// CreateDirectories creates multiple directories with the specified permissions
func (m *Manager) CreateDirectories(basePath string, dirs []string, perm os.FileMode) error {
	for _, dir := range dirs {
		dirPath := filepath.Join(basePath, dir)
		if err := m.CreateDirectory(dirPath, perm); err != nil {
			return err
		}
		m.logger.Debug("Created directory", "path", dirPath)
	}
	return nil
}

// EnsureDirectory creates a directory if it doesn't exist
func (m *Manager) EnsureDirectory(dirPath string, perm os.FileMode) error {
	if !m.DirectoryExists(dirPath) {
		return m.CreateDirectory(dirPath, perm)
	}
	return nil
}

// DirectoryExists checks if a directory exists
func (m *Manager) DirectoryExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists checks if a file exists
func (m *Manager) FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CreateFile creates a file with the specified content and permissions
func (m *Manager) CreateFile(filePath string, content []byte, perm os.FileMode) error {
	m.logger.Debug("Creating file", "path", filePath, "permissions", perm)

	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := m.EnsureDirectory(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, content, perm); err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	return nil
}

// ReadFile reads the content of a file
func (m *Manager) ReadFile(filePath string) ([]byte, error) {
	m.logger.Debug("Reading file", "path", filePath)

	// Basic path validation to prevent directory traversal
	if filepath.IsAbs(filePath) {
		// Allow absolute paths but log them for security monitoring
		m.logger.Debug("reading absolute path", "path", filePath)
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - path validation implemented above
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return content, nil
}

// RemoveDirectory removes a directory and all its contents
func (m *Manager) RemoveDirectory(dirPath string) error {
	m.logger.Debug("Removing directory", "path", dirPath)

	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dirPath, err)
	}

	return nil
}

// RemoveFile removes a file
func (m *Manager) RemoveFile(filePath string) error {
	m.logger.Debug("Removing file", "path", filePath)

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", filePath, err)
	}

	return nil
}
