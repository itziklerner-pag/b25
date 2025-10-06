package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b25/strategy-engine/internal/config"
	"github.com/b25/strategy-engine/internal/engine"
	"github.com/b25/strategy-engine/pkg/logger"
	"github.com/b25/strategy-engine/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultConfigPath = "config.yaml"
	shutdownTimeout   = 10 * time.Second
)

func main() {
	// Load configuration
	configPath := getConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Strategy Engine",
		"version", "1.0.0",
		"port", cfg.Server.Port,
		"mode", cfg.Engine.Mode,
	)

	// Initialize metrics
	metricsCollector := metrics.New(cfg.Metrics.Namespace)

	// Create strategy engine
	eng, err := engine.New(cfg, log, metricsCollector)
	if err != nil {
		log.Fatal("Failed to create strategy engine", "error", err)
	}

	// Start the engine
	if err := eng.Start(); err != nil {
		log.Fatal("Failed to start strategy engine", "error", err)
	}

	// Start HTTP server for health checks and metrics
	srv := startHTTPServer(cfg, log, eng)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down strategy engine...")

	// Stop the engine
	if err := eng.Stop(); err != nil {
		log.Error("Error stopping engine", "error", err)
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Strategy engine exited")
}

// setCORSHeaders sets CORS headers for health endpoints
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
}

// authMiddleware checks API key if authentication is enabled
func authMiddleware(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for OPTIONS requests
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}

		// Skip auth if not enabled
		if !cfg.Server.EnableAuth {
			next(w, r)
			return
		}

		// Check API key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != cfg.Server.APIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized","message":"Invalid or missing API key"}`))
			return
		}

		next(w, r)
	}
}

func startHTTPServer(cfg *config.Config, log *logger.Logger, eng *engine.Engine) *http.Server {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		setCORSHeaders(w)

		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"strategy-engine"}`))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		setCORSHeaders(w)

		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// Metrics endpoint (Prometheus)
	if cfg.Metrics.Enabled {
		mux.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	// Status endpoint (protected if auth enabled)
	mux.HandleFunc("/status", authMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		setCORSHeaders(w)

		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		metrics := eng.GetMetrics()

		// Simple JSON serialization
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"mode":"%s","active_strategies":%d,"signal_queue_size":%d}`,
			metrics["mode"],
			metrics["active_strategies"],
			metrics["signal_queue_size"],
		)
	}))

	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        mux,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		log.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", "error", err)
		}
	}()

	return srv
}

func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return defaultConfigPath
}
