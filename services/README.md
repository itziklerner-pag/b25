# Services

Backend microservices for the HFT trading system.

## Services Overview

| Service | Language | Purpose | Latency Target |
|---------|----------|---------|----------------|
| [market-data](./market-data/) | Rust | Real-time data ingestion & order book | <100μs |
| [order-execution](./order-execution/) | Go | Order lifecycle management | <10ms |
| [strategy-engine](./strategy-engine/) | Go | Trading strategy execution | <500μs |
| [account-monitor](./account-monitor/) | Go | Balance, P&L, reconciliation | <100ms |
| [dashboard-server](./dashboard-server/) | Go | WebSocket state aggregation | <50ms |
| [risk-manager](./risk-manager/) | Go | Risk management & limits | <10ms |
| [configuration](./configuration/) | Go | Configuration management | N/A |
| [metrics](./metrics/) | Config | Prometheus/Grafana setup | N/A |

## Development

Each service directory contains:
- `README.md` - Service documentation
- `Dockerfile` - Container definition
- `config.example.yaml` - Configuration template
- Source code in language-specific structure

## Building Services

### All Services (Docker)
```bash
# From repo root
docker-compose -f docker/docker-compose.dev.yml build
```

### Individual Service
```bash
cd services/<service-name>
# See service README for build instructions
```

## Running Services

### With Docker Compose (Recommended)
```bash
docker-compose -f docker/docker-compose.dev.yml up
```

### Standalone (Development)
```bash
cd services/<service-name>
# Follow service-specific instructions
```

## Testing

Each service has its own test suite:
```bash
cd services/<service-name>
# See service README for test commands
```

## Health Checks

All services expose health endpoints:
```bash
curl http://localhost:<port>/health
```

## Metrics

All services expose Prometheus metrics:
```bash
curl http://localhost:<port>/metrics
```

## Development Plans

Detailed development plans for each service are in:
- `../../docs/service-plans/`

## Dependencies

### Common Dependencies
- Redis (hot cache)
- PostgreSQL (configuration)
- TimescaleDB (time-series data)
- NATS (pub/sub messaging)

### Starting Dependencies
```bash
docker-compose -f docker/docker-compose.dev.yml up redis postgres timescaledb nats
```
