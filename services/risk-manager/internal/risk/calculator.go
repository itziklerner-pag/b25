package risk

import (
	"fmt"
	"math"
)

// Calculator computes risk metrics
type Calculator struct {
	maxLeverage        float64
	maxDrawdownPercent float64
}

// NewCalculator creates a new risk calculator
func NewCalculator(maxLeverage, maxDrawdownPercent float64) *Calculator {
	return &Calculator{
		maxLeverage:        maxLeverage,
		maxDrawdownPercent: maxDrawdownPercent,
	}
}

// AccountState represents the current state of a trading account
type AccountState struct {
	Equity            float64
	Balance           float64
	UnrealizedPnL     float64
	MarginUsed        float64
	AvailableMargin   float64
	Positions         []Position
	PendingOrders     []Order
	PeakEquity        float64
	DailyStartEquity  float64
}

// Position represents a trading position
type Position struct {
	Symbol        string
	Side          string // "LONG" or "SHORT"
	Quantity      float64
	EntryPrice    float64
	CurrentPrice  float64
	UnrealizedPnL float64
	Notional      float64
	Margin        float64
}

// Order represents a pending order
type Order struct {
	OrderID    string
	Symbol     string
	Side       string // "BUY" or "SELL"
	Quantity   float64
	Price      float64
	OrderType  string // "MARKET", "LIMIT"
	StrategyID string
}

// RiskMetrics contains all calculated risk metrics
type RiskMetrics struct {
	MarginRatio          float64
	Leverage             float64
	DrawdownDaily        float64
	DrawdownMax          float64
	DailyPnL             float64
	UnrealizedPnL        float64
	TotalEquity          float64
	TotalMarginUsed      float64
	PositionConcentration map[string]float64 // symbol -> percentage
	OpenPositions        int
	PendingOrders        int
}

// CalculateMetrics computes all risk metrics for the account
func (c *Calculator) CalculateMetrics(state AccountState) RiskMetrics {
	metrics := RiskMetrics{
		TotalEquity:           state.Equity,
		TotalMarginUsed:       state.MarginUsed,
		UnrealizedPnL:         state.UnrealizedPnL,
		OpenPositions:         len(state.Positions),
		PendingOrders:         len(state.PendingOrders),
		PositionConcentration: make(map[string]float64),
	}

	// Margin ratio: equity / margin used
	if state.MarginUsed > 0 {
		metrics.MarginRatio = state.Equity / state.MarginUsed
	} else {
		metrics.MarginRatio = math.Inf(1) // Infinite margin ratio when no margin used
	}

	// Leverage: total position notional / equity
	totalNotional := c.calculateTotalNotional(state.Positions)
	if state.Equity > 0 {
		metrics.Leverage = totalNotional / state.Equity
	}

	// Daily P&L
	metrics.DailyPnL = state.Equity - state.DailyStartEquity

	// Daily drawdown
	if state.DailyStartEquity > 0 {
		metrics.DrawdownDaily = (state.DailyStartEquity - state.Equity) / state.DailyStartEquity
		if metrics.DrawdownDaily < 0 {
			metrics.DrawdownDaily = 0 // No drawdown if we're up
		}
	}

	// Max drawdown
	if state.PeakEquity > 0 {
		metrics.DrawdownMax = (state.PeakEquity - state.Equity) / state.PeakEquity
		if metrics.DrawdownMax < 0 {
			metrics.DrawdownMax = 0
		}
	}

	// Position concentration
	if state.Equity > 0 {
		for _, pos := range state.Positions {
			concentration := math.Abs(pos.Notional) / state.Equity
			metrics.PositionConcentration[pos.Symbol] = concentration
		}
	}

	return metrics
}

// SimulateOrder simulates adding an order to the account state
func (c *Calculator) SimulateOrder(state AccountState, order Order, currentPrice float64) (AccountState, error) {
	// Create a copy of the state
	simState := state
	simState.PendingOrders = append([]Order{}, state.PendingOrders...)
	simState.Positions = append([]Position{}, state.Positions...)

	// Calculate order notional
	price := order.Price
	if price == 0 {
		price = currentPrice
	}

	orderNotional := order.Quantity * price

	// Estimate margin requirement (simplified, typically 1/leverage)
	estimatedMargin := orderNotional / c.maxLeverage

	// Check if we have enough available margin
	if estimatedMargin > simState.AvailableMargin {
		return simState, fmt.Errorf("insufficient margin: need %.2f, available %.2f",
			estimatedMargin, simState.AvailableMargin)
	}

	// Update margin used
	simState.MarginUsed += estimatedMargin
	simState.AvailableMargin -= estimatedMargin

	// Add simulated position (simplified)
	existingPos := c.findPosition(simState.Positions, order.Symbol)
	if existingPos != nil {
		// Update existing position
		c.updatePosition(existingPos, order, price)
	} else {
		// Create new position
		newPos := Position{
			Symbol:       order.Symbol,
			Side:         c.orderSideToPositionSide(order.Side),
			Quantity:     order.Quantity,
			EntryPrice:   price,
			CurrentPrice: currentPrice,
			Notional:     orderNotional,
			Margin:       estimatedMargin,
		}
		simState.Positions = append(simState.Positions, newPos)
	}

	return simState, nil
}

// CalculateLimitUtilization calculates how much of each limit is being used
func (c *Calculator) CalculateLimitUtilization(metrics RiskMetrics, limits map[string]float64) map[string]float64 {
	utilization := make(map[string]float64)

	// Leverage utilization
	if maxLeverage, ok := limits["max_leverage"]; ok && maxLeverage > 0 {
		utilization["leverage"] = metrics.Leverage / maxLeverage
	}

	// Drawdown utilization
	if maxDrawdown, ok := limits["max_drawdown"]; ok && maxDrawdown > 0 {
		utilization["drawdown"] = metrics.DrawdownMax / maxDrawdown
	}

	// Margin utilization
	if minMarginRatio, ok := limits["min_margin_ratio"]; ok && minMarginRatio > 0 {
		// Higher is better for margin ratio, so invert the utilization
		utilization["margin"] = minMarginRatio / metrics.MarginRatio
	}

	return utilization
}

// calculateTotalNotional sums up all position notionals
func (c *Calculator) calculateTotalNotional(positions []Position) float64 {
	total := 0.0
	for _, pos := range positions {
		total += math.Abs(pos.Notional)
	}
	return total
}

// findPosition finds a position by symbol
func (c *Calculator) findPosition(positions []Position, symbol string) *Position {
	for i := range positions {
		if positions[i].Symbol == symbol {
			return &positions[i]
		}
	}
	return nil
}

// updatePosition updates an existing position with a new order
func (c *Calculator) updatePosition(pos *Position, order Order, price float64) {
	orderSide := c.orderSideToPositionSide(order.Side)

	if pos.Side == orderSide {
		// Same side - increase position
		totalQuantity := pos.Quantity + order.Quantity
		weightedPrice := (pos.EntryPrice*pos.Quantity + price*order.Quantity) / totalQuantity
		pos.EntryPrice = weightedPrice
		pos.Quantity = totalQuantity
		pos.Notional = totalQuantity * price
	} else {
		// Opposite side - reduce or flip position
		if order.Quantity < pos.Quantity {
			// Partial close
			pos.Quantity -= order.Quantity
			pos.Notional = pos.Quantity * price
		} else if order.Quantity == pos.Quantity {
			// Full close
			pos.Quantity = 0
			pos.Notional = 0
		} else {
			// Flip position
			pos.Quantity = order.Quantity - pos.Quantity
			pos.Side = orderSide
			pos.EntryPrice = price
			pos.Notional = pos.Quantity * price
		}
	}
}

// orderSideToPositionSide converts order side to position side
func (c *Calculator) orderSideToPositionSide(orderSide string) string {
	if orderSide == "BUY" {
		return "LONG"
	}
	return "SHORT"
}

// ValidateMetrics checks if metrics are within acceptable ranges
func (c *Calculator) ValidateMetrics(metrics RiskMetrics) []string {
	var violations []string

	// Check leverage
	if metrics.Leverage > c.maxLeverage {
		violations = append(violations, fmt.Sprintf(
			"Leverage %.2fx exceeds maximum %.2fx",
			metrics.Leverage, c.maxLeverage,
		))
	}

	// Check margin ratio
	if metrics.MarginRatio < 1.0 {
		violations = append(violations, fmt.Sprintf(
			"Margin ratio %.2f is below minimum 1.0 (insufficient margin)",
			metrics.MarginRatio,
		))
	}

	// Check drawdown
	if metrics.DrawdownMax > c.maxDrawdownPercent {
		violations = append(violations, fmt.Sprintf(
			"Max drawdown %.2f%% exceeds maximum %.2f%%",
			metrics.DrawdownMax*100, c.maxDrawdownPercent*100,
		))
	}

	return violations
}
