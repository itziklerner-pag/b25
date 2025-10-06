package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/yourusername/b25/services/dashboard-server/internal/types"
	pb "github.com/yourusername/b25/services/order-execution/proto"
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

	// Service clients
	orderGRPCClient pb.OrderServiceClient
	orderGRPCConn   *grpc.ClientConn
	httpClient      *http.Client
}

// Config holds aggregator configuration
type Config struct {
	RedisURL             string
	OrderServiceGRPC     string
	StrategyServiceHTTP  string
	AccountServiceGRPC   string
}

// NewAggregator creates a new state aggregator
func NewAggregator(logger zerolog.Logger, cfg Config) *Aggregator {
	ctx, cancel := context.WithCancel(context.Background())

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
		DB:   0,
	})

	agg := &Aggregator{
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
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}

	// Initialize gRPC connection to Order Execution service
	if cfg.OrderServiceGRPC != "" {
		conn, err := grpc.Dial(cfg.OrderServiceGRPC,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to Order Execution service")
		} else {
			agg.orderGRPCConn = conn
			agg.orderGRPCClient = pb.NewOrderServiceClient(conn)
			logger.Info().Msg("Connected to Order Execution gRPC service")
		}
	}

	return agg
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
	if a.orderGRPCConn != nil {
		a.orderGRPCConn.Close()
	}
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

	// Load from Redis cache and query backend services
	a.loadMarketDataFromRedis()
	a.loadOrdersFromService()
	a.loadStrategiesFromService()
	a.loadAccountData()

	// Only initialize demo data if no real data is available
	if len(a.marketData) == 0 {
		a.logger.Warn().Msg("No market data found, initializing minimal demo data")
		a.initializeMinimalDemoData()
	}

	a.logger.Info().
		Int("market_data", len(a.marketData)).
		Int("orders", len(a.orders)).
		Int("positions", len(a.positions)).
		Int("strategies", len(a.strategies)).
		Msg("Initial state loaded")
}

// subscribeToUpdates subscribes to Redis pub/sub for real-time updates
func (a *Aggregator) subscribeToUpdates() {
	// Subscribe to market data updates
	pubsub := a.redisClient.PSubscribe(a.ctx,
		"market_data:*",
		"orderbook:*",
		"trades:*",
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
		Str("pattern", msg.Pattern).
		Msg("Received pub/sub message")

	// Parse message based on channel pattern
	switch {
	case strings.HasPrefix(msg.Channel, "market_data:"):
		a.handleMarketDataUpdate(msg)
	case strings.HasPrefix(msg.Channel, "orderbook:"):
		a.handleOrderbookUpdate(msg)
	case strings.HasPrefix(msg.Channel, "trades:"):
		a.handleTradeUpdate(msg)
	case strings.HasPrefix(msg.Channel, "orders:"):
		a.handleOrderUpdate(msg)
	case strings.HasPrefix(msg.Channel, "positions:"):
		a.handlePositionUpdate(msg)
	case strings.HasPrefix(msg.Channel, "account:"):
		a.handleAccountUpdate(msg)
	case strings.HasPrefix(msg.Channel, "strategies:"):
		a.handleStrategyUpdate(msg)
	default:
		a.logger.Debug().Str("channel", msg.Channel).Msg("Unknown channel pattern")
	}
}

// handleMarketDataUpdate handles market data pub/sub updates
func (a *Aggregator) handleMarketDataUpdate(msg *redis.Message) {
	var md types.MarketData
	if err := json.Unmarshal([]byte(msg.Payload), &md); err != nil {
		a.logger.Error().Err(err).Msg("Failed to parse market data update")
		return
	}

	// Extract symbol from channel name (e.g., "market_data:BTCUSDT")
	parts := strings.Split(msg.Channel, ":")
	if len(parts) >= 2 {
		md.Symbol = parts[1]
	}

	a.UpdateMarketData(md.Symbol, &md)
	a.logger.Info().
		Str("symbol", md.Symbol).
		Float64("price", md.LastPrice).
		Msg("Updated market data from pub/sub")
}

// handleOrderbookUpdate handles orderbook pub/sub updates
func (a *Aggregator) handleOrderbookUpdate(msg *redis.Message) {
	// Extract symbol and fetch latest market data
	parts := strings.Split(msg.Channel, ":")
	if len(parts) >= 2 {
		symbol := parts[1]
		a.loadMarketDataForSymbol(symbol)
		a.logger.Debug().Str("symbol", symbol).Msg("Orderbook update triggered market data refresh")
	}
}

// handleTradeUpdate handles trade pub/sub updates
func (a *Aggregator) handleTradeUpdate(msg *redis.Message) {
	// Extract symbol and fetch latest market data
	parts := strings.Split(msg.Channel, ":")
	if len(parts) >= 2 {
		symbol := parts[1]
		a.loadMarketDataForSymbol(symbol)
		a.logger.Debug().Str("symbol", symbol).Msg("Trade update triggered market data refresh")
	}
}

// handleOrderUpdate handles order pub/sub updates
func (a *Aggregator) handleOrderUpdate(msg *redis.Message) {
	// Reload orders from service
	a.logger.Info().Msg("Order update received, reloading orders")
	a.loadOrdersFromService()
}

// handlePositionUpdate handles position pub/sub updates
func (a *Aggregator) handlePositionUpdate(msg *redis.Message) {
	var pos types.Position
	if err := json.Unmarshal([]byte(msg.Payload), &pos); err != nil {
		a.logger.Error().Err(err).Msg("Failed to parse position update")
		return
	}

	a.UpdatePosition(pos.Symbol, &pos)
	a.logger.Info().Str("symbol", pos.Symbol).Msg("Updated position from pub/sub")
}

// handleAccountUpdate handles account pub/sub updates
func (a *Aggregator) handleAccountUpdate(msg *redis.Message) {
	var acc types.Account
	if err := json.Unmarshal([]byte(msg.Payload), &acc); err != nil {
		a.logger.Error().Err(err).Msg("Failed to parse account update")
		return
	}

	a.UpdateAccount(&acc)
	a.logger.Info().Msg("Updated account from pub/sub")
}

// handleStrategyUpdate handles strategy pub/sub updates
func (a *Aggregator) handleStrategyUpdate(msg *redis.Message) {
	// Reload strategies from service
	a.logger.Info().Msg("Strategy update received, reloading strategies")
	a.loadStrategiesFromService()
}

// periodicRefresh periodically refreshes state from backend services
func (a *Aggregator) periodicRefresh() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.logger.Info().Msg("Performing periodic state refresh")
			a.loadMarketDataFromRedis()
			a.loadOrdersFromService()
			a.loadStrategiesFromService()
			a.notifyUpdate()
		}
	}
}

// loadMarketDataFromRedis loads market data from Redis
func (a *Aggregator) loadMarketDataFromRedis() {
	marketDataKeys, err := a.redisClient.Keys(a.ctx, "market_data:*").Result()
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to load market data keys from Redis")
		return
	}

	count := 0
	updated := 0
	for _, key := range marketDataKeys {
		data, err := a.redisClient.Get(a.ctx, key).Result()
		if err != nil {
			continue
		}

		var md types.MarketData
		if err := json.Unmarshal([]byte(data), &md); err != nil {
			continue
		}

		// Extract symbol from key
		parts := strings.Split(key, ":")
		if len(parts) >= 2 {
			md.Symbol = parts[1]
		}

		// Check if data actually changed before updating
		a.mu.RLock()
		oldMD, exists := a.marketData[md.Symbol]
		hasChanged := !exists || oldMD.LastPrice != md.LastPrice || oldMD.BidPrice != md.BidPrice || oldMD.AskPrice != md.AskPrice
		a.mu.RUnlock()

		if hasChanged {
			a.UpdateMarketData(md.Symbol, &md)
			updated++
		}
		count++
	}

	if count > 0 {
		a.logger.Info().
			Int("total", count).
			Int("updated", updated).
			Msg("Loaded market data from Redis")
	}
}

// loadMarketDataForSymbol loads market data for a specific symbol
func (a *Aggregator) loadMarketDataForSymbol(symbol string) {
	key := fmt.Sprintf("market_data:%s", symbol)
	data, err := a.redisClient.Get(a.ctx, key).Result()
	if err != nil {
		return
	}

	var md types.MarketData
	if err := json.Unmarshal([]byte(data), &md); err != nil {
		return
	}

	md.Symbol = symbol
	a.UpdateMarketData(symbol, &md)
}

// loadOrdersFromService loads orders from Order Execution service
func (a *Aggregator) loadOrdersFromService() {
	if a.orderGRPCClient == nil {
		a.logger.Debug().Msg("Order gRPC client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()

	// Query orders via gRPC
	resp, err := a.orderGRPCClient.GetOrders(ctx, &pb.GetOrdersRequest{
		Limit: 100, // Get last 100 orders
	})
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to load orders from service")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert proto orders to internal format
	a.orders = make([]*types.Order, 0, len(resp.Orders))
	for _, pbOrder := range resp.Orders {
		order := &types.Order{
			ID:        pbOrder.OrderId,
			Symbol:    pbOrder.Symbol,
			Side:      mapProtoSideToString(pbOrder.Side),
			Type:      mapProtoTypeToString(pbOrder.Type),
			Price:     pbOrder.Price,
			Quantity:  pbOrder.Quantity,
			Filled:    pbOrder.FilledQuantity,
			Status:    mapProtoStateToString(pbOrder.State),
			CreatedAt: time.Unix(pbOrder.CreatedAt, 0),
			UpdatedAt: time.Unix(pbOrder.UpdatedAt, 0),
		}
		a.orders = append(a.orders, order)
	}

	a.logger.Info().Int("count", len(a.orders)).Msg("Loaded orders from service")
	a.notifyUpdate()
}

// loadStrategiesFromService loads strategies from Strategy Engine service
func (a *Aggregator) loadStrategiesFromService() {
	// Query Strategy Engine HTTP API
	resp, err := a.httpClient.Get("http://localhost:8082/status")
	if err != nil {
		a.logger.Debug().Err(err).Msg("Failed to load strategies from service")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.logger.Debug().Int("status", resp.StatusCode).Msg("Strategy service returned non-200 status")
		return
	}

	var statusResp struct {
		Mode             string `json:"mode"`
		ActiveStrategies int    `json:"active_strategies"`
		SignalQueueSize  int    `json:"signal_queue_size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		a.logger.Error().Err(err).Msg("Failed to parse strategy service response")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Update strategies based on status
	// Create synthetic strategy entries based on active count
	a.strategies = make(map[string]*types.Strategy)
	strategyNames := []string{"Momentum", "Market Making", "Scalping"}

	for i := 0; i < statusResp.ActiveStrategies && i < len(strategyNames); i++ {
		id := fmt.Sprintf("strat-%d", i+1)
		status := "running"
		if statusResp.Mode == "simulation" {
			status = "simulation"
		}

		a.strategies[id] = &types.Strategy{
			ID:        id,
			Name:      strategyNames[i],
			Status:    status,
			PnL:       0, // TODO: Get real PnL from service
			Trades:    0, // TODO: Get real trade count
			WinRate:   0, // TODO: Get real win rate
			UpdatedAt: time.Now(),
		}
	}

	a.logger.Info().
		Int("count", len(a.strategies)).
		Str("mode", statusResp.Mode).
		Msg("Loaded strategies from service")
	a.notifyUpdate()
}

// loadAccountData loads account data (demo for now)
func (a *Aggregator) loadAccountData() {
	// TODO: Query from Account Monitor gRPC when available
	// For now, use demo data
	a.mu.Lock()
	defer a.mu.Unlock()

	a.account = &types.Account{
		TotalBalance:     10000.0,
		AvailableBalance: 8500.0,
		MarginUsed:       1500.0,
		UnrealizedPnL:    0.0,
		Balances: map[string]float64{
			"USDT": 10000.0,
			"BTC":  0.05,
			"ETH":  0.5,
		},
		UpdatedAt: time.Now(),
	}
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

	a.logger.Debug().
		Str("symbol", symbol).
		Float64("price", data.LastPrice).
		Uint64("sequence", a.sequence).
		Msg("Market data updated")

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

	a.logger.Info().
		Str("order_id", order.ID).
		Str("symbol", order.Symbol).
		Str("status", order.Status).
		Uint64("sequence", a.sequence).
		Msg("Order updated")

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

	a.logger.Info().
		Str("symbol", symbol).
		Float64("pnl", position.UnrealizedPnL).
		Uint64("sequence", a.sequence).
		Msg("Position updated")

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

	a.logger.Info().
		Float64("balance", account.TotalBalance).
		Float64("pnl", account.UnrealizedPnL).
		Uint64("sequence", a.sequence).
		Msg("Account updated")

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

	a.logger.Info().
		Str("strategy_id", id).
		Str("status", strategy.Status).
		Uint64("sequence", a.sequence).
		Msg("Strategy updated")

	a.notifyUpdate()
}

// notifyUpdate sends a notification on the update channel
func (a *Aggregator) notifyUpdate() {
	select {
	case a.updateChan <- struct{}{}:
		a.logger.Debug().Msg("Update notification sent")
	default:
		// Channel full, skip notification
		a.logger.Warn().Msg("Update channel full, skipping notification")
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

// initializeMinimalDemoData initializes minimal demo data only if no real data exists
func (a *Aggregator) initializeMinimalDemoData() {
	a.logger.Info().Msg("Initializing minimal demo data")

	// Add demo market data only for symbols not already present
	symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	for _, symbol := range symbols {
		if _, exists := a.marketData[symbol]; !exists {
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
	}
}

// Helper functions to map proto enums to strings
func mapProtoSideToString(side pb.OrderSide) string {
	switch side {
	case pb.OrderSide_BUY:
		return "BUY"
	case pb.OrderSide_SELL:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

func mapProtoTypeToString(orderType pb.OrderType) string {
	switch orderType {
	case pb.OrderType_MARKET:
		return "MARKET"
	case pb.OrderType_LIMIT:
		return "LIMIT"
	case pb.OrderType_STOP_MARKET:
		return "STOP_MARKET"
	case pb.OrderType_STOP_LIMIT:
		return "STOP_LIMIT"
	default:
		return "LIMIT"
	}
}

func mapProtoStateToString(state pb.OrderState) string {
	switch state {
	case pb.OrderState_NEW:
		return "NEW"
	case pb.OrderState_SUBMITTED:
		return "SUBMITTED"
	case pb.OrderState_PARTIALLY_FILLED:
		return "PARTIALLY_FILLED"
	case pb.OrderState_FILLED:
		return "FILLED"
	case pb.OrderState_CANCELED:
		return "CANCELED"
	case pb.OrderState_REJECTED:
		return "REJECTED"
	case pb.OrderState_EXPIRED:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}
