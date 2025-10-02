package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/internal/router"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Create test config
	cfg := getTestConfig()

	// Create logger and metrics
	log := logger.Default()
	m := metrics.New("test")

	// Create router
	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	r.Handler().ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestVersionEndpoint(t *testing.T) {
	cfg := getTestConfig()
	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	r.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "api-gateway", response["service"])
}

func TestAuthenticationRequired(t *testing.T) {
	cfg := getTestConfig()
	cfg.Auth.Enabled = true
	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	// Request without authentication
	req := httptest.NewRequest("GET", "/api/v1/account/balance", nil)
	w := httptest.NewRecorder()

	r.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyAuthentication(t *testing.T) {
	cfg := getTestConfig()
	cfg.Auth.Enabled = true
	cfg.Auth.APIKeys = []config.APIKey{
		{Key: "test-api-key", Role: "viewer"},
	}
	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	// Request with valid API key
	req := httptest.NewRequest("GET", "/api/v1/account/balance", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	r.Handler().ServeHTTP(w, req)

	// Should not be 401 (might be 502 due to backend not available in test)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestRateLimiting(t *testing.T) {
	cfg := getTestConfig()
	cfg.RateLimit.Enabled = true
	cfg.RateLimit.Global.RequestsPerSecond = 2
	cfg.RateLimit.Global.Burst = 2

	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	// Make requests up to the limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.Handler().ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			// Third request should be rate limited
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestCORSHeaders(t *testing.T) {
	cfg := getTestConfig()
	cfg.CORS.Enabled = true
	cfg.CORS.AllowedOrigins = []string{"http://localhost:3000"}

	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	r.Handler().ServeHTTP(w, req)

	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRequestIDMiddleware(t *testing.T) {
	cfg := getTestConfig()
	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.Handler().ServeHTTP(w, req)

	// Should have X-Request-ID header in response
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestCustomRequestID(t *testing.T) {
	cfg := getTestConfig()
	log := logger.Default()
	m := metrics.New("test")

	r, err := router.New(cfg, log, m)
	assert.NoError(t, err)

	customID := "custom-request-id-123"
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Request-ID", customID)
	w := httptest.NewRecorder()

	r.Handler().ServeHTTP(w, req)

	// Should echo back the custom request ID
	assert.Equal(t, customID, w.Header().Get("X-Request-ID"))
}

// Helper function to create test config
func getTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           8080,
			Mode:           "test",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    120 * time.Second,
			MaxHeaderBytes: 1048576,
		},
		Services: config.ServicesConfig{
			MarketData: config.ServiceConfig{
				URL:        "http://localhost:9090",
				Timeout:    5 * time.Second,
				MaxRetries: 3,
			},
			OrderExecution: config.ServiceConfig{
				URL:        "http://localhost:9091",
				Timeout:    10 * time.Second,
				MaxRetries: 2,
			},
			AccountMonitor: config.ServiceConfig{
				URL:        "http://localhost:9093",
				Timeout:    5 * time.Second,
				MaxRetries: 3,
			},
		},
		Auth: config.AuthConfig{
			Enabled:   false,
			JWTSecret: "test-secret",
		},
		RateLimit: config.RateLimitConfig{
			Enabled: false,
		},
		CORS: config.CORSConfig{
			Enabled: false,
		},
		CircuitBreaker: config.CircuitBreakerConfig{
			Enabled: false,
		},
		Cache: config.CacheConfig{
			Enabled: false,
		},
		Validation: config.ValidationConfig{
			Enabled:        true,
			MaxRequestSize: 10485760,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Metrics: config.MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "test",
		},
		Health: config.HealthConfig{
			Enabled:        true,
			Path:           "/health",
			CheckServices:  false,
			ServiceTimeout: 2 * time.Second,
		},
		Features: config.FeaturesConfig{
			EnableTracing:      true,
			EnableCompression:  true,
			EnableRequestID:    true,
			EnableAccessLog:    false,
			EnableErrorDetails: false,
		},
		Timeout: config.TimeoutConfig{
			Default: 30 * time.Second,
			Max:     300 * time.Second,
		},
		Retry: config.RetryConfig{
			Enabled:              true,
			MaxAttempts:          3,
			BackoffMultiplier:    2,
			InitialInterval:      100 * time.Millisecond,
			MaxInterval:          5 * time.Second,
			RetryableStatusCodes: []int{502, 503, 504},
		},
	}
}
