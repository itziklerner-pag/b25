use ahash::AHashMap;
use serde::{Deserialize, Deserializer, Serialize, Serializer};
use std::collections::BTreeMap;
use std::sync::RwLock;
use chrono::Utc;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceLevel {
    pub price: f64,
    pub quantity: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderBook {
    pub symbol: String,
    pub bids: BTreeMap<OrderedFloat, f64>, // price -> quantity
    pub asks: BTreeMap<OrderedFloat, f64>, // price -> quantity
    pub last_update_id: u64,
    pub timestamp: i64,
}

// Wrapper for f64 to make it Ord for BTreeMap
#[derive(Debug, Clone, Copy, PartialEq, PartialOrd)]
pub struct OrderedFloat(pub f64);

impl Eq for OrderedFloat {}

impl Ord for OrderedFloat {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.0.partial_cmp(&other.0).unwrap_or(std::cmp::Ordering::Equal)
    }
}

impl From<f64> for OrderedFloat {
    fn from(f: f64) -> Self {
        OrderedFloat(f)
    }
}

// Custom serialization for OrderedFloat
impl Serialize for OrderedFloat {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_f64(self.0)
    }
}

// Custom deserialization for OrderedFloat
impl<'de> Deserialize<'de> for OrderedFloat {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        let f = f64::deserialize(deserializer)?;
        Ok(OrderedFloat(f))
    }
}

impl OrderBook {
    pub fn new(symbol: String) -> Self {
        Self {
            symbol,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            last_update_id: 0,
            timestamp: Utc::now().timestamp_micros(),
        }
    }

    /// Apply a depth update (delta)
    pub fn apply_update(&mut self, update: &DepthUpdate) -> Result<(), String> {
        // Sequence validation
        if update.first_update_id > self.last_update_id + 1 {
            return Err(format!(
                "Sequence gap detected: expected {}, got {}",
                self.last_update_id + 1,
                update.first_update_id
            ));
        }

        // Update bids
        for level in &update.bids {
            let price = OrderedFloat(level.price);
            if level.quantity == 0.0 {
                self.bids.remove(&price);
            } else {
                self.bids.insert(price, level.quantity);
            }
        }

        // Update asks
        for level in &update.asks {
            let price = OrderedFloat(level.price);
            if level.quantity == 0.0 {
                self.asks.remove(&price);
            } else {
                self.asks.insert(price, level.quantity);
            }
        }

        self.last_update_id = update.last_update_id;
        self.timestamp = Utc::now().timestamp_micros();

        Ok(())
    }

    /// Get top N levels
    pub fn get_top_levels(&self, depth: usize) -> (Vec<PriceLevel>, Vec<PriceLevel>) {
        let bids: Vec<PriceLevel> = self
            .bids
            .iter()
            .rev()
            .take(depth)
            .map(|(price, qty)| PriceLevel {
                price: price.0,
                quantity: *qty,
            })
            .collect();

        let asks: Vec<PriceLevel> = self
            .asks
            .iter()
            .take(depth)
            .map(|(price, qty)| PriceLevel {
                price: price.0,
                quantity: *qty,
            })
            .collect();

        (bids, asks)
    }

    /// Get mid price
    pub fn mid_price(&self) -> Option<f64> {
        let best_bid = self.bids.iter().next_back()?.0.0;
        let best_ask = self.asks.iter().next()?.0.0;
        Some((best_bid + best_ask) / 2.0)
    }

    /// Get spread
    pub fn spread(&self) -> Option<f64> {
        let best_bid = self.bids.iter().next_back()?.0.0;
        let best_ask = self.asks.iter().next()?.0.0;
        Some(best_ask - best_bid)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepthUpdate {
    pub symbol: String,
    pub first_update_id: u64,
    pub last_update_id: u64,
    pub bids: Vec<PriceLevel>,
    pub asks: Vec<PriceLevel>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Trade {
    pub symbol: String,
    pub trade_id: u64,
    pub price: f64,
    pub quantity: f64,
    pub timestamp: i64,
    pub is_buyer_maker: bool,
}

pub struct OrderBookManager {
    books: RwLock<AHashMap<String, OrderBook>>,
    depth: usize,
}

impl OrderBookManager {
    pub fn new(depth: usize) -> Self {
        Self {
            books: RwLock::new(AHashMap::new()),
            depth,
        }
    }

    pub fn get_or_create(&self, symbol: &str) -> OrderBook {
        let books = self.books.read().unwrap();
        if let Some(book) = books.get(symbol) {
            return book.clone();
        }
        drop(books);

        let mut books = self.books.write().unwrap();
        books
            .entry(symbol.to_string())
            .or_insert_with(|| OrderBook::new(symbol.to_string()))
            .clone()
    }

    pub fn update(&self, symbol: &str, update: DepthUpdate) -> Result<OrderBook, String> {
        let mut books = self.books.write().unwrap();
        let book = books
            .entry(symbol.to_string())
            .or_insert_with(|| OrderBook::new(symbol.to_string()));

        book.apply_update(&update)?;
        Ok(book.clone())
    }

    pub fn get(&self, symbol: &str) -> Option<OrderBook> {
        let books = self.books.read().unwrap();
        books.get(symbol).cloned()
    }

    pub fn snapshot(&self, symbol: &str, depth: usize) -> Option<(Vec<PriceLevel>, Vec<PriceLevel>)> {
        let books = self.books.read().unwrap();
        books.get(symbol).map(|book| book.get_top_levels(depth))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_orderbook_update() {
        let mut book = OrderBook::new("BTCUSDT".to_string());

        let update = DepthUpdate {
            symbol: "BTCUSDT".to_string(),
            first_update_id: 1,
            last_update_id: 1,
            bids: vec![
                PriceLevel { price: 50000.0, quantity: 1.5 },
                PriceLevel { price: 49999.0, quantity: 2.0 },
            ],
            asks: vec![
                PriceLevel { price: 50001.0, quantity: 1.0 },
                PriceLevel { price: 50002.0, quantity: 3.0 },
            ],
        };

        assert!(book.apply_update(&update).is_ok());
        assert_eq!(book.bids.len(), 2);
        assert_eq!(book.asks.len(), 2);

        let mid = book.mid_price().unwrap();
        assert!((mid - 50000.5).abs() < 0.01);
    }

    #[test]
    fn test_sequence_validation() {
        let mut book = OrderBook::new("BTCUSDT".to_string());
        book.last_update_id = 100;

        let update = DepthUpdate {
            symbol: "BTCUSDT".to_string(),
            first_update_id: 105, // Gap!
            last_update_id: 105,
            bids: vec![],
            asks: vec![],
        };

        assert!(book.apply_update(&update).is_err());
    }
}
