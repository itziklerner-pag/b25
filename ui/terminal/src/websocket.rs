use crate::config::ConnectionConfig;
use crate::state::StateUpdate;
use crate::types::*;
use anyhow::Result;
use futures::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use tokio::sync::mpsc;
use tokio_tungstenite::{connect_async, tungstenite::Message};

#[derive(Debug, Deserialize)]
#[serde(tag = "type")]
pub enum DashboardMessage {
    #[serde(rename = "positions")]
    Positions { data: Vec<Position> },

    #[serde(rename = "orders")]
    Orders { data: Vec<Order> },

    #[serde(rename = "orderbook")]
    OrderBook { data: OrderBook },

    #[serde(rename = "fills")]
    Fills { data: Vec<Fill> },

    #[serde(rename = "signals")]
    Signals { data: Vec<Signal> },

    #[serde(rename = "alerts")]
    Alerts { data: Vec<Alert> },

    #[serde(rename = "ping")]
    Ping { timestamp: i64 },
}

#[derive(Debug, Serialize)]
#[serde(tag = "type")]
pub enum ClientMessage {
    #[serde(rename = "subscribe")]
    Subscribe { channels: Vec<String> },

    #[serde(rename = "pong")]
    Pong { timestamp: i64 },
}

pub struct WsClient {
    config: ConnectionConfig,
    state_tx: mpsc::Sender<StateUpdate>,
}

impl WsClient {
    pub fn new(config: ConnectionConfig, state_tx: mpsc::Sender<StateUpdate>) -> Self {
        Self { config, state_tx }
    }

    pub async fn connect_with_retry(&self) -> Result<()> {
        let mut backoff = 1;

        loop {
            // Update connection status to Connecting
            let _ = self
                .state_tx
                .send(StateUpdate::ConnectionStatus(
                    ConnectionStatus::Connecting,
                    0,
                ))
                .await;

            match self.connect().await {
                Ok(_) => {
                    tracing::info!("WebSocket connected successfully");
                    backoff = 1; // Reset backoff on success
                }
                Err(e) => {
                    tracing::error!("WebSocket connection failed: {}", e);

                    // Update connection status to Error
                    let _ = self
                        .state_tx
                        .send(StateUpdate::ConnectionStatus(ConnectionStatus::Error, 0))
                        .await;

                    // Calculate backoff
                    let sleep_duration = Duration::from_millis(
                        backoff.min(self.config.max_reconnect_interval_ms),
                    );
                    tracing::info!("Retrying in {:?}...", sleep_duration);

                    tokio::time::sleep(sleep_duration).await;

                    // Exponential backoff
                    backoff = (backoff * 2).min(self.config.max_reconnect_interval_ms);
                }
            }
        }
    }

    async fn connect(&self) -> Result<()> {
        tracing::info!("Connecting to WebSocket: {}", self.config.dashboard_url);

        let (ws_stream, _) = connect_async(&self.config.dashboard_url).await?;
        let (mut write, mut read) = ws_stream.split();

        // Send subscription message
        let subscribe_msg = ClientMessage::Subscribe {
            channels: vec![
                "positions".to_string(),
                "orders".to_string(),
                "orderbook".to_string(),
                "fills".to_string(),
                "signals".to_string(),
                "alerts".to_string(),
            ],
        };

        let msg_json = serde_json::to_string(&subscribe_msg)?;
        write.send(Message::Text(msg_json)).await?;

        tracing::info!("Subscription sent");

        // Update connection status to Connected
        let _ = self
            .state_tx
            .send(StateUpdate::ConnectionStatus(
                ConnectionStatus::Connected,
                0,
            ))
            .await;

        // Message processing loop
        while let Some(msg) = read.next().await {
            match msg {
                Ok(Message::Text(text)) => {
                    let start = std::time::Instant::now();

                    match serde_json::from_str::<DashboardMessage>(&text) {
                        Ok(dashboard_msg) => {
                            self.handle_message(dashboard_msg).await?;

                            let latency = start.elapsed().as_millis() as u64;
                            let _ = self
                                .state_tx
                                .send(StateUpdate::ConnectionStatus(
                                    ConnectionStatus::Connected,
                                    latency,
                                ))
                                .await;
                        }
                        Err(e) => {
                            tracing::error!("Failed to parse message: {}", e);
                            tracing::debug!("Message content: {}", text);
                        }
                    }
                }
                Ok(Message::Close(_)) => {
                    tracing::warn!("WebSocket closed by server");
                    return Err(anyhow::anyhow!("Connection closed"));
                }
                Ok(Message::Ping(data)) => {
                    write.send(Message::Pong(data)).await?;
                }
                Ok(_) => {}
                Err(e) => {
                    tracing::error!("WebSocket error: {}", e);
                    return Err(e.into());
                }
            }
        }

        Ok(())
    }

    async fn handle_message(&self, msg: DashboardMessage) -> Result<()> {
        let update = match msg {
            DashboardMessage::Positions { data } => StateUpdate::Positions(data),
            DashboardMessage::Orders { data } => StateUpdate::Orders(data),
            DashboardMessage::OrderBook { data } => StateUpdate::OrderBook(data),
            DashboardMessage::Fills { data } => StateUpdate::Fills(data),
            DashboardMessage::Signals { data } => StateUpdate::Signals(data),
            DashboardMessage::Alerts { data } => StateUpdate::Alerts(data),
            DashboardMessage::Ping { timestamp } => {
                // Handle ping/pong if needed
                tracing::trace!("Received ping: {}", timestamp);
                return Ok(());
            }
        };

        self.state_tx.send(update).await?;
        Ok(())
    }
}
