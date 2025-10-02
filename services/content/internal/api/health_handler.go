package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

type HealthStatus struct {
	Status    string        `json:"status"`
	Uptime    string        `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
	Version   string        `json:"version"`
}

// Health returns service health status
// @Summary Health check
// @Tags system
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	uptime := time.Since(startTime)

	status := HealthStatus{
		Status:    "healthy",
		Uptime:    uptime.String(),
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	c.JSON(http.StatusOK, status)
}

// Ready returns readiness status
// @Summary Readiness check
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ready [get]
func (h *Handler) Ready(c *gin.Context) {
	// TODO: Check database connectivity
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
