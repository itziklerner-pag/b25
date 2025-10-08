package validator

import (
	"fmt"
	"math"

	"github.com/yourusername/b25/services/order-execution/internal/models"
)

// SymbolInfo represents exchange symbol information
type SymbolInfo struct {
	Symbol           string
	PricePrecision   int
	QuantityPrecision int
	MinNotional      float64
	MinQuantity      float64
	MaxQuantity      float64
	TickSize         float64
	StepSize         float64
}

// RiskLimits represents risk management limits
type RiskLimits struct {
	MaxOrderValue    float64
	MaxPositionSize  float64
	MaxDailyOrders   int
	MaxOpenOrders    int
	AllowedSymbols   map[string]bool
}

// Validator validates orders against exchange rules and risk limits
type Validator struct {
	symbolInfo map[string]*SymbolInfo
	riskLimits *RiskLimits
}

// NewValidator creates a new order validator
func NewValidator(riskLimits *RiskLimits) *Validator {
	return &Validator{
		symbolInfo: make(map[string]*SymbolInfo),
		riskLimits: riskLimits,
	}
}

// RegisterSymbol adds symbol information for validation
func (v *Validator) RegisterSymbol(info *SymbolInfo) {
	v.symbolInfo[info.Symbol] = info
}

// GetRiskLimits returns the current risk limits
func (v *Validator) GetRiskLimits() *RiskLimits {
	return v.riskLimits
}

// ValidateOrder validates an order against all rules
func (v *Validator) ValidateOrder(order *models.Order) error {
	// Check symbol is registered
	symbolInfo, exists := v.symbolInfo[order.Symbol]
	if !exists {
		return fmt.Errorf("symbol %s not registered", order.Symbol)
	}

	// Check allowed symbols
	if v.riskLimits.AllowedSymbols != nil {
		if !v.riskLimits.AllowedSymbols[order.Symbol] {
			return fmt.Errorf("symbol %s not allowed", order.Symbol)
		}
	}

	// Validate quantity
	if err := v.validateQuantity(order, symbolInfo); err != nil {
		return err
	}

	// Validate price
	if err := v.validatePrice(order, symbolInfo); err != nil {
		return err
	}

	// Validate notional value
	if err := v.validateNotional(order, symbolInfo); err != nil {
		return err
	}

	// Validate order value against risk limits
	if err := v.validateRiskLimits(order); err != nil {
		return err
	}

	// Validate time in force
	if err := v.validateTimeInForce(order); err != nil {
		return err
	}

	return nil
}

// validateQuantity validates order quantity
func (v *Validator) validateQuantity(order *models.Order, info *SymbolInfo) error {
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	if order.Quantity < info.MinQuantity {
		return fmt.Errorf("quantity %.8f below minimum %.8f", order.Quantity, info.MinQuantity)
	}

	if order.Quantity > info.MaxQuantity {
		return fmt.Errorf("quantity %.8f above maximum %.8f", order.Quantity, info.MaxQuantity)
	}

	// Check step size precision
	if info.StepSize > 0 {
		remainder := math.Mod(order.Quantity, info.StepSize)
		if remainder > 1e-8 { // Floating point tolerance
			return fmt.Errorf("quantity %.8f not a multiple of step size %.8f", order.Quantity, info.StepSize)
		}
	}

	// Check precision
	precision := getPrecision(order.Quantity)
	if precision > info.QuantityPrecision {
		return fmt.Errorf("quantity precision %d exceeds maximum %d", precision, info.QuantityPrecision)
	}

	return nil
}

// validatePrice validates order price
func (v *Validator) validatePrice(order *models.Order, info *SymbolInfo) error {
	// Market orders don't need price validation
	if order.Type == models.OrderTypeMarket || order.Type == models.OrderTypeStopMarket {
		return nil
	}

	if order.Price <= 0 {
		return fmt.Errorf("price must be positive for limit orders")
	}

	// Check tick size
	if info.TickSize > 0 {
		remainder := math.Mod(order.Price, info.TickSize)
		if remainder > 1e-8 { // Floating point tolerance
			return fmt.Errorf("price %.8f not a multiple of tick size %.8f", order.Price, info.TickSize)
		}
	}

	// Check precision
	precision := getPrecision(order.Price)
	if precision > info.PricePrecision {
		return fmt.Errorf("price precision %d exceeds maximum %d", precision, info.PricePrecision)
	}

	return nil
}

// validateNotional validates minimum notional value
func (v *Validator) validateNotional(order *models.Order, info *SymbolInfo) error {
	var notional float64

	switch order.Type {
	case models.OrderTypeMarket, models.OrderTypeStopMarket:
		// For market orders, we can't validate notional without current price
		// This should be checked by the executor with real-time price
		return nil
	case models.OrderTypeLimit, models.OrderTypePostOnly:
		notional = order.Price * order.Quantity
	case models.OrderTypeStopLimit:
		notional = order.Price * order.Quantity
	default:
		return fmt.Errorf("unsupported order type: %s", order.Type)
	}

	if notional < info.MinNotional {
		return fmt.Errorf("order notional %.2f below minimum %.2f", notional, info.MinNotional)
	}

	return nil
}

// validateRiskLimits validates against risk management limits
func (v *Validator) validateRiskLimits(order *models.Order) error {
	var orderValue float64

	// Calculate order value
	if order.Type == models.OrderTypeMarket || order.Type == models.OrderTypeStopMarket {
		// For market orders, use a conservative estimate
		// This will be validated again with actual execution price
		return nil
	} else {
		orderValue = order.Price * order.Quantity
	}

	if orderValue > v.riskLimits.MaxOrderValue {
		return fmt.Errorf("order value %.2f exceeds maximum %.2f", orderValue, v.riskLimits.MaxOrderValue)
	}

	return nil
}

// validateTimeInForce validates time in force is compatible with order type
func (v *Validator) validateTimeInForce(order *models.Order) error {
	// POST_ONLY orders must use GTX or GTC
	if order.PostOnly || order.Type == models.OrderTypePostOnly {
		if order.TimeInForce != models.TimeInForceGTX && order.TimeInForce != models.TimeInForceGTC {
			return fmt.Errorf("post-only orders must use GTX or GTC time in force")
		}
	}

	// Market orders typically use IOC or FOK
	if order.Type == models.OrderTypeMarket {
		if order.TimeInForce == models.TimeInForceGTC || order.TimeInForce == models.TimeInForceGTX {
			return fmt.Errorf("market orders cannot use GTC or GTX time in force")
		}
	}

	return nil
}

// getPrecision returns the decimal precision of a float
func getPrecision(val float64) int {
	str := fmt.Sprintf("%.8f", val)
	// Remove trailing zeros
	for len(str) > 0 && str[len(str)-1] == '0' {
		str = str[:len(str)-1]
	}
	// Find decimal point
	dotIndex := -1
	for i, c := range str {
		if c == '.' {
			dotIndex = i
			break
		}
	}
	if dotIndex == -1 {
		return 0
	}
	return len(str) - dotIndex - 1
}

// ValidatePositionSize validates position size against limits
func (v *Validator) ValidatePositionSize(symbol string, currentSize, orderQuantity float64, side models.OrderSide) error {
	// Calculate new position size
	var newSize float64
	if side == models.OrderSideBuy {
		newSize = currentSize + orderQuantity
	} else {
		newSize = currentSize - orderQuantity
	}

	absNewSize := math.Abs(newSize)
	if absNewSize > v.riskLimits.MaxPositionSize {
		return fmt.Errorf("new position size %.8f would exceed maximum %.8f", absNewSize, v.riskLimits.MaxPositionSize)
	}

	return nil
}

// ValidateOrderCount validates order count against daily limits
func (v *Validator) ValidateOrderCount(dailyOrders, openOrders int) error {
	if dailyOrders >= v.riskLimits.MaxDailyOrders {
		return fmt.Errorf("daily order limit %d reached", v.riskLimits.MaxDailyOrders)
	}

	if openOrders >= v.riskLimits.MaxOpenOrders {
		return fmt.Errorf("open order limit %d reached", v.riskLimits.MaxOpenOrders)
	}

	return nil
}
