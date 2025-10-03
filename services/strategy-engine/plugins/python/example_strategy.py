#!/usr/bin/env python3
"""
Example Python-based trading strategy

This is a template for creating Python-based strategies that can be executed
by the Strategy Engine. In production, this would communicate with the engine
via gRPC, ZeroMQ, or another IPC mechanism.
"""

import json
import time
from typing import List, Dict, Any, Optional
from dataclasses import dataclass, asdict
from datetime import datetime


@dataclass
class MarketData:
    """Market data structure"""
    symbol: str
    timestamp: datetime
    sequence: int
    last_price: float
    bid_price: float
    ask_price: float
    bid_size: float
    ask_size: float
    volume: float
    type: str = "tick"


@dataclass
class Signal:
    """Trading signal structure"""
    id: str
    strategy: str
    symbol: str
    side: str  # buy, sell
    order_type: str  # market, limit
    quantity: float
    price: float = 0.0
    priority: int = 5
    timestamp: datetime = None
    metadata: Dict[str, Any] = None

    def __post_init__(self):
        if self.timestamp is None:
            self.timestamp = datetime.now()
        if self.metadata is None:
            self.metadata = {}


@dataclass
class Fill:
    """Order fill structure"""
    fill_id: str
    order_id: str
    symbol: str
    side: str
    price: float
    quantity: float
    fee: float
    timestamp: datetime
    strategy: str


@dataclass
class Position:
    """Position structure"""
    symbol: str
    side: str  # long, short, flat
    quantity: float
    avg_entry_price: float
    current_price: float
    unrealized_pnl: float
    realized_pnl: float
    timestamp: datetime
    strategy: str


class ExampleStrategy:
    """
    Example momentum-based strategy in Python
    """

    def __init__(self, name: str = "python_example"):
        self.name = name
        self.running = False
        self.config = {}
        self.metrics = {}
        self.price_history: Dict[str, List[float]] = {}
        self.lookback_period = 20
        self.threshold = 0.015  # 1.5% momentum threshold

    def init(self, config: Dict[str, Any]) -> None:
        """Initialize the strategy"""
        self.config = config
        self.lookback_period = config.get('lookback_period', 20)
        self.threshold = config.get('threshold', 0.015)
        self.metrics['initialized_at'] = time.time()
        print(f"[{self.name}] Initialized with config: {config}")

    def start(self) -> None:
        """Start the strategy"""
        self.running = True
        self.metrics['started_at'] = time.time()
        print(f"[{self.name}] Started")

    def stop(self) -> None:
        """Stop the strategy"""
        self.running = False
        self.metrics['stopped_at'] = time.time()
        print(f"[{self.name}] Stopped")

    def on_market_data(self, data: MarketData) -> List[Signal]:
        """Process market data and generate signals"""
        if not self.running:
            return []

        signals = []

        # Update price history
        if data.symbol not in self.price_history:
            self.price_history[data.symbol] = []

        self.price_history[data.symbol].append(data.last_price)

        # Keep only last N prices
        if len(self.price_history[data.symbol]) > self.lookback_period:
            self.price_history[data.symbol] = \
                self.price_history[data.symbol][-self.lookback_period:]

        # Calculate momentum
        if len(self.price_history[data.symbol]) >= 2:
            momentum = self.calculate_momentum(data.symbol)

            # Generate buy signal on positive momentum
            if momentum > self.threshold:
                signal = Signal(
                    id=f"{self.name}_{data.symbol}_{int(time.time() * 1000000)}",
                    strategy=self.name,
                    symbol=data.symbol,
                    side="buy",
                    order_type="market",
                    quantity=10.0,
                    priority=5,
                    metadata={
                        "momentum": momentum,
                        "reason": "positive_momentum"
                    }
                )
                signals.append(signal)
                self.increment_metric('signals_generated')

            # Generate sell signal on negative momentum
            elif momentum < -self.threshold:
                signal = Signal(
                    id=f"{self.name}_{data.symbol}_{int(time.time() * 1000000)}",
                    strategy=self.name,
                    symbol=data.symbol,
                    side="sell",
                    order_type="market",
                    quantity=10.0,
                    priority=5,
                    metadata={
                        "momentum": momentum,
                        "reason": "negative_momentum"
                    }
                )
                signals.append(signal)
                self.increment_metric('signals_generated')

        self.increment_metric('market_data_processed')
        return signals

    def on_fill(self, fill: Fill) -> None:
        """Handle fill event"""
        if not self.running:
            return

        self.increment_metric('fills_received')
        self.metrics['last_fill_price'] = fill.price
        self.metrics['last_fill_time'] = time.time()
        print(f"[{self.name}] Fill: {fill.symbol} {fill.side} "
              f"{fill.quantity}@{fill.price}")

    def on_position_update(self, position: Position) -> None:
        """Handle position update"""
        if not self.running:
            return

        self.increment_metric('position_updates')
        self.metrics[f'position_{position.symbol}'] = position.quantity
        self.metrics[f'pnl_{position.symbol}'] = (
            position.unrealized_pnl + position.realized_pnl
        )

    def calculate_momentum(self, symbol: str) -> float:
        """Calculate momentum indicator"""
        prices = self.price_history[symbol]
        if len(prices) < 2:
            return 0.0

        current_price = prices[-1]
        old_price = prices[0]

        if old_price == 0:
            return 0.0

        return (current_price - old_price) / old_price

    def increment_metric(self, key: str) -> None:
        """Increment a metric counter"""
        self.metrics[key] = self.metrics.get(key, 0) + 1

    def get_metrics(self) -> Dict[str, Any]:
        """Get strategy metrics"""
        return self.metrics.copy()


def main():
    """
    Example main function for testing the strategy standalone
    """
    strategy = ExampleStrategy()

    config = {
        'lookback_period': 20,
        'threshold': 0.015,
    }

    strategy.init(config)
    strategy.start()

    # Simulate some market data
    for i in range(30):
        data = MarketData(
            symbol="BTCUSDT",
            timestamp=datetime.now(),
            sequence=i,
            last_price=50000 + (i * 100),
            bid_price=49995 + (i * 100),
            ask_price=50005 + (i * 100),
            bid_size=10.0,
            ask_size=10.0,
            volume=1000.0,
        )

        signals = strategy.on_market_data(data)
        if signals:
            print(f"Generated {len(signals)} signals:")
            for signal in signals:
                print(f"  {signal}")

        time.sleep(0.1)

    strategy.stop()
    print(f"\nFinal metrics: {json.dumps(strategy.get_metrics(), indent=2)}")


if __name__ == "__main__":
    main()
