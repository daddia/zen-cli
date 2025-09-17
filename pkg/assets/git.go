package assets

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/errors"
)

// GitCLIRepository implements GitRepository using Git CLI wrapper
type GitCLIRepository struct {
	repoPath     string
	logger       logging.Logger
	auth         AuthProvider
	authProvider string
}

// NewGitCLIRepository creates a new Git CLI repository wrapper
func NewGitCLIRepository(repoPath string, logger logging.Logger, auth AuthProvider, authProvider string) *GitCLIRepository {
	return &GitCLIRepository{
		repoPath:     repoPath,
		logger:       logger,
		auth:         auth,
		authProvider: authProvider,
	}
}

// Clone clones the repository to local cache
func (g *GitCLIRepository) Clone(ctx context.Context, url, branch string, shallow bool) error {
	g.logger.Info("cloning repository", "url", g.sanitizeURL(url), "branch", branch, "shallow", shallow)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(g.repoPath), 0755); err != nil {
		return errors.Wrap(err, "failed to create repository parent directory")
	}

	// Remove existing directory if it exists
	if err := os.RemoveAll(g.repoPath); err != nil {
		return errors.Wrap(err, "failed to remove existing repository directory")
	}

	// Build git clone command
	args := []string{"clone"}

	if shallow {
		args = append(args, "--depth", "1")
	}

	if branch != "" {
		args = append(args, "--branch", branch)
	}

	args = append(args, url, g.repoPath)

	// Execute git clone with authentication
	if err := g.executeGitCommand(ctx, "", args...); err != nil {
		return errors.Wrap(err, "git clone failed")
	}

	g.logger.Info("repository cloned successfully", "path", g.repoPath)
	return nil
}

// Pull updates the local repository
func (g *GitCLIRepository) Pull(ctx context.Context) error {
	g.logger.Debug("pulling repository updates")

	// Check if repository exists
	if !g.repositoryExists() {
		return &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: "repository not found, clone required",
		}
	}

	// Execute git pull
	if err := g.executeGitCommand(ctx, g.repoPath, "pull"); err != nil {
		return errors.Wrap(err, "git pull failed")
	}

	g.logger.Debug("repository updated successfully")
	return nil
}

// GetFile retrieves a file from the repository
func (g *GitCLIRepository) GetFile(ctx context.Context, path string) ([]byte, error) {
	g.logger.Debug("getting file from repository", "path", path)

	if !g.repositoryExists() {
		return nil, &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: "repository not found",
		}
	}

	filePath := filepath.Join(g.repoPath, path)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, &AssetClientError{
				Code:    ErrorCodeAssetNotFound,
				Message: fmt.Sprintf("file '%s' not found in repository", path),
			}
		}
		return nil, errors.Wrap(err, "failed to stat file")
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	g.logger.Debug("file retrieved successfully", "path", path, "size", len(content))
	return content, nil
}

// ListFiles lists all files in the repository
func (g *GitCLIRepository) ListFiles(ctx context.Context, pattern string) ([]string, error) {
	g.logger.Debug("listing repository files", "pattern", pattern)

	if !g.repositoryExists() {
		return nil, &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: "repository not found",
		}
	}

	var files []string

	err := filepath.WalkDir(g.repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(g.repoPath, path)
		if err != nil {
			return err
		}

		// Apply pattern matching if specified
		if pattern == "" || g.matchPattern(relPath, pattern) {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to walk repository directory")
	}

	g.logger.Debug("files listed successfully", "count", len(files))
	return files, nil
}

// GetLastCommit returns the last commit hash
func (g *GitCLIRepository) GetLastCommit(ctx context.Context) (string, error) {
	g.logger.Debug("getting last commit hash")

	if !g.repositoryExists() {
		return "", &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: "repository not found",
		}
	}

	// Execute git rev-parse HEAD
	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", errors.Wrap(err, "failed to get commit hash")
	}

	commit := strings.TrimSpace(output)
	g.logger.Debug("last commit retrieved", "commit", commit[:8]+"...")
	return commit, nil
}

// IsClean returns true if the repository has no uncommitted changes
func (g *GitCLIRepository) IsClean(ctx context.Context) (bool, error) {
	g.logger.Debug("checking repository status")

	if !g.repositoryExists() {
		return false, &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: "repository not found",
		}
	}

	// Execute git status --porcelain
	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "status", "--porcelain")
	if err != nil {
		return false, errors.Wrap(err, "failed to check repository status")
	}

	clean := strings.TrimSpace(output) == ""
	g.logger.Debug("repository status checked", "clean", clean)
	return clean, nil
}

// Private helper methods

func (g *GitCLIRepository) executeGitCommand(ctx context.Context, workDir string, args ...string) error {
	_, err := g.executeGitCommandWithOutput(ctx, workDir, args...)
	return err
}

func (g *GitCLIRepository) executeGitCommandWithOutput(ctx context.Context, workDir string, args ...string) (string, error) {
	// Create command with timeout
	cmd := exec.CommandContext(ctx, "git", args...)

	if workDir != "" {
		cmd.Dir = workDir
	}

	// Set up authentication if needed
	if err := g.setupAuthentication(cmd); err != nil {
		return "", errors.Wrap(err, "failed to setup authentication")
	}

	// Set up environment
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0", // Disable interactive prompts
		"GIT_ASKPASS=echo",      // Use echo as askpass to prevent hanging
	)

	g.logger.Debug("executing git command", "args", g.sanitizeArgs(args), "workdir", workDir)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		g.logger.Error("git command failed",
			"args", g.sanitizeArgs(args),
			"error", err,
			"output", string(output))

		return "", &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: fmt.Sprintf("git command failed: %v", err),
			Details: map[string]interface{}{
				"command": g.sanitizeArgs(args),
				"output":  string(output),
			},
		}
	}

	g.logger.Debug("git command completed successfully", "args", g.sanitizeArgs(args))
	return string(output), nil
}

func (g *GitCLIRepository) setupAuthentication(cmd *exec.Cmd) error {
	// Get credentials from auth provider
	credentials, err := g.auth.GetCredentials(g.authProvider)
	if err != nil {
		return errors.Wrap(err, "failed to get credentials")
	}

	if credentials == "" {
		return nil // No credentials available
	}

	// Set up credential helper for HTTPS authentication
	cmd.Env = append(cmd.Env,
		"GIT_ASKPASS=echo",
		"GIT_USERNAME=token",
		fmt.Sprintf("GIT_PASSWORD=%s", credentials),
	)

	// For HTTPS URLs, we can use the credential helper approach
	cmd.Env = append(cmd.Env, "GIT_CONFIG_COUNT=1")
	cmd.Env = append(cmd.Env, "GIT_CONFIG_KEY_0=credential.helper")
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_CONFIG_VALUE_0=!echo 'username=token'; echo 'password=%s'", credentials))

	return nil
}

func (g *GitCLIRepository) repositoryExists() bool {
	gitDir := filepath.Join(g.repoPath, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

func (g *GitCLIRepository) matchPattern(path, pattern string) bool {
	// Simple pattern matching - can be enhanced with more sophisticated matching
	if pattern == "*" || pattern == "" {
		return true
	}

	// Check if path contains pattern
	return strings.Contains(strings.ToLower(path), strings.ToLower(pattern))
}

func (g *GitCLIRepository) sanitizeURL(url string) string {
	// Remove credentials from URL for logging
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			return "https://***@" + parts[len(parts)-1]
		}
	}
	return url
}

func (g *GitCLIRepository) sanitizeArgs(args []string) []string {
	sanitized := make([]string, len(args))
	copy(sanitized, args)

	// Replace URLs with sanitized versions
	for i, arg := range sanitized {
		if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "git@") {
			sanitized[i] = g.sanitizeURL(arg)
		}
	}

	return sanitized
}

// GitCommandError represents a Git command execution error
type GitCommandError struct {
	Command []string
	Output  string
	Err     error
}

func (e GitCommandError) Error() string {
	return fmt.Sprintf("git command failed: %v, output: %s", e.Err, e.Output)
}

// IsGitNotFound checks if the error is due to Git not being found
func IsGitNotFound(err error) bool {
	if execErr, ok := err.(*exec.Error); ok {
		return execErr.Err == exec.ErrNotFound
	}
	return false
}

// ValidateGitInstallation checks if Git is installed and accessible
func ValidateGitInstallation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if IsGitNotFound(err) {
			return &AssetClientError{
				Code:    ErrorCodeConfigurationError,
				Message: "Git is not installed or not found in PATH",
				Details: "Please install Git and ensure it's available in your system PATH",
			}
		}
		return errors.Wrap(err, "failed to check Git installation")
	}

	version := strings.TrimSpace(string(output))
	if !strings.Contains(version, "git version") {
		return &AssetClientError{
			Code:    ErrorCodeConfigurationError,
			Message: "Invalid Git installation detected",
			Details: fmt.Sprintf("Unexpected git --version output: %s", version),
		}
	}

	return nil
}
