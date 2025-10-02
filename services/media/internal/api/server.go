package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/b25/services/media/internal/cache"
	"github.com/yourorg/b25/services/media/internal/config"
	"github.com/yourorg/b25/services/media/internal/processing"
	"github.com/yourorg/b25/services/media/internal/quota"
	"github.com/yourorg/b25/services/media/internal/security"
	"github.com/yourorg/b25/services/media/internal/storage"
)

// ServerConfig holds dependencies for the API server
type ServerConfig struct {
	Config          *config.Config
	DB              *sql.DB
	Cache           cache.Cache
	Storage         storage.Storage
	ImageProcessor  *processing.ImageProcessor
	VideoProcessor  *processing.VideoProcessor
	SecurityScanner security.Scanner
	QuotaManager    *quota.QuotaManager
}

// Server represents the API server
type Server struct {
	config  ServerConfig
	handler *Handler
}

// NewServer creates a new API server
func NewServer(cfg ServerConfig) *Server {
	handler := NewHandler(cfg)

	return &Server{
		config:  cfg,
		handler: handler,
	}
}

// Router sets up and returns the HTTP router
func (s *Server) Router() *gin.Engine {
	// Set Gin mode
	if s.config.Config.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	// Health and metrics endpoints
	if s.config.Config.Health.Path != "" {
		router.GET(s.config.Config.Health.Path, s.handler.Health)
	}

	if s.config.Config.Metrics.Enabled {
		router.GET(s.config.Config.Metrics.Path, s.handler.Metrics)
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Media upload endpoints
		media := v1.Group("/media")
		{
			media.POST("/upload", s.handler.UploadMedia)
			media.POST("/upload/multipart/init", s.handler.InitMultipartUpload)
			media.POST("/upload/multipart/:session_id/chunk", s.handler.UploadChunk)
			media.POST("/upload/multipart/:session_id/complete", s.handler.CompleteMultipartUpload)
			media.DELETE("/upload/multipart/:session_id", s.handler.AbortMultipartUpload)

			// Media retrieval and management
			media.GET("/:id", s.handler.GetMedia)
			media.GET("/:id/download", s.handler.DownloadMedia)
			media.GET("/:id/stream", s.handler.StreamMedia)
			media.GET("", s.handler.ListMedia)
			media.DELETE("/:id", s.handler.DeleteMedia)
			media.GET("/:id/metadata", s.handler.GetMetadata)

			// Image operations
			media.GET("/:id/thumbnail/:size", s.handler.GetThumbnail)
			media.POST("/:id/resize", s.handler.ResizeImage)
			media.POST("/:id/convert", s.handler.ConvertFormat)

			// Video operations
			media.GET("/:id/variants", s.handler.GetVideoVariants)
			media.GET("/:id/playlist", s.handler.GetHLSPlaylist)
		}

		// Quota management endpoints
		quotaGroup := v1.Group("/quota")
		{
			quotaGroup.GET("/user/:user_id", s.handler.GetUserQuota)
			quotaGroup.GET("/org/:org_id", s.handler.GetOrgQuota)
			quotaGroup.PUT("/user/:user_id", s.handler.SetUserQuota)
			quotaGroup.PUT("/org/:org_id", s.handler.SetOrgQuota)
		}

		// Statistics endpoints
		stats := v1.Group("/stats")
		{
			stats.GET("/user/:user_id", s.handler.GetUserStats)
			stats.GET("/org/:org_id", s.handler.GetOrgStats)
		}

		// Admin endpoints
		admin := v1.Group("/admin")
		{
			admin.POST("/quota/recalculate/:entity_id", s.handler.RecalculateQuota)
			admin.GET("/storage/status", s.handler.GetStorageStatus)
		}
	}

	// Serve static media files (for local storage)
	if s.config.Config.Storage.Type == "local" {
		router.Static("/media", s.config.Config.Storage.Local.BasePath)
	}

	return router
}
