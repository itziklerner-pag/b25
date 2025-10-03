package validator

import (
	"encoding/json"
	"fmt"

	"github.com/b25/services/configuration/internal/domain"
	"gopkg.in/yaml.v3"
)

// Validator handles validation of configuration values
type Validator struct {
	// Add custom validators for specific config types
	customValidators map[domain.ConfigType]func(value json.RawMessage) error
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := &Validator{
		customValidators: make(map[domain.ConfigType]func(value json.RawMessage) error),
	}
	v.registerDefaultValidators()
	return v
}

// registerDefaultValidators registers built-in validators
func (v *Validator) registerDefaultValidators() {
	v.customValidators[domain.ConfigTypeStrategy] = v.validateStrategyConfig
	v.customValidators[domain.ConfigTypeRiskLimit] = v.validateRiskLimitConfig
	v.customValidators[domain.ConfigTypeTradingPair] = v.validateTradingPairConfig
	v.customValidators[domain.ConfigTypeSystem] = v.validateSystemConfig
}

// Validate validates a configuration value
func (v *Validator) Validate(configType domain.ConfigType, format domain.ConfigFormat, value json.RawMessage) error {
	// Validate format syntax
	if err := v.validateFormat(format, value); err != nil {
		return fmt.Errorf("format validation failed: %w", err)
	}

	// Run type-specific validation
	if validator, exists := v.customValidators[configType]; exists {
		if err := validator(value); err != nil {
			return fmt.Errorf("type validation failed: %w", err)
		}
	}

	return nil
}

// validateFormat validates the syntax of the configuration format
func (v *Validator) validateFormat(format domain.ConfigFormat, value json.RawMessage) error {
	switch format {
	case domain.ConfigFormatJSON:
		var temp interface{}
		if err := json.Unmarshal(value, &temp); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	case domain.ConfigFormatYAML:
		var temp interface{}
		if err := yaml.Unmarshal(value, &temp); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
	default:
		return domain.ErrInvalidFormat
	}
	return nil
}

// validateStrategyConfig validates strategy configuration
func (v *Validator) validateStrategyConfig(value json.RawMessage) error {
	var config struct {
		Name        string  `json:"name"`
		Type        string  `json:"type"`
		Enabled     bool    `json:"enabled"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := json.Unmarshal(value, &config); err != nil {
		return fmt.Errorf("invalid strategy config structure: %w", err)
	}

	if config.Name == "" {
		return domain.NewValidationError("name", "strategy name is required")
	}

	if config.Type == "" {
		return domain.NewValidationError("type", "strategy type is required")
	}

	// Validate strategy type
	validTypes := map[string]bool{
		"market_making": true,
		"arbitrage":     true,
		"momentum":      true,
		"mean_reversion": true,
	}

	if !validTypes[config.Type] {
		return domain.NewValidationError("type", fmt.Sprintf("invalid strategy type: %s", config.Type))
	}

	return nil
}

// validateRiskLimitConfig validates risk limit configuration
func (v *Validator) validateRiskLimitConfig(value json.RawMessage) error {
	var config struct {
		MaxPositionSize  float64 `json:"max_position_size"`
		MaxLossPerTrade  float64 `json:"max_loss_per_trade"`
		MaxDailyLoss     float64 `json:"max_daily_loss"`
		MaxLeverage      float64 `json:"max_leverage"`
		StopLossPercent  float64 `json:"stop_loss_percent"`
	}

	if err := json.Unmarshal(value, &config); err != nil {
		return fmt.Errorf("invalid risk limit config structure: %w", err)
	}

	if config.MaxPositionSize <= 0 {
		return domain.NewValidationError("max_position_size", "must be greater than 0")
	}

	if config.MaxLossPerTrade <= 0 {
		return domain.NewValidationError("max_loss_per_trade", "must be greater than 0")
	}

	if config.MaxDailyLoss <= 0 {
		return domain.NewValidationError("max_daily_loss", "must be greater than 0")
	}

	if config.MaxLeverage <= 0 || config.MaxLeverage > 100 {
		return domain.NewValidationError("max_leverage", "must be between 0 and 100")
	}

	if config.StopLossPercent <= 0 || config.StopLossPercent > 100 {
		return domain.NewValidationError("stop_loss_percent", "must be between 0 and 100")
	}

	return nil
}

// validateTradingPairConfig validates trading pair configuration
func (v *Validator) validateTradingPairConfig(value json.RawMessage) error {
	var config struct {
		Symbol          string  `json:"symbol"`
		BaseCurrency    string  `json:"base_currency"`
		QuoteCurrency   string  `json:"quote_currency"`
		MinOrderSize    float64 `json:"min_order_size"`
		MaxOrderSize    float64 `json:"max_order_size"`
		PricePrecision  int     `json:"price_precision"`
		QuantityPrecision int   `json:"quantity_precision"`
		Enabled         bool    `json:"enabled"`
	}

	if err := json.Unmarshal(value, &config); err != nil {
		return fmt.Errorf("invalid trading pair config structure: %w", err)
	}

	if config.Symbol == "" {
		return domain.NewValidationError("symbol", "symbol is required")
	}

	if config.BaseCurrency == "" {
		return domain.NewValidationError("base_currency", "base currency is required")
	}

	if config.QuoteCurrency == "" {
		return domain.NewValidationError("quote_currency", "quote currency is required")
	}

	if config.MinOrderSize <= 0 {
		return domain.NewValidationError("min_order_size", "must be greater than 0")
	}

	if config.MaxOrderSize <= 0 {
		return domain.NewValidationError("max_order_size", "must be greater than 0")
	}

	if config.MinOrderSize > config.MaxOrderSize {
		return domain.NewValidationError("min_order_size", "must be less than max_order_size")
	}

	if config.PricePrecision < 0 || config.PricePrecision > 8 {
		return domain.NewValidationError("price_precision", "must be between 0 and 8")
	}

	if config.QuantityPrecision < 0 || config.QuantityPrecision > 8 {
		return domain.NewValidationError("quantity_precision", "must be between 0 and 8")
	}

	return nil
}

// validateSystemConfig validates system configuration
func (v *Validator) validateSystemConfig(value json.RawMessage) error {
	var config struct {
		Name  string                 `json:"name"`
		Value interface{}            `json:"value"`
		Type  string                 `json:"type"`
	}

	if err := json.Unmarshal(value, &config); err != nil {
		return fmt.Errorf("invalid system config structure: %w", err)
	}

	if config.Name == "" {
		return domain.NewValidationError("name", "system config name is required")
	}

	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"object":  true,
		"array":   true,
	}

	if config.Type != "" && !validTypes[config.Type] {
		return domain.NewValidationError("type", fmt.Sprintf("invalid type: %s", config.Type))
	}

	return nil
}

// RegisterCustomValidator allows registering custom validators for specific config types
func (v *Validator) RegisterCustomValidator(configType domain.ConfigType, validatorFunc func(value json.RawMessage) error) {
	v.customValidators[configType] = validatorFunc
}
