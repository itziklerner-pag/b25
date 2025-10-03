use serde::{Deserialize, Serialize};
use std::fs;
use anyhow::Result;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub symbols: Vec<String>,
    pub exchange_ws_url: String,
    pub redis_url: String,
    pub order_book_depth: usize,
    pub health_port: u16,
    pub shm_name: String,
    pub reconnect_delay_ms: u64,
    pub max_reconnect_delay_ms: u64,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            symbols: vec!["BTCUSDT".to_string(), "ETHUSDT".to_string()],
            exchange_ws_url: "wss://fstream.binance.com/stream".to_string(),
            redis_url: "redis://127.0.0.1:6379".to_string(),
            order_book_depth: 20,
            health_port: 9090,
            shm_name: "market_data_shm".to_string(),
            reconnect_delay_ms: 1000,
            max_reconnect_delay_ms: 60000,
        }
    }
}

impl Config {
    pub fn from_file(path: &str) -> Result<Self> {
        let contents = fs::read_to_string(path)?;
        let config: Config = serde_yaml::from_str(&contents)?;
        Ok(config)
    }
}
