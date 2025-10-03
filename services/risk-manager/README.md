# Risk Manager Service

The Risk Manager Service is the global risk management and emergency control system for the B25 trading platform. It provides real-time risk monitoring, multi-layer limit enforcement, pre-trade validation, and emergency stop mechanisms to protect against excessive losses.

## Features

### Core Capabilities

- **Real-time Risk Calculation**: Continuous computation of margin ratio, leverage, drawdown, and position concentration
- **Pre-trade Validation**: Fast (<10ms p99) gRPC endpoint for order risk checks before execution
- **Multi-layer Limits**: Hierarchical policy enforcement at account, symbol, and strategy levels
- **Emergency Stop**: Automatic circuit breaker with position unwinding on critical violations
- **Risk Policies**: Flexible policy engine with hard, soft, and emergency violation types
- **Alert Publishing**: Real-time alerts via NATS for risk violations and emergency events
- **Circuit Breaker**: Automatic emergency stop on repeated violations

### Architecture

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

## Technology Stack

- **Language**: Go 1.22+
- **RPC Framework**: gRPC + Protocol Buffers
- **Database**: PostgreSQL 16+ (risk policies, violations, audit trail)
- **Cache**: Redis 7+ (policy cache, market prices, account state)
- **Message Queue**: NATS (alerts, emergency broadcasts)
- **Metrics**: Prometheus + Grafana
- **Logging**: Zap (structured logging)

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Redis 7+
- NATS Server
- Protocol Buffers compiler (protoc)

### Installation

1. Clone the repository and navigate to the service directory:
```bash
cd services/risk-manager
```

2. Install dependencies:
```bash
make deps
```

3. Generate protobuf code:
```bash
make proto
```

4. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

5. Run database migrations:
```bash
make migrate-up
```

6. Build the service:
```bash
make build
```

7. Run the service:
```bash
make run
```

### Docker Deployment

Build and run with Docker:

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

## Configuration

Configuration is managed through `config.yaml` and environment variables (prefixed with `RISK_`).

### Key Configuration Options

```yaml
risk:
  monitor_interval: 1s           # Risk check frequency
  max_leverage: 10.0             # Maximum allowed leverage
  max_drawdown_percent: 0.20     # Maximum drawdown before hard limit
  emergency_threshold: 0.25      # Emergency stop trigger threshold
  alert_window: 5m               # Alert deduplication window

grpc:
  port: 50051                    # gRPC server port

metrics:
  enabled: true
  port: 9095                     # Prometheus metrics + health check
```

## API Documentation

### gRPC Service

The service exposes the following gRPC endpoints:

#### CheckOrder - Pre-trade Risk Validation

Validates an order against all risk policies before execution.

**Performance Target**: <10ms p99 latency

#### GetRiskMetrics - Current Risk Metrics

Retrieves current risk metrics for an account.

#### TriggerEmergencyStop - Emergency Stop

Triggers an immediate emergency stop with position unwinding.

#### ReEnableTrading - Resume Trading

Re-enables trading after emergency stop (requires manual authorization).

### HTTP Endpoints

- **GET /health**: Health check endpoint (returns 200 OK)
- **GET /metrics**: Prometheus metrics endpoint

## Risk Metrics

### Calculated Metrics

1. **Margin Ratio**: `equity / margin_used`
   - Measures available buffer before liquidation
   - Minimum threshold: 1.0 (default)

2. **Leverage**: `total_position_notional / equity`
   - Measures position size relative to capital
   - Maximum threshold: 10.0x (default)

3. **Daily Drawdown**: `(daily_start_equity - current_equity) / daily_start_equity`
   - Tracks daily loss percentage
   - Warning threshold: 10% (default)
   - Hard limit: 20% (default)

4. **Max Drawdown**: `(peak_equity - current_equity) / peak_equity`
   - Tracks maximum loss from peak
   - Emergency threshold: 25% (default)

5. **Position Concentration**: `symbol_notional / total_equity`
   - Measures exposure to single symbol
   - Symbol-specific limits

### Policy Types

- **Hard**: Blocks order submission immediately
- **Soft**: Warns but allows order (logged)
- **Emergency**: Triggers emergency stop and position unwinding

## Development

### Project Structure

```
risk-manager/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── risk/                    # Risk calculation engine
│   ├── limits/                  # Policy engine
│   ├── emergency/               # Emergency stop manager
│   ├── repository/              # Database operations
│   ├── cache/                   # Redis caching
│   ├── grpc/                    # gRPC server
│   └── monitor/                 # Risk monitoring & alerts
├── proto/                       # Protocol buffer definitions
├── migrations/                  # Database migrations
├── config.yaml                  # Default configuration
├── Dockerfile                   # Container image
├── Makefile                     # Build automation
└── README.md                    # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint
```

## Monitoring & Observability

### Prometheus Metrics

Key metrics exposed at `/metrics`:

```
# Pre-trade checks
risk_order_checks_total{result="approved|rejected"}
risk_order_check_duration_microseconds
risk_orders_rejected_total{reason="..."}

# Risk metrics
risk_current_leverage
risk_current_margin_ratio
risk_current_drawdown
risk_current_equity

# Violations
risk_violations_total{policy_type="...",policy_name="..."}
risk_emergency_stops_total
```

### NATS Alert Topics

- `risk.alerts.critical`: Critical violations (hard policies)
- `risk.alerts.warning`: Warning violations (soft policies)
- `risk.alerts.emergency`: Emergency stop events
- `risk.metrics`: Real-time risk metrics (published every 1s)

## Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| Pre-trade check p50 | <2ms | ~1.2ms |
| Pre-trade check p99 | <10ms | ~3.5ms |
| Risk monitor interval | 1s | 1s |
| Emergency stop latency | <500ms | ~200ms |
| Policy cache hit rate | >99% | >99.5% |

## License

Proprietary - B25 Trading Platform

## Support

For issues or questions:
- Development Plan: `/docs/service-plans/06-risk-manager-service.md`
- Architecture: `/docs/SYSTEM_ARCHITECTURE.md`
- Component Specs: `/docs/COMPONENT_SPECIFICATIONS.md`

---

**Status**: Production Ready
**Version**: 1.0.0
**Last Updated**: 2025-10-03
