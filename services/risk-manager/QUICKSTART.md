# Risk Manager Service - Quick Start Guide

This guide will help you get the Risk Manager Service up and running in under 5 minutes.

## Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- Make

## Option 1: Docker Compose (Recommended)

The fastest way to run the complete stack:

```bash
# Start all services (PostgreSQL, Redis, NATS, Risk Manager, Prometheus, Grafana)
docker-compose up -d

# View logs
docker-compose logs -f risk-manager

# Check health
curl http://localhost:9095/health

# View metrics
curl http://localhost:9095/metrics

# Stop all services
docker-compose down
```

Services will be available at:
- Risk Manager gRPC: `localhost:50051`
- Risk Manager Health/Metrics: `http://localhost:9095`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- NATS: `localhost:4222`

## Option 2: Local Development

For active development with hot reloading:

### 1. Start Infrastructure

```bash
# Start only the infrastructure services
docker-compose up -d postgres redis nats
```

### 2. Run Migrations

```bash
make migrate-up
```

### 3. Run the Service

```bash
# Standard run
make run

# OR with hot reloading (requires air: go install github.com/cosmtrek/air@latest)
make dev
```

## Testing the Service

### 1. Health Check

```bash
curl http://localhost:9095/health
```

Expected: `200 OK`

### 2. Metrics Check

```bash
curl http://localhost:9095/metrics | grep risk_
```

You should see metrics like:
```
risk_current_leverage
risk_current_margin_ratio
risk_emergency_stop_active
```

### 3. gRPC Test with grpcurl

Install grpcurl:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

List available services:
```bash
grpcurl -plaintext localhost:50051 list
```

Check order risk:
```bash
grpcurl -plaintext -d '{
  "order_id": "test-order-1",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": 0.1,
  "price": 50000,
  "order_type": "LIMIT",
  "account_id": "test-account"
}' localhost:50051 riskmanager.RiskManager/CheckOrder
```

Get risk metrics:
```bash
grpcurl -plaintext -d '{
  "account_id": "test-account"
}' localhost:50051 riskmanager.RiskManager/GetRiskMetrics
```

Trigger emergency stop:
```bash
grpcurl -plaintext -d '{
  "reason": "Test emergency stop",
  "triggered_by": "developer"
}' localhost:50051 riskmanager.RiskManager/TriggerEmergencyStop
```

Check emergency stop status:
```bash
grpcurl -plaintext -d '{}' localhost:50051 riskmanager.RiskManager/GetEmergencyStopStatus
```

Re-enable trading:
```bash
grpcurl -plaintext -d '{
  "authorized_by": "developer",
  "reason": "Test completed"
}' localhost:50051 riskmanager.RiskManager/ReEnableTrading
```

## Database Access

Connect to PostgreSQL:
```bash
docker exec -it risk-manager-postgres psql -U postgres -d risk_manager
```

View policies:
```sql
SELECT id, name, type, metric, threshold, enabled FROM risk_policies;
```

View violations:
```sql
SELECT * FROM risk_violations ORDER BY violation_time DESC LIMIT 10;
```

View emergency stops:
```sql
SELECT * FROM emergency_stops ORDER BY trigger_time DESC;
```

## NATS Monitoring

View NATS monitoring dashboard:
```bash
open http://localhost:8222
```

Subscribe to alerts:
```bash
# Install NATS CLI: https://github.com/nats-io/natscli
nats sub "risk.alerts.>"
```

## Prometheus & Grafana

### Prometheus Queries

Open Prometheus: `http://localhost:9090`

Useful queries:
```promql
# Current leverage
risk_current_leverage

# Order check latency (p99)
histogram_quantile(0.99, rate(risk_order_check_duration_microseconds_bucket[5m]))

# Rejection rate
rate(risk_orders_rejected_total[1m])

# Emergency stops
increase(risk_emergency_stops_total[1h])
```

### Grafana Dashboard

1. Open Grafana: `http://localhost:3000` (admin/admin)
2. Add Prometheus datasource: `http://prometheus:9090`
3. Import or create dashboards with the metrics above

## Common Commands

```bash
# Build
make build

# Run tests
make test

# Generate protobuf code
make proto

# Format code
make fmt

# Run linter
make lint

# View logs (Docker)
docker-compose logs -f risk-manager

# Restart service (Docker)
docker-compose restart risk-manager

# Clean up
make clean
docker-compose down -v
```

## Configuration

Edit `config.yaml` or use environment variables:

```bash
export RISK_DATABASE_HOST=localhost
export RISK_REDIS_HOST=localhost
export RISK_NATS_URL=nats://localhost:4222
export RISK_LOGGING_LEVEL=debug
```

See `.env.example` for all available environment variables.

## Troubleshooting

### Service won't start

1. Check if ports are available:
```bash
lsof -i :50051  # gRPC
lsof -i :9095   # Metrics
lsof -i :5432   # PostgreSQL
lsof -i :6379   # Redis
lsof -i :4222   # NATS
```

2. Check logs:
```bash
docker-compose logs risk-manager
```

3. Verify infrastructure is running:
```bash
docker-compose ps
```

### Database connection failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Test connection
docker exec -it risk-manager-postgres pg_isready -U postgres

# Run migrations
make migrate-up
```

### Redis connection failed

```bash
# Check Redis is running
docker-compose ps redis

# Test connection
docker exec -it risk-manager-redis redis-cli ping
```

### NATS connection failed

```bash
# Check NATS is running
docker-compose ps nats

# Test connection
curl http://localhost:8222/healthz
```

## Next Steps

1. Read the full [README.md](README.md)
2. Review the [service plan](../../docs/service-plans/06-risk-manager-service.md)
3. Explore the API with grpcurl
4. Set up Grafana dashboards
5. Customize risk policies in the database
6. Integrate with other services (Account Monitor, Order Execution)

## Development Workflow

Typical development cycle:

1. Make code changes
2. Service auto-reloads (if using `make dev`)
3. Test with grpcurl or client
4. Check metrics in Prometheus
5. View alerts in NATS
6. Commit and push

Happy coding!
