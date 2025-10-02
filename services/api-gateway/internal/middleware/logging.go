package middleware

import (
	"time"

	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingMiddleware handles request/response logging
type LoggingMiddleware struct {
	log     *logger.Logger
	metrics *metrics.Collector
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(log *logger.Logger, m *metrics.Collector) *LoggingMiddleware {
	return &LoggingMiddleware{
		log:     log,
		metrics: m,
	}
}

// RequestID adds a unique request ID to each request
func (l *LoggingMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// AccessLog logs each HTTP request
func (l *LoggingMiddleware) AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		requestID, _ := c.Get("request_id")

		// Get request/response sizes
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		// Log the request
		l.log.Infow("HTTP request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"request_size", requestSize,
			"response_size", c.Writer.Size(),
		)

		// Record metrics
		l.metrics.RecordRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.Seconds(),
			requestSize,
			int64(c.Writer.Size()),
		)
	}
}

// ErrorLog logs errors that occur during request processing
func (l *LoggingMiddleware) ErrorLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			requestID, _ := c.Get("request_id")

			for _, err := range c.Errors {
				l.log.Errorw("Request error",
					"request_id", requestID,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"error", err.Error(),
					"type", err.Type,
				)
			}
		}
	}
}

// Recovery recovers from panics and logs them
func (l *LoggingMiddleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				l.log.Errorw("Panic recovered",
					"request_id", requestID,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"error", err,
				)

				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ConnectionCounter tracks active connections
func (l *LoggingMiddleware) ConnectionCounter() gin.HandlerFunc {
	return func(c *gin.Context) {
		l.metrics.IncrementActiveConnections()
		defer l.metrics.DecrementActiveConnections()

		c.Next()
	}
}
