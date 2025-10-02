# Sub-Systems Architecture - Microservices Design

**Purpose:** Definition of autonomous sub-systems for independent development, testing, and containerized deployment.

**Last Updated:** 2025-10-02
**Version:** 1.0

---

## Sub-System Overview

Each sub-system is designed to:
- Run as an independent microservice container
- Have its own technology stack
- Include isolated testing infrastructure
- Provide observability UI/tooling
- Communicate via well-defined APIs

---

## 1. Market Data Service

**Responsibility:** Real-time market data ingestion, order book management, and distribution

**Core Functions:**
- WebSocket connection management to exchanges
- Order book replica maintenance and validation
- Trade stream processing and normalization
- Data distribution via pub/sub and shared memory

**Technology Independence:**
- Can use Rust/C++ for ultra-low latency
- Or Go/Node.js for easier development
- Independent data serialization format choice

**Testing:**
- Mock WebSocket server for exchange simulation
- Order book accuracy validation suite
- Latency benchmarking tools
- Reconnection scenario testing

**Observability UI:**
- WebSocket connection health dashboard
- Order book visualization
- Message throughput metrics
- Latency histograms (p50, p95, p99)

**Interfaces:**
- **Output:** Pub/sub topics (market_data.{symbol}.orderbook, market_data.{symbol}.trades)
- **Output:** Shared memory ring buffer (local IPC)
- **Output:** Time-series DB writes
- **Health:** HTTP health check endpoint

---

## 2. Order Execution Service

**Responsibility:** Order lifecycle management and exchange communication

**Core Functions:**
- Order validation and pre-flight checks
- Exchange REST API communication
- Order state tracking (NEW → FILLED/CANCELED/REJECTED)
- Fill event processing
- Rate limiting and circuit breaker

**Technology Independence:**
- Any language with async HTTP support
- Exchange adapter pattern for multi-exchange support

**Testing:**
- Exchange API mock server
- Order validation test suite
- Rate limiter verification
- Circuit breaker scenario testing
- Idempotency verification

**Observability UI:**
- Active orders dashboard
- Order submission latency charts
- Circuit breaker state indicators
- Maker/taker ratio visualization
- Order rejection reason analysis

**Interfaces:**
- **Input:** RPC endpoint for order requests
- **Output:** Pub/sub for fill events and order updates
- **Output:** Time-series DB for order history
- **Health:** HTTP health check endpoint

---

## 3. Strategy Engine Service

**Responsibility:** Trading strategy execution and signal generation

**Core Functions:**
- Plugin-based strategy loading
- Market data consumption and routing
- Signal generation and aggregation
- Risk filtering
- Position tracking
- Hot-reload capability

**Technology Independence:**
- Core engine in any language
- Strategy plugins can be scripts (Python/JS) or compiled
- Flexible plugin interface design

**Testing:**
- Strategy backtesting framework
- Mock market data replay
- Signal validation suite
- Risk filter test scenarios
- Plugin isolation testing

**Observability UI:**
- Active strategy dashboard
- Signal generation timeline
- Strategy performance comparison
- Position tracking per strategy
- Signal rejection reasons

**Interfaces:**
- **Input:** Subscribe to market data topics
- **Input:** Subscribe to fill events
- **Output:** RPC calls to order execution service
- **Query:** gRPC API for strategy state
- **Health:** HTTP health check endpoint

---

## 4. Account Monitor Service

**Responsibility:** Balance tracking, P&L calculation, and position reconciliation

**Core Functions:**
- Real-time balance tracking
- Position state management
- P&L calculation (realized and unrealized)
- Position reconciliation with exchange
- Risk threshold monitoring and alerting

**Technology Independence:**
- Any language with REST/WebSocket support
- Independent database schema design

**Testing:**
- P&L calculation unit tests
- Reconciliation algorithm verification
- Alert threshold testing
- Balance update race condition testing

**Observability UI:**
- Account balance dashboard
- Position breakdown visualization
- P&L charts (daily, cumulative)
- Win rate and trade statistics
- Risk metrics (margin ratio, leverage, drawdown)

**Interfaces:**
- **Input:** Subscribe to fill events
- **Input:** Exchange user data stream WebSocket
- **Query:** gRPC API for account state queries
- **Output:** Pub/sub for alerts
- **Output:** Time-series DB for P&L history
- **Health:** HTTP health check endpoint

---

## 5. Dashboard Server Service

**Responsibility:** State aggregation and real-time broadcasting to UI clients

**Core Functions:**
- Multi-source state aggregation
- WebSocket server for clients
- Update rate differentiation (TUI vs Web)
- Efficient message serialization
- Client subscription management

**Technology Independence:**
- Any language with WebSocket support
- Choice of serialization format (JSON, MessagePack, Protobuf)

**Testing:**
- WebSocket client simulator
- Load testing (100+ concurrent clients)
- Message serialization benchmarks
- State aggregation accuracy tests

**Observability UI:**
- Connected clients monitor
- Message throughput metrics
- Broadcast latency tracking
- Client subscription visualization

**Interfaces:**
- **Input:** Query/subscribe to all data sources (market data, orders, account)
- **Output:** WebSocket server for TUI/Web clients
- **Output:** REST API for historical queries
- **Health:** HTTP health check endpoint

---

## 6. Risk Manager Service

**Responsibility:** Global risk management and emergency controls

**Core Functions:**
- Real-time risk metric calculation
- Multi-layer limit enforcement (position, leverage, drawdown)
- Emergency stop mechanism
- Risk policy configuration
- Pre-trade risk checks (as service for execution engine)

**Technology Independence:**
- Any language with fast computation
- Rule engine can be simple or complex (Drools, custom)

**Testing:**
- Risk calculation unit tests
- Limit enforcement scenario testing
- Emergency stop simulation
- Policy change testing

**Observability UI:**
- Risk metrics dashboard
- Limit utilization indicators
- Historical risk violations
- Emergency stop status
- Risk policy configuration UI

**Interfaces:**
- **Input:** Query account state from account monitor
- **Input:** Query market data for mark prices
- **Input:** RPC for pre-trade risk checks
- **Output:** Pub/sub for risk alerts
- **Output:** Config DB for risk policies
- **Health:** HTTP health check endpoint

---

## 7. Configuration Service

**Responsibility:** Centralized configuration management

**Core Functions:**
- Strategy configuration storage
- Risk limit definitions
- Symbol metadata management
- System settings
- Configuration versioning
- Hot-reload trigger mechanism

**Technology Independence:**
- Any language with database access
- Config format agnostic (JSON, YAML, TOML)

**Testing:**
- Configuration validation tests
- Version rollback testing
- Hot-reload simulation
- Access control testing

**Observability UI:**
- Configuration browser and editor
- Version history and diff viewer
- Active configuration dashboard
- Validation status indicators

**Interfaces:**
- **Storage:** Relational database (PostgreSQL, MySQL, etc.)
- **Query:** gRPC/REST API for config reads
- **Update:** gRPC/REST API for config writes
- **Notify:** Pub/sub for config change events
- **Health:** HTTP health check endpoint

---

## 8. Metrics & Observability Service

**Responsibility:** System-wide metrics collection and visualization

**Core Functions:**
- Metrics scraping from all services
- Time-series storage
- Alert rule evaluation
- Metrics API for dashboards
- Log aggregation (optional)

**Technology Independence:**
- Prometheus + Grafana (standard choice)
- Or InfluxDB + Chronograf
- Or custom solution

**Testing:**
- Metric collection verification
- Alert rule testing
- Query performance testing
- Retention policy testing

**Observability UI:**
- Grafana dashboards (or equivalent)
- System health overview
- Performance metrics
- Business metrics (P&L, trades, etc.)
- Alert management

**Interfaces:**
- **Input:** Scrape metrics from all service endpoints
- **Input:** Receive metrics via push (optional)
- **Storage:** Time-series database
- **Query:** Prometheus/InfluxDB query API
- **Health:** HTTP health check endpoint

---

## 9. Terminal UI (TUI) Service

**Responsibility:** Real-time terminal-based user interface

**Core Functions:**
- WebSocket client to dashboard server
- Real-time rendering (100ms updates)
- Keyboard-driven navigation
- Multi-panel layout (positions, orders, orderbook, AI signals)
- Manual trading controls

**Technology Independence:**
- Language: Rust (ratatui), Go (tview), Python (textual), etc.
- Runs as standalone binary

**Testing:**
- Rendering tests (snapshot testing)
- WebSocket reconnection testing
- Keyboard input handling tests
- Performance testing (low CPU usage)

**Observability UI:**
- Self-monitoring (connection status, latency)
- Embedded system health indicators

**Interfaces:**
- **Input:** WebSocket client to dashboard server
- **Output:** Direct rendering to terminal

---

## 10. Web Dashboard Service

**Responsibility:** Web-based user interface

**Core Functions:**
- WebSocket client to dashboard server
- Responsive web UI (250ms updates)
- Interactive charts and visualizations
- Manual trading interface
- Mobile-responsive design

**Technology Independence:**
- Frontend: React/Vue/Svelte/Angular
- Backend: Optional API gateway (if needed)
- Runs as static site or with lightweight server

**Testing:**
- Component testing (Jest, Vitest)
- E2E testing (Playwright, Cypress)
- Visual regression testing
- WebSocket reconnection testing

**Observability UI:**
- Built-in system health indicators
- Connection status display
- Latency metrics

**Interfaces:**
- **Input:** WebSocket client to dashboard server
- **Input:** Optional REST API for historical queries
- **Output:** Static assets served via HTTP

---

## Communication Architecture

### Inter-Service Communication

**Synchronous (Request-Reply):**
- gRPC for typed, efficient RPC
- REST/HTTP for simpler integrations

**Asynchronous (Pub/Sub):**
- Redis Pub/Sub (simple, in-memory)
- NATS (lightweight, cloud-native)
- RabbitMQ (feature-rich, persistent)
- Kafka (high-throughput, persistent)

**Shared Memory:**
- Only within single-host deployments
- Market data pipeline → Strategy engine (same machine)

### Data Flow Summary

```
Exchange WebSocket → Market Data Service
                   ↓ (pub/sub)
                Strategy Engine Service
                   ↓ (RPC)
                Order Execution Service
                   ↓ (REST/WS)
                Exchange REST API
                   ↓ (WebSocket user data)
                Account Monitor Service

All Services → Dashboard Server Service → TUI/Web Clients
All Services → Metrics & Observability Service
All Services → Configuration Service (query)
```

---

## Deployment Recommendations

### Container Per Service

```yaml
# docker-compose.yml example structure
services:
  market-data:
    build: ./market-data
    networks: [trading-net]
    ports: [9090:9090]  # metrics

  order-execution:
    build: ./order-execution
    networks: [trading-net]
    ports: [9091:9091]

  strategy-engine:
    build: ./strategy-engine
    networks: [trading-net]
    ports: [9092:9092]

  # ... etc
```

### Technology Stack Independence

Each service directory contains:
```
service-name/
├── src/               # Source code in any language
├── tests/             # Test suite
├── Dockerfile         # Container definition
├── README.md          # Service documentation
├── config.example     # Configuration template
└── ui/                # Observability UI (if applicable)
```

---

## Development Workflow

### Independent Development

1. **Service Owner:** Each service can have dedicated developer(s)
2. **Technology Choice:** Team picks best tool for the job
3. **Contract-First:** Define APIs/interfaces before implementation
4. **Mock Dependencies:** Use mocks for other services during development

### Integration Testing

1. **Contract Testing:** Verify API contracts
2. **Docker Compose:** Test all services together locally
3. **Staging Environment:** Full integration testing
4. **Canary Deployments:** Gradual rollout to production

---

## Scaling Strategy

### Horizontal Scaling

- **Strategy Engine:** Multiple instances behind load balancer
- **Dashboard Server:** Multiple instances for HA
- **Order Execution:** Single instance (or active-passive HA)

### Vertical Scaling

- **Market Data Service:** Dedicated high-performance server
- **Metrics Service:** Larger storage capacity

---

## Monitoring Requirements

Each service must expose:

1. **Health Check:** `/health` endpoint (liveness, readiness, startup)
2. **Metrics:** `/metrics` endpoint (Prometheus format)
3. **Logging:** Structured JSON logs to stdout/stderr
4. **Tracing:** Correlation IDs in all inter-service calls

---

**End of Document**

This architecture enables:
- ✅ Independent development and deployment
- ✅ Technology diversity (use best tool per service)
- ✅ Isolated testing and validation
- ✅ Microservices/container deployment
- ✅ Built-in observability per service
