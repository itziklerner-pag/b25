# Service Monitor Issue

The ServiceMonitor shows services as "Degraded" because:

1. It tries to fetch: `http://localhost:8080/health`
2. Browser is on: `https://mm.itziklerner.com`
3. This fails with ERR_CONNECTION_REFUSED

## Solution:

ServiceMonitor should fetch through Nginx proxy:
- Market Data: `https://mm.itziklerner.com/api/services/market-data/health`
- Order Execution: `https://mm.itziklerner.com/api/services/order-execution/health`
- etc.

Need to:
1. Add Nginx proxy routes for `/api/services/*` 
2. Update ServiceMonitor to use domain URLs
3. Or disable health checks and show status from WebSocket data instead
