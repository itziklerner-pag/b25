package admin

const adminPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Gateway - Admin Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
            color: #e2e8f0;
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        .header {
            background: rgba(30, 41, 59, 0.8);
            backdrop-filter: blur(10px);
            border-radius: 16px;
            padding: 32px;
            margin-bottom: 24px;
            border: 1px solid rgba(148, 163, 184, 0.1);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }

        .header h1 {
            font-size: 32px;
            font-weight: 700;
            background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 8px;
        }

        .header .subtitle {
            color: #94a3b8;
            font-size: 14px;
        }

        .status-bar {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 24px;
        }

        .status-card {
            background: rgba(30, 41, 59, 0.8);
            backdrop-filter: blur(10px);
            border-radius: 12px;
            padding: 20px;
            border: 1px solid rgba(148, 163, 184, 0.1);
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
        }

        .status-card .label {
            color: #94a3b8;
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 8px;
        }

        .status-card .value {
            font-size: 24px;
            font-weight: 600;
            color: #f1f5f9;
        }

        .status-card.healthy .value {
            color: #10b981;
        }

        .status-card.warning .value {
            color: #f59e0b;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
            gap: 24px;
            margin-bottom: 24px;
        }

        .card {
            background: rgba(30, 41, 59, 0.8);
            backdrop-filter: blur(10px);
            border-radius: 16px;
            padding: 24px;
            border: 1px solid rgba(148, 163, 184, 0.1);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }

        .card h2 {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 20px;
            color: #f1f5f9;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .card h2::before {
            content: '';
            width: 4px;
            height: 20px;
            background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
            border-radius: 2px;
        }

        .info-grid {
            display: grid;
            gap: 12px;
        }

        .info-row {
            display: flex;
            justify-content: space-between;
            padding: 12px;
            background: rgba(15, 23, 42, 0.5);
            border-radius: 8px;
            border: 1px solid rgba(148, 163, 184, 0.05);
        }

        .info-row .key {
            color: #94a3b8;
            font-size: 14px;
        }

        .info-row .value {
            color: #f1f5f9;
            font-weight: 500;
            font-size: 14px;
            font-family: 'Monaco', 'Menlo', monospace;
        }

        .endpoint-list {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .endpoint-item {
            background: rgba(15, 23, 42, 0.5);
            border-radius: 8px;
            padding: 16px;
            border: 1px solid rgba(148, 163, 184, 0.05);
            transition: all 0.2s;
        }

        .endpoint-item:hover {
            background: rgba(15, 23, 42, 0.7);
            border-color: rgba(59, 130, 246, 0.3);
            transform: translateX(4px);
        }

        .endpoint-item .method {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: 600;
            margin-right: 12px;
            font-family: 'Monaco', 'Menlo', monospace;
        }

        .method.get { background: #10b981; color: #fff; }
        .method.post { background: #3b82f6; color: #fff; }
        .method.put { background: #f59e0b; color: #fff; }
        .method.delete { background: #ef4444; color: #fff; }
        .method.ws { background: #8b5cf6; color: #fff; }

        .endpoint-item .path {
            color: #e2e8f0;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 13px;
        }

        .endpoint-item .description {
            color: #94a3b8;
            font-size: 12px;
            margin-top: 8px;
        }

        .service-item {
            background: rgba(15, 23, 42, 0.5);
            border-radius: 8px;
            padding: 16px;
            border: 1px solid rgba(148, 163, 184, 0.05);
            margin-bottom: 12px;
        }

        .service-item .service-name {
            color: #3b82f6;
            font-weight: 600;
            font-size: 14px;
            margin-bottom: 8px;
        }

        .service-item .service-url {
            color: #94a3b8;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 12px;
        }

        .health-status {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
        }

        .health-status.healthy {
            background: rgba(16, 185, 129, 0.2);
            color: #10b981;
        }

        .health-status.unhealthy {
            background: rgba(239, 68, 68, 0.2);
            color: #ef4444;
        }

        .health-status::before {
            content: '';
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: currentColor;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .feature-badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 600;
            margin: 4px;
        }

        .feature-badge.enabled {
            background: rgba(16, 185, 129, 0.2);
            color: #10b981;
        }

        .feature-badge.disabled {
            background: rgba(100, 116, 139, 0.2);
            color: #64748b;
        }

        .test-panel {
            background: rgba(15, 23, 42, 0.5);
            border-radius: 8px;
            padding: 16px;
            margin-top: 16px;
        }

        .test-input {
            width: 100%;
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.2);
            border-radius: 6px;
            padding: 10px;
            color: #e2e8f0;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 13px;
            margin-bottom: 12px;
        }

        .test-button {
            background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }

        .test-button:hover {
            transform: translateY(-2px);
        }

        .response-viewer {
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 6px;
            padding: 16px;
            margin-top: 12px;
            max-height: 400px;
            overflow-y: auto;
        }

        .response-viewer pre {
            color: #e2e8f0;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 12px;
            white-space: pre-wrap;
            word-wrap: break-word;
        }

        .loading {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid rgba(59, 130, 246, 0.3);
            border-top-color: #3b82f6;
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        @media (max-width: 768px) {
            .grid {
                grid-template-columns: 1fr;
            }
            .status-bar {
                grid-template-columns: 1fr;
            }
        }

        /* Scrollbar styling */
        ::-webkit-scrollbar {
            width: 8px;
            height: 8px;
        }

        ::-webkit-scrollbar-track {
            background: rgba(15, 23, 42, 0.5);
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
            <h1>API Gateway</h1>
            <p class="subtitle">Central routing hub for all microservices</p>
        </div>

        <div class="status-bar">
            <div class="status-card healthy">
                <div class="label">Status</div>
                <div class="value" id="service-status">Loading...</div>
            </div>
            <div class="status-card">
                <div class="label">Version</div>
                <div class="value" id="service-version">-</div>
            </div>
            <div class="status-card">
                <div class="label">Uptime</div>
                <div class="value" id="service-uptime">-</div>
            </div>
            <div class="status-card">
                <div class="label">Port</div>
                <div class="value" id="service-port">-</div>
            </div>
            <div class="status-card">
                <div class="label">Goroutines</div>
                <div class="value" id="goroutines">-</div>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h2>Service Information</h2>
                <div class="info-grid">
                    <div class="info-row">
                        <span class="key">Mode</span>
                        <span class="value" id="server-mode">-</span>
                    </div>
                    <div class="info-row">
                        <span class="key">Go Version</span>
                        <span class="value" id="go-version">-</span>
                    </div>
                    <div class="info-row">
                        <span class="key">CPU Cores</span>
                        <span class="value" id="num-cpu">-</span>
                    </div>
                    <div class="info-row">
                        <span class="key">Start Time</span>
                        <span class="value" id="start-time">-</span>
                    </div>
                </div>
            </div>

            <div class="card">
                <h2>Backend Services</h2>
                <div id="backend-services">
                    <div class="loading"></div>
                </div>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h2>Configuration</h2>
                <div id="config-details">
                    <div class="loading"></div>
                </div>
            </div>

            <div class="card">
                <h2>Features</h2>
                <div id="features">
                    <div class="loading"></div>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>Available Endpoints</h2>
            <div class="endpoint-list" id="endpoints">
                <div class="loading"></div>
            </div>
        </div>

        <div class="card">
            <h2>API Tester</h2>
            <div class="test-panel">
                <select class="test-input" id="test-method">
                    <option value="GET">GET</option>
                    <option value="POST">POST</option>
                    <option value="PUT">PUT</option>
                    <option value="DELETE">DELETE</option>
                </select>
                <input type="text" class="test-input" id="test-endpoint" placeholder="Enter endpoint (e.g., /health)" value="/health">
                <textarea class="test-input" id="test-body" placeholder="Request body (JSON)" rows="4" style="display: none;"></textarea>
                <button class="test-button" onclick="testEndpoint()">Send Request</button>
                <div class="response-viewer" id="response-viewer" style="display: none;">
                    <pre id="response-content"></pre>
                </div>
            </div>
        </div>
    </div>

    <script>
        const API_BASE = window.location.origin + '/services/api-gateway';

        // Show/hide request body based on method
        document.getElementById('test-method').addEventListener('change', function(e) {
            const bodyField = document.getElementById('test-body');
            bodyField.style.display = ['POST', 'PUT'].includes(e.target.value) ? 'block' : 'none';
        });

        async function loadServiceInfo() {
            try {
                const response = await fetch(API_BASE + '/api/service-info');
                const data = await response.json();

                document.getElementById('service-status').innerHTML = '<span class="health-status healthy">Healthy</span>';
                document.getElementById('service-version').textContent = data.version;
                document.getElementById('service-uptime').textContent = data.uptime;
                document.getElementById('service-port').textContent = data.port;
                document.getElementById('goroutines').textContent = data.goroutines;
                document.getElementById('server-mode').textContent = data.mode;
                document.getElementById('go-version').textContent = data.go_version;
                document.getElementById('num-cpu').textContent = data.num_cpu;
                document.getElementById('start-time').textContent = new Date(data.start_time).toLocaleString();

                renderBackendServices(data.config.services);
                renderConfig(data.config);
                renderFeatures(data.config.features);
                renderEndpoints();
            } catch (error) {
                console.error('Failed to load service info:', error);
                document.getElementById('service-status').innerHTML = '<span class="health-status unhealthy">Error</span>';
            }
        }

        function renderBackendServices(services) {
            const container = document.getElementById('backend-services');
            container.innerHTML = '';

            for (const [name, config] of Object.entries(services)) {
                const div = document.createElement('div');
                div.className = 'service-item';
                div.innerHTML = ` + "`" + `
                    <div class="service-name">${formatServiceName(name)}</div>
                    <div class="service-url">${config.url}</div>
                    <div style="margin-top: 6px; color: #64748b; font-size: 11px;">
                        Timeout: ${config.timeout}
                    </div>
                ` + "`" + `;
                container.appendChild(div);
            }
        }

        function renderConfig(config) {
            const container = document.getElementById('config-details');
            container.innerHTML = '';

            const configs = [
                { label: 'Authentication', value: config.auth.enabled ? 'Enabled' : 'Disabled', color: config.auth.enabled ? '#10b981' : '#64748b' },
                { label: 'Rate Limiting', value: config.rate_limit.enabled ? ` + "`" + `${config.rate_limit.global_rps} req/s` + "`" + ` : 'Disabled', color: config.rate_limit.enabled ? '#10b981' : '#64748b' },
                { label: 'CORS', value: config.cors.enabled ? 'Enabled' : 'Disabled', color: config.cors.enabled ? '#10b981' : '#64748b' },
                { label: 'Circuit Breaker', value: config.circuit_breaker.enabled ? 'Enabled' : 'Disabled', color: config.circuit_breaker.enabled ? '#10b981' : '#64748b' },
                { label: 'Cache', value: config.cache.enabled ? ` + "`" + `TTL: ${config.cache.default_ttl}` + "`" + ` : 'Disabled', color: config.cache.enabled ? '#10b981' : '#64748b' },
                { label: 'WebSocket', value: config.websocket.enabled ? ` + "`" + `Max: ${config.websocket.max_connections}` + "`" + ` : 'Disabled', color: config.websocket.enabled ? '#10b981' : '#64748b' },
            ];

            configs.forEach(item => {
                const div = document.createElement('div');
                div.className = 'info-row';
                div.innerHTML = ` + "`" + `
                    <span class="key">${item.label}</span>
                    <span class="value" style="color: ${item.color}">${item.value}</span>
                ` + "`" + `;
                container.appendChild(div);
            });
        }

        function renderFeatures(features) {
            const container = document.getElementById('features');
            container.innerHTML = '';

            for (const [name, enabled] of Object.entries(features)) {
                const badge = document.createElement('span');
                badge.className = ` + "`" + `feature-badge ${enabled ? 'enabled' : 'disabled'}` + "`" + `;
                badge.textContent = formatFeatureName(name);
                container.appendChild(badge);
            }
        }

        function renderEndpoints() {
            const container = document.getElementById('endpoints');
            container.innerHTML = '';

            const endpoints = [
                { method: 'GET', path: '/health', description: 'Health check endpoint' },
                { method: 'GET', path: '/health/liveness', description: 'Kubernetes liveness probe' },
                { method: 'GET', path: '/health/readiness', description: 'Kubernetes readiness probe' },
                { method: 'GET', path: '/metrics', description: 'Prometheus metrics' },
                { method: 'GET', path: '/version', description: 'Service version information' },
                { method: 'GET', path: '/api/v1/market-data/symbols', description: 'Get trading symbols' },
                { method: 'GET', path: '/api/v1/market-data/orderbook/:symbol', description: 'Get order book' },
                { method: 'GET', path: '/api/v1/market-data/trades/:symbol', description: 'Get recent trades' },
                { method: 'GET', path: '/api/v1/market-data/ticker/:symbol', description: 'Get ticker data' },
                { method: 'POST', path: '/api/v1/orders', description: 'Create new order (auth required)' },
                { method: 'GET', path: '/api/v1/orders', description: 'List all orders (auth required)' },
                { method: 'GET', path: '/api/v1/orders/:id', description: 'Get order details (auth required)' },
                { method: 'DELETE', path: '/api/v1/orders/:id', description: 'Cancel order (auth required)' },
                { method: 'GET', path: '/api/v1/strategies', description: 'List strategies (auth required)' },
                { method: 'POST', path: '/api/v1/strategies/:id/start', description: 'Start strategy (auth required)' },
                { method: 'POST', path: '/api/v1/strategies/:id/stop', description: 'Stop strategy (auth required)' },
                { method: 'GET', path: '/api/v1/account/balance', description: 'Get account balance' },
                { method: 'GET', path: '/api/v1/account/positions', description: 'Get open positions' },
                { method: 'GET', path: '/api/v1/account/pnl', description: 'Get P&L report' },
                { method: 'GET', path: '/api/v1/risk/limits', description: 'Get risk limits (auth required)' },
                { method: 'GET', path: '/api/v1/risk/status', description: 'Get risk status (auth required)' },
                { method: 'WS', path: '/ws', description: 'WebSocket connection for real-time updates' },
            ];

            endpoints.forEach(endpoint => {
                const div = document.createElement('div');
                div.className = 'endpoint-item';
                div.innerHTML = ` + "`" + `
                    <div>
                        <span class="method ${endpoint.method.toLowerCase()}">${endpoint.method}</span>
                        <span class="path">${endpoint.path}</span>
                    </div>
                    <div class="description">${endpoint.description}</div>
                ` + "`" + `;
                container.appendChild(div);
            });
        }

        async function testEndpoint() {
            const method = document.getElementById('test-method').value;
            const endpoint = document.getElementById('test-endpoint').value;
            const body = document.getElementById('test-body').value;
            const responseViewer = document.getElementById('response-viewer');
            const responseContent = document.getElementById('response-content');

            responseViewer.style.display = 'block';
            responseContent.textContent = 'Loading...';

            try {
                const options = {
                    method: method,
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };

                if (['POST', 'PUT'].includes(method) && body) {
                    options.body = body;
                }

                const response = await fetch(API_BASE + endpoint, options);
                const contentType = response.headers.get('content-type');

                let data;
                if (contentType && contentType.includes('application/json')) {
                    data = await response.json();
                    responseContent.textContent = JSON.stringify(data, null, 2);
                } else {
                    data = await response.text();
                    responseContent.textContent = data;
                }
            } catch (error) {
                responseContent.textContent = ` + "`" + `Error: ${error.message}` + "`" + `;
            }
        }

        function formatServiceName(name) {
            return name.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ');
        }

        function formatFeatureName(name) {
            return name.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
        }

        // Auto-refresh every 5 seconds
        setInterval(loadServiceInfo, 5000);

        // Initial load
        loadServiceInfo();
    </script>
</body>
</html>
`
