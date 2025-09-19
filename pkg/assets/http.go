package assets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/errors"
)

// HTTPManifestClient handles HTTP-based manifest downloading
type HTTPManifestClient struct {
	httpClient   *http.Client
	logger       logging.Logger
	auth         AuthProvider
	authProvider string
}

// NewHTTPManifestClient creates a new HTTP manifest client
func NewHTTPManifestClient(logger logging.Logger, auth AuthProvider, authProvider string) *HTTPManifestClient {
	return &HTTPManifestClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       logger,
		auth:         auth,
		authProvider: authProvider,
	}
}

// DownloadManifest downloads the manifest.yaml file from the repository using HTTP API
func (h *HTTPManifestClient) DownloadManifest(ctx context.Context, repoURL, branch string) ([]byte, error) {
	h.logger.Debug("downloading manifest via HTTP API", "repo", h.sanitizeURL(repoURL), "branch", branch)

	// Parse repository URL to determine provider and construct API URL
	apiURL, err := h.buildAPIURL(repoURL, branch, "manifest.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build API URL")
	}

	h.logger.Debug("constructed API URL", "url", apiURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	// Add authentication headers
	if err := h.addAuthHeaders(req); err != nil {
		return nil, errors.Wrap(err, "failed to add authentication headers")
	}

	// Set appropriate headers
	req.Header.Set("Accept", "application/vnd.github.v3.raw") // GitHub raw content
	req.Header.Set("User-Agent", "zen-cli/1.0")

	// Make the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, &AssetClientError{
			Code:    ErrorCodeNetworkError,
			Message: fmt.Sprintf("HTTP request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == 404 {
		return nil, &AssetClientError{
			Code:    ErrorCodeAssetNotFound,
			Message: "manifest.yaml not found in repository",
		}
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, &AssetClientError{
			Code:    ErrorCodeAuthenticationFailed,
			Message: "authentication failed or insufficient permissions",
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &AssetClientError{
			Code:    ErrorCodeRepositoryError,
			Message: fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode),
		}
	}

	// Read response body
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	h.logger.Debug("manifest downloaded successfully", "size", len(content))
	return content, nil
}

// buildAPIURL constructs the appropriate API URL for different Git providers
func (h *HTTPManifestClient) buildAPIURL(repoURL, branch, filePath string) (string, error) {
	// Parse the repository URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", errors.Wrap(err, "invalid repository URL")
	}

	// Remove .git suffix from path but keep leading slash for API
	repoPath := strings.TrimSuffix(parsedURL.Path, ".git")

	switch parsedURL.Host {
	case "github.com":
		// GitHub API: https://api.github.com/repos/owner/repo/contents/path?ref=branch
		return fmt.Sprintf("https://api.github.com/repos%s/contents/%s?ref=%s",
			repoPath, filePath, branch), nil

	case "gitlab.com":
		// GitLab API: https://gitlab.com/api/v4/projects/owner%2Frepo/repository/files/path/raw?ref=branch
		encodedPath := url.QueryEscape(strings.TrimPrefix(repoPath, "/"))
		encodedFile := url.QueryEscape(filePath)
		return fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/repository/files/%s/raw?ref=%s",
			encodedPath, encodedFile, branch), nil

	default:
		// For other Git hosts, try a generic approach (may not work for all)
		return "", fmt.Errorf("unsupported Git provider: %s", parsedURL.Host)
	}
}

// addAuthHeaders adds authentication headers to the request
func (h *HTTPManifestClient) addAuthHeaders(req *http.Request) error {
	if h.auth == nil {
		h.logger.Debug("no auth provider configured, using anonymous access")
		return nil // No authentication configured
	}

	// Get credentials from auth provider
	token, err := h.auth.GetCredentials(h.authProvider)
	if err != nil {
		h.logger.Debug("failed to get credentials, using anonymous access", "error", err)
		return nil // Continue without authentication for public repositories
	}

	if token == "" {
		h.logger.Debug("no token available, using anonymous access")
		return nil // No token available - ok for public repositories
	}

	// Add appropriate auth header based on provider
	switch h.authProvider {
	case "github":
		req.Header.Set("Authorization", "Bearer "+token)
		h.logger.Debug("added GitHub authentication header")
	case "gitlab":
		req.Header.Set("Private-Token", token)
		h.logger.Debug("added GitLab authentication header")
	default:
		req.Header.Set("Authorization", "Bearer "+token)
		h.logger.Debug("added generic authentication header")
	}

	return nil
}

// sanitizeURL removes sensitive information from URLs for logging
func (h *HTTPManifestClient) sanitizeURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid-url]"
	}

	// Remove user info (tokens, passwords)
	parsedURL.User = nil
	return parsedURL.String()
}

// GitHubResponse represents the GitHub API response for file contents
type GitHubResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

// downloadFromGitHub downloads content using GitHub's API with proper content handling
func (h *HTTPManifestClient) downloadFromGitHub(ctx context.Context, apiURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	// Add authentication headers
	if err := h.addAuthHeaders(req); err != nil {
		return nil, errors.Wrap(err, "failed to add authentication headers")
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "zen-cli/1.0")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API request failed with status %d", resp.StatusCode)
	}

	var githubResp GitHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&githubResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode GitHub API response")
	}

	// For small files, content is base64 encoded in the response
	if githubResp.Encoding == "base64" && githubResp.Content != "" {
		// Decode base64 content
		content, err := h.decodeBase64Content(githubResp.Content)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode base64 content")
		}
		return content, nil
	}

	// For larger files, use the download URL
	if githubResp.DownloadURL != "" {
		return h.downloadFromURL(ctx, githubResp.DownloadURL)
	}

	return nil, fmt.Errorf("unable to download content from GitHub")
}

// decodeBase64Content decodes base64 content from GitHub API
func (h *HTTPManifestClient) decodeBase64Content(content string) ([]byte, error) {
	// Remove whitespace and newlines from base64 content
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")

	// GitHub uses standard base64 encoding
	return base64.StdEncoding.DecodeString(content)
}

// downloadFromURL downloads content from a direct URL
func (h *HTTPManifestClient) downloadFromURL(ctx context.Context, downloadURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create download request")
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "download request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
