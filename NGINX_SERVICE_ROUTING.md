# B25 Service Routing Configuration

## Summary

Successfully configured Nginx reverse proxy routes for all B25 services to enable proper health check monitoring from the ServiceMonitor component.

## Changes Made

### 1. Nginx Configuration (`/etc/nginx/sites-available/mm.itziklerner.com`)

Added service proxy routes for all B25 backend services:

```nginx
# Service API Proxies (for health checks and monitoring)
location /services/market-data/ {
    proxy_pass http://localhost:8080/;
    # CORS headers and proxy settings
}

location /services/order-execution/ {
    proxy_pass http://localhost:8081/;
}

location /services/strategy-engine/ {
    proxy_pass http://localhost:8082/;
}

location /services/risk-manager/ {
    proxy_pass http://localhost:8083/;
}

location /services/account-monitor/ {
    proxy_pass http://localhost:8084/;
}

location /services/configuration/ {
    proxy_pass http://localhost:8085/;
}

location /services/dashboard-server/ {
    proxy_pass http://localhost:8086/;
}

location /services/api-gateway/ {
    proxy_pass http://localhost:8000/;
}

location /services/auth-service/ {
    proxy_pass http://localhost:9097/;
}

location /services/prometheus/ {
    proxy_pass http://localhost:9090/;
}

location /services/grafana-internal/ {
    proxy_pass http://localhost:3001/;
}
```

**Features:**
- CORS headers enabled for cross-origin requests
- OPTIONS preflight request handling
- 10s connection timeout, 30s read/send timeout
- Standard proxy headers (X-Real-IP, X-Forwarded-For, etc.)

### 2. ServiceMonitor Component (`/home/mm/dev/b25/ui/web/src/components/ServiceMonitor.tsx`)

Updated health check URLs from localhost to proxied endpoints:

**Before:**
```typescript
url: 'https://mm.itziklerner.com/api/services:8080'
```

**After:**
```typescript
url: 'https://mm.itziklerner.com/services/market-data/health'
```

**Complete URL Mapping:**
- Market Data: `https://mm.itziklerner.com/services/market-data/health` → `http://localhost:8080/health`
- Order Execution: `https://mm.itziklerner.com/services/order-execution/health` → `http://localhost:8081/health`
- Strategy Engine: `https://mm.itziklerner.com/services/strategy-engine/health` → `http://localhost:8082/health`
- Risk Manager: `https://mm.itziklerner.com/services/risk-manager/health` → `http://localhost:8083/health`
- Account Monitor: `https://mm.itziklerner.com/services/account-monitor/health` → `http://localhost:8084/health`
- Configuration: `https://mm.itziklerner.com/services/configuration/health` → `http://localhost:8085/health`
- Dashboard Server: `https://mm.itziklerner.com/services/dashboard-server/health` → `http://localhost:8086/health`
- API Gateway: `https://mm.itziklerner.com/services/api-gateway/health` → `http://localhost:8000/health`
- Auth Service: `https://mm.itziklerner.com/services/auth-service/health` → `http://localhost:9097/health`
- Prometheus: `https://mm.itziklerner.com/services/prometheus/health` → `http://localhost:9090/health`
- Grafana: `https://mm.itziklerner.com/services/grafana-internal/health` → `http://localhost:3001/health`

### 3. TypeScript Types (`/home/mm/dev/b25/ui/web/src/types/index.ts`)

Added `changes` property to WebSocketMessage interface to fix TypeScript compilation errors:

```typescript
export interface WebSocketMessage {
  type: string;
  data?: any;
  changes?: any;  // Added this field
  timestamp?: number;
  channel?: string;
}
```

## Verification Results

### Service Health Status

```
✓ Market Data (8080)       - HEALTHY
✓ Order Execution (8081)   - HEALTHY
✓ Strategy Engine (8082)   - HEALTHY
✓ Risk Manager (8083)      - HEALTHY
✗ Account Monitor (8084)   - NOT FOUND (404) - Service not implementing /health endpoint
✓ Configuration (8085)     - HEALTHY
✓ Dashboard Server (8086)  - HEALTHY
✓ API Gateway (8000)       - HEALTHY (aggregates all service health)
✗ Auth Service (9097)      - SERVICE UNAVAILABLE (502) - Service not running
✗ Prometheus (9090)        - NOT FOUND (404) - Service not implementing /health endpoint
~ Grafana (3001)           - DEGRADED (302 redirect to login)
```

**Summary:**
- 7 services HEALTHY
- 1 service DEGRADED
- 3 services UNHEALTHY

### Testing Commands

Test individual services:
```bash
curl https://mm.itziklerner.com/services/market-data/health
curl https://mm.itziklerner.com/services/order-execution/health
curl https://mm.itziklerner.com/services/strategy-engine/health
curl https://mm.itziklerner.com/services/risk-manager/health
curl https://mm.itziklerner.com/services/configuration/health
curl https://mm.itziklerner.com/services/dashboard-server/health
curl https://mm.itziklerner.com/services/api-gateway/health
```

Run full verification:
```bash
/home/mm/dev/b25/verify-service-routes.sh
```

### Nginx Operations

Test configuration:
```bash
sudo nginx -t
```

Reload Nginx:
```bash
sudo systemctl reload nginx
```

Check Nginx status:
```bash
sudo systemctl status nginx
```

View Nginx logs:
```bash
tail -f /var/log/nginx/b25-access.log
tail -f /var/log/nginx/b25-error.log
```

## Frontend Build

The frontend was successfully rebuilt with the new service URLs:

```bash
cd /home/mm/dev/b25/ui/web
npm run build
```

Build output location: `/home/mm/dev/b25/ui/web/dist/`

## User Interface

Access the Service Monitor at: https://mm.itziklerner.com/system

**Expected Behavior:**
- Services with green indicators are HEALTHY
- Services auto-refresh every 10 seconds
- Click on Market Data Service card to view detailed monitoring
- Status counts at the top show healthy/degraded/unhealthy services
- No more localhost connection errors
- No more CORS errors in browser console

## Known Issues

1. **Account Monitor (8084)** - Not implementing `/health` endpoint, returns 404
2. **Auth Service (9097)** - Service not running, returns 502
3. **Prometheus (9090)** - Not implementing standard `/health` endpoint
4. **Grafana (3001)** - Redirects to login page instead of health check

## Recommendations

1. **Implement /health endpoints** for Account Monitor and Prometheus services
2. **Start Auth Service** if it's required for the system
3. **Configure Grafana** to allow unauthenticated health checks or use a different endpoint
4. **Add service-specific health checks** for infrastructure services (Redis, PostgreSQL, NATS, TimescaleDB) through a dedicated monitoring service

## Architecture Benefits

This configuration provides:

1. **Security**: Services are not directly exposed, only accessible through Nginx proxy
2. **CORS Handling**: Centralized CORS policy at the proxy level
3. **SSL/TLS**: All service communication from the browser uses HTTPS
4. **Load Balancing**: Can easily add multiple backend instances per service
5. **Monitoring**: Centralized access logs for all service health checks
6. **Caching**: Can add caching layers at the proxy level if needed
7. **Rate Limiting**: Can implement rate limiting per service route
8. **Request Routing**: Can route to different backends based on headers, paths, etc.

## Files Modified

1. `/etc/nginx/sites-available/mm.itziklerner.com` - Added service proxy routes
2. `/home/mm/dev/b25/ui/web/src/components/ServiceMonitor.tsx` - Updated health check URLs
3. `/home/mm/dev/b25/ui/web/src/types/index.ts` - Added `changes` field to WebSocketMessage
4. `/home/mm/dev/b25/ui/web/dist/*` - Rebuilt frontend assets

## Files Created

1. `/home/mm/dev/b25/verify-service-routes.sh` - Service health verification script
2. `/home/mm/dev/b25/NGINX_SERVICE_ROUTING.md` - This documentation file

---

**Date:** 2025-10-06
**Status:** Implemented and Verified
**Next Steps:** Fix unhealthy services and implement missing health endpoints
