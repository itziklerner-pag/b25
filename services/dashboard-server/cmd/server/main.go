package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/yourusername/b25/services/dashboard-server/internal/aggregator"
	"github.com/yourusername/b25/services/dashboard-server/internal/broadcaster"
	"github.com/yourusername/b25/services/dashboard-server/internal/server"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = logger

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	logger.Info().
		Str("version", "1.0.0").
		Str("log_level", level.String()).
		Msg("Starting Dashboard Server Service")

	// Create state aggregator with service connections
	aggConfig := aggregator.Config{
		RedisURL:            config.RedisURL,
		OrderServiceGRPC:    config.OrderServiceGRPC,
		StrategyServiceHTTP: config.StrategyServiceHTTP,
		AccountServiceGRPC:  config.AccountServiceGRPC,
	}
	stateAggregator := aggregator.NewAggregator(logger, aggConfig)
	if err := stateAggregator.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start state aggregator")
	}
	defer stateAggregator.Stop()

	// Create broadcaster
	broadcaster := broadcaster.NewBroadcaster(logger, stateAggregator)
	broadcaster.Start()
	defer broadcaster.Stop()

	// Create WebSocket server
	serverConfig := server.Config{
		AllowedOrigins: config.AllowedOrigins,
		APIKey:         config.APIKey,
	}
	wsServer := server.NewServer(logger, stateAggregator, broadcaster, serverConfig)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsServer.HandleWebSocket)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/debug", wsServer.HandleDebug)
	mux.HandleFunc("/api/v1/history", wsServer.HandleHistory)
	mux.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		sig := <-sigCh

		logger.Info().
			Str("signal", sig.String()).
			Msg("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("Server shutdown error")
		}
	}()

	logger.Info().
		Int("port", config.Port).
		Str("order_service", config.OrderServiceGRPC).
		Str("strategy_service", config.StrategyServiceHTTP).
		Msg("Dashboard Server started successfully")

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Server failed")
	}

	logger.Info().Msg("Server stopped gracefully")
}

type Config struct {
	Port                int
	LogLevel            string
	RedisURL            string
	OrderServiceGRPC    string
	StrategyServiceHTTP string
	AccountServiceGRPC  string
	AllowedOrigins      []string
	APIKey              string // Optional API key for authentication
}

func loadConfig() (*Config, error) {
	// Try to load from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/dashboard-server/")

	// Set defaults
	viper.SetDefault("server.port", 8086)
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("redis.url", "localhost:6379")
	viper.SetDefault("backend_services.order_execution.url", "localhost:50051")
	viper.SetDefault("backend_services.strategy_engine.url", "http://localhost:8082")
	viper.SetDefault("backend_services.account_monitor.url", "localhost:50055")
	viper.SetDefault("websocket.allowed_origins", []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:8080",
	})

	// Read config file (ignore error if doesn't exist)
	_ = viper.ReadInConfig()

	viper.SetEnvPrefix("DASHBOARD")
	viper.AutomaticEnv()

	return &Config{
		Port:                viper.GetInt("server.port"),
		LogLevel:            viper.GetString("server.log_level"),
		RedisURL:            viper.GetString("redis.url"),
		OrderServiceGRPC:    viper.GetString("backend_services.order_execution.url"),
		StrategyServiceHTTP: viper.GetString("backend_services.strategy_engine.url"),
		AccountServiceGRPC:  viper.GetString("backend_services.account_monitor.url"),
		AllowedOrigins:      viper.GetStringSlice("websocket.allowed_origins"),
		APIKey:              viper.GetString("security.api_key"),
	}, nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"dashboard-server"}`))
}

func loggingMiddleware(logger zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Dur("duration", time.Since(start)).
			Msg("HTTP request")
	})
}
