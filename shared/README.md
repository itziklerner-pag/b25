# B25 Shared Libraries

This directory contains shared protobuf definitions and libraries used across all B25 HFT trading system services.

## Structure

```
shared/
├── proto/                  # Protocol buffer definitions
│   ├── common.proto       # Common types (timestamp, enums, etc.)
│   ├── market_data.proto  # Market data messages
│   ├── orders.proto       # Order and fill messages
│   ├── account.proto      # Account and position messages
│   ├── config.proto       # Configuration messages
│   └── gen/              # Generated code
│       ├── go/           # Go protobuf code
│       └── rust/         # Rust protobuf code
│
├── lib/
│   ├── go/               # Go shared libraries
│   │   ├── types/        # Common Go types
│   │   ├── utils/        # Utility functions
│   │   └── metrics/      # Prometheus metrics helpers
│   │
│   └── rust/             # Rust shared libraries
│       └── common/       # Common Rust utilities
│
└── schemas/              # JSON schemas (if needed)
```

## Protobuf Definitions

### common.proto
- `Timestamp`: High-precision timestamp with nanosecond accuracy
- `Exchange`: Supported exchanges enum
- `Side`: Buy/Sell enum
- `OrderType`: Limit, Market, Stop, etc.
- `TimeInForce`: GTC, IOC, FOK, POST_ONLY
- `OrderStatus`: NEW, SUBMITTED, FILLED, CANCELED, etc.
- `Decimal`: Precise financial number representation
- Common error and metadata types

### market_data.proto
- `OrderBook`: Complete order book snapshot
- `OrderBookUpdate`: Incremental order book updates
- `Trade`: Individual trade
- `AggregatedTrade`: Aggregated trades
- `Ticker`: 24hr statistics
- `Candlestick`: OHLCV data
- `MarketMetrics`: Derived metrics (spread, imbalance, microprice)
- Subscription and snapshot request/response messages

### orders.proto
- `OrderRequest`: Order placement request
- `OrderResponse`: Order placement response
- `Order`: Complete order state
- `OrderUpdate`: Order state change events
- `Fill`: Fill execution event
- `CancelOrderRequest/Response`: Order cancellation
- `ValidationResult`: Order validation results
- `RateLimiterStatus`: Rate limiter state
- `OrderMetrics`: Order statistics

### account.proto
- `Balance`: Asset balance
- `Position`: Trading position
- `AccountState`: Complete account snapshot
- `AccountUpdate`: Account state changes
- `PnLCalculation`: Profit/loss calculations
- `PerformanceMetrics`: Trading performance stats
- `RiskMetrics`: Risk exposure metrics
- `ReconciliationResult`: Exchange reconciliation

### config.proto
- `StrategyConfig`: Strategy configuration
- `RiskLimits`: Risk limit definitions
- `TradingPairConfig`: Trading pair settings
- `ExchangeConfig`: Exchange connection settings
- `SystemConfig`: System-wide configuration
- `AlertRule`: Alert rule definitions
- Configuration query and update messages

## Go Libraries

### types/
- `Decimal`: High-precision decimal type using big.Rat
- `Timestamp`: Nanosecond-precision timestamp
- `OrderBook`: In-memory order book with efficient updates

### utils/
- `IDGenerator`: Unique ID generation (orders, requests, traces)
- `CircuitBreaker`: Circuit breaker pattern implementation
- `RateLimiter`: Token bucket rate limiter

### metrics/
- `Metrics`: Prometheus metrics for HFT system
  - Latency metrics (p50, p95, p99)
  - Throughput counters
  - Business metrics (PnL, positions)
  - System health metrics

## Rust Libraries

### common/
- `Decimal`: High-precision decimal using rust_decimal
- `Timestamp`: Nanosecond-precision timestamp
- `OrderBook`: Thread-safe order book implementation
- `CircuitBreaker`: Async circuit breaker
- `RateLimiter`: Async token bucket rate limiter
- `IDGenerator`: Unique ID generation functions
- `B25Error`: Common error types

## Usage

### Generating Protobuf Code

```bash
# Generate all protobuf code
make proto

# Generate only Go code
make proto-go

# Generate only Rust code
make proto-rust

# Using buf (alternative)
make proto-buf
```

### Building Libraries

```bash
# Build Go libraries
make go-build

# Build Rust libraries
make rust-build
```

### Testing

```bash
# Run all tests
make test

# Test Go only
make go-test

# Test Rust only
make rust-test
```

### Using in Go Services

```go
import (
    "github.com/b25/shared/proto/gen/go/common"
    "github.com/b25/shared/proto/gen/go/orders"
    "github.com/b25/shared/lib/go/types"
    "github.com/b25/shared/lib/go/utils"
    "github.com/b25/shared/lib/go/metrics"
)

// Create a decimal
price := types.NewDecimal("42000.50")

// Generate an order ID
orderID := utils.GenerateClientOrderID("strategy1")

// Use circuit breaker
cb := utils.NewCircuitBreaker(utils.CircuitBreakerConfig{
    MaxFailures: 5,
    Timeout:     30 * time.Second,
})

err := cb.Execute(ctx, func() error {
    return submitOrder(order)
})

// Record metrics
metrics := metrics.NewMetrics("b25")
metrics.OrdersSubmittedTotal.WithLabelValues("binance", "BTCUSDT", "BUY", "LIMIT").Inc()
```

### Using in Rust Services

```rust
use b25_common::{
    Decimal, Timestamp, OrderBook, CircuitBreaker,
    RateLimiter, generate_client_order_id, B25Error
};

// Create a decimal
let price = Decimal::from_str("42000.50")?;

// Generate an order ID
let order_id = generate_client_order_id("strategy1");

// Use circuit breaker
let cb = CircuitBreaker::new(CircuitBreakerConfig::default());
let result = cb.execute_async(|| async {
    submit_order(order).await
}).await?;

// Use rate limiter
let limiter = RateLimiter::new(10, 10); // 10/sec, burst 10
limiter.wait().await;
```

## Installation Requirements

### Go
- Go 1.21+
- Protocol Buffers compiler (protoc)
- Go protobuf plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Rust
- Rust 1.75+
- Protocol Buffers compiler (protoc)

### Buf (Optional)
```bash
# Install buf for easier protobuf management
curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.30.0/buf-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

Or use the provided make target:
```bash
make install-tools
```

## Design Principles

1. **Zero-copy where possible**: Use protobuf for efficient serialization
2. **Type safety**: Strong typing prevents runtime errors
3. **Precision**: Decimal types avoid floating point errors
4. **Performance**: Optimized for sub-millisecond latency
5. **Reusability**: Common patterns extracted into shared libraries
6. **Testability**: All utilities include comprehensive tests

## Contributing

When adding new shared types:
1. Define protobuf messages in appropriate .proto file
2. Regenerate code with `make proto`
3. Add corresponding utility functions if needed
4. Include tests for new functionality
5. Update this README

## License

Copyright (c) 2025 B25 Trading Systems
