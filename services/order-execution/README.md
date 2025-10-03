# Order Execution Service

High-performance order lifecycle management and exchange communication service for HFT trading systems.

**Language**: Go 1.21+
**Target Latency**: <10ms
**Development Plan**: `../../docs/service-plans/02-order-execution-service.md`

## Features

### Core Functionality
- **gRPC Server**: High-performance order management API
- **Exchange Integration**: Binance Futures REST API with HMAC SHA256 signing
- **Order Validation**: Price precision, minimum notional, and risk limit checks
- **State Machine**: Order lifecycle (NEW → SUBMITTED → FILLED/CANCELED/REJECTED)
- **Rate Limiting**: Client-side protection against exchange rate limits
- **Circuit Breaker**: Automatic fallback on exchange API failures
- **Redis Caching**: Fast order state retrieval
- **NATS Pub/Sub**: Real-time fill event distribution
- **Maker-Fee Optimization**: POST_ONLY order support for maker rebates

### Architecture Highlights
- Multi-tier rate limiting (per-second, per-minute)
- Circuit breaker with adaptive thresholds
- Order state caching with TTL
- Real-time order update streaming
- Comprehensive Prometheus metrics
- Health check endpoints (liveness, readiness)

## Quick Start

### Prerequisites
- Go 1.21+
- Redis (for caching)
- NATS (for pub/sub)
- Binance Futures API credentials (testnet or production)

### Installation

```bash
# Clone and navigate to service directory
cd services/order-execution

# Install dependencies
go mod download

# Copy configuration
cp config.example.yaml config.yaml

# Edit config.yaml with your credentials
nano config.yaml
```

### Running Locally

```bash
# Build
go build -o bin/order-execution ./cmd/server

# Run
./bin/order-execution

# Or run directly
go run ./cmd/server
```

### Running with Docker

```bash
# Build image
docker build -t order-execution:latest .

# Run container
docker run -d \
  --name order-execution \
  -p 50051:50051 \
  -p 9091:9091 \
  -e BINANCE_API_KEY=your_key \
  -e BINANCE_SECRET_KEY=your_secret \
  -e REDIS_ADDRESS=redis:6379 \
  -e NATS_ADDRESS=nats://nats:4222 \
  order-execution:latest
```

## Configuration

Configuration can be provided via:
1. `config.yaml` file
2. Environment variables (override file config)

### Key Configuration Options

```yaml
server:
  grpc_port: 50051        # gRPC server port
  http_port: 9091         # HTTP metrics/health port

exchange:
  api_key: "${BINANCE_API_KEY}"
  secret_key: "${BINANCE_SECRET_KEY}"
  testnet: true           # Use testnet for development

rate_limit:
  requests_per_second: 10 # Conservative limit
  burst: 20               # Burst allowance
```

See `config.example.yaml` for full configuration options.

## API

### gRPC Endpoints

#### CreateOrder
Create a new order and submit to exchange.

```protobuf
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse)
```

**Example Request**:
```json
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "LIMIT",
  "quantity": 0.001,
  "price": 45000.0,
  "time_in_force": "GTC",
  "post_only": true
}
```

#### CancelOrder
Cancel an existing order.

```protobuf
rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse)
```

#### GetOrder
Retrieve order details.

```protobuf
rpc GetOrder(GetOrderRequest) returns (GetOrderResponse)
```

#### StreamOrderUpdates
Stream real-time order updates.

```protobuf
rpc StreamOrderUpdates(StreamOrderUpdatesRequest) returns (stream OrderUpdate)
```

### HTTP Endpoints

- `GET /health` - Comprehensive health check (JSON)
- `GET /health/live` - Liveness probe (K8s)
- `GET /health/ready` - Readiness probe (K8s)
- `GET /metrics` - Prometheus metrics

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/validator
```

### Integration Tests

```bash
# Run with integration tag
go test -tags=integration ./...
```

### Load Testing

```bash
# Install ghz (gRPC load testing tool)
go install github.com/bojand/ghz/cmd/ghz@latest

# Run load test
ghz --insecure \
  --proto proto/order.proto \
  --call order.OrderService/CreateOrder \
  -d '{"symbol":"BTCUSDT","side":"BUY","type":"LIMIT","quantity":0.001,"price":45000}' \
  -c 10 -n 1000 \
  localhost:50051
```

## Metrics

### Order Metrics
- `order_execution_orders_created_total` - Total orders created
- `order_execution_orders_filled_total` - Total orders filled
- `order_execution_orders_canceled_total` - Total orders canceled
- `order_execution_orders_rejected_total` - Total orders rejected
- `order_execution_order_latency_seconds` - Order creation latency

### Exchange Metrics
- `order_execution_exchange_requests_total` - Exchange API requests
- `order_execution_exchange_errors_total` - Exchange API errors
- `order_execution_exchange_latency_seconds` - Exchange API latency

### System Metrics
- `order_execution_rate_limit_hits_total` - Rate limit hits
- `order_execution_circuit_breaker_state` - Circuit breaker state
- `order_execution_cache_hits_total` - Cache hits
- `order_execution_cache_misses_total` - Cache misses

## Development

### Project Structure

```
order-execution/
├── cmd/
│   └── server/          # Main entry point
├── internal/
│   ├── executor/        # Order execution logic
│   ├── validator/       # Order validation
│   ├── exchange/        # Binance API client
│   ├── ratelimit/       # Rate limiting
│   ├── circuitbreaker/  # Circuit breaker
│   ├── metrics/         # Prometheus metrics
│   ├── health/          # Health checks
│   └── models/          # Data models
├── proto/               # gRPC protobuf definitions
├── config/              # Configuration files
├── Dockerfile           # Container image
└── go.mod              # Go dependencies
```

### Generating Protobuf Code

```bash
# Install protoc compiler
# Install Go protobuf plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/order.proto
```

## Production Deployment

### Environment Variables

```bash
# Required
BINANCE_API_KEY=your_api_key
BINANCE_SECRET_KEY=your_secret_key

# Optional (with defaults)
REDIS_ADDRESS=localhost:6379
NATS_ADDRESS=nats://localhost:4222
LOG_LEVEL=info
LOG_FORMAT=json
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-execution
spec:
  replicas: 3
  selector:
    matchLabels:
      app: order-execution
  template:
    metadata:
      labels:
        app: order-execution
    spec:
      containers:
      - name: order-execution
        image: order-execution:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 9091
          name: http
        env:
        - name: BINANCE_API_KEY
          valueFrom:
            secretKeyRef:
              name: binance-credentials
              key: api-key
        - name: BINANCE_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: binance-credentials
              key: secret-key
        livenessProbe:
          httpGet:
            path: /health/live
            port: 9091
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 9091
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Performance Tuning

#### Rate Limiting
Binance Futures has the following limits:
- **2400 requests/minute** (40 req/s)
- **1200 requests/minute per IP** for certain endpoints

Configure conservative limits:
```yaml
rate_limit:
  requests_per_second: 10
  burst: 20
```

#### Circuit Breaker
```yaml
circuit_breaker:
  failure_threshold: 5      # Open after 5 failures
  timeout: 30               # Stay open for 30s
  max_requests: 3           # Allow 3 requests in half-open
```

#### Redis Caching
```yaml
performance:
  order_cache_ttl: 86400    # 24 hours
```

## Troubleshooting

### Common Issues

#### 1. Rate Limit Errors
```
Error: rate limit exceeded
```

**Solution**: Reduce `requests_per_second` in config or increase `burst`.

#### 2. Circuit Breaker Open
```
Error: circuit breaker binance_create_order is open
```

**Solution**: Check exchange API status. Circuit breaker will auto-recover after timeout.

#### 3. Order Validation Failures
```
Error: validation failed: quantity precision exceeds maximum
```

**Solution**: Check symbol trading rules. Service auto-loads from exchange on startup.

### Logs

```bash
# View structured JSON logs
tail -f logs/order-execution.log | jq

# Filter by level
tail -f logs/order-execution.log | jq 'select(.level == "error")'

# Filter by order ID
tail -f logs/order-execution.log | jq 'select(.order_id == "abc-123")'
```

## Security

### API Credentials
- **Never commit** API keys to version control
- Use environment variables or secret management
- Rotate keys regularly
- Use testnet for development

### Network Security
- Use TLS for gRPC in production
- Implement API key authentication
- Rate limit client requests
- Monitor for suspicious activity

## Monitoring

### Grafana Dashboards

Import the included Grafana dashboard for comprehensive monitoring:
- Order throughput and latency
- Exchange API performance
- Circuit breaker state
- Cache hit rates
- Error rates and types

### Alerts

Recommended Prometheus alerts:
```yaml
# High error rate
- alert: HighOrderRejectionRate
  expr: rate(order_execution_orders_rejected_total[5m]) > 0.1

# Circuit breaker open
- alert: CircuitBreakerOpen
  expr: order_execution_circuit_breaker_state == 2

# High latency
- alert: HighOrderLatency
  expr: histogram_quantile(0.95, order_execution_order_latency_seconds) > 1
```

## Contributing

See `../../CONTRIBUTING.md` for development guidelines.

## License

See `../../LICENSE` for license information.

## Support

- **Documentation**: `../../docs/service-plans/02-order-execution-service.md`
- **Issues**: GitHub Issues
- **Slack**: #order-execution channel
