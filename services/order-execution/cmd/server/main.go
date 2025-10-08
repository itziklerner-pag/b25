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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"

	"github.com/yourusername/b25/services/order-execution/internal/admin"
	"github.com/yourusername/b25/services/order-execution/internal/executor"
	"github.com/yourusername/b25/services/order-execution/internal/health"
	pb "github.com/yourusername/b25/services/order-execution/proto"
)

const (
	version = "1.0.0"
)

// Config holds application configuration
type Config struct {
	Server struct {
		GRPCPort  int    `yaml:"grpc_port"`
		HTTPPort  int    `yaml:"http_port"`
		Host      string `yaml:"host"`
	} `yaml:"server"`

	Exchange struct {
		APIKey    string `yaml:"api_key"`
		SecretKey string `yaml:"secret_key"`
		Testnet   bool   `yaml:"testnet"`
	} `yaml:"exchange"`

	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	NATS struct {
		Address string `yaml:"address"`
	} `yaml:"nats"`

	RateLimit struct {
		RequestsPerSecond int `yaml:"requests_per_second"`
		Burst             int `yaml:"burst"`
	} `yaml:"rate_limit"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

func main() {
	// Initialize logger
	logger, err := initLogger("info", "json")
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting order execution service", zap.String("version", version))

	// Set version in admin package
	admin.SetVersion(version)

	// Load configuration
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Override logger with configured level
	logger, err = initLogger(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		logger.Fatal("failed to reinitialize logger", zap.Error(err))
	}

	// Create executor
	executorCfg := executor.Config{
		ExchangeAPIKey:    cfg.Exchange.APIKey,
		ExchangeSecretKey: cfg.Exchange.SecretKey,
		TestnetMode:       cfg.Exchange.Testnet,
		RedisAddr:         cfg.Redis.Address,
		RedisPassword:     cfg.Redis.Password,
		RedisDB:           cfg.Redis.DB,
		NATSAddr:          cfg.NATS.Address,
		RateLimitRPS:      cfg.RateLimit.RequestsPerSecond,
		RateLimitBurst:    cfg.RateLimit.Burst,
	}

	orderExecutor, err := executor.NewOrderExecutor(executorCfg, logger)
	if err != nil {
		logger.Fatal("failed to create order executor", zap.Error(err))
	}
	defer orderExecutor.Close()

	// Start gRPC server
	grpcServer := startGRPCServer(cfg, orderExecutor, logger)

	// Start HTTP server (health + metrics + admin)
	httpServer := startHTTPServer(cfg, orderExecutor, logger)

	// Wait for shutdown signal
	waitForShutdown(logger, grpcServer, httpServer)

	logger.Info("service stopped")
}

// loadConfig loads configuration from file
func loadConfig(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Use environment variables or defaults
		return &Config{
			Server: struct {
				GRPCPort int    `yaml:"grpc_port"`
				HTTPPort int    `yaml:"http_port"`
				Host     string `yaml:"host"`
			}{
				GRPCPort: 50051,
				HTTPPort: 9091,
				Host:     "0.0.0.0",
			},
			Exchange: struct {
				APIKey    string `yaml:"api_key"`
				SecretKey string `yaml:"secret_key"`
				Testnet   bool   `yaml:"testnet"`
			}{
				APIKey:    getEnv("BINANCE_API_KEY", ""),
				SecretKey: getEnv("BINANCE_SECRET_KEY", ""),
				Testnet:   getEnv("BINANCE_TESTNET", "true") == "true",
			},
			Redis: struct {
				Address  string `yaml:"address"`
				Password string `yaml:"password"`
				DB       int    `yaml:"db"`
			}{
				Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       0,
			},
			NATS: struct {
				Address string `yaml:"address"`
			}{
				Address: getEnv("NATS_ADDRESS", "nats://localhost:4222"),
			},
			RateLimit: struct {
				RequestsPerSecond int `yaml:"requests_per_second"`
				Burst             int `yaml:"burst"`
			}{
				RequestsPerSecond: 10,
				Burst:             20,
			},
			Logging: struct {
				Level  string `yaml:"level"`
				Format string `yaml:"format"`
			}{
				Level:  getEnv("LOG_LEVEL", "info"),
				Format: getEnv("LOG_FORMAT", "json"),
			},
		}, nil
	}

	// Load from file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables in the YAML content
	expandedData := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
		cfg.Exchange.APIKey = apiKey
	}
	if secretKey := os.Getenv("BINANCE_SECRET_KEY"); secretKey != "" {
		cfg.Exchange.SecretKey = secretKey
	}

	return &cfg, nil
}

// initLogger initializes the logger
func initLogger(level, format string) (*zap.Logger, error) {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}

// startGRPCServer starts the gRPC server
func startGRPCServer(cfg *Config, orderExecutor *executor.OrderExecutor, logger *zap.Logger) *grpc.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	grpcService := executor.NewGRPCServer(orderExecutor)
	pb.RegisterOrderServiceServer(grpcServer, grpcService)

	// Enable reflection for development
	reflection.Register(grpcServer)

	go func() {
		logger.Info("starting gRPC server", zap.String("address", addr))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	return grpcServer
}

// startHTTPServer starts the HTTP server for health checks, metrics, and admin
func startHTTPServer(cfg *Config, orderExecutor *executor.OrderExecutor, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()

	// Create health checker
	healthChecker := health.NewHealthChecker(
		orderExecutor.GetRedisClient(),
		orderExecutor.GetNATSConn(),
		version,
	)

	// Create admin handler
	adminConfig := &admin.Config{
		HTTPPort:       cfg.Server.HTTPPort,
		GRPCPort:       cfg.Server.GRPCPort,
		TestnetMode:    cfg.Exchange.Testnet,
		RateLimitRPS:   cfg.RateLimit.RequestsPerSecond,
		RateLimitBurst: cfg.RateLimit.Burst,
	}
	adminHandler := admin.NewHandler(orderExecutor, logger, adminConfig)

	// Register endpoints
	mux.HandleFunc("/", adminHandler.AdminPageHandler()) // Show admin page by default
	mux.HandleFunc("/admin", adminHandler.AdminPageHandler())
	mux.HandleFunc("/api/service-info", adminHandler.ServiceInfoHandler())
	mux.HandleFunc("/api/endpoints", adminHandler.EndpointsHandler())
	mux.HandleFunc("/health", healthChecker.HTTPHandler())
	mux.HandleFunc("/health/ready", healthChecker.ReadinessHandler())
	mux.HandleFunc("/health/live", healthChecker.LivenessHandler())
	mux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("address", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve HTTP", zap.Error(err))
		}
	}()

	return server
}

// rootHandler returns a simple service info handler for root path
func rootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service":"Order Execution Service","version":"%s","status":"running"}`, version)
	}
}

// waitForShutdown waits for shutdown signal and performs graceful shutdown
func waitForShutdown(logger *zap.Logger, grpcServer *grpc.Server, httpServer *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	logger.Info("shutting down HTTP server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Shutdown gRPC server
	logger.Info("shutting down gRPC server")
	grpcServer.GracefulStop()
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
