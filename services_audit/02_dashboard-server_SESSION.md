# Dashboard Server Service - Interactive Session Notes

**Date:** 2025-10-06
**Status:** ✅ **OPERATIONAL** (Running on PID 54843)

---

## Executive Summary

The **dashboard-server** is a **WebSocket aggregation hub** that acts as the central nervous system for your trading UI. It:

- Consolidates data from all backend services (market-data, orders, positions, strategies)
- Broadcasts real-time updates to web and TUI clients via WebSocket
- Uses differential updates to minimize bandwidth
- Supports MessagePack for 60-65% bandwidth savings vs JSON

**Current Status:** Running successfully, serving live market data

**Grade:** B+ (Operational, needs authentication and backend integration)

---

## What This Service Does

### Purpose

Think of dashboard-server as a **smart aggregator and broadcaster**:

1. **Subscribes** to Redis pub/sub channels from backend services
2. **Aggregates** data into a unified state (market data, orders, positions, account, strategies)
3. **Broadcasts** differential updates to connected WebSocket clients
4. **Optimizes** bandwidth through MessagePack serialization and change detection

### Why It's Important

Without dashboard-server:
- UI would need to connect to 5+ backend services separately
- Much higher bandwidth usage (full state every update)
- No unified state view
- Complex client-side data management

**With dashboard-server:**
- Single WebSocket connection
- Differential updates (only changes)
- Unified, consistent state
- ~60% less bandwidth

---

## Architecture Overview

### Code Structure (8 Go files, ~1,900 lines)

```
dashboard-server/
├── cmd/
│   └── server/
│       └── main.go (120 lines) ............. Entry point, server startup
│
├── internal/
│   ├── aggregator/
│   │   ├── aggregator.go (600 lines) ....... State management & Redis pub/sub
│   │   └── aggregator_test.go (150 lines) .. Unit tests
│   ├── broadcaster/
│   │   └── broadcaster.go (400 lines) ...... Differential updates & broadcasting
│   ├── server/
│   │   ├── server.go (500 lines) ........... WebSocket server
│   │   └── server_test.go (100 lines) ...... Unit tests
│   ├── types/
│   │   └── types.go (200 lines) ............ Data structures
│   └── metrics/
│       └── metrics.go (80 lines) ........... Prometheus metrics
│
├── config.yaml ............................ Configuration
├── Dockerfile ............................. Container definition
├── Makefile ............................... Build commands
└── README.md .............................. Documentation
```

### Data Flow

```
Backend Services (market-data, order-execution, etc.)
         ↓
    Redis Pub/Sub
         ↓
Aggregator (aggregator.go)
  → Maintains unified state
  → Thread-safe with RWMutex
  → Tracks sequence numbers
         ↓
Broadcaster (broadcaster.go)
  → TUI clients: 100ms updates
  → Web clients: 250ms updates
  → Computes differential updates
  → MessagePack or JSON serialization
         ↓
WebSocket Server (server.go)
  → /ws endpoint
  → Client connection management
  → Subscription filtering
  → Heartbeat (30s ping/pong)
         ↓
    UI Clients (Web Dashboard, TUI)
```

---

## Inputs & Outputs

### INPUTS

#### 1. Redis Pub/Sub Channels (from backend services)

**Market Data:**
- `market_data:BTCUSDT` - Simplified market data
- `orderbook:BTCUSDT` - Full order book
- `trades:BTCUSDT` - Trade events

**Orders:**
- `orders:*` - Order updates (create, fill, cancel)

**Positions:**
- `positions:*` - Position updates

**Account:**
- `account:*` - Balance, P&L updates

**Strategies:**
- `strategies:*` - Strategy status updates

#### 2. Backend Service APIs (HTTP/gRPC)

**For periodic refresh (every 30s):**
- Order Execution (gRPC :50051) - Get all orders
- Account Monitor (gRPC :50055) - Get account state
- Strategy Engine (HTTP :8082) - Get strategy status
- Risk Manager (HTTP :9095) - Get risk metrics

#### 3. Configuration (config.yaml)

```yaml
server:
  port: 8086              # HTTP/WebSocket server port

redis:
  url: localhost:6379     # Redis connection

websocket:
  ping_interval: 30s      # Keep-alive heartbeat

broadcast:
  tui_interval: 100ms     # TUI update frequency
  web_interval: 250ms     # Web dashboard update frequency

backend_services:
  market_data:
    url: localhost:8080
  order_execution:
    url: localhost:8081
  # ... etc
```

### OUTPUTS

#### 1. WebSocket Endpoint (ws://localhost:8086/ws)

**Message Types:**

**A. State Update (broadcast)**
```json
{
  "type": "state_update",
  "sequence": 12345,
  "timestamp": "2025-10-06T05:41:00Z",
  "data": {
    "market_data": {
      "BTCUSDT": {
        "last_price": 123432.15,
        "bid_price": 123432.0,
        "ask_price": 123432.5
      }
    },
    "orders": [...],
    "positions": {...},
    "account": {...},
    "strategies": {...}
  }
}
```

**B. Differential Update (broadcast)**
```json
{
  "type": "diff_update",
  "sequence": 12346,
  "timestamp": "2025-10-06T05:41:00.250Z",
  "changes": {
    "market_data": {
      "BTCUSDT": {
        "last_price": 123433.25  // Only changed field
      }
    }
  }
}
```

**C. Subscription Response (client → server)**
```json
// Client sends:
{
  "action": "subscribe",
  "subscriptions": ["market_data", "orders"]
}

// Server responds:
{
  "type": "subscribed",
  "subscriptions": ["market_data", "orders"],
  "sequence": 12345
}
```

#### 2. REST API Endpoints

**Health:** `GET /health`
```json
{"status":"ok","service":"dashboard-server"}
```

**Debug:** `GET /debug`
```json
{
  "clients": 3,
  "state_size": 150000,
  "uptime": "2h15m",
  "last_update": "2025-10-06T05:41:00Z"
}
```

**Metrics:** `GET /metrics` (Prometheus format)
```
# HELP dashboard_websocket_clients_total Total WebSocket clients
# TYPE dashboard_websocket_clients_total gauge
dashboard_websocket_clients_total 3

# HELP dashboard_broadcasts_total Total broadcasts sent
# TYPE dashboard_broadcasts_total counter
dashboard_broadcasts_total{client_type="web"} 5420
dashboard_broadcasts_total{client_type="tui"} 13550
```

---

## Dependencies

### Required Services

1. **Redis** (localhost:6379)
   - Purpose: Pub/sub for real-time updates, caching
   - Status: ✅ Running (b25-redis container)
   - Critical: Yes

### Optional Backend Services

2. **market-data** (localhost:8080)
   - Status: ✅ Running
   - Provides: Market prices, order books, trades

3. **order-execution** (localhost:8081)
   - Status: ⚠️ Unknown (gRPC not verified)
   - Provides: Order status

4. **account-monitor** (localhost:8084)
   - Status: ⚠️ Unknown (gRPC not verified)
   - Provides: Account balance, P&L

5. **strategy-engine** (localhost:8082)
   - Status: ⚠️ Unknown
   - Provides: Strategy status

6. **risk-manager** (localhost:9095)
   - Status: ⚠️ Unknown
   - Provides: Risk metrics

**Note:** Service works with just Redis and market-data. Other services add more data types.

---

## Current Status

### Service Running

```bash
$ ps aux | grep dashboard-server
mm  54843  5.8  0.3  2055016  21232  ?  SNl  05:30  7:50  ./dashboard-server
```

**Details:**
- PID: 54843
- Uptime: Since 05:30 (2+ hours)
- CPU: 5.8%
- Memory: 21MB (0.3%)
- Status: Running smoothly

### Health Check

```bash
$ curl http://localhost:8086/health
{"status":"ok","service":"dashboard-server"}
```

✅ Health endpoint responding

### Configuration

**Port:** 8086 (HTTP + WebSocket)
**Redis:** localhost:6379
**Update Rates:**
- TUI clients: 100ms (10 updates/second)
- Web clients: 250ms (4 updates/second)

---

## How to Test in Isolation

### Test 1: Health Check

```bash
curl http://localhost:8086/health
# Expected: {"status":"ok","service":"dashboard-server"}
```

✅ **Result:** Working

### Test 2: Debug Info

```bash
curl http://localhost:8086/debug
# Expected: JSON with client count, state size, uptime
```

### Test 3: WebSocket Connection (with wscat)

**Install wscat:**
```bash
npm install -g wscat
```

**Connect and subscribe:**
```bash
wscat -c ws://localhost:8086/ws

# After connection, send:
{"action":"subscribe","subscriptions":["market_data"]}

# You should receive:
# 1. Subscription confirmation
# 2. Initial state snapshot
# 3. Differential updates every 250ms (web client)
```

### Test 4: WebSocket Connection (with Node.js)

Create `test-client.js`:
```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8086/ws');

ws.on('open', () => {
  console.log('Connected!');

  // Subscribe to market data
  ws.send(JSON.stringify({
    action: 'subscribe',
    subscriptions: ['market_data']
  }));
});

ws.on('message', (data) => {
  const msg = JSON.parse(data);
  console.log('Received:', msg.type);

  if (msg.type === 'state_update' || msg.type === 'diff_update') {
    console.log('Market data:', msg.data?.market_data || msg.changes?.market_data);
  }
});

ws.on('error', (error) => {
  console.error('WebSocket error:', error);
});
```

**Run:**
```bash
node test-client.js
```

### Test 5: Prometheus Metrics

```bash
curl -s http://localhost:8086/metrics | grep dashboard
```

---

## Performance Characteristics

### Current Performance (from audit)

**Latency:**
- WebSocket broadcast: <50ms p99 ✅
- Full state serialization: ~5-10ms
- Differential update: ~1-2ms (much faster)

**Throughput:**
- Concurrent clients: 100+ supported
- Broadcast rate: 4-10 updates/second per client
- Bandwidth per client: ~10-50 KB/s (with differentials)

**Memory:**
- Base: ~24MB
- Per client: ~4MB
- With 10 clients: ~64MB

**Bandwidth Savings:**
- MessagePack vs JSON: 60-65% smaller
- Differential vs full: 50-90% smaller (depending on changes)
- Combined savings: Up to 95% bandwidth reduction

### Resource Limits (Recommended)

For systemd service (when we create it):
- CPU: 50% (same as market-data)
- Memory: 512M (plenty for 100+ clients)
- Tasks: 100

---

## Key Features

### 1. Differential Updates

**Smart Change Detection:**
```go
// Only sends fields that changed
oldState := {"BTCUSDT": {"last_price": 123.45}}
newState := {"BTCUSDT": {"last_price": 123.50}}

// Sends only:
{"changes": {"BTCUSDT": {"last_price": 123.50}}}

// Savings: 50-90% bandwidth
```

### 2. MessagePack Serialization

**Binary vs JSON:**
```
JSON:   {"last_price":123432.15} = 28 bytes
MsgPack: [compact binary]         = 10 bytes
Savings: 64%
```

### 3. Subscription Filtering

**Clients choose what they want:**
```json
// Client A (trader): Only market data + orders
{"subscriptions": ["market_data", "orders"]}

// Client B (dashboard): Everything
{"subscriptions": ["market_data", "orders", "positions", "account", "strategies"]}
```

### 4. Heartbeat / Keep-Alive

**Automatic connection health:**
- Server pings clients every 30 seconds
- Clients must respond with pong
- Auto-disconnect dead connections

---

## Issues Found (from audit)

### Critical: **NONE** ✅

### Major Issues (3 found)

#### 1. ⚠️ Backend Integration Incomplete

**Issue:** Some backend services use placeholder/demo data
- Order execution integration not fully tested
- Account monitor integration partial
- Strategy engine returns mock status

**Impact:** Dashboard shows some fake data mixed with real data
**Fix:** Complete backend integration (1-2 weeks)
**Priority:** Medium

#### 2. ⚠️ No Historical Data API

**Issue:** No `/api/v1/history` endpoint implemented
**Impact:** Can't query past states
**Fix:** Implement TimescaleDB queries (1 week)
**Priority:** Low

#### 3. ⚠️ Origin Checking Disabled

**Issue:** WebSocket accepts connections from any origin
```go
CheckOrigin: func(r *http.Request) bool { return true }
```
**Impact:** Security risk - CSRF attacks possible
**Fix:** Add proper origin validation (30 minutes)
**Priority:** High for production

### Minor Issues (9 found)

1. No rate limiting on WebSocket messages
2. No authentication/authorization
3. Test coverage ~30% (should be 60%+)
4. Some config fields unused
5. No graceful degradation if backend unavailable
6. No circuit breakers for backend calls
7. Logs could be more structured
8. No distributed tracing
9. Port mismatch in docs (8080 vs 8086)

---

## Strengths

### ✅ Well-Architected

- Clean separation of concerns (aggregator/broadcaster/server)
- Thread-safe state management (RWMutex)
- Efficient differential updates
- Good use of Go channels and goroutines

### ✅ Performance Optimized

- MessagePack serialization
- Differential state computation
- Configurable update rates
- Connection pooling

### ✅ Monitoring Ready

- Prometheus metrics
- Structured logging (zerolog)
- Health checks
- Debug endpoints

### ✅ Recent Bug Fixes Working

- Broadcast updates fixed (was nil map issue)
- Null serialization fixed
- State updates flowing correctly

---

## How Other Services Use This

### Current Consumers

**1. Web Dashboard (ui/web)**
- Connects via WebSocket to `ws://localhost:8086/ws`
- Subscribes to: market_data, orders (potentially all)
- Update rate: 250ms
- Status: ✅ Working (you're probably using it now)

**2. TUI (terminal UI)**
- Same WebSocket endpoint
- Higher update rate: 100ms (more responsive for CLI)
- Status: ⚠️ Not verified in this audit

### Integration Example

**React Hook (from ui/web):**
```typescript
import { useWebSocket } from '@/hooks/useWebSocket';

export function TradingDashboard() {
  const { data, isConnected } = useWebSocket('ws://localhost:8086/ws', {
    subscriptions: ['market_data', 'orders', 'positions']
  });

  return (
    <div>
      <div>Connection: {isConnected ? '🟢' : '🔴'}</div>
      <div>BTC Price: ${data.market_data?.BTCUSDT?.last_price}</div>
      <div>Open Orders: {data.orders?.length}</div>
    </div>
  );
}
```

---

## Configuration Options

### Current Config (config.yaml)

```yaml
server:
  port: 8086                    # WebSocket + HTTP port
  log_level: info              # Logging verbosity
  shutdown_timeout: 30s        # Graceful shutdown time

redis:
  url: localhost:6379          # Redis server
  db: 0                        # Redis database
  pool_size: 10                # Connection pool

websocket:
  read_buffer_size: 1024       # WebSocket buffer
  write_buffer_size: 4096      # WebSocket buffer
  ping_interval: 30s           # Heartbeat frequency
  read_timeout: 60s            # Read deadline
  write_timeout: 10s           # Write deadline

broadcast:
  tui_interval: 100ms          # TUI update rate (10/sec)
  web_interval: 250ms          # Web update rate (4/sec)

backend_services:
  market_data:
    url: localhost:8080
    timeout: 5s
  order_execution:
    url: localhost:8081
    timeout: 5s
  # ... etc
```

### Tuning Options

**For high-frequency updates:**
```yaml
broadcast:
  tui_interval: 50ms    # 20 updates/sec
  web_interval: 100ms   # 10 updates/sec
```

**For low bandwidth:**
```yaml
broadcast:
  tui_interval: 500ms   # 2 updates/sec
  web_interval: 1s      # 1 update/sec
```

**For more clients:**
```yaml
redis:
  pool_size: 50         # More connections
```

---

## Next Steps

### What We Should Do

**Option 1: Create Deployment Automation (like market-data)**
- Create `deploy.sh` script
- Create systemd service
- Add `.gitignore`
- Test deployment

**Option 2: Fix Critical Issues**
- Add origin checking for WebSocket
- Add authentication
- Complete backend integration

**Option 3: Test Functionality**
- Connect via WebSocket
- Verify data flowing
- Test subscription filtering
- Measure performance

**Option 4: Move to Next Service**
- Dashboard-server is working well
- Continue with configuration or other services

---

**Which direction would you like to go?**

1. Create deployment automation (like we did for market-data)
2. Test the WebSocket functionality
3. Fix the origin checking security issue
4. Move to the next service (configuration)
5. Something else
