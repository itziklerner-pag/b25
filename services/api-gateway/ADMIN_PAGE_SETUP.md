# API Gateway - Admin Page Setup

## Overview
Comprehensive admin dashboard for the API Gateway service, providing real-time monitoring, configuration details, and API testing capabilities.

## Files Created

### 1. `/home/mm/dev/b25/services/api-gateway/internal/admin/admin.go`
**Purpose**: Backend handler logic for admin endpoints

**Key Features**:
- `HandleAdminPage()`: Serves the HTML admin interface
- `HandleServiceInfo()`: Returns JSON with service metadata
- Service information includes:
  - Version, uptime, port, mode
  - Go version, CPU cores, goroutines
  - Configuration details (services, auth, rate limits, CORS, etc.)
  - Feature flags

**Endpoints Added**:
- `GET /admin` - Admin dashboard UI
- `GET /` - Redirects to admin dashboard (default route)
- `GET /api/service-info` - Service metadata JSON API

### 2. `/home/mm/dev/b25/services/api-gateway/internal/admin/page.go`
**Purpose**: HTML/CSS/JavaScript embedded admin interface

**UI Features**:
- **Dark Theme**: Modern gradient design matching account-monitor style
- **Real-time Monitoring**: Auto-refreshes every 5 seconds
- **Status Dashboard**: Shows health, version, uptime, port, goroutines
- **Service Information**: Runtime details (Go version, CPU, start time)
- **Backend Services**: Lists all 7 backend microservices with URLs and timeouts
- **Configuration Panel**: Displays all gateway settings
- **Features Display**: Shows enabled/disabled feature flags
- **Endpoints List**: Comprehensive list of all available API routes
- **Interactive API Tester**: Test any endpoint directly from the dashboard
  - Supports GET, POST, PUT, DELETE methods
  - JSON response viewer
  - Auto-formatted output

**Design Highlights**:
- Responsive grid layout
- Glass morphism effects with backdrop blur
- Gradient accents (blue to purple)
- Smooth animations and transitions
- Mobile-optimized
- Custom scrollbar styling

### 3. `/home/mm/dev/b25/services/api-gateway/internal/router/router.go`
**Updated**: Added admin handler integration

**Changes**:
- Imported admin package
- Initialized admin handler
- Registered admin routes (public, no auth required for internal access)
- Routes added before other groups to ensure priority

## Configuration

### Backend Services Displayed:
1. **Market Data** - Market data aggregation service
2. **Order Execution** - Trade execution service
3. **Strategy Engine** - Trading strategy management
4. **Account Monitor** - Account state tracking
5. **Dashboard Server** - Real-time dashboard WebSocket server
6. **Risk Manager** - Risk management and limits
7. **Configuration** - Centralized configuration service

### Gateway Features Monitored:
- Authentication (JWT/API Keys)
- Rate Limiting (Global, Per-IP, Per-Endpoint)
- CORS configuration
- Circuit Breaker status
- Redis caching
- WebSocket support
- Feature flags (Tracing, Compression, Request ID, Access Log, Error Details)

## API Endpoints Documented

The admin page displays all 20+ available endpoints including:

**Public Routes**:
- `/health` - Health check
- `/health/liveness` - Kubernetes liveness probe
- `/health/readiness` - Kubernetes readiness probe
- `/metrics` - Prometheus metrics
- `/version` - Service version

**Market Data** (v1):
- `GET /api/v1/market-data/symbols`
- `GET /api/v1/market-data/orderbook/:symbol`
- `GET /api/v1/market-data/trades/:symbol`
- `GET /api/v1/market-data/ticker/:symbol`

**Orders** (Auth Required):
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders` - List orders
- `GET /api/v1/orders/:id` - Get order details
- `DELETE /api/v1/orders/:id` - Cancel order
- `GET /api/v1/orders/active` - Active orders
- `GET /api/v1/orders/history` - Order history

**Strategies** (Auth Required):
- `GET /api/v1/strategies` - List strategies
- `POST /api/v1/strategies` - Create strategy (admin only)
- `GET /api/v1/strategies/:id` - Get strategy
- `PUT /api/v1/strategies/:id` - Update strategy (admin only)
- `DELETE /api/v1/strategies/:id` - Delete strategy (admin only)
- `POST /api/v1/strategies/:id/start` - Start strategy
- `POST /api/v1/strategies/:id/stop` - Stop strategy
- `GET /api/v1/strategies/:id/status` - Strategy status

**Account**:
- `GET /api/v1/account/balance`
- `GET /api/v1/account/positions`
- `GET /api/v1/account/pnl`
- `GET /api/v1/account/pnl/daily`
- `GET /api/v1/account/trades`

**Risk** (Auth Required):
- `GET /api/v1/risk/limits`
- `PUT /api/v1/risk/limits` (admin only)
- `GET /api/v1/risk/status`
- `POST /api/v1/risk/emergency-stop` (admin only)

**WebSocket**:
- `WS /ws` - Real-time updates

## Build Instructions

```bash
cd /home/mm/dev/b25/services/api-gateway
go build -o bin/api-gateway ./cmd/server
```

## Access Points

### Via Nginx (Production):
```
http://your-domain/services/api-gateway/
http://your-domain/services/api-gateway/admin
http://your-domain/services/api-gateway/api/service-info
```

### Direct Access (Development):
```
http://localhost:8000/
http://localhost:8000/admin
http://localhost:8000/api/service-info
```

## JavaScript Configuration

The admin page uses the following API base URL configuration:
```javascript
const API_BASE = window.location.origin + '/services/api-gateway';
```

This ensures the page works correctly behind nginx reverse proxy.

## Auto-Refresh

The service information automatically refreshes every 5 seconds via:
```javascript
setInterval(loadServiceInfo, 5000);
```

This keeps the dashboard up-to-date with:
- Current uptime
- Active goroutines
- Service health status
- Real-time configuration

## Styling Details

### Color Palette:
- **Background**: Dark gradient (#0f172a to #1e293b)
- **Cards**: Translucent dark (#1e293b with 80% opacity)
- **Accents**: Blue to purple gradient (#3b82f6 to #8b5cf6)
- **Text**: Light gray (#e2e8f0)
- **Secondary**: Muted gray (#94a3b8)

### Typography:
- **Font**: System fonts (-apple-system, Segoe UI, Roboto)
- **Code**: Monaco, Menlo (monospace)

### Components:
- **Status Badges**: Color-coded (green=healthy, red=error, amber=warning)
- **Method Tags**: HTTP method color coding
- **Feature Badges**: Enabled (green) / Disabled (gray)
- **Loading States**: Animated spinner

## Testing the API Tester

1. Navigate to the admin page
2. Scroll to "API Tester" section
3. Select HTTP method (GET/POST/PUT/DELETE)
4. Enter endpoint path (e.g., `/health`)
5. Click "Send Request"
6. View formatted JSON response

Example test:
```
Method: GET
Endpoint: /health
Response: { "status": "ok", ... }
```

## Security Notes

- Admin page is currently public (no authentication)
- Designed for internal network access only
- Sensitive config values are masked (e.g., Redis passwords)
- API testing respects gateway authentication when configured

## Next Steps (Optional Enhancements)

1. Add authentication to admin page
2. Add real-time metrics charts
3. Add request/response logging viewer
4. Add circuit breaker status per service
5. Add cache hit/miss statistics
6. Add WebSocket connection monitor
7. Add performance profiling endpoint
8. Add configuration hot-reload button

## Troubleshooting

### Admin page not loading
1. Check if service is running: `ps aux | grep api-gateway`
2. Check if port 8000 is accessible
3. Check nginx configuration for `/services/api-gateway/` route

### Service info returns 404
1. Verify routes are registered in router.go
2. Check if admin handler is initialized
3. Rebuild the service after changes

### Styles not rendering
1. Check browser console for errors
2. Verify HTML is being served correctly
3. Check Content-Type header is `text/html`

## Maintenance

- Update version constant in `/internal/admin/admin.go` when releasing
- Add new endpoints to the endpoints list in page.go
- Update service configurations when adding new backend services
- Keep auto-refresh interval reasonable to avoid overwhelming the service

## Performance Considerations

- HTML/CSS/JS is embedded in binary (single file deployment)
- No external dependencies or CDNs
- Minimal JavaScript (vanilla, no frameworks)
- Auto-refresh is lightweight (only fetches JSON)
- Response caching handled by browser

## Compatibility

- Works with all modern browsers
- Mobile responsive design
- No external dependencies
- Compatible with nginx reverse proxy
- CORS-friendly
