package types

import (
	"time"
)

// ClientType represents the type of client connected
type ClientType int

const (
	ClientTypeTUI ClientType = iota // 100ms updates
	ClientTypeWeb                    // 250ms updates
)

func (ct ClientType) String() string {
	switch ct {
	case ClientTypeTUI:
		return "TUI"
	case ClientTypeWeb:
		return "Web"
	default:
		return "Unknown"
	}
}

// SerializationFormat represents the message serialization format
type SerializationFormat int

const (
	FormatMessagePack SerializationFormat = iota
	FormatJSON
)

func (sf SerializationFormat) String() string {
	switch sf {
	case FormatMessagePack:
		return "MessagePack"
	case FormatJSON:
		return "JSON"
	default:
		return "Unknown"
	}
}

// State represents the aggregated state from all services
type State struct {
	MarketData map[string]*MarketData `msgpack:"market_data" json:"market_data"`
	Orders     []*Order               `msgpack:"orders" json:"orders"`
	Positions  map[string]*Position   `msgpack:"positions" json:"positions"`
	Account    *Account               `msgpack:"account" json:"account"`
	Strategies map[string]*Strategy   `msgpack:"strategies" json:"strategies"`
	Timestamp  time.Time              `msgpack:"timestamp" json:"timestamp"`
	Sequence   uint64                 `msgpack:"seq" json:"seq"`
}

// MarketData represents market data for a symbol
type MarketData struct {
	Symbol    string  `msgpack:"symbol" json:"symbol"`
	LastPrice float64 `msgpack:"last_price" json:"last_price"`
	BidPrice  float64 `msgpack:"bid_price" json:"bid_price"`
	AskPrice  float64 `msgpack:"ask_price" json:"ask_price"`
	Volume24h float64 `msgpack:"volume_24h" json:"volume_24h"`
	High24h   float64 `msgpack:"high_24h" json:"high_24h"`
	Low24h    float64 `msgpack:"low_24h" json:"low_24h"`
	UpdatedAt time.Time `msgpack:"updated_at" json:"updated_at"`
}

// Order represents a trading order
type Order struct {
	ID         string    `msgpack:"id" json:"id"`
	Symbol     string    `msgpack:"symbol" json:"symbol"`
	Side       string    `msgpack:"side" json:"side"`
	Type       string    `msgpack:"type" json:"type"`
	Price      float64   `msgpack:"price" json:"price"`
	Quantity   float64   `msgpack:"quantity" json:"quantity"`
	Filled     float64   `msgpack:"filled" json:"filled"`
	Status     string    `msgpack:"status" json:"status"`
	CreatedAt  time.Time `msgpack:"created_at" json:"created_at"`
	UpdatedAt  time.Time `msgpack:"updated_at" json:"updated_at"`
}

// Position represents a trading position
type Position struct {
	Symbol           string  `msgpack:"symbol" json:"symbol"`
	Side             string  `msgpack:"side" json:"side"`
	Quantity         float64 `msgpack:"quantity" json:"quantity"`
	EntryPrice       float64 `msgpack:"entry_price" json:"entry_price"`
	MarkPrice        float64 `msgpack:"mark_price" json:"mark_price"`
	UnrealizedPnL    float64 `msgpack:"unrealized_pnl" json:"unrealized_pnl"`
	RealizedPnL      float64 `msgpack:"realized_pnl" json:"realized_pnl"`
	LiquidationPrice float64 `msgpack:"liquidation_price" json:"liquidation_price"`
	UpdatedAt        time.Time `msgpack:"updated_at" json:"updated_at"`
}

// Account represents account information
type Account struct {
	TotalBalance     float64            `msgpack:"total_balance" json:"total_balance"`
	AvailableBalance float64            `msgpack:"available_balance" json:"available_balance"`
	MarginUsed       float64            `msgpack:"margin_used" json:"margin_used"`
	UnrealizedPnL    float64            `msgpack:"unrealized_pnl" json:"unrealized_pnl"`
	Balances         map[string]float64 `msgpack:"balances" json:"balances"`
	UpdatedAt        time.Time          `msgpack:"updated_at" json:"updated_at"`
}

// Strategy represents a trading strategy
type Strategy struct {
	ID        string  `msgpack:"id" json:"id"`
	Name      string  `msgpack:"name" json:"name"`
	Status    string  `msgpack:"status" json:"status"`
	PnL       float64 `msgpack:"pnl" json:"pnl"`
	Trades    int     `msgpack:"trades" json:"trades"`
	WinRate   float64 `msgpack:"win_rate" json:"win_rate"`
	UpdatedAt time.Time `msgpack:"updated_at" json:"updated_at"`
}

// ClientMessage represents messages sent from client to server
type ClientMessage struct {
	Type     string   `json:"type"`
	Channels []string `json:"channels,omitempty"`
}

// ServerMessage represents messages sent from server to client
type ServerMessage struct {
	Type      string                 `msgpack:"type" json:"type"`
	Sequence  uint64                 `msgpack:"seq" json:"seq"`
	Timestamp time.Time              `msgpack:"timestamp" json:"timestamp"`
	Data      *State                 `msgpack:"data,omitempty" json:"data,omitempty"`
	Changes   map[string]interface{} `msgpack:"changes,omitempty" json:"changes,omitempty"`
	Error     string                 `msgpack:"error,omitempty" json:"error,omitempty"`
}
