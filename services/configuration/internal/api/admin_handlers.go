package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ServeAdmin serves the admin dashboard HTML
func (h *Handler) ServeAdmin(c *gin.Context) {
	c.File("./web/index.html")
}

// ServiceInfo returns service information
func (h *Handler) ServiceInfo(c *gin.Context) {
	info := gin.H{
		"service": "configuration",
		"version": "1.0.0",
		"name":    "Configuration Service",
		"description": "Centralized configuration management for B25 trading platform",
		"port":    8085,
		"endpoints": []gin.H{
			{
				"path":        "/",
				"method":      "GET",
				"description": "Admin dashboard",
			},
			{
				"path":        "/admin",
				"method":      "GET",
				"description": "Admin dashboard",
			},
			{
				"path":        "/health",
				"method":      "GET",
				"description": "Health check endpoint",
			},
			{
				"path":        "/ready",
				"method":      "GET",
				"description": "Readiness check endpoint",
			},
			{
				"path":        "/metrics",
				"method":      "GET",
				"description": "Prometheus metrics",
			},
			{
				"path":        "/api/service-info",
				"method":      "GET",
				"description": "Service information",
			},
			{
				"path":        "/api/v1/configurations",
				"method":      "POST",
				"description": "Create new configuration",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations",
				"method":      "GET",
				"description": "List all configurations",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id",
				"method":      "GET",
				"description": "Get configuration by ID",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/key/:key",
				"method":      "GET",
				"description": "Get configuration by key",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id",
				"method":      "PUT",
				"description": "Update configuration",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id/activate",
				"method":      "POST",
				"description": "Activate configuration",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id/deactivate",
				"method":      "POST",
				"description": "Deactivate configuration",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id",
				"method":      "DELETE",
				"description": "Delete configuration",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id/versions",
				"method":      "GET",
				"description": "Get configuration versions",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id/rollback",
				"method":      "POST",
				"description": "Rollback configuration",
				"protected":   true,
			},
			{
				"path":        "/api/v1/configurations/:id/audit-logs",
				"method":      "GET",
				"description": "Get configuration audit logs",
				"protected":   true,
			},
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    info,
	})
}
