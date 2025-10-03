use crate::Decimal;
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

/// Represents a price level in the order book.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderBookLevel {
    pub price: Decimal,
    pub quantity: Decimal,
    pub order_count: usize,
}

/// Represents one side of the order book (bids or asks).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderBookSide {
    levels: BTreeMap<String, OrderBookLevel>, // price string -> level
    is_ask: bool,
}

impl OrderBookSide {
    /// Creates a new order book side.
    pub fn new(is_ask: bool) -> Self {
        OrderBookSide {
            levels: BTreeMap::new(),
            is_ask,
        }
    }

    /// Updates a price level in the order book.
    pub fn update(&mut self, price: Decimal, quantity: Decimal, order_count: usize) {
        let price_key = price.to_string();

        if quantity.is_zero() {
            self.levels.remove(&price_key);
        } else {
            let level = OrderBookLevel {
                price,
                quantity,
                order_count,
            };
            self.levels.insert(price_key, level);
        }
    }

    /// Returns the best price level (highest bid or lowest ask).
    pub fn get_best(&self) -> Option<&OrderBookLevel> {
        if self.is_ask {
            // For asks, we want the lowest price
            self.get_sorted_levels().first().map(|(_, level)| level)
        } else {
            // For bids, we want the highest price
            self.get_sorted_levels().last().map(|(_, level)| level)
        }
    }

    /// Returns the top n levels.
    pub fn get_depth(&self, n: usize) -> Vec<&OrderBookLevel> {
        let sorted = self.get_sorted_levels();
        if self.is_ask {
            sorted.iter().take(n).map(|(_, level)| level).collect()
        } else {
            sorted.iter().rev().take(n).map(|(_, level)| level).collect()
        }
    }

    /// Returns the total volume up to depth n.
    pub fn get_total_volume(&self, n: usize) -> Decimal {
        self.get_depth(n)
            .iter()
            .fold(Decimal::zero(), |acc, level| acc + level.quantity)
    }

    /// Clears all levels.
    pub fn clear(&mut self) {
        self.levels.clear();
    }

    /// Returns all levels sorted by price.
    fn get_sorted_levels(&self) -> Vec<(&String, &OrderBookLevel)> {
        self.levels.iter().collect()
    }

    /// Returns the number of levels.
    pub fn len(&self) -> usize {
        self.levels.len()
    }

    /// Returns true if there are no levels.
    pub fn is_empty(&self) -> bool {
        self.levels.is_empty()
    }
}

/// Represents a complete order book with bids and asks.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderBook {
    pub symbol: String,
    pub bids: OrderBookSide,
    pub asks: OrderBookSide,
    pub sequence_number: i64,
}

impl OrderBook {
    /// Creates a new order book.
    pub fn new(symbol: String) -> Self {
        OrderBook {
            symbol,
            bids: OrderBookSide::new(false),
            asks: OrderBookSide::new(true),
            sequence_number: 0,
        }
    }

    /// Returns the mid price (average of best bid and ask).
    pub fn get_mid_price(&self) -> Option<Decimal> {
        let best_bid = self.bids.get_best()?;
        let best_ask = self.asks.get_best()?;

        Some((best_bid.price + best_ask.price) / Decimal::from_i64(2))
    }

    /// Returns the spread in basis points.
    pub fn get_spread_bps(&self) -> Option<Decimal> {
        let best_bid = self.bids.get_best()?;
        let best_ask = self.asks.get_best()?;

        let spread = best_ask.price - best_bid.price;
        let mid = (best_bid.price + best_ask.price) / Decimal::from_i64(2);

        if mid.is_zero() {
            return Some(Decimal::zero());
        }

        // Spread in basis points = (spread / mid) * 10000
        Some((spread / mid) * Decimal::from_i64(10000))
    }

    /// Returns the volume-weighted mid price (microprice).
    pub fn get_micro_price(&self) -> Option<Decimal> {
        let best_bid = self.bids.get_best()?;
        let best_ask = self.asks.get_best()?;

        let bid_qty = best_bid.quantity;
        let ask_qty = best_ask.quantity;
        let total_qty = bid_qty + ask_qty;

        if total_qty.is_zero() {
            return self.get_mid_price();
        }

        // Microprice = (bid_price * ask_qty + ask_price * bid_qty) / (bid_qty + ask_qty)
        let numerator = best_bid.price * ask_qty + best_ask.price * bid_qty;
        Some(numerator / total_qty)
    }

    /// Returns the order book imbalance (-1 to 1).
    pub fn get_imbalance(&self, depth: usize) -> Decimal {
        let bid_volume = self.bids.get_total_volume(depth);
        let ask_volume = self.asks.get_total_volume(depth);

        let total_volume = bid_volume + ask_volume;
        if total_volume.is_zero() {
            return Decimal::zero();
        }

        // Imbalance = (bid_volume - ask_volume) / (bid_volume + ask_volume)
        (bid_volume - ask_volume) / total_volume
    }

    /// Clears the order book.
    pub fn clear(&mut self) {
        self.bids.clear();
        self.asks.clear();
        self.sequence_number = 0;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_order_book_operations() {
        let mut ob = OrderBook::new("BTCUSDT".to_string());

        // Add some bids and asks
        ob.bids.update(Decimal::from_i64(100), Decimal::from_i64(10), 1);
        ob.bids.update(Decimal::from_i64(99), Decimal::from_i64(20), 1);
        ob.asks.update(Decimal::from_i64(101), Decimal::from_i64(15), 1);
        ob.asks.update(Decimal::from_i64(102), Decimal::from_i64(25), 1);

        // Test mid price
        let mid = ob.get_mid_price().unwrap();
        assert_eq!(mid, Decimal::from_str("100.5").unwrap());

        // Test imbalance
        let imbalance = ob.get_imbalance(10);
        assert!(imbalance < Decimal::zero()); // More ask volume
    }
}
