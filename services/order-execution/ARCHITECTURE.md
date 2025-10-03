# Order Execution Service - Architecture

## Overview

The Order Execution Service is a high-performance, production-ready microservice designed for low-latency order management in HFT trading systems. It handles the complete order lifecycle from validation through exchange submission to fill notification.

## Design Principles

1. **Low Latency**: Target <10ms order submission
2. **Reliability**: Circuit breakers, retries, and graceful degradation
3. **Observability**: Comprehensive metrics and structured logging
4. **Safety**: Multi-layer validation and risk limits
5. **Scalability**: Stateless design with distributed caching

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Order Execution Service                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐      ┌───────────────┐      ┌──────────────┐ │
│  │  gRPC Server │─────▶│   Executor    │─────▶│   Exchange   │ │
│  │  (Port 50051)│      │               │      │  API Client  │ │
│  └──────────────┘      └───────────────┘      └──────────────┘ │
│         │                      │                       │         │
│         │                      │                       │         │
│         ▼                      ▼                       ▼         │
│  ┌──────────────┐      ┌───────────────┐      ┌──────────────┐ │
│  │  Validator   │      │  Rate Limiter │      │   Circuit    │ │
│  │              │      │               │      │   Breaker    │ │
│  └──────────────┘      └───────────────┘      └──────────────┘ │
│         │                      │                       │         │
│         ▼                      ▼                       ▼         │
│  ┌──────────────┐      ┌───────────────┐      ┌──────────────┐ │
│  │    Redis     │      │     NATS      │      │  Prometheus  │ │
│  │  (Cache)     │      │  (Events)     │      │  (Metrics)   │ │
│  └──────────────┘      └───────────────┘      └──────────────┘ │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              HTTP Server (Port 9091)                       │  │
│  │  /health  /health/ready  /health/live  /metrics           │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. gRPC Server (`internal/executor/grpc_server.go`)

**Purpose**: External API interface for order management

**Endpoints**:
- `CreateOrder`: Submit new orders
- `CancelOrder`: Cancel existing orders
- `GetOrder`: Query order status
- `GetOrders`: Bulk order queries
- `StreamOrderUpdates`: Real-time order updates stream

**Features**:
- Protocol Buffers for efficient serialization
- Bidirectional streaming support
- Context-aware request handling
- Error translation to gRPC status codes

### 2. Order Executor (`internal/executor/executor.go`)

**Purpose**: Core order lifecycle management

**Responsibilities**:
- Order validation orchestration
- Exchange API communication
- State machine enforcement
- Cache management
- Event publishing

**State Machine**:
```
NEW → SUBMITTED → PARTIALLY_FILLED → FILLED
  ↓       ↓              ↓
REJECTED  CANCELED    CANCELED
```

**Key Methods**:
- `CreateOrder()`: Validate → Submit → Update cache → Publish event
- `CancelOrder()`: Verify state → Submit cancel → Update cache
- `GetOrder()`: Check memory → Check Redis → Return

### 3. Validator (`internal/validator/validator.go`)

**Purpose**: Multi-layer order validation

**Validation Layers**:

1. **Symbol Validation**
   - Symbol is registered
   - Symbol is allowed for trading

2. **Quantity Validation**
   - Within min/max bounds
   - Matches step size precision
   - Respects quantity precision limits

3. **Price Validation**
   - Matches tick size
   - Respects price precision limits
   - Within reasonable bounds

4. **Notional Validation**
   - Order value exceeds minimum notional
   - Order value within limits

5. **Risk Validation**
   - Order value within risk limits
   - Position size within limits
   - Order count within daily limits

6. **Time-in-Force Validation**
   - Compatible with order type
   - POST_ONLY uses GTX

### 4. Exchange Client (`internal/exchange/binance.go`)

**Purpose**: Binance Futures API communication

**Features**:
- **HMAC SHA256 Signing**: Secure request authentication
- **Request Building**: Proper query string formatting
- **Response Parsing**: JSON to internal models
- **Error Handling**: Exchange error translation
- **Testnet Support**: Development environment

**API Methods**:
- `CreateOrder()`: POST /fapi/v1/order
- `CancelOrder()`: DELETE /fapi/v1/order
- `GetOrder()`: GET /fapi/v1/order
- `GetExchangeInfo()`: GET /fapi/v1/exchangeInfo
- `GetAccountInfo()`: GET /fapi/v2/account

**Security**:
- API keys never logged
- Timestamps prevent replay attacks
- Query string sorted for consistent signing

### 5. Rate Limiter (`internal/ratelimit/ratelimit.go`)

**Purpose**: Protect against exchange rate limits

**Implementations**:

1. **Simple Rate Limiter**
   - Token bucket algorithm
   - Per-key rate limiting
   - Configurable RPS and burst

2. **Multi-Tier Limiter**
   - Multiple time windows
   - Cascade limiting
   - Per-second, per-minute, per-hour

3. **Weighted Limiter**
   - Different weights per operation
   - Expensive operations consume more tokens

**Configuration**:
```yaml
rate_limit:
  requests_per_second: 10
  burst: 20
```

### 6. Circuit Breaker (`internal/circuitbreaker/circuitbreaker.go`)

**Purpose**: Prevent cascade failures

**States**:
- **Closed**: Normal operation, requests pass through
- **Open**: Threshold exceeded, requests fail fast
- **Half-Open**: Testing recovery, limited requests

**Configuration**:
```yaml
circuit_breaker:
  failure_threshold: 5    # Open after 5 failures
  timeout: 30            # Stay open for 30s
  max_requests: 3        # Allow 3 requests in half-open
```

**Features**:
- Per-endpoint circuit breakers
- Automatic recovery testing
- State change callbacks
- Metrics integration

### 7. Metrics (`internal/metrics/metrics.go`)

**Purpose**: Observability and monitoring

**Metric Types**:

1. **Counters**:
   - `orders_created_total`
   - `orders_filled_total`
   - `orders_canceled_total`
   - `orders_rejected_total`
   - `exchange_errors_total`

2. **Histograms**:
   - `order_latency_seconds`
   - `cancel_latency_seconds`
   - `exchange_latency_seconds`

3. **Gauges**:
   - `order_state{state="SUBMITTED"}`
   - `circuit_breaker_state{name="binance"}`

**Usage**:
```go
metrics.OrdersCreated.Inc()
metrics.OrderLatency.Observe(duration.Seconds())
metrics.RecordOrderState("SUBMITTED")
```

### 8. Health Checks (`internal/health/health.go`)

**Purpose**: Service health monitoring

**Checks**:
1. **Redis**: Connection and ping
2. **NATS**: Connection status
3. **System**: Basic health

**Endpoints**:
- `/health`: Full health check (JSON)
- `/health/ready`: Readiness probe (K8s)
- `/health/live`: Liveness probe (K8s)

**States**:
- `healthy`: All checks pass
- `degraded`: Some non-critical failures
- `unhealthy`: Critical failures

## Data Flow

### Order Creation Flow

```
1. Client Request
   └─▶ gRPC CreateOrder()

2. Validation
   ├─▶ Symbol registered?
   ├─▶ Quantity valid?
   ├─▶ Price valid?
   ├─▶ Notional valid?
   └─▶ Risk limits OK?

3. Rate Limiting
   └─▶ Token available?

4. Circuit Breaker
   └─▶ Circuit closed?

5. Exchange Submission
   ├─▶ Build request
   ├─▶ Sign request (HMAC SHA256)
   ├─▶ POST to exchange
   └─▶ Parse response

6. State Update
   ├─▶ Update order state
   ├─▶ Store in Redis
   └─▶ Store in memory

7. Event Publishing
   └─▶ Publish to NATS

8. Response
   └─▶ Return to client
```

### Caching Strategy

```
┌─────────────────────────────────────────┐
│         Order Retrieval Strategy         │
├─────────────────────────────────────────┤
│                                          │
│  1. Check Memory Cache                   │
│     └─▶ Hit? Return immediately          │
│                                          │
│  2. Check Redis Cache                    │
│     └─▶ Hit? Cache in memory, return    │
│                                          │
│  3. Query Database (if implemented)      │
│     └─▶ Cache in Redis + memory, return │
│                                          │
│  4. Not Found                            │
│     └─▶ Return error                     │
│                                          │
└─────────────────────────────────────────┘
```

**Cache Invalidation**:
- TTL-based: 24 hours
- Event-based: On order updates
- Manual: Admin endpoints

## Error Handling

### Error Types

1. **Validation Errors**
   - Status: `INVALID_ARGUMENT`
   - Recovery: Client fix
   - Example: Invalid quantity

2. **Rate Limit Errors**
   - Status: `RESOURCE_EXHAUSTED`
   - Recovery: Retry with backoff
   - Example: Too many requests

3. **Circuit Breaker Errors**
   - Status: `UNAVAILABLE`
   - Recovery: Wait for auto-recovery
   - Example: Exchange down

4. **Exchange Errors**
   - Status: Varies by error
   - Recovery: Varies
   - Example: Insufficient balance

### Retry Strategy

```yaml
Validation Errors: No retry (client error)
Rate Limits: Exponential backoff
Circuit Breaker: Wait for recovery
Transient Errors: 3 retries with jitter
```

## Performance Optimizations

### 1. Order Submission Path

```
Target: <10ms total latency

Breakdown:
- gRPC receive:        <1ms
- Validation:          <1ms
- Rate limit check:    <1ms
- Exchange API call:   5-8ms
- State update:        <1ms
- Event publish:       <1ms
Total:                 8-12ms
```

### 2. Caching

- **Memory Cache**: O(1) order lookup
- **Redis Cache**: <1ms network latency
- **Cache Warm-up**: On service start

### 3. Connection Pooling

- **HTTP Client**: Reuse connections
- **Redis**: Connection pool (10 conns)
- **NATS**: Single persistent connection

### 4. Async Operations

- Event publishing: Fire and forget
- Metrics recording: Non-blocking
- Log writing: Buffered

## Security Considerations

### 1. API Key Management

```go
// Never log API keys
logger.Info("creating order",
    zap.String("symbol", symbol))  // ✓
    // NEVER: zap.String("api_key", apiKey)  // ✗

// Environment variables
apiKey := os.Getenv("BINANCE_API_KEY")

// Secrets manager (production)
apiKey := secretsManager.GetSecret("binance-api-key")
```

### 2. Request Signing

```go
// HMAC SHA256 signature
signature := sign(queryString, secretKey)

// Timestamp prevents replay
timestamp := time.Now().UnixMilli()
```

### 3. Input Validation

All inputs validated before processing:
- Symbol whitelist
- Quantity bounds
- Price bounds
- Order value limits

### 4. Rate Limiting

Multiple layers:
- Application-level
- Per-user limits
- Global limits

## Scaling Considerations

### Horizontal Scaling

**Stateless Design**:
- No local state (beyond caches)
- Can run multiple instances
- Load balancer distributes requests

**Shared State**:
- Redis for distributed cache
- NATS for event distribution
- PostgreSQL for persistent storage (optional)

### Vertical Scaling

**Resource Requirements**:
- CPU: 2 cores minimum
- Memory: 1GB minimum
- Network: Low latency critical

**Bottlenecks**:
- Exchange API latency (5-10ms)
- Network I/O
- Rate limits

## Monitoring & Alerting

### Key Metrics to Monitor

1. **Latency**
   - P50, P95, P99 order latency
   - Alert if P99 > 50ms

2. **Error Rate**
   - Order rejection rate
   - Exchange error rate
   - Alert if > 1% errors

3. **Throughput**
   - Orders per second
   - Fills per second

4. **Circuit Breaker**
   - State changes
   - Alert on OPEN state

5. **Cache Performance**
   - Hit rate (should be >80%)
   - Miss rate

### Grafana Dashboard

Panels:
- Order throughput (line chart)
- Latency percentiles (heatmap)
- Error rates (counter)
- Circuit breaker state (gauge)
- Cache hit ratio (gauge)

## Testing Strategy

### Unit Tests

```bash
go test ./internal/validator -v
go test ./internal/ratelimit -v
go test ./internal/circuitbreaker -v
```

### Integration Tests

```bash
go test -tags=integration ./...
```

Test scenarios:
- Full order lifecycle
- Error handling
- Rate limiting
- Circuit breaker behavior

### Load Tests

```bash
ghz --insecure \
    --proto proto/order.proto \
    --call order.OrderService/CreateOrder \
    -c 100 -n 10000 \
    localhost:50051
```

Target: 1000 orders/second sustained

## Deployment

### Docker

```bash
docker build -t order-execution .
docker run -p 50051:50051 -p 9091:9091 order-execution
```

### Kubernetes

```yaml
resources:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "2000m"
```

### Environment Variables

Required:
- `BINANCE_API_KEY`
- `BINANCE_SECRET_KEY`

Optional:
- `REDIS_ADDRESS`
- `NATS_ADDRESS`
- `LOG_LEVEL`

## Future Enhancements

1. **Multi-Exchange Support**
   - Abstract exchange interface
   - FTX, Bybit, OKX clients

2. **Advanced Order Types**
   - Iceberg orders
   - TWAP orders
   - Smart order routing

3. **Position Management**
   - Position tracking
   - P&L calculation
   - Risk scoring

4. **Machine Learning**
   - Optimal execution timing
   - Slippage prediction
   - Fill rate optimization

5. **Database Persistence**
   - PostgreSQL/TimescaleDB
   - Historical order data
   - Analytics queries

## References

- [Binance Futures API](https://binance-docs.github.io/apidocs/futures/en/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Rate Limiting Strategies](https://en.wikipedia.org/wiki/Token_bucket)
