package strategies

import (
	"fmt"
	"sync"
	"time"
)

// MarketMakingStrategy implements a market making strategy
type MarketMakingStrategy struct {
	*BaseStrategy

	// Configuration
	spread         float64
	orderSize      float64
	maxInventory   float64
	inventorySkew  float64
	minSpreadBps   float64

	// State
	positions      map[string]*Position
	activeOrders   map[string]bool
	mu             sync.RWMutex
}

// NewMarketMakingStrategy creates a new market making strategy
func NewMarketMakingStrategy() *MarketMakingStrategy {
	return &MarketMakingStrategy{
		BaseStrategy: NewBaseStrategy("market_making"),
		positions:    make(map[string]*Position),
		activeOrders: make(map[string]bool),
	}
}

// Init initializes the market making strategy
func (s *MarketMakingStrategy) Init(config map[string]interface{}) error {
	if err := s.BaseStrategy.Init(config); err != nil {
		return err
	}

	s.spread = s.GetConfigFloat64("spread", 0.001) // 10 bps default
	s.orderSize = s.GetConfigFloat64("order_size", 100.0)
	s.maxInventory = s.GetConfigFloat64("max_inventory", 1000.0)
	s.inventorySkew = s.GetConfigFloat64("inventory_skew", 0.5)
	s.minSpreadBps = s.GetConfigFloat64("min_spread_bps", 5.0)

	s.UpdateMetric("spread", s.spread)
	s.UpdateMetric("order_size", s.orderSize)
	s.UpdateMetric("max_inventory", s.maxInventory)

	return nil
}

// OnMarketData processes market data and generates signals
func (s *MarketMakingStrategy) OnMarketData(data *MarketData) ([]*Signal, error) {
	if !s.IsRunning() {
		return nil, nil
	}

	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Microseconds()
		s.UpdateMetric("last_market_data_latency_us", latency)
	}()

	s.IncrementMetric("market_data_processed")

	// Check if spread is wide enough
	currentSpreadBps := CalculateSpreadBps(data.BidPrice, data.AskPrice)
	if currentSpreadBps < s.minSpreadBps {
		s.IncrementMetric("spread_too_narrow")
		return nil, nil
	}

	// Generate market making quotes
	signals := s.generateQuotes(data)

	if len(signals) > 0 {
		s.IncrementMetric("signals_generated")
		s.UpdateMetric("last_signal_time", time.Now().Unix())
	}

	return signals, nil
}

// OnFill handles fill events
func (s *MarketMakingStrategy) OnFill(fill *Fill) error {
	if !s.IsRunning() {
		return nil
	}

	s.IncrementMetric("fills_received")
	s.UpdateMetric("last_fill_time", time.Now().Unix())

	// Track realized PnL
	s.IncrementMetric(fmt.Sprintf("volume_%s", fill.Symbol))

	return nil
}

// OnPositionUpdate handles position updates
func (s *MarketMakingStrategy) OnPositionUpdate(position *Position) error {
	if !s.IsRunning() {
		return nil
	}

	s.mu.Lock()
	s.positions[position.Symbol] = position
	s.mu.Unlock()

	s.IncrementMetric("position_updates")
	s.UpdateMetric(fmt.Sprintf("position_%s", position.Symbol), position.Quantity)
	s.UpdateMetric(fmt.Sprintf("inventory_%s", position.Symbol), position.Quantity)

	return nil
}

// generateQuotes generates bid and ask quotes
func (s *MarketMakingStrategy) generateQuotes(data *MarketData) []*Signal {
	signals := make([]*Signal, 0, 2)

	s.mu.RLock()
	position, hasPosition := s.positions[data.Symbol]
	s.mu.RUnlock()

	currentInventory := 0.0
	if hasPosition {
		currentInventory = position.Quantity
		if position.Side == SideSell {
			currentInventory = -currentInventory
		}
	}

	// Calculate inventory skew
	inventoryRatio := currentInventory / s.maxInventory
	skew := s.inventorySkew * inventoryRatio

	// Calculate mid price
	midPrice := CalculateMidPrice(data.BidPrice, data.AskPrice)

	// Calculate bid and ask prices with inventory skew
	bidOffset := s.spread * (1 + skew)
	askOffset := s.spread * (1 - skew)

	bidPrice := midPrice * (1 - bidOffset)
	askPrice := midPrice * (1 + askOffset)

	// Generate bid signal (only if not at max inventory)
	if currentInventory < s.maxInventory {
		bidSignal := s.CreateSignal(
			data.Symbol,
			SideBuy,
			OrderTypeLimit,
			s.orderSize,
			bidPrice,
			PriorityMedium,
		)
		bidSignal.TimeInForce = TimeInForceGTC
		bidSignal.Metadata["side"] = "bid"
		bidSignal.Metadata["mid_price"] = midPrice
		bidSignal.Metadata["inventory"] = currentInventory
		signals = append(signals, bidSignal)
	}

	// Generate ask signal (only if have inventory or can short)
	if currentInventory > -s.maxInventory {
		askSignal := s.CreateSignal(
			data.Symbol,
			SideSell,
			OrderTypeLimit,
			s.orderSize,
			askPrice,
			PriorityMedium,
		)
		askSignal.TimeInForce = TimeInForceGTC
		askSignal.Metadata["side"] = "ask"
		askSignal.Metadata["mid_price"] = midPrice
		askSignal.Metadata["inventory"] = currentInventory
		signals = append(signals, askSignal)
	}

	s.UpdateMetric(fmt.Sprintf("bid_price_%s", data.Symbol), bidPrice)
	s.UpdateMetric(fmt.Sprintf("ask_price_%s", data.Symbol), askPrice)
	s.UpdateMetric(fmt.Sprintf("spread_bps_%s", data.Symbol), (askPrice-bidPrice)/midPrice*10000)

	return signals
}
