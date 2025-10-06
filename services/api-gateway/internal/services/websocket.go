package services

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// WebSocketProxy handles WebSocket connections and proxies them to backend services
type WebSocketProxy struct {
	config *config.Config
	log    *logger.Logger
}

// NewWebSocketProxy creates a new WebSocket proxy
func NewWebSocketProxy(cfg *config.Config, log *logger.Logger) *WebSocketProxy {
	return &WebSocketProxy{
		config: cfg,
		log:    log,
	}
}

// getServiceURL returns the service URL for a given service name
func (p *WebSocketProxy) getServiceURL(serviceName string) (config.ServiceConfig, bool) {
	switch serviceName {
	case "market_data":
		return p.config.Services.MarketData, true
	case "order_execution":
		return p.config.Services.OrderExecution, true
	case "strategy_engine":
		return p.config.Services.StrategyEngine, true
	case "account_monitor":
		return p.config.Services.AccountMonitor, true
	case "dashboard_server":
		return p.config.Services.DashboardServer, true
	case "risk_manager":
		return p.config.Services.RiskManager, true
	case "configuration":
		return p.config.Services.Configuration, true
	default:
		return config.ServiceConfig{}, false
	}
}

// ProxyWebSocket handles WebSocket upgrade and proxies the connection
func (p *WebSocketProxy) ProxyWebSocket(c *gin.Context, targetService string) {
	// Get target service URL
	serviceURL, exists := p.getServiceURL(targetService)
	if !exists {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unknown service"})
		return
	}

	// Parse target URL
	target, err := url.Parse(serviceURL.URL)
	if err != nil {
		p.log.Error("Failed to parse service URL", "service", targetService, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid service configuration"})
		return
	}

	// Change scheme to ws/wss
	if target.Scheme == "http" {
		target.Scheme = "ws"
	} else if target.Scheme == "https" {
		target.Scheme = "wss"
	}

	// Build target URL with path
	target.Path = c.Request.URL.Path
	target.RawQuery = c.Request.URL.RawQuery

	// Create request to backend
	backendURL := target.String()
	p.log.Debug("Proxying WebSocket connection", "backend", backendURL)

	// Create backend request
	req, err := http.NewRequest(c.Request.Method, backendURL, nil)
	if err != nil {
		p.log.Error("Failed to create backend request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy request"})
		return
	}

	// Copy headers
	copyHeaders(req.Header, c.Request.Header)

	// Add forwarding headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", c.Request.URL.Scheme)
	req.Header.Set("X-Forwarded-Host", c.Request.Host)
	req.Header.Set("X-Gateway-Version", "1.0.0")

	// Get request ID from context
	if requestID, exists := c.Get("request_id"); exists {
		req.Header.Set("X-Request-ID", requestID.(string))
	}

	// Perform the upgrade
	client := &http.Client{
		Timeout: p.config.WebSocket.HandshakeTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		p.log.Error("Failed to connect to backend", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to backend service"})
		return
	}
	defer resp.Body.Close()

	// Check if upgrade was successful
	if resp.StatusCode != http.StatusSwitchingProtocols {
		p.log.Warn("Backend did not upgrade to WebSocket", "status", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "Backend did not accept WebSocket upgrade"})
		return
	}

	// Get the underlying connection
	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		p.log.Error("Response writer does not support hijacking")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade not supported"})
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		p.log.Error("Failed to hijack connection", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer clientConn.Close()

	// Get backend connection
	backendConn, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		p.log.Error("Backend response body is not a ReadWriteCloser")
		return
	}
	defer backendConn.Close()

	// Write upgrade response to client
	err = resp.Write(clientConn)
	if err != nil {
		p.log.Error("Failed to write upgrade response", "error", err)
		return
	}

	// Start bidirectional copy
	errChan := make(chan error, 2)

	// Copy from client to backend
	go func() {
		_, err := io.Copy(backendConn, clientConn)
		errChan <- err
	}()

	// Copy from backend to client
	go func() {
		_, err := io.Copy(clientConn, backendConn)
		errChan <- err
	}()

	// Wait for one direction to complete
	err = <-errChan
	if err != nil && err != io.EOF {
		p.log.Debug("WebSocket connection closed", "error", err)
	} else {
		p.log.Debug("WebSocket connection closed normally")
	}
}

// copyHeaders copies HTTP headers from source to destination
func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		// Skip headers that shouldn't be forwarded
		lowerKey := strings.ToLower(key)
		if lowerKey == "connection" || lowerKey == "upgrade" || lowerKey == "keep-alive" {
			continue
		}

		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// HandleWebSocket is a convenience method for handling WebSocket routes
func (p *WebSocketProxy) HandleWebSocket(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p.ProxyWebSocket(c, service)
	}
}

// GetWebSocketStats returns statistics about active WebSocket connections
func (p *WebSocketProxy) GetWebSocketStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         p.config.WebSocket.Enabled,
		"max_connections": p.config.WebSocket.MaxConnections,
		// In a production implementation, you would track active connections
		// and return them here
	}
}
