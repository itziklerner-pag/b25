package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"github.com/sony/gobreaker"

	"github.com/yourusername/b25/services/order-execution/internal/circuitbreaker"
	"github.com/yourusername/b25/services/order-execution/internal/exchange"
	"github.com/yourusername/b25/services/order-execution/internal/metrics"
	"github.com/yourusername/b25/services/order-execution/internal/models"
	"github.com/yourusername/b25/services/order-execution/internal/ratelimit"
	"github.com/yourusername/b25/services/order-execution/internal/validator"
)

// OrderExecutor handles order execution and lifecycle management
type OrderExecutor struct {
	exchangeClient *exchange.BinanceClient
	validator      *validator.Validator
	rateLimiter    *ratelimit.RateLimiter
	circuitBreaker *circuitbreaker.CircuitBreaker
	redisClient    *redis.Client
	natsConn       *nats.Conn
	logger         *zap.Logger
	metrics        *metrics.Metrics

	orders sync.Map // orderID -> *models.Order
	mu     sync.RWMutex
}

// Config holds executor configuration
type Config struct {
	ExchangeAPIKey    string
	ExchangeSecretKey string
	TestnetMode       bool
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	NATSAddr          string
	RateLimitRPS      int
	RateLimitBurst    int
}

// NewOrderExecutor creates a new order executor
func NewOrderExecutor(cfg Config, logger *zap.Logger) (*OrderExecutor, error) {
	// Initialize exchange client
	exchangeClient := exchange.NewBinanceClient(cfg.ExchangeAPIKey, cfg.ExchangeSecretKey, cfg.TestnetMode)

	// Initialize validator with risk limits
	riskLimits := &validator.RiskLimits{
		MaxOrderValue:   1000000, // $1M
		MaxPositionSize: 10,      // 10 BTC equivalent
		MaxDailyOrders:  10000,
		MaxOpenOrders:   500,
		AllowedSymbols: map[string]bool{
			"BTCUSDT": true,
			"ETHUSDT": true,
		},
	}
	validatorInstance := validator.NewValidator(riskLimits)

	// Initialize rate limiter
	rateLimiter := ratelimit.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	// Initialize circuit breaker
	cbConfig := circuitbreaker.DefaultConfig()
	cbConfig.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		logger.Warn("circuit breaker state change",
			zap.String("name", name),
			zap.Any("from", from),
			zap.Any("to", to),
		)
	}
	circuitBreaker := circuitbreaker.NewCircuitBreaker(cbConfig)

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Initialize NATS
	natsConn, err := nats.Connect(cfg.NATSAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Initialize metrics
	metricsInstance := metrics.NewMetrics()

	executor := &OrderExecutor{
		exchangeClient: exchangeClient,
		validator:      validatorInstance,
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
		redisClient:    redisClient,
		natsConn:       natsConn,
		logger:         logger,
		metrics:        metricsInstance,
	}

	// Load exchange info and register symbols
	if err := executor.loadExchangeInfo(); err != nil {
		logger.Warn("failed to load exchange info", zap.Error(err))
	}

	return executor, nil
}

// CreateOrder creates and submits a new order
func (e *OrderExecutor) CreateOrder(ctx context.Context, order *models.Order) error {
	startTime := time.Now()

	// Generate order ID if not set
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}

	// Set initial state
	order.State = models.OrderStateNew
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	e.logger.Info("creating order",
		zap.String("order_id", order.OrderID),
		zap.String("symbol", order.Symbol),
		zap.String("side", string(order.Side)),
		zap.String("type", string(order.Type)),
		zap.Float64("quantity", order.Quantity),
		zap.Float64("price", order.Price),
	)

	// Validate order
	if err := e.validator.ValidateOrder(order); err != nil {
		e.logger.Error("order validation failed", zap.Error(err))
		e.metrics.OrdersRejected.Inc()
		order.State = models.OrderStateRejected
		e.cacheOrder(ctx, order)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Store order
	e.orders.Store(order.OrderID, order)
	e.cacheOrder(ctx, order)

	// Submit to exchange
	if err := e.submitOrder(ctx, order); err != nil {
		e.logger.Error("failed to submit order", zap.Error(err))
		e.metrics.OrdersRejected.Inc()
		order.State = models.OrderStateRejected
		e.updateOrder(ctx, order, "REJECTED")
		return fmt.Errorf("submission failed: %w", err)
	}

	// Record metrics
	duration := time.Since(startTime)
	e.metrics.OrdersCreated.Inc()
	e.metrics.OrderLatency.Observe(duration.Seconds())

	e.logger.Info("order created successfully",
		zap.String("order_id", order.OrderID),
		zap.String("exchange_order_id", order.ExchangeOrderID),
		zap.Duration("latency", duration),
	)

	return nil
}

// submitOrder submits order to exchange
func (e *OrderExecutor) submitOrder(ctx context.Context, order *models.Order) error {
	// Apply rate limiting
	if err := e.rateLimiter.Wait(ctx, "order_submission"); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Execute with circuit breaker
	result, err := e.circuitBreaker.ExecuteWithContext(ctx, "binance_create_order", func() (interface{}, error) {
		return e.exchangeClient.CreateOrder(order)
	})

	if err != nil {
		e.metrics.ExchangeErrors.Inc()
		return err
	}

	resp := result.(*exchange.BinanceOrderResponse)

	// Update order with exchange response
	order.ExchangeOrderID = strconv.FormatInt(resp.OrderID, 10)
	order.State = models.OrderState(exchange.MapBinanceOrderStatus(resp.Status))
	order.UpdatedAt = time.Now()

	// Parse filled quantity
	if resp.ExecutedQty != "" {
		filledQty, _ := strconv.ParseFloat(resp.ExecutedQty, 64)
		order.FilledQuantity = filledQty
	}

	// Parse average price
	if resp.AvgPrice != "" {
		avgPrice, _ := strconv.ParseFloat(resp.AvgPrice, 64)
		order.AveragePrice = avgPrice
	}

	// Update order state
	e.updateOrder(ctx, order, "SUBMITTED")

	return nil
}

// CancelOrder cancels an existing order
func (e *OrderExecutor) CancelOrder(ctx context.Context, orderID, symbol string) error {
	startTime := time.Now()

	// Get order
	orderInterface, exists := e.orders.Load(orderID)
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	order := orderInterface.(*models.Order)

	// Check if order can be canceled
	if order.IsTerminal() {
		return fmt.Errorf("order is in terminal state: %s", order.State)
	}

	e.logger.Info("canceling order",
		zap.String("order_id", orderID),
		zap.String("symbol", symbol),
	)

	// Apply rate limiting
	if err := e.rateLimiter.Wait(ctx, "order_cancellation"); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Execute with circuit breaker
	_, err := e.circuitBreaker.ExecuteWithContext(ctx, "binance_cancel_order", func() (interface{}, error) {
		return e.exchangeClient.CancelOrder(symbol, order.ExchangeOrderID)
	})

	if err != nil {
		e.metrics.ExchangeErrors.Inc()
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Update order state
	order.State = models.OrderStateCanceled
	order.UpdatedAt = time.Now()
	e.updateOrder(ctx, order, "CANCELED")

	// Record metrics
	duration := time.Since(startTime)
	e.metrics.OrdersCanceled.Inc()

	e.logger.Info("order canceled successfully",
		zap.String("order_id", orderID),
		zap.Duration("latency", duration),
	)

	return nil
}

// GetOrder retrieves an order
func (e *OrderExecutor) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	// Try memory cache first
	if orderInterface, exists := e.orders.Load(orderID); exists {
		return orderInterface.(*models.Order), nil
	}

	// Try Redis cache
	order, err := e.getOrderFromCache(ctx, orderID)
	if err == nil && order != nil {
		return order, nil
	}

	return nil, fmt.Errorf("order not found: %s", orderID)
}

// updateOrder updates order state and publishes event
func (e *OrderExecutor) updateOrder(ctx context.Context, order *models.Order, updateType string) {
	// Update in memory
	e.orders.Store(order.OrderID, order)

	// Update in Redis
	e.cacheOrder(ctx, order)

	// Publish update event
	update := &models.OrderUpdate{
		Order:      order,
		UpdateType: updateType,
		Timestamp:  time.Now(),
	}

	e.publishOrderUpdate(update)

	// Record state metrics
	e.metrics.RecordOrderState(string(order.State))
}

// cacheOrder stores order in Redis
func (e *OrderExecutor) cacheOrder(ctx context.Context, order *models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("order:%s", order.OrderID)
	return e.redisClient.Set(ctx, key, data, 24*time.Hour).Err()
}

// getOrderFromCache retrieves order from Redis
func (e *OrderExecutor) getOrderFromCache(ctx context.Context, orderID string) (*models.Order, error) {
	key := fmt.Sprintf("order:%s", orderID)
	data, err := e.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// publishOrderUpdate publishes order update to NATS
func (e *OrderExecutor) publishOrderUpdate(update *models.OrderUpdate) {
	data, err := json.Marshal(update)
	if err != nil {
		e.logger.Error("failed to marshal order update", zap.Error(err))
		return
	}

	subject := fmt.Sprintf("orders.updates.%s", update.Order.Symbol)
	if err := e.natsConn.Publish(subject, data); err != nil {
		e.logger.Error("failed to publish order update",
			zap.String("subject", subject),
			zap.Error(err),
		)
		return
	}

	e.metrics.EventsPublished.Inc()
}

// loadExchangeInfo loads exchange information and registers symbols
func (e *OrderExecutor) loadExchangeInfo() error {
	info, err := e.exchangeClient.GetExchangeInfo()
	if err != nil {
		return err
	}

	for _, symbol := range info.Symbols {
		if symbol.Status != "TRADING" {
			continue
		}

		symbolInfo := &validator.SymbolInfo{
			Symbol:            symbol.Symbol,
			PricePrecision:    symbol.PricePrecision,
			QuantityPrecision: symbol.QuantityPrecision,
		}

		// Parse filters
		for _, filter := range symbol.Filters {
			switch filter.FilterType {
			case "PRICE_FILTER":
				if tickSize, err := strconv.ParseFloat(filter.TickSize, 64); err == nil {
					symbolInfo.TickSize = tickSize
				}
			case "LOT_SIZE":
				if minQty, err := strconv.ParseFloat(filter.MinQty, 64); err == nil {
					symbolInfo.MinQuantity = minQty
				}
				if maxQty, err := strconv.ParseFloat(filter.MaxQty, 64); err == nil {
					symbolInfo.MaxQuantity = maxQty
				}
				if stepSize, err := strconv.ParseFloat(filter.StepSize, 64); err == nil {
					symbolInfo.StepSize = stepSize
				}
			case "MIN_NOTIONAL":
				if minNotional, err := strconv.ParseFloat(filter.MinNotional, 64); err == nil {
					symbolInfo.MinNotional = minNotional
				}
			}
		}

		e.validator.RegisterSymbol(symbolInfo)
	}

	e.logger.Info("loaded exchange info", zap.Int("symbols", len(info.Symbols)))
	return nil
}

// Close closes all connections
func (e *OrderExecutor) Close() error {
	if e.natsConn != nil {
		e.natsConn.Close()
	}
	if e.redisClient != nil {
		return e.redisClient.Close()
	}
	return nil
}
