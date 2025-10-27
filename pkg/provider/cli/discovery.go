package cli

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/provider"
)

// DiscoveryCache provides thread-safe caching of binary discovery results
//
// Caching is important for performance since binary discovery can be expensive:
//   - PATH lookups require filesystem operations
//   - Version detection requires executing the binary
//   - These operations are repeated frequently during provider initialization
//
// The cache uses a time-based TTL to ensure freshness while avoiding
// excessive filesystem operations.
type DiscoveryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	path      string
	version   string
	timestamp time.Time
	err       error
}

// NewDiscoveryCache creates a new discovery cache with the specified TTL
//
// A typical TTL is 5-10 minutes, balancing freshness with performance.
// Set TTL to 0 to disable caching (not recommended for production).
func NewDiscoveryCache(ttl time.Duration) *DiscoveryCache {
	return &DiscoveryCache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached discovery result if available and not expired
func (c *DiscoveryCache) Get(name string) (path, version string, err error, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[name]
	if !exists {
		return "", "", nil, false
	}

	// Check if entry has expired
	if c.ttl > 0 && time.Since(entry.timestamp) > c.ttl {
		return "", "", nil, false
	}

	return entry.path, entry.version, entry.err, true
}

// Set stores a discovery result in the cache
func (c *DiscoveryCache) Set(name, path, version string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[name] = &cacheEntry{
		path:      path,
		version:   version,
		timestamp: time.Now(),
		err:       err,
	}
}

// Clear removes all entries from the cache
func (c *DiscoveryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// DiscoveryResult contains the result of binary discovery
type DiscoveryResult struct {
	// Path is the absolute path to the binary
	Path string

	// Version is the detected version string
	Version string

	// Available indicates whether the binary is available and meets requirements
	Available bool

	// Reason provides explanation when Available is false
	Reason string
}

// Discovery provides binary discovery and version detection
type Discovery struct {
	cache  *DiscoveryCache
	logger logging.Logger
}

// NewDiscovery creates a new binary discovery instance
func NewDiscovery(cache *DiscoveryCache, logger logging.Logger) *Discovery {
	if cache == nil {
		// Default cache with 5 minute TTL
		cache = NewDiscoveryCache(5 * time.Minute)
	}
	return &Discovery{
		cache:  cache,
		logger: logger,
	}
}

// FindBinary locates a binary in the system PATH
//
// Returns:
//   - Absolute path to the binary
//   - error if the binary is not found
//
// The result is cached to avoid repeated filesystem lookups.
func (d *Discovery) FindBinary(name string) (string, error) {
	d.logger.Debug("finding binary", "name", name)

	// Check cache first
	if path, _, cachedErr, found := d.cache.Get(name); found {
		d.logger.Debug("binary discovery cache hit", "name", name, "path", path)
		return path, cachedErr
	}

	// Look up binary in PATH
	path, err := exec.LookPath(name)
	if err != nil {
		d.logger.Debug("binary not found in PATH", "name", name, "error", err)
		notFoundErr := provider.ErrNotFound(name, "binary")

		// Cache the error to avoid repeated lookups
		d.cache.Set(name, "", "", notFoundErr)
		return "", notFoundErr
	}

	d.logger.Debug("binary found", "name", name, "path", path)

	// Cache the successful result
	d.cache.Set(name, path, "", nil)
	return path, nil
}

// GetVersion executes the binary with version arguments and parses the version
//
// Parameters:
//   - ctx: context for timeout/cancellation
//   - path: absolute path to the binary
//   - versionArgs: arguments to get version (e.g., ["--version"], ["-v"])
//
// Returns:
//   - Detected version string (e.g., "2.39.0")
//   - error if version detection fails
//
// The version string is extracted using common patterns. If multiple patterns
// match, the first match is returned. Custom version parsing can be implemented
// by calling ParseVersion directly.
func (d *Discovery) GetVersion(ctx context.Context, path string, versionArgs []string) (string, error) {
	d.logger.Debug("getting version", "binary", path, "args", versionArgs)

	// Execute the binary with version arguments
	cmd := exec.CommandContext(ctx, path, versionArgs...)

	// Set timeout for version detection (should be fast)
	if ctx.Err() == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, path, versionArgs...)
	}

	// Capture combined output (version might be on stdout or stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.logger.Debug("failed to get version",
			"binary", path,
			"args", versionArgs,
			"error", err,
			"output", string(output))
		return "", provider.WrapError(
			provider.ErrorCodeExecution,
			"failed to execute version command",
			path,
			err,
		)
	}

	// Parse version from output
	version, err := ParseVersion(string(output))
	if err != nil {
		d.logger.Debug("failed to parse version",
			"binary", path,
			"output", string(output),
			"error", err)
		return "", provider.WrapError(
			provider.ErrorCodeParse,
			"failed to parse version from output",
			path,
			err,
		)
	}

	d.logger.Debug("version detected", "binary", path, "version", version)
	return version, nil
}

// ParseVersion extracts a semantic version string from command output
//
// Supports common version formats:
//   - Semantic versioning: "1.2.3", "2.39.0", "v3.1.0"
//   - Git-style: "git version 2.39.0"
//   - Node-style: "v18.12.0"
//   - Simple: "1.2", "2.0"
//
// Returns the first version string found in the output.
// Returns error if no version pattern is detected.
func ParseVersion(output string) (string, error) {
	// Common version patterns (in order of specificity)
	patterns := []*regexp.Regexp{
		// Semantic version with optional 'v' prefix: v1.2.3, 1.2.3-beta
		regexp.MustCompile(`v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?)`),

		// Two-part version with optional 'v' prefix: v1.2, 1.2
		regexp.MustCompile(`v?(\d+\.\d+)`),

		// Single version number (less common, but supported)
		regexp.MustCompile(`v?(\d+)`),
	}

	output = strings.TrimSpace(output)

	// Try each pattern in order
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(output)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("no version pattern found in output: %s", output)
}

// ValidateVersion checks if a version meets a minimum version requirement
//
// Supports simple version comparison for semantic versions.
// Both current and required should be in the format "X.Y.Z" or "X.Y".
//
// Returns:
//   - true if current >= required
//   - false otherwise
//
// For more complex version constraints (e.g., ">=1.2.0,<2.0.0"), use a
// dedicated version constraint library like github.com/hashicorp/go-version.
func ValidateVersion(current, required string) (bool, error) {
	// Normalize versions (remove 'v' prefix if present)
	current = strings.TrimPrefix(current, "v")
	required = strings.TrimPrefix(required, "v")

	// Parse current version
	currentParts, err := parseVersionParts(current)
	if err != nil {
		return false, fmt.Errorf("invalid current version %q: %w", current, err)
	}

	// Parse required version
	requiredParts, err := parseVersionParts(required)
	if err != nil {
		return false, fmt.Errorf("invalid required version %q: %w", required, err)
	}

	// Compare versions part by part (major, minor, patch)
	for i := 0; i < len(requiredParts); i++ {
		// If current version has fewer parts, pad with zeros
		if i >= len(currentParts) {
			return false, nil
		}

		if currentParts[i] < requiredParts[i] {
			return false, nil
		}
		if currentParts[i] > requiredParts[i] {
			return true, nil
		}
		// If equal, continue to next part
	}

	// Versions are equal (or current has more parts)
	return true, nil
}

// parseVersionParts splits a version string into numeric parts
func parseVersionParts(version string) ([]int, error) {
	// Remove any pre-release or build metadata (e.g., "1.2.3-beta" -> "1.2.3")
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid version part %q: %w", part, err)
		}
		result = append(result, num)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("version has no numeric parts")
	}

	return result, nil
}

// Discover performs complete binary discovery including version detection
//
// This is a convenience method that combines FindBinary and GetVersion
// with version validation.
//
// Parameters:
//   - ctx: context for timeout/cancellation
//   - name: binary name to find
//   - versionArgs: arguments to get version (e.g., ["--version"])
//   - minVersion: minimum required version (empty = no requirement)
//
// Returns:
//   - DiscoveryResult with path, version, and availability status
//   - error only if an unexpected error occurs (not for "not found")
//
// The DiscoveryResult.Available field indicates whether the binary meets
// all requirements. The Reason field explains why if not available.
func (d *Discovery) Discover(ctx context.Context, name string, versionArgs []string, minVersion string) (DiscoveryResult, error) {
	result := DiscoveryResult{
		Available: false,
	}

	// Find binary in PATH
	path, err := d.FindBinary(name)
	if err != nil {
		result.Reason = "binary not found in PATH"
		return result, nil // Not an error, just not available
	}

	result.Path = path

	// Get version if version args provided
	if len(versionArgs) > 0 {
		version, err := d.GetVersion(ctx, path, versionArgs)
		if err != nil {
			result.Reason = fmt.Sprintf("failed to detect version: %v", err)
			return result, nil // Not an error, just not available
		}

		result.Version = version

		// Validate version if minimum required
		if minVersion != "" {
			valid, err := ValidateVersion(version, minVersion)
			if err != nil {
				result.Reason = fmt.Sprintf("failed to validate version: %v", err)
				return result, nil
			}

			if !valid {
				result.Reason = fmt.Sprintf("version %s does not meet requirement %s", version, minVersion)
				return result, nil
			}
		}
	}

	// Binary is available and meets all requirements
	result.Available = true
	return result, nil
}
