# Go Plugin Strategies

This directory contains Go plugin-based trading strategies that can be loaded dynamically by the Strategy Engine.

## Building Plugins

To build a plugin strategy:

```bash
go build -buildmode=plugin -o <name>.so <name>.go
```

Example:
```bash
go build -buildmode=plugin -o example_plugin.so example_plugin.go
```

## Plugin Interface

Each plugin must implement the `strategies.Strategy` interface and export a `NewStrategy` function:

```go
func NewStrategy() strategies.Strategy {
    return &YourStrategy{
        BaseStrategy: strategies.NewBaseStrategy("your_strategy_name"),
    }
}
```

## Required Methods

- `Name() string` - Returns strategy name
- `Init(config map[string]interface{}) error` - Initializes the strategy
- `OnMarketData(data *MarketData) ([]*Signal, error)` - Processes market data
- `OnFill(fill *Fill) error` - Handles order fills
- `OnPositionUpdate(position *Position) error` - Handles position updates
- `Start() error` - Starts the strategy
- `Stop() error` - Stops the strategy
- `IsRunning() bool` - Returns running status
- `GetMetrics() map[string]interface{}` - Returns strategy metrics

## Hot Reloading

When hot reload is enabled in the engine configuration, new plugins placed in this directory will be automatically loaded without restarting the engine.

**Note:** Go plugins cannot be unloaded. Modified plugins require a full engine restart.

## Example Strategies

- `example_plugin.so` - A simple example plugin strategy
