use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::fs;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub connection: ConnectionConfig,
    pub ui: UiConfig,
    pub panels: PanelsConfig,
    pub keyboard: KeyboardConfig,
    pub performance: PerformanceConfig,
    pub logging: LoggingConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionConfig {
    pub dashboard_url: String,
    pub reconnect_interval_ms: u64,
    pub max_reconnect_interval_ms: u64,
    pub timeout_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UiConfig {
    pub refresh_rate_ms: u64,
    pub color_scheme: String,
    pub show_milliseconds: bool,
    pub stale_data_threshold_s: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PanelsConfig {
    pub default_symbol: String,
    pub max_fills_display: usize,
    pub max_signals_display: usize,
    pub max_alerts_display: usize,
    pub orderbook_depth_levels: usize,
    pub alert_auto_dismiss_s: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KeyboardConfig {
    pub quit_keys: Vec<String>,
    pub help_key: String,
    pub reload_key: String,
    pub command_key: String,
    pub search_key: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceConfig {
    pub enable_dirty_flag_optimization: bool,
    pub max_cpu_percent: f64,
    pub max_memory_mb: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoggingConfig {
    pub level: String,
    pub file: String,
    pub json: bool,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            connection: ConnectionConfig {
                dashboard_url: "ws://localhost:8080/ws".to_string(),
                reconnect_interval_ms: 1000,
                max_reconnect_interval_ms: 30000,
                timeout_ms: 5000,
            },
            ui: UiConfig {
                refresh_rate_ms: 100,
                color_scheme: "default".to_string(),
                show_milliseconds: false,
                stale_data_threshold_s: 5,
            },
            panels: PanelsConfig {
                default_symbol: "BTCUSDT".to_string(),
                max_fills_display: 50,
                max_signals_display: 20,
                max_alerts_display: 100,
                orderbook_depth_levels: 10,
                alert_auto_dismiss_s: 30,
            },
            keyboard: KeyboardConfig {
                quit_keys: vec!["q".to_string()],
                help_key: "?".to_string(),
                reload_key: "r".to_string(),
                command_key: ":".to_string(),
                search_key: "/".to_string(),
            },
            performance: PerformanceConfig {
                enable_dirty_flag_optimization: true,
                max_cpu_percent: 5.0,
                max_memory_mb: 50,
            },
            logging: LoggingConfig {
                level: "info".to_string(),
                file: String::new(),
                json: false,
            },
        }
    }
}

impl Config {
    pub fn load(path: &str) -> Result<Self> {
        let contents = fs::read_to_string(path)?;
        let config: Config = serde_yaml::from_str(&contents)?;
        Ok(config)
    }
}
