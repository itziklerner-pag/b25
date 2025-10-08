package admin

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"
)

var (
	startTime = time.Now()
	version   = "1.0.0"
)

// Handler provides admin interface endpoints
type Handler struct {
	logger *zap.Logger
	config *Config
}

// Config holds admin configuration
type Config struct {
	HTTPPort       int
	GRPCPort       int
	TestnetMode    bool
	RateLimitRPS   int
	RateLimitBurst int
}

// NewHandler creates a new admin handler
func NewHandler(exec interface{}, logger *zap.Logger, cfg *Config) *Handler {
	return &Handler{
		logger: logger,
		config: cfg,
	}
}

// SetVersion sets the service version
func SetVersion(v string) {
	version = v
}

// AdminPageHandler serves the admin HTML page
func (h *Handler) AdminPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(adminHTML))
	}
}

// ServiceInfoHandler returns service information
func (h *Handler) ServiceInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(startTime)

		// Get system stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		info := map[string]interface{}{
			"service":       "Order Execution Service",
			"version":       version,
			"status":        "running",
			"uptime":        uptime.String(),
			"uptimeSeconds": int64(uptime.Seconds()),
			"ports": map[string]int{
				"http": h.config.HTTPPort,
				"grpc": h.config.GRPCPort,
			},
			"configuration": map[string]interface{}{
				"exchange": map[string]interface{}{
					"name":    "Binance Futures",
					"testnet": h.config.TestnetMode,
					"mode":    h.getMode(),
				},
				"rateLimit": map[string]interface{}{
					"requestsPerSecond": h.config.RateLimitRPS,
					"burst":             h.config.RateLimitBurst,
				},
			},
			"system": map[string]interface{}{
				"goroutines": runtime.NumGoroutine(),
				"memoryMB":   memStats.Alloc / 1024 / 1024,
				"gcCount":    memStats.NumGC,
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}

// getMode returns a readable mode string
func (h *Handler) getMode() string {
	if h.config.TestnetMode {
		return "Testnet (Paper Trading)"
	}
	return "Production (Live Trading)"
}

// EndpointsHandler returns available endpoints
func (h *Handler) EndpointsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoints := map[string]interface{}{
			"grpc": map[string]interface{}{
				"port": h.config.GRPCPort,
				"methods": []map[string]string{
					{
						"name":        "CreateOrder",
						"description": "Create a new order on the exchange",
						"request":     "OrderRequest",
						"response":    "OrderResponse",
					},
					{
						"name":        "CancelOrder",
						"description": "Cancel an existing order",
						"request":     "CancelOrderRequest",
						"response":    "CancelOrderResponse",
					},
					{
						"name":        "GetOrder",
						"description": "Get order details",
						"request":     "GetOrderRequest",
						"response":    "OrderResponse",
					},
					{
						"name":        "GetOrderStatus",
						"description": "Get order status",
						"request":     "GetOrderStatusRequest",
						"response":    "OrderStatusResponse",
					},
				},
			},
			"http": map[string]interface{}{
				"port": h.config.HTTPPort,
				"endpoints": []map[string]string{
					{
						"path":        "/health",
						"method":      "GET",
						"description": "Detailed health check",
					},
					{
						"path":        "/health/ready",
						"method":      "GET",
						"description": "Readiness probe",
					},
					{
						"path":        "/health/live",
						"method":      "GET",
						"description": "Liveness probe",
					},
					{
						"path":        "/metrics",
						"method":      "GET",
						"description": "Prometheus metrics",
					},
					{
						"path":        "/",
						"method":      "GET",
						"description": "Service information",
					},
					{
						"path":        "/admin",
						"method":      "GET",
						"description": "Admin dashboard",
					},
					{
						"path":        "/api/service-info",
						"method":      "GET",
						"description": "Service information JSON",
					},
					{
						"path":        "/api/endpoints",
						"method":      "GET",
						"description": "Available endpoints",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}
}

const adminHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Order Execution Service - Admin Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #0f0f1e 0%, #1a1a2e 100%);
            color: #e0e0e0;
            padding: 24px;
            min-height: 100vh;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        .header {
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 32px;
            margin-bottom: 24px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }

        .header h1 {
            font-size: 32px;
            font-weight: 700;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin-bottom: 8px;
        }

        .header .subtitle {
            color: #a0a0a0;
            font-size: 14px;
            display: flex;
            align-items: center;
            gap: 16px;
            flex-wrap: wrap;
        }

        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
        }

        .status-badge.running {
            background: rgba(16, 185, 129, 0.2);
            color: #10b981;
            border: 1px solid rgba(16, 185, 129, 0.3);
        }

        .status-badge.testnet {
            background: rgba(251, 191, 36, 0.2);
            color: #fbbf24;
            border: 1px solid rgba(251, 191, 36, 0.3);
        }

        .status-badge.production {
            background: rgba(239, 68, 68, 0.2);
            color: #ef4444;
            border: 1px solid rgba(239, 68, 68, 0.3);
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: currentColor;
            animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 24px;
            margin-bottom: 24px;
        }

        .card {
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 24px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
        }

        .card h2 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 16px;
            color: #f0f0f0;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .info-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }

        .info-row:last-child {
            border-bottom: none;
        }

        .info-label {
            color: #a0a0a0;
            font-size: 14px;
        }

        .info-value {
            color: #f0f0f0;
            font-weight: 600;
            font-size: 14px;
            text-align: right;
        }

        .endpoint-list {
            list-style: none;
        }

        .endpoint-item {
            padding: 12px 16px;
            background: rgba(255, 255, 255, 0.03);
            border-radius: 8px;
            margin-bottom: 8px;
            border-left: 3px solid #667eea;
        }

        .endpoint-item:last-child {
            margin-bottom: 0;
        }

        .endpoint-method {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: 700;
            margin-right: 8px;
            background: rgba(102, 126, 234, 0.2);
            color: #667eea;
        }

        .endpoint-path {
            font-family: 'Monaco', 'Courier New', monospace;
            font-size: 13px;
            color: #e0e0e0;
        }

        .endpoint-description {
            font-size: 12px;
            color: #a0a0a0;
            margin-top: 4px;
        }

        .test-section {
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 24px;
            margin-bottom: 24px;
        }

        .test-section h2 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 16px;
            color: #f0f0f0;
        }

        .test-buttons {
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
        }

        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            font-weight: 600;
            font-size: 14px;
            cursor: pointer;
            transition: all 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }

        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .btn-primary:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }

        .btn-secondary {
            background: rgba(255, 255, 255, 0.1);
            color: #e0e0e0;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .btn-secondary:hover {
            background: rgba(255, 255, 255, 0.15);
        }

        .test-result {
            margin-top: 16px;
            padding: 16px;
            border-radius: 8px;
            font-family: 'Monaco', 'Courier New', monospace;
            font-size: 12px;
            display: none;
            max-height: 400px;
            overflow-y: auto;
        }

        .test-result.success {
            background: rgba(16, 185, 129, 0.1);
            border: 1px solid rgba(16, 185, 129, 0.3);
            color: #10b981;
        }

        .test-result.error {
            background: rgba(239, 68, 68, 0.1);
            border: 1px solid rgba(239, 68, 68, 0.3);
            color: #ef4444;
        }

        .footer {
            text-align: center;
            color: #666;
            font-size: 12px;
            margin-top: 48px;
            padding-top: 24px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
        }

        .loading {
            display: inline-block;
            width: 12px;
            height: 12px;
            border: 2px solid rgba(255, 255, 255, 0.3);
            border-top-color: #667eea;
            border-radius: 50%;
            animation: spin 0.6s linear infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .auto-refresh {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 12px;
            color: #a0a0a0;
        }

        .refresh-dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: #10b981;
            animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
        }

        pre {
            white-space: pre-wrap;
            word-wrap: break-word;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Order Execution Service</h1>
            <div class="subtitle">
                <span class="status-badge running">
                    <span class="status-dot"></span>
                    <span id="serviceStatus">Running</span>
                </span>
                <span id="modeBadge" class="status-badge">Loading...</span>
                <span>Version: <span id="version">-</span></span>
                <span>Uptime: <span id="uptime">-</span></span>
                <div class="auto-refresh">
                    <span class="refresh-dot"></span>
                    <span>Auto-refresh: 5s</span>
                </div>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h2>Service Info</h2>
                <div class="info-row">
                    <span class="info-label">HTTP Port</span>
                    <span class="info-value" id="httpPort">-</span>
                </div>
                <div class="info-row">
                    <span class="info-label">gRPC Port</span>
                    <span class="info-value" id="grpcPort">-</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Goroutines</span>
                    <span class="info-value" id="goroutines">-</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Memory Usage</span>
                    <span class="info-value" id="memory">-</span>
                </div>
            </div>

            <div class="card">
                <h2>Exchange Config</h2>
                <div class="info-row">
                    <span class="info-label">Exchange</span>
                    <span class="info-value">Binance Futures</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Mode</span>
                    <span class="info-value" id="exchangeMode">-</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Rate Limit (RPS)</span>
                    <span class="info-value" id="rateLimit">-</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Burst Limit</span>
                    <span class="info-value" id="burstLimit">-</span>
                </div>
            </div>
        </div>

        <div class="grid">
            <div class="card" style="grid-column: 1 / -1;">
                <h2>Available Endpoints</h2>

                <h3 style="color: #a0a0a0; font-size: 14px; margin: 16px 0 12px 0;">HTTP Endpoints (Port: <span id="httpPortEndpoints">9091</span>)</h3>
                <ul class="endpoint-list" id="httpEndpoints">
                    <li class="endpoint-item">
                        <span class="endpoint-method">GET</span>
                        <span class="endpoint-path">Loading...</span>
                    </li>
                </ul>

                <h3 style="color: #a0a0a0; font-size: 14px; margin: 24px 0 12px 0;">gRPC Methods (Port: <span id="grpcPortEndpoints">50051</span>)</h3>
                <ul class="endpoint-list" id="grpcMethods">
                    <li class="endpoint-item">
                        <span class="endpoint-method">RPC</span>
                        <span class="endpoint-path">Loading...</span>
                    </li>
                </ul>
            </div>
        </div>

        <div class="test-section">
            <h2>Interactive Testing</h2>
            <div class="test-buttons">
                <button class="btn btn-primary" onclick="testHealth()">
                    <span>Test Health</span>
                </button>
                <button class="btn btn-primary" onclick="testReadiness()">
                    <span>Test Readiness</span>
                </button>
                <button class="btn btn-primary" onclick="testLiveness()">
                    <span>Test Liveness</span>
                </button>
                <button class="btn btn-secondary" onclick="testMetrics()">
                    <span>View Metrics</span>
                </button>
                <button class="btn btn-secondary" onclick="refreshData()">
                    <span>Refresh Data</span>
                </button>
            </div>
            <div id="testResult" class="test-result"></div>
        </div>

        <div class="footer">
            <p>Order Execution Service | Binance Futures Trading Engine</p>
            <p style="margin-top: 8px;">Last updated: <span id="lastUpdate">-</span></p>
        </div>
    </div>

    <script>
        var API_BASE = window.location.origin + '/services/order-execution';
        var autoRefreshInterval;

        window.addEventListener('DOMContentLoaded', function() {
            refreshData();
            startAutoRefresh();
        });

        function startAutoRefresh() {
            autoRefreshInterval = setInterval(refreshData, 5000);
        }

        function stopAutoRefresh() {
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
            }
        }

        function refreshData() {
            Promise.all([
                loadServiceInfo(),
                loadEndpoints()
            ]).then(function() {
                updateLastUpdateTime();
            }).catch(function(error) {
                console.error('Error refreshing data:', error);
            });
        }

        function loadServiceInfo() {
            return fetch(API_BASE + '/api/service-info')
                .then(function(response) { return response.json(); })
                .then(function(data) {
                    document.getElementById('version').textContent = data.version || '-';
                    document.getElementById('uptime').textContent = data.uptime || '-';

                    var modeBadge = document.getElementById('modeBadge');
                    if (data.configuration && data.configuration.exchange && data.configuration.exchange.testnet) {
                        modeBadge.textContent = 'Testnet Mode';
                        modeBadge.className = 'status-badge testnet';
                    } else {
                        modeBadge.textContent = 'Production Mode';
                        modeBadge.className = 'status-badge production';
                    }

                    document.getElementById('httpPort').textContent = (data.ports && data.ports.http) || '9091';
                    document.getElementById('grpcPort').textContent = (data.ports && data.ports.grpc) || '50051';
                    document.getElementById('goroutines').textContent = (data.system && data.system.goroutines) || '-';
                    document.getElementById('memory').textContent = ((data.system && data.system.memoryMB) || 0) + ' MB';

                    document.getElementById('exchangeMode').textContent = (data.configuration && data.configuration.exchange && data.configuration.exchange.mode) || '-';
                    document.getElementById('rateLimit').textContent = (data.configuration && data.configuration.rateLimit && data.configuration.rateLimit.requestsPerSecond) || '-';
                    document.getElementById('burstLimit').textContent = (data.configuration && data.configuration.rateLimit && data.configuration.rateLimit.burst) || '-';
                })
                .catch(function(error) {
                    console.error('Error loading service info:', error);
                });
        }

        function loadEndpoints() {
            return fetch(API_BASE + '/api/endpoints')
                .then(function(response) { return response.json(); })
                .then(function(data) {
                    var httpList = document.getElementById('httpEndpoints');
                    httpList.innerHTML = '';

                    if (data.http && data.http.endpoints) {
                        document.getElementById('httpPortEndpoints').textContent = data.http.port || '9091';
                        data.http.endpoints.forEach(function(endpoint) {
                            var li = document.createElement('li');
                            li.className = 'endpoint-item';
                            li.innerHTML = '<span class="endpoint-method">' + endpoint.method + '</span>' +
                                '<span class="endpoint-path">' + endpoint.path + '</span>' +
                                '<div class="endpoint-description">' + endpoint.description + '</div>';
                            httpList.appendChild(li);
                        });
                    }

                    var grpcList = document.getElementById('grpcMethods');
                    grpcList.innerHTML = '';

                    if (data.grpc && data.grpc.methods) {
                        document.getElementById('grpcPortEndpoints').textContent = data.grpc.port || '50051';
                        data.grpc.methods.forEach(function(method) {
                            var li = document.createElement('li');
                            li.className = 'endpoint-item';
                            li.innerHTML = '<span class="endpoint-method">RPC</span>' +
                                '<span class="endpoint-path">' + method.name + '</span>' +
                                '<div class="endpoint-description">' + method.description + ' (' + method.request + ' → ' + method.response + ')</div>';
                            grpcList.appendChild(li);
                        });
                    }
                })
                .catch(function(error) {
                    console.error('Error loading endpoints:', error);
                });
        }

        function updateLastUpdateTime() {
            var now = new Date();
            document.getElementById('lastUpdate').textContent = now.toLocaleTimeString();
        }

        function testHealth() {
            testEndpoint('/health', 'Health Check');
        }

        function testReadiness() {
            testEndpoint('/health/ready', 'Readiness Check');
        }

        function testLiveness() {
            testEndpoint('/health/live', 'Liveness Check');
        }

        function testMetrics() {
            testEndpoint('/metrics', 'Prometheus Metrics', true);
        }

        function testEndpoint(path, name, isText) {
            var resultDiv = document.getElementById('testResult');
            resultDiv.style.display = 'block';
            resultDiv.className = 'test-result';
            resultDiv.innerHTML = '<div class="loading"></div> Testing ' + name + '...';

            fetch(API_BASE + path)
                .then(function(response) {
                    var contentType = response.headers.get('content-type');
                    if (isText || !contentType || contentType.indexOf('application/json') === -1) {
                        return response.text();
                    } else {
                        return response.json();
                    }
                })
                .then(function(data) {
                    resultDiv.className = 'test-result success';
                    var displayData = typeof data === 'object' ? JSON.stringify(data, null, 2) : data;
                    resultDiv.innerHTML = '<strong>Success - ' + name + '</strong><br><br><pre>' + displayData + '</pre>';
                })
                .catch(function(error) {
                    resultDiv.className = 'test-result error';
                    resultDiv.innerHTML = '<strong>Error - ' + name + '</strong><br><br>' + error.message;
                });
        }
    </script>
</body>
</html>
`
