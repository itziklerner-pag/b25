package api

import (
	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/middleware"
	"github.com/b25/services/notification/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(
	cfg *config.Config,
	logger *zap.Logger,
	notificationService *service.NotificationService,
	userService *service.UserService,
	templateService *service.TemplateService,
) *gin.Engine {
	// Set gin mode
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// Health check endpoints
	router.GET("/health", HealthCheck)
	router.GET("/ready", ReadinessCheck(notificationService))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Notification handlers
		notificationHandler := NewNotificationHandler(notificationService, logger)
		notifications := v1.Group("/notifications")
		{
			notifications.POST("", notificationHandler.CreateNotification)
			notifications.GET("", notificationHandler.ListNotifications)
			notifications.GET("/:id", notificationHandler.GetNotification)
			notifications.GET("/user/:user_id", notificationHandler.GetUserNotifications)
			notifications.GET("/correlation/:correlation_id", notificationHandler.GetByCorrelationID)
		}

		// User handlers
		userHandler := NewUserHandler(userService, logger)
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)

			// Preferences
			users.GET("/:id/preferences", userHandler.GetPreferences)
			users.POST("/:id/preferences", userHandler.CreatePreference)
			users.PUT("/:id/preferences/:pref_id", userHandler.UpdatePreference)
			users.DELETE("/:id/preferences/:pref_id", userHandler.DeletePreference)

			// Devices
			users.GET("/:id/devices", userHandler.GetDevices)
			users.POST("/:id/devices", userHandler.RegisterDevice)
			users.PUT("/:id/devices/:device_id", userHandler.UpdateDevice)
			users.DELETE("/:id/devices/:device_id", userHandler.DeleteDevice)
		}

		// Template handlers
		templateHandler := NewTemplateHandler(templateService, logger)
		templates := v1.Group("/templates")
		{
			templates.POST("", templateHandler.CreateTemplate)
			templates.GET("", templateHandler.ListTemplates)
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
			templates.POST("/:id/test", templateHandler.TestTemplate)
		}
	}

	return router
}

// SetupMetricsRouter creates a separate router for metrics
func SetupMetricsRouter() *gin.Engine {
	router := gin.New()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return router
}
