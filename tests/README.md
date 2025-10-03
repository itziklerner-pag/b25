# B25 Trading System - Test Suite

Comprehensive test suite for the B25 high-frequency trading system, including integration tests, end-to-end tests, and performance benchmarks.

## Structure

```
tests/
├── integration/              # Integration tests
│   ├── market_data_test.go          # Market data pipeline tests
│   ├── order_flow_test.go           # Order submission and fill tests
│   ├── account_reconciliation_test.go # Position reconciliation tests
│   └── strategy_execution_test.go    # Strategy signal generation tests
│
├── e2e/                      # End-to-end tests
│   ├── trading_flow_test.go         # Complete trading cycle tests
│   ├── failover_test.go             # Service failure scenarios
│   └── latency_benchmark_test.go    # Performance benchmarks
│
└── testutil/                 # Test utilities
    ├── exchange/             # Mock exchange server
    ├── generators/           # Test data generators
    └── docker/              # Docker Compose for test environment
```

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Running test infrastructure (Redis, NATS, PostgreSQL)

## Quick Start

### 1. Start Test Infrastructure

```bash
cd tests/testutil/docker
docker-compose -f docker-compose.test.yml up -d
```

This starts:
- Redis (port 6380)
- NATS (port 4223)
- PostgreSQL (port 5433)
- TimescaleDB (port 5434)
- Mock Exchange Server (ports 8545, 8546)

### 2. Run Integration Tests

```bash
cd tests/integration
go test -v ./...
```

### 3. Run E2E Tests

```bash
cd tests/e2e
go test -v ./...
```

### 4. Run Performance Benchmarks

```bash
cd tests/e2e
go test -v -run=Benchmark
```

## Test Categories

### Integration Tests

Test individual service integrations with infrastructure:

- **Market Data Pipeline** (`market_data_test.go`)
  - Data ingestion and processing
  - Order book handling
  - Latency measurements
  - Multi-symbol support

- **Order Flow** (`order_flow_test.go`)
  - Order submission and lifecycle
  - Fill processing
  - Order cancellation
  - Rejection handling
  - Concurrent order execution

- **Account Reconciliation** (`account_reconciliation_test.go`)
  - Position tracking and reconciliation
  - Balance management
  - P&L calculation
  - Multi-symbol reconciliation

- **Strategy Execution** (`strategy_execution_test.go`)
  - Signal generation from market data
  - Signal to order conversion
  - Risk management
  - Multi-strategy execution

### End-to-End Tests

Test complete system workflows:

- **Trading Flow** (`trading_flow_test.go`)
  - Complete trading cycle (Market Data → Signal → Order → Fill → Position)
  - Round-trip latency
  - Multi-symbol trading
  - High-frequency scenarios

- **Failover** (`failover_test.go`)
  - NATS connection recovery
  - Redis connection handling
  - Message queue backlog
  - Circuit breaker activation
  - Graceful degradation

- **Performance Benchmarks** (`latency_benchmark_test.go`)
  - Market data ingestion latency
  - Order execution latency
  - Redis read/write performance
  - NATS publish/subscribe latency
  - End-to-end system latency
  - Throughput measurements

## Performance Targets

| Metric | Target | Critical |
|--------|--------|----------|
| Market Data Latency | < 100μs | < 500μs |
| Order Execution | < 10ms | < 50ms |
| Strategy Decision | < 500μs | < 2ms |
| Redis Operations | < 1ms | < 5ms |
| NATS Publish | < 500μs | < 2ms |

## Environment Variables

Configure test environment using these variables:

```bash
# Infrastructure
export REDIS_ADDR="localhost:6380"
export NATS_ADDR="nats://localhost:4223"
export POSTGRES_ADDR="localhost:5433"
export POSTGRES_USER="testuser"
export POSTGRES_PASSWORD="testpass"
export POSTGRES_DB="b25_test"

# Mock Exchange
export MOCK_EXCHANGE_HTTP="http://localhost:8545"
export MOCK_EXCHANGE_WS="ws://localhost:8546"
```

## Test Utilities

### Mock Exchange Server

Simulates cryptocurrency exchange with:
- REST API for orders, account info
- WebSocket for real-time updates
- Configurable latency and fill behavior
- Order book simulation

**Configuration:**
```yaml
http_addr: ":8545"
ws_addr: ":8546"
order_latency: 10ms
fill_delay: 50ms
reject_rate: 0.0
```

### Test Data Generators

Generate realistic test data:

- `OrderGenerator` - Trading orders
- `MarketDataGenerator` - Ticks, order books, candles
- `AccountDataGenerator` - Positions, balances
- `StrategyDataGenerator` - Trading signals, metrics
- `ScenarioGenerator` - Complete trading scenarios

**Usage:**
```go
gen := generators.NewOrderGenerator()
order := gen.GenerateMarketOrder("BTCUSDT", "BUY", 0.1)
```

## Running Specific Tests

### Run single test
```bash
go test -v -run TestOrderSubmission
```

### Run test suite
```bash
go test -v -run TestOrderFlowSuite
```

### Run with timeout
```bash
go test -v -timeout 30m ./...
```

### Run in parallel
```bash
go test -v -parallel 4 ./...
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Integration Tests
  run: |
    docker-compose -f tests/testutil/docker/docker-compose.test.yml up -d
    sleep 10
    cd tests/integration && go test -v ./...

- name: Run E2E Tests
  run: |
    cd tests/e2e && go test -v ./...
```

### Test Coverage

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Troubleshooting

### Tests are skipping

Tests skip when services aren't running. Ensure test infrastructure is up:

```bash
docker-compose -f tests/testutil/docker/docker-compose.test.yml ps
```

### Connection errors

Check service health:

```bash
# Redis
redis-cli -p 6380 ping

# NATS
curl http://localhost:8223

# PostgreSQL
psql -h localhost -p 5433 -U testuser -d b25_test
```

### Clean test environment

```bash
# Stop all services
docker-compose -f tests/testutil/docker/docker-compose.test.yml down -v

# Remove test data
docker volume prune -f

# Restart
docker-compose -f tests/testutil/docker/docker-compose.test.yml up -d
```

## Performance Testing

### Run latency benchmarks
```bash
go test -v -run=TestLatencyBenchmark
```

### Run throughput tests
```bash
go test -v -run=TestThroughput
```

### Generate performance report
```bash
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Use `SetupTest` and `TearDownTest` for cleanup
3. **Timeouts**: Set reasonable timeouts for async operations
4. **Logging**: Use `t.Logf()` for debugging information
5. **Skipping**: Skip tests gracefully when services are unavailable
6. **Assertions**: Use testify assertions for clear error messages

## Contributing

When adding new tests:

1. Follow existing test structure
2. Use testify suite for organization
3. Include performance assertions
4. Document test scenarios
5. Update this README

## License

Same as main project.
