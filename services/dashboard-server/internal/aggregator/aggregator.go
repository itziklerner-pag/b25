package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"

	"github.com/yourusername/b25/services/dashboard-server/internal/types"
)

// Aggregator manages state aggregation from multiple sources
type Aggregator struct {
	mu            sync.RWMutex
	marketData    map[string]*types.MarketData
	orders        []*types.Order
	positions     map[string]*types.Position
	account       *types.Account
	strategies    map[string]*types.Strategy
	lastUpdate    time.Time
	sequence      uint64
	logger        zerolog.Logger
	redisClient   *redis.Client
	ctx           context.Context
	cancel        context.CancelFunc
	updateChan    chan struct{}
}

// NewAggregator creates a new state aggregator
func NewAggregator(logger zerolog.Logger, redisURL string) *Aggregator {
	ctx, cancel := context.WithCancel(context.Background())

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   0,
	})

	return &Aggregator{
		marketData:  make(map[string]*types.MarketData),
		orders:      make([]*types.Order, 0),
		positions:   make(map[string]*types.Position),
		account:     &types.Account{Balances: make(map[string]float64)},
		strategies:  make(map[string]*types.Strategy),
		lastUpdate:  time.Now(),
		logger:      logger,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
		updateChan:  make(chan struct{}, 100),
	}
}

// Start begins the aggregation process
func (a *Aggregator) Start() error {
	// Test Redis connection
	if err := a.redisClient.Ping(a.ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	a.logger.Info().Msg("State aggregator started")

	// Load initial state
	go a.loadInitialState()

	// Subscribe to Redis pub/sub for real-time updates
	go a.subscribeToUpdates()

	// Periodic full refresh from backend services
	go a.periodicRefresh()

	return nil
}

// Stop stops the aggregation process
func (a *Aggregator) Stop() {
	a.cancel()
	a.redisClient.Close()
	a.logger.Info().Msg("State aggregator stopped")
}

// GetFullState returns a snapshot of the current state
func (a *Aggregator) GetFullState() *types.State {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return &types.State{
		MarketData: a.copyMarketData(),
		Orders:     a.copyOrders(),
		Positions:  a.copyPositions(),
		Account:    a.copyAccount(),
		Strategies: a.copyStrategies(),
		Timestamp:  time.Now(),
		Sequence:   a.sequence,
	}
}

// GetUpdateChannel returns a channel that signals when state is updated
func (a *Aggregator) GetUpdateChannel() <-chan struct{} {
	return a.updateChan
}

// loadInitialState loads the initial state from backend services
func (a *Aggregator) loadInitialState() {
	a.logger.Info().Msg("Loading initial state...")

	// TODO: Query backend services via gRPC/REST
	// For now, load from Redis cache if available
	a.loadFromRedis()

	// Initialize with demo data if empty
	if len(a.marketData) == 0 {
		a.initializeDemoData()
	}

	a.logger.Info().
		Int("market_data", len(a.marketData)).
		Int("orders", len(a.orders)).
		Int("positions", len(a.positions)).
		Msg("Initial state loaded")
}

// subscribeToUpdates subscribes to Redis pub/sub for real-time updates
func (a *Aggregator) subscribeToUpdates() {
	pubsub := a.redisClient.Subscribe(a.ctx,
		"market_data:*",
		"orders:*",
		"positions:*",
		"account:*",
		"strategies:*",
	)
	defer pubsub.Close()

	ch := pubsub.Channel()

	a.logger.Info().Msg("Subscribed to Redis pub/sub channels")

	for {
		select {
		case <-a.ctx.Done():
			return
		case msg := <-ch:
			a.handlePubSubMessage(msg)
		}
	}
}

// handlePubSubMessage handles incoming pub/sub messages
func (a *Aggregator) handlePubSubMessage(msg *redis.Message) {
	a.logger.Debug().
		Str("channel", msg.Channel).
		Msg("Received pub/sub message")

	// Parse message and update state
	// TODO: Implement proper message parsing based on channel pattern
	a.notifyUpdate()
}

// periodicRefresh periodically refreshes state from backend services
func (a *Aggregator) periodicRefresh() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.logger.Debug().Msg("Performing periodic state refresh")
			a.loadFromRedis()
			a.notifyUpdate()
		}
	}
}

// loadFromRedis loads state from Redis cache
func (a *Aggregator) loadFromRedis() {
	// Load market data
	marketDataKeys, err := a.redisClient.Keys(a.ctx, "market_data:*").Result()
	if err == nil {
		for _, key := range marketDataKeys {
			data, err := a.redisClient.Get(a.ctx, key).Result()
			if err == nil {
				var md types.MarketData
				if json.Unmarshal([]byte(data), &md) == nil {
					a.UpdateMarketData(md.Symbol, &md)
				}
			}
		}
	}

	// TODO: Load other state types from Redis
}

// UpdateMarketData updates market data for a symbol
func (a *Aggregator) UpdateMarketData(symbol string, data *types.MarketData) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data.UpdatedAt = time.Now()
	a.marketData[symbol] = data
	a.lastUpdate = time.Now()
	a.sequence++

	// Cache in Redis
	if jsonData, err := json.Marshal(data); err == nil {
		a.redisClient.Set(a.ctx, fmt.Sprintf("market_data:%s", symbol), jsonData, 5*time.Minute)
	}

	a.notifyUpdate()
}

// UpdateOrder updates or adds an order
func (a *Aggregator) UpdateOrder(order *types.Order) {
	a.mu.Lock()
	defer a.mu.Unlock()

	order.UpdatedAt = time.Now()

	// Replace or append order
	found := false
	for i, o := range a.orders {
		if o.ID == order.ID {
			a.orders[i] = order
			found = true
			break
		}
	}
	if !found {
		a.orders = append(a.orders, order)
	}

	a.lastUpdate = time.Now()
	a.sequence++
	a.notifyUpdate()
}

// UpdatePosition updates a position
func (a *Aggregator) UpdatePosition(symbol string, position *types.Position) {
	a.mu.Lock()
	defer a.mu.Unlock()

	position.UpdatedAt = time.Now()
	a.positions[symbol] = position
	a.lastUpdate = time.Now()
	a.sequence++
	a.notifyUpdate()
}

// UpdateAccount updates account information
func (a *Aggregator) UpdateAccount(account *types.Account) {
	a.mu.Lock()
	defer a.mu.Unlock()

	account.UpdatedAt = time.Now()
	a.account = account
	a.lastUpdate = time.Now()
	a.sequence++
	a.notifyUpdate()
}

// UpdateStrategy updates a strategy
func (a *Aggregator) UpdateStrategy(id string, strategy *types.Strategy) {
	a.mu.Lock()
	defer a.mu.Unlock()

	strategy.UpdatedAt = time.Now()
	a.strategies[id] = strategy
	a.lastUpdate = time.Now()
	a.sequence++
	a.notifyUpdate()
}

// notifyUpdate sends a notification on the update channel
func (a *Aggregator) notifyUpdate() {
	select {
	case a.updateChan <- struct{}{}:
	default:
		// Channel full, skip notification
	}
}

// Helper functions to create copies for thread safety
func (a *Aggregator) copyMarketData() map[string]*types.MarketData {
	copy := make(map[string]*types.MarketData)
	for k, v := range a.marketData {
		md := *v
		copy[k] = &md
	}
	return copy
}

func (a *Aggregator) copyOrders() []*types.Order {
	copy := make([]*types.Order, len(a.orders))
	for i, o := range a.orders {
		order := *o
		copy[i] = &order
	}
	return copy
}

func (a *Aggregator) copyPositions() map[string]*types.Position {
	copy := make(map[string]*types.Position)
	for k, v := range a.positions {
		pos := *v
		copy[k] = &pos
	}
	return copy
}

func (a *Aggregator) copyAccount() *types.Account {
	if a.account == nil {
		return &types.Account{Balances: make(map[string]float64)}
	}
	acc := *a.account
	acc.Balances = make(map[string]float64)
	for k, v := range a.account.Balances {
		acc.Balances[k] = v
	}
	return &acc
}

func (a *Aggregator) copyStrategies() map[string]*types.Strategy {
	copy := make(map[string]*types.Strategy)
	for k, v := range a.strategies {
		strat := *v
		copy[k] = &strat
	}
	return copy
}

// initializeDemoData initializes demo data for testing
func (a *Aggregator) initializeDemoData() {
	a.logger.Info().Msg("Initializing demo data")

	// Add demo market data
	symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	for _, symbol := range symbols {
		a.UpdateMarketData(symbol, &types.MarketData{
			Symbol:    symbol,
			LastPrice: 50000.0,
			BidPrice:  49999.0,
			AskPrice:  50001.0,
			Volume24h: 1000000.0,
			High24h:   51000.0,
			Low24h:    49000.0,
		})
	}

	// Add demo account
	a.UpdateAccount(&types.Account{
		TotalBalance:     10000.0,
		AvailableBalance: 8000.0,
		MarginUsed:       2000.0,
		UnrealizedPnL:    150.0,
		Balances: map[string]float64{
			"USDT": 10000.0,
			"BTC":  0.1,
		},
	})
}
