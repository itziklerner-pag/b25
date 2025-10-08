# Risk Manager Service - Admin Page Setup Summary

## Completed Tasks

### 1. Admin Page Created
**File**: `/home/mm/dev/b25/services/risk-manager/admin/index.html`

Modern, responsive admin dashboard featuring:
- Real-time service health status badge
- Risk metrics display cards (Leverage, Margin Ratio, Daily Drawdown, Max Drawdown)
- Account information cards (Total Equity, Position Notional, Active Violations, Uptime)
- Active risk policies section with visual type indicators (Hard, Soft, Emergency)
- Recent violations log
- Service information panel
- Auto-refresh every 5 seconds
- Beautiful gradient background with glassmorphism design
- Mobile-responsive layout

### 2. HTTP Routes Registered
Updated `/home/mm/dev/b25/services/risk-manager/cmd/server/main.go` to add:

**New Endpoints**:
- `GET /` - Admin dashboard (root path)
- `GET /admin` - Admin dashboard (explicit path)
- `GET /api/service-info` - Service metadata endpoint
- `GET /health` - Health check endpoint (already existed, enhanced with CORS)
- `GET /metrics` - Prometheus metrics (already existed)

**API Base**: `window.location.origin + '/services/risk-manager'`

### 3. Service Info Endpoint
Returns JSON with:
```json
{
  "name": "Risk Manager Service",
  "version": "1.0.0",
  "port": 9095,
  "grpc_port": 50052,
  "environment": "development",
  "start_time": "2025-10-08T...",
  "uptime": "..."
}
```

## Critical Startup Issues Identified

### Issue 1: Missing Database Tables
**Error**: `pq: relation "risk_policies" does not exist`

**Impact**:
- Service cannot load risk policies from database
- Falls back to default hardcoded policies
- Cannot record violations to database

**Log Evidence**:
```
Line 4: failed to load policies from database - error: "pq: relation \"risk_policies\" does not exist"
Line 10: failed to record violation - error: "pq: relation \"risk_violations\" does not exist"
```

**Resolution Required**:
```bash
cd /home/mm/dev/b25/services/risk-manager
make migrate-up
```

This will create the required database schema:
- `risk_policies` - Policy definitions table
- `risk_violations` - Violation records table
- Other risk management tables

### Issue 2: Emergency Stop Triggered
**Error**: Circuit breaker tripped, emergency stop activated

**Impact**:
- Trading halted
- Emergency stop is active and preventing normal operation
- Repeated violation attempts due to missing database tables

**Log Evidence**:
```
Line 23: circuit breaker tripped
Line 25: circuit breaker tripped - triggering emergency stop
Line 26: EMERGENCY STOP TRIGGERED - reason: "Circuit breaker tripped due to repeated violations"
```

**Root Cause**:
The missing database tables cause every policy check to fail when trying to record violations. The repeated failures trigger the circuit breaker, which activates the emergency stop mechanism.

**Resolution**:
1. Run database migrations (fixes root cause)
2. Restart the service (clears emergency stop state)

### Issue 3: Mock Data Mode
**Warning**: Account Monitor client not configured

**Impact**:
- Service is using mock account data
- Not safe for production use
- Risk calculations based on dummy values

**Log Evidence**:
```
Line 91-93: Account Monitor client not configured - using mock data - NOT SAFE FOR PRODUCTION
```

**Resolution**:
This is a known limitation. The service requires the Account Monitor service to have proper protobuf definitions before real integration can be enabled.

## Current Service Status

### Running Components
- gRPC Server: Port 50052 (Running)
- HTTP Metrics Server: Port 9095 (Running)
- Risk Monitor: Running (but triggering violations)
- NATS Connection: Connected
- Redis Connection: Connected
- PostgreSQL Connection: Connected

### Failed Components
- Database schema not initialized
- Emergency stop active
- Policy enforcement degraded (using defaults only)

## How to Fix and Deploy

### Step 1: Run Database Migrations
```bash
cd /home/mm/dev/b25/services/risk-manager
make migrate-up
```

### Step 2: Build the Service (if needed)
```bash
make build
```

The binary will be created at: `/home/mm/dev/b25/services/risk-manager/bin/risk-manager`

### Step 3: Restart the Service
```bash
# Using your deployment system
sudo systemctl restart risk-manager

# Or if using supervisorctl
sudo supervisorctl restart risk-manager
```

### Step 4: Verify Admin Page
Navigate to: `http://your-server/services/risk-manager/` or `http://your-server/services/risk-manager/admin`

### Step 5: Check Health
```bash
curl http://localhost:9095/health
curl http://localhost:9095/api/service-info
```

## Admin Page Features

### Risk Metrics Display
- **Leverage**: Current account leverage ratio with color coding
  - Green: < 5.0x (low risk)
  - Yellow: 5.0-10.0x (moderate)
  - Red: > 10.0x (high risk)

- **Margin Ratio**: Available margin buffer
  - Green: > 1.5 (healthy)
  - Yellow: 1.0-1.5 (warning)
  - Red: < 1.0 (critical)

- **Daily Drawdown**: Loss from daily start
  - Green: < 5%
  - Yellow: 5-10%
  - Red: > 10%

- **Max Drawdown**: Loss from peak equity
  - Green: < 12.5%
  - Yellow: 12.5-25%
  - Red: > 25% (emergency threshold)

### Active Policies Display
Shows all configured risk policies with:
- Policy name
- Type badge (Hard/Soft/Emergency)
- Metric being monitored
- Operator (less_than, greater_than, etc.)
- Threshold value
- Scope (account/symbol/strategy)

### Default Policies
1. **Max Leverage Limit** (Hard)
   - Metric: leverage
   - Threshold: ≤ 10.0x
   - Blocks orders exceeding leverage

2. **Min Margin Ratio** (Hard)
   - Metric: margin_ratio
   - Threshold: ≥ 1.0
   - Prevents margin calls

3. **Max Drawdown Emergency Stop** (Emergency)
   - Metric: drawdown_max
   - Threshold: > 25%
   - Triggers emergency trading halt

### Auto-Refresh
The admin page automatically refreshes every 5 seconds to display:
- Current risk metrics
- Service health status
- Active violations
- System uptime

## Architecture Integration

### Nginx Routing
The admin page is accessible via the nginx reverse proxy at:
- `/services/risk-manager/` - Admin dashboard
- `/services/risk-manager/admin` - Admin dashboard
- `/services/risk-manager/health` - Health check
- `/services/risk-manager/api/service-info` - Service info
- `/services/risk-manager/metrics` - Prometheus metrics

### Service Ports
- **HTTP**: 9095 (metrics, health, admin)
- **gRPC**: 50052 (risk checking, emergency control)

### Dependencies
- PostgreSQL (risk policies, violations)
- Redis (policy cache, market prices)
- NATS (alerts, emergency broadcasts)
- Account Monitor gRPC (when available)

## Next Steps

1. **Immediate**: Run database migrations to fix the critical database table errors
2. **Immediate**: Restart service to clear emergency stop state
3. **Short-term**: Implement real Account Monitor integration when protobuf is ready
4. **Future**: Add more advanced API endpoints for risk metrics retrieval
5. **Future**: Add violation history API for the admin dashboard

## Known Limitations

1. Risk metrics currently display mock data (will show real data once Account Monitor is integrated)
2. Violation log shows placeholder text (will populate from database after migrations)
3. Emergency stop can only be cleared by service restart (manual re-enable endpoint not yet exposed via HTTP)

## Files Modified

1. `/home/mm/dev/b25/services/risk-manager/admin/index.html` - Created
2. `/home/mm/dev/b25/services/risk-manager/cmd/server/main.go` - Updated (added HTTP routes)

## Testing Checklist

- [ ] Admin page loads at `/services/risk-manager/`
- [ ] Health check returns 200 OK
- [ ] Service info endpoint returns JSON
- [ ] Status badge shows "Service Healthy"
- [ ] Risk metrics display (with mock data initially)
- [ ] Policies section shows 3 default policies
- [ ] Auto-refresh works every 5 seconds
- [ ] Mobile responsive layout works
- [ ] No console errors in browser

---

**Status**: Admin page created and routes registered. Service requires database migrations before full functionality.

**Next Action**: Run `make migrate-up` to initialize database schema, then restart the service.
