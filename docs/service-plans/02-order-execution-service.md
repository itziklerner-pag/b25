# Order Execution Service - Development Plan

**Service:** Order Execution Service
**Purpose:** Order lifecycle management and exchange communication
**Version:** 1.0
**Last Updated:** 2025-10-02

---

## Table of Contents

1. [Technology Stack Recommendation](#1-technology-stack-recommendation)
2. [Architecture Design](#2-architecture-design)
3. [Development Phases](#3-development-phases)
4. [Implementation Details](#4-implementation-details)
5. [Testing Strategy](#5-testing-strategy)
6. [Deployment](#6-deployment)
7. [Observability](#7-observability)

---

## 1. Technology Stack Recommendation

### Primary Language: **Go**

**Rationale:**
- Excellent async HTTP support with native goroutines
- Strong concurrency primitives for rate limiting and circuit breakers
- Fast compilation and deployment
- Rich ecosystem for HTTP clients and gRPC
- Built-in profiling and performance monitoring
- Type safety for financial operations
- Low memory footprint

**Alternative Considerations:**
- **Rust:** Best performance, but longer development time
- **Node.js:** Fast development, but less optimal for CPU-bound operations
- **Python:** Excellent for prototyping, but GIL limitations for high concurrency

### Core Dependencies

```go
// HTTP Client
"net/http"                              // Standard library with connection pooling
"golang.org/x/time/rate"                // Token bucket rate limiter

// RPC Framework
"google.golang.org/grpc"                // gRPC for RPC endpoint
"google.golang.org/protobuf"            // Protocol buffers

// Alternative: ZeroMQ
"github.com/pebbe/zmq4"                 // ZeroMQ REP socket

// State Storage
"github.com/redis/go-redis/v9"          // Redis client

// Pub/Sub
"github.com/nats-io/nats.go"            // NATS for event publishing
// Alternative: Redis Pub/Sub

// Configuration
"github.com/spf13/viper"                // Configuration management
"github.com/joho/godotenv"              // Environment variables

// Cryptography
"crypto/hmac"                           // HMAC for exchange signatures
"crypto/sha256"                         // SHA256 hashing

// Utilities
"github.com/google/uuid"                // UUID generation for request IDs
"github.com/sony/gobreaker"             // Circuit breaker implementation
"go.uber.org/zap"                       // Structured logging
"github.com/prometheus/client_golang"   // Metrics

// Testing
"github.com/stretchr/testify"           // Test assertions
"github.com/jarcoal/httpmock"           // HTTP mocking
"github.com/testcontainers/testcontainers-go" // Integration tests
```

### Recommended Stack Summary

| Component | Technology | Why |
|-----------|-----------|-----|
| Language | Go 1.22+ | Performance, concurrency, type safety |
| RPC Framework | gRPC | Typed interfaces, streaming support, industry standard |
| HTTP Client | net/http | Built-in, excellent connection pooling |
| State Store | Redis 7+ | Fast, supports TTL, pub/sub, atomic operations |
| Pub/Sub | NATS | Lightweight, cloud-native, high performance |
| Circuit Breaker | sony/gobreaker | Battle-tested, configurable |
| Rate Limiter | golang.org/x/time/rate | Token bucket, built-in |
| Testing | testify + httpmock | Comprehensive, easy mocking |
| Metrics | Prometheus | Industry standard |
| Logging | zap | Fast structured logging |

---

## 2. Architecture Design

### 2.1 Order Lifecycle State Machine

```
┌─────────────────────────────────────────────────────────┐
│                    Order State Machine                   │
└─────────────────────────────────────────────────────────┘

    [Client Request]
         |
         v
    ┌─────────┐
    │  INIT   │ Initial state
    └────┬────┘
         |
         v
    ┌─────────┐
    │VALIDATING│ Pre-flight checks
    └────┬────┘
         |
         ├─── REJECTED (validation failed)
         |
         v
    ┌─────────┐
    │ PENDING │ Queued for submission
    └────┬────┘
         |
         v
    ┌──────────┐
    │SUBMITTING│ Sending to exchange
    └────┬─────┘
         |
         ├─── REJECTED (exchange rejection)
         |
         v
    ┌─────────┐
    │   NEW   │ Accepted by exchange
    └────┬────┘
         |
         ├────────────────┬────────────────┐
         v                v                v
    ┌─────────┐    ┌──────────┐    ┌─────────┐
    │ FILLED  │    │PARTIALLY │    │CANCELED │
    │         │    │  FILLED  │    │         │
    └─────────┘    └────┬─────┘    └─────────┘
                        |
                        v
                   ┌─────────┐
                   │ FILLED  │
                   └─────────┘

Terminal States: FILLED, CANCELED, REJECTED
```

### 2.2 System Architecture

```
┌────────────────────────────────────────────────────────┐
│         Order Execution Service Architecture           │
└────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                      RPC Server                         │
│                  (gRPC / ZeroMQ REP)                    │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                  Request Handler                        │
│  - Idempotency Check (request_id)                      │
│  - Request Deserialization                              │
│  - Correlation ID Assignment                            │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                 Validation Pipeline                     │
│  1. Symbol validation                                   │
│  2. Quantity/price checks                               │
│  3. Order type validation                               │
│  4. Risk checks (via Risk Manager RPC - optional)       │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                  Rate Limiter                           │
│  - Token bucket per endpoint type                       │
│  - Configurable limits per exchange                     │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                  Circuit Breaker                        │
│  - State: CLOSED / OPEN / HALF_OPEN                     │
│  - Failure threshold tracking                           │
│  - Automatic recovery testing                           │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│              Exchange Adapter Layer                     │
│  - Binance Futures Adapter                              │
│  - OKX Adapter (future)                                 │
│  - Bybit Adapter (future)                               │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                 HTTP Client Pool                        │
│  - Connection pooling (max 100 conns)                   │
│  - Keep-alive                                           │
│  - Timeout configuration                                │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
               [Exchange REST API]
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│              Response Handler                           │
│  - Parse exchange response                              │
│  - Map to internal order state                          │
│  - Error classification                                 │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                State Manager                            │
│  - Redis state persistence                              │
│  - Order history tracking                               │
│  - Fill event aggregation                               │
└──────────────────────┬──────────────────────────────────┘
                       |
                       v
┌─────────────────────────────────────────────────────────┐
│                Event Publisher                          │
│  - Pub/sub for order updates                            │
│  - Pub/sub for fill events                              │
│  - Time-series DB writes                                │
└─────────────────────────────────────────────────────────┘
```

### 2.3 Validation Pipeline

```go
type ValidationRule interface {
    Validate(order *Order) error
}

type ValidationPipeline struct {
    rules []ValidationRule
}

// Validation Rules (in order)
1. SymbolValidator      - Symbol exists and is tradeable
2. QuantityValidator    - Min/max quantity, step size
3. PriceValidator       - Price precision, tick size, min notional
4. OrderTypeValidator   - Valid order type for symbol
5. TimeInForceValidator - Valid TIF for order type
6. BalanceValidator     - Sufficient balance (optional, can be done by exchange)
7. RiskValidator        - Risk manager pre-trade check (RPC call)
```

### 2.4 Rate Limiter Design

```go
type RateLimiter struct {
    // Binance Futures has multiple rate limits:
    // - Order limit: 300 orders/10s per account
    // - Request weight: 2400/min per IP
    orderLimiter   *rate.Limiter  // 30 orders/sec (safety buffer)
    weightLimiter  *rate.Limiter  // 40 weight/sec (2400/min)

    // Per-symbol limits (optional)
    symbolLimiters map[string]*rate.Limiter
}

// Token bucket algorithm:
// - orderLimiter: burst=50, rate=30/sec
// - weightLimiter: burst=100, rate=40/sec
```

### 2.5 Circuit Breaker Pattern

```go
type CircuitBreakerConfig struct {
    MaxRequests     uint32  // Max requests in half-open state
    Interval        time.Duration // Reset interval
    Timeout         time.Duration // Time to wait in open state
    ReadyToTrip     func(counts gobreaker.Counts) bool
}

// Default Configuration:
// - Open circuit after 5 consecutive failures
// - Half-open after 30 seconds
// - Allow 3 requests in half-open state
// - Close after 2 consecutive successes
```

### 2.6 Idempotency Mechanism

```go
// Idempotency Key: client-provided request_id
// Storage: Redis with 24-hour TTL
// Key format: "idempotency:{request_id}"
// Value: JSON-serialized response

type IdempotencyCache struct {
    redis *redis.Client
    ttl   time.Duration // 24 hours
}

// Workflow:
// 1. Check if request_id exists in Redis
// 2. If exists, return cached response immediately
// 3. If not, process request
// 4. Cache response with request_id as key
// 5. Set TTL to 24 hours
```

### 2.7 Component Interaction Diagram

```
Strategy Engine                     Risk Manager
      |                                  |
      | [RPC: PlaceOrder]                |
      v                                  |
┌──────────────────────────────────┐    |
│   Order Execution Service        │    |
│                                  │    |
│  ┌────────────────────────┐     │    |
│  │ Idempotency Check      │     │    |
│  └───────────┬────────────┘     │    |
│              v                   │    |
│  ┌────────────────────────┐     │    |
│  │ Validation Pipeline    │◄────┼────┘ [RPC: PreTradeCheck]
│  └───────────┬────────────┘     │
│              v                   │
│  ┌────────────────────────┐     │
│  │ Rate Limiter           │     │
│  └───────────┬────────────┘     │
│              v                   │
│  ┌────────────────────────┐     │
│  │ Circuit Breaker        │     │
│  └───────────┬────────────┘     │
│              v                   │
│  ┌────────────────────────┐     │
│  │ Exchange Adapter       │     │
│  └───────────┬────────────┘     │
│              v                   │
└──────────────┼───────────────────┘
               |
               v
        Exchange REST API
               |
               v
        [Response Handler]
               |
               v
        ┌──────────────┐
        │ State Manager│────► Redis
        └──────┬───────┘
               |
               v
        ┌──────────────┐
        │Event Publisher│───► NATS (order_updates, fill_events)
        └──────────────┘
```

---

## 3. Development Phases

### Phase 1: Foundation & RPC Server (Days 1-2)

**Goal:** Basic service skeleton with RPC endpoint

**Tasks:**
- [ ] Project structure setup
- [ ] gRPC service definition (protobuf)
- [ ] RPC server implementation
- [ ] Health check endpoint
- [ ] Configuration management (Viper)
- [ ] Structured logging (Zap)
- [ ] Basic metrics (Prometheus)
- [ ] Docker development environment

**Deliverables:**
- Working gRPC server accepting order requests
- Health check endpoint at `/health`
- Metrics endpoint at `/metrics`
- Dockerfile and docker-compose.yml

**Time Estimate:** 2 days

---

### Phase 2: Order Validation Engine (Days 3-4)

**Goal:** Comprehensive pre-flight validation

**Tasks:**
- [ ] Define Order and OrderRequest structs
- [ ] Implement validation pipeline interface
- [ ] Symbol validator (with configurable symbol list)
- [ ] Quantity validator (min/max, step size)
- [ ] Price validator (precision, tick size, min notional)
- [ ] Order type validator
- [ ] Time-in-force validator
- [ ] Unit tests for all validators
- [ ] Error response standardization

**Deliverables:**
- Complete validation pipeline
- 90%+ test coverage for validators
- Validation error codes and messages

**Time Estimate:** 2 days

---

### Phase 3: Exchange API Integration (Days 5-7)

**Goal:** Binance Futures REST API integration

**Tasks:**
- [ ] HTTP client with connection pooling
- [ ] HMAC signature generation
- [ ] Timestamp synchronization
- [ ] Binance Futures adapter implementation
  - [ ] POST /fapi/v1/order (New Order)
  - [ ] GET /fapi/v1/order (Query Order)
  - [ ] DELETE /fapi/v1/order (Cancel Order)
- [ ] Response parsing and error mapping
- [ ] Exchange error code handling
- [ ] Request/response logging
- [ ] Integration tests with mock server

**Deliverables:**
- Working Binance Futures adapter
- HMAC signature implementation
- Exchange error mapping
- Integration test suite

**Time Estimate:** 3 days

---

### Phase 4: State Management & Pub/Sub (Days 8-9)

**Goal:** Order state persistence and event publishing

**Tasks:**
- [ ] Redis client setup
- [ ] Order state persistence
  - [ ] Key schema: `order:{order_id}`
  - [ ] TTL management (30 days)
- [ ] Idempotency cache
  - [ ] Key schema: `idempotency:{request_id}`
  - [ ] 24-hour TTL
- [ ] Order history tracking
- [ ] NATS client setup
- [ ] Event publisher implementation
  - [ ] order.created events
  - [ ] order.filled events
  - [ ] order.canceled events
  - [ ] order.rejected events
- [ ] Fill event aggregation
- [ ] Integration tests with Redis and NATS

**Deliverables:**
- Redis state management
- NATS event publishing
- Idempotency implementation
- Order history API

**Time Estimate:** 2 days

---

### Phase 5: Circuit Breaker & Rate Limiting (Days 10-11)

**Goal:** Resilience and rate limiting

**Tasks:**
- [ ] Rate limiter implementation
  - [ ] Order rate limiter (30/sec)
  - [ ] Weight rate limiter (40/sec)
  - [ ] Per-symbol limiters (optional)
- [ ] Circuit breaker integration
  - [ ] Configuration (thresholds, timeouts)
  - [ ] State management (closed/open/half-open)
  - [ ] Metrics for circuit state
- [ ] Retry logic with exponential backoff
- [ ] Rate limit response handling (429)
- [ ] Circuit breaker tests
- [ ] Load testing

**Deliverables:**
- Production-ready rate limiting
- Circuit breaker implementation
- Resilience test suite
- Load test results

**Time Estimate:** 2 days

---

### Phase 6: Testing & Observability (Days 12-14)

**Goal:** Comprehensive testing and monitoring

**Tasks:**
- [ ] Mock exchange server implementation
- [ ] End-to-end test suite
  - [ ] Happy path: order placement
  - [ ] Validation errors
  - [ ] Exchange errors
  - [ ] Rate limiting
  - [ ] Circuit breaker scenarios
  - [ ] Idempotency
- [ ] Integration tests with testcontainers (Redis)
- [ ] Performance benchmarks
- [ ] Observability dashboard
  - [ ] Grafana dashboard JSON
  - [ ] Key metrics visualization
- [ ] Documentation
  - [ ] API documentation
  - [ ] Deployment guide
  - [ ] Runbook

**Deliverables:**
- 85%+ test coverage
- Mock exchange server
- Grafana dashboard
- Complete documentation

**Time Estimate:** 3 days

---

### Development Phase Summary

| Phase | Days | Focus | Key Deliverable |
|-------|------|-------|----------------|
| 1 | 1-2 | Foundation | gRPC server |
| 2 | 3-4 | Validation | Validation pipeline |
| 3 | 5-7 | Exchange API | Binance adapter |
| 4 | 8-9 | State & Events | Redis + NATS |
| 5 | 10-11 | Resilience | Rate limiter + Circuit breaker |
| 6 | 12-14 | Testing | Test suite + Dashboard |

**Total:** 14 days (~2 weeks with buffer)

---

## 4. Implementation Details

### 4.1 Project Structure

```
order-execution-service/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point
├── internal/
│   ├── config/
│   │   └── config.go                  # Configuration management
│   ├── handler/
│   │   ├── grpc_handler.go            # gRPC request handlers
│   │   └── health_handler.go          # Health check handler
│   ├── validator/
│   │   ├── validator.go               # Validation interface
│   │   ├── symbol_validator.go
│   │   ├── quantity_validator.go
│   │   ├── price_validator.go
│   │   └── pipeline.go                # Validation pipeline
│   ├── exchange/
│   │   ├── adapter.go                 # Exchange adapter interface
│   │   ├── binance/
│   │   │   ├── adapter.go             # Binance implementation
│   │   │   ├── client.go              # HTTP client
│   │   │   ├── signer.go              # HMAC signature
│   │   │   └── types.go               # Binance types
│   │   └── mock/
│   │       └── adapter.go             # Mock for testing
│   ├── ratelimit/
│   │   └── limiter.go                 # Rate limiter
│   ├── circuitbreaker/
│   │   └── breaker.go                 # Circuit breaker wrapper
│   ├── state/
│   │   ├── manager.go                 # State management
│   │   ├── redis_store.go             # Redis storage
│   │   └── idempotency.go             # Idempotency cache
│   ├── events/
│   │   ├── publisher.go               # Event publisher interface
│   │   └── nats_publisher.go          # NATS implementation
│   ├── models/
│   │   ├── order.go                   # Order model
│   │   ├── enums.go                   # Enums (state, side, type)
│   │   └── errors.go                  # Error types
│   └── metrics/
│       └── metrics.go                 # Prometheus metrics
├── proto/
│   └── order_service.proto            # gRPC service definition
├── pkg/
│   └── logger/
│       └── logger.go                  # Logger wrapper
├── test/
│   ├── integration/
│   │   └── order_test.go              # Integration tests
│   ├── mock/
│   │   └── exchange_server.go         # Mock exchange server
│   └── fixtures/
│       └── orders.json                # Test fixtures
├── configs/
│   ├── config.yaml                    # Default config
│   └── config.example.yaml            # Example config
├── deployments/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── k8s/
│       ├── deployment.yaml
│       └── service.yaml
├── scripts/
│   ├── generate-proto.sh              # Protobuf generation
│   └── test.sh                        # Test runner
├── docs/
│   ├── API.md                         # API documentation
│   └── ARCHITECTURE.md                # Architecture doc
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 4.2 Core Data Models

```go
// internal/models/order.go

package models

import (
    "time"
)

// Order represents an order in the system
type Order struct {
    ID              string          `json:"id" redis:"id"`
    RequestID       string          `json:"request_id" redis:"request_id"`
    ExchangeOrderID string          `json:"exchange_order_id,omitempty" redis:"exchange_order_id"`
    Symbol          string          `json:"symbol" redis:"symbol"`
    Side            OrderSide       `json:"side" redis:"side"`
    Type            OrderType       `json:"type" redis:"type"`
    TimeInForce     TimeInForce     `json:"time_in_force" redis:"time_in_force"`
    Quantity        float64         `json:"quantity" redis:"quantity"`
    Price           float64         `json:"price,omitempty" redis:"price"`
    State           OrderState      `json:"state" redis:"state"`
    FilledQuantity  float64         `json:"filled_quantity" redis:"filled_quantity"`
    AveragePrice    float64         `json:"average_price,omitempty" redis:"average_price"`
    CreatedAt       time.Time       `json:"created_at" redis:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at" redis:"updated_at"`
    ErrorCode       string          `json:"error_code,omitempty" redis:"error_code"`
    ErrorMessage    string          `json:"error_message,omitempty" redis:"error_message"`
}

// OrderState represents the lifecycle state of an order
type OrderState string

const (
    OrderStateInit        OrderState = "INIT"
    OrderStateValidating  OrderState = "VALIDATING"
    OrderStatePending     OrderState = "PENDING"
    OrderStateSubmitting  OrderState = "SUBMITTING"
    OrderStateNew         OrderState = "NEW"
    OrderStatePartiallyFilled OrderState = "PARTIALLY_FILLED"
    OrderStateFilled      OrderState = "FILLED"
    OrderStateCanceled    OrderState = "CANCELED"
    OrderStateRejected    OrderState = "REJECTED"
)

// IsTerminal returns true if the order state is terminal
func (s OrderState) IsTerminal() bool {
    return s == OrderStateFilled || s == OrderStateCanceled || s == OrderStateRejected
}

// OrderSide represents the order side (buy/sell)
type OrderSide string

const (
    OrderSideBuy  OrderSide = "BUY"
    OrderSideSell OrderSide = "SELL"
)

// OrderType represents the order type
type OrderType string

const (
    OrderTypeLimit           OrderType = "LIMIT"
    OrderTypeMarket          OrderType = "MARKET"
    OrderTypeStopLoss        OrderType = "STOP_LOSS"
    OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
    OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
    OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"
)

// TimeInForce represents the time in force
type TimeInForce string

const (
    TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
    TimeInForceIOC TimeInForce = "IOC" // Immediate Or Cancel
    TimeInForceFOK TimeInForce = "FOK" // Fill Or Kill
)

// OrderRequest represents a request to place an order
type OrderRequest struct {
    RequestID   string      `json:"request_id" validate:"required"`
    Symbol      string      `json:"symbol" validate:"required"`
    Side        OrderSide   `json:"side" validate:"required,oneof=BUY SELL"`
    Type        OrderType   `json:"type" validate:"required"`
    TimeInForce TimeInForce `json:"time_in_force,omitempty"`
    Quantity    float64     `json:"quantity" validate:"required,gt=0"`
    Price       float64     `json:"price,omitempty"`
}

// OrderResponse represents the response after order submission
type OrderResponse struct {
    Order   *Order `json:"order"`
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
}

// FillEvent represents a fill event
type FillEvent struct {
    OrderID         string    `json:"order_id"`
    ExchangeOrderID string    `json:"exchange_order_id"`
    Symbol          string    `json:"symbol"`
    Side            OrderSide `json:"side"`
    Price           float64   `json:"price"`
    Quantity        float64   `json:"quantity"`
    Commission      float64   `json:"commission"`
    CommissionAsset string    `json:"commission_asset"`
    TradeID         string    `json:"trade_id"`
    Timestamp       time.Time `json:"timestamp"`
}
```

### 4.3 gRPC Service Definition

```protobuf
// proto/order_service.proto

syntax = "proto3";

package order;

option go_package = "github.com/yourusername/order-execution-service/proto";

import "google/protobuf/timestamp.proto";

service OrderService {
  // PlaceOrder submits a new order to the exchange
  rpc PlaceOrder(OrderRequest) returns (OrderResponse);

  // GetOrder retrieves order information
  rpc GetOrder(GetOrderRequest) returns (OrderResponse);

  // CancelOrder cancels an existing order
  rpc CancelOrder(CancelOrderRequest) returns (OrderResponse);

  // GetOrderHistory retrieves order history
  rpc GetOrderHistory(GetOrderHistoryRequest) returns (OrderHistoryResponse);
}

message OrderRequest {
  string request_id = 1;    // Client-provided idempotency key
  string symbol = 2;        // Trading symbol (e.g., "BTCUSDT")
  string side = 3;          // BUY or SELL
  string type = 4;          // LIMIT, MARKET, etc.
  string time_in_force = 5; // GTC, IOC, FOK
  double quantity = 6;      // Order quantity
  double price = 7;         // Order price (optional for MARKET)
}

message OrderResponse {
  Order order = 1;
  bool success = 2;
  string error = 3;
}

message Order {
  string id = 1;
  string request_id = 2;
  string exchange_order_id = 3;
  string symbol = 4;
  string side = 5;
  string type = 6;
  string time_in_force = 7;
  double quantity = 8;
  double price = 9;
  string state = 10;
  double filled_quantity = 11;
  double average_price = 12;
  google.protobuf.Timestamp created_at = 13;
  google.protobuf.Timestamp updated_at = 14;
  string error_code = 15;
  string error_message = 16;
}

message GetOrderRequest {
  string order_id = 1;
}

message CancelOrderRequest {
  string order_id = 1;
}

message GetOrderHistoryRequest {
  string symbol = 1;
  int32 limit = 2;
  int64 start_time = 3;
  int64 end_time = 4;
}

message OrderHistoryResponse {
  repeated Order orders = 1;
}
```

### 4.4 Order Validation Implementation

```go
// internal/validator/pipeline.go

package validator

import (
    "context"
    "fmt"

    "order-execution-service/internal/models"
)

// Validator interface for validation rules
type Validator interface {
    Validate(ctx context.Context, req *models.OrderRequest) error
}

// Pipeline executes validation rules in sequence
type Pipeline struct {
    validators []Validator
}

// NewPipeline creates a new validation pipeline
func NewPipeline(validators ...Validator) *Pipeline {
    return &Pipeline{
        validators: validators,
    }
}

// Validate executes all validators
func (p *Pipeline) Validate(ctx context.Context, req *models.OrderRequest) error {
    for _, v := range p.validators {
        if err := v.Validate(ctx, req); err != nil {
            return err
        }
    }
    return nil
}

// internal/validator/symbol_validator.go

package validator

import (
    "context"
    "fmt"

    "order-execution-service/internal/models"
)

// SymbolValidator validates symbol
type SymbolValidator struct {
    allowedSymbols map[string]bool
}

// NewSymbolValidator creates a new symbol validator
func NewSymbolValidator(symbols []string) *SymbolValidator {
    allowed := make(map[string]bool)
    for _, s := range symbols {
        allowed[s] = true
    }
    return &SymbolValidator{
        allowedSymbols: allowed,
    }
}

// Validate checks if symbol is allowed
func (v *SymbolValidator) Validate(ctx context.Context, req *models.OrderRequest) error {
    if !v.allowedSymbols[req.Symbol] {
        return fmt.Errorf("invalid symbol: %s", req.Symbol)
    }
    return nil
}

// internal/validator/quantity_validator.go

package validator

import (
    "context"
    "fmt"
    "math"

    "order-execution-service/internal/models"
)

// SymbolInfo contains trading rules for a symbol
type SymbolInfo struct {
    MinQuantity float64
    MaxQuantity float64
    StepSize    float64
}

// QuantityValidator validates order quantity
type QuantityValidator struct {
    symbolInfo map[string]*SymbolInfo
}

// NewQuantityValidator creates a new quantity validator
func NewQuantityValidator(symbolInfo map[string]*SymbolInfo) *QuantityValidator {
    return &QuantityValidator{
        symbolInfo: symbolInfo,
    }
}

// Validate checks quantity constraints
func (v *QuantityValidator) Validate(ctx context.Context, req *models.OrderRequest) error {
    info, ok := v.symbolInfo[req.Symbol]
    if !ok {
        return fmt.Errorf("no quantity rules for symbol: %s", req.Symbol)
    }

    // Check minimum quantity
    if req.Quantity < info.MinQuantity {
        return fmt.Errorf("quantity %.8f below minimum %.8f", req.Quantity, info.MinQuantity)
    }

    // Check maximum quantity
    if req.Quantity > info.MaxQuantity {
        return fmt.Errorf("quantity %.8f exceeds maximum %.8f", req.Quantity, info.MaxQuantity)
    }

    // Check step size
    remainder := math.Mod(req.Quantity, info.StepSize)
    if remainder > 1e-8 { // Floating point tolerance
        return fmt.Errorf("quantity %.8f does not match step size %.8f", req.Quantity, info.StepSize)
    }

    return nil
}

// internal/validator/price_validator.go

package validator

import (
    "context"
    "fmt"
    "math"

    "order-execution-service/internal/models"
)

// PriceInfo contains price rules for a symbol
type PriceInfo struct {
    MinPrice    float64
    MaxPrice    float64
    TickSize    float64
    MinNotional float64 // Minimum order value (price * quantity)
}

// PriceValidator validates order price
type PriceValidator struct {
    priceInfo map[string]*PriceInfo
}

// NewPriceValidator creates a new price validator
func NewPriceValidator(priceInfo map[string]*PriceInfo) *PriceValidator {
    return &PriceValidator{
        priceInfo: priceInfo,
    }
}

// Validate checks price constraints
func (v *PriceValidator) Validate(ctx context.Context, req *models.OrderRequest) error {
    // Market orders don't have price
    if req.Type == models.OrderTypeMarket {
        return nil
    }

    info, ok := v.priceInfo[req.Symbol]
    if !ok {
        return fmt.Errorf("no price rules for symbol: %s", req.Symbol)
    }

    // Check minimum price
    if req.Price < info.MinPrice {
        return fmt.Errorf("price %.8f below minimum %.8f", req.Price, info.MinPrice)
    }

    // Check maximum price
    if req.Price > info.MaxPrice {
        return fmt.Errorf("price %.8f exceeds maximum %.8f", req.Price, info.MaxPrice)
    }

    // Check tick size
    remainder := math.Mod(req.Price, info.TickSize)
    if remainder > 1e-8 {
        return fmt.Errorf("price %.8f does not match tick size %.8f", req.Price, info.TickSize)
    }

    // Check minimum notional
    notional := req.Price * req.Quantity
    if notional < info.MinNotional {
        return fmt.Errorf("notional %.8f below minimum %.8f", notional, info.MinNotional)
    }

    return nil
}
```

### 4.5 HMAC Signature Generation (Binance)

```go
// internal/exchange/binance/signer.go

package binance

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "net/url"
    "sort"
    "strconv"
    "time"
)

// Signer handles HMAC signature generation for Binance API
type Signer struct {
    apiSecret string
}

// NewSigner creates a new signer
func NewSigner(apiSecret string) *Signer {
    return &Signer{
        apiSecret: apiSecret,
    }
}

// SignRequest signs a request with HMAC SHA256
func (s *Signer) SignRequest(params map[string]string) (string, error) {
    // Add timestamp if not present
    if _, ok := params["timestamp"]; !ok {
        params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
    }

    // Create query string from params
    queryString := s.buildQueryString(params)

    // Create HMAC signature
    signature := s.createSignature(queryString)

    return signature, nil
}

// buildQueryString creates a sorted query string from params
func (s *Signer) buildQueryString(params map[string]string) string {
    // Sort keys for consistent signing
    keys := make([]string, 0, len(params))
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // Build query string
    values := url.Values{}
    for _, k := range keys {
        values.Add(k, params[k])
    }

    return values.Encode()
}

// createSignature creates HMAC SHA256 signature
func (s *Signer) createSignature(queryString string) string {
    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(queryString))
    return hex.EncodeToString(mac.Sum(nil))
}

// Example usage:
// params := map[string]string{
//     "symbol": "BTCUSDT",
//     "side": "BUY",
//     "type": "LIMIT",
//     "quantity": "0.001",
//     "price": "50000.00",
//     "timeInForce": "GTC",
// }
// signature := signer.SignRequest(params)
// params["signature"] = signature
```

### 4.6 Exchange API Integration (Binance Futures)

```go
// internal/exchange/binance/client.go

package binance

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strconv"
    "time"

    "order-execution-service/internal/models"
)

const (
    BinanceFuturesBaseURL = "https://fapi.binance.com"

    // Endpoints
    OrderEndpoint       = "/fapi/v1/order"
    ExchangeInfoEndpoint = "/fapi/v1/exchangeInfo"
)

// Client is a Binance Futures API client
type Client struct {
    apiKey     string
    apiSecret  string
    baseURL    string
    httpClient *http.Client
    signer     *Signer
}

// NewClient creates a new Binance client
func NewClient(apiKey, apiSecret string) *Client {
    return &Client{
        apiKey:    apiKey,
        apiSecret: apiSecret,
        baseURL:   BinanceFuturesBaseURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        signer: NewSigner(apiSecret),
    }
}

// PlaceOrder submits a new order to Binance Futures
func (c *Client) PlaceOrder(ctx context.Context, req *models.OrderRequest) (*BinanceOrderResponse, error) {
    // Build parameters
    params := map[string]string{
        "symbol":      req.Symbol,
        "side":        string(req.Side),
        "type":        string(req.Type),
        "quantity":    strconv.FormatFloat(req.Quantity, 'f', -1, 64),
        "timestamp":   strconv.FormatInt(time.Now().UnixMilli(), 10),
    }

    // Add optional parameters
    if req.Price > 0 {
        params["price"] = strconv.FormatFloat(req.Price, 'f', -1, 64)
    }
    if req.TimeInForce != "" {
        params["timeInForce"] = string(req.TimeInForce)
    }

    // Add unique client order ID for idempotency
    params["newClientOrderId"] = req.RequestID

    // Sign request
    signature, err := c.signer.SignRequest(params)
    if err != nil {
        return nil, fmt.Errorf("failed to sign request: %w", err)
    }
    params["signature"] = signature

    // Build URL
    reqURL := c.baseURL + OrderEndpoint + "?" + c.buildQueryString(params)

    // Create HTTP request
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Add headers
    httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)
    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    // Execute request
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    // Handle errors
    if resp.StatusCode != http.StatusOK {
        var errResp BinanceErrorResponse
        if err := json.Unmarshal(body, &errResp); err != nil {
            return nil, fmt.Errorf("exchange error (status %d): %s", resp.StatusCode, string(body))
        }
        return nil, fmt.Errorf("exchange error: %s (code: %d)", errResp.Msg, errResp.Code)
    }

    // Parse response
    var orderResp BinanceOrderResponse
    if err := json.Unmarshal(body, &orderResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &orderResp, nil
}

// GetOrder retrieves order information
func (c *Client) GetOrder(ctx context.Context, symbol, orderID string) (*BinanceOrderResponse, error) {
    params := map[string]string{
        "symbol":    symbol,
        "orderId":   orderID,
        "timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10),
    }

    signature, err := c.signer.SignRequest(params)
    if err != nil {
        return nil, fmt.Errorf("failed to sign request: %w", err)
    }
    params["signature"] = signature

    reqURL := c.baseURL + OrderEndpoint + "?" + c.buildQueryString(params)

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("exchange error (status %d): %s", resp.StatusCode, string(body))
    }

    var orderResp BinanceOrderResponse
    if err := json.Unmarshal(body, &orderResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &orderResp, nil
}

// CancelOrder cancels an existing order
func (c *Client) CancelOrder(ctx context.Context, symbol, orderID string) (*BinanceOrderResponse, error) {
    params := map[string]string{
        "symbol":    symbol,
        "orderId":   orderID,
        "timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10),
    }

    signature, err := c.signer.SignRequest(params)
    if err != nil {
        return nil, fmt.Errorf("failed to sign request: %w", err)
    }
    params["signature"] = signature

    reqURL := c.baseURL + OrderEndpoint + "?" + c.buildQueryString(params)

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("exchange error (status %d): %s", resp.StatusCode, string(body))
    }

    var orderResp BinanceOrderResponse
    if err := json.Unmarshal(body, &orderResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &orderResp, nil
}

func (c *Client) buildQueryString(params map[string]string) string {
    values := url.Values{}
    for k, v := range params {
        values.Add(k, v)
    }
    return values.Encode()
}

// internal/exchange/binance/types.go

package binance

// BinanceOrderResponse represents a response from Binance
type BinanceOrderResponse struct {
    OrderID             int64   `json:"orderId"`
    Symbol              string  `json:"symbol"`
    Status              string  `json:"status"`
    ClientOrderID       string  `json:"clientOrderId"`
    Price               string  `json:"price"`
    AvgPrice            string  `json:"avgPrice"`
    OrigQty             string  `json:"origQty"`
    ExecutedQty         string  `json:"executedQty"`
    CumQty              string  `json:"cumQty"`
    CumQuote            string  `json:"cumQuote"`
    TimeInForce         string  `json:"timeInForce"`
    Type                string  `json:"type"`
    ReduceOnly          bool    `json:"reduceOnly"`
    ClosePosition       bool    `json:"closePosition"`
    Side                string  `json:"side"`
    PositionSide        string  `json:"positionSide"`
    StopPrice           string  `json:"stopPrice"`
    WorkingType         string  `json:"workingType"`
    PriceProtect        bool    `json:"priceProtect"`
    OrigType            string  `json:"origType"`
    UpdateTime          int64   `json:"updateTime"`
}

// BinanceErrorResponse represents an error from Binance
type BinanceErrorResponse struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
}

// Binance order status mapping
const (
    BinanceStatusNew             = "NEW"
    BinanceStatusPartiallyFilled = "PARTIALLY_FILLED"
    BinanceStatusFilled          = "FILLED"
    BinanceStatusCanceled        = "CANCELED"
    BinanceStatusRejected        = "REJECTED"
    BinanceStatusExpired         = "EXPIRED"
)

// MapBinanceStatus maps Binance status to internal OrderState
func MapBinanceStatus(binanceStatus string) string {
    switch binanceStatus {
    case BinanceStatusNew:
        return "NEW"
    case BinanceStatusPartiallyFilled:
        return "PARTIALLY_FILLED"
    case BinanceStatusFilled:
        return "FILLED"
    case BinanceStatusCanceled:
        return "CANCELED"
    case BinanceStatusRejected:
        return "REJECTED"
    case BinanceStatusExpired:
        return "CANCELED"
    default:
        return "UNKNOWN"
    }
}
```

### 4.7 State Management with Redis

```go
// internal/state/redis_store.go

package state

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"

    "order-execution-service/internal/models"
)

const (
    OrderKeyPrefix       = "order:"
    IdempotencyKeyPrefix = "idempotency:"
    OrderTTL             = 30 * 24 * time.Hour // 30 days
    IdempotencyTTL       = 24 * time.Hour      // 24 hours
)

// RedisStore implements state storage with Redis
type RedisStore struct {
    client *redis.Client
}

// NewRedisStore creates a new Redis store
func NewRedisStore(addr, password string, db int) (*RedisStore, error) {
    client := redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     password,
        DB:           db,
        PoolSize:     100,
        MinIdleConns: 10,
    })

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &RedisStore{
        client: client,
    }, nil
}

// SaveOrder saves an order to Redis
func (s *RedisStore) SaveOrder(ctx context.Context, order *models.Order) error {
    key := OrderKeyPrefix + order.ID

    // Serialize order to JSON
    data, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("failed to marshal order: %w", err)
    }

    // Save with TTL
    if err := s.client.Set(ctx, key, data, OrderTTL).Err(); err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }

    return nil
}

// GetOrder retrieves an order from Redis
func (s *RedisStore) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
    key := OrderKeyPrefix + orderID

    data, err := s.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, fmt.Errorf("order not found: %s", orderID)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }

    var order models.Order
    if err := json.Unmarshal(data, &order); err != nil {
        return nil, fmt.Errorf("failed to unmarshal order: %w", err)
    }

    return &order, nil
}

// CheckIdempotency checks if a request_id has been processed
func (s *RedisStore) CheckIdempotency(ctx context.Context, requestID string) (*models.OrderResponse, bool, error) {
    key := IdempotencyKeyPrefix + requestID

    data, err := s.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, false, nil // Not found, proceed
    }
    if err != nil {
        return nil, false, fmt.Errorf("failed to check idempotency: %w", err)
    }

    // Found cached response
    var response models.OrderResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, false, fmt.Errorf("failed to unmarshal cached response: %w", err)
    }

    return &response, true, nil
}

// CacheResponse caches an order response for idempotency
func (s *RedisStore) CacheResponse(ctx context.Context, requestID string, response *models.OrderResponse) error {
    key := IdempotencyKeyPrefix + requestID

    data, err := json.Marshal(response)
    if err != nil {
        return fmt.Errorf("failed to marshal response: %w", err)
    }

    if err := s.client.Set(ctx, key, data, IdempotencyTTL).Err(); err != nil {
        return fmt.Errorf("failed to cache response: %w", err)
    }

    return nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
    return s.client.Close()
}
```

### 4.8 Event Publishing with NATS

```go
// internal/events/nats_publisher.go

package events

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/nats-io/nats.go"

    "order-execution-service/internal/models"
)

const (
    OrderCreatedTopic  = "order.created"
    OrderUpdatedTopic  = "order.updated"
    OrderFilledTopic   = "order.filled"
    OrderCanceledTopic = "order.canceled"
    OrderRejectedTopic = "order.rejected"
    FillEventTopic     = "fill.event"
)

// NATSPublisher publishes events to NATS
type NATSPublisher struct {
    conn *nats.Conn
}

// NewNATSPublisher creates a new NATS publisher
func NewNATSPublisher(url string) (*NATSPublisher, error) {
    conn, err := nats.Connect(url,
        nats.MaxReconnects(-1),
        nats.ReconnectWait(2*time.Second),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    return &NATSPublisher{
        conn: conn,
    }, nil
}

// PublishOrderCreated publishes an order created event
func (p *NATSPublisher) PublishOrderCreated(ctx context.Context, order *models.Order) error {
    return p.publishOrder(OrderCreatedTopic, order)
}

// PublishOrderUpdated publishes an order updated event
func (p *NATSPublisher) PublishOrderUpdated(ctx context.Context, order *models.Order) error {
    return p.publishOrder(OrderUpdatedTopic, order)
}

// PublishOrderFilled publishes an order filled event
func (p *NATSPublisher) PublishOrderFilled(ctx context.Context, order *models.Order) error {
    return p.publishOrder(OrderFilledTopic, order)
}

// PublishOrderCanceled publishes an order canceled event
func (p *NATSPublisher) PublishOrderCanceled(ctx context.Context, order *models.Order) error {
    return p.publishOrder(OrderCanceledTopic, order)
}

// PublishOrderRejected publishes an order rejected event
func (p *NATSPublisher) PublishOrderRejected(ctx context.Context, order *models.Order) error {
    return p.publishOrder(OrderRejectedTopic, order)
}

// PublishFillEvent publishes a fill event
func (p *NATSPublisher) PublishFillEvent(ctx context.Context, fill *models.FillEvent) error {
    data, err := json.Marshal(fill)
    if err != nil {
        return fmt.Errorf("failed to marshal fill event: %w", err)
    }

    if err := p.conn.Publish(FillEventTopic, data); err != nil {
        return fmt.Errorf("failed to publish fill event: %w", err)
    }

    return nil
}

func (p *NATSPublisher) publishOrder(topic string, order *models.Order) error {
    data, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("failed to marshal order: %w", err)
    }

    if err := p.conn.Publish(topic, data); err != nil {
        return fmt.Errorf("failed to publish to %s: %w", topic, err)
    }

    return nil
}

// Close closes the NATS connection
func (p *NATSPublisher) Close() error {
    p.conn.Close()
    return nil
}
```

### 4.9 Rate Limiter Implementation

```go
// internal/ratelimit/limiter.go

package ratelimit

import (
    "context"
    "fmt"
    "time"

    "golang.org/x/time/rate"
)

// Limiter implements rate limiting for exchange API calls
type Limiter struct {
    // Binance Futures limits:
    // - Order limit: 300 orders/10s = 30/sec (with buffer)
    // - Request weight: 2400/min = 40/sec
    orderLimiter  *rate.Limiter
    weightLimiter *rate.Limiter
}

// NewLimiter creates a new rate limiter
func NewLimiter() *Limiter {
    return &Limiter{
        // Order limiter: 30 orders/sec with burst of 50
        orderLimiter: rate.NewLimiter(rate.Limit(30), 50),

        // Weight limiter: 40 weight/sec with burst of 100
        weightLimiter: rate.NewLimiter(rate.Limit(40), 100),
    }
}

// WaitForOrder waits for order rate limit token
func (l *Limiter) WaitForOrder(ctx context.Context) error {
    if err := l.orderLimiter.Wait(ctx); err != nil {
        return fmt.Errorf("order rate limit wait failed: %w", err)
    }
    return nil
}

// WaitForWeight waits for weight rate limit token
func (l *Limiter) WaitForWeight(ctx context.Context, weight int) error {
    // Reserve tokens for the weight
    reservation := l.weightLimiter.ReserveN(time.Now(), weight)
    if !reservation.OK() {
        return fmt.Errorf("weight rate limit exceeded")
    }

    // Wait for the reservation
    delay := reservation.Delay()
    if delay > 0 {
        select {
        case <-time.After(delay):
            return nil
        case <-ctx.Done():
            reservation.Cancel()
            return ctx.Err()
        }
    }

    return nil
}

// AllowOrder checks if an order can be placed immediately
func (l *Limiter) AllowOrder() bool {
    return l.orderLimiter.Allow()
}

// AllowWeight checks if a request with given weight can be made immediately
func (l *Limiter) AllowWeight(weight int) bool {
    return l.weightLimiter.AllowN(time.Now(), weight)
}
```

### 4.10 Circuit Breaker Implementation

```go
// internal/circuitbreaker/breaker.go

package circuitbreaker

import (
    "context"
    "fmt"
    "time"

    "github.com/sony/gobreaker"
)

// Breaker wraps gobreaker for exchange API calls
type Breaker struct {
    cb *gobreaker.CircuitBreaker
}

// NewBreaker creates a new circuit breaker
func NewBreaker(name string) *Breaker {
    settings := gobreaker.Settings{
        Name:        name,
        MaxRequests: 3,                   // Allow 3 requests in half-open state
        Interval:    10 * time.Second,    // Reset counts after 10 seconds
        Timeout:     30 * time.Second,    // Wait 30 seconds before half-open
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Open circuit after 5 consecutive failures
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            // Log state changes
            fmt.Printf("Circuit breaker %s: %s -> %s\n", name, from, to)
        },
    }

    return &Breaker{
        cb: gobreaker.NewCircuitBreaker(settings),
    }
}

// Execute runs a function through the circuit breaker
func (b *Breaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
    return b.cb.Execute(fn)
}

// State returns the current circuit breaker state
func (b *Breaker) State() gobreaker.State {
    return b.cb.State()
}

// Counts returns the current circuit breaker counts
func (b *Breaker) Counts() gobreaker.Counts {
    return b.cb.Counts()
}
```

---

## 5. Testing Strategy

### 5.1 Mock Exchange Server

```go
// test/mock/exchange_server.go

package mock

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strconv"
    "sync"
    "time"
)

// ExchangeServer is a mock exchange server for testing
type ExchangeServer struct {
    server  *httptest.Server
    orders  map[string]*MockOrder
    mu      sync.RWMutex
    nextID  int64
}

// MockOrder represents a mock order
type MockOrder struct {
    OrderID       int64  `json:"orderId"`
    ClientOrderID string `json:"clientOrderId"`
    Symbol        string `json:"symbol"`
    Status        string `json:"status"`
    Side          string `json:"side"`
    Type          string `json:"type"`
    Price         string `json:"price"`
    OrigQty       string `json:"origQty"`
    ExecutedQty   string `json:"executedQty"`
    TimeInForce   string `json:"timeInForce"`
    UpdateTime    int64  `json:"updateTime"`
}

// NewExchangeServer creates a new mock exchange server
func NewExchangeServer() *ExchangeServer {
    es := &ExchangeServer{
        orders: make(map[string]*MockOrder),
        nextID: 1000000,
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/fapi/v1/order", es.handleOrder)

    es.server = httptest.NewServer(mux)

    return es
}

// URL returns the server URL
func (es *ExchangeServer) URL() string {
    return es.server.URL
}

// Close closes the server
func (es *ExchangeServer) Close() {
    es.server.Close()
}

// handleOrder handles order requests
func (es *ExchangeServer) handleOrder(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        es.placeOrder(w, r)
    case http.MethodGet:
        es.getOrder(w, r)
    case http.MethodDelete:
        es.cancelOrder(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// placeOrder handles order placement
func (es *ExchangeServer) placeOrder(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    symbol := r.URL.Query().Get("symbol")
    side := r.URL.Query().Get("side")
    orderType := r.URL.Query().Get("type")
    quantity := r.URL.Query().Get("quantity")
    price := r.URL.Query().Get("price")
    timeInForce := r.URL.Query().Get("timeInForce")
    clientOrderID := r.URL.Query().Get("newClientOrderId")

    // Validate required fields
    if symbol == "" || side == "" || orderType == "" || quantity == "" {
        es.sendError(w, -1102, "Mandatory parameter missing")
        return
    }

    // Generate order ID
    es.mu.Lock()
    orderID := es.nextID
    es.nextID++
    es.mu.Unlock()

    // Create mock order
    order := &MockOrder{
        OrderID:       orderID,
        ClientOrderID: clientOrderID,
        Symbol:        symbol,
        Status:        "NEW",
        Side:          side,
        Type:          orderType,
        Price:         price,
        OrigQty:       quantity,
        ExecutedQty:   "0",
        TimeInForce:   timeInForce,
        UpdateTime:    time.Now().UnixMilli(),
    }

    // Store order
    es.mu.Lock()
    es.orders[strconv.FormatInt(orderID, 10)] = order
    es.mu.Unlock()

    // Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

// getOrder handles order query
func (es *ExchangeServer) getOrder(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("orderId")

    es.mu.RLock()
    order, exists := es.orders[orderID]
    es.mu.RUnlock()

    if !exists {
        es.sendError(w, -2013, "Order does not exist")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

// cancelOrder handles order cancellation
func (es *ExchangeServer) cancelOrder(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("orderId")

    es.mu.Lock()
    order, exists := es.orders[orderID]
    if exists {
        order.Status = "CANCELED"
    }
    es.mu.Unlock()

    if !exists {
        es.sendError(w, -2013, "Order does not exist")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

// sendError sends an error response
func (es *ExchangeServer) sendError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "code": code,
        "msg":  msg,
    })
}

// SimulateFill simulates order fill
func (es *ExchangeServer) SimulateFill(orderID string, qty string) {
    es.mu.Lock()
    defer es.mu.Unlock()

    if order, exists := es.orders[orderID]; exists {
        order.ExecutedQty = qty
        if order.ExecutedQty == order.OrigQty {
            order.Status = "FILLED"
        } else {
            order.Status = "PARTIALLY_FILLED"
        }
        order.UpdateTime = time.Now().UnixMilli()
    }
}
```

### 5.2 Validation Test Cases

```go
// internal/validator/validator_test.go

package validator

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"

    "order-execution-service/internal/models"
)

func TestSymbolValidator(t *testing.T) {
    validator := NewSymbolValidator([]string{"BTCUSDT", "ETHUSDT"})

    tests := []struct {
        name    string
        symbol  string
        wantErr bool
    }{
        {"Valid BTC", "BTCUSDT", false},
        {"Valid ETH", "ETHUSDT", false},
        {"Invalid symbol", "XYZUSDT", true},
        {"Empty symbol", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &models.OrderRequest{Symbol: tt.symbol}
            err := validator.Validate(context.Background(), req)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestQuantityValidator(t *testing.T) {
    symbolInfo := map[string]*SymbolInfo{
        "BTCUSDT": {
            MinQuantity: 0.001,
            MaxQuantity: 1000.0,
            StepSize:    0.001,
        },
    }

    validator := NewQuantityValidator(symbolInfo)

    tests := []struct {
        name     string
        quantity float64
        wantErr  bool
    }{
        {"Valid quantity", 0.1, false},
        {"Min quantity", 0.001, false},
        {"Max quantity", 1000.0, false},
        {"Below min", 0.0001, true},
        {"Above max", 1001.0, true},
        {"Invalid step", 0.0015, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &models.OrderRequest{
                Symbol:   "BTCUSDT",
                Quantity: tt.quantity,
            }
            err := validator.Validate(context.Background(), req)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestPriceValidator(t *testing.T) {
    priceInfo := map[string]*PriceInfo{
        "BTCUSDT": {
            MinPrice:    0.01,
            MaxPrice:    1000000.0,
            TickSize:    0.01,
            MinNotional: 5.0,
        },
    }

    validator := NewPriceValidator(priceInfo)

    tests := []struct {
        name     string
        price    float64
        quantity float64
        wantErr  bool
    }{
        {"Valid price", 50000.00, 0.001, false},
        {"Below min notional", 1.00, 0.001, true},
        {"Invalid tick", 50000.005, 0.001, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &models.OrderRequest{
                Symbol:   "BTCUSDT",
                Type:     models.OrderTypeLimit,
                Price:    tt.price,
                Quantity: tt.quantity,
            }
            err := validator.Validate(context.Background(), req)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 5.3 Integration Test Suite

```go
// test/integration/order_test.go

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "order-execution-service/internal/models"
    "order-execution-service/test/mock"
)

func TestOrderPlacement_HappyPath(t *testing.T) {
    // Setup mock exchange server
    mockServer := mock.NewExchangeServer()
    defer mockServer.Close()

    // TODO: Setup service with mock server URL

    // Place order
    req := &models.OrderRequest{
        RequestID:   "test-" + time.Now().Format("20060102150405"),
        Symbol:      "BTCUSDT",
        Side:        models.OrderSideBuy,
        Type:        models.OrderTypeLimit,
        TimeInForce: models.TimeInForceGTC,
        Quantity:    0.001,
        Price:       50000.00,
    }

    resp, err := placeOrder(context.Background(), req)
    require.NoError(t, err)
    assert.True(t, resp.Success)
    assert.NotNil(t, resp.Order)
    assert.Equal(t, models.OrderStateNew, resp.Order.State)
}

func TestOrderPlacement_Idempotency(t *testing.T) {
    mockServer := mock.NewExchangeServer()
    defer mockServer.Close()

    req := &models.OrderRequest{
        RequestID:   "idempotency-test",
        Symbol:      "BTCUSDT",
        Side:        models.OrderSideBuy,
        Type:        models.OrderTypeLimit,
        TimeInForce: models.TimeInForceGTC,
        Quantity:    0.001,
        Price:       50000.00,
    }

    // First request
    resp1, err := placeOrder(context.Background(), req)
    require.NoError(t, err)

    // Second request with same request_id
    resp2, err := placeOrder(context.Background(), req)
    require.NoError(t, err)

    // Should return same order
    assert.Equal(t, resp1.Order.ID, resp2.Order.ID)
}

func TestOrderPlacement_ValidationError(t *testing.T) {
    mockServer := mock.NewExchangeServer()
    defer mockServer.Close()

    req := &models.OrderRequest{
        RequestID:   "validation-test",
        Symbol:      "BTCUSDT",
        Side:        models.OrderSideBuy,
        Type:        models.OrderTypeLimit,
        TimeInForce: models.TimeInForceGTC,
        Quantity:    0.0001, // Below minimum
        Price:       50000.00,
    }

    resp, err := placeOrder(context.Background(), req)
    require.NoError(t, err)
    assert.False(t, resp.Success)
    assert.Equal(t, models.OrderStateRejected, resp.Order.State)
    assert.Contains(t, resp.Error, "quantity")
}
```

### 5.4 Circuit Breaker Test Scenarios

```go
// internal/circuitbreaker/breaker_test.go

package circuitbreaker

import (
    "context"
    "errors"
    "testing"

    "github.com/sony/gobreaker"
    "github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
    breaker := NewBreaker("test")

    // Simulate 5 consecutive failures
    for i := 0; i < 5; i++ {
        _, err := breaker.Execute(context.Background(), func() (interface{}, error) {
            return nil, errors.New("simulated failure")
        })
        assert.Error(t, err)
    }

    // Circuit should be open now
    assert.Equal(t, gobreaker.StateOpen, breaker.State())

    // Next request should fail immediately
    _, err := breaker.Execute(context.Background(), func() (interface{}, error) {
        return "success", nil
    })
    assert.Error(t, err)
    assert.Equal(t, gobreaker.ErrOpenState, err)
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
    // This test requires waiting for timeout
    // Implementation depends on test duration requirements
}
```

### 5.5 Rate Limiter Tests

```go
// internal/ratelimit/limiter_test.go

package ratelimit

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestRateLimiter_OrderLimit(t *testing.T) {
    limiter := NewLimiter()

    // Should allow first 50 orders (burst)
    for i := 0; i < 50; i++ {
        assert.True(t, limiter.AllowOrder(), "Should allow order %d", i)
    }

    // 51st order should be rate limited
    assert.False(t, limiter.AllowOrder(), "Should rate limit 51st order")
}

func TestRateLimiter_WeightLimit(t *testing.T) {
    limiter := NewLimiter()

    // Should allow first 100 weight (burst)
    assert.True(t, limiter.AllowWeight(100))

    // Next weight should be rate limited
    assert.False(t, limiter.AllowWeight(1))
}

func TestRateLimiter_WaitForOrder(t *testing.T) {
    limiter := NewLimiter()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Drain burst
    for i := 0; i < 50; i++ {
        limiter.AllowOrder()
    }

    // Wait should succeed
    start := time.Now()
    err := limiter.WaitForOrder(ctx)
    duration := time.Since(start)

    assert.NoError(t, err)
    assert.Greater(t, duration, time.Duration(0), "Should have waited")
}
```

---

## 6. Deployment

### 6.1 Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order-execution-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/order-execution-service .
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose ports
EXPOSE 50051 8080 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./order-execution-service"]
```

### 6.2 Docker Compose

```yaml
# docker-compose.yml

version: '3.8'

services:
  order-execution:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: order-execution-service
    ports:
      - "50051:50051"  # gRPC
      - "8080:8080"    # HTTP (health check)
      - "9091:9091"    # Metrics
    environment:
      - LOG_LEVEL=info
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
      - REDIS_DB=0
      - NATS_URL=nats://nats:4222
      - BINANCE_API_KEY=${BINANCE_API_KEY}
      - BINANCE_API_SECRET=${BINANCE_API_SECRET}
      - RATE_LIMIT_ORDERS_PER_SEC=30
      - RATE_LIMIT_WEIGHT_PER_SEC=40
      - CIRCUIT_BREAKER_THRESHOLD=5
      - CIRCUIT_BREAKER_TIMEOUT=30s
    depends_on:
      - redis
      - nats
    networks:
      - trading-net
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: order-execution-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    networks:
      - trading-net
    restart: unless-stopped

  nats:
    image: nats:2.10-alpine
    container_name: order-execution-nats
    ports:
      - "4222:4222"  # Client
      - "8222:8222"  # HTTP monitoring
    command: "-js -m 8222"
    networks:
      - trading-net
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    container_name: order-execution-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - trading-net
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: order-execution-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
      - ./configs/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./configs/grafana/datasources:/etc/grafana/provisioning/datasources
    depends_on:
      - prometheus
    networks:
      - trading-net
    restart: unless-stopped

networks:
  trading-net:
    driver: bridge

volumes:
  redis-data:
  prometheus-data:
  grafana-data:
```

### 6.3 Configuration Management

```yaml
# configs/config.yaml

server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9091
  shutdown_timeout: 30s

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 100
  min_idle_conns: 10

nats:
  url: "nats://localhost:4222"
  max_reconnects: -1
  reconnect_wait: 2s

exchange:
  name: "binance_futures"
  base_url: "https://fapi.binance.com"
  api_key: "${BINANCE_API_KEY}"
  api_secret: "${BINANCE_API_SECRET}"
  timeout: 10s
  max_idle_conns: 100
  max_idle_conns_per_host: 100

rate_limit:
  orders_per_sec: 30
  orders_burst: 50
  weight_per_sec: 40
  weight_burst: 100

circuit_breaker:
  max_requests: 3
  interval: 10s
  timeout: 30s
  failure_threshold: 5

validation:
  symbols:
    - "BTCUSDT"
    - "ETHUSDT"
    - "BNBUSDT"

  symbol_info:
    BTCUSDT:
      min_quantity: 0.001
      max_quantity: 1000.0
      step_size: 0.001
      min_price: 0.01
      max_price: 1000000.0
      tick_size: 0.01
      min_notional: 5.0

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  path: "/metrics"
```

### 6.4 Secrets Management

```bash
# .env.example

# Exchange API credentials (NEVER commit actual values)
BINANCE_API_KEY=your_api_key_here
BINANCE_API_SECRET=your_api_secret_here

# Redis (if password protected)
REDIS_PASSWORD=

# Service configuration
LOG_LEVEL=info
ENVIRONMENT=production
```

### 6.5 Kubernetes Deployment (Optional)

```yaml
# deployments/k8s/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-execution-service
  labels:
    app: order-execution
spec:
  replicas: 2
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
        image: order-execution-service:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 8080
          name: http
        - containerPort: 9091
          name: metrics
        env:
        - name: REDIS_ADDR
          value: "redis-service:6379"
        - name: NATS_URL
          value: "nats://nats-service:4222"
        - name: BINANCE_API_KEY
          valueFrom:
            secretKeyRef:
              name: exchange-secrets
              key: api-key
        - name: BINANCE_API_SECRET
          valueFrom:
            secretKeyRef:
              name: exchange-secrets
              key: api-secret
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: order-execution-service
spec:
  selector:
    app: order-execution
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9091
    targetPort: 9091
  type: ClusterIP
```

### 6.6 Health Check Implementation

```go
// internal/handler/health_handler.go

package handler

import (
    "encoding/json"
    "net/http"
    "time"
)

type HealthHandler struct {
    startTime time.Time
    redis     HealthChecker
    nats      HealthChecker
}

type HealthChecker interface {
    Ping() error
}

type HealthResponse struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Uptime    string            `json:"uptime"`
    Checks    map[string]string `json:"checks"`
}

func NewHealthHandler(redis, nats HealthChecker) *HealthHandler {
    return &HealthHandler{
        startTime: time.Now(),
        redis:     redis,
        nats:      nats,
    }
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    checks := make(map[string]string)
    overall := "healthy"

    // Check Redis
    if err := h.redis.Ping(); err != nil {
        checks["redis"] = "unhealthy: " + err.Error()
        overall = "degraded"
    } else {
        checks["redis"] = "healthy"
    }

    // Check NATS
    if err := h.nats.Ping(); err != nil {
        checks["nats"] = "unhealthy: " + err.Error()
        overall = "degraded"
    } else {
        checks["nats"] = "healthy"
    }

    response := HealthResponse{
        Status:    overall,
        Timestamp: time.Now(),
        Uptime:    time.Since(h.startTime).String(),
        Checks:    checks,
    }

    w.Header().Set("Content-Type", "application/json")
    if overall != "healthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    json.NewEncoder(w).Encode(response)
}
```

---

## 7. Observability

### 7.1 Key Metrics

```go
// internal/metrics/metrics.go

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Order metrics
    OrdersPlaced = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_placed_total",
            Help: "Total number of orders placed",
        },
        []string{"symbol", "side", "type"},
    )

    OrdersFilled = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_filled_total",
            Help: "Total number of orders filled",
        },
        []string{"symbol", "side"},
    )

    OrdersRejected = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_rejected_total",
            Help: "Total number of orders rejected",
        },
        []string{"symbol", "reason"},
    )

    OrderSubmissionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "order_submission_duration_seconds",
            Help:    "Order submission latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"symbol", "exchange"},
    )

    // Circuit breaker metrics
    CircuitBreakerState = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
        },
        []string{"name"},
    )

    // Rate limiter metrics
    RateLimitReached = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rate_limit_reached_total",
            Help: "Total number of times rate limit was reached",
        },
        []string{"type"}, // order or weight
    )

    // Validation metrics
    ValidationFailures = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "validation_failures_total",
            Help: "Total number of validation failures",
        },
        []string{"rule"},
    )

    // Exchange API metrics
    ExchangeAPIRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "exchange_api_requests_total",
            Help: "Total number of exchange API requests",
        },
        []string{"exchange", "endpoint", "status"},
    )

    ExchangeAPIErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "exchange_api_errors_total",
            Help: "Total number of exchange API errors",
        },
        []string{"exchange", "error_code"},
    )

    // Cache metrics
    CacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits (idempotency)",
        },
    )

    CacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
    )
)
```

### 7.2 Prometheus Configuration

```yaml
# configs/prometheus.yml

global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'order-execution-service'
    static_configs:
      - targets: ['order-execution:9091']
        labels:
          service: 'order-execution'
          environment: 'production'
```

### 7.3 Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Order Execution Service",
    "panels": [
      {
        "title": "Order Placement Rate",
        "targets": [
          {
            "expr": "rate(orders_placed_total[5m])"
          }
        ]
      },
      {
        "title": "Order Submission Latency (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(order_submission_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Circuit Breaker State",
        "targets": [
          {
            "expr": "circuit_breaker_state"
          }
        ]
      },
      {
        "title": "Order Rejection Reasons",
        "targets": [
          {
            "expr": "orders_rejected_total"
          }
        ]
      },
      {
        "title": "Exchange API Error Rate",
        "targets": [
          {
            "expr": "rate(exchange_api_errors_total[5m])"
          }
        ]
      },
      {
        "title": "Rate Limit Events",
        "targets": [
          {
            "expr": "rate(rate_limit_reached_total[5m])"
          }
        ]
      },
      {
        "title": "Cache Hit Ratio",
        "targets": [
          {
            "expr": "cache_hits_total / (cache_hits_total + cache_misses_total)"
          }
        ]
      }
    ]
  }
}
```

### 7.4 Logging Standards

```go
// pkg/logger/logger.go

package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger(level string) (*zap.Logger, error) {
    config := zap.Config{
        Level:            parseLevel(level),
        Encoding:         "json",
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
        EncoderConfig: zapcore.EncoderConfig{
            TimeKey:        "timestamp",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            MessageKey:     "message",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
    }

    return config.Build()
}

// Log format:
// {
//   "timestamp": "2025-10-02T10:30:00.000Z",
//   "level": "info",
//   "message": "Order placed successfully",
//   "order_id": "123456",
//   "symbol": "BTCUSDT",
//   "side": "BUY",
//   "quantity": 0.001,
//   "correlation_id": "abc-123"
// }
```

### 7.5 Dashboard UI Requirements

**Required Panels:**

1. **Active Orders Dashboard**
   - Current open orders
   - Order count by state
   - Real-time order updates

2. **Order Submission Latency**
   - p50, p95, p99 latencies
   - Breakdown by exchange
   - Time-series chart

3. **Circuit Breaker Status**
   - Current state indicator
   - State transition history
   - Failure count

4. **Maker/Taker Ratio**
   - Fill distribution
   - Fee analysis
   - Time-series breakdown

5. **Order Rejection Analysis**
   - Rejection reasons breakdown
   - Rejection rate over time
   - Top rejection causes

6. **Rate Limit Monitoring**
   - Current usage vs. limits
   - Rate limit events
   - Per-endpoint breakdown

7. **System Health**
   - Uptime
   - Dependency health (Redis, NATS)
   - Error rates

---

## Summary

This development plan provides a complete roadmap for building a production-ready Order Execution Service over 14 days. The plan emphasizes:

1. **Technology Stack:** Go for performance, gRPC for typed RPC, Redis for state, NATS for pub/sub
2. **Architecture:** Clean separation of concerns with validation pipeline, rate limiting, and circuit breaker
3. **Phased Development:** 6 distinct phases building from foundation to full production readiness
4. **Implementation Details:** Complete code examples for critical components
5. **Testing:** Comprehensive test strategy with mock server and integration tests
6. **Deployment:** Docker-based deployment with health checks and configuration management
7. **Observability:** Full metrics, logging, and dashboard specifications

**Next Steps:**
1. Set up development environment
2. Initialize Go project structure
3. Begin Phase 1: Foundation & RPC Server
4. Follow the 14-day development roadmap

The service will be production-ready with proper error handling, rate limiting, circuit breaking, idempotency, and comprehensive observability.
