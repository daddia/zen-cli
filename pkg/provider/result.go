package provider

import (
	"time"
)

// Result represents the unified result of a provider operation
//
// Result provides a transport-agnostic structure that can hold output from both
// CLI commands and API requests. The relevant fields are populated based on the
// provider kind:
//   - CLI providers: ExitCode, Stdout, Stderr, Duration
//   - API providers: ExitCode (HTTP status), Body, Headers, Meta, Duration
type Result struct {
	// ExitCode is the result status code
	// For CLI providers: the process exit code (0 = success, non-zero = error)
	// For API providers: the HTTP status code (200 = success, 4xx/5xx = error)
	ExitCode int

	// Stdout contains the standard output from CLI commands (CLI providers only)
	// Empty for API providers
	Stdout []byte

	// Stderr contains the standard error output from CLI commands (CLI providers only)
	// Empty for API providers
	Stderr []byte

	// Body contains the response body from API requests (API providers only)
	// This is the raw response body (JSON, XML, etc.) that can be parsed by the caller
	// Empty for CLI providers
	Body []byte

	// Headers contains HTTP response headers (API providers only)
	// Useful for extracting rate limit info, pagination links, etc.
	// Empty for CLI providers
	Headers map[string]string

	// Meta contains additional provider-specific metadata
	// Examples:
	//   - Rate limit information: {"rate_limit_remaining": "4999"}
	//   - Pagination info: {"next_page": "2", "total_pages": "10"}
	//   - Git commit hash: {"commit": "a1b2c3d"}
	// This field is optional and provider-specific
	Meta map[string]any

	// Duration is the time taken to execute the operation
	// Useful for performance monitoring and debugging
	Duration time.Duration
}

// Success returns true if the operation completed successfully
//
// For CLI providers: exit code is 0
// For API providers: HTTP status code is 2xx (200-299)
func (r Result) Success() bool {
	// For API providers: HTTP status codes 200-299 indicate success
	if r.ExitCode >= 200 && r.ExitCode < 300 {
		return true
	}
	// For CLI providers: only exit code 0 indicates success
	return r.ExitCode == 0
}

// IsClientError returns true if the error is a client error (4xx for APIs)
//
// Only relevant for API providers. Returns false for CLI providers.
func (r Result) IsClientError() bool {
	return r.ExitCode >= 400 && r.ExitCode < 500
}

// IsServerError returns true if the error is a server error (5xx for APIs)
//
// Only relevant for API providers. Returns false for CLI providers.
func (r Result) IsServerError() bool {
	return r.ExitCode >= 500 && r.ExitCode < 600
}

// Output returns the primary output content based on provider kind
//
// For CLI providers: returns Stdout
// For API providers: returns Body
//
// This is a convenience method for getting the main output without
// needing to check the provider kind.
func (r Result) Output() []byte {
	if len(r.Body) > 0 {
		return r.Body
	}
	return r.Stdout
}
