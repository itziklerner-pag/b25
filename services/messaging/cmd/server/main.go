package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b25/services/messaging/internal/api"
	"github.com/b25/services/messaging/internal/auth"
	"github.com/b25/services/messaging/internal/repository"
	"github.com/b25/services/messaging/internal/service"
	"github.com/b25/services/messaging/internal/websocket"
	"github.com/b25/services/messaging/pkg/config"
	"github.com/b25/services/messaging/pkg/logger"
	"github.com/b25/services/messaging/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func main() {
	// Load configuration
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Msg("Starting Messaging Service")

	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewPostgresRepository(db)

	// Initialize WebSocket hub
	hub := websocket.NewHub(log)
	go hub.Run()

	// Initialize service
	svc := service.NewMessagingService(repo, hub, log)

	// Initialize WebSocket handler
	wsHandler := websocket.NewHandler(svc, log)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)

	// Initialize API handler
	apiHandler := api.NewHandler(svc, log)

	// Setup router
	router := setupRouter(apiHandler, wsHandler, jwtManager, hub, log)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Server.HTTPPort).Msg("Server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hub.Shutdown()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func setupRouter(
	apiHandler *api.Handler,
	wsHandler *websocket.Handler,
	jwtManager *auth.JWTManager,
	hub *websocket.Hub,
	log zerolog.Logger,
) *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.CORS)
	router.Use(middleware.Logging(log))

	// Health check (no auth required)
	router.HandleFunc("/health", apiHandler.Health).Methods("GET")

	// Metrics (no auth required)
	router.Handle("/metrics", promhttp.Handler())

	// WebSocket endpoint
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, jwtManager, hub, wsHandler, log)
	})

	// API routes (require authentication)
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(jwtManager.Middleware)

	// Conversations
	api.HandleFunc("/conversations", apiHandler.CreateConversation).Methods("POST")
	api.HandleFunc("/conversations", apiHandler.ListConversations).Methods("GET")
	api.HandleFunc("/conversations/{id}", apiHandler.GetConversation).Methods("GET")
	api.HandleFunc("/conversations/{id}", apiHandler.UpdateConversation).Methods("PUT")
	api.HandleFunc("/conversations/{id}/members", apiHandler.AddMember).Methods("POST")
	api.HandleFunc("/conversations/{id}/members/{userId}", apiHandler.RemoveMember).Methods("DELETE")

	// Messages
	api.HandleFunc("/conversations/{id}/messages", apiHandler.SendMessage).Methods("POST")
	api.HandleFunc("/conversations/{id}/messages", apiHandler.GetMessages).Methods("GET")
	api.HandleFunc("/messages/{id}", apiHandler.EditMessage).Methods("PUT")
	api.HandleFunc("/messages/{id}", apiHandler.DeleteMessage).Methods("DELETE")
	api.HandleFunc("/messages/{id}/reactions", apiHandler.AddReaction).Methods("POST")
	api.HandleFunc("/messages/{id}/reactions/{emoji}", apiHandler.RemoveReaction).Methods("DELETE")
	api.HandleFunc("/messages/{id}/read", apiHandler.MarkAsRead).Methods("POST")

	// Search
	api.HandleFunc("/search/messages", apiHandler.SearchMessages).Methods("GET")

	// Users
	api.HandleFunc("/users/online", apiHandler.GetOnlineUsers).Methods("GET")

	return router
}

func handleWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	jwtManager *auth.JWTManager,
	hub *websocket.Hub,
	handler *websocket.Handler,
	log zerolog.Logger,
) {
	// Extract and verify JWT from query parameter
	claims, err := jwtManager.ExtractTokenFromQuery(r)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket authentication failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO: Implement proper origin check
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Create client
	client := websocket.NewClient(claims.UserID, conn, hub, handler, log)

	// Register client
	hub.Register(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
