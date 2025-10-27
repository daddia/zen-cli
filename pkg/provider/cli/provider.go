package cli

import (
	"github.com/daddia/zen/pkg/provider"
)

// CLIProvider extends the base Provider interface with CLI-specific capabilities
//
// CLIProvider is designed for providers that wrap external CLI tools.
// It embeds the standard provider.Provider interface and adds methods specific
// to CLI execution:
//   - BinaryPath() - returns the absolute path to the CLI binary
//   - ExecArgsFor() - maps operations to command-line arguments
//   - WorkDir() - returns the working directory for execution
//   - Env() - returns environment variables for execution
//
// This sub-interface allows CLI-specific code to access additional capabilities
// while maintaining compatibility with the unified Provider interface.
//
// Example implementation:
//
//	type GitProvider struct {
//	    binaryPath string
//	    workDir    string
//	    env        []string
//	}
//
//	func (g *GitProvider) BinaryPath() string {
//	    return g.binaryPath
//	}
//
//	func (g *GitProvider) ExecArgsFor(op string, params map[string]any) ([]string, error) {
//	    switch op {
//	    case "git.clone":
//	        return []string{"clone", params["url"].(string)}, nil
//	    case "git.status":
//	        return []string{"status", "--porcelain"}, nil
//	    default:
//	        return nil, provider.ErrInvalidOp("git", op)
//	    }
//	}
//
// CLIProvider implementations must be safe for concurrent use by multiple goroutines.
type CLIProvider interface {
	// Embed the base Provider interface
	provider.Provider

	// BinaryPath returns the absolute path to the CLI binary
	//
	// This is the path that will be executed. It should be an absolute path
	// to ensure consistent execution regardless of the caller's working directory.
	//
	// Example: "/usr/local/bin/git"
	BinaryPath() string

	// ExecArgsFor converts an operation and parameters into CLI arguments
	//
	// This method provides declarative operation mapping:
	//   - Input: operation name (e.g., "git.clone") and parameters
	//   - Output: command-line arguments for the binary
	//
	// The operation string follows a hierarchical naming convention:
	//   - Format: "<provider>.<action>"
	//   - Examples: "git.clone", "git.commit", "terraform.apply"
	//
	// Parameters are operation-specific and provided as a map:
	//   - git.clone: {"url": "...", "directory": "...", "depth": 1}
	//   - git.commit: {"message": "...", "files": [...]}
	//
	// Returns:
	//   - []string: command-line arguments to execute
	//   - error: ErrInvalidOp if operation is not supported
	//
	// Implementation requirements:
	//   - Validate all required parameters are present
	//   - Sanitize inputs to prevent command injection
	//   - Return ErrInvalidOp for unsupported operations
	//   - Document supported operations and their parameters
	//
	// Example:
	//
	//	func (g *GitProvider) ExecArgsFor(op string, params map[string]any) ([]string, error) {
	//	    switch op {
	//	    case "git.clone":
	//	        url, ok := params["url"].(string)
	//	        if !ok {
	//	            return nil, fmt.Errorf("url parameter required")
	//	        }
	//	        args := []string{"clone"}
	//	        if depth, ok := params["depth"].(int); ok && depth > 0 {
	//	            args = append(args, "--depth", strconv.Itoa(depth))
	//	        }
	//	        args = append(args, url)
	//	        return args, nil
	//	    default:
	//	        return nil, provider.ErrInvalidOp("git", op)
	//	    }
	//	}
	ExecArgsFor(op string, params map[string]any) ([]string, error)

	// WorkDir returns the working directory for command execution
	//
	// This is the directory where the CLI command will be executed.
	// It can be used to:
	//   - Execute commands in a specific repository (e.g., git operations)
	//   - Control where output files are created
	//   - Provide context for relative path arguments
	//
	// Returns:
	//   - string: absolute path to working directory
	//   - empty string: use the current process's working directory
	//
	// The working directory should be validated to exist before execution.
	//
	// Example:
	//
	//	func (g *GitProvider) WorkDir() string {
	//	    return g.repoPath // e.g., "/home/user/repo"
	//	}
	WorkDir() string

	// Env returns environment variables for command execution
	//
	// These environment variables are added to the minimal base environment
	// used by the CLI executor. Use sparingly for security.
	//
	// Common use cases:
	//   - Authentication: GIT_ASKPASS, GITHUB_TOKEN
	//   - Configuration: GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL
	//   - Behavior control: GIT_TERMINAL_PROMPT=0, TF_IN_AUTOMATION=1
	//
	// Format: Each string should be "KEY=VALUE"
	//
	// Security considerations:
	//   - Only pass required environment variables
	//   - Never pass credentials directly (use credential helpers)
	//   - Avoid passing parent process environment wholesale
	//
	// Returns:
	//   - []string: environment variables in "KEY=VALUE" format
	//   - nil or empty slice: no additional environment variables
	//
	// Example:
	//
	//	func (g *GitProvider) Env() []string {
	//	    return []string{
	//	        "GIT_TERMINAL_PROMPT=0",
	//	        "GIT_AUTHOR_NAME=Zen CLI",
	//	        "GIT_AUTHOR_EMAIL=zen@example.com",
	//	    }
	//	}
	Env() []string
}
