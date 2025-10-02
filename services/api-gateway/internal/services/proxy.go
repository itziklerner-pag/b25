package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/b25/api-gateway/internal/breaker"
	"github.com/b25/api-gateway/internal/cache"
	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// ProxyService handles proxying requests to backend services
type ProxyService struct {
	config  *config.Config
	log     *logger.Logger
	metrics *metrics.Collector
	breaker *breaker.Manager
	cache   *cache.Cache
	client  *http.Client
}

// NewProxyService creates a new proxy service
func NewProxyService(
	cfg *config.Config,
	log *logger.Logger,
	m *metrics.Collector,
	b *breaker.Manager,
	c *cache.Cache,
) *ProxyService {
	return &ProxyService{
		config:  cfg,
		log:     log,
		metrics: m,
		breaker: b,
		cache:   c,
		client: &http.Client{
			Timeout: cfg.Timeout.Default,
		},
	}
}

// ProxyRequest proxies a request to a backend service
func (p *ProxyService) ProxyRequest(c *gin.Context, service string) {
	start := time.Now()

	// Get service URL
	serviceURL := p.config.GetServiceURL(service)
	if serviceURL == "" {
		p.log.Errorw("Unknown service", "service", service)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unknown service"})
		return
	}

	// Check cache for GET requests
	if c.Request.Method == "GET" && p.cache != nil {
		cacheKey := p.cache.GenerateKey(c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
		if cached, err := p.cache.Get(c.Request.Context(), cacheKey); err == nil {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
	}

	// Build target URL
	targetURL := fmt.Sprintf("%s%s", serviceURL, c.Request.URL.Path)
	if c.Request.URL.RawQuery != "" {
		targetURL = fmt.Sprintf("%s?%s", targetURL, c.Request.URL.RawQuery)
	}

	// Execute through circuit breaker
	result, err := p.breaker.Execute(service, func() (interface{}, error) {
		return p.executeRequest(c, service, targetURL)
	})

	duration := time.Since(start)

	if err != nil {
		p.metrics.RecordUpstreamError(service, "circuit_breaker")
		p.log.Errorw("Circuit breaker error",
			"service", service,
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Service temporarily unavailable",
		})
		return
	}

	resp := result.(*http.Response)
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.metrics.RecordUpstreamError(service, "read_error")
		p.log.Errorw("Failed to read response body", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read response"})
		return
	}

	// Record metrics
	p.metrics.RecordUpstreamRequest(service, c.Request.Method, resp.StatusCode, duration.Seconds())

	// Cache successful GET responses
	if c.Request.Method == "GET" && resp.StatusCode == http.StatusOK && p.cache != nil {
		cacheKey := p.cache.GenerateKey(c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
		ttl := p.cache.GetTTL(c.Request.URL.Path)
		if err := p.cache.Set(c.Request.Context(), cacheKey, body, ttl); err != nil {
			p.log.Warnw("Failed to cache response", "error", err)
		}
	}

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Send response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// executeRequest executes the actual HTTP request
func (p *ProxyService) executeRequest(c *gin.Context, service, targetURL string) (*http.Response, error) {
	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create new request
	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		targetURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add gateway headers
	if p.config.Transformation.Enabled {
		for key, value := range p.config.Transformation.RequestHeaders {
			// Handle template variables
			if value == "${client_ip}" {
				value = c.ClientIP()
			}
			req.Header.Set(key, value)
		}
	}

	// Get timeout for this endpoint
	timeout := p.getTimeout(c.Request.URL.Path)
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request with retry
	return p.executeWithRetry(req, service)
}

// executeWithRetry executes a request with retry logic
func (p *ProxyService) executeWithRetry(req *http.Request, service string) (*http.Response, error) {
	if !p.config.Retry.Enabled {
		return p.client.Do(req)
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt < p.config.Retry.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Calculate backoff
			backoff := p.config.Retry.InitialInterval * time.Duration(1<<uint(attempt-1))
			if backoff > p.config.Retry.MaxInterval {
				backoff = p.config.Retry.MaxInterval
			}
			time.Sleep(backoff)

			p.log.Infow("Retrying request",
				"service", service,
				"attempt", attempt+1,
				"max_attempts", p.config.Retry.MaxAttempts,
			)
		}

		resp, err = p.client.Do(req)
		if err != nil {
			p.log.Warnw("Request failed",
				"service", service,
				"attempt", attempt+1,
				"error", err,
			)
			continue
		}

		// Check if status code is retryable
		if !p.isRetryableStatusCode(resp.StatusCode) {
			return resp, nil
		}

		// Close response body before retry
		resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// isRetryableStatusCode checks if a status code is retryable
func (p *ProxyService) isRetryableStatusCode(statusCode int) bool {
	for _, code := range p.config.Retry.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// getTimeout returns the timeout for a given endpoint
func (p *ProxyService) getTimeout(path string) time.Duration {
	if timeout, exists := p.config.Timeout.Endpoints[path]; exists {
		return timeout
	}
	return p.config.Timeout.Default
}
