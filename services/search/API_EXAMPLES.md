# Search Service API Examples

Complete examples for all API endpoints.

## Base URL

```
http://localhost:9097
```

## Table of Contents

- [Search](#search)
- [Autocomplete](#autocomplete)
- [Indexing](#indexing)
- [Analytics](#analytics)
- [Health Checks](#health-checks)

---

## Search

### Basic Search

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "size": 10
  }'
```

### Search Specific Index

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "momentum",
    "index": "trades",
    "size": 20
  }'
```

### Search with Filters

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTC",
    "index": "trades",
    "filters": {
      "side": "BUY",
      "strategy": "momentum"
    },
    "size": 50
  }'
```

### Search with Sorting

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "ETHUSDT",
    "index": "trades",
    "sort": [
      {
        "field": "timestamp",
        "order": "desc"
      },
      {
        "field": "quantity",
        "order": "asc"
      }
    ],
    "size": 25
  }'
```

### Search with Date Range

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "trades",
    "index": "trades",
    "date_range": {
      "field": "timestamp",
      "from": "2025-10-01T00:00:00Z",
      "to": "2025-10-03T23:59:59Z"
    },
    "size": 100
  }'
```

### Search with Pagination

```bash
# First page
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "from": 0,
    "size": 50
  }'

# Second page
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "from": 50,
    "size": 50
  }'
```

### Search with Highlighting

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "momentum strategy",
    "index": "strategies",
    "highlight": true,
    "size": 10
  }'
```

### Search with Facets

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "*",
    "index": "trades",
    "facets": ["symbol", "strategy", "side"],
    "size": 0
  }'
```

### Advanced Search Query

```bash
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTC high frequency",
    "index": "trades",
    "filters": {
      "side": "BUY",
      "strategy": "momentum"
    },
    "sort": [
      {
        "field": "timestamp",
        "order": "desc"
      }
    ],
    "date_range": {
      "field": "timestamp",
      "from": "2025-10-01T00:00:00Z",
      "to": "2025-10-31T23:59:59Z"
    },
    "from": 0,
    "size": 50,
    "highlight": true,
    "facets": ["symbol", "strategy"],
    "min_score": 0.5
  }'
```

---

## Autocomplete

### Basic Autocomplete

```bash
curl -X POST http://localhost:9097/api/v1/autocomplete \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTC",
    "size": 10
  }'
```

### Autocomplete for Specific Index

```bash
curl -X POST http://localhost:9097/api/v1/autocomplete \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mom",
    "index": "strategies",
    "field": "name",
    "size": 5
  }'
```

### Symbol Autocomplete

```bash
curl -X POST http://localhost:9097/api/v1/autocomplete \
  -H "Content-Type: application/json" \
  -d '{
    "query": "ETH",
    "index": "trades",
    "field": "symbol",
    "size": 10
  }'
```

---

## Indexing

### Index Single Trade

```bash
curl -X POST http://localhost:9097/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "index": "trades",
    "id": "trade-001",
    "document": {
      "symbol": "BTCUSDT",
      "side": "BUY",
      "type": "MARKET",
      "quantity": 1.5,
      "price": 50000.0,
      "value": 75000.0,
      "commission": 15.0,
      "pnl": 1500.0,
      "strategy": "momentum",
      "order_id": "order-001",
      "timestamp": "2025-10-03T12:00:00Z",
      "execution_time_us": 500
    }
  }'
```

### Index Single Order

```bash
curl -X POST http://localhost:9097/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "index": "orders",
    "id": "order-001",
    "document": {
      "symbol": "ETHUSDT",
      "side": "SELL",
      "type": "LIMIT",
      "status": "FILLED",
      "quantity": 10.0,
      "price": 3000.0,
      "filled_quantity": 10.0,
      "avg_fill_price": 3000.0,
      "strategy": "arbitrage",
      "time_in_force": "GTC",
      "created_at": "2025-10-03T12:00:00Z",
      "updated_at": "2025-10-03T12:01:00Z"
    }
  }'
```

### Index Single Strategy

```bash
curl -X POST http://localhost:9097/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "index": "strategies",
    "id": "strategy-001",
    "document": {
      "name": "Momentum Trading",
      "type": "MOMENTUM",
      "status": "ACTIVE",
      "symbols": ["BTCUSDT", "ETHUSDT"],
      "parameters": {
        "timeframe": "1h",
        "threshold": 0.02,
        "stop_loss": 0.01
      },
      "performance": {
        "total_trades": 150,
        "win_rate": 0.65,
        "total_pnl": 15000.0,
        "sharpe_ratio": 2.5,
        "max_drawdown": 0.05
      },
      "created_at": "2025-10-01T00:00:00Z",
      "updated_at": "2025-10-03T12:00:00Z"
    }
  }'
```

### Index Market Data

```bash
curl -X POST http://localhost:9097/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "index": "market_data",
    "document": {
      "symbol": "BTCUSDT",
      "timestamp": "2025-10-03T12:00:00Z",
      "open": 49800.0,
      "high": 50200.0,
      "low": 49500.0,
      "close": 50000.0,
      "volume": 150.5,
      "vwap": 49950.0,
      "trades": 1523
    }
  }'
```

### Index Log Entry

```bash
curl -X POST http://localhost:9097/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "index": "logs",
    "id": "log-001",
    "document": {
      "level": "INFO",
      "service": "order-execution",
      "message": "Order executed successfully",
      "timestamp": "2025-10-03T12:00:00Z",
      "fields": {
        "order_id": "order-001",
        "symbol": "BTCUSDT",
        "latency_ms": 5
      },
      "trace_id": "trace-123",
      "span_id": "span-456"
    }
  }'
```

### Bulk Index Multiple Documents

```bash
curl -X POST http://localhost:9097/api/v1/index/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "documents": [
      {
        "index": "trades",
        "id": "trade-002",
        "document": {
          "symbol": "ETHUSDT",
          "side": "BUY",
          "quantity": 5.0,
          "price": 3000.0,
          "timestamp": "2025-10-03T12:00:00Z"
        }
      },
      {
        "index": "trades",
        "id": "trade-003",
        "document": {
          "symbol": "BNBUSDT",
          "side": "SELL",
          "quantity": 20.0,
          "price": 500.0,
          "timestamp": "2025-10-03T12:01:00Z"
        }
      },
      {
        "index": "orders",
        "id": "order-002",
        "document": {
          "symbol": "ADAUSDT",
          "side": "BUY",
          "status": "NEW",
          "quantity": 1000.0,
          "price": 0.50,
          "created_at": "2025-10-03T12:02:00Z"
        }
      }
    ]
  }'
```

---

## Analytics

### Track Click on Search Result

```bash
curl -X POST http://localhost:9097/api/v1/analytics/click \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "index": "trades",
    "document_id": "trade-001",
    "position": 1
  }'
```

### Get Popular Searches

```bash
# Get top 10 popular searches
curl http://localhost:9097/api/v1/analytics/popular?limit=10

# Get top 50 popular searches
curl http://localhost:9097/api/v1/analytics/popular?limit=50
```

### Get Search Statistics

```bash
curl http://localhost:9097/api/v1/analytics/stats
```

**Example Response:**
```json
{
  "searches_by_index": {
    "trades": 1250,
    "orders": 850,
    "strategies": 320
  },
  "avg_latency_ms": 42.5,
  "total_searches_today": 2420,
  "popular_searches": [
    {
      "query": "BTCUSDT",
      "search_count": 450,
      "last_used": "2025-10-03T12:00:00Z"
    }
  ]
}
```

---

## Health Checks

### Service Health

```bash
curl http://localhost:9097/health
```

**Example Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "elasticsearch": {
    "status": "healthy",
    "latency": "5ms"
  },
  "redis": {
    "status": "healthy",
    "latency": "2ms"
  },
  "nats": {
    "status": "healthy",
    "latency": "1ms"
  },
  "timestamp": "2025-10-03T12:00:00Z"
}
```

### Readiness Probe

```bash
curl http://localhost:9097/ready
```

### Metrics (Prometheus)

```bash
curl http://localhost:9098/metrics
```

---

## Complete Workflow Example

### 1. Index Sample Data

```bash
# Index multiple trades
curl -X POST http://localhost:9097/api/v1/index/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "documents": [
      {
        "index": "trades",
        "document": {
          "symbol": "BTCUSDT",
          "side": "BUY",
          "strategy": "momentum",
          "quantity": 1.5,
          "price": 50000.0,
          "timestamp": "2025-10-03T10:00:00Z"
        }
      },
      {
        "index": "trades",
        "document": {
          "symbol": "BTCUSDT",
          "side": "SELL",
          "strategy": "momentum",
          "quantity": 1.5,
          "price": 51000.0,
          "timestamp": "2025-10-03T11:00:00Z"
        }
      },
      {
        "index": "trades",
        "document": {
          "symbol": "ETHUSDT",
          "side": "BUY",
          "strategy": "arbitrage",
          "quantity": 10.0,
          "price": 3000.0,
          "timestamp": "2025-10-03T10:30:00Z"
        }
      }
    ]
  }'
```

### 2. Search the Data

```bash
# Search for BTC trades
curl -X POST http://localhost:9097/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "index": "trades",
    "sort": [{"field": "timestamp", "order": "desc"}],
    "size": 10
  }'
```

### 3. Use Autocomplete

```bash
# Get suggestions
curl -X POST http://localhost:9097/api/v1/autocomplete \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTC",
    "size": 5
  }'
```

### 4. Track Analytics

```bash
# Track a click
curl -X POST http://localhost:9097/api/v1/analytics/click \
  -H "Content-Type: application/json" \
  -d '{
    "query": "BTCUSDT",
    "index": "trades",
    "document_id": "trade-001",
    "position": 1
  }'

# View popular searches
curl http://localhost:9097/api/v1/analytics/popular?limit=10
```

### 5. Monitor Health

```bash
# Check service health
curl http://localhost:9097/health

# View metrics
curl http://localhost:9098/metrics
```

---

## Using with Programming Languages

### Python

```python
import requests
import json

BASE_URL = "http://localhost:9097"

# Search example
def search_trades(query):
    url = f"{BASE_URL}/api/v1/search"
    payload = {
        "query": query,
        "index": "trades",
        "size": 10
    }
    response = requests.post(url, json=payload)
    return response.json()

# Index example
def index_trade(trade_data):
    url = f"{BASE_URL}/api/v1/index"
    payload = {
        "index": "trades",
        "document": trade_data
    }
    response = requests.post(url, json=payload)
    return response.json()

# Usage
results = search_trades("BTCUSDT")
print(f"Found {results['total_hits']} results")
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:9097';

// Search example
async function searchTrades(query) {
  const response = await axios.post(`${BASE_URL}/api/v1/search`, {
    query: query,
    index: 'trades',
    size: 10
  });
  return response.data;
}

// Index example
async function indexTrade(tradeData) {
  const response = await axios.post(`${BASE_URL}/api/v1/index`, {
    index: 'trades',
    document: tradeData
  });
  return response.data;
}

// Usage
searchTrades('BTCUSDT')
  .then(results => console.log(`Found ${results.total_hits} results`))
  .catch(error => console.error('Error:', error));
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

const baseURL = "http://localhost:9097"

type SearchRequest struct {
    Query string `json:"query"`
    Index string `json:"index"`
    Size  int    `json:"size"`
}

func searchTrades(query string) (*SearchResponse, error) {
    req := SearchRequest{
        Query: query,
        Index: "trades",
        Size:  10,
    }

    body, _ := json.Marshal(req)
    resp, err := http.Post(
        baseURL+"/api/v1/search",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result SearchResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

---

## Rate Limiting

Default rate limits:
- **100 requests/second** per client IP
- **Burst of 200 requests** allowed

Rate limit headers in response:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1696348800
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request",
  "message": "Query field is required"
}
```

### 429 Too Many Requests
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests"
}
```

### 500 Internal Server Error
```json
{
  "error": "Search failed",
  "message": "Elasticsearch connection timeout"
}
```

---

**For more information, see the full [README.md](README.md)**
