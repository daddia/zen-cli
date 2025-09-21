package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/errors"
)

// Repository represents a Git repository interface
type Repository interface {
	// Basic repository operations
	Clone(ctx context.Context, url, branch string, shallow bool) error
	Pull(ctx context.Context) error
	GetFile(ctx context.Context, path string) ([]byte, error)
	ListFiles(ctx context.Context, pattern string) ([]string, error)
	GetLastCommit(ctx context.Context) (string, error)
	IsClean(ctx context.Context) (bool, error)

	// Generic Git command execution - access to ALL Git commands
	ExecuteCommand(ctx context.Context, args ...string) (string, error)
	ExecuteCommandWithOutput(ctx context.Context, args ...string) ([]byte, error)

	// Branching operations
	CreateBranch(ctx context.Context, name string) error
	DeleteBranch(ctx context.Context, name string, force bool) error
	ListBranches(ctx context.Context, remote bool) ([]Branch, error)
	SwitchBranch(ctx context.Context, name string) error
	GetCurrentBranch(ctx context.Context) (string, error)

	// Commit operations
	Commit(ctx context.Context, message string, files ...string) error
	GetCommitHistory(ctx context.Context, limit int) ([]Commit, error)
	ShowCommit(ctx context.Context, hash string) (CommitDetails, error)
	AddFiles(ctx context.Context, files ...string) error

	// Merge and Rebase
	Merge(ctx context.Context, branch string, strategy string) error
	Rebase(ctx context.Context, branch string, interactive bool) error

	// Stashing
	Stash(ctx context.Context, message string) error
	StashPop(ctx context.Context, index int) error
	ListStashes(ctx context.Context) ([]Stash, error)

	// Remote operations
	AddRemote(ctx context.Context, name, url string) error
	ListRemotes(ctx context.Context) ([]Remote, error)
	Fetch(ctx context.Context, remote string) error
	Push(ctx context.Context, remote, branch string) error

	// Configuration
	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key, value string, global bool) error

	// Advanced operations
	Diff(ctx context.Context, options DiffOptions) (string, error)
	Log(ctx context.Context, options LogOptions) ([]Commit, error)
	Blame(ctx context.Context, file string) ([]BlameLine, error)
	Tag(ctx context.Context, name, message string) error
	ListTags(ctx context.Context) ([]Tag, error)
	Status(ctx context.Context) (StatusInfo, error)
}

// AuthProvider provides authentication for Git operations
type AuthProvider interface {
	// GetCredentials returns credentials for the specified provider
	GetCredentials(provider string) (string, error)
}

// CLIRepository implements Repository using Git CLI wrapper
type CLIRepository struct {
	repoPath     string
	logger       logging.Logger
	auth         AuthProvider
	authProvider string
}

// NewCLIRepository creates a new Git CLI repository wrapper
func NewCLIRepository(repoPath string, logger logging.Logger, auth AuthProvider, authProvider string) *CLIRepository {
	return &CLIRepository{
		repoPath:     repoPath,
		logger:       logger,
		auth:         auth,
		authProvider: authProvider,
	}
}

// Clone clones the repository to local cache
func (g *CLIRepository) Clone(ctx context.Context, url, branch string, shallow bool) error {
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
func (g *CLIRepository) Pull(ctx context.Context) error {
	g.logger.Debug("pulling repository updates")

	// Check if repository exists
	if !g.repositoryExists() {
		return &GitError{
			Code:    ErrorCodeRepositoryNotFound,
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
func (g *CLIRepository) GetFile(ctx context.Context, path string) ([]byte, error) {
	g.logger.Debug("getting file from repository", "path", path)

	if !g.repositoryExists() {
		return nil, &GitError{
			Code:    ErrorCodeRepositoryNotFound,
			Message: "repository not found",
		}
	}

	filePath := filepath.Join(g.repoPath, path)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, &GitError{
				Code:    ErrorCodeFileNotFound,
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
func (g *CLIRepository) ListFiles(ctx context.Context, pattern string) ([]string, error) {
	g.logger.Debug("listing repository files", "pattern", pattern)

	if !g.repositoryExists() {
		return nil, &GitError{
			Code:    ErrorCodeRepositoryNotFound,
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
func (g *CLIRepository) GetLastCommit(ctx context.Context) (string, error) {
	g.logger.Debug("getting last commit hash")

	if !g.repositoryExists() {
		return "", &GitError{
			Code:    ErrorCodeRepositoryNotFound,
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
func (g *CLIRepository) IsClean(ctx context.Context) (bool, error) {
	g.logger.Debug("checking repository status")

	if !g.repositoryExists() {
		return false, &GitError{
			Code:    ErrorCodeRepositoryNotFound,
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

// ExecuteCommand executes any Git command with full flexibility
func (g *CLIRepository) ExecuteCommand(ctx context.Context, args ...string) (string, error) {
	g.logger.Debug("executing generic git command", "args", g.sanitizeArgs(args))

	if !g.repositoryExists() {
		return "", &GitError{
			Code:    ErrorCodeRepositoryNotFound,
			Message: "repository not found",
		}
	}

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, args...)
	if err != nil {
		return "", err
	}

	return output, nil
}

// ExecuteCommandWithOutput executes any Git command and returns raw bytes
func (g *CLIRepository) ExecuteCommandWithOutput(ctx context.Context, args ...string) ([]byte, error) {
	g.logger.Debug("executing generic git command with output", "args", g.sanitizeArgs(args))

	if !g.repositoryExists() {
		return nil, &GitError{
			Code:    ErrorCodeRepositoryNotFound,
			Message: "repository not found",
		}
	}

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, args...)
	if err != nil {
		return nil, err
	}

	return []byte(output), nil
}

// Private helper methods

func (g *CLIRepository) executeGitCommand(ctx context.Context, workDir string, args ...string) error {
	_, err := g.executeGitCommandWithOutput(ctx, workDir, args...)
	return err
}

func (g *CLIRepository) executeGitCommandWithOutput(ctx context.Context, workDir string, args ...string) (string, error) {
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

		return "", &GitError{
			Code:    ErrorCodeCommandFailed,
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

func (g *CLIRepository) setupAuthentication(cmd *exec.Cmd) error {
	if g.auth == nil {
		return nil // No authentication configured
	}

	// Get credentials from auth provider
	credentials, err := g.auth.GetCredentials(g.authProvider)
	if err != nil {
		g.logger.Debug("no credentials available for git operations", "error", err)
		return nil // Continue without authentication
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

func (g *CLIRepository) repositoryExists() bool {
	gitDir := filepath.Join(g.repoPath, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

func (g *CLIRepository) matchPattern(path, pattern string) bool {
	// Simple pattern matching - can be enhanced with more sophisticated matching
	if pattern == "*" || pattern == "" {
		return true
	}

	// Check if path contains pattern
	return strings.Contains(strings.ToLower(path), strings.ToLower(pattern))
}

func (g *CLIRepository) sanitizeURL(url string) string {
	// Remove credentials from URL for logging
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			return "https://***@" + parts[len(parts)-1]
		}
	}
	return url
}

func (g *CLIRepository) sanitizeArgs(args []string) []string {
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

// Branching operations

// CreateBranch creates a new branch
func (g *CLIRepository) CreateBranch(ctx context.Context, name string) error {
	g.logger.Debug("creating branch", "name", name)
	return g.executeGitCommand(ctx, g.repoPath, "branch", name)
}

// DeleteBranch deletes a branch
func (g *CLIRepository) DeleteBranch(ctx context.Context, name string, force bool) error {
	g.logger.Debug("deleting branch", "name", name, "force", force)

	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, name)

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// ListBranches lists all branches
func (g *CLIRepository) ListBranches(ctx context.Context, remote bool) ([]Branch, error) {
	g.logger.Debug("listing branches", "remote", remote)

	args := []string{"branch"}
	if remote {
		args = append(args, "-r")
	}
	args = append(args, "--format", "%(refname:short)|%(objectname)|%(contents:subject)")

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, args...)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			branch := Branch{
				Name:     parts[0],
				Commit:   parts[1],
				Message:  parts[2],
				IsRemote: remote,
			}
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

// SwitchBranch switches to a branch
func (g *CLIRepository) SwitchBranch(ctx context.Context, name string) error {
	g.logger.Debug("switching branch", "name", name)
	return g.executeGitCommand(ctx, g.repoPath, "checkout", name)
}

// GetCurrentBranch returns the current branch name
func (g *CLIRepository) GetCurrentBranch(ctx context.Context) (string, error) {
	g.logger.Debug("getting current branch")

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// Commit operations

// Commit commits changes to the repository
func (g *CLIRepository) Commit(ctx context.Context, message string, files ...string) error {
	g.logger.Debug("committing changes", "message", message, "files", files)

	// Add files if specified
	if len(files) > 0 {
		args := append([]string{"add"}, files...)
		if err := g.executeGitCommand(ctx, g.repoPath, args...); err != nil {
			return err
		}
	}

	// Commit with message
	return g.executeGitCommand(ctx, g.repoPath, "commit", "-m", message)
}

// AddFiles adds files to the staging area
func (g *CLIRepository) AddFiles(ctx context.Context, files ...string) error {
	g.logger.Debug("adding files to staging", "files", files)

	args := append([]string{"add"}, files...)
	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// GetCommitHistory returns commit history
func (g *CLIRepository) GetCommitHistory(ctx context.Context, limit int) ([]Commit, error) {
	g.logger.Debug("getting commit history", "limit", limit)

	args := []string{"log", "--format=%H|%an|%ae|%ad|%s", "--date=iso"}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, args...)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			// Parse date
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
			commit := Commit{
				Hash:      parts[0],
				Author:    parts[1],
				Email:     parts[2],
				Date:      date,
				Message:   parts[4],
				ShortHash: parts[0][:8],
			}
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

// ShowCommit shows detailed commit information
func (g *CLIRepository) ShowCommit(ctx context.Context, hash string) (CommitDetails, error) {
	g.logger.Debug("showing commit details", "hash", hash)

	// Get basic commit info
	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "show", "--stat", "--format=%H|%an|%ae|%ad|%s", hash)
	if err != nil {
		return CommitDetails{}, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return CommitDetails{}, &GitError{
			Code:    ErrorCodeFileNotFound,
			Message: "commit not found",
		}
	}

	// Parse commit info
	parts := strings.Split(lines[0], "|")
	if len(parts) < 5 {
		return CommitDetails{}, &GitError{
			Code:    ErrorCodeCommandFailed,
			Message: "invalid commit format",
		}
	}

	date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
	commit := Commit{
		Hash:      parts[0],
		Author:    parts[1],
		Email:     parts[2],
		Date:      date,
		Message:   parts[4],
		ShortHash: parts[0][:8],
	}

	// Get diff
	diffOutput, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "show", "--format=", hash)
	if err != nil {
		diffOutput = ""
	}

	return CommitDetails{
		Commit: commit,
		Diff:   diffOutput,
	}, nil
}

// Remote operations

// AddRemote adds a remote repository
func (g *CLIRepository) AddRemote(ctx context.Context, name, url string) error {
	g.logger.Debug("adding remote", "name", name, "url", g.sanitizeURL(url))
	return g.executeGitCommand(ctx, g.repoPath, "remote", "add", name, url)
}

// ListRemotes lists all remotes
func (g *CLIRepository) ListRemotes(ctx context.Context) ([]Remote, error) {
	g.logger.Debug("listing remotes")

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "remote", "-v")
	if err != nil {
		return nil, err
	}

	var remotes []Remote
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remote := Remote{
				Name: parts[0],
				URL:  parts[1],
			}
			remotes = append(remotes, remote)
		}
	}

	return remotes, nil
}

// Fetch fetches from a remote
func (g *CLIRepository) Fetch(ctx context.Context, remote string) error {
	g.logger.Debug("fetching from remote", "remote", remote)

	args := []string{"fetch"}
	if remote != "" {
		args = append(args, remote)
	}

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// Push pushes to a remote
func (g *CLIRepository) Push(ctx context.Context, remote, branch string) error {
	g.logger.Debug("pushing to remote", "remote", remote, "branch", branch)

	args := []string{"push"}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// Configuration

// GetConfig gets a Git configuration value
func (g *CLIRepository) GetConfig(ctx context.Context, key string) (string, error) {
	g.logger.Debug("getting config", "key", key)

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "config", "--get", key)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// SetConfig sets a Git configuration value
func (g *CLIRepository) SetConfig(ctx context.Context, key, value string, global bool) error {
	g.logger.Debug("setting config", "key", key, "global", global)

	args := []string{"config"}
	if global {
		args = append(args, "--global")
	}
	args = append(args, key, value)

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// Advanced operations

// Diff shows differences
func (g *CLIRepository) Diff(ctx context.Context, options DiffOptions) (string, error) {
	g.logger.Debug("showing diff", "options", options)

	args := []string{"diff"}

	if options.Staged {
		args = append(args, "--cached")
	}
	if options.Context > 0 {
		args = append(args, fmt.Sprintf("-%d", options.Context))
	}
	if options.Commit != "" {
		args = append(args, options.Commit)
	}
	if options.BaseCommit != "" {
		args = append(args, options.BaseCommit)
	}
	if len(options.Files) > 0 {
		args = append(args, "--")
		args = append(args, options.Files...)
	}

	return g.executeGitCommandWithOutput(ctx, g.repoPath, args...)
}

// Log shows commit log
func (g *CLIRepository) Log(ctx context.Context, options LogOptions) ([]Commit, error) {
	g.logger.Debug("showing log", "options", options)

	args := []string{"log", "--format=%H|%an|%ae|%ad|%s", "--date=iso"}

	if options.Limit > 0 {
		args = append(args, fmt.Sprintf("-%d", options.Limit))
	}
	if options.Since != "" {
		args = append(args, "--since", options.Since)
	}
	if options.Until != "" {
		args = append(args, "--until", options.Until)
	}
	if options.Author != "" {
		args = append(args, "--author", options.Author)
	}
	if options.Grep != "" {
		args = append(args, "--grep", options.Grep)
	}
	if options.Oneline {
		args = append(args, "--oneline")
	}
	if options.Graph {
		args = append(args, "--graph")
	}
	if options.All {
		args = append(args, "--all")
	}
	if len(options.Files) > 0 {
		args = append(args, "--")
		args = append(args, options.Files...)
	}

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, args...)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
			commit := Commit{
				Hash:      parts[0],
				Author:    parts[1],
				Email:     parts[2],
				Date:      date,
				Message:   parts[4],
				ShortHash: parts[0][:8],
			}
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

// Blame shows blame information for a file
func (g *CLIRepository) Blame(ctx context.Context, file string) ([]BlameLine, error) {
	g.logger.Debug("showing blame", "file", file)

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "blame", "-l", file)
	if err != nil {
		return nil, err
	}

	var blameLines []BlameLine
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		// Parse blame line format: commit author date line_num content
		parts := strings.Split(line, " ")
		if len(parts) >= 4 {
			blameLine := BlameLine{
				Commit:   parts[0],
				Author:   parts[1],
				Date:     parts[2],
				LineNum:  i + 1,
				Content:  strings.Join(parts[3:], " "),
				Filename: file,
			}
			blameLines = append(blameLines, blameLine)
		}
	}

	return blameLines, nil
}

// Tag creates a tag
func (g *CLIRepository) Tag(ctx context.Context, name, message string) error {
	g.logger.Debug("creating tag", "name", name, "message", message)

	args := []string{"tag"}
	if message != "" {
		args = append(args, "-a", name, "-m", message)
	} else {
		args = append(args, name)
	}

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// ListTags lists all tags
func (g *CLIRepository) ListTags(ctx context.Context) ([]Tag, error) {
	g.logger.Debug("listing tags")

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "tag", "-l", "--format=%(refname:short)|%(objectname)|%(contents:subject)|%(creatordate:iso)")
	if err != nil {
		return nil, err
	}

	var tags []Tag
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
			tag := Tag{
				Name:    parts[0],
				Commit:  parts[1],
				Message: parts[2],
				Date:    date,
			}
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// Status shows repository status
func (g *CLIRepository) Status(ctx context.Context) (StatusInfo, error) {
	g.logger.Debug("showing status")

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "status", "--porcelain", "-b")
	if err != nil {
		return StatusInfo{}, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return StatusInfo{}, nil
	}

	status := StatusInfo{
		Clean:          true,
		StagedFiles:    []string{},
		ModifiedFiles:  []string{},
		UntrackedFiles: []string{},
		DeletedFiles:   []string{},
		RenamedFiles:   make(map[string]string),
	}

	// Parse branch info from first line
	if strings.HasPrefix(lines[0], "## ") {
		branchInfo := strings.TrimPrefix(lines[0], "## ")
		if parts := strings.Split(branchInfo, "..."); len(parts) > 0 {
			status.Branch = parts[0]
		}
		lines = lines[1:]
	}

	// Parse file status
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		filename := line[3:]

		status.Clean = false

		switch statusCode[0] {
		case 'A', 'M', 'D':
			status.StagedFiles = append(status.StagedFiles, filename)
		}

		switch statusCode[1] {
		case 'M':
			status.ModifiedFiles = append(status.ModifiedFiles, filename)
		case 'D':
			status.DeletedFiles = append(status.DeletedFiles, filename)
		case '?':
			status.UntrackedFiles = append(status.UntrackedFiles, filename)
		}

		if statusCode[0] == 'R' {
			// Renamed file
			parts := strings.Split(filename, " -> ")
			if len(parts) == 2 {
				status.RenamedFiles[parts[0]] = parts[1]
			}
		}
	}

	return status, nil
}

// Merge merges a branch into current branch
func (g *CLIRepository) Merge(ctx context.Context, branch string, strategy string) error {
	g.logger.Debug("merging branch", "branch", branch, "strategy", strategy)

	args := []string{"merge"}
	if strategy != "" {
		args = append(args, "-s", strategy)
	}
	args = append(args, branch)

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// Rebase rebases current branch onto another branch
func (g *CLIRepository) Rebase(ctx context.Context, branch string, interactive bool) error {
	g.logger.Debug("rebasing branch", "branch", branch, "interactive", interactive)

	args := []string{"rebase"}
	if interactive {
		args = append(args, "-i")
	}
	args = append(args, branch)

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// Stash stashes current changes
func (g *CLIRepository) Stash(ctx context.Context, message string) error {
	g.logger.Debug("stashing changes", "message", message)

	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// StashPop applies and removes a stash
func (g *CLIRepository) StashPop(ctx context.Context, index int) error {
	g.logger.Debug("applying stash", "index", index)

	args := []string{"stash", "pop"}
	if index > 0 {
		args = append(args, fmt.Sprintf("stash@{%d}", index))
	}

	return g.executeGitCommand(ctx, g.repoPath, args...)
}

// ListStashes lists all stashes
func (g *CLIRepository) ListStashes(ctx context.Context) ([]Stash, error) {
	g.logger.Debug("listing stashes")

	output, err := g.executeGitCommandWithOutput(ctx, g.repoPath, "stash", "list", "--format=%gd|%gs|%gd|%gs")
	if err != nil {
		return nil, err
	}

	var stashes []Stash
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			stash := Stash{
				Index:   i,
				Message: parts[1],
				Branch:  "current", // Could be enhanced to parse actual branch
			}
			stashes = append(stashes, stash)
		}
	}

	return stashes, nil
}

// ValidateGitInstallation checks if Git is installed and accessible
func ValidateGitInstallation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if IsGitNotFound(err) {
			return &GitError{
				Code:    ErrorCodeConfigError,
				Message: "Git is not installed or not found in PATH",
				Details: "Please install Git and ensure it's available in your system PATH",
			}
		}
		return errors.Wrap(err, "failed to check Git installation")
	}

	version := strings.TrimSpace(string(output))
	if !strings.Contains(version, "git version") {
		return &GitError{
			Code:    ErrorCodeConfigError,
			Message: "Invalid Git installation detected",
			Details: fmt.Sprintf("Unexpected git --version output: %s", version),
		}
	}

	return nil
}

// IsGitNotFound checks if the error is due to Git not being found
func IsGitNotFound(err error) bool {
	if execErr, ok := err.(*exec.Error); ok {
		return execErr.Err == exec.ErrNotFound
	}
	return false
}
