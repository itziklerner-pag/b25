package router

import (
	"github.com/b25/api-gateway/internal/admin"
	"github.com/b25/api-gateway/internal/breaker"
	"github.com/b25/api-gateway/internal/cache"
	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/internal/handlers"
	"github.com/b25/api-gateway/internal/middleware"
	"github.com/b25/api-gateway/internal/services"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// Router handles HTTP routing
type Router struct {
	engine *gin.Engine
	config *config.Config
	log    *logger.Logger
}

// New creates a new router
func New(cfg *config.Config, log *logger.Logger, m *metrics.Collector) (*Router, error) {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create Gin engine
	engine := gin.New()

	// Initialize components
	cache, err := cache.NewCache(cfg.Cache, log, m)
	if err != nil {
		return nil, err
	}

	breakerMgr := breaker.NewManager(cfg.CircuitBreaker, log, m)
	proxyService := services.NewProxyService(cfg, log, m, breakerMgr, cache)
	wsProxy := services.NewWebSocketProxy(cfg, log)

	// Initialize middleware
	loggingMw := middleware.NewLoggingMiddleware(log, m)
	authMw := middleware.NewAuthMiddleware(cfg.Auth, log, m)
	rateLimitMw := middleware.NewRateLimitMiddleware(cfg.RateLimit, log, m)
	corsMw := middleware.NewCORSMiddleware(cfg.CORS)
	validationMw := middleware.NewValidationMiddleware(cfg.Validation, log)
	securityMw := middleware.NewSecurityMiddleware(cfg.TLS.Enabled)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(cfg, log, breakerMgr)
	metricsHandler := handlers.NewMetricsHandler()
	versionHandler := handlers.NewVersionHandler("1.0.0", "", "")
	adminHandler := admin.NewHandler(cfg, log)

	// Global middleware
	engine.Use(loggingMw.Recovery())
	engine.Use(loggingMw.RequestID())
	engine.Use(securityMw.SecurityHeaders())
	engine.Use(corsMw.Handle())
	engine.Use(loggingMw.ConnectionCounter())

	if cfg.Features.EnableAccessLog {
		engine.Use(loggingMw.AccessLog())
	}

	engine.Use(loggingMw.ErrorLog())
	engine.Use(validationMw.ValidateRequestSize())

	// Rate limiting middleware
	if cfg.RateLimit.Enabled {
		engine.Use(rateLimitMw.GlobalLimit())
		engine.Use(rateLimitMw.IPLimit())
		engine.Use(rateLimitMw.EndpointLimit())
		engine.Use(rateLimitMw.RateLimitHeaders())
	}

	// Admin routes (public for internal access)
	engine.GET("/admin", gin.WrapF(adminHandler.HandleAdminPage))
	engine.GET("/", gin.WrapF(adminHandler.HandleAdminPage)) // Default to admin page
	engine.GET("/api/service-info", gin.WrapF(adminHandler.HandleServiceInfo))

	// Public routes (no auth required)
	public := engine.Group("/")
	{
		// Health checks
		if cfg.Health.Enabled {
			public.GET(cfg.Health.Path, healthHandler.Health)
			public.GET("/health/liveness", healthHandler.Liveness)
			public.GET("/health/readiness", healthHandler.Readiness)
		}

		// Metrics
		if cfg.Metrics.Enabled {
			public.GET(cfg.Metrics.Path, metricsHandler.Metrics())
		}

		// Version
		public.GET("/version", versionHandler.Version)
	}

	// API routes (with authentication)
	api := engine.Group("/api")
	if cfg.Auth.Enabled {
		api.Use(authMw.Authenticate())
	}

	// API v1 routes
	v1 := api.Group("/v1")
	{
		// Market Data Service routes
		marketData := v1.Group("/market-data")
		{
			marketData.GET("/symbols", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "market_data")
			})
			marketData.GET("/orderbook/:symbol", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "market_data")
			})
			marketData.GET("/trades/:symbol", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "market_data")
			})
			marketData.GET("/ticker/:symbol", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "market_data")
			})
		}

		// Order Execution Service routes
		orders := v1.Group("/orders")
		orders.Use(authMw.RequireRole("operator", "admin"))
		{
			orders.POST("", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "order_execution")
			})
			orders.GET("/:id", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "order_execution")
			})
			orders.GET("", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "order_execution")
			})
			orders.DELETE("/:id", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "order_execution")
			})
			orders.GET("/active", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "order_execution")
			})
			orders.GET("/history", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "order_execution")
			})
		}

		// Strategy Engine Service routes
		strategies := v1.Group("/strategies")
		strategies.Use(authMw.RequireRole("operator", "admin"))
		{
			strategies.GET("", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.GET("/:id", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.POST("", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.PUT("/:id", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.DELETE("/:id", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.POST("/:id/start", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.POST("/:id/stop", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
			strategies.GET("/:id/status", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "strategy_engine")
			})
		}

		// Account Monitor Service routes
		account := v1.Group("/account")
		{
			account.GET("/balance", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "account_monitor")
			})
			account.GET("/positions", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "account_monitor")
			})
			account.GET("/pnl", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "account_monitor")
			})
			account.GET("/pnl/daily", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "account_monitor")
			})
			account.GET("/trades", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "account_monitor")
			})
		}

		// Risk Manager Service routes
		risk := v1.Group("/risk")
		risk.Use(authMw.RequireRole("operator", "admin"))
		{
			risk.GET("/limits", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "risk_manager")
			})
			risk.PUT("/limits", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "risk_manager")
			})
			risk.GET("/status", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "risk_manager")
			})
			risk.POST("/emergency-stop", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "risk_manager")
			})
		}

		// Configuration Service routes
		configRoutes := v1.Group("/config")
		configRoutes.Use(authMw.RequireRole("operator", "admin"))
		{
			configRoutes.GET("", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "configuration")
			})
			configRoutes.GET("/:key", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "configuration")
			})
			configRoutes.PUT("/:key", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "configuration")
			})
			configRoutes.DELETE("/:key", authMw.RequireRole("admin"), func(c *gin.Context) {
				proxyService.ProxyRequest(c, "configuration")
			})
		}

		// Dashboard Server routes (WebSocket handled separately)
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/status", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "dashboard_server")
			})
			dashboard.GET("/summary", func(c *gin.Context) {
				proxyService.ProxyRequest(c, "dashboard_server")
			})
		}
	}

	// WebSocket routes (if enabled)
	if cfg.WebSocket.Enabled {
		// WebSocket endpoint for dashboard server
		// This proxies WebSocket connections to the dashboard-server backend
		wsRoutes := engine.Group("/ws")
		if cfg.Auth.Enabled {
			wsRoutes.Use(authMw.Authenticate())
		}
		{
			// Main WebSocket endpoint - proxies to dashboard server
			wsRoutes.GET("", wsProxy.HandleWebSocket("dashboard_server"))

			// Additional WebSocket endpoints can be added here
			// wsRoutes.GET("/market-data", wsProxy.HandleWebSocket("market_data"))
		}
	}

	return &Router{
		engine: engine,
		config: cfg,
		log:    log,
	}, nil
}

// Handler returns the HTTP handler
func (r *Router) Handler() *gin.Engine {
	return r.engine
}
