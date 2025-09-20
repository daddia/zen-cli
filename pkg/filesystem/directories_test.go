package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := logging.NewBasic()
	manager := New(logger)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
}

func TestCreateZenWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	zenDir := filepath.Join(tempDir, ".zen")

	logger := logging.NewBasic()
	manager := New(logger)

	err := manager.CreateZenWorkspace(zenDir)
	require.NoError(t, err)

	// Check main .zen directory exists
	assert.DirExists(t, zenDir)

	// Check essential subdirectories exist
	expectedDirs := []string{
		"assets",
		"cache",
		"logs",
		"work",
		"metadata",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(zenDir, dir)
		assert.DirExists(t, dirPath, "Directory %s should exist", dir)
	}
}

func TestCreateTaskDirectory(t *testing.T) {
	tempDir := t.TempDir()
	taskDir := filepath.Join(tempDir, "PROJ-123")

	logger := logging.NewBasic()
	manager := New(logger)

	err := manager.CreateTaskDirectory(taskDir)
	require.NoError(t, err)

	// Check main task directory exists
	assert.DirExists(t, taskDir)

	// Check essential subdirectories exist
	expectedDirs := []string{
		".zenflow",
		"metadata",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(taskDir, dir)
		assert.DirExists(t, dirPath, "Directory %s should exist", dir)
	}
}

func TestCreateWorkTypeDirectory(t *testing.T) {
	tempDir := t.TempDir()
	taskDir := filepath.Join(tempDir, "PROJ-123")

	logger := logging.NewBasic()
	manager := New(logger)

	// Create task directory first
	err := manager.CreateTaskDirectory(taskDir)
	require.NoError(t, err)

	// Create work-type directory
	err = manager.CreateWorkTypeDirectory(taskDir, "research")
	require.NoError(t, err)

	// Check work-type directory exists
	researchDir := filepath.Join(taskDir, "research")
	assert.DirExists(t, researchDir)
}

func TestEnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "directory")

	logger := logging.NewBasic()
	manager := New(logger)

	// Directory shouldn't exist initially
	assert.NoDirExists(t, testDir)

	// Ensure directory creates it
	err := manager.EnsureDirectory(testDir)
	require.NoError(t, err)
	assert.DirExists(t, testDir)

	// Calling again should not error
	err = manager.EnsureDirectory(testDir)
	require.NoError(t, err)
	assert.DirExists(t, testDir)
}

func TestDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing")
	nonExistingDir := filepath.Join(tempDir, "nonexisting")

	logger := logging.NewBasic()
	manager := New(logger)

	// Create existing directory
	err := os.MkdirAll(existingDir, 0755)
	require.NoError(t, err)

	// Test existing directory
	assert.True(t, manager.DirectoryExists(existingDir))

	// Test non-existing directory
	assert.False(t, manager.DirectoryExists(nonExistingDir))

	// Test with file instead of directory
	filePath := filepath.Join(tempDir, "testfile")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)
	assert.False(t, manager.DirectoryExists(filePath))
}

func TestGetWorkTypeDirectories(t *testing.T) {
	logger := logging.NewBasic()
	manager := New(logger)

	workTypes := manager.GetWorkTypeDirectories()

	expected := []string{
		"research",
		"spikes",
		"design",
		"execution",
		"outcomes",
	}

	assert.Equal(t, expected, workTypes)
}
