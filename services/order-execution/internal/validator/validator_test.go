package validator

import (
	"testing"

	"github.com/yourusername/b25/services/order-execution/internal/models"
)

func TestValidateQuantity(t *testing.T) {
	symbolInfo := &SymbolInfo{
		Symbol:            "BTCUSDT",
		PricePrecision:    2,
		QuantityPrecision: 3,
		MinNotional:       10.0,
		MinQuantity:       0.001,
		MaxQuantity:       100.0,
		TickSize:          0.01,
		StepSize:          0.001,
	}

	riskLimits := &RiskLimits{
		MaxOrderValue:   1000000,
		MaxPositionSize: 10,
		MaxDailyOrders:  10000,
		MaxOpenOrders:   500,
		AllowedSymbols: map[string]bool{
			"BTCUSDT": true,
		},
	}

	v := NewValidator(riskLimits)
	v.RegisterSymbol(symbolInfo)

	tests := []struct {
		name    string
		order   *models.Order
		wantErr bool
	}{
		{
			name: "valid limit order",
			order: &models.Order{
				Symbol:      "BTCUSDT",
				Side:        models.OrderSideBuy,
				Type:        models.OrderTypeLimit,
				Quantity:    0.001,
				Price:       45000.00,
				TimeInForce: models.TimeInForceGTC,
			},
			wantErr: false,
		},
		{
			name: "quantity too small",
			order: &models.Order{
				Symbol:      "BTCUSDT",
				Side:        models.OrderSideBuy,
				Type:        models.OrderTypeLimit,
				Quantity:    0.0001,
				Price:       45000.00,
				TimeInForce: models.TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "quantity too large",
			order: &models.Order{
				Symbol:      "BTCUSDT",
				Side:        models.OrderSideBuy,
				Type:        models.OrderTypeLimit,
				Quantity:    101.0,
				Price:       45000.00,
				TimeInForce: models.TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "invalid price precision",
			order: &models.Order{
				Symbol:      "BTCUSDT",
				Side:        models.OrderSideBuy,
				Type:        models.OrderTypeLimit,
				Quantity:    0.001,
				Price:       45000.123,
				TimeInForce: models.TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "post only with gtx",
			order: &models.Order{
				Symbol:      "BTCUSDT",
				Side:        models.OrderSideBuy,
				Type:        models.OrderTypePostOnly,
				Quantity:    0.001,
				Price:       45000.00,
				TimeInForce: models.TimeInForceGTX,
				PostOnly:    true,
			},
			wantErr: false,
		},
		{
			name: "post only with invalid tif",
			order: &models.Order{
				Symbol:      "BTCUSDT",
				Side:        models.OrderSideBuy,
				Type:        models.OrderTypePostOnly,
				Quantity:    0.001,
				Price:       45000.00,
				TimeInForce: models.TimeInForceIOC,
				PostOnly:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateOrder(tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPrecision(t *testing.T) {
	tests := []struct {
		value    float64
		expected int
	}{
		{45000.0, 0},
		{45000.1, 1},
		{45000.12, 2},
		{0.001, 3},
		{0.0001, 4},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := getPrecision(tt.value)
			if result != tt.expected {
				t.Errorf("getPrecision(%f) = %d, want %d", tt.value, result, tt.expected)
			}
		})
	}
}
