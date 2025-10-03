# Strategy Engine Service

The Strategy Engine is a high-performance, plugin-based trading strategy execution system that processes market data, generates trading signals, and submits orders while enforcing risk management rules.

## Features

- **Plugin-Based Architecture**: Support for Go plugins and Python scripts
- **Multiple Strategy Types**: Built-in momentum, market-making, and scalping strategies
- **Real-Time Processing**: Subscribe to market data via Redis pub/sub
- **Event-Driven**: React to fills and position updates via NATS
- **Risk Management**: Comprehensive pre-trade risk checks
- **Signal Aggregation**: Prioritize and filter signals before execution
- **Hot Reload**: Dynamically load new strategies without restarts
- **Multiple Modes**: Live, Simulation, and Observation modes
- **Low Latency**: Target <500μs processing time per signal
- **Observability**: Prometheus metrics and structured logging

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Strategy Engine                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │  Momentum    │  │ Market Making│  │  Scalping    │ │
│  │  Strategy    │  │  Strategy    │  │  Strategy    │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
│         │                 │                 │          │
│         └─────────┬───────┴─────────┬───────┘          │
│                   │                 │                  │
│           ┌───────▼─────────────────▼────────┐        │
│           │   Signal Aggregator & Prioritizer │        │
│           └───────┬────────────────────────────┘        │
│                   │                                     │
│           ┌───────▼────────┐                           │
│           │  Risk Manager  │                           │
│           └───────┬────────┘                           │
│                   │                                     │
│           ┌───────▼────────┐                           │
│           │ Order Executor │                           │
│           └────────────────┘                           │
└─────────────────────────────────────────────────────────┘
         ▲            ▲            │
         │            │            ▼
    Redis Pub/Sub  NATS      gRPC Order
   (Market Data)  (Fills)    Execution
```

## Quick Start

```bash
# Build
go build -o bin/strategy-engine ./cmd/server

# Run
./bin/strategy-engine

# With custom config
CONFIG_PATH=config.yaml ./bin/strategy-engine
```

## Configuration

Copy the example configuration:

```bash
cp config.example.yaml config.yaml
```

### Key Configuration Options

```yaml
engine:
  mode: "simulation"  # live, simulation, observation
  signalBufferSize: 1000
  hotReload: true
  pluginsDir: "./plugins"

strategies:
  enabled:
    - momentum
    - market_making
    - scalping

risk:
  enabled: true
  maxPositionSize: 1000.0
  maxDailyLoss: 5000.0
  maxDrawdown: 0.10
```

## Execution Modes

### Live Mode

Execute real trades:

```yaml
engine:
  mode: "live"
```

### Simulation Mode

Generate signals without executing trades:

```yaml
engine:
  mode: "simulation"
```

### Observation Mode

Monitor strategies without generating signals:

```yaml
engine:
  mode: "observation"
```

## Built-in Strategies

### Momentum Strategy

Trades based on price momentum over a lookback period.

**Configuration:**
```yaml
strategies:
  configs:
    momentum:
      lookback_period: 20
      threshold: 0.02  # 2% momentum
      max_position: 1000.0
```

### Market Making Strategy

Provides liquidity by placing bid/ask quotes with inventory management.

**Configuration:**
```yaml
strategies:
  configs:
    market_making:
      spread: 0.001  # 10 bps
      order_size: 100.0
      max_inventory: 1000.0
      inventory_skew: 0.5
```

### Scalping Strategy

Fast in-and-out trades targeting small profits.

**Configuration:**
```yaml
strategies:
  configs:
    scalping:
      profit_target: 0.001  # 10 bps
      stop_loss: 0.0005  # 5 bps
      max_hold_time_seconds: 60
```

## Creating Custom Strategies

### Go Strategy

Implement the `Strategy` interface:

```go
type Strategy interface {
    Name() string
    Init(config map[string]interface{}) error
    OnMarketData(data *MarketData) ([]*Signal, error)
    OnFill(fill *Fill) error
    OnPositionUpdate(position *Position) error
    Start() error
    Stop() error
    IsRunning() bool
    GetMetrics() map[string]interface{}
}
```

### Go Plugin Strategy

See `plugins/go/example_plugin.go` for a complete example.

Build:
```bash
go build -buildmode=plugin -o my_strategy.so my_strategy.go
```

### Python Strategy

See `plugins/python/example_strategy.py` for a complete example.

## Risk Management

The engine includes comprehensive risk checks:

- **Position Limits**: Maximum position size per symbol
- **Order Value Limits**: Maximum notional value per order
- **Daily Loss Limits**: Stop trading after daily loss threshold
- **Drawdown Limits**: Monitor account drawdown
- **Rate Limiting**: Maximum orders per second/minute
- **Symbol Whitelist/Blacklist**: Control tradable symbols

## API Endpoints

### Health Check

```bash
curl http://localhost:9092/health
```

Response:
```json
{
  "status": "healthy",
  "service": "strategy-engine"
}
```

### Status

```bash
curl http://localhost:9092/status
```

Response:
```json
{
  "mode": "simulation",
  "active_strategies": 3,
  "signal_queue_size": 0
}
```

### Metrics

Prometheus metrics at:
```bash
curl http://localhost:9092/metrics
```

## Metrics

### Strategy Metrics

- `strategy_signals_total` - Total signals generated by strategy
- `strategy_errors_total` - Total strategy errors
- `strategy_latency_microseconds` - Strategy processing latency

### Market Data Metrics

- `market_data_received_total` - Market data messages received
- `market_data_latency_microseconds` - Market data processing latency

### Order Metrics

- `orders_submitted_total` - Orders submitted to execution
- `orders_rejected_total` - Orders rejected by risk manager
- `order_latency_microseconds` - Order submission latency

### Risk Metrics

- `risk_checks_total` - Risk checks performed
- `risk_violations_total` - Risk violations detected

### Engine Metrics

- `active_strategies` - Number of active strategies
- `plugin_reloads_total` - Plugin reload count
- `signal_queue_size` - Current signal queue size

## Performance

### Latency Targets

- Market data processing: <100μs
- Signal generation: <500μs
- Risk validation: <50μs
- Order submission: <10ms

### Throughput

- Market data: 100,000+ updates/sec
- Signal generation: 10,000+ signals/sec
- Order submission: 1,000+ orders/sec

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

## Docker

### Build

```bash
docker build -t strategy-engine:latest .
```

### Run

```bash
docker run -d \
  --name strategy-engine \
  -p 9092:9092 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/plugins:/app/plugins \
  strategy-engine:latest
```

## Troubleshooting

### Strategies Not Loading

Check logs for initialization errors:
```bash
grep "Failed to load strategy" /var/log/strategy-engine.log
```

### High Latency

1. Check Redis connection latency
2. Review strategy complexity
3. Enable debug logging
4. Profile with pprof

### Risk Violations

Review risk configuration and adjust limits:
```yaml
risk:
  maxPositionSize: 2000.0  # Increase limit
  maxDailyLoss: 10000.0
```

## Development

For detailed development plans, see:
`../../docs/service-plans/03-strategy-engine-service.md`

## License

Proprietary - All rights reserved
