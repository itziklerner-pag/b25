// +build plugin

package main

import (
	"time"

	"github.com/b25/strategy-engine/internal/strategies"
)

// ExamplePluginStrategy is an example plugin-based strategy
type ExamplePluginStrategy struct {
	*strategies.BaseStrategy
	threshold float64
}

// NewStrategy is the required entry point for the plugin
func NewStrategy() strategies.Strategy {
	return &ExamplePluginStrategy{
		BaseStrategy: strategies.NewBaseStrategy("example_plugin"),
	}
}

// Init initializes the strategy
func (s *ExamplePluginStrategy) Init(config map[string]interface{}) error {
	if err := s.BaseStrategy.Init(config); err != nil {
		return err
	}

	s.threshold = s.GetConfigFloat64("threshold", 0.01)
	return nil
}

// OnMarketData processes market data
func (s *ExamplePluginStrategy) OnMarketData(data *strategies.MarketData) ([]*strategies.Signal, error) {
	if !s.IsRunning() {
		return nil, nil
	}

	s.IncrementMetric("market_data_processed")

	signals := make([]*strategies.Signal, 0)

	// Example logic: buy when spread is tight
	spread := strategies.CalculateSpreadBps(data.BidPrice, data.AskPrice)
	if spread < s.threshold*10000 {
		signal := s.CreateSignal(
			data.Symbol,
			strategies.SideBuy,
			strategies.OrderTypeLimit,
			10.0,
			data.BidPrice,
			strategies.PriorityMedium,
		)
		signal.TimeInForce = strategies.TimeInForceGTC
		signal.Metadata["spread_bps"] = spread
		signals = append(signals, signal)
	}

	return signals, nil
}

// OnFill handles fill events
func (s *ExamplePluginStrategy) OnFill(fill *strategies.Fill) error {
	if !s.IsRunning() {
		return nil
	}

	s.IncrementMetric("fills_received")
	s.UpdateMetric("last_fill_time", time.Now().Unix())

	return nil
}

// OnPositionUpdate handles position updates
func (s *ExamplePluginStrategy) OnPositionUpdate(position *strategies.Position) error {
	if !s.IsRunning() {
		return nil
	}

	s.IncrementMetric("position_updates")
	s.UpdateMetric("current_position", position.Quantity)

	return nil
}

// Build instructions:
// go build -buildmode=plugin -o example_plugin.so example_plugin.go
