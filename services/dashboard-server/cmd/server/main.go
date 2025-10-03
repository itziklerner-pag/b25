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

	// Create state aggregator
	stateAggregator := aggregator.NewAggregator(logger, config.RedisURL)
	if err := stateAggregator.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start state aggregator")
	}
	defer stateAggregator.Stop()

	// Create broadcaster
	broadcaster := broadcaster.NewBroadcaster(logger, stateAggregator)
	broadcaster.Start()
	defer broadcaster.Stop()

	// Create WebSocket server
	wsServer := server.NewServer(logger, stateAggregator, broadcaster)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsServer.HandleWebSocket)
	mux.HandleFunc("/health", handleHealth)
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
		Msg("Dashboard Server started successfully")

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Server failed")
	}

	logger.Info().Msg("Server stopped gracefully")
}

type Config struct {
	Port     int
	LogLevel string
	RedisURL string
}

func loadConfig() (*Config, error) {
	viper.SetDefault("port", 8080)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("redis_url", "localhost:6379")

	viper.SetEnvPrefix("DASHBOARD")
	viper.AutomaticEnv()

	return &Config{
		Port:     viper.GetInt("port"),
		LogLevel: viper.GetString("log_level"),
		RedisURL: viper.GetString("redis_url"),
	}, nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
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
