# Dashboard Server Service - Development Plan

**Service ID:** 05
**Service Name:** Dashboard Server Service
**Last Updated:** 2025-10-02
**Version:** 1.0
**Status:** Planning

---

## Table of Contents

1. [Service Overview](#service-overview)
2. [Technology Stack Recommendation](#technology-stack-recommendation)
3. [Architecture Design](#architecture-design)
4. [Development Phases](#development-phases)
5. [Implementation Details](#implementation-details)
6. [Testing Strategy](#testing-strategy)
7. [Deployment](#deployment)
8. [Observability](#observability)
9. [Code Examples](#code-examples)

---

## Service Overview

### Responsibility

State aggregation and real-time broadcasting to UI clients (TUI and Web Dashboard).

### Core Functions

- Multi-source state aggregation from all backend services
- WebSocket server for real-time client connections
- Update rate differentiation (TUI: 100ms, Web: 250ms)
- Efficient message serialization with minimal bandwidth
- Client subscription management and filtering
- Connection pooling and lifecycle management

### Key Requirements

- Support 100+ concurrent WebSocket connections
- Low latency state broadcasts (<50ms from source to client)
- Bandwidth optimization through efficient serialization
- Graceful handling of backend service failures
- Client reconnection support with state recovery

---

## Technology Stack Recommendation

### Primary Recommendation: Go

**Rationale:**
- Native WebSocket support with `gorilla/websocket` or `nhooyr.io/websocket`
- Excellent concurrency model (goroutines) for managing multiple connections
- Low memory footprint per connection (~4KB per goroutine)
- Fast compilation and deployment
- Built-in profiling tools (pprof)
- Strong standard library for HTTP/WebSocket servers

**Language:** Go 1.21+

**WebSocket Library:**
```
Primary: github.com/gorilla/websocket (battle-tested, widely adopted)
Alternative: nhooyr.io/websocket (modern, context-aware)
```

**Serialization Format:**
```
Primary: MessagePack (github.com/vmihailenco/msgpack/v5)
  - 2-10x smaller than JSON
  - Faster serialization/deserialization
  - Schema-free but type-safe

Alternative: Protocol Buffers
  - Best performance but requires schema management
  - Use if extreme performance is needed

Fallback: JSON (encoding/json)
  - For debugging and web clients that prefer JSON
  - Enable with query parameter: ws://host?format=json
```

**Client Libraries for Data Sources:**
```
- gRPC: google.golang.org/grpc
- REST: net/http (standard library)
- Redis Pub/Sub: github.com/redis/go-redis/v9
- NATS: github.com/nats-io/nats.go
```

**Testing Frameworks:**
```
- Unit Testing: testing (standard library)
- WebSocket Testing: gorilla/websocket + httptest
- Load Testing: github.com/tsenart/vegeta or custom Go benchmark
- Mocking: github.com/stretchr/testify/mock
- Assertion: github.com/stretchr/testify/assert
```

**Additional Libraries:**
```
- Logging: github.com/rs/zerolog (structured, fast)
- Metrics: github.com/prometheus/client_golang
- Configuration: github.com/spf13/viper
- Context Management: context (standard library)
```

### Alternative Stack: Node.js (TypeScript)

**Use Case:** If team is primarily JavaScript-focused

**Technology:**
- Runtime: Node.js 20+ with TypeScript 5+
- WebSocket: `ws` library or `socket.io`
- Serialization: `@msgpack/msgpack` or `protobufjs`
- Testing: Jest + `ws` for WebSocket testing

**Trade-offs:**
- Easier for JS teams but higher memory usage per connection
- Single-threaded event loop (need cluster mode for multi-core)

### Alternative Stack: Rust

**Use Case:** If extreme performance is critical

**Technology:**
- Framework: `tokio` + `axum` or `actix-web`
- WebSocket: `tokio-tungstenite`
- Serialization: `rmp-serde` (MessagePack) or `prost` (Protobuf)
- Testing: `cargo test` + `tokio-test`

**Trade-offs:**
- Absolute best performance and safety
- Steeper learning curve, slower development

---

## Architecture Design

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Dashboard Server Service                  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              WebSocket Server Layer                   │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐     │  │
│  │  │ TUI Client │  │ TUI Client │  │ Web Client │ ... │  │
│  │  │  (100ms)   │  │  (100ms)   │  │  (250ms)   │     │  │
│  │  └────────────┘  └────────────┘  └────────────┘     │  │
│  └───────────┬──────────────────────────────────────────┘  │
│              │                                               │
│  ┌───────────▼──────────────────────────────────────────┐  │
│  │         Connection Manager                            │  │
│  │  - Client registry (map[clientID]*Client)             │  │
│  │  - Subscription routing                               │  │
│  │  - Rate limiting per client type                      │  │
│  └───────────┬──────────────────────────────────────────┘  │
│              │                                               │
│  ┌───────────▼──────────────────────────────────────────┐  │
│  │         State Aggregation Engine                      │  │
│  │  - Market data state cache                            │  │
│  │  - Order state cache                                  │  │
│  │  - Account state cache                                │  │
│  │  - P&L state cache                                    │  │
│  │  - Strategy state cache                               │  │
│  └───────────┬──────────────────────────────────────────┘  │
│              │                                               │
│  ┌───────────▼──────────────────────────────────────────┐  │
│  │         Broadcast Scheduler                           │  │
│  │  - TUI broadcast ticker (100ms)                       │  │
│  │  - Web broadcast ticker (250ms)                       │  │
│  │  - Differential updates (send only changes)           │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
└────────┬──────────────────────────────────────────┬─────────┘
         │                                           │
         │ Input Sources                             │ Output
         │                                           │
    ┌────▼────────────────────────────────┐    ┌───▼──────┐
    │  gRPC/REST APIs:                    │    │ Metrics  │
    │  - Market Data Service              │    │ (Prom)   │
    │  - Order Execution Service          │    └──────────┘
    │  - Account Monitor Service          │
    │  - Strategy Engine Service          │
    │  - Risk Manager Service             │
    │                                     │
    │  Pub/Sub Subscriptions:             │
    │  - Fill events                      │
    │  - Order updates                    │
    │  - Market data updates              │
    │  - Risk alerts                      │
    └─────────────────────────────────────┘
```

### Component Breakdown

#### 1. WebSocket Server Layer

**Responsibilities:**
- Accept WebSocket connections from clients
- Handle WebSocket handshake and upgrade
- Manage connection lifecycle (connect, disconnect, ping/pong)
- Parse incoming client messages (subscriptions, commands)
- Send serialized state updates to clients

**Key Features:**
- Client authentication via JWT tokens (optional)
- Connection heartbeat with ping/pong (30s interval)
- Automatic reconnection support with sequence numbers
- Graceful shutdown with connection draining

#### 2. Connection Manager

**Responsibilities:**
- Maintain registry of all active connections
- Track client metadata (type: TUI/Web, subscriptions, last update time)
- Route messages to appropriate clients based on subscriptions
- Implement per-client rate limiting

**Data Structures:**
```go
type Client struct {
    ID            string
    Type          ClientType  // TUI or Web
    Conn          *websocket.Conn
    Subscriptions map[string]bool
    SendChan      chan []byte
    LastUpdate    time.Time
    Context       context.Context
    Cancel        context.CancelFunc
}

type ClientType int
const (
    ClientTypeTUI ClientType = iota  // 100ms updates
    ClientTypeWeb                     // 250ms updates
)
```

#### 3. State Aggregation Engine

**Responsibilities:**
- Query data from backend services on startup
- Subscribe to real-time update streams (pub/sub)
- Maintain in-memory cache of current state
- Compute derived metrics (e.g., total P&L, position summary)
- Handle backend service failures gracefully

**State Categories:**
```
1. Market Data State
   - Order books per symbol
   - Latest trade prices
   - 24h volume, high, low

2. Order State
   - Active orders (open, partially filled)
   - Recent order history (last 100)

3. Account State
   - Current balances per asset
   - Total account value in USD
   - Available margin

4. Position State
   - Open positions per symbol
   - Unrealized P&L
   - Average entry price

5. Strategy State
   - Active strategies
   - Strategy P&L
   - Recent signals

6. Risk State
   - Current leverage
   - Risk limits utilization
   - Active alerts
```

**Caching Strategy:**
- Use sync.RWMutex for thread-safe state access
- Cache invalidation on pub/sub updates
- Periodic full refresh from backend services (every 60s)

#### 4. Broadcast Scheduler

**Responsibilities:**
- Run periodic broadcast tickers for different client types
- Serialize state snapshots efficiently
- Implement differential updates (only send changed fields)
- Batch updates when possible

**Broadcasting Strategy:**
```go
// TUI Ticker: 100ms interval
tuiTicker := time.NewTicker(100 * time.Millisecond)

// Web Ticker: 250ms interval
webTicker := time.NewTicker(250 * time.Millisecond)

// On each tick:
// 1. Get current state snapshot
// 2. Compare with previous state
// 3. Serialize differential update
// 4. Send to all subscribed clients of that type
```

**Differential Update Algorithm:**
```
1. Maintain lastSentState per client
2. On broadcast:
   - Compare currentState with lastSentState
   - Build patch object with only changed fields
   - Serialize patch with MessagePack
   - Send to client
   - Update lastSentState = currentState
3. On full refresh request:
   - Send complete state snapshot
```

### Message Format Design

#### Client-to-Server Messages

```json
// Subscribe to specific data types
{
  "type": "subscribe",
  "channels": ["market_data", "orders", "positions", "account"]
}

// Unsubscribe
{
  "type": "unsubscribe",
  "channels": ["market_data"]
}

// Request full state refresh
{
  "type": "refresh"
}

// Manual trading command (optional)
{
  "type": "place_order",
  "payload": {
    "symbol": "BTCUSDT",
    "side": "BUY",
    "quantity": 0.01,
    "price": 50000
  }
}
```

#### Server-to-Client Messages

```json
// Full state snapshot (MessagePack encoded)
{
  "type": "snapshot",
  "seq": 12345,
  "timestamp": "2025-10-02T10:30:00Z",
  "data": {
    "market_data": { /* ... */ },
    "orders": { /* ... */ },
    "positions": { /* ... */ },
    "account": { /* ... */ }
  }
}

// Differential update (MessagePack encoded)
{
  "type": "update",
  "seq": 12346,
  "timestamp": "2025-10-02T10:30:00.100Z",
  "changes": {
    "positions.BTCUSDT.unrealized_pnl": 125.50,
    "account.total_balance": 10250.75
  }
}

// Error message
{
  "type": "error",
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests"
}
```

### Serialization Strategy

**MessagePack Advantages:**
- Binary format, smaller payload size
- Fast serialization/deserialization
- Schema-free (flexible)
- Wide language support

**Implementation:**
```go
// Serialize state to MessagePack
data, err := msgpack.Marshal(state)

// Deserialize from MessagePack
var state State
err := msgpack.Unmarshal(data, &state)
```

**JSON Fallback:**
- For debugging: `ws://localhost:8080?format=json`
- For web clients that prefer JSON over MessagePack
- Trade-off: ~2-5x larger payload size

---

## Development Phases

### Phase 1: WebSocket Server Setup (Week 1)

**Goals:**
- Set up basic WebSocket server with connection handling
- Implement ping/pong heartbeat mechanism
- Create client connection manager

**Deliverables:**
- WebSocket server listening on port 8080
- Connection acceptance and lifecycle management
- Basic client registry with add/remove operations
- Health check endpoint (`GET /health`)
- Metrics endpoint (`GET /metrics`)

**Tasks:**
1. Initialize Go project with modules
2. Implement WebSocket upgrade handler
3. Create Client struct and connection manager
4. Implement ping/pong with 30s interval
5. Add graceful shutdown logic
6. Write unit tests for connection handling

**Success Criteria:**
- Server accepts WebSocket connections
- Ping/pong keeps connections alive
- Graceful disconnection works
- Can handle 100+ concurrent connections

---

### Phase 2: State Aggregation from Data Sources (Week 2)

**Goals:**
- Connect to backend services (gRPC/REST)
- Subscribe to pub/sub topics for real-time updates
- Build in-memory state cache with thread-safe access

**Deliverables:**
- gRPC clients for querying backend services
- Pub/Sub subscribers (Redis/NATS)
- State cache implementation with sync.RWMutex
- State refresh logic (initial load + periodic updates)

**Tasks:**
1. Implement gRPC clients for all backend services
2. Create state cache data structures
3. Implement pub/sub subscription handlers
4. Build state aggregation logic
5. Add error handling for backend failures
6. Write integration tests with mock backends

**Success Criteria:**
- Successfully queries initial state from all services
- Receives real-time updates via pub/sub
- State cache correctly updated on events
- Graceful degradation when backend service is down

---

### Phase 3: Update Broadcasting Mechanism (Week 3)

**Goals:**
- Implement broadcast schedulers for TUI and Web clients
- Build efficient serialization with MessagePack
- Create differential update algorithm

**Deliverables:**
- TUI broadcast ticker (100ms)
- Web broadcast ticker (250ms)
- MessagePack serialization
- Differential update logic

**Tasks:**
1. Create broadcast scheduler goroutines
2. Implement state snapshot serialization
3. Build differential update algorithm
4. Add per-client lastSentState tracking
5. Optimize broadcast performance
6. Write benchmark tests

**Success Criteria:**
- TUI clients receive updates every 100ms
- Web clients receive updates every 250ms
- Differential updates reduce bandwidth by >70%
- CPU usage remains <5% under 100 connections

---

### Phase 4: Client Subscription System (Week 4)

**Goals:**
- Implement subscription management
- Support selective data streaming
- Add client message parsing

**Deliverables:**
- Subscription message handlers
- Channel filtering logic
- Client command processing

**Tasks:**
1. Implement subscribe/unsubscribe message handling
2. Add channel filtering to broadcast logic
3. Build client command router
4. Implement full state refresh on request
5. Add subscription validation
6. Write unit tests for subscription logic

**Success Criteria:**
- Clients can subscribe to specific channels
- Only subscribed data is sent to clients
- Refresh command sends full state snapshot
- Invalid subscriptions return error messages

---

### Phase 5: Message Serialization Optimization (Week 5)

**Goals:**
- Optimize MessagePack encoding
- Implement compression (optional)
- Add JSON fallback mode

**Deliverables:**
- Optimized MessagePack encoding
- Optional gzip compression for large payloads
- JSON mode support via query parameter
- Bandwidth usage metrics

**Tasks:**
1. Profile MessagePack serialization performance
2. Implement struct tags for efficient encoding
3. Add optional gzip compression (if payload > 1KB)
4. Create JSON serialization mode
5. Add bandwidth metrics to Prometheus
6. Run serialization benchmarks

**Success Criteria:**
- MessagePack payloads 3-5x smaller than JSON
- Serialization time <1ms per message
- Compression reduces bandwidth by additional 30-50%
- JSON mode works for debugging

---

### Phase 6: Load Testing and Observability (Week 6)

**Goals:**
- Comprehensive load testing
- Performance profiling and optimization
- Production-ready observability

**Deliverables:**
- Load testing suite simulating 100+ clients
- Performance benchmarks and reports
- Prometheus metrics dashboard
- Structured logging with zerolog

**Tasks:**
1. Build WebSocket client simulator for load testing
2. Run load tests with 100, 250, 500 concurrent clients
3. Profile CPU and memory usage (pprof)
4. Optimize hot paths identified by profiling
5. Add comprehensive Prometheus metrics
6. Create Grafana dashboard
7. Implement structured logging
8. Write performance test suite

**Success Criteria:**
- Handles 100+ concurrent clients with <10MB memory per client
- Broadcast latency p99 <50ms
- CPU usage <20% under full load
- All metrics exposed to Prometheus
- Logs are structured and queryable

---

## Implementation Details

### WebSocket Connection Handling

```go
package server

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/rs/zerolog"
    "github.com/vmihailenco/msgpack/v5"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        // TODO: Implement proper origin checking in production
        return true
    },
}

type Server struct {
    clients       map[string]*Client
    clientsMu     sync.RWMutex
    stateCache    *StateCache
    logger        zerolog.Logger
    shutdownCtx   context.Context
    shutdownCancel context.CancelFunc
}

type Client struct {
    ID            string
    Type          ClientType
    Conn          *websocket.Conn
    Subscriptions map[string]bool
    SendChan      chan []byte
    LastUpdate    time.Time
    LastState     *State
    Context       context.Context
    Cancel        context.CancelFunc
    Format        SerializationFormat
}

type ClientType int

const (
    ClientTypeTUI ClientType = iota
    ClientTypeWeb
)

type SerializationFormat int

const (
    FormatMessagePack SerializationFormat = iota
    FormatJSON
)

func NewServer(logger zerolog.Logger) *Server {
    ctx, cancel := context.WithCancel(context.Background())
    return &Server{
        clients:        make(map[string]*Client),
        stateCache:     NewStateCache(),
        logger:         logger,
        shutdownCtx:    ctx,
        shutdownCancel: cancel,
    }
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
        return
    }

    // Parse client type from query parameter
    clientType := ClientTypeTUI
    if r.URL.Query().Get("type") == "web" {
        clientType = ClientTypeWeb
    }

    // Parse serialization format
    format := FormatMessagePack
    if r.URL.Query().Get("format") == "json" {
        format = FormatJSON
    }

    client := s.createClient(conn, clientType, format)
    s.registerClient(client)

    s.logger.Info().
        Str("client_id", client.ID).
        Str("client_type", clientType.String()).
        Str("format", format.String()).
        Msg("Client connected")

    // Start goroutines for this client
    go s.clientReader(client)
    go s.clientWriter(client)
}

func (s *Server) createClient(conn *websocket.Conn, clientType ClientType, format SerializationFormat) *Client {
    ctx, cancel := context.WithCancel(s.shutdownCtx)
    return &Client{
        ID:            generateClientID(),
        Type:          clientType,
        Conn:          conn,
        Subscriptions: make(map[string]bool),
        SendChan:      make(chan []byte, 256),
        LastUpdate:    time.Now(),
        Context:       ctx,
        Cancel:        cancel,
        Format:        format,
    }
}

func (s *Server) registerClient(client *Client) {
    s.clientsMu.Lock()
    s.clients[client.ID] = client
    s.clientsMu.Unlock()
}

func (s *Server) unregisterClient(client *Client) {
    s.clientsMu.Lock()
    delete(s.clients, client.ID)
    s.clientsMu.Unlock()

    client.Cancel()
    close(client.SendChan)

    s.logger.Info().
        Str("client_id", client.ID).
        Msg("Client disconnected")
}

func (s *Server) clientReader(client *Client) {
    defer s.unregisterClient(client)

    client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    client.Conn.SetPongHandler(func(string) error {
        client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })

    for {
        select {
        case <-client.Context.Done():
            return
        default:
            _, message, err := client.Conn.ReadMessage()
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                    s.logger.Error().Err(err).Str("client_id", client.ID).Msg("WebSocket read error")
                }
                return
            }

            s.handleClientMessage(client, message)
        }
    }
}

func (s *Server) clientWriter(client *Client) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-client.Context.Done():
            return
        case message := <-client.SendChan:
            client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := client.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
                s.logger.Error().Err(err).Str("client_id", client.ID).Msg("WebSocket write error")
                return
            }
        case <-ticker.C:
            client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (s *Server) handleClientMessage(client *Client, message []byte) {
    var msg ClientMessage
    if err := json.Unmarshal(message, &msg); err != nil {
        s.logger.Error().Err(err).Str("client_id", client.ID).Msg("Failed to parse client message")
        return
    }

    switch msg.Type {
    case "subscribe":
        s.handleSubscribe(client, msg.Channels)
    case "unsubscribe":
        s.handleUnsubscribe(client, msg.Channels)
    case "refresh":
        s.handleRefresh(client)
    default:
        s.logger.Warn().Str("type", msg.Type).Msg("Unknown message type")
    }
}

func (s *Server) handleSubscribe(client *Client, channels []string) {
    for _, ch := range channels {
        client.Subscriptions[ch] = true
    }
    s.logger.Info().
        Str("client_id", client.ID).
        Strs("channels", channels).
        Msg("Client subscribed")
}

func (s *Server) handleUnsubscribe(client *Client, channels []string) {
    for _, ch := range channels {
        delete(client.Subscriptions, ch)
    }
    s.logger.Info().
        Str("client_id", client.ID).
        Strs("channels", channels).
        Msg("Client unsubscribed")
}

func (s *Server) handleRefresh(client *Client) {
    state := s.stateCache.GetFullState()
    s.sendFullState(client, state)
}

type ClientMessage struct {
    Type     string   `json:"type"`
    Channels []string `json:"channels,omitempty"`
}

func generateClientID() string {
    return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

func (ct ClientType) String() string {
    switch ct {
    case ClientTypeTUI:
        return "TUI"
    case ClientTypeWeb:
        return "Web"
    default:
        return "Unknown"
    }
}

func (sf SerializationFormat) String() string {
    switch sf {
    case FormatMessagePack:
        return "MessagePack"
    case FormatJSON:
        return "JSON"
    default:
        return "Unknown"
    }
}
```

### State Aggregation Logic

```go
package server

import (
    "sync"
    "time"
)

type StateCache struct {
    mu            sync.RWMutex
    marketData    map[string]*MarketData
    orders        []*Order
    positions     map[string]*Position
    account       *Account
    strategies    map[string]*Strategy
    lastUpdate    time.Time
}

type State struct {
    MarketData map[string]*MarketData `msgpack:"market_data" json:"market_data"`
    Orders     []*Order               `msgpack:"orders" json:"orders"`
    Positions  map[string]*Position   `msgpack:"positions" json:"positions"`
    Account    *Account               `msgpack:"account" json:"account"`
    Strategies map[string]*Strategy   `msgpack:"strategies" json:"strategies"`
    Timestamp  time.Time              `msgpack:"timestamp" json:"timestamp"`
    Sequence   uint64                 `msgpack:"seq" json:"seq"`
}

type MarketData struct {
    Symbol    string  `msgpack:"symbol" json:"symbol"`
    LastPrice float64 `msgpack:"last_price" json:"last_price"`
    BidPrice  float64 `msgpack:"bid_price" json:"bid_price"`
    AskPrice  float64 `msgpack:"ask_price" json:"ask_price"`
    Volume24h float64 `msgpack:"volume_24h" json:"volume_24h"`
    High24h   float64 `msgpack:"high_24h" json:"high_24h"`
    Low24h    float64 `msgpack:"low_24h" json:"low_24h"`
}

type Order struct {
    ID         string    `msgpack:"id" json:"id"`
    Symbol     string    `msgpack:"symbol" json:"symbol"`
    Side       string    `msgpack:"side" json:"side"`
    Type       string    `msgpack:"type" json:"type"`
    Price      float64   `msgpack:"price" json:"price"`
    Quantity   float64   `msgpack:"quantity" json:"quantity"`
    Filled     float64   `msgpack:"filled" json:"filled"`
    Status     string    `msgpack:"status" json:"status"`
    CreatedAt  time.Time `msgpack:"created_at" json:"created_at"`
}

type Position struct {
    Symbol          string  `msgpack:"symbol" json:"symbol"`
    Side            string  `msgpack:"side" json:"side"`
    Quantity        float64 `msgpack:"quantity" json:"quantity"`
    EntryPrice      float64 `msgpack:"entry_price" json:"entry_price"`
    MarkPrice       float64 `msgpack:"mark_price" json:"mark_price"`
    UnrealizedPnL   float64 `msgpack:"unrealized_pnl" json:"unrealized_pnl"`
    RealizedPnL     float64 `msgpack:"realized_pnl" json:"realized_pnl"`
    LiquidationPrice float64 `msgpack:"liquidation_price" json:"liquidation_price"`
}

type Account struct {
    TotalBalance     float64            `msgpack:"total_balance" json:"total_balance"`
    AvailableBalance float64            `msgpack:"available_balance" json:"available_balance"`
    MarginUsed       float64            `msgpack:"margin_used" json:"margin_used"`
    UnrealizedPnL    float64            `msgpack:"unrealized_pnl" json:"unrealized_pnl"`
    Balances         map[string]float64 `msgpack:"balances" json:"balances"`
}

type Strategy struct {
    ID        string  `msgpack:"id" json:"id"`
    Name      string  `msgpack:"name" json:"name"`
    Status    string  `msgpack:"status" json:"status"`
    PnL       float64 `msgpack:"pnl" json:"pnl"`
    Trades    int     `msgpack:"trades" json:"trades"`
    WinRate   float64 `msgpack:"win_rate" json:"win_rate"`
}

func NewStateCache() *StateCache {
    return &StateCache{
        marketData: make(map[string]*MarketData),
        orders:     make([]*Order, 0),
        positions:  make(map[string]*Position),
        account:    &Account{Balances: make(map[string]float64)},
        strategies: make(map[string]*Strategy),
        lastUpdate: time.Now(),
    }
}

func (sc *StateCache) GetFullState() *State {
    sc.mu.RLock()
    defer sc.mu.RUnlock()

    return &State{
        MarketData: sc.marketData,
        Orders:     sc.orders,
        Positions:  sc.positions,
        Account:    sc.account,
        Strategies: sc.strategies,
        Timestamp:  time.Now(),
    }
}

func (sc *StateCache) UpdateMarketData(symbol string, data *MarketData) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.marketData[symbol] = data
    sc.lastUpdate = time.Now()
}

func (sc *StateCache) UpdateOrder(order *Order) {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    // Replace or append order
    found := false
    for i, o := range sc.orders {
        if o.ID == order.ID {
            sc.orders[i] = order
            found = true
            break
        }
    }
    if !found {
        sc.orders = append(sc.orders, order)
    }

    sc.lastUpdate = time.Now()
}

func (sc *StateCache) UpdatePosition(symbol string, position *Position) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.positions[symbol] = position
    sc.lastUpdate = time.Now()
}

func (sc *StateCache) UpdateAccount(account *Account) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.account = account
    sc.lastUpdate = time.Now()
}

func (sc *StateCache) UpdateStrategy(id string, strategy *Strategy) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.strategies[id] = strategy
    sc.lastUpdate = time.Now()
}
```

### Broadcast Scheduling

```go
package server

import (
    "time"

    "github.com/vmihailenco/msgpack/v5"
)

func (s *Server) StartBroadcasters() {
    go s.tuiBroadcaster()
    go s.webBroadcaster()
}

func (s *Server) tuiBroadcaster() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    var sequence uint64

    for {
        select {
        case <-s.shutdownCtx.Done():
            return
        case <-ticker.C:
            sequence++
            s.broadcastToClients(ClientTypeTUI, sequence)
        }
    }
}

func (s *Server) webBroadcaster() {
    ticker := time.NewTicker(250 * time.Millisecond)
    defer ticker.Stop()

    var sequence uint64

    for {
        select {
        case <-s.shutdownCtx.Done():
            return
        case <-ticker.C:
            sequence++
            s.broadcastToClients(ClientTypeWeb, sequence)
        }
    }
}

func (s *Server) broadcastToClients(clientType ClientType, sequence uint64) {
    currentState := s.stateCache.GetFullState()
    currentState.Sequence = sequence

    s.clientsMu.RLock()
    defer s.clientsMu.RUnlock()

    for _, client := range s.clients {
        if client.Type != clientType {
            continue
        }

        // Check if client has subscriptions
        if len(client.Subscriptions) == 0 {
            continue
        }

        // Filter state based on subscriptions
        filteredState := s.filterStateBySubscriptions(currentState, client.Subscriptions)

        // Generate differential update if possible
        var message []byte
        var err error

        if client.LastState != nil {
            diff := s.computeDiff(client.LastState, filteredState)
            if len(diff) > 0 {
                msg := ServerMessage{
                    Type:      "update",
                    Sequence:  sequence,
                    Timestamp: time.Now(),
                    Changes:   diff,
                }
                message, err = s.serializeMessage(client.Format, msg)
            }
        } else {
            // Send full snapshot for first update
            msg := ServerMessage{
                Type:      "snapshot",
                Sequence:  sequence,
                Timestamp: time.Now(),
                Data:      filteredState,
            }
            message, err = s.serializeMessage(client.Format, msg)
        }

        if err != nil {
            s.logger.Error().Err(err).Str("client_id", client.ID).Msg("Failed to serialize message")
            continue
        }

        if message != nil {
            select {
            case client.SendChan <- message:
                client.LastState = filteredState
                client.LastUpdate = time.Now()
            default:
                s.logger.Warn().Str("client_id", client.ID).Msg("Client send buffer full, dropping message")
            }
        }
    }
}

func (s *Server) filterStateBySubscriptions(state *State, subscriptions map[string]bool) *State {
    filtered := &State{
        Timestamp: state.Timestamp,
        Sequence:  state.Sequence,
    }

    if subscriptions["market_data"] {
        filtered.MarketData = state.MarketData
    }
    if subscriptions["orders"] {
        filtered.Orders = state.Orders
    }
    if subscriptions["positions"] {
        filtered.Positions = state.Positions
    }
    if subscriptions["account"] {
        filtered.Account = state.Account
    }
    if subscriptions["strategies"] {
        filtered.Strategies = state.Strategies
    }

    return filtered
}

func (s *Server) computeDiff(oldState, newState *State) map[string]interface{} {
    diff := make(map[string]interface{})

    // Compare market data
    if oldState.MarketData != nil && newState.MarketData != nil {
        for symbol, newMD := range newState.MarketData {
            oldMD, exists := oldState.MarketData[symbol]
            if !exists || oldMD.LastPrice != newMD.LastPrice {
                diff["market_data."+symbol+".last_price"] = newMD.LastPrice
            }
            if !exists || oldMD.BidPrice != newMD.BidPrice {
                diff["market_data."+symbol+".bid_price"] = newMD.BidPrice
            }
            if !exists || oldMD.AskPrice != newMD.AskPrice {
                diff["market_data."+symbol+".ask_price"] = newMD.AskPrice
            }
        }
    }

    // Compare positions
    if oldState.Positions != nil && newState.Positions != nil {
        for symbol, newPos := range newState.Positions {
            oldPos, exists := oldState.Positions[symbol]
            if !exists || oldPos.UnrealizedPnL != newPos.UnrealizedPnL {
                diff["positions."+symbol+".unrealized_pnl"] = newPos.UnrealizedPnL
            }
            if !exists || oldPos.MarkPrice != newPos.MarkPrice {
                diff["positions."+symbol+".mark_price"] = newPos.MarkPrice
            }
        }
    }

    // Compare account
    if oldState.Account != nil && newState.Account != nil {
        if oldState.Account.TotalBalance != newState.Account.TotalBalance {
            diff["account.total_balance"] = newState.Account.TotalBalance
        }
        if oldState.Account.UnrealizedPnL != newState.Account.UnrealizedPnL {
            diff["account.unrealized_pnl"] = newState.Account.UnrealizedPnL
        }
    }

    // TODO: Add more field comparisons as needed

    return diff
}

func (s *Server) serializeMessage(format SerializationFormat, msg ServerMessage) ([]byte, error) {
    switch format {
    case FormatMessagePack:
        return msgpack.Marshal(msg)
    case FormatJSON:
        return json.Marshal(msg)
    default:
        return nil, fmt.Errorf("unsupported serialization format: %v", format)
    }
}

type ServerMessage struct {
    Type      string                 `msgpack:"type" json:"type"`
    Sequence  uint64                 `msgpack:"seq" json:"seq"`
    Timestamp time.Time              `msgpack:"timestamp" json:"timestamp"`
    Data      *State                 `msgpack:"data,omitempty" json:"data,omitempty"`
    Changes   map[string]interface{} `msgpack:"changes,omitempty" json:"changes,omitempty"`
}

func (s *Server) sendFullState(client *Client, state *State) {
    msg := ServerMessage{
        Type:      "snapshot",
        Timestamp: time.Now(),
        Data:      state,
    }

    message, err := s.serializeMessage(client.Format, msg)
    if err != nil {
        s.logger.Error().Err(err).Str("client_id", client.ID).Msg("Failed to serialize full state")
        return
    }

    select {
    case client.SendChan <- message:
        client.LastState = state
    default:
        s.logger.Warn().Str("client_id", client.ID).Msg("Client send buffer full")
    }
}
```

### Main Server Setup

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "your-project/server"
)

func main() {
    // Setup logger
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    log.Logger = logger

    // Create server
    srv := server.NewServer(logger)

    // Start background tasks
    srv.StartBroadcasters()

    // Setup HTTP routes
    mux := http.NewServeMux()
    mux.HandleFunc("/ws", srv.HandleWebSocket)
    mux.HandleFunc("/health", handleHealth)
    mux.Handle("/metrics", promhttp.Handler())

    // Start HTTP server
    httpServer := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    // Graceful shutdown
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
        <-sigCh

        logger.Info().Msg("Shutting down server...")

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := httpServer.Shutdown(ctx); err != nil {
            logger.Error().Err(err).Msg("Server shutdown error")
        }
    }()

    logger.Info().Str("addr", httpServer.Addr).Msg("Starting Dashboard Server")
    if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
        logger.Fatal().Err(err).Msg("Server failed")
    }
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}
```

---

## Testing Strategy

### Unit Tests

```go
package server

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestStateCache_UpdateMarketData(t *testing.T) {
    cache := NewStateCache()

    md := &MarketData{
        Symbol:    "BTCUSDT",
        LastPrice: 50000.0,
        BidPrice:  49999.0,
        AskPrice:  50001.0,
    }

    cache.UpdateMarketData("BTCUSDT", md)

    state := cache.GetFullState()
    assert.Equal(t, 1, len(state.MarketData))
    assert.Equal(t, 50000.0, state.MarketData["BTCUSDT"].LastPrice)
}

func TestComputeDiff(t *testing.T) {
    s := &Server{}

    oldState := &State{
        MarketData: map[string]*MarketData{
            "BTCUSDT": {Symbol: "BTCUSDT", LastPrice: 50000.0},
        },
    }

    newState := &State{
        MarketData: map[string]*MarketData{
            "BTCUSDT": {Symbol: "BTCUSDT", LastPrice: 51000.0},
        },
    }

    diff := s.computeDiff(oldState, newState)
    assert.Contains(t, diff, "market_data.BTCUSDT.last_price")
    assert.Equal(t, 51000.0, diff["market_data.BTCUSDT.last_price"])
}
```

### WebSocket Integration Tests

```go
package server

import (
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gorilla/websocket"
    "github.com/rs/zerolog"
    "github.com/stretchr/testify/assert"
)

func TestWebSocketConnection(t *testing.T) {
    logger := zerolog.Nop()
    srv := NewServer(logger)

    // Create test server
    server := httptest.NewServer(http.HandlerFunc(srv.HandleWebSocket))
    defer server.Close()

    // Connect WebSocket client
    wsURL := "ws" + server.URL[4:] + "?type=tui"
    ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    assert.NoError(t, err)
    defer ws.Close()

    // Send subscribe message
    subscribeMsg := `{"type":"subscribe","channels":["market_data","orders"]}`
    err = ws.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
    assert.NoError(t, err)

    // Wait for response
    time.Sleep(100 * time.Millisecond)

    // Verify client is registered
    srv.clientsMu.RLock()
    assert.Equal(t, 1, len(srv.clients))
    srv.clientsMu.RUnlock()
}
```

### Load Testing

```go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

func main() {
    numClients := 100
    var wg sync.WaitGroup

    for i := 0; i < numClients; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            runClient(id)
        }(i)
    }

    wg.Wait()
}

func runClient(id int) {
    url := "ws://localhost:8080/ws?type=tui"
    ws, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        fmt.Printf("Client %d: Failed to connect: %v\n", id, err)
        return
    }
    defer ws.Close()

    // Subscribe
    subscribeMsg := `{"type":"subscribe","channels":["market_data","orders","positions","account"]}`
    ws.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))

    // Receive messages for 60 seconds
    messageCount := 0
    timeout := time.After(60 * time.Second)

    for {
        select {
        case <-timeout:
            fmt.Printf("Client %d: Received %d messages\n", id, messageCount)
            return
        default:
            ws.SetReadDeadline(time.Now().Add(2 * time.Second))
            _, _, err := ws.ReadMessage()
            if err == nil {
                messageCount++
            }
        }
    }
}
```

### Serialization Benchmarks

```go
package server

import (
    "encoding/json"
    "testing"

    "github.com/vmihailenco/msgpack/v5"
)

func BenchmarkMessagePackSerialization(b *testing.B) {
    state := createLargeState()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := msgpack.Marshal(state)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkJSONSerialization(b *testing.B) {
    state := createLargeState()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := json.Marshal(state)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func createLargeState() *State {
    state := &State{
        MarketData: make(map[string]*MarketData),
        Orders:     make([]*Order, 0),
        Positions:  make(map[string]*Position),
        Account:    &Account{Balances: make(map[string]float64)},
        Strategies: make(map[string]*Strategy),
    }

    // Add 50 symbols
    for i := 0; i < 50; i++ {
        symbol := fmt.Sprintf("SYMBOL%d", i)
        state.MarketData[symbol] = &MarketData{
            Symbol:    symbol,
            LastPrice: 100.0 + float64(i),
            BidPrice:  99.5 + float64(i),
            AskPrice:  100.5 + float64(i),
        }
    }

    // Add 20 orders
    for i := 0; i < 20; i++ {
        state.Orders = append(state.Orders, &Order{
            ID:       fmt.Sprintf("order-%d", i),
            Symbol:   "BTCUSDT",
            Side:     "BUY",
            Quantity: 0.1,
            Price:    50000.0,
        })
    }

    return state
}
```

---

## Deployment

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashboard-server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/dashboard-server .

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./dashboard-server"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  dashboard-server:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
      - MARKET_DATA_SERVICE_URL=market-data:9090
      - ORDER_EXECUTION_SERVICE_URL=order-execution:9091
      - ACCOUNT_MONITOR_SERVICE_URL=account-monitor:9093
      - REDIS_URL=redis:6379
    networks:
      - trading-net
    depends_on:
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - trading-net

networks:
  trading-net:
    driver: bridge
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard-server
  labels:
    app: dashboard-server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: dashboard-server
  template:
    metadata:
      labels:
        app: dashboard-server
    spec:
      containers:
      - name: dashboard-server
        image: your-registry/dashboard-server:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: dashboard-server
spec:
  selector:
    app: dashboard-server
  ports:
  - port: 8080
    targetPort: 8080
  type: LoadBalancer
```

### Load Balancing Considerations

**WebSocket Load Balancing:**

For horizontal scaling with multiple Dashboard Server instances, configure load balancer for WebSocket sticky sessions:

**Nginx Configuration:**
```nginx
upstream dashboard_servers {
    ip_hash;  # Sticky sessions based on client IP
    server dashboard-server-1:8080;
    server dashboard-server-2:8080;
}

server {
    listen 80;

    location /ws {
        proxy_pass http://dashboard_servers;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    location /health {
        proxy_pass http://dashboard_servers;
    }

    location /metrics {
        proxy_pass http://dashboard_servers;
    }
}
```

**Alternative: Use Redis for Shared State**

If clients can connect to any server instance:
- Store client subscription state in Redis
- Broadcast updates via Redis Pub/Sub to all server instances
- Each server forwards to its connected clients

---

## Observability

### Prometheus Metrics

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // WebSocket connection metrics
    ConnectedClients = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "dashboard_connected_clients",
            Help: "Number of connected WebSocket clients",
        },
        []string{"client_type"},
    )

    MessagesSent = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "dashboard_messages_sent_total",
            Help: "Total number of messages sent to clients",
        },
        []string{"client_type", "message_type"},
    )

    MessagesReceived = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "dashboard_messages_received_total",
            Help: "Total number of messages received from clients",
        },
        []string{"message_type"},
    )

    BroadcastLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "dashboard_broadcast_latency_seconds",
            Help:    "Broadcast latency in seconds",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
        },
        []string{"client_type"},
    )

    SerializationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "dashboard_serialization_duration_seconds",
            Help:    "Serialization duration in seconds",
            Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10),
        },
        []string{"format"},
    )

    MessageSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "dashboard_message_size_bytes",
            Help:    "Size of messages in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 2, 10),
        },
        []string{"format", "message_type"},
    )

    StateUpdateLag = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "dashboard_state_update_lag_seconds",
            Help: "Lag between state source update and cache update",
        },
        []string{"source"},
    )

    ClientSubscriptions = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "dashboard_client_subscriptions",
            Help: "Number of active subscriptions per channel",
        },
        []string{"channel"},
    )
)

// Helper functions to update metrics
func IncrementConnectedClients(clientType string) {
    ConnectedClients.WithLabelValues(clientType).Inc()
}

func DecrementConnectedClients(clientType string) {
    ConnectedClients.WithLabelValues(clientType).Dec()
}

func RecordMessageSent(clientType, messageType string) {
    MessagesSent.WithLabelValues(clientType, messageType).Inc()
}

func RecordBroadcastLatency(clientType string, duration float64) {
    BroadcastLatency.WithLabelValues(clientType).Observe(duration)
}

func RecordSerializationDuration(format string, duration float64) {
    SerializationDuration.WithLabelValues(format).Observe(duration)
}

func RecordMessageSize(format, messageType string, size int) {
    MessageSize.WithLabelValues(format, messageType).Observe(float64(size))
}
```

### Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "Dashboard Server Service",
    "panels": [
      {
        "title": "Connected Clients",
        "targets": [
          {
            "expr": "sum(dashboard_connected_clients) by (client_type)"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Broadcast Latency (p99)",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(dashboard_broadcast_latency_seconds_bucket[5m]))"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Messages Sent Rate",
        "targets": [
          {
            "expr": "rate(dashboard_messages_sent_total[1m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Message Size Distribution",
        "targets": [
          {
            "expr": "rate(dashboard_message_size_bytes_bucket[5m])"
          }
        ],
        "type": "heatmap"
      },
      {
        "title": "Serialization Performance",
        "targets": [
          {
            "expr": "rate(dashboard_serialization_duration_seconds_sum[1m]) / rate(dashboard_serialization_duration_seconds_count[1m])"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

### Structured Logging Example

```go
// In broadcast function
func (s *Server) broadcastToClients(clientType ClientType, sequence uint64) {
    start := time.Now()

    s.logger.Debug().
        Str("client_type", clientType.String()).
        Uint64("sequence", sequence).
        Msg("Starting broadcast")

    // ... broadcast logic ...

    duration := time.Since(start)

    s.logger.Info().
        Str("client_type", clientType.String()).
        Uint64("sequence", sequence).
        Dur("duration", duration).
        Int("clients_notified", notifiedCount).
        Msg("Broadcast completed")

    metrics.RecordBroadcastLatency(clientType.String(), duration.Seconds())
}
```

---

## Code Examples

### Complete Client Example (Go)

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/gorilla/websocket"
    "github.com/vmihailenco/msgpack/v5"
)

type State struct {
    MarketData map[string]*MarketData `msgpack:"market_data"`
    Orders     []*Order               `msgpack:"orders"`
    Positions  map[string]*Position   `msgpack:"positions"`
    Account    *Account               `msgpack:"account"`
}

type MarketData struct {
    Symbol    string  `msgpack:"symbol"`
    LastPrice float64 `msgpack:"last_price"`
}

type Order struct {
    ID     string `msgpack:"id"`
    Symbol string `msgpack:"symbol"`
    Status string `msgpack:"status"`
}

type Position struct {
    Symbol        string  `msgpack:"symbol"`
    UnrealizedPnL float64 `msgpack:"unrealized_pnl"`
}

type Account struct {
    TotalBalance float64 `msgpack:"total_balance"`
}

type ServerMessage struct {
    Type      string                 `msgpack:"type"`
    Sequence  uint64                 `msgpack:"seq"`
    Timestamp time.Time              `msgpack:"timestamp"`
    Data      *State                 `msgpack:"data,omitempty"`
    Changes   map[string]interface{} `msgpack:"changes,omitempty"`
}

func main() {
    // Connect to Dashboard Server
    url := "ws://localhost:8080/ws?type=tui&format=msgpack"
    ws, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer ws.Close()

    // Subscribe to all channels
    subscribeMsg := `{"type":"subscribe","channels":["market_data","orders","positions","account"]}`
    if err := ws.WriteMessage(websocket.TextMessage, []byte(subscribeMsg)); err != nil {
        log.Fatal("Failed to subscribe:", err)
    }

    fmt.Println("Connected and subscribed. Receiving updates...")

    // Read messages
    for {
        _, message, err := ws.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            return
        }

        var serverMsg ServerMessage
        if err := msgpack.Unmarshal(message, &serverMsg); err != nil {
            log.Println("Unmarshal error:", err)
            continue
        }

        switch serverMsg.Type {
        case "snapshot":
            fmt.Printf("[SNAPSHOT] Seq: %d\n", serverMsg.Sequence)
            if serverMsg.Data != nil && serverMsg.Data.Account != nil {
                fmt.Printf("  Account Balance: $%.2f\n", serverMsg.Data.Account.TotalBalance)
            }

        case "update":
            fmt.Printf("[UPDATE] Seq: %d, Changes: %d\n", serverMsg.Sequence, len(serverMsg.Changes))
            for key, value := range serverMsg.Changes {
                fmt.Printf("  %s = %v\n", key, value)
            }
        }
    }
}
```

### JavaScript/TypeScript Client Example

```typescript
import * as msgpack from '@msgpack/msgpack';

interface ServerMessage {
  type: 'snapshot' | 'update' | 'error';
  seq?: number;
  timestamp?: string;
  data?: State;
  changes?: Record<string, any>;
}

interface State {
  market_data?: Record<string, MarketData>;
  orders?: Order[];
  positions?: Record<string, Position>;
  account?: Account;
}

interface MarketData {
  symbol: string;
  last_price: number;
  bid_price: number;
  ask_price: number;
}

interface Order {
  id: string;
  symbol: string;
  side: string;
  quantity: number;
  status: string;
}

interface Position {
  symbol: string;
  unrealized_pnl: number;
}

interface Account {
  total_balance: number;
}

class DashboardClient {
  private ws: WebSocket | null = null;
  private currentState: State = {};
  private onStateUpdate?: (state: State) => void;

  connect(url: string, clientType: 'tui' | 'web' = 'web') {
    const wsUrl = `${url}?type=${clientType}&format=msgpack`;
    this.ws = new WebSocket(wsUrl);
    this.ws.binaryType = 'arraybuffer';

    this.ws.onopen = () => {
      console.log('Connected to Dashboard Server');
      this.subscribe(['market_data', 'orders', 'positions', 'account']);
    };

    this.ws.onmessage = async (event) => {
      const buffer = new Uint8Array(event.data);
      const message = msgpack.decode(buffer) as ServerMessage;
      this.handleMessage(message);
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('Disconnected from Dashboard Server');
    };
  }

  subscribe(channels: string[]) {
    const message = {
      type: 'subscribe',
      channels: channels,
    };
    this.ws?.send(JSON.stringify(message));
  }

  private handleMessage(message: ServerMessage) {
    switch (message.type) {
      case 'snapshot':
        console.log(`[SNAPSHOT] Seq: ${message.seq}`);
        if (message.data) {
          this.currentState = message.data;
          this.notifyUpdate();
        }
        break;

      case 'update':
        console.log(`[UPDATE] Seq: ${message.seq}, Changes: ${Object.keys(message.changes || {}).length}`);
        if (message.changes) {
          this.applyChanges(message.changes);
          this.notifyUpdate();
        }
        break;

      case 'error':
        console.error('Server error:', message);
        break;
    }
  }

  private applyChanges(changes: Record<string, any>) {
    for (const [path, value] of Object.entries(changes)) {
      this.setNestedValue(this.currentState, path, value);
    }
  }

  private setNestedValue(obj: any, path: string, value: any) {
    const keys = path.split('.');
    let current = obj;

    for (let i = 0; i < keys.length - 1; i++) {
      const key = keys[i];
      if (!current[key]) {
        current[key] = {};
      }
      current = current[key];
    }

    current[keys[keys.length - 1]] = value;
  }

  private notifyUpdate() {
    if (this.onStateUpdate) {
      this.onStateUpdate(this.currentState);
    }
  }

  onUpdate(callback: (state: State) => void) {
    this.onStateUpdate = callback;
  }

  disconnect() {
    this.ws?.close();
  }
}

// Usage
const client = new DashboardClient();
client.connect('ws://localhost:8080/ws', 'web');

client.onUpdate((state) => {
  console.log('State updated:', state);

  // Update UI with new state
  if (state.account) {
    document.getElementById('balance').textContent = `$${state.account.total_balance.toFixed(2)}`;
  }

  if (state.positions) {
    const totalPnl = Object.values(state.positions).reduce(
      (sum, pos) => sum + pos.unrealized_pnl,
      0
    );
    document.getElementById('pnl').textContent = `$${totalPnl.toFixed(2)}`;
  }
});
```

---

## Summary

This development plan provides a complete roadmap for building the Dashboard Server Service with:

1. **Go-based WebSocket server** with excellent concurrency support
2. **MessagePack serialization** for 3-5x bandwidth reduction
3. **Differential updates** to minimize data transfer
4. **Rate-differentiated broadcasting** (TUI: 100ms, Web: 250ms)
5. **Comprehensive testing strategy** including load testing for 100+ clients
6. **Production-ready observability** with Prometheus metrics and Grafana dashboards
7. **Containerized deployment** with Docker and Kubernetes support
8. **Client examples** in both Go and TypeScript

The service is designed to be:
- **Scalable**: Handles 100+ concurrent connections with low resource usage
- **Efficient**: Optimized serialization and differential updates
- **Reliable**: Graceful error handling and automatic reconnection
- **Observable**: Rich metrics and structured logging
- **Maintainable**: Clean architecture with separation of concerns

Total estimated development time: **6 weeks** for full implementation and testing.

---

**Next Steps:**

1. Review and approve technology choices
2. Set up development environment and repository structure
3. Begin Phase 1: WebSocket Server Setup
4. Iterate through development phases with regular testing
5. Deploy to staging environment for integration testing
6. Performance tuning and optimization
7. Production deployment with monitoring

**Dependencies:**

- Backend services must expose gRPC/REST APIs
- Pub/Sub infrastructure (Redis/NATS) must be available
- Prometheus and Grafana for observability

**Success Metrics:**

- [ ] Support 100+ concurrent WebSocket connections
- [ ] Broadcast latency p99 < 50ms
- [ ] Memory usage < 10MB per connection
- [ ] MessagePack payloads 3-5x smaller than JSON
- [ ] 99.9% uptime in production
- [ ] Zero message loss during normal operation
- [ ] Graceful degradation when backend services fail

---

**End of Document**
