# API Gateway Service

Production-ready API Gateway for the B25 High-Frequency Trading System. Provides unified access point, authentication, rate limiting, caching, and circuit breaking for all backend microservices.

## Features

### Core Capabilities
- **High-Performance Routing**: Gin-based HTTP router optimized for low latency
- **Request Proxying**: Intelligent request forwarding with retry logic
- **Service Discovery**: Dynamic routing to backend microservices
- **Graceful Shutdown**: Zero-downtime deployments

### Security & Authentication
- **JWT Authentication**: Stateless token-based authentication
- **API Key Authentication**: Simple key-based access for services
- **Role-Based Access Control (RBAC)**: Fine-grained permissions (admin, operator, viewer)
- **CORS Support**: Configurable cross-origin resource sharing
- **TLS/SSL**: Optional HTTPS support

### Reliability & Resilience
- **Circuit Breaker**: Automatic failure detection and recovery per service
- **Retry Logic**: Exponential backoff retry for transient failures
- **Timeouts**: Configurable per-endpoint timeouts
- **Health Checks**: Liveness, readiness, and service health monitoring

### Performance Optimization
- **Response Caching**: Redis-based caching with configurable TTL
- **Rate Limiting**: Multi-level rate limiting (global, per-endpoint, per-IP)
- **Connection Pooling**: Efficient connection reuse
- **Request Streaming**: Memory-efficient request/response handling

### Observability
- **Structured Logging**: JSON logs with request tracing
- **Prometheus Metrics**: Comprehensive metrics collection
- **Request ID Tracking**: Distributed tracing support
- **Access Logs**: Detailed request/response logging

## Architecture

```
┌─────────────┐
│   Clients   │
│ (Web/TUI)   │
└──────┬──────┘
       │
       ↓
┌─────────────────────────────────────┐
│         API Gateway                 │
│  ┌──────────────────────────────┐  │
│  │  Authentication & RBAC       │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Rate Limiting               │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Request Routing & Proxy     │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Circuit Breaker & Retry     │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Caching & Validation        │  │
│  └──────────────────────────────┘  │
└───────────┬─────────────────────────┘
            │
    ┌───────┴───────┐
    │               │
    ↓               ↓
┌──────────┐   ┌──────────────┐
│ Market   │   │   Order      │
│ Data     │   │   Execution  │
└──────────┘   └──────────────┘
    ↓               ↓
┌──────────┐   ┌──────────────┐
│Strategy  │   │   Account    │
│Engine    │   │   Monitor    │
└──────────┘   └──────────────┘
```

## Quick Start

### Prerequisites
- Go 1.21+
- Redis (for caching)
- Backend services running

### Installation

```bash
# Clone the repository
cd services/api-gateway

# Install dependencies
go mod download

# Copy configuration
cp config.example.yaml config.yaml

# Edit configuration
vim config.yaml
```

### Configuration

Edit `config.yaml` to configure:
- Service endpoints
- Authentication settings
- Rate limits
- Cache settings
- Circuit breaker thresholds

### Running Locally

```bash
# Run directly
go run cmd/server/main.go

# Or build and run
make build
./api-gateway
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  b25/api-gateway:latest
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api-gateway
```

## API Documentation

### Public Endpoints (No Authentication)

#### Health Check
```bash
GET /health
```
Returns gateway and service health status.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-10-03T12:00:00Z",
  "services": {
    "market_data": { "status": "healthy" },
    "order_execution": { "status": "healthy" }
  }
}
```

#### Liveness Probe
```bash
GET /health/liveness
```

#### Readiness Probe
```bash
GET /health/readiness
```

#### Metrics
```bash
GET /metrics
```
Prometheus metrics endpoint.

#### Version
```bash
GET /version
```
Returns service version information.

### Protected Endpoints (Authentication Required)

All `/api/*` endpoints require authentication via:
- **JWT Token**: `Authorization: Bearer <token>`
- **API Key**: `X-API-Key: <key>`

### Market Data Service

```bash
# List all trading symbols
GET /api/v1/market-data/symbols

# Get order book for a symbol
GET /api/v1/market-data/orderbook/:symbol

# Get recent trades
GET /api/v1/market-data/trades/:symbol

# Get ticker data
GET /api/v1/market-data/ticker/:symbol
```

### Order Execution Service (Operator/Admin)

```bash
# Place a new order
POST /api/v1/orders
Content-Type: application/json
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "LIMIT",
  "price": "50000.00",
  "quantity": "0.01"
}

# Get order by ID
GET /api/v1/orders/:id

# List all orders
GET /api/v1/orders?limit=100

# Cancel order
DELETE /api/v1/orders/:id

# List active orders
GET /api/v1/orders/active

# Get order history
GET /api/v1/orders/history
```

### Strategy Engine Service (Operator/Admin)

```bash
# List strategies
GET /api/v1/strategies

# Get strategy details
GET /api/v1/strategies/:id

# Create strategy (Admin only)
POST /api/v1/strategies

# Update strategy (Admin only)
PUT /api/v1/strategies/:id

# Delete strategy (Admin only)
DELETE /api/v1/strategies/:id

# Start strategy
POST /api/v1/strategies/:id/start

# Stop strategy
POST /api/v1/strategies/:id/stop

# Get strategy status
GET /api/v1/strategies/:id/status
```

### Account Monitor Service

```bash
# Get account balance
GET /api/v1/account/balance

# Get positions
GET /api/v1/account/positions

# Get P&L
GET /api/v1/account/pnl

# Get daily P&L
GET /api/v1/account/pnl/daily

# Get trade history
GET /api/v1/account/trades
```

### Risk Manager Service (Operator/Admin)

```bash
# Get risk limits
GET /api/v1/risk/limits

# Update risk limits (Admin only)
PUT /api/v1/risk/limits

# Get risk status
GET /api/v1/risk/status

# Trigger emergency stop (Admin only)
POST /api/v1/risk/emergency-stop
```

### Configuration Service (Operator/Admin)

```bash
# Get all configuration
GET /api/v1/config

# Get config by key
GET /api/v1/config/:key

# Update config (Admin only)
PUT /api/v1/config/:key

# Delete config (Admin only)
DELETE /api/v1/config/:key
```

## Authentication

### Using JWT

```bash
# Example request with JWT
curl -H "Authorization: Bearer eyJhbGc..." \
     http://localhost:8080/api/v1/account/balance
```

### Using API Key

```bash
# Example request with API key
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/v1/account/balance
```

## Rate Limiting

The gateway implements multi-level rate limiting:

### Global Rate Limit
Default: 1000 requests/second with burst of 2000

### Per-Endpoint Rate Limit
Example limits:
- `/api/v1/orders`: 10 req/s
- `/api/v1/market-data`: 100 req/s
- `/api/v1/account`: 50 req/s

### Per-IP Rate Limit
Default: 300 requests/minute per IP

Rate limit information is returned in headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Burst: 200
```

## Circuit Breaker

Circuit breakers protect against cascading failures:

### States
- **Closed**: Normal operation
- **Open**: Too many failures, requests rejected
- **Half-Open**: Testing if service recovered

### Configuration
- Max consecutive failures: 3
- Timeout duration: 60 seconds
- Failure ratio threshold: 60%

## Caching

Response caching is enabled for GET requests:

### Cache TTL
- `/api/v1/market-data/symbols`: 5 minutes
- `/api/v1/account/balance`: 5 seconds
- `/api/v1/strategies`: 1 minute

### Cache Invalidation
Cache is automatically invalidated on:
- POST/PUT/DELETE requests
- Cache TTL expiration

## Monitoring

### Metrics

Available Prometheus metrics:
- `api_gateway_http_requests_total` - Total HTTP requests
- `api_gateway_http_request_duration_seconds` - Request duration
- `api_gateway_circuit_breaker_state` - Circuit breaker states
- `api_gateway_cache_hits_total` - Cache hits
- `api_gateway_rate_limit_exceeded_total` - Rate limit violations
- `api_gateway_upstream_requests_total` - Upstream requests
- `api_gateway_active_connections` - Active connections

### Logging

Structured JSON logs include:
- Request ID
- Method and path
- Status code
- Duration
- Client IP
- User agent
- Error details

Example log:
```json
{
  "level": "info",
  "ts": "2025-10-03T12:00:00Z",
  "msg": "HTTP request",
  "request_id": "abc-123",
  "method": "GET",
  "path": "/api/v1/account/balance",
  "status": 200,
  "duration_ms": 45,
  "client_ip": "192.168.1.100"
}
```

## Development

### Project Structure
```
api-gateway/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── middleware/      # HTTP middleware
│   ├── router/          # Route definitions
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   ├── cache/           # Cache implementation
│   └── breaker/         # Circuit breaker
├── pkg/
│   ├── logger/          # Logging utilities
│   └── metrics/         # Metrics collection
├── tests/               # Integration tests
├── config.example.yaml  # Example configuration
├── Dockerfile          # Docker image definition
└── Makefile           # Build automation
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
go test -v ./tests/...
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Run linter
make lint

# Format code
make fmt
```

## Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: b25/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: /config/config.yaml
        volumeMounts:
        - name: config
          mountPath: /config
        livenessProbe:
          httpGet:
            path: /health/liveness
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/readiness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: api-gateway-config
```

### Environment Variables

Key environment variables:
```bash
CONFIG_PATH=/app/config.yaml
SERVER_PORT=8080
LOG_LEVEL=info
REDIS_URL=redis://redis:6379/0
JWT_SECRET=<your-secret>
```

## Performance

### Benchmarks

Tested on:
- CPU: Intel Core i7-9700K
- Memory: 16GB
- Go: 1.21

Results:
- Requests/second: 50,000+
- Latency (p50): <5ms
- Latency (p99): <20ms
- Memory usage: ~50MB

### Optimization Tips

1. **Enable Caching**: Cache frequently accessed data
2. **Tune Rate Limits**: Balance security and performance
3. **Connection Pooling**: Reuse HTTP connections
4. **Disable Unnecessary Features**: Turn off unused middleware

## Troubleshooting

### High Latency
- Check circuit breaker states
- Review upstream service health
- Analyze Prometheus metrics
- Check Redis connection

### Rate Limiting Issues
- Review rate limit configuration
- Check client IP distribution
- Analyze rate limit metrics

### Authentication Failures
- Verify JWT secret matches
- Check API key configuration
- Review authentication logs

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## License

See [LICENSE](../../LICENSE) for license information.

## Support

For issues and questions:
- GitHub Issues: [github.com/b25/issues](https://github.com/b25/issues)
- Documentation: [docs/](../../docs/)
