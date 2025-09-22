package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
)

// ClientPool manages HTTP connections with connection pooling and middleware
type ClientPool struct {
	logger     logging.Logger
	clients    map[string]*http.Client
	middleware []MiddlewareFunc
	config     *ClientPoolConfig
	metrics    *ClientPoolMetrics
	mu         sync.RWMutex
}

// HTTPClientInterface defines the HTTP client interface
type HTTPClientInterface interface {
	// Do executes an HTTP request
	Do(req *http.Request) (*http.Response, error)

	// Get performs a GET request
	Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error)

	// Post performs a POST request
	Post(ctx context.Context, url string, headers map[string]string, body []byte) (*http.Response, error)

	// Put performs a PUT request
	Put(ctx context.Context, url string, headers map[string]string, body []byte) (*http.Response, error)

	// Delete performs a DELETE request
	Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error)

	// Close closes the client and cleans up resources
	Close() error
}

// ConnectionPoolInterface defines the connection pool interface
type ConnectionPoolInterface interface {
	// GetClient returns an HTTP client for the given key
	GetClient(key string) (*http.Client, error)

	// CreateClient creates a new HTTP client with the given configuration
	CreateClient(key string, config *ClientConfig) (*http.Client, error)

	// RemoveClient removes a client from the pool
	RemoveClient(key string) error

	// GetPoolMetrics returns connection pool metrics
	GetPoolMetrics() *ClientPoolMetrics

	// CloseAll closes all clients in the pool
	CloseAll() error
}

// MiddlewareInterface defines HTTP middleware interface
type MiddlewareInterface interface {
	// Handle processes an HTTP request/response
	Handle(req *http.Request, next http.RoundTripper) (*http.Response, error)

	// Name returns the middleware name
	Name() string
}

// ClientPoolConfig contains HTTP client pool configuration
type ClientPoolConfig struct {
	MaxIdleConns        int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost     int           `json:"max_conns_per_host" yaml:"max_conns_per_host"`
	IdleConnTimeout     time.Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout time.Duration `json:"tls_handshake_timeout" yaml:"tls_handshake_timeout"`
	DialTimeout         time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	KeepAlive           time.Duration `json:"keep_alive" yaml:"keep_alive"`
	DisableCompression  bool          `json:"disable_compression" yaml:"disable_compression"`
	InsecureSkipVerify  bool          `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// ClientConfig contains configuration for individual HTTP clients
type ClientConfig struct {
	Timeout         time.Duration     `json:"timeout" yaml:"timeout"`
	Headers         map[string]string `json:"headers" yaml:"headers"`
	UserAgent       string            `json:"user_agent" yaml:"user_agent"`
	FollowRedirects bool              `json:"follow_redirects" yaml:"follow_redirects"`
	MaxRedirects    int               `json:"max_redirects" yaml:"max_redirects"`
}

// ClientPoolMetrics contains HTTP client pool performance metrics
type ClientPoolMetrics struct {
	TotalClients    int           `json:"total_clients"`
	ActiveConns     int           `json:"active_connections"`
	IdleConns       int           `json:"idle_connections"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessfulReqs  int64         `json:"successful_requests"`
	FailedReqs      int64         `json:"failed_requests"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastRequestTime time.Time     `json:"last_request_time"`
	mu              sync.RWMutex
}

// MiddlewareFunc represents an HTTP middleware function
type MiddlewareFunc func(req *http.Request, next http.RoundTripper) (*http.Response, error)

// RoundTripperFunc is an adapter to use functions as http.RoundTripper
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewClientPool creates a new HTTP client pool
func NewClientPool(logger logging.Logger, config *ClientPoolConfig) *ClientPool {
	if config == nil {
		config = DefaultClientPoolConfig()
	}

	return &ClientPool{
		logger:     logger,
		clients:    make(map[string]*http.Client),
		middleware: make([]MiddlewareFunc, 0),
		config:     config,
		metrics:    &ClientPoolMetrics{},
	}
}

// DefaultClientPoolConfig returns default client pool configuration
func DefaultClientPoolConfig() *ClientPoolConfig {
	return &ClientPoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DialTimeout:         30 * time.Second,
		KeepAlive:           30 * time.Second,
		DisableCompression:  false,
		InsecureSkipVerify:  false,
	}
}

// GetClient returns an HTTP client for the given key
func (cp *ClientPool) GetClient(key string) (*http.Client, error) {
	cp.mu.RLock()
	client, exists := cp.clients[key]
	cp.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Create new client with default configuration
	return cp.CreateClient(key, &ClientConfig{
		Timeout:         30 * time.Second,
		FollowRedirects: true,
		MaxRedirects:    10,
	})
}

// CreateClient creates a new HTTP client with the given configuration
func (cp *ClientPool) CreateClient(key string, config *ClientConfig) (*http.Client, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Check if client already exists
	if _, exists := cp.clients[key]; exists {
		return nil, fmt.Errorf("client already exists for key: %s", key)
	}

	// Create custom transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cp.config.DialTimeout,
			KeepAlive: cp.config.KeepAlive,
		}).DialContext,
		MaxIdleConns:        cp.config.MaxIdleConns,
		MaxIdleConnsPerHost: cp.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cp.config.MaxConnsPerHost,
		IdleConnTimeout:     cp.config.IdleConnTimeout,
		TLSHandshakeTimeout: cp.config.TLSHandshakeTimeout,
		DisableCompression:  cp.config.DisableCompression,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cp.config.InsecureSkipVerify,
		},
	}

	// Apply middleware to transport
	var roundTripper http.RoundTripper = transport
	for i := len(cp.middleware) - 1; i >= 0; i-- {
		middleware := cp.middleware[i]
		roundTripper = RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return middleware(req, roundTripper)
		})
	}

	// Create HTTP client
	client := &http.Client{
		Transport: roundTripper,
		Timeout:   config.Timeout,
	}

	// Configure redirect policy
	if !config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if config.MaxRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		}
	}

	cp.clients[key] = client

	cp.logger.Debug("created HTTP client", "key", key, "timeout", config.Timeout)

	return client, nil
}

// RemoveClient removes a client from the pool
func (cp *ClientPool) RemoveClient(key string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	client, exists := cp.clients[key]
	if !exists {
		return fmt.Errorf("client not found for key: %s", key)
	}

	// Close idle connections
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	delete(cp.clients, key)

	cp.logger.Debug("removed HTTP client", "key", key)

	return nil
}

// GetPoolMetrics returns connection pool metrics
func (cp *ClientPool) GetPoolMetrics() *ClientPoolMetrics {
	cp.metrics.mu.RLock()
	defer cp.metrics.mu.RUnlock()

	cp.mu.RLock()
	totalClients := len(cp.clients)
	cp.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	return &ClientPoolMetrics{
		TotalClients:    totalClients,
		TotalRequests:   cp.metrics.TotalRequests,
		SuccessfulReqs:  cp.metrics.SuccessfulReqs,
		FailedReqs:      cp.metrics.FailedReqs,
		AverageLatency:  cp.metrics.AverageLatency,
		LastRequestTime: cp.metrics.LastRequestTime,
	}
}

// CloseAll closes all clients in the pool
func (cp *ClientPool) CloseAll() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for key, client := range cp.clients {
		// Close idle connections
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
		cp.logger.Debug("closed HTTP client", "key", key)
	}

	cp.clients = make(map[string]*http.Client)

	cp.logger.Info("closed all HTTP clients")

	return nil
}

// AddMiddleware adds middleware to the client pool
func (cp *ClientPool) AddMiddleware(middleware MiddlewareFunc) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.middleware = append(cp.middleware, middleware)

	cp.logger.Debug("added middleware to client pool")
}

// Middleware implementations

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(logger logging.Logger) MiddlewareFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		start := time.Now()

		logger.Debug("HTTP request",
			"method", req.Method,
			"url", req.URL.String(),
			"user_agent", req.UserAgent())

		resp, err := next.RoundTrip(req)
		duration := time.Since(start)

		if err != nil {
			logger.Warn("HTTP request failed",
				"method", req.Method,
				"url", req.URL.String(),
				"duration", duration,
				"error", err)
			return nil, err
		}

		logger.Debug("HTTP response",
			"method", req.Method,
			"url", req.URL.String(),
			"status", resp.StatusCode,
			"duration", duration)

		return resp, nil
	}
}

// MetricsMiddleware tracks HTTP request metrics
func MetricsMiddleware(metrics *ClientPoolMetrics) MiddlewareFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		start := time.Now()

		resp, err := next.RoundTrip(req)
		duration := time.Since(start)

		// Update metrics
		metrics.mu.Lock()
		metrics.TotalRequests++
		if err != nil {
			metrics.FailedReqs++
		} else {
			metrics.SuccessfulReqs++
		}

		// Update average latency (simple moving average)
		if metrics.TotalRequests == 1 {
			metrics.AverageLatency = duration
		} else {
			metrics.AverageLatency = (metrics.AverageLatency + duration) / 2
		}

		metrics.LastRequestTime = time.Now()
		metrics.mu.Unlock()

		return resp, err
	}
}

// RetryMiddleware adds retry logic to HTTP requests
func RetryMiddleware(maxRetries int, baseDelay time.Duration) MiddlewareFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			// Clone request for retry (body might be consumed)
			reqClone := req.Clone(req.Context())

			resp, err := next.RoundTrip(reqClone)
			if err == nil && resp.StatusCode < 500 {
				return resp, nil
			}

			lastErr = err
			if resp != nil {
				resp.Body.Close()
			}

			// Don't retry on last attempt
			if attempt == maxRetries {
				break
			}

			// Check if error is retryable
			if err != nil || isRetryableStatusCode(resp.StatusCode) {
				// Calculate backoff delay
				delay := baseDelay * time.Duration(1<<attempt)
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}

				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(delay):
					// Continue to next attempt
				}
			} else {
				// Not retryable
				break
			}
		}

		return nil, lastErr
	}
}

// AuthMiddleware adds authentication to HTTP requests
func AuthMiddleware(authFunc func(*http.Request) error) MiddlewareFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		// Add authentication
		if err := authFunc(req); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}

		return next.RoundTrip(req)
	}
}

// TimeoutMiddleware adds timeout to HTTP requests
func TimeoutMiddleware(timeout time.Duration) MiddlewareFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()

		reqWithTimeout := req.WithContext(ctx)
		return next.RoundTrip(reqWithTimeout)
	}
}

// UserAgentMiddleware adds User-Agent header to HTTP requests
func UserAgentMiddleware(userAgent string) MiddlewareFunc {
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", userAgent)
		}
		return next.RoundTrip(req)
	}
}

// Helper functions

// isRetryableStatusCode checks if an HTTP status code is retryable
func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// PooledHTTPClient implements HTTPClientInterface using the client pool
type PooledHTTPClient struct {
	pool *ClientPool
	key  string
}

// NewPooledHTTPClient creates a new pooled HTTP client
func NewPooledHTTPClient(pool *ClientPool, key string) *PooledHTTPClient {
	return &PooledHTTPClient{
		pool: pool,
		key:  key,
	}
}

// Do executes an HTTP request
func (pc *PooledHTTPClient) Do(req *http.Request) (*http.Response, error) {
	client, err := pc.pool.GetClient(pc.key)
	if err != nil {
		return nil, fmt.Errorf("failed to get HTTP client: %w", err)
	}

	return client.Do(req)
}

// Get performs a GET request
func (pc *PooledHTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return pc.Do(req)
}

// Post performs a POST request
func (pc *PooledHTTPClient) Post(ctx context.Context, url string, headers map[string]string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return pc.Do(req)
}

// Put performs a PUT request
func (pc *PooledHTTPClient) Put(ctx context.Context, url string, headers map[string]string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return pc.Do(req)
}

// Delete performs a DELETE request
func (pc *PooledHTTPClient) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return pc.Do(req)
}

// Close closes the client and cleans up resources
func (pc *PooledHTTPClient) Close() error {
	return pc.pool.RemoveClient(pc.key)
}
