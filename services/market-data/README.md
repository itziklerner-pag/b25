# Market Data Service

Real-time market data ingestion, order book maintenance, and distribution service built in Rust for ultra-low latency.

## Overview

- **Language**: Rust 1.75+
- **Latency Target**: <100μs (p99) processing latency
- **Throughput**: 10,000+ updates/second per symbol
- **Exchange**: Binance Futures (configurable)

## Architecture

### Core Components

1. **WebSocket Client** (`websocket.rs`)
   - Connects to Binance Futures WebSocket API
   - Subscribes to depth updates and aggregate trades
   - Exponential backoff reconnection logic
   - Automatic ping/pong for connection health

2. **Order Book Manager** (`orderbook.rs`)
   - Maintains local order book replica
   - Processes delta updates with sequence validation
   - Thread-safe using RwLock
   - Efficient BTreeMap for price level sorting

3. **Publisher** (`publisher.rs`)
   - Redis pub/sub for distributed consumers
   - Shared memory ring buffer for local IPC
   - Dual-channel distribution for flexibility

4. **Health Server** (`health.rs`)
   - HTTP endpoints on port 9090
   - `/health` - Service health status
   - `/metrics` - Prometheus metrics export
   - `/ready` - Readiness probe

5. **Metrics** (`metrics.rs`)
   - Processing latency tracking
   - WebSocket connection status
   - Message throughput counters
   - Sequence error detection

## Building

### Prerequisites
```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Install system dependencies (Ubuntu/Debian)
sudo apt-get install pkg-config libssl-dev
```

### Development Build
```bash
cargo build
```

### Production Build (Optimized)
```bash
cargo build --release
```

### Docker Build
```bash
docker build -t b25/market-data .
```

## Configuration

Copy `config.example.yaml` to `config.yaml` and configure:

```yaml
symbols:
  - BTCUSDT
  - ETHUSDT
  - BNBUSDT
  - SOLUSDT

exchange_ws_url: "wss://fstream.binance.com/stream"
redis_url: "redis://127.0.0.1:6379"
order_book_depth: 20
health_port: 9090
shm_name: "market_data_shm"
reconnect_delay_ms: 1000
max_reconnect_delay_ms: 60000
```

## Running

### Local Development
```bash
# Run with default config
cargo run --release

# With custom config
RUST_LOG=debug cargo run --release
```

### Docker
```bash
# Run with mounted config
docker run -p 9090:9090 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  b25/market-data

# With docker-compose
docker-compose up market-data
```

## Testing

```bash
# Unit tests
cargo test

# Run specific test
cargo test orderbook_update

# With output
cargo test -- --nocapture

# Benchmarks (if implemented)
cargo bench
```

## API & Integration

### Redis Pub/Sub Channels

**Order Book Updates:**
```
Channel: orderbook:{SYMBOL}
Payload: {
  "symbol": "BTCUSDT",
  "bids": [[50000.0, 1.5], ...],
  "asks": [[50001.0, 2.0], ...],
  "last_update_id": 12345,
  "timestamp": 1234567890
}
```

**Trades:**
```
Channel: trades:{SYMBOL}
Payload: {
  "symbol": "BTCUSDT",
  "trade_id": 123456,
  "price": 50000.0,
  "quantity": 1.5,
  "timestamp": 1234567890,
  "is_buyer_maker": true
}
```

### Consuming Data

**Python Example:**
```python
import redis
import json

r = redis.Redis(host='localhost', port=6379)
pubsub = r.pubsub()
pubsub.subscribe('orderbook:BTCUSDT')

for message in pubsub.listen():
    if message['type'] == 'message':
        data = json.loads(message['data'])
        print(f"Order book update: {data}")
```

**Rust Example:**
```rust
use redis::Commands;

let client = redis::Client::open("redis://127.0.0.1/")?;
let mut con = client.get_connection()?;
let mut pubsub = con.as_pubsub();
pubsub.subscribe("orderbook:BTCUSDT")?;

loop {
    let msg = pubsub.get_message()?;
    let payload: String = msg.get_payload()?;
    println!("Received: {}", payload);
}
```

## Metrics

Prometheus metrics exposed on `http://localhost:9090/metrics`:

### WebSocket Metrics
- `websocket_connected{symbol}` - Connection status (1=connected, 0=disconnected)
- `websocket_disconnects_total{symbol}` - Total disconnections

### Processing Metrics
- `messages_processed_total{symbol,type}` - Messages processed by type
- `messages_error_total{symbol}` - Processing errors
- `processing_latency_microseconds{symbol}` - Processing latency histogram

### Order Book Metrics
- `orderbook_updates_total{symbol}` - Order book updates
- `sequence_errors_total{symbol}` - Sequence validation errors

### Trade Metrics
- `trades_processed_total{symbol}` - Trades processed

### Redis Metrics
- `redis_publishes_total{symbol,type}` - Successful publishes
- `redis_errors_total{symbol}` - Redis errors

## Health Checks

### Health Endpoint
```bash
curl http://localhost:9090/health
```

Response:
```json
{
  "status": "healthy",
  "service": "market-data",
  "version": "0.1.0"
}
```

### Readiness Endpoint
```bash
curl http://localhost:9090/ready
```

## Performance Optimization

### Latency Targets
- Message processing: <100μs (p99)
- Order book update: <50μs
- Redis publish: <200μs

### Optimization Techniques
1. **Zero-copy parsing** - Direct deserialization from WebSocket
2. **Lock-free shared memory** - Crossbeam queue for IPC
3. **Connection pooling** - Redis connection manager
4. **Batch processing** - Group updates when possible
5. **CPU pinning** - Pin threads to cores (TODO)

### Monitoring Latency
```bash
# Watch metrics for high latency
watch -n 1 'curl -s http://localhost:9090/metrics | grep processing_latency'
```

## Troubleshooting

### WebSocket Connection Issues
- Check network connectivity to Binance
- Verify exchange_ws_url in config
- Look for rate limiting (HTTP 429)

### Sequence Errors
- Normal after reconnection
- Service will auto-recover
- Monitor `sequence_errors_total` metric

### Redis Connection Issues
- Ensure Redis is running: `redis-cli ping`
- Check redis_url in config
- Verify network connectivity

### High Latency
- Check CPU usage
- Monitor Redis latency: `redis-cli --latency`
- Review `processing_latency_microseconds` histogram

## Development Roadmap

### Completed
- [x] WebSocket client with reconnection
- [x] Order book maintenance
- [x] Redis pub/sub distribution
- [x] Shared memory ring buffer
- [x] Health check endpoints
- [x] Prometheus metrics
- [x] Sequence validation

### TODO
- [ ] True shared memory using `shared_memory` crate
- [ ] Order book snapshot recovery
- [ ] QuestDB time-series storage
- [ ] WebSocket compression support
- [ ] Multi-exchange support
- [ ] Rate limiting / backpressure
- [ ] CPU pinning for threads
- [ ] Integration tests
- [ ] Load testing / benchmarks

## Dependencies

### Runtime
- Redis 6.0+ (for pub/sub)

### Optional
- QuestDB (for historical data)
- Prometheus (for metrics scraping)
- Grafana (for visualization)

## License

Proprietary - B25 Trading Platform

## Support

For issues or questions, contact the B25 platform team.
