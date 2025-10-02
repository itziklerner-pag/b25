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

	"github.com/b25/analytics/internal/aggregation"
	"github.com/b25/analytics/internal/api"
	"github.com/b25/analytics/internal/cache"
	"github.com/b25/analytics/internal/config"
	"github.com/b25/analytics/internal/ingestion"
	"github.com/b25/analytics/internal/logger"
	internalmetrics "github.com/b25/analytics/internal/metrics"
	"github.com/b25/analytics/internal/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
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
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Analytics Service",
		zap.String("version", version),
		zap.String("config", *configPath),
	)

	// Initialize database repository
	repo, err := repository.NewRepository(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize repository", zap.Error(err))
	}
	defer repo.Close()
	log.Info("Database connection established")

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(&cfg.Redis, cfg.Analytics.Query.CacheTTL, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis cache", zap.Error(err))
	}
	defer redisCache.Close()
	log.Info("Redis cache initialized")

	// Initialize Prometheus metrics
	prometheusMetrics := internalmetrics.NewMetrics()
	log.Info("Prometheus metrics initialized")

	// Initialize event consumer
	consumer := ingestion.NewConsumer(&cfg.Kafka, &cfg.Analytics.Ingestion, repo, log)
	if err := consumer.Start(); err != nil {
		log.Fatal("Failed to start event consumer", zap.Error(err))
	}
	log.Info("Event consumer started")

	// Initialize aggregation engine
	aggregationEngine, err := aggregation.NewEngine(&cfg.Analytics.Aggregation, repo, log)
	if err != nil {
		log.Fatal("Failed to create aggregation engine", zap.Error(err))
	}
	if err := aggregationEngine.Start(); err != nil {
		log.Fatal("Failed to start aggregation engine", zap.Error(err))
	}
	log.Info("Aggregation engine started")

	// Initialize API handler and router
	handler := api.NewHandler(repo, redisCache, log)
	router := api.SetupRouter(cfg, handler, log)

	// HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Info("Starting HTTP server",
			zap.String("address", httpServer.Addr),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Prometheus metrics server
	if cfg.Metrics.Enabled {
		metricsServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
			Handler: promhttp.Handler(),
		}

		go func() {
			log.Info("Starting Prometheus metrics server",
				zap.String("address", metricsServer.Addr),
				zap.String("path", cfg.Metrics.Path),
			)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("Metrics server error", zap.Error(err))
			}
		}()
	}

	// Start background jobs
	go runBackgroundJobs(repo, redisCache, log, cfg)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Stop consumer
	if err := consumer.Stop(); err != nil {
		log.Error("Error stopping consumer", zap.Error(err))
	}

	// Stop aggregation engine
	if err := aggregationEngine.Stop(); err != nil {
		log.Error("Error stopping aggregation engine", zap.Error(err))
	}

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error", zap.Error(err))
	}

	log.Info("Service stopped gracefully")
}

// runBackgroundJobs runs periodic background tasks
func runBackgroundJobs(repo *repository.Repository, cache *cache.RedisCache, log *zap.Logger, cfg *config.Config) {
	// Refresh dashboard metrics periodically
	dashboardTicker := time.NewTicker(10 * time.Second)
	defer dashboardTicker.Stop()

	// Cleanup old data daily
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-dashboardTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := repo.RefreshDashboardMaterializedView(ctx); err != nil {
				log.Error("Failed to refresh dashboard metrics", zap.Error(err))
			}
			cancel()

		case <-cleanupTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			if err := repo.CleanupOldData(ctx); err != nil {
				log.Error("Failed to cleanup old data", zap.Error(err))
			} else {
				log.Info("Old data cleanup completed")
			}
			cancel()
		}
	}
}
