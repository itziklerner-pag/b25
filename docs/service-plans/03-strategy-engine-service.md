# Strategy Engine Service - Development Plan

**Service:** Strategy Engine Service
**Purpose:** Trading strategy execution and signal generation with hot-reload capability
**Target Latency:** <500μs per strategy tick
**Last Updated:** 2025-10-02
**Version:** 1.0

---

## 1. Technology Stack Recommendation

### 1.1 Core Engine Language

**Recommended: Go**

**Rationale:**
- Excellent balance between performance and development speed
- Native concurrency primitives (goroutines) for parallel strategy execution
- Fast compilation for rapid iteration
- Built-in reflection for plugin loading
- Strong ecosystem for message queues and gRPC
- Easier to meet <500μs latency target than Python
- Better memory management than interpreted languages

**Alternative Options:**
- **Rust:** Maximum performance (~200μs possible), but slower development cycle
- **Python:** Fastest development, but difficult to meet latency requirements (use for strategies, not core)

### 1.2 Plugin System Approach

**Recommended: Multi-Layer Plugin Architecture**

```
Layer 1: Native Go Plugins (go plugin package)
  - For production-critical, high-performance strategies
  - Direct memory access, no serialization overhead
  - Compiled .so files hot-reloaded via plugin.Open()

Layer 2: Embedded Python (via gRPC or cgo)
  - For rapid strategy development and backtesting
  - Python subprocess with gRPC communication
  - Accept ~1-2ms latency penalty for flexibility

Layer 3: WASM Plugins (future extensibility)
  - Language-agnostic strategy deployment
  - Sandboxed execution environment
  - wasmer-go or wazero runtime
```

**Initial Focus:** Layer 1 (Go plugins) + Layer 2 (Python strategies)

### 1.3 Strategy Language Options

**For Rapid Development:**
```python
# Python Strategy Interface
class Strategy:
    def on_tick(self, orderbook: OrderBook, trades: List[Trade]) -> Signal:
        # Strategy logic here
        pass

    def on_fill(self, fill: Fill) -> None:
        # Position tracking
        pass
```

**For Production Performance:**
```go
// Go Strategy Plugin Interface
type Strategy interface {
    OnTick(ctx context.Context, data *MarketData) (*Signal, error)
    OnFill(ctx context.Context, fill *Fill) error
    GetState(ctx context.Context) (*StrategyState, error)
    HotReload(ctx context.Context, config *Config) error
}
```

### 1.4 Message Queue Client

**Recommended: NATS Client**

```go
import "github.com/nats-io/nats.go"
```

**Rationale:**
- Lightweight, low-latency pub/sub (<1ms)
- Native Go support with excellent performance
- Built-in request/reply for RPC calls
- Supports both pub/sub and queuing
- Easier operational overhead than Kafka

**Alternative:** Redis Pub/Sub (simpler, but less features)

### 1.5 Testing Frameworks

**Go Testing Stack:**
```go
// Core testing
"testing"                              // Standard library
"github.com/stretchr/testify/assert"   // Assertions
"github.com/stretchr/testify/mock"     // Mocking

// Benchmarking
"github.com/montanaflynn/stats"        // Statistical analysis

// Integration testing
"github.com/testcontainers/testcontainers-go" // Docker containers
```

**Python Strategy Testing:**
```python
pytest                  # Test framework
pytest-benchmark       # Performance testing
hypothesis             # Property-based testing
backtrader             # Backtesting framework
```

---

## 2. Architecture Design

### 2.1 High-Level Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Strategy Engine Service                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐      ┌──────────────┐      ┌───────────────┐  │
│  │   Plugin    │──────│   Strategy   │──────│    Signal     │  │
│  │   Loader    │      │   Registry   │      │  Aggregator   │  │
│  └─────────────┘      └──────────────┘      └───────────────┘  │
│         │                     │                      │           │
│         │                     ▼                      │           │
│  ┌─────────────┐      ┌──────────────┐      ┌───────────────┐  │
│  │ Hot-Reload  │      │   Market     │──────│     Risk      │  │
│  │  Manager    │      │Data Router   │      │    Filter     │  │
│  └─────────────┘      └──────────────┘      └───────────────┘  │
│                               │                      │           │
│                               ▼                      ▼           │
│                       ┌──────────────┐      ┌───────────────┐  │
│                       │  Position    │      │  Order Queue  │  │
│                       │   Tracker    │      │  (to exec)    │  │
│                       └──────────────┘      └───────────────┘  │
│                                                                   │
├─────────────────────────────────────────────────────────────────┤
│  Inputs: Market Data (NATS), Fill Events (NATS)                 │
│  Outputs: Order Requests (gRPC), Strategy State (gRPC)          │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Plugin Interface Definition

```go
// pkg/strategy/interface.go

package strategy

import (
    "context"
    "time"
)

// MarketData contains all market information for a single tick
type MarketData struct {
    Symbol       string
    Timestamp    time.Time
    OrderBook    *OrderBook
    RecentTrades []Trade
    OHLCV        *OHLCV
}

// OrderBook represents the order book snapshot
type OrderBook struct {
    Bids []PriceLevel  // Sorted descending
    Asks []PriceLevel  // Sorted ascending
}

type PriceLevel struct {
    Price    float64
    Quantity float64
}

type Trade struct {
    Price     float64
    Quantity  float64
    Timestamp time.Time
    Side      string // "buy" or "sell"
}

type OHLCV struct {
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume float64
}

// Signal represents a trading signal from a strategy
type Signal struct {
    StrategyID string
    Symbol     string
    Direction  Direction // LONG, SHORT, CLOSE
    Confidence float64   // 0.0 to 1.0
    Size       float64   // Optional size hint
    Metadata   map[string]interface{}
    Timestamp  time.Time
}

type Direction int

const (
    DirectionNone Direction = iota
    DirectionLong
    DirectionShort
    DirectionClose
)

// Fill represents an order execution event
type Fill struct {
    OrderID      string
    StrategyID   string
    Symbol       string
    Side         string
    Price        float64
    Quantity     float64
    Commission   float64
    Timestamp    time.Time
}

// StrategyState contains current strategy state for monitoring
type StrategyState struct {
    StrategyID     string
    IsActive       bool
    Position       float64
    UnrealizedPnL  float64
    RealizedPnL    float64
    SignalCount    int64
    LastSignalTime time.Time
    Metadata       map[string]interface{}
}

// Config contains strategy configuration
type Config struct {
    Params map[string]interface{}
}

// Strategy is the main plugin interface that all strategies must implement
type Strategy interface {
    // Initialize is called once when the strategy is loaded
    Initialize(ctx context.Context, config *Config) error

    // OnTick is called when new market data arrives
    // Must return within <500μs for production strategies
    OnTick(ctx context.Context, data *MarketData) (*Signal, error)

    // OnFill is called when an order is executed
    OnFill(ctx context.Context, fill *Fill) error

    // GetState returns current strategy state for monitoring
    GetState(ctx context.Context) (*StrategyState, error)

    // HotReload updates strategy configuration without restart
    HotReload(ctx context.Context, config *Config) error

    // Shutdown is called before strategy unload
    Shutdown(ctx context.Context) error
}

// Plugin represents the plugin metadata and constructor
type Plugin struct {
    Name        string
    Version     string
    Description string
    NewStrategy func() Strategy
}
```

### 2.3 Strategy Lifecycle Management

**State Machine:**
```
UNLOADED → LOADING → INITIALIZED → ACTIVE → PAUSED → SHUTTING_DOWN → UNLOADED
                          ↓            ↓
                      RELOAD_PENDING  ↓
                          ↓            ↓
                      INITIALIZED  UNLOADED
```

**Lifecycle Manager:**
```go
// internal/lifecycle/manager.go

package lifecycle

import (
    "context"
    "sync"
    "time"
)

type State int

const (
    StateUnloaded State = iota
    StateLoading
    StateInitialized
    StateActive
    StatePaused
    StateReloadPending
    StateShuttingDown
)

type StrategyLifecycle struct {
    ID         string
    Strategy   strategy.Strategy
    State      State
    Config     *strategy.Config
    LastTick   time.Time
    ErrorCount int
    mu         sync.RWMutex
}

type LifecycleManager struct {
    strategies map[string]*StrategyLifecycle
    mu         sync.RWMutex
}

func NewLifecycleManager() *LifecycleManager {
    return &LifecycleManager{
        strategies: make(map[string]*StrategyLifecycle),
    }
}

func (m *LifecycleManager) Load(ctx context.Context, id string, s strategy.Strategy, config *strategy.Config) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    lc := &StrategyLifecycle{
        ID:       id,
        Strategy: s,
        State:    StateLoading,
        Config:   config,
    }

    if err := s.Initialize(ctx, config); err != nil {
        return err
    }

    lc.State = StateInitialized
    m.strategies[id] = lc
    return nil
}

func (m *LifecycleManager) Activate(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    lc, exists := m.strategies[id]
    if !exists {
        return fmt.Errorf("strategy %s not found", id)
    }

    lc.mu.Lock()
    defer lc.mu.Unlock()

    if lc.State == StateInitialized || lc.State == StatePaused {
        lc.State = StateActive
        return nil
    }

    return fmt.Errorf("cannot activate strategy in state %v", lc.State)
}

func (m *LifecycleManager) Pause(id string) error {
    // Similar pattern...
}

func (m *LifecycleManager) Reload(ctx context.Context, id string, newConfig *strategy.Config) error {
    m.mu.RLock()
    lc := m.strategies[id]
    m.mu.RUnlock()

    if lc == nil {
        return fmt.Errorf("strategy %s not found", id)
    }

    lc.mu.Lock()
    lc.State = StateReloadPending
    lc.mu.Unlock()

    // Wait for current tick to complete (with timeout)
    // Then call HotReload
    if err := lc.Strategy.HotReload(ctx, newConfig); err != nil {
        lc.mu.Lock()
        lc.State = StateActive
        lc.mu.Unlock()
        return err
    }

    lc.mu.Lock()
    lc.Config = newConfig
    lc.State = StateActive
    lc.mu.Unlock()

    return nil
}

func (m *LifecycleManager) Unload(ctx context.Context, id string) error {
    // Shutdown strategy gracefully and remove from registry
}
```

### 2.4 Signal Aggregation Algorithms

```go
// internal/aggregator/aggregator.go

package aggregator

import (
    "context"
    "time"
)

type AggregationMethod int

const (
    MethodMajorityVote AggregationMethod = iota
    MethodWeightedAverage
    MethodUnanimous
    MethodFirstSignal
)

type Aggregator struct {
    method        AggregationMethod
    windowSize    time.Duration
    signalBuffer  map[string][]strategy.Signal // key: symbol
}

func NewAggregator(method AggregationMethod, windowSize time.Duration) *Aggregator {
    return &Aggregator{
        method:       method,
        windowSize:   windowSize,
        signalBuffer: make(map[string][]strategy.Signal),
    }
}

// AddSignal adds a signal to the aggregation buffer
func (a *Aggregator) AddSignal(signal strategy.Signal) {
    a.signalBuffer[signal.Symbol] = append(a.signalBuffer[signal.Symbol], signal)
}

// Aggregate combines signals for a symbol
func (a *Aggregator) Aggregate(ctx context.Context, symbol string) (*strategy.Signal, error) {
    signals := a.signalBuffer[symbol]
    if len(signals) == 0 {
        return nil, nil
    }

    // Remove stale signals
    cutoff := time.Now().Add(-a.windowSize)
    freshSignals := make([]strategy.Signal, 0, len(signals))
    for _, sig := range signals {
        if sig.Timestamp.After(cutoff) {
            freshSignals = append(freshSignals, sig)
        }
    }
    a.signalBuffer[symbol] = freshSignals

    if len(freshSignals) == 0 {
        return nil, nil
    }

    switch a.method {
    case MethodMajorityVote:
        return a.majorityVote(freshSignals), nil
    case MethodWeightedAverage:
        return a.weightedAverage(freshSignals), nil
    case MethodUnanimous:
        return a.unanimous(freshSignals), nil
    case MethodFirstSignal:
        return &freshSignals[0], nil
    default:
        return nil, fmt.Errorf("unknown aggregation method: %v", a.method)
    }
}

// majorityVote returns the direction with most votes, weighted by confidence
func (a *Aggregator) majorityVote(signals []strategy.Signal) *strategy.Signal {
    votes := make(map[strategy.Direction]float64)
    totalConfidence := 0.0

    for _, sig := range signals {
        votes[sig.Direction] += sig.Confidence
        totalConfidence += sig.Confidence
    }

    // Find direction with highest weighted votes
    var maxDirection strategy.Direction
    var maxVotes float64

    for dir, v := range votes {
        if v > maxVotes {
            maxVotes = v
            maxDirection = dir
        }
    }

    // Average confidence
    avgConfidence := totalConfidence / float64(len(signals))

    return &strategy.Signal{
        StrategyID: "aggregated",
        Symbol:     signals[0].Symbol,
        Direction:  maxDirection,
        Confidence: avgConfidence,
        Timestamp:  time.Now(),
    }
}

// weightedAverage combines signals using confidence-weighted averaging
func (a *Aggregator) weightedAverage(signals []strategy.Signal) *strategy.Signal {
    // Convert directions to numeric values: LONG=1, CLOSE=0, SHORT=-1
    // Weight by confidence and average
    var weightedSum float64
    var totalWeight float64

    for _, sig := range signals {
        value := a.directionToValue(sig.Direction)
        weight := sig.Confidence
        weightedSum += value * weight
        totalWeight += weight
    }

    avgValue := weightedSum / totalWeight
    direction := a.valueToDirection(avgValue)

    return &strategy.Signal{
        StrategyID: "aggregated",
        Symbol:     signals[0].Symbol,
        Direction:  direction,
        Confidence: totalWeight / float64(len(signals)),
        Timestamp:  time.Now(),
    }
}

func (a *Aggregator) directionToValue(d strategy.Direction) float64 {
    switch d {
    case strategy.DirectionLong:
        return 1.0
    case strategy.DirectionShort:
        return -1.0
    case strategy.DirectionClose:
        return 0.0
    default:
        return 0.0
    }
}

func (a *Aggregator) valueToDirection(v float64) strategy.Direction {
    if v > 0.3 {
        return strategy.DirectionLong
    } else if v < -0.3 {
        return strategy.DirectionShort
    }
    return strategy.DirectionClose
}

// unanimous requires all strategies to agree
func (a *Aggregator) unanimous(signals []strategy.Signal) *strategy.Signal {
    if len(signals) == 0 {
        return nil
    }

    firstDirection := signals[0].Direction
    for _, sig := range signals[1:] {
        if sig.Direction != firstDirection {
            // No consensus
            return &strategy.Signal{
                StrategyID: "aggregated",
                Symbol:     signals[0].Symbol,
                Direction:  strategy.DirectionNone,
                Confidence: 0.0,
                Timestamp:  time.Now(),
            }
        }
    }

    // All agree
    avgConfidence := 0.0
    for _, sig := range signals {
        avgConfidence += sig.Confidence
    }
    avgConfidence /= float64(len(signals))

    return &strategy.Signal{
        StrategyID: "aggregated",
        Symbol:     signals[0].Symbol,
        Direction:  firstDirection,
        Confidence: avgConfidence,
        Timestamp:  time.Now(),
    }
}
```

### 2.5 Risk Filter Pipeline

```go
// internal/risk/filter.go

package risk

import (
    "context"
    "fmt"
)

type FilterResult struct {
    Passed  bool
    Reason  string
    Signal  *strategy.Signal
}

type Filter interface {
    Name() string
    Check(ctx context.Context, signal *strategy.Signal) (*FilterResult, error)
}

// Pipeline executes a chain of risk filters
type Pipeline struct {
    filters []Filter
}

func NewPipeline(filters ...Filter) *Pipeline {
    return &Pipeline{filters: filters}
}

func (p *Pipeline) Execute(ctx context.Context, signal *strategy.Signal) (*FilterResult, error) {
    for _, filter := range p.filters {
        result, err := filter.Check(ctx, signal)
        if err != nil {
            return nil, fmt.Errorf("filter %s error: %w", filter.Name(), err)
        }
        if !result.Passed {
            return result, nil
        }
    }

    return &FilterResult{
        Passed: true,
        Signal: signal,
    }, nil
}

// ConfidenceFilter rejects signals below confidence threshold
type ConfidenceFilter struct {
    minConfidence float64
}

func NewConfidenceFilter(minConfidence float64) *ConfidenceFilter {
    return &ConfidenceFilter{minConfidence: minConfidence}
}

func (f *ConfidenceFilter) Name() string {
    return "confidence_filter"
}

func (f *ConfidenceFilter) Check(ctx context.Context, signal *strategy.Signal) (*FilterResult, error) {
    if signal.Confidence < f.minConfidence {
        return &FilterResult{
            Passed: false,
            Reason: fmt.Sprintf("confidence %.2f below threshold %.2f",
                signal.Confidence, f.minConfidence),
        }, nil
    }
    return &FilterResult{Passed: true, Signal: signal}, nil
}

// PositionLimitFilter prevents position size from exceeding limits
type PositionLimitFilter struct {
    positionTracker *PositionTracker
    maxPosition     float64
}

func NewPositionLimitFilter(tracker *PositionTracker, maxPosition float64) *PositionLimitFilter {
    return &PositionLimitFilter{
        positionTracker: tracker,
        maxPosition:     maxPosition,
    }
}

func (f *PositionLimitFilter) Name() string {
    return "position_limit_filter"
}

func (f *PositionLimitFilter) Check(ctx context.Context, signal *strategy.Signal) (*FilterResult, error) {
    currentPos := f.positionTracker.GetPosition(signal.StrategyID, signal.Symbol)

    // Check if signal would exceed limits
    var newPos float64
    switch signal.Direction {
    case strategy.DirectionLong:
        newPos = currentPos + signal.Size
    case strategy.DirectionShort:
        newPos = currentPos - signal.Size
    case strategy.DirectionClose:
        newPos = 0
    }

    if abs(newPos) > f.maxPosition {
        return &FilterResult{
            Passed: false,
            Reason: fmt.Sprintf("position %.2f would exceed limit %.2f",
                newPos, f.maxPosition),
        }, nil
    }

    return &FilterResult{Passed: true, Signal: signal}, nil
}

// DrawdownFilter prevents trading if strategy drawdown exceeds limit
type DrawdownFilter struct {
    positionTracker *PositionTracker
    maxDrawdown     float64
}

func (f *DrawdownFilter) Name() string {
    return "drawdown_filter"
}

func (f *DrawdownFilter) Check(ctx context.Context, signal *strategy.Signal) (*FilterResult, error) {
    state := f.positionTracker.GetStrategyState(signal.StrategyID)

    // Calculate drawdown from peak
    drawdown := (state.PeakPnL - state.RealizedPnL) / state.PeakPnL

    if drawdown > f.maxDrawdown {
        return &FilterResult{
            Passed: false,
            Reason: fmt.Sprintf("drawdown %.2f%% exceeds limit %.2f%%",
                drawdown*100, f.maxDrawdown*100),
        }, nil
    }

    return &FilterResult{Passed: true, Signal: signal}, nil
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
```

### 2.6 Hot-Reload Mechanism

```go
// internal/hotreload/watcher.go

package hotreload

import (
    "context"
    "log"
    "path/filepath"
    "time"

    "github.com/fsnotify/fsnotify"
)

type ReloadEvent struct {
    StrategyID string
    PluginPath string
    ConfigPath string
    Timestamp  time.Time
}

type Watcher struct {
    pluginDir  string
    configDir  string
    watcher    *fsnotify.Watcher
    reloadChan chan ReloadEvent
}

func NewWatcher(pluginDir, configDir string) (*Watcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    w := &Watcher{
        pluginDir:  pluginDir,
        configDir:  configDir,
        watcher:    watcher,
        reloadChan: make(chan ReloadEvent, 10),
    }

    // Watch plugin directory
    if err := watcher.Add(pluginDir); err != nil {
        return nil, err
    }

    // Watch config directory
    if err := watcher.Add(configDir); err != nil {
        return nil, err
    }

    return w, nil
}

func (w *Watcher) Start(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return w.watcher.Close()
        case event, ok := <-w.watcher.Events:
            if !ok {
                return nil
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
                w.handleFileChange(event.Name)
            }
        case err, ok := <-w.watcher.Errors:
            if !ok {
                return nil
            }
            log.Printf("watcher error: %v", err)
        }
    }
}

func (w *Watcher) handleFileChange(path string) {
    ext := filepath.Ext(path)
    basename := filepath.Base(path)

    // Ignore temporary files
    if basename[0] == '.' || basename[len(basename)-1] == '~' {
        return
    }

    strategyID := filepath.Base(filepath.Dir(path))

    switch ext {
    case ".so": // Go plugin
        w.reloadChan <- ReloadEvent{
            StrategyID: strategyID,
            PluginPath: path,
            Timestamp:  time.Now(),
        }
    case ".yaml", ".json": // Config file
        w.reloadChan <- ReloadEvent{
            StrategyID: strategyID,
            ConfigPath: path,
            Timestamp:  time.Now(),
        }
    }
}

func (w *Watcher) ReloadEvents() <-chan ReloadEvent {
    return w.reloadChan
}

// Reload coordinator
type ReloadCoordinator struct {
    watcher         *Watcher
    lifecycleManager *lifecycle.LifecycleManager
    pluginLoader    *PluginLoader
}

func NewReloadCoordinator(
    watcher *Watcher,
    lm *lifecycle.LifecycleManager,
    pl *PluginLoader,
) *ReloadCoordinator {
    return &ReloadCoordinator{
        watcher:          watcher,
        lifecycleManager: lm,
        pluginLoader:     pl,
    }
}

func (rc *ReloadCoordinator) Start(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        case event := <-rc.watcher.ReloadEvents():
            if err := rc.handleReload(ctx, event); err != nil {
                log.Printf("reload error for %s: %v", event.StrategyID, err)
            }
        }
    }
}

func (rc *ReloadCoordinator) handleReload(ctx context.Context, event ReloadEvent) error {
    log.Printf("Hot-reloading strategy: %s", event.StrategyID)

    if event.PluginPath != "" {
        // Reload plugin
        newStrategy, err := rc.pluginLoader.Load(event.PluginPath)
        if err != nil {
            return fmt.Errorf("failed to load plugin: %w", err)
        }

        // Replace strategy instance
        if err := rc.lifecycleManager.ReplaceStrategy(ctx, event.StrategyID, newStrategy); err != nil {
            return fmt.Errorf("failed to replace strategy: %w", err)
        }
    }

    if event.ConfigPath != "" {
        // Reload config
        newConfig, err := loadConfig(event.ConfigPath)
        if err != nil {
            return fmt.Errorf("failed to load config: %w", err)
        }

        // Apply new config via hot-reload
        if err := rc.lifecycleManager.Reload(ctx, event.StrategyID, newConfig); err != nil {
            return fmt.Errorf("failed to reload config: %w", err)
        }
    }

    log.Printf("Successfully hot-reloaded strategy: %s", event.StrategyID)
    return nil
}
```

---

## 3. Development Phases

### Phase 1: Core Engine and Plugin Interface (Days 1-2)

**Deliverables:**
- Plugin interface definition (`pkg/strategy/interface.go`)
- Basic plugin loader for Go plugins
- Strategy registry
- Lifecycle manager (load, unload, pause)
- Unit tests for plugin loading

**Tasks:**
1. Define Strategy interface and data structures
2. Implement Go plugin loader using `plugin` package
3. Create strategy registry with map-based storage
4. Implement basic lifecycle state machine
5. Write tests with mock strategies

**Validation:**
- Load and unload 10 strategies without errors
- State transitions work correctly
- Mock strategy OnTick executes <100μs

### Phase 2: Market Data Consumption (Day 3)

**Deliverables:**
- NATS subscriber for market data topics
- Market data router to active strategies
- Data deserialization and validation
- Concurrent strategy execution

**Tasks:**
1. Set up NATS connection and subscriptions
2. Implement market data deserialization (JSON or Protobuf)
3. Create router to dispatch data to strategies
4. Implement goroutine pool for parallel strategy execution
5. Add error handling and circuit breakers

**Validation:**
- Subscribe to market_data.*.orderbook topics
- Route data to 5 concurrent strategies
- Measure end-to-end latency <1ms (including network)

### Phase 3: Signal Generation and Aggregation (Days 4-5)

**Deliverables:**
- Signal aggregator with multiple algorithms
- Signal buffer with time-window management
- Signal validation and sanitization
- Metrics collection for signals

**Tasks:**
1. Implement signal aggregation algorithms (majority vote, weighted)
2. Create time-windowed signal buffer
3. Add signal validation (bounds checking, completeness)
4. Implement Prometheus metrics for signal counts
5. Write unit tests for aggregation logic

**Validation:**
- Majority vote aggregates 3 conflicting signals correctly
- Weighted average handles confidence values
- Stale signals are pruned from buffer

### Phase 4: Risk Filtering (Days 6-7)

**Deliverables:**
- Risk filter pipeline
- Confidence filter
- Position limit filter
- Drawdown filter
- Filter result tracking

**Tasks:**
1. Implement risk filter interface
2. Create filter pipeline executor
3. Implement confidence threshold filter
4. Implement position limit filter (requires position tracker)
5. Implement drawdown filter
6. Add filter bypass flags for testing

**Validation:**
- Pipeline rejects low-confidence signals
- Position limits are enforced
- Filters execute in <50μs total

### Phase 5: Position Tracking (Days 8-9)

**Deliverables:**
- Position tracker with per-strategy positions
- Fill event subscriber (NATS)
- Realized/unrealized P&L calculation
- gRPC API for position queries

**Tasks:**
1. Implement position tracker with thread-safe state
2. Subscribe to fill events from Order Execution Service
3. Update positions on fill events
4. Calculate P&L (mark-to-market)
5. Create gRPC service for position queries
6. Add position persistence (optional, SQLite)

**Validation:**
- Positions update correctly on fills
- P&L matches manual calculation
- gRPC API returns positions <10ms

### Phase 6: Example Strategies (Days 10-11)

**Deliverables:**
- Simple momentum strategy (Go plugin)
- Market making strategy (Go plugin)
- Python strategy wrapper (gRPC-based)
- Strategy backtesting harness

**Tasks:**
1. Implement momentum crossover strategy
2. Implement basic market making strategy
3. Create Python strategy gRPC server template
4. Build backtesting framework with CSV replay
5. Document strategy development guide

**Validation:**
- Strategies load and execute without errors
- Backtesting produces expected signals on historical data
- Python strategy latency <2ms

### Phase 7: Hot-Reload and Testing (Days 12-13)

**Deliverables:**
- File watcher for plugin directory
- Hot-reload coordinator
- Comprehensive integration tests
- Performance benchmarks
- Docker container

**Tasks:**
1. Implement fsnotify-based file watcher
2. Create reload coordinator for zero-downtime reloads
3. Write integration tests with testcontainers
4. Run performance benchmarks (target <500μs)
5. Create Dockerfile and docker-compose integration

**Validation:**
- Strategy hot-reloads without dropping ticks
- Integration test passes with all services
- Latency benchmark <500μs p99

---

## 4. Implementation Details

### 4.1 Example Strategy: Momentum Crossover

```go
// strategies/momentum/momentum.go

package main

import (
    "context"
    "fmt"

    "github.com/b25/strategy-engine/pkg/strategy"
)

// MomentumStrategy implements simple moving average crossover
type MomentumStrategy struct {
    shortWindow int
    longWindow  int

    prices      []float64
    position    float64
}

func NewMomentumStrategy() *MomentumStrategy {
    return &MomentumStrategy{
        shortWindow: 5,
        longWindow:  20,
        prices:      make([]float64, 0, 100),
    }
}

func (s *MomentumStrategy) Initialize(ctx context.Context, config *strategy.Config) error {
    if sw, ok := config.Params["short_window"].(float64); ok {
        s.shortWindow = int(sw)
    }
    if lw, ok := config.Params["long_window"].(float64); ok {
        s.longWindow = int(lw)
    }
    return nil
}

func (s *MomentumStrategy) OnTick(ctx context.Context, data *strategy.MarketData) (*strategy.Signal, error) {
    // Get mid price
    midPrice := (data.OrderBook.Bids[0].Price + data.OrderBook.Asks[0].Price) / 2

    // Add to price history
    s.prices = append(s.prices, midPrice)
    if len(s.prices) > s.longWindow {
        s.prices = s.prices[1:]
    }

    // Need enough data
    if len(s.prices) < s.longWindow {
        return nil, nil
    }

    // Calculate moving averages
    shortMA := s.sma(s.shortWindow)
    longMA := s.sma(s.longWindow)

    // Generate signal
    var signal *strategy.Signal

    if shortMA > longMA && s.position <= 0 {
        // Golden cross - go long
        signal = &strategy.Signal{
            StrategyID: "momentum",
            Symbol:     data.Symbol,
            Direction:  strategy.DirectionLong,
            Confidence: 0.7,
            Size:       1.0,
            Timestamp:  data.Timestamp,
        }
    } else if shortMA < longMA && s.position >= 0 {
        // Death cross - go short
        signal = &strategy.Signal{
            StrategyID: "momentum",
            Symbol:     data.Symbol,
            Direction:  strategy.DirectionShort,
            Confidence: 0.7,
            Size:       1.0,
            Timestamp:  data.Timestamp,
        }
    }

    return signal, nil
}

func (s *MomentumStrategy) sma(window int) float64 {
    start := len(s.prices) - window
    sum := 0.0
    for i := start; i < len(s.prices); i++ {
        sum += s.prices[i]
    }
    return sum / float64(window)
}

func (s *MomentumStrategy) OnFill(ctx context.Context, fill *strategy.Fill) error {
    // Update position
    if fill.Side == "buy" {
        s.position += fill.Quantity
    } else {
        s.position -= fill.Quantity
    }
    return nil
}

func (s *MomentumStrategy) GetState(ctx context.Context) (*strategy.StrategyState, error) {
    return &strategy.StrategyState{
        StrategyID: "momentum",
        IsActive:   true,
        Position:   s.position,
        Metadata: map[string]interface{}{
            "short_window": s.shortWindow,
            "long_window":  s.longWindow,
        },
    }, nil
}

func (s *MomentumStrategy) HotReload(ctx context.Context, config *strategy.Config) error {
    // Update parameters without losing position
    if sw, ok := config.Params["short_window"].(float64); ok {
        s.shortWindow = int(sw)
    }
    if lw, ok := config.Params["long_window"].(float64); ok {
        s.longWindow = int(lw)
    }
    return nil
}

func (s *MomentumStrategy) Shutdown(ctx context.Context) error {
    // Clean up resources
    return nil
}

// Plugin export
var Plugin = strategy.Plugin{
    Name:        "momentum",
    Version:     "1.0.0",
    Description: "Simple moving average crossover strategy",
    NewStrategy: func() strategy.Strategy {
        return NewMomentumStrategy()
    },
}
```

**Build command:**
```bash
go build -buildmode=plugin -o momentum.so strategies/momentum/momentum.go
```

### 4.2 Example Strategy: Market Maker

```go
// strategies/market_maker/market_maker.go

package main

import (
    "context"

    "github.com/b25/strategy-engine/pkg/strategy"
)

type MarketMakerStrategy struct {
    spread      float64
    targetSize  float64
    maxPosition float64

    position float64
}

func NewMarketMakerStrategy() *MarketMakerStrategy {
    return &MarketMakerStrategy{
        spread:      0.001, // 0.1% spread
        targetSize:  1.0,
        maxPosition: 5.0,
    }
}

func (s *MarketMakerStrategy) Initialize(ctx context.Context, config *strategy.Config) error {
    if spread, ok := config.Params["spread"].(float64); ok {
        s.spread = spread
    }
    if size, ok := config.Params["target_size"].(float64); ok {
        s.targetSize = size
    }
    if max, ok := config.Params["max_position"].(float64); ok {
        s.maxPosition = max
    }
    return nil
}

func (s *MarketMakerStrategy) OnTick(ctx context.Context, data *strategy.MarketData) (*strategy.Signal, error) {
    // Get best bid/ask
    bestBid := data.OrderBook.Bids[0].Price
    bestAsk := data.OrderBook.Asks[0].Price
    midPrice := (bestBid + bestAsk) / 2

    // Calculate target prices
    targetBid := midPrice * (1 - s.spread/2)
    targetAsk := midPrice * (1 + s.spread/2)

    // Determine if we should provide liquidity
    // Skew based on current position (inventory management)
    positionSkew := s.position / s.maxPosition // -1 to 1

    var signal *strategy.Signal

    if positionSkew < 0.5 && s.position < s.maxPosition {
        // We're short or neutral, willing to buy
        signal = &strategy.Signal{
            StrategyID: "market_maker",
            Symbol:     data.Symbol,
            Direction:  strategy.DirectionLong,
            Confidence: 0.6,
            Size:       s.targetSize,
            Metadata: map[string]interface{}{
                "limit_price": targetBid,
                "order_type":  "limit",
            },
            Timestamp: data.Timestamp,
        }
    } else if positionSkew > -0.5 && s.position > -s.maxPosition {
        // We're long or neutral, willing to sell
        signal = &strategy.Signal{
            StrategyID: "market_maker",
            Symbol:     data.Symbol,
            Direction:  strategy.DirectionShort,
            Confidence: 0.6,
            Size:       s.targetSize,
            Metadata: map[string]interface{}{
                "limit_price": targetAsk,
                "order_type":  "limit",
            },
            Timestamp: data.Timestamp,
        }
    }

    return signal, nil
}

func (s *MarketMakerStrategy) OnFill(ctx context.Context, fill *strategy.Fill) error {
    if fill.Side == "buy" {
        s.position += fill.Quantity
    } else {
        s.position -= fill.Quantity
    }
    return nil
}

func (s *MarketMakerStrategy) GetState(ctx context.Context) (*strategy.StrategyState, error) {
    return &strategy.StrategyState{
        StrategyID: "market_maker",
        IsActive:   true,
        Position:   s.position,
        Metadata: map[string]interface{}{
            "spread":       s.spread,
            "target_size":  s.targetSize,
            "max_position": s.maxPosition,
        },
    }, nil
}

func (s *MarketMakerStrategy) HotReload(ctx context.Context, config *strategy.Config) error {
    // Update parameters
    if spread, ok := config.Params["spread"].(float64); ok {
        s.spread = spread
    }
    return nil
}

func (s *MarketMakerStrategy) Shutdown(ctx context.Context) error {
    return nil
}

var Plugin = strategy.Plugin{
    Name:        "market_maker",
    Version:     "1.0.0",
    Description: "Basic market making strategy with inventory management",
    NewStrategy: func() strategy.Strategy {
        return NewMarketMakerStrategy()
    },
}
```

### 4.3 Python Strategy Wrapper

**Python Strategy Template:**
```python
# strategies/python/base_strategy.py

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Optional, Dict, Any, List
from datetime import datetime

@dataclass
class PriceLevel:
    price: float
    quantity: float

@dataclass
class OrderBook:
    bids: List[PriceLevel]
    asks: List[PriceLevel]

@dataclass
class Trade:
    price: float
    quantity: float
    timestamp: datetime
    side: str

@dataclass
class MarketData:
    symbol: str
    timestamp: datetime
    orderbook: OrderBook
    recent_trades: List[Trade]

@dataclass
class Signal:
    strategy_id: str
    symbol: str
    direction: str  # "LONG", "SHORT", "CLOSE"
    confidence: float
    size: float
    metadata: Dict[str, Any]
    timestamp: datetime

@dataclass
class Fill:
    order_id: str
    strategy_id: str
    symbol: str
    side: str
    price: float
    quantity: float
    commission: float
    timestamp: datetime

class BaseStrategy(ABC):
    def __init__(self, strategy_id: str):
        self.strategy_id = strategy_id
        self.position = 0.0

    @abstractmethod
    def initialize(self, config: Dict[str, Any]) -> None:
        """Called once on strategy load"""
        pass

    @abstractmethod
    def on_tick(self, data: MarketData) -> Optional[Signal]:
        """Called on every market data update"""
        pass

    def on_fill(self, fill: Fill) -> None:
        """Called when order is filled"""
        if fill.side == "buy":
            self.position += fill.quantity
        else:
            self.position -= fill.quantity

    def get_state(self) -> Dict[str, Any]:
        """Return current strategy state"""
        return {
            "strategy_id": self.strategy_id,
            "is_active": True,
            "position": self.position,
        }

    def hot_reload(self, config: Dict[str, Any]) -> None:
        """Update configuration without restart"""
        self.initialize(config)

    def shutdown(self) -> None:
        """Cleanup before unload"""
        pass
```

**Example Python Strategy:**
```python
# strategies/python/rsi_strategy.py

import numpy as np
from base_strategy import BaseStrategy, MarketData, Signal
from typing import Optional, Dict, Any
from datetime import datetime

class RSIStrategy(BaseStrategy):
    def __init__(self):
        super().__init__("rsi_strategy")
        self.period = 14
        self.overbought = 70
        self.oversold = 30
        self.prices = []

    def initialize(self, config: Dict[str, Any]) -> None:
        self.period = config.get("period", 14)
        self.overbought = config.get("overbought", 70)
        self.oversold = config.get("oversold", 30)

    def on_tick(self, data: MarketData) -> Optional[Signal]:
        # Calculate mid price
        mid_price = (data.orderbook.bids[0].price + data.orderbook.asks[0].price) / 2
        self.prices.append(mid_price)

        # Keep only what we need
        if len(self.prices) > self.period + 1:
            self.prices = self.prices[-(self.period + 1):]

        # Need enough data
        if len(self.prices) < self.period + 1:
            return None

        # Calculate RSI
        rsi = self._calculate_rsi()

        # Generate signal
        if rsi < self.oversold and self.position <= 0:
            return Signal(
                strategy_id=self.strategy_id,
                symbol=data.symbol,
                direction="LONG",
                confidence=min(1.0, (self.oversold - rsi) / self.oversold),
                size=1.0,
                metadata={"rsi": rsi},
                timestamp=datetime.now(),
            )
        elif rsi > self.overbought and self.position >= 0:
            return Signal(
                strategy_id=self.strategy_id,
                symbol=data.symbol,
                direction="SHORT",
                confidence=min(1.0, (rsi - self.overbought) / (100 - self.overbought)),
                size=1.0,
                metadata={"rsi": rsi},
                timestamp=datetime.now(),
            )

        return None

    def _calculate_rsi(self) -> float:
        deltas = np.diff(self.prices)
        gains = np.where(deltas > 0, deltas, 0)
        losses = np.where(deltas < 0, -deltas, 0)

        avg_gain = np.mean(gains[-self.period:])
        avg_loss = np.mean(losses[-self.period:])

        if avg_loss == 0:
            return 100

        rs = avg_gain / avg_loss
        rsi = 100 - (100 / (1 + rs))
        return rsi


# Factory function for gRPC server
def create_strategy():
    return RSIStrategy()
```

**gRPC Service for Python Strategies:**
```python
# strategies/python/grpc_server.py

import grpc
from concurrent import futures
import strategy_pb2
import strategy_pb2_grpc
from importlib import import_module
import time

class StrategyService(strategy_pb2_grpc.StrategyServiceServicer):
    def __init__(self, strategy_module):
        module = import_module(strategy_module)
        self.strategy = module.create_strategy()

    def Initialize(self, request, context):
        config = dict(request.config)
        self.strategy.initialize(config)
        return strategy_pb2.InitializeResponse(success=True)

    def OnTick(self, request, context):
        # Convert protobuf to MarketData
        market_data = self._convert_market_data(request)

        start = time.perf_counter()
        signal = self.strategy.on_tick(market_data)
        latency = (time.perf_counter() - start) * 1000  # ms

        if signal is None:
            return strategy_pb2.OnTickResponse(has_signal=False, latency_ms=latency)

        return strategy_pb2.OnTickResponse(
            has_signal=True,
            signal=self._convert_signal(signal),
            latency_ms=latency,
        )

    def OnFill(self, request, context):
        fill = self._convert_fill(request)
        self.strategy.on_fill(fill)
        return strategy_pb2.OnFillResponse(success=True)

    def GetState(self, request, context):
        state = self.strategy.get_state()
        return strategy_pb2.GetStateResponse(state=state)

    # ... conversion methods ...

def serve(strategy_module, port=50051):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    strategy_pb2_grpc.add_StrategyServiceServicer_to_server(
        StrategyService(strategy_module), server
    )
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    print(f"Python strategy server listening on port {port}")
    server.wait_for_termination()

if __name__ == '__main__':
    import sys
    if len(sys.argv) < 2:
        print("Usage: python grpc_server.py <strategy_module>")
        sys.exit(1)

    serve(sys.argv[1])
```

---

## 5. Testing Strategy

### 5.1 Unit Testing

```go
// internal/aggregator/aggregator_test.go

package aggregator

import (
    "testing"
    "time"

    "github.com/b25/strategy-engine/pkg/strategy"
    "github.com/stretchr/testify/assert"
)

func TestMajorityVote(t *testing.T) {
    agg := NewAggregator(MethodMajorityVote, 5*time.Second)

    // Add 3 LONG signals, 1 SHORT
    signals := []strategy.Signal{
        {Direction: strategy.DirectionLong, Confidence: 0.8, Symbol: "BTCUSDT", Timestamp: time.Now()},
        {Direction: strategy.DirectionLong, Confidence: 0.7, Symbol: "BTCUSDT", Timestamp: time.Now()},
        {Direction: strategy.DirectionLong, Confidence: 0.6, Symbol: "BTCUSDT", Timestamp: time.Now()},
        {Direction: strategy.DirectionShort, Confidence: 0.5, Symbol: "BTCUSDT", Timestamp: time.Now()},
    }

    for _, sig := range signals {
        agg.AddSignal(sig)
    }

    result, err := agg.Aggregate(context.Background(), "BTCUSDT")
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, strategy.DirectionLong, result.Direction)
}

func TestWeightedAverage(t *testing.T) {
    agg := NewAggregator(MethodWeightedAverage, 5*time.Second)

    // Two opposing signals
    signals := []strategy.Signal{
        {Direction: strategy.DirectionLong, Confidence: 0.9, Symbol: "BTCUSDT", Timestamp: time.Now()},
        {Direction: strategy.DirectionShort, Confidence: 0.3, Symbol: "BTCUSDT", Timestamp: time.Now()},
    }

    for _, sig := range signals {
        agg.AddSignal(sig)
    }

    result, err := agg.Aggregate(context.Background(), "BTCUSDT")
    assert.NoError(t, err)
    assert.NotNil(t, result)
    // Weighted average should favor LONG due to higher confidence
    assert.Equal(t, strategy.DirectionLong, result.Direction)
}

func TestStaleSignalPruning(t *testing.T) {
    agg := NewAggregator(MethodMajorityVote, 1*time.Second)

    // Add old signal
    oldSignal := strategy.Signal{
        Direction: strategy.DirectionLong,
        Confidence: 0.8,
        Symbol: "BTCUSDT",
        Timestamp: time.Now().Add(-2 * time.Second), // Old
    }
    agg.AddSignal(oldSignal)

    // Should be pruned
    result, err := agg.Aggregate(context.Background(), "BTCUSDT")
    assert.NoError(t, err)
    assert.Nil(t, result) // No fresh signals
}
```

### 5.2 Backtesting Framework

```go
// internal/backtest/backtester.go

package backtest

import (
    "context"
    "encoding/csv"
    "os"
    "time"

    "github.com/b25/strategy-engine/pkg/strategy"
)

type BacktestResult struct {
    Signals       []strategy.Signal
    FinalPosition float64
    PnL           float64
    SignalCount   int
    AvgLatency    time.Duration
}

type Backtester struct {
    strategy strategy.Strategy
    csvPath  string
}

func NewBacktester(s strategy.Strategy, csvPath string) *Backtester {
    return &Backtester{
        strategy: s,
        csvPath:  csvPath,
    }
}

func (b *Backtester) Run(ctx context.Context) (*BacktestResult, error) {
    // Initialize strategy
    config := &strategy.Config{Params: make(map[string]interface{})}
    if err := b.strategy.Initialize(ctx, config); err != nil {
        return nil, err
    }

    // Open CSV file
    file, err := os.Open(b.csvPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    result := &BacktestResult{
        Signals: make([]strategy.Signal, 0),
    }

    var totalLatency time.Duration

    // Replay market data
    for _, record := range records[1:] { // Skip header
        marketData := b.parseRecord(record)

        start := time.Now()
        signal, err := b.strategy.OnTick(ctx, marketData)
        latency := time.Since(start)
        totalLatency += latency

        if err != nil {
            return nil, err
        }

        if signal != nil {
            result.Signals = append(result.Signals, *signal)
            result.SignalCount++
        }
    }

    result.AvgLatency = totalLatency / time.Duration(len(records)-1)

    // Get final state
    state, err := b.strategy.GetState(ctx)
    if err != nil {
        return nil, err
    }

    result.FinalPosition = state.Position
    result.PnL = state.RealizedPnL

    return result, nil
}

func (b *Backtester) parseRecord(record []string) *strategy.MarketData {
    // Parse CSV record into MarketData
    // Format: timestamp,symbol,bid_price,bid_qty,ask_price,ask_qty
    // ... parsing logic ...
    return &strategy.MarketData{
        // ... populated fields
    }
}
```

**Backtest Test:**
```go
// strategies/momentum/momentum_test.go

package main

import (
    "context"
    "testing"

    "github.com/b25/strategy-engine/internal/backtest"
    "github.com/stretchr/testify/assert"
)

func TestMomentumBacktest(t *testing.T) {
    strategy := NewMomentumStrategy()
    bt := backtest.NewBacktester(strategy, "testdata/btcusdt_1min.csv")

    result, err := bt.Run(context.Background())
    assert.NoError(t, err)

    // Validate results
    assert.Greater(t, result.SignalCount, 0, "Should generate signals")
    assert.Less(t, result.AvgLatency.Microseconds(), int64(500), "Latency should be <500μs")

    t.Logf("Backtest Results:")
    t.Logf("  Signals: %d", result.SignalCount)
    t.Logf("  Final Position: %.2f", result.FinalPosition)
    t.Logf("  P&L: %.2f", result.PnL)
    t.Logf("  Avg Latency: %v", result.AvgLatency)
}
```

### 5.3 Performance Benchmarks

```go
// internal/benchmarks/strategy_bench_test.go

package benchmarks

import (
    "context"
    "testing"
    "time"

    "github.com/b25/strategy-engine/pkg/strategy"
)

func BenchmarkStrategyOnTick(b *testing.B) {
    // Load strategy
    s := loadTestStrategy()
    config := &strategy.Config{Params: make(map[string]interface{})}
    s.Initialize(context.Background(), config)

    // Prepare market data
    marketData := &strategy.MarketData{
        Symbol:    "BTCUSDT",
        Timestamp: time.Now(),
        OrderBook: &strategy.OrderBook{
            Bids: []strategy.PriceLevel{{Price: 50000, Quantity: 1}},
            Asks: []strategy.PriceLevel{{Price: 50001, Quantity: 1}},
        },
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := s.OnTick(context.Background(), marketData)
        if err != nil {
            b.Fatal(err)
        }
    }

    b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/1000.0, "μs/op")
}

func BenchmarkSignalAggregation(b *testing.B) {
    agg := aggregator.NewAggregator(aggregator.MethodMajorityVote, 5*time.Second)

    signals := []strategy.Signal{
        {Direction: strategy.DirectionLong, Confidence: 0.8, Symbol: "BTCUSDT", Timestamp: time.Now()},
        {Direction: strategy.DirectionLong, Confidence: 0.7, Symbol: "BTCUSDT", Timestamp: time.Now()},
        {Direction: strategy.DirectionShort, Confidence: 0.5, Symbol: "BTCUSDT", Timestamp: time.Now()},
    }

    for _, sig := range signals {
        agg.AddSignal(sig)
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := agg.Aggregate(context.Background(), "BTCUSDT")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkRiskFilter(b *testing.B) {
    filter := risk.NewConfidenceFilter(0.5)
    signal := &strategy.Signal{
        Direction:  strategy.DirectionLong,
        Confidence: 0.6,
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := filter.Check(context.Background(), signal)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Run benchmarks:**
```bash
go test -bench=. -benchmem ./internal/benchmarks/
```

**Target results:**
```
BenchmarkStrategyOnTick-8         5000    450 μs/op    <-- Must be <500μs
BenchmarkSignalAggregation-8    100000     15 μs/op
BenchmarkRiskFilter-8          1000000      2 μs/op
```

### 5.4 Integration Testing

```go
// test/integration/strategy_engine_test.go

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestStrategyEngineIntegration(t *testing.T) {
    ctx := context.Background()

    // Start NATS container
    natsContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "nats:latest",
            ExposedPorts: []string{"4222/tcp"},
            WaitingFor:   wait.ForLog("Server is ready"),
        },
        Started: true,
    })
    assert.NoError(t, err)
    defer natsContainer.Terminate(ctx)

    // Get NATS connection string
    natsHost, _ := natsContainer.Host(ctx)
    natsPort, _ := natsContainer.MappedPort(ctx, "4222")
    natsURL := fmt.Sprintf("nats://%s:%s", natsHost, natsPort.Port())

    // Start strategy engine
    engine := NewStrategyEngine(natsURL)
    go engine.Start(ctx)
    time.Sleep(1 * time.Second)

    // Load test strategy
    err = engine.LoadStrategy(ctx, "test_strategy", testStrategyPath, testConfig)
    assert.NoError(t, err)

    // Publish market data
    nc, _ := nats.Connect(natsURL)
    defer nc.Close()

    marketDataJSON := `{"symbol":"BTCUSDT","timestamp":"2025-10-02T10:00:00Z","orderbook":{"bids":[{"price":50000,"quantity":1}],"asks":[{"price":50001,"quantity":1}]}}`
    nc.Publish("market_data.BTCUSDT.orderbook", []byte(marketDataJSON))

    // Wait for signal
    time.Sleep(500 * time.Millisecond)

    // Verify strategy state
    state, err := engine.GetStrategyState(ctx, "test_strategy")
    assert.NoError(t, err)
    assert.True(t, state.IsActive)

    // Shutdown
    engine.Shutdown(ctx)
}
```

---

## 6. Deployment

### 6.1 Dockerfile

```dockerfile
# Dockerfile

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build main service
RUN CGO_ENABLED=1 GOOS=linux go build -o strategy-engine cmd/strategy-engine/main.go

# Build example strategies as plugins
RUN go build -buildmode=plugin -o plugins/momentum.so strategies/momentum/momentum.go
RUN go build -buildmode=plugin -o plugins/market_maker.so strategies/market_maker/market_maker.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates python3 python3-dev py3-pip gcc musl-dev

WORKDIR /app

# Copy binary and plugins
COPY --from=builder /app/strategy-engine .
COPY --from=builder /app/plugins ./plugins

# Copy Python strategy support
COPY strategies/python ./strategies/python
RUN pip3 install -r strategies/python/requirements.txt

# Configuration
COPY config.yaml .

# Expose metrics port
EXPOSE 9092

# Health check
HEALTHCHECK --interval=10s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9092/health || exit 1

CMD ["./strategy-engine", "--config", "config.yaml"]
```

### 6.2 Configuration

```yaml
# config.yaml

service:
  name: strategy-engine
  port: 9092
  metrics_port: 9092

nats:
  url: nats://nats:4222
  topics:
    market_data: "market_data.*.orderbook"
    fills: "fills"

plugins:
  directory: ./plugins
  config_directory: ./configs/strategies
  hot_reload: true

strategies:
  - id: momentum_btc
    plugin: momentum.so
    symbols:
      - BTCUSDT
    config:
      short_window: 5
      long_window: 20
    enabled: true

  - id: market_maker_eth
    plugin: market_maker.so
    symbols:
      - ETHUSDT
    config:
      spread: 0.001
      target_size: 0.1
      max_position: 1.0
    enabled: true

  - id: rsi_python
    type: python
    module: rsi_strategy
    port: 50051  # gRPC port for Python service
    symbols:
      - BTCUSDT
    config:
      period: 14
      overbought: 70
      oversold: 30
    enabled: true

aggregation:
  method: weighted_average  # majority_vote, weighted_average, unanimous
  window_size: 5s

risk:
  filters:
    - type: confidence
      min_confidence: 0.6
    - type: position_limit
      max_position: 10.0
    - type: drawdown
      max_drawdown: 0.2  # 20%

position_tracker:
  persistence: true
  db_path: ./data/positions.db

logging:
  level: info
  format: json

metrics:
  enabled: true
  prometheus_port: 9092
```

### 6.3 Docker Compose Integration

```yaml
# docker-compose.yml (excerpt)

services:
  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "8222:8222"
    networks:
      - trading-net

  strategy-engine:
    build: ./strategy-engine
    depends_on:
      - nats
    environment:
      - NATS_URL=nats://nats:4222
      - LOG_LEVEL=info
    volumes:
      - ./strategy-engine/configs:/app/configs
      - ./strategy-engine/plugins:/app/plugins
      - strategy-data:/app/data
    ports:
      - "9092:9092"  # Metrics
    networks:
      - trading-net
    restart: unless-stopped

volumes:
  strategy-data:

networks:
  trading-net:
    driver: bridge
```

---

## 7. Observability

### 7.1 Metrics Collection

```go
// internal/metrics/metrics.go

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Strategy execution metrics
    StrategyTickDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "strategy_tick_duration_microseconds",
            Help:    "Time taken for strategy OnTick execution",
            Buckets: []float64{10, 50, 100, 250, 500, 1000, 2000, 5000},
        },
        []string{"strategy_id"},
    )

    SignalsGenerated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "strategy_signals_total",
            Help: "Total number of signals generated",
        },
        []string{"strategy_id", "symbol", "direction"},
    )

    SignalsFiltered = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "strategy_signals_filtered_total",
            Help: "Total number of signals filtered by risk",
        },
        []string{"filter_name", "reason"},
    )

    StrategyErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "strategy_errors_total",
            Help: "Total number of strategy errors",
        },
        []string{"strategy_id", "error_type"},
    )

    ActiveStrategies = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "strategy_active_count",
            Help: "Number of active strategies",
        },
    )

    StrategyPosition = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "strategy_position",
            Help: "Current position per strategy",
        },
        []string{"strategy_id", "symbol"},
    )

    StrategyPnL = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "strategy_pnl",
            Help: "Realized P&L per strategy",
        },
        []string{"strategy_id"},
    )

    HotReloads = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "strategy_hot_reloads_total",
            Help: "Total number of hot reloads",
        },
        []string{"strategy_id", "success"},
    )
)
```

### 7.2 Grafana Dashboard

**Dashboard JSON Template (excerpt):**
```json
{
  "dashboard": {
    "title": "Strategy Engine Performance",
    "panels": [
      {
        "title": "Strategy Execution Latency (p99)",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(strategy_tick_duration_microseconds_bucket[5m]))",
            "legendFormat": "{{strategy_id}}"
          }
        ],
        "yaxes": [
          {"label": "Latency (μs)", "max": 500}
        ]
      },
      {
        "title": "Signals Generated by Strategy",
        "targets": [
          {
            "expr": "rate(strategy_signals_total[5m])",
            "legendFormat": "{{strategy_id}} - {{direction}}"
          }
        ]
      },
      {
        "title": "Risk Filter Rejections",
        "targets": [
          {
            "expr": "rate(strategy_signals_filtered_total[5m])",
            "legendFormat": "{{filter_name}}: {{reason}}"
          }
        ]
      },
      {
        "title": "Strategy Positions",
        "targets": [
          {
            "expr": "strategy_position",
            "legendFormat": "{{strategy_id}} - {{symbol}}"
          }
        ]
      },
      {
        "title": "Strategy P&L",
        "targets": [
          {
            "expr": "strategy_pnl",
            "legendFormat": "{{strategy_id}}"
          }
        ]
      }
    ]
  }
}
```

### 7.3 Signal Tracking UI

**Simple Web Dashboard:**
```go
// internal/webui/handler.go

package webui

import (
    "encoding/json"
    "net/http"
    "time"
)

type SignalEvent struct {
    Timestamp  time.Time `json:"timestamp"`
    StrategyID string    `json:"strategy_id"`
    Symbol     string    `json:"symbol"`
    Direction  string    `json:"direction"`
    Confidence float64   `json:"confidence"`
    Filtered   bool      `json:"filtered"`
    FilterReason string  `json:"filter_reason,omitempty"`
}

type UIHandler struct {
    signalHistory []SignalEvent
    maxHistory    int
}

func NewUIHandler() *UIHandler {
    return &UIHandler{
        signalHistory: make([]SignalEvent, 0, 1000),
        maxHistory:    1000,
    }
}

func (h *UIHandler) AddSignal(event SignalEvent) {
    h.signalHistory = append(h.signalHistory, event)
    if len(h.signalHistory) > h.maxHistory {
        h.signalHistory = h.signalHistory[1:]
    }
}

func (h *UIHandler) HandleSignalHistory(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(h.signalHistory)
}

func (h *UIHandler) HandleStrategyStates(w http.ResponseWriter, r *http.Request) {
    // Query all strategy states from lifecycle manager
    // Return as JSON
}
```

**Simple HTML Dashboard:**
```html
<!-- webui/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Strategy Engine Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>Strategy Engine - Live Signals</h1>

    <div id="signals-table">
        <table>
            <thead>
                <tr>
                    <th>Timestamp</th>
                    <th>Strategy</th>
                    <th>Symbol</th>
                    <th>Direction</th>
                    <th>Confidence</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody id="signals-body"></tbody>
        </table>
    </div>

    <script>
        async function fetchSignals() {
            const response = await fetch('/api/signals');
            const signals = await response.json();

            const tbody = document.getElementById('signals-body');
            tbody.innerHTML = '';

            signals.slice(-20).reverse().forEach(signal => {
                const row = tbody.insertRow();
                row.innerHTML = `
                    <td>${new Date(signal.timestamp).toLocaleTimeString()}</td>
                    <td>${signal.strategy_id}</td>
                    <td>${signal.symbol}</td>
                    <td class="${signal.direction.toLowerCase()}">${signal.direction}</td>
                    <td>${signal.confidence.toFixed(2)}</td>
                    <td>${signal.filtered ? 'Filtered: ' + signal.filter_reason : 'Passed'}</td>
                `;
            });
        }

        setInterval(fetchSignals, 1000);
        fetchSignals();
    </script>
</body>
</html>
```

---

## 8. Additional Considerations

### 8.1 Error Handling

```go
// internal/engine/error_handling.go

package engine

import (
    "context"
    "log"
    "time"
)

type ErrorHandler struct {
    maxErrors     int
    resetDuration time.Duration
    errorCounts   map[string]int
    lastReset     time.Time
}

func (h *ErrorHandler) HandleStrategyError(strategyID string, err error) {
    h.errorCounts[strategyID]++

    log.Printf("Strategy %s error: %v (count: %d)", strategyID, err, h.errorCounts[strategyID])

    // Auto-pause strategy if too many errors
    if h.errorCounts[strategyID] >= h.maxErrors {
        log.Printf("Pausing strategy %s due to excessive errors", strategyID)
        // Pause strategy via lifecycle manager
    }
}

func (h *ErrorHandler) ResetCounts() {
    if time.Since(h.lastReset) > h.resetDuration {
        h.errorCounts = make(map[string]int)
        h.lastReset = time.Now()
    }
}
```

### 8.2 Plugin Isolation

```go
// Strategies run in goroutines with panic recovery
func (e *Engine) executeStrategy(ctx context.Context, s *StrategyLifecycle, data *MarketData) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Strategy %s panicked: %v", s.ID, r)
            metrics.StrategyErrors.WithLabelValues(s.ID, "panic").Inc()
            // Auto-pause strategy
            e.lifecycleManager.Pause(s.ID)
        }
    }()

    signal, err := s.Strategy.OnTick(ctx, data)
    // ... handle signal
}
```

### 8.3 Resource Limits

```go
// Limit CPU/memory per strategy using cgroups (in production)
// Or use timeout contexts to prevent runaway strategies

func (e *Engine) executeWithTimeout(ctx context.Context, s *StrategyLifecycle, data *MarketData) (*Signal, error) {
    ctx, cancel := context.WithTimeout(ctx, 500*time.Microsecond)
    defer cancel()

    signalChan := make(chan *Signal, 1)
    errChan := make(chan error, 1)

    go func() {
        signal, err := s.Strategy.OnTick(ctx, data)
        if err != nil {
            errChan <- err
        } else {
            signalChan <- signal
        }
    }()

    select {
    case <-ctx.Done():
        return nil, fmt.Errorf("strategy timeout")
    case err := <-errChan:
        return nil, err
    case signal := <-signalChan:
        return signal, nil
    }
}
```

---

## 9. Development Timeline

**Total Estimated Time: 13 days (2 weeks + 1 day buffer)**

| Phase | Days | Deliverables | Validation |
|-------|------|--------------|------------|
| 1. Core Engine | 2 | Plugin interface, loader, lifecycle | Load 10 strategies |
| 2. Market Data | 1 | NATS consumer, router | <1ms latency |
| 3. Signals | 2 | Aggregation, validation | Algorithms tested |
| 4. Risk Filters | 2 | Filter pipeline, 3 filters | <50μs execution |
| 5. Positions | 2 | Tracker, P&L, gRPC API | Accurate P&L |
| 6. Example Strategies | 2 | 2 Go + 1 Python strategy | Backtests pass |
| 7. Hot-reload & Testing | 2 | File watcher, integration tests | <500μs p99 |

---

## 10. Success Criteria

**Technical:**
- [x] Strategy OnTick latency <500μs (p99)
- [x] Hot-reload without dropping market data ticks
- [x] Plugin isolation (one crash doesn't affect others)
- [x] 100% test coverage for aggregation and risk filters
- [x] Backtest framework validates strategies

**Functional:**
- [x] Load 5+ concurrent strategies
- [x] Aggregate signals with 3 different algorithms
- [x] Filter signals through risk pipeline
- [x] Track positions and P&L accurately
- [x] Python strategy support via gRPC

**Operational:**
- [x] Prometheus metrics exposed
- [x] Health check endpoint
- [x] Graceful shutdown
- [x] Docker containerization
- [x] Integration with Market Data Service

---

## 11. Next Steps After Completion

1. **Performance Optimization:** Profile and optimize hot paths to achieve <200μs latency
2. **Advanced Strategies:** Implement statistical arbitrage, order book imbalance
3. **ML Integration:** Add ML model inference strategies (TensorFlow Serving, ONNX)
4. **Multi-Symbol Strategies:** Support basket/pairs trading
5. **Advanced Backtesting:** Walk-forward analysis, parameter optimization
6. **Strategy Marketplace:** Plugin repository for sharing strategies
7. **Live Paper Trading:** Test strategies with real data, simulated execution

---

**End of Development Plan**

This plan provides a clear path from initial setup to production-ready Strategy Engine Service with hot-reload capabilities, multi-language plugin support, and robust risk management.
