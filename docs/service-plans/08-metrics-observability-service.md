# Metrics & Observability Service - Development Plan

**Service ID:** 08
**Service Name:** Metrics & Observability Service
**Purpose:** System-wide metrics collection, visualization, alerting, and log aggregation
**Last Updated:** 2025-10-02
**Version:** 1.0

---

## Table of Contents

1. [Technology Stack Recommendation](#1-technology-stack-recommendation)
2. [Architecture Design](#2-architecture-design)
3. [Development Phases](#3-development-phases)
4. [Implementation Details](#4-implementation-details)
5. [Dashboard Specifications](#5-dashboard-specifications)
6. [Alert Rules](#6-alert-rules)
7. [Deployment](#7-deployment)
8. [Testing](#8-testing)
9. [Configuration Files](#9-configuration-files)

---

## 1. Technology Stack Recommendation

### 1.1 Recommended Stack: Prometheus + Grafana

**Primary Choice: Prometheus Ecosystem**

**Rationale:**
- Industry standard for microservices monitoring
- Pull-based model ideal for service discovery
- Powerful PromQL query language
- Excellent Grafana integration
- Active community and extensive exporters
- Battle-tested in production HFT environments

**Stack Components:**

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Metrics Storage** | Prometheus | Time-series database and scraping engine |
| **Visualization** | Grafana | Dashboard creation and visualization |
| **Alerting** | Alertmanager | Alert routing, grouping, and notification |
| **Log Aggregation** | Loki | Optional log aggregation (Prometheus-like for logs) |
| **Service Discovery** | File SD / Docker SD | Dynamic target discovery |
| **Exporters** | Node Exporter, cAdvisor | System and container metrics |

### 1.2 Alternative Stack Comparison

#### Option 2: InfluxDB + Chronograf

**Pros:**
- Better write performance for high-cardinality metrics
- More flexible data model (tags vs labels)
- Built-in retention policies
- SQL-like query language (InfluxQL/Flux)

**Cons:**
- Smaller ecosystem than Prometheus
- Grafana integration less mature than native Prometheus
- Fewer community exporters
- Push model requires more infrastructure

**Use Case:** Choose if you need >10M metrics/sec or extreme cardinality

#### Option 3: VictoriaMetrics + Grafana

**Pros:**
- Prometheus-compatible but more efficient
- Better compression and query performance
- Lower resource usage
- Drop-in replacement for Prometheus

**Cons:**
- Smaller community
- Less mature than Prometheus
- Commercial features for clustering

**Use Case:** Choose for large-scale deployments (>100 services)

### 1.3 Final Recommendation

**Selected Stack:**
```yaml
Core Stack:
  - Prometheus 2.48+ (metrics storage + scraping)
  - Grafana 10.2+ (visualization)
  - Alertmanager 0.26+ (alerting)
  - Loki 2.9+ (log aggregation)
  - Promtail (log shipping)

Supporting Tools:
  - Node Exporter 1.7+ (host metrics)
  - cAdvisor (container metrics)
  - Blackbox Exporter (endpoint probing)
  - Custom exporters (per service)
```

**Technology Justification:**
- **Proven:** Used by 80%+ of containerized deployments
- **Integrated:** Native Grafana support with templating
- **Extensible:** Easy custom exporter development
- **Cloud-native:** Kubernetes-native but works standalone
- **Cost:** 100% open source with no licensing fees

---

## 2. Architecture Design

### 2.1 Metrics Collection Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Prometheus Server                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Scraper   │  │   TSDB      │  │   Alert Evaluator   │ │
│  │   Engine    │→ │   Storage   │→ │   (PromQL Rules)    │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└────────┬──────────────────┬────────────────────┬────────────┘
         │                  │                    │
         │ (scrape)         │ (query)            │ (alerts)
         ↓                  ↓                    ↓
┌─────────────────┐  ┌─────────────┐  ┌──────────────────┐
│ Service Metrics │  │   Grafana   │  │  Alertmanager    │
│   Endpoints     │  │  Dashboards │  │   (Routing)      │
│  /metrics:9090  │  │             │  └────────┬─────────┘
│  /metrics:9091  │  │             │           │
│  /metrics:909X  │  │             │           ↓
└─────────────────┘  └─────────────┘  ┌──────────────────┐
         ↑                             │  Notifications   │
         │                             │  - Slack         │
┌────────┴─────────┐                  │  - Email         │
│  Custom Metrics  │                  │  - PagerDuty     │
│    Exporters     │                  └──────────────────┘
└──────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      Loki (Logs)                            │
│  ┌─────────────┐  ┌─────────────┐                          │
│  │  Ingester   │→ │  Log Store  │                          │
│  └─────────────┘  └─────────────┘                          │
└────────┬────────────────────────────────────────────────────┘
         ↑
         │ (push)
┌────────┴─────────┐
│    Promtail      │
│  (Log Shipper)   │
└────────┬─────────┘
         ↑
         │ (docker logs)
┌────────┴─────────┐
│  All Services    │
│  (stdout/stderr) │
└──────────────────┘
```

### 2.2 Time-Series Storage Design

**Prometheus TSDB Configuration:**

```yaml
Storage Layout:
  - Data Directory: /prometheus/data
  - Retention Time: 15 days (configurable)
  - Retention Size: 50GB (configurable)
  - WAL Compression: Enabled
  - Block Size: 2h (default)

Data Model:
  Metric Name: {label1="value1", label2="value2"} value timestamp

  Example:
    order_execution_latency_ms{
      service="order-execution",
      exchange="binance",
      symbol="BTCUSDT",
      order_type="limit",
      quantile="0.95"
    } 4.2 1696348800000
```

**Cardinality Management:**
- Maximum unique label combinations: <1M per metric
- Label design: Avoid high-cardinality labels (user IDs, order IDs)
- Use relabeling to drop unnecessary labels
- Monitor cardinality with `prometheus_tsdb_symbol_table_size_bytes`

**Performance Tuning:**
```yaml
Scrape Configuration:
  - Default Scrape Interval: 15s
  - High-Priority Services: 5s (market-data, order-execution)
  - Low-Priority Services: 30s (config-service)
  - Scrape Timeout: 10s

Query Performance:
  - Query Max Samples: 50M
  - Query Timeout: 2m
  - Max Concurrent Queries: 20
```

### 2.3 Alert Rule Evaluation

```
Alert Evaluation Pipeline:

1. Scrape Metrics (every 15s)
   ↓
2. Store in TSDB
   ↓
3. Evaluate Rules (every 15s)
   ↓
4. Check Alert State Transitions
   - inactive → pending (condition met, waiting for 'for' duration)
   - pending → firing (condition held for 'for' duration)
   - firing → inactive (condition no longer met)
   ↓
5. Send to Alertmanager
   ↓
6. Group & Deduplicate
   ↓
7. Route to Receivers (Slack, Email, PagerDuty)
   ↓
8. Apply Inhibition Rules (suppress related alerts)
```

**Alert State Management:**
- Alerts stored in Prometheus until resolved
- Alertmanager maintains firing state with deduplication
- Silences managed in Alertmanager (not Prometheus)
- Alert annotations support templating with metric values

### 2.4 Dashboard Organization

**Grafana Folder Structure:**
```
Grafana/
├── System Health/
│   ├── 01-Overview Dashboard (all services health)
│   ├── 02-Infrastructure Metrics (CPU, RAM, disk, network)
│   └── 03-Service Discovery (target status)
│
├── Performance/
│   ├── 10-Latency Dashboard (p50, p95, p99 for all services)
│   ├── 11-Throughput Dashboard (requests/sec, messages/sec)
│   └── 12-Error Rates (4xx, 5xx, failures)
│
├── Business Metrics/
│   ├── 20-P&L Dashboard (realized, unrealized, cumulative)
│   ├── 21-Trading Activity (orders, fills, volume)
│   ├── 22-Strategy Performance (per-strategy P&L, win rate)
│   └── 23-Fee Analysis (maker vs taker, fee costs)
│
├── Service Dashboards/
│   ├── 30-Market Data Service
│   ├── 31-Order Execution Service
│   ├── 32-Strategy Engine Service
│   ├── 33-Account Monitor Service
│   ├── 34-Risk Manager Service
│   └── 35-Configuration Service
│
└── Alerts/
    └── 40-Alert Dashboard (active alerts, alert history)
```

### 2.5 Log Aggregation Pipeline

**Loki Architecture:**

```yaml
Components:
  - Distributor: Receives logs from Promtail
  - Ingester: Batches logs and writes to storage
  - Querier: Handles LogQL queries
  - Index Store: Metadata index (labels, timestamps)
  - Chunk Store: Actual log content

Storage Backend:
  - Filesystem (development/small deployments)
  - S3/GCS (production, long-term retention)

Retention:
  - Recent Logs: 7 days (queryable)
  - Archive: 90 days (compressed storage)
```

**Log Label Strategy:**
```yaml
Labels (indexed, for filtering):
  - service_name: market-data, order-execution, etc.
  - level: debug, info, warn, error
  - environment: dev, staging, prod

Content (full-text search):
  - Message: log message content
  - Trace ID: correlation ID for distributed tracing
  - Additional fields: JSON structured logging
```

---

## 3. Development Phases

### Phase 1: Prometheus & Node Exporter Setup (Week 1, Days 1-2)

**Objectives:**
- Deploy Prometheus server
- Configure Node Exporter for host metrics
- Set up basic scraping configuration

**Deliverables:**
- Running Prometheus instance
- Host metrics collection (CPU, RAM, disk, network)
- Prometheus web UI accessible
- Basic health checks

**Tasks:**
1. Create `prometheus/` directory structure
2. Write `prometheus.yml` configuration
3. Create Docker Compose service for Prometheus
4. Deploy Node Exporter on host
5. Verify metrics scraping with Prometheus UI
6. Configure data retention (15 days, 50GB)

**Acceptance Criteria:**
- [ ] Prometheus UI shows "up" for Node Exporter target
- [ ] Host metrics queryable (e.g., `node_cpu_seconds_total`)
- [ ] TSDB storing data with correct retention policy

---

### Phase 2: Service Discovery & Scraping Config (Week 1, Days 3-4)

**Objectives:**
- Configure scraping for all microservices
- Implement service discovery mechanism
- Add cAdvisor for container metrics

**Deliverables:**
- All services exposing `/metrics` endpoints
- Prometheus scraping all service targets
- Container metrics collection
- Service-level metrics visible

**Tasks:**
1. Define `/metrics` endpoint standard for all services
2. Configure file-based or Docker service discovery
3. Add scrape configs for each service:
   - market-data (port 9090)
   - order-execution (port 9091)
   - strategy-engine (port 9092)
   - account-monitor (port 9093)
   - risk-manager (port 9094)
   - config-service (port 9095)
   - dashboard-server (port 9096)
4. Deploy cAdvisor for container metrics
5. Create relabeling rules for clean metric naming

**Acceptance Criteria:**
- [ ] All services show "up" in Prometheus targets page
- [ ] Service-specific metrics queryable (e.g., `order_execution_total`)
- [ ] Container metrics available (CPU, memory per container)
- [ ] No scrape timeout errors

---

### Phase 3: Grafana Dashboard Setup (Week 1, Days 5-6)

**Objectives:**
- Deploy Grafana
- Create initial dashboards
- Configure Prometheus datasource

**Deliverables:**
- Grafana running with Prometheus datasource
- System Health Overview dashboard
- Infrastructure Metrics dashboard
- Service-specific dashboards (templates)

**Tasks:**
1. Deploy Grafana container
2. Configure Prometheus as datasource
3. Create dashboard templates:
   - System Health Overview
   - Infrastructure Metrics
   - Latency Dashboard (USE method)
   - Error Rate Dashboard (RED method)
4. Set up dashboard variables (service selector, time range)
5. Configure dashboard auto-refresh (15s)
6. Create dashboard JSON exports for version control

**Acceptance Criteria:**
- [ ] Grafana accessible at http://localhost:3000
- [ ] Prometheus datasource connected and querying
- [ ] 4+ dashboards created and functional
- [ ] Variables working correctly (service selector)
- [ ] Dashboards auto-refreshing

---

### Phase 4: Alert Rule Configuration (Week 2, Days 1-3)

**Objectives:**
- Define comprehensive alert rules
- Deploy Alertmanager
- Configure notification channels

**Deliverables:**
- Alert rules for critical system events
- Alertmanager routing configuration
- Slack/Email notification integration
- Alert dashboard in Grafana

**Tasks:**
1. Write alert rule groups:
   - Infrastructure alerts (CPU, memory, disk)
   - Service health alerts (down, restart, errors)
   - Performance alerts (latency, throughput)
   - Business alerts (P&L drawdown, order failures)
2. Deploy Alertmanager
3. Configure alert routing:
   - Critical → PagerDuty + Slack
   - Warning → Slack
   - Info → Log only
4. Set up inhibition rules (suppress cascading alerts)
5. Create silences for maintenance windows
6. Build Alert Overview dashboard

**Acceptance Criteria:**
- [ ] Alert rules loaded in Prometheus
- [ ] Test alert fires and routes correctly
- [ ] Slack webhook receives notifications
- [ ] Alert dashboard shows active alerts
- [ ] Inhibition rules prevent alert storms

---

### Phase 5: Loki Log Aggregation (Week 2, Days 4-5)

**Objectives:**
- Deploy Loki for log aggregation
- Configure Promtail log shippers
- Integrate logs in Grafana

**Deliverables:**
- Loki running and ingesting logs
- Promtail collecting logs from all services
- Log exploration UI in Grafana
- Correlated metrics + logs view

**Tasks:**
1. Deploy Loki container
2. Configure Loki storage (filesystem or S3)
3. Deploy Promtail on each service node
4. Configure Promtail to scrape Docker logs
5. Define log labels (service, level, environment)
6. Add Loki datasource to Grafana
7. Create log exploration dashboard
8. Link logs to metrics (via trace IDs)

**Acceptance Criteria:**
- [ ] Loki ingesting logs from all services
- [ ] Logs queryable via LogQL in Grafana
- [ ] Log labels correctly applied
- [ ] Logs linked to metrics in dashboards
- [ ] Log retention policy enforced (7 days)

---

### Phase 6: Custom Metrics Exporters (Week 2, Day 6 - Week 3, Day 2)

**Objectives:**
- Build custom exporters for business metrics
- Instrument services with Prometheus client libraries
- Expose advanced HFT-specific metrics

**Deliverables:**
- Custom exporter for exchange API metrics
- Business metrics endpoints in services
- Trading-specific metrics (order flow, slippage, etc.)

**Tasks:**
1. Add Prometheus client libraries to each service:
   - Go: `github.com/prometheus/client_golang`
   - Python: `prometheus_client`
   - Node.js: `prom-client`
   - Rust: `prometheus`
2. Instrument services with custom metrics:
   - **Market Data:** WebSocket latency, orderbook depth, message rate
   - **Order Execution:** Order submission latency, fill rate, rejection reasons
   - **Strategy Engine:** Signal generation rate, active strategies, position size
   - **Account Monitor:** Balance, unrealized P&L, margin ratio
   - **Risk Manager:** Risk limit utilization, violations, emergency stops
3. Create histogram metrics for latency (proper buckets)
4. Create gauge metrics for state (positions, balances)
5. Create counter metrics for events (orders, fills, errors)
6. Add metric labels for dimensionality (symbol, exchange, strategy)

**Acceptance Criteria:**
- [ ] All services expose custom business metrics
- [ ] Metrics follow naming conventions (unit suffixes)
- [ ] Histogram buckets tuned for HFT latencies (0.1ms-10ms)
- [ ] Metrics documentation added to service READMEs
- [ ] No high-cardinality labels (verified)

---

## 4. Implementation Details

### 4.1 Prometheus Configuration

**File:** `/home/mm/dev/b25/prometheus/prometheus.yml`

```yaml
# Prometheus Configuration for B25 Trading System
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'b25-trading'
    environment: 'production'

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

# Load alert rules
rule_files:
  - /etc/prometheus/alerts/*.yml

# Scrape configurations
scrape_configs:
  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          service: 'prometheus'

  # Node Exporter (host metrics)
  - job_name: 'node-exporter'
    scrape_interval: 30s
    static_configs:
      - targets: ['node-exporter:9100']
        labels:
          service: 'node-exporter'
          instance: 'trading-server-01'

  # cAdvisor (container metrics)
  - job_name: 'cadvisor'
    scrape_interval: 30s
    static_configs:
      - targets: ['cadvisor:8080']
        labels:
          service: 'cadvisor'

  # Market Data Service
  - job_name: 'market-data'
    scrape_interval: 5s  # High-priority service
    scrape_timeout: 4s
    static_configs:
      - targets: ['market-data:9090']
        labels:
          service: 'market-data'
          tier: 'critical'

  # Order Execution Service
  - job_name: 'order-execution'
    scrape_interval: 5s  # High-priority service
    scrape_timeout: 4s
    static_configs:
      - targets: ['order-execution:9091']
        labels:
          service: 'order-execution'
          tier: 'critical'

  # Strategy Engine Service
  - job_name: 'strategy-engine'
    scrape_interval: 10s
    static_configs:
      - targets: ['strategy-engine:9092']
        labels:
          service: 'strategy-engine'
          tier: 'core'

  # Account Monitor Service
  - job_name: 'account-monitor'
    scrape_interval: 10s
    static_configs:
      - targets: ['account-monitor:9093']
        labels:
          service: 'account-monitor'
          tier: 'core'

  # Risk Manager Service
  - job_name: 'risk-manager'
    scrape_interval: 10s
    static_configs:
      - targets: ['risk-manager:9094']
        labels:
          service: 'risk-manager'
          tier: 'core'

  # Configuration Service
  - job_name: 'config-service'
    scrape_interval: 30s
    static_configs:
      - targets: ['config-service:9095']
        labels:
          service: 'config-service'
          tier: 'support'

  # Dashboard Server Service
  - job_name: 'dashboard-server'
    scrape_interval: 15s
    static_configs:
      - targets: ['dashboard-server:9096']
        labels:
          service: 'dashboard-server'
          tier: 'support'

  # Blackbox Exporter (endpoint probing)
  - job_name: 'blackbox'
    metrics_path: /probe
    params:
      module: [http_2xx]
    static_configs:
      - targets:
          - http://market-data:8080/health
          - http://order-execution:8080/health
          - http://strategy-engine:8080/health
          - http://account-monitor:8080/health
          - http://risk-manager:8080/health
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: blackbox-exporter:9115

# Storage configuration
storage:
  tsdb:
    path: /prometheus/data
    retention.time: 15d
    retention.size: 50GB
```

### 4.2 Scrape Target Definitions

**Standard Metrics Endpoint per Service:**

Each service MUST implement:
```
GET /metrics
Content-Type: text/plain; version=0.0.4

# Response format (Prometheus exposition format)
# HELP metric_name Description of the metric
# TYPE metric_name counter|gauge|histogram|summary
metric_name{label1="value1",label2="value2"} value timestamp
```

**Metric Naming Conventions:**

```
Pattern: {service}_{subsystem}_{name}_{unit}

Examples:
  - market_data_websocket_latency_seconds
  - order_execution_orders_total
  - strategy_engine_signals_generated_total
  - account_monitor_balance_usd
  - risk_manager_limit_utilization_ratio
```

**Label Standards:**

```yaml
Common Labels (all metrics):
  - service: market-data, order-execution, etc.
  - instance: unique service instance ID
  - environment: dev, staging, prod

Domain-Specific Labels:
  - exchange: binance, bybit, okx
  - symbol: BTCUSDT, ETHUSDT, etc.
  - order_type: limit, market, stop
  - strategy_id: scalper_v1, arbitrage_v2
  - side: buy, sell
  - status: success, failure, pending
```

### 4.3 Key Dashboard Layouts

#### Dashboard 1: System Health Overview

**File:** `/home/mm/dev/b25/grafana/dashboards/01-system-health-overview.json`

**Panels:**

1. **Service Status Grid** (3x3 grid)
   - Each service: Green (up) / Red (down)
   - Query: `up{job=~"market-data|order-execution|..."}`

2. **Request Rate Timeline** (Graph)
   - Requests/sec per service
   - Query: `rate(http_requests_total[1m])`

3. **Error Rate Timeline** (Graph)
   - Errors/sec per service
   - Query: `rate(http_requests_total{status=~"5.."}[1m])`

4. **Latency Heatmap** (Heatmap)
   - P95 latency distribution across services
   - Query: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`

5. **Active Alerts** (Table)
   - Currently firing alerts
   - Query: `ALERTS{alertstate="firing"}`

6. **Resource Utilization** (Gauges)
   - CPU, Memory, Disk usage
   - Query: `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`

#### Dashboard 2: Performance Metrics

**File:** `/home/mm/dev/b25/grafana/dashboards/10-performance-metrics.json`

**Panels:**

1. **Order-to-Fill Latency** (Graph with quantiles)
   - P50, P95, P99, Max
   - Query: `histogram_quantile(0.95, rate(order_execution_latency_ms_bucket[5m]))`

2. **WebSocket Latency** (Graph)
   - Market data feed latency
   - Query: `market_data_websocket_latency_ms{quantile="0.95"}`

3. **Signal-to-Order Latency** (Graph)
   - Strategy decision to order submission
   - Query: `strategy_engine_signal_to_order_latency_ms{quantile="0.99"}`

4. **Throughput Metrics** (Graph)
   - Orders/sec, Fills/sec, Signals/sec
   - Query: `rate(order_execution_orders_total[1m])`

5. **Message Processing Rate** (Graph)
   - Market data messages/sec
   - Query: `rate(market_data_messages_processed_total[1m])`

#### Dashboard 3: Business Metrics (P&L)

**File:** `/home/mm/dev/b25/grafana/dashboards/20-pnl-dashboard.json`

**Panels:**

1. **Cumulative P&L** (Graph)
   - Total realized + unrealized P&L over time
   - Query: `account_monitor_realized_pnl_usd + account_monitor_unrealized_pnl_usd`

2. **Daily P&L** (Bar Chart)
   - P&L per day (last 30 days)
   - Query: `increase(account_monitor_realized_pnl_usd[1d])`

3. **Win Rate** (Gauge)
   - Percentage of profitable trades
   - Query: `(account_monitor_winning_trades_total / account_monitor_total_trades_total) * 100`

4. **Trading Volume** (Graph)
   - USD volume traded per hour
   - Query: `sum(rate(order_execution_volume_usd_total[1h]))`

5. **Position Sizes** (Table)
   - Current positions by symbol
   - Query: `account_monitor_position_size{symbol=~".*"}`

6. **Fee Analysis** (Pie Chart)
   - Maker vs Taker fees
   - Query: `sum(account_monitor_fees_paid_usd{type="maker"})` vs `sum(account_monitor_fees_paid_usd{type="taker"})`

7. **Sharpe Ratio** (Stat)
   - Risk-adjusted return metric
   - Query: `account_monitor_sharpe_ratio`

8. **Max Drawdown** (Gauge)
   - Largest peak-to-trough decline
   - Query: `account_monitor_max_drawdown_pct`

#### Dashboard 4: Market Data Service

**File:** `/home/mm/dev/b25/grafana/dashboards/30-market-data-service.json`

**Panels:**

1. **WebSocket Connection Status** (Stat)
   - Connected exchanges count
   - Query: `market_data_websocket_connected{status="connected"}`

2. **Order Book Depth** (Graph)
   - Bid/ask levels by symbol
   - Query: `market_data_orderbook_depth_levels{side="bid|ask"}`

3. **Message Latency** (Histogram)
   - Exchange → Our system latency
   - Query: `histogram_quantile(0.99, rate(market_data_message_latency_ms_bucket[5m]))`

4. **Reconnection Events** (Graph)
   - WebSocket disconnect/reconnect timeline
   - Query: `rate(market_data_reconnections_total[5m])`

5. **Data Gaps Detected** (Graph)
   - Missing sequence numbers or timestamps
   - Query: `rate(market_data_gaps_detected_total[5m])`

#### Dashboard 5: Order Execution Service

**File:** `/home/mm/dev/b25/grafana/dashboards/31-order-execution-service.json`

**Panels:**

1. **Order Status Breakdown** (Pie Chart)
   - NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED
   - Query: `sum by (status) (order_execution_orders_total)`

2. **Order Submission Latency** (Graph)
   - Time from request to exchange acknowledgment
   - Query: `histogram_quantile(0.95, rate(order_execution_submission_latency_ms_bucket[5m]))`

3. **Fill Rate** (Gauge)
   - Percentage of orders filled vs submitted
   - Query: `(order_execution_filled_orders_total / order_execution_submitted_orders_total) * 100`

4. **Circuit Breaker Status** (Stat)
   - CLOSED (normal), OPEN (tripped), HALF_OPEN (testing)
   - Query: `order_execution_circuit_breaker_state`

5. **Rate Limiter Hits** (Graph)
   - Orders throttled by rate limiter
   - Query: `rate(order_execution_rate_limited_total[5m])`

6. **Maker vs Taker Orders** (Graph)
   - Order type distribution
   - Query: `rate(order_execution_orders_total{type="maker|taker"}[5m])`

### 4.4 Alert Rule Definitions

**File:** `/home/mm/dev/b25/prometheus/alerts/infrastructure.yml`

```yaml
groups:
  - name: infrastructure_alerts
    interval: 30s
    rules:
      # Critical: Service Down
      - alert: ServiceDown
        expr: up{tier="critical"} == 0
        for: 1m
        labels:
          severity: critical
          tier: infrastructure
        annotations:
          summary: "Critical service {{ $labels.service }} is down"
          description: "{{ $labels.service }} on {{ $labels.instance }} has been down for more than 1 minute."
          runbook_url: "https://wiki.b25.com/runbooks/service-down"

      # Critical: High CPU Usage
      - alert: HighCPUUsage
        expr: 100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 90
        for: 5m
        labels:
          severity: warning
          tier: infrastructure
        annotations:
          summary: "High CPU usage on {{ $labels.instance }}"
          description: "CPU usage is {{ $value }}% on {{ $labels.instance }}."

      # Critical: High Memory Usage
      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 90
        for: 5m
        labels:
          severity: warning
          tier: infrastructure
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"
          description: "Memory usage is {{ $value }}% on {{ $labels.instance }}."

      # Critical: Disk Space Low
      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100 < 10
        for: 5m
        labels:
          severity: critical
          tier: infrastructure
        annotations:
          summary: "Disk space critically low on {{ $labels.instance }}"
          description: "Only {{ $value }}% disk space remaining on {{ $labels.instance }}."

      # Warning: High Network Errors
      - alert: HighNetworkErrors
        expr: rate(node_network_receive_errs_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
          tier: infrastructure
        annotations:
          summary: "High network errors on {{ $labels.instance }}"
          description: "Network interface {{ $labels.device }} is experiencing {{ $value }} errors/sec."
```

**File:** `/home/mm/dev/b25/prometheus/alerts/services.yml`

```yaml
groups:
  - name: service_health_alerts
    interval: 15s
    rules:
      # Critical: Service Restart Loop
      - alert: ServiceRestartLoop
        expr: rate(container_last_seen[5m]) > 5
        for: 2m
        labels:
          severity: critical
          tier: service
        annotations:
          summary: "Service {{ $labels.name }} is restarting repeatedly"
          description: "Container {{ $labels.name }} has restarted {{ $value }} times in the last 5 minutes."

      # Critical: High Error Rate
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
          tier: service
        annotations:
          summary: "High error rate on {{ $labels.service }}"
          description: "Error rate is {{ $value | humanizePercentage }} on {{ $labels.service }}."

      # Warning: Slow Response Time
      - alert: SlowResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
          tier: service
        annotations:
          summary: "Slow response time on {{ $labels.service }}"
          description: "P95 latency is {{ $value }}s on {{ $labels.service }}."

      # Critical: WebSocket Disconnected
      - alert: WebSocketDisconnected
        expr: market_data_websocket_connected{exchange=~"binance|bybit"} == 0
        for: 30s
        labels:
          severity: critical
          tier: service
        annotations:
          summary: "WebSocket disconnected from {{ $labels.exchange }}"
          description: "Market data feed from {{ $labels.exchange }} has been disconnected for 30 seconds."

      # Warning: Order Execution Queue Backlog
      - alert: OrderExecutionBacklog
        expr: order_execution_queue_depth > 100
        for: 1m
        labels:
          severity: warning
          tier: service
        annotations:
          summary: "Order execution queue backlog"
          description: "{{ $value }} orders waiting in execution queue."
```

**File:** `/home/mm/dev/b25/prometheus/alerts/business.yml`

```yaml
groups:
  - name: business_alerts
    interval: 15s
    rules:
      # Critical: P&L Drawdown
      - alert: HighDrawdown
        expr: account_monitor_max_drawdown_pct > 5
        for: 1m
        labels:
          severity: critical
          tier: business
        annotations:
          summary: "High drawdown detected"
          description: "Max drawdown is {{ $value }}%, exceeding 5% threshold."
          action: "Consider emergency stop if drawdown continues."

      # Critical: Low Account Balance
      - alert: LowAccountBalance
        expr: account_monitor_balance_usd < 1000
        for: 1m
        labels:
          severity: critical
          tier: business
        annotations:
          summary: "Account balance critically low"
          description: "Account balance is ${{ $value }}, below $1000 minimum."

      # Warning: High Order Rejection Rate
      - alert: HighOrderRejectionRate
        expr: rate(order_execution_orders_total{status="rejected"}[5m]) / rate(order_execution_orders_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          tier: business
        annotations:
          summary: "High order rejection rate"
          description: "{{ $value | humanizePercentage }} of orders are being rejected."

      # Info: Strategy Stopped
      - alert: StrategyInactive
        expr: strategy_engine_active_strategies == 0
        for: 5m
        labels:
          severity: info
          tier: business
        annotations:
          summary: "No active strategies running"
          description: "All trading strategies have been stopped or disabled."

      # Warning: Position Size Limit
      - alert: PositionSizeLimitApproaching
        expr: (abs(account_monitor_position_size) / account_monitor_position_limit) > 0.9
        for: 1m
        labels:
          severity: warning
          tier: business
        annotations:
          summary: "Position size approaching limit for {{ $labels.symbol }}"
          description: "Position size is {{ $value | humanizePercentage }} of limit."

      # Critical: Risk Limit Violation
      - alert: RiskLimitViolation
        expr: risk_manager_limit_violations_total > 0
        for: 30s
        labels:
          severity: critical
          tier: business
        annotations:
          summary: "Risk limit violation detected"
          description: "{{ $value }} risk limit violations in the last 30 seconds."
          action: "Review positions and risk parameters immediately."

      # Warning: Low Win Rate
      - alert: LowWinRate
        expr: (account_monitor_winning_trades_total / account_monitor_total_trades_total) < 0.4
        for: 1h
        labels:
          severity: warning
          tier: business
        annotations:
          summary: "Low win rate detected"
          description: "Win rate is {{ $value | humanizePercentage }}, below 40% threshold."

      # Info: High Maker Fee Ratio
      - alert: LowMakerRatio
        expr: (account_monitor_maker_orders_total / account_monitor_total_orders_total) < 0.7
        for: 1h
        labels:
          severity: info
          tier: business
        annotations:
          summary: "Low maker order ratio"
          description: "Only {{ $value | humanizePercentage }} of orders are maker orders (target: >70%)."
```

### 4.5 Retention Policies

**Prometheus Retention:**
```yaml
Time-Based Retention:
  - Default: 15 days
  - Critical services (market-data, order-execution): 30 days

Size-Based Retention:
  - Max Storage: 50GB
  - Block compression: Enabled (saves ~50% space)

Downsampling (future):
  - Raw data: 15 days
  - 5-minute aggregates: 90 days
  - 1-hour aggregates: 1 year
```

**Loki Retention:**
```yaml
Recent Logs:
  - Retention: 7 days
  - Storage: Filesystem (fast access)

Archive Logs:
  - Retention: 90 days
  - Storage: S3/GCS (compressed, cheaper)

Compaction:
  - Enabled: true
  - Interval: 1h
```

---

## 5. Dashboard Specifications

### 5.1 System Health Overview Dashboard

**Dashboard ID:** 01-system-health-overview
**Refresh Rate:** 15s
**Time Range:** Last 1 hour (default)

**Layout:**

```
┌─────────────────────────────────────────────────────────────┐
│  System Health Overview                    [Last 1h ▼]      │
├─────────────────────────────────────────────────────────────┤
│  Service Status Grid                                        │
│  ┌──────────┬──────────┬──────────┬──────────┐            │
│  │ Market   │ Order    │ Strategy │ Account  │            │
│  │ Data     │ Exec     │ Engine   │ Monitor  │            │
│  │   🟢      │   🟢      │   🟢      │   🟢      │            │
│  ├──────────┼──────────┼──────────┼──────────┤            │
│  │ Risk     │ Config   │ Dashboard│ Metrics  │            │
│  │ Manager  │ Service  │ Server   │ Service  │            │
│  │   🟢      │   🟢      │   🟢      │   🟢      │            │
│  └──────────┴──────────┴──────────┴──────────┘            │
├─────────────────────────────────────────────────────────────┤
│  Request Rate (req/s)              Error Rate (err/s)       │
│  [────────Graph────────]           [────────Graph────────]  │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│  P95 Latency Heatmap                                        │
│  [─────────────────Heatmap─────────────────]                │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│  Active Alerts                 Resource Utilization         │
│  [──────Table──────]           CPU: 45%  MEM: 62%           │
│                                DISK: 38%  NET: 12%           │
└─────────────────────────────────────────────────────────────┘
```

**Panel Queries:**

```promql
# Service Status Grid
up{job=~"market-data|order-execution|strategy-engine|account-monitor|risk-manager|config-service|dashboard-server"}

# Request Rate
sum by (service) (rate(http_requests_total[1m]))

# Error Rate
sum by (service) (rate(http_requests_total{status=~"5.."}[1m]))

# P95 Latency Heatmap
histogram_quantile(0.95, sum by (le, service) (rate(http_request_duration_seconds_bucket[5m])))

# Active Alerts
ALERTS{alertstate="firing"}

# CPU Utilization
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory Utilization
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# Disk Utilization
(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100

# Network Utilization
rate(node_network_receive_bytes_total[1m]) + rate(node_network_transmit_bytes_total[1m])
```

### 5.2 Performance Metrics Dashboard

**Dashboard ID:** 10-performance-metrics
**Refresh Rate:** 10s
**Time Range:** Last 15 minutes (default)

**Key Panels:**

1. **Latency Quantiles Panel** (Graph)
   ```promql
   histogram_quantile(0.50, rate(order_execution_latency_ms_bucket[1m]))  # P50
   histogram_quantile(0.95, rate(order_execution_latency_ms_bucket[1m]))  # P95
   histogram_quantile(0.99, rate(order_execution_latency_ms_bucket[1m]))  # P99
   max(order_execution_latency_ms)  # Max
   ```

2. **Throughput Panel** (Graph)
   ```promql
   sum(rate(order_execution_orders_total[1m]))  # Orders/sec
   sum(rate(market_data_messages_processed_total[1m]))  # Messages/sec
   sum(rate(strategy_engine_signals_generated_total[1m]))  # Signals/sec
   ```

3. **WebSocket Latency Panel** (Graph)
   ```promql
   histogram_quantile(0.95, rate(market_data_websocket_latency_ms_bucket[1m]))
   ```

### 5.3 Business Metrics (P&L) Dashboard

**Dashboard ID:** 20-pnl-dashboard
**Refresh Rate:** 30s
**Time Range:** Last 24 hours (default)

**Key Panels:**

1. **Cumulative P&L** (Graph)
   ```promql
   account_monitor_realized_pnl_usd + account_monitor_unrealized_pnl_usd
   ```

2. **Daily P&L** (Bar Chart)
   ```promql
   increase(account_monitor_realized_pnl_usd[1d])
   ```

3. **Win Rate** (Gauge)
   ```promql
   (account_monitor_winning_trades_total / account_monitor_total_trades_total) * 100
   ```

4. **Trading Volume** (Graph)
   ```promql
   sum(rate(order_execution_volume_usd_total[1h]))
   ```

5. **Position Table** (Table)
   ```promql
   account_monitor_position_size{symbol=~".*"}
   account_monitor_position_entry_price{symbol=~".*"}
   account_monitor_unrealized_pnl_usd{symbol=~".*"}
   ```

6. **Fee Breakdown** (Pie Chart)
   ```promql
   sum(account_monitor_fees_paid_usd{type="maker"})  # Maker fees
   sum(account_monitor_fees_paid_usd{type="taker"})  # Taker fees
   ```

### 5.4 Per-Service Dashboards

Each service gets a dedicated dashboard with service-specific metrics.

**Example: Market Data Service Dashboard**

**Dashboard ID:** 30-market-data-service
**Refresh Rate:** 5s

**Panels:**

1. **WebSocket Status** (Stat)
   ```promql
   market_data_websocket_connected{exchange=~".*"}
   ```

2. **Order Book Depth** (Graph)
   ```promql
   market_data_orderbook_depth_levels{side="bid"}
   market_data_orderbook_depth_levels{side="ask"}
   ```

3. **Message Latency** (Heatmap)
   ```promql
   histogram_quantile(0.99, rate(market_data_message_latency_ms_bucket[5m]))
   ```

4. **Reconnection Events** (Graph)
   ```promql
   rate(market_data_reconnections_total[5m])
   ```

5. **Message Processing Rate** (Graph)
   ```promql
   rate(market_data_messages_processed_total[1m])
   ```

---

## 6. Alert Rules

### 6.1 Alert Severity Levels

| Severity | Response Time | Notification | Use Case |
|----------|--------------|--------------|----------|
| **Critical** | Immediate (< 5 min) | PagerDuty + Slack | Service down, data loss, security breach |
| **Warning** | Soon (< 1 hour) | Slack | Performance degradation, capacity issues |
| **Info** | Awareness only | Log + Slack | Configuration changes, strategy updates |

### 6.2 Critical Alerts (Immediate Response)

```yaml
# Service availability
- ServiceDown (any critical service down for >1min)
- ServiceRestartLoop (>5 restarts in 5min)
- WebSocketDisconnected (market data feed down >30s)

# Data integrity
- OrderBookStale (no updates in 10s)
- PositionMismatch (local vs exchange position differs)

# Financial risk
- HighDrawdown (>5% from peak)
- LowAccountBalance (<$1000)
- RiskLimitViolation (position/leverage/exposure limits)

# Infrastructure
- DiskSpaceLow (<10% free)
- DatabaseDown (PostgreSQL unreachable)
```

### 6.3 Warning Alerts (Investigate Soon)

```yaml
# Performance
- HighCPUUsage (>90% for 5min)
- HighMemoryUsage (>90% for 5min)
- SlowResponseTime (P95 >1s for 5min)

# Business metrics
- HighOrderRejectionRate (>10% of orders)
- LowWinRate (<40% over 1 hour)
- PositionSizeLimitApproaching (>90% of limit)

# Capacity
- HighNetworkErrors (>10 errors/sec)
- OrderExecutionBacklog (>100 orders queued)
```

### 6.4 Info Alerts (Awareness)

```yaml
# Operational changes
- StrategyInactive (no active strategies for 5min)
- ConfigurationChanged (new config deployed)
- LowMakerRatio (<70% maker orders over 1 hour)

# Maintenance
- PrometheusTargetDown (scrape target unavailable)
- LogIngestionSlow (Loki lag >30s)
```

### 6.5 Alert Routing Configuration

**File:** `/home/mm/dev/b25/alertmanager/alertmanager.yml`

```yaml
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'

route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'

  routes:
    # Critical alerts → PagerDuty + Slack
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      continue: true

    - match:
        severity: critical
      receiver: 'slack-critical'

    # Warning alerts → Slack only
    - match:
        severity: warning
      receiver: 'slack-warnings'

    # Info alerts → Slack (low-priority channel)
    - match:
        severity: info
      receiver: 'slack-info'

receivers:
  - name: 'default'
    slack_configs:
      - channel: '#trading-alerts'
        title: 'Alert: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
        description: '{{ .GroupLabels.alertname }}'

  - name: 'slack-critical'
    slack_configs:
      - channel: '#trading-alerts-critical'
        color: 'danger'
        title: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
        text: |
          *Summary:* {{ .CommonAnnotations.summary }}
          *Description:* {{ .CommonAnnotations.description }}
          *Runbook:* {{ .CommonAnnotations.runbook_url }}

  - name: 'slack-warnings'
    slack_configs:
      - channel: '#trading-alerts'
        color: 'warning'
        title: '⚠️  WARNING: {{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.description }}'

  - name: 'slack-info'
    slack_configs:
      - channel: '#trading-info'
        color: 'good'
        title: 'ℹ️  INFO: {{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.description }}'

inhibit_rules:
  # If service is down, suppress all other alerts for that service
  - source_match:
      alertname: 'ServiceDown'
    target_match_re:
      alertname: '.*'
    equal: ['service']

  # If disk is full, suppress high memory alerts (may be swap-related)
  - source_match:
      alertname: 'DiskSpaceLow'
    target_match:
      alertname: 'HighMemoryUsage'
    equal: ['instance']
```

### 6.6 Alert Testing

**Test Alert Command:**
```bash
# Send test alert to Alertmanager
curl -X POST http://localhost:9093/api/v1/alerts \
  -H 'Content-Type: application/json' \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "service": "test"
    },
    "annotations": {
      "summary": "This is a test alert",
      "description": "Testing alert routing and notifications"
    }
  }]'
```

**Verify Alert Rules:**
```bash
# Check Prometheus alert rules
promtool check rules /etc/prometheus/alerts/*.yml

# Test specific alert expression
promtool query instant http://localhost:9090 \
  'up{tier="critical"} == 0'
```

---

## 7. Deployment

### 7.1 Docker Compose Setup

**File:** `/home/mm/dev/b25/docker-compose.metrics.yml`

```yaml
version: '3.8'

services:
  # Prometheus server
  prometheus:
    image: prom/prometheus:v2.48.0
    container_name: b25-prometheus
    restart: unless-stopped
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=15d'
      - '--storage.tsdb.retention.size=50GB'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./prometheus/alerts:/etc/prometheus/alerts:ro
      - prometheus-data:/prometheus
    networks:
      - trading-net
    depends_on:
      - alertmanager

  # Alertmanager
  alertmanager:
    image: prom/alertmanager:v0.26.0
    container_name: b25-alertmanager
    restart: unless-stopped
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
      - alertmanager-data:/alertmanager
    networks:
      - trading-net

  # Grafana
  grafana:
    image: grafana/grafana:10.2.2
    container_name: b25-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-clock-panel
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning:ro
      - ./grafana/dashboards:/var/lib/grafana/dashboards:ro
    networks:
      - trading-net
    depends_on:
      - prometheus
      - loki

  # Loki (log aggregation)
  loki:
    image: grafana/loki:2.9.3
    container_name: b25-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - ./loki/loki-config.yml:/etc/loki/local-config.yaml:ro
      - loki-data:/loki
    networks:
      - trading-net

  # Promtail (log shipper)
  promtail:
    image: grafana/promtail:2.9.3
    container_name: b25-promtail
    restart: unless-stopped
    volumes:
      - ./promtail/promtail-config.yml:/etc/promtail/config.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    command: -config.file=/etc/promtail/config.yml
    networks:
      - trading-net
    depends_on:
      - loki

  # Node Exporter (host metrics)
  node-exporter:
    image: prom/node-exporter:v1.7.0
    container_name: b25-node-exporter
    restart: unless-stopped
    command:
      - '--path.procfs=/host/proc'
      - '--path.sysfs=/host/sys'
      - '--path.rootfs=/rootfs'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    networks:
      - trading-net

  # cAdvisor (container metrics)
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:v0.47.2
    container_name: b25-cadvisor
    restart: unless-stopped
    privileged: true
    ports:
      - "8080:8080"
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:rw
      - /sys:/sys:ro
      - /var/lib/docker:/var/lib/docker:ro
    networks:
      - trading-net

  # Blackbox Exporter (endpoint probing)
  blackbox-exporter:
    image: prom/blackbox-exporter:v0.24.0
    container_name: b25-blackbox-exporter
    restart: unless-stopped
    ports:
      - "9115:9115"
    volumes:
      - ./blackbox/blackbox.yml:/etc/blackbox_exporter/config.yml:ro
    networks:
      - trading-net

volumes:
  prometheus-data:
    driver: local
  alertmanager-data:
    driver: local
  grafana-data:
    driver: local
  loki-data:
    driver: local

networks:
  trading-net:
    external: true
```

### 7.2 Persistent Storage Configuration

**Volume Mounts:**

```yaml
Prometheus Data:
  - Host Path: /var/lib/b25/prometheus
  - Container Path: /prometheus
  - Size: 50GB minimum
  - Backup: Daily snapshots to S3

Grafana Data:
  - Host Path: /var/lib/b25/grafana
  - Container Path: /var/lib/grafana
  - Size: 5GB
  - Backup: Dashboard JSON exports in git

Loki Data:
  - Host Path: /var/lib/b25/loki
  - Container Path: /loki
  - Size: 20GB minimum
  - Backup: Optional S3 backend for long-term retention

Alertmanager Data:
  - Host Path: /var/lib/b25/alertmanager
  - Container Path: /alertmanager
  - Size: 1GB
  - Backup: Not critical (ephemeral alert state)
```

**Backup Strategy:**

```bash
#!/bin/bash
# /home/mm/dev/b25/scripts/backup-metrics.sh

# Snapshot Prometheus data
docker exec b25-prometheus promtool tsdb snapshot /prometheus --output-dir=/prometheus/snapshots

# Backup snapshot to S3
aws s3 sync /var/lib/b25/prometheus/snapshots s3://b25-backups/prometheus/$(date +%Y-%m-%d)/

# Export Grafana dashboards
curl -u admin:${GRAFANA_ADMIN_PASSWORD} \
  http://localhost:3000/api/search?type=dash-db | \
  jq -r '.[] | .uri' | \
  xargs -I {} curl -u admin:${GRAFANA_ADMIN_PASSWORD} \
    http://localhost:3000/api/dashboards/{} > /var/lib/b25/grafana/backups/dashboard-$(date +%Y-%m-%d).json

# Cleanup old snapshots (keep last 7 days)
find /var/lib/b25/prometheus/snapshots -type d -mtime +7 -exec rm -rf {} \;
```

### 7.3 High Availability Considerations

**Prometheus HA Setup (Optional):**

```yaml
# For production: Run 2+ Prometheus instances with same config
# Use federation or Thanos for global querying

services:
  prometheus-1:
    image: prom/prometheus:v2.48.0
    # ... same config as primary

  prometheus-2:
    image: prom/prometheus:v2.48.0
    # ... same config as primary

  # Thanos Query (aggregates multiple Prometheus instances)
  thanos-query:
    image: quay.io/thanos/thanos:v0.32.5
    command:
      - 'query'
      - '--http-address=0.0.0.0:9090'
      - '--store=prometheus-1:10901'
      - '--store=prometheus-2:10901'
```

**Alertmanager HA Setup:**

```yaml
# Alertmanager natively supports clustering
services:
  alertmanager-1:
    image: prom/alertmanager:v0.26.0
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--cluster.peer=alertmanager-2:9094'

  alertmanager-2:
    image: prom/alertmanager:v0.26.0
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--cluster.peer=alertmanager-1:9094'
```

**Grafana HA Setup:**

```yaml
# Use shared database (PostgreSQL) + load balancer
services:
  grafana-db:
    image: postgres:15
    environment:
      POSTGRES_DB: grafana
      POSTGRES_USER: grafana
      POSTGRES_PASSWORD: ${GRAFANA_DB_PASSWORD}

  grafana-1:
    image: grafana/grafana:10.2.2
    environment:
      - GF_DATABASE_TYPE=postgres
      - GF_DATABASE_HOST=grafana-db:5432
      - GF_DATABASE_NAME=grafana
      - GF_DATABASE_USER=grafana
      - GF_DATABASE_PASSWORD=${GRAFANA_DB_PASSWORD}

  grafana-2:
    # ... same config as grafana-1
```

### 7.4 Deployment Commands

**Initial Deployment:**

```bash
# Create trading network (if not exists)
docker network create trading-net

# Create data directories
sudo mkdir -p /var/lib/b25/{prometheus,grafana,loki,alertmanager}
sudo chown -R 65534:65534 /var/lib/b25/prometheus  # nobody user
sudo chown -R 472:472 /var/lib/b25/grafana  # grafana user
sudo chown -R 10001:10001 /var/lib/b25/loki  # loki user

# Start metrics stack
cd /home/mm/dev/b25
docker compose -f docker-compose.metrics.yml up -d

# Verify all services running
docker compose -f docker-compose.metrics.yml ps

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Access UIs
echo "Prometheus: http://localhost:9090"
echo "Grafana: http://localhost:3000 (admin/admin)"
echo "Alertmanager: http://localhost:9093"
```

**Configuration Reload:**

```bash
# Reload Prometheus config (without restart)
curl -X POST http://localhost:9090/-/reload

# Reload Alertmanager config
curl -X POST http://localhost:9093/-/reload

# Restart Grafana (if needed)
docker compose -f docker-compose.metrics.yml restart grafana
```

**Health Checks:**

```bash
# Check Prometheus health
curl http://localhost:9090/-/healthy

# Check Alertmanager health
curl http://localhost:9093/-/healthy

# Check Grafana health
curl http://localhost:3000/api/health

# Check Loki health
curl http://localhost:3100/ready
```

---

## 8. Testing

### 8.1 Metric Collection Verification

**Test 1: Verify All Targets Are Up**

```bash
# Query Prometheus for target status
curl -s http://localhost:9090/api/v1/targets | \
  jq '.data.activeTargets[] | select(.health != "up") | {job: .labels.job, health: .health, error: .lastError}'

# Expected: Empty output (all targets healthy)
```

**Test 2: Verify Metrics Are Being Scraped**

```bash
# Check if specific metric exists
curl -s "http://localhost:9090/api/v1/query?query=up" | jq '.data.result[] | {job: .metric.job, value: .value[1]}'

# Expected: value="1" for all jobs
```

**Test 3: Verify Custom Metrics**

```bash
# Check if service-specific metrics exist
for metric in \
  "market_data_websocket_latency_ms" \
  "order_execution_orders_total" \
  "strategy_engine_signals_generated_total" \
  "account_monitor_balance_usd"; do

  echo "Testing metric: $metric"
  curl -s "http://localhost:9090/api/v1/query?query=$metric" | \
    jq -r ".data.result | length"
done

# Expected: Non-zero count for each metric
```

### 8.2 Alert Rule Testing

**Test 1: Validate Alert Rule Syntax**

```bash
# Check alert rule files for syntax errors
promtool check rules /home/mm/dev/b25/prometheus/alerts/*.yml

# Expected: "SUCCESS: X rules found"
```

**Test 2: Trigger Test Alert**

```bash
# Manually send alert to Alertmanager
curl -X POST http://localhost:9093/api/v1/alerts \
  -H 'Content-Type: application/json' \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "service": "test",
      "instance": "test-instance"
    },
    "annotations": {
      "summary": "Test alert for notification verification",
      "description": "This is a test alert to verify the notification pipeline."
    },
    "startsAt": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
    "endsAt": "'"$(date -u -d '+5 minutes' +%Y-%m-%dT%H:%M:%SZ)"'"
  }]'

# Verify alert appears in Alertmanager UI
curl -s http://localhost:9093/api/v1/alerts | jq '.data[] | select(.labels.alertname == "TestAlert")'

# Expected: Alert visible in Alertmanager, notification sent to Slack
```

**Test 3: Verify Alert Evaluation**

```bash
# Check if Prometheus is evaluating rules
curl -s http://localhost:9090/api/v1/rules | jq '.data.groups[] | {name: .name, rules: .rules | length}'

# Query specific alert state
curl -s http://localhost:9090/api/v1/rules | \
  jq '.data.groups[].rules[] | select(.type == "alerting") | {alert: .name, state: .state}'

# Expected: Rules loaded and evaluating (state: "inactive" or "firing")
```

### 8.3 Dashboard Rendering Tests

**Test 1: Verify Grafana Datasource**

```bash
# Test Prometheus datasource connection
curl -s -u admin:${GRAFANA_ADMIN_PASSWORD} \
  http://localhost:3000/api/datasources | \
  jq '.[] | select(.type == "prometheus") | {name: .name, url: .url}'

# Test datasource query
curl -s -u admin:${GRAFANA_ADMIN_PASSWORD} \
  "http://localhost:3000/api/datasources/proxy/1/api/v1/query?query=up" | \
  jq '.data.result | length'

# Expected: Datasource connected, query returns results
```

**Test 2: Verify Dashboard Provisioning**

```bash
# List all dashboards
curl -s -u admin:${GRAFANA_ADMIN_PASSWORD} \
  http://localhost:3000/api/search?type=dash-db | \
  jq '.[] | {title: .title, uid: .uid}'

# Expected: All dashboards from provisioning directory
```

**Test 3: Test Dashboard Queries**

```bash
# Get dashboard JSON
DASHBOARD_UID="system-health-overview"
curl -s -u admin:${GRAFANA_ADMIN_PASSWORD} \
  "http://localhost:3000/api/dashboards/uid/$DASHBOARD_UID" | \
  jq '.dashboard.panels[] | {title: .title, type: .type}'

# Expected: All panels loaded correctly
```

### 8.4 Log Aggregation Tests

**Test 1: Verify Loki Ingestion**

```bash
# Check if Loki is receiving logs
curl -s "http://localhost:3100/loki/api/v1/labels" | jq '.data[]'

# Expected: Labels like "service_name", "level", etc.
```

**Test 2: Query Logs via LogQL**

```bash
# Query logs for specific service
curl -s -G "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service_name="market-data"}' \
  --data-urlencode "start=$(date -u -d '1 hour ago' +%s)000000000" \
  --data-urlencode "end=$(date -u +%s)000000000" | \
  jq '.data.result | length'

# Expected: Log entries returned
```

**Test 3: Verify Promtail Scraping**

```bash
# Check Promtail targets
curl -s http://localhost:9080/targets | jq '.activeTargets[] | {job: .labels.job, health: .health}'

# Expected: Docker container targets with "up" health
```

### 8.5 Performance Testing

**Test 1: Query Performance**

```bash
# Measure query execution time
time curl -s "http://localhost:9090/api/v1/query?query=rate(http_requests_total[5m])" > /dev/null

# Expected: < 100ms for simple queries
```

**Test 2: Scrape Duration**

```bash
# Check scrape duration per job
curl -s "http://localhost:9090/api/v1/query?query=scrape_duration_seconds" | \
  jq '.data.result[] | {job: .metric.job, duration: .value[1]}'

# Expected: < 1s per scrape (preferably < 100ms)
```

**Test 3: Storage Size**

```bash
# Check TSDB storage size
du -sh /var/lib/b25/prometheus/

# Check number of time series
curl -s "http://localhost:9090/api/v1/query?query=prometheus_tsdb_head_series" | \
  jq '.data.result[0].value[1]'

# Expected: Storage growth within limits, series count < 1M
```

### 8.6 Integration Tests

**Test 1: End-to-End Alert Flow**

```bash
#!/bin/bash
# Test complete alert flow: metric → alert → notification

# 1. Simulate high error rate
for i in {1..100}; do
  curl -s http://localhost:8080/simulate-error > /dev/null
done

# 2. Wait for alert evaluation (15s interval + 2m for duration)
sleep 135

# 3. Check if alert fired
ALERT_STATE=$(curl -s "http://localhost:9090/api/v1/query?query=ALERTS{alertname='HighErrorRate'}" | \
  jq -r '.data.result[0].metric.alertstate')

if [ "$ALERT_STATE" == "firing" ]; then
  echo "✓ Alert fired successfully"
else
  echo "✗ Alert did not fire (state: $ALERT_STATE)"
  exit 1
fi

# 4. Verify notification sent (check Slack/email)
# Manual verification required
```

**Test 2: Dashboard Refresh**

```bash
# Open dashboard in headless browser and verify auto-refresh
# (requires Selenium or Playwright)

# Simple curl test: verify dashboard loads
DASHBOARD_UID="system-health-overview"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  -u admin:${GRAFANA_ADMIN_PASSWORD} \
  "http://localhost:3000/d/$DASHBOARD_UID")

if [ "$HTTP_CODE" == "200" ]; then
  echo "✓ Dashboard loads successfully"
else
  echo "✗ Dashboard failed to load (HTTP $HTTP_CODE)"
  exit 1
fi
```

### 8.7 Automated Test Suite

**File:** `/home/mm/dev/b25/tests/metrics/test_observability.sh`

```bash
#!/bin/bash
set -e

echo "=== Metrics & Observability Test Suite ==="

# Test 1: Service Health
echo -n "Testing service health... "
UNHEALTHY=$(curl -s http://localhost:9090/api/v1/targets | \
  jq '.data.activeTargets[] | select(.health != "up")' | wc -l)
if [ "$UNHEALTHY" -eq 0 ]; then
  echo "✓ All targets healthy"
else
  echo "✗ $UNHEALTHY targets unhealthy"
  exit 1
fi

# Test 2: Metrics Existence
echo -n "Testing metric existence... "
METRICS=(
  "up"
  "market_data_websocket_latency_ms"
  "order_execution_orders_total"
  "account_monitor_balance_usd"
)
for metric in "${METRICS[@]}"; do
  COUNT=$(curl -s "http://localhost:9090/api/v1/query?query=$metric" | \
    jq '.data.result | length')
  if [ "$COUNT" -gt 0 ]; then
    echo -n "."
  else
    echo "✗ Metric $metric not found"
    exit 1
  fi
done
echo " ✓"

# Test 3: Alert Rules Loaded
echo -n "Testing alert rules... "
RULES=$(curl -s http://localhost:9090/api/v1/rules | \
  jq '.data.groups[].rules[] | select(.type == "alerting")' | wc -l)
if [ "$RULES" -gt 0 ]; then
  echo "✓ $RULES alert rules loaded"
else
  echo "✗ No alert rules found"
  exit 1
fi

# Test 4: Grafana Dashboards
echo -n "Testing Grafana dashboards... "
DASHBOARDS=$(curl -s -u admin:${GRAFANA_ADMIN_PASSWORD} \
  http://localhost:3000/api/search?type=dash-db | jq '. | length')
if [ "$DASHBOARDS" -gt 0 ]; then
  echo "✓ $DASHBOARDS dashboards provisioned"
else
  echo "✗ No dashboards found"
  exit 1
fi

# Test 5: Loki Ingestion
echo -n "Testing Loki log ingestion... "
LABELS=$(curl -s "http://localhost:3100/loki/api/v1/labels" | jq '.data | length')
if [ "$LABELS" -gt 0 ]; then
  echo "✓ Loki ingesting logs"
else
  echo "✗ No logs in Loki"
  exit 1
fi

echo "=== All tests passed ✓ ==="
```

---

## 9. Configuration Files

### 9.1 Complete Prometheus Configuration

**File:** `/home/mm/dev/b25/prometheus/prometheus.yml`

(Already provided in Section 4.1)

### 9.2 Alertmanager Configuration

**File:** `/home/mm/dev/b25/alertmanager/alertmanager.yml`

(Already provided in Section 6.5)

### 9.3 Loki Configuration

**File:** `/home/mm/dev/b25/loki/loki-config.yml`

```yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2023-01-01
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/boltdb-shipper-active
    cache_location: /loki/boltdb-shipper-cache
    cache_ttl: 24h
    shared_store: filesystem
  filesystem:
    directory: /loki/chunks

compactor:
  working_directory: /loki/boltdb-shipper-compactor
  shared_store: filesystem

limits_config:
  retention_period: 168h  # 7 days
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 10
  ingestion_burst_size_mb: 20

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: true
  retention_period: 168h
```

### 9.4 Promtail Configuration

**File:** `/home/mm/dev/b25/promtail/promtail-config.yml`

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # Docker container logs
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
    relabel_configs:
      # Extract container name
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container_name'

      # Extract service name (from container label)
      - source_labels: ['__meta_docker_container_label_com_docker_compose_service']
        target_label: 'service_name'

      # Only scrape trading system containers
      - source_labels: ['__meta_docker_container_label_com_docker_compose_project']
        regex: 'b25'
        action: keep

    pipeline_stages:
      # Parse JSON logs
      - json:
          expressions:
            level: level
            message: message
            timestamp: timestamp
            trace_id: trace_id

      # Extract log level
      - labels:
          level:

      # Parse timestamp
      - timestamp:
          source: timestamp
          format: RFC3339Nano

      # Drop debug logs in production
      - match:
          selector: '{level="debug"}'
          action: drop
          drop_counter_reason: debug_logs_dropped
```

### 9.5 Blackbox Exporter Configuration

**File:** `/home/mm/dev/b25/blackbox/blackbox.yml`

```yaml
modules:
  http_2xx:
    prober: http
    timeout: 5s
    http:
      valid_http_versions: ["HTTP/1.1", "HTTP/2.0"]
      valid_status_codes: [200]
      method: GET
      fail_if_ssl: false
      fail_if_not_ssl: false
      preferred_ip_protocol: "ip4"

  http_post_2xx:
    prober: http
    timeout: 5s
    http:
      method: POST
      headers:
        Content-Type: application/json
      body: '{"health": "check"}'
      valid_status_codes: [200, 201]

  tcp_connect:
    prober: tcp
    timeout: 5s

  icmp:
    prober: icmp
    timeout: 5s
```

### 9.6 Grafana Provisioning

**File:** `/home/mm/dev/b25/grafana/provisioning/datasources/datasources.yml`

```yaml
apiVersion: 1

datasources:
  # Prometheus datasource
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
    jsonData:
      timeInterval: "15s"
      queryTimeout: "120s"
      httpMethod: "POST"

  # Loki datasource
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: false
    jsonData:
      maxLines: 1000
```

**File:** `/home/mm/dev/b25/grafana/provisioning/dashboards/dashboards.yml`

```yaml
apiVersion: 1

providers:
  - name: 'B25 Trading Dashboards'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

### 9.7 Grafana Dashboard JSON (Example)

**File:** `/home/mm/dev/b25/grafana/dashboards/01-system-health-overview.json`

```json
{
  "dashboard": {
    "id": null,
    "uid": "system-health-overview",
    "title": "System Health Overview",
    "tags": ["b25", "health", "overview"],
    "timezone": "browser",
    "schemaVersion": 36,
    "version": 1,
    "refresh": "15s",
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "panels": [
      {
        "id": 1,
        "type": "stat",
        "title": "Service Status",
        "gridPos": {"x": 0, "y": 0, "w": 24, "h": 4},
        "targets": [
          {
            "expr": "up{job=~\"market-data|order-execution|strategy-engine|account-monitor|risk-manager|config-service|dashboard-server\"}",
            "refId": "A",
            "legendFormat": "{{job}}"
          }
        ],
        "options": {
          "colorMode": "background",
          "graphMode": "none",
          "reduceOptions": {
            "values": false,
            "calcs": ["lastNotNull"]
          },
          "orientation": "horizontal"
        },
        "fieldConfig": {
          "defaults": {
            "mappings": [
              {"type": "value", "value": "1", "text": "UP", "color": "green"},
              {"type": "value", "value": "0", "text": "DOWN", "color": "red"}
            ],
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {"value": 0, "color": "red"},
                {"value": 1, "color": "green"}
              ]
            }
          }
        }
      },
      {
        "id": 2,
        "type": "graph",
        "title": "Request Rate",
        "gridPos": {"x": 0, "y": 4, "w": 12, "h": 8},
        "targets": [
          {
            "expr": "sum by (service) (rate(http_requests_total[1m]))",
            "refId": "A",
            "legendFormat": "{{service}}"
          }
        ],
        "yaxes": [
          {"label": "req/s", "format": "short"},
          {"show": false}
        ]
      },
      {
        "id": 3,
        "type": "graph",
        "title": "Error Rate",
        "gridPos": {"x": 12, "y": 4, "w": 12, "h": 8},
        "targets": [
          {
            "expr": "sum by (service) (rate(http_requests_total{status=~\"5..\"}[1m]))",
            "refId": "A",
            "legendFormat": "{{service}}"
          }
        ],
        "yaxes": [
          {"label": "err/s", "format": "short"},
          {"show": false}
        ],
        "alert": {
          "name": "High Error Rate Dashboard Alert",
          "conditions": [
            {
              "evaluator": {"type": "gt", "params": [10]},
              "operator": {"type": "and"},
              "query": {"params": ["A", "5m", "now"]},
              "reducer": {"type": "avg"},
              "type": "query"
            }
          ]
        }
      }
    ]
  }
}
```

---

## Summary

This development plan provides a complete blueprint for implementing the **Metrics & Observability Service** for the B25 Trading System. The plan includes:

1. **Technology Stack:** Prometheus + Grafana + Alertmanager + Loki (industry-standard, battle-tested)
2. **6-Phase Implementation:** From basic setup to advanced custom exporters (2-3 weeks)
3. **Complete Configurations:** Ready-to-use Prometheus, Alertmanager, Loki, Grafana configs
4. **Comprehensive Dashboards:** System health, performance, business metrics, per-service dashboards
5. **Intelligent Alerting:** 3-tier severity (critical/warning/info) with proper routing
6. **Production-Ready Deployment:** Docker Compose with HA considerations and backup strategies
7. **Extensive Testing:** Automated test suite covering all aspects of the observability stack

**Key Deliverables:**
- `/home/mm/dev/b25/prometheus/` - Prometheus config and alert rules
- `/home/mm/dev/b25/grafana/` - Dashboard JSONs and provisioning
- `/home/mm/dev/b25/loki/` - Log aggregation config
- `/home/mm/dev/b25/alertmanager/` - Alert routing and notifications
- `/home/mm/dev/b25/docker-compose.metrics.yml` - Complete deployment stack

**Next Steps:**
1. Review and approve technology choices
2. Begin Phase 1 implementation (Prometheus setup)
3. Define service-specific metrics for each microservice
4. Customize alert thresholds based on operational requirements
5. Set up Slack/PagerDuty integrations for notifications

This observability stack will provide complete visibility into the trading system's performance, health, and business metrics, enabling rapid incident response and continuous optimization.
