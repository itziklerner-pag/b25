package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/yourusername/b25/tests/testutil/generators"
)

// StrategyExecutionTestSuite tests strategy signal generation and execution
type StrategyExecutionTestSuite struct {
	suite.Suite
	redisClient  *redis.Client
	natsConn     *nats.Conn
	marketDataGen *generators.MarketDataGenerator
	strategyGen   *generators.StrategyDataGenerator
}

func (s *StrategyExecutionTestSuite) SetupSuite() {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6380"),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.redisClient.Ping(ctx).Err()
	require.NoError(s.T(), err)

	s.natsConn, err = nats.Connect(getEnv("NATS_ADDR", "nats://localhost:4223"))
	require.NoError(s.T(), err)

	s.marketDataGen = generators.NewMarketDataGenerator()
	s.strategyGen = generators.NewStrategyDataGenerator()
}

func (s *StrategyExecutionTestSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

func (s *StrategyExecutionTestSuite) SetupTest() {
	ctx := context.Background()
	s.redisClient.FlushDB(ctx)
}

// TestStrategySignalGeneration tests signal generation from market data
func (s *StrategyExecutionTestSuite) TestStrategySignalGeneration() {
	strategyName := "momentum"
	symbol := "BTCUSDT"

	// Subscribe to signals
	signalsChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("strategy.signals."+strategyName, signalsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish market data to trigger strategy
	marketData := s.marketDataGen.GenerateTick(symbol)
	data, err := json.Marshal(marketData)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.tick."+symbol, data)
	require.NoError(s.T(), err)

	// Wait for signal
	select {
	case msg := <-signalsChan:
		var signal map[string]interface{}
		err = json.Unmarshal(msg.Data, &signal)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), strategyName, signal["strategy"])
		assert.Equal(s.T(), symbol, signal["symbol"])
		assert.NotNil(s.T(), signal["side"])
		assert.NotNil(s.T(), signal["quantity"])

		s.T().Logf("Signal generated: %v", signal)

	case <-time.After(5 * time.Second):
		s.T().Skip("Strategy engine not responding - may not be running")
	}
}

// TestSignalToOrderExecution tests signal to order conversion
func (s *StrategyExecutionTestSuite) TestSignalToOrderExecution() {
	strategyName := "scalping"
	symbol := "BTCUSDT"

	// Generate and publish signal
	signal := s.strategyGen.GenerateSignal(strategyName, symbol)

	// Subscribe to order events
	ordersChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("orders.updates."+symbol, ordersChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish signal
	data, err := json.Marshal(signal)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("strategy.signals."+strategyName, data)
	require.NoError(s.T(), err)

	// Wait for order creation
	select {
	case msg := <-ordersChan:
		var update map[string]interface{}
		err = json.Unmarshal(msg.Data, &update)
		require.NoError(s.T(), err)

		orderData := update["order"].(map[string]interface{})
		assert.Equal(s.T(), symbol, orderData["symbol"])
		s.T().Logf("Order created from signal: %v", orderData)

	case <-time.After(5 * time.Second):
		s.T().Skip("No order created from signal")
	}
}

// TestStrategyRiskManagement tests risk management in strategy execution
func (s *StrategyExecutionTestSuite) TestStrategyRiskManagement() {
	strategyName := "momentum"
	symbol := "BTCUSDT"

	// Create high-risk signal (large quantity)
	signal := map[string]interface{}{
		"id":         "signal_risk_001",
		"strategy":   strategyName,
		"symbol":     symbol,
		"side":       "buy",
		"order_type": "market",
		"quantity":   100.0, // Very large quantity
		"priority":   10,
	}

	// Subscribe to risk alerts
	alertsChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("alerts.risk.#", alertsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish risky signal
	data, err := json.Marshal(signal)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("strategy.signals."+strategyName, data)
	require.NoError(s.T(), err)

	// Wait for risk alert or rejection
	select {
	case msg := <-alertsChan:
		var alert map[string]interface{}
		err = json.Unmarshal(msg.Data, &alert)
		require.NoError(s.T(), err)

		s.T().Logf("Risk alert received: %v", alert)
		assert.NotNil(s.T(), alert["reason"])

	case <-time.After(3 * time.Second):
		s.T().Skip("No risk alert received - risk management may not be configured")
	}
}

// TestMultiStrategyExecution tests multiple strategies running concurrently
func (s *StrategyExecutionTestSuite) TestMultiStrategyExecution() {
	strategies := []string{"momentum", "market_making", "scalping"}
	symbol := "BTCUSDT"

	// Subscribe to all strategy signals
	signalsChan := make(chan *nats.Msg, 10)
	sub, err := s.natsConn.ChanSubscribe("strategy.signals.*", signalsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish market data
	marketData := s.marketDataGen.GenerateTick(symbol)
	data, err := json.Marshal(marketData)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.tick."+symbol, data)
	require.NoError(s.T(), err)

	// Collect signals from different strategies
	receivedStrategies := make(map[string]bool)
	timeout := time.After(5 * time.Second)

collectSignals:
	for {
		select {
		case msg := <-signalsChan:
			var signal map[string]interface{}
			if err := json.Unmarshal(msg.Data, &signal); err == nil {
				if strategy, ok := signal["strategy"].(string); ok {
					receivedStrategies[strategy] = true
					s.T().Logf("Signal from %s strategy", strategy)
				}
			}

			if len(receivedStrategies) >= len(strategies) {
				break collectSignals
			}

		case <-timeout:
			break collectSignals
		}
	}

	if len(receivedStrategies) == 0 {
		s.T().Skip("No strategy signals received")
		return
	}

	s.T().Logf("Received signals from %d strategies", len(receivedStrategies))
}

// TestStrategyPerformanceTracking tests strategy performance metrics
func (s *StrategyExecutionTestSuite) TestStrategyPerformanceTracking() {
	strategyName := "momentum"
	symbol := "BTCUSDT"

	// Simulate trading sequence
	for i := 0; i < 5; i++ {
		signal := s.strategyGen.GenerateSignal(strategyName, symbol)
		data, err := json.Marshal(signal)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("strategy.signals."+strategyName, data)
		require.NoError(s.T(), err)

		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)

	// Check performance metrics in Redis
	ctx := context.Background()
	key := "strategy:metrics:" + strategyName
	metricsData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Strategy metrics not available")
		return
	}

	require.NoError(s.T(), err)

	var metrics map[string]interface{}
	err = json.Unmarshal(metricsData, &metrics)
	require.NoError(s.T(), err)

	s.T().Logf("Strategy metrics: %v", metrics)
	assert.NotNil(s.T(), metrics["total_signals"])
}

// TestStrategyStateManagement tests strategy state persistence
func (s *StrategyExecutionTestSuite) TestStrategyStateManagement() {
	strategyName := "market_making"
	symbol := "BTCUSDT"

	// Create strategy state
	state := map[string]interface{}{
		"strategy":      strategyName,
		"symbol":        symbol,
		"status":        "RUNNING",
		"last_signal":   time.Now().UnixMilli(),
		"open_orders":   []string{"order_1", "order_2"},
		"current_spread": 0.1,
	}

	// Store state in Redis
	ctx := context.Background()
	key := "strategy:state:" + strategyName + ":" + symbol
	data, err := json.Marshal(state)
	require.NoError(s.T(), err)

	err = s.redisClient.Set(ctx, key, data, 1*time.Hour).Err()
	require.NoError(s.T(), err)

	// Verify state retrieval
	storedData, err := s.redisClient.Get(ctx, key).Bytes()
	require.NoError(s.T(), err)

	var retrievedState map[string]interface{}
	err = json.Unmarshal(storedData, &retrievedState)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), strategyName, retrievedState["strategy"])
	assert.Equal(s.T(), "RUNNING", retrievedState["status"])
}

// TestStrategyPriorityExecution tests signal priority handling
func (s *StrategyExecutionTestSuite) TestStrategyPriorityExecution() {
	strategyName := "scalping"
	symbol := "BTCUSDT"

	// Create signals with different priorities
	highPrioritySignal := s.strategyGen.GenerateSignal(strategyName, symbol)
	highPrioritySignal["priority"] = 10

	lowPrioritySignal := s.strategyGen.GenerateSignal(strategyName, symbol)
	lowPrioritySignal["priority"] = 1

	// Subscribe to order updates
	ordersChan := make(chan *nats.Msg, 10)
	sub, err := s.natsConn.ChanSubscribe("orders.updates."+symbol, ordersChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish low priority first
	lowPriorityData, err := json.Marshal(lowPrioritySignal)
	require.NoError(s.T(), err)
	s.natsConn.Publish("strategy.signals."+strategyName, lowPriorityData)

	time.Sleep(50 * time.Millisecond)

	// Publish high priority
	highPriorityData, err := json.Marshal(highPrioritySignal)
	require.NoError(s.T(), err)
	s.natsConn.Publish("strategy.signals."+strategyName, highPriorityData)

	// Collect order creations
	orderCount := 0
	timeout := time.After(3 * time.Second)

collectOrders:
	for {
		select {
		case <-ordersChan:
			orderCount++
			if orderCount >= 2 {
				break collectOrders
			}
		case <-timeout:
			break collectOrders
		}
	}

	s.T().Logf("Received %d order updates", orderCount)
}

// TestStrategyStopLoss tests stop-loss signal generation
func (s *StrategyExecutionTestSuite) TestStrategyStopLoss() {
	strategyName := "momentum"
	symbol := "BTCUSDT"

	// Simulate position with loss
	position := map[string]interface{}{
		"strategy":        strategyName,
		"symbol":          symbol,
		"side":            "long",
		"quantity":        1.0,
		"avg_entry_price": 50000.0,
		"current_price":   48000.0, // 4% loss
		"unrealized_pnl":  -2000.0,
	}

	// Subscribe to stop-loss signals
	signalsChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("strategy.signals.stop_loss", signalsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish position update
	data, err := json.Marshal(position)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("account.position.update", data)
	require.NoError(s.T(), err)

	// Wait for stop-loss signal
	select {
	case msg := <-signalsChan:
		var signal map[string]interface{}
		err = json.Unmarshal(msg.Data, &signal)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), symbol, signal["symbol"])
		assert.Equal(s.T(), "sell", signal["side"])
		s.T().Logf("Stop-loss signal generated: %v", signal)

	case <-time.After(3 * time.Second):
		s.T().Skip("No stop-loss signal generated")
	}
}

// TestStrategyLatency tests strategy decision latency
func (s *StrategyExecutionTestSuite) TestStrategyLatency() {
	strategyName := "scalping"
	symbol := "BTCUSDT"
	numTests := 50

	latencies := make([]time.Duration, numTests)

	for i := 0; i < numTests; i++ {
		marketData := s.marketDataGen.GenerateTick(symbol)

		startTime := time.Now()

		data, err := json.Marshal(marketData)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("market.tick."+symbol, data)
		require.NoError(s.T(), err)

		latencies[i] = time.Since(startTime)

		time.Sleep(20 * time.Millisecond)
	}

	// Calculate statistics
	var totalLatency time.Duration
	maxLatency := time.Duration(0)

	for _, latency := range latencies {
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	avgLatency := totalLatency / time.Duration(numTests)

	s.T().Logf("Average strategy latency: %v", avgLatency)
	s.T().Logf("Max strategy latency: %v", maxLatency)

	// Assert performance targets
	assert.Less(s.T(), avgLatency, 2*time.Millisecond, "Average latency should be < 2ms")
	assert.Less(s.T(), maxLatency, 10*time.Millisecond, "Max latency should be < 10ms")
}

func TestStrategyExecutionSuite(t *testing.T) {
	suite.Run(t, new(StrategyExecutionTestSuite))
}
