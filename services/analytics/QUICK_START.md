# Analytics Service - Quick Start Guide

## 5-Minute Setup

### Prerequisites
- Docker and Docker Compose installed
- 4GB RAM available
- Ports available: 9097, 9098, 9099, 5432, 6379, 9092

### Step 1: Clone and Navigate
```bash
cd /home/mm/dev/b25/services/analytics
```

### Step 2: Start Dependencies
```bash
docker-compose up -d postgres redis kafka
# Wait 30 seconds for services to be ready
```

### Step 3: Configure
```bash
cp config.example.yaml config.yaml
# config.yaml is already configured for docker-compose
```

### Step 4: Run Database Migrations
```bash
docker-compose exec postgres psql -U analytics_user -d analytics -f /docker-entrypoint-initdb.d/001_initial_schema.sql
```

### Step 5: Initialize Kafka Topics
```bash
./scripts/init-topics.sh
```

### Step 6: Start Analytics Service
```bash
# Option A: Using Docker
docker-compose up -d analytics

# Option B: Running locally
make run
```

### Step 7: Verify Service is Running
```bash
curl http://localhost:9097/health
# Should return: {"status":"healthy",...}
```

### Step 8: Send Test Events
```bash
./scripts/test-event.sh
```

### Step 9: Query Metrics
```bash
curl http://localhost:9097/api/v1/dashboard/metrics
```

## Common Commands

```bash
# View logs
docker-compose logs -f analytics

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build analytics

# Run tests
make test

# Check Prometheus metrics
curl http://localhost:9098/metrics
```

## API Quick Reference

```bash
# Track an event
curl -X POST http://localhost:9097/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "order.placed",
    "user_id": "user123",
    "properties": {"symbol": "BTCUSDT", "price": 50000}
  }'

# Query events
curl "http://localhost:9097/api/v1/events?limit=10"

# Get dashboard metrics
curl http://localhost:9097/api/v1/dashboard/metrics

# Get event statistics
curl http://localhost:9097/api/v1/events/stats
```

## Monitoring

- **Service API**: http://localhost:9097
- **Prometheus Metrics**: http://localhost:9098/metrics
- **Prometheus UI**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)

## Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose logs analytics

# Verify dependencies
docker-compose ps
```

### Database connection errors
```bash
# Restart PostgreSQL
docker-compose restart postgres

# Check PostgreSQL logs
docker-compose logs postgres
```

### No events being ingested
```bash
# Check Kafka
docker-compose logs kafka

# Verify topics exist
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

## Next Steps

1. Review full documentation in [README.md](README.md)
2. Explore API endpoints
3. Set up Grafana dashboards
4. Configure production settings
5. Set up monitoring and alerts

## Development Mode

```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Run with hot reload
make dev
```

## Production Deployment

See [README.md](README.md) for production deployment instructions.

---
Need help? Check the [Implementation Summary](IMPLEMENTATION_SUMMARY.md) for detailed information.
