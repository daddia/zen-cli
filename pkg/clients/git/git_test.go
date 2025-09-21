package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthProvider implements AuthProvider for testing
type mockAuthProvider struct {
	credentials map[string]string
	errors      map[string]error
}

func newMockAuthProvider() *mockAuthProvider {
	return &mockAuthProvider{
		credentials: make(map[string]string),
		errors:      make(map[string]error),
	}
}

func (m *mockAuthProvider) GetCredentials(provider string) (string, error) {
	if err, exists := m.errors[provider]; exists {
		return "", err
	}
	return m.credentials[provider], nil
}

func TestNewCLIRepository(t *testing.T) {
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository("/test/path", logger, auth, "github")

	assert.NotNil(t, repo)
	assert.Equal(t, "/test/path", repo.repoPath)
	assert.Equal(t, logger, repo.logger)
	assert.Equal(t, auth, repo.auth)
	assert.Equal(t, "github", repo.authProvider)
}

func TestCLIRepository_RepositoryExists(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")

	// Initially should not exist
	assert.False(t, repo.repositoryExists())

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Now should exist
	assert.True(t, repo.repositoryExists())
}

func TestCLIRepository_GetFile_RepositoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Try to get file from non-existent repository
	content, err := repo.GetFile(ctx, "test.txt")

	assert.Nil(t, content)
	assert.Error(t, err)

	var gitErr *GitError
	assert.ErrorAs(t, err, &gitErr)
	assert.Equal(t, ErrorCodeRepositoryNotFound, gitErr.Code)
	assert.Contains(t, gitErr.Message, "repository not found")
}

func TestCLIRepository_GetFile_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Create .git directory to simulate repository exists
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Try to get non-existent file
	content, err := repo.GetFile(ctx, "nonexistent.txt")

	assert.Nil(t, content)
	assert.Error(t, err)

	var gitErr *GitError
	assert.ErrorAs(t, err, &gitErr)
	assert.Equal(t, ErrorCodeFileNotFound, gitErr.Code)
	assert.Contains(t, gitErr.Message, "not found in repository")
}

func TestCLIRepository_GetFile_Success(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create test file
	testContent := "Hello, World!"
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// Get file
	content, err := repo.GetFile(ctx, "test.txt")

	require.NoError(t, err)
	assert.Equal(t, []byte(testContent), content)
}

func TestCLIRepository_ListFiles_RepositoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Try to list files from non-existent repository
	files, err := repo.ListFiles(ctx, "")

	assert.Nil(t, files)
	assert.Error(t, err)

	var gitErr *GitError
	assert.ErrorAs(t, err, &gitErr)
	assert.Equal(t, ErrorCodeRepositoryNotFound, gitErr.Code)
}

func TestCLIRepository_ListFiles_Success(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create test files
	err = os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "file2.md"), []byte("content2"), 0644)
	require.NoError(t, err)

	// Create subdirectory with file
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("content3"), 0644)
	require.NoError(t, err)

	// List all files
	files, err := repo.ListFiles(ctx, "")

	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Contains(t, files, "file1.txt")
	assert.Contains(t, files, "file2.md")
	assert.Contains(t, files, filepath.Join("subdir", "file3.txt"))
}

func TestCLIRepository_MatchPattern(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")

	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"test.txt", "", true},          // Empty pattern matches all
		{"test.txt", "*", true},         // Wildcard matches all
		{"test.txt", "test", true},      // Partial match
		{"test.txt", "txt", true},       // Extension match
		{"test.txt", "md", false},       // No match
		{"path/to/file.go", "go", true}, // Nested file match
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.pattern, func(t *testing.T) {
			result := repo.matchPattern(tt.path, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCLIRepository_SanitizeURL(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			input:    "https://token:x-oauth-basic@github.com/user/repo.git",
			expected: "https://***@github.com/user/repo.git",
		},
		{
			input:    "https://user:password@gitlab.com/user/repo.git",
			expected: "https://***@gitlab.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := repo.sanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCLIRepository_SanitizeArgs(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")

	args := []string{
		"clone",
		"https://token:secret@github.com/user/repo.git",
		"destination",
	}

	sanitized := repo.sanitizeArgs(args)

	assert.Equal(t, "clone", sanitized[0])
	assert.Equal(t, "https://***@github.com/user/repo.git", sanitized[1])
	assert.Equal(t, "destination", sanitized[2])
}

func TestCLIRepository_IsClean_RepositoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Try to check status of non-existent repository
	clean, err := repo.IsClean(ctx)

	assert.False(t, clean)
	assert.Error(t, err)

	var gitErr *GitError
	assert.ErrorAs(t, err, &gitErr)
	assert.Equal(t, ErrorCodeRepositoryNotFound, gitErr.Code)
}

func TestCLIRepository_GetLastCommit_RepositoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Try to get commit from non-existent repository
	commit, err := repo.GetLastCommit(ctx)

	assert.Empty(t, commit)
	assert.Error(t, err)

	var gitErr *GitError
	assert.ErrorAs(t, err, &gitErr)
	assert.Equal(t, ErrorCodeRepositoryNotFound, gitErr.Code)
}

func TestCLIRepository_ExecuteCommand_RepositoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	logger := logging.NewBasic()
	auth := newMockAuthProvider()

	repo := NewCLIRepository(tempDir, logger, auth, "github")
	ctx := context.Background()

	// Try to execute command on non-existent repository
	output, err := repo.ExecuteCommand(ctx, "status")

	assert.Empty(t, output)
	assert.Error(t, err)

	var gitErr *GitError
	assert.ErrorAs(t, err, &gitErr)
	assert.Equal(t, ErrorCodeRepositoryNotFound, gitErr.Code)
}

func TestValidateGitInstallation(t *testing.T) {
	ctx := context.Background()

	// This test will depend on whether Git is installed on the system
	err := ValidateGitInstallation(ctx)

	if err != nil {
		// If Git is not installed, should get appropriate error
		var gitErr *GitError
		if assert.ErrorAs(t, err, &gitErr) {
			assert.Equal(t, ErrorCodeConfigError, gitErr.Code)
		}
	}
	// If Git is installed, err should be nil
}

func TestIsGitNotFound(t *testing.T) {
	// Test with different error types
	result := IsGitNotFound(assert.AnError)
	assert.False(t, result)

	// Test with nil error
	result = IsGitNotFound(nil)
	assert.False(t, result)

	// Note: Testing actual exec.Error would require executing invalid commands
	// which could be flaky, so we test the negative cases instead
}

func TestCommandError(t *testing.T) {
	cmdErr := CommandError{
		Command: []string{"git", "status"},
		Output:  "fatal: not a git repository",
		Err:     assert.AnError,
	}

	errorMsg := cmdErr.Error()
	assert.Contains(t, errorMsg, "git command failed")
	assert.Contains(t, errorMsg, "fatal: not a git repository")
}

func TestGitError(t *testing.T) {
	gitErr := GitError{
		Code:    ErrorCodeRepositoryNotFound,
		Message: "repository not found",
		Details: "test details",
	}

	assert.Equal(t, "repository not found", gitErr.Error())
	assert.Equal(t, ErrorCodeRepositoryNotFound, gitErr.Code)
	assert.Equal(t, "test details", gitErr.Details)
}
