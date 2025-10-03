package strategies

import (
	"fmt"
	"sync"
	"time"
)

// MomentumStrategy implements a momentum-based trading strategy
type MomentumStrategy struct {
	*BaseStrategy

	// Configuration
	lookbackPeriod int
	threshold      float64
	maxPosition    float64

	// State
	priceHistory map[string]*PriceHistory
	positions    map[string]*Position
	mu           sync.RWMutex
}

// PriceHistory stores historical price data
type PriceHistory struct {
	Prices     []float64
	Timestamps []time.Time
	MaxSize    int
}

// NewMomentumStrategy creates a new momentum strategy
func NewMomentumStrategy() *MomentumStrategy {
	return &MomentumStrategy{
		BaseStrategy: NewBaseStrategy("momentum"),
		priceHistory: make(map[string]*PriceHistory),
		positions:    make(map[string]*Position),
	}
}

// Init initializes the momentum strategy
func (s *MomentumStrategy) Init(config map[string]interface{}) error {
	if err := s.BaseStrategy.Init(config); err != nil {
		return err
	}

	s.lookbackPeriod = s.GetConfigInt("lookback_period", 20)
	s.threshold = s.GetConfigFloat64("threshold", 0.02) // 2% momentum threshold
	s.maxPosition = s.GetConfigFloat64("max_position", 1000.0)

	s.UpdateMetric("lookback_period", s.lookbackPeriod)
	s.UpdateMetric("threshold", s.threshold)
	s.UpdateMetric("max_position", s.maxPosition)

	return nil
}

// OnMarketData processes market data and generates signals
func (s *MomentumStrategy) OnMarketData(data *MarketData) ([]*Signal, error) {
	if !s.IsRunning() {
		return nil, nil
	}

	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Microseconds()
		s.UpdateMetric("last_market_data_latency_us", latency)
	}()

	s.IncrementMetric("market_data_processed")

	// Update price history
	s.updatePriceHistory(data.Symbol, data.LastPrice, data.Timestamp)

	// Calculate momentum
	momentum, err := s.calculateMomentum(data.Symbol)
	if err != nil {
		s.IncrementMetric("errors")
		return nil, err
	}

	s.UpdateMetric(fmt.Sprintf("momentum_%s", data.Symbol), momentum)

	// Generate signals based on momentum
	signals := s.generateSignals(data, momentum)

	if len(signals) > 0 {
		s.IncrementMetric("signals_generated")
		s.UpdateMetric("last_signal_time", time.Now().Unix())
	}

	return signals, nil
}

// OnFill handles fill events
func (s *MomentumStrategy) OnFill(fill *Fill) error {
	if !s.IsRunning() {
		return nil
	}

	s.IncrementMetric("fills_received")
	s.UpdateMetric("last_fill_time", time.Now().Unix())
	s.UpdateMetric(fmt.Sprintf("last_fill_%s", fill.Symbol), map[string]interface{}{
		"price":    fill.Price,
		"quantity": fill.Quantity,
		"side":     fill.Side,
	})

	return nil
}

// OnPositionUpdate handles position updates
func (s *MomentumStrategy) OnPositionUpdate(position *Position) error {
	if !s.IsRunning() {
		return nil
	}

	s.mu.Lock()
	s.positions[position.Symbol] = position
	s.mu.Unlock()

	s.IncrementMetric("position_updates")
	s.UpdateMetric(fmt.Sprintf("position_%s", position.Symbol), position.Quantity)
	s.UpdateMetric(fmt.Sprintf("pnl_%s", position.Symbol), position.UnrealizedPnL + position.RealizedPnL)

	return nil
}

// updatePriceHistory updates the price history for a symbol
func (s *MomentumStrategy) updatePriceHistory(symbol string, price float64, timestamp time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	history, exists := s.priceHistory[symbol]
	if !exists {
		history = &PriceHistory{
			Prices:     make([]float64, 0, s.lookbackPeriod),
			Timestamps: make([]time.Time, 0, s.lookbackPeriod),
			MaxSize:    s.lookbackPeriod,
		}
		s.priceHistory[symbol] = history
	}

	history.Prices = append(history.Prices, price)
	history.Timestamps = append(history.Timestamps, timestamp)

	// Keep only the last N prices
	if len(history.Prices) > history.MaxSize {
		history.Prices = history.Prices[1:]
		history.Timestamps = history.Timestamps[1:]
	}
}

// calculateMomentum calculates the momentum indicator
func (s *MomentumStrategy) calculateMomentum(symbol string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.priceHistory[symbol]
	if !exists || len(history.Prices) < 2 {
		return 0, fmt.Errorf("insufficient price history for %s", symbol)
	}

	// Simple momentum: (current_price - old_price) / old_price
	currentPrice := history.Prices[len(history.Prices)-1]
	oldPrice := history.Prices[0]

	if oldPrice == 0 {
		return 0, fmt.Errorf("invalid price history for %s", symbol)
	}

	momentum := (currentPrice - oldPrice) / oldPrice
	return momentum, nil
}

// generateSignals generates trading signals based on momentum
func (s *MomentumStrategy) generateSignals(data *MarketData, momentum float64) []*Signal {
	signals := make([]*Signal, 0)

	s.mu.RLock()
	position, hasPosition := s.positions[data.Symbol]
	s.mu.RUnlock()

	currentPosition := 0.0
	if hasPosition {
		currentPosition = position.Quantity
		if position.Side == SideSell {
			currentPosition = -currentPosition
		}
	}

	// Buy signal: strong positive momentum and not at max position
	if momentum > s.threshold && currentPosition < s.maxPosition {
		quantity := s.calculateOrderQuantity(data.Symbol, SideBuy, currentPosition)
		if quantity > 0 {
			signal := s.CreateSignal(
				data.Symbol,
				SideBuy,
				OrderTypeMarket,
				quantity,
				0,
				PriorityMedium,
			)
			signal.TimeInForce = TimeInForceIOC
			signal.Metadata["momentum"] = momentum
			signal.Metadata["reason"] = "positive_momentum"
			signals = append(signals, signal)
		}
	}

	// Sell signal: strong negative momentum and have position
	if momentum < -s.threshold && currentPosition > 0 {
		quantity := s.calculateOrderQuantity(data.Symbol, SideSell, currentPosition)
		if quantity > 0 {
			signal := s.CreateSignal(
				data.Symbol,
				SideSell,
				OrderTypeMarket,
				quantity,
				0,
				PriorityMedium,
			)
			signal.TimeInForce = TimeInForceIOC
			signal.Metadata["momentum"] = momentum
			signal.Metadata["reason"] = "negative_momentum"
			signals = append(signals, signal)
		}
	}

	return signals
}

// calculateOrderQuantity calculates the order quantity
func (s *MomentumStrategy) calculateOrderQuantity(symbol, side string, currentPosition float64) float64 {
	// Simple implementation: trade 10% of max position
	baseQuantity := s.maxPosition * 0.1

	if side == SideBuy {
		remaining := s.maxPosition - currentPosition
		if baseQuantity > remaining {
			return remaining
		}
	} else if side == SideSell {
		if baseQuantity > currentPosition {
			return currentPosition
		}
	}

	return baseQuantity
}
