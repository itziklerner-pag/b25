package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b25/services/notification/internal/api"
	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/b25/services/notification/internal/providers"
	"github.com/b25/services/notification/internal/queue"
	"github.com/b25/services/notification/internal/repository"
	"github.com/b25/services/notification/internal/service"
	"github.com/b25/services/notification/internal/templates"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	logger.Info("starting notification service",
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

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}
	logger.Info("database connection established")

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("failed to ping Redis", zap.Error(err))
	}
	logger.Info("Redis connection established")

	// Initialize repositories
	notifRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateRepo := repository.NewTemplateRepository(db)

	// Initialize template engine
	templateEngine := templates.NewTemplateEngine()

	// Load templates from database
	ctx := context.Background()
	dbTemplates, err := templateRepo.ListActive(ctx)
	if err != nil {
		logger.Warn("failed to load templates from database", zap.Error(err))
	} else {
		for _, tmpl := range dbTemplates {
			if err := templateEngine.RegisterTemplate(tmpl); err != nil {
				logger.Error("failed to register template",
					zap.String("template", tmpl.Name),
					zap.Error(err),
				)
			}
		}
		logger.Info("templates loaded", zap.Int("count", templateEngine.GetTemplateCount()))
	}

	// Initialize providers
	providers := initProviders(ctx, cfg, logger)

	// Initialize queue
	q, err := queue.NewAsynqQueue(&cfg.Queue, logger)
	if err != nil {
		logger.Fatal("failed to initialize queue", zap.Error(err))
	}
	defer q.Close()

	// Initialize rate limiter
	rateLimiter := service.NewRedisRateLimiter(redisClient, &cfg.RateLimit)

	// Initialize services
	notificationService := service.NewNotificationService(
		cfg,
		logger,
		notifRepo,
		userRepo,
		templateRepo,
		templateEngine,
		q,
		providers,
		rateLimiter,
	)

	userService := service.NewUserService(cfg, logger, userRepo)
	templateService := service.NewTemplateService(cfg, logger, templateRepo, templateEngine)

	// Initialize queue worker
	taskHandler := queue.NewTaskHandler(notificationService, logger)
	worker := queue.NewWorker(&cfg.Queue, logger, taskHandler)
	worker.RegisterHandlers()

	// Start worker in background
	go func() {
		if err := worker.Start(); err != nil {
			logger.Fatal("failed to start worker", zap.Error(err))
		}
	}()

	// Initialize HTTP router
	router := api.SetupRouter(cfg, logger, notificationService, userService, templateService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		metricsRouter := api.SetupMetricsRouter()
		metricsSrv := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
			Handler: metricsRouter,
		}
		go func() {
			logger.Info("metrics server started", zap.Int("port", cfg.Metrics.Port))
			if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("metrics server error", zap.Error(err))
			}
		}()
	}

	// Start server in background
	go func() {
		logger.Info("HTTP server started", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Start periodic tasks
	go startPeriodicTasks(ctx, notificationService, logger)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down notification service...")

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Stop worker
	worker.Stop()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("notification service stopped")
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

func initProviders(ctx context.Context, cfg *config.Config, logger *zap.Logger) map[models.NotificationChannel]providers.Provider {
	providerMap := make(map[models.NotificationChannel]providers.Provider)

	// Initialize email provider
	if cfg.Email.Provider == "sendgrid" {
		emailProvider, err := providers.NewSendGridProvider(&cfg.Email)
		if err != nil {
			logger.Error("failed to initialize SendGrid provider", zap.Error(err))
		} else {
			providerMap[models.ChannelEmail] = emailProvider
			logger.Info("SendGrid provider initialized")
		}
	}

	// Initialize SMS provider
	if cfg.SMS.Provider == "twilio" {
		smsProvider, err := providers.NewTwilioProvider(&cfg.SMS)
		if err != nil {
			logger.Error("failed to initialize Twilio provider", zap.Error(err))
		} else {
			providerMap[models.ChannelSMS] = smsProvider
			logger.Info("Twilio provider initialized")
		}
	}

	// Initialize push provider
	if cfg.Push.Provider == "fcm" {
		pushProvider, err := providers.NewFCMProvider(ctx, &cfg.Push)
		if err != nil {
			logger.Error("failed to initialize FCM provider", zap.Error(err))
		} else {
			providerMap[models.ChannelPush] = pushProvider
			logger.Info("FCM provider initialized")
		}
	}

	return providerMap
}

func startPeriodicTasks(ctx context.Context, notificationService *service.NotificationService, logger *zap.Logger) {
	// Process scheduled notifications every minute
	scheduledTicker := time.NewTicker(1 * time.Minute)
	defer scheduledTicker.Stop()

	// Process retries every 30 seconds
	retryTicker := time.NewTicker(30 * time.Second)
	defer retryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping periodic tasks")
			return
		case <-scheduledTicker.C:
			if err := notificationService.ProcessScheduledNotifications(ctx); err != nil {
				logger.Error("failed to process scheduled notifications", zap.Error(err))
			}
		case <-retryTicker.C:
			if err := notificationService.ProcessRetries(ctx); err != nil {
				logger.Error("failed to process retries", zap.Error(err))
			}
		}
	}
}
