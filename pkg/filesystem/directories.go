package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daddia/zen/internal/logging"
)

// Manager handles directory creation and management for Zen CLI
type Manager struct {
	logger logging.Logger
}

// New creates a new directory manager
func New(logger logging.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// CreateZenWorkspace creates the minimal .zen directory structure for workspace initialization
func (m *Manager) CreateZenWorkspace(zenDir string) error {
	m.logger.Debug("Creating zen workspace structure", "zen_dir", zenDir)

	// Create main .zen directory
	if err := os.MkdirAll(zenDir, 0755); err != nil {
		return fmt.Errorf("failed to create .zen directory: %w", err)
	}

	// Create essential subdirectories only
	essentialDirs := []string{
		"assets",   // Asset cache and manifest
		"cache",    // CLI caches
		"logs",     // CLI run logs, sync traces
		"work",     // Work directory (tasks will be created here)
		"metadata", // External system integration
	}

	for _, subdir := range essentialDirs {
		dirPath := filepath.Join(zenDir, subdir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
		m.logger.Debug("Created directory", "path", dirPath)
	}

	return nil
}

// CreateTaskDirectory creates the minimal task directory structure
func (m *Manager) CreateTaskDirectory(taskDir string) error {
	m.logger.Debug("Creating task directory structure", "task_dir", taskDir)

	// Create main task directory
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Create essential subdirectories only
	essentialDirs := []string{
		".zenflow", // Zenflow state tracking
		"metadata", // External system snapshots
	}

	for _, subdir := range essentialDirs {
		dirPath := filepath.Join(taskDir, subdir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
		m.logger.Debug("Created directory", "path", dirPath)
	}

	return nil
}

// CreateWorkTypeDirectory creates a work-type directory on demand
func (m *Manager) CreateWorkTypeDirectory(taskDir, workType string) error {
	dirPath := filepath.Join(taskDir, workType)
	m.logger.Debug("Creating work-type directory", "path", dirPath)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create work-type directory %s: %w", dirPath, err)
	}

	return nil
}

// EnsureDirectory creates a directory if it doesn't exist
func (m *Manager) EnsureDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		m.logger.Debug("Creating directory", "path", dirPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
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

// GetWorkTypeDirectories returns the list of work-type directories that can be created on demand
func (m *Manager) GetWorkTypeDirectories() []string {
	return []string{
		"research",  // Investigation and discovery work
		"spikes",    // Technical exploration and prototyping
		"design",    // Specifications and planning artifacts
		"execution", // Implementation evidence and results
		"outcomes",  // Learning, metrics, and retrospectives
	}
}
