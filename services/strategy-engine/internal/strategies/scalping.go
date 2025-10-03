package strategies

import (
	"fmt"
	"sync"
	"time"
)

// ScalpingStrategy implements a scalping strategy
type ScalpingStrategy struct {
	*BaseStrategy

	// Configuration
	targetSpreadBps float64
	profitTarget    float64
	stopLoss        float64
	maxHoldTime     time.Duration
	maxPosition     float64

	// State
	positions   map[string]*ScalpPosition
	mu          sync.RWMutex
}

// ScalpPosition represents a scalping position
type ScalpPosition struct {
	Symbol      string
	EntryPrice  float64
	Quantity    float64
	Side        string
	EntryTime   time.Time
	TargetPrice float64
	StopPrice   float64
}

// NewScalpingStrategy creates a new scalping strategy
func NewScalpingStrategy() *ScalpingStrategy {
	return &ScalpingStrategy{
		BaseStrategy: NewBaseStrategy("scalping"),
		positions:    make(map[string]*ScalpPosition),
	}
}

// Init initializes the scalping strategy
func (s *ScalpingStrategy) Init(config map[string]interface{}) error {
	if err := s.BaseStrategy.Init(config); err != nil {
		return err
	}

	s.targetSpreadBps = s.GetConfigFloat64("target_spread_bps", 5.0)
	s.profitTarget = s.GetConfigFloat64("profit_target", 0.001) // 10 bps
	s.stopLoss = s.GetConfigFloat64("stop_loss", 0.0005) // 5 bps
	s.maxHoldTime = time.Duration(s.GetConfigInt("max_hold_time_seconds", 60)) * time.Second
	s.maxPosition = s.GetConfigFloat64("max_position", 500.0)

	s.UpdateMetric("target_spread_bps", s.targetSpreadBps)
	s.UpdateMetric("profit_target", s.profitTarget)
	s.UpdateMetric("stop_loss", s.stopLoss)
	s.UpdateMetric("max_hold_time_seconds", s.maxHoldTime.Seconds())

	return nil
}

// OnMarketData processes market data and generates signals
func (s *ScalpingStrategy) OnMarketData(data *MarketData) ([]*Signal, error) {
	if !s.IsRunning() {
		return nil, nil
	}

	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Microseconds()
		s.UpdateMetric("last_market_data_latency_us", latency)
	}()

	s.IncrementMetric("market_data_processed")

	signals := make([]*Signal, 0)

	// Check existing positions for exit conditions
	exitSignals := s.checkExitConditions(data)
	signals = append(signals, exitSignals...)

	// Generate entry signals if no position
	s.mu.RLock()
	_, hasPosition := s.positions[data.Symbol]
	s.mu.RUnlock()

	if !hasPosition {
		entrySignals := s.generateEntrySignals(data)
		signals = append(signals, entrySignals...)
	}

	if len(signals) > 0 {
		s.IncrementMetric("signals_generated")
		s.UpdateMetric("last_signal_time", time.Now().Unix())
	}

	return signals, nil
}

// OnFill handles fill events
func (s *ScalpingStrategy) OnFill(fill *Fill) error {
	if !s.IsRunning() {
		return nil
	}

	s.IncrementMetric("fills_received")
	s.UpdateMetric("last_fill_time", time.Now().Unix())

	// Track position on entry
	s.mu.Lock()
	defer s.mu.Unlock()

	if fill.Side == SideBuy {
		targetPrice := fill.Price * (1 + s.profitTarget)
		stopPrice := fill.Price * (1 - s.stopLoss)

		s.positions[fill.Symbol] = &ScalpPosition{
			Symbol:      fill.Symbol,
			EntryPrice:  fill.Price,
			Quantity:    fill.Quantity,
			Side:        SideBuy,
			EntryTime:   fill.Timestamp,
			TargetPrice: targetPrice,
			StopPrice:   stopPrice,
		}

		s.UpdateMetric(fmt.Sprintf("entry_%s", fill.Symbol), map[string]interface{}{
			"price":  fill.Price,
			"side":   SideBuy,
			"target": targetPrice,
			"stop":   stopPrice,
		})
	} else {
		// Exit position
		delete(s.positions, fill.Symbol)
		s.IncrementMetric("positions_closed")
	}

	return nil
}

// OnPositionUpdate handles position updates
func (s *ScalpingStrategy) OnPositionUpdate(position *Position) error {
	if !s.IsRunning() {
		return nil
	}

	s.IncrementMetric("position_updates")
	s.UpdateMetric(fmt.Sprintf("pnl_%s", position.Symbol), position.UnrealizedPnL+position.RealizedPnL)

	return nil
}

// generateEntrySignals generates entry signals
func (s *ScalpingStrategy) generateEntrySignals(data *MarketData) []*Signal {
	signals := make([]*Signal, 0)

	// Check if spread is tight enough for scalping
	spreadBps := CalculateSpreadBps(data.BidPrice, data.AskPrice)
	if spreadBps > s.targetSpreadBps {
		s.IncrementMetric("spread_too_wide")
		return signals
	}

	// Look for entry opportunities based on order book imbalance
	if len(data.Bids) > 0 && len(data.Asks) > 0 {
		bidVolume := 0.0
		askVolume := 0.0

		for _, level := range data.Bids {
			bidVolume += level.Quantity
		}
		for _, level := range data.Asks {
			askVolume += level.Quantity
		}

		imbalance := (bidVolume - askVolume) / (bidVolume + askVolume)
		s.UpdateMetric(fmt.Sprintf("imbalance_%s", data.Symbol), imbalance)

		// Strong bid pressure - buy signal
		if imbalance > 0.3 {
			signal := s.CreateSignal(
				data.Symbol,
				SideBuy,
				OrderTypeLimit,
				s.maxPosition,
				data.BidPrice, // Join the bid
				PriorityHigh,
			)
			signal.TimeInForce = TimeInForceIOC
			signal.Metadata["imbalance"] = imbalance
			signal.Metadata["reason"] = "bid_pressure"
			signals = append(signals, signal)
		}
	}

	return signals
}

// checkExitConditions checks if positions should be exited
func (s *ScalpingStrategy) checkExitConditions(data *MarketData) []*Signal {
	signals := make([]*Signal, 0)

	s.mu.RLock()
	position, hasPosition := s.positions[data.Symbol]
	s.mu.RUnlock()

	if !hasPosition {
		return signals
	}

	now := time.Now()
	currentPrice := data.LastPrice

	// Check profit target
	if currentPrice >= position.TargetPrice {
		signal := s.createExitSignal(position, data, "profit_target")
		signals = append(signals, signal)
		s.IncrementMetric("profit_target_exits")
	}

	// Check stop loss
	if currentPrice <= position.StopPrice {
		signal := s.createExitSignal(position, data, "stop_loss")
		signals = append(signals, signal)
		s.IncrementMetric("stop_loss_exits")
	}

	// Check max hold time
	if now.Sub(position.EntryTime) > s.maxHoldTime {
		signal := s.createExitSignal(position, data, "max_hold_time")
		signals = append(signals, signal)
		s.IncrementMetric("time_exits")
	}

	return signals
}

// createExitSignal creates an exit signal
func (s *ScalpingStrategy) createExitSignal(position *ScalpPosition, data *MarketData, reason string) *Signal {
	signal := s.CreateSignal(
		position.Symbol,
		SideSell,
		OrderTypeMarket,
		position.Quantity,
		0,
		PriorityCritical,
	)
	signal.TimeInForce = TimeInForceIOC
	signal.Metadata["reason"] = reason
	signal.Metadata["entry_price"] = position.EntryPrice
	signal.Metadata["hold_time"] = time.Since(position.EntryTime).Seconds()

	return signal
}
