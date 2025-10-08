# Dashboard Server - Admin Page

## Overview

The Dashboard Server now includes a comprehensive admin page that provides real-time monitoring, health checks, and testing capabilities for the WebSocket aggregation service.

## Access

### Direct Access
- **URL**: `http://localhost:8086/` or `http://localhost:8086/admin`
- **Port**: 8086 (default HTTP port)

### Via Nginx Proxy
- **URL**: `http://your-domain/services/dashboard-server/`
- **Route**: `/services/dashboard-server/`

## Features

### 1. Real-time Service Monitoring
- **Service Uptime**: Shows how long the service has been running
- **Connected Clients**: Live count of active WebSocket connections
- **State Sequence**: Total number of state updates processed
- **Goroutines**: Current Go runtime concurrency metrics

### 2. Service Information Dashboard
Displays comprehensive service details:
- Service name and version (v1.0.0)
- Go runtime version
- CPU cores available
- WebSocket serialization format (JSON/MessagePack)
- Last state update timestamp

### 3. Aggregated State Metrics
Real-time counts of:
- Market Data streams (symbols being tracked)
- Active Orders in the system
- Open Positions across all accounts
- Active Strategies running

### 4. Backend Service Health
Visual indicators for all connected services:
- **Order Execution Service** (gRPC)
- **Strategy Engine** (HTTP)
- **Account Monitor** (gRPC)
- **Redis** (Pub/Sub & Cache)

Each service shows:
- Connection status (connected/disconnected)
- Visual pulse indicator
- Last check timestamp

### 5. Interactive API Testing
Built-in testing tools for all endpoints:
- **Test Health**: Checks `/health` endpoint
- **Test Debug**: Retrieves full state dump from `/debug`
- **Test History API**: Queries historical data endpoint
- **Test WebSocket**: Establishes live WebSocket connection and subscribes to all channels

Test results display:
- HTTP status codes
- Response data (formatted JSON)
- WebSocket message flow
- Real-time subscription confirmations

## API Endpoints

### Admin Endpoints

#### `GET /` or `GET /admin`
Returns the admin dashboard HTML page.

**Response**: HTML page with embedded CSS and JavaScript

#### `GET /api/service-info`
Returns comprehensive service information in JSON format.

**Response**:
```json
{
  "service": {
    "name": "dashboard-server",
    "version": "1.0.0",
    "uptime": "2h30m15s",
    "started": "2025-10-08T10:30:00Z"
  },
  "runtime": {
    "go_version": "go1.21.0",
    "num_goroutine": 45,
    "num_cpu": 8
  },
  "websocket": {
    "connected_clients": 3,
    "total_clients": 3,
    "format": "JSON/MessagePack"
  },
  "state": {
    "sequence": 12456,
    "last_update": "2025-10-08T13:00:15Z",
    "market_data_count": 5,
    "orders_count": 12,
    "positions_count": 3,
    "strategies_count": 2
  },
  "backend": {
    "order_execution": {
      "status": "connected",
      "last_checked": "2025-10-08T13:00:15Z"
    },
    "strategy_engine": {
      "status": "connected",
      "last_checked": "2025-10-08T13:00:15Z"
    },
    "account_monitor": {
      "status": "connected",
      "last_checked": "2025-10-08T13:00:15Z"
    },
    "redis": {
      "status": "connected",
      "last_checked": "2025-10-08T13:00:15Z"
    }
  },
  "health": {
    "status": "healthy",
    "checks": {
      "aggregator": "ok",
      "websocket": "ok",
      "redis": "ok"
    }
  }
}
```

### Existing Endpoints

All existing endpoints remain functional:

- `GET /health` - Health check
- `GET /debug` - Full state dump
- `GET /api/v1/history` - Historical data queries
- `GET /ws` - WebSocket connection
- `GET /metrics` - Prometheus metrics

## Auto-refresh

The admin page automatically refreshes service information every **5 seconds** to provide real-time monitoring without manual intervention.

## Design

### Dark Theme
The admin page features a modern dark theme with:
- Gradient backgrounds (slate/navy blue)
- Glass-morphism effects
- Smooth animations and transitions
- Responsive card-based layout
- Hover effects for better UX

### Color Coding
- **Green** (Connected/Healthy): #10b981
- **Yellow** (Warning): #fbbf24
- **Red** (Error/Disconnected): #ef4444
- **Blue** (Primary actions): #3b82f6
- **Purple** (Secondary actions): #8b5cf6

### Typography
- System font stack for optimal performance
- Clear hierarchy with varied sizes
- Monospace font for code/test results

## Architecture

### File Structure
```
dashboard-server/
├── internal/
│   ├── admin/
│   │   └── admin.go           # Admin handler with embedded HTML
│   ├── server/
│   │   └── server.go          # WebSocket server (added GetClientCount)
│   └── aggregator/
│       └── aggregator.go      # State aggregation
└── cmd/
    └── server/
        └── main.go            # Route registration
```

### Implementation Details

1. **Embedded HTML**: The entire admin page is embedded in the Go binary as a constant string, eliminating external dependencies.

2. **CORS Enabled**: All admin endpoints support CORS for development and testing.

3. **WebSocket Testing**: The built-in WebSocket tester demonstrates proper connection setup with:
   - Correct URL format (`ws://` protocol)
   - Query parameters for client type and format
   - Channel subscriptions
   - Message parsing

4. **Zero External Assets**: All CSS and JavaScript are inline - no CDN dependencies.

## Usage Examples

### Testing Health Endpoint
Click "Test Health" button to verify service health:
```json
{
  "status": "ok",
  "service": "dashboard-server"
}
```

### Testing WebSocket Connection
Click "Test WebSocket" to establish a connection and see:
1. Connection success message
2. Subscription confirmations
3. Snapshot data with counts
4. Auto-disconnect after 5 seconds

### Monitoring State Updates
Watch the "State Sequence" counter increment as:
- Market data updates arrive from Redis
- Orders are created/updated
- Strategies change status
- Positions are opened/closed

## Troubleshooting

### Admin Page Not Loading
- Verify service is running: `curl http://localhost:8086/health`
- Check port 8086 is not blocked
- Review service logs for startup errors

### Service Info Shows "Error"
- Check `/api/service-info` endpoint directly
- Verify all backend services are accessible
- Review network connectivity

### WebSocket Test Fails
- Ensure WebSocket endpoint is accessible
- Check allowed origins configuration
- Verify no firewall blocking WebSocket protocol

### Backend Service Shows Disconnected
- Verify the backend service is running
- Check connection strings in config
- Review service logs for connection errors

## Development

### Building
```bash
cd /home/mm/dev/b25/services/dashboard-server
make build
```

### Running Locally
```bash
./bin/dashboard-server
```

### Testing Admin Page
```bash
# Start service
./bin/dashboard-server

# In browser, navigate to:
http://localhost:8086/

# Or test via curl:
curl http://localhost:8086/api/service-info
```

## Security Considerations

1. **Production Deployment**: Consider adding authentication for the admin page in production
2. **API Key**: Configure `security.api_key` in config for WebSocket authentication
3. **Allowed Origins**: Restrict WebSocket origins in production config
4. **Network Access**: Use firewall rules to limit admin page access

## Future Enhancements

Potential improvements:
- Authentication/authorization layer
- Historical metrics charts
- Log streaming interface
- Configuration management UI
- Client connection details (IP, duration, subscriptions)
- Performance metrics graphs
- Alert configuration interface

## Credits

Design inspired by modern admin dashboards with focus on:
- Developer experience
- Real-time monitoring
- Interactive testing
- Mobile-responsive layouts
- Accessibility (WCAG compliance)
