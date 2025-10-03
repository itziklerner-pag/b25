package models

import "time"

// OrderSide represents buy or sell
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents the order type
type OrderType string

const (
	OrderTypeMarket     OrderType = "MARKET"
	OrderTypeLimit      OrderType = "LIMIT"
	OrderTypeStopMarket OrderType = "STOP_MARKET"
	OrderTypeStopLimit  OrderType = "STOP_LIMIT"
	OrderTypePostOnly   OrderType = "POST_ONLY"
)

// OrderState represents the order lifecycle state
type OrderState string

const (
	OrderStateNew             OrderState = "NEW"
	OrderStateSubmitted       OrderState = "SUBMITTED"
	OrderStatePartiallyFilled OrderState = "PARTIALLY_FILLED"
	OrderStateFilled          OrderState = "FILLED"
	OrderStateCanceled        OrderState = "CANCELED"
	OrderStateRejected        OrderState = "REJECTED"
	OrderStateExpired         OrderState = "EXPIRED"
)

// TimeInForce represents order time in force
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill or Kill
	TimeInForceGTX TimeInForce = "GTX" // Good Till Crossing (Post-only)
)

// Order represents a trading order
type Order struct {
	OrderID         string      `json:"order_id"`
	ClientOrderID   string      `json:"client_order_id,omitempty"`
	ExchangeOrderID string      `json:"exchange_order_id,omitempty"`
	Symbol          string      `json:"symbol"`
	Side            OrderSide   `json:"side"`
	Type            OrderType   `json:"type"`
	State           OrderState  `json:"state"`
	TimeInForce     TimeInForce `json:"time_in_force"`
	Quantity        float64     `json:"quantity"`
	Price           float64     `json:"price,omitempty"`
	StopPrice       float64     `json:"stop_price,omitempty"`
	FilledQuantity  float64     `json:"filled_quantity"`
	AveragePrice    float64     `json:"average_price"`
	Fee             float64     `json:"fee"`
	FeeAsset        string      `json:"fee_asset"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	UserID          string      `json:"user_id"`
	ReduceOnly      bool        `json:"reduce_only"`
	PostOnly        bool        `json:"post_only"`
}

// Fill represents an order fill
type Fill struct {
	FillID    string    `json:"fill_id"`
	OrderID   string    `json:"order_id"`
	Symbol    string    `json:"symbol"`
	Side      OrderSide `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Fee       float64   `json:"fee"`
	FeeAsset  string    `json:"fee_asset"`
	Timestamp time.Time `json:"timestamp"`
	IsMaker   bool      `json:"is_maker"`
}

// OrderUpdate represents an order state change
type OrderUpdate struct {
	Order      *Order    `json:"order"`
	UpdateType string    `json:"update_type"` // CREATED, UPDATED, FILLED, CANCELED, REJECTED
	Timestamp  time.Time `json:"timestamp"`
}

// StateTransition represents allowed state transitions
var StateTransitions = map[OrderState][]OrderState{
	OrderStateNew: {
		OrderStateSubmitted,
		OrderStateRejected,
	},
	OrderStateSubmitted: {
		OrderStatePartiallyFilled,
		OrderStateFilled,
		OrderStateCanceled,
		OrderStateRejected,
		OrderStateExpired,
	},
	OrderStatePartiallyFilled: {
		OrderStateFilled,
		OrderStateCanceled,
	},
}

// CanTransition checks if a state transition is valid
func (o *Order) CanTransition(newState OrderState) bool {
	allowedStates, exists := StateTransitions[o.State]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == newState {
			return true
		}
	}
	return false
}

// IsTerminal checks if the order is in a terminal state
func (o *Order) IsTerminal() bool {
	return o.State == OrderStateFilled ||
		o.State == OrderStateCanceled ||
		o.State == OrderStateRejected ||
		o.State == OrderStateExpired
}

// IsFilled checks if the order is completely filled
func (o *Order) IsFilled() bool {
	return o.FilledQuantity >= o.Quantity
}
