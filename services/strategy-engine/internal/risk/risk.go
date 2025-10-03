package risk

import (
	"fmt"
	"sync"
	"time"

	"github.com/b25/strategy-engine/internal/config"
	"github.com/b25/strategy-engine/internal/strategies"
	"github.com/b25/strategy-engine/pkg/logger"
	"github.com/b25/strategy-engine/pkg/metrics"
)

// Manager handles risk management and filtering
type Manager struct {
	cfg     *config.RiskConfig
	logger  *logger.Logger
	metrics *metrics.Collector

	// State tracking
	positions      map[string]float64
	dailyPnL       float64
	dailyVolume    map[string]float64
	orderCounts    *OrderCounter
	accountBalance float64
	maxDrawdown    float64
	peakBalance    float64
	mu             sync.RWMutex
}

// OrderCounter tracks order rates
type OrderCounter struct {
	secondCounter  map[int64]int
	minuteCounter  map[int64]int
	mu             sync.RWMutex
}

// NewManager creates a new risk manager
func NewManager(cfg *config.RiskConfig, log *logger.Logger, m *metrics.Collector) *Manager {
	return &Manager{
		cfg:            cfg,
		logger:         log,
		metrics:        m,
		positions:      make(map[string]float64),
		dailyVolume:    make(map[string]float64),
		orderCounts:    newOrderCounter(),
		accountBalance: 100000.0, // Default starting balance
		peakBalance:    100000.0,
	}
}

func newOrderCounter() *OrderCounter {
	return &OrderCounter{
		secondCounter: make(map[int64]int),
		minuteCounter: make(map[int64]int),
	}
}

// ValidateSignal validates a signal against risk rules
func (m *Manager) ValidateSignal(signal *strategies.Signal) error {
	if !m.cfg.Enabled {
		m.metrics.RiskChecks.WithLabelValues("disabled", "pass").Inc()
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check symbol whitelist/blacklist
	if err := m.checkSymbol(signal.Symbol); err != nil {
		m.metrics.RiskChecks.WithLabelValues("symbol", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("symbol_blocked", signal.Strategy).Inc()
		return err
	}

	// Check order value limit
	if err := m.checkOrderValue(signal); err != nil {
		m.metrics.RiskChecks.WithLabelValues("order_value", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("order_value_exceeded", signal.Strategy).Inc()
		return err
	}

	// Check position size limit
	if err := m.checkPositionSize(signal); err != nil {
		m.metrics.RiskChecks.WithLabelValues("position_size", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("position_size_exceeded", signal.Strategy).Inc()
		return err
	}

	// Check daily loss limit
	if err := m.checkDailyLoss(); err != nil {
		m.metrics.RiskChecks.WithLabelValues("daily_loss", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("daily_loss_exceeded", signal.Strategy).Inc()
		return err
	}

	// Check drawdown limit
	if err := m.checkDrawdown(); err != nil {
		m.metrics.RiskChecks.WithLabelValues("drawdown", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("drawdown_exceeded", signal.Strategy).Inc()
		return err
	}

	// Check account balance
	if err := m.checkAccountBalance(); err != nil {
		m.metrics.RiskChecks.WithLabelValues("account_balance", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("low_balance", signal.Strategy).Inc()
		return err
	}

	// Check order rate limits
	if err := m.checkOrderRateLimits(); err != nil {
		m.metrics.RiskChecks.WithLabelValues("rate_limit", "fail").Inc()
		m.metrics.RiskViolations.WithLabelValues("rate_limit_exceeded", signal.Strategy).Inc()
		return err
	}

	m.metrics.RiskChecks.WithLabelValues("all", "pass").Inc()
	return nil
}

// checkSymbol checks if symbol is allowed
func (m *Manager) checkSymbol(symbol string) error {
	// Check blacklist first
	for _, blocked := range m.cfg.BlockedSymbols {
		if symbol == blocked {
			return fmt.Errorf("symbol %s is blocked", symbol)
		}
	}

	// If whitelist is defined, check it
	if len(m.cfg.AllowedSymbols) > 0 {
		allowed := false
		for _, allowedSymbol := range m.cfg.AllowedSymbols {
			if symbol == allowedSymbol {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("symbol %s is not in allowed list", symbol)
		}
	}

	return nil
}

// checkOrderValue checks if order value is within limits
func (m *Manager) checkOrderValue(signal *strategies.Signal) error {
	if m.cfg.MaxOrderValue <= 0 {
		return nil
	}

	orderValue := signal.Quantity * signal.Price
	if signal.OrderType == strategies.OrderTypeMarket {
		// Use a conservative estimate for market orders
		orderValue = signal.Quantity * 1000 // Placeholder
	}

	if orderValue > m.cfg.MaxOrderValue {
		return fmt.Errorf("order value %.2f exceeds limit %.2f", orderValue, m.cfg.MaxOrderValue)
	}

	return nil
}

// checkPositionSize checks if position size is within limits
func (m *Manager) checkPositionSize(signal *strategies.Signal) error {
	if m.cfg.MaxPositionSize <= 0 {
		return nil
	}

	currentPosition := m.positions[signal.Symbol]
	newPosition := currentPosition

	if signal.Side == strategies.SideBuy {
		newPosition += signal.Quantity
	} else {
		newPosition -= signal.Quantity
	}

	if abs(newPosition) > m.cfg.MaxPositionSize {
		return fmt.Errorf("position size %.2f exceeds limit %.2f", abs(newPosition), m.cfg.MaxPositionSize)
	}

	return nil
}

// checkDailyLoss checks if daily loss limit is exceeded
func (m *Manager) checkDailyLoss() error {
	if m.cfg.MaxDailyLoss <= 0 {
		return nil
	}

	if m.dailyPnL < -m.cfg.MaxDailyLoss {
		return fmt.Errorf("daily loss %.2f exceeds limit %.2f", abs(m.dailyPnL), m.cfg.MaxDailyLoss)
	}

	return nil
}

// checkDrawdown checks if drawdown limit is exceeded
func (m *Manager) checkDrawdown() error {
	if m.cfg.MaxDrawdown <= 0 {
		return nil
	}

	currentDrawdown := (m.peakBalance - m.accountBalance) / m.peakBalance
	if currentDrawdown > m.cfg.MaxDrawdown {
		return fmt.Errorf("drawdown %.2f%% exceeds limit %.2f%%", currentDrawdown*100, m.cfg.MaxDrawdown*100)
	}

	return nil
}

// checkAccountBalance checks if account balance is above minimum
func (m *Manager) checkAccountBalance() error {
	if m.cfg.MinAccountBalance <= 0 {
		return nil
	}

	if m.accountBalance < m.cfg.MinAccountBalance {
		return fmt.Errorf("account balance %.2f below minimum %.2f", m.accountBalance, m.cfg.MinAccountBalance)
	}

	return nil
}

// checkOrderRateLimits checks if order rate limits are exceeded
func (m *Manager) checkOrderRateLimits() error {
	now := time.Now()
	currentSecond := now.Unix()
	currentMinute := now.Unix() / 60

	m.orderCounts.mu.RLock()
	ordersThisSecond := m.orderCounts.secondCounter[currentSecond]
	ordersThisMinute := m.orderCounts.minuteCounter[currentMinute]
	m.orderCounts.mu.RUnlock()

	if m.cfg.MaxOrdersPerSecond > 0 && ordersThisSecond >= m.cfg.MaxOrdersPerSecond {
		return fmt.Errorf("order rate limit exceeded: %d orders/second", ordersThisSecond)
	}

	if m.cfg.MaxOrdersPerMinute > 0 && ordersThisMinute >= m.cfg.MaxOrdersPerMinute {
		return fmt.Errorf("order rate limit exceeded: %d orders/minute", ordersThisMinute)
	}

	return nil
}

// RecordOrder records an order for rate limiting
func (m *Manager) RecordOrder() {
	now := time.Now()
	currentSecond := now.Unix()
	currentMinute := now.Unix() / 60

	m.orderCounts.mu.Lock()
	m.orderCounts.secondCounter[currentSecond]++
	m.orderCounts.minuteCounter[currentMinute]++

	// Clean up old entries
	for second := range m.orderCounts.secondCounter {
		if currentSecond-second > 2 {
			delete(m.orderCounts.secondCounter, second)
		}
	}
	for minute := range m.orderCounts.minuteCounter {
		if currentMinute-minute > 2 {
			delete(m.orderCounts.minuteCounter, minute)
		}
	}
	m.orderCounts.mu.Unlock()
}

// UpdatePosition updates position tracking
func (m *Manager) UpdatePosition(symbol string, quantity float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.positions[symbol] = quantity
}

// UpdatePnL updates daily PnL tracking
func (m *Manager) UpdatePnL(pnl float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dailyPnL += pnl
	m.accountBalance += pnl

	// Update peak balance and max drawdown
	if m.accountBalance > m.peakBalance {
		m.peakBalance = m.accountBalance
	}

	drawdown := (m.peakBalance - m.accountBalance) / m.peakBalance
	if drawdown > m.maxDrawdown {
		m.maxDrawdown = drawdown
	}
}

// ResetDaily resets daily counters
func (m *Manager) ResetDaily() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dailyPnL = 0
	m.dailyVolume = make(map[string]float64)

	m.logger.Info("Daily risk counters reset")
}

// GetMetrics returns risk metrics
func (m *Manager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"account_balance": m.accountBalance,
		"daily_pnl":       m.dailyPnL,
		"peak_balance":    m.peakBalance,
		"max_drawdown":    m.maxDrawdown,
		"positions":       len(m.positions),
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
