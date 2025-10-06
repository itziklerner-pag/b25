# Risk Manager Service - Technical Audit Report

**Service**: Risk Manager
**Location**: `/home/mm/dev/b25/services/risk-manager`
**Language**: Go 1.22+
**Status**: Production Ready (v1.0.0)
**Audit Date**: 2025-10-06
**Last Updated**: 2025-10-03

---

## Executive Summary

The Risk Manager Service is the global risk management and emergency control system for the B25 trading platform. It provides real-time risk monitoring, multi-layer limit enforcement, pre-trade validation, and emergency stop mechanisms to protect against excessive losses. The service is designed for low-latency operations (<10ms p99 for pre-trade checks) and high reliability.

**Key Strengths**:
- Well-architected with clear separation of concerns
- Fast pre-trade validation with caching
- Multi-layer policy system (hard, soft, emergency)
- Circuit breaker pattern for automatic emergency stops
- Comprehensive metrics and monitoring
- Emergency stop functionality with position unwinding support

**Critical Issues**:
- Uses mock account state data instead of real Account Monitor integration
- Missing actual integration with market data service for live prices
- No tests found in codebase (0% test coverage)
- Database migrations exist but no connection pooling optimization
- Dockerfile has merge conflict markers (Git merge not completed)

---

## 1. Purpose

The Risk Manager Service serves as the **central risk control system** for the B25 trading platform with the following responsibilities:

### Primary Functions
1. **Pre-trade Risk Validation**: Fast gRPC endpoint (<10ms p99) that validates orders against risk policies before they are submitted to exchanges
2. **Real-time Risk Monitoring**: Continuous monitoring (1s interval) of account metrics against configured policies
3. **Policy Enforcement**: Multi-layer hierarchical policy system with hard, soft, and emergency violation types
4. **Emergency Stop Control**: Automatic circuit breaker that triggers emergency stops on critical violations and coordinates position unwinding
5. **Alert Publishing**: Real-time alert distribution via NATS for risk violations and emergency events
6. **Risk Metrics Calculation**: Computes key risk metrics including leverage, margin ratio, drawdown, and position concentration

### Business Logic
- Blocks orders that would violate hard policies (e.g., max leverage, min margin ratio)
- Issues warnings for soft policy violations (logged but not blocking)
- Triggers emergency stops for critical violations (e.g., max drawdown exceeded)
- Maintains circuit breaker that triggers emergency stop after 5 violations in 1 minute
- Provides gRPC API for other services to check order risk before execution

---

## 2. Technology Stack

### Core Technologies
- **Language**: Go 1.22+
- **RPC Framework**: gRPC with Protocol Buffers v3
- **HTTP Server**: Go standard library (metrics/health endpoints)
- **Concurrency**: Go goroutines with sync primitives

### Dependencies (from go.mod)

#### Database & Storage
- **PostgreSQL**: `github.com/lib/pq` v1.10.9 - Risk policies, violations, emergency stops
- **SQLx**: `github.com/jmoiron/sqlx` v1.3.5 - Database operations
- **Redis**: `github.com/redis/go-redis/v9` v9.4.0 - Policy cache, market prices, account state

#### Messaging
- **NATS**: `github.com/nats-io/nats.go` v1.31.0 - Alert publishing, metrics broadcasting

#### Configuration & Logging
- **Viper**: `github.com/spf13/viper` v1.18.2 - Configuration management
- **Zap**: `go.uber.org/zap` v1.26.0 - Structured logging

#### Monitoring
- **Prometheus**: `github.com/prometheus/client_golang` v1.18.0 - Metrics collection

#### gRPC
- **gRPC**: `google.golang.org/grpc` v1.60.1
- **Protobuf**: `google.golang.org/protobuf` v1.32.0

#### Utilities
- **UUID**: `github.com/google/uuid` v1.5.0 - Unique identifiers

### Infrastructure
- **Container**: Docker with multi-stage builds
- **Orchestration**: Docker Compose (includes PostgreSQL, Redis, NATS, Prometheus, Grafana)
- **Hot Reload**: Air (development mode)

---

## 3. Data Flow

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Risk Manager Service                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐      ┌──────────────────┐               │
│  │  Pre-Trade API   │      │  Emergency Stop  │               │
│  │   (gRPC Server)  │      │    Controller    │               │
│  └────────┬─────────┘      └────────┬─────────┘               │
│           │                          │                          │
│           v                          v                          │
│  ┌──────────────────────────────────────────┐                  │
│  │      Risk Calculation Engine              │                  │
│  │  - Margin Ratio    - Drawdown            │                  │
│  │  - Leverage        - Position Limits     │                  │
│  │  - Concentration   - Velocity Limits     │                  │
│  └──────────────┬───────────────────────────┘                  │
│                 │                                               │
│                 v                                               │
│  ┌──────────────────────────────────────────┐                  │
│  │      Policy Enforcement Engine            │                  │
│  │  - Multi-layer Rules                      │                  │
│  │  - Hierarchical Limits                    │                  │
│  │  - Conditional Logic                      │                  │
│  └──────────────┬───────────────────────────┘                  │
│                 │                                               │
│  ┌──────────────┴───────────────────────────┐                  │
│  │      Real-Time Monitor                    │                  │
│  │  - Position Tracker   - Alert Manager     │                  │
│  │  - Limit Watcher      - Violation Logger  │                  │
│  └───────────────────────────────────────────┘                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
           │                     │                     │
           v                     v                     v
    ┌──────────┐         ┌──────────┐         ┌──────────┐
    │PostgreSQL│         │  Redis   │         │   NATS   │
    │ Policies │         │  Cache   │         │  Alerts  │
    └──────────┘         └──────────┘         └──────────┘
```

### Detailed Data Flow

#### 1. Pre-Trade Order Validation (gRPC CheckOrder)

```
Order Execution Service
        │
        │ 1. gRPC CheckOrder(OrderRiskRequest)
        v
   Risk Manager
        │
        │ 2. Check emergency stop status
        v
   Emergency Stop Manager
        │
        │ 3. Get current market price
        v
   Market Price Cache (Redis)
        │
        │ 4. Get account state (MOCK - should be Account Monitor)
        v
   Mock Account State
        │
        │ 5. Simulate order impact on account
        v
   Risk Calculator
        │
        │ 6. Calculate post-trade metrics
        v
   Risk Metrics (leverage, margin ratio, drawdown, etc.)
        │
        │ 7. Evaluate against policies
        v
   Policy Engine
        │
        │ 8. Filter hard/emergency violations
        │
        v
   OrderRiskResponse (approved: bool, violations: []string)
        │
        v
   Order Execution Service
```

**Processing Steps**:
1. Receive order details (symbol, side, quantity, price, account_id, strategy_id)
2. Check if emergency stop is active (blocks all orders if true)
3. Fetch current market price from Redis cache (or use order price for limit orders)
4. Get current account state (CURRENTLY USING MOCK DATA)
5. Simulate order execution to calculate post-trade state
6. Calculate post-trade risk metrics (leverage, margin ratio, drawdown, concentration)
7. Evaluate metrics against all applicable policies (account, symbol, strategy scopes)
8. Filter only hard and emergency violations (soft violations are warnings only)
9. Return approval decision with violations and processing time (<10ms target)

#### 2. Real-Time Risk Monitoring (Background Loop)

```
Risk Monitor (1s interval)
        │
        │ 1. Get account state (MOCK)
        v
   Mock Account State
        │
        │ 2. Calculate current risk metrics
        v
   Risk Calculator
        │
        │ 3. Publish metrics to NATS
        v
   NATS (risk.metrics topic)
        │
        │ 4. Evaluate against all policies
        v
   Policy Engine
        │
        │ 5. Group violations by type
        v
   Emergency/Hard/Soft Violations
        │
        ├─→ Emergency: Trigger emergency stop → NATS alert
        │
        ├─→ Hard: Log + Publish alert + Circuit breaker
        │
        └─→ Soft: Log + Publish warning alert
```

**Processing Steps**:
1. Every 1 second, fetch current account state (MOCK)
2. Calculate real-time risk metrics
3. Publish metrics to NATS for dashboard consumption
4. Evaluate metrics against all active policies
5. Handle violations based on type:
   - **Emergency**: Trigger emergency stop immediately, publish critical alert
   - **Hard**: Log violation, publish critical alert, record in circuit breaker
   - **Soft**: Log violation, publish warning alert
6. Circuit breaker triggers emergency stop if 5 violations in 1 minute

#### 3. Emergency Stop Flow

```
Trigger Source (Policy Violation / Manual / Circuit Breaker)
        │
        │ 1. Trigger emergency stop
        v
   Emergency Stop Manager
        │
        │ 2. Update status (is_stopped = true)
        │
        ├─→ 3. Publish emergency alert to NATS
        │
        └─→ 4. Block all future orders
                │
                │ 5. Should trigger position unwinding (NOT IMPLEMENTED)
                │
                v
           Order Execution Service (should cancel orders/close positions)
```

#### 4. Policy Refresh Loop (30s interval)

```
Policy Refresh Timer (30s)
        │
        │ 1. Load active policies from PostgreSQL
        v
   Policy Repository
        │
        │ 2. Update Redis cache
        v
   Policy Cache
        │
        │ 3. Update in-memory policy engine
        v
   Policy Engine (reloaded policies)
```

---

## 4. Inputs

### 4.1 gRPC Inputs (Port 50052)

#### CheckOrder - Pre-trade validation
**Protobuf Message**: `OrderRiskRequest`
```protobuf
message OrderRiskRequest {
  string order_id = 1;       // Unique order identifier
  string symbol = 2;         // Trading pair (e.g., "BTCUSDT")
  string side = 3;           // "BUY" or "SELL"
  double quantity = 4;       // Order quantity
  double price = 5;          // Order price (0 for market orders)
  string order_type = 6;     // "MARKET" or "LIMIT"
  string strategy_id = 7;    // Strategy identifier
  string account_id = 8;     // Account identifier
  int64 timestamp = 9;       // Order timestamp
}
```

**Caller**: Order Execution Service (before submitting orders to exchange)

#### GetRiskMetrics - Current risk metrics
**Protobuf Message**: `RiskMetricsRequest`
```protobuf
message RiskMetricsRequest {
  string account_id = 1;     // Account to get metrics for
}
```

**Caller**: Dashboard, monitoring systems

#### TriggerEmergencyStop - Manual emergency stop
**Protobuf Message**: `EmergencyStopRequest`
```protobuf
message EmergencyStopRequest {
  string reason = 1;         // Reason for emergency stop
  string triggered_by = 2;   // Who triggered it (user/system)
  bool force = 3;            // Force stop even if already stopped
}
```

**Caller**: Dashboard admin panel, monitoring systems

#### ReEnableTrading - Resume after emergency stop
**Protobuf Message**: `ReEnableTradingRequest`
```protobuf
message ReEnableTradingRequest {
  string authorized_by = 1;  // Who authorized re-enabling
  string reason = 2;         // Reason for re-enabling
}
```

**Caller**: Dashboard admin panel (requires manual authorization)

### 4.2 Database Inputs (PostgreSQL)

**Connection**: `postgresql://b25:password@localhost:5432/b25_config`

#### Risk Policies (risk_policies table)
```sql
- id: UUID (primary key)
- name: VARCHAR(255) - Policy name
- type: 'hard' | 'soft' | 'emergency'
- metric: 'leverage' | 'margin_ratio' | 'drawdown_daily' | 'drawdown_max' | 'concentration_<symbol>'
- operator: 'less_than' | 'greater_than' | 'equal' | etc.
- threshold: NUMERIC(20,8) - Limit value
- scope: 'account' | 'symbol' | 'strategy'
- scope_id: VARCHAR(100) - Symbol or strategy ID (null for account scope)
- enabled: BOOLEAN
- priority: INTEGER (higher priority evaluated first)
```

**Default Policies Loaded**:
1. Max Account Leverage: ≤ 10.0x (hard)
2. Min Margin Ratio: ≥ 1.0 (hard)
3. Daily Drawdown Warning: > 10% (soft)
4. Max Drawdown Hard Limit: > 20% (hard)
5. Emergency Drawdown Stop: > 25% (emergency)

#### Violation Recording (risk_violations table)
Records historical violations for analysis

#### Emergency Stops (emergency_stops table)
Records emergency stop events with snapshots

### 4.3 Cache Inputs (Redis)

**Connection**: `localhost:6379` (DB 0 for policies, DB 1 for market data)

#### Policy Cache
- **Key**: `risk:policies`
- **Type**: JSON array of policies
- **TTL**: 1 second
- **Fallback**: Local in-memory cache + PostgreSQL

#### Market Price Cache
- **Key**: `market:prices:{symbol}` (e.g., `market:prices:BTCUSDT`)
- **Type**: Float64
- **TTL**: 100ms
- **Source**: Market Data Service (should populate this)

#### Account State Cache (not fully implemented)
- **Key**: `account:state:{account_id}`
- **Type**: JSON object
- **TTL**: 100ms
- **Source**: Account Monitor Service (CURRENTLY USING MOCK)

### 4.4 Configuration (config.yaml + Environment Variables)

**Config File**: `/home/mm/dev/b25/services/risk-manager/config.yaml`

**Environment Prefix**: `RISK_`

Key configuration parameters:
```yaml
server:
  port: 8083                      # HTTP metrics/health port

database:
  host: localhost
  port: 5432
  database: b25_config

redis:
  host: localhost
  port: 6379
  db: 0                           # Policy cache DB

nats:
  url: nats://localhost:4222
  alert_subject: risk.alerts
  emergency_topic: risk.emergency

grpc:
  port: 50052                     # gRPC server port

risk:
  monitor_interval: 1s            # Risk check frequency
  cache_ttl: 100ms                # Market price cache TTL
  policy_cache_ttl: 1s            # Policy cache TTL
  max_leverage: 10.0              # Default max leverage
  max_drawdown_percent: 0.20      # Default max drawdown (20%)
  emergency_threshold: 0.25       # Emergency stop threshold (25%)
  alert_window: 5m                # Alert deduplication window
  account_monitor_url: localhost:50053  # Account Monitor gRPC (not used)
  market_data_redis_db: 1         # Market data Redis DB
```

---

## 5. Outputs

### 5.1 gRPC Outputs

#### OrderRiskResponse
```protobuf
message OrderRiskResponse {
  bool approved = 1;                       // true = order can proceed, false = blocked
  repeated string violations = 2;          // List of policy violations (empty if approved)
  RiskMetrics post_trade_metrics = 3;      // Projected metrics after order execution
  int64 processing_time_us = 4;            // Processing latency in microseconds
  string rejection_reason = 5;             // Reason code if rejected
}
```

**Rejection Reasons**:
- `emergency_stop_active` - Trading halted due to emergency stop
- `policy_violation` - Hard or emergency policy violated
- `simulation_failed` - Order simulation failed (e.g., insufficient margin)

#### RiskMetrics
```protobuf
message RiskMetrics {
  double margin_ratio = 1;                           // equity / margin_used
  double leverage = 2;                               // total_notional / equity
  double drawdown_daily = 3;                         // Daily loss percentage
  double drawdown_max = 4;                           // Max loss from peak
  double daily_pnl = 5;                              // Daily P&L
  double unrealized_pnl = 6;                         // Unrealized P&L
  double total_equity = 7;                           // Total account equity
  double total_margin_used = 8;                      // Total margin in use
  map<string, double> position_concentration = 9;    // symbol -> concentration %
  map<string, double> limit_utilization = 10;        // limit -> utilization %
  int32 open_positions = 11;                         // Number of open positions
  int32 pending_orders = 12;                         // Number of pending orders
}
```

### 5.2 NATS Outputs

#### Alert Topics

**Subject**: `risk.alerts.critical` (hard and emergency violations)
```json
{
  "level": "critical",
  "policy_id": "uuid",
  "policy_name": "Max Account Leverage",
  "policy_type": "hard",
  "metric": "leverage",
  "metric_value": 12.5,
  "threshold_value": 10.0,
  "message": "Policy 'Max Account Leverage' violated: leverage (12.5000) greater_than 10.0000",
  "timestamp": 1728000000
}
```

**Subject**: `risk.alerts.warning` (soft violations)
```json
{
  "level": "warning",
  "policy_id": "uuid",
  "policy_name": "Daily Drawdown Warning",
  "policy_type": "soft",
  "metric": "drawdown_daily",
  "metric_value": 0.12,
  "threshold_value": 0.10,
  "message": "Policy 'Daily Drawdown Warning' violated: drawdown_daily (0.1200) greater_than 0.1000",
  "timestamp": 1728000000
}
```

**Subject**: `risk.alerts.emergency` (emergency stop events)
```json
{
  "level": "emergency",
  "type": "emergency_stop",
  "reason": "Max drawdown exceeded 25%",
  "triggered_by": "risk_monitor",
  "timestamp": 1728000000
}
```

#### Metrics Topic

**Subject**: `risk.metrics` (published every 1 second)
```json
{
  "margin_ratio": 10.0,
  "leverage": 2.5,
  "drawdown_daily": 0.05,
  "drawdown_max": 0.08,
  "daily_pnl": -500.0,
  "unrealized_pnl": 150.0,
  "total_equity": 100000.0,
  "total_margin_used": 10000.0,
  "position_concentration": {
    "BTCUSDT": 0.30,
    "ETHUSDT": 0.20
  },
  "open_positions": 2,
  "pending_orders": 1,
  "timestamp": 1728000000
}
```

**Consumers**: Dashboard Server (for real-time risk monitoring)

### 5.3 HTTP Outputs

#### GET /health (Port 8083)
**Response**: `200 OK` with body `OK`
**Headers**: CORS enabled (`Access-Control-Allow-Origin: *`)

#### GET /metrics (Port 8083)
**Response**: Prometheus format metrics

**Key Metrics**:
```
# Pre-trade checks
risk_order_checks_total{result="approved"} 12345
risk_order_checks_total{result="rejected"} 67
risk_order_check_duration_microseconds_bucket{le="1000"} 11000
risk_orders_rejected_total{reason="policy_violation"} 45
risk_orders_rejected_total{reason="emergency_stop_active"} 22

# Risk metrics
risk_current_leverage 2.5
risk_current_margin_ratio 10.0
risk_current_drawdown 0.08
risk_current_equity 100000.0
risk_open_positions 2
risk_pending_orders 1

# Violations
risk_violations_total{policy_type="hard",policy_name="Max Leverage"} 5
risk_emergency_stops_total 2
risk_emergency_stop_active 0

# Alerts
risk_alerts_published_total{level="critical"} 10
risk_alerts_published_total{level="warning"} 25
risk_alerts_deduplicated_total 15

# Circuit breaker
risk_circuit_breaker_trips_total 1
risk_circuit_breaker_violation_count 3
```

### 5.4 Database Outputs (PostgreSQL)

#### Risk Violations Log
```sql
INSERT INTO risk_violations (
  policy_id,
  metric_value,
  threshold_value,
  context,
  action_taken
) VALUES (...)
```

**Purpose**: Audit trail of all violations for compliance and analysis

#### Emergency Stop Records
```sql
INSERT INTO emergency_stops (
  trigger_reason,
  triggered_by,
  account_state,
  positions_snapshot,
  orders_canceled,
  positions_closed
) VALUES (...)
```

**Purpose**: Historical record of emergency stops with full context

---

## 6. Dependencies

### 6.1 Required External Services

#### PostgreSQL (Database)
- **Host**: `localhost:5432`
- **Database**: `b25_config`
- **Schema**: See `/home/mm/dev/b25/services/risk-manager/migrations/001_initial_schema.up.sql`
- **Purpose**: Store risk policies, violation history, emergency stop records
- **Connection Pool**: 25 max open, 5 max idle, 5m lifetime
- **Critical**: Service cannot start without database connection

#### Redis (Cache)
- **Host**: `localhost:6379`
- **DB 0**: Policy cache
- **DB 1**: Market data prices (populated by Market Data Service)
- **Purpose**: Fast policy lookup, market price cache, account state cache
- **Connection Pool**: 10 connections, 3 max retries
- **Critical**: Service starts but performance degrades without Redis

#### NATS (Message Queue)
- **Host**: `nats://localhost:4222`
- **Purpose**: Publish alerts, metrics, emergency stop events
- **Max Reconnects**: 10
- **Reconnect Wait**: 2s
- **Critical**: Service starts but alerts won't be delivered without NATS

### 6.2 Service Dependencies (Not Fully Implemented)

#### Account Monitor Service (MISSING INTEGRATION)
- **Expected**: gRPC client to `localhost:50053`
- **Purpose**: Get real-time account state (equity, positions, margin)
- **Current Status**: Using mock data in `getMockAccountState()`
- **Impact**: **CRITICAL** - Risk calculations use fake data, not production-ready

#### Market Data Service (PARTIAL INTEGRATION)
- **Expected**: Writes to Redis DB 1 at `market:prices:{symbol}`
- **Purpose**: Current market prices for order simulation
- **Current Status**: Reads from Redis but falls back to order price
- **Impact**: **HIGH** - Market orders may use stale prices

### 6.3 Optional Dependencies

#### Prometheus (Monitoring)
- **Scrapes**: `http://localhost:8083/metrics`
- **Purpose**: Metrics collection and alerting

#### Grafana (Visualization)
- **Purpose**: Dashboard visualization of risk metrics

---

## 7. Configuration

### 7.1 Configuration Sources

Configuration is loaded from three sources (in order of precedence):
1. Environment variables (prefixed with `RISK_`)
2. Config file (`config.yaml` in current dir, `./config/`, or `/etc/risk-manager/`)
3. Default values (in `internal/config/config.go`)

### 7.2 Key Configuration Parameters

#### Server Configuration
```yaml
server:
  port: 8083                      # HTTP server port (metrics + health)
  mode: development               # "development" or "production"
  read_timeout: 10s               # HTTP read timeout
  write_timeout: 10s              # HTTP write timeout
  shutdown_timeout: 15s           # Graceful shutdown timeout
```

#### Database Configuration
```yaml
database:
  host: localhost
  port: 5432
  user: b25
  password: JDExqQGCJxncMuKrRwpAmg==   # SECURITY: Password in plaintext
  database: b25_config
  ssl_mode: disable
  max_open_conns: 25              # Max database connections
  max_idle_conns: 5               # Max idle connections
  conn_max_lifetime: 5m           # Connection max lifetime
```

#### Redis Configuration
```yaml
redis:
  host: localhost
  port: 6379
  password: ""                    # No password
  db: 0                           # Policy cache DB
  max_retries: 3
  pool_size: 10
```

#### NATS Configuration
```yaml
nats:
  url: nats://localhost:4222
  max_reconnect: 10
  reconnect_wait: 2s
  alert_subject: risk.alerts      # Alert topic prefix
  emergency_topic: risk.emergency # Emergency stop topic
```

#### gRPC Configuration
```yaml
grpc:
  port: 50052                     # gRPC server port
  max_connection_idle: 5m
  max_connection_age: 30m
  keep_alive_interval: 30s
  keep_alive_timeout: 10s
```

#### Risk Configuration
```yaml
risk:
  monitor_interval: 1s            # How often to check risk (background loop)
  cache_ttl: 100ms                # Market price cache TTL
  policy_cache_ttl: 1s            # Policy cache TTL
  max_leverage: 10.0              # Default max leverage
  max_drawdown_percent: 0.20      # Default max drawdown (20%)
  emergency_threshold: 0.25       # Emergency stop threshold (25%)
  alert_window: 5m                # Alert deduplication window
  account_monitor_url: localhost:50053  # Account Monitor gRPC endpoint
  market_data_redis_db: 1         # Redis DB for market data
```

**Purpose of Each**:
- `monitor_interval`: Frequency of background risk checks (balances overhead vs timeliness)
- `cache_ttl`: How long to cache market prices (affects price staleness)
- `policy_cache_ttl`: How long to cache policies (affects policy update propagation)
- `max_leverage`: Default maximum leverage allowed (can be overridden per policy)
- `max_drawdown_percent`: Hard limit for drawdown before blocking orders
- `emergency_threshold`: Drawdown level that triggers automatic emergency stop
- `alert_window`: Prevents duplicate alerts for same violation within window
- `account_monitor_url`: gRPC endpoint for Account Monitor (NOT CURRENTLY USED)
- `market_data_redis_db`: Redis DB where Market Data Service writes prices

#### Logging Configuration
```yaml
logging:
  level: info                     # debug, info, warn, error
  format: json                    # json or console
```

#### Metrics Configuration
```yaml
metrics:
  enabled: true
  port: 8083                      # Metrics port (shares with health endpoint)
```

### 7.3 Environment Variable Overrides

All config values can be overridden with environment variables:

```bash
# Example environment variables
export RISK_DATABASE_HOST=postgres.example.com
export RISK_REDIS_HOST=redis.example.com
export RISK_NATS_URL=nats://nats.example.com:4222
export RISK_LOGGING_LEVEL=debug
export RISK_GRPC_PORT=50052
export RISK_RISK_MAX_LEVERAGE=5.0
```

---

## 8. Code Structure

### 8.1 Directory Layout

```
/home/mm/dev/b25/services/risk-manager/
├── cmd/
│   └── server/
│       └── main.go                    # Application entry point (384 lines)
├── internal/
│   ├── cache/
│   │   └── policy_cache.go            # Redis caching layer (229 lines)
│   ├── config/
│   │   └── config.go                  # Configuration management (210 lines)
│   ├── emergency/
│   │   └── stop.go                    # Emergency stop manager (242 lines)
│   ├── grpc/
│   │   └── server.go                  # gRPC service implementation (289 lines)
│   ├── limits/
│   │   └── policy.go                  # Policy engine (228 lines)
│   ├── monitor/
│   │   ├── metrics.go                 # Prometheus metrics (218 lines)
│   │   ├── monitor.go                 # Risk monitoring loop (270 lines)
│   │   └── publisher.go               # NATS alert publisher (153 lines)
│   └── repository/
│       └── policy_repository.go       # Database operations (279 lines)
├── proto/
│   ├── risk_manager.proto             # gRPC service definition (122 lines)
│   ├── risk_manager.pb.go             # Generated protobuf code
│   └── risk_manager_grpc.pb.go        # Generated gRPC code
├── migrations/
│   ├── 001_initial_schema.up.sql      # Database schema (80 lines)
│   └── 001_initial_schema.down.sql    # Rollback migration
├── config.yaml                        # Default configuration
├── Dockerfile                         # Container image (HAS MERGE CONFLICT)
├── docker-compose.yml                 # Full stack deployment
├── Makefile                           # Build automation
├── go.mod                             # Go dependencies
├── .air.toml                          # Hot reload config
├── README.md                          # Service documentation
└── QUICKSTART.md                      # Quick start guide
```

### 8.2 Key Files and Responsibilities

#### cmd/server/main.go
**Lines**: 384
**Responsibilities**:
- Application initialization and bootstrapping
- Dependency injection and component wiring
- Loads configuration
- Initializes logger (Zap)
- Connects to PostgreSQL, Redis, NATS
- Loads policies from database
- Starts gRPC server (port 50052)
- Starts HTTP metrics server (port 8083)
- Starts risk monitor background loop
- Starts policy refresh loop (30s interval)
- Graceful shutdown handling

**Key Functions**:
- `main()`: Entry point, orchestrates service startup
- `initLogger()`: Creates Zap logger (JSON or console format)
- `initDatabase()`: Connects to PostgreSQL with connection pooling
- `initRedis()`: Creates Redis client
- `initNATS()`: Connects to NATS with reconnect handlers
- `startGRPCServer()`: Launches gRPC server with health checks
- `startMetricsServer()`: Launches HTTP server for metrics/health
- `startPolicyRefreshLoop()`: Background loop to refresh policies every 30s
- `getDefaultPolicies()`: Returns default policies if DB is empty

**Notable Code**:
- Sets CORS headers on health endpoint (lines 278-282)
- Uses mock account state (should integrate with Account Monitor)
- Policy refresh every 30 seconds to pick up database changes

#### internal/grpc/server.go
**Lines**: 289
**Responsibilities**:
- Implements gRPC RiskManager service
- Pre-trade order validation (CheckOrder)
- Batch order validation (CheckOrderBatch)
- Risk metrics retrieval (GetRiskMetrics)
- Emergency stop control (TriggerEmergencyStop, ReEnableTrading, GetEmergencyStopStatus)

**Key Methods**:
- `CheckOrder()`: Fast order risk validation (<10ms target)
  - Checks emergency stop status
  - Gets current market price from cache
  - Simulates order execution
  - Calculates post-trade metrics
  - Evaluates against policies
  - Returns approval/rejection

- `GetRiskMetrics()`: Returns current risk metrics for account
- `TriggerEmergencyStop()`: Triggers emergency stop (manual or automated)
- `ReEnableTrading()`: Re-enables trading after emergency stop (requires authorization)

**Notable Code**:
- Uses `getMockAccountState()` instead of real Account Monitor client (line 276)
- Processing time tracking for latency monitoring
- Filters only hard/emergency violations for blocking (soft = warnings only)

#### internal/risk/calculator.go
**Lines**: 288
**Responsibilities**:
- Core risk metric calculations
- Order simulation and impact analysis
- Position and margin calculations

**Key Structs**:
- `Calculator`: Main risk calculation engine
- `AccountState`: Current account snapshot
- `Position`: Individual position details
- `Order`: Pending order details
- `RiskMetrics`: Calculated risk metrics

**Key Methods**:
- `CalculateMetrics()`: Computes all risk metrics from account state
  - Margin ratio: equity / margin_used
  - Leverage: total_notional / equity
  - Daily drawdown: (daily_start - current) / daily_start
  - Max drawdown: (peak - current) / peak
  - Position concentration: symbol_notional / equity

- `SimulateOrder()`: Projects account state after order execution
  - Estimates margin requirement
  - Checks available margin
  - Updates positions (increase, reduce, or flip)

- `ValidateMetrics()`: Checks metrics against hardcoded limits
- `CalculateLimitUtilization()`: Shows how much of each limit is being used

**Risk Formulas**:
```go
// Margin ratio (higher is better)
margin_ratio = equity / margin_used

// Leverage (lower is better)
leverage = total_position_notional / equity

// Daily drawdown (0 to 1, lower is better)
drawdown_daily = (daily_start_equity - current_equity) / daily_start_equity

// Max drawdown (0 to 1, lower is better)
drawdown_max = (peak_equity - current_equity) / peak_equity

// Position concentration (0 to 1, lower is better)
concentration = symbol_notional / total_equity
```

#### internal/limits/policy.go
**Lines**: 228
**Responsibilities**:
- Policy definition and management
- Policy evaluation engine
- Violation detection

**Key Types**:
- `PolicyType`: hard, soft, emergency
- `PolicyScope`: account, symbol, strategy
- `PolicyOperator`: less_than, greater_than, equal, etc.

**Key Structs**:
- `Policy`: Risk policy definition
- `PolicyViolation`: Detected violation
- `PolicyEngine`: Policy evaluation engine

**Key Methods**:
- `LoadPolicies()`: Loads policies into engine
- `GetApplicablePolicies()`: Filters policies by scope (account/symbol/strategy)
- `EvaluatePolicy()`: Checks single policy against metrics
- `EvaluateAll()`: Checks all applicable policies
- `HasEmergencyViolations()`: Checks if any emergency violations exist

**Policy Evaluation Logic**:
```go
// Example: Max Leverage Policy
{
  Metric: "leverage",
  Operator: "less_than_or_equal",
  Threshold: 10.0,
  Type: "hard"
}

// Violated if: leverage > 10.0
// Result: Order blocked
```

#### internal/emergency/stop.go
**Lines**: 242
**Responsibilities**:
- Emergency stop state management
- Circuit breaker pattern implementation
- Emergency stop trigger and recovery

**Key Structs**:
- `StopManager`: Manages emergency stop state
- `StopStatus`: Current stop status details
- `CircuitBreaker`: Automatic emergency stop on repeated violations

**Key Methods**:
- `Trigger()`: Triggers emergency stop
- `IsActive()`: Checks if stop is active
- `ReEnable()`: Re-enables trading after stop
- `ShouldBlockOrders()`: Determines if orders should be blocked

**Circuit Breaker**:
- Threshold: 5 violations
- Window: 1 minute
- Action: Triggers emergency stop automatically

**Emergency Stop Flow**:
1. Trigger called (manual or automatic)
2. Set `is_stopped = true`
3. Publish emergency alert to NATS
4. Block all future orders (via `ShouldBlockOrders()`)
5. Wait for manual re-enable (requires authorization)

#### internal/monitor/monitor.go
**Lines**: 270
**Responsibilities**:
- Background risk monitoring (1s interval)
- Continuous policy evaluation
- Violation handling and escalation

**Key Struct**:
- `RiskMonitor`: Background monitoring service

**Key Methods**:
- `Run()`: Main monitoring loop (1s ticker)
- `checkRisk()`: Single risk check iteration
- `handleViolations()`: Processes detected violations by type

**Monitoring Loop**:
```
Every 1 second:
1. Get account state (MOCK)
2. Calculate risk metrics
3. Publish metrics to NATS (for dashboard)
4. Evaluate all policies
5. Handle violations:
   - Emergency: Trigger emergency stop
   - Hard: Log, alert, circuit breaker
   - Soft: Log, alert
6. Record violations to database
```

**Circuit Breaker Integration**:
- 5 hard violations in 1 minute → automatic emergency stop

#### internal/monitor/publisher.go
**Lines**: 153
**Responsibilities**:
- NATS alert publishing
- Alert deduplication
- Metrics broadcasting

**Key Struct**:
- `NATSAlertPublisher`: NATS publisher
- `AlertDeduplicator`: Prevents duplicate alerts

**Key Methods**:
- `PublishAlert()`: Publishes policy violation alert
- `PublishMetrics()`: Publishes risk metrics
- `PublishEmergencyAlert()`: Publishes emergency stop alert

**Alert Deduplication**:
- Window: 5 minutes (configurable)
- Key: `{policy_id}:{metric}`
- Prevents spam from same violation

#### internal/cache/policy_cache.go
**Lines**: 229
**Responsibilities**:
- Two-tier caching (Redis + in-memory)
- Policy cache management
- Market price cache
- Account state cache

**Key Structs**:
- `PolicyCache`: Policy caching with local fallback
- `MarketPriceCache`: Market price caching
- `AccountStateCache`: Account state caching (not fully used)

**Cache Strategy**:
1. Check local in-memory cache (fastest)
2. If miss/expired, check Redis
3. If Redis hit, update local cache
4. If Redis miss, return error (caller loads from DB)

**TTLs**:
- Policy cache: 1 second
- Price cache: 100ms
- Account state: 100ms

#### internal/repository/policy_repository.go
**Lines**: 279
**Responsibilities**:
- PostgreSQL database operations
- Policy CRUD operations
- Violation recording
- Emergency stop recording

**Key Methods**:
- `GetActive()`: Load all active policies
- `Create()`: Create new policy
- `Update()`: Update policy (with optimistic locking)
- `RecordViolation()`: Log policy violation
- `RecordEmergencyStop()`: Log emergency stop event

**Optimistic Locking**:
- Uses version field for concurrent updates
- Prevents lost updates

---

## 9. Testing in Isolation

### 9.1 Prerequisites

**Required Software**:
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL client (psql)
- Redis CLI
- grpcurl (for gRPC testing)
- NATS CLI (optional, for monitoring alerts)

**Installation**:
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 9.2 Quick Start - Docker Compose (Recommended)

**Full stack with all dependencies**:

```bash
# Navigate to service directory
cd /home/mm/dev/b25/services/risk-manager

# Start all services (PostgreSQL, Redis, NATS, Risk Manager, Prometheus, Grafana)
docker-compose up -d

# View logs
docker-compose logs -f risk-manager

# Expected output:
# {"level":"info","msg":"starting risk manager service","version":"1.0.0","mode":"development"}
# {"level":"info","msg":"database connection established"}
# {"level":"info","msg":"Redis connection established"}
# {"level":"info","msg":"NATS connection established"}
# {"level":"info","msg":"policies loaded from database","count":5}
# {"level":"info","msg":"gRPC server started","addr":":50052"}
# {"level":"info","msg":"metrics server started","addr":":8083"}
# {"level":"info","msg":"risk monitor started","interval":"1s"}

# Check health
curl http://localhost:8083/health
# Expected: OK

# Stop services
docker-compose down
```

**Services exposed**:
- Risk Manager gRPC: `localhost:50052`
- Risk Manager HTTP: `http://localhost:8083`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- NATS: `localhost:4222`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

### 9.3 Local Development Setup

**Step 1: Start Infrastructure Only**

```bash
cd /home/mm/dev/b25/services/risk-manager

# Start only PostgreSQL, Redis, NATS
docker-compose up -d postgres redis nats

# Wait for services to be ready (10 seconds)
sleep 10
```

**Step 2: Run Database Migrations**

```bash
# Run migrations to create schema and default policies
make migrate-up

# Expected output:
# Running migrations...
# migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/risk_manager?sslmode=disable" up
# 1/u initial_schema (timestamp ms)

# Verify tables were created
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager -c "\dt"

# Expected tables:
# risk_policies
# risk_violations
# emergency_stops
```

**Step 3: Verify Database Setup**

```bash
# Connect to database
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager

# Check default policies
SELECT id, name, type, metric, threshold, enabled FROM risk_policies;

# Expected 5 policies:
# 1. Max Account Leverage (hard, leverage ≤ 10.0)
# 2. Min Margin Ratio (hard, margin_ratio ≥ 1.0)
# 3. Daily Drawdown Warning (soft, drawdown_daily > 0.10)
# 4. Max Drawdown Hard Limit (hard, drawdown_max > 0.20)
# 5. Emergency Drawdown Stop (emergency, drawdown_max > 0.25)

# Exit psql
\q
```

**Step 4: Build Service**

```bash
# Build binary
make build

# Expected output:
# Building Risk Manager Service...
# go build -o bin/risk-manager cmd/server/main.go

# Verify binary was created
ls -lh bin/risk-manager
```

**Step 5: Run Service Locally**

```bash
# Run service (uses config.yaml in current directory)
make run

# Or run directly
./bin/risk-manager

# Or run with hot reload (requires air)
make dev

# Expected startup logs:
# {"level":"info","ts":...,"msg":"starting risk manager service","version":"1.0.0"}
# {"level":"info","ts":...,"msg":"database connection established"}
# {"level":"info","ts":...,"msg":"Redis connection established"}
# {"level":"info","ts":...,"msg":"NATS connection established"}
# {"level":"info","ts":...,"msg":"policies loaded from database","count":5}
# {"level":"info","ts":...,"msg":"gRPC server started","addr":":50052"}
# {"level":"info","ts":...,"msg":"metrics server started","addr":":8083"}
# {"level":"info","ts":...,"msg":"risk monitor started","interval":"1s"}
```

### 9.4 Health Check Tests

**Test 1: HTTP Health Endpoint**

```bash
# Test health endpoint
curl -v http://localhost:8083/health

# Expected response:
# HTTP/1.1 200 OK
# Access-Control-Allow-Origin: *
# Content-Type: text/plain
#
# OK

# Test CORS preflight
curl -X OPTIONS http://localhost:8083/health

# Expected: 200 OK
```

**Test 2: Prometheus Metrics**

```bash
# Get all metrics
curl http://localhost:8083/metrics | head -50

# Search for risk metrics
curl http://localhost:8083/metrics | grep "^risk_"

# Expected metrics:
# risk_current_leverage 0
# risk_current_margin_ratio +Inf
# risk_current_drawdown 0
# risk_current_equity 100000
# risk_open_positions 0
# risk_pending_orders 0
# risk_emergency_stop_active 0
```

**Test 3: gRPC Health Check**

```bash
# Install grpcurl if not already installed
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:50052 list

# Expected output:
# grpc.health.v1.Health
# riskmanager.RiskManager

# Check health
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check

# Expected:
# {
#   "status": "SERVING"
# }
```

### 9.5 Functional Tests with Mock Data

**Test 1: Check Order - Approved (Low Risk)**

```bash
# Submit low-risk order (should be approved)
grpcurl -plaintext -d '{
  "order_id": "test-order-001",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": 0.01,
  "price": 50000,
  "order_type": "LIMIT",
  "account_id": "test-account",
  "strategy_id": "test-strategy"
}' localhost:50052 riskmanager.RiskManager/CheckOrder

# Expected response:
# {
#   "approved": true,
#   "postTradeMetrics": {
#     "marginRatio": 9.524,
#     "leverage": 0.105,
#     "drawdownDaily": 0.020,
#     "totalEquity": 100000,
#     "totalMarginUsed": 10500,
#     ...
#   },
#   "processingTimeUs": "1500"  # Should be < 10000 microseconds
# }
```

**Test 2: Check Order - Rejected (Insufficient Margin)**

```bash
# Submit order requiring more margin than available
grpcurl -plaintext -d '{
  "order_id": "test-order-002",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": 100,
  "price": 50000,
  "order_type": "LIMIT",
  "account_id": "test-account",
  "strategy_id": "test-strategy"
}' localhost:50052 riskmanager.RiskManager/CheckOrder

# Expected response:
# {
#   "approved": false,
#   "violations": [
#     "insufficient margin: need 500000.00, available 90000.00"
#   ],
#   "rejectionReason": "simulation_failed",
#   "processingTimeUs": "800"
# }
```

**Test 3: Get Risk Metrics**

```bash
# Get current risk metrics for account
grpcurl -plaintext -d '{
  "account_id": "test-account"
}' localhost:50052 riskmanager.RiskManager/GetRiskMetrics

# Expected response:
# {
#   "metrics": {
#     "marginRatio": 10,
#     "leverage": 0,
#     "drawdownDaily": 0.020408163,
#     "drawdownMax": 0.047619048,
#     "dailyPnl": 2000,
#     "totalEquity": 100000,
#     "totalMarginUsed": 10000,
#     "openPositions": 0,
#     "pendingOrders": 0
#   },
#   "timestamp": "1728000000"
# }
```

**Test 4: Trigger Emergency Stop**

```bash
# Trigger emergency stop manually
grpcurl -plaintext -d '{
  "reason": "Testing emergency stop functionality",
  "triggered_by": "test_user",
  "force": false
}' localhost:50052 riskmanager.RiskManager/TriggerEmergencyStop

# Expected response:
# {
#   "success": true,
#   "message": "Emergency stop activated",
#   "status": {
#     "isStopped": true,
#     "stoppedAt": "1728000000",
#     "stopReason": "Testing emergency stop functionality",
#     "triggeredBy": "test_user",
#     "ordersCanceled": 0,
#     "positionsClosed": 0,
#     "completed": false
#   }
# }

# Check emergency stop status
grpcurl -plaintext -d '{}' localhost:50052 riskmanager.RiskManager/GetEmergencyStopStatus

# Try to check order (should be blocked)
grpcurl -plaintext -d '{
  "order_id": "test-order-003",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": 0.01,
  "price": 50000,
  "order_type": "LIMIT",
  "account_id": "test-account"
}' localhost:50052 riskmanager.RiskManager/CheckOrder

# Expected:
# {
#   "approved": false,
#   "violations": ["Emergency stop is active - all trading halted"],
#   "rejectionReason": "emergency_stop_active"
# }
```

**Test 5: Re-enable Trading**

```bash
# Re-enable trading after emergency stop
grpcurl -plaintext -d '{
  "authorized_by": "test_admin",
  "reason": "Emergency resolved, testing completed"
}' localhost:50052 riskmanager.RiskManager/ReEnableTrading

# Expected response:
# {
#   "success": true,
#   "message": "Trading re-enabled successfully",
#   "reEnabledAt": "1728000100"
# }

# Verify orders are now allowed
grpcurl -plaintext -d '{
  "order_id": "test-order-004",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": 0.01,
  "price": 50000,
  "order_type": "LIMIT",
  "account_id": "test-account"
}' localhost:50052 riskmanager.RiskManager/CheckOrder

# Expected: "approved": true
```

### 9.6 Database Testing

**Test 1: View Policies**

```bash
# Connect to database
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager

# View all policies
SELECT id, name, type, metric, operator, threshold, enabled, priority
FROM risk_policies
ORDER BY priority DESC;

# Expected: 5 default policies
```

**Test 2: Add Custom Policy**

```bash
# Still in psql, insert a custom policy
INSERT INTO risk_policies (
  name, type, metric, operator, threshold, scope, enabled, priority
) VALUES (
  'BTCUSDT Position Limit',
  'hard',
  'concentration_BTCUSDT',
  'less_than_or_equal',
  0.50,
  'symbol',
  true,
  120
);

# Wait 30 seconds for policy refresh loop to pick up the change
# (or restart service)

# Verify policy was loaded (check logs)
# Expected log: "policies refreshed" with count=6
```

**Test 3: View Violations History**

```bash
# View recent violations
SELECT
  v.id,
  p.name as policy_name,
  v.violation_time,
  v.metric_value,
  v.threshold_value,
  v.action_taken
FROM risk_violations v
LEFT JOIN risk_policies p ON v.policy_id = p.id
ORDER BY v.violation_time DESC
LIMIT 10;

# Exit psql
\q
```

### 9.7 NATS Alert Testing

**Install NATS CLI** (optional):
```bash
# Install NATS CLI
curl -sf https://binaries.nats.dev/nats-io/natscli/nats@latest | sh

# Or with Go
go install github.com/nats-io/natscli/nats@latest
```

**Test 1: Subscribe to All Risk Alerts**

```bash
# Subscribe to all risk alert topics
nats sub "risk.alerts.>"

# In another terminal, trigger an action that causes a violation
# (e.g., emergency stop or modify account state to violate policy)

# Expected output (when violation occurs):
# [#1] Received on "risk.alerts.emergency"
# {"level":"emergency","type":"emergency_stop","reason":"...","triggered_by":"...","timestamp":1728000000}
```

**Test 2: Monitor Risk Metrics Stream**

```bash
# Subscribe to risk metrics (published every 1 second)
nats sub "risk.metrics"

# Expected output (every 1 second):
# [#1] Received on "risk.metrics"
# {"margin_ratio":10,"leverage":0,"drawdown_daily":0.02,"...","timestamp":1728000000}
# [#2] Received on "risk.metrics"
# {"margin_ratio":10,"leverage":0,"drawdown_daily":0.02,"...","timestamp":1728000001}
```

### 9.8 Redis Cache Testing

**Test 1: Verify Policy Cache**

```bash
# Connect to Redis
docker exec -it risk-manager-redis redis-cli

# Check policy cache (should be JSON array)
GET risk:policies

# Expected: JSON array of 5 (or 6 if custom policy added) policies

# Check TTL
TTL risk:policies

# Expected: ~1 second (refreshes continuously)
```

**Test 2: Simulate Market Price Data**

```bash
# Still in redis-cli
# Switch to market data DB
SELECT 1

# Set mock market prices (simulates Market Data Service)
SET market:prices:BTCUSDT 50000.50
SET market:prices:ETHUSDT 3000.25
SET market:prices:BNBUSDT 400.10

# Set TTL to 100ms (as configured)
EXPIRE market:prices:BTCUSDT 1
EXPIRE market:prices:ETHUSDT 1
EXPIRE market:prices:BNBUSDT 1

# Exit redis-cli
EXIT
```

**Test 3: Test Order with Market Price**

```bash
# Submit market order (price = 0, should use cached price)
grpcurl -plaintext -d '{
  "order_id": "test-market-order",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": 0.1,
  "price": 0,
  "order_type": "MARKET",
  "account_id": "test-account"
}' localhost:50052 riskmanager.RiskManager/CheckOrder

# Expected: Uses price from Redis (50000.50)
# If Redis key expired or not found, request will fail with error
```

### 9.9 Performance Testing

**Test 1: Measure Pre-trade Check Latency**

```bash
# Create script to measure latency
cat > test_latency.sh << 'EOF'
#!/bin/bash
for i in {1..100}; do
  result=$(grpcurl -plaintext -d '{
    "order_id": "perf-test-'$i'",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "quantity": 0.01,
    "price": 50000,
    "order_type": "LIMIT",
    "account_id": "test-account"
  }' localhost:50052 riskmanager.RiskManager/CheckOrder 2>&1)

  # Extract processing time
  latency=$(echo "$result" | grep -o '"processingTimeUs": "[0-9]*"' | grep -o '[0-9]*')
  echo "Request $i: ${latency}µs"
done
EOF

chmod +x test_latency.sh
./test_latency.sh

# Expected: Most requests < 10000µs (10ms), typical ~1000-3000µs
```

**Test 2: Check Prometheus Latency Metrics**

```bash
# View latency histogram
curl -s http://localhost:8083/metrics | grep risk_order_check_duration

# Expected output:
# risk_order_check_duration_microseconds_bucket{le="100"} 0
# risk_order_check_duration_microseconds_bucket{le="500"} 15
# risk_order_check_duration_microseconds_bucket{le="1000"} 45
# risk_order_check_duration_microseconds_bucket{le="2500"} 85
# risk_order_check_duration_microseconds_bucket{le="5000"} 98
# risk_order_check_duration_microseconds_bucket{le="10000"} 100
```

### 9.10 Expected Test Results Summary

| Test | Expected Result | Success Criteria |
|------|----------------|------------------|
| Health Check | `200 OK` | HTTP status 200 |
| gRPC Service List | Shows `riskmanager.RiskManager` | Service visible |
| Low-risk Order | `approved: true` | Processing time < 10ms |
| High-risk Order | `approved: false` with violations | Correct rejection reason |
| Emergency Stop | Orders blocked after trigger | All subsequent orders rejected |
| Re-enable Trading | Orders allowed after re-enable | Emergency stop cleared |
| Policy Cache | Redis contains 5-6 policies | TTL ~1 second |
| Market Price Cache | Redis contains price data in DB 1 | Price retrieved successfully |
| Metrics Publishing | NATS receives metrics every 1s | Continuous stream |
| Database Violations | Violations logged to PostgreSQL | Rows created in risk_violations |
| Latency p99 | < 10ms (10000µs) | 99% of requests under target |

---

## 10. Health Checks

### 10.1 Startup Health Checks

**Service is healthy when**:
1. HTTP server responds on port 8083
2. gRPC server responds on port 50052
3. Database connection established
4. Redis connection established
5. NATS connection established
6. Policies loaded from database (at least default 5 policies)

**Startup Verification Commands**:

```bash
# 1. Check HTTP health endpoint
curl -f http://localhost:8083/health || echo "FAIL: HTTP health check"

# 2. Check gRPC health
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check || echo "FAIL: gRPC health check"

# 3. Check PostgreSQL connection
docker exec -it risk-manager-postgres pg_isready -U postgres || echo "FAIL: PostgreSQL connection"

# 4. Check Redis connection
docker exec -it risk-manager-redis redis-cli ping || echo "FAIL: Redis connection"

# 5. Check NATS connection
curl http://localhost:8222/healthz || echo "FAIL: NATS connection"

# 6. Check policy count
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager -tAc "SELECT COUNT(*) FROM risk_policies WHERE enabled=true;" || echo "FAIL: Policy check"

# Expected: All checks pass
```

### 10.2 Runtime Health Monitoring

**Key Metrics to Monitor**:

#### Availability Metrics
```bash
# HTTP health endpoint uptime
curl http://localhost:8083/health
# Target: 99.9% uptime

# gRPC health status
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check
# Target: SERVING status
```

#### Performance Metrics
```bash
# Pre-trade check latency
curl -s http://localhost:8083/metrics | grep risk_order_check_duration_microseconds

# Target: p50 < 2ms, p99 < 10ms
# Check p99 calculation:
# histogram_quantile(0.99, rate(risk_order_check_duration_microseconds_bucket[5m]))
```

#### Functional Metrics
```bash
# Order check success rate
curl -s http://localhost:8083/metrics | grep risk_order_checks_total

# Calculate success rate:
# approved / (approved + rejected)
# Target: > 90% (depends on trading strategy aggressiveness)

# Emergency stop status
curl -s http://localhost:8083/metrics | grep risk_emergency_stop_active

# Expected: 0 (not active) in normal operation
```

#### Resource Metrics
```bash
# Database connections
docker exec -it risk-manager-postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity WHERE datname='risk_manager';"

# Target: < 25 (max_open_conns)

# Redis connection count
docker exec -it risk-manager-redis redis-cli CLIENT LIST | wc -l

# Target: < 10 (pool_size)

# NATS connection status
curl http://localhost:8222/connz

# Target: 1 active connection from risk-manager
```

### 10.3 Dependency Health Checks

**PostgreSQL Health**:
```bash
# Check database is accepting connections
docker exec -it risk-manager-postgres pg_isready -U postgres -d risk_manager

# Check for long-running queries (> 10 seconds)
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager -c "
  SELECT pid, now() - query_start as duration, query
  FROM pg_stat_activity
  WHERE state = 'active' AND now() - query_start > interval '10 seconds';
"

# Expected: No long-running queries

# Check table sizes
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager -c "
  SELECT schemaname, tablename,
         pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
  FROM pg_tables
  WHERE schemaname = 'public'
  ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"

# Monitor growth of risk_violations table (grows over time)
```

**Redis Health**:
```bash
# Check Redis is responding
docker exec -it risk-manager-redis redis-cli ping
# Expected: PONG

# Check memory usage
docker exec -it risk-manager-redis redis-cli INFO memory | grep used_memory_human

# Check key count
docker exec -it risk-manager-redis redis-cli DBSIZE

# Check policy cache exists
docker exec -it risk-manager-redis redis-cli EXISTS risk:policies
# Expected: 1 (exists)
```

**NATS Health**:
```bash
# Check NATS monitoring endpoint
curl http://localhost:8222/varz | jq '.connections'

# Check subscriptions
curl http://localhost:8222/subsz

# Expected: Subscriptions from dashboard-server
```

### 10.4 Alert Configuration

**Critical Alerts** (require immediate attention):

1. **Emergency Stop Active**
   - Metric: `risk_emergency_stop_active == 1`
   - Action: Investigate trigger reason, review account state

2. **Service Down**
   - Metric: `up{job="risk-manager"} == 0`
   - Action: Check logs, restart service

3. **Database Connection Lost**
   - Log: "failed to ping database"
   - Action: Check PostgreSQL availability

4. **High Pre-trade Check Latency**
   - Metric: `histogram_quantile(0.99, risk_order_check_duration_microseconds) > 10000`
   - Action: Check database/Redis performance, reduce load

5. **Order Rejection Rate > 50%**
   - Metric: `rate(risk_orders_rejected_total[5m]) / rate(risk_order_checks_total[5m]) > 0.5`
   - Action: Review policies, check for account issues

**Warning Alerts** (monitor closely):

1. **Policy Refresh Failures**
   - Log: "failed to refresh policies"
   - Action: Check database connectivity

2. **NATS Alert Publishing Failures**
   - Log: "failed to publish alert"
   - Action: Check NATS connectivity

3. **Circuit Breaker Approaching Threshold**
   - Metric: `risk_circuit_breaker_violation_count >= 3`
   - Action: Review recent violations

### 10.5 Health Check Automation

**Docker Health Check** (built into Dockerfile):
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1
```

**Kubernetes Readiness Probe** (example):
```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8083
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3

livenessProbe:
  httpGet:
    path: /health
    port: 8083
  initialDelaySeconds: 30
  periodSeconds: 30
  timeoutSeconds: 3
  failureThreshold: 3
```

**Monitoring Script** (simple bash):
```bash
#!/bin/bash
# health_monitor.sh

while true; do
  # Check HTTP health
  if ! curl -sf http://localhost:8083/health > /dev/null; then
    echo "CRITICAL: Risk Manager HTTP health check failed"
    # Send alert
  fi

  # Check gRPC health
  if ! grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check > /dev/null 2>&1; then
    echo "CRITICAL: Risk Manager gRPC health check failed"
    # Send alert
  fi

  # Check emergency stop status
  emergency=$(curl -s http://localhost:8083/metrics | grep "risk_emergency_stop_active" | awk '{print $2}')
  if [ "$emergency" == "1" ]; then
    echo "WARNING: Emergency stop is active"
    # Send alert
  fi

  sleep 60
done
```

---

## 11. Performance Characteristics

### 11.1 Latency Targets

#### Pre-trade Check (CheckOrder)

**Target**: < 10ms p99 (documented in README.md and code comments)

**Current Performance** (from README):
- p50: ~1.2ms
- p99: ~3.5ms
- **Status**: ✅ Exceeding target

**Latency Breakdown** (estimated):
1. gRPC overhead: ~0.1ms
2. Emergency stop check: ~0.05ms (in-memory)
3. Price cache lookup (Redis): ~0.5ms
4. Account state retrieval: ~0ms (MOCK, should be ~1-2ms with real Account Monitor)
5. Order simulation: ~0.2ms (in-memory calculation)
6. Metrics calculation: ~0.1ms
7. Policy evaluation: ~0.3ms (in-memory, cached policies)
8. Response serialization: ~0.1ms

**Total Estimated**: ~1.3ms (aligns with current p50)

**Optimization Opportunities**:
- Policy cache hit rate > 99.5% (from README)
- Local in-memory policy cache reduces Redis lookups
- Minimal database queries in hot path

#### Background Risk Monitoring

**Interval**: 1 second (configurable via `risk.monitor_interval`)

**Processing Time**: < 100ms per iteration (to avoid blocking next tick)

**Components**:
1. Account state fetch: ~10ms (MOCK, will be higher with real integration)
2. Metrics calculation: ~5ms
3. NATS publish: ~10ms
4. Policy evaluation: ~5ms
5. Violation handling: ~20ms (database writes, NATS alerts)

**Total**: ~50ms per iteration (leaves 950ms buffer)

#### Emergency Stop

**Trigger Latency**: < 500ms (target from README)
**Current**: ~200ms (from README)

**Steps**:
1. State update: ~1ms (in-memory)
2. NATS alert publish: ~10ms
3. Database record: ~50ms (async, doesn't block)

**Blocking Time**: ~11ms (state + alert)

### 11.2 Throughput Targets

#### Pre-trade Checks

**Target**: 1000+ requests/second per instance

**Factors**:
- gRPC connection pooling
- Cached policies (no DB query per request)
- Cached prices (Redis lookup, not calculation)
- Mock account state (instant, real integration will reduce throughput)

**Estimated with Real Integrations**:
- Account Monitor gRPC call: +2ms → ~500 req/s
- Can scale horizontally with multiple instances

#### Policy Updates

**Refresh Rate**: Every 30 seconds (from `startPolicyRefreshLoop` in main.go)

**Database Query**: Single query to load all active policies

**Impact**: Minimal (background goroutine, doesn't block request handling)

#### NATS Publishing

**Metrics**: 1 message/second (from monitoring loop)
**Alerts**: Variable (depends on violations)
**Throughput**: NATS can handle 10,000+ msg/s easily

### 11.3 Resource Requirements

#### Memory

**Base Usage**: ~50-100 MB (Go runtime + caches)

**Caches**:
- Policy cache: ~10 KB (5-10 policies × ~1 KB each)
- Local policy cache: ~10 KB (in-memory copy)
- Alert deduplicator: ~1 KB per unique violation in 5-minute window

**Growth Factors**:
- Alert deduplicator grows with unique violations
- Cleanup runs every 5 minutes to remove old entries

**Estimated Total**: 100-150 MB under normal load

#### CPU

**Idle**: < 1% (monitoring loop + policy refresh)

**Under Load**:
- 100 req/s: ~10-15% CPU
- 1000 req/s: ~80-90% CPU (single core)

**Bottlenecks**:
- Risk calculation (floating-point math)
- Policy evaluation (loop over policies)
- JSON serialization/deserialization

**Scaling**: Can scale horizontally (stateless except for emergency stop)

#### Network

**Inbound** (gRPC):
- CheckOrder request: ~200 bytes
- CheckOrder response: ~500 bytes
- At 1000 req/s: ~0.7 MB/s

**Outbound** (NATS):
- Metrics: ~500 bytes/second
- Alerts: ~500 bytes per alert

**Database**:
- Policy refresh: ~10 KB every 30s
- Violation logging: ~1 KB per violation

**Total Bandwidth**: < 1 MB/s under normal load

#### Database Connections

**Max Open**: 25 (configured)
**Max Idle**: 5 (configured)
**Lifetime**: 5 minutes

**Typical Usage**:
- Policy refresh: 1 connection every 30s
- Violation logging: 1 connection per violation
- Emergency stop: 1 connection per event

**Peak**: ~5-10 connections under moderate load

### 11.4 Scalability Limits

#### Single Instance Limits

**Theoretical Max** (with current mock data):
- Pre-trade checks: ~2000 req/s (500µs per request)
- Limited by CPU (risk calculation)

**With Real Integrations**:
- Pre-trade checks: ~500 req/s (2ms Account Monitor call + 1ms processing)
- Limited by Account Monitor response time

#### Horizontal Scaling

**Stateless Design**: ✅ Can run multiple instances

**Shared State**:
- Emergency stop: Stored in StopManager (in-memory per instance)
- **Problem**: Emergency stop not synchronized across instances
- **Solution Needed**: Move emergency stop state to Redis or database

**Load Balancing**:
- gRPC: Use client-side or proxy load balancing
- No sticky sessions required

**Coordination**:
- Policy refresh: All instances refresh independently (acceptable)
- Monitoring loop: All instances run independently (can cause duplicate alerts)
  - **Mitigation**: Alert deduplication helps but not perfect
  - **Better Solution**: Use single monitor instance or leader election

#### Vertical Scaling

**CPU**: Linear scaling up to ~8 cores (diminishing returns after)
**Memory**: Service is lightweight, memory not a bottleneck
**Disk**: Minimal disk I/O (database is remote)

### 11.5 Performance Monitoring

**Key Metrics**:

```promql
# Latency percentiles
histogram_quantile(0.50, rate(risk_order_check_duration_microseconds_bucket[5m]))
histogram_quantile(0.99, rate(risk_order_check_duration_microseconds_bucket[5m]))

# Request rate
rate(risk_order_checks_total[1m])

# Error rate
rate(risk_orders_rejected_total{reason!~"policy_violation"}[5m])

# Cache hit rate
rate(risk_policy_cache_hits_total[5m]) / rate(risk_policy_cache_requests_total[5m])
# Note: These metrics not currently exposed, need to add

# Emergency stop frequency
increase(risk_emergency_stops_total[1h])
```

**Grafana Dashboard Queries**:

```promql
# Pre-trade Check Latency (line graph)
histogram_quantile(0.50, rate(risk_order_check_duration_microseconds_bucket[5m])) / 1000 # Convert to ms
histogram_quantile(0.99, rate(risk_order_check_duration_microseconds_bucket[5m])) / 1000

# Order Approval Rate (gauge)
rate(risk_order_checks_total{result="approved"}[5m]) / rate(risk_order_checks_total[5m]) * 100

# Current Risk Metrics (gauges)
risk_current_leverage
risk_current_margin_ratio
risk_current_drawdown
risk_current_equity

# Violation Rate (bar chart)
sum by (policy_name) (rate(risk_violations_total[1h]))
```

---

## 12. Current Issues

### 12.1 Critical Issues

#### 1. Mock Account State Data (SEVERITY: CRITICAL)

**Location**:
- `/home/mm/dev/b25/services/risk-manager/internal/grpc/server.go:276`
- `/home/mm/dev/b25/services/risk-manager/internal/monitor/monitor.go:222`

**Issue**: Service uses hardcoded mock account state instead of real data from Account Monitor Service.

**Code**:
```go
// getMockAccountState returns mock account state (replace with real Account Monitor client)
func (s *RiskServer) getMockAccountState(accountID string) risk.AccountState {
    return risk.AccountState{
        Equity:           100000.0,
        Balance:          100000.0,
        UnrealizedPnL:    0.0,
        MarginUsed:       10000.0,
        AvailableMargin:  90000.0,
        Positions:        []risk.Position{},
        PendingOrders:    []risk.Order{},
        PeakEquity:       105000.0,
        DailyStartEquity: 98000.0,
    }
}
```

**Impact**:
- All risk calculations use fake data
- Pre-trade checks approve/reject based on incorrect account state
- Monitoring loop calculates metrics from fake positions
- **Service is NOT production-ready without this integration**

**Fix Required**:
1. Implement gRPC client to Account Monitor (URL configured as `localhost:50053`)
2. Replace `getMockAccountState()` calls with real gRPC calls
3. Handle Account Monitor unavailability (fallback? circuit breaker?)
4. Add caching layer to reduce Account Monitor load

**Estimated Effort**: 4-8 hours

---

#### 2. Dockerfile Contains Git Merge Conflict (SEVERITY: HIGH)

**Location**: `/home/mm/dev/b25/services/risk-manager/Dockerfile`

**Issue**: Dockerfile has unresolved Git merge conflict markers.

**Lines 1-96**:
```dockerfile
<<<<<<< HEAD
# Build stage
=======
# Multi-stage build for Go Risk Manager Service

# Builder stage - compiles the application
>>>>>>> refs/remotes/origin/main
FROM golang:1.22-alpine AS builder
...
```

**Impact**:
- Docker build will fail
- Cannot build production images
- Service cannot be deployed to containerized environments

**Fix Required**:
1. Resolve merge conflict manually
2. Choose correct version (likely origin/main has better comments)
3. Test Docker build

**Command to Fix**:
```bash
cd /home/mm/dev/b25/services/risk-manager
git checkout Dockerfile  # Reset to clean version
# Or manually edit to remove conflict markers
docker build -t risk-manager:test .  # Verify build works
```

**Estimated Effort**: 10 minutes

---

#### 3. No Test Coverage (SEVERITY: HIGH)

**Issue**: No test files found in entire codebase.

**Search Results**:
```bash
find /home/mm/dev/b25/services/risk-manager -name "*_test.go"
# Result: (empty)
```

**Impact**:
- No confidence in code correctness
- Refactoring is risky
- Cannot verify risk calculation formulas
- No regression testing

**Areas Needing Tests**:
1. **Risk Calculator** (`internal/risk/calculator.go`):
   - Metric calculations (leverage, margin ratio, drawdown)
   - Order simulation logic
   - Position updates (increase, reduce, flip)

2. **Policy Engine** (`internal/limits/policy.go`):
   - Policy evaluation logic
   - Operator comparisons
   - Scope filtering

3. **Emergency Stop** (`internal/emergency/stop.go`):
   - Trigger/re-enable flow
   - Circuit breaker logic

4. **gRPC Handlers** (`internal/grpc/server.go`):
   - Order validation flow
   - Emergency stop blocking

**Fix Required**:
Create comprehensive test suite with:
- Unit tests for all calculators and business logic
- Integration tests for database operations
- End-to-end tests for gRPC endpoints

**Estimated Effort**: 2-3 days

---

### 12.2 High Priority Issues

#### 4. Missing Market Data Integration (SEVERITY: HIGH)

**Location**: `/home/mm/dev/b25/services/risk-manager/internal/cache/policy_cache.go:124`

**Issue**: Market price cache expects Market Data Service to populate Redis, but integration not documented.

**Code**:
```go
// GetPrice retrieves price for a symbol
func (c *MarketPriceCache) GetPrice(ctx context.Context, symbol string) (float64, error) {
    key := fmt.Sprintf(priceCacheKey, symbol)  // "market:prices:{symbol}"
    price, err := c.redis.Get(ctx, key).Float64()
    if err != nil {
        if err == redis.Nil {
            return 0, fmt.Errorf("price not found for symbol %s", symbol)
        }
        return 0, fmt.Errorf("get price from redis: %w", err)
    }
    return price, nil
}
```

**Impact**:
- Market orders will fail if price not in Redis
- Fallback to order price for limit orders (acceptable)
- No documented process for populating prices

**Fix Required**:
1. Verify Market Data Service writes to `market:prices:{symbol}` in Redis DB 1
2. Document integration contract
3. Add fallback price source (exchange API?)
4. Add monitoring for stale prices

**Estimated Effort**: 2-4 hours (coordination with Market Data Service owner)

---

#### 5. Hardcoded Database Password in Config (SEVERITY: HIGH - Security)

**Location**: `/home/mm/dev/b25/services/risk-manager/config.yaml:12`

**Issue**: Database password stored in plaintext in config file.

**Code**:
```yaml
database:
  user: b25
  password: JDExqQGCJxncMuKrRwpAmg==   # Plaintext password
```

**Impact**:
- Security risk if config file is committed to version control
- Password visible to anyone with file access
- Violates security best practices

**Fix Required**:
1. Remove password from `config.yaml`
2. Use environment variable: `export RISK_DATABASE_PASSWORD=...`
3. Or use secrets management (Vault, Kubernetes secrets, etc.)
4. Update documentation to reflect environment variable usage

**Quick Fix**:
```bash
# Remove password from config.yaml
sed -i 's/password: .*/password: ""/' config.yaml

# Set environment variable
export RISK_DATABASE_PASSWORD='JDExqQGCJxncMuKrRwpAmg=='

# Run service (will pick up env var)
make run
```

**Estimated Effort**: 30 minutes

---

### 12.3 Medium Priority Issues

#### 6. Emergency Stop Not Synchronized Across Instances

**Location**: `/home/mm/dev/b25/services/risk-manager/internal/emergency/stop.go`

**Issue**: Emergency stop state stored in-memory, not shared across service instances.

**Impact**:
- In multi-instance deployment, emergency stop only blocks orders on triggering instance
- Other instances continue accepting orders
- Inconsistent behavior

**Fix Required**:
1. Move emergency stop state to Redis or PostgreSQL
2. All instances check shared state before approving orders
3. Use pub/sub for real-time emergency stop notifications

**Estimated Effort**: 4-6 hours

---

#### 7. No Position Unwinding Implementation

**Location**: `/home/mm/dev/b25/services/risk-manager/internal/emergency/stop.go:80`

**Issue**: Emergency stop triggers but doesn't actually close positions or cancel orders.

**Code Comments**:
```go
// Trigger initiates an emergency stop
func (m *StopManager) Trigger(...) error {
    // ... update status ...
    // Publish emergency alert
    if err := m.alertPublisher.PublishEmergencyAlert(ctx, reason, triggeredBy); err != nil {
        m.logger.Error("failed to publish emergency alert", zap.Error(err))
        // Continue with emergency stop even if alert fails
    }
    return nil
    // TODO: Trigger position unwinding via Order Execution Service
}
```

**Impact**:
- Emergency stop blocks future orders but doesn't protect existing positions
- Losses can continue to accumulate
- Manual intervention required to close positions

**Fix Required**:
1. Add gRPC client to Order Execution Service
2. Call CancelAllOrders() on emergency stop
3. Call CloseAllPositions() on emergency stop
4. Track progress (orders_canceled, positions_closed)
5. Mark completed when done

**Estimated Effort**: 6-8 hours

---

#### 8. Missing Cache Hit Rate Metrics

**Issue**: Code references policy cache hit rate > 99.5% but metrics not exposed.

**Impact**:
- Cannot verify cache performance claims
- Cannot detect cache degradation

**Fix Required**:
Add metrics in `internal/cache/policy_cache.go`:
```go
var (
    policyCacheHits = promauto.NewCounter(prometheus.CounterOpts{
        Name: "risk_policy_cache_hits_total",
        Help: "Total policy cache hits",
    })
    policyCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
        Name: "risk_policy_cache_misses_total",
        Help: "Total policy cache misses",
    })
)
```

**Estimated Effort**: 1 hour

---

#### 9. No Rate Limiting on gRPC Endpoints

**Issue**: No rate limiting to prevent DoS or runaway clients.

**Impact**:
- Service vulnerable to overload from misbehaving clients
- No backpressure mechanism

**Fix Required**:
Add gRPC interceptor with rate limiting:
```go
import "golang.org/x/time/rate"

rateLimiter := rate.NewLimiter(1000, 100) // 1000 req/s, burst 100

grpcSrv := grpcServer.NewServer(
    grpcServer.UnaryInterceptor(rateLimitInterceptor(rateLimiter)),
    // ... other options
)
```

**Estimated Effort**: 2-3 hours

---

### 12.4 Low Priority Issues

#### 10. Duplicate Monitoring Alerts

**Issue**: If multiple Risk Manager instances run monitoring loops, they publish duplicate alerts.

**Impact**:
- Alert noise
- Higher NATS traffic

**Fix**:
- Use leader election (only one instance runs monitoring)
- Or accept duplication (alert deduplicator mitigates)

**Estimated Effort**: 4-6 hours (leader election)

---

#### 11. No Graceful Degradation

**Issue**: Service fails to start if PostgreSQL is unavailable.

**Impact**:
- Cannot run in degraded mode with default policies
- Tight coupling to database

**Fix**:
Allow service to start with default policies if DB is down:
```go
policies, err := policyRepo.GetActive(ctx)
if err != nil {
    logger.Warn("failed to load policies from database, using defaults", zap.Error(err))
    policies = getDefaultPolicies()
} else {
    logger.Info("policies loaded from database", zap.Int("count", len(policies)))
}
```

**Note**: This pattern already exists in `main.go:91-98`, so this issue is partially mitigated.

---

#### 12. No Metrics for NATS Publishing Failures

**Issue**: NATS publishing failures logged but not metrified.

**Impact**:
- Cannot monitor alert delivery reliability

**Fix**:
Add counter for NATS publish failures:
```go
natsPublishFailures = promauto.NewCounter(prometheus.CounterOpts{
    Name: "risk_nats_publish_failures_total",
    Help: "Total NATS publish failures",
})
```

**Estimated Effort**: 30 minutes

---

### 12.5 Issue Summary Table

| # | Issue | Severity | Impact | Effort | Status |
|---|-------|----------|--------|--------|--------|
| 1 | Mock Account State | CRITICAL | Not production-ready | 4-8h | Open |
| 2 | Dockerfile Merge Conflict | HIGH | Build broken | 10m | Open |
| 3 | No Test Coverage | HIGH | Low confidence | 2-3d | Open |
| 4 | Missing Market Data Integration | HIGH | Market orders may fail | 2-4h | Open |
| 5 | Hardcoded DB Password | HIGH | Security risk | 30m | Open |
| 6 | Emergency Stop Not Synced | MEDIUM | Multi-instance issue | 4-6h | Open |
| 7 | No Position Unwinding | MEDIUM | Losses continue | 6-8h | Open |
| 8 | Missing Cache Metrics | MEDIUM | Can't verify performance | 1h | Open |
| 9 | No Rate Limiting | MEDIUM | DoS vulnerable | 2-3h | Open |
| 10 | Duplicate Alerts | LOW | Alert noise | 4-6h | Open |
| 11 | No Graceful Degradation | LOW | Tight coupling | N/A | Mitigated |
| 12 | Missing NATS Metrics | LOW | Monitoring gap | 30m | Open |

---

## 13. Recommendations

### 13.1 Immediate Actions (Before Production)

#### 1. Implement Account Monitor Integration (CRITICAL)

**Priority**: P0 - Blocker for production use

**Tasks**:
1. Create gRPC client to Account Monitor Service
2. Define clear API contract (required fields, error handling)
3. Replace all `getMockAccountState()` calls with real client
4. Add caching layer (100ms TTL as configured)
5. Implement circuit breaker for Account Monitor failures
6. Add monitoring for Account Monitor latency

**Acceptance Criteria**:
- Risk calculations use real account data
- Service handles Account Monitor unavailability gracefully
- Latency impact < 5ms (within 10ms p99 target)

**Code Template**:
```go
// internal/grpc/account_client.go
type AccountClient struct {
    conn *grpc.ClientConn
    client accountpb.AccountMonitorClient
    cache *AccountStateCache
}

func (c *AccountClient) GetAccountState(ctx context.Context, accountID string) (risk.AccountState, error) {
    // Check cache first
    if cached, err := c.cache.Get(ctx, accountID); err == nil {
        return cached, nil
    }

    // Call Account Monitor
    resp, err := c.client.GetAccountState(ctx, &accountpb.AccountRequest{
        AccountId: accountID,
    })
    if err != nil {
        return risk.AccountState{}, fmt.Errorf("account monitor call failed: %w", err)
    }

    // Convert to risk.AccountState
    state := convertToAccountState(resp)

    // Cache result
    c.cache.Set(ctx, accountID, state)

    return state, nil
}
```

---

#### 2. Fix Dockerfile Merge Conflict (HIGH)

**Priority**: P0 - Prevents deployment

**Tasks**:
1. Resolve merge conflict in Dockerfile
2. Choose appropriate version (likely origin/main)
3. Test Docker build
4. Test Docker image runs correctly

**Command**:
```bash
cd /home/mm/dev/b25/services/risk-manager

# Option 1: Accept origin/main version
git checkout origin/main -- Dockerfile

# Option 2: Manually edit to remove conflict markers
# Edit Dockerfile and remove all lines with <<<<<<<, =======, >>>>>>>

# Test build
docker build -t risk-manager:test .

# Test run
docker run --rm risk-manager:test --help
```

---

#### 3. Add Basic Test Coverage (HIGH)

**Priority**: P1 - Critical for confidence

**Recommended Test Coverage**:
- **Risk Calculator**: 80%+ (core business logic)
- **Policy Engine**: 80%+ (critical validation)
- **gRPC Handlers**: 60%+ (integration-heavy)
- **Overall**: 70%+

**Test Structure**:
```
internal/risk/calculator_test.go
internal/limits/policy_test.go
internal/emergency/stop_test.go
internal/grpc/server_test.go
internal/cache/policy_cache_test.go
```

**Sample Test**:
```go
// internal/risk/calculator_test.go
package risk

import "testing"

func TestCalculateMetrics(t *testing.T) {
    calc := NewCalculator(10.0, 0.20)

    state := AccountState{
        Equity:           100000,
        Balance:          100000,
        MarginUsed:       20000,
        Positions: []Position{
            {Symbol: "BTCUSDT", Notional: 50000},
        },
        PeakEquity:       110000,
        DailyStartEquity: 105000,
    }

    metrics := calc.CalculateMetrics(state)

    // Test margin ratio: 100000 / 20000 = 5.0
    if metrics.MarginRatio != 5.0 {
        t.Errorf("MarginRatio = %f, want 5.0", metrics.MarginRatio)
    }

    // Test leverage: 50000 / 100000 = 0.5
    if metrics.Leverage != 0.5 {
        t.Errorf("Leverage = %f, want 0.5", metrics.Leverage)
    }

    // Test drawdown: (110000 - 100000) / 110000 = 0.0909
    expected := 0.09090909090909091
    if abs(metrics.DrawdownMax - expected) > 0.0001 {
        t.Errorf("DrawdownMax = %f, want %f", metrics.DrawdownMax, expected)
    }
}
```

---

#### 4. Remove Hardcoded Database Password (HIGH - Security)

**Priority**: P1 - Security risk

**Tasks**:
1. Remove password from `config.yaml`
2. Update deployment scripts to use environment variables
3. Document environment variable in README
4. Add warning in config.yaml file

**Config File Change**:
```yaml
# config.yaml
database:
  host: localhost
  port: 5432
  user: b25
  password: ""  # Set via RISK_DATABASE_PASSWORD environment variable
  database: b25_config
```

**Documentation Addition**:
```markdown
## Environment Variables

Required environment variables:

RISK_DATABASE_PASSWORD - PostgreSQL password (REQUIRED)

Optional:
RISK_DATABASE_HOST - Override database host
RISK_REDIS_HOST - Override Redis host
...
```

---

### 13.2 Short-term Improvements (1-2 Weeks)

#### 5. Implement Emergency Stop Position Unwinding

**Tasks**:
1. Design Order Execution Service integration
2. Implement gRPC calls to cancel orders and close positions
3. Track unwinding progress (orders_canceled, positions_closed)
4. Add timeout and retry logic
5. Update emergency stop status when complete

**Integration Points**:
- Order Execution Service: `CancelAllOrders(accountID)`
- Order Execution Service: `CloseAllPositions(accountID)`

---

#### 6. Add Market Data Service Integration

**Tasks**:
1. Coordinate with Market Data Service team on Redis schema
2. Verify Market Data Service writes to `market:prices:{symbol}` in DB 1
3. Add monitoring for price staleness
4. Add fallback price source (exchange API or cached last price)
5. Document integration in README

**Redis Key Contract**:
```
Key: market:prices:{symbol}
Value: Float64 (current price)
TTL: 100ms (Market Data Service should update continuously)
DB: 1 (as configured in risk.market_data_redis_db)
```

---

#### 7. Synchronize Emergency Stop Across Instances

**Option A: Redis-based State** (Recommended)
```go
// Store emergency stop in Redis
func (m *StopManager) IsActive() bool {
    val, err := m.redis.Get(ctx, "emergency_stop:active").Bool()
    if err != nil {
        return false
    }
    return val
}

func (m *StopManager) Trigger(...) error {
    m.redis.Set(ctx, "emergency_stop:active", true, 0)
    // ... rest of logic
}
```

**Option B: PostgreSQL-based State**
```sql
CREATE TABLE emergency_stop_state (
    id INT PRIMARY KEY DEFAULT 1,
    is_active BOOLEAN DEFAULT FALSE,
    stopped_at TIMESTAMPTZ,
    reason TEXT
);
```

---

#### 8. Add Rate Limiting

**Implementation**:
```go
// internal/grpc/middleware.go
func rateLimitInterceptor(limiter *rate.Limiter) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        if !limiter.Allow() {
            return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
        }
        return handler(ctx, req)
    }
}
```

**Configuration**:
```yaml
grpc:
  rate_limit: 1000  # req/s
  rate_limit_burst: 100
```

---

### 13.3 Medium-term Enhancements (1-3 Months)

#### 9. Add Advanced Policy Features

**Enhancements**:
1. **Time-based Policies**: Different limits for different times of day
2. **Volatility-adjusted Limits**: Reduce leverage during high volatility
3. **Strategy-specific Limits**: Different limits per trading strategy
4. **Symbol Blacklist**: Block trading on specific symbols
5. **Velocity Limits**: Max orders per minute, max position changes per hour

**Example**:
```go
type Policy struct {
    // ... existing fields
    TimeWindows []TimeWindow  // Active only during these times
    VolatilityMultiplier float64  // Adjust threshold based on volatility
}
```

---

#### 10. Implement Policy Versioning and A/B Testing

**Features**:
1. Test new policies on subset of accounts
2. Gradual rollout of policy changes
3. Compare violation rates between policy versions

**Schema Addition**:
```sql
ALTER TABLE risk_policies ADD COLUMN rollout_percentage INT DEFAULT 100;
ALTER TABLE risk_policies ADD COLUMN test_group VARCHAR(50);
```

---

#### 11. Add Machine Learning Risk Scoring

**Concept**:
- Train ML model on historical violations
- Predict likelihood of violation before it occurs
- Early warning system for risk managers

**Integration**:
```go
func (s *RiskServer) CheckOrder(...) {
    // ... existing validation

    // ML risk score
    riskScore := s.mlModel.PredictRiskScore(accountState, order)
    if riskScore > 0.8 {
        // High risk warning (soft violation)
    }
}
```

---

#### 12. Create Risk Dashboard

**Features**:
- Real-time risk metrics visualization
- Policy violation history charts
- Emergency stop timeline
- Account risk heatmap
- Alert feed

**Tech Stack**:
- Frontend: React + D3.js
- Backend: Subscribe to NATS risk.metrics topic
- Storage: InfluxDB for time-series data

---

### 13.4 Operational Improvements

#### 13. Add Comprehensive Monitoring

**Metrics to Add**:
```go
// Cache performance
policy_cache_hits_total
policy_cache_misses_total
price_cache_hits_total
price_cache_misses_total

// NATS reliability
nats_publish_success_total
nats_publish_failures_total
nats_reconnects_total

// Database performance
db_query_duration_seconds{operation="get_policies"}
db_query_duration_seconds{operation="record_violation"}
db_connection_pool_size

// Account Monitor integration
account_monitor_call_duration_seconds
account_monitor_failures_total
```

---

#### 14. Implement Distributed Tracing

**Add OpenTelemetry**:
```go
import "go.opentelemetry.io/otel"

func (s *RiskServer) CheckOrder(ctx context.Context, req *pb.OrderRiskRequest) (*pb.OrderRiskResponse, error) {
    ctx, span := otel.Tracer("risk-manager").Start(ctx, "CheckOrder")
    defer span.End()

    // Trace through all components
    // - Account Monitor call
    // - Price cache lookup
    // - Policy evaluation
    // - etc.
}
```

**Benefits**:
- End-to-end request tracing
- Identify bottlenecks
- Visualize service dependencies

---

#### 15. Add Chaos Engineering Tests

**Scenarios**:
1. Database connection failure
2. Redis unavailability
3. NATS connection loss
4. Account Monitor timeout
5. High latency (network degradation)

**Tool**: Chaos Mesh or Toxiproxy

---

#### 16. Improve Documentation

**Add**:
1. **Architecture Decision Records (ADRs)**: Why certain design choices were made
2. **Runbook**: Step-by-step incident response procedures
3. **API Examples**: More gRPC request/response examples
4. **Performance Tuning Guide**: Configuration recommendations for different loads
5. **Security Audit**: Document security considerations and threat model

---

### 13.5 Code Quality Improvements

#### 17. Add Linting and Static Analysis

**Tools**:
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Add .golangci.yml config
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosec  # Security checks
    - gocyclo  # Cyclomatic complexity
    - ineffassign
    - unused

# Run
make lint
```

---

#### 18. Add Pre-commit Hooks

**Setup**:
```bash
# .git/hooks/pre-commit
#!/bin/bash
make fmt
make vet
make test
```

---

#### 19. Improve Error Handling

**Add Custom Error Types**:
```go
type RiskError struct {
    Code    string
    Message string
    Details map[string]interface{}
}

var (
    ErrInsufficientMargin = &RiskError{Code: "INSUFFICIENT_MARGIN", ...}
    ErrEmergencyStop = &RiskError{Code: "EMERGENCY_STOP", ...}
    ErrPolicyViolation = &RiskError{Code: "POLICY_VIOLATION", ...}
)
```

---

#### 20. Add Structured Logging with Context

**Improve**:
```go
// Add request ID to all logs
func (s *RiskServer) CheckOrder(ctx context.Context, req *pb.OrderRiskRequest) (*pb.OrderRiskResponse, error) {
    requestID := uuid.New().String()
    logger := s.logger.With(
        zap.String("request_id", requestID),
        zap.String("order_id", req.OrderId),
        zap.String("account_id", req.AccountId),
    )

    logger.Info("processing order risk check")
    // ... rest of logic
}
```

---

### 13.6 Recommendations Summary

| Category | Recommendation | Priority | Effort | Impact |
|----------|---------------|----------|--------|--------|
| **Immediate** | Account Monitor Integration | P0 | 4-8h | CRITICAL |
| **Immediate** | Fix Dockerfile | P0 | 10m | HIGH |
| **Immediate** | Add Tests | P1 | 2-3d | HIGH |
| **Immediate** | Remove Hardcoded Password | P1 | 30m | HIGH |
| **Short-term** | Position Unwinding | P1 | 6-8h | HIGH |
| **Short-term** | Market Data Integration | P1 | 2-4h | MEDIUM |
| **Short-term** | Sync Emergency Stop | P2 | 4-6h | MEDIUM |
| **Short-term** | Add Rate Limiting | P2 | 2-3h | MEDIUM |
| **Medium-term** | Advanced Policies | P3 | 1-2w | MEDIUM |
| **Medium-term** | Policy Versioning | P3 | 1w | LOW |
| **Medium-term** | ML Risk Scoring | P4 | 2-3w | LOW |
| **Medium-term** | Risk Dashboard | P4 | 2-3w | MEDIUM |
| **Operational** | Monitoring | P2 | 1-2d | HIGH |
| **Operational** | Distributed Tracing | P3 | 2-3d | MEDIUM |
| **Operational** | Chaos Testing | P4 | 1w | LOW |
| **Operational** | Documentation | P3 | 1w | MEDIUM |
| **Code Quality** | Linting | P3 | 2h | LOW |
| **Code Quality** | Pre-commit Hooks | P3 | 1h | LOW |
| **Code Quality** | Error Handling | P3 | 1-2d | MEDIUM |
| **Code Quality** | Structured Logging | P3 | 4h | LOW |

---

## Conclusion

The Risk Manager Service is a well-architected, performance-focused service with clear separation of concerns and comprehensive policy engine. However, it has **critical blockers** that prevent production deployment:

**Blockers**:
1. Mock account state data (not using real Account Monitor)
2. Dockerfile merge conflict (prevents builds)
3. No test coverage (0%)

**After addressing blockers**, the service provides:
- Fast pre-trade validation (<10ms p99)
- Flexible multi-layer policy system
- Emergency stop protection
- Real-time risk monitoring
- Comprehensive metrics and alerting

**Production Readiness Score**: 4/10 (after fixing blockers: 7/10)

**Recommended Path to Production**:
1. Week 1: Fix blockers (Account Monitor integration, Dockerfile, basic tests, security)
2. Week 2: Add position unwinding, market data integration, rate limiting
3. Week 3: Improve monitoring, documentation, operational readiness
4. Week 4: Load testing, chaos testing, final security audit

---

**End of Audit Report**
