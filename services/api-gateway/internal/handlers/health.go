package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/b25/api-gateway/internal/breaker"
	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	config  *config.Config
	log     *logger.Logger
	breaker *breaker.Manager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config, log *logger.Logger, b *breaker.Manager) *HealthHandler {
	return &HealthHandler{
		config:  cfg,
		log:     log,
		breaker: b,
	}
}

// setCORSHeaders sets CORS headers for health endpoints
func setCORSHeaders(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Health returns the health status of the gateway
func (h *HealthHandler) Health(c *gin.Context) {
	// Set CORS headers
	setCORSHeaders(c)

	if !h.config.Health.Enabled {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	response := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Check services if enabled
	if h.config.Health.CheckServices {
		services := h.checkServices()
		response["services"] = services

		// Set overall status based on services
		allHealthy := true
		for _, status := range services {
			if statusMap, ok := status.(gin.H); ok {
				if statusMap["status"] != "healthy" {
					allHealthy = false
					break
				}
			}
		}

		if !allHealthy {
			response["status"] = "degraded"
		}
	}

	c.JSON(http.StatusOK, response)
}

// Liveness returns liveness probe status
func (h *HealthHandler) Liveness(c *gin.Context) {
	// Set CORS headers
	setCORSHeaders(c)

	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Readiness returns readiness probe status
func (h *HealthHandler) Readiness(c *gin.Context) {
	// Set CORS headers
	setCORSHeaders(c)

	// Check if gateway is ready to serve traffic
	ready := true
	reasons := make([]string, 0)

	// Check circuit breakers
	if h.breaker != nil {
		services := []string{
			"market_data",
			"order_execution",
			"strategy_engine",
			"account_monitor",
			"dashboard_server",
			"risk_manager",
			"configuration",
		}

		for _, service := range services {
			state := h.breaker.GetState(service)
			if state == gobreaker.StateOpen {
				ready = false
				reasons = append(reasons, "circuit breaker open for "+service)
			}
		}
	}

	if ready {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"reasons": reasons,
		})
	}
}

// checkServices checks the health of backend services
func (h *HealthHandler) checkServices() map[string]interface{} {
	services := map[string]string{
		"market_data":      h.config.Services.MarketData.URL,
		"order_execution":  h.config.Services.OrderExecution.URL,
		"strategy_engine":  h.config.Services.StrategyEngine.URL,
		"account_monitor":  h.config.Services.AccountMonitor.URL,
		"dashboard_server": h.config.Services.DashboardServer.URL,
		"risk_manager":     h.config.Services.RiskManager.URL,
		"configuration":    h.config.Services.Configuration.URL,
	}

	results := make(map[string]interface{})
	client := &http.Client{
		Timeout: h.config.Health.ServiceTimeout,
	}

	for name, url := range services {
		status := h.checkService(client, name, url+"/health")
		results[name] = status
	}

	return results
}

// checkService checks a single service
func (h *HealthHandler) checkService(client *http.Client, name, url string) gin.H {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Health.ServiceTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return gin.H{
			"status": "healthy",
		}
	}

	return gin.H{
		"status":      "unhealthy",
		"status_code": resp.StatusCode,
	}
}
