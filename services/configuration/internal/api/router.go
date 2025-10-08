package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter sets up the HTTP router
func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	// Serve static files for admin UI
	router.Static("/web", "./web")

	// Admin routes - serve the admin dashboard
	router.GET("/", handler.ServeAdmin)
	router.GET("/admin", handler.ServeAdmin)

	// Health endpoints
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Service info endpoint (public)
	router.GET("/api/service-info", handler.ServiceInfo)

	// API v1 routes - protected with API key authentication
	v1 := router.Group("/api/v1")
	v1.Use(APIKeyMiddleware())
	{
		// Configuration routes
		configs := v1.Group("/configurations")
		{
			configs.POST("", handler.CreateConfiguration)
			configs.GET("", handler.ListConfigurations)
			configs.GET("/:id", handler.GetConfiguration)
			configs.GET("/key/:key", handler.GetConfigurationByKey)
			configs.PUT("/:id", handler.UpdateConfiguration)
			configs.POST("/:id/activate", handler.ActivateConfiguration)
			configs.POST("/:id/deactivate", handler.DeactivateConfiguration)
			configs.DELETE("/:id", handler.DeleteConfiguration)

			// Version management
			configs.GET("/:id/versions", handler.GetVersions)
			configs.POST("/:id/rollback", handler.RollbackConfiguration)

			// Audit logs
			configs.GET("/:id/audit-logs", handler.GetAuditLogs)
		}
	}

	return router
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
