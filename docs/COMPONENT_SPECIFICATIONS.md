# Component Specifications - Technology-Agnostic

**Purpose:** Detailed functional specifications for each component of the HFT trading system, independent of implementation technology.

**Last Updated:** 2025-10-01
**Version:** 1.0

---

## Table of Contents

1. [Market Data Pipeline](#market-data-pipeline)
2. [Order Execution Engine](#order-execution-engine)
3. [Strategy Engine](#strategy-engine)
4. [Account Monitor](#account-monitor)
5. [Dashboard Server](#dashboard-server)
6. [Risk Manager](#risk-manager)
7. [Configuration Service](#configuration-service)

---

## Market Data Pipeline

### Purpose

Ingest raw market data from cryptocurrency exchanges, normalize it, maintain local order book replicas, and distribute to downstream consumers with minimal latency.

### Functional Requirements

#### FR-MD-001: WebSocket Connection Management
- **Description:** Establish and maintain persistent WebSocket connections to exchange
- **Inputs:** Exchange WebSocket URL, authentication credentials (if needed), symbol list
- **Outputs:** Connection status events
- **Behavior:**
  1. Connect to exchange WebSocket endpoint
  2. Authenticate if required
  3. Subscribe to depth and trade streams for configured symbols
  4. Monitor connection health via heartbeat/ping-pong
  5. Detect disconnection via missing heartbeats or errors
  6. Reconnect automatically with exponential backoff
  7. Resubscribe to all streams after reconnection

#### FR-MD-002: Order Book Maintenance
- **Description:** Maintain real-time local replica of exchange order book
- **Inputs:** Order book snapshot, incremental depth updates
- **Outputs:** Normalized order book state
- **Behavior:**
  1. Fetch initial order book snapshot via REST API or WebSocket
  2. Apply incremental updates to local order book
  3. Validate sequence numbers for gap detection
  4. Discard stale updates (lower sequence than current)
  5. Request new snapshot if gaps detected
  6. Maintain configurable number of price levels (e.g., top 20)
  7. Calculate derived metrics: best bid/ask, mid price, spread

**Order Book Data Structure:**
```
OrderBook {
  symbol: string
  timestamp: integer (milliseconds)
  last_update_id: integer (sequence number)
  bids: List<PriceLevel>  // sorted descending
  asks: List<PriceLevel>  // sorted ascending
}

PriceLevel {
  price: decimal
  quantity: decimal
}
```

#### FR-MD-003: Trade Stream Processing
- **Description:** Process aggregated trade events from exchange
- **Inputs:** Trade event messages
- **Outputs:** Normalized trade records
- **Behavior:**
  1. Parse trade event message
  2. Extract: symbol, price, quantity, timestamp, buyer_is_maker
  3. Validate data integrity
  4. Calculate trade flow metrics (net buy/sell volume)
  5. Publish to consumers

**Trade Data Structure:**
```
Trade {
  symbol: string
  trade_id: integer
  price: decimal
  quantity: decimal
  timestamp: integer (milliseconds)
  buyer_is_maker: boolean
}
```

#### FR-MD-004: Data Normalization
- **Description:** Convert exchange-specific formats to internal standard
- **Inputs:** Raw exchange messages
- **Outputs:** Normalized data structures
- **Behavior:**
  1. Parse exchange-specific JSON/binary format
  2. Map fields to internal schema
  3. Convert data types (string to decimal, etc.)
  4. Apply precision and scaling
  5. Add metadata (source, received timestamp)

#### FR-MD-005: Data Distribution
- **Description:** Publish normalized data to consumers efficiently
- **Inputs:** Normalized order book and trade data
- **Outputs:** Serialized messages to shared memory, message queue, or database
- **Behavior:**
  1. Serialize data to efficient format (binary preferred)
  2. Write to shared memory ring buffer (local consumers)
  3. Publish to message queue topic (distributed consumers)
  4. Insert into time-series database (historical storage)
  5. Rate limit database writes to avoid overload

### Non-Functional Requirements

#### NFR-MD-001: Latency
- **Target:** <100μs from WebSocket message receipt to publication
- **Measurement:** Instrument with high-resolution timestamps
- **Optimization:** Zero-copy buffers, lock-free data structures, minimal allocations

#### NFR-MD-002: Throughput
- **Target:** 10,000 messages/second per symbol
- **Measurement:** Count messages processed per second
- **Optimization:** Batch processing, efficient serialization

#### NFR-MD-003: Reliability
- **Target:** 99.99% uptime during market hours
- **Measurement:** Track connection uptime, downtime incidents
- **Mechanism:** Automatic reconnection, health monitoring, alerting

#### NFR-MD-004: Data Accuracy
- **Target:** Zero data loss, zero incorrect order books
- **Measurement:** Periodic reconciliation with exchange REST API
- **Mechanism:** Sequence number validation, snapshot resynchronization

### Error Handling

| Error Condition | Response |
|-----------------|----------|
| WebSocket disconnection | Log error, attempt reconnection with exponential backoff, alert if >30s |
| Sequence number gap | Log warning, request new snapshot, resume processing |
| Malformed message | Log error, skip message, increment error counter |
| Rate limit exceeded | Back off, reduce subscription frequency, alert operator |
| Authentication failure | Log critical error, halt pipeline, alert operator immediately |

### Configuration Parameters

```
configuration MarketDataPipeline {
  symbols: List<String>                // Trading pairs to subscribe
  exchange_ws_url: String              // WebSocket endpoint
  exchange_api_url: String             // REST API endpoint
  api_key: String (optional)           // API key for authentication
  api_secret: String (optional)        // API secret for authentication

  order_book_depth: Integer            // Number of price levels (default: 20)
  ring_buffer_size: Integer            // Shared memory size (default: 8192)
  enable_database_writes: Boolean      // Write to time-series DB (default: true)
  database_write_interval_ms: Integer  // Batch write interval (default: 100)

  reconnect_max_attempts: Integer      // Max reconnection attempts (default: 10)
  reconnect_base_delay_ms: Integer     // Base delay (default: 1000)
  reconnect_max_delay_ms: Integer      // Max delay (default: 60000)

  heartbeat_interval_ms: Integer       // Heartbeat check interval (default: 5000)
  heartbeat_timeout_ms: Integer        // Connection timeout (default: 30000)

  metrics_port: Integer                // Prometheus metrics port (default: 9090)
  log_level: String                    // Logging level (default: "info")
}
```

### Metrics Exported

- `websocket_connected{symbol}` (gauge, 0/1)
- `messages_received_total{symbol,type}` (counter)
- `messages_processed_total{symbol,type}` (counter)
- `messages_dropped_total{symbol,reason}` (counter)
- `processing_latency_microseconds{symbol}` (histogram)
- `order_book_updates_total{symbol}` (counter)
- `trades_processed_total{symbol}` (counter)
- `reconnection_attempts_total{symbol}` (counter)
- `errors_total{symbol,type}` (counter)

### API / Interfaces

**Output: Shared Memory**
```
SharedMemoryWriter.write(message: MarketDataMessage)

where MarketDataMessage is one of:
  - OrderBookSnapshot
  - OrderBookUpdate
  - Trade
```

**Output: Message Queue**
```
Publisher.publish(topic: String, message: Bytes)

Topics:
  - "market_data.{symbol}.orderbook"
  - "market_data.{symbol}.trades"
```

**Output: Time-Series Database**
```
Database.batch_insert(table: String, records: List<Record>)

Tables:
  - market_data_snapshots
  - trades
```

**Health Check Endpoint**
```
GET /health
Response:
  status: "healthy" | "degraded" | "unhealthy"
  connections: {
    "BTCUSDT": "connected",
    "ETHUSDT": "connected"
  }
  uptime_seconds: 12345
  errors_last_minute: 2
```

---

## Order Execution Engine

### Purpose

Receive order requests from strategy engine, validate them, submit to exchange, track order lifecycle, and manage fills.

### Functional Requirements

#### FR-OE-001: Order Request Handling
- **Description:** Accept order requests from strategy engine via RPC
- **Inputs:** Order request message
- **Outputs:** Order acknowledgment or rejection
- **Behavior:**
  1. Receive order request via request-reply protocol
  2. Assign unique client order ID
  3. Validate request format
  4. Queue for processing
  5. Return acknowledgment with request ID and order ID

**Order Request Structure:**
```
OrderRequest {
  request_id: string (UUID)
  symbol: string
  side: "buy" | "sell"
  type: "limit" | "market" | "stop_limit" | "take_profit_limit"
  time_in_force: "GTC" | "IOC" | "FOK" | "POST_ONLY"
  price: decimal (optional, required for limit orders)
  quantity: decimal
  stop_price: decimal (optional, for stop orders)
  execution_mode: "live" | "simulation" | "observation"
}
```

#### FR-OE-002: Order Validation
- **Description:** Pre-flight validation before submission to exchange
- **Inputs:** Order request
- **Outputs:** Validation result (pass/fail with reason)
- **Validation Rules:**
  1. **Notional Value:** `price * quantity >= min_notional` (e.g., 5 USDT for Binance)
  2. **Price Precision:** Price matches symbol precision (e.g., 2 decimals for BTCUSDT)
  3. **Quantity Precision:** Quantity matches symbol precision
  4. **Price Bounds:** Price within reasonable range (e.g., ±10% from mark price)
  5. **Position Limits:** Adding position won't exceed max position size
  6. **Available Margin:** Sufficient margin for order
  7. **Rate Limits:** Not exceeding order rate limits
  8. **Circuit Breaker:** Circuit breaker not open

**Validation Response:**
```
ValidationResult {
  valid: boolean
  order_id: string (if valid)
  error_code: string (if invalid)
  error_message: string (if invalid)
}
```

#### FR-OE-003: Order Submission
- **Description:** Submit validated orders to exchange REST API
- **Inputs:** Validated order
- **Outputs:** Exchange order ID and status
- **Behavior:**
  1. Construct exchange-specific order payload
  2. Sign request with HMAC-SHA256 (or exchange-required method)
  3. Submit via HTTPS POST
  4. Parse exchange response
  5. Update local order state
  6. Handle exchange errors

**Exchange API Request (Example for Binance):**
```
POST /fapi/v1/order
Headers:
  X-MBX-APIKEY: {api_key}
Body:
  symbol: "BTCUSDT"
  side: "BUY"
  type: "LIMIT"
  timeInForce: "GTC"
  quantity: "0.001"
  price: "67250.00"
  recvWindow: 5000
  timestamp: 1234567890123
  signature: {HMAC-SHA256 signature}
```

#### FR-OE-004: Order State Management
- **Description:** Track order through its complete lifecycle
- **States:**
  - NEW: Just created, not yet submitted
  - PENDING_SUBMIT: Queued for submission
  - SUBMITTED: Sent to exchange, awaiting acknowledgment
  - PARTIALLY_FILLED: Some quantity filled
  - FILLED: Completely filled
  - CANCELED: Canceled by user or system
  - REJECTED: Rejected by exchange
  - EXPIRED: Expired due to time-in-force

**State Transitions:**
```
NEW → PENDING_SUBMIT → SUBMITTED → PARTIALLY_FILLED → FILLED
                                 → CANCELED
                                 → REJECTED
                                 → EXPIRED
```

#### FR-OE-005: Fill Processing
- **Description:** Process order fill notifications from exchange
- **Inputs:** Fill event from user data stream WebSocket
- **Outputs:** Fill notification to subscribers, P&L calculation input
- **Behavior:**
  1. Receive fill event from WebSocket
  2. Match to local order by order ID
  3. Update order state (PARTIALLY_FILLED or FILLED)
  4. Calculate fill P&L if closing position
  5. Publish fill event to subscribers
  6. Update metrics

**Fill Event Structure:**
```
FillEvent {
  order_id: integer
  client_order_id: string
  symbol: string
  side: "buy" | "sell"
  price: decimal
  quantity: decimal
  commission: decimal
  commission_asset: string
  is_maker: boolean
  timestamp: integer
}
```

#### FR-OE-006: Order Cancellation
- **Description:** Cancel active orders on request
- **Inputs:** Cancel request with order ID or symbol
- **Outputs:** Cancellation confirmation
- **Behavior:**
  1. Receive cancel request
  2. Validate order exists and is cancelable
  3. Submit cancel request to exchange
  4. Update local order state to CANCELED
  5. Publish cancellation event

#### FR-OE-007: Execution Modes
- **Description:** Support different execution modes for testing and operation
- **Modes:**
  - **Live:** Submit orders to real exchange with real money
  - **Simulation:** Generate order IDs locally, simulate fills, don't submit to exchange
  - **Observation:** Log signals only, don't generate orders

### Non-Functional Requirements

#### NFR-OE-001: Latency
- **Target:** <10ms from order request to exchange submission
- **Measurement:** Instrument with timestamps
- **Optimization:** Connection pooling, pre-computed signatures where possible

#### NFR-OE-002: Reliability
- **Target:** 99.9% successful order submissions (excluding exchange rejections)
- **Measurement:** Track submission success rate
- **Mechanism:** Retry with exponential backoff (max 3 attempts), circuit breaker

#### NFR-OE-003: Idempotency
- **Target:** Duplicate requests don't create duplicate orders
- **Mechanism:** Use client_order_id as idempotency key, cache recent submissions

### Error Handling

| Error Condition | Response |
|-----------------|----------|
| Exchange rate limit (429) | Back off exponentially, retry up to 3 times, alert if persistent |
| Exchange unavailable (503) | Retry after delay, open circuit breaker if repeated |
| Authentication failure (401) | Halt trading immediately, open circuit breaker, alert operator |
| Invalid order (4xx) | Log error, return rejection to strategy, don't retry |
| Network timeout | Retry up to 3 times, then mark as failed, alert if frequent |
| Order not found (cancel) | Log warning, assume already canceled, continue |

### Configuration Parameters

```
configuration OrderExecutionEngine {
  binance_api_url: String              // REST API base URL
  binance_api_key: String              // API key
  binance_api_secret: String           // API secret

  execution_mode: String               // "live" | "simulation" | "observation"

  zeromq_request_address: String       // ZeroMQ REP socket address
  zeromq_publish_address: String       // ZeroMQ PUB socket address

  max_orders_per_second: Integer       // Rate limit (default: 10)
  request_timeout_ms: Integer          // HTTP request timeout (default: 5000)
  max_retry_attempts: Integer          // Max retries (default: 3)

  circuit_breaker_threshold: Integer   // Failures to open CB (default: 5)
  circuit_breaker_timeout_ms: Integer  // CB timeout (default: 30000)

  connection_pool_size: Integer        // HTTP connections (default: 20)

  metrics_port: Integer                // Prometheus port (default: 9091)
  log_level: String                    // Log level (default: "info")
}
```

### Metrics Exported

- `order_requests_total{result}` (counter, result=success|rejected|failed)
- `orders_submitted_total{symbol,side}` (counter)
- `orders_filled_total{symbol,side}` (counter)
- `orders_canceled_total{symbol}` (counter)
- `order_submit_latency_milliseconds{symbol}` (histogram)
- `circuit_breaker_state{service}` (gauge, 0=closed, 1=open, 2=half_open)
- `maker_fills_total{symbol}` (counter)
- `taker_fills_total{symbol}` (counter)
- `errors_total{type}` (counter)

### API / Interfaces

**Input: Order Request (RPC)**
```
Request:
  OrderRequest (as defined above)

Response:
  OrderResponse {
    request_id: string
    status: "accepted" | "rejected"
    order_id: string (if accepted)
    error: string (if rejected)
  }
```

**Output: Fill Events (Pub/Sub)**
```
Topic: "fills.{symbol}"
Payload: FillEvent (as defined above)
```

**Output: Order Updates (Pub/Sub)**
```
Topic: "orders.{symbol}"
Payload: OrderUpdate {
  order_id: string
  status: OrderState
  filled_quantity: decimal
  remaining_quantity: decimal
  average_price: decimal
}
```

---

## Strategy Engine

### Purpose

Execute trading strategies, generate signals, aggregate multi-strategy signals, apply risk filters, and send order requests to execution engine.

### Functional Requirements

#### FR-SE-001: Strategy Plugin System
- **Description:** Load and manage multiple trading strategies dynamically
- **Inputs:** Strategy configuration files
- **Outputs:** Loaded strategy instances
- **Behavior:**
  1. Scan strategy directory or registry
  2. Load strategy plugins (scripts, compiled libraries, etc.)
  3. Initialize each strategy with its configuration
  4. Maintain registry of active strategies
  5. Support hot-reload (reload without restart)

**Strategy Interface (Generic):**
```
interface Strategy {
  name(): string

  initialize(config: Configuration): void

  on_market_data(data: MarketData): List<Signal>

  on_order_fill(fill: FillEvent): void

  on_position_update(position: Position): void

  get_state(): StrategyState
}
```

#### FR-SE-002: Market Data Consumption
- **Description:** Subscribe to market data and feed to strategies
- **Inputs:** Market data from market data pipeline
- **Outputs:** Strategy invocations
- **Behavior:**
  1. Connect to market data source (shared memory, message queue)
  2. Deserialize market data messages
  3. Route to appropriate strategies based on symbol
  4. Invoke strategy's `on_market_data()` method
  5. Collect returned signals

#### FR-SE-003: Signal Generation
- **Description:** Strategies generate trading signals based on market data
- **Signal Structure:**
```
Signal {
  symbol: string
  side: "buy" | "sell"
  signal_strength: decimal (0.0 to 1.0)
  price_hint: decimal (optional, suggested limit price)
  reason: string (explanation)
  urgency: "passive" | "normal" | "aggressive"
  strategy_name: string
  timestamp: integer
}
```

**Urgency Levels:**
- **Passive:** POST_ONLY orders, willing to wait for maker fill
- **Normal:** POST_ONLY with competitive pricing, prefer maker
- **Aggressive:** Allow taker execution if maker fails

#### FR-SE-004: Signal Aggregation
- **Description:** Combine signals from multiple strategies
- **Inputs:** Signals from all active strategies
- **Outputs:** Aggregated signal or decision
- **Aggregation Methods:**
  1. **Majority Vote:** Execute if >50% of strategies agree
  2. **Weighted Average:** Weight by strategy performance
  3. **First Signal:** Execute first signal above threshold
  4. **Ensemble:** ML model combines signals

**Aggregated Signal:**
```
AggregatedSignal {
  symbol: string
  side: "buy" | "sell"
  combined_strength: decimal
  contributing_strategies: List<string>
  final_price_hint: decimal
  urgency: "passive" | "normal" | "aggressive"
}
```

#### FR-SE-005: Risk Filtering
- **Description:** Apply risk management rules before sending orders
- **Inputs:** Aggregated signals, current positions, risk limits
- **Outputs:** Filtered signals (approved or rejected)
- **Risk Checks:**
  1. Position size limit: Won't exceed max position for symbol
  2. Total exposure limit: Won't exceed portfolio-level max
  3. Order frequency limit: Not exceeding max orders per second
  4. Daily P&L limit: Not past daily loss limit
  5. Drawdown limit: Not in max drawdown state

**Risk Filter Response:**
```
RiskFilterResult {
  approved: boolean
  signal: AggregatedSignal (if approved)
  rejection_reason: string (if rejected)
}
```

#### FR-SE-006: Order Generation
- **Description:** Convert approved signals to order requests
- **Inputs:** Approved signals
- **Outputs:** Order requests to execution engine
- **Behavior:**
  1. Determine order type based on signal urgency
  2. Calculate order quantity based on signal strength and position sizing rules
  3. Set time-in-force (POST_ONLY for passive/normal)
  4. Set limit price from price_hint (or best bid/ask)
  5. Construct OrderRequest message
  6. Send to execution engine via RPC
  7. Handle execution engine response

#### FR-SE-007: Position Tracking
- **Description:** Maintain current position state for each symbol
- **Inputs:** Fill events from execution engine
- **Outputs:** Updated position state
- **Behavior:**
  1. Subscribe to fill events
  2. Update position on each fill
  3. Calculate average entry price
  4. Track unrealized P&L
  5. Notify strategies of position updates

**Position Structure:**
```
Position {
  symbol: string
  side: "long" | "short" | "flat"
  size: decimal
  entry_price: decimal
  current_price: decimal
  unrealized_pnl: decimal
  unrealized_pnl_percent: decimal
}
```

### Non-Functional Requirements

#### NFR-SE-001: Latency
- **Target:** <500μs from market data to signal generation
- **Measurement:** Instrument strategy execution time
- **Optimization:** Minimize allocations, use efficient data structures

#### NFR-SE-002: Extensibility
- **Target:** Add new strategies without modifying core engine
- **Mechanism:** Plugin interface, dynamic loading, hot-reload

#### NFR-SE-003: Isolation
- **Target:** Strategy exceptions don't crash engine or affect other strategies
- **Mechanism:** Catch and log exceptions, disable faulty strategy

### Error Handling

| Error Condition | Response |
|-----------------|----------|
| Strategy exception | Log error, skip signal generation for that strategy, increment error counter, disable if repeated |
| Market data unavailable | Use last known data, log warning, alert if prolonged |
| Execution engine unavailable | Queue orders if possible, alert operator, enter observation mode |
| Risk filter rejection | Log info, don't send order, publish rejection event |

### Configuration Parameters

```
configuration StrategyEngine {
  strategies: List<StrategyConfig>     // Strategy configurations

  market_data_source: String           // "shared_memory" | "zeromq" | "redis"
  market_data_address: String          // Connection string

  execution_engine_address: String     // RPC address for order requests
  account_monitor_address: String      // gRPC address for account queries

  signal_aggregation_method: String    // "majority" | "weighted" | "first" | "ensemble"

  risk_limits: RiskLimits              // Global risk limits

  enable_hot_reload: Boolean           // Allow strategy hot-reload (default: false)

  metrics_port: Integer                // Prometheus port (default: 9092)
  log_level: String                    // Log level (default: "info")
}

StrategyConfig {
  name: string
  type: string                         // Strategy type (e.g., "momentum")
  enabled: boolean
  symbol: string
  order_size: decimal
  parameters: Map<string, any>         // Strategy-specific params
}

RiskLimits {
  max_position_size_usdt: decimal
  max_positions_per_symbol: integer
  max_daily_drawdown_pct: decimal
  max_leverage: decimal
  max_orders_per_second: integer
}
```

### Metrics Exported

- `strategies_active{name}` (gauge)
- `signals_generated_total{strategy,symbol,side}` (counter)
- `signals_aggregated_total{symbol}` (counter)
- `signals_rejected_by_risk{reason}` (counter)
- `orders_generated_total{strategy,symbol}` (counter)
- `strategy_execution_latency_microseconds{strategy}` (histogram)
- `strategy_errors_total{strategy,type}` (counter)

---

## Account Monitor

### Purpose

Track account balances, positions, profit/loss, reconcile with exchange, and enforce risk limits.

### Functional Requirements

#### FR-AM-001: Balance Tracking
- **Description:** Monitor real-time account balance
- **Inputs:** Balance updates from exchange user data stream, fill events
- **Outputs:** Current balance state
- **Behavior:**
  1. Query initial balance via REST API
  2. Subscribe to balance update events
  3. Update balance on fills and funding events
  4. Persist balance snapshots to database
  5. Provide query interface for current balance

**Balance Structure:**
```
AccountBalance {
  asset: string (e.g., "USDT")
  total_balance: decimal
  available_balance: decimal
  in_orders: decimal
  timestamp: integer
}
```

#### FR-AM-002: Position Tracking
- **Description:** Track open positions for all symbols
- **Inputs:** Fill events, position updates from exchange
- **Outputs:** Current positions
- **Behavior:**
  1. Initialize position state from exchange REST API
  2. Update positions on fill events
  3. Calculate average entry price on position increases
  4. Mark position as closed when size reaches zero
  5. Track position P&L

**Position Structure (Detailed):**
```
Position {
  symbol: string
  side: "long" | "short"
  size: decimal
  entry_price: decimal
  mark_price: decimal
  liquidation_price: decimal
  unrealized_pnl: decimal
  unrealized_pnl_percent: decimal
  leverage: decimal
  margin_used: decimal
  timestamp: integer
}
```

#### FR-AM-003: P&L Calculation
- **Description:** Calculate realized and unrealized profit/loss
- **Inputs:** Fill events, current mark prices, positions
- **Outputs:** P&L metrics
- **Calculations:**
  - **Realized P&L:** Sum of P&L from closed positions
  - **Unrealized P&L:** Mark-to-market of open positions
  - **Total P&L:** Realized + Unrealized
  - **Daily P&L:** P&L since midnight UTC
  - **Fees Paid:** Sum of all trading fees

**P&L Structure:**
```
PnLSummary {
  realized_pnl: decimal
  unrealized_pnl: decimal
  total_pnl: decimal
  daily_pnl: decimal
  fees_paid: decimal
  win_rate: decimal (percentage)
  sharpe_ratio: decimal (optional)
  max_drawdown: decimal
  max_drawdown_percent: decimal
}
```

#### FR-AM-004: Position Reconciliation
- **Description:** Ensure local positions match exchange state
- **Inputs:** Local positions, exchange positions via REST API
- **Outputs:** Reconciliation report, correction events
- **Behavior:**
  1. Query exchange positions periodically (e.g., every 5 seconds)
  2. Compare with local position state
  3. If mismatch detected:
     - Log discrepancy with details
     - Use exchange as source of truth
     - Update local state
     - Alert operator
  4. If mismatch exceeds threshold (e.g., >5%):
     - Log critical alert
     - Halt trading
     - Require manual intervention

**Reconciliation Report:**
```
ReconciliationReport {
  timestamp: integer
  positions_checked: integer
  mismatches_found: integer
  discrepancies: List<Discrepancy>
}

Discrepancy {
  symbol: string
  local_size: decimal
  exchange_size: decimal
  difference: decimal
  difference_percent: decimal
}
```

#### FR-AM-005: Risk Threshold Monitoring
- **Description:** Monitor risk metrics and generate alerts
- **Inputs:** Current positions, balance, P&L, risk limits
- **Outputs:** Alert events
- **Monitored Metrics:**
  - Daily drawdown vs. max allowed
  - Margin ratio vs. minimum required
  - Position size vs. max allowed per symbol
  - Total exposure vs. max allowed
  - Liquidation price proximity

**Alert Structure:**
```
Alert {
  id: string
  severity: "info" | "warning" | "error" | "critical"
  category: "position" | "balance" | "pnl" | "risk" | "system"
  message: string
  timestamp: integer
  acknowledged: boolean
}
```

#### FR-AM-006: Query API
- **Description:** Provide interface for other components to query account state
- **Endpoints:**
  - `GetAccountState()` → Full account snapshot
  - `GetBalance()` → Current balance
  - `GetPosition(symbol)` → Position for specific symbol
  - `GetPositions()` → All open positions
  - `GetPnL()` → Current P&L summary
  - `SubscribeAccountUpdates()` → Streaming updates

### Non-Functional Requirements

#### NFR-AM-001: Accuracy
- **Target:** 100% reconciliation with exchange (zero undetected mismatches)
- **Measurement:** Count reconciliation mismatches, track detection time
- **Mechanism:** Frequent reconciliation, multiple validation points

#### NFR-AM-002: Latency (Non-Critical)
- **Target:** <100ms for query responses
- **Note:** Not on critical trading path, latency less stringent

### Configuration Parameters

```
configuration AccountMonitor {
  binance_api_url: String
  binance_api_key: String
  binance_api_secret: String

  grpc_server_address: String          // gRPC server listen address

  reconciliation_interval_seconds: Integer   // Default: 5

  alert_thresholds: AlertThresholds

  metrics_port: Integer                // Prometheus port
  log_level: String
}

AlertThresholds {
  max_daily_drawdown_pct: decimal      // E.g., 5.0
  min_margin_ratio_pct: decimal        // E.g., 20.0
  max_position_size_usdt: decimal      // E.g., 50000
  liquidation_proximity_pct: decimal   // E.g., 5.0 (alert if liq price within 5%)
}
```

### Metrics Exported

- `account_balance_usdt` (gauge)
- `available_balance_usdt` (gauge)
- `total_pnl_usdt` (gauge)
- `realized_pnl_usdt` (gauge)
- `unrealized_pnl_usdt` (gauge)
- `daily_pnl_usdt` (gauge)
- `position_count` (gauge)
- `position_value_usdt{symbol}` (gauge)
- `margin_ratio` (gauge)
- `reconciliation_mismatches_total` (counter)
- `alerts_generated_total{severity,category}` (counter)

---

## Dashboard Server

### Purpose

Aggregate system state from all components and broadcast to dashboard clients (TUI and web) via WebSocket.

### Functional Requirements

#### FR-DS-001: State Aggregation
- **Description:** Collect state from all system components
- **Inputs:**
  - Market data from market data pipeline
  - Order state from execution engine
  - Account state from account monitor
  - System metrics from all services
- **Outputs:** Unified dashboard state
- **Behavior:**
  1. Query or subscribe to each data source
  2. Merge into unified state structure
  3. Cache aggregated state in memory
  4. Update on change events

#### FR-DS-002: WebSocket Server
- **Description:** Manage WebSocket connections to dashboard clients
- **Inputs:** Client connection requests
- **Outputs:** Accepted connections
- **Behavior:**
  1. Listen for WebSocket connections
  2. Authenticate clients (if required)
  3. Assign client type (TUI or Web)
  4. Add to connection pool
  5. Send initial full state snapshot
  6. Handle client disconnections

#### FR-DS-003: Update Broadcasting
- **Description:** Send periodic state updates to connected clients
- **Inputs:** State changes, update timer
- **Outputs:** Serialized state messages via WebSocket
- **Behavior:**
  1. Maintain update timer per client type
    - TUI: 100ms (10 updates/sec)
    - Web: 250ms (4 updates/sec)
  2. Serialize current state efficiently
  3. Broadcast to all connected clients of that type
  4. Track send failures, disconnect dead clients

#### FR-DS-004: Message Serialization
- **Description:** Efficient serialization of state messages
- **Format:** Binary (e.g., MessagePack, Protocol Buffers) preferred for efficiency
- **Messages:**
  - Full state snapshot (on connection)
  - Incremental updates (periodic)
  - Event notifications (on-demand)
  - Heartbeat (keep-alive)

**Message Types:**
```
MessageType:
  FULL_STATE
  POSITIONS_UPDATE
  ORDERBOOK_UPDATE
  ORDER_UPDATE
  PNL_UPDATE
  HEALTH_UPDATE
  AI_SIGNAL
  ALERT
  HEARTBEAT
```

#### FR-DS-005: Client Subscriptions
- **Description:** Allow clients to subscribe to specific data types
- **Inputs:** Subscription request from client
- **Outputs:** Filtered updates matching subscription
- **Behavior:**
  1. Receive subscription message
  2. Update client subscription filter
  3. Only send matching updates to that client

### Configuration Parameters

```
configuration DashboardServer {
  websocket_listen_address: String     // E.g., "0.0.0.0:8080"

  tui_update_rate_ms: Integer          // Default: 100
  web_update_rate_ms: Integer          // Default: 250

  max_clients: Integer                 // Default: 100
  heartbeat_interval_ms: Integer       // Default: 5000

  enable_authentication: Boolean       // Default: false

  metrics_port: Integer
  log_level: String
}
```

### Metrics Exported

- `websocket_clients_connected{type}` (gauge, type=tui|web)
- `messages_sent_total{type}` (counter)
- `message_send_errors_total` (counter)
- `broadcast_latency_milliseconds` (histogram)

---

**End of Component Specifications**

This document provides detailed functional and non-functional specifications for each major component. Use these specifications to guide implementation in your chosen technology stack.
