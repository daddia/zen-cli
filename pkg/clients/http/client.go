package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/clients"
	"golang.org/x/time/rate"
)

// Client provides a shared HTTP client with common functionality
type Client struct {
	httpClient     *http.Client
	logger         logging.Logger
	rateLimiter    *rate.Limiter
	baseURL        string
	defaultHeaders map[string]string
	timeout        time.Duration
	retryConfig    RetryConfig
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	RetryableStatus []int
}

// NewClient creates a new HTTP client
func NewClient(config clients.HTTPConfig, logger logging.Logger) *Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		MaxIdleConnsPerHost: 5,
		ForceAttemptHTTP2:   true,
	}

	httpClient := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	// Create rate limiter
	var rateLimiter *rate.Limiter
	if config.RateLimits.RequestsPerMinute > 0 {
		limit := rate.Limit(float64(config.RateLimits.RequestsPerMinute) / 60.0) // Convert to per-second
		burst := config.RateLimits.BurstSize
		if burst == 0 {
			burst = 10 // Default burst size
		}
		rateLimiter = rate.NewLimiter(limit, burst)
	}

	return &Client{
		httpClient:     httpClient,
		logger:         logger,
		rateLimiter:    rateLimiter,
		baseURL:        config.BaseURL,
		defaultHeaders: config.Headers,
		timeout:        config.Timeout,
		retryConfig: RetryConfig{
			MaxRetries:      config.Retries,
			BaseDelay:       1 * time.Second,
			MaxDelay:        30 * time.Second,
			RetryableStatus: []int{429, 500, 502, 503, 504},
		},
	}
}

// Request represents an HTTP request
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	Body        []byte
	QueryParams map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	Duration   time.Duration
}

// Do executes an HTTP request with retry logic and rate limiting
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// Check rate limiting
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, &clients.ClientError{
				Code:      clients.ErrorCodeRateLimited,
				Message:   "rate limit exceeded",
				Retryable: true,
			}
		}
	}

	// Build full URL
	fullURL, err := c.buildURL(req.URL, req.QueryParams)
	if err != nil {
		return nil, &clients.ClientError{
			Code:    clients.ErrorCodeInvalidRequest,
			Message: fmt.Sprintf("invalid URL: %v", err),
		}
	}

	// Execute with retry logic
	return c.executeWithRetry(ctx, req.Method, fullURL, req.Headers, req.Body)
}

// executeWithRetry executes an HTTP request with exponential backoff retry
func (c *Client) executeWithRetry(ctx context.Context, method, url string, headers map[string]string, body []byte) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.executeRequest(ctx, method, url, headers, body)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryable(err, resp) {
			return resp, err
		}

		// Don't wait after the last attempt
		if attempt == c.retryConfig.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := c.calculateBackoff(attempt)

		c.logger.Debug("retrying HTTP request",
			"attempt", attempt+1,
			"delay", delay,
			"error", err)

		// Wait before retry
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}

	return nil, lastErr
}

// executeRequest executes a single HTTP request
func (c *Client) executeRequest(ctx context.Context, method, url string, headers map[string]string, body []byte) (*Response, error) {
	start := time.Now()

	// Create request
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, &clients.ClientError{
			Code:    clients.ErrorCodeInvalidRequest,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	// Set default headers
	for key, value := range c.defaultHeaders {
		req.Header.Set(key, value)
	}

	// Set request-specific headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return nil, &clients.ClientError{
			Code:      clients.ErrorCodeConnectionFailed,
			Message:   fmt.Sprintf("HTTP request failed: %v", err),
			Retryable: true,
		}
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &clients.ClientError{
			Code:    clients.ErrorCodeInternalError,
			Message: fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
		Duration:   duration,
	}

	// Log request details
	c.logger.Debug("HTTP request completed",
		"method", method,
		"url", c.sanitizeURL(url),
		"status", resp.StatusCode,
		"duration", duration,
		"response_size", len(respBody))

	return response, nil
}

// buildURL builds a full URL with query parameters
func (c *Client) buildURL(path string, queryParams map[string]string) (string, error) {
	var fullURL string

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		fullURL = path
	} else {
		if c.baseURL == "" {
			return "", fmt.Errorf("no base URL configured")
		}
		fullURL = strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	}

	// Add query parameters
	if len(queryParams) > 0 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}

		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	return fullURL, nil
}

// isRetryable determines if an error or response is retryable
func (c *Client) isRetryable(err error, resp *Response) bool {
	// Check if it's a retryable client error
	if clientErr, ok := err.(*clients.ClientError); ok {
		return clientErr.Retryable
	}

	// Check response status codes
	if resp != nil {
		for _, status := range c.retryConfig.RetryableStatus {
			if resp.StatusCode == status {
				return true
			}
		}
	}

	return false
}

// calculateBackoff calculates the backoff delay for retry attempts
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with jitter
	delay := c.retryConfig.BaseDelay * time.Duration(1<<attempt)

	// Cap at max delay
	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}

	// Add jitter (±10%)
	jitter := time.Duration(float64(delay) * 0.1)
	jitterMultiplier := 2*time.Now().UnixNano()%2 - 1
	delay += time.Duration(float64(jitter) * float64(jitterMultiplier))

	return delay
}

// sanitizeURL removes sensitive information from URLs for logging
func (c *Client) sanitizeURL(rawURL string) string {
	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove user info
	u.User = nil

	// Remove query parameters that might contain sensitive data
	if u.RawQuery != "" {
		u.RawQuery = "..."
	}

	return u.String()
}

// Close cleans up the HTTP client resources
func (c *Client) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}

// SetDefaultHeader sets a default header for all requests
func (c *Client) SetDefaultHeader(key, value string) {
	if c.defaultHeaders == nil {
		c.defaultHeaders = make(map[string]string)
	}
	c.defaultHeaders[key] = value
}

// SetRateLimit updates the rate limiting configuration
func (c *Client) SetRateLimit(requestsPerMinute int, burstSize int) {
	if requestsPerMinute > 0 {
		limit := rate.Limit(float64(requestsPerMinute) / 60.0)
		if burstSize == 0 {
			burstSize = 10
		}
		c.rateLimiter = rate.NewLimiter(limit, burstSize)
	} else {
		c.rateLimiter = nil
	}
}
