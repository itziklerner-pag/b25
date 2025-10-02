package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourorg/b25/services/payment/internal/api"
	"github.com/yourorg/b25/services/payment/internal/config"
	"github.com/yourorg/b25/services/payment/internal/database"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/payment"
	"github.com/yourorg/b25/services/payment/internal/repository"
	"github.com/yourorg/b25/services/payment/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize logger
	appLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.Database.URL); err != nil {
		appLogger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize Redis for caching
	redisClient := database.NewRedisClient(cfg.Redis)
	defer redisClient.Close()

	// Initialize repositories
	txRepo := repository.NewTransactionRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	paymentMethodRepo := repository.NewPaymentMethodRepository(db)
	webhookRepo := repository.NewWebhookEventRepository(db)

	// Initialize Stripe client
	stripeClient := payment.NewStripeClient(cfg.Stripe.SecretKey)

	// Initialize services
	paymentService := service.NewPaymentService(
		txRepo,
		paymentMethodRepo,
		stripeClient,
		redisClient,
		appLogger,
	)

	subscriptionService := service.NewSubscriptionService(
		subscriptionRepo,
		invoiceRepo,
		stripeClient,
		appLogger,
	)

	invoiceService := service.NewInvoiceService(
		invoiceRepo,
		txRepo,
		stripeClient,
		appLogger,
	)

	refundService := service.NewRefundService(
		txRepo,
		stripeClient,
		appLogger,
	)

	webhookService := service.NewWebhookService(
		webhookRepo,
		paymentService,
		subscriptionService,
		cfg.Stripe.WebhookSecret,
		appLogger,
	)

	// Initialize HTTP router
	router := setupRouter(cfg)

	// Register API routes
	api.RegisterRoutes(
		router,
		paymentService,
		subscriptionService,
		invoiceService,
		refundService,
		webhookService,
		cfg,
		appLogger,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("Starting payment service", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server exited")
}

func setupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.GinLogger())

	return router
}
