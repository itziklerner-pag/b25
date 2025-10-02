# Search Service

A high-performance search service for the B25 HFT trading system, providing full-text search, autocomplete, and analytics across trades, orders, strategies, market data, and logs.

## Features

- **Full-Text Search**: Fast and accurate search across all indexed documents
- **Real-Time Indexing**: Automatic indexing of new documents via NATS message queue
- **Autocomplete**: Smart query suggestions for improved user experience
- **Advanced Filtering**: Filter by date ranges, fields, and custom criteria
- **Faceted Search**: Aggregate and group results by various dimensions
- **Search Analytics**: Track popular searches, click-through rates, and search patterns
- **High Performance**: Built on Elasticsearch for sub-second query response times
- **Scalable Architecture**: Horizontal scaling support with load balancing
- **Comprehensive API**: RESTful API with JSON responses
- **Observability**: Prometheus metrics, structured logging, and health checks

## Architecture

### Components

1. **API Layer** (`internal/api/`)
   - HTTP handlers for search, autocomplete, and indexing
   - Rate limiting and CORS middleware
   - Prometheus metrics collection

2. **Search Engine** (`internal/search/`)
   - Elasticsearch client wrapper
   - Query builder and result parser
   - Index management

3. **Indexer** (`internal/indexer/`)
   - Real-time document indexing from NATS
   - Batch processing with configurable workers
   - Automatic retries and error handling

4. **Analytics** (`internal/analytics/`)
   - Search query tracking
   - Click-through rate monitoring
   - Popular search queries
   - Redis-based storage

### Data Flow

```
NATS Messages → Indexer → Elasticsearch → Search API → Client
                    ↓
                 Analytics (Redis)
```

## Installation

### Prerequisites

- Go 1.21 or higher
- Elasticsearch 8.x
- Redis 6.x or higher
- NATS 2.x or higher

### Build from Source

```bash
# Clone repository
git clone <repo-url>
cd services/search

# Download dependencies
go mod download

# Build binary
make build

# Or use Go directly
go build -o search-service ./cmd/server
```

### Docker

```bash
# Build Docker image
make docker-build

# Or use Docker directly
docker build -t b25/search-service:latest .
```

## Configuration

Configuration is managed via YAML files and environment variables. See `config.example.yaml` for all available options.

### Key Configuration Sections

#### Server
```yaml
server:
  host: "0.0.0.0"
  port: 9097
  read_timeout: 30s
  write_timeout: 30s
```

#### Elasticsearch
```yaml
elasticsearch:
  addresses:
    - "http://localhost:9200"
  username: "elastic"
  password: "changeme"
  indices:
    trades:
      name: "b25-trades"
      shards: 5
      replicas: 1
```

#### Redis
```yaml
redis:
  address: "localhost:6379"
  password: ""
  db: 0
  cache_ttl: 300s
```

#### NATS
```yaml
nats:
  url: "nats://localhost:4222"
  subjects:
    trades: "trades.>"
    orders: "orders.>"
```

### Environment Variables

All configuration values can be overridden using environment variables with the `SEARCH_` prefix:

```bash
export SEARCH_SERVER_PORT=9097
export SEARCH_ELASTICSEARCH_ADDRESSES=http://localhost:9200
export SEARCH_REDIS_ADDRESS=localhost:6379
export SEARCH_NATS_URL=nats://localhost:4222
```

## Usage

### Starting the Service

```bash
# Using binary
./search-service -config config.yaml

# Using Make
make run

# Using Docker
docker run -p 9097:9097 -p 9098:9098 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  b25/search-service:latest
```

### API Endpoints

#### Search

**POST** `/api/v1/search`

Search for documents across indices.

```json
{
  "query": "BTCUSDT",
  "index": "trades",
  "filters": {
    "strategy": "momentum"
  },
  "sort": [
    {
      "field": "timestamp",
      "order": "desc"
    }
  ],
  "from": 0,
  "size": 50,
  "highlight": true,
  "facets": ["strategy", "symbol"],
  "date_range": {
    "field": "timestamp",
    "from": "2025-01-01T00:00:00Z",
    "to": "2025-12-31T23:59:59Z"
  }
}
```

**Response:**
```json
{
  "query": "BTCUSDT",
  "total_hits": 150,
  "max_score": 0.95,
  "results": [
    {
      "index": "b25-trades",
      "id": "trade-123",
      "score": 0.95,
      "source": {
        "symbol": "BTCUSDT",
        "side": "BUY",
        "quantity": 1.5,
        "price": 50000.0
      },
      "highlights": {
        "symbol": ["<em>BTCUSDT</em>"]
      }
    }
  ],
  "facets": {
    "strategy": [
      {"key": "momentum", "count": 80},
      {"key": "arbitrage", "count": 70}
    ]
  },
  "took_ms": 45,
  "from": 0,
  "size": 50
}
```

#### Autocomplete

**POST** `/api/v1/autocomplete`

Get autocomplete suggestions.

```json
{
  "query": "BTC",
  "index": "trades",
  "size": 10
}
```

**Response:**
```json
{
  "query": "BTC",
  "suggestions": [
    "BTCUSDT",
    "BTCUSD",
    "BTCEUR"
  ],
  "took_ms": 12
}
```

#### Index Document

**POST** `/api/v1/index`

Index a single document.

```json
{
  "index": "trades",
  "id": "trade-456",
  "document": {
    "symbol": "ETHUSDT",
    "side": "SELL",
    "quantity": 10.0,
    "price": 3000.0,
    "timestamp": "2025-10-03T12:00:00Z"
  }
}
```

#### Bulk Index

**POST** `/api/v1/index/bulk`

Index multiple documents at once.

```json
{
  "documents": [
    {
      "index": "trades",
      "document": {...}
    },
    {
      "index": "orders",
      "document": {...}
    }
  ]
}
```

#### Analytics

**GET** `/api/v1/analytics/popular?limit=10`

Get popular search queries.

**GET** `/api/v1/analytics/stats`

Get search statistics.

**POST** `/api/v1/analytics/click`

Track a click on a search result.

#### Health Checks

**GET** `/health`

Service health status.

**GET** `/ready`

Readiness probe for Kubernetes.

## Indexed Document Types

### Trades
- Trade ID, symbol, side, type
- Quantity, price, value, commission, P&L
- Strategy, order ID
- Timestamp, execution time

### Orders
- Order ID, symbol, side, type, status
- Quantity, price, filled quantity
- Strategy, time in force
- Created, updated, cancelled timestamps

### Strategies
- Strategy ID, name, type, status
- Symbols, parameters
- Performance metrics
- Created, updated timestamps

### Market Data
- Symbol, timestamp
- OHLCV (Open, High, Low, Close, Volume)
- VWAP, trade count

### Logs
- Log ID, level, service, message
- Timestamp, custom fields
- Trace ID, span ID

## Monitoring

### Metrics

Prometheus metrics are exposed on port 9098 at `/metrics`.

Key metrics:
- `search_http_requests_total` - Total HTTP requests
- `search_queries_total` - Total search queries
- `search_query_duration_seconds` - Query latency
- `search_index_operations_total` - Indexing operations
- `search_elasticsearch_health_status` - Elasticsearch health

### Health Checks

- **Liveness**: `GET /health` - Returns service and dependency health
- **Readiness**: `GET /ready` - Returns whether service is ready to serve traffic

### Logging

Structured JSON logging with configurable levels (debug, info, warn, error).

```json
{
  "level": "info",
  "timestamp": "2025-10-03T12:00:00Z",
  "message": "Search completed",
  "query": "BTCUSDT",
  "total_hits": 150,
  "latency": "45ms"
}
```

## Performance

### Benchmarks

- Search query latency: <50ms (p95)
- Autocomplete latency: <20ms (p95)
- Index throughput: >10,000 docs/sec
- Concurrent queries: >1,000 req/sec

### Optimization Tips

1. **Index Sharding**: Adjust shard count based on data volume
2. **Replica Count**: Increase replicas for read-heavy workloads
3. **Batch Size**: Tune indexer batch size for throughput vs latency
4. **Worker Count**: Scale indexer workers based on message volume
5. **Cache TTL**: Adjust Redis TTL for frequently accessed data

## Development

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make integration-test

# Test coverage
make test-coverage

# Benchmarks
make benchmark
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Tidy dependencies
make mod
```

### Local Development Setup

1. Start dependencies:
```bash
# Using Docker Compose (from repo root)
docker-compose -f docker/docker-compose.dev.yml up -d elasticsearch redis nats
```

2. Copy configuration:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

3. Run service:
```bash
make run
```

## Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  search:
    image: b25/search-service:latest
    ports:
      - "9097:9097"
      - "9098:9098"
    environment:
      - SEARCH_ELASTICSEARCH_ADDRESSES=http://elasticsearch:9200
      - SEARCH_REDIS_ADDRESS=redis:6379
      - SEARCH_NATS_URL=nats://nats:4222
    depends_on:
      - elasticsearch
      - redis
      - nats
```

### Kubernetes

See `k8s/` directory in the repository root for Kubernetes manifests.

```bash
kubectl apply -f k8s/deployments/search-service.yaml
```

## Troubleshooting

### Common Issues

**Elasticsearch connection failed**
- Check Elasticsearch is running and accessible
- Verify credentials in configuration
- Check network connectivity

**Index creation failed**
- Ensure Elasticsearch user has proper permissions
- Check Elasticsearch cluster health
- Review Elasticsearch logs

**High query latency**
- Check Elasticsearch cluster health and resources
- Review query patterns and optimize filters
- Increase Elasticsearch cluster size
- Add more replicas for read scaling

**Indexer queue full**
- Increase queue buffer size in configuration
- Add more indexer workers
- Check Elasticsearch indexing performance

**Memory issues**
- Reduce batch size
- Decrease worker count
- Increase container memory limits

## Security

- **Authentication**: Use Elasticsearch API keys or basic auth
- **TLS/SSL**: Enable TLS for Elasticsearch connections in production
- **Rate Limiting**: Configured per endpoint (default: 100 req/sec)
- **CORS**: Configure allowed origins in production
- **Secrets**: Use environment variables or secrets management
- **Network**: Run services in private network, expose only necessary ports

## Performance Tuning

### Elasticsearch Settings

```yaml
# Increase heap size for better performance
ES_JAVA_OPTS: "-Xms4g -Xmx4g"

# Adjust refresh interval for indexing performance
index.refresh_interval: "30s"

# Increase bulk queue size
thread_pool.bulk.queue_size: 1000
```

### Service Settings

```yaml
# Indexer optimization
indexer:
  batch_size: 2000    # Larger batches for throughput
  workers: 8          # More workers for parallel processing
  queue_buffer: 20000 # Larger queue for burst handling

# Search optimization
search:
  timeout: 10s        # Increase for complex queries
  max_results: 50000  # Adjust based on use case
```

## API Rate Limits

Default rate limits (configurable):
- Search: 100 requests/second
- Autocomplete: 200 requests/second
- Indexing: 50 requests/second
- Analytics: 100 requests/second

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Submit a pull request

## License

See LICENSE file in repository root.

## Support

- GitHub Issues: Report bugs and request features
- Documentation: See `docs/` directory
- API Documentation: OpenAPI spec available at `/api/docs`

---

**Built for the B25 HFT Trading System**
High-performance search with real-time indexing and comprehensive analytics.
