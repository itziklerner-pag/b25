package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// corsMiddleware adds CORS headers to Gin responses
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	// Apply CORS headers
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "configuration-service",
		"version": "1.0.0",
	})
}

// ReadinessCheck handles readiness check requests
func (h *Handler) ReadinessCheck(c *gin.Context) {
	// Apply CORS headers
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	checks := gin.H{}
	allHealthy := true

	// Check database connectivity
	if err := h.db.Ping(); err != nil {
		checks["database"] = "error: " + err.Error()
		allHealthy = false
	} else {
		checks["database"] = "ok"
	}

	// Check NATS connectivity
	if h.natsConn == nil || !h.natsConn.IsConnected() {
		checks["nats"] = "disconnected"
		allHealthy = false
	} else {
		checks["nats"] = "ok"
	}

	if !allHealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": checks,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": checks,
	})
}
