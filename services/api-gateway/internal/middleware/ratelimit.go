package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	config       config.RateLimitConfig
	log          *logger.Logger
	metrics      *metrics.Collector
	globalLimiter *rate.Limiter
	ipLimiters   map[string]*rate.Limiter
	mu           sync.RWMutex
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(cfg config.RateLimitConfig, log *logger.Logger, m *metrics.Collector) *RateLimitMiddleware {
	var globalLimiter *rate.Limiter
	if cfg.Enabled && cfg.Global.RequestsPerSecond > 0 {
		globalLimiter = rate.NewLimiter(
			rate.Limit(cfg.Global.RequestsPerSecond),
			cfg.Global.Burst,
		)
	}

	return &RateLimitMiddleware{
		config:        cfg,
		log:           log,
		metrics:       m,
		globalLimiter: globalLimiter,
		ipLimiters:    make(map[string]*rate.Limiter),
	}
}

// GlobalLimit applies global rate limiting
func (r *RateLimitMiddleware) GlobalLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.config.Enabled || r.globalLimiter == nil {
			c.Next()
			return
		}

		if !r.globalLimiter.Allow() {
			r.metrics.RecordRateLimitExceeded(c.Request.URL.Path, "global")
			r.log.Warn("Global rate limit exceeded", "path", c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Global rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointLimit applies endpoint-specific rate limiting
func (r *RateLimitMiddleware) EndpointLimit() gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)

	// Pre-create limiters for configured endpoints
	for path, limits := range r.config.Endpoints {
		limiters[path] = rate.NewLimiter(
			rate.Limit(limits.RequestsPerSecond),
			limits.Burst,
		)
	}

	return func(c *gin.Context) {
		if !r.config.Enabled {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		limiter, exists := limiters[path]

		if !exists {
			c.Next()
			return
		}

		if !limiter.Allow() {
			r.metrics.RecordRateLimitExceeded(path, "endpoint")
			r.log.Warn("Endpoint rate limit exceeded",
				"path", path,
				"client_ip", c.ClientIP(),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Endpoint rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPLimit applies per-IP rate limiting
func (r *RateLimitMiddleware) IPLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.config.Enabled || r.config.PerIP.RequestsPerMinute <= 0 {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := r.getIPLimiter(ip)

		if !limiter.Allow() {
			r.metrics.RecordRateLimitExceeded(c.Request.URL.Path, "ip")
			r.log.Warn("IP rate limit exceeded",
				"client_ip", ip,
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "IP rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getIPLimiter gets or creates a rate limiter for an IP address
func (r *RateLimitMiddleware) getIPLimiter(ip string) *rate.Limiter {
	r.mu.RLock()
	limiter, exists := r.ipLimiters[ip]
	r.mu.RUnlock()

	if exists {
		return limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	limiter, exists = r.ipLimiters[ip]
	if exists {
		return limiter
	}

	// Create new limiter (convert requests per minute to requests per second)
	rps := float64(r.config.PerIP.RequestsPerMinute) / 60.0
	limiter = rate.NewLimiter(rate.Limit(rps), r.config.PerIP.Burst)
	r.ipLimiters[ip] = limiter

	// Clean up old limiters periodically to prevent memory leak
	if len(r.ipLimiters) > 10000 {
		go r.cleanupLimiters()
	}

	return limiter
}

// cleanupLimiters removes idle limiters
func (r *RateLimitMiddleware) cleanupLimiters() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Simple cleanup: remove half of the limiters
	// In production, you'd want a more sophisticated approach (LRU cache, TTL, etc.)
	if len(r.ipLimiters) > 5000 {
		newMap := make(map[string]*rate.Limiter)
		count := 0
		for k, v := range r.ipLimiters {
			if count < 5000 {
				newMap[k] = v
				count++
			}
		}
		r.ipLimiters = newMap
		r.log.Info("Cleaned up IP rate limiters", "remaining", len(r.ipLimiters))
	}
}

// RateLimitHeaders adds rate limit information to response headers
func (r *RateLimitMiddleware) RateLimitHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.config.Enabled {
			c.Next()
			return
		}

		// Check endpoint-specific limits
		path := c.Request.URL.Path
		if limits, exists := r.config.Endpoints[path]; exists {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limits.RequestsPerSecond))
			c.Header("X-RateLimit-Burst", fmt.Sprintf("%d", limits.Burst))
		} else if r.config.Global.RequestsPerSecond > 0 {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", r.config.Global.RequestsPerSecond))
			c.Header("X-RateLimit-Burst", fmt.Sprintf("%d", r.config.Global.Burst))
		}

		c.Next()
	}
}
