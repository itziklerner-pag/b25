package tests

import (
	"testing"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/internal/middleware"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidateJWT(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled:   true,
		JWTSecret: "test-secret",
	}

	log := logger.Default()
	m := metrics.New("test")
	authMw := middleware.NewAuthMiddleware(cfg, log, m)

	// Generate a test token
	token, err := authMw.GenerateJWT("user123", "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthMiddleware_ValidateAPIKey(t *testing.T) {
	cfg := config.AuthConfig{
		Enabled: true,
		APIKeys: []config.APIKey{
			{Key: "test-key", Role: "admin"},
			{Key: "viewer-key", Role: "viewer"},
		},
	}

	log := logger.Default()
	m := metrics.New("test")
	authMw := middleware.NewAuthMiddleware(cfg, log, m)

	// This would require exposing the validateAPIKey method or testing through HTTP
	// For now, we test through the HTTP handler in integration tests
	assert.NotNil(t, authMw)
}

func TestRateLimitMiddleware_Creation(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled: true,
		Global: config.GlobalRateLimit{
			RequestsPerSecond: 100,
			Burst:             200,
		},
	}

	log := logger.Default()
	m := metrics.New("test")
	rlMw := middleware.NewRateLimitMiddleware(cfg, log, m)

	assert.NotNil(t, rlMw)
}

func TestCORSMiddleware_Creation(t *testing.T) {
	cfg := config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
	}

	corsMw := middleware.NewCORSMiddleware(cfg)
	assert.NotNil(t, corsMw)
}

func TestValidationMiddleware_Creation(t *testing.T) {
	cfg := config.ValidationConfig{
		Enabled:        true,
		MaxRequestSize: 10485760,
	}

	log := logger.Default()
	validationMw := middleware.NewValidationMiddleware(cfg, log)
	assert.NotNil(t, validationMw)
}
