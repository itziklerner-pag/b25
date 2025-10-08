package monitor

import (
	"encoding/json"
	"net/http"
)

const adminPageHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Account Monitor - Service Admin</title>
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

        .status-degraded {
            background: #d29922;
            color: #000;
        }

        .status-unhealthy {
            background: #da3633;
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

        .refresh-time {
            color: #8b949e;
            font-size: 12px;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ Account Monitor Service</h1>
            <p>Real-time account monitoring and reconciliation</p>
            <div id="status-badge" class="status-badge">Loading...</div>
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
                    <span class="stat-value" id="service-version">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Uptime</span>
                    <span class="stat-value" id="service-uptime">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">HTTP Port</span>
                    <span class="stat-value">8087</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">gRPC Port</span>
                    <span class="stat-value">50054</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Metrics Port</span>
                    <span class="stat-value">9094</span>
                </div>
            </div>

            <div class="card">
                <h2>
                    <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z" clip-rule="evenodd"/></svg>
                    Health Checks
                </h2>
                <div class="stat-row">
                    <span class="stat-label"><span class="check-indicator" id="db-indicator"></span>Database</span>
                    <span class="stat-value" id="db-status">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label"><span class="check-indicator" id="redis-indicator"></span>Redis</span>
                    <span class="stat-value" id="redis-status">-</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label"><span class="check-indicator" id="ws-indicator"></span>WebSocket</span>
                    <span class="stat-value" id="ws-status">-</span>
                </div>
            </div>

            <div class="card">
                <h2>
                    <svg class="icon" fill="currentColor" viewBox="0 0 20 20"><path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z"/></svg>
                    Configuration
                </h2>
                <div class="stat-row">
                    <span class="stat-label">Exchange</span>
                    <span class="stat-value">Binance</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Testnet</span>
                    <span class="stat-value">No (Production)</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Reconciliation Interval</span>
                    <span class="stat-value">5s</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Alerts Enabled</span>
                    <span class="stat-value">Yes</span>
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
                    <div><span class="method">GET</span>/api/account</div>
                    <div>Account state</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/api/positions</div>
                    <div>Current positions</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/api/pnl</div>
                    <div>Profit & Loss</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/api/balance</div>
                    <div>Account balance</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/api/alerts</div>
                    <div>Active alerts</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">WS</span>/ws</div>
                    <div>WebSocket stream</div>
                </li>
                <li class="endpoint-item">
                    <div><span class="method">GET</span>/metrics</div>
                    <div>Prometheus metrics (port 9094)</div>
                </li>
            </ul>
        </div>

        <div class="test-section">
            <h2 style="margin-bottom: 16px; color: #58a6ff;">🧪 Service Testing</h2>
            <p style="color: #8b949e; margin-bottom: 16px;">Test API endpoints and verify service functionality</p>

            <div class="test-controls">
                <button onclick="testHealth()">Test Health</button>
                <button onclick="testReady()">Test Ready</button>
                <button onclick="testAccount()">Test Account</button>
                <button onclick="testPositions()">Test Positions</button>
                <button onclick="testPnL()">Test P&L</button>
                <button onclick="testBalance()">Test Balance</button>
                <button onclick="testAlerts()">Test Alerts</button>
                <button class="secondary" onclick="clearOutput()">Clear Output</button>
            </div>

            <div class="output-box" id="test-output">Click a test button to run endpoint tests...</div>
        </div>
    </div>

    <script>
        const API_BASE = window.location.origin + '/services/account-monitor';

        async function fetchHealth() {
            try {
                const response = await fetch(API_BASE + '/health');
                const data = await response.json();
                updateHealthDisplay(data);
            } catch (error) {
                console.error('Failed to fetch health:', error);
                document.getElementById('status-badge').textContent = 'Offline';
                document.getElementById('status-badge').className = 'status-badge status-unhealthy';
            }
        }

        function updateHealthDisplay(data) {
            // Update status badge
            const statusBadge = document.getElementById('status-badge');
            statusBadge.textContent = data.status.toUpperCase();
            statusBadge.className = 'status-badge status-' + data.status;

            // Update service info
            document.getElementById('service-version').textContent = data.version || 'N/A';
            document.getElementById('service-uptime').textContent = data.uptime || 'N/A';

            // Update health checks
            updateCheck('db', data.checks?.database);
            updateCheck('redis', data.checks?.redis);
            updateCheck('ws', data.checks?.websocket);

            // Update timestamp
            document.getElementById('last-refresh').textContent = 'Last updated: ' + new Date().toLocaleTimeString();
        }

        function updateCheck(id, check) {
            const indicator = document.getElementById(id + '-indicator');
            const status = document.getElementById(id + '-status');

            if (check && check.status === 'ok') {
                indicator.className = 'check-indicator check-ok';
                status.textContent = 'OK';
            } else {
                indicator.className = 'check-indicator check-error';
                status.textContent = check?.message || 'Error';
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
                const data = await response.json();
                appendOutput('Success ' + name + ' - Status ' + response.status);
                appendOutput('Response: ' + JSON.stringify(data, null, 2));
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

        function testAccount() {
            testEndpoint('/api/account', 'Account State');
        }

        function testPositions() {
            testEndpoint('/api/positions', 'Positions');
        }

        function testPnL() {
            testEndpoint('/api/pnl', 'P&L');
        }

        function testBalance() {
            testEndpoint('/api/balance', 'Balance');
        }

        function testAlerts() {
            testEndpoint('/api/alerts', 'Alerts');
        }

        // Initial load and auto-refresh
        fetchHealth();
        setInterval(fetchHealth, 5000); // Refresh every 5 seconds
    </script>
</body>
</html>
`

// HandleAdminPage serves the admin dashboard page
func (m *AccountMonitor) HandleAdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(adminPageHTML))
}

// HandleServiceInfo returns detailed service information
func (m *AccountMonitor) HandleServiceInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service": "account-monitor",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"http_port":    8087,
			"grpc_port":    50054,
			"metrics_port": 9094,
		},
		"features": []string{
			"Real-time account monitoring",
			"Position tracking",
			"Balance reconciliation",
			"P&L calculation",
			"Alert management",
			"WebSocket streaming",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(info)
}
