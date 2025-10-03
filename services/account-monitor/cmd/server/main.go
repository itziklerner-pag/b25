package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/yourorg/b25/services/account-monitor/internal/alert"
	"github.com/yourorg/b25/services/account-monitor/internal/balance"
	"github.com/yourorg/b25/services/account-monitor/internal/calculator"
	"github.com/yourorg/b25/services/account-monitor/internal/config"
	"github.com/yourorg/b25/services/account-monitor/internal/exchange"
	"github.com/yourorg/b25/services/account-monitor/internal/grpcserver"
	"github.com/yourorg/b25/services/account-monitor/internal/health"
	"github.com/yourorg/b25/services/account-monitor/internal/monitor"
	"github.com/yourorg/b25/services/account-monitor/internal/position"
	"github.com/yourorg/b25/services/account-monitor/internal/reconciliation"
	"github.com/yourorg/b25/services/account-monitor/internal/storage"
)

const (
	serviceName    = "account-monitor"
	serviceVersion = "1.0.0"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Set log level
	if cfg.Logging.Level == "debug" {
		logger, _ = zap.NewDevelopment()
	}

	logger.Info("Starting Account Monitor Service",
		zap.String("version", serviceVersion),
		zap.String("grpc_port", fmt.Sprintf("%d", cfg.GRPC.Port)),
		zap.String("http_port", fmt.Sprintf("%d", cfg.HTTP.Port)),
		zap.String("metrics_port", fmt.Sprintf("%d", cfg.Metrics.Port)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	logger.Info("Initializing storage connections")
	pgPool, err := storage.NewPostgresPool(ctx, cfg.Database.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pgPool.Close()

	redisClient := storage.NewRedisClient(cfg.Database.Redis)
	defer redisClient.Close()

	// Run database migrations
	if err := storage.RunMigrations(ctx, pgPool); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize NATS connection
	natsConn, err := storage.NewNATSConnection(cfg.PubSub.NATS)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()

	// Initialize managers
	positionMgr := position.NewManager(redisClient, logger)
	balanceMgr := balance.NewManager(redisClient, logger)
	pnlCalc := calculator.NewPnLCalculator(positionMgr, balanceMgr, pgPool, logger)

	// Initialize exchange client
	exchangeClient := exchange.NewBinanceClient(cfg.Exchange, logger)
	wsClient := exchange.NewWebSocketClient(cfg.Exchange, logger)

	// Initialize reconciler
	reconciler := reconciliation.NewReconciler(
		positionMgr,
		balanceMgr,
		exchangeClient,
		cfg.Reconciliation,
		logger,
	)

	// Initialize alert manager
	alertMgr := alert.NewManager(
		cfg.Alerts,
		redisClient,
		natsConn,
		pgPool,
		logger,
	)

	// Initialize account monitor
	accountMonitor := monitor.NewAccountMonitor(
		positionMgr,
		balanceMgr,
		pnlCalc,
		reconciler,
		alertMgr,
		wsClient,
		natsConn,
		logger,
	)

	// Start services
	var wg sync.WaitGroup

	// 1. Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startGRPCServer(cfg, accountMonitor, logger); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// 2. Start HTTP server (health checks, dashboard)
	wg.Add(1)
	go func() {
		defer wg.Done()
		healthChecker := health.NewChecker(pgPool, redisClient, wsClient, logger)
		if err := startHTTPServer(cfg, healthChecker, accountMonitor, logger); err != nil {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// 3. Start metrics server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startMetricsServer(cfg, logger); err != nil {
			logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	// 4. Start account monitor (WebSocket, reconciliation, alerts)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := accountMonitor.Start(ctx); err != nil {
			logger.Error("Account monitor failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received, gracefully shutting down...")

	// Cancel context and wait for goroutines
	cancel()
	wg.Wait()

	logger.Info("Account Monitor Service stopped")
}

func startGRPCServer(cfg *config.Config, monitor *monitor.AccountMonitor, logger *zap.Logger) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", cfg.GRPC.Port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(uint32(cfg.GRPC.MaxConnections)),
	)

	grpcserver.RegisterAccountMonitorServer(grpcServer, monitor)

	logger.Info("gRPC server starting", zap.Int("port", cfg.GRPC.Port))
	return grpcServer.Serve(lis)
}

func startHTTPServer(cfg *config.Config, healthChecker *health.Checker, monitor *monitor.AccountMonitor, logger *zap.Logger) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthChecker.HandleHealth)
	mux.HandleFunc("/ready", healthChecker.HandleReady)

	// Dashboard API endpoints
	if cfg.HTTP.DashboardEnabled {
		mux.HandleFunc("/api/account", monitor.HandleAccountState)
		mux.HandleFunc("/api/positions", monitor.HandlePositions)
		mux.HandleFunc("/api/pnl", monitor.HandlePnL)
		mux.HandleFunc("/api/balance", monitor.HandleBalance)
		mux.HandleFunc("/api/alerts", monitor.HandleAlerts)
		mux.HandleFunc("/ws", monitor.HandleWebSocket)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("HTTP server starting", zap.Int("port", cfg.HTTP.Port))
	return server.ListenAndServe()
}

func startMetricsServer(cfg *config.Config, logger *zap.Logger) error {
	mux := http.NewServeMux()
	mux.Handle(cfg.Metrics.Path, promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
		Handler: mux,
	}

	logger.Info("Metrics server starting", zap.Int("port", cfg.Metrics.Port))
	return server.ListenAndServe()
}

func init() {
	// Set default viper settings
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/account-monitor")
	viper.AutomaticEnv()
}
