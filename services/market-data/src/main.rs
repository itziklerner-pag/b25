mod config;
mod orderbook;
mod publisher;
mod websocket;
mod metrics;
mod shm;
mod health;

use anyhow::Result;
use std::sync::Arc;
use tokio::signal;
use tracing::{info, error};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::config::Config;
use crate::orderbook::OrderBookManager;
use crate::publisher::Publisher;
use crate::websocket::WebSocketClient;
use crate::health::HealthServer;

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "market_data_service=debug,info".into()),
        )
        .with(tracing_subscriber::fmt::layer().with_target(true))
        .init();

    info!("Starting Market Data Service");

    // Load configuration
    let config = Config::from_file("config.yaml")
        .or_else(|_| Config::from_file("config.example.yaml"))
        .unwrap_or_else(|_| {
            info!("No config file found, using defaults");
            Config::default()
        });

    info!("Configuration loaded: {} symbols", config.symbols.len());

    // Initialize shared components
    let orderbook_manager = Arc::new(OrderBookManager::new(config.order_book_depth));
    let publisher = Arc::new(
        Publisher::new(&config.redis_url, &config.shm_name)
            .await
            .expect("Failed to initialize publisher")
    );

    // Start health check server
    let health_server = HealthServer::new(config.health_port);
    let health_handle = tokio::spawn(async move {
        if let Err(e) = health_server.start().await {
            error!("Health server error: {}", e);
        }
    });

    // Start WebSocket clients for each symbol
    let mut ws_handles = Vec::new();

    for symbol in &config.symbols {
        let symbol = symbol.clone();
        let ws_url = config.exchange_ws_url.clone();
        let orderbook_manager = Arc::clone(&orderbook_manager);
        let publisher = Arc::clone(&publisher);

        let handle = tokio::spawn(async move {
            let client = WebSocketClient::new(
                symbol.clone(),
                ws_url,
                orderbook_manager,
                publisher,
            );

            loop {
                info!("Starting WebSocket client for {}", symbol);
                match client.connect_and_run().await {
                    Ok(_) => {
                        info!("WebSocket client for {} exited normally", symbol);
                        break;
                    }
                    Err(e) => {
                        error!("WebSocket client error for {}: {}", symbol, e);
                        // Exponential backoff
                        tokio::time::sleep(tokio::time::Duration::from_secs(5)).await;
                    }
                }
            }
        });

        ws_handles.push(handle);
    }

    info!("All WebSocket clients started");

    // Wait for shutdown signal
    match signal::ctrl_c().await {
        Ok(()) => {
            info!("Shutdown signal received");
        }
        Err(err) => {
            error!("Unable to listen for shutdown signal: {}", err);
        }
    }

    // Cleanup
    for handle in ws_handles {
        handle.abort();
    }
    health_handle.abort();

    info!("Market Data Service stopped");
    Ok(())
}
