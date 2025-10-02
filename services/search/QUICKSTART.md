# Search Service - Quick Start Guide

Get the B25 Search Service running in under 5 minutes!

## Prerequisites

- Docker and Docker Compose installed
- 4GB RAM available
- Ports 9097, 9098, 9200, 6379, 4222 available

## Quick Start

### 1. Start All Services

```bash
cd services/search
docker-compose up -d
```

This will start:
- Elasticsearch (port 9200)
- Redis (port 6379)
- NATS (port 4222)
- Search Service (ports 9097, 9098)

### 2. Wait for Services to be Ready

```bash
# Wait for health checks (about 30-60 seconds)
docker-compose ps

# Check search service health
curl http://localhost:9097/health
```

### 3. Try Your First Search

```bash
# Index a sample trade
curl -X POST http://localhost:9097/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "index": "trades",
    "id": "trade-001",
    "document": {
      "symbol": "BTCUSDT",
      "side": "BUY",
      "quantity": 1.5,
      "price": 50000.0,
      "strategy": "momentum",
      "timestamp": "2025-10-03T12:00:00Z"
    }
  }'

# Search for it
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "index": "trades",
    "size": 10
  }'
```

### 4. Try Autocomplete

```bash
curl -X POST http://localhost:9097/api/v1/autocomplete \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTC",
    "size": 5
  }'
```

### 5. Check Metrics

```bash
curl http://localhost:9098/metrics
```

## View Logs

```bash
# All services
docker-compose logs -f

# Search service only
docker-compose logs -f search-service

# Elasticsearch
docker-compose logs -f elasticsearch
```

## Stop Services

```bash
docker-compose down

# Remove data volumes (clean slate)
docker-compose down -v
```

## What's Next?

- Read the full [README.md](README.md) for detailed documentation
- Explore the [API endpoints](README.md#api-endpoints)
- Configure for production use
- Integrate with your trading system via NATS

## Common Commands

```bash
# Restart search service
docker-compose restart search-service

# View Elasticsearch indices
curl http://localhost:9200/_cat/indices?v

# Check Redis keys
docker-compose exec redis redis-cli KEYS "*"

# Execute into search service container
docker-compose exec search-service sh
```

## Troubleshooting

**Services won't start:**
```bash
docker-compose down -v
docker system prune -f
docker-compose up -d
```

**Port conflicts:**
Edit `docker-compose.yml` and change conflicting port mappings.

**Out of memory:**
Increase Docker memory limit in Docker Desktop settings.

## Example: Bulk Indexing

```bash
curl -X POST http://localhost:9097/api/v1/index/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "documents": [
      {
        "index": "trades",
        "document": {
          "symbol": "ETHUSDT",
          "side": "SELL",
          "quantity": 10.0,
          "price": 3000.0,
          "timestamp": "2025-10-03T12:01:00Z"
        }
      },
      {
        "index": "trades",
        "document": {
          "symbol": "BNBUSDT",
          "side": "BUY",
          "quantity": 5.0,
          "price": 500.0,
          "timestamp": "2025-10-03T12:02:00Z"
        }
      }
    ]
  }'
```

## Example: Advanced Search

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTC",
    "index": "trades",
    "filters": {
      "side": "BUY"
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
    "facets": ["symbol", "strategy"],
    "date_range": {
      "field": "timestamp",
      "from": "2025-10-01T00:00:00Z",
      "to": "2025-10-31T23:59:59Z"
    }
  }'
```

## Production Deployment

For production deployment:
1. Use external Elasticsearch cluster
2. Enable TLS/SSL
3. Configure authentication
4. Set up proper monitoring
5. Configure backups
6. Review security settings

See [README.md](README.md#deployment) for details.

---

**Need Help?**
- Check [README.md](README.md) for full documentation
- Review logs: `docker-compose logs -f`
- Check health: `curl http://localhost:9097/health`
