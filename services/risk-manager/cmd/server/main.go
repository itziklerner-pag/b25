package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b25/services/risk-manager/internal/cache"
	"github.com/b25/services/risk-manager/internal/client"
	"github.com/b25/services/risk-manager/internal/config"
	"github.com/b25/services/risk-manager/internal/emergency"
	"github.com/b25/services/risk-manager/internal/grpc"
	"github.com/b25/services/risk-manager/internal/limits"
	"github.com/b25/services/risk-manager/internal/middleware"
	"github.com/b25/services/risk-manager/internal/monitor"
	"github.com/b25/services/risk-manager/internal/repository"
	"github.com/b25/services/risk-manager/internal/risk"
	pb "github.com/b25/services/risk-manager/proto"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	grpcServer "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting risk manager service",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode),
	)

	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient := initRedis(cfg.Redis)
	defer redisClient.Close()

	// Test connections
	ctx := context.Background()
	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}
	logger.Info("database connection established")

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to ping Redis", zap.Error(err))
	}
	logger.Info("Redis connection established")

	// Initialize NATS
	nc, err := initNATS(cfg.NATS, logger)
	if err != nil {
		logger.Fatal("failed to initialize NATS", zap.Error(err))
	}
	defer nc.Close()
	logger.Info("NATS connection established")

	// Initialize Account Monitor client (with fallback to mock if unavailable)
	var accountMonitorClient *client.AccountMonitorClient
	if cfg.Risk.AccountMonitorURL != "" {
		accountMonitorClient, err = client.NewAccountMonitorClient(
			cfg.Risk.AccountMonitorURL,
			"default_user", // TODO: Get from config or context
			logger,
		)
		if err != nil {
			logger.Warn("failed to connect to Account Monitor - will use mock data",
				zap.String("url", cfg.Risk.AccountMonitorURL),
				zap.Error(err),
			)
			accountMonitorClient = nil
		} else {
			defer accountMonitorClient.Close()
			logger.Info("Account Monitor client connected")
		}
	} else {
		logger.Warn("Account Monitor URL not configured - using mock data - NOT SAFE FOR PRODUCTION")
	}

	// Initialize components
	policyRepo := repository.NewPolicyRepository(db)
	policyCache := cache.NewPolicyCache(redisClient, cfg.Risk.PolicyCacheTTL)
	priceCache := cache.NewMarketPriceCache(redisClient, cfg.Risk.CacheTTL)

	// Load policies from database
	policies, err := policyRepo.GetActive(ctx)
	if err != nil {
		logger.Warn("failed to load policies from database", zap.Error(err))
		policies = getDefaultPolicies()
	} else {
		logger.Info("policies loaded from database", zap.Int("count", len(policies)))
	}

	// Cache policies
	if err := policyCache.SetPolicies(ctx, policies); err != nil {
		logger.Error("failed to cache policies", zap.Error(err))
	}

	// Initialize policy engine
	policyEngine := limits.NewPolicyEngine()
	policyEngine.LoadPolicies(policies)

	// Initialize risk calculator
	calculator := risk.NewCalculator(cfg.Risk.MaxLeverage, cfg.Risk.MaxDrawdownPercent)

	// Initialize alert publisher
	alertPublisher := monitor.NewNATSAlertPublisher(nc, logger, cfg.NATS.AlertSubject, cfg.Risk.AlertWindow)
	go alertPublisher.StartCleanup(ctx)

	// Initialize emergency stop manager
	stopManager := emergency.NewStopManager(logger, alertPublisher)

	// Initialize metrics collector
	metricsCollector := monitor.NewMetricsCollector()

	// Initialize gRPC server
	riskServer := grpc.NewRiskServer(
		logger,
		calculator,
		policyEngine,
		policyCache,
		priceCache,
		stopManager,
		accountMonitorClient, // Pass Account Monitor client (or nil for mock)
		metricsCollector,     // Pass metrics collector
	)

	// Start gRPC server
	grpcSrv := startGRPCServer(cfg, logger, riskServer)
	defer grpcSrv.GracefulStop()

	// Start metrics server
	if cfg.Metrics.Enabled {
		go startMetricsServer(cfg, logger)
	}

	// Start risk monitor
	riskMonitor := monitor.NewRiskMonitor(
		logger,
		calculator,
		policyEngine,
		policyRepo,
		stopManager,
		alertPublisher,
		cfg.Risk.MonitorInterval,
		accountMonitorClient,  // Pass Account Monitor client (or nil for mock)
		"default_user",        // TODO: Get from config
	)

	go func() {
		if err := riskMonitor.Run(ctx); err != nil {
			logger.Error("risk monitor error", zap.Error(err))
		}
	}()

	// Start policy refresh loop
	go startPolicyRefreshLoop(ctx, logger, policyRepo, policyCache, policyEngine)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down risk manager service...")

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Wait for shutdown
	<-shutdownCtx.Done()

	logger.Info("risk manager service stopped")
}

func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	return logger, nil
}

func initDatabase(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func initRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:       cfg.GetRedisAddr(),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
	})
}

func initNATS(cfg config.NATSConfig, logger *zap.Logger) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnect),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Warn("NATS connection closed")
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return nc, nil
}

func startGRPCServer(cfg *config.Config, logger *zap.Logger, riskServer *grpc.RiskServer) *grpcServer.Server {
	// Create listener
	lis, err := net.Listen("tcp", cfg.GRPC.GetGRPCAddr())
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Create interceptors
	authInterceptor := middleware.NewAuthInterceptor(logger, cfg.GRPC.APIKey, cfg.GRPC.AuthEnabled)
	loggingInterceptor := middleware.NewLoggingInterceptor(logger)

	// Log authentication status
	if cfg.GRPC.AuthEnabled {
		if cfg.GRPC.APIKey == "" {
			logger.Fatal("gRPC authentication enabled but API key not configured")
		}
		logger.Info("gRPC authentication enabled")
	} else {
		logger.Warn("gRPC authentication disabled - NOT SAFE FOR PRODUCTION")
	}

	// Create gRPC server with options
	grpcSrv := grpcServer.NewServer(
		grpcServer.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.GRPC.MaxConnectionIdle,
			MaxConnectionAge:  cfg.GRPC.MaxConnectionAge,
			Time:              cfg.GRPC.KeepAliveInterval,
			Timeout:           cfg.GRPC.KeepAliveTimeout,
		}),
		grpcServer.ChainUnaryInterceptor(
			loggingInterceptor.Unary(),
			authInterceptor.Unary(),
		),
		grpcServer.ChainStreamInterceptor(
			loggingInterceptor.Stream(),
			authInterceptor.Stream(),
		),
	)

	// Register services
	pb.RegisterRiskManagerServer(grpcSrv, riskServer)

	// Register health check
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthServer)

	// Start server in background
	go func() {
		logger.Info("gRPC server started", zap.String("addr", cfg.GRPC.GetGRPCAddr()))
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal("gRPC server error", zap.Error(err))
		}
	}()

	return grpcSrv
}

// setCORSHeaders sets CORS headers for health endpoints
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func startMetricsServer(cfg *config.Config, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		setCORSHeaders(w)

		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	logger.Info("metrics server started", zap.String("addr", srv.Addr))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("metrics server error", zap.Error(err))
	}
}

func startPolicyRefreshLoop(ctx context.Context, logger *zap.Logger, repo *repository.PolicyRepository, cache *cache.PolicyCache, engine *limits.PolicyEngine) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			policies, err := repo.GetActive(ctx)
			if err != nil {
				logger.Error("failed to refresh policies", zap.Error(err))
				continue
			}

			// Update cache
			if err := cache.SetPolicies(ctx, policies); err != nil {
				logger.Error("failed to update policy cache", zap.Error(err))
			}

			// Update engine
			engine.LoadPolicies(policies)

			logger.Debug("policies refreshed", zap.Int("count", len(policies)))

		case <-ctx.Done():
			return
		}
	}
}

func getDefaultPolicies() []*limits.Policy {
	return []*limits.Policy{
		{
			ID:        "default-leverage",
			Name:      "Max Leverage Limit",
			Type:      limits.PolicyTypeHard,
			Metric:    "leverage",
			Operator:  limits.OperatorLessThanOrEqual,
			Threshold: 10.0,
			Scope:     limits.PolicyScopeAccount,
			Enabled:   true,
			Priority:  100,
		},
		{
			ID:        "default-margin",
			Name:      "Min Margin Ratio",
			Type:      limits.PolicyTypeHard,
			Metric:    "margin_ratio",
			Operator:  limits.OperatorGreaterThanOrEqual,
			Threshold: 1.0,
			Scope:     limits.PolicyScopeAccount,
			Enabled:   true,
			Priority:  100,
		},
		{
			ID:        "default-drawdown",
			Name:      "Max Drawdown Emergency Stop",
			Type:      limits.PolicyTypeEmergency,
			Metric:    "drawdown_max",
			Operator:  limits.OperatorGreaterThan,
			Threshold: 0.25,
			Scope:     limits.PolicyScopeAccount,
			Action:    "emergency_stop",
			Enabled:   true,
			Priority:  200,
		},
	}
}
