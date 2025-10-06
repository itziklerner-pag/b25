use anyhow::{Result, Context};
use serde::Deserialize;
use tracing::{info, error};

use crate::orderbook::{OrderBook, OrderedFloat};
use std::collections::BTreeMap;

#[derive(Debug, Deserialize)]
struct BinanceSnapshot {
    #[serde(rename = "lastUpdateId")]
    last_update_id: u64,
    bids: Vec<(String, String)>, // [price, quantity]
    asks: Vec<(String, String)>,
}

pub struct SnapshotFetcher {
    rest_api_url: String,
}

impl SnapshotFetcher {
    pub fn new(rest_api_url: String) -> Self {
        Self { rest_api_url }
    }

    /// Fetch orderbook snapshot from Binance REST API
    pub async fn fetch_snapshot(&self, symbol: &str, limit: usize) -> Result<OrderBook> {
        let url = format!(
            "{}/fapi/v1/depth?symbol={}&limit={}",
            self.rest_api_url, symbol, limit
        );

        info!("Fetching snapshot for {} from {}", symbol, url);

        // Use reqwest to fetch snapshot
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(10))
            .build()
            .context("Failed to create HTTP client")?;

        let response = client
            .get(&url)
            .send()
            .await
            .context("Failed to send snapshot request")?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().await.unwrap_or_default();
            anyhow::bail!("Snapshot request failed with status {}: {}", status, body);
        }

        let snapshot: BinanceSnapshot = response
            .json()
            .await
            .context("Failed to parse snapshot response")?;

        // Convert to internal OrderBook format
        let mut orderbook = OrderBook::new(symbol.to_string());
        orderbook.last_update_id = snapshot.last_update_id;
        orderbook.timestamp = chrono::Utc::now().timestamp_micros();

        // Parse bids
        for (price_str, qty_str) in snapshot.bids {
            let price: f64 = price_str.parse().unwrap_or(0.0);
            let qty: f64 = qty_str.parse().unwrap_or(0.0);
            if price > 0.0 && qty > 0.0 {
                orderbook.bids.insert(OrderedFloat(price), qty);
            }
        }

        // Parse asks
        for (price_str, qty_str) in snapshot.asks {
            let price: f64 = price_str.parse().unwrap_or(0.0);
            let qty: f64 = qty_str.parse().unwrap_or(0.0);
            if price > 0.0 && qty > 0.0 {
                orderbook.asks.insert(OrderedFloat(price), qty);
            }
        }

        info!(
            "Fetched snapshot for {}: {} bids, {} asks, last_update_id={}",
            symbol,
            orderbook.bids.len(),
            orderbook.asks.len(),
            orderbook.last_update_id
        );

        Ok(orderbook)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_fetch_snapshot_live() {
        let fetcher = SnapshotFetcher::new("https://fapi.binance.com".to_string());
        let result = fetcher.fetch_snapshot("BTCUSDT", 20).await;

        assert!(result.is_ok());
        let book = result.unwrap();
        assert!(book.bids.len() > 0);
        assert!(book.asks.len() > 0);
        assert!(book.last_update_id > 0);
    }
}
