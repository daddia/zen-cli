package plugin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
)

// HostAPIImpl implements the HostAPI interface
type HostAPIImpl struct {
	logger     logging.Logger
	auth       auth.Manager
	httpClient *http.Client
}

// NewHostAPI creates a new host API implementation
func NewHostAPI(logger logging.Logger, auth auth.Manager) HostAPI {
	return &HostAPIImpl{
		logger: logger,
		auth:   auth,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 5,
			},
		},
	}
}

// HTTPRequest performs an HTTP request on behalf of a plugin
func (h *HostAPIImpl) HTTPRequest(method, url string, headers map[string]string, body []byte) ([]byte, error) {
	h.logger.Debug("plugin HTTP request", "method", method, "url", h.sanitizeURL(url))

	// Validate URL and method
	if err := h.validateHTTPRequest(method, url); err != nil {
		return nil, &PluginError{
			Code:      ErrCodeSecurityViolation,
			Message:   fmt.Sprintf("HTTP request validation failed: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Create request
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, &PluginError{
			Code:      ErrCodeInvalidArguments,
			Message:   fmt.Sprintf("failed to create HTTP request: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default headers
	req.Header.Set("User-Agent", "Zen-CLI-Plugin/1.0")

	// Execute request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, &PluginError{
			Code:      ErrCodeRuntimeError,
			Message:   fmt.Sprintf("HTTP request failed: %v", err),
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &PluginError{
			Code:      ErrCodeRuntimeError,
			Message:   fmt.Sprintf("failed to read response body: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, &PluginError{
			Code:    ErrCodeRuntimeError,
			Message: fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode),
			Details: map[string]interface{}{
				"status_code": resp.StatusCode,
				"response":    string(responseBody),
			},
			Timestamp: time.Now(),
		}
	}

	h.logger.Debug("plugin HTTP request completed",
		"method", method,
		"url", h.sanitizeURL(url),
		"status", resp.StatusCode,
		"response_size", len(responseBody))

	return responseBody, nil
}

// GetConfig retrieves a configuration value
func (h *HostAPIImpl) GetConfig(key string) (string, error) {
	h.logger.Debug("plugin config access", "key", key)

	// TODO: Implement configuration access
	// This would integrate with the existing config system
	// For now, return empty value

	return "", nil
}

// GetCredentials retrieves credentials for a provider
func (h *HostAPIImpl) GetCredentials(credentialRef string) (string, error) {
	h.logger.Debug("plugin credential access", "credential_ref", credentialRef)

	// Use the existing auth manager to get credentials
	credentials, err := h.auth.GetCredentials(credentialRef)
	if err != nil {
		return "", &PluginError{
			Code:      ErrCodeRuntimeError,
			Message:   fmt.Sprintf("failed to get credentials: %v", err),
			Timestamp: time.Now(),
		}
	}

	return credentials, nil
}

// Log writes a log message
func (h *HostAPIImpl) Log(level string, message string) error {
	switch level {
	case "info":
		h.logger.Info("plugin log", "message", message)
	case "warn":
		h.logger.Warn("plugin log", "message", message)
	case "error":
		h.logger.Error("plugin log", "message", message)
	case "debug":
		h.logger.Debug("plugin log", "message", message)
	default:
		h.logger.Info("plugin log", "level", level, "message", message)
	}

	return nil
}

// GetTask retrieves task data (if permitted)
func (h *HostAPIImpl) GetTask(taskID string) ([]byte, error) {
	h.logger.Debug("plugin task access", "task_id", taskID)

	// TODO: Implement task data access
	// This would integrate with the task management system
	// For now, return placeholder data

	return []byte(`{"id":"` + taskID + `","title":"Sample Task"}`), nil
}

// UpdateTask updates task data (if permitted)
func (h *HostAPIImpl) UpdateTask(taskID string, data []byte) error {
	h.logger.Debug("plugin task update", "task_id", taskID, "data_size", len(data))

	// TODO: Implement task update
	// This would integrate with the task management system

	return nil
}

// ValidateConfig validates plugin configuration
func (h *HostAPIImpl) ValidateConfig(configJSON string, schemaName string) error {
	h.logger.Debug("plugin config validation", "schema", schemaName)

	// TODO: Implement configuration validation
	// This would use JSON schema validation

	return nil
}

// validateHTTPRequest validates an HTTP request for security
func (h *HostAPIImpl) validateHTTPRequest(method, url string) error {
	// Validate HTTP method
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
		"HEAD":   true,
	}

	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}

	// Validate URL scheme
	if !h.isAllowedURL(url) {
		return fmt.Errorf("URL not allowed: %s", h.sanitizeURL(url))
	}

	return nil
}

// isAllowedURL checks if a URL is allowed for plugin access
func (h *HostAPIImpl) isAllowedURL(url string) bool {
	// TODO: Implement URL allowlist based on plugin permissions
	// For now, allow HTTPS URLs only
	return len(url) > 8 && url[:8] == "https://"
}

// sanitizeURL removes sensitive information from URLs for logging
func (h *HostAPIImpl) sanitizeURL(url string) string {
	// Remove query parameters and fragments that might contain sensitive data
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx] + "?..."
	}
	if idx := strings.Index(url, "#"); idx != -1 {
		url = url[:idx] + "#..."
	}
	return url
}
