package git

import (
	"fmt"
	"time"
)

// Git data structures for enhanced operations

// Branch represents a Git branch
type Branch struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"is_current"`
	IsRemote  bool   `json:"is_remote"`
	Commit    string `json:"commit"`
	Message   string `json:"message"`
}

// Commit represents a Git commit
type Commit struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
	ShortHash string    `json:"short_hash"`
}

// CommitDetails represents detailed commit information
type CommitDetails struct {
	Commit
	Files      []string `json:"files"`
	Insertions int      `json:"insertions"`
	Deletions  int      `json:"deletions"`
	Diff       string   `json:"diff"`
}

// Stash represents a Git stash
type Stash struct {
	Index   int       `json:"index"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
	Branch  string    `json:"branch"`
}

// Remote represents a Git remote
type Remote struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Tag represents a Git tag
type Tag struct {
	Name    string    `json:"name"`
	Commit  string    `json:"commit"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

// BlameLine represents a line in git blame output
type BlameLine struct {
	Commit   string `json:"commit"`
	Author   string `json:"author"`
	Date     string `json:"date"`
	LineNum  int    `json:"line_num"`
	Content  string `json:"content"`
	Filename string `json:"filename"`
}

// StatusInfo represents git status information
type StatusInfo struct {
	Branch         string            `json:"branch"`
	Clean          bool              `json:"clean"`
	StagedFiles    []string          `json:"staged_files"`
	ModifiedFiles  []string          `json:"modified_files"`
	UntrackedFiles []string          `json:"untracked_files"`
	DeletedFiles   []string          `json:"deleted_files"`
	RenamedFiles   map[string]string `json:"renamed_files"`
}

// DiffOptions represents options for git diff
type DiffOptions struct {
	Files      []string `json:"files,omitempty"`
	Staged     bool     `json:"staged,omitempty"`
	Unstaged   bool     `json:"unstaged,omitempty"`
	Commit     string   `json:"commit,omitempty"`
	BaseCommit string   `json:"base_commit,omitempty"`
	Context    int      `json:"context,omitempty"`
}

// LogOptions represents options for git log
type LogOptions struct {
	Limit   int      `json:"limit,omitempty"`
	Since   string   `json:"since,omitempty"`
	Until   string   `json:"until,omitempty"`
	Author  string   `json:"author,omitempty"`
	Grep    string   `json:"grep,omitempty"`
	Files   []string `json:"files,omitempty"`
	Oneline bool     `json:"oneline,omitempty"`
	Graph   bool     `json:"graph,omitempty"`
	All     bool     `json:"all,omitempty"`
}

// GitError represents Git operation specific errors
type GitError struct {
	Code       GitErrorCode `json:"code"`
	Message    string       `json:"message"`
	Details    interface{}  `json:"details,omitempty"`
	RetryAfter int          `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e GitError) Error() string {
	return e.Message
}

// GitErrorCode represents specific Git operation error codes
type GitErrorCode string

const (
	ErrorCodeRepositoryNotFound GitErrorCode = "repository_not_found"
	ErrorCodeFileNotFound       GitErrorCode = "file_not_found"
	ErrorCodeCommandFailed      GitErrorCode = "command_failed"
	ErrorCodeAuthFailed         GitErrorCode = "auth_failed"
	ErrorCodeNetworkError       GitErrorCode = "network_error"
	ErrorCodeConfigError        GitErrorCode = "config_error"
)

// CommandError represents a Git command execution error
type CommandError struct {
	Command []string
	Output  string
	Err     error
}

func (e CommandError) Error() string {
	return fmt.Sprintf("git command failed: %v, output: %s", e.Err, e.Output)
}
