# Quick Start Guide

Get the Market Data Service running in 5 minutes.

## Option 1: Docker Compose (Recommended)

This is the fastest way to get started. No Rust installation required.

```bash
# 1. Start the service with Redis
make compose-up

# 2. Check health
make health

# 3. View logs
make compose-logs

# 4. Check metrics
make metrics

# 5. Subscribe to data (in another terminal)
python3 examples/consumer.py BTCUSDT

# 6. Stop the service
make compose-down
```

## Option 2: Local Development

If you want to develop/modify the service:

```bash
# 1. Install Rust (if not already installed)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# 2. Install system dependencies
make install-deps

# 3. Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# 4. Build and run
make release
make run

# 5. In another terminal, subscribe to data
python3 examples/consumer.py BTCUSDT
```

## Verification

### Check Service Health
```bash
curl http://localhost:9090/health
```

Expected output:
```json
{
  "status": "healthy",
  "service": "market-data",
  "version": "0.1.0"
}
```

### Check Metrics
```bash
curl http://localhost:9090/metrics | grep websocket_connected
```

Expected output (when connected):
```
websocket_connected{symbol="BTCUSDT"} 1
websocket_connected{symbol="ETHUSDT"} 1
```

### Subscribe to Order Book Updates
```bash
# Using redis-cli
redis-cli SUBSCRIBE 'orderbook:BTCUSDT'

# Using Python consumer
python3 examples/consumer.py BTCUSDT

# Using Rust consumer (requires compilation)
cargo run --example consumer BTCUSDT
```

## Configuration

Edit `config.yaml` (or `config.example.yaml`) to customize:

```yaml
symbols:
  - BTCUSDT    # Add or remove symbols
  - ETHUSDT

exchange_ws_url: "wss://fstream.binance.com/stream"
redis_url: "redis://127.0.0.1:6379"
order_book_depth: 20
health_port: 9090
```

## Monitoring

### With Docker Compose (includes Prometheus & Grafana)
```bash
make compose-monitoring

# Access Grafana at http://localhost:3000
# Default credentials: admin/admin
# Access Prometheus at http://localhost:9091
```

### Watch Metrics in Terminal
```bash
make watch-metrics
```

## Troubleshooting

### Service won't start
- Check Redis is running: `redis-cli ping`
- Check port 9090 is available: `lsof -i :9090`
- Check logs: `make compose-logs`

### No data in Redis
- Verify WebSocket connection: `curl http://localhost:9090/metrics | grep websocket_connected`
- Check for errors in logs
- Verify network connectivity to Binance

### High latency
- Check CPU usage: `top`
- Monitor metrics: `make watch-metrics`
- Look for warnings in logs

## Next Steps

1. **Integrate with your application**: See Redis pub/sub examples in README.md
2. **Monitor performance**: Set up Prometheus/Grafana dashboards
3. **Add more symbols**: Edit config.yaml and restart
4. **Customize for other exchanges**: Modify websocket.rs

## Common Commands

```bash
make help              # Show all available commands
make build             # Build debug version
make release           # Build release version
make test              # Run tests
make docker-build      # Build Docker image
make health            # Check service health
make metrics           # View Prometheus metrics
make compose-up        # Start with Docker Compose
make compose-down      # Stop Docker Compose
```

## Performance Tips

1. **Use release builds**: Always run with `--release` flag for production
2. **Tune Redis**: Use `appendonly no` for pub/sub only workloads
3. **Monitor latency**: Watch `processing_latency_microseconds` metric
4. **CPU affinity**: Pin to specific cores for consistent latency
5. **Network**: Use low-latency network connection to exchange

## Support

- Check logs: `make compose-logs`
- View metrics: `make metrics`
- See README.md for detailed documentation
- Report issues to the B25 team
