use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Position {
    pub symbol: String,
    pub size: f64,
    pub entry_price: f64,
    pub current_price: f64,
    pub side: PositionSide,
    pub unrealized_pnl: f64,
    pub pnl_percent: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PositionSide {
    Long,
    Short,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Order {
    pub id: String,
    pub symbol: String,
    pub side: OrderSide,
    pub order_type: OrderType,
    pub price: f64,
    pub size: f64,
    pub filled_size: f64,
    pub status: OrderStatus,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OrderSide {
    Buy,
    Sell,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OrderType {
    Limit,
    Market,
    StopLimit,
    StopMarket,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OrderStatus {
    New,
    PartiallyFilled,
    Filled,
    Canceled,
    Rejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderBook {
    pub symbol: String,
    pub bids: Vec<(f64, f64)>, // (price, size)
    pub asks: Vec<(f64, f64)>, // (price, size)
    pub timestamp: DateTime<Utc>,
}

impl OrderBook {
    pub fn spread(&self) -> f64 {
        if let (Some(best_bid), Some(best_ask)) = (
            self.bids.first().map(|b| b.0),
            self.asks.first().map(|a| a.0),
        ) {
            best_ask - best_bid
        } else {
            0.0
        }
    }

    pub fn mid_price(&self) -> f64 {
        if let (Some(best_bid), Some(best_ask)) = (
            self.bids.first().map(|b| b.0),
            self.asks.first().map(|a| a.0),
        ) {
            (best_ask + best_bid) / 2.0
        } else {
            0.0
        }
    }

    pub fn spread_percent(&self) -> f64 {
        let mid = self.mid_price();
        if mid > 0.0 {
            (self.spread() / mid) * 100.0
        } else {
            0.0
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Fill {
    pub id: String,
    pub order_id: String,
    pub symbol: String,
    pub side: OrderSide,
    pub price: f64,
    pub size: f64,
    pub fee: f64,
    pub pnl: f64,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Signal {
    pub id: String,
    pub strategy: String,
    pub symbol: String,
    pub signal_type: SignalType,
    pub strength: f64, // 0.0 to 1.0
    pub target_price: f64,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum SignalType {
    Long,
    Short,
    Neutral,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Alert {
    pub id: String,
    pub level: AlertLevel,
    pub message: String,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AlertLevel {
    Info,
    Warning,
    Error,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConnectionStatus {
    Disconnected,
    Connecting,
    Connected,
    Error,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Panel {
    Positions,
    Orders,
    Fills,
    OrderBook,
    Signals,
    Alerts,
}

impl Panel {
    pub fn next(&self) -> Self {
        match self {
            Panel::Positions => Panel::Orders,
            Panel::Orders => Panel::Fills,
            Panel::Fills => Panel::OrderBook,
            Panel::OrderBook => Panel::Signals,
            Panel::Signals => Panel::Alerts,
            Panel::Alerts => Panel::Positions,
        }
    }

    pub fn prev(&self) -> Self {
        match self {
            Panel::Positions => Panel::Alerts,
            Panel::Orders => Panel::Positions,
            Panel::Fills => Panel::Orders,
            Panel::OrderBook => Panel::Fills,
            Panel::Signals => Panel::OrderBook,
            Panel::Alerts => Panel::Signals,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum InputMode {
    Normal,
    Command,
}
