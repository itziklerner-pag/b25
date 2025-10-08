package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
)

const version = "1.0.0"

var startTime = time.Now()

// Handler manages admin endpoints
type Handler struct {
	config *config.Config
	logger *logger.Logger
}

// NewHandler creates a new admin handler
func NewHandler(cfg *config.Config, log *logger.Logger) *Handler {
	return &Handler{
		config: cfg,
		logger: log,
	}
}

// ServiceInfo represents service metadata
type ServiceInfo struct {
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Uptime      string                 `json:"uptime"`
	Port        int                    `json:"port"`
	Mode        string                 `json:"mode"`
	GoVersion   string                 `json:"go_version"`
	NumCPU      int                    `json:"num_cpu"`
	Goroutines  int                    `json:"goroutines"`
	StartTime   time.Time              `json:"start_time"`
	CurrentTime time.Time              `json:"current_time"`
	Config      map[string]interface{} `json:"config"`
}

// HandleAdminPage serves the admin HTML page
func (h *Handler) HandleAdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(adminPageHTML))
}

// HandleServiceInfo returns service metadata
func (h *Handler) HandleServiceInfo(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	info := ServiceInfo{
		Service:     "API Gateway",
		Version:     version,
		Uptime:      formatDuration(uptime),
		Port:        h.config.Server.Port,
		Mode:        h.config.Server.Mode,
		GoVersion:   runtime.Version(),
		NumCPU:      runtime.NumCPU(),
		Goroutines:  runtime.NumGoroutine(),
		StartTime:   startTime,
		CurrentTime: time.Now(),
		Config:      h.buildConfigInfo(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// buildConfigInfo creates a sanitized config map for display
func (h *Handler) buildConfigInfo() map[string]interface{} {
	return map[string]interface{}{
		"services": map[string]interface{}{
			"market_data": map[string]interface{}{
				"url":     h.config.Services.MarketData.URL,
				"timeout": h.config.Services.MarketData.Timeout.String(),
			},
			"order_execution": map[string]interface{}{
				"url":     h.config.Services.OrderExecution.URL,
				"timeout": h.config.Services.OrderExecution.Timeout.String(),
			},
			"strategy_engine": map[string]interface{}{
				"url":     h.config.Services.StrategyEngine.URL,
				"timeout": h.config.Services.StrategyEngine.Timeout.String(),
			},
			"account_monitor": map[string]interface{}{
				"url":     h.config.Services.AccountMonitor.URL,
				"timeout": h.config.Services.AccountMonitor.Timeout.String(),
			},
			"dashboard_server": map[string]interface{}{
				"url":     h.config.Services.DashboardServer.URL,
				"timeout": h.config.Services.DashboardServer.Timeout.String(),
			},
			"risk_manager": map[string]interface{}{
				"url":     h.config.Services.RiskManager.URL,
				"timeout": h.config.Services.RiskManager.Timeout.String(),
			},
			"configuration": map[string]interface{}{
				"url":     h.config.Services.Configuration.URL,
				"timeout": h.config.Services.Configuration.Timeout.String(),
			},
		},
		"auth": map[string]interface{}{
			"enabled":       h.config.Auth.Enabled,
			"jwt_expiry":    h.config.Auth.JWTExpiry.String(),
			"api_key_count": len(h.config.Auth.APIKeys),
		},
		"rate_limit": map[string]interface{}{
			"enabled":               h.config.RateLimit.Enabled,
			"global_rps":            h.config.RateLimit.Global.RequestsPerSecond,
			"per_ip_rpm":            h.config.RateLimit.PerIP.RequestsPerMinute,
			"endpoint_limits_count": len(h.config.RateLimit.Endpoints),
		},
		"cors": map[string]interface{}{
			"enabled":           h.config.CORS.Enabled,
			"allowed_origins":   h.config.CORS.AllowedOrigins,
			"allowed_methods":   h.config.CORS.AllowedMethods,
			"allow_credentials": h.config.CORS.AllowCredentials,
		},
		"circuit_breaker": map[string]interface{}{
			"enabled":      h.config.CircuitBreaker.Enabled,
			"max_requests": h.config.CircuitBreaker.MaxRequests,
			"interval":     h.config.CircuitBreaker.Interval.String(),
			"timeout":      h.config.CircuitBreaker.Timeout.String(),
		},
		"cache": map[string]interface{}{
			"enabled":     h.config.Cache.Enabled,
			"default_ttl": h.config.Cache.DefaultTTL.String(),
			"redis_url":   maskRedisURL(h.config.Cache.RedisURL),
		},
		"websocket": map[string]interface{}{
			"enabled":         h.config.WebSocket.Enabled,
			"max_connections": h.config.WebSocket.MaxConnections,
			"ping_interval":   h.config.WebSocket.PingInterval.String(),
		},
		"features": map[string]interface{}{
			"tracing":       h.config.Features.EnableTracing,
			"compression":   h.config.Features.EnableCompression,
			"request_id":    h.config.Features.EnableRequestID,
			"access_log":    h.config.Features.EnableAccessLog,
			"error_details": h.config.Features.EnableErrorDetails,
		},
	}
}

// maskRedisURL masks sensitive parts of Redis URL
func maskRedisURL(url string) string {
	if url == "" {
		return ""
	}
	// Simple masking - in production you'd want more sophisticated handling
	return "redis://***:***@***"
}

// formatDuration formats a duration into human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return formatWithUnits(days, hours, minutes, seconds, "d", "h", "m", "s")
	}
	if hours > 0 {
		return formatWithUnits(hours, minutes, seconds, 0, "h", "m", "s", "")
	}
	if minutes > 0 {
		return formatWithUnits(minutes, seconds, 0, 0, "m", "s", "", "")
	}
	return formatWithUnits(seconds, 0, 0, 0, "s", "", "", "")
}

func formatWithUnits(a, b, c, d int, ua, ub, uc, ud string) string {
	result := ""
	if a > 0 && ua != "" {
		result += formatUnit(a, ua)
	}
	if b > 0 && ub != "" {
		if result != "" {
			result += " "
		}
		result += formatUnit(b, ub)
	}
	if c > 0 && uc != "" {
		if result != "" {
			result += " "
		}
		result += formatUnit(c, uc)
	}
	if d > 0 && ud != "" {
		if result != "" {
			result += " "
		}
		result += formatUnit(d, ud)
	}
	return result
}

func formatUnit(value int, unit string) string {
	return fmt.Sprintf("%d%s", value, unit)
}
