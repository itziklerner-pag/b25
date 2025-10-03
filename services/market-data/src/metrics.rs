use lazy_static::lazy_static;
use prometheus::{
    register_counter_vec, register_gauge_vec, register_histogram_vec,
    CounterVec, GaugeVec, HistogramVec, TextEncoder, Encoder,
};

lazy_static! {
    // WebSocket connection status
    pub static ref WS_CONNECTED: GaugeVec = register_gauge_vec!(
        "websocket_connected",
        "WebSocket connection status (1=connected, 0=disconnected)",
        &["symbol"]
    )
    .unwrap();

    pub static ref WS_DISCONNECTS: CounterVec = register_counter_vec!(
        "websocket_disconnects_total",
        "Total number of WebSocket disconnections",
        &["symbol"]
    )
    .unwrap();

    // Message processing
    pub static ref MESSAGES_PROCESSED: CounterVec = register_counter_vec!(
        "messages_processed_total",
        "Total number of messages processed",
        &["symbol", "type"]
    )
    .unwrap();

    pub static ref MESSAGES_ERROR: CounterVec = register_counter_vec!(
        "messages_error_total",
        "Total number of message processing errors",
        &["symbol"]
    )
    .unwrap();

    pub static ref PROCESSING_LATENCY: HistogramVec = register_histogram_vec!(
        "processing_latency_microseconds",
        "Message processing latency in microseconds",
        &["symbol"],
        vec![1.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 1000.0]
    )
    .unwrap();

    // Order book updates
    pub static ref ORDERBOOK_UPDATES: CounterVec = register_counter_vec!(
        "orderbook_updates_total",
        "Total number of order book updates",
        &["symbol"]
    )
    .unwrap();

    pub static ref SEQUENCE_ERRORS: CounterVec = register_counter_vec!(
        "sequence_errors_total",
        "Total number of sequence validation errors",
        &["symbol"]
    )
    .unwrap();

    // Trades
    pub static ref TRADES_PROCESSED: CounterVec = register_counter_vec!(
        "trades_processed_total",
        "Total number of trades processed",
        &["symbol"]
    )
    .unwrap();

    // Redis publishing
    pub static ref REDIS_PUBLISHES: CounterVec = register_counter_vec!(
        "redis_publishes_total",
        "Total number of Redis publishes",
        &["symbol", "type"]
    )
    .unwrap();

    pub static ref REDIS_ERRORS: CounterVec = register_counter_vec!(
        "redis_errors_total",
        "Total number of Redis errors",
        &["symbol"]
    )
    .unwrap();
}

pub fn encode_metrics() -> Result<String, Box<dyn std::error::Error>> {
    let encoder = TextEncoder::new();
    let metric_families = prometheus::gather();
    let mut buffer = Vec::new();
    encoder.encode(&metric_families, &mut buffer)?;
    Ok(String::from_utf8(buffer)?)
}
