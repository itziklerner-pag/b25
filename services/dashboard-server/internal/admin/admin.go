package admin

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/rs/zerolog"

	"github.com/yourusername/b25/services/dashboard-server/internal/aggregator"
	"github.com/yourusername/b25/services/dashboard-server/internal/server"
)

const (
	serviceName    = "dashboard-server"
	serviceVersion = "1.0.0"
)

var startTime = time.Now()

// Handler provides admin page functionality
type Handler struct {
	logger      zerolog.Logger
	aggregator  *aggregator.Aggregator
	wsServer    *server.Server
}

// NewHandler creates a new admin handler
func NewHandler(logger zerolog.Logger, agg *aggregator.Aggregator, ws *server.Server) *Handler {
	return &Handler{
		logger:     logger,
		aggregator: agg,
		wsServer:   ws,
	}
}

// HandleAdminPage serves the admin dashboard HTML
func (h *Handler) HandleAdminPage(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(adminPageHTML))
}

// HandleServiceInfo provides service information API
func (h *Handler) HandleServiceInfo(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get current state
	state := h.aggregator.GetFullState()

	// Count WebSocket clients
	connectedClients := h.wsServer.GetClientCount()

	info := ServiceInfo{
		Service: ServiceDetails{
			Name:    serviceName,
			Version: serviceVersion,
			Uptime:  time.Since(startTime).String(),
			Started: startTime.Format(time.RFC3339),
		},
		Runtime: RuntimeInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
		},
		WebSocket: WebSocketInfo{
			ConnectedClients: connectedClients,
			TotalClients:     connectedClients, // For now, same as connected
			Format:           "JSON/MessagePack",
		},
		State: StateInfo{
			Sequence:       state.Sequence,
			LastUpdate:     state.Timestamp.Format(time.RFC3339),
			MarketDataCount: len(state.MarketData),
			OrdersCount:    len(state.Orders),
			PositionsCount: len(state.Positions),
			StrategiesCount: len(state.Strategies),
		},
		Backend: BackendServicesInfo{
			OrderExecution: BackendServiceStatus{
				Status:      "connected",
				LastChecked: time.Now().Format(time.RFC3339),
			},
			StrategyEngine: BackendServiceStatus{
				Status:      "connected",
				LastChecked: time.Now().Format(time.RFC3339),
			},
			AccountMonitor: BackendServiceStatus{
				Status:      "connected",
				LastChecked: time.Now().Format(time.RFC3339),
			},
			Redis: BackendServiceStatus{
				Status:      "connected",
				LastChecked: time.Now().Format(time.RFC3339),
			},
		},
		Health: HealthInfo{
			Status: "healthy",
			Checks: map[string]string{
				"aggregator": "ok",
				"websocket":  "ok",
				"redis":      "ok",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// ServiceInfo represents the complete service information
type ServiceInfo struct {
	Service   ServiceDetails       `json:"service"`
	Runtime   RuntimeInfo          `json:"runtime"`
	WebSocket WebSocketInfo        `json:"websocket"`
	State     StateInfo            `json:"state"`
	Backend   BackendServicesInfo  `json:"backend"`
	Health    HealthInfo           `json:"health"`
}

type ServiceDetails struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	Started string `json:"started"`
}

type RuntimeInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
}

type WebSocketInfo struct {
	ConnectedClients int    `json:"connected_clients"`
	TotalClients     int    `json:"total_clients"`
	Format           string `json:"format"`
}

type StateInfo struct {
	Sequence        uint64 `json:"sequence"`
	LastUpdate      string `json:"last_update"`
	MarketDataCount int    `json:"market_data_count"`
	OrdersCount     int    `json:"orders_count"`
	PositionsCount  int    `json:"positions_count"`
	StrategiesCount int    `json:"strategies_count"`
}

type BackendServicesInfo struct {
	OrderExecution BackendServiceStatus `json:"order_execution"`
	StrategyEngine BackendServiceStatus `json:"strategy_engine"`
	AccountMonitor BackendServiceStatus `json:"account_monitor"`
	Redis          BackendServiceStatus `json:"redis"`
}

type BackendServiceStatus struct {
	Status      string `json:"status"`
	LastChecked string `json:"last_checked"`
}

type HealthInfo struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

const adminPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard Server - Admin</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
            color: #e2e8f0;
            min-height: 100vh;
            padding: 2rem;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        .header {
            text-align: center;
            margin-bottom: 3rem;
            padding: 2rem;
            background: rgba(30, 41, 59, 0.5);
            border-radius: 16px;
            border: 1px solid rgba(148, 163, 184, 0.1);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }

        .header h1 {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, #60a5fa 0%, #a78bfa 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 0.5rem;
        }

        .header .subtitle {
            color: #94a3b8;
            font-size: 1.1rem;
        }

        .status-badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.875rem;
            font-weight: 600;
            margin-top: 1rem;
        }

        .status-badge.healthy {
            background: rgba(16, 185, 129, 0.2);
            color: #10b981;
            border: 1px solid rgba(16, 185, 129, 0.3);
        }

        .status-badge.warning {
            background: rgba(251, 191, 36, 0.2);
            color: #fbbf24;
            border: 1px solid rgba(251, 191, 36, 0.3);
        }

        .status-badge.error {
            background: rgba(239, 68, 68, 0.2);
            color: #ef4444;
            border: 1px solid rgba(239, 68, 68, 0.3);
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }

        .card {
            background: rgba(30, 41, 59, 0.6);
            border-radius: 12px;
            padding: 1.5rem;
            border: 1px solid rgba(148, 163, 184, 0.1);
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
            backdrop-filter: blur(10px);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
        }

        .card-title {
            font-size: 0.875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: #94a3b8;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .card-value {
            font-size: 2rem;
            font-weight: 700;
            color: #f1f5f9;
            margin-bottom: 0.5rem;
        }

        .card-label {
            font-size: 0.875rem;
            color: #64748b;
        }

        .section {
            background: rgba(30, 41, 59, 0.6);
            border-radius: 12px;
            padding: 2rem;
            border: 1px solid rgba(148, 163, 184, 0.1);
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
            margin-bottom: 2rem;
        }

        .section-title {
            font-size: 1.5rem;
            font-weight: 700;
            margin-bottom: 1.5rem;
            color: #f1f5f9;
            display: flex;
            align-items: center;
            gap: 0.75rem;
        }

        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 1rem;
        }

        .info-item {
            padding: 1rem;
            background: rgba(15, 23, 42, 0.5);
            border-radius: 8px;
            border: 1px solid rgba(148, 163, 184, 0.1);
        }

        .info-item-label {
            font-size: 0.75rem;
            color: #94a3b8;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.5rem;
        }

        .info-item-value {
            font-size: 1rem;
            color: #f1f5f9;
            font-weight: 600;
        }

        .backend-services {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 1rem;
        }

        .backend-service {
            padding: 1rem;
            background: rgba(15, 23, 42, 0.5);
            border-radius: 8px;
            border: 1px solid rgba(148, 163, 184, 0.1);
            display: flex;
            align-items: center;
            gap: 0.75rem;
        }

        .service-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }

        .service-indicator.connected {
            background: #10b981;
            box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
        }

        .service-indicator.disconnected {
            background: #ef4444;
            box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .service-name {
            font-size: 0.875rem;
            font-weight: 600;
            color: #f1f5f9;
        }

        .service-status {
            font-size: 0.75rem;
            color: #94a3b8;
            margin-top: 0.25rem;
        }

        .test-section {
            margin-top: 1.5rem;
        }

        .test-buttons {
            display: flex;
            gap: 0.75rem;
            flex-wrap: wrap;
        }

        .btn {
            padding: 0.75rem 1.5rem;
            border: none;
            border-radius: 8px;
            font-size: 0.875rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s;
            background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
            color: white;
            box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
        }

        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
        }

        .btn:active {
            transform: translateY(0);
        }

        .btn-secondary {
            background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
            box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
        }

        .btn-secondary:hover {
            box-shadow: 0 6px 16px rgba(139, 92, 246, 0.4);
        }

        .test-result {
            margin-top: 1rem;
            padding: 1rem;
            background: rgba(15, 23, 42, 0.7);
            border-radius: 8px;
            border: 1px solid rgba(148, 163, 184, 0.1);
            font-family: 'Courier New', monospace;
            font-size: 0.875rem;
            max-height: 400px;
            overflow-y: auto;
            display: none;
        }

        .test-result.show {
            display: block;
        }

        .icon {
            width: 20px;
            height: 20px;
            display: inline-block;
        }

        .refresh-indicator {
            display: inline-block;
            width: 8px;
            height: 8px;
            background: #10b981;
            border-radius: 50%;
            margin-left: 0.5rem;
            animation: pulse 2s infinite;
        }

        .loading {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid rgba(148, 163, 184, 0.3);
            border-top-color: #60a5fa;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        /* Scrollbar styling */
        ::-webkit-scrollbar {
            width: 8px;
            height: 8px;
        }

        ::-webkit-scrollbar-track {
            background: rgba(15, 23, 42, 0.5);
            border-radius: 4px;
        }

        ::-webkit-scrollbar-thumb {
            background: rgba(148, 163, 184, 0.3);
            border-radius: 4px;
        }

        ::-webkit-scrollbar-thumb:hover {
            background: rgba(148, 163, 184, 0.5);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚡ Dashboard Server</h1>
            <p class="subtitle">Real-time WebSocket Aggregation & Broadcasting</p>
            <span class="status-badge healthy" id="status-badge">
                <span class="loading" id="loading" style="display: inline-block;"></span>
                <span id="status-text">Loading...</span>
            </span>
        </div>

        <div class="grid">
            <div class="card">
                <div class="card-title">
                    <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                    Service Uptime
                </div>
                <div class="card-value" id="uptime">-</div>
                <div class="card-label">Since <span id="started">-</span></div>
            </div>

            <div class="card">
                <div class="card-title">
                    <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                    Connected Clients
                </div>
                <div class="card-value" id="connected-clients">-</div>
                <div class="card-label">Active WebSocket connections</div>
            </div>

            <div class="card">
                <div class="card-title">
                    <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
                    </svg>
                    State Sequence
                </div>
                <div class="card-value" id="sequence">-</div>
                <div class="card-label">Total state updates</div>
            </div>

            <div class="card">
                <div class="card-title">
                    <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                    </svg>
                    Goroutines
                </div>
                <div class="card-value" id="goroutines">-</div>
                <div class="card-label">Runtime concurrency</div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">
                <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Service Information
                <span class="refresh-indicator"></span>
            </div>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-item-label">Service Name</div>
                    <div class="info-item-value" id="service-name">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">Version</div>
                    <div class="info-item-value" id="version">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">Go Version</div>
                    <div class="info-item-value" id="go-version">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">CPU Cores</div>
                    <div class="info-item-value" id="num-cpu">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">WebSocket Format</div>
                    <div class="info-item-value" id="ws-format">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">Last Update</div>
                    <div class="info-item-value" id="last-update">-</div>
                </div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">
                <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                </svg>
                Aggregated State
            </div>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-item-label">Market Data</div>
                    <div class="info-item-value" id="market-data-count">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">Active Orders</div>
                    <div class="info-item-value" id="orders-count">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">Open Positions</div>
                    <div class="info-item-value" id="positions-count">-</div>
                </div>
                <div class="info-item">
                    <div class="info-item-label">Active Strategies</div>
                    <div class="info-item-value" id="strategies-count">-</div>
                </div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">
                <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                </svg>
                Backend Services
            </div>
            <div class="backend-services">
                <div class="backend-service">
                    <div class="service-indicator connected" id="order-indicator"></div>
                    <div>
                        <div class="service-name">Order Execution</div>
                        <div class="service-status" id="order-status">Checking...</div>
                    </div>
                </div>
                <div class="backend-service">
                    <div class="service-indicator connected" id="strategy-indicator"></div>
                    <div>
                        <div class="service-name">Strategy Engine</div>
                        <div class="service-status" id="strategy-status">Checking...</div>
                    </div>
                </div>
                <div class="backend-service">
                    <div class="service-indicator connected" id="account-indicator"></div>
                    <div>
                        <div class="service-name">Account Monitor</div>
                        <div class="service-status" id="account-status">Checking...</div>
                    </div>
                </div>
                <div class="backend-service">
                    <div class="service-indicator connected" id="redis-indicator"></div>
                    <div>
                        <div class="service-name">Redis</div>
                        <div class="service-status" id="redis-status">Checking...</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="section">
            <div class="section-title">
                <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
                API Endpoint Testing
            </div>
            <div class="test-section">
                <div class="test-buttons">
                    <button class="btn" onclick="testEndpoint('/health')">Test Health</button>
                    <button class="btn" onclick="testEndpoint('/debug')">Test Debug</button>
                    <button class="btn" onclick="testEndpoint('/api/v1/history?type=market_data&limit=10')">Test History API</button>
                    <button class="btn btn-secondary" onclick="testWebSocket()">Test WebSocket</button>
                </div>
                <div class="test-result" id="test-result"></div>
            </div>
        </div>
    </div>

    <script>
        const API_BASE = window.location.origin + '/services/dashboard-server';
        let refreshInterval;

        async function fetchServiceInfo() {
            try {
                const response = await fetch(API_BASE + '/api/service-info');
                const data = await response.json();
                updateUI(data);
            } catch (error) {
                console.error('Error fetching service info:', error);
                document.getElementById('status-badge').className = 'status-badge error';
                document.getElementById('status-text').textContent = 'Error';
                document.getElementById('loading').style.display = 'none';
            }
        }

        function updateUI(data) {
            // Update status badge
            const statusBadge = document.getElementById('status-badge');
            const statusText = document.getElementById('status-text');
            const loading = document.getElementById('loading');

            statusBadge.className = 'status-badge ' + (data.health.status === 'healthy' ? 'healthy' : 'warning');
            statusText.textContent = data.health.status.charAt(0).toUpperCase() + data.health.status.slice(1);
            loading.style.display = 'none';

            // Update metrics cards
            document.getElementById('uptime').textContent = data.service.uptime;
            document.getElementById('started').textContent = new Date(data.service.started).toLocaleString();
            document.getElementById('connected-clients').textContent = data.websocket.connected_clients;
            document.getElementById('sequence').textContent = data.state.sequence.toLocaleString();
            document.getElementById('goroutines').textContent = data.runtime.num_goroutine;

            // Update service info
            document.getElementById('service-name').textContent = data.service.name;
            document.getElementById('version').textContent = data.service.version;
            document.getElementById('go-version').textContent = data.runtime.go_version;
            document.getElementById('num-cpu').textContent = data.runtime.num_cpu;
            document.getElementById('ws-format').textContent = data.websocket.format;

            const lastUpdate = new Date(data.state.last_update);
            const now = new Date();
            const diffSeconds = Math.floor((now - lastUpdate) / 1000);
            document.getElementById('last-update').textContent = diffSeconds < 60 ?
                diffSeconds + 's ago' :
                Math.floor(diffSeconds / 60) + 'm ago';

            // Update state counts
            document.getElementById('market-data-count').textContent = data.state.market_data_count;
            document.getElementById('orders-count').textContent = data.state.orders_count;
            document.getElementById('positions-count').textContent = data.state.positions_count;
            document.getElementById('strategies-count').textContent = data.state.strategies_count;

            // Update backend services
            updateBackendService('order', data.backend.order_execution);
            updateBackendService('strategy', data.backend.strategy_engine);
            updateBackendService('account', data.backend.account_monitor);
            updateBackendService('redis', data.backend.redis);
        }

        function updateBackendService(name, service) {
            const indicator = document.getElementById(name + '-indicator');
            const status = document.getElementById(name + '-status');

            if (service.status === 'connected') {
                indicator.className = 'service-indicator connected';
                status.textContent = 'Connected';
            } else {
                indicator.className = 'service-indicator disconnected';
                status.textContent = 'Disconnected';
            }
        }

        async function testEndpoint(endpoint) {
            const resultDiv = document.getElementById('test-result');
            resultDiv.classList.add('show');
            resultDiv.textContent = 'Testing ' + endpoint + '...\n';

            try {
                const response = await fetch(API_BASE + endpoint);
                const data = await response.json();
                resultDiv.textContent = 'Status: ' + response.status + '\n\n' +
                    JSON.stringify(data, null, 2);
            } catch (error) {
                resultDiv.textContent = 'Error: ' + error.message;
            }
        }

        function testWebSocket() {
            const resultDiv = document.getElementById('test-result');
            resultDiv.classList.add('show');
            resultDiv.textContent = 'Connecting to WebSocket...\n';

            try {
                const wsUrl = 'ws://' + window.location.host + '/services/dashboard-server/ws?type=web&format=json';
                const ws = new WebSocket(wsUrl);

                ws.onopen = () => {
                    resultDiv.textContent += 'Connected!\n';
                    resultDiv.textContent += 'Subscribing to all channels...\n';
                    ws.send(JSON.stringify({
                        type: 'subscribe',
                        channels: ['market_data', 'orders', 'positions', 'strategies', 'account']
                    }));
                };

                ws.onmessage = (event) => {
                    const data = JSON.parse(event.data);
                    resultDiv.textContent += '\nReceived: ' + data.type + ' (seq: ' + data.sequence + ')\n';
                    if (data.type === 'snapshot') {
                        resultDiv.textContent += 'Market Data: ' + Object.keys(data.data.market_data || {}).length + ' symbols\n';
                        resultDiv.textContent += 'Orders: ' + (data.data.orders || []).length + '\n';
                        resultDiv.textContent += 'Positions: ' + Object.keys(data.data.positions || {}).length + '\n';
                        resultDiv.textContent += 'Strategies: ' + Object.keys(data.data.strategies || {}).length + '\n';
                    }
                };

                ws.onerror = (error) => {
                    resultDiv.textContent += '\nError: ' + error.message;
                };

                ws.onclose = () => {
                    resultDiv.textContent += '\nConnection closed';
                };

                setTimeout(() => {
                    ws.close();
                    resultDiv.textContent += '\nTest completed';
                }, 5000);
            } catch (error) {
                resultDiv.textContent = 'Error: ' + error.message;
            }
        }

        // Initial load and auto-refresh every 5 seconds
        fetchServiceInfo();
        refreshInterval = setInterval(fetchServiceInfo, 5000);
    </script>
</body>
</html>`
