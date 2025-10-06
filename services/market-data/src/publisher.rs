use anyhow::{Result, Context};
use redis::aio::ConnectionManager;
use redis::{AsyncCommands, Client};
use serde::Serialize;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{debug, error};

use crate::orderbook::{OrderBook, Trade};
use crate::shm::SharedMemoryRing;
use crate::metrics;

pub struct Publisher {
    redis_client: Client,
    redis_conn: Arc<RwLock<ConnectionManager>>,
    shm_ring: Arc<SharedMemoryRing>,
}

#[derive(Serialize)]
struct MarketData {
    symbol: String,
    last_price: f64,
    bid_price: f64,
    ask_price: f64,
    volume_24h: f64,
    high_24h: f64,
    low_24h: f64,
    updated_at: String,
}

impl Publisher {
    pub async fn new(redis_url: &str, shm_name: &str) -> Result<Self> {
        let redis_client = Client::open(redis_url)
            .context("Failed to create Redis client")?;

        let conn_manager = ConnectionManager::new(redis_client.clone())
            .await
            .context("Failed to create Redis connection manager")?;

        let shm_ring = SharedMemoryRing::new(shm_name, 1024 * 1024) // 1MB ring buffer
            .context("Failed to create shared memory ring")?;

        Ok(Self {
            redis_client,
            redis_conn: Arc::new(RwLock::new(conn_manager)),
            shm_ring: Arc::new(shm_ring),
        })
    }

    pub async fn publish_orderbook(&self, book: &OrderBook) -> Result<()> {
        // 1. Publish full orderbook to orderbook:SYMBOL channel
        let orderbook_channel = format!("orderbook:{}", book.symbol);
        let orderbook_payload = serde_json::to_string(book)
            .context("Failed to serialize order book")?;

        match self.publish_redis(&orderbook_channel, &orderbook_payload).await {
            Ok(_) => {
                metrics::REDIS_PUBLISHES
                    .with_label_values(&[&book.symbol, "orderbook"])
                    .inc();
            }
            Err(e) => {
                error!("Failed to publish orderbook to Redis: {}", e);
                metrics::REDIS_ERRORS
                    .with_label_values(&[&book.symbol])
                    .inc();
            }
        }

        // 2. Create simplified market data and store in market_data:SYMBOL key
        let best_bid = book.bids.iter().next_back().map(|(p, _)| p.0).unwrap_or(0.0);
        let best_ask = book.asks.iter().next().map(|(p, _)| p.0).unwrap_or(0.0);
        let last_price = if best_bid > 0.0 && best_ask > 0.0 {
            (best_bid + best_ask) / 2.0
        } else {
            0.0
        };

        let market_data = MarketData {
            symbol: book.symbol.clone(),
            last_price,
            bid_price: best_bid,
            ask_price: best_ask,
            volume_24h: 0.0, // TODO: Track from trades
            high_24h: 0.0,   // TODO: Track from trades
            low_24h: 0.0,    // TODO: Track from trades
            updated_at: chrono::Utc::now().to_rfc3339(),
        };

        let market_data_payload = serde_json::to_string(&market_data)
            .context("Failed to serialize market data")?;

        // Store in Redis key with 5 minute expiration
        let market_data_key = format!("market_data:{}", book.symbol);
        match self.set_redis(&market_data_key, &market_data_payload, 300).await {
            Ok(_) => {
                debug!("Stored market data for {} in Redis key", book.symbol);
            }
            Err(e) => {
                error!("Failed to store market data in Redis: {}", e);
                metrics::REDIS_ERRORS
                    .with_label_values(&[&book.symbol])
                    .inc();
            }
        }

        // 3. Publish to market_data:SYMBOL channel for dashboard aggregator
        let market_data_channel = format!("market_data:{}", book.symbol);
        match self.publish_redis(&market_data_channel, &market_data_payload).await {
            Ok(_) => {
                metrics::REDIS_PUBLISHES
                    .with_label_values(&[&book.symbol, "market_data"])
                    .inc();
            }
            Err(e) => {
                error!("Failed to publish market data to Redis: {}", e);
            }
        }

        // 4. Write full orderbook to shared memory for ultra-low latency local consumers
        if let Err(e) = self.shm_ring.write(orderbook_payload.as_bytes()) {
            error!("Failed to write to shared memory: {}", e);
        }

        debug!("Published order book and market data for {}", book.symbol);
        Ok(())
    }

    pub async fn publish_trade(&self, trade: &Trade) -> Result<()> {
        let channel = format!("trades:{}", trade.symbol);
        let payload = serde_json::to_string(trade)
            .context("Failed to serialize trade")?;

        // Publish to Redis
        match self.publish_redis(&channel, &payload).await {
            Ok(_) => {
                metrics::REDIS_PUBLISHES
                    .with_label_values(&[&trade.symbol, "trade"])
                    .inc();
            }
            Err(e) => {
                error!("Failed to publish trade to Redis: {}", e);
                metrics::REDIS_ERRORS
                    .with_label_values(&[&trade.symbol])
                    .inc();
            }
        }

        debug!("Published trade for {}", trade.symbol);
        Ok(())
    }

    async fn publish_redis(&self, channel: &str, payload: &str) -> Result<()> {
        let mut conn = self.redis_conn.write().await;
        conn.publish::<_, _, ()>(channel, payload)
            .await
            .context("Redis publish failed")?;
        Ok(())
    }

    async fn set_redis(&self, key: &str, value: &str, ttl_seconds: u64) -> Result<()> {
        let mut conn = self.redis_conn.write().await;
        conn.set_ex::<_, _, ()>(key, value, ttl_seconds)
            .await
            .context("Redis SET failed")?;
        Ok(())
    }

    pub async fn health_check(&self) -> bool {
        let mut conn = self.redis_conn.write().await;
        redis::cmd("PING")
            .query_async::<_, String>(&mut *conn)
            .await
            .is_ok()
    }
}
