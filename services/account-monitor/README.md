# Account Monitor Service

Real-time account monitoring service for tracking balances, positions, P&L, and reconciliation with exchange.

**Language**: Go 1.21+
**Port**: 9093 (Health), 50051 (gRPC), 8080 (HTTP/WebSocket)
**Development Plan**: `../../docs/service-plans/04-account-monitor-service.md`

## Features

- Real-time position tracking from NATS fill events
- WebSocket connection to exchange user data stream
- Balance tracking (free, locked, total)
- P&L calculation (realized and unrealized)
- Periodic reconciliation with exchange (every 5s)
- Risk threshold monitoring (margin ratio, leverage, drawdown)
- Alert generation on violations
- gRPC query API for account state
- TimescaleDB for historical P&L storage
- Redis for current state caching
- Prometheus metrics exposure

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL/TimescaleDB
- Redis
- NATS

### Installation

```bash
# Clone repository
cd services/account-monitor

# Install dependencies
go mod download

# Copy config
cp config.example.yaml config.yaml

# Edit configuration
# - Set BINANCE_API_KEY and BINANCE_SECRET_KEY in environment
# - Set POSTGRES_PASSWORD in environment
# - Configure NATS URL

# Run migrations (handled automatically on startup)

# Run service
go run cmd/server/main.go
```

### Using Docker

```bash
# Build image
docker build -t account-monitor:latest .

# Run with docker-compose
docker-compose up -d
```

## Configuration

Configuration is loaded from `config.yaml` and environment variables.

### Environment Variables

- `BINANCE_API_KEY`: Binance API key
- `BINANCE_SECRET_KEY`: Binance secret key
- `POSTGRES_PASSWORD`: PostgreSQL password

### Key Settings

```yaml
reconciliation:
  enabled: true
  interval: 5s  # Reconcile every 5 seconds

alerts:
  thresholds:
    min_balance: "1000.0"
    max_drawdown_pct: "-5.0"
    max_margin_ratio: "0.8"
```

## API Endpoints

### HTTP/REST API

- `GET /health` - Health check
- `GET /ready` - Readiness probe
- `GET /api/account` - Full account state (balances, positions, P&L)
- `GET /api/positions` - All positions
- `GET /api/pnl` - Current P&L report
- `GET /api/balance` - All balances
- `GET /api/alerts` - Recent alerts
- `GET /ws` - WebSocket for real-time updates

### gRPC API

See `pkg/proto/account_monitor.proto` for full API definition.

- `GetAccountState` - Get current account state
- `GetPosition` - Get position for symbol
- `GetAllPositions` - Get all positions
- `GetPnL` - Get P&L report
- `GetBalance` - Get balance for asset
- `GetPnLHistory` - Get historical P&L

## Architecture

### Components

1. **Position Manager** - Tracks positions with state machine (LONG/SHORT/FLAT)
2. **Balance Manager** - Tracks asset balances (free, locked, total)
3. **P&L Calculator** - Calculates realized and unrealized P&L
4. **Reconciler** - Periodic sync with exchange to detect drifts
5. **Alert Manager** - Monitors thresholds and publishes alerts
6. **WebSocket Client** - Receives real-time updates from exchange
7. **NATS Subscriber** - Receives fill events from order execution

### Data Flow

```
NATS (fills) ─────────┐
                      ▼
Exchange WebSocket ──► Position Manager ──► P&L Calculator
                      │                      │
                      ├─► Balance Manager ───┤
                      │                      │
                      ▼                      ▼
                   Reconciler ────────────► Alerts
                      │                      │
                      ▼                      ▼
                  TimescaleDB            NATS (alerts)
                  Redis Cache
```

### Position Tracking Logic

Positions are tracked using a weighted average entry price:

- **Opening**: Set entry price to fill price
- **Adding**: Calculate weighted average entry price
- **Reducing**: Realize P&L on closed portion
- **Reversing**: Realize P&L and set new entry price

### Reconciliation

Every 5 seconds (configurable), the service:
1. Fetches account info from exchange
2. Compares local state with exchange state
3. Detects drifts beyond tolerance (0.00001 for balance, 0.0001 for positions)
4. Auto-corrects local state
5. Publishes drift alerts

## Metrics

Exposed on `http://localhost:9093/metrics`

### Key Metrics

- `account_positions_total{symbol}` - Number of open positions
- `account_position_value_usd{symbol}` - Position value in USD
- `account_realized_pnl_usd` - Total realized P&L
- `account_unrealized_pnl_usd{symbol}` - Unrealized P&L per symbol
- `account_balance{asset}` - Balance by asset
- `account_equity_usd` - Total account equity
- `reconciliation_drift_abs` - Drift histogram
- `reconciliation_duration_seconds` - Reconciliation duration
- `alerts_triggered_total{type,severity}` - Alert counts
- `websocket_reconnects_total` - WebSocket reconnection count
- `websocket_messages_received_total{type}` - WebSocket message counts

## Database Schema

### pnl_snapshots (TimescaleDB Hypertable)

```sql
CREATE TABLE pnl_snapshots (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20),
    realized_pnl DECIMAL(20, 8) NOT NULL,
    unrealized_pnl DECIMAL(20, 8) NOT NULL,
    total_pnl DECIMAL(20, 8) NOT NULL,
    total_fees DECIMAL(20, 8) NOT NULL,
    net_pnl DECIMAL(20, 8) NOT NULL,
    win_rate DECIMAL(5, 2),
    total_trades INT
);
```

### alerts

```sql
CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    symbol VARCHAR(20),
    message TEXT NOT NULL,
    value DECIMAL(20, 8),
    threshold DECIMAL(20, 8),
    timestamp TIMESTAMPTZ NOT NULL
);
```

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/position/...
```

## Development

### Project Structure

```
account-monitor/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── alert/
│   │   └── manager.go             # Alert management
│   ├── balance/
│   │   └── manager.go             # Balance tracking
│   ├── calculator/
│   │   └── pnl.go                 # P&L calculation
│   ├── config/
│   │   └── config.go              # Configuration
│   ├── exchange/
│   │   ├── binance.go             # Binance REST client
│   │   └── websocket.go           # Binance WebSocket client
│   ├── grpcserver/
│   │   └── server.go              # gRPC server
│   ├── health/
│   │   └── checker.go             # Health checks
│   ├── metrics/
│   │   └── prometheus.go          # Prometheus metrics
│   ├── monitor/
│   │   └── monitor.go             # Main monitor orchestrator
│   ├── position/
│   │   └── manager.go             # Position tracking
│   ├── reconciliation/
│   │   └── reconciler.go          # Reconciliation logic
│   └── storage/
│       ├── postgres.go            # PostgreSQL client
│       ├── redis.go               # Redis client
│       └── nats.go                # NATS client
├── pkg/
│   └── proto/
│       └── account_monitor.proto  # gRPC definitions
├── config.example.yaml
├── Dockerfile
├── go.mod
└── README.md
```

### Building

```bash
# Build binary
go build -o bin/account-monitor ./cmd/server

# Build Docker image
docker build -t account-monitor:latest .

# Generate protobuf code (if proto changes)
protoc --go_out=. --go-grpc_out=. pkg/proto/account_monitor.proto
```

## Production Deployment

### Kubernetes

See `../../k8s/deployments/account-monitor.yaml` for deployment configuration.

### Environment Variables (Production)

```bash
BINANCE_API_KEY=<your-api-key>
BINANCE_SECRET_KEY=<your-secret-key>
POSTGRES_PASSWORD=<secure-password>
```

### Scaling Considerations

- Single instance per trading account (stateful service)
- Horizontal scaling not recommended due to state management
- Use leader election if running multiple instances
- Redis persistence for state recovery

## Monitoring

### Grafana Dashboards

Import dashboards from `../../services/metrics/grafana/dashboards/account-monitor.json`

### Alerts

Configure alerts in Prometheus/Alertmanager:
- High drift detection rate
- WebSocket disconnections
- P&L anomalies
- Balance threshold breaches

## Troubleshooting

### WebSocket Connection Issues

```bash
# Check WebSocket status
curl http://localhost:8080/health

# Check logs
docker logs account-monitor
```

### Reconciliation Drifts

Check logs for drift messages:
```
{"level":"warn","msg":"Drifts detected during reconciliation","balance_drifts":1}
```

Investigate exchange API issues or network latency.

### Database Migration Issues

Migrations run automatically on startup. To manually run:
```bash
psql -h localhost -U trading -d trading -f internal/storage/migrations/001_initial.sql
```

## Contributing

See `../../CONTRIBUTING.md` for contribution guidelines.

## License

See `../../LICENSE`

---

**Status**: Production Ready
**Maintainer**: Trading Infrastructure Team
**Last Updated**: 2025-10-03
