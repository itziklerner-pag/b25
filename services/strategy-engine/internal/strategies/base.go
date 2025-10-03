package strategies

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// BaseStrategy provides common functionality for strategies
type BaseStrategy struct {
	name      string
	running   bool
	mu        sync.RWMutex
	config    map[string]interface{}
	metrics   map[string]interface{}
	metricsMu sync.RWMutex
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(name string) *BaseStrategy {
	return &BaseStrategy{
		name:    name,
		running: false,
		metrics: make(map[string]interface{}),
	}
}

// Name returns the strategy name
func (s *BaseStrategy) Name() string {
	return s.name
}

// Init initializes the strategy
func (s *BaseStrategy) Init(config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	return nil
}

// Start starts the strategy
func (s *BaseStrategy) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
	s.updateMetric("start_time", time.Now().Unix())
	return nil
}

// Stop stops the strategy
func (s *BaseStrategy) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.updateMetric("stop_time", time.Now().Unix())
	return nil
}

// IsRunning returns whether the strategy is running
func (s *BaseStrategy) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetMetrics returns strategy metrics
func (s *BaseStrategy) GetMetrics() map[string]interface{} {
	s.metricsMu.RLock()
	defer s.metricsMu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := make(map[string]interface{}, len(s.metrics))
	for k, v := range s.metrics {
		metrics[k] = v
	}
	return metrics
}

// GetConfig returns configuration value
func (s *BaseStrategy) GetConfig(key string, defaultValue interface{}) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.config[key]; ok {
		return val
	}
	return defaultValue
}

// GetConfigFloat64 returns float64 configuration value
func (s *BaseStrategy) GetConfigFloat64(key string, defaultValue float64) float64 {
	val := s.GetConfig(key, defaultValue)
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultValue
	}
}

// GetConfigInt returns int configuration value
func (s *BaseStrategy) GetConfigInt(key string, defaultValue int) int {
	val := s.GetConfig(key, defaultValue)
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

// GetConfigString returns string configuration value
func (s *BaseStrategy) GetConfigString(key string, defaultValue string) string {
	val := s.GetConfig(key, defaultValue)
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

// GetConfigBool returns bool configuration value
func (s *BaseStrategy) GetConfigBool(key string, defaultValue bool) bool {
	val := s.GetConfig(key, defaultValue)
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// CreateSignal creates a new signal
func (s *BaseStrategy) CreateSignal(symbol, side, orderType string, quantity, price float64, priority int) *Signal {
	signal := &Signal{
		ID:        uuid.New().String(),
		Strategy:  s.name,
		Symbol:    symbol,
		Side:      side,
		OrderType: orderType,
		Quantity:  quantity,
		Priority:  priority,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if orderType == OrderTypeLimit || orderType == OrderTypeStopLimit {
		signal.Price = price
	}

	return signal
}

// IncrementMetric increments a metric counter
func (s *BaseStrategy) IncrementMetric(key string) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	if val, ok := s.metrics[key]; ok {
		if count, ok := val.(int64); ok {
			s.metrics[key] = count + 1
			return
		}
	}
	s.metrics[key] = int64(1)
}

// UpdateMetric updates a metric value (public version)
func (s *BaseStrategy) UpdateMetric(key string, value interface{}) {
	s.updateMetric(key, value)
}

// updateMetric updates a metric value (internal)
func (s *BaseStrategy) updateMetric(key string, value interface{}) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	s.metrics[key] = value
}

// CalculateSpread calculates bid-ask spread
func CalculateSpread(bidPrice, askPrice float64) float64 {
	if bidPrice <= 0 || askPrice <= 0 {
		return 0
	}
	return askPrice - bidPrice
}

// CalculateSpreadBps calculates bid-ask spread in basis points
func CalculateSpreadBps(bidPrice, askPrice float64) float64 {
	if bidPrice <= 0 || askPrice <= 0 {
		return 0
	}
	midPrice := (bidPrice + askPrice) / 2
	return ((askPrice - bidPrice) / midPrice) * 10000
}

// CalculateMidPrice calculates mid price
func CalculateMidPrice(bidPrice, askPrice float64) float64 {
	if bidPrice <= 0 || askPrice <= 0 {
		return 0
	}
	return (bidPrice + askPrice) / 2
}
