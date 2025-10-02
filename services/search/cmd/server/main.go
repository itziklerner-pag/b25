package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/yourorg/b25/services/search/internal/analytics"
	"github.com/yourorg/b25/services/search/internal/api"
	"github.com/yourorg/b25/services/search/internal/config"
	"github.com/yourorg/b25/services/search/internal/indexer"
	"github.com/yourorg/b25/services/search/internal/search"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger, err := setupLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Search Service",
		zap.String("version", version),
		zap.Int("port", cfg.Server.Port),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Elasticsearch client
	esClient, err := search.NewElasticsearchClient(&cfg.Elasticsearch, logger)
	if err != nil {
		logger.Fatal("Failed to create Elasticsearch client", zap.Error(err))
	}

	// Create indices
	logger.Info("Creating Elasticsearch indices...")
	if err := esClient.CreateIndices(ctx); err != nil {
		logger.Fatal("Failed to create indices", zap.Error(err))
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Info("Connected to Redis", zap.String("address", cfg.Redis.Address))

	// Initialize NATS connection
	natsConn, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()
	logger.Info("Connected to NATS", zap.String("url", cfg.NATS.URL))

	// Initialize analytics
	analyticsTracker := analytics.NewAnalytics(redisClient, &cfg.Analytics, cfg.Redis.AnalyticsTTL, logger)
	analyticsTracker.StartCleanupWorker(ctx)

	// Initialize indexer
	indexerService := indexer.NewIndexer(esClient, natsConn, &cfg.Indexer, &cfg.NATS.Subjects, logger)
	if err := indexerService.Start(); err != nil {
		logger.Fatal("Failed to start indexer", zap.Error(err))
	}
	defer indexerService.Stop()

	// Initialize API handler
	handler := api.NewHandler(esClient, analyticsTracker, &cfg.Search, logger)

	// Setup HTTP router
	router := api.SetupRouter(handler, cfg, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start metrics server if enabled
	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		metricsRouter := api.SetupMetricsRouter()
		metricsServer = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port),
			Handler: metricsRouter,
		}

		go func() {
			logger.Info("Starting metrics server", zap.Int("port", cfg.Metrics.Port))
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Metrics server error", zap.Error(err))
			}
		}()
	}

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	if metricsServer != nil {
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Metrics server forced to shutdown", zap.Error(err))
		}
	}

	logger.Info("Server stopped")
}

// setupLogger creates a configured logger
func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	// Determine log level
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Determine encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Determine output
	var writer zapcore.WriteSyncer
	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writer = zapcore.AddSync(file)
	} else {
		writer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, writer, level)

	// Build logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
