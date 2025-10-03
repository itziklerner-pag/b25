package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/exchange"
)

type Checker struct {
	db        *pgxpool.Pool
	redis     *redis.Client
	wsClient  *exchange.WebSocketClient
	logger    *zap.Logger
	startTime time.Time
}

type HealthStatus struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Uptime  string            `json:"uptime"`
	Checks  map[string]Check  `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func NewChecker(db *pgxpool.Pool, redis *redis.Client, wsClient *exchange.WebSocketClient, logger *zap.Logger) *Checker {
	return &Checker{
		db:        db,
		redis:     redis,
		wsClient:  wsClient,
		logger:    logger,
		startTime: time.Now(),
	}
}

// HandleHealth handles health check requests
func (h *Checker) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := HealthStatus{
		Status:  "healthy",
		Version: "1.0.0",
		Uptime:  time.Since(h.startTime).String(),
		Checks:  make(map[string]Check),
	}

	// Check database
	dbCheck := h.checkDatabase(ctx)
	status.Checks["database"] = dbCheck
	if dbCheck.Status != "ok" {
		status.Status = "unhealthy"
	}

	// Check Redis
	redisCheck := h.checkRedis(ctx)
	status.Checks["redis"] = redisCheck
	if redisCheck.Status != "ok" {
		status.Status = "unhealthy"
	}

	// Check WebSocket
	wsCheck := h.checkWebSocket()
	status.Checks["websocket"] = wsCheck
	if wsCheck.Status != "ok" {
		status.Status = "degraded"
	}

	statusCode := http.StatusOK
	if status.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(status)
}

// HandleReady handles readiness probe requests
func (h *Checker) HandleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check only critical dependencies
	if err := h.db.Ping(ctx); err != nil {
		http.Error(w, "database not ready", http.StatusServiceUnavailable)
		return
	}

	if _, err := h.redis.Ping(ctx).Result(); err != nil {
		http.Error(w, "redis not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func (h *Checker) checkDatabase(ctx context.Context) Check {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return Check{Status: "error", Message: err.Error()}
	}
	return Check{Status: "ok"}
}

func (h *Checker) checkRedis(ctx context.Context) Check {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if _, err := h.redis.Ping(ctx).Result(); err != nil {
		return Check{Status: "error", Message: err.Error()}
	}
	return Check{Status: "ok"}
}

func (h *Checker) checkWebSocket() Check {
	if !h.wsClient.IsConnected() {
		return Check{Status: "error", Message: "WebSocket disconnected"}
	}
	return Check{Status: "ok"}
}
