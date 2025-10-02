# Analytics Service

High-performance analytics service for the B25 HFT trading system. Provides event tracking, metrics aggregation, and real-time analytics dashboard data.

## Features

- **High-Throughput Event Ingestion**: Kafka-based event consumer with batch processing
- **Time-Series Storage**: Optimized PostgreSQL schema with partitioning
- **Metrics Aggregation**: Multi-interval aggregation (1m, 5m, 15m, 1h, 1d)
- **Real-Time Analytics**: Dashboard metrics with Redis caching
- **RESTful API**: Comprehensive endpoints for event tracking and querying
- **Custom Events**: User-defined event types with schema validation
- **Data Retention**: Automated cleanup policies for old data
- **Prometheus Metrics**: Built-in observability

## Architecture

```
┌─────────────────┐
│  Kafka Topics   │
│  (Events)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Event Consumer │ ──────► PostgreSQL
│  (Batch Ingest) │         (Time-Series)
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Aggregation    │ ──────► Redis Cache
│  Engine         │         (Real-Time)
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  REST API       │
│  (HTTP Server)  │
└─────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- Kafka 3.0+ (optional, for event ingestion)

### Installation

1. Clone the repository and navigate to the service directory:
```bash
cd services/analytics
```

2. Copy and configure the example config:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

3. Run database migrations:
```bash
make migrate-up DB_URL="postgres://user:password@localhost/analytics"
```

4. Build and run:
```bash
make build
make run
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run -p 9097:9097 -p 9098:9098 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  b25/analytics:latest
```

### Docker Compose

```bash
# Start all dependencies (PostgreSQL, Redis, Kafka)
docker-compose up -d

# Service will be available at:
# - HTTP API: http://localhost:9097
# - Metrics: http://localhost:9098/metrics
# - Health: http://localhost:9099/health
```

## Configuration

The service is configured via `config.yaml`. Key sections:

### Server Configuration
```yaml
server:
  host: "0.0.0.0"
  port: 9097
  read_timeout: 30s
  write_timeout: 30s
```

### Database Configuration
```yaml
database:
  host: "localhost"
  port: 5432
  database: "analytics"
  user: "analytics_user"
  password: "your_password"
  max_connections: 50
```

### Event Ingestion
```yaml
analytics:
  ingestion:
    batch_size: 1000      # Events per batch
    batch_timeout: 5s     # Max wait time before flush
    workers: 4            # Number of worker goroutines
    buffer_size: 10000    # Event buffer size
```

See `config.example.yaml` for full configuration options.

## API Endpoints

### Event Tracking

**POST** `/api/v1/events`

Track a new event:
```json
{
  "event_type": "order.placed",
  "user_id": "user123",
  "session_id": "session456",
  "properties": {
    "symbol": "BTCUSDT",
    "price": 50000,
    "quantity": 0.1
  },
  "timestamp": "2025-10-03T12:00:00Z"
}
```

**GET** `/api/v1/events`

Query events:
```
GET /api/v1/events?start_time=2025-10-01T00:00:00Z&end_time=2025-10-03T23:59:59Z&event_type=order.placed&limit=100
```

### Metrics

**GET** `/api/v1/metrics`

Get aggregated metrics:
```
GET /api/v1/metrics?metric_name=events.count.order.placed&interval=1h&start_time=2025-10-01T00:00:00Z
```

**GET** `/api/v1/dashboard/metrics`

Get real-time dashboard metrics:
```json
{
  "active_users": 150,
  "events_per_second": 45.2,
  "orders_placed": 1000,
  "orders_filled": 950,
  "active_strategies": 5,
  "total_volume": 1500000.0
}
```

### Event Statistics

**GET** `/api/v1/events/stats`

Get event counts by type:
```json
{
  "total_events": 10000,
  "by_type": {
    "order.placed": 4500,
    "order.filled": 4200,
    "order.canceled": 300
  }
}
```

### Custom Events

**POST** `/api/v1/custom-events`

Define a custom event type:
```json
{
  "name": "custom.trading.signal",
  "display_name": "Trading Signal",
  "description": "Custom trading signal event",
  "schema": {
    "properties": {
      "signal_strength": {"type": "number"},
      "strategy": {"type": "string"}
    }
  },
  "is_active": true
}
```

## Database Schema

### Events Table (Partitioned)
```sql
CREATE TABLE events (
    id UUID,
    event_type VARCHAR(255),
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    properties JSONB,
    timestamp TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);
```

### Metric Aggregations
```sql
CREATE TABLE metric_aggregations (
    metric_name VARCHAR(255),
    interval VARCHAR(10),
    time_bucket TIMESTAMPTZ,
    count BIGINT,
    sum, avg, min, max DOUBLE PRECISION,
    p50, p95, p99 DOUBLE PRECISION,
    dimensions JSONB
);
```

## Performance

- **Event Ingestion**: 10,000+ events/second per worker
- **Batch Processing**: 1000 events per batch (configurable)
- **Query Latency**: <100ms for aggregated metrics (cached)
- **Database**: Partitioned tables for optimal time-series queries

## Data Retention

Automated cleanup policies:
- Raw events: 90 days (configurable)
- 1-minute aggregates: 365 days
- Hourly aggregates: 2 years
- Daily aggregates: 10 years

Cleanup runs daily and can be triggered manually:
```sql
SELECT cleanup_old_data();
```

## Monitoring

### Prometheus Metrics

Service exposes metrics at `http://localhost:9098/metrics`:

- `analytics_events_ingested_total` - Total events ingested
- `analytics_events_failed_total` - Failed event ingestions
- `analytics_batches_processed_total` - Batches processed
- `analytics_batch_duration_seconds` - Batch processing duration
- `analytics_query_duration_seconds` - Database query duration
- `analytics_cache_hits_total` - Cache hits
- `analytics_cache_misses_total` - Cache misses

### Health Checks

- **Liveness**: `GET /health` - Service is running
- **Readiness**: `GET /ready` - Service is ready to accept traffic
- **Health**: `GET /healthz` - Detailed health status

## Development

### Running Tests
```bash
# Run all tests
make test

# With coverage
make test-coverage

# Run specific package tests
go test -v ./internal/models/...
```

### Code Quality
```bash
# Format code
make fmt

# Run linters
make lint

# Vet code
go vet ./...
```

### Hot Reload Development
```bash
# Install air for hot reloading
go install github.com/cosmtrek/air@latest

# Run with hot reload
make dev
```

## Integration with B25 System

The analytics service integrates with other B25 services:

1. **Event Sources**: Consumes events from Kafka topics published by:
   - Order Execution Service
   - Strategy Engine
   - Market Data Pipeline
   - Account Monitor

2. **Event Types**: Tracks trading-specific events:
   - `order.placed`, `order.filled`, `order.canceled`
   - `strategy.started`, `strategy.stopped`
   - `signal.generated`
   - `position.opened`, `position.closed`
   - `balance.updated`

3. **Dashboard Server**: Provides metrics for:
   - Terminal UI
   - Web Dashboard

## Troubleshooting

### High Memory Usage
- Reduce `ingestion.buffer_size` in config
- Decrease `ingestion.batch_size`
- Check for slow database queries

### Slow Queries
- Ensure database indexes are present
- Enable query caching (Redis)
- Partition old data

### Event Loss
- Check Kafka consumer lag
- Increase `ingestion.workers`
- Monitor `analytics_events_failed_total` metric

## License

Part of the B25 HFT Trading System - See root LICENSE file

## Support

For issues and questions, see the main B25 documentation or create an issue in the repository.
