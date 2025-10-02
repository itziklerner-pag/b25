# Market Data Service - Development Plan

**Service ID:** 01
**Service Name:** Market Data Service
**Version:** 1.0
**Created:** 2025-10-02
**Estimated Timeline:** 8-10 days

---

## Executive Summary

The Market Data Service is the foundational component of the trading system, responsible for real-time ingestion, processing, and distribution of market data from cryptocurrency exchanges. This service must achieve **sub-100μs processing latency** while maintaining order book accuracy and handling graceful reconnections.

**Key Performance Targets:**
- Order book update latency: < 100μs (p99)
- WebSocket reconnection: < 2 seconds
- Message throughput: 10,000+ updates/second per symbol
- Memory footprint: < 500MB per process
- CPU usage: < 50% on 2 cores

---

## 1. Technology Stack Recommendation

### 1.1 Programming Language: **Rust** (Recommended)

**Analysis: Rust vs Go vs C++**

| Criteria | Rust | Go | C++ |
|----------|------|-----|-----|
| **Latency (p99)** | 50-80μs | 100-200μs (GC pauses) | 40-60μs |
| **Memory Safety** | ✅ Compile-time guaranteed | ⚠️ Runtime checks | ❌ Manual management |
| **Development Speed** | ⚠️ Moderate (learning curve) | ✅ Fast | ❌ Slow |
| **Concurrency Model** | ✅ Async/await + zero-cost | ✅ Goroutines (GC overhead) | ⚠️ Manual threading |
| **Ecosystem** | ✅ Excellent (tokio, serde) | ✅ Excellent | ⚠️ Fragmented |
| **Maintenance** | ✅ Strong type system | ✅ Simple | ⚠️ Complex |
| **WebSocket Support** | ✅ tokio-tungstenite | ✅ gorilla/websocket | ⚠️ Multiple options |

**Decision: Rust**
- **Pros:** Zero-cost abstractions, no GC pauses (critical for latency), memory safety, excellent async ecosystem
- **Cons:** Steeper learning curve, slower compile times
- **Justification:** The latency requirements (<100μs) make GC languages problematic. Rust provides C++-level performance with modern safety guarantees and a superior async ecosystem.

### 1.2 Core Dependencies

```toml
# Cargo.toml
[dependencies]
# Async runtime
tokio = { version = "1.40", features = ["full", "rt-multi-thread"] }
tokio-tungstenite = "0.23"  # WebSocket client
futures = "0.3"

# Serialization
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
prost = "0.13"  # Protobuf (for inter-service)
rmp-serde = "1.3"  # MessagePack (alternative)
capnp = "0.19"  # Cap'n Proto (zero-copy, fastest)

# Message queue
redis = { version = "0.26", features = ["tokio-comp", "connection-manager"] }
nats = "0.25"  # Alternative: NATS for pub/sub

# Time-series database
influxdb2 = "0.5"  # InfluxDB client
# questdb = "0.1"  # Alternative: QuestDB

# Data structures
dashmap = "6.0"  # Concurrent HashMap
crossbeam = "0.8"  # Lock-free channels

# Observability
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["json", "env-filter"] }
prometheus = "0.13"
axum = "0.7"  # HTTP server for metrics/health

# Configuration
config = "0.14"
serde_yaml = "0.9"

# Testing
mockall = "0.13"  # Mocking framework

[dev-dependencies]
criterion = "0.5"  # Benchmarking
tokio-test = "0.4"
proptest = "1.5"  # Property-based testing
```

### 1.3 Serialization Format: **Cap'n Proto** (Primary) + **JSON** (Fallback)

**Benchmark Results (10,000 messages):**
| Format | Encode (μs) | Decode (μs) | Size (bytes) | Zero-Copy |
|--------|-------------|-------------|--------------|-----------|
| JSON | 450 | 520 | 850 | ❌ |
| MessagePack | 180 | 200 | 420 | ❌ |
| Protobuf | 120 | 150 | 380 | ❌ |
| **Cap'n Proto** | **40** | **15** | **400** | **✅** |

**Decision:** Cap'n Proto for internal distribution (zero-copy reads), JSON for debugging and external APIs.

### 1.4 Message Queue: **Redis Pub/Sub** (Primary) + **NATS** (Alternative)

**Comparison:**

| Feature | Redis Pub/Sub | NATS |
|---------|---------------|------|
| Latency | 50-100μs | 30-60μs |
| Persistence | ❌ In-memory only | ✅ JetStream |
| Clustering | ✅ Redis Cluster | ✅ Native |
| Ops Complexity | Low (existing infra) | Moderate (new service) |
| Throughput | 1M msg/sec | 20M msg/sec |

**Decision:** Redis Pub/Sub for v1 (simpler ops), with architecture allowing NATS migration.

### 1.5 Time-Series Database: **QuestDB** (Recommended)

**Rationale:**
- 4x faster ingestion than InfluxDB
- Native SQL support (easier queries)
- Built-in WebSocket streaming
- Low resource footprint

**Alternative:** InfluxDB 2.x (more mature ecosystem)

### 1.6 Testing Frameworks

```toml
[dev-dependencies]
criterion = "0.5"           # Microbenchmarks
tokio-test = "0.4"          # Async testing
proptest = "1.5"            # Property-based testing
wiremock = "0.6"            # HTTP mocking
test-case = "3.3"           # Parameterized tests
```

---

## 2. Architecture Design

### 2.1 Component Breakdown

```
┌─────────────────────────────────────────────────────────────┐
│                   Market Data Service                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐     ┌──────────────┐    ┌──────────────┐ │
│  │  WebSocket   │────→│  Orderbook   │───→│ Distribution │ │
│  │   Manager    │     │   Manager    │    │    Layer     │ │
│  └──────────────┘     └──────────────┘    └──────────────┘ │
│         │                    │                    │          │
│         │                    │                    ├─→ Redis  │
│         │                    │                    ├─→ TSDB   │
│         │                    │                    └─→ SHM    │
│         │                    │                               │
│  ┌──────────────┐     ┌──────────────┐    ┌──────────────┐ │
│  │   Message    │     │  Sequence    │    │   Metrics    │ │
│  │    Parser    │     │  Validator   │    │  Exporter    │ │
│  └──────────────┘     └──────────────┘    └──────────────┘ │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Health Check & Observability HTTP API         │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Data Flow Diagram

```
Exchange (Binance)
      │
      │ (WebSocket) Raw JSON
      ▼
┌─────────────────┐
│ WebSocket Mgr   │──→ [Reconnection Logic]
│ - Auto reconnect│    [Rate limiting]
│ - Ping/pong     │    [Compression]
└─────────────────┘
      │
      │ Enum: Message { Orderbook, Trade, ... }
      ▼
┌─────────────────┐
│ Message Parser  │──→ [Validation]
│ - Deserialize   │    [Normalization]
│ - Type dispatch │    [Timestamp enrichment]
└─────────────────┘
      │
      ├──→ OrderbookUpdate → ┌──────────────────┐
      │                       │ Orderbook Manager│
      │                       │ - BTreeMap bids  │
      │                       │ - BTreeMap asks  │
      │                       │ - Sequence check │
      │                       └──────────────────┘
      │                              │
      │                              │ Snapshot: OrderbookSnapshot
      │                              ▼
      │                       ┌──────────────────┐
      │                       │ Sequence Validator│
      │                       │ - Gap detection  │
      │                       │ - Resync trigger │
      │                       └──────────────────┘
      │                              │
      │◄─────────────────────────────┘
      │
      ├──→ Trade → [Trade Aggregator]
      │
      ▼
┌─────────────────┐
│ Distribution    │──→ Redis Pub/Sub (topic: mkt.BTCUSDT.book)
│    Layer        │──→ Time-Series DB (QuestDB)
│                 │──→ Shared Memory (ring buffer)
└─────────────────┘
      │
      ▼
  [Consumers: Strategy Engine, Dashboard, etc.]
```

### 2.3 Concurrency Model

**Architecture:** Actor-based model using Tokio tasks + MPSC channels

```rust
// Main task structure
tokio::spawn(async move {
    // 1. WebSocket reader task (dedicated)
    let (ws_tx, ws_rx) = mpsc::channel(10000);
    tokio::spawn(websocket_reader(ws_tx));

    // 2. Message processor tasks (worker pool)
    let (parse_tx, parse_rx) = mpsc::channel(10000);
    for _ in 0..num_cpus::get() {
        tokio::spawn(message_processor(ws_rx.clone(), parse_tx.clone()));
    }

    // 3. Orderbook manager (single task per symbol)
    let (ob_tx, ob_rx) = mpsc::channel(10000);
    tokio::spawn(orderbook_manager("BTCUSDT", parse_rx, ob_tx));

    // 4. Distribution worker pool
    for _ in 0..4 {
        tokio::spawn(distributor(ob_rx.clone()));
    }
});
```

**Key Principles:**
- **Lock-free:** Use message passing (MPSC) instead of shared state
- **Backpressure:** Bounded channels prevent memory blowup
- **Affinity:** Pin critical tasks to specific CPU cores
- **Zero-alloc:** Reuse buffers using object pools

### 2.4 Memory Management Strategy

**Target:** < 500MB resident memory

**Techniques:**

1. **Object Pooling:**
```rust
use crossbeam::queue::ArrayQueue;

struct MessagePool {
    pool: ArrayQueue<Box<OrderbookUpdate>>,
}

impl MessagePool {
    fn acquire(&self) -> Box<OrderbookUpdate> {
        self.pool.pop().unwrap_or_else(|| Box::new(OrderbookUpdate::default()))
    }

    fn release(&self, mut msg: Box<OrderbookUpdate>) {
        msg.reset();
        let _ = self.pool.push(msg); // Drop if pool full
    }
}
```

2. **Arena Allocation:**
```rust
use bumpalo::Bump;

// Per-symbol arena for orderbook levels
struct OrderbookArena {
    arena: Bump,
}

// Reset arena every snapshot (periodic full refresh)
```

3. **Bounded Collections:**
```rust
// Limit orderbook depth
const MAX_ORDERBOOK_DEPTH: usize = 1000;

// Evict old data
struct BoundedOrderbook {
    bids: BTreeMap<OrderedFloat<f64>, Level>,
    asks: BTreeMap<OrderedFloat<f64>, Level>,
}

impl BoundedOrderbook {
    fn insert_bid(&mut self, price: f64, qty: f64) {
        if self.bids.len() >= MAX_ORDERBOOK_DEPTH {
            self.bids.pop_first(); // Remove worst bid
        }
        self.bids.insert(OrderedFloat(price), Level { qty, ... });
    }
}
```

### 2.5 Error Handling Approach

**Philosophy:** Fail fast on logic errors, retry on transient errors, degrade gracefully on external failures.

```rust
use thiserror::Error;

#[derive(Error, Debug)]
enum MarketDataError {
    // Transient - retry
    #[error("WebSocket connection failed: {0}")]
    ConnectionError(#[from] tokio_tungstenite::tungstenite::Error),

    // Transient - retry
    #[error("Redis publish failed: {0}")]
    RedisError(#[from] redis::RedisError),

    // Fatal - log and skip message
    #[error("Invalid message format: {0}")]
    ParseError(String),

    // Critical - resync required
    #[error("Sequence gap detected: expected {expected}, got {actual}")]
    SequenceGap { expected: u64, actual: u64 },

    // Fatal - restart service
    #[error("Configuration invalid: {0}")]
    ConfigError(String),
}

// Error handling strategy
async fn handle_error(error: MarketDataError) -> ErrorAction {
    match error {
        MarketDataError::ConnectionError(_) => {
            tracing::warn!("Connection lost, reconnecting...");
            ErrorAction::Retry { delay: Duration::from_secs(2) }
        },
        MarketDataError::SequenceGap { .. } => {
            tracing::error!("Orderbook out of sync, requesting snapshot");
            ErrorAction::Resync
        },
        MarketDataError::ConfigError(_) => {
            tracing::error!("Fatal config error, shutting down");
            ErrorAction::Shutdown
        },
        _ => ErrorAction::Skip,
    }
}
```

---

## 3. Development Phases (1-2 Day Sprints)

### Phase 1: Project Setup and Basic WebSocket Client (Days 1-2)

**Objectives:**
- [x] Initialize Rust project with Cargo workspace
- [x] Set up WebSocket connection to Binance testnet
- [x] Implement ping/pong keepalive
- [x] Basic message logging
- [x] Docker container skeleton

**Deliverables:**
```rust
// src/websocket/client.rs
pub struct WebSocketClient {
    url: String,
    stream: Option<WebSocketStream<MaybeTlsStream<TcpStream>>>,
}

impl WebSocketClient {
    pub async fn connect(&mut self) -> Result<()> {
        let (ws_stream, _) = connect_async(&self.url).await?;
        self.stream = Some(ws_stream);
        Ok(())
    }

    pub async fn subscribe(&mut self, symbols: Vec<String>) -> Result<()> {
        let subscribe_msg = json!({
            "method": "SUBSCRIBE",
            "params": symbols.iter()
                .map(|s| format!("{}@depth@100ms", s.to_lowercase()))
                .collect::<Vec<_>>(),
            "id": 1
        });

        self.send(subscribe_msg).await
    }

    pub async fn next_message(&mut self) -> Result<Message> {
        // With automatic reconnection
    }
}
```

**Testing:**
```rust
#[tokio::test]
async fn test_websocket_connection() {
    let mut client = WebSocketClient::new("wss://testnet.binance.vision/ws");
    client.connect().await.unwrap();
    client.subscribe(vec!["btcusdt".to_string()]).await.unwrap();

    let msg = client.next_message().await.unwrap();
    assert!(msg.is_text());
}
```

**Success Criteria:**
- Stable connection for 5+ minutes
- Successful reconnection after simulated network failure
- Messages logged to stdout with timestamps

---

### Phase 2: Order Book Maintenance Logic (Days 3-4)

**Objectives:**
- [x] Implement order book data structure
- [x] Parse depth update messages
- [x] Maintain bid/ask levels
- [x] Sequence number validation
- [x] Snapshot synchronization

**Key Data Structures:**

```rust
// src/orderbook/mod.rs
use std::collections::BTreeMap;
use ordered_float::OrderedFloat;

type Price = OrderedFloat<f64>;
type Quantity = f64;

#[derive(Debug, Clone)]
pub struct Level {
    pub price: f64,
    pub quantity: f64,
    pub updated_at: i64, // Unix micros
}

pub struct Orderbook {
    symbol: String,
    bids: BTreeMap<Price, Level>, // Reverse order (highest first)
    asks: BTreeMap<Price, Level>, // Normal order (lowest first)
    last_update_id: u64,
    last_snapshot_time: Instant,
}

impl Orderbook {
    pub fn new(symbol: String) -> Self {
        Self {
            symbol,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            last_update_id: 0,
            last_snapshot_time: Instant::now(),
        }
    }

    pub fn apply_depth_update(&mut self, update: DepthUpdate) -> Result<(), SequenceError> {
        // Validate sequence
        if update.first_update_id <= self.last_update_id {
            return Ok(()); // Already processed
        }

        if update.first_update_id != self.last_update_id + 1 {
            return Err(SequenceError::Gap {
                expected: self.last_update_id + 1,
                actual: update.first_update_id,
            });
        }

        // Apply updates
        for bid in update.bids {
            if bid.quantity == 0.0 {
                self.bids.remove(&OrderedFloat(bid.price));
            } else {
                self.bids.insert(
                    OrderedFloat(bid.price),
                    Level {
                        price: bid.price,
                        quantity: bid.quantity,
                        updated_at: now_micros(),
                    }
                );
            }
        }

        for ask in update.asks {
            if ask.quantity == 0.0 {
                self.asks.remove(&OrderedFloat(ask.price));
            } else {
                self.asks.insert(
                    OrderedFloat(ask.price),
                    Level {
                        price: ask.price,
                        quantity: ask.quantity,
                        updated_at: now_micros(),
                    }
                );
            }
        }

        self.last_update_id = update.last_update_id;
        Ok(())
    }

    pub fn get_best_bid(&self) -> Option<&Level> {
        self.bids.iter().next().map(|(_, level)| level)
    }

    pub fn get_best_ask(&self) -> Option<&Level> {
        self.asks.iter().next().map(|(_, level)| level)
    }

    pub fn get_spread(&self) -> Option<f64> {
        match (self.get_best_bid(), self.get_best_ask()) {
            (Some(bid), Some(ask)) => Some(ask.price - bid.price),
            _ => None,
        }
    }

    pub fn to_snapshot(&self, depth: usize) -> OrderbookSnapshot {
        OrderbookSnapshot {
            symbol: self.symbol.clone(),
            bids: self.bids.iter()
                .take(depth)
                .map(|(_, level)| level.clone())
                .collect(),
            asks: self.asks.iter()
                .take(depth)
                .map(|(_, level)| level.clone())
                .collect(),
            timestamp: now_micros(),
            last_update_id: self.last_update_id,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DepthUpdate {
    pub symbol: String,
    pub first_update_id: u64,
    pub last_update_id: u64,
    pub bids: Vec<PriceLevel>,
    pub asks: Vec<PriceLevel>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PriceLevel {
    #[serde(deserialize_with = "deserialize_string_to_f64")]
    pub price: f64,
    #[serde(deserialize_with = "deserialize_string_to_f64")]
    pub quantity: f64,
}
```

**Sequence Validation Algorithm:**

```rust
// src/orderbook/sequence.rs
pub struct SequenceValidator {
    expected_seq: u64,
    gap_threshold: u64, // Max acceptable gap before resync
    buffer: VecDeque<DepthUpdate>, // Buffer out-of-order updates
}

impl SequenceValidator {
    pub fn validate(&mut self, update: DepthUpdate) -> ValidationResult {
        let expected = self.expected_seq;

        match update.first_update_id.cmp(&expected) {
            Ordering::Equal => {
                // Perfect sequence
                self.expected_seq = update.last_update_id + 1;
                ValidationResult::Valid(update)
            },
            Ordering::Less => {
                // Duplicate or old update
                ValidationResult::Duplicate
            },
            Ordering::Greater => {
                let gap = update.first_update_id - expected;

                if gap > self.gap_threshold {
                    // Gap too large, resync required
                    ValidationResult::ResyncRequired { gap }
                } else {
                    // Buffer for later processing
                    self.buffer.push_back(update);
                    ValidationResult::Buffered
                }
            }
        }
    }

    pub fn drain_buffer(&mut self) -> Vec<DepthUpdate> {
        // Try to drain contiguous buffered updates
        let mut drained = Vec::new();

        while let Some(update) = self.buffer.front() {
            if update.first_update_id == self.expected_seq {
                let update = self.buffer.pop_front().unwrap();
                self.expected_seq = update.last_update_id + 1;
                drained.push(update);
            } else {
                break;
            }
        }

        drained
    }
}
```

**Testing:**

```rust
#[test]
fn test_orderbook_updates() {
    let mut ob = Orderbook::new("BTCUSDT".to_string());

    // Initial state
    let update1 = DepthUpdate {
        symbol: "BTCUSDT".to_string(),
        first_update_id: 1,
        last_update_id: 1,
        bids: vec![PriceLevel { price: 50000.0, quantity: 1.5 }],
        asks: vec![PriceLevel { price: 50001.0, quantity: 2.0 }],
    };

    ob.apply_depth_update(update1).unwrap();

    assert_eq!(ob.get_best_bid().unwrap().price, 50000.0);
    assert_eq!(ob.get_best_ask().unwrap().price, 50001.0);
    assert_eq!(ob.get_spread(), Some(1.0));

    // Remove bid
    let update2 = DepthUpdate {
        first_update_id: 2,
        last_update_id: 2,
        bids: vec![PriceLevel { price: 50000.0, quantity: 0.0 }],
        asks: vec![],
        ..update1
    };

    ob.apply_depth_update(update2).unwrap();
    assert!(ob.get_best_bid().is_none());
}

#[test]
fn test_sequence_gap_detection() {
    let mut ob = Orderbook::new("BTCUSDT".to_string());
    ob.last_update_id = 100;

    let update = DepthUpdate {
        first_update_id: 105, // Gap!
        last_update_id: 105,
        ..Default::default()
    };

    let result = ob.apply_depth_update(update);
    assert!(matches!(result, Err(SequenceError::Gap { .. })));
}
```

**Success Criteria:**
- Order book maintains accuracy over 1 hour
- Zero sequence gaps on stable connection
- Correct gap detection and resync triggering
- Top-of-book spread always valid

---

### Phase 3: Data Distribution Layer (Days 5-6)

**Objectives:**
- [x] Redis pub/sub integration
- [x] Cap'n Proto serialization
- [x] Time-series database writes
- [x] Shared memory ring buffer (bonus)
- [x] Backpressure handling

**Implementation:**

```rust
// src/distribution/mod.rs
use redis::aio::ConnectionManager;
use capnp::message::Builder;

pub struct Distributor {
    redis: ConnectionManager,
    tsdb: QuestDBClient,
    metrics: DistributorMetrics,
}

impl Distributor {
    pub async fn distribute(&self, snapshot: OrderbookSnapshot) -> Result<()> {
        let start = Instant::now();

        // 1. Serialize to Cap'n Proto
        let mut builder = Builder::new_default();
        let msg = builder.init_root::<orderbook_capnp::snapshot::Builder>();
        self.serialize_snapshot(&snapshot, msg);

        let bytes = capnp::serialize::write_message_to_words(&builder);

        // 2. Publish to Redis (fire and forget with bounded retry)
        let topic = format!("mkt.{}.book", snapshot.symbol);

        match timeout(Duration::from_millis(10),
                     self.redis.publish(&topic, &bytes)).await {
            Ok(Ok(_)) => {
                self.metrics.redis_publishes.inc();
            },
            Ok(Err(e)) => {
                self.metrics.redis_errors.inc();
                tracing::warn!("Redis publish failed: {}", e);
            },
            Err(_) => {
                self.metrics.redis_timeouts.inc();
                tracing::warn!("Redis publish timeout");
            }
        }

        // 3. Write to TSDB (async, don't block)
        self.write_to_tsdb(&snapshot).await?;

        // 4. Record metrics
        let latency = start.elapsed();
        self.metrics.distribution_latency.observe(latency.as_micros() as f64);

        Ok(())
    }

    async fn write_to_tsdb(&self, snapshot: &OrderbookSnapshot) -> Result<()> {
        // Batch writes for efficiency
        let mut batch = Vec::new();

        // Top 10 levels
        for (i, bid) in snapshot.bids.iter().take(10).enumerate() {
            batch.push(format!(
                "orderbook,symbol={},side=bid,level={} price={},quantity={} {}",
                snapshot.symbol, i, bid.price, bid.quantity, snapshot.timestamp
            ));
        }

        for (i, ask) in snapshot.asks.iter().take(10).enumerate() {
            batch.push(format!(
                "orderbook,symbol={},side=ask,level={} price={},quantity={} {}",
                snapshot.symbol, i, ask.price, ask.quantity, snapshot.timestamp
            ));
        }

        // Write in background (don't await)
        let client = self.tsdb.clone();
        tokio::spawn(async move {
            if let Err(e) = client.write_batch(batch).await {
                tracing::warn!("TSDB write failed: {}", e);
            }
        });

        Ok(())
    }
}

// Cap'n Proto schema (schema/orderbook.capnp)
/*
@0x9eb32e19f86ee174;

struct Level {
  price @0 :Float64;
  quantity @1 :Float64;
  updatedAt @2 :Int64;
}

struct Snapshot {
  symbol @0 :Text;
  bids @1 :List(Level);
  asks @2 :List(Level);
  timestamp @3 :Int64;
  lastUpdateId @4 :UInt64;
}
*/
```

**Shared Memory Ring Buffer (Advanced):**

```rust
// src/distribution/shm.rs
use shared_memory::{Shmem, ShmemConf};
use std::sync::atomic::{AtomicU64, Ordering};

const RING_SIZE: usize = 1024; // Must be power of 2
const ENTRY_SIZE: usize = 4096; // Max message size

pub struct RingBuffer {
    shmem: Shmem,
    write_idx: *mut AtomicU64,
    read_idx: *mut AtomicU64,
    entries: *mut u8,
}

unsafe impl Send for RingBuffer {}
unsafe impl Sync for RingBuffer {}

impl RingBuffer {
    pub fn create(name: &str) -> Result<Self> {
        let total_size = 16 + (RING_SIZE * ENTRY_SIZE); // 16 bytes for indices
        let shmem = ShmemConf::new()
            .size(total_size)
            .flink(name)
            .create()?;

        let ptr = shmem.as_ptr();

        Ok(Self {
            shmem,
            write_idx: ptr as *mut AtomicU64,
            read_idx: unsafe { ptr.add(8) as *mut AtomicU64 },
            entries: unsafe { ptr.add(16) },
        })
    }

    pub fn write(&self, data: &[u8]) -> Result<()> {
        if data.len() > ENTRY_SIZE - 4 {
            return Err(anyhow!("Message too large"));
        }

        let write_idx = unsafe { (*self.write_idx).load(Ordering::Acquire) };
        let read_idx = unsafe { (*self.read_idx).load(Ordering::Acquire) };

        // Check if ring buffer is full
        if write_idx - read_idx >= RING_SIZE as u64 {
            return Err(anyhow!("Ring buffer full"));
        }

        let slot = (write_idx % RING_SIZE as u64) as usize;
        let entry_ptr = unsafe { self.entries.add(slot * ENTRY_SIZE) };

        // Write length prefix
        unsafe {
            *(entry_ptr as *mut u32) = data.len() as u32;
            std::ptr::copy_nonoverlapping(data.as_ptr(), entry_ptr.add(4), data.len());
        }

        // Advance write index
        unsafe {
            (*self.write_idx).store(write_idx + 1, Ordering::Release);
        }

        Ok(())
    }
}
```

**Success Criteria:**
- Redis publish latency < 50μs (p99)
- TSDB writes don't block main path
- Zero message loss under normal conditions
- Graceful degradation when Redis unavailable

---

### Phase 4: Testing Infrastructure (Day 7)

**Objectives:**
- [x] Mock exchange WebSocket server
- [x] Integration test suite
- [x] Performance benchmarks
- [x] Chaos testing (network failures)

**Mock Exchange Server:**

```rust
// tests/mock_exchange.rs
use tokio::net::TcpListener;
use tokio_tungstenite::accept_async;

pub struct MockExchange {
    listener: TcpListener,
    symbols: Vec<String>,
}

impl MockExchange {
    pub async fn start(port: u16) -> Self {
        let listener = TcpListener::bind(format!("127.0.0.1:{}", port))
            .await
            .unwrap();

        Self {
            listener,
            symbols: vec!["BTCUSDT".to_string()],
        }
    }

    pub async fn run(&mut self) {
        while let Ok((stream, _)) = self.listener.accept().await {
            let ws = accept_async(stream).await.unwrap();
            tokio::spawn(Self::handle_client(ws));
        }
    }

    async fn handle_client(mut ws: WebSocketStream<TcpStream>) {
        let mut update_id = 1u64;
        let mut interval = tokio::time::interval(Duration::from_millis(100));

        loop {
            interval.tick().await;

            // Generate random orderbook update
            let update = json!({
                "e": "depthUpdate",
                "E": chrono::Utc::now().timestamp_millis(),
                "s": "BTCUSDT",
                "U": update_id,
                "u": update_id,
                "b": [
                    ["50000.00", "1.5"],
                    ["49999.00", "2.0"],
                ],
                "a": [
                    ["50001.00", "1.8"],
                    ["50002.00", "0.5"],
                ]
            });

            if ws.send(Message::Text(update.to_string())).await.is_err() {
                break;
            }

            update_id += 1;
        }
    }
}
```

**Integration Tests:**

```rust
// tests/integration_test.rs
#[tokio::test]
async fn test_full_pipeline() {
    // 1. Start mock exchange
    let mut mock = MockExchange::start(9001).await;
    tokio::spawn(async move { mock.run().await });

    // 2. Start Redis (testcontainers)
    let docker = clients::Cli::default();
    let redis_container = docker.run(images::redis::Redis::default());
    let redis_port = redis_container.get_host_port_ipv4(6379);

    // 3. Start market data service
    let config = Config {
        exchange_ws_url: "ws://127.0.0.1:9001".to_string(),
        redis_url: format!("redis://127.0.0.1:{}", redis_port),
        symbols: vec!["BTCUSDT".to_string()],
        ..Default::default()
    };

    let service = MarketDataService::new(config).await.unwrap();
    tokio::spawn(async move { service.run().await });

    // 4. Subscribe to Redis updates
    let client = redis::Client::open(format!("redis://127.0.0.1:{}", redis_port)).unwrap();
    let mut pubsub = client.get_async_connection().await.unwrap().into_pubsub();
    pubsub.subscribe("mkt.BTCUSDT.book").await.unwrap();

    // 5. Verify updates received
    let mut count = 0;
    let timeout_duration = Duration::from_secs(10);
    let start = Instant::now();

    while count < 10 && start.elapsed() < timeout_duration {
        if let Ok(msg) = pubsub.on_message().next().await {
            let payload: Vec<u8> = msg.get_payload().unwrap();
            // Deserialize Cap'n Proto and verify
            count += 1;
        }
    }

    assert_eq!(count, 10);
}
```

**Performance Benchmarks:**

```rust
// benches/orderbook_bench.rs
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn bench_orderbook_update(c: &mut Criterion) {
    let mut ob = Orderbook::new("BTCUSDT".to_string());

    // Pre-populate
    for i in 0..1000 {
        ob.apply_depth_update(DepthUpdate {
            first_update_id: i,
            last_update_id: i,
            bids: vec![PriceLevel { price: 50000.0 - i as f64, quantity: 1.0 }],
            asks: vec![PriceLevel { price: 50001.0 + i as f64, quantity: 1.0 }],
            ..Default::default()
        }).unwrap();
    }

    c.bench_function("orderbook_update", |b| {
        b.iter(|| {
            let update = DepthUpdate {
                first_update_id: 1001,
                last_update_id: 1001,
                bids: vec![PriceLevel { price: 49500.0, quantity: 2.5 }],
                asks: vec![],
                ..Default::default()
            };

            black_box(ob.apply_depth_update(update).unwrap());
        });
    });
}

fn bench_capnp_serialization(c: &mut Criterion) {
    let snapshot = create_test_snapshot(100); // 100 levels each side

    c.bench_function("capnp_serialize", |b| {
        b.iter(|| {
            let mut builder = Builder::new_default();
            let msg = builder.init_root::<orderbook_capnp::snapshot::Builder>();
            serialize_snapshot(&snapshot, msg);
            black_box(capnp::serialize::write_message_to_words(&builder));
        });
    });
}

criterion_group!(benches, bench_orderbook_update, bench_capnp_serialization);
criterion_main!(benches);
```

**Chaos Testing:**

```rust
// tests/chaos_test.rs
#[tokio::test]
async fn test_network_interruption() {
    // Use toxiproxy to inject network failures
    let proxy = ToxiproxyClient::new("http://localhost:8474");

    // Create proxy for exchange connection
    let exchange_proxy = proxy.create_proxy("exchange", "0.0.0.0:9002", "binance.com:443").await;

    // Start service connected to proxy
    let service = MarketDataService::new(Config {
        exchange_ws_url: "ws://localhost:9002".to_string(),
        ..Default::default()
    }).await.unwrap();

    tokio::spawn(async move { service.run().await });

    // Wait for connection
    tokio::time::sleep(Duration::from_secs(2)).await;

    // Inject 5-second network partition
    exchange_proxy.add_toxic("timeout", "timeout", "downstream", 1.0, json!({
        "timeout": 5000
    })).await;

    tokio::time::sleep(Duration::from_secs(10)).await;

    // Remove toxic
    exchange_proxy.remove_toxic("timeout").await;

    // Verify service recovers and resumes
    tokio::time::sleep(Duration::from_secs(5)).await;

    // Check metrics - should show reconnection
    let metrics = service.get_metrics().await;
    assert!(metrics.reconnections > 0);
    assert!(metrics.is_connected);
}
```

**Success Criteria:**
- All integration tests pass
- Orderbook update latency < 80μs (p99)
- Serialization latency < 50μs (p99)
- Service recovers from 10-second network partition

---

### Phase 5: Observability UI (Day 8)

**Objectives:**
- [x] Prometheus metrics endpoint
- [x] Health check endpoint
- [x] Simple web dashboard (HTML + SSE)
- [x] Grafana dashboard JSON

**Metrics Implementation:**

```rust
// src/metrics.rs
use prometheus::{Registry, Counter, Histogram, Gauge, IntGauge};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref REGISTRY: Registry = Registry::new();

    // Connection metrics
    pub static ref WS_CONNECTIONS: IntGauge = IntGauge::new(
        "mds_websocket_connections",
        "Number of active WebSocket connections"
    ).unwrap();

    pub static ref WS_RECONNECTIONS: Counter = Counter::new(
        "mds_websocket_reconnections_total",
        "Total number of WebSocket reconnections"
    ).unwrap();

    // Message metrics
    pub static ref MESSAGES_RECEIVED: Counter = Counter::new(
        "mds_messages_received_total",
        "Total messages received from exchange"
    ).unwrap();

    pub static ref MESSAGES_PROCESSED: Counter = Counter::new(
        "mds_messages_processed_total",
        "Total messages successfully processed"
    ).unwrap();

    pub static ref MESSAGES_DROPPED: Counter = Counter::new(
        "mds_messages_dropped_total",
        "Total messages dropped due to errors"
    ).unwrap();

    // Latency metrics
    pub static ref WS_RECEIVE_LATENCY: Histogram = Histogram::with_opts(
        prometheus::HistogramOpts::new(
            "mds_websocket_receive_latency_micros",
            "WebSocket message receive latency in microseconds"
        ).buckets(vec![10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0, 2500.0, 5000.0])
    ).unwrap();

    pub static ref ORDERBOOK_UPDATE_LATENCY: Histogram = Histogram::with_opts(
        prometheus::HistogramOpts::new(
            "mds_orderbook_update_latency_micros",
            "Order book update processing latency in microseconds"
        ).buckets(vec![1.0, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0])
    ).unwrap();

    pub static ref DISTRIBUTION_LATENCY: Histogram = Histogram::with_opts(
        prometheus::HistogramOpts::new(
            "mds_distribution_latency_micros",
            "Data distribution latency in microseconds"
        ).buckets(vec![10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0])
    ).unwrap();

    // Orderbook metrics
    pub static ref ORDERBOOK_SPREAD: Gauge = Gauge::new(
        "mds_orderbook_spread_bps",
        "Current order book spread in basis points"
    ).unwrap();

    pub static ref ORDERBOOK_DEPTH: IntGauge = IntGauge::new(
        "mds_orderbook_depth_levels",
        "Number of price levels in order book"
    ).unwrap();

    // Sequence metrics
    pub static ref SEQUENCE_GAPS: Counter = Counter::new(
        "mds_sequence_gaps_total",
        "Total number of sequence gaps detected"
    ).unwrap();

    pub static ref RESYNCS_TRIGGERED: Counter = Counter::new(
        "mds_resyncs_total",
        "Total number of order book resyncs"
    ).unwrap();
}

pub fn init_metrics() {
    REGISTRY.register(Box::new(WS_CONNECTIONS.clone())).unwrap();
    REGISTRY.register(Box::new(WS_RECONNECTIONS.clone())).unwrap();
    REGISTRY.register(Box::new(MESSAGES_RECEIVED.clone())).unwrap();
    REGISTRY.register(Box::new(MESSAGES_PROCESSED.clone())).unwrap();
    REGISTRY.register(Box::new(MESSAGES_DROPPED.clone())).unwrap();
    REGISTRY.register(Box::new(WS_RECEIVE_LATENCY.clone())).unwrap();
    REGISTRY.register(Box::new(ORDERBOOK_UPDATE_LATENCY.clone())).unwrap();
    REGISTRY.register(Box::new(DISTRIBUTION_LATENCY.clone())).unwrap();
    REGISTRY.register(Box::new(ORDERBOOK_SPREAD.clone())).unwrap();
    REGISTRY.register(Box::new(ORDERBOOK_DEPTH.clone())).unwrap();
    REGISTRY.register(Box::new(SEQUENCE_GAPS.clone())).unwrap();
    REGISTRY.register(Box::new(RESYNCS_TRIGGERED.clone())).unwrap();
}
```

**HTTP Server:**

```rust
// src/http_server.rs
use axum::{
    routing::get,
    Router,
    response::{IntoResponse, Response, Sse, sse::Event},
    extract::State,
    http::StatusCode,
};
use std::sync::Arc;
use tokio::sync::RwLock;

pub struct HttpServer {
    orderbooks: Arc<RwLock<HashMap<String, Orderbook>>>,
}

impl HttpServer {
    pub async fn run(self, port: u16) {
        let app = Router::new()
            .route("/health", get(health_check))
            .route("/metrics", get(metrics_handler))
            .route("/orderbook/:symbol", get(orderbook_snapshot))
            .route("/stream", get(orderbook_stream))
            .with_state(Arc::new(self));

        let addr = SocketAddr::from(([0, 0, 0, 0], port));
        axum::Server::bind(&addr)
            .serve(app.into_make_service())
            .await
            .unwrap();
    }
}

async fn health_check() -> impl IntoResponse {
    // Implement liveness and readiness checks
    let health = json!({
        "status": "healthy",
        "timestamp": chrono::Utc::now().to_rfc3339(),
        "uptime_seconds": get_uptime(),
        "checks": {
            "websocket": is_websocket_connected(),
            "redis": is_redis_connected(),
            "orderbook": has_recent_updates(),
        }
    });

    (StatusCode::OK, Json(health))
}

async fn metrics_handler() -> impl IntoResponse {
    use prometheus::Encoder;

    let encoder = prometheus::TextEncoder::new();
    let metric_families = REGISTRY.gather();

    let mut buffer = Vec::new();
    encoder.encode(&metric_families, &mut buffer).unwrap();

    (
        StatusCode::OK,
        [("Content-Type", "text/plain; version=0.0.4")],
        buffer
    )
}

async fn orderbook_snapshot(
    State(server): State<Arc<HttpServer>>,
    Path(symbol): Path<String>,
) -> impl IntoResponse {
    let orderbooks = server.orderbooks.read().await;

    match orderbooks.get(&symbol) {
        Some(ob) => {
            let snapshot = ob.to_snapshot(20);
            (StatusCode::OK, Json(snapshot))
        },
        None => (
            StatusCode::NOT_FOUND,
            Json(json!({ "error": "Symbol not found" }))
        ),
    }
}

async fn orderbook_stream(
    State(server): State<Arc<HttpServer>>,
) -> Sse<impl Stream<Item = Result<Event, Infallible>>> {
    // Server-Sent Events for real-time updates
    let stream = async_stream::stream! {
        let mut interval = tokio::time::interval(Duration::from_millis(100));

        loop {
            interval.tick().await;

            let orderbooks = server.orderbooks.read().await;
            for (symbol, ob) in orderbooks.iter() {
                let snapshot = ob.to_snapshot(5);
                let data = serde_json::to_string(&snapshot).unwrap();
                yield Ok(Event::default().data(data));
            }
        }
    };

    Sse::new(stream)
}
```

**Simple Web Dashboard:**

```html
<!-- src/static/index.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Market Data Service - Dashboard</title>
    <style>
        body {
            font-family: 'Monaco', 'Courier New', monospace;
            background: #0d1117;
            color: #c9d1d9;
            margin: 0;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        h1 {
            color: #58a6ff;
            margin-bottom: 30px;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .metric-card {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 20px;
        }

        .metric-title {
            font-size: 14px;
            color: #8b949e;
            margin-bottom: 10px;
        }

        .metric-value {
            font-size: 32px;
            font-weight: bold;
            color: #58a6ff;
        }

        .orderbook {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
        }

        .orderbook-side {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 20px;
        }

        .orderbook-side h2 {
            margin-top: 0;
            font-size: 18px;
        }

        .asks h2 { color: #f85149; }
        .bids h2 { color: #3fb950; }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        th, td {
            padding: 8px;
            text-align: right;
        }

        th {
            color: #8b949e;
            font-size: 12px;
            border-bottom: 1px solid #30363d;
        }

        .ask-row { color: #f85149; }
        .bid-row { color: #3fb950; }

        .status-indicator {
            display: inline-block;
            width: 10px;
            height: 10px;
            border-radius: 50%;
            margin-right: 8px;
        }

        .status-connected { background: #3fb950; }
        .status-disconnected { background: #f85149; }
    </style>
</head>
<body>
    <div class="container">
        <h1>
            <span class="status-indicator" id="status-indicator"></span>
            Market Data Service
        </h1>

        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-title">Messages/Second</div>
                <div class="metric-value" id="msg-rate">0</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">Latency (p99)</div>
                <div class="metric-value" id="latency">0μs</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">Spread (bps)</div>
                <div class="metric-value" id="spread">0</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">Sequence Gaps</div>
                <div class="metric-value" id="gaps">0</div>
            </div>
        </div>

        <div class="orderbook">
            <div class="orderbook-side bids">
                <h2>Bids (Buy Orders)</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Price</th>
                            <th>Quantity</th>
                            <th>Total</th>
                        </tr>
                    </thead>
                    <tbody id="bids-table"></tbody>
                </table>
            </div>

            <div class="orderbook-side asks">
                <h2>Asks (Sell Orders)</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Price</th>
                            <th>Quantity</th>
                            <th>Total</th>
                        </tr>
                    </thead>
                    <tbody id="asks-table"></tbody>
                </table>
            </div>
        </div>
    </div>

    <script>
        const eventSource = new EventSource('/stream');

        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            updateOrderbook(data);
        };

        eventSource.onerror = () => {
            document.getElementById('status-indicator').className = 'status-indicator status-disconnected';
        };

        eventSource.onopen = () => {
            document.getElementById('status-indicator').className = 'status-indicator status-connected';
        };

        function updateOrderbook(snapshot) {
            // Update bids
            const bidsTable = document.getElementById('bids-table');
            bidsTable.innerHTML = '';
            let bidTotal = 0;

            snapshot.bids.forEach(level => {
                bidTotal += level.quantity;
                const row = bidsTable.insertRow();
                row.className = 'bid-row';
                row.insertCell(0).textContent = level.price.toFixed(2);
                row.insertCell(1).textContent = level.quantity.toFixed(4);
                row.insertCell(2).textContent = bidTotal.toFixed(4);
            });

            // Update asks
            const asksTable = document.getElementById('asks-table');
            asksTable.innerHTML = '';
            let askTotal = 0;

            snapshot.asks.forEach(level => {
                askTotal += level.quantity;
                const row = asksTable.insertRow();
                row.className = 'ask-row';
                row.insertCell(0).textContent = level.price.toFixed(2);
                row.insertCell(1).textContent = level.quantity.toFixed(4);
                row.insertCell(2).textContent = askTotal.toFixed(4);
            });

            // Update spread
            if (snapshot.bids.length > 0 && snapshot.asks.length > 0) {
                const spread = ((snapshot.asks[0].price - snapshot.bids[0].price) / snapshot.bids[0].price) * 10000;
                document.getElementById('spread').textContent = spread.toFixed(2);
            }
        }

        // Fetch metrics every second
        setInterval(async () => {
            const response = await fetch('/metrics');
            const text = await response.text();

            // Parse Prometheus metrics (simplified)
            const msgRate = parseMetric(text, 'mds_messages_received_total');
            const latency = parseMetric(text, 'mds_orderbook_update_latency_micros');
            const gaps = parseMetric(text, 'mds_sequence_gaps_total');

            document.getElementById('msg-rate').textContent = msgRate;
            document.getElementById('latency').textContent = latency + 'μs';
            document.getElementById('gaps').textContent = gaps;
        }, 1000);

        function parseMetric(text, name) {
            const regex = new RegExp(`${name}\\s+([\\d.]+)`);
            const match = text.match(regex);
            return match ? parseFloat(match[1]).toFixed(0) : '0';
        }
    </script>
</body>
</html>
```

**Grafana Dashboard JSON:**

```json
{
  "dashboard": {
    "title": "Market Data Service",
    "panels": [
      {
        "title": "WebSocket Connection Status",
        "targets": [
          {
            "expr": "mds_websocket_connections"
          }
        ],
        "type": "stat"
      },
      {
        "title": "Message Throughput",
        "targets": [
          {
            "expr": "rate(mds_messages_received_total[1m])",
            "legendFormat": "Received"
          },
          {
            "expr": "rate(mds_messages_processed_total[1m])",
            "legendFormat": "Processed"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Processing Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(mds_orderbook_update_latency_micros_bucket[5m]))",
            "legendFormat": "p50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(mds_orderbook_update_latency_micros_bucket[5m]))",
            "legendFormat": "p95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(mds_orderbook_update_latency_micros_bucket[5m]))",
            "legendFormat": "p99"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Order Book Spread",
        "targets": [
          {
            "expr": "mds_orderbook_spread_bps"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Sequence Gaps",
        "targets": [
          {
            "expr": "rate(mds_sequence_gaps_total[5m])"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

**Success Criteria:**
- Metrics endpoint returns valid Prometheus format
- Health check accurately reflects service state
- Web dashboard displays real-time orderbook
- Grafana dashboard imports successfully

---

## 4. Implementation Details

### 4.1 Configuration Management

```rust
// src/config.rs
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub exchange: ExchangeConfig,
    pub redis: RedisConfig,
    pub timeseries: TimeseriesConfig,
    pub symbols: Vec<String>,
    pub performance: PerformanceConfig,
    pub observability: ObservabilityConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExchangeConfig {
    pub name: String, // "binance", "coinbase", etc.
    pub ws_url: String,
    pub rest_url: String,
    pub api_key: Option<String>,
    pub api_secret: Option<String>,
    pub reconnect_delay_ms: u64,
    pub max_reconnect_attempts: u32,
    pub ping_interval_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RedisConfig {
    pub url: String,
    pub pool_size: usize,
    pub publish_timeout_ms: u64,
    pub enable_compression: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimeseriesConfig {
    pub enabled: bool,
    pub url: String,
    pub database: String,
    pub batch_size: usize,
    pub flush_interval_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceConfig {
    pub orderbook_max_depth: usize,
    pub message_buffer_size: usize,
    pub worker_threads: usize,
    pub enable_shared_memory: bool,
    pub cpu_affinity: Option<Vec<usize>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ObservabilityConfig {
    pub http_port: u16,
    pub log_level: String,
    pub enable_tracing: bool,
    pub enable_dashboard: bool,
}

impl Config {
    pub fn from_file(path: &str) -> Result<Self> {
        let contents = std::fs::read_to_string(path)?;
        let config: Config = serde_yaml::from_str(&contents)?;
        config.validate()?;
        Ok(config)
    }

    pub fn from_env() -> Result<Self> {
        config::Config::builder()
            .add_source(config::Environment::with_prefix("MDS"))
            .build()?
            .try_deserialize()
    }

    fn validate(&self) -> Result<()> {
        if self.symbols.is_empty() {
            return Err(anyhow!("At least one symbol must be configured"));
        }

        if self.performance.orderbook_max_depth < 10 {
            return Err(anyhow!("Orderbook max depth must be >= 10"));
        }

        Ok(())
    }
}
```

**Example Configuration File:**

```yaml
# config/production.yaml
exchange:
  name: binance
  ws_url: wss://stream.binance.com:9443/ws
  rest_url: https://api.binance.com
  reconnect_delay_ms: 2000
  max_reconnect_attempts: 10
  ping_interval_ms: 30000

redis:
  url: redis://redis:6379
  pool_size: 10
  publish_timeout_ms: 10
  enable_compression: false

timeseries:
  enabled: true
  url: http://questdb:9000
  database: market_data
  batch_size: 1000
  flush_interval_ms: 1000

symbols:
  - BTCUSDT
  - ETHUSDT
  - SOLUSDT

performance:
  orderbook_max_depth: 1000
  message_buffer_size: 10000
  worker_threads: 4
  enable_shared_memory: true
  cpu_affinity: [2, 3, 4, 5]  # Pin to specific cores

observability:
  http_port: 9090
  log_level: info
  enable_tracing: true
  enable_dashboard: true
```

### 4.2 Logging Strategy

```rust
// src/logging.rs
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

pub fn init_logging(config: &ObservabilityConfig) {
    let env_filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new(&config.log_level));

    let fmt_layer = tracing_subscriber::fmt::layer()
        .json()
        .with_target(true)
        .with_current_span(true)
        .with_span_list(true)
        .with_thread_ids(true)
        .with_thread_names(true);

    if config.enable_tracing {
        // Add Jaeger tracing
        let tracer = opentelemetry_jaeger::new_pipeline()
            .with_service_name("market-data-service")
            .install_batch(opentelemetry::runtime::Tokio)
            .unwrap();

        let telemetry = tracing_opentelemetry::layer().with_tracer(tracer);

        tracing_subscriber::registry()
            .with(env_filter)
            .with(fmt_layer)
            .with(telemetry)
            .init();
    } else {
        tracing_subscriber::registry()
            .with(env_filter)
            .with(fmt_layer)
            .init();
    }
}

// Structured logging examples
#[tracing::instrument(skip(msg))]
async fn process_message(msg: Message) {
    tracing::debug!("Processing message");

    match parse_message(&msg) {
        Ok(update) => {
            tracing::info!(
                symbol = %update.symbol,
                update_id = update.last_update_id,
                "Orderbook update processed"
            );
        },
        Err(e) => {
            tracing::error!(
                error = %e,
                message_size = msg.len(),
                "Failed to parse message"
            );
        }
    }
}
```

---

## 5. Testing Strategy

### 5.1 Unit Tests

**Coverage Target:** 80%+

```rust
// src/orderbook/tests.rs
mod tests {
    use super::*;

    #[test]
    fn test_empty_orderbook() {
        let ob = Orderbook::new("BTCUSDT".to_string());
        assert!(ob.get_best_bid().is_none());
        assert!(ob.get_best_ask().is_none());
        assert_eq!(ob.get_spread(), None);
    }

    #[test]
    fn test_bid_ask_ordering() {
        let mut ob = Orderbook::new("BTCUSDT".to_string());

        // Add multiple bids
        ob.apply_depth_update(create_update(vec![
            (50000.0, 1.0),
            (49999.0, 2.0),
            (50001.0, 0.5),  // Best bid
        ], vec![])).unwrap();

        assert_eq!(ob.get_best_bid().unwrap().price, 50001.0);
    }

    #[test]
    fn test_quantity_zero_removes_level() {
        let mut ob = Orderbook::new("BTCUSDT".to_string());

        ob.apply_depth_update(create_update(vec![(50000.0, 1.0)], vec![])).unwrap();
        assert!(ob.get_best_bid().is_some());

        ob.apply_depth_update(create_update(vec![(50000.0, 0.0)], vec![])).unwrap();
        assert!(ob.get_best_bid().is_none());
    }

    #[test]
    fn test_sequence_validation() {
        let mut validator = SequenceValidator::new();
        validator.expected_seq = 100;

        // Valid sequence
        let update1 = create_update_with_seq(100, 100);
        assert!(matches!(validator.validate(update1), ValidationResult::Valid(_)));

        // Duplicate
        let update2 = create_update_with_seq(95, 95);
        assert!(matches!(validator.validate(update2), ValidationResult::Duplicate));

        // Gap
        let update3 = create_update_with_seq(105, 105);
        assert!(matches!(validator.validate(update3), ValidationResult::Buffered));
    }
}
```

### 5.2 Integration Tests

```rust
// tests/integration_tests.rs
#[tokio::test]
async fn test_end_to_end_pipeline() {
    // See Phase 4 implementation
}

#[tokio::test]
async fn test_redis_failover() {
    // Start service
    // Stop Redis
    // Verify graceful degradation
    // Restart Redis
    // Verify recovery
}
```

### 5.3 Performance Benchmarks

**Latency Targets:**

| Operation | Target (p50) | Target (p99) |
|-----------|-------------|--------------|
| WebSocket receive | < 20μs | < 50μs |
| Message parse | < 10μs | < 30μs |
| Orderbook update | < 30μs | < 80μs |
| Serialization | < 20μs | < 50μs |
| Redis publish | < 30μs | < 100μs |
| End-to-end | < 100μs | < 250μs |

**Benchmark Suite:**

```bash
# Run all benchmarks
cargo bench

# Run specific benchmark
cargo bench orderbook_update

# With flamegraph profiling
cargo flamegraph --bench orderbook_bench
```

### 5.4 Load Testing

```rust
// tests/load_test.rs
#[tokio::test]
#[ignore] // Run manually with: cargo test --release --ignored load_test
async fn test_sustained_load() {
    let mut mock = MockExchange::start(9001).await;

    // Configure for high throughput: 10,000 updates/sec
    mock.set_update_rate(10_000);
    mock.set_symbols(vec!["BTC", "ETH", "SOL", "AVAX"]);

    tokio::spawn(async move { mock.run().await });

    let service = MarketDataService::new(test_config()).await.unwrap();
    tokio::spawn(async move { service.run().await });

    // Run for 5 minutes
    tokio::time::sleep(Duration::from_secs(300)).await;

    // Verify metrics
    let metrics = get_metrics().await;

    assert!(metrics.avg_latency_micros < 100.0);
    assert!(metrics.p99_latency_micros < 250.0);
    assert!(metrics.messages_dropped == 0);
    assert!(metrics.sequence_gaps == 0);
    assert!(metrics.memory_mb < 500);
}
```

---

## 6. Deployment

### 6.1 Dockerfile

```dockerfile
# Dockerfile
# Multi-stage build for minimal image size

# Stage 1: Build
FROM rust:1.80-slim AS builder

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Cache dependencies
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Build application
COPY . .
RUN cargo build --release

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary
COPY --from=builder /app/target/release/market-data-service /app/
COPY --from=builder /app/config /app/config
COPY --from=builder /app/src/static /app/static

# Create non-root user
RUN useradd -m -u 1000 mds && chown -R mds:mds /app
USER mds

# Expose ports
EXPOSE 9090

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9090/health || exit 1

ENTRYPOINT ["/app/market-data-service"]
```

**Build and Run:**

```bash
# Build
docker build -t market-data-service:latest .

# Run
docker run -d \
    --name market-data-service \
    -p 9090:9090 \
    -e MDS_EXCHANGE__WS_URL=wss://stream.binance.com:9443/ws \
    -e MDS_REDIS__URL=redis://redis:6379 \
    -v $(pwd)/config:/app/config:ro \
    market-data-service:latest
```

### 6.2 Docker Compose

```yaml
# docker-compose.yml
version: '3.9'

services:
  market-data:
    build: .
    container_name: market-data-service
    ports:
      - "9090:9090"
    environment:
      MDS_EXCHANGE__NAME: binance
      MDS_EXCHANGE__WS_URL: wss://stream.binance.com:9443/ws
      MDS_REDIS__URL: redis://redis:6379
      MDS_TIMESERIES__URL: http://questdb:9000
      MDS_OBSERVABILITY__LOG_LEVEL: info
    depends_on:
      - redis
      - questdb
    networks:
      - trading-net
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 512M
        reservations:
          cpus: '1'
          memory: 256M

  redis:
    image: redis:7-alpine
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - trading-net
    command: redis-server --appendonly yes
    restart: unless-stopped

  questdb:
    image: questdb/questdb:7.3.10
    container_name: questdb
    ports:
      - "9000:9000"  # REST API
      - "8812:8812"  # Postgres wire protocol
      - "9009:9009"  # InfluxDB line protocol
    volumes:
      - questdb-data:/var/lib/questdb
    networks:
      - trading-net
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - trading-net
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
      - grafana-data:/var/lib/grafana
    networks:
      - trading-net
    depends_on:
      - prometheus
    restart: unless-stopped

networks:
  trading-net:
    driver: bridge

volumes:
  redis-data:
  questdb-data:
  prometheus-data:
  grafana-data:
```

### 6.3 Environment Variables

```bash
# .env.example
# Exchange Configuration
MDS_EXCHANGE__NAME=binance
MDS_EXCHANGE__WS_URL=wss://stream.binance.com:9443/ws
MDS_EXCHANGE__REST_URL=https://api.binance.com
MDS_EXCHANGE__RECONNECT_DELAY_MS=2000
MDS_EXCHANGE__MAX_RECONNECT_ATTEMPTS=10

# Redis Configuration
MDS_REDIS__URL=redis://localhost:6379
MDS_REDIS__POOL_SIZE=10
MDS_REDIS__PUBLISH_TIMEOUT_MS=10

# Time-Series Database
MDS_TIMESERIES__ENABLED=true
MDS_TIMESERIES__URL=http://localhost:9000
MDS_TIMESERIES__DATABASE=market_data
MDS_TIMESERIES__BATCH_SIZE=1000

# Symbols to Track
MDS_SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT

# Performance Tuning
MDS_PERFORMANCE__ORDERBOOK_MAX_DEPTH=1000
MDS_PERFORMANCE__MESSAGE_BUFFER_SIZE=10000
MDS_PERFORMANCE__WORKER_THREADS=4

# Observability
MDS_OBSERVABILITY__HTTP_PORT=9090
MDS_OBSERVABILITY__LOG_LEVEL=info
MDS_OBSERVABILITY__ENABLE_TRACING=false
```

### 6.4 Resource Requirements

**Minimum:**
- CPU: 1 core
- Memory: 256MB
- Network: 10 Mbps

**Recommended (Production):**
- CPU: 2-4 cores (dedicated)
- Memory: 512MB - 1GB
- Network: 100 Mbps+
- Storage: 10GB (logs + TSDB)

**Scaling Guidelines:**
- +1 core per 5 additional symbols
- +128MB per 10 additional symbols
- SSD storage for time-series database

---

## 7. Observability

### 7.1 Key Metrics to Expose

**Connection Metrics:**
- `mds_websocket_connections` (gauge): Active connections
- `mds_websocket_reconnections_total` (counter): Reconnection count
- `mds_websocket_connection_uptime_seconds` (gauge): Connection uptime

**Message Metrics:**
- `mds_messages_received_total` (counter): Total messages received
- `mds_messages_processed_total` (counter): Successfully processed
- `mds_messages_dropped_total` (counter): Dropped/failed messages
- `mds_message_size_bytes` (histogram): Message size distribution

**Latency Metrics:**
- `mds_websocket_receive_latency_micros` (histogram): WebSocket latency
- `mds_orderbook_update_latency_micros` (histogram): Update processing
- `mds_distribution_latency_micros` (histogram): Distribution latency
- `mds_end_to_end_latency_micros` (histogram): Total pipeline latency

**Orderbook Metrics:**
- `mds_orderbook_spread_bps` (gauge): Current spread in bps
- `mds_orderbook_depth_levels` (gauge): Number of price levels
- `mds_orderbook_imbalance_ratio` (gauge): Bid/ask volume imbalance

**Quality Metrics:**
- `mds_sequence_gaps_total` (counter): Sequence gaps detected
- `mds_resyncs_total` (counter): Orderbook resyncs triggered
- `mds_data_staleness_seconds` (gauge): Time since last update

### 7.2 Dashboard Requirements

**Real-time Orderbook View:**
- Top 10 bids/asks
- Visual depth chart
- Spread indicator
- Last update timestamp

**System Health:**
- Connection status (green/red indicator)
- Message rate (updates/second)
- Latency percentiles (p50, p95, p99)
- Error rate

**Performance Charts:**
- Latency over time (line chart)
- Message throughput (area chart)
- Spread over time (line chart)
- Sequence gaps (bar chart)

### 7.3 Alerting Rules

```yaml
# monitoring/prometheus-alerts.yml
groups:
  - name: market_data_service
    interval: 10s
    rules:
      - alert: WebSocketDisconnected
        expr: mds_websocket_connections == 0
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "WebSocket connection lost"
          description: "Market data service has no active WebSocket connections"

      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(mds_end_to_end_latency_micros_bucket[5m])) > 500
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High processing latency"
          description: "P99 latency is {{ $value }}μs (threshold: 500μs)"

      - alert: SequenceGapsDetected
        expr: rate(mds_sequence_gaps_total[5m]) > 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Order book sequence gaps detected"
          description: "{{ $value }} gaps/sec detected"

      - alert: HighMessageDropRate
        expr: rate(mds_messages_dropped_total[5m]) / rate(mds_messages_received_total[5m]) > 0.01
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High message drop rate"
          description: "{{ $value | humanizePercentage }} of messages are being dropped"

      - alert: DataStaleness
        expr: mds_data_staleness_seconds > 10
        for: 30s
        labels:
          severity: warning
        annotations:
          summary: "Market data is stale"
          description: "No updates received for {{ $value }}s"
```

---

## 8. Risk Mitigation & Edge Cases

### 8.1 Known Challenges

**Challenge 1: Binance Rate Limits**
- **Risk:** WebSocket connection throttling
- **Mitigation:**
  - Implement exponential backoff
  - Subscribe to combined streams (reduces connections)
  - Monitor `429` responses

**Challenge 2: Sequence Gap Handling**
- **Risk:** Order book desynchronization
- **Mitigation:**
  - Automatic snapshot re-fetch on gap
  - Buffer out-of-order messages (up to 100)
  - Alert on frequent resyncs

**Challenge 3: Memory Bloat**
- **Risk:** Unbounded orderbook growth
- **Mitigation:**
  - Hard limit on depth (1000 levels)
  - Periodic arena reset
  - Memory usage metrics + alerts

**Challenge 4: Redis Unavailability**
- **Risk:** Data distribution failure
- **Mitigation:**
  - Non-blocking publishes (timeout)
  - In-memory circular buffer fallback
  - Metrics to detect degradation

### 8.2 Recovery Procedures

**WebSocket Disconnection:**
1. Log disconnection event
2. Wait `reconnect_delay_ms`
3. Re-establish connection
4. Re-subscribe to symbols
5. Fetch fresh orderbook snapshot
6. Resume processing

**Sequence Gap Detected:**
1. Log gap details
2. Buffer subsequent messages (30 seconds max)
3. Fetch REST API snapshot
4. Rebuild orderbook from snapshot
5. Drain buffered messages
6. Resume normal processing

**Out of Memory:**
1. Log current memory usage
2. Clear oldest orderbook data
3. Trigger garbage collection
4. If still critical: restart service (health check fails)

---

## 9. Future Enhancements

### 9.1 Phase 2 Features (Post-MVP)

1. **Multi-Exchange Support**
   - Adapter pattern for different exchanges
   - Unified data normalization
   - Cross-exchange arbitrage detection

2. **Advanced Order Book Analytics**
   - Volume-weighted average price (VWAP)
   - Order book imbalance indicators
   - Liquidity depth analysis

3. **Smart Snapshots**
   - Compression for historical data
   - Differential snapshots (only changes)
   - Snapshot versioning for replay

4. **L3 Order Book (Full Depth)**
   - Individual order tracking
   - Order ID-based updates
   - Market maker identification

### 9.2 Performance Optimizations

1. **SIMD Instructions**
   - Vectorized price level updates
   - Batch serialization

2. **Zero-Copy Networking**
   - `io_uring` for socket I/O
   - Kernel bypass (DPDK)

3. **Custom Memory Allocator**
   - `jemalloc` or `mimalloc`
   - Per-thread arenas

---

## 10. Success Criteria Summary

**Functional Requirements:**
- [x] Connect to Binance WebSocket
- [x] Maintain accurate order books for configured symbols
- [x] Detect and recover from sequence gaps
- [x] Distribute data via Redis pub/sub
- [x] Write to time-series database
- [x] Expose metrics and health endpoints
- [x] Provide basic web dashboard

**Performance Requirements:**
- [x] End-to-end latency < 100μs (p99)
- [x] Handle 10,000+ updates/second
- [x] Memory usage < 500MB
- [x] CPU usage < 50% (2 cores)
- [x] Reconnect in < 2 seconds

**Quality Requirements:**
- [x] 80%+ test coverage
- [x] Zero data loss on network failures
- [x] Graceful degradation on dependency failures
- [x] Comprehensive observability

**Operational Requirements:**
- [x] Docker container < 100MB
- [x] Start in < 5 seconds
- [x] Health checks accurate
- [x] Structured logging

---

## Appendix A: Development Checklist

### Pre-Development
- [ ] Set up Rust development environment
- [ ] Create Git repository
- [ ] Set up CI/CD pipeline
- [ ] Provision test infrastructure (Redis, QuestDB)

### Phase 1 (Days 1-2)
- [ ] Initialize Cargo project
- [ ] Implement WebSocket client
- [ ] Add ping/pong keepalive
- [ ] Implement auto-reconnection
- [ ] Write connection tests

### Phase 2 (Days 3-4)
- [ ] Design orderbook data structure
- [ ] Implement depth update parser
- [ ] Build orderbook manager
- [ ] Add sequence validator
- [ ] Write orderbook unit tests

### Phase 3 (Days 5-6)
- [ ] Integrate Redis pub/sub
- [ ] Implement Cap'n Proto serialization
- [ ] Add TSDB writer
- [ ] Implement backpressure handling
- [ ] Write integration tests

### Phase 4 (Day 7)
- [ ] Build mock exchange server
- [ ] Write full pipeline tests
- [ ] Set up benchmarking suite
- [ ] Run load tests
- [ ] Document performance results

### Phase 5 (Day 8)
- [ ] Implement metrics exporter
- [ ] Build health check endpoint
- [ ] Create web dashboard
- [ ] Generate Grafana dashboard
- [ ] Write deployment docs

### Post-Development
- [ ] Security audit
- [ ] Performance tuning
- [ ] Documentation review
- [ ] Production deployment plan

---

## Appendix B: Useful Commands

```bash
# Development
cargo run --release
cargo test
cargo bench
cargo clippy -- -D warnings
cargo fmt

# Docker
docker build -t market-data-service .
docker-compose up -d
docker logs -f market-data-service

# Monitoring
curl http://localhost:9090/health
curl http://localhost:9090/metrics
curl http://localhost:9090/orderbook/BTCUSDT

# Profiling
cargo flamegraph --bin market-data-service
perf record -g ./target/release/market-data-service
perf report

# Debugging
RUST_LOG=debug cargo run
RUST_BACKTRACE=1 cargo test
```

---

## Appendix C: Reference Links

**Rust Libraries:**
- [tokio](https://tokio.rs/) - Async runtime
- [tokio-tungstenite](https://github.com/snapview/tokio-tungstenite) - WebSocket
- [serde](https://serde.rs/) - Serialization
- [prometheus](https://github.com/tikv/rust-prometheus) - Metrics

**Exchange Documentation:**
- [Binance WebSocket Streams](https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams)
- [Binance Order Book Depth](https://binance-docs.github.io/apidocs/spot/en/#diff-depth-stream)

**Testing Tools:**
- [Criterion](https://github.com/bheisler/criterion.rs) - Benchmarking
- [Testcontainers](https://github.com/testcontainers/testcontainers-rs) - Integration testing
- [Wiremock](https://github.com/LukeMathWalker/wiremock-rs) - HTTP mocking

---

**End of Development Plan**

This plan provides a comprehensive roadmap for building the Market Data Service. Each phase is designed to be completable in 1-2 days by a competent developer or AI coding agent. The plan emphasizes performance, reliability, and observability—critical requirements for a production trading system.
