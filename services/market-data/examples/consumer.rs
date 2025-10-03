/// Example consumer for Market Data Service in Rust
/// Subscribes to order book updates via Redis pub/sub
///
/// Usage: cargo run --example consumer [SYMBOL]

use redis::{Commands, ControlFlow};
use serde::Deserialize;
use std::env;
use chrono::{DateTime, Utc};

#[derive(Debug, Deserialize)]
struct OrderBook {
    symbol: String,
    bids: std::collections::BTreeMap<String, f64>,
    asks: std::collections::BTreeMap<String, f64>,
    last_update_id: u64,
    timestamp: i64,
}

#[derive(Debug, Deserialize)]
struct Trade {
    symbol: String,
    trade_id: u64,
    price: f64,
    quantity: f64,
    timestamp: i64,
    is_buyer_maker: bool,
}

fn format_orderbook(book: OrderBook) {
    println!("\n{}", "=".repeat(60));
    let dt = DateTime::from_timestamp_micros(book.timestamp).unwrap_or_default();
    println!("Symbol: {} | Time: {}", book.symbol, dt.format("%H:%M:%S%.6f"));
    println!("{}", "=".repeat(60));

    // Top 5 asks (reversed for display)
    println!("\nAsks (sellers):");
    let mut asks: Vec<_> = book.asks.iter().collect();
    asks.sort_by(|a, b| b.0.cmp(a.0));
    for (price, qty) in asks.iter().take(5).rev() {
        let price_f: f64 = price.parse().unwrap_or(0.0);
        println!("  {:>12.2} | {:>10.4}", price_f, qty);
    }

    println!("\n{}", "-".repeat(40));

    // Top 5 bids
    println!("\nBids (buyers):");
    let mut bids: Vec<_> = book.bids.iter().collect();
    bids.sort_by(|a, b| b.0.cmp(a.0));
    for (price, qty) in bids.iter().take(5) {
        let price_f: f64 = price.parse().unwrap_or(0.0);
        println!("  {:>12.2} | {:>10.4}", price_f, qty);
    }

    // Calculate spread
    if let (Some(best_bid), Some(best_ask)) = (bids.first(), asks.last()) {
        let bid_price: f64 = best_bid.0.parse().unwrap_or(0.0);
        let ask_price: f64 = best_ask.0.parse().unwrap_or(0.0);
        let spread = ask_price - bid_price;
        let spread_bps = (spread / bid_price) * 10000.0;
        println!("\nSpread: ${:.2} ({:.2} bps)", spread, spread_bps);
    }
}

fn format_trade(trade: Trade) {
    let dt = DateTime::from_timestamp_millis(trade.timestamp).unwrap_or_default();
    let side = if trade.is_buyer_maker { "SELL" } else { "BUY" };
    println!(
        "[{}] {} {:4} {:>10.4} @ ${:>10.2}",
        dt.format("%H:%M:%S"),
        trade.symbol,
        side,
        trade.quantity,
        trade.price
    );
}

fn main() -> redis::RedisResult<()> {
    let args: Vec<String> = env::args().collect();
    let symbol = args.get(1).map(|s| s.as_str()).unwrap_or("BTCUSDT");

    println!("Connecting to Redis...");
    let client = redis::Client::open("redis://127.0.0.1/")?;
    let mut con = client.get_connection()?;

    // Test connection
    redis::cmd("PING").query::<String>(&mut con)?;
    println!("Connected to Redis successfully");

    let orderbook_channel = format!("orderbook:{}", symbol);
    let trade_channel = format!("trades:{}", symbol);

    println!("\nSubscribing to:");
    println!("  - {}", orderbook_channel);
    println!("  - {}", trade_channel);
    println!("\nWaiting for messages... (Ctrl+C to quit)\n");

    let mut pubsub = con.as_pubsub();
    pubsub.subscribe(&orderbook_channel)?;
    pubsub.subscribe(&trade_channel)?;

    loop {
        let msg = pubsub.get_message()?;
        let payload: String = msg.get_payload()?;
        let channel: String = msg.get_channel_name();

        if channel.contains("orderbook") {
            if let Ok(book) = serde_json::from_str::<OrderBook>(&payload) {
                format_orderbook(book);
            }
        } else if channel.contains("trades") {
            if let Ok(trade) = serde_json::from_str::<Trade>(&payload) {
                format_trade(trade);
            }
        }
    }
}
