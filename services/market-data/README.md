# Market Data Service

Real-time market data ingestion, order book maintenance, and distribution service.

## Overview

- **Language**: Rust 1.75+
- **Latency Target**: <100μs (p99)
- **Throughput**: 10,000+ updates/second per symbol

## Development Plan

See detailed development plan: `../../docs/service-plans/01-market-data-service.md`

## Architecture

- WebSocket client to exchange
- Order book maintenance with sequence validation
- Cap'n Proto serialization
- Redis Pub/Sub distribution
- QuestDB time-series storage

## Building

```bash
# Development build
cargo build

# Production build (optimized)
cargo build --release

# With Docker
docker build -t b25/market-data .
```

## Configuration

Copy `config.example.yaml` to `config.yaml` and configure:

```yaml
symbols: ["BTCUSDT", "ETHUSDT"]
exchange_ws_url: "wss://fstream.binance.com/stream"
redis_url: "redis://localhost:6379"
questdb_url: "http://localhost:9000"
order_book_depth: 20
```

## Running

```bash
# Local
cargo run --release

# Docker
docker run -v $(pwd)/config.yaml:/app/config.yaml b25/market-data

# With docker-compose
docker-compose up market-data
```

## Testing

```bash
# Unit tests
cargo test

# Integration tests (requires mock exchange)
cargo test --test integration

# Benchmarks
cargo bench
```

## Metrics

Exposed on `http://localhost:9090/metrics`:

- `websocket_connected{symbol}`
- `messages_processed_total{symbol,type}`
- `processing_latency_microseconds{symbol}`
- `order_book_updates_total{symbol}`

## Health Check

```bash
curl http://localhost:9090/health
```

## Dependencies

- Redis (pub/sub)
- QuestDB (time-series storage)

## Implementation Status

- [ ] WebSocket client
- [ ] Order book maintenance
- [ ] Redis distribution
- [ ] QuestDB integration
- [ ] Testing suite
- [ ] Observability UI
