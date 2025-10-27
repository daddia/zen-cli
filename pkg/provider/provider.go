package provider

import (
	"context"
	"io"
)

// Kind represents the type of provider implementation
type Kind string

const (
	// KindCLI represents a provider that wraps external CLI tools
	KindCLI Kind = "cli"

	// KindAPI represents a provider that wraps REST/GraphQL APIs
	KindAPI Kind = "api"
)

// Info contains metadata and availability information about a provider
type Info struct {
	// Name is the unique identifier for the provider (e.g., "git", "github", "jira")
	Name string

	// Kind indicates whether this is a CLI or API provider
	Kind Kind

	// Version is the detected version of the tool or API
	// For CLI providers: the binary version (e.g., "2.39.0")
	// For API providers: the API version or server version
	Version string

	// Available indicates whether the provider is ready for use
	// For CLI providers: binary exists and meets minimum version requirements
	// For API providers: credentials are configured and endpoint is reachable
	Available bool

	// Reason provides human-readable explanation when Available is false
	// Examples: "binary not found in PATH", "missing API credentials", "version 2.35+ required"
	Reason string

	// Capabilities is a map of supported operations for this provider
	// Key is the operation name (e.g., "git.clone", "pull_request.list")
	// Value indicates whether the operation is currently available
	Capabilities map[string]bool

	// BinaryPath is the absolute path to the CLI binary (CLI providers only)
	// Empty for API providers
	BinaryPath string

	// BaseURL is the API endpoint URL (API providers only)
	// Empty for CLI providers
	BaseURL string
}

// Provider is the unified interface for both CLI and API providers
//
// Providers enable Zen CLI to integrate with external tools and services through
// a transport-agnostic interface. Each provider wrapper lives in its own repository
// (e.g., zen-git, zen-jira) and is compiled into Zen CLI as a Go module dependency.
//
// Provider implementations must be safe for concurrent use by multiple goroutines.
type Provider interface {
	// Info returns metadata and availability information about the provider
	//
	// This method should be lightweight and fast (< 100ms) as it may be called
	// frequently for provider discovery and status checks.
	//
	// The context can be used for cancellation and timeouts.
	Info(ctx context.Context) (Info, error)

	// Execute runs a named operation with the given parameters
	//
	// Operations are provider-specific strings that map to CLI commands or API endpoints:
	//   - CLI providers: "git.clone", "git.commit", "git.push"
	//   - API providers: "pull_request.list", "issue.create", "task.update"
	//
	// Parameters are passed as a map to support flexible, operation-specific arguments:
	//   - CLI providers: map["url"]="...", map["directory"]="..."
	//   - API providers: map["owner"]="...", map["repo"]="...", map["state"]="open"
	//
	// The Result contains the complete output (CLI stdout/stderr or API response body).
	//
	// Returns ErrInvalidOp if the operation is not supported by this provider.
	// Returns ErrExecution if the operation fails during execution.
	// Returns ErrTimeout if the context deadline is exceeded.
	Execute(ctx context.Context, op string, params map[string]any) (Result, error)

	// Stream executes an operation and returns a streaming reader for the output
	//
	// This is useful for long-running operations where you want to process output
	// incrementally rather than buffering it all in memory:
	//   - CLI providers: stream stdout from long-running commands
	//   - API providers: stream large API responses or webhook data
	//
	// The caller is responsible for closing the returned ReadCloser.
	//
	// Returns ErrInvalidOp if the operation doesn't support streaming.
	// Returns ErrExecution if the operation fails to start.
	Stream(ctx context.Context, op string, params map[string]any) (io.ReadCloser, error)
}
