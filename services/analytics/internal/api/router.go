package api

import (
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(cfg *config.Config, handler *Handler, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(LoggerMiddleware(logger))
	router.Use(gin.Recovery())

	if cfg.Security.CORS.Enabled {
		router.Use(CORSMiddleware(cfg.Security.CORS))
	}

	if cfg.Security.RateLimit.Enabled {
		router.Use(RateLimitMiddleware(cfg.Security.RateLimit))
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

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		logger.Info("HTTP request",
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

// RateLimitMiddleware implements basic rate limiting
func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis-based rate limiting
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting with Redis
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
