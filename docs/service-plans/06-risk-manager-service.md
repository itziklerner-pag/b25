# Risk Manager Service - Development Plan

**Service:** Risk Manager Service
**Version:** 1.0
**Created:** 2025-10-02
**Status:** Planning Phase

---

## Table of Contents

1. [Overview](#overview)
2. [Technology Stack Recommendation](#technology-stack-recommendation)
3. [Architecture Design](#architecture-design)
4. [Development Phases](#development-phases)
5. [Implementation Details](#implementation-details)
6. [Testing Strategy](#testing-strategy)
7. [Deployment](#deployment)
8. [Observability](#observability)
9. [Performance Targets](#performance-targets)
10. [Risk Formulas Reference](#risk-formulas-reference)

---

## Overview

### Purpose

The Risk Manager Service is the global risk management and emergency control system for the trading platform. It provides:

- Real-time risk metric calculation and monitoring
- Multi-layer limit enforcement (position size, leverage, drawdown)
- Emergency stop mechanism with immediate position liquidation
- Pre-trade risk checks as a service for the execution engine
- Risk policy configuration and versioning

### Key Responsibilities

1. **Risk Calculation Engine:** Compute margin ratio, leverage, drawdown, position concentration
2. **Policy Management:** Store and enforce risk limits across multiple dimensions
3. **Pre-Trade Validation:** Synchronous RPC service for order validation before submission
4. **Real-Time Monitoring:** Continuous position and account monitoring with alerting
5. **Emergency Controls:** Circuit breaker and emergency stop with position unwinding

### Critical Requirements

- **Low Latency:** Pre-trade checks must complete in <10ms (p99)
- **High Availability:** 99.9% uptime with automatic failover
- **Accuracy:** Risk calculations must be precise to prevent false positives
- **Auditability:** All risk policy changes and violations must be logged
- **Safety:** Emergency stop must be reliable and fast

---

## Technology Stack Recommendation

### Primary Language: **Go**

**Rationale:**
- Excellent performance for real-time calculations (better than Python, easier than Rust)
- Strong concurrency primitives (goroutines) for parallel risk checks
- Fast gRPC implementation for pre-trade check API
- Easy deployment and monitoring
- Good balance between development speed and performance

**Alternative:** Rust for maximum performance if <5ms p99 latency required

### Core Technologies

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Runtime** | Go 1.22+ | Performance, concurrency, type safety |
| **RPC Framework** | gRPC + protobuf | Low-latency pre-trade checks, type safety |
| **Rule Engine** | Custom Go-based | Simple, fast, auditable (avoid Drools complexity) |
| **Database** | PostgreSQL 16+ | ACID compliance for policies, JSON support |
| **Cache Layer** | Redis 7+ | Sub-ms policy lookups, atomic operations |
| **Pub/Sub** | NATS or Redis | Alert distribution, emergency stop broadcast |
| **Config Management** | Viper + YAML/JSON | Hot-reload support, environment overrides |
| **Metrics** | Prometheus client | Standard observability integration |
| **Logging** | Zap (structured) | High-performance structured logging |

### Testing Frameworks

| Type | Framework | Purpose |
|------|-----------|---------|
| **Unit Testing** | Go testing + testify | Risk calculation verification |
| **Property Testing** | gopter | Risk formula edge cases |
| **Integration Testing** | testcontainers-go | Database and cache testing |
| **Load Testing** | k6 | Pre-trade check latency validation |
| **Chaos Testing** | toxiproxy | Emergency stop reliability |

### Development Tools

- **Protocol Buffers:** protoc + protoc-gen-go-grpc for API contracts
- **Database Migrations:** golang-migrate for schema versioning
- **Code Generation:** go generate for risk rule templates
- **Linting:** golangci-lint with strict rules
- **Dependency Management:** Go modules with vendoring

---

## Architecture Design

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

### Component Details

#### 1. Risk Calculation Engine

**Purpose:** Core computation engine for all risk metrics

**Inputs:**
- Account state (from Account Monitor Service via gRPC)
- Current positions and pending orders
- Market prices (from Market Data Service)
- Risk policies (from database/cache)

**Outputs:**
- Calculated risk metrics (margin ratio, leverage, etc.)
- Limit utilization percentages
- Risk score aggregates

**Key Algorithms:**
```go
// Margin Ratio Calculation
marginRatio = totalEquity / totalRequiredMargin

// Leverage Calculation
leverage = totalPositionNotional / totalEquity

// Drawdown Calculation
drawdown = (peakEquity - currentEquity) / peakEquity

// Position Concentration
concentration = singleSymbolNotional / totalPortfolioValue
```

**Performance:**
- Cache market prices (update every 100ms)
- Batch position queries
- Pre-compute intermediate values
- Target: <5ms per calculation

#### 2. Policy Enforcement Engine

**Purpose:** Rule evaluation and limit enforcement

**Policy Hierarchy:**
```
Global Risk Limits
  ├─ Account-Level Limits
  │   ├─ Max Leverage: 10x
  │   ├─ Max Drawdown: 20%
  │   └─ Daily Loss Limit: $10,000
  │
  ├─ Symbol-Level Limits
  │   ├─ Max Position Size: $100,000
  │   ├─ Max Concentration: 30%
  │   └─ Max Orders Per Minute: 20
  │
  └─ Strategy-Level Limits
      ├─ Per-Strategy Max Allocation: $50,000
      └─ Per-Strategy Max Leverage: 5x
```

**Rule Types:**
- **Hard Limits:** Block order submission (pre-trade checks)
- **Soft Limits:** Warn but allow (with logging)
- **Emergency Limits:** Trigger emergency stop
- **Conditional Rules:** Time-based or event-based

**Policy Schema:**
```yaml
risk_policies:
  - id: "account-leverage-limit"
    type: "hard"
    metric: "leverage"
    operator: "less_than_or_equal"
    threshold: 10.0
    scope: "account"
    enabled: true

  - id: "btc-position-limit"
    type: "hard"
    metric: "position_notional"
    operator: "less_than_or_equal"
    threshold: 100000.0
    scope: "symbol"
    symbol: "BTCUSDT"
    enabled: true

  - id: "daily-drawdown-emergency"
    type: "emergency"
    metric: "drawdown_daily"
    operator: "greater_than"
    threshold: 0.20
    scope: "account"
    action: "emergency_stop"
    enabled: true
```

#### 3. Pre-Trade Risk Check API

**Purpose:** Synchronous validation service for Order Execution Service

**gRPC Interface:**
```protobuf
service RiskManager {
  // Pre-trade validation (must be fast: <10ms p99)
  rpc CheckOrder(OrderRiskRequest) returns (OrderRiskResponse);

  // Batch validation for multiple orders
  rpc CheckOrderBatch(BatchOrderRiskRequest) returns (BatchOrderRiskResponse);

  // Get current risk metrics
  rpc GetRiskMetrics(RiskMetricsRequest) returns (RiskMetricsResponse);

  // Emergency stop trigger
  rpc TriggerEmergencyStop(EmergencyStopRequest) returns (EmergencyStopResponse);
}

message OrderRiskRequest {
  string order_id = 1;
  string symbol = 2;
  string side = 3;        // BUY or SELL
  double quantity = 4;
  double price = 5;       // 0 for market orders
  string order_type = 6;  // MARKET, LIMIT
  string strategy_id = 7;
  int64 timestamp = 8;
}

message OrderRiskResponse {
  bool approved = 1;
  repeated string violations = 2;  // Empty if approved
  RiskMetrics post_trade_metrics = 3;
  int64 processing_time_us = 4;
}

message RiskMetrics {
  double margin_ratio = 1;
  double leverage = 2;
  double drawdown = 3;
  double daily_pnl = 4;
  map<string, double> position_concentration = 5;
  map<string, double> limit_utilization = 6;
}
```

**Validation Flow:**
```
Order Request
    ↓
Query Current Account State (cached)
    ↓
Simulate Post-Trade State
    ↓
Calculate Post-Trade Risk Metrics
    ↓
Evaluate All Applicable Policies
    ↓
    ├─ All Pass → Return Approved
    └─ Any Fail → Return Rejected with Violations
```

#### 4. Emergency Stop Mechanism

**Purpose:** Immediate system shutdown and position unwinding

**Trigger Conditions:**
- Drawdown exceeds emergency threshold (e.g., 25%)
- Manual trigger via admin API
- External system failure detected
- Repeated risk violations within time window

**Emergency Stop Workflow:**
```
Trigger Event
    ↓
Lock All Order Submissions (atomic flag in Redis)
    ↓
Broadcast Emergency Stop Alert (NATS pub/sub)
    ↓
Cancel All Open Orders (parallel API calls)
    ↓
Calculate Position Unwind Orders (prioritize by risk)
    ↓
Submit Market Orders to Close Positions
    ↓
Monitor Fills Until All Positions Closed
    ↓
Log Complete Audit Trail
    ↓
Require Manual Re-Enable
```

**Safety Mechanisms:**
- Distributed lock to prevent duplicate emergency stops
- Idempotent operations (safe to retry)
- Complete audit logging
- Manual confirmation required to re-enable trading

#### 5. Real-Time Monitoring

**Purpose:** Continuous position and risk monitoring

**Monitoring Loop:**
```go
for {
    // Every 1 second
    accountState := fetchAccountState()
    positions := fetchCurrentPositions()
    prices := fetchMarketPrices()

    metrics := calculateRiskMetrics(accountState, positions, prices)
    violations := evaluatePolicies(metrics)

    if len(violations) > 0 {
        publishAlerts(violations)
        logViolations(violations)

        if hasEmergencyViolation(violations) {
            triggerEmergencyStop()
        }
    }

    updateMetricsDashboard(metrics)
    sleep(1 * time.Second)
}
```

**Alert Channels:**
- Pub/sub for real-time dashboards
- Webhook for external notifications
- Database for audit trail
- Emergency stop trigger for critical violations

---

## Development Phases

### Phase 1: Risk Calculation Engine (Weeks 1-2)

**Goal:** Build core risk metric calculation library

**Deliverables:**
- [ ] Risk calculation functions (margin, leverage, drawdown)
- [ ] Position aggregation logic
- [ ] Mark price integration
- [ ] Unit tests with property-based testing
- [ ] Benchmark suite

**Dependencies:**
- Account Monitor Service API contract
- Market Data Service API contract

**Acceptance Criteria:**
- All risk formulas validated against manual calculations
- 100% test coverage on calculation logic
- Calculation performance: <1ms for single account
- Property tests pass 10,000 iterations

### Phase 2: Policy Management System (Weeks 2-3)

**Goal:** Implement policy storage and evaluation engine

**Deliverables:**
- [ ] PostgreSQL schema for risk policies
- [ ] Policy CRUD API (internal gRPC)
- [ ] Policy evaluation engine
- [ ] Policy versioning and audit trail
- [ ] Redis caching layer
- [ ] Integration tests

**Schema Design:**
```sql
CREATE TABLE risk_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- hard, soft, emergency
    metric VARCHAR(100) NOT NULL,
    operator VARCHAR(50) NOT NULL,
    threshold NUMERIC(20, 8) NOT NULL,
    scope VARCHAR(50) NOT NULL, -- account, symbol, strategy
    scope_id VARCHAR(100), -- symbol name or strategy ID
    action VARCHAR(50), -- block, warn, emergency_stop
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

CREATE TABLE risk_violations (
    id BIGSERIAL PRIMARY KEY,
    policy_id UUID REFERENCES risk_policies(id),
    violation_time TIMESTAMPTZ DEFAULT NOW(),
    metric_value NUMERIC(20, 8),
    threshold_value NUMERIC(20, 8),
    context JSONB, -- account state, order details, etc.
    action_taken VARCHAR(100),
    resolved BOOLEAN DEFAULT false
);

CREATE TABLE emergency_stops (
    id BIGSERIAL PRIMARY KEY,
    trigger_time TIMESTAMPTZ DEFAULT NOW(),
    trigger_reason TEXT,
    triggered_by VARCHAR(100), -- manual, auto, policy_id
    account_state JSONB,
    positions_snapshot JSONB,
    orders_canceled INTEGER,
    positions_closed INTEGER,
    completed_at TIMESTAMPTZ,
    re_enabled_at TIMESTAMPTZ,
    re_enabled_by VARCHAR(100)
);

CREATE INDEX idx_policies_enabled ON risk_policies(enabled);
CREATE INDEX idx_policies_scope ON risk_policies(scope, scope_id);
CREATE INDEX idx_violations_time ON risk_violations(violation_time);
CREATE INDEX idx_violations_policy ON risk_violations(policy_id);
```

**Acceptance Criteria:**
- Policies stored and retrieved correctly
- Cache invalidation on policy updates
- Policy versioning tracks all changes
- Evaluation engine handles all operators correctly

### Phase 3: Pre-Trade Check RPC Service (Weeks 3-4)

**Goal:** Build fast gRPC service for order validation

**Deliverables:**
- [ ] gRPC server implementation
- [ ] Order simulation logic
- [ ] Multi-layer validation flow
- [ ] Request/response logging
- [ ] Load testing suite
- [ ] Client SDK generation

**Implementation:**
```go
// Core validation function
func (s *RiskServer) CheckOrder(
    ctx context.Context,
    req *pb.OrderRiskRequest,
) (*pb.OrderRiskResponse, error) {
    startTime := time.Now()

    // 1. Fetch current account state (cached)
    accountState, err := s.accountClient.GetAccountState(ctx)
    if err != nil {
        return nil, status.Errorf(codes.Unavailable, "account state unavailable: %v", err)
    }

    // 2. Fetch applicable policies (cached)
    policies := s.policyCache.GetApplicablePolicies(req.Symbol, req.StrategyId)

    // 3. Simulate post-trade state
    postTradeState := s.simulateOrder(accountState, req)

    // 4. Calculate post-trade risk metrics
    metrics := s.calculateRiskMetrics(postTradeState)

    // 5. Evaluate all policies
    violations := s.evaluatePolicies(metrics, policies)

    // 6. Build response
    approved := len(violations) == 0
    processingTime := time.Since(startTime).Microseconds()

    // 7. Log for audit
    s.logOrderCheck(req, approved, violations, processingTime)

    return &pb.OrderRiskResponse{
        Approved:          approved,
        Violations:        violations,
        PostTradeMetrics:  metrics,
        ProcessingTimeUs:  processingTime,
    }, nil
}
```

**Performance Optimizations:**
- Cache account state (100ms TTL)
- Cache policies (1s TTL with pub/sub invalidation)
- Cache market prices (100ms TTL)
- Connection pooling for database
- gRPC keepalive for persistent connections

**Acceptance Criteria:**
- p99 latency < 10ms under normal load
- p99 latency < 50ms under 10x load
- No false positives in validation
- 100% uptime during load tests
- Graceful degradation on dependency failures

### Phase 4: Real-Time Monitoring and Alerts (Weeks 4-5)

**Goal:** Continuous risk monitoring with alerting

**Deliverables:**
- [ ] Background monitoring loop
- [ ] Alert generation and routing
- [ ] Violation logging
- [ ] Dashboard metrics publishing
- [ ] Alert aggregation (prevent spam)

**Monitoring Implementation:**
```go
type RiskMonitor struct {
    accountClient  AccountServiceClient
    marketClient   MarketDataClient
    policyEngine   PolicyEngine
    alertPublisher AlertPublisher
    metrics        MetricsCollector
}

func (m *RiskMonitor) Run(ctx context.Context) error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := m.checkRisk(ctx); err != nil {
                log.Error("risk check failed", zap.Error(err))
                m.metrics.IncRiskCheckErrors()
            }

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (m *RiskMonitor) checkRisk(ctx context.Context) error {
    // Fetch current state
    accountState, err := m.accountClient.GetAccountState(ctx)
    if err != nil {
        return fmt.Errorf("fetch account state: %w", err)
    }

    // Calculate risk metrics
    metrics := m.calculateRiskMetrics(accountState)

    // Evaluate all active policies
    violations := m.policyEngine.EvaluateAll(metrics)

    // Publish metrics to dashboard
    m.publishMetrics(metrics)

    // Handle violations
    if len(violations) > 0 {
        m.handleViolations(violations, accountState)
    }

    return nil
}

func (m *RiskMonitor) handleViolations(
    violations []PolicyViolation,
    accountState AccountState,
) {
    // Group violations by severity
    var emergencyViolations []PolicyViolation
    var hardViolations []PolicyViolation
    var softViolations []PolicyViolation

    for _, v := range violations {
        switch v.Policy.Type {
        case PolicyTypeEmergency:
            emergencyViolations = append(emergencyViolations, v)
        case PolicyTypeHard:
            hardViolations = append(hardViolations, v)
        case PolicyTypeSoft:
            softViolations = append(softViolations, v)
        }
    }

    // Trigger emergency stop if needed
    if len(emergencyViolations) > 0 {
        m.triggerEmergencyStop(emergencyViolations, accountState)
        return
    }

    // Publish alerts for hard and soft violations
    for _, v := range hardViolations {
        m.alertPublisher.PublishAlert(AlertLevelCritical, v)
    }
    for _, v := range softViolations {
        m.alertPublisher.PublishAlert(AlertLevelWarning, v)
    }

    // Log all violations
    m.logViolations(violations)
}
```

**Alert Deduplication:**
```go
type AlertDeduplicator struct {
    recentAlerts map[string]time.Time
    mutex        sync.RWMutex
    window       time.Duration
}

func (d *AlertDeduplicator) ShouldAlert(alertKey string) bool {
    d.mutex.RLock()
    lastSent, exists := d.recentAlerts[alertKey]
    d.mutex.RUnlock()

    if !exists || time.Since(lastSent) > d.window {
        d.mutex.Lock()
        d.recentAlerts[alertKey] = time.Now()
        d.mutex.Unlock()
        return true
    }

    return false
}
```

**Acceptance Criteria:**
- Violations detected within 2 seconds
- Alerts published within 100ms of detection
- No duplicate alerts within 5-minute window
- All violations logged to database
- Metrics dashboard updates in real-time

### Phase 5: Emergency Stop Mechanism (Week 5)

**Goal:** Reliable emergency shutdown and position unwinding

**Deliverables:**
- [ ] Emergency stop trigger logic
- [ ] Order cancellation workflow
- [ ] Position unwinding algorithm
- [ ] Distributed locking mechanism
- [ ] Complete audit logging
- [ ] Manual re-enable flow

**Emergency Stop Implementation:**
```go
type EmergencyStop struct {
    orderClient    OrderExecutionClient
    accountClient  AccountServiceClient
    lockManager    DistributedLockManager
    auditLogger    AuditLogger
    alertPublisher AlertPublisher
}

func (e *EmergencyStop) Trigger(
    ctx context.Context,
    reason string,
    triggeredBy string,
) error {
    // Acquire distributed lock to prevent duplicate triggers
    lock, err := e.lockManager.AcquireLock(ctx, "emergency_stop", 5*time.Minute)
    if err != nil {
        return fmt.Errorf("acquire emergency stop lock: %w", err)
    }
    defer lock.Release()

    log.Warn("EMERGENCY STOP TRIGGERED",
        zap.String("reason", reason),
        zap.String("triggered_by", triggeredBy))

    // Step 1: Block all new order submissions
    if err := e.setTradingDisabled(ctx); err != nil {
        return fmt.Errorf("disable trading: %w", err)
    }

    // Step 2: Broadcast emergency stop alert
    e.alertPublisher.PublishEmergencyStopAlert(reason)

    // Step 3: Capture current state snapshot
    accountState, err := e.accountClient.GetAccountState(ctx)
    if err != nil {
        return fmt.Errorf("capture account state: %w", err)
    }

    // Step 4: Cancel all open orders
    canceledCount, err := e.cancelAllOrders(ctx)
    if err != nil {
        log.Error("failed to cancel all orders", zap.Error(err))
    }

    // Step 5: Close all positions
    closedCount, err := e.closeAllPositions(ctx, accountState.Positions)
    if err != nil {
        return fmt.Errorf("close positions: %w", err)
    }

    // Step 6: Log complete audit trail
    e.auditLogger.LogEmergencyStop(EmergencyStopEvent{
        TriggerTime:    time.Now(),
        Reason:         reason,
        TriggeredBy:    triggeredBy,
        AccountState:   accountState,
        OrdersCanceled: canceledCount,
        PositionsClosed: closedCount,
    })

    log.Warn("EMERGENCY STOP COMPLETED",
        zap.Int("orders_canceled", canceledCount),
        zap.Int("positions_closed", closedCount))

    return nil
}

func (e *EmergencyStop) cancelAllOrders(ctx context.Context) (int, error) {
    openOrders, err := e.orderClient.GetOpenOrders(ctx)
    if err != nil {
        return 0, err
    }

    var wg sync.WaitGroup
    canceledCount := atomic.Int32{}
    errChan := make(chan error, len(openOrders))

    // Cancel orders in parallel
    for _, order := range openOrders {
        wg.Add(1)
        go func(orderID string) {
            defer wg.Done()

            if err := e.orderClient.CancelOrder(ctx, orderID); err != nil {
                errChan <- fmt.Errorf("cancel order %s: %w", orderID, err)
            } else {
                canceledCount.Add(1)
            }
        }(order.ID)
    }

    wg.Wait()
    close(errChan)

    // Collect errors
    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return int(canceledCount.Load()), fmt.Errorf("cancel errors: %v", errs)
    }

    return int(canceledCount.Load()), nil
}

func (e *EmergencyStop) closeAllPositions(
    ctx context.Context,
    positions []Position,
) (int, error) {
    // Sort by risk priority (largest positions first)
    sort.Slice(positions, func(i, j int) bool {
        return math.Abs(positions[i].Notional) > math.Abs(positions[j].Notional)
    })

    closedCount := 0

    // Close positions sequentially to avoid overwhelming exchange
    for _, pos := range positions {
        if pos.Quantity == 0 {
            continue
        }

        // Create market order to close position
        closeOrder := &Order{
            Symbol:   pos.Symbol,
            Side:     reversePositionSide(pos.Side),
            Quantity: math.Abs(pos.Quantity),
            Type:     "MARKET",
            Reason:   "emergency_stop",
        }

        if err := e.orderClient.SubmitOrder(ctx, closeOrder); err != nil {
            log.Error("failed to close position",
                zap.String("symbol", pos.Symbol),
                zap.Error(err))
            continue
        }

        closedCount++

        // Small delay to avoid rate limits
        time.Sleep(100 * time.Millisecond)
    }

    return closedCount, nil
}
```

**Acceptance Criteria:**
- Emergency stop completes within 30 seconds
- All open orders canceled successfully
- All positions closed or flagged for manual intervention
- Complete audit trail in database
- Idempotent (safe to retry)
- Trading re-enable requires manual confirmation

### Phase 6: Configuration UI and Testing (Week 6)

**Goal:** Admin interface and comprehensive testing

**Deliverables:**
- [ ] Risk policy configuration API
- [ ] Admin UI for policy management (optional)
- [ ] Policy validation and testing tools
- [ ] End-to-end integration tests
- [ ] Chaos engineering tests
- [ ] Load testing and performance tuning
- [ ] Documentation

**Admin API:**
```protobuf
service RiskAdmin {
  // Policy management
  rpc CreatePolicy(CreatePolicyRequest) returns (Policy);
  rpc UpdatePolicy(UpdatePolicyRequest) returns (Policy);
  rpc DeletePolicy(DeletePolicyRequest) returns (Empty);
  rpc ListPolicies(ListPoliciesRequest) returns (ListPoliciesResponse);

  // Emergency stop management
  rpc GetEmergencyStopStatus(Empty) returns (EmergencyStopStatus);
  rpc EnableTrading(EnableTradingRequest) returns (Empty);

  // Testing and validation
  rpc TestPolicy(TestPolicyRequest) returns (TestPolicyResponse);
  rpc SimulateRiskCheck(SimulateRiskCheckRequest) returns (OrderRiskResponse);
}
```

**Acceptance Criteria:**
- All CRUD operations work correctly
- Policy validation prevents invalid configurations
- Emergency stop status queryable
- Trading re-enable flow works
- Complete integration test coverage
- Load tests pass performance targets

---

## Implementation Details

### Risk Metric Calculations

#### 1. Margin Ratio

**Formula:**
```
Margin Ratio = Total Equity / Total Required Margin
```

**Implementation:**
```go
func CalculateMarginRatio(accountState AccountState) float64 {
    totalEquity := accountState.Balance + accountState.UnrealizedPnL
    totalRequiredMargin := 0.0

    for _, pos := range accountState.Positions {
        // Initial margin requirement (e.g., 1/leverage)
        requiredMargin := math.Abs(pos.Notional) / pos.Leverage
        totalRequiredMargin += requiredMargin
    }

    if totalRequiredMargin == 0 {
        return math.Inf(1) // Infinite margin ratio (no positions)
    }

    return totalEquity / totalRequiredMargin
}
```

**Risk Threshold:**
- Healthy: > 2.0
- Warning: 1.5 - 2.0
- Critical: 1.0 - 1.5
- Liquidation Risk: < 1.0

#### 2. Account Leverage

**Formula:**
```
Leverage = Total Position Notional / Total Equity
```

**Implementation:**
```go
func CalculateLeverage(accountState AccountState) float64 {
    totalEquity := accountState.Balance + accountState.UnrealizedPnL
    totalNotional := 0.0

    for _, pos := range accountState.Positions {
        totalNotional += math.Abs(pos.Notional)
    }

    if totalEquity == 0 {
        return 0.0
    }

    return totalNotional / totalEquity
}
```

**Risk Threshold:**
- Conservative: < 3x
- Moderate: 3x - 5x
- Aggressive: 5x - 10x
- Extreme: > 10x

#### 3. Drawdown (Multiple Types)

**Daily Drawdown:**
```
Daily DD = (Day Start Equity - Current Equity) / Day Start Equity
```

**Peak Drawdown:**
```
Peak DD = (Peak Equity - Current Equity) / Peak Equity
```

**Implementation:**
```go
type DrawdownCalculator struct {
    dayStartEquity float64
    peakEquity     float64
    mutex          sync.RWMutex
}

func (d *DrawdownCalculator) UpdateEquity(currentEquity float64) {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    if currentEquity > d.peakEquity {
        d.peakEquity = currentEquity
    }
}

func (d *DrawdownCalculator) ResetDaily(startEquity float64) {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    d.dayStartEquity = startEquity
}

func (d *DrawdownCalculator) GetDailyDrawdown(currentEquity float64) float64 {
    d.mutex.RLock()
    defer d.mutex.RUnlock()

    if d.dayStartEquity == 0 {
        return 0.0
    }

    return (d.dayStartEquity - currentEquity) / d.dayStartEquity
}

func (d *DrawdownCalculator) GetPeakDrawdown(currentEquity float64) float64 {
    d.mutex.RLock()
    defer d.mutex.RUnlock()

    if d.peakEquity == 0 {
        return 0.0
    }

    return (d.peakEquity - currentEquity) / d.peakEquity
}
```

**Risk Threshold:**
- Acceptable: < 10%
- Warning: 10% - 20%
- Critical: 20% - 25%
- Emergency: > 25%

#### 4. Position Concentration

**Formula:**
```
Concentration = Single Position Notional / Total Portfolio Value
```

**Implementation:**
```go
func CalculatePositionConcentration(
    accountState AccountState,
) map[string]float64 {
    totalEquity := accountState.Balance + accountState.UnrealizedPnL
    concentrations := make(map[string]float64)

    for _, pos := range accountState.Positions {
        notional := math.Abs(pos.Notional)
        concentration := notional / totalEquity
        concentrations[pos.Symbol] = concentration
    }

    return concentrations
}
```

**Risk Threshold:**
- Diversified: < 20% per position
- Concentrated: 20% - 40% per position
- High Risk: > 40% per position

#### 5. Order Velocity Limits

**Purpose:** Prevent erratic trading behavior

**Implementation:**
```go
type VelocityLimiter struct {
    orderCounts map[string]*TokenBucket
    mutex       sync.RWMutex
}

type TokenBucket struct {
    tokens       int
    maxTokens    int
    refillRate   int           // tokens per second
    lastRefill   time.Time
    mutex        sync.Mutex
}

func (v *VelocityLimiter) CheckOrderRate(
    symbol string,
    maxOrdersPerMinute int,
) bool {
    v.mutex.RLock()
    bucket, exists := v.orderCounts[symbol]
    v.mutex.RUnlock()

    if !exists {
        v.mutex.Lock()
        bucket = &TokenBucket{
            tokens:     maxOrdersPerMinute,
            maxTokens:  maxOrdersPerMinute,
            refillRate: maxOrdersPerMinute / 60,
            lastRefill: time.Now(),
        }
        v.orderCounts[symbol] = bucket
        v.mutex.Unlock()
    }

    return bucket.TryConsume()
}

func (t *TokenBucket) TryConsume() bool {
    t.mutex.Lock()
    defer t.mutex.Unlock()

    // Refill tokens based on elapsed time
    elapsed := time.Since(t.lastRefill)
    tokensToAdd := int(elapsed.Seconds()) * t.refillRate

    if tokensToAdd > 0 {
        t.tokens = min(t.tokens+tokensToAdd, t.maxTokens)
        t.lastRefill = time.Now()
    }

    // Try to consume token
    if t.tokens > 0 {
        t.tokens--
        return true
    }

    return false
}
```

### Policy Enforcement Examples

#### Example 1: Simple Leverage Limit

```yaml
policy:
  id: "max-leverage-10x"
  name: "Maximum Account Leverage"
  type: "hard"
  metric: "leverage"
  operator: "less_than_or_equal"
  threshold: 10.0
  scope: "account"
  enabled: true
```

**Evaluation:**
```go
func (e *PolicyEngine) evaluateLeveragePolicy(
    policy Policy,
    metrics RiskMetrics,
) *PolicyViolation {
    currentLeverage := metrics.Leverage

    switch policy.Operator {
    case "less_than_or_equal":
        if currentLeverage > policy.Threshold {
            return &PolicyViolation{
                PolicyID:    policy.ID,
                PolicyName:  policy.Name,
                MetricValue: currentLeverage,
                Threshold:   policy.Threshold,
                Message:     fmt.Sprintf("Leverage %.2f exceeds limit %.2f",
                    currentLeverage, policy.Threshold),
            }
        }
    }

    return nil
}
```

#### Example 2: Symbol-Specific Position Limit

```yaml
policy:
  id: "btc-position-limit"
  name: "BTC Maximum Position Size"
  type: "hard"
  metric: "position_notional"
  operator: "less_than_or_equal"
  threshold: 100000.0
  scope: "symbol"
  scope_id: "BTCUSDT"
  enabled: true
```

#### Example 3: Time-Based Trading Restriction

```yaml
policy:
  id: "no-trading-outside-hours"
  name: "Trading Hours Restriction"
  type: "hard"
  metric: "time_of_day"
  operator: "between"
  threshold_min: "09:30"
  threshold_max: "16:00"
  timezone: "America/New_York"
  scope: "account"
  enabled: true
```

#### Example 4: Conditional Drawdown Policy

```yaml
policy:
  id: "progressive-drawdown-limit"
  name: "Progressive Drawdown Limit"
  type: "hard"
  metric: "drawdown_daily"
  rules:
    - condition: "account_size < 10000"
      threshold: 0.05  # 5% limit for small accounts
    - condition: "account_size >= 10000"
      threshold: 0.10  # 10% limit for larger accounts
  scope: "account"
  enabled: true
```

### Pre-Trade Simulation

**Purpose:** Calculate post-trade state without executing

```go
func (s *RiskServer) simulateOrder(
    currentState AccountState,
    order *OrderRiskRequest,
) AccountState {
    simulatedState := currentState.Clone()

    // Calculate order notional
    orderNotional := order.Quantity * order.Price
    if order.OrderType == "MARKET" {
        // Use current market price
        orderNotional = order.Quantity * s.getCurrentPrice(order.Symbol)
    }

    // Update balance (assume order fills immediately)
    if order.Side == "BUY" {
        simulatedState.Balance -= orderNotional
    } else {
        simulatedState.Balance += orderNotional
    }

    // Update or create position
    position := simulatedState.GetPosition(order.Symbol)
    if position == nil {
        // New position
        simulatedState.Positions = append(simulatedState.Positions, Position{
            Symbol:   order.Symbol,
            Side:     order.Side,
            Quantity: order.Quantity,
            EntryPrice: order.Price,
            Notional: orderNotional,
        })
    } else {
        // Update existing position
        if position.Side == order.Side {
            // Add to position
            totalQuantity := position.Quantity + order.Quantity
            avgPrice := (position.EntryPrice*position.Quantity + order.Price*order.Quantity) / totalQuantity
            position.Quantity = totalQuantity
            position.EntryPrice = avgPrice
            position.Notional += orderNotional
        } else {
            // Reduce or reverse position
            if order.Quantity >= position.Quantity {
                // Reverse position
                position.Side = order.Side
                position.Quantity = order.Quantity - position.Quantity
                position.EntryPrice = order.Price
            } else {
                // Reduce position
                position.Quantity -= order.Quantity
            }
        }
    }

    return simulatedState
}
```

---

## Testing Strategy

### Unit Testing

**Risk Calculation Tests:**
```go
func TestMarginRatioCalculation(t *testing.T) {
    tests := []struct {
        name           string
        accountState   AccountState
        expectedRatio  float64
    }{
        {
            name: "no positions",
            accountState: AccountState{
                Balance:       10000,
                UnrealizedPnL: 0,
                Positions:     []Position{},
            },
            expectedRatio: math.Inf(1),
        },
        {
            name: "single position 2x leverage",
            accountState: AccountState{
                Balance:       10000,
                UnrealizedPnL: 500,
                Positions: []Position{
                    {
                        Symbol:   "BTCUSDT",
                        Notional: 20000,
                        Leverage: 2.0,
                    },
                },
            },
            expectedRatio: 1.05, // 10500 / 10000
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ratio := CalculateMarginRatio(tt.accountState)
            assert.InDelta(t, tt.expectedRatio, ratio, 0.01)
        })
    }
}
```

**Property-Based Testing:**
```go
func TestLeverageProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("leverage is always non-negative", prop.ForAll(
        func(balance, notional float64) bool {
            if balance <= 0 {
                return true // skip invalid inputs
            }

            accountState := AccountState{
                Balance: balance,
                Positions: []Position{
                    {Notional: notional},
                },
            }

            leverage := CalculateLeverage(accountState)
            return leverage >= 0
        },
        gen.Float64Range(1, 1000000),
        gen.Float64Range(0, 10000000),
    ))

    properties.Property("leverage increases with notional", prop.ForAll(
        func(balance, notional1, notional2 float64) bool {
            if balance <= 0 || notional2 <= notional1 {
                return true
            }

            state1 := AccountState{
                Balance:   balance,
                Positions: []Position{{Notional: notional1}},
            }
            state2 := AccountState{
                Balance:   balance,
                Positions: []Position{{Notional: notional2}},
            }

            lev1 := CalculateLeverage(state1)
            lev2 := CalculateLeverage(state2)

            return lev2 >= lev1
        },
        gen.Float64Range(1000, 100000),
        gen.Float64Range(100, 50000),
        gen.Float64Range(100, 100000),
    ))

    properties.TestingRun(t)
}
```

### Integration Testing

**Database Integration:**
```go
func TestPolicyStorage(t *testing.T) {
    // Use testcontainers for real PostgreSQL
    ctx := context.Background()

    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:16",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_DB":       "risk_test",
                "POSTGRES_PASSWORD": "test",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)

    // Get connection string
    host, _ := postgres.Host(ctx)
    port, _ := postgres.MappedPort(ctx, "5432")
    connStr := fmt.Sprintf("postgres://postgres:test@%s:%s/risk_test", host, port.Port())

    // Run migrations
    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)

    // Test policy CRUD
    repo := NewPolicyRepository(db)

    policy := &Policy{
        Name:      "test-policy",
        Type:      "hard",
        Metric:    "leverage",
        Operator:  "less_than_or_equal",
        Threshold: 10.0,
        Scope:     "account",
        Enabled:   true,
    }

    // Create
    created, err := repo.CreatePolicy(ctx, policy)
    require.NoError(t, err)
    assert.NotEmpty(t, created.ID)

    // Read
    retrieved, err := repo.GetPolicy(ctx, created.ID)
    require.NoError(t, err)
    assert.Equal(t, policy.Name, retrieved.Name)

    // Update
    retrieved.Threshold = 5.0
    updated, err := repo.UpdatePolicy(ctx, retrieved)
    require.NoError(t, err)
    assert.Equal(t, 5.0, updated.Threshold)

    // Delete
    err = repo.DeletePolicy(ctx, created.ID)
    require.NoError(t, err)
}
```

### Limit Enforcement Scenarios

```go
func TestLimitEnforcementScenarios(t *testing.T) {
    testCases := []struct {
        name          string
        policies      []Policy
        accountState  AccountState
        order         OrderRiskRequest
        shouldApprove bool
        expectedViolations int
    }{
        {
            name: "order within limits",
            policies: []Policy{
                {
                    Name:      "max-leverage",
                    Metric:    "leverage",
                    Operator:  "less_than_or_equal",
                    Threshold: 10.0,
                },
            },
            accountState: AccountState{
                Balance:   10000,
                Positions: []Position{},
            },
            order: OrderRiskRequest{
                Symbol:   "BTCUSDT",
                Side:     "BUY",
                Quantity: 0.1,
                Price:    50000,
            },
            shouldApprove: true,
            expectedViolations: 0,
        },
        {
            name: "order exceeds leverage limit",
            policies: []Policy{
                {
                    Name:      "max-leverage",
                    Metric:    "leverage",
                    Operator:  "less_than_or_equal",
                    Threshold: 5.0,
                },
            },
            accountState: AccountState{
                Balance:   10000,
                Positions: []Position{},
            },
            order: OrderRiskRequest{
                Symbol:   "BTCUSDT",
                Side:     "BUY",
                Quantity: 2.0,
                Price:    50000, // 100k notional / 10k balance = 10x leverage
            },
            shouldApprove: false,
            expectedViolations: 1,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            server := &RiskServer{
                policyEngine: NewPolicyEngine(tc.policies),
            }

            resp, err := server.CheckOrder(context.Background(), &tc.order)
            require.NoError(t, err)

            assert.Equal(t, tc.shouldApprove, resp.Approved)
            assert.Equal(t, tc.expectedViolations, len(resp.Violations))
        })
    }
}
```

### Emergency Stop Simulation

```go
func TestEmergencyStopWorkflow(t *testing.T) {
    // Mock dependencies
    orderClient := &MockOrderExecutionClient{
        openOrders: []Order{
            {ID: "order1", Symbol: "BTCUSDT"},
            {ID: "order2", Symbol: "ETHUSDT"},
        },
    }

    accountClient := &MockAccountClient{
        positions: []Position{
            {Symbol: "BTCUSDT", Quantity: 0.5, Notional: 25000},
            {Symbol: "ETHUSDT", Quantity: 10, Notional: 30000},
        },
    }

    emergencyStop := &EmergencyStop{
        orderClient:   orderClient,
        accountClient: accountClient,
        lockManager:   NewMockLockManager(),
    }

    // Trigger emergency stop
    err := emergencyStop.Trigger(context.Background(), "test trigger", "manual")
    require.NoError(t, err)

    // Verify all orders canceled
    assert.Equal(t, 2, orderClient.canceledCount)

    // Verify all positions closed
    assert.Equal(t, 2, orderClient.closeOrdersSubmitted)

    // Verify trading disabled
    assert.True(t, emergencyStop.isTradingDisabled())
}
```

### Load Testing

**k6 Load Test Script:**
```javascript
import grpc from 'k6/net/grpc';
import { check } from 'k6';

const client = new grpc.Client();
client.load(['../proto'], 'risk_manager.proto');

export let options = {
  stages: [
    { duration: '1m', target: 100 },  // Ramp to 100 RPS
    { duration: '3m', target: 100 },  // Stay at 100 RPS
    { duration: '1m', target: 500 },  // Spike to 500 RPS
    { duration: '2m', target: 500 },  // Stay at 500 RPS
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    'grpc_req_duration{method="CheckOrder"}': ['p(99)<10'], // p99 < 10ms
    'grpc_req_failed{method="CheckOrder"}': ['rate<0.01'],  // <1% errors
  },
};

export default () => {
  client.connect('localhost:50051', { plaintext: true });

  const request = {
    order_id: `order_${Date.now()}`,
    symbol: 'BTCUSDT',
    side: 'BUY',
    quantity: 0.1,
    price: 50000,
    order_type: 'LIMIT',
    strategy_id: 'test_strategy',
    timestamp: Date.now(),
  };

  const response = client.invoke('risk.RiskManager/CheckOrder', request);

  check(response, {
    'status is OK': (r) => r.status === grpc.StatusOK,
    'processing time < 10ms': (r) => r.message.processing_time_us < 10000,
  });

  client.close();
};
```

### Chaos Engineering

**Test Network Failures:**
```go
func TestPreTradeCheckWithNetworkFailures(t *testing.T) {
    // Use toxiproxy to simulate network issues
    proxy := toxiproxy.NewProxy()
    proxy.Toxics.Add("latency", "latency", "downstream", 1.0, toxiproxy.Attributes{
        "latency": 5000, // 5s latency
    })

    server := &RiskServer{
        accountClient: NewAccountClient(proxy.Listen),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    resp, err := server.CheckOrder(ctx, &OrderRiskRequest{
        Symbol: "BTCUSDT",
        Side:   "BUY",
        Quantity: 0.1,
    })

    // Should timeout gracefully
    assert.Error(t, err)
    assert.Nil(t, resp)

    // Should log degradation
    // Should NOT crash
}
```

---

## Deployment

### Dockerfile

```dockerfile
# Multi-stage build for optimal image size
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protobuf-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf code
RUN make proto

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /risk-manager \
    ./cmd/risk-manager

# Final stage
FROM alpine:3.19

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /risk-manager /app/risk-manager

# Copy configuration templates
COPY --from=builder /app/configs /app/configs

# Create non-root user
RUN adduser -D -u 1000 riskmanager
USER riskmanager

# Expose ports
EXPOSE 50051 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/risk-manager", "healthcheck"]

# Run the service
ENTRYPOINT ["/app/risk-manager"]
CMD ["serve"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  risk-manager:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "50051:50051"  # gRPC
      - "9090:9090"    # Metrics
    environment:
      - DATABASE_URL=postgres://risk:password@postgres:5432/risk_db
      - REDIS_URL=redis://redis:6379
      - NATS_URL=nats://nats:4222
      - LOG_LEVEL=info
      - ACCOUNT_SERVICE_URL=account-monitor:50052
      - MARKET_SERVICE_URL=market-data:50053
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      nats:
        condition: service_started
    networks:
      - trading-net
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=risk_db
      - POSTGRES_USER=risk
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U risk"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - trading-net

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - trading-net

  nats:
    image: nats:2.10-alpine
    command: "--cluster_name risk-cluster --max_payload 10MB"
    ports:
      - "4222:4222"
    networks:
      - trading-net

networks:
  trading-net:
    driver: bridge

volumes:
  postgres-data:
  redis-data:
```

### Database Migrations

**Migration 001_initial_schema.up.sql:**
```sql
-- Risk policies table
CREATE TABLE risk_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('hard', 'soft', 'emergency')),
    metric VARCHAR(100) NOT NULL,
    operator VARCHAR(50) NOT NULL,
    threshold NUMERIC(20, 8) NOT NULL,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('account', 'symbol', 'strategy')),
    scope_id VARCHAR(100),
    action VARCHAR(50),
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_policies_enabled ON risk_policies(enabled) WHERE enabled = true;
CREATE INDEX idx_policies_scope ON risk_policies(scope, scope_id);
CREATE INDEX idx_policies_type ON risk_policies(type);

-- Risk violations table
CREATE TABLE risk_violations (
    id BIGSERIAL PRIMARY KEY,
    policy_id UUID REFERENCES risk_policies(id),
    violation_time TIMESTAMPTZ DEFAULT NOW(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value NUMERIC(20, 8) NOT NULL,
    threshold_value NUMERIC(20, 8) NOT NULL,
    context JSONB,
    action_taken VARCHAR(100),
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_violations_time ON risk_violations(violation_time DESC);
CREATE INDEX idx_violations_policy ON risk_violations(policy_id);
CREATE INDEX idx_violations_unresolved ON risk_violations(resolved) WHERE resolved = false;

-- Emergency stops table
CREATE TABLE emergency_stops (
    id BIGSERIAL PRIMARY KEY,
    trigger_time TIMESTAMPTZ DEFAULT NOW(),
    trigger_reason TEXT NOT NULL,
    triggered_by VARCHAR(100) NOT NULL,
    account_state JSONB,
    positions_snapshot JSONB,
    orders_canceled INTEGER DEFAULT 0,
    positions_closed INTEGER DEFAULT 0,
    completed_at TIMESTAMPTZ,
    re_enabled_at TIMESTAMPTZ,
    re_enabled_by VARCHAR(100)
);

CREATE INDEX idx_emergency_stops_time ON emergency_stops(trigger_time DESC);

-- Pre-trade check audit log
CREATE TABLE pre_trade_checks (
    id BIGSERIAL PRIMARY KEY,
    check_time TIMESTAMPTZ DEFAULT NOW(),
    order_id VARCHAR(100),
    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL,
    quantity NUMERIC(20, 8) NOT NULL,
    price NUMERIC(20, 8),
    approved BOOLEAN NOT NULL,
    violations TEXT[],
    processing_time_us INTEGER,
    account_leverage NUMERIC(10, 4),
    account_margin_ratio NUMERIC(10, 4)
);

CREATE INDEX idx_pre_trade_checks_time ON pre_trade_checks(check_time DESC);
CREATE INDEX idx_pre_trade_checks_symbol ON pre_trade_checks(symbol);
CREATE INDEX idx_pre_trade_checks_approved ON pre_trade_checks(approved);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_risk_policies_updated_at
    BEFORE UPDATE ON risk_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Health Checks

```go
type HealthChecker struct {
    db          *sql.DB
    redis       *redis.Client
    nats        *nats.Conn
    dependencies map[string]DependencyChecker
}

type HealthStatus struct {
    Status       string                 `json:"status"`
    Timestamp    time.Time              `json:"timestamp"`
    Version      string                 `json:"version"`
    Dependencies map[string]DepStatus   `json:"dependencies"`
}

type DepStatus struct {
    Status    string        `json:"status"`
    Latency   time.Duration `json:"latency_ms"`
    Error     string        `json:"error,omitempty"`
}

func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Timestamp:    time.Now(),
        Version:      version.Version,
        Dependencies: make(map[string]DepStatus),
    }

    // Check database
    dbStart := time.Now()
    if err := h.db.PingContext(ctx); err != nil {
        status.Dependencies["postgres"] = DepStatus{
            Status:  "unhealthy",
            Latency: time.Since(dbStart),
            Error:   err.Error(),
        }
    } else {
        status.Dependencies["postgres"] = DepStatus{
            Status:  "healthy",
            Latency: time.Since(dbStart),
        }
    }

    // Check Redis
    redisStart := time.Now()
    if err := h.redis.Ping(ctx).Err(); err != nil {
        status.Dependencies["redis"] = DepStatus{
            Status:  "unhealthy",
            Latency: time.Since(redisStart),
            Error:   err.Error(),
        }
    } else {
        status.Dependencies["redis"] = DepStatus{
            Status:  "healthy",
            Latency: time.Since(redisStart),
        }
    }

    // Check NATS
    if h.nats.IsConnected() {
        status.Dependencies["nats"] = DepStatus{
            Status: "healthy",
        }
    } else {
        status.Dependencies["nats"] = DepStatus{
            Status: "unhealthy",
            Error:  "not connected",
        }
    }

    // Check external services
    for name, checker := range h.dependencies {
        depStart := time.Now()
        if err := checker.Check(ctx); err != nil {
            status.Dependencies[name] = DepStatus{
                Status:  "unhealthy",
                Latency: time.Since(depStart),
                Error:   err.Error(),
            }
        } else {
            status.Dependencies[name] = DepStatus{
                Status:  "healthy",
                Latency: time.Since(depStart),
            }
        }
    }

    // Overall status
    allHealthy := true
    for _, dep := range status.Dependencies {
        if dep.Status != "healthy" {
            allHealthy = false
            break
        }
    }

    if allHealthy {
        status.Status = "healthy"
    } else {
        status.Status = "degraded"
    }

    return status
}
```

### Kubernetes Deployment (Optional)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: risk-manager
  labels:
    app: risk-manager
spec:
  replicas: 2
  selector:
    matchLabels:
      app: risk-manager
  template:
    metadata:
      labels:
        app: risk-manager
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: risk-manager
        image: risk-manager:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 9090
          name: metrics
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: risk-manager-secrets
              key: database-url
        - name: REDIS_URL
          value: "redis://redis:6379"
        - name: NATS_URL
          value: "nats://nats:4222"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command: ["/app/risk-manager", "healthcheck"]
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          exec:
            command: ["/app/risk-manager", "healthcheck"]
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: risk-manager
spec:
  selector:
    app: risk-manager
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

---

## Observability

### Prometheus Metrics

```go
var (
    // Pre-trade check metrics
    preTradeCheckDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "risk_pre_trade_check_duration_seconds",
            Help:    "Duration of pre-trade risk checks",
            Buckets: []float64{.001, .005, .01, .025, .05, .1},
        },
        []string{"approved"},
    )

    preTradeCheckTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "risk_pre_trade_check_total",
            Help: "Total number of pre-trade checks",
        },
        []string{"approved", "reason"},
    )

    // Risk metric gauges
    currentLeverage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "risk_current_leverage",
            Help: "Current account leverage",
        },
    )

    currentMarginRatio = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "risk_current_margin_ratio",
            Help: "Current margin ratio",
        },
    )

    currentDrawdown = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "risk_current_drawdown",
            Help: "Current drawdown percentage",
        },
        []string{"type"}, // daily, peak
    )

    positionConcentration = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "risk_position_concentration",
            Help: "Position concentration by symbol",
        },
        []string{"symbol"},
    )

    // Policy violations
    policyViolations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "risk_policy_violations_total",
            Help: "Total policy violations",
        },
        []string{"policy_id", "policy_name", "severity"},
    )

    // Emergency stops
    emergencyStops = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "risk_emergency_stops_total",
            Help: "Total emergency stops triggered",
        },
    )

    // Limit utilization
    limitUtilization = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "risk_limit_utilization",
            Help: "Risk limit utilization percentage",
        },
        []string{"limit_type", "scope"},
    )
)

// Update metrics
func (s *RiskServer) recordPreTradeCheck(
    approved bool,
    duration time.Duration,
    violations []string,
) {
    approvedStr := "true"
    if !approved {
        approvedStr = "false"
    }

    preTradeCheckDuration.WithLabelValues(approvedStr).Observe(duration.Seconds())

    if approved {
        preTradeCheckTotal.WithLabelValues("true", "").Inc()
    } else {
        for _, violation := range violations {
            preTradeCheckTotal.WithLabelValues("false", violation).Inc()
        }
    }
}

func (m *RiskMonitor) publishMetrics(metrics RiskMetrics) {
    currentLeverage.Set(metrics.Leverage)
    currentMarginRatio.Set(metrics.MarginRatio)
    currentDrawdown.WithLabelValues("daily").Set(metrics.DailyDrawdown)
    currentDrawdown.WithLabelValues("peak").Set(metrics.PeakDrawdown)

    for symbol, concentration := range metrics.PositionConcentration {
        positionConcentration.WithLabelValues(symbol).Set(concentration)
    }

    for limitType, utilization := range metrics.LimitUtilization {
        limitUtilization.WithLabelValues(limitType, "account").Set(utilization)
    }
}
```

### Grafana Dashboard Requirements

**Dashboard 1: Risk Overview**
- Current leverage (gauge)
- Current margin ratio (gauge)
- Daily drawdown (gauge with threshold markers)
- Peak drawdown (time series)
- Position concentration (bar chart)
- Limit utilization (multi-gauge)

**Dashboard 2: Pre-Trade Checks**
- Request rate (time series)
- Approval rate percentage (stat)
- p50/p95/p99 latency (time series)
- Rejection reasons (pie chart)
- Processing time distribution (heatmap)

**Dashboard 3: Policy Violations**
- Violations over time (time series by policy)
- Violation severity breakdown (stacked bar)
- Most violated policies (table)
- Violation resolution time (histogram)

**Dashboard 4: Emergency Stops**
- Emergency stop history (timeline)
- Trigger reasons (pie chart)
- Recovery time (stat)
- Positions closed per stop (bar chart)

**Dashboard 5: System Health**
- Service uptime
- Dependency health status
- Database connection pool usage
- Redis cache hit rate
- gRPC request rate

### Alert Rules

**Prometheus Alert Rules:**
```yaml
groups:
  - name: risk_manager_alerts
    interval: 30s
    rules:
      # High leverage warning
      - alert: HighLeverageWarning
        expr: risk_current_leverage > 7
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Account leverage is high"
          description: "Current leverage {{ $value }} exceeds warning threshold of 7x"

      # Critical leverage
      - alert: CriticalLeverage
        expr: risk_current_leverage > 9
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "Account leverage is critical"
          description: "Current leverage {{ $value }} approaching max limit of 10x"

      # Low margin ratio
      - alert: LowMarginRatio
        expr: risk_current_margin_ratio < 1.5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Margin ratio is low"
          description: "Current margin ratio {{ $value }} below safe threshold of 1.5"

      # High drawdown
      - alert: HighDrawdown
        expr: risk_current_drawdown{type="daily"} > 0.15
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Daily drawdown is high"
          description: "Daily drawdown {{ $value | humanizePercentage }} exceeds 15%"

      # Pre-trade check latency
      - alert: HighPreTradeCheckLatency
        expr: histogram_quantile(0.99, rate(risk_pre_trade_check_duration_seconds_bucket[5m])) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Pre-trade check latency is high"
          description: "p99 latency {{ $value | humanizeDuration }} exceeds 50ms"

      # High rejection rate
      - alert: HighRejectionRate
        expr: |
          sum(rate(risk_pre_trade_check_total{approved="false"}[5m])) /
          sum(rate(risk_pre_trade_check_total[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Pre-trade check rejection rate is high"
          description: "{{ $value | humanizePercentage }} of orders being rejected"

      # Service down
      - alert: RiskManagerDown
        expr: up{job="risk-manager"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Risk Manager service is down"
          description: "Risk Manager has been down for more than 1 minute"
```

### Logging Strategy

```go
// Structured logging with zap
logger, _ := zap.NewProduction()

// Pre-trade check logging
logger.Info("pre-trade check",
    zap.String("order_id", req.OrderId),
    zap.String("symbol", req.Symbol),
    zap.Float64("quantity", req.Quantity),
    zap.Bool("approved", approved),
    zap.Strings("violations", violations),
    zap.Duration("processing_time", processingTime),
)

// Policy violation logging
logger.Warn("policy violation detected",
    zap.String("policy_id", policy.ID),
    zap.String("policy_name", policy.Name),
    zap.String("metric", policy.Metric),
    zap.Float64("value", metricValue),
    zap.Float64("threshold", policy.Threshold),
    zap.String("action", policy.Action),
)

// Emergency stop logging
logger.Error("EMERGENCY STOP TRIGGERED",
    zap.String("reason", reason),
    zap.String("triggered_by", triggeredBy),
    zap.Int("orders_canceled", canceledCount),
    zap.Int("positions_closed", closedCount),
    zap.Time("trigger_time", triggerTime),
)

// Error logging with stack traces
logger.Error("risk check failed",
    zap.Error(err),
    zap.Stack("stack"),
    zap.String("context", "account_state_fetch"),
)
```

---

## Performance Targets

### Latency Requirements

| Operation | p50 | p95 | p99 | p99.9 |
|-----------|-----|-----|-----|-------|
| Pre-trade check | <2ms | <5ms | <10ms | <50ms |
| Risk metric calculation | <1ms | <2ms | <5ms | <10ms |
| Policy evaluation | <500us | <1ms | <2ms | <5ms |
| Emergency stop trigger | <100ms | <200ms | <500ms | <1s |

### Throughput Requirements

| Operation | Target RPS | Peak RPS |
|-----------|-----------|----------|
| Pre-trade checks | 1,000 | 5,000 |
| Risk metric queries | 100 | 500 |
| Policy updates | 10 | 50 |

### Resource Limits

| Resource | Normal | Peak |
|----------|--------|------|
| CPU | 250m | 500m |
| Memory | 256Mi | 512Mi |
| Database connections | 10 | 25 |
| Redis connections | 5 | 10 |

### Availability Targets

- **Uptime:** 99.9% (43 minutes downtime/month)
- **RTO (Recovery Time Objective):** <5 minutes
- **RPO (Recovery Point Objective):** <1 minute

---

## Risk Formulas Reference

### 1. Margin Ratio

```
Margin Ratio = Total Equity / Total Required Margin

Where:
  Total Equity = Balance + Unrealized PnL
  Total Required Margin = Σ(Position Notional / Position Leverage)
```

**Interpretation:**
- > 2.0: Safe
- 1.5 - 2.0: Caution
- 1.0 - 1.5: Risk of liquidation
- < 1.0: Liquidation imminent

### 2. Account Leverage

```
Leverage = Total Position Notional / Total Equity

Where:
  Total Position Notional = Σ|Position Notional|
  Total Equity = Balance + Unrealized PnL
```

**Example:**
- Balance: $10,000
- Position 1: $30,000 notional
- Position 2: $20,000 notional
- Leverage = ($30,000 + $20,000) / $10,000 = 5x

### 3. Daily Drawdown

```
Daily Drawdown = (Day Start Equity - Current Equity) / Day Start Equity

Where:
  Day Start Equity = Equity at market open (00:00 UTC)
  Current Equity = Balance + Unrealized PnL
```

**Example:**
- Start: $10,000
- Current: $8,500
- Daily DD = ($10,000 - $8,500) / $10,000 = 15%

### 4. Peak Drawdown

```
Peak Drawdown = (Peak Equity - Current Equity) / Peak Equity

Where:
  Peak Equity = Highest equity value since inception
  Current Equity = Balance + Unrealized PnL
```

### 5. Position Concentration

```
Concentration = Single Position Notional / Total Portfolio Value

Where:
  Single Position Notional = |Position Value|
  Total Portfolio Value = Balance + Unrealized PnL
```

**Risk Levels:**
- < 20%: Diversified
- 20-40%: Concentrated
- > 40%: High risk

### 6. Value at Risk (VaR) - Optional

```
VaR = Portfolio Value × Volatility × Z-score

Where:
  Volatility = Historical price volatility (σ)
  Z-score = Confidence level (1.65 for 95%, 2.33 for 99%)
```

**Example (95% confidence, 1-day VaR):**
- Portfolio: $10,000
- Daily volatility: 2%
- VaR = $10,000 × 0.02 × 1.65 = $330
- Interpretation: 95% chance loss won't exceed $330 in one day

---

## Appendix: Configuration Example

**config.yaml:**
```yaml
server:
  grpc_port: 50051
  metrics_port: 9090
  graceful_shutdown_timeout: 30s

database:
  url: ${DATABASE_URL}
  max_connections: 25
  max_idle_connections: 10
  connection_lifetime: 5m

redis:
  url: ${REDIS_URL}
  max_retries: 3
  pool_size: 10
  cache_ttl:
    account_state: 100ms
    policies: 1s
    market_prices: 100ms

nats:
  url: ${NATS_URL}
  max_reconnects: 10
  reconnect_wait: 2s

risk:
  monitoring_interval: 1s
  alert_deduplication_window: 5m
  emergency_stop_timeout: 30s

  default_limits:
    max_leverage: 10.0
    max_drawdown: 0.25
    max_position_concentration: 0.40
    max_orders_per_minute: 20

external_services:
  account_monitor:
    url: ${ACCOUNT_SERVICE_URL}
    timeout: 1s
    retry_attempts: 2

  market_data:
    url: ${MARKET_SERVICE_URL}
    timeout: 500ms
    retry_attempts: 1

logging:
  level: ${LOG_LEVEL:-info}
  format: json
  output: stdout

metrics:
  enabled: true
  path: /metrics
```

---

**End of Development Plan**

This comprehensive plan provides:
- Clear technology choices with rationale
- Detailed architecture and component design
- 6-phase development roadmap
- Complete implementation examples with code
- Extensive testing strategy
- Production-ready deployment configuration
- Full observability setup
- Performance targets and SLAs
- Risk calculation formulas and reference

The Risk Manager Service is now ready for development!
