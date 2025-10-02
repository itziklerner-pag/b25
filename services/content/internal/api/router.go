package api

import (
	"github.com/b25/services/content/internal/middleware"
	"github.com/b25/services/content/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func SetupRouter(
	handler *Handler,
	authService service.AuthService,
	logger *zap.Logger,
) *gin.Engine {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware())

	// Health and metrics endpoints
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Ready)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.GET("/me", middleware.AuthMiddleware(authService), handler.GetMe)
		}

		// Content routes
		content := v1.Group("/content")
		{
			// Public routes
			content.GET("", handler.SearchContent)                  // List/search content
			content.GET("/:id", handler.GetContent)                 // Get by ID
			content.GET("/slug/:slug", handler.GetContentBySlug)    // Get by slug
			content.GET("/:id/versions", handler.GetContentVersions) // Get version history
			content.GET("/:id/versions/:version", handler.GetContentVersion)

			// Protected routes (require authentication)
			protected := content.Group("")
			protected.Use(middleware.AuthMiddleware(authService))
			{
				protected.POST("", handler.CreateContent)                 // Create
				protected.PUT("/:id", handler.UpdateContent)              // Update
				protected.DELETE("/:id", handler.DeleteContent)           // Delete
				protected.POST("/:id/publish", handler.PublishContent)    // Publish
				protected.POST("/:id/archive", handler.ArchiveContent)    // Archive
			}
		}

		// Media routes
		media := v1.Group("/media")
		media.Use(middleware.AuthMiddleware(authService))
		{
			media.POST("/upload", handler.UploadMedia)
			media.DELETE("/:id", handler.DeleteMedia)
		}
	}

	// Serve uploaded files (if using local storage)
	router.Static("/uploads", "./uploads")

	return router
}
