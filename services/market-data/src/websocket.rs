use anyhow::{Result, Context};
use futures_util::{SinkExt, StreamExt};
use serde::Deserialize;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::sleep;
use tokio_tungstenite::{connect_async, tungstenite::Message};
use tracing::{debug, info, warn};

use crate::orderbook::{DepthUpdate, OrderBookManager, PriceLevel, Trade};
use crate::publisher::Publisher;
use crate::snapshot::SnapshotFetcher;
use crate::metrics;

pub struct WebSocketClient {
    symbol: String,
    ws_url: String,
    orderbook_manager: Arc<OrderBookManager>,
    publisher: Arc<Publisher>,
    _snapshot_fetcher: Arc<SnapshotFetcher>,
    order_book_depth: usize,
    reconnect_delay: Duration,
    max_reconnect_delay: Duration,
}

#[derive(Debug, Deserialize)]
#[serde(tag = "e")]
enum BinanceMessage {
    #[serde(rename = "depthUpdate")]
    DepthUpdate(BinanceDepthUpdate),
    #[serde(rename = "aggTrade")]
    AggTrade(BinanceAggTrade),
}

#[derive(Debug, Deserialize)]
struct BinanceDepthUpdate {
    #[serde(rename = "s")]
    symbol: String,
    #[serde(rename = "U")]
    first_update_id: u64,
    #[serde(rename = "u")]
    last_update_id: u64,
    #[serde(rename = "b")]
    bids: Vec<(String, String)>, // [price, quantity]
    #[serde(rename = "a")]
    asks: Vec<(String, String)>,
}

#[derive(Debug, Deserialize)]
struct BinanceAggTrade {
    #[serde(rename = "s")]
    symbol: String,
    #[serde(rename = "a")]
    trade_id: u64,
    #[serde(rename = "p")]
    price: String,
    #[serde(rename = "q")]
    quantity: String,
    #[serde(rename = "T")]
    timestamp: i64,
    #[serde(rename = "m")]
    is_buyer_maker: bool,
}

impl WebSocketClient {
    pub fn new(
        symbol: String,
        ws_url: String,
        orderbook_manager: Arc<OrderBookManager>,
        publisher: Arc<Publisher>,
        snapshot_fetcher: Arc<SnapshotFetcher>,
        order_book_depth: usize,
    ) -> Self {
        Self {
            symbol,
            ws_url,
            orderbook_manager,
            publisher,
            _snapshot_fetcher: snapshot_fetcher,
            order_book_depth,
            reconnect_delay: Duration::from_millis(1000),
            max_reconnect_delay: Duration::from_secs(60),
        }
    }

    pub async fn connect_and_run(&self) -> Result<()> {
        let mut current_delay = self.reconnect_delay;

        loop {
            match self.run_connection().await {
                Ok(_) => {
                    info!("WebSocket connection closed normally for {}", self.symbol);
                    return Ok(());
                }
                Err(e) => {
                    warn!("WebSocket error for {}: {}", self.symbol, e);
                    metrics::WS_DISCONNECTS.with_label_values(&[&self.symbol]).inc();

                    // Exponential backoff
                    warn!(
                        "Reconnecting {} in {:?}...",
                        self.symbol, current_delay
                    );
                    sleep(current_delay).await;
                    current_delay = std::cmp::min(current_delay * 2, self.max_reconnect_delay);
                }
            }
        }
    }

    async fn run_connection(&self) -> Result<()> {
        // Skip REST snapshot fetch (geo-blocked) - build orderbook from WebSocket
        info!("Building orderbook for {} from WebSocket updates (REST API geo-blocked)", self.symbol);

        // Connect to WebSocket for incremental updates
        let streams = format!(
            "{}@depth@100ms/{}@aggTrade",
            self.symbol.to_lowercase(),
            self.symbol.to_lowercase()
        );
        let url = format!("{}?streams={}", self.ws_url, streams);

        info!("Connecting to {} for {}", url, self.symbol);

        let (ws_stream, _) = connect_async(&url)
            .await
            .context("Failed to connect to WebSocket")?;

        info!("Connected to WebSocket for {}", self.symbol);
        metrics::WS_CONNECTED.with_label_values(&[&self.symbol]).set(1.0);

        let (mut write, mut read) = ws_stream.split();

        // Send ping periodically
        let symbol_clone = self.symbol.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(30));
            loop {
                interval.tick().await;
                if write.send(Message::Ping(vec![])).await.is_err() {
                    warn!("Failed to send ping for {}", symbol_clone);
                    break;
                }
            }
        });

        // Process messages
        while let Some(msg) = read.next().await {
            let msg = msg.context("WebSocket message error")?;

            match msg {
                Message::Text(text) => {
                    if let Err(e) = self.process_message(&text).await {
                        debug!("Error processing message: {}", e);
                        metrics::MESSAGES_ERROR
                            .with_label_values(&[&self.symbol])
                            .inc();
                    }
                }
                Message::Pong(_) => {
                    debug!("Received pong for {}", self.symbol);
                }
                Message::Close(_) => {
                    info!("WebSocket closed for {}", self.symbol);
                    metrics::WS_CONNECTED.with_label_values(&[&self.symbol]).set(0.0);
                    break;
                }
                _ => {}
            }
        }

        Ok(())
    }

    async fn process_message(&self, text: &str) -> Result<()> {
        // Binance sends wrapped messages
        #[derive(Deserialize)]
        struct Wrapper {
            stream: String,
            data: serde_json::Value,
        }

        let wrapper: Wrapper = serde_json::from_str(text)
            .context("Failed to parse wrapper")?;

        // Parse the inner message based on stream type
        if wrapper.stream.contains("depth") {
            let depth_update: BinanceDepthUpdate = serde_json::from_value(wrapper.data)
                .context("Failed to parse depth update")?;
            self.handle_depth_update(depth_update).await?;
        } else if wrapper.stream.contains("aggTrade") {
            let trade: BinanceAggTrade = serde_json::from_value(wrapper.data)
                .context("Failed to parse trade")?;
            self.handle_trade(trade).await?;
        }

        Ok(())
    }

    async fn handle_depth_update(&self, update: BinanceDepthUpdate) -> Result<()> {
        // Convert Binance format to internal format
        let depth_update = DepthUpdate {
            symbol: update.symbol.clone(),
            first_update_id: update.first_update_id,
            last_update_id: update.last_update_id,
            bids: update
                .bids
                .iter()
                .map(|(p, q)| PriceLevel {
                    price: p.parse().unwrap_or(0.0),
                    quantity: q.parse().unwrap_or(0.0),
                })
                .collect(),
            asks: update
                .asks
                .iter()
                .map(|(p, q)| PriceLevel {
                    price: p.parse().unwrap_or(0.0),
                    quantity: q.parse().unwrap_or(0.0),
                })
                .collect(),
        };

        // Update order book
        match self.orderbook_manager.update(&self.symbol, depth_update.clone()) {
            Ok(book) => {
                // Only publish if we have meaningful data
                if book.bids.len() > 0 && book.asks.len() > 0 {
                    metrics::ORDERBOOK_UPDATES
                        .with_label_values(&[&self.symbol])
                        .inc();

                    // Publish to Redis and shared memory
                    self.publisher
                        .publish_orderbook(&book)
                        .await
                        .context("Failed to publish orderbook")?;

                    debug!(
                        "Updated order book for {}: {} bids, {} asks, last_update_id={}",
                        self.symbol,
                        book.bids.len(),
                        book.asks.len(),
                        book.last_update_id
                    );
                }
            }
            Err(e) => {
                // Sequence error - reset orderbook to accept next update as baseline
                debug!("Sequence error for {}: {}. Resetting to accept next update.", self.symbol, e);
                metrics::SEQUENCE_ERRORS
                    .with_label_values(&[&self.symbol])
                    .inc();

                // Reset orderbook - next update will be accepted as baseline
                let mut books = self.orderbook_manager.books.write().unwrap();
                books.remove(&self.symbol);
            }
        }

        Ok(())
    }

    async fn handle_trade(&self, trade_data: BinanceAggTrade) -> Result<()> {
        let trade = Trade {
            symbol: trade_data.symbol.clone(),
            trade_id: trade_data.trade_id,
            price: trade_data.price.parse().unwrap_or(0.0),
            quantity: trade_data.quantity.parse().unwrap_or(0.0),
            timestamp: trade_data.timestamp,
            is_buyer_maker: trade_data.is_buyer_maker,
        };

        metrics::TRADES_PROCESSED
            .with_label_values(&[&self.symbol])
            .inc();

        // Publish trade
        self.publisher
            .publish_trade(&trade)
            .await
            .context("Failed to publish trade")?;

        debug!(
            "Processed trade for {}: {} @ {}",
            self.symbol, trade.quantity, trade.price
        );

        Ok(())
    }
}
