# Dashboard Server Service

WebSocket-based state aggregation and real-time broadcasting service for B25 HFT Trading System.

## Overview

The Dashboard Server aggregates trading state from multiple backend services and broadcasts updates to connected UI clients (TUI and Web) with optimized update rates and efficient serialization.

**Language**: Go 1.21+
**Port**: 8080
**Development Plan**: [docs/service-plans/05-dashboard-server-service.md](../../docs/service-plans/05-dashboard-server-service.md)

## Features

- **WebSocket Server**: Real-time bidirectional communication with clients
- **Multi-Source Aggregation**: Consolidates data from market data, orders, account, positions, and strategies
- **Rate-Differentiated Broadcasting**: 100ms for TUI clients, 250ms for Web clients
- **Efficient Serialization**: MessagePack (default) or JSON with 3-5x bandwidth reduction
- **Differential Updates**: Only sends changed fields to minimize bandwidth
- **Client Subscription Management**: Clients can subscribe to specific data channels
- **Redis State Cache**: Fast state retrieval and pub/sub integration
- **REST API**: Historical data queries
- **Health Checks**: HTTP health endpoint for monitoring
- **Prometheus Metrics**: Comprehensive observability

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Dashboard Server Service                    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           WebSocket Server (port 8080)                │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐     │  │
│  │  │ TUI Client │  │ TUI Client │  │ Web Client │ ... │  │
│  │  └────────────┘  └────────────┘  └────────────┘     │  │
│  └───────────┬──────────────────────────────────────────┘  │
│              │                                               │
│  ┌───────────▼──────────────────────────────────────────┐  │
│  │         Broadcaster (100ms TUI / 250ms Web)           │  │
│  │  - Differential updates                               │  │
│  │  - MessagePack/JSON serialization                     │  │
│  └───────────┬──────────────────────────────────────────┘  │
│              │                                               │
│  ┌───────────▼──────────────────────────────────────────┐  │
│  │         State Aggregator                              │  │
│  │  - Market data cache                                  │  │
│  │  - Orders cache                                       │  │
│  │  - Positions cache                                    │  │
│  │  - Account cache                                      │  │
│  │  - Strategies cache                                   │  │
│  └───────────┬──────────────────────────────────────────┘  │
│              │                                               │
└──────────────┼───────────────────────────────────────────────┘
               │
    ┌──────────▼──────────┐
    │   Redis (Cache +    │
    │      Pub/Sub)       │
    └─────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Redis 7.0+
- Docker (optional)

### Local Development

1. **Clone and navigate**:
```bash
cd services/dashboard-server
```

2. **Install dependencies**:
```bash
go mod download
```

3. **Start Redis**:
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

4. **Configure environment**:
```bash
cp .env.example .env
# Edit .env as needed
```

5. **Run the server**:
```bash
go run cmd/server/main.go
```

6. **Verify health**:
```bash
curl http://localhost:8080/health
# Response: {"status":"ok","service":"dashboard-server"}
```

### Build Binary

```bash
go build -o bin/dashboard-server ./cmd/server
./bin/dashboard-server
```

### Docker

```bash
# Build image
docker build -t dashboard-server:latest .

# Run container
docker run -d \
  --name dashboard-server \
  -p 8080:8080 \
  -e DASHBOARD_REDIS_URL=redis:6379 \
  dashboard-server:latest
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DASHBOARD_PORT` | `8080` | HTTP server port |
| `DASHBOARD_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `DASHBOARD_REDIS_URL` | `localhost:6379` | Redis connection URL |

### Configuration File

Copy `config.example.yaml` to `config.yaml` and customize settings.

## WebSocket API

### Connection

**TUI Client (100ms updates)**:
```
ws://localhost:8080/ws?type=tui&format=msgpack
```

**Web Client (250ms updates)**:
```
ws://localhost:8080/ws?type=web&format=json
```

### Client Messages

**Subscribe to channels**:
```json
{
  "type": "subscribe",
  "channels": ["market_data", "orders", "positions", "account", "strategies"]
}
```

**Unsubscribe**:
```json
{
  "type": "unsubscribe",
  "channels": ["market_data"]
}
```

**Request full state refresh**:
```json
{
  "type": "refresh"
}
```

### Server Messages

**Full state snapshot** (MessagePack or JSON):
```json
{
  "type": "snapshot",
  "seq": 12345,
  "timestamp": "2025-10-03T10:30:00Z",
  "data": {
    "market_data": {
      "BTCUSDT": {
        "symbol": "BTCUSDT",
        "last_price": 50000.0,
        "bid_price": 49999.0,
        "ask_price": 50001.0
      }
    },
    "orders": [...],
    "positions": {...},
    "account": {...},
    "strategies": {...}
  }
}
```

**Differential update**:
```json
{
  "type": "update",
  "seq": 12346,
  "timestamp": "2025-10-03T10:30:00.100Z",
  "changes": {
    "market_data.BTCUSDT.last_price": 50100.0,
    "account.total_balance": 10250.75,
    "positions.BTCUSDT.unrealized_pnl": 125.50
  }
}
```

**Error message**:
```json
{
  "type": "error",
  "error": "Invalid subscription channel"
}
```

## REST API

### Historical Data Query

```bash
GET /api/v1/history?type=market_data&symbol=BTCUSDT&limit=100
```

**Response**:
```json
{
  "type": "market_data",
  "symbol": "BTCUSDT",
  "limit": "100",
  "data": []
}
```

## Testing

### Run Tests

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### WebSocket Client Example (Go)

```go
package main

import (
    "fmt"
    "log"

    "github.com/gorilla/websocket"
    "github.com/vmihailenco/msgpack/v5"
)

func main() {
    // Connect to Dashboard Server
    url := "ws://localhost:8080/ws?type=tui&format=msgpack"
    ws, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer ws.Close()

    // Subscribe to channels
    subscribeMsg := `{"type":"subscribe","channels":["market_data","orders","account"]}`
    ws.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))

    // Read messages
    for {
        _, message, err := ws.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            return
        }

        var serverMsg ServerMessage
        msgpack.Unmarshal(message, &serverMsg)

        fmt.Printf("Received: %s, Seq: %d\n", serverMsg.Type, serverMsg.Sequence)
    }
}
```

### WebSocket Client Example (JavaScript)

```javascript
import * as msgpack from '@msgpack/msgpack';

const ws = new WebSocket('ws://localhost:8080/ws?type=web&format=msgpack');
ws.binaryType = 'arraybuffer';

ws.onopen = () => {
    console.log('Connected');

    // Subscribe
    const msg = {
        type: 'subscribe',
        channels: ['market_data', 'orders', 'account']
    };
    ws.send(JSON.stringify(msg));
};

ws.onmessage = async (event) => {
    const buffer = new Uint8Array(event.data);
    const message = msgpack.decode(buffer);

    console.log('Received:', message.type, 'Seq:', message.seq);

    if (message.type === 'snapshot') {
        console.log('Full state:', message.data);
    } else if (message.type === 'update') {
        console.log('Changes:', message.changes);
    }
};
```

## Metrics

The service exposes Prometheus metrics at `http://localhost:8080/metrics`:

### Key Metrics

- `dashboard_connected_clients` - Number of connected WebSocket clients (by type)
- `dashboard_messages_sent_total` - Total messages sent to clients
- `dashboard_messages_received_total` - Total messages received from clients
- `dashboard_broadcast_latency_seconds` - Broadcast latency histogram
- `dashboard_serialization_duration_seconds` - Serialization time
- `dashboard_message_size_bytes` - Message size distribution
- `dashboard_client_subscriptions` - Active subscriptions per channel
- `dashboard_active_connections` - Total active WebSocket connections

### Grafana Dashboard

Import the dashboard JSON from `../../services/metrics/grafana/dashboards/dashboard-server.json` (to be created).

## Performance

### Benchmarks

- **Concurrent Connections**: Supports 100+ concurrent WebSocket clients
- **Broadcast Latency (p99)**: <50ms from state update to client
- **Memory Usage**: ~4MB per connection (Go goroutines)
- **MessagePack Payload**: 3-5x smaller than JSON
- **Serialization**: <1ms per message

### Load Testing

```bash
# Run load test with 100 clients
go run tests/load/websocket_load.go -clients=100 -duration=60s
```

## Production Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  dashboard-server:
    image: dashboard-server:latest
    ports:
      - "8080:8080"
    environment:
      - DASHBOARD_PORT=8080
      - DASHBOARD_LOG_LEVEL=info
      - DASHBOARD_REDIS_URL=redis:6379
    depends_on:
      - redis
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    restart: unless-stopped
```

### Kubernetes

See `../../k8s/deployments/dashboard-server.yaml` for Kubernetes manifests.

### Health Checks

```bash
# Health endpoint
curl http://localhost:8080/health

# Metrics endpoint
curl http://localhost:8080/metrics
```

## Development

### Project Structure

```
dashboard-server/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── server/
│   │   └── server.go            # WebSocket server
│   ├── aggregator/
│   │   └── aggregator.go        # State aggregation
│   ├── broadcaster/
│   │   └── broadcaster.go       # State broadcasting
│   ├── metrics/
│   │   └── metrics.go           # Prometheus metrics
│   └── types/
│       └── types.go             # Shared types
├── go.mod
├── go.sum
├── Dockerfile
├── .env.example
├── config.example.yaml
└── README.md
```

### Adding New State Types

1. Add type definition to `internal/types/types.go`
2. Add cache in `internal/aggregator/aggregator.go`
3. Add update method in aggregator
4. Add subscription channel support
5. Update diff computation in broadcaster

## Troubleshooting

### Common Issues

**Connection refused**:
- Verify Redis is running: `redis-cli ping`
- Check firewall rules

**High latency**:
- Check Redis performance
- Review broadcast interval settings
- Monitor Prometheus metrics

**Memory issues**:
- Check number of connected clients
- Review state cache size
- Monitor goroutine count: `curl http://localhost:8080/debug/pprof/goroutine`

### Debug Mode

```bash
DASHBOARD_LOG_LEVEL=debug go run cmd/server/main.go
```

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## License

See [LICENSE](../../LICENSE) for license information.

## Documentation

- [Development Plan](../../docs/service-plans/05-dashboard-server-service.md)
- [System Architecture](../../docs/SYSTEM_ARCHITECTURE.md)
- [API Documentation](../../docs/api/)

## Support

For issues and questions:
- GitHub Issues: [b25 issues](https://github.com/yourusername/b25/issues)
- Documentation: [docs/](../../docs/)

---

**Dashboard Server Service** - Part of the B25 HFT Trading System
