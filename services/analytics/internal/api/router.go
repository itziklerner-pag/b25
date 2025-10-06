package api

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(cfg *config.Config, handler *Handler, logger *zap.Logger, redisClient *redis.Client) *gin.Engine {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(RequestIDMiddleware())
	router.Use(LoggerMiddleware(logger))
	router.Use(gin.Recovery())

	if cfg.Security.CORS.Enabled {
		router.Use(CORSMiddleware(cfg.Security.CORS))
	}

	if cfg.Security.RateLimit.Enabled {
		router.Use(RateLimitMiddleware(redisClient, cfg.Security.RateLimit, logger))
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Event tracking
		v1.POST("/events", handler.TrackEvent)
		v1.GET("/events", handler.GetEvents)
		v1.GET("/events/stats", handler.GetEventStats)

		// Metrics
		v1.GET("/metrics", handler.GetMetrics)
		v1.GET("/dashboard/metrics", handler.GetDashboardMetrics)

		// Custom events
		v1.POST("/custom-events", handler.CreateCustomEvent)
		v1.GET("/custom-events/:name", handler.GetCustomEvent)

		// Internal metrics
		v1.GET("/internal/ingestion-metrics", handler.GetIngestionMetrics)
	}

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)
	router.GET("/healthz", handler.HealthCheck)
	router.GET("/ready", handler.HealthCheck)

	return router
}

// RequestIDMiddleware generates and adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (timestamp + random component)
			requestID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix()%1000000)
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		requestID, _ := c.Get("request_id")

		c.Next()

		latency := time.Since(start)

		logger.Info("HTTP request",
			zap.String("request_id", fmt.Sprintf("%v", requestID)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implements Redis-based rate limiting
func RateLimitMiddleware(redisClient *redis.Client, cfg config.RateLimitConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Use client IP as rate limit key
		key := fmt.Sprintf("ratelimit:%s", c.ClientIP())

		// Increment counter
		pipe := redisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err := pipe.Exec(ctx)

		if err != nil {
			logger.Warn("Rate limit check failed", zap.Error(err))
			// On error, allow the request (fail open)
			c.Next()
			return
		}

		count := incr.Val()

		// Check if limit exceeded
		if count > int64(cfg.RequestsPerMinute) {
			logger.Warn("Rate limit exceeded",
				zap.String("client_ip", c.ClientIP()),
				zap.Int64("requests", count),
				zap.Int("limit", cfg.RequestsPerMinute),
			)
			c.JSON(429, gin.H{
				"error": "rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", cfg.RequestsPerMinute-int(count)))

		c.Next()
	}
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
