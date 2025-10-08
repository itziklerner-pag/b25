# Account Monitor Service - Admin Page

## Overview
The account-monitor service now includes a comprehensive admin page for monitoring, testing, and managing the service.

## Access
- **Main URL**: http://localhost:8087/
- **Admin URL**: http://localhost:8087/admin
- **Health API**: http://localhost:8087/health
- **Service Info API**: http://localhost:8087/api/service-info

## Features

### 1. Real-time Health Monitoring
- **Service Status**: Overall health status (healthy/degraded/unhealthy)
- **Component Checks**:
  - Database (PostgreSQL) connection status
  - Redis connection status
  - WebSocket connection to Binance
- **Auto-refresh**: Updates every 5 seconds

### 2. Service Information Dashboard
- Service version
- Uptime tracking
- Port information (HTTP: 8087, gRPC: 50054, Metrics: 9094)
- Configuration details

### 3. Complete API Endpoint Documentation
All available endpoints are listed and documented:
- `GET /health` - Health check with detailed component status
- `GET /ready` - Readiness probe for Kubernetes
- `GET /api/account` - Account state information
- `GET /api/positions` - Current trading positions
- `GET /api/pnl` - Profit & Loss calculations
- `GET /api/balance` - Account balance details
- `GET /api/alerts` - Active system alerts
- `WS /ws` - WebSocket stream for real-time updates
- `GET /metrics` - Prometheus metrics (port 9094)

### 4. Interactive Testing
One-click testing for all endpoints:
- Test individual endpoints with instant results
- View JSON responses in formatted output
- Track test execution history
- Clear output console
- Error handling and display

### 5. Visual Design
- Modern dark theme UI
- Color-coded status indicators
- Responsive grid layout
- Real-time status badges
- Live uptime tracking

## Service Status

### Current Status: ⚠️ DEGRADED

**What's Working:**
- ✅ HTTP Server (port 8087)
- ✅ gRPC Server (port 50054)
- ✅ Metrics Server (port 9094)
- ✅ Database Connection (PostgreSQL)
- ✅ Redis Connection
- ✅ NATS Messaging
- ✅ Reconciliation Engine
- ✅ Alert Manager
- ✅ Position Tracking
- ✅ Balance Monitoring

**What's Not Working:**
- ⚠️ WebSocket Connection to Binance (handshake failing)
  - This is likely due to testnet/production API endpoint mismatch
  - Service continues to operate in degraded mode
  - Reconciliation and monitoring still functional

## Configuration

The service is configured to run with:
- **Exchange**: Binance (Production)
- **HTTP Port**: 8087
- **gRPC Port**: 50054
- **Metrics Port**: 9094
- **Reconciliation Interval**: 5 seconds
- **Alerts**: Enabled
- **Dashboard**: Enabled

## Deployment Notes

### Starting the Service
```bash
BINANCE_API_KEY='your_api_key' \
BINANCE_SECRET_KEY='your_secret_key' \
POSTGRES_PASSWORD='your_postgres_password' \
./bin/account-monitor
```

### Required Environment Variables
- `BINANCE_API_KEY` - Binance API key
- `BINANCE_SECRET_KEY` - Binance secret key
- `POSTGRES_PASSWORD` - PostgreSQL database password

### Dependencies
- PostgreSQL (port 5433, database: b25_timeseries)
- Redis (port 6379)
- NATS (port 4222)

## Integration with Other Services

The account-monitor service integrates with:
- **Order Execution Service** - Receives fill events
- **Strategy Engine** - Provides position and P&L data
- **Dashboard Server** - Streams real-time account state
- **API Gateway** - Exposed via gateway on port 8000

## API Gateway Configuration

Updated API Gateway to use correct port:
```yaml
services:
  account_monitor:
    url: "http://localhost:8087"
    timeout: 5s
    max_retries: 3
```

## Next Steps

To fully integrate the admin page with your settings page:

1. **Add Link to Settings Page**: Add a service card in your settings/admin interface that links to http://localhost:8087/admin

2. **Fix WebSocket Connection**: Update the Binance WebSocket configuration to use the correct endpoint based on whether you're using testnet or production keys

3. **Set up Reverse Proxy**: If accessing from a different host, configure nginx/traefik to proxy to the admin page

4. **Add Authentication**: Consider adding API key authentication to the admin page for production use

## Troubleshooting

### Admin Page Not Loading
```bash
# Check if service is running
ps aux | grep account-monitor

# Check logs
tail -f /tmp/account-monitor-clean.log

# Verify port is listening
lsof -i :8087
```

### Health Check Failing
```bash
# Test health endpoint
curl http://localhost:8087/health

# Check database connection
export PGPASSWORD='L9JYNAeS3qdtqa6CrExpMA=='
psql -h localhost -p 5433 -U b25 -d b25_timeseries -c "SELECT 1"

# Check Redis connection
redis-cli -h localhost -p 6379 ping
```

### WebSocket Issues
The WebSocket connection to Binance may fail due to:
- Production keys being used with testnet endpoint (or vice versa)
- Network/firewall restrictions
- Invalid API keys

To fix, update `config.yaml`:
```yaml
exchange:
  testnet: true  # Set to true for testnet, false for production
```

## Screenshots

The admin page includes:
- Header with real-time status badge
- Service information card with version and uptime
- Health checks card with component status
- Configuration details card
- Complete endpoint documentation
- Interactive testing section with live output

---

**Created**: 2025-10-08
**Service Version**: 1.0.0
**Admin Page**: http://localhost:8087/
