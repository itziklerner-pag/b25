package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/b25/services/media/internal/api"
	"github.com/yourorg/b25/services/media/internal/cache"
	"github.com/yourorg/b25/services/media/internal/config"
	"github.com/yourorg/b25/services/media/internal/database"
	"github.com/yourorg/b25/services/media/internal/processing"
	"github.com/yourorg/b25/services/media/internal/quota"
	"github.com/yourorg/b25/services/media/internal/security"
	"github.com/yourorg/b25/services/media/internal/storage"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg)

	log.Info("Starting Media Service...")

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Initialize storage backend
	storageBackend, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize image processor
	imageProcessor := processing.NewImageProcessor(cfg.Processing.Image)

	// Initialize video processor
	videoProcessor := processing.NewVideoProcessor(cfg.Processing.Video)

	// Initialize security scanner
	var securityScanner security.Scanner
	if cfg.Security.EnableVirusScan {
		securityScanner, err = security.NewClamAVScanner(cfg.Security.ClamAV)
		if err != nil {
			log.Warnf("Failed to initialize virus scanner: %v", err)
			securityScanner = security.NewNoOpScanner()
		}
	} else {
		securityScanner = security.NewNoOpScanner()
	}

	// Initialize quota manager
	quotaManager := quota.NewQuotaManager(db, cfg.Quota)
	if cfg.Quota.Enabled {
		go quotaManager.StartPeriodicCheck(context.Background())
	}

	// Initialize API server
	apiServer := api.NewServer(api.ServerConfig{
		Config:           cfg,
		DB:               db,
		Cache:            redisCache,
		Storage:          storageBackend,
		ImageProcessor:   imageProcessor,
		VideoProcessor:   videoProcessor,
		SecurityScanner:  securityScanner,
		QuotaManager:     quotaManager,
	})

	// Setup HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      apiServer.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Media Service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Media Service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Media Service stopped")
}

func setupLogging(cfg *config.Config) {
	// Set log level
	level, err := log.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

	// Set log format
	if cfg.Logging.Format == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set output
	if cfg.Logging.Output == "file" && cfg.Logging.FilePath != "" {
		file, err := os.OpenFile(cfg.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Warnf("Failed to open log file, using stdout: %v", err)
		}
	}
}
