package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	dirPath := filepath.Join(tempDir, "test-dir")
	err := manager.CreateDirectory(dirPath, 0755)

	require.NoError(t, err)
	assert.DirExists(t, dirPath)
}

func TestCreateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	dirs := []string{"dir1", "dir2", "dir3"}
	err := manager.CreateDirectories(tempDir, dirs, 0755)

	require.NoError(t, err)

	for _, dir := range dirs {
		assert.DirExists(t, filepath.Join(tempDir, dir))
	}
}

func TestDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	// Test existing directory
	assert.True(t, manager.DirectoryExists(tempDir))

	// Test non-existing directory
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	assert.False(t, manager.DirectoryExists(nonExistentDir))
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test existing file
	assert.True(t, manager.FileExists(testFile))

	// Test non-existing file
	nonExistentFile := filepath.Join(tempDir, "non-existent.txt")
	assert.False(t, manager.FileExists(nonExistentFile))

	// Test that directory returns false for FileExists
	assert.False(t, manager.FileExists(tempDir))
}

func TestCreateFile(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	filePath := filepath.Join(tempDir, "subdir", "test.txt")
	content := []byte("test content")

	err := manager.CreateFile(filePath, content, 0644)
	require.NoError(t, err)

	assert.FileExists(t, filePath)

	// Verify content
	readContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, readContent)
}

func TestReadFile(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	filePath := filepath.Join(tempDir, "test.txt")
	expectedContent := []byte("test content")

	err := os.WriteFile(filePath, expectedContent, 0644)
	require.NoError(t, err)

	content, err := manager.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestEnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	manager := New(logger)

	// Test creating new directory
	newDir := filepath.Join(tempDir, "new-dir")
	err := manager.EnsureDirectory(newDir, 0755)
	require.NoError(t, err)
	assert.DirExists(t, newDir)

	// Test ensuring existing directory (should not error)
	err = manager.EnsureDirectory(newDir, 0755)
	require.NoError(t, err)
	assert.DirExists(t, newDir)
}
