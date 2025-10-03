package balance

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
	redisBalancePrefix = "balance:"
	redisBalanceTTL    = 24 * time.Hour
)

type Manager struct {
	balances map[string]*Balance
	mu       sync.RWMutex
	redis    *redis.Client
	logger   *zap.Logger
}

type Balance struct {
	Asset      string          `json:"asset"`
	Free       decimal.Decimal `json:"free"`       // Available for trading
	Locked     decimal.Decimal `json:"locked"`     // In open orders
	Total      decimal.Decimal `json:"total"`      // Free + Locked
	USDValue   decimal.Decimal `json:"usd_value"`  // Mark-to-market value
	LastUpdate time.Time       `json:"last_update"`
}

type AccountEquity struct {
	TotalBalance    decimal.Decimal `json:"total_balance"`
	UnrealizedPnL   decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL     decimal.Decimal `json:"realized_pnl"`
	TotalEquity     decimal.Decimal `json:"total_equity"`
	Margin          decimal.Decimal `json:"margin"`
	MarginRatio     decimal.Decimal `json:"margin_ratio"`
	AvailableMargin decimal.Decimal `json:"available_margin"`
}

func NewManager(redisClient *redis.Client, logger *zap.Logger) *Manager {
	return &Manager{
		balances: make(map[string]*Balance),
		redis:    redisClient,
		logger:   logger,
	}
}

// UpdateBalance updates balance for an asset
func (m *Manager) UpdateBalance(asset string, free, locked decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance := &Balance{
		Asset:      asset,
		Free:       free,
		Locked:     locked,
		Total:      free.Add(locked),
		LastUpdate: time.Now(),
	}

	m.balances[asset] = balance

	// Update metrics
	metrics.AccountBalance.WithLabelValues(asset).Set(balance.Total.InexactFloat64())

	// Persist to Redis
	if err := m.saveToRedis(context.Background(), balance); err != nil {
		m.logger.Warn("Failed to save balance to Redis", zap.Error(err))
	}

	m.logger.Debug("Balance updated",
		zap.String("asset", asset),
		zap.String("free", free.String()),
		zap.String("locked", locked.String()),
		zap.String("total", balance.Total.String()),
	)

	return nil
}

// GetBalance retrieves balance for an asset
func (m *Manager) GetBalance(asset string) (*Balance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	balance, exists := m.balances[asset]
	if !exists {
		return nil, fmt.Errorf("balance not found for asset: %s", asset)
	}

	balanceCopy := *balance
	return &balanceCopy, nil
}

// GetAllBalances returns all balances
func (m *Manager) GetAllBalances() map[string]*Balance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	balances := make(map[string]*Balance)
	for asset, bal := range m.balances {
		balCopy := *bal
		balances[asset] = &balCopy
	}

	return balances
}

// SetBalance directly sets balance (used for reconciliation)
func (m *Manager) SetBalance(asset string, total decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance := &Balance{
		Asset:      asset,
		Free:       total, // Assume all free for reconciliation
		Locked:     decimal.Zero,
		Total:      total,
		LastUpdate: time.Now(),
	}

	m.balances[asset] = balance

	return m.saveToRedis(context.Background(), balance)
}

// CalculateTotalEquity calculates total account equity in USD
func (m *Manager) CalculateTotalEquity(priceMap map[string]decimal.Decimal) decimal.Decimal {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalEquity := decimal.Zero

	for asset, balance := range m.balances {
		// USDT and USD are already in USD
		if asset == "USDT" || asset == "USD" {
			totalEquity = totalEquity.Add(balance.Total)
			continue
		}

		// Convert to USD using price
		if price, ok := priceMap[asset+"USDT"]; ok {
			usdValue := balance.Total.Mul(price)
			totalEquity = totalEquity.Add(usdValue)
			balance.USDValue = usdValue
		}
	}

	// Update metrics
	metrics.AccountEquity.Set(totalEquity.InexactFloat64())

	return totalEquity
}

// saveToRedis persists balance to Redis
func (m *Manager) saveToRedis(ctx context.Context, balance *Balance) error {
	key := redisBalancePrefix + balance.Asset
	data, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	return m.redis.Set(ctx, key, data, redisBalanceTTL).Err()
}

// RestoreFromRedis restores balances from Redis
func (m *Manager) RestoreFromRedis(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pattern := redisBalancePrefix + "*"
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		data, err := m.redis.Get(ctx, key).Bytes()
		if err != nil {
			m.logger.Warn("Failed to get balance from Redis", zap.String("key", key), zap.Error(err))
			continue
		}

		var balance Balance
		if err := json.Unmarshal(data, &balance); err != nil {
			m.logger.Warn("Failed to unmarshal balance", zap.String("key", key), zap.Error(err))
			continue
		}

		m.balances[balance.Asset] = &balance
		m.logger.Info("Restored balance from Redis", zap.String("asset", balance.Asset))
	}

	return nil
}

// SnapshotToRedis saves all balances to Redis
func (m *Manager) SnapshotToRedis(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, balance := range m.balances {
		if err := m.saveToRedis(ctx, balance); err != nil {
			return err
		}
	}

	return nil
}
