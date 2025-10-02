package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/yourorg/b25/services/search/internal/config"
)

// SetupRouter configures and returns the HTTP router
func SetupRouter(handler *Handler, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(MetricsMiddleware())

	// CORS
	if cfg.CORS.Enabled {
		corsConfig := cors.Config{
			AllowOrigins:     cfg.CORS.AllowedOrigins,
			AllowMethods:     cfg.CORS.AllowedMethods,
			AllowHeaders:     cfg.CORS.AllowedHeaders,
			MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
			AllowCredentials: true,
		}
		router.Use(cors.New(corsConfig))
	}

	// Rate limiting
	if cfg.RateLimit.Enabled {
		limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerSecond), cfg.RateLimit.Burst)
		router.Use(RateLimitMiddleware(limiter))
	}

	// Health checks
	router.GET(cfg.Health.Path, handler.Health)
	router.GET(cfg.Health.ReadinessPath, handler.Readiness)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Search endpoints
		v1.POST("/search", handler.Search)
		v1.POST("/autocomplete", handler.Autocomplete)

		// Indexing endpoints
		v1.POST("/index", handler.Index)
		v1.POST("/index/bulk", handler.BulkIndex)

		// Analytics endpoints
		analytics := v1.Group("/analytics")
		{
			analytics.POST("/click", handler.TrackClick)
			analytics.GET("/popular", handler.GetPopularSearches)
			analytics.GET("/stats", handler.GetSearchStats)
		}
	}

	return router
}

// SetupMetricsRouter creates a separate router for metrics
func SetupMetricsRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		if statusCode >= 500 {
			logger.Error("HTTP request", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}
	}
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, ErrorResponse{
				Error:   "Rate limit exceeded",
				Message: "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// MetricsMiddleware collects Prometheus metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()

		// Update metrics
		httpRequestsTotal.WithLabelValues(method, path, string(rune(statusCode))).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
