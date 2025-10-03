# Test Architecture

Comprehensive overview of the B25 Trading System test architecture, patterns, and best practices.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                   Test Architecture                      │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐      │
│  │Integration │  │    E2E     │  │ Performance  │      │
│  │   Tests    │  │   Tests    │  │  Benchmarks  │      │
│  └─────┬──────┘  └─────┬──────┘  └──────┬───────┘      │
│        │                │                │              │
│        └────────────────┴────────────────┘              │
│                         │                                │
│                         ▼                                │
│              ┌──────────────────┐                       │
│              │  Test Utilities  │                       │
│              ├──────────────────┤                       │
│              │ - Mock Exchange  │                       │
│              │ - Data Generators│                       │
│              │ - Docker Setup   │                       │
│              └──────────────────┘                       │
│                         │                                │
│                         ▼                                │
│        ┌────────────────────────────────┐              │
│        │   Test Infrastructure          │              │
│        ├────────────────────────────────┤              │
│        │ Redis │ NATS │ PostgreSQL      │              │
│        │ TimescaleDB │ Mock Exchange    │              │
│        └────────────────────────────────┘              │
└─────────────────────────────────────────────────────────┘
```

## Test Layers

### 1. Unit Tests (Not Included - Service-Specific)

Located within each service directory:
```
services/
├── order-execution/
│   └── internal/
│       └── executor/
│           └── executor_test.go
```

**Purpose:** Test individual functions and methods in isolation

**Characteristics:**
- Fast execution (< 1ms per test)
- No external dependencies
- Mock all dependencies
- High code coverage target (80%+)

### 2. Integration Tests (`tests/integration/`)

Test service integration with infrastructure components.

```
tests/integration/
├── market_data_test.go           # Market data pipeline
├── order_flow_test.go            # Order execution flow
├── account_reconciliation_test.go # Position tracking
└── strategy_execution_test.go     # Strategy signals
```

**Purpose:** Validate service-to-infrastructure integration

**Characteristics:**
- Medium execution time (100ms - 5s per test)
- Uses real infrastructure (Redis, NATS, PostgreSQL)
- Tests data flow and state management
- Validates messaging patterns

**Test Pattern:**
```go
type IntegrationTestSuite struct {
    suite.Suite
    redisClient *redis.Client
    natsConn    *nats.Conn
    // ... other infrastructure
}

func (s *IntegrationTestSuite) SetupSuite() {
    // Connect to infrastructure
}

func (s *IntegrationTestSuite) SetupTest() {
    // Clean state before each test
}

func (s *IntegrationTestSuite) TestFeature() {
    // Arrange: Setup test data
    // Act: Execute operation
    // Assert: Verify results
}
```

### 3. End-to-End Tests (`tests/e2e/`)

Test complete system workflows.

```
tests/e2e/
├── trading_flow_test.go          # Complete trading cycle
├── failover_test.go              # Failure scenarios
└── latency_benchmark_test.go     # Performance benchmarks
```

**Purpose:** Validate complete business workflows

**Characteristics:**
- Longer execution time (5s - 60s per test)
- Tests entire system interaction
- Validates business logic end-to-end
- Performance and reliability testing

**Test Pattern:**
```go
func (s *E2ETestSuite) TestCompleteTradingCycle() {
    // Track events through entire pipeline
    events := []string{}

    // Subscribe to all relevant events
    // 1. Market Data
    // 2. Strategy Signal
    // 3. Order Creation
    // 4. Fill
    // 5. Position Update

    // Trigger flow
    s.publishMarketData(...)

    // Collect and verify events
    assert.Contains(events, "MARKET_DATA_RECEIVED")
    assert.Contains(events, "ORDER_FILLED")
    assert.Contains(events, "POSITION_UPDATED")
}
```

### 4. Performance Benchmarks

Measure system performance under various conditions.

**Metrics Tracked:**
- Latency (p50, p95, p99)
- Throughput (requests/sec)
- Resource usage (CPU, memory)
- Error rates

**Benchmark Pattern:**
```go
func (s *BenchmarkSuite) TestOrderLatency() {
    latencies := make([]time.Duration, 1000)

    for i := 0; i < 1000; i++ {
        start := time.Now()
        // Execute operation
        latencies[i] = time.Since(start)
    }

    stats := calculateStats(latencies)
    assert.Less(s.T(), stats.P99, 50*time.Millisecond)
}
```

## Test Utilities

### Mock Exchange Server

Simulates a cryptocurrency exchange for testing.

**Features:**
- REST API endpoints (create order, cancel, query)
- WebSocket streaming (market data, order updates)
- Configurable latency and behavior
- Order book simulation

**Configuration:**
```go
config := &MockExchangeConfig{
    HTTPAddr:           ":8545",
    WSAddr:             ":8546",
    OrderLatency:       10 * time.Millisecond,
    FillDelay:          50 * time.Millisecond,
    RejectRate:         0.0,
    PartialFillEnabled: false,
    MarketDataEnabled:  true,
}
```

**Usage in Tests:**
```go
// Mock exchange is automatically available at localhost:8545
order := createTestOrder()
response := httpPost("http://localhost:8545/api/v3/order", order)
assert.Equal(t, "FILLED", response.Status)
```

### Data Generators

Generate realistic test data.

**Available Generators:**

1. **OrderGenerator**
   ```go
   gen := generators.NewOrderGenerator()
   order := gen.GenerateMarketOrder("BTCUSDT", "BUY", 0.1)
   limitOrder := gen.GenerateLimitOrder("BTCUSDT", "SELL", 50000, 0.5)
   ```

2. **MarketDataGenerator**
   ```go
   gen := generators.NewMarketDataGenerator()
   tick := gen.GenerateTick("BTCUSDT")
   orderBook := gen.GenerateOrderBook("BTCUSDT", 10)
   candle := gen.GenerateCandle("BTCUSDT", 1*time.Minute)
   ```

3. **AccountDataGenerator**
   ```go
   gen := generators.NewAccountDataGenerator()
   position := gen.GeneratePosition("BTCUSDT")
   balance := gen.GenerateBalance()
   ```

4. **StrategyDataGenerator**
   ```go
   gen := generators.NewStrategyDataGenerator()
   signal := gen.GenerateSignal("momentum", "BTCUSDT")
   metrics := gen.GenerateStrategyMetrics("scalping")
   ```

5. **ScenarioGenerator**
   ```go
   gen := generators.NewScenarioGenerator()
   scenario := gen.GenerateTradingScenario("BTCUSDT", 10)
   hfScenario := gen.GenerateHighFrequencyScenario("BTCUSDT", 5*time.Second, 100*time.Millisecond)
   ```

## Test Infrastructure

### Docker Compose Setup

All infrastructure runs in Docker containers for isolation and consistency.

**Services:**

| Service | Port | Purpose |
|---------|------|---------|
| redis-test | 6380 | Caching, pub/sub |
| nats-test | 4223 | Message queue |
| postgres-test | 5433 | Relational data |
| timescale-test | 5434 | Time-series data |
| mock-exchange | 8545, 8546 | Exchange simulation |

**Health Checks:**
- All services have health checks
- Tests wait for healthy state before running
- Automatic retry on failure

**Data Persistence:**
- Volumes for data persistence (optional)
- Cleanup between test runs
- Database initialization scripts

## Test Patterns

### 1. Arrange-Act-Assert (AAA)

```go
func (s *TestSuite) TestOrderCreation() {
    // Arrange
    order := s.orderGen.GenerateMarketOrder("BTCUSDT", "BUY", 0.1)

    // Act
    err := s.createOrder(order)

    // Assert
    assert.NoError(s.T(), err)
    assert.Equal(s.T(), "SUBMITTED", order.State)
}
```

### 2. Event-Driven Testing

```go
func (s *TestSuite) TestEventFlow() {
    eventsChan := make(chan Event, 10)

    // Subscribe to events
    s.natsConn.Subscribe("events.*", func(msg *nats.Msg) {
        eventsChan <- parseEvent(msg)
    })

    // Trigger action
    s.publishMarketData(...)

    // Collect events
    events := collectEvents(eventsChan, 5*time.Second)

    // Verify event sequence
    assert.Equal(s.T(), []string{"DATA", "SIGNAL", "ORDER"}, events)
}
```

### 3. Timeout Pattern

```go
func (s *TestSuite) TestWithTimeout() {
    resultChan := make(chan Result, 1)

    // Async operation
    go func() {
        result := doAsyncOperation()
        resultChan <- result
    }()

    // Wait with timeout
    select {
    case result := <-resultChan:
        assert.NotNil(s.T(), result)
    case <-time.After(5 * time.Second):
        s.T().Fatal("Operation timeout")
    }
}
```

### 4. Graceful Skip Pattern

```go
func (s *TestSuite) TestWithDependency() {
    // Attempt operation
    result, err := s.callService()

    // Skip if service unavailable
    if err == ErrServiceUnavailable {
        s.T().Skip("Service not running - skipping test")
        return
    }

    // Continue if service available
    require.NoError(s.T(), err)
    assert.NotNil(s.T(), result)
}
```

### 5. Retry Pattern

```go
func (s *TestSuite) TestWithRetry() {
    var result Result
    var err error

    // Retry up to 3 times
    for i := 0; i < 3; i++ {
        result, err = s.operation()
        if err == nil {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }

    require.NoError(s.T(), err)
    assert.NotNil(s.T(), result)
}
```

## Best Practices

### 1. Test Independence

- Each test should be independent
- Use `SetupTest()` for clean state
- Don't rely on test execution order
- Clean up resources in `TearDownTest()`

### 2. Meaningful Assertions

```go
// Good
assert.Equal(s.T(), "FILLED", order.State, "Order should be filled after execution")

// Avoid
assert.True(s.T(), order.State == "FILLED")
```

### 3. Error Handling

```go
// Always check and handle errors
result, err := s.operation()
require.NoError(s.T(), err, "Operation should not fail")

// Use require for critical assertions
require.NotNil(s.T(), result, "Result cannot be nil")

// Use assert for non-critical assertions
assert.Greater(s.T(), result.Count, 0)
```

### 4. Logging for Debugging

```go
func (s *TestSuite) TestComplexOperation() {
    s.T().Log("Starting complex operation test")

    result := s.doOperation()
    s.T().Logf("Operation result: %+v", result)

    assert.NotNil(s.T(), result)
}
```

### 5. Performance Considerations

```go
// Avoid in tests
time.Sleep(5 * time.Second)  // Fixed sleep

// Prefer
waitForCondition(func() bool {
    return order.State == "FILLED"
}, 5*time.Second)  // Wait with timeout
```

## Performance Targets

### Latency Requirements

| Component | Target | Critical |
|-----------|--------|----------|
| Market Data Ingestion | < 100μs | < 500μs |
| Order Execution | < 10ms | < 50ms |
| Strategy Decision | < 500μs | < 2ms |
| Redis Operations | < 1ms | < 5ms |
| NATS Publish | < 500μs | < 2ms |
| End-to-End Flow | < 100ms | < 500ms |

### Throughput Requirements

| Operation | Target | Critical |
|-----------|--------|----------|
| Market Data Ticks | 10,000/sec | 1,000/sec |
| Order Submissions | 1,000/sec | 100/sec |
| Order Updates | 5,000/sec | 500/sec |
| Strategy Signals | 500/sec | 50/sec |

### Resource Limits

| Resource | Limit |
|----------|-------|
| Memory per service | < 512MB |
| CPU per service | < 50% |
| Network latency | < 1ms (local) |
| Database connections | < 100 |

## Continuous Integration

### Pre-commit Checks

```bash
# Run before committing
make lint          # Code quality
make quick-test    # Fast tests
make coverage      # Coverage check
```

### CI Pipeline

```yaml
stages:
  - lint
  - unit-test
  - integration-test
  - e2e-test
  - benchmark
  - coverage-report
```

### Quality Gates

- Code coverage: > 80%
- All tests passing
- No linter errors
- Performance benchmarks within targets
- No test skips in CI (all services must run)

## Debugging Tests

### Enable Verbose Logging

```bash
go test -v -run TestSpecificTest
```

### Use Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test -- -test.run TestSpecificTest
```

### Inspect Infrastructure

```bash
# Redis
redis-cli -p 6380 MONITOR

# NATS
curl http://localhost:8223/connz

# PostgreSQL
psql -h localhost -p 5433 -U testuser -d b25_test
```

### View Logs

```bash
# All services
docker-compose -f testutil/docker/docker-compose.test.yml logs -f

# Specific service
docker-compose -f testutil/docker/docker-compose.test.yml logs -f redis-test
```

## Extending Tests

### Adding New Integration Test

1. Create test file: `tests/integration/new_feature_test.go`
2. Use suite structure:
   ```go
   type NewFeatureTestSuite struct {
       suite.Suite
       // dependencies
   }
   ```
3. Implement SetupSuite, SetupTest, tests
4. Register suite: `suite.Run(t, new(NewFeatureTestSuite))`

### Adding New E2E Test

1. Create test file: `tests/e2e/new_flow_test.go`
2. Follow event-driven pattern
3. Track complete workflow
4. Add performance assertions

### Adding New Mock

1. Create in `testutil/mocks/`
2. Implement interface
3. Add configuration options
4. Document usage

## Metrics and Monitoring

### Test Metrics Tracked

- **Execution time** per test
- **Pass/fail rate** per suite
- **Coverage** per package
- **Performance trends** over time
- **Flaky test** detection

### Performance Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof -run TestBenchmark
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -run TestBenchmark
go tool pprof mem.prof

# Trace
go test -trace=trace.out -run TestBenchmark
go tool trace trace.out
```

## Resources

- [Testing in Go](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Docker Compose](https://docs.docker.com/compose/)
- [NATS Testing](https://docs.nats.io/nats-concepts/core-nats/testing)
- [Redis Testing](https://redis.io/topics/testing)
