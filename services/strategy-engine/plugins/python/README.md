# Python Plugin Strategies

This directory contains Python-based trading strategies that can be executed by the Strategy Engine.

## Overview

Python strategies provide flexibility for rapid prototyping and testing of trading algorithms. They communicate with the Strategy Engine via IPC (Inter-Process Communication).

## Requirements

```bash
pip install -r requirements.txt
```

## Strategy Structure

Each Python strategy must implement the following methods:

- `init(config: Dict[str, Any]) -> None` - Initialize the strategy
- `start() -> None` - Start the strategy
- `stop() -> None` - Stop the strategy
- `on_market_data(data: MarketData) -> List[Signal]` - Process market data
- `on_fill(fill: Fill) -> None` - Handle order fills
- `on_position_update(position: Position) -> None` - Handle position updates
- `get_metrics() -> Dict[str, Any]` - Return strategy metrics

## Running Standalone

For testing, you can run a strategy standalone:

```bash
python example_strategy.py
```

## Running with Engine

When the Strategy Engine is configured to use Python strategies, it will:

1. Start a Python process for each strategy
2. Establish IPC communication (gRPC/ZeroMQ)
3. Send market data and receive signals
4. Manage the strategy lifecycle

## Configuration

Strategies are configured via the engine's `config.yaml`:

```yaml
strategies:
  enabled:
    - python_example
  configs:
    python_example:
      lookback_period: 20
      threshold: 0.015
  pythonPath: /usr/bin/python3
  pythonVenv: /path/to/venv
```

## Example Strategies

- `example_strategy.py` - Momentum-based strategy example
- `market_making.py` - Market making strategy example (TODO)
- `arbitrage.py` - Arbitrage strategy example (TODO)

## Performance Considerations

- Python strategies have higher latency than Go strategies
- Use for strategies that don't require sub-millisecond execution
- Consider using NumPy/Pandas for efficient data processing
- Profile your strategy to identify bottlenecks

## Development Tips

1. Test strategies standalone first
2. Use proper logging
3. Handle errors gracefully
4. Monitor memory usage
5. Use type hints for better code quality
6. Write unit tests

## IPC Communication (Production)

In production, the Strategy Engine would communicate with Python strategies using:

- **gRPC**: For low-latency, typed communication
- **ZeroMQ**: For high-performance message passing
- **Named Pipes/Sockets**: For simple local IPC

Current implementation is a placeholder. Full IPC integration requires:

1. Protocol definition (protobuf for gRPC)
2. Python server implementation
3. Go client integration in engine
4. Error handling and reconnection logic
