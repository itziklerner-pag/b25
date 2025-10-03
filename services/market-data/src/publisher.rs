use anyhow::{Result, Context};
use redis::aio::ConnectionManager;
use redis::{AsyncCommands, Client};
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
        let channel = format!("orderbook:{}", book.symbol);
        let payload = serde_json::to_string(book)
            .context("Failed to serialize order book")?;

        // Publish to Redis
        match self.publish_redis(&channel, &payload).await {
            Ok(_) => {
                metrics::REDIS_PUBLISHES
                    .with_label_values(&[&book.symbol, "orderbook"])
                    .inc();
            }
            Err(e) => {
                error!("Failed to publish to Redis: {}", e);
                metrics::REDIS_ERRORS
                    .with_label_values(&[&book.symbol])
                    .inc();
            }
        }

        // Write to shared memory for ultra-low latency local consumers
        if let Err(e) = self.shm_ring.write(payload.as_bytes()) {
            error!("Failed to write to shared memory: {}", e);
        }

        debug!("Published order book for {}", book.symbol);
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
        conn.publish(channel, payload)
            .await
            .context("Redis publish failed")?;
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
