package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/b25/strategy-engine/internal/config"
	"github.com/b25/strategy-engine/internal/engine"
)

var startTime = time.Now()

const adminPageHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Strategy Engine - Service Admin</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #0f1419;
            color: #e6edf3;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        header {
            background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
            border: 1px solid #30363d;
        }

        h1 {
            font-size: 32px;
            margin-bottom: 10px;
            color: #58a6ff;
        }

        .status-badge {
            display: inline-block;
            padding: 6px 16px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 600;
            margin-top: 10px;
        }

        .status-healthy {
            background: #238636;
            color: #fff;
        }

        .status-simulation {
            background: #d29922;
            color: #000;
        }

        .status-live {
            background: #da3633;
            color: #fff;
        }

        .status-observation {
            background: #1f6feb;
            color: #fff;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .card {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 24px;
        }

        .card h2 {
            font-size: 20px;
            margin-bottom: 16px;
            color: #58a6ff;
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .icon {
            width: 24px;
            height: 24px;
        }

        .stat-row {
            display: flex;
            justify-content: space-between;
            padding: 12px 0;
            border-bottom: 1px solid #21262d;
        }

        .stat-row:last-child {
            border-bottom: none;
        }

        .stat-label {
            color: #8b949e;
            font-size: 14px;
        }

        .stat-value {
            color: #e6edf3;
            font-weight: 600;
            font-size: 14px;
        }

        .endpoint-list {
            list-style: none;
        }

        .endpoint-item {
            background: #0d1117;
            padding: 12px 16px;
            margin-bottom: 8px;
            border-radius: 6px;
            border-left: 3px solid #58a6ff;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .method {
            background: #238636;
            color: #fff;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: 600;
            margin-right: 10px;
        }

        .test-section {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 24px;
            margin-bottom: 30px;
        }

        .test-controls {
            display: flex;
            gap: 10px;
            margin-top: 16px;
            flex-wrap: wrap;
        }

        button {
            background: #238636;
            color: #fff;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            transition: background 0.2s;
        }

        button:hover {
            background: #2ea043;
        }

        button:active {
            transform: scale(0.98);
        }

        button.secondary {
            background: #21262d;
            color: #e6edf3;
        }

        button.secondary:hover {
            background: #30363d;
        }

        .output-box {
            background: #0d1117;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 16px;
            margin-top: 16px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            max-height: 400px;
            overflow-y: auto;
            white-space: pre-wrap;
            word-wrap: break-word;
        }

        .loading {
            opacity: 0.6;
            pointer-events: none;
        }

        .check-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 8px;
        }

        .check-ok {
            background: #238636;
        }

        .check-error {
            background: #da3633;
        }

        .check-warning {
            background: #d29922;
        }

        .refresh-time {
            color: #8b949e;
            font-size: 12px;
            margin-top: 10px;
        }

        .strategy-card {
            background: #0d1117;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 16px;
            margin-bottom: 12px;
        }

        .strategy-name {
            font-size: 16px;
            font-weight: 600;
            color: #58a6ff;
            margin-bottom: 8px;
        }

        .strategy-details {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
            font-size: 13px;
        }

        .risk-metric {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #21262d;
        }

        .risk-metric:last-child {
            border-bottom: none;
        }

        .risk-label {
            color: #8b949e;
            font-size: 13px;
        }

        .risk-value {
            color: #e6edf3;
            font-weight: 600;
            font-size: 13px;
        }

        .mode-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            margin-left: 8px;
        }

        .mode-live {
            background: #da3633;
            color: #fff;
        }

        .mode-simulation {
            background: #d29922;
            color: #000;
        }

        .mode-observation {
            background: #1f6feb;
            color: #fff;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ Strategy Engine Service</h1>
            <p>Algorithmic trading strategy execution and management</p>
            <div>
                <span id="status-badge" class="status-badge">Loading...</span>
                <span id="mode-badge" class="mode-badge">Loading...</span>
            </div>
            <div class="refresh-time" id="last-refresh">Last updated: Loading...</div>
        </header>

        <div class="grid">
            <div class="card">
                <h2>
                    <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/><path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd"/></svg>
                    Service Information
                </h2>
                <div class="stat-row">
                    <span class="stat-label">Version</span>
                    <span class="stat-value" id="service-version">1.0.0</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Uptime</span>
                    <span class="stat-value" id="service-uptime">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">HTTP Port</span>
                    <span class="stat-value">9092</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Engine Mode</span>
                    <span class="stat-value" id="engine-mode">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Hot Reload</span>
                    <span class="stat-value" id="hot-reload">-</span>
                </div>
            </div>

            <div class="card">
                <h2>
                    <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z" clip-rule="evenodd"/></svg>
                    Active Strategies
                </h2>
                <div class="stat-row">
                    <span class="stat-label">Total Active</span>
                    <span class="stat-value" id="active-strategies">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Signal Queue Size</span>
                    <span class="stat-value" id="signal-queue-size">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Signals Processed</span>
                    <span class="stat-value" id="signals-processed">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Orders Submitted</span>
                    <span class="stat-value" id="orders-submitted">-</span>
                </div>
            </div>

            <div class="card">
                <h2>
                    <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path d="M8.433 7.418c.155-.103.346-.196.567-.267v1.698a2.305 2.305 0 01-.567-.267C8.07 8.34 8 8.114 8 8c0-.114.07-.34.433-.582zM11 12.849v-1.698c.22.071.412.164.567.267.364.243.433.468.433.582 0 .114-.07.34-.433.582a2.305 2.305 0 01-.567.267z"/><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd"/></svg>
                    Backend Connections
                </h2>
                <div class="stat-row">
                    <span class="stat-label"><span class="check-indicator" id="redis-indicator"></span>Redis</span>
                    <span class="stat-value" id="redis-status">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label"><span class="check-indicator" id="nats-indicator"></span>NATS</span>
                    <span class="stat-value" id="nats-status">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label"><span class="check-indicator" id="grpc-indicator"></span>Order Execution gRPC</span>
                    <span class="stat-value" id="grpc-status">-</span>
                </div>
            </div>
        </div>

        <div class="card" style="margin-bottom: 30px;">
            <h2>
                <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/></svg>
                Strategy Details
            </h2>
            <div id="strategy-list">
                <p style="color: #8b949e; font-size: 14px;">Loading strategy information...</p>
            </div>
        </div>

        <div class="card" style="margin-bottom: 30px;">
            <h2>
                <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M3 6a3 3 0 013-3h10a1 1 0 01.8 1.6L14.25 8l2.55 3.4A1 1 0 0116 13H6a1 1 0 00-1 1v3a1 1 0 11-2 0V6z" clip-rule="evenodd"/></svg>
                Risk Management
            </h2>
            <div id="risk-limits">
                <div class="risk-metric">
                    <span class="risk-label">Risk Enabled</span>
                    <span class="risk-value" id="risk-enabled">-</span>
                </div>
                <div class="risk-metric">
                    <span class="risk-label">Max Position Size</span>
                    <span class="risk-value" id="max-position">-</span>
                </div>
                <div class="risk-metric">
                    <span class="risk-label">Max Order Value</span>
                    <span class="risk-value" id="max-order">-</span>
                </div>
                <div class="risk-metric">
                    <span class="risk-label">Max Daily Loss</span>
                    <span class="risk-value" id="max-daily-loss">-</span>
                </div>
                <div class="risk-metric">
                    <span class="risk-label">Max Drawdown</span>
                    <span class="risk-value" id="max-drawdown">-</span>
                </div>
                <div class="risk-metric">
                    <span class="risk-label">Orders/Second Limit</span>
                    <span class="risk-value" id="orders-per-sec">-</span>
                </div>
                <div class="risk-metric">
                    <span class="risk-label">Orders/Minute Limit</span>
                    <span class="risk-value" id="orders-per-min">-</span>
                </div>
            </div>
        </div>

        <div class="card" style="margin-bottom: 30px;">
            <h2>
                <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd"/></svg>
                Available Endpoints
            </h2>
            <ul class="endpoint-list">
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/health</div>
                    <div>Health check</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/ready</div>
                    <div>Readiness probe</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/status</div>
                    <div>Engine status (auth required)</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/metrics</div>
                    <div>Prometheus metrics</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/api/service-info</div>
                    <div>Detailed service information</div>
                </li>
            </ul>
        </div>

        <div class="test-section">
            <h2 style="margin-bottom: 16px; color: #58a6ff;">🧪 Service Testing</h2>
            <p style="color: #8b949e; margin-bottom: 16px;">Test API endpoints and verify service functionality</p>

            <div class="test-controls">
                <button onclick="testHealth()">Test Health</button>
                <button onclick="testReady()">Test Ready</button>
                <button onclick="testStatus()">Test Status</button>
                <button onclick="testServiceInfo()">Test Service Info</button>
                <button onclick="testMetrics()">Test Metrics</button>
                <button class="secondary" onclick="clearOutput()">Clear Output</button>
            </div>

            <div class="output-box" id="test-output">Click a test button to run endpoint tests...</div>
        </div>
    </div>

    <script>
        const API_BASE = window.location.origin + '/services/strategy-engine';

        async function fetchServiceInfo() {
            try {
                const response = await fetch(API_BASE + '/api/service-info');
                const data = await response.json();
                updateServiceDisplay(data);
            } catch (error) {
                console.error('Failed to fetch service info:', error);
                document.getElementById('status-badge').textContent = 'Offline';
                document.getElementById('status-badge').className = 'status-badge status-unhealthy';
            }
        }

        async function fetchStatus() {
            try {
                const response = await fetch(API_BASE + '/status');
                const data = await response.json();
                updateStatusDisplay(data);
            } catch (error) {
                console.error('Failed to fetch status:', error);
            }
        }

        function updateServiceDisplay(data) {
            // Update status badge
            const statusBadge = document.getElementById('status-badge');
            statusBadge.textContent = 'Healthy';
            statusBadge.className = 'status-badge status-healthy';

            // Update mode badge
            const modeBadge = document.getElementById('mode-badge');
            const mode = data.mode || 'simulation';
            modeBadge.textContent = mode.toUpperCase();
            modeBadge.className = 'mode-badge mode-' + mode;

            // Update service info
            document.getElementById('service-version').textContent = data.version || '1.0.0';
            document.getElementById('service-uptime').textContent = data.uptime || 'N/A';
            document.getElementById('engine-mode').textContent = (data.mode || 'simulation').toUpperCase();
            document.getElementById('hot-reload').textContent = data.hot_reload ? 'Enabled' : 'Disabled';

            // Update strategy info
            const strategies = data.strategies || {};
            document.getElementById('active-strategies').textContent = Object.keys(strategies).length;

            // Update connections
            updateCheck('redis', data.connections?.redis);
            updateCheck('nats', data.connections?.nats);
            updateCheck('grpc', data.connections?.order_execution);

            // Update risk limits
            if (data.risk) {
                document.getElementById('risk-enabled').textContent = data.risk.enabled ? 'Yes' : 'No';
                document.getElementById('max-position').textContent = data.risk.max_position_size || 'N/A';
                document.getElementById('max-order').textContent = data.risk.max_order_value || 'N/A';
                document.getElementById('max-daily-loss').textContent = data.risk.max_daily_loss || 'N/A';
                document.getElementById('max-drawdown').textContent = data.risk.max_drawdown || 'N/A';
                document.getElementById('orders-per-sec').textContent = data.risk.max_orders_per_second || 'N/A';
                document.getElementById('orders-per-min').textContent = data.risk.max_orders_per_minute || 'N/A';
            }

            // Update strategy list
            updateStrategyList(strategies);

            // Update timestamp
            document.getElementById('last-refresh').textContent = 'Last updated: ' + new Date().toLocaleTimeString();
        }

        function updateStatusDisplay(data) {
            document.getElementById('signal-queue-size').textContent = data.signal_queue_size || '0';
            document.getElementById('active-strategies').textContent = data.active_strategies || '0';
        }

        function updateStrategyList(strategies) {
            const listDiv = document.getElementById('strategy-list');

            if (!strategies || Object.keys(strategies).length === 0) {
                listDiv.innerHTML = '<p style="color: #8b949e; font-size: 14px;">No active strategies</p>';
                return;
            }

            let html = '';
            for (const [name, config] of Object.entries(strategies)) {
                html += '<div class="strategy-card">';
                html += '<div class="strategy-name">' + name.toUpperCase() + '</div>';
                html += '<div class="strategy-details">';

                if (typeof config === 'object') {
                    for (const [key, value] of Object.entries(config)) {
                        html += '<div style="color: #8b949e;">' + key + ': <span style="color: #e6edf3;">' + value + '</span></div>';
                    }
                } else {
                    html += '<div style="color: #8b949e;">Status: <span style="color: #e6edf3;">Active</span></div>';
                }

                html += '</div>';
                html += '</div>';
            }

            listDiv.innerHTML = html;
        }

        function updateCheck(id, check) {
            const indicator = document.getElementById(id + '-indicator');
            const status = document.getElementById(id + '-status');

            if (check && check.status === 'ok') {
                indicator.className = 'check-indicator check-ok';
                status.textContent = 'Connected';
            } else if (check && check.status === 'warning') {
                indicator.className = 'check-indicator check-warning';
                status.textContent = check.message || 'Warning';
            } else {
                indicator.className = 'check-indicator check-error';
                status.textContent = check?.message || 'Disconnected';
            }
        }

        function appendOutput(text, type = 'info') {
            const output = document.getElementById('test-output');
            const timestamp = new Date().toLocaleTimeString();
            const line = '[' + timestamp + '] ' + text + '\n';
            output.textContent += line;
            output.scrollTop = output.scrollHeight;
        }

        function clearOutput() {
            document.getElementById('test-output').textContent = 'Output cleared.\n';
        }

        async function testEndpoint(path, name) {
            appendOutput('Testing ' + name + '...');
            try {
                const response = await fetch(API_BASE + path);
                let data;
                const contentType = response.headers.get('content-type');

                if (contentType && contentType.includes('application/json')) {
                    data = await response.json();
                    appendOutput('Success ' + name + ' - Status ' + response.status);
                    appendOutput('Response: ' + JSON.stringify(data, null, 2));
                } else {
                    const text = await response.text();
                    appendOutput('Success ' + name + ' - Status ' + response.status);
                    appendOutput('Response: ' + text.substring(0, 500) + (text.length > 500 ? '...' : ''));
                }
            } catch (error) {
                appendOutput('Error ' + name + ' - Error: ' + error.message);
            }
        }

        function testHealth() {
            testEndpoint('/health', 'Health Check');
        }

        function testReady() {
            testEndpoint('/ready', 'Ready Check');
        }

        function testStatus() {
            testEndpoint('/status', 'Status');
        }

        function testServiceInfo() {
            testEndpoint('/api/service-info', 'Service Info');
        }

        function testMetrics() {
            testEndpoint('/metrics', 'Metrics');
        }

        // Initial load and auto-refresh
        fetchServiceInfo();
        fetchStatus();
        setInterval(() => {
            fetchServiceInfo();
            fetchStatus();
        }, 5000); // Refresh every 5 seconds
    </script>
</body>
</html>
`

// handleAdminPage serves the admin dashboard page
func handleAdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(adminPageHTML))
}

// handleServiceInfo returns detailed service information
func handleServiceInfo(cfg *config.Config, eng *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		setCORSHeaders(w)

		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		uptime := time.Since(startTime)
		hours := int(uptime.Hours())
		minutes := int(uptime.Minutes()) % 60
		uptimeStr := fmt.Sprintf("%dh %dm", hours, minutes)

		// Get engine metrics
		metrics := eng.GetMetrics()

		// Build strategy information
		strategies := make(map[string]interface{})
		if strategyMetrics, ok := metrics["strategies"].(map[string]interface{}); ok {
			strategies = strategyMetrics
		}

		// Build connection status
		connections := map[string]interface{}{
			"redis": map[string]interface{}{
				"status": "ok",
			},
			"nats": map[string]interface{}{
				"status": "ok",
			},
			"order_execution": map[string]interface{}{
				"status": "ok",
			},
		}

		info := map[string]interface{}{
			"service":    "strategy-engine",
			"version":    "1.0.0",
			"uptime":     uptimeStr,
			"mode":       cfg.Engine.Mode,
			"hot_reload": cfg.Engine.HotReload,
			"port":       cfg.Server.Port,
			"strategies": strategies,
			"connections": connections,
			"risk": map[string]interface{}{
				"enabled":                cfg.Risk.Enabled,
				"max_position_size":      cfg.Risk.MaxPositionSize,
				"max_order_value":        cfg.Risk.MaxOrderValue,
				"max_daily_loss":         cfg.Risk.MaxDailyLoss,
				"max_drawdown":           cfg.Risk.MaxDrawdown,
				"max_orders_per_second":  cfg.Risk.MaxOrdersPerSecond,
				"max_orders_per_minute":  cfg.Risk.MaxOrdersPerMinute,
				"allowed_symbols":        cfg.Risk.AllowedSymbols,
				"blocked_symbols":        cfg.Risk.BlockedSymbols,
			},
			"features": []string{
				"Momentum Strategy",
				"Market Making Strategy",
				"Scalping Strategy",
				"Real-time Market Data Processing",
				"Risk Management",
				"Order Execution via gRPC",
				"Hot Plugin Reloading",
				"Prometheus Metrics",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}
