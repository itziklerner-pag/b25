package validator

import (
	"encoding/json"
	"testing"

	"github.com/b25/services/configuration/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidateStrategyConfig(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name: "valid strategy config",
			value: `{
				"name": "Market Making",
				"type": "market_making",
				"enabled": true,
				"parameters": {"spread": 0.002}
			}`,
			wantErr: false,
		},
		{
			name: "missing name",
			value: `{
				"type": "market_making",
				"enabled": true
			}`,
			wantErr: true,
		},
		{
			name: "invalid strategy type",
			value: `{
				"name": "Test",
				"type": "invalid_type",
				"enabled": true
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(
				domain.ConfigTypeStrategy,
				domain.ConfigFormatJSON,
				json.RawMessage(tt.value),
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRiskLimitConfig(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name: "valid risk limit config",
			value: `{
				"max_position_size": 10000,
				"max_loss_per_trade": 500,
				"max_daily_loss": 2000,
				"max_leverage": 10,
				"stop_loss_percent": 5
			}`,
			wantErr: false,
		},
		{
			name: "invalid max_position_size",
			value: `{
				"max_position_size": 0,
				"max_loss_per_trade": 500,
				"max_daily_loss": 2000,
				"max_leverage": 10,
				"stop_loss_percent": 5
			}`,
			wantErr: true,
		},
		{
			name: "invalid leverage",
			value: `{
				"max_position_size": 10000,
				"max_loss_per_trade": 500,
				"max_daily_loss": 2000,
				"max_leverage": 150,
				"stop_loss_percent": 5
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(
				domain.ConfigTypeRiskLimit,
				domain.ConfigFormatJSON,
				json.RawMessage(tt.value),
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTradingPairConfig(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name: "valid trading pair config",
			value: `{
				"symbol": "BTC/USDT",
				"base_currency": "BTC",
				"quote_currency": "USDT",
				"min_order_size": 0.001,
				"max_order_size": 10,
				"price_precision": 2,
				"quantity_precision": 8,
				"enabled": true
			}`,
			wantErr: false,
		},
		{
			name: "missing symbol",
			value: `{
				"base_currency": "BTC",
				"quote_currency": "USDT",
				"min_order_size": 0.001,
				"max_order_size": 10,
				"price_precision": 2,
				"quantity_precision": 8,
				"enabled": true
			}`,
			wantErr: true,
		},
		{
			name: "min_order_size greater than max_order_size",
			value: `{
				"symbol": "BTC/USDT",
				"base_currency": "BTC",
				"quote_currency": "USDT",
				"min_order_size": 10,
				"max_order_size": 1,
				"price_precision": 2,
				"quantity_precision": 8,
				"enabled": true
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(
				domain.ConfigTypeTradingPair,
				domain.ConfigFormatJSON,
				json.RawMessage(tt.value),
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		format  domain.ConfigFormat
		value   string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			format:  domain.ConfigFormatJSON,
			value:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			format:  domain.ConfigFormatJSON,
			value:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "valid YAML",
			format:  domain.ConfigFormatYAML,
			value:   "key: value\nkey2: value2",
			wantErr: false,
		},
		{
			name:    "invalid YAML",
			format:  domain.ConfigFormatYAML,
			value:   ":\ninvalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFormat(tt.format, json.RawMessage(tt.value))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
