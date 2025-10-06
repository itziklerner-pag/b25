package engine

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/b25/strategy-engine/internal/config"
	grpcclient "github.com/b25/strategy-engine/internal/grpc"
	"github.com/b25/strategy-engine/internal/pubsub"
	"github.com/b25/strategy-engine/internal/risk"
	"github.com/b25/strategy-engine/internal/strategies"
	"github.com/b25/strategy-engine/pkg/logger"
	"github.com/b25/strategy-engine/pkg/metrics"
)

// Engine is the main strategy engine
type Engine struct {
	cfg      *config.Config
	logger   *logger.Logger
	metrics  *metrics.Collector

	// Components
	registry       *strategies.Registry
	riskManager    *risk.Manager
	redisSubscriber *pubsub.RedisSubscriber
	natsSubscriber *pubsub.NATSSubscriber
	orderClient    *grpcclient.OrderExecutionClient

	// Active strategies
	activeStrategies map[string]strategies.Strategy
	strategyMu       sync.RWMutex

	// Signal processing
	signalQueue chan *strategies.Signal
	signalMu    sync.Mutex

	// Plugin management
	pluginLoader *PluginLoader

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new strategy engine
func New(cfg *config.Config, log *logger.Logger, m *metrics.Collector) (*Engine, error) {
	// Create risk manager
	riskMgr := risk.NewManager(&cfg.Risk, log, m)

	// Create strategy registry
	registry := strategies.NewRegistry()

	// Create Redis subscriber
	redisSubscriber, err := pubsub.NewRedisSubscriber(&cfg.Redis, log, m)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis subscriber: %w", err)
	}

	// Create NATS subscriber
	natsSubscriber, err := pubsub.NewNATSSubscriber(&cfg.NATS, log, m)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS subscriber: %w", err)
	}

	// Create gRPC client for order execution
	orderClient, err := grpcclient.NewOrderExecutionClient(&cfg.GRPC, log, m)
	if err != nil {
		return nil, fmt.Errorf("failed to create order execution client: %w", err)
	}

	// Create plugin loader
	pluginLoader := NewPluginLoader(cfg.Engine.PluginsDir, log, m)

	ctx, cancel := context.WithCancel(context.Background())

	engine := &Engine{
		cfg:              cfg,
		logger:           log,
		metrics:          m,
		registry:         registry,
		riskManager:      riskMgr,
		redisSubscriber:  redisSubscriber,
		natsSubscriber:   natsSubscriber,
		orderClient:      orderClient,
		activeStrategies: make(map[string]strategies.Strategy),
		signalQueue:      make(chan *strategies.Signal, cfg.Engine.SignalBufferSize),
		pluginLoader:     pluginLoader,
		ctx:              ctx,
		cancel:           cancel,
	}

	return engine, nil
}

// Start starts the strategy engine
func (e *Engine) Start() error {
	e.logger.Info("Starting strategy engine",
		"mode", e.cfg.Engine.Mode,
		"hot_reload", e.cfg.Engine.HotReload,
	)

	// Load and initialize strategies
	if err := e.loadStrategies(); err != nil {
		return fmt.Errorf("failed to load strategies: %w", err)
	}

	// Start signal processor
	e.wg.Add(1)
	go e.processSignals()

	// Start market data subscription
	e.wg.Add(1)
	go e.subscribeMarketData()

	// Start fill event subscription
	e.wg.Add(1)
	go e.subscribeFills()

	// Start position update subscription
	e.wg.Add(1)
	go e.subscribePositions()

	// Start plugin hot reload if enabled
	if e.cfg.Engine.HotReload {
		e.wg.Add(1)
		go e.hotReloadLoop()
	}

	// Start daily reset timer
	e.wg.Add(1)
	go e.dailyResetLoop()

	e.logger.Info("Strategy engine started")
	return nil
}

// Stop stops the strategy engine
func (e *Engine) Stop() error {
	e.logger.Info("Stopping strategy engine...")

	// Cancel context to stop all goroutines
	e.cancel()

	// Stop all strategies
	e.strategyMu.Lock()
	for name, strategy := range e.activeStrategies {
		if err := strategy.Stop(); err != nil {
			e.logger.Error("Failed to stop strategy",
				"strategy", name,
				"error", err,
			)
		}
	}
	e.strategyMu.Unlock()

	// Close connections
	if err := e.redisSubscriber.Close(); err != nil {
		e.logger.Error("Failed to close Redis connection", "error", err)
	}

	if err := e.natsSubscriber.Close(); err != nil {
		e.logger.Error("Failed to close NATS connection", "error", err)
	}

	if err := e.orderClient.Close(); err != nil {
		e.logger.Error("Failed to close gRPC connection", "error", err)
	}

	// Wait for all goroutines to finish
	e.wg.Wait()

	e.logger.Info("Strategy engine stopped")
	return nil
}

// loadStrategies loads and initializes strategies
func (e *Engine) loadStrategies() error {
	e.logger.Info("Loading strategies", "enabled", e.cfg.Strategies.Enabled)

	for _, strategyName := range e.cfg.Strategies.Enabled {
		// Get strategy config
		strategyConfig, ok := e.cfg.Strategies.Configs[strategyName].(map[string]interface{})
		if !ok {
			strategyConfig = make(map[string]interface{})
		}

		// Create strategy instance
		strategy, err := e.registry.Create(strategyName)
		if err != nil {
			e.logger.Error("Failed to create strategy",
				"strategy", strategyName,
				"error", err,
			)
			continue
		}

		// Initialize strategy
		if err := strategy.Init(strategyConfig); err != nil {
			e.logger.Error("Failed to initialize strategy",
				"strategy", strategyName,
				"error", err,
			)
			continue
		}

		// Start strategy
		if err := strategy.Start(); err != nil {
			e.logger.Error("Failed to start strategy",
				"strategy", strategyName,
				"error", err,
			)
			continue
		}

		// Add to active strategies
		e.strategyMu.Lock()
		e.activeStrategies[strategyName] = strategy
		e.strategyMu.Unlock()

		e.logger.Info("Strategy loaded and started", "strategy", strategyName)
	}

	e.metrics.ActiveStrategies.Set(float64(len(e.activeStrategies)))

	return nil
}

// subscribeMarketData subscribes to market data from Redis
func (e *Engine) subscribeMarketData() {
	defer e.wg.Done()

	// Use configurable market data channels from config
	channels := e.cfg.Redis.MarketDataChannels
	if len(channels) == 0 {
		// Fallback to default channels if not configured
		channels = []string{"market:btcusdt", "market:ethusdt"}
		e.logger.Warn("No market data channels configured, using defaults", "channels", channels)
	}

	e.logger.Info("Subscribing to market data channels", "channels", channels)

	handler := func(data *strategies.MarketData) error {
		return e.handleMarketData(data)
	}

	if err := e.redisSubscriber.Subscribe(e.ctx, channels, handler); err != nil {
		e.logger.Error("Market data subscription error", "error", err)
	}
}

// subscribeFills subscribes to fill events from NATS
func (e *Engine) subscribeFills() {
	defer e.wg.Done()

	handler := func(fill *strategies.Fill) error {
		return e.handleFill(fill)
	}

	if err := e.natsSubscriber.SubscribeFills(e.ctx, e.cfg.NATS.FillSubject, handler); err != nil {
		e.logger.Error("Fill subscription error", "error", err)
	}
}

// subscribePositions subscribes to position updates from NATS
func (e *Engine) subscribePositions() {
	defer e.wg.Done()

	handler := func(position *strategies.Position) error {
		return e.handlePosition(position)
	}

	if err := e.natsSubscriber.SubscribePositions(e.ctx, e.cfg.NATS.PositionSubject, handler); err != nil {
		e.logger.Error("Position subscription error", "error", err)
	}
}

// handleMarketData handles incoming market data
func (e *Engine) handleMarketData(data *strategies.MarketData) error {
	startTime := time.Now()

	e.strategyMu.RLock()
	strategyList := make([]strategies.Strategy, 0, len(e.activeStrategies))
	for _, strategy := range e.activeStrategies {
		strategyList = append(strategyList, strategy)
	}
	e.strategyMu.RUnlock()

	// Process market data in parallel for all strategies
	var wg sync.WaitGroup
	for _, strategy := range strategyList {
		wg.Add(1)
		go func(s strategies.Strategy) {
			defer wg.Done()

			signals, err := s.OnMarketData(data)
			if err != nil {
				e.logger.Error("Strategy OnMarketData error",
					"strategy", s.Name(),
					"error", err,
				)
				e.metrics.StrategyErrors.WithLabelValues(s.Name(), "market_data").Inc()
				return
			}

			// Queue signals for processing
			for _, signal := range signals {
				select {
				case e.signalQueue <- signal:
					e.metrics.StrategySignals.WithLabelValues(
						signal.Strategy,
						signal.Symbol,
						signal.Side,
					).Inc()
				default:
					e.logger.Warn("Signal queue full, dropping signal",
						"strategy", signal.Strategy,
						"symbol", signal.Symbol,
					)
					e.metrics.SignalsDropped.WithLabelValues(
						signal.Strategy,
						signal.Symbol,
					).Inc()
				}
			}
		}(strategy)
	}

	wg.Wait()

	latency := time.Since(startTime).Microseconds()
	e.metrics.ProcessingTime.WithLabelValues("market_data_handler").Observe(float64(latency))

	return nil
}

// handleFill handles incoming fill events
func (e *Engine) handleFill(fill *strategies.Fill) error {
	e.strategyMu.RLock()
	strategy, exists := e.activeStrategies[fill.Strategy]
	e.strategyMu.RUnlock()

	if !exists {
		return fmt.Errorf("strategy not found: %s", fill.Strategy)
	}

	startTime := time.Now()

	if err := strategy.OnFill(fill); err != nil {
		e.logger.Error("Strategy OnFill error",
			"strategy", fill.Strategy,
			"error", err,
		)
		e.metrics.StrategyErrors.WithLabelValues(fill.Strategy, "fill").Inc()
		return err
	}

	latency := time.Since(startTime).Microseconds()
	e.metrics.StrategyLatency.WithLabelValues(fill.Strategy, "on_fill").Observe(float64(latency))

	// Update risk manager
	e.riskManager.UpdatePnL(fill.Quantity * fill.Price * -1) // Simplified PnL

	return nil
}

// handlePosition handles incoming position updates
func (e *Engine) handlePosition(position *strategies.Position) error {
	e.strategyMu.RLock()
	strategy, exists := e.activeStrategies[position.Strategy]
	e.strategyMu.RUnlock()

	if !exists {
		return fmt.Errorf("strategy not found: %s", position.Strategy)
	}

	startTime := time.Now()

	if err := strategy.OnPositionUpdate(position); err != nil {
		e.logger.Error("Strategy OnPositionUpdate error",
			"strategy", position.Strategy,
			"error", err,
		)
		e.metrics.StrategyErrors.WithLabelValues(position.Strategy, "position").Inc()
		return err
	}

	latency := time.Since(startTime).Microseconds()
	e.metrics.StrategyLatency.WithLabelValues(position.Strategy, "on_position").Observe(float64(latency))

	// Update risk manager
	e.riskManager.UpdatePosition(position.Symbol, position.Quantity)

	return nil
}

// processSignals processes queued signals
func (e *Engine) processSignals() {
	defer e.wg.Done()

	ticker := time.NewTicker(100 * time.Microsecond)
	defer ticker.Stop()

	signals := make([]*strategies.Signal, 0, 100)

	for {
		select {
		case <-e.ctx.Done():
			return

		case <-ticker.C:
			// Collect signals from queue
			signals = signals[:0]
			for len(signals) < 100 {
				select {
				case signal := <-e.signalQueue:
					signals = append(signals, signal)
				default:
					goto process
				}
			}

		process:
			if len(signals) == 0 {
				continue
			}

			e.metrics.SignalQueueSize.Set(float64(len(e.signalQueue)))

			// Sort signals by priority (highest first)
			sort.Slice(signals, func(i, j int) bool {
				return signals[i].Priority > signals[j].Priority
			})

			// Process signals
			validSignals := make([]*strategies.Signal, 0, len(signals))
			for _, signal := range signals {
				// Validate against risk rules
				if err := e.riskManager.ValidateSignal(signal); err != nil {
					e.logger.Warn("Signal rejected by risk manager",
						"strategy", signal.Strategy,
						"symbol", signal.Symbol,
						"reason", err.Error(),
					)
					continue
				}

				validSignals = append(validSignals, signal)
			}

			// Submit valid signals
			if len(validSignals) > 0 {
				if e.cfg.Engine.Mode == "live" {
					if err := e.orderClient.SubmitOrders(e.ctx, validSignals); err != nil {
						e.logger.Error("Failed to submit orders", "error", err)
					}

					// Record orders for rate limiting
					for range validSignals {
						e.riskManager.RecordOrder()
					}
				} else {
					e.logger.Info("Simulation mode - orders not submitted",
						"count", len(validSignals),
					)
				}
			}
		}
	}
}

// hotReloadLoop handles hot reloading of plugins
func (e *Engine) hotReloadLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.cfg.Engine.ReloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return

		case <-ticker.C:
			if err := e.pluginLoader.Reload(); err != nil {
				e.logger.Error("Plugin reload failed", "error", err)
			} else {
				e.metrics.PluginReloads.Inc()
			}
		}
	}
}

// dailyResetLoop resets daily counters
func (e *Engine) dailyResetLoop() {
	defer e.wg.Done()

	// Calculate time until next midnight UTC
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	duration := next.Sub(now)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return

		case <-timer.C:
			e.logger.Info("Performing daily reset")
			e.riskManager.ResetDaily()

			// Reset timer for next day
			timer.Reset(24 * time.Hour)
		}
	}
}

// GetMetrics returns engine metrics
func (e *Engine) GetMetrics() map[string]interface{} {
	e.strategyMu.RLock()
	defer e.strategyMu.RUnlock()

	strategyMetrics := make(map[string]interface{})
	for name, strategy := range e.activeStrategies {
		strategyMetrics[name] = strategy.GetMetrics()
	}

	return map[string]interface{}{
		"active_strategies": len(e.activeStrategies),
		"signal_queue_size": len(e.signalQueue),
		"mode":              e.cfg.Engine.Mode,
		"strategies":        strategyMetrics,
		"risk":              e.riskManager.GetMetrics(),
	}
}
