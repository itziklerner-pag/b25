# Account Monitor Service - Development Plan

**Service ID:** 04
**Purpose:** Real-time account state tracking, position management, P&L calculation, and reconciliation
**Last Updated:** 2025-10-02
**Version:** 1.0

---

## 1. Technology Stack Recommendation

### Core Language & Framework
**Recommended: Go (Primary) or Python (Alternative)**

**Go Advantages:**
- Excellent concurrency model (goroutines) for handling multiple WebSocket streams
- Built-in race detection for concurrent balance updates
- Fast P&L calculations with compiled performance
- Strong gRPC ecosystem
- Efficient memory usage for position tracking

**Python Advantages:**
- Rich numerical libraries (numpy, pandas) for P&L analytics
- Faster development for complex financial calculations
- Easier backtesting and validation of formulas

**Decision Matrix:**
- **Go:** For production performance and reliability
- **Python:** For rapid prototyping and complex analytics

### Technology Stack (Go-based)

```yaml
Language: Go 1.21+
Web Framework: None (standard lib sufficient for gRPC/HTTP)
gRPC Framework: google.golang.org/grpc
WebSocket Client: gorilla/websocket
Database:
  - Time-series: TimescaleDB (PostgreSQL extension)
  - Cache: Redis (position snapshots)
Database Client:
  - pgx (PostgreSQL driver)
  - go-redis (Redis client)
Pub/Sub: Redis Pub/Sub or NATS
Testing:
  - testing (stdlib)
  - testify (assertions)
  - gomock (mocking)
  - go-sqlmock (DB mocking)
Observability:
  - prometheus/client_golang (metrics)
  - zerolog (structured logging)
  - opentelemetry-go (tracing)
Configuration: viper
Hot Reload: fsnotify
```

### Alternative Stack (Python-based)

```yaml
Language: Python 3.11+
Framework: FastAPI (HTTP) + grpcio
WebSocket Client: websockets or python-binance
Database:
  - Time-series: TimescaleDB or InfluxDB
  - Cache: Redis
Database Client:
  - asyncpg (PostgreSQL)
  - redis-py (Redis)
Pub/Sub: redis-py or nats.py
Testing:
  - pytest
  - pytest-asyncio
  - pytest-mock
  - hypothesis (property-based testing)
Observability:
  - prometheus_client
  - structlog
  - opentelemetry-api
Configuration: pydantic-settings
```

---

## 2. Architecture Design

### 2.1 High-Level Component Architecture

```
┌─────────────────────────────────────────────────────────┐
│           Account Monitor Service                       │
│                                                          │
│  ┌──────────────────┐      ┌────────────────────────┐  │
│  │ gRPC Server      │      │ WebSocket Client       │  │
│  │ - AccountQuery   │      │ - User Data Stream     │  │
│  │ - PositionQuery  │      │ - Balance Updates      │  │
│  │ - P&LQuery       │      │ - Position Updates     │  │
│  └────────┬─────────┘      └──────────┬─────────────┘  │
│           │                            │                │
│           ▼                            ▼                │
│  ┌─────────────────────────────────────────────────┐   │
│  │         Position State Manager                  │   │
│  │  - State machine (NONE→LONG→FLAT→SHORT)       │   │
│  │  - Concurrent access synchronization           │   │
│  │  - Entry/exit price tracking                   │   │
│  └────────────────────┬────────────────────────────┘   │
│                       │                                 │
│           ┌───────────┴───────────┐                    │
│           ▼                       ▼                     │
│  ┌─────────────────┐     ┌──────────────────────┐     │
│  │ P&L Calculator  │     │ Balance Tracker      │     │
│  │ - Realized P&L  │     │ - Available balance  │     │
│  │ - Unrealized P&L│     │ - Locked balance     │     │
│  │ - Fee tracking  │     │ - Total equity       │     │
│  └─────────────────┘     └──────────────────────┘     │
│           │                       │                     │
│           ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────┐   │
│  │         Reconciliation Engine                   │   │
│  │  - Periodic sync with exchange account endpoint │   │
│  │  - Drift detection and correction               │   │
│  │  - Discrepancy alerting                         │   │
│  └────────────────────────────────────────────────┘    │
│           │                                             │
│           ▼                                             │
│  ┌─────────────────────────────────────────────────┐   │
│  │         Alert System                            │   │
│  │  - Threshold monitoring (balance, drawdown)     │   │
│  │  - Risk metrics evaluation                      │   │
│  │  - Alert publishing                             │   │
│  └─────────────────────────────────────────────────┘   │
│           │                                             │
│           ▼                                             │
│  ┌─────────────────────────────────────────────────┐   │
│  │         Persistence Layer                       │   │
│  │  - TimescaleDB (P&L history, trades)            │   │
│  │  - Redis (current positions snapshot)           │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Position State Machine

```go
type PositionState int

const (
    PositionNone PositionState = iota  // FLAT - no position
    PositionLong                        // LONG - positive quantity
    PositionShort                       // SHORT - negative quantity
)

type Position struct {
    Symbol        string
    State         PositionState
    Quantity      decimal.Decimal
    EntryPrice    decimal.Decimal
    CurrentPrice  decimal.Decimal
    UnrealizedPnL decimal.Decimal
    RealizedPnL   decimal.Decimal
    LastUpdate    time.Time
    TotalFees     decimal.Decimal
}

// State transitions:
// NONE → LONG (buy order fills)
// NONE → SHORT (sell order fills)
// LONG → NONE (sell reduces quantity to 0)
// LONG → SHORT (sell overshoots, creates short)
// SHORT → NONE (buy reduces quantity to 0)
// SHORT → LONG (buy overshoots, creates long)
```

### 2.3 Balance Tracking System

```go
type AccountBalance struct {
    Asset           string
    Free            decimal.Decimal  // Available for trading
    Locked          decimal.Decimal  // In open orders
    Total           decimal.Decimal  // Free + Locked
    USDValue        decimal.Decimal  // Mark-to-market value
    LastUpdate      time.Time
}

type AccountEquity struct {
    TotalBalance    decimal.Decimal  // Sum of all asset values
    UnrealizedPnL   decimal.Decimal  // From open positions
    RealizedPnL     decimal.Decimal  // Closed position profits
    TotalEquity     decimal.Decimal  // Balance + UnrealizedPnL
    Margin          decimal.Decimal  // Used margin (futures)
    MarginRatio     decimal.Decimal  // Margin / Total Equity
    AvailableMargin decimal.Decimal  // For new positions
}
```

### 2.4 P&L Calculation Engine

**Realized P&L Formula:**
```
For each fill that closes/reduces a position:

realized_pnl = (exit_price - entry_price) * quantity_closed * direction
- direction: +1 for long, -1 for short
- quantity_closed: amount of position reduced
- fees: trading fees on both entry and exit
```

**Unrealized P&L Formula:**
```
For open positions:

unrealized_pnl = (current_price - entry_price) * open_quantity * direction
- current_price: last traded price or mark price
- open_quantity: current position size
- direction: +1 for long, -1 for short
```

**Fee Tracking:**
```
total_fees = maker_fees + taker_fees
fee_rate = exchange.fee_structure[maker|taker]
fee_amount = fill_quantity * fill_price * fee_rate
```

### 2.5 Reconciliation Algorithm

**Periodic Reconciliation (Every 60 seconds):**

```go
func ReconcileAccount(ctx context.Context) error {
    // 1. Fetch current state from exchange
    exchangeAccount, err := exchangeClient.GetAccountInfo(ctx)

    // 2. Compare balances
    for asset, localBalance := range localState.Balances {
        exchangeBalance := exchangeAccount.Balances[asset]

        drift := exchangeBalance.Total.Sub(localBalance.Total)
        if drift.Abs().GreaterThan(TOLERANCE) {
            // Drift detected
            logger.Warn("Balance drift detected",
                "asset", asset,
                "local", localBalance.Total,
                "exchange", exchangeBalance.Total,
                "drift", drift)

            // Correct local state
            localState.Balances[asset] = exchangeBalance

            // Publish alert
            publishAlert(AlertBalanceDrift, asset, drift)
        }
    }

    // 3. Compare positions
    for symbol, localPos := range localState.Positions {
        exchangePos := exchangeAccount.Positions[symbol]

        qtyDrift := exchangePos.Quantity.Sub(localPos.Quantity)
        if qtyDrift.Abs().GreaterThan(POSITION_TOLERANCE) {
            // Position drift detected
            logger.Warn("Position drift detected",
                "symbol", symbol,
                "local_qty", localPos.Quantity,
                "exchange_qty", exchangePos.Quantity,
                "drift", qtyDrift)

            // Recalculate P&L with corrected position
            correctedPnL := calculatePnL(exchangePos)

            // Update local state
            localState.Positions[symbol] = exchangePos

            // Publish alert
            publishAlert(AlertPositionDrift, symbol, qtyDrift)
        }
    }

    return nil
}
```

**Drift Tolerance:**
- Balance: 0.00001 (accounting for floating point)
- Position: 0.0001 quantity
- If drift exceeds threshold: Alert + Auto-correct

### 2.6 Alert System

```go
type AlertType string

const (
    AlertLowBalance      AlertType = "LOW_BALANCE"
    AlertHighDrawdown    AlertType = "HIGH_DRAWDOWN"
    AlertMarginRatio     AlertType = "HIGH_MARGIN_RATIO"
    AlertBalanceDrift    AlertType = "BALANCE_DRIFT"
    AlertPositionDrift   AlertType = "POSITION_DRIFT"
    AlertUnexpectedFill  AlertType = "UNEXPECTED_FILL"
)

type Alert struct {
    Type      AlertType
    Severity  string  // INFO, WARNING, CRITICAL
    Symbol    string
    Message   string
    Value     decimal.Decimal
    Threshold decimal.Decimal
    Timestamp time.Time
}

type AlertThresholds struct {
    MinBalance        decimal.Decimal  // Alert if below
    MaxDrawdown       decimal.Decimal  // Alert if exceeded (e.g., -5%)
    MaxMarginRatio    decimal.Decimal  // Alert if above (e.g., 0.8)
    BalanceDriftPct   decimal.Decimal  // Alert if drift > X%
    PositionDriftPct  decimal.Decimal  // Alert if drift > X%
}

func EvaluateAlerts(state *AccountState) []Alert {
    alerts := []Alert{}

    // Check balance threshold
    if state.Equity.TotalBalance.LessThan(thresholds.MinBalance) {
        alerts = append(alerts, Alert{
            Type: AlertLowBalance,
            Severity: "WARNING",
            Message: "Balance below threshold",
            Value: state.Equity.TotalBalance,
            Threshold: thresholds.MinBalance,
        })
    }

    // Check drawdown
    if state.Equity.RealizedPnL.Div(initialBalance).LessThan(thresholds.MaxDrawdown.Neg()) {
        alerts = append(alerts, Alert{
            Type: AlertHighDrawdown,
            Severity: "CRITICAL",
            Message: "Drawdown exceeds threshold",
            Value: state.Equity.RealizedPnL,
            Threshold: thresholds.MaxDrawdown,
        })
    }

    // Check margin ratio (futures)
    if state.Equity.MarginRatio.GreaterThan(thresholds.MaxMarginRatio) {
        alerts = append(alerts, Alert{
            Type: AlertMarginRatio,
            Severity: "CRITICAL",
            Message: "Margin ratio too high",
            Value: state.Equity.MarginRatio,
            Threshold: thresholds.MaxMarginRatio,
        })
    }

    return alerts
}
```

---

## 3. Development Phases

### Phase 1: gRPC Server & Basic Queries (Week 1)
**Duration:** 3-5 days

**Tasks:**
- Set up Go project structure
- Define Protocol Buffers schema
- Implement gRPC server
- Create basic query endpoints (account state, positions)
- Add health check endpoint
- Write unit tests

**Deliverables:**
```protobuf
// account_monitor.proto
service AccountMonitor {
    rpc GetAccountState(AccountRequest) returns (AccountState);
    rpc GetPosition(PositionRequest) returns (Position);
    rpc GetAllPositions(Empty) returns (PositionList);
    rpc GetPnL(PnLRequest) returns (PnLReport);
    rpc GetBalance(BalanceRequest) returns (Balance);
}
```

**Acceptance Criteria:**
- [ ] gRPC server starts successfully
- [ ] All query methods return mock data
- [ ] Health check returns service status
- [ ] Unit test coverage > 80%

### Phase 2: Exchange WebSocket Client (Week 1-2)
**Duration:** 4-6 days

**Tasks:**
- Implement Binance user data stream WebSocket client
- Handle balance update events
- Handle position update events (futures)
- Handle order execution events
- Implement reconnection logic with exponential backoff
- Add event parsing and validation

**Deliverables:**
```go
type WebSocketClient struct {
    conn          *websocket.Conn
    listenKey     string
    eventHandlers map[string]EventHandler
    reconnectCh   chan struct{}
}

func (c *WebSocketClient) Connect(ctx context.Context) error
func (c *WebSocketClient) SubscribeBalanceUpdates(handler BalanceUpdateHandler)
func (c *WebSocketClient) SubscribePositionUpdates(handler PositionUpdateHandler)
func (c *WebSocketClient) SubscribeOrderUpdates(handler OrderUpdateHandler)
```

**Acceptance Criteria:**
- [ ] WebSocket connects to Binance testnet
- [ ] Balance updates trigger handlers
- [ ] Reconnection works after network failure
- [ ] Events are parsed correctly
- [ ] Integration test with mock exchange

### Phase 3: Balance & Position Tracking (Week 2)
**Duration:** 4-6 days

**Tasks:**
- Implement Position state manager with sync.RWMutex
- Implement Balance tracker
- Handle position state transitions
- Track entry/exit prices with FIFO/LIFO logic
- Implement weighted average price calculation
- Add Redis for position snapshots

**Deliverables:**
```go
type PositionManager struct {
    positions map[string]*Position
    mu        sync.RWMutex
    redis     *redis.Client
}

func (pm *PositionManager) UpdatePosition(fill Fill) error
func (pm *PositionManager) GetPosition(symbol string) (*Position, error)
func (pm *PositionManager) GetAllPositions() ([]*Position, error)
func (pm *PositionManager) SnapshotToRedis(ctx context.Context) error
func (pm *PositionManager) RestoreFromRedis(ctx context.Context) error
```

**Acceptance Criteria:**
- [ ] Position updates handle all state transitions
- [ ] Concurrent access is thread-safe (race detector passes)
- [ ] Entry price calculation is accurate
- [ ] Positions persist to Redis
- [ ] Restore from Redis on service restart

### Phase 4: P&L Calculation (Week 3)
**Duration:** 5-7 days

**Tasks:**
- Implement realized P&L calculator
- Implement unrealized P&L calculator
- Track fees per trade
- Implement cumulative P&L tracking
- Add TimescaleDB for P&L history
- Create P&L aggregation queries (daily, weekly, monthly)

**Deliverables:**
```go
type PnLCalculator struct {
    positions     *PositionManager
    priceProvider PriceProvider
    db            *pgxpool.Pool
}

func (p *PnLCalculator) CalculateRealizedPnL(fill Fill) (decimal.Decimal, error)
func (p *PnLCalculator) CalculateUnrealizedPnL(position Position) (decimal.Decimal, error)
func (p *PnLCalculator) GetTotalPnL() (*PnLReport, error)
func (p *PnLCalculator) GetPnLHistory(from, to time.Time) ([]*PnLSnapshot, error)
func (p *PnLCalculator) StorePnLSnapshot(ctx context.Context, snapshot *PnLSnapshot) error
```

**Database Schema:**
```sql
CREATE TABLE pnl_snapshots (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20),
    realized_pnl DECIMAL(20, 8) NOT NULL,
    unrealized_pnl DECIMAL(20, 8) NOT NULL,
    total_pnl DECIMAL(20, 8) NOT NULL,
    total_fees DECIMAL(20, 8) NOT NULL,
    equity DECIMAL(20, 8) NOT NULL
);

SELECT create_hypertable('pnl_snapshots', 'timestamp');

CREATE INDEX idx_pnl_snapshots_timestamp ON pnl_snapshots(timestamp DESC);
CREATE INDEX idx_pnl_snapshots_symbol ON pnl_snapshots(symbol, timestamp DESC);
```

**Acceptance Criteria:**
- [ ] Realized P&L calculation matches manual calculation
- [ ] Unrealized P&L updates with price changes
- [ ] Fees are tracked accurately
- [ ] P&L history persists to TimescaleDB
- [ ] Query performance < 100ms for 1 week of data

### Phase 5: Reconciliation Mechanism (Week 3-4)
**Duration:** 4-5 days

**Tasks:**
- Implement periodic reconciliation job
- Add exchange account info API client
- Compare local state with exchange state
- Detect and log drifts
- Auto-correct on drift detection
- Add reconciliation metrics

**Deliverables:**
```go
type Reconciler struct {
    localState    *AccountState
    exchangeAPI   ExchangeClient
    alertPublisher AlertPublisher
    ticker        *time.Ticker
}

func (r *Reconciler) Start(ctx context.Context, interval time.Duration)
func (r *Reconciler) ReconcileNow(ctx context.Context) (*ReconciliationReport, error)
func (r *Reconciler) detectDrifts() []Drift
func (r *Reconciler) correctDrifts(drifts []Drift) error
```

**Acceptance Criteria:**
- [ ] Reconciliation runs every 60 seconds
- [ ] Drifts are detected correctly
- [ ] Auto-correction updates local state
- [ ] Alerts are published on drift
- [ ] Reconciliation report includes all drifts

### Phase 6: Alert System (Week 4)
**Duration:** 3-4 days

**Tasks:**
- Implement threshold monitoring
- Add configurable alert rules
- Publish alerts to Redis Pub/Sub
- Add alert history storage
- Create alert suppression logic (no spam)

**Deliverables:**
```go
type AlertManager struct {
    thresholds     *AlertThresholds
    publisher      PubSubPublisher
    db             *pgxpool.Pool
    suppressionMap map[string]time.Time  // Last alert time
}

func (am *AlertManager) EvaluateAlerts(state *AccountState) []Alert
func (am *AlertManager) PublishAlert(alert Alert) error
func (am *AlertManager) SuppressAlert(alertType AlertType, duration time.Duration) bool
func (am *AlertManager) GetAlertHistory(from, to time.Time) ([]*Alert, error)
```

**Acceptance Criteria:**
- [ ] Alerts trigger when thresholds are exceeded
- [ ] Alerts publish to Redis Pub/Sub
- [ ] Alert suppression prevents spam (max 1 per minute)
- [ ] Alert history persists to database
- [ ] Configurable thresholds via config file

### Phase 7: Observability UI (Week 5)
**Duration:** 5-7 days

**Tasks:**
- Create HTTP server for dashboard
- Build account balance view
- Build position breakdown view
- Build P&L chart (daily, cumulative)
- Build trade statistics view
- Add WebSocket updates for real-time data

**Deliverables:**
- Web dashboard with:
  - Real-time balance display
  - Position table with P&L
  - P&L time-series chart
  - Win rate and statistics
  - Alert feed

**Acceptance Criteria:**
- [ ] Dashboard displays real-time account state
- [ ] P&L chart updates every second
- [ ] Position table shows unrealized P&L
- [ ] Statistics calculate correctly
- [ ] WebSocket updates have < 100ms latency

---

## 4. Implementation Details

### 4.1 Position Tracking Logic

```go
package position

import (
    "sync"
    "github.com/shopspring/decimal"
)

type Position struct {
    Symbol        string
    Quantity      decimal.Decimal  // Positive = Long, Negative = Short
    EntryPrice    decimal.Decimal  // Weighted average entry price
    CurrentPrice  decimal.Decimal
    RealizedPnL   decimal.Decimal
    UnrealizedPnL decimal.Decimal
    TotalFees     decimal.Decimal
    Trades        []Trade
}

type Trade struct {
    ID           string
    Timestamp    time.Time
    Side         string  // BUY or SELL
    Quantity     decimal.Decimal
    Price        decimal.Decimal
    Fee          decimal.Decimal
    FeeCurrency  string
}

type PositionManager struct {
    positions map[string]*Position
    mu        sync.RWMutex
}

func (pm *PositionManager) UpdatePosition(symbol string, fill Fill) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    pos, exists := pm.positions[symbol]
    if !exists {
        pos = &Position{
            Symbol:      symbol,
            Quantity:    decimal.Zero,
            EntryPrice:  decimal.Zero,
            RealizedPnL: decimal.Zero,
            Trades:      []Trade{},
        }
        pm.positions[symbol] = pos
    }

    // Create trade record
    trade := Trade{
        ID:          fill.ID,
        Timestamp:   fill.Timestamp,
        Side:        fill.Side,
        Quantity:    fill.Quantity,
        Price:       fill.Price,
        Fee:         fill.Fee,
        FeeCurrency: fill.FeeCurrency,
    }
    pos.Trades = append(pos.Trades, trade)

    // Calculate P&L and update position
    fillQty := fill.Quantity
    if fill.Side == "SELL" {
        fillQty = fillQty.Neg()
    }

    oldQty := pos.Quantity
    newQty := oldQty.Add(fillQty)

    // Case 1: Opening new position
    if oldQty.IsZero() {
        pos.EntryPrice = fill.Price
        pos.Quantity = newQty
    } else if oldQty.Sign() == fillQty.Sign() {
        // Case 2: Adding to position (same direction)
        // Calculate weighted average entry price
        oldValue := oldQty.Mul(pos.EntryPrice)
        newValue := fillQty.Abs().Mul(fill.Price)
        totalValue := oldValue.Add(newValue)
        pos.EntryPrice = totalValue.Div(newQty.Abs())
        pos.Quantity = newQty
    } else {
        // Case 3: Reducing or reversing position (opposite direction)
        closedQty := decimal.Min(oldQty.Abs(), fillQty.Abs())

        // Calculate realized P&L for closed portion
        var pnl decimal.Decimal
        if oldQty.IsPositive() {
            // Closing long: (sell_price - entry_price) * qty
            pnl = fill.Price.Sub(pos.EntryPrice).Mul(closedQty)
        } else {
            // Closing short: (entry_price - buy_price) * qty
            pnl = pos.EntryPrice.Sub(fill.Price).Mul(closedQty)
        }

        // Subtract fees from P&L
        pnl = pnl.Sub(fill.Fee)
        pos.RealizedPnL = pos.RealizedPnL.Add(pnl)

        // Update quantity
        pos.Quantity = newQty

        // If position reversed, update entry price
        if newQty.Sign() != oldQty.Sign() && !newQty.IsZero() {
            pos.EntryPrice = fill.Price
        }

        // If position closed completely, reset entry price
        if newQty.IsZero() {
            pos.EntryPrice = decimal.Zero
        }
    }

    // Update total fees
    pos.TotalFees = pos.TotalFees.Add(fill.Fee)

    return nil
}

func (pm *PositionManager) CalculateUnrealizedPnL(symbol string, currentPrice decimal.Decimal) (decimal.Decimal, error) {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    pos, exists := pm.positions[symbol]
    if !exists || pos.Quantity.IsZero() {
        return decimal.Zero, nil
    }

    var unrealizedPnL decimal.Decimal
    if pos.Quantity.IsPositive() {
        // Long position: (current_price - entry_price) * qty
        unrealizedPnL = currentPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
    } else {
        // Short position: (entry_price - current_price) * |qty|
        unrealizedPnL = pos.EntryPrice.Sub(currentPrice).Mul(pos.Quantity.Abs())
    }

    pos.CurrentPrice = currentPrice
    pos.UnrealizedPnL = unrealizedPnL

    return unrealizedPnL, nil
}
```

### 4.2 P&L Calculation Implementation

```go
package pnl

import (
    "context"
    "time"
    "github.com/shopspring/decimal"
)

type PnLCalculator struct {
    positionMgr   *position.PositionManager
    priceProvider PriceProvider
    db            *pgxpool.Pool
}

type PnLReport struct {
    Timestamp       time.Time
    RealizedPnL     decimal.Decimal
    UnrealizedPnL   decimal.Decimal
    TotalPnL        decimal.Decimal
    TotalFees       decimal.Decimal
    NetPnL          decimal.Decimal
    WinRate         decimal.Decimal
    TotalTrades     int
    WinningTrades   int
    LosingTrades    int
    AverageWin      decimal.Decimal
    AverageLoss     decimal.Decimal
    ProfitFactor    decimal.Decimal
}

func (p *PnLCalculator) GetCurrentPnL(ctx context.Context) (*PnLReport, error) {
    positions := p.positionMgr.GetAllPositions()

    var totalRealizedPnL decimal.Decimal
    var totalUnrealizedPnL decimal.Decimal
    var totalFees decimal.Decimal

    for _, pos := range positions {
        // Get current price
        currentPrice, err := p.priceProvider.GetPrice(pos.Symbol)
        if err != nil {
            return nil, err
        }

        // Calculate unrealized P&L
        unrealizedPnL, err := p.positionMgr.CalculateUnrealizedPnL(pos.Symbol, currentPrice)
        if err != nil {
            return nil, err
        }

        totalRealizedPnL = totalRealizedPnL.Add(pos.RealizedPnL)
        totalUnrealizedPnL = totalUnrealizedPnL.Add(unrealizedPnL)
        totalFees = totalFees.Add(pos.TotalFees)
    }

    totalPnL := totalRealizedPnL.Add(totalUnrealizedPnL)
    netPnL := totalPnL.Sub(totalFees)

    // Calculate statistics
    stats := p.calculateStatistics(positions)

    report := &PnLReport{
        Timestamp:     time.Now(),
        RealizedPnL:   totalRealizedPnL,
        UnrealizedPnL: totalUnrealizedPnL,
        TotalPnL:      totalPnL,
        TotalFees:     totalFees,
        NetPnL:        netPnL,
        WinRate:       stats.WinRate,
        TotalTrades:   stats.TotalTrades,
        WinningTrades: stats.WinningTrades,
        LosingTrades:  stats.LosingTrades,
        AverageWin:    stats.AverageWin,
        AverageLoss:   stats.AverageLoss,
        ProfitFactor:  stats.ProfitFactor,
    }

    return report, nil
}

type TradeStatistics struct {
    WinRate       decimal.Decimal
    TotalTrades   int
    WinningTrades int
    LosingTrades  int
    AverageWin    decimal.Decimal
    AverageLoss   decimal.Decimal
    ProfitFactor  decimal.Decimal
}

func (p *PnLCalculator) calculateStatistics(positions []*position.Position) TradeStatistics {
    var totalWins decimal.Decimal
    var totalLosses decimal.Decimal
    winCount := 0
    lossCount := 0

    for _, pos := range positions {
        for _, trade := range pos.Trades {
            // Only count closed trades
            if trade.RealizedPnL.IsZero() {
                continue
            }

            if trade.RealizedPnL.IsPositive() {
                totalWins = totalWins.Add(trade.RealizedPnL)
                winCount++
            } else {
                totalLosses = totalLosses.Add(trade.RealizedPnL.Abs())
                lossCount++
            }
        }
    }

    totalTrades := winCount + lossCount
    var winRate decimal.Decimal
    if totalTrades > 0 {
        winRate = decimal.NewFromInt(int64(winCount)).Div(decimal.NewFromInt(int64(totalTrades))).Mul(decimal.NewFromInt(100))
    }

    var avgWin decimal.Decimal
    if winCount > 0 {
        avgWin = totalWins.Div(decimal.NewFromInt(int64(winCount)))
    }

    var avgLoss decimal.Decimal
    if lossCount > 0 {
        avgLoss = totalLosses.Div(decimal.NewFromInt(int64(lossCount)))
    }

    var profitFactor decimal.Decimal
    if !totalLosses.IsZero() {
        profitFactor = totalWins.Div(totalLosses)
    }

    return TradeStatistics{
        WinRate:       winRate,
        TotalTrades:   totalTrades,
        WinningTrades: winCount,
        LosingTrades:  lossCount,
        AverageWin:    avgWin,
        AverageLoss:   avgLoss,
        ProfitFactor:  profitFactor,
    }
}

func (p *PnLCalculator) StorePnLSnapshot(ctx context.Context, report *PnLReport) error {
    query := `
        INSERT INTO pnl_snapshots (
            timestamp, realized_pnl, unrealized_pnl, total_pnl,
            total_fees, net_pnl, win_rate, total_trades
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

    _, err := p.db.Exec(ctx, query,
        report.Timestamp,
        report.RealizedPnL,
        report.UnrealizedPnL,
        report.TotalPnL,
        report.TotalFees,
        report.NetPnL,
        report.WinRate,
        report.TotalTrades,
    )

    return err
}

func (p *PnLCalculator) GetPnLHistory(ctx context.Context, from, to time.Time, interval string) ([]*PnLReport, error) {
    // Aggregate P&L by time interval (1h, 1d, 1w)
    query := `
        SELECT
            time_bucket($1, timestamp) as bucket,
            AVG(realized_pnl) as avg_realized,
            AVG(unrealized_pnl) as avg_unrealized,
            AVG(total_pnl) as avg_total,
            SUM(total_fees) as sum_fees,
            AVG(win_rate) as avg_win_rate
        FROM pnl_snapshots
        WHERE timestamp >= $2 AND timestamp <= $3
        GROUP BY bucket
        ORDER BY bucket DESC
    `

    rows, err := p.db.Query(ctx, query, interval, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    reports := []*PnLReport{}
    for rows.Next() {
        var r PnLReport
        err := rows.Scan(
            &r.Timestamp,
            &r.RealizedPnL,
            &r.UnrealizedPnL,
            &r.TotalPnL,
            &r.TotalFees,
            &r.WinRate,
        )
        if err != nil {
            return nil, err
        }
        r.NetPnL = r.TotalPnL.Sub(r.TotalFees)
        reports = append(reports, &r)
    }

    return reports, nil
}
```

### 4.3 Reconciliation Logic Implementation

```go
package reconciliation

import (
    "context"
    "time"
    "github.com/shopspring/decimal"
)

type Reconciler struct {
    positionMgr    *position.PositionManager
    balanceMgr     *balance.BalanceManager
    exchangeClient ExchangeClient
    alertPublisher AlertPublisher
    logger         *zerolog.Logger

    balanceTolerance  decimal.Decimal
    positionTolerance decimal.Decimal
}

type ReconciliationReport struct {
    Timestamp       time.Time
    BalanceDrifts   []BalanceDrift
    PositionDrifts  []PositionDrift
    Corrected       bool
    Error           error
}

type BalanceDrift struct {
    Asset         string
    LocalBalance  decimal.Decimal
    ExchangeBalance decimal.Decimal
    Drift         decimal.Decimal
    DriftPercent  decimal.Decimal
}

type PositionDrift struct {
    Symbol           string
    LocalQuantity    decimal.Decimal
    ExchangeQuantity decimal.Decimal
    Drift            decimal.Decimal
    DriftPercent     decimal.Decimal
}

func NewReconciler(
    positionMgr *position.PositionManager,
    balanceMgr *balance.BalanceManager,
    exchangeClient ExchangeClient,
    alertPublisher AlertPublisher,
    logger *zerolog.Logger,
) *Reconciler {
    return &Reconciler{
        positionMgr:       positionMgr,
        balanceMgr:        balanceMgr,
        exchangeClient:    exchangeClient,
        alertPublisher:    alertPublisher,
        logger:            logger,
        balanceTolerance:  decimal.NewFromFloat(0.00001),  // 0.001%
        positionTolerance: decimal.NewFromFloat(0.0001),   // 0.01%
    }
}

func (r *Reconciler) Start(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            report, err := r.ReconcileNow(ctx)
            if err != nil {
                r.logger.Error().Err(err).Msg("Reconciliation failed")
                continue
            }

            if len(report.BalanceDrifts) > 0 || len(report.PositionDrifts) > 0 {
                r.logger.Warn().
                    Int("balance_drifts", len(report.BalanceDrifts)).
                    Int("position_drifts", len(report.PositionDrifts)).
                    Msg("Drifts detected during reconciliation")
            }

        case <-ctx.Done():
            return
        }
    }
}

func (r *Reconciler) ReconcileNow(ctx context.Context) (*ReconciliationReport, error) {
    report := &ReconciliationReport{
        Timestamp: time.Now(),
    }

    // Fetch exchange account state
    exchangeAccount, err := r.exchangeClient.GetAccountInfo(ctx)
    if err != nil {
        report.Error = err
        return report, err
    }

    // 1. Reconcile Balances
    balanceDrifts := r.reconcileBalances(exchangeAccount.Balances)
    report.BalanceDrifts = balanceDrifts

    // 2. Reconcile Positions
    positionDrifts := r.reconcilePositions(exchangeAccount.Positions)
    report.PositionDrifts = positionDrifts

    // 3. Auto-correct if drifts detected
    if len(balanceDrifts) > 0 || len(positionDrifts) > 0 {
        err := r.correctDrifts(balanceDrifts, positionDrifts)
        if err != nil {
            report.Error = err
            return report, err
        }
        report.Corrected = true

        // Publish alerts
        r.publishDriftAlerts(balanceDrifts, positionDrifts)
    }

    return report, nil
}

func (r *Reconciler) reconcileBalances(exchangeBalances map[string]ExchangeBalance) []BalanceDrift {
    drifts := []BalanceDrift{}

    localBalances := r.balanceMgr.GetAllBalances()

    for asset, localBal := range localBalances {
        exchangeBal, exists := exchangeBalances[asset]
        if !exists {
            continue
        }

        drift := exchangeBal.Total.Sub(localBal.Total)
        driftPct := decimal.Zero
        if !localBal.Total.IsZero() {
            driftPct = drift.Div(localBal.Total).Mul(decimal.NewFromInt(100))
        }

        if drift.Abs().GreaterThan(r.balanceTolerance) {
            drifts = append(drifts, BalanceDrift{
                Asset:           asset,
                LocalBalance:    localBal.Total,
                ExchangeBalance: exchangeBal.Total,
                Drift:           drift,
                DriftPercent:    driftPct,
            })
        }
    }

    return drifts
}

func (r *Reconciler) reconcilePositions(exchangePositions map[string]ExchangePosition) []PositionDrift {
    drifts := []PositionDrift{}

    localPositions := r.positionMgr.GetAllPositions()

    for symbol, localPos := range localPositions {
        exchangePos, exists := exchangePositions[symbol]
        if !exists {
            continue
        }

        drift := exchangePos.Quantity.Sub(localPos.Quantity)
        driftPct := decimal.Zero
        if !localPos.Quantity.IsZero() {
            driftPct = drift.Div(localPos.Quantity.Abs()).Mul(decimal.NewFromInt(100))
        }

        if drift.Abs().GreaterThan(r.positionTolerance) {
            drifts = append(drifts, PositionDrift{
                Symbol:           symbol,
                LocalQuantity:    localPos.Quantity,
                ExchangeQuantity: exchangePos.Quantity,
                Drift:            drift,
                DriftPercent:     driftPct,
            })
        }
    }

    return drifts
}

func (r *Reconciler) correctDrifts(balanceDrifts []BalanceDrift, positionDrifts []PositionDrift) error {
    // Correct balance drifts
    for _, drift := range balanceDrifts {
        err := r.balanceMgr.SetBalance(drift.Asset, drift.ExchangeBalance)
        if err != nil {
            return err
        }

        r.logger.Info().
            Str("asset", drift.Asset).
            Str("old_balance", drift.LocalBalance.String()).
            Str("new_balance", drift.ExchangeBalance.String()).
            Str("drift", drift.Drift.String()).
            Msg("Corrected balance drift")
    }

    // Correct position drifts
    for _, drift := range positionDrifts {
        err := r.positionMgr.SetPosition(drift.Symbol, drift.ExchangeQuantity)
        if err != nil {
            return err
        }

        r.logger.Info().
            Str("symbol", drift.Symbol).
            Str("old_qty", drift.LocalQuantity.String()).
            Str("new_qty", drift.ExchangeQuantity.String()).
            Str("drift", drift.Drift.String()).
            Msg("Corrected position drift")
    }

    return nil
}

func (r *Reconciler) publishDriftAlerts(balanceDrifts []BalanceDrift, positionDrifts []PositionDrift) {
    for _, drift := range balanceDrifts {
        alert := Alert{
            Type:      AlertBalanceDrift,
            Severity:  "WARNING",
            Symbol:    drift.Asset,
            Message:   fmt.Sprintf("Balance drift detected: %s%%", drift.DriftPercent.StringFixed(2)),
            Value:     drift.Drift,
            Timestamp: time.Now(),
        }
        r.alertPublisher.Publish(alert)
    }

    for _, drift := range positionDrifts {
        alert := Alert{
            Type:      AlertPositionDrift,
            Severity:  "WARNING",
            Symbol:    drift.Symbol,
            Message:   fmt.Sprintf("Position drift detected: %s%%", drift.DriftPercent.StringFixed(2)),
            Value:     drift.Drift,
            Timestamp: time.Now(),
        }
        r.alertPublisher.Publish(alert)
    }
}
```

---

## 5. Testing Strategy

### 5.1 P&L Calculation Unit Tests

```go
package pnl_test

import (
    "testing"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
)

func TestRealizedPnL_LongPosition(t *testing.T) {
    pm := position.NewPositionManager()

    // Buy 1 BTC @ 50000
    fill1 := Fill{
        ID:       "1",
        Symbol:   "BTCUSDT",
        Side:     "BUY",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(50000),
        Fee:      decimal.NewFromFloat(25),  // 0.05% fee
    }
    pm.UpdatePosition("BTCUSDT", fill1)

    // Sell 1 BTC @ 55000
    fill2 := Fill{
        ID:       "2",
        Symbol:   "BTCUSDT",
        Side:     "SELL",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(55000),
        Fee:      decimal.NewFromFloat(27.5),  // 0.05% fee
    }
    pm.UpdatePosition("BTCUSDT", fill2)

    pos, _ := pm.GetPosition("BTCUSDT")

    // Expected P&L = (55000 - 50000) * 1 - 25 - 27.5 = 4947.5
    expectedPnL := decimal.NewFromFloat(4947.5)
    assert.True(t, pos.RealizedPnL.Equal(expectedPnL),
        "Expected P&L %s, got %s", expectedPnL, pos.RealizedPnL)
    assert.True(t, pos.Quantity.IsZero(), "Position should be closed")
}

func TestRealizedPnL_ShortPosition(t *testing.T) {
    pm := position.NewPositionManager()

    // Sell 1 BTC @ 50000 (short)
    fill1 := Fill{
        ID:       "1",
        Symbol:   "BTCUSDT",
        Side:     "SELL",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(50000),
        Fee:      decimal.NewFromFloat(25),
    }
    pm.UpdatePosition("BTCUSDT", fill1)

    // Buy 1 BTC @ 45000 (cover short)
    fill2 := Fill{
        ID:       "2",
        Symbol:   "BTCUSDT",
        Side:     "BUY",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(45000),
        Fee:      decimal.NewFromFloat(22.5),
    }
    pm.UpdatePosition("BTCUSDT", fill2)

    pos, _ := pm.GetPosition("BTCUSDT")

    // Expected P&L = (50000 - 45000) * 1 - 25 - 22.5 = 4952.5
    expectedPnL := decimal.NewFromFloat(4952.5)
    assert.True(t, pos.RealizedPnL.Equal(expectedPnL))
}

func TestWeightedAveragePrice(t *testing.T) {
    pm := position.NewPositionManager()

    // Buy 1 BTC @ 50000
    fill1 := Fill{
        Symbol:   "BTCUSDT",
        Side:     "BUY",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(50000),
    }
    pm.UpdatePosition("BTCUSDT", fill1)

    // Buy 1 BTC @ 52000
    fill2 := Fill{
        Symbol:   "BTCUSDT",
        Side:     "BUY",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(52000),
    }
    pm.UpdatePosition("BTCUSDT", fill2)

    pos, _ := pm.GetPosition("BTCUSDT")

    // Expected avg price = (50000 + 52000) / 2 = 51000
    expectedAvg := decimal.NewFromFloat(51000)
    assert.True(t, pos.EntryPrice.Equal(expectedAvg))
    assert.True(t, pos.Quantity.Equal(decimal.NewFromFloat(2.0)))
}

func TestPositionReversal(t *testing.T) {
    pm := position.NewPositionManager()

    // Buy 1 BTC @ 50000 (long)
    fill1 := Fill{
        Symbol:   "BTCUSDT",
        Side:     "BUY",
        Quantity: decimal.NewFromFloat(1.0),
        Price:    decimal.NewFromFloat(50000),
        Fee:      decimal.Zero,
    }
    pm.UpdatePosition("BTCUSDT", fill1)

    // Sell 2 BTC @ 52000 (close long, open short)
    fill2 := Fill{
        Symbol:   "BTCUSDT",
        Side:     "SELL",
        Quantity: decimal.NewFromFloat(2.0),
        Price:    decimal.NewFromFloat(52000),
        Fee:      decimal.Zero,
    }
    pm.UpdatePosition("BTCUSDT", fill2)

    pos, _ := pm.GetPosition("BTCUSDT")

    // Should have 1 BTC short position
    assert.True(t, pos.Quantity.Equal(decimal.NewFromFloat(-1.0)))
    assert.True(t, pos.EntryPrice.Equal(decimal.NewFromFloat(52000)))

    // Realized P&L from closing long = (52000 - 50000) * 1 = 2000
    assert.True(t, pos.RealizedPnL.Equal(decimal.NewFromFloat(2000)))
}
```

### 5.2 Reconciliation Test Scenarios

```go
package reconciliation_test

import (
    "testing"
    "context"
    "github.com/shopspring/decimal"
)

func TestReconciliation_BalanceDrift(t *testing.T) {
    // Setup mocks
    mockExchange := &MockExchangeClient{}
    mockExchange.On("GetAccountInfo").Return(&ExchangeAccount{
        Balances: map[string]ExchangeBalance{
            "USDT": {Total: decimal.NewFromFloat(10000)},
        },
    }, nil)

    balanceMgr := balance.NewBalanceManager()
    balanceMgr.SetBalance("USDT", decimal.NewFromFloat(9900))  // Drift of 100

    reconciler := NewReconciler(nil, balanceMgr, mockExchange, nil, logger)

    report, err := reconciler.ReconcileNow(context.Background())

    assert.NoError(t, err)
    assert.Len(t, report.BalanceDrifts, 1)
    assert.Equal(t, "USDT", report.BalanceDrifts[0].Asset)
    assert.True(t, report.BalanceDrifts[0].Drift.Equal(decimal.NewFromFloat(100)))
    assert.True(t, report.Corrected)

    // Verify correction
    correctedBalance := balanceMgr.GetBalance("USDT")
    assert.True(t, correctedBalance.Equal(decimal.NewFromFloat(10000)))
}

func TestReconciliation_PositionDrift(t *testing.T) {
    mockExchange := &MockExchangeClient{}
    mockExchange.On("GetAccountInfo").Return(&ExchangeAccount{
        Positions: map[string]ExchangePosition{
            "BTCUSDT": {Quantity: decimal.NewFromFloat(1.5)},
        },
    }, nil)

    positionMgr := position.NewPositionManager()
    positionMgr.SetPosition("BTCUSDT", decimal.NewFromFloat(1.0))  // Drift of 0.5

    reconciler := NewReconciler(positionMgr, nil, mockExchange, nil, logger)

    report, err := reconciler.ReconcileNow(context.Background())

    assert.NoError(t, err)
    assert.Len(t, report.PositionDrifts, 1)
    assert.True(t, report.Corrected)

    // Verify correction
    pos, _ := positionMgr.GetPosition("BTCUSDT")
    assert.True(t, pos.Quantity.Equal(decimal.NewFromFloat(1.5)))
}

func TestReconciliation_WithinTolerance(t *testing.T) {
    // Drift of 0.00001 should not trigger alert
    mockExchange := &MockExchangeClient{}
    mockExchange.On("GetAccountInfo").Return(&ExchangeAccount{
        Balances: map[string]ExchangeBalance{
            "USDT": {Total: decimal.NewFromFloat(10000.00001)},
        },
    }, nil)

    balanceMgr := balance.NewBalanceManager()
    balanceMgr.SetBalance("USDT", decimal.NewFromFloat(10000.00000))

    reconciler := NewReconciler(nil, balanceMgr, mockExchange, nil, logger)

    report, err := reconciler.ReconcileNow(context.Background())

    assert.NoError(t, err)
    assert.Len(t, report.BalanceDrifts, 0)
    assert.False(t, report.Corrected)
}
```

### 5.3 Alert Threshold Tests

```go
package alert_test

func TestAlert_LowBalance(t *testing.T) {
    thresholds := AlertThresholds{
        MinBalance: decimal.NewFromFloat(1000),
    }

    state := &AccountState{
        Equity: AccountEquity{
            TotalBalance: decimal.NewFromFloat(900),
        },
    }

    alertMgr := NewAlertManager(thresholds, nil, nil)
    alerts := alertMgr.EvaluateAlerts(state)

    assert.Len(t, alerts, 1)
    assert.Equal(t, AlertLowBalance, alerts[0].Type)
    assert.Equal(t, "WARNING", alerts[0].Severity)
}

func TestAlert_HighDrawdown(t *testing.T) {
    initialBalance := decimal.NewFromFloat(10000)
    thresholds := AlertThresholds{
        MaxDrawdown: decimal.NewFromFloat(-5),  // -5%
    }

    state := &AccountState{
        Equity: AccountEquity{
            RealizedPnL: decimal.NewFromFloat(-600),  // -6% drawdown
        },
    }

    alertMgr := NewAlertManager(thresholds, nil, nil)
    alertMgr.SetInitialBalance(initialBalance)
    alerts := alertMgr.EvaluateAlerts(state)

    assert.Len(t, alerts, 1)
    assert.Equal(t, AlertHighDrawdown, alerts[0].Type)
    assert.Equal(t, "CRITICAL", alerts[0].Severity)
}

func TestAlert_Suppression(t *testing.T) {
    alertMgr := NewAlertManager(AlertThresholds{}, mockPublisher, db)

    alert := Alert{Type: AlertLowBalance}

    // First alert should publish
    err := alertMgr.PublishAlert(alert)
    assert.NoError(t, err)
    assert.Equal(t, 1, mockPublisher.PublishCount)

    // Second alert within 1 minute should be suppressed
    err = alertMgr.PublishAlert(alert)
    assert.NoError(t, err)
    assert.Equal(t, 1, mockPublisher.PublishCount)  // Still 1

    // After 1 minute, should publish again
    time.Sleep(61 * time.Second)
    err = alertMgr.PublishAlert(alert)
    assert.NoError(t, err)
    assert.Equal(t, 2, mockPublisher.PublishCount)
}
```

### 5.4 Race Condition Testing

```go
package position_test

import (
    "testing"
    "sync"
)

func TestPositionManager_ConcurrentUpdates(t *testing.T) {
    pm := position.NewPositionManager()

    var wg sync.WaitGroup
    numGoroutines := 100

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            fill := Fill{
                Symbol:   "BTCUSDT",
                Side:     "BUY",
                Quantity: decimal.NewFromFloat(0.1),
                Price:    decimal.NewFromFloat(50000),
            }

            err := pm.UpdatePosition("BTCUSDT", fill)
            assert.NoError(t, err)
        }(i)
    }

    wg.Wait()

    pos, _ := pm.GetPosition("BTCUSDT")
    expected := decimal.NewFromFloat(10.0)  // 100 * 0.1
    assert.True(t, pos.Quantity.Equal(expected))
}

// Run with: go test -race
func TestBalanceManager_RaceConditions(t *testing.T) {
    bm := balance.NewBalanceManager()
    bm.SetBalance("USDT", decimal.NewFromFloat(10000))

    var wg sync.WaitGroup

    // Concurrent reads and writes
    for i := 0; i < 50; i++ {
        wg.Add(2)

        // Writer
        go func() {
            defer wg.Done()
            bm.UpdateBalance("USDT", decimal.NewFromFloat(100))
        }()

        // Reader
        go func() {
            defer wg.Done()
            _ = bm.GetBalance("USDT")
        }()
    }

    wg.Wait()
}
```

---

## 6. Deployment

### 6.1 Dockerfile

```dockerfile
# Multi-stage build for Go service
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o account-monitor ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/account-monitor .

# Copy config
COPY config.yaml .

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Expose ports
EXPOSE 50051 8080 9090

CMD ["./account-monitor"]
```

### 6.2 Docker Compose Configuration

```yaml
version: '3.8'

services:
  account-monitor:
    build: .
    container_name: account-monitor
    restart: unless-stopped
    environment:
      - BINANCE_API_KEY=${BINANCE_API_KEY}
      - BINANCE_SECRET_KEY=${BINANCE_SECRET_KEY}
      - BINANCE_TESTNET=${BINANCE_TESTNET:-true}
      - POSTGRES_HOST=timescaledb
      - POSTGRES_PORT=5432
      - POSTGRES_DB=trading
      - POSTGRES_USER=trading
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - LOG_LEVEL=info
      - RECONCILE_INTERVAL=60s
    ports:
      - "50051:50051"  # gRPC
      - "8080:8080"    # HTTP/WebSocket
      - "9090:9090"    # Metrics
    depends_on:
      - timescaledb
      - redis
    networks:
      - trading-net
    volumes:
      - ./config.yaml:/root/config.yaml:ro
      - ./logs:/root/logs
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  timescaledb:
    image: timescale/timescaledb:latest-pg14
    container_name: timescaledb
    restart: unless-stopped
    environment:
      - POSTGRES_DB=trading
      - POSTGRES_USER=trading
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
    volumes:
      - timescaledb-data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - trading-net

  redis:
    image: redis:7-alpine
    container_name: redis
    restart: unless-stopped
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - trading-net

volumes:
  timescaledb-data:
  redis-data:

networks:
  trading-net:
    driver: bridge
```

### 6.3 API Secrets Management

**Option 1: Environment Variables (Development)**
```bash
# .env file
BINANCE_API_KEY=your_api_key_here
BINANCE_SECRET_KEY=your_secret_key_here
POSTGRES_PASSWORD=secure_password
```

**Option 2: Docker Secrets (Production)**
```yaml
# docker-compose.yml
services:
  account-monitor:
    secrets:
      - binance_api_key
      - binance_secret_key
      - postgres_password
    environment:
      - BINANCE_API_KEY_FILE=/run/secrets/binance_api_key
      - BINANCE_SECRET_KEY_FILE=/run/secrets/binance_secret_key

secrets:
  binance_api_key:
    external: true
  binance_secret_key:
    external: true
```

**Option 3: HashiCorp Vault (Production)**
```go
import "github.com/hashicorp/vault/api"

func LoadSecretsFromVault() error {
    client, err := api.NewClient(&api.Config{
        Address: os.Getenv("VAULT_ADDR"),
    })

    client.SetToken(os.Getenv("VAULT_TOKEN"))

    secret, err := client.Logical().Read("secret/data/trading/binance")
    if err != nil {
        return err
    }

    data := secret.Data["data"].(map[string]interface{})
    os.Setenv("BINANCE_API_KEY", data["api_key"].(string))
    os.Setenv("BINANCE_SECRET_KEY", data["secret_key"].(string))

    return nil
}
```

### 6.4 Health Check Implementation

```go
package health

import (
    "context"
    "net/http"
    "encoding/json"
    "time"
)

type HealthStatus struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Uptime      string            `json:"uptime"`
    Checks      map[string]Check  `json:"checks"`
}

type Check struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
}

type HealthChecker struct {
    db            *pgxpool.Pool
    redis         *redis.Client
    wsClient      *WebSocketClient
    startTime     time.Time
}

func (h *HealthChecker) HandleHealth(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    status := HealthStatus{
        Status:  "healthy",
        Version: "1.0.0",
        Uptime:  time.Since(h.startTime).String(),
        Checks:  make(map[string]Check),
    }

    // Check database
    dbCheck := h.checkDatabase(ctx)
    status.Checks["database"] = dbCheck
    if dbCheck.Status != "ok" {
        status.Status = "unhealthy"
    }

    // Check Redis
    redisCheck := h.checkRedis(ctx)
    status.Checks["redis"] = redisCheck
    if redisCheck.Status != "ok" {
        status.Status = "unhealthy"
    }

    // Check WebSocket
    wsCheck := h.checkWebSocket()
    status.Checks["websocket"] = wsCheck
    if wsCheck.Status != "ok" {
        status.Status = "degraded"
    }

    statusCode := http.StatusOK
    if status.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(status)
}

func (h *HealthChecker) checkDatabase(ctx context.Context) Check {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    err := h.db.Ping(ctx)
    if err != nil {
        return Check{Status: "error", Message: err.Error()}
    }
    return Check{Status: "ok"}
}

func (h *HealthChecker) checkRedis(ctx context.Context) Check {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    _, err := h.redis.Ping(ctx).Result()
    if err != nil {
        return Check{Status: "error", Message: err.Error()}
    }
    return Check{Status: "ok"}
}

func (h *HealthChecker) checkWebSocket() Check {
    if !h.wsClient.IsConnected() {
        return Check{Status: "error", Message: "WebSocket disconnected"}
    }
    return Check{Status: "ok"}
}
```

---

## 7. Observability

### 7.1 Metrics Exposition

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Position metrics
    PositionCount = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "account_positions_total",
            Help: "Current number of open positions",
        },
        []string{"symbol"},
    )

    PositionValue = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "account_position_value_usd",
            Help: "Position value in USD",
        },
        []string{"symbol"},
    )

    // P&L metrics
    RealizedPnL = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "account_realized_pnl_usd",
            Help: "Total realized P&L in USD",
        },
    )

    UnrealizedPnL = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "account_unrealized_pnl_usd",
            Help: "Unrealized P&L per symbol",
        },
        []string{"symbol"},
    )

    // Balance metrics
    AccountBalance = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "account_balance",
            Help: "Account balance by asset",
        },
        []string{"asset"},
    )

    AccountEquity = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "account_equity_usd",
            Help: "Total account equity in USD",
        },
    )

    // Reconciliation metrics
    ReconciliationDrift = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "reconciliation_drift_abs",
            Help:    "Absolute drift values during reconciliation",
            Buckets: prometheus.ExponentialBuckets(0.0001, 10, 8),
        },
    )

    ReconciliationDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "reconciliation_duration_seconds",
            Help:    "Duration of reconciliation process",
            Buckets: prometheus.DefBuckets,
        },
    )

    // Alert metrics
    AlertsTriggered = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "alerts_triggered_total",
            Help: "Number of alerts triggered by type",
        },
        []string{"type", "severity"},
    )

    // WebSocket metrics
    WebSocketReconnects = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "websocket_reconnects_total",
            Help: "Number of WebSocket reconnections",
        },
    )

    WebSocketMessagesReceived = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "websocket_messages_received_total",
            Help: "Number of WebSocket messages received by type",
        },
        []string{"type"},
    )
)
```

### 7.2 Account Dashboard Requirements

**Dashboard Layout:**

```
┌─────────────────────────────────────────────────────────────┐
│  Account Monitor Dashboard                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Account Equity: $12,450.00  (+$450.00 | +3.75%)           │
│                                                             │
│  ┌─────────────────────┐  ┌─────────────────────────────┐ │
│  │  Balance Breakdown  │  │   Position Summary          │ │
│  ├─────────────────────┤  ├─────────────────────────────┤ │
│  │  USDT:   $10,000    │  │  Open Positions: 3          │ │
│  │  BTC:    $2,000     │  │  Total Value: $5,000        │ │
│  │  ETH:    $450       │  │  Unrealized P&L: +$150      │ │
│  └─────────────────────┘  └─────────────────────────────┘ │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Open Positions                                       │ │
│  ├──────┬──────┬────────┬──────────┬───────────┬────────┤ │
│  │Symbol│ Qty  │  Entry │ Current  │Unrealized │  P&L % │ │
│  ├──────┼──────┼────────┼──────────┼───────────┼────────┤ │
│  │BTCUSD│ 0.05 │ 50,000 │  51,000  │   +$50    │  +2%   │ │
│  │ETHUSD│ 2.0  │  3,000 │   3,050  │  +$100    │ +1.6%  │ │
│  │SOLUSD│-10.0 │    150 │     145  │   +$50    │ +3.3%  │ │
│  └──────┴──────┴────────┴──────────┴───────────┴────────┘ │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  P&L Chart (24h)                                      │ │
│  │                                                       │ │
│  │   500 ┤                                        ╭──    │ │
│  │       │                                   ╭────╯      │ │
│  │   250 ┤                          ╭────────╯           │ │
│  │       │                     ╭────╯                    │ │
│  │     0 ┼─────────────────────╯                         │ │
│  │       └────────────────────────────────────────────   │ │
│  │       0h     6h     12h     18h     24h               │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌─────────────────────┐  ┌─────────────────────────────┐ │
│  │  Trade Statistics   │  │   Risk Metrics              │ │
│  ├─────────────────────┤  ├─────────────────────────────┤ │
│  │  Win Rate:    65%   │  │  Margin Ratio:    45%       │ │
│  │  Total Trades: 42   │  │  Max Drawdown:   -3.2%      │ │
│  │  Avg Win:  $85.50   │  │  Sharpe Ratio:    1.8       │ │
│  │  Avg Loss: -$45.20  │  │  Current Leverage: 2.5x     │ │
│  │  Profit Factor: 1.9 │  │  Available Margin: $5,500   │ │
│  └─────────────────────┘  └─────────────────────────────┘ │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Recent Alerts                                        │ │
│  ├───────────┬─────────────────────────────────────────┤ │
│  │ 10:45 AM  │ [WARNING] Balance drift detected: USDT  │ │
│  │ 09:30 AM  │ [INFO] Reconciliation completed         │ │
│  │ 08:15 AM  │ [CRITICAL] High margin ratio: 85%       │ │
│  └───────────┴─────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**Dashboard API:**

```go
// WebSocket message format
type DashboardUpdate struct {
    Type      string      `json:"type"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

// Account state update
type AccountStateUpdate struct {
    Equity        decimal.Decimal       `json:"equity"`
    RealizedPnL   decimal.Decimal       `json:"realized_pnl"`
    UnrealizedPnL decimal.Decimal       `json:"unrealized_pnl"`
    Balances      map[string]Balance    `json:"balances"`
    Positions     []Position            `json:"positions"`
}

// P&L chart data
type PnLChartData struct {
    Interval  string                `json:"interval"`  // 1m, 5m, 1h, 1d
    DataPoints []PnLDataPoint       `json:"data_points"`
}

type PnLDataPoint struct {
    Timestamp time.Time       `json:"timestamp"`
    Value     decimal.Decimal `json:"value"`
}

// Statistics update
type StatisticsUpdate struct {
    WinRate       decimal.Decimal `json:"win_rate"`
    TotalTrades   int             `json:"total_trades"`
    WinningTrades int             `json:"winning_trades"`
    LosingTrades  int             `json:"losing_trades"`
    AverageWin    decimal.Decimal `json:"average_win"`
    AverageLoss   decimal.Decimal `json:"average_loss"`
    ProfitFactor  decimal.Decimal `json:"profit_factor"`
    SharpeRatio   decimal.Decimal `json:"sharpe_ratio"`
}
```

### 7.3 P&L Visualization

**Time-series Queries for Charts:**

```sql
-- Hourly P&L for last 24 hours
SELECT
    time_bucket('1 hour', timestamp) as hour,
    FIRST(realized_pnl, timestamp) as realized,
    FIRST(unrealized_pnl, timestamp) as unrealized,
    FIRST(total_pnl, timestamp) as total
FROM pnl_snapshots
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour ASC;

-- Daily P&L for last 30 days
SELECT
    time_bucket('1 day', timestamp) as day,
    MAX(realized_pnl) - MIN(realized_pnl) as daily_realized,
    LAST(total_pnl, timestamp) as eod_total
FROM pnl_snapshots
WHERE timestamp >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day ASC;

-- Cumulative P&L
SELECT
    timestamp,
    realized_pnl,
    SUM(realized_pnl) OVER (ORDER BY timestamp) as cumulative_pnl
FROM pnl_snapshots
WHERE timestamp >= NOW() - INTERVAL '7 days'
ORDER BY timestamp ASC;
```

---

## 8. Project Structure

```
account-monitor/
├── cmd/
│   └── server/
│       └── main.go                  # Service entry point
├── internal/
│   ├── position/
│   │   ├── manager.go               # Position tracking logic
│   │   ├── manager_test.go
│   │   └── types.go
│   ├── pnl/
│   │   ├── calculator.go            # P&L calculation
│   │   ├── calculator_test.go
│   │   └── types.go
│   ├── balance/
│   │   ├── manager.go               # Balance tracking
│   │   └── types.go
│   ├── reconciliation/
│   │   ├── reconciler.go            # Reconciliation logic
│   │   ├── reconciler_test.go
│   │   └── types.go
│   ├── alert/
│   │   ├── manager.go               # Alert evaluation
│   │   ├── manager_test.go
│   │   └── types.go
│   ├── exchange/
│   │   ├── binance.go               # Binance client
│   │   ├── websocket.go             # WebSocket handler
│   │   └── types.go
│   ├── grpc/
│   │   ├── server.go                # gRPC server
│   │   └── handlers.go
│   ├── http/
│   │   ├── server.go                # HTTP server
│   │   ├── dashboard.go             # Dashboard API
│   │   └── health.go
│   ├── storage/
│   │   ├── postgres.go              # PostgreSQL client
│   │   ├── redis.go                 # Redis client
│   │   └── migrations/
│   │       └── 001_initial.sql
│   └── metrics/
│       └── prometheus.go            # Metrics definitions
├── pkg/
│   └── proto/
│       └── account_monitor.proto    # Protocol Buffers
├── ui/
│   ├── dashboard/
│   │   ├── index.html
│   │   ├── app.js
│   │   └── styles.css
│   └── assets/
├── config/
│   ├── config.yaml                  # Default config
│   └── config.example.yaml
├── scripts/
│   ├── init_db.sh
│   └── test_reconciliation.sh
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 9. Configuration File

```yaml
# config.yaml
service:
  name: account-monitor
  version: 1.0.0

grpc:
  port: 50051
  max_connections: 100

http:
  port: 8080
  dashboard_enabled: true

metrics:
  port: 9090
  path: /metrics

exchange:
  name: binance
  testnet: true
  api_key_env: BINANCE_API_KEY
  secret_key_env: BINANCE_SECRET_KEY
  websocket:
    reconnect_interval: 5s
    ping_interval: 30s

database:
  postgres:
    host: localhost
    port: 5432
    database: trading
    user: trading
    password_env: POSTGRES_PASSWORD
    max_connections: 10

  redis:
    host: localhost
    port: 6379
    db: 0

reconciliation:
  enabled: true
  interval: 60s
  balance_tolerance: 0.00001
  position_tolerance: 0.0001

alerts:
  enabled: true
  thresholds:
    min_balance: 1000.0
    max_drawdown_pct: -5.0
    max_margin_ratio: 0.8
    balance_drift_pct: 1.0
    position_drift_pct: 1.0
  suppression_duration: 60s

pubsub:
  provider: redis  # redis, nats, kafka
  topics:
    alerts: "trading.alerts"
    pnl_updates: "trading.pnl"

logging:
  level: info  # debug, info, warn, error
  format: json
  output: stdout
```

---

## 10. Development Timeline

**Total Estimated Duration: 4-5 weeks**

| Week | Phase | Deliverables |
|------|-------|--------------|
| 1 | Phase 1-2 | gRPC server + WebSocket client |
| 2 | Phase 3 | Position & balance tracking |
| 3 | Phase 4-5 | P&L calculation + reconciliation |
| 4 | Phase 6-7 | Alerts + observability UI |
| 5 | Testing & deployment | Production readiness |

---

## 11. Success Criteria

- [ ] All gRPC queries return accurate data within 50ms
- [ ] WebSocket maintains connection with < 0.1% downtime
- [ ] Position tracking handles all state transitions correctly
- [ ] P&L calculations match manual verification (100% accuracy)
- [ ] Reconciliation detects drifts within 60 seconds
- [ ] Alert system triggers within 1 second of threshold breach
- [ ] Dashboard updates in real-time (< 100ms latency)
- [ ] Race detector passes on all concurrent operations
- [ ] Test coverage > 80%
- [ ] Service passes 24-hour stability test under load

---

**End of Development Plan**
