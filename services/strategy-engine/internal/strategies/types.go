package strategies

import (
	"time"
)

// Strategy defines the interface that all trading strategies must implement
type Strategy interface {
	// Name returns the strategy name
	Name() string

	// Init initializes the strategy with configuration
	Init(config map[string]interface{}) error

	// OnMarketData is called when new market data arrives
	OnMarketData(data *MarketData) ([]*Signal, error)

	// OnFill is called when an order is filled
	OnFill(fill *Fill) error

	// OnPositionUpdate is called when position changes
	OnPositionUpdate(position *Position) error

	// Start starts the strategy
	Start() error

	// Stop stops the strategy
	Stop() error

	// IsRunning returns whether the strategy is running
	IsRunning() bool

	// GetMetrics returns strategy metrics
	GetMetrics() map[string]interface{}
}

// MarketData represents market data
type MarketData struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Sequence  uint64    `json:"sequence"`

	// Price data
	LastPrice    float64 `json:"last_price"`
	BidPrice     float64 `json:"bid_price"`
	AskPrice     float64 `json:"ask_price"`
	BidSize      float64 `json:"bid_size"`
	AskSize      float64 `json:"ask_size"`
	Volume       float64 `json:"volume"`
	VolumeQuote  float64 `json:"volume_quote"`

	// Order book levels (top 5)
	Bids []PriceLevel `json:"bids,omitempty"`
	Asks []PriceLevel `json:"asks,omitempty"`

	// OHLCV data (if aggregated)
	Open  float64 `json:"open,omitempty"`
	High  float64 `json:"high,omitempty"`
	Low   float64 `json:"low,omitempty"`
	Close float64 `json:"close,omitempty"`

	// Additional metadata
	Type string `json:"type"` // tick, trade, book, candle
}

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// Fill represents an order fill
type Fill struct {
	FillID    string    `json:"fill_id"`
	OrderID   string    `json:"order_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"` // buy, sell
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Fee       float64   `json:"fee"`
	Timestamp time.Time `json:"timestamp"`
	Strategy  string    `json:"strategy"`
}

// Position represents a trading position
type Position struct {
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // long, short, flat
	Quantity      float64   `json:"quantity"`
	AvgEntryPrice float64   `json:"avg_entry_price"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	RealizedPnL   float64   `json:"realized_pnl"`
	Timestamp     time.Time `json:"timestamp"`
	Strategy      string    `json:"strategy"`
}

// Signal represents a trading signal
type Signal struct {
	ID        string    `json:"id"`
	Strategy  string    `json:"strategy"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"` // buy, sell
	OrderType string    `json:"order_type"` // market, limit, stop, stop_limit
	Price     float64   `json:"price,omitempty"`
	Quantity  float64   `json:"quantity"`
	StopPrice float64   `json:"stop_price,omitempty"`
	Priority  int       `json:"priority"` // 1-10, 10 is highest
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Risk parameters
	MaxSlippage float64 `json:"max_slippage,omitempty"`
	TimeInForce string  `json:"time_in_force,omitempty"` // GTC, IOC, FOK
}

// SignalPriority constants
const (
	PriorityLow      = 1
	PriorityMedium   = 5
	PriorityHigh     = 8
	PriorityCritical = 10
)

// OrderSide constants
const (
	SideBuy  = "buy"
	SideSell = "sell"
)

// OrderType constants
const (
	OrderTypeMarket     = "market"
	OrderTypeLimit      = "limit"
	OrderTypeStop       = "stop"
	OrderTypeStopLimit  = "stop_limit"
)

// TimeInForce constants
const (
	TimeInForceGTC = "GTC" // Good Till Cancel
	TimeInForceIOC = "IOC" // Immediate Or Cancel
	TimeInForceFOK = "FOK" // Fill Or Kill
)

// StrategyType constants
const (
	StrategyTypeMomentum     = "momentum"
	StrategyTypeMarketMaking = "market_making"
	StrategyTypeScalping     = "scalping"
	StrategyTypeArbitrage    = "arbitrage"
	StrategyTypeMeanReversion = "mean_reversion"
)
