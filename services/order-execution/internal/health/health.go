package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
)

// Status represents health check status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Check represents a single health check result
type Check struct {
	Name     string        `json:"name"`
	Status   Status        `json:"status"`
	Message  string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration_ms"`
}

// Response represents the overall health check response
type Response struct {
	Status    Status           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Checks    map[string]Check `json:"checks"`
	Version   string           `json:"version,omitempty"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	redisClient *redis.Client
	natsConn    *nats.Conn
	version     string
	mu          sync.RWMutex
	lastCheck   *Response
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(redisClient *redis.Client, natsConn *nats.Conn, version string) *HealthChecker {
	return &HealthChecker{
		redisClient: redisClient,
		natsConn:    natsConn,
		version:     version,
	}
}

// Check performs all health checks
func (h *HealthChecker) Check(ctx context.Context) *Response {
	response := &Response{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Checks:    make(map[string]Check),
		Version:   h.version,
	}

	// Run checks concurrently
	var wg sync.WaitGroup
	checkResults := make(chan Check, 3)

	// Redis check
	wg.Add(1)
	go func() {
		defer wg.Done()
		checkResults <- h.checkRedis(ctx)
	}()

	// NATS check
	wg.Add(1)
	go func() {
		defer wg.Done()
		checkResults <- h.checkNATS()
	}()

	// System check
	wg.Add(1)
	go func() {
		defer wg.Done()
		checkResults <- h.checkSystem()
	}()

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(checkResults)
	}()

	// Collect results
	for check := range checkResults {
		response.Checks[check.Name] = check

		// Update overall status
		if check.Status == StatusUnhealthy {
			response.Status = StatusUnhealthy
		} else if check.Status == StatusDegraded && response.Status == StatusHealthy {
			response.Status = StatusDegraded
		}
	}

	// Cache result
	h.mu.Lock()
	h.lastCheck = response
	h.mu.Unlock()

	return response
}

// checkRedis checks Redis connectivity
func (h *HealthChecker) checkRedis(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name: "redis",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := h.redisClient.Ping(timeoutCtx).Err()
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = err.Error()
	} else {
		check.Status = StatusHealthy
	}

	return check
}

// checkNATS checks NATS connectivity
func (h *HealthChecker) checkNATS() Check {
	start := time.Now()
	check := Check{
		Name: "nats",
	}

	if h.natsConn == nil {
		check.Status = StatusUnhealthy
		check.Message = "NATS connection not initialized"
		check.Duration = time.Since(start)
		return check
	}

	if !h.natsConn.IsConnected() {
		check.Status = StatusUnhealthy
		check.Message = "NATS not connected"
	} else {
		check.Status = StatusHealthy
	}

	check.Duration = time.Since(start)
	return check
}

// checkSystem checks system resources
func (h *HealthChecker) checkSystem() Check {
	start := time.Now()
	check := Check{
		Name:     "system",
		Status:   StatusHealthy,
		Duration: time.Since(start),
	}

	// Add system checks like memory, goroutines, etc.
	// For now, always healthy

	return check
}

// GetLastCheck returns the last health check result
func (h *HealthChecker) GetLastCheck() *Response {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastCheck
}

// HTTPHandler returns an HTTP handler for health checks
func (h *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Set HTTP status based on health
		switch response.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // Still OK, but degraded
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(response)
	}
}

// ReadinessHandler returns a readiness probe handler
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.Check(ctx)

		// Readiness is stricter - must be fully healthy
		if response.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("not ready"))
		}
	}
}

// LivenessHandler returns a liveness probe handler
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Liveness is very simple - just check if we can respond
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alive"))
	}
}
