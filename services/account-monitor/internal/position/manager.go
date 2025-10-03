package position

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/metrics"
)

const (
	redisPositionPrefix = "position:"
	redisPositionTTL    = 24 * time.Hour
)

type Manager struct {
	positions map[string]*Position
	mu        sync.RWMutex
	redis     *redis.Client
	logger    *zap.Logger
}

type Position struct {
	Symbol        string          `json:"symbol"`
	Quantity      decimal.Decimal `json:"quantity"`       // Positive = Long, Negative = Short
	EntryPrice    decimal.Decimal `json:"entry_price"`    // Weighted average
	CurrentPrice  decimal.Decimal `json:"current_price"`
	RealizedPnL   decimal.Decimal `json:"realized_pnl"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	TotalFees     decimal.Decimal `json:"total_fees"`
	LastUpdate    time.Time       `json:"last_update"`
	Trades        []Trade         `json:"trades"`
}

type Trade struct {
	ID          string          `json:"id"`
	Timestamp   time.Time       `json:"timestamp"`
	Side        string          `json:"side"` // BUY or SELL
	Quantity    decimal.Decimal `json:"quantity"`
	Price       decimal.Decimal `json:"price"`
	Fee         decimal.Decimal `json:"fee"`
	FeeCurrency string          `json:"fee_currency"`
}

type Fill struct {
	ID          string
	Symbol      string
	Side        string // BUY or SELL
	Quantity    decimal.Decimal
	Price       decimal.Decimal
	Fee         decimal.Decimal
	FeeCurrency string
	Timestamp   time.Time
}

func NewManager(redisClient *redis.Client, logger *zap.Logger) *Manager {
	return &Manager{
		positions: make(map[string]*Position),
		redis:     redisClient,
		logger:    logger,
	}
}

// UpdatePosition updates position based on fill
func (m *Manager) UpdatePosition(fill Fill) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, exists := m.positions[fill.Symbol]
	if !exists {
		pos = &Position{
			Symbol:      fill.Symbol,
			Quantity:    decimal.Zero,
			EntryPrice:  decimal.Zero,
			RealizedPnL: decimal.Zero,
			Trades:      []Trade{},
		}
		m.positions[fill.Symbol] = pos
	}

	// Create trade record
	trade := Trade{
		ID:          fill.ID,
		Timestamp:   fill.Timestamp,
		Side:        fill.Side,
		Quantity:    fill.Quantity,
		Price:       fill.Price,
		Fee:         fill.Fee,
		FeeCurrency: fill.FeeCurrency,
	}
	pos.Trades = append(pos.Trades, trade)

	// Calculate new position
	fillQty := fill.Quantity
	if fill.Side == "SELL" {
		fillQty = fillQty.Neg()
	}

	oldQty := pos.Quantity
	newQty := oldQty.Add(fillQty)

	// Case 1: Opening new position
	if oldQty.IsZero() {
		pos.EntryPrice = fill.Price
		pos.Quantity = newQty
	} else if oldQty.Sign() == fillQty.Sign() {
		// Case 2: Adding to position (same direction)
		oldValue := oldQty.Mul(pos.EntryPrice)
		newValue := fillQty.Abs().Mul(fill.Price)
		totalValue := oldValue.Add(newValue)
		pos.EntryPrice = totalValue.Div(newQty.Abs())
		pos.Quantity = newQty
	} else {
		// Case 3: Reducing or reversing position
		closedQty := decimal.Min(oldQty.Abs(), fillQty.Abs())

		// Calculate realized P&L for closed portion
		var pnl decimal.Decimal
		if oldQty.IsPositive() {
			// Closing long: (sell_price - entry_price) * qty
			pnl = fill.Price.Sub(pos.EntryPrice).Mul(closedQty)
		} else {
			// Closing short: (entry_price - buy_price) * qty
			pnl = pos.EntryPrice.Sub(fill.Price).Mul(closedQty)
		}

		// Subtract fees from P&L
		pnl = pnl.Sub(fill.Fee)
		pos.RealizedPnL = pos.RealizedPnL.Add(pnl)

		// Update quantity
		pos.Quantity = newQty

		// If position reversed, update entry price
		if !newQty.IsZero() && newQty.Sign() != oldQty.Sign() {
			pos.EntryPrice = fill.Price
		}

		// If position closed completely, reset entry price
		if newQty.IsZero() {
			pos.EntryPrice = decimal.Zero
		}
	}

	// Update total fees
	pos.TotalFees = pos.TotalFees.Add(fill.Fee)
	pos.LastUpdate = fill.Timestamp

	// Update metrics
	metrics.PositionCount.WithLabelValues(fill.Symbol).Set(1)
	if pos.Quantity.IsZero() {
		metrics.PositionCount.WithLabelValues(fill.Symbol).Set(0)
	}
	metrics.RealizedPnL.Set(pos.RealizedPnL.InexactFloat64())

	// Persist to Redis
	if err := m.saveToRedis(context.Background(), pos); err != nil {
		m.logger.Warn("Failed to save position to Redis", zap.Error(err))
	}

	m.logger.Debug("Position updated",
		zap.String("symbol", fill.Symbol),
		zap.String("side", fill.Side),
		zap.String("quantity", pos.Quantity.String()),
		zap.String("entry_price", pos.EntryPrice.String()),
		zap.String("realized_pnl", pos.RealizedPnL.String()),
	)

	return nil
}

// CalculateUnrealizedPnL calculates unrealized P&L for a position
func (m *Manager) CalculateUnrealizedPnL(symbol string, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, exists := m.positions[symbol]
	if !exists || pos.Quantity.IsZero() {
		return decimal.Zero, nil
	}

	var unrealizedPnL decimal.Decimal
	if pos.Quantity.IsPositive() {
		// Long position: (current_price - entry_price) * qty
		unrealizedPnL = currentPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
	} else {
		// Short position: (entry_price - current_price) * |qty|
		unrealizedPnL = pos.EntryPrice.Sub(currentPrice).Mul(pos.Quantity.Abs())
	}

	pos.CurrentPrice = currentPrice
	pos.UnrealizedPnL = unrealizedPnL

	// Update metrics
	metrics.UnrealizedPnL.WithLabelValues(symbol).Set(unrealizedPnL.InexactFloat64())
	posValue := currentPrice.Mul(pos.Quantity.Abs())
	metrics.PositionValue.WithLabelValues(symbol).Set(posValue.InexactFloat64())

	return unrealizedPnL, nil
}

// GetPosition retrieves a position by symbol
func (m *Manager) GetPosition(symbol string) (*Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pos, exists := m.positions[symbol]
	if !exists {
		return nil, fmt.Errorf("position not found for symbol: %s", symbol)
	}

	// Return a copy to prevent external modification
	posCopy := *pos
	return &posCopy, nil
}

// GetAllPositions returns all positions
func (m *Manager) GetAllPositions() map[string]*Position {
	m.mu.RLock()
	defer m.mu.RUnlock()

	positions := make(map[string]*Position)
	for symbol, pos := range m.positions {
		posCopy := *pos
		positions[symbol] = &posCopy
	}

	return positions
}

// SetPosition directly sets a position (used for reconciliation)
func (m *Manager) SetPosition(symbol string, quantity decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, exists := m.positions[symbol]
	if !exists {
		pos = &Position{
			Symbol:      symbol,
			Quantity:    quantity,
			RealizedPnL: decimal.Zero,
			Trades:      []Trade{},
		}
		m.positions[symbol] = pos
	} else {
		pos.Quantity = quantity
	}

	pos.LastUpdate = time.Now()

	return m.saveToRedis(context.Background(), pos)
}

// saveToRedis persists position to Redis
func (m *Manager) saveToRedis(ctx context.Context, pos *Position) error {
	key := redisPositionPrefix + pos.Symbol
	data, err := json.Marshal(pos)
	if err != nil {
		return err
	}

	return m.redis.Set(ctx, key, data, redisPositionTTL).Err()
}

// RestoreFromRedis restores positions from Redis
func (m *Manager) RestoreFromRedis(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pattern := redisPositionPrefix + "*"
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		data, err := m.redis.Get(ctx, key).Bytes()
		if err != nil {
			m.logger.Warn("Failed to get position from Redis", zap.String("key", key), zap.Error(err))
			continue
		}

		var pos Position
		if err := json.Unmarshal(data, &pos); err != nil {
			m.logger.Warn("Failed to unmarshal position", zap.String("key", key), zap.Error(err))
			continue
		}

		m.positions[pos.Symbol] = &pos
		m.logger.Info("Restored position from Redis", zap.String("symbol", pos.Symbol))
	}

	return nil
}

// SnapshotToRedis saves all positions to Redis
func (m *Manager) SnapshotToRedis(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, pos := range m.positions {
		if err := m.saveToRedis(ctx, pos); err != nil {
			return err
		}
	}

	return nil
}
