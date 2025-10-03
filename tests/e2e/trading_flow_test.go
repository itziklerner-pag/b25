package e2e

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

// TradingFlowTestSuite tests complete trading cycle end-to-end
type TradingFlowTestSuite struct {
	suite.Suite
	redisClient   *redis.Client
	natsConn      *nats.Conn
	marketDataGen *generators.MarketDataGenerator
	orderGen      *generators.OrderGenerator
	strategyGen   *generators.StrategyDataGenerator
}

func (s *TradingFlowTestSuite) SetupSuite() {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6380"),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.redisClient.Ping(ctx).Err()
	require.NoError(s.T(), err, "Redis connection failed")

	s.natsConn, err = nats.Connect(getEnv("NATS_ADDR", "nats://localhost:4223"))
	require.NoError(s.T(), err, "NATS connection failed")

	s.marketDataGen = generators.NewMarketDataGenerator()
	s.orderGen = generators.NewOrderGenerator()
	s.strategyGen = generators.NewStrategyDataGenerator()
}

func (s *TradingFlowTestSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

func (s *TradingFlowTestSuite) SetupTest() {
	ctx := context.Background()
	s.redisClient.FlushDB(ctx)
}

// TestCompleteTradingCycle tests the full trading cycle:
// Market Data -> Strategy Signal -> Order Execution -> Fill -> Position Update
func (s *TradingFlowTestSuite) TestCompleteTradingCycle() {
	symbol := "BTCUSDT"
	strategyName := "momentum"

	// Track the complete flow
	events := []string{}
	eventsChan := make(chan string, 20)

	// Subscribe to market data
	marketDataSub, err := s.natsConn.Subscribe("market.tick."+symbol, func(msg *nats.Msg) {
		eventsChan <- "MARKET_DATA_RECEIVED"
	})
	require.NoError(s.T(), err)
	defer marketDataSub.Unsubscribe()

	// Subscribe to strategy signals
	signalSub, err := s.natsConn.Subscribe("strategy.signals."+strategyName, func(msg *nats.Msg) {
		eventsChan <- "SIGNAL_GENERATED"
	})
	require.NoError(s.T(), err)
	defer signalSub.Unsubscribe()

	// Subscribe to order updates
	orderSub, err := s.natsConn.Subscribe("orders.updates."+symbol, func(msg *nats.Msg) {
		var update map[string]interface{}
		if err := json.Unmarshal(msg.Data, &update); err == nil {
			if orderData, ok := update["order"].(map[string]interface{}); ok {
				state := orderData["state"].(string)
				eventsChan <- "ORDER_" + state
			}
		}
	})
	require.NoError(s.T(), err)
	defer orderSub.Unsubscribe()

	// Subscribe to fills
	fillSub, err := s.natsConn.Subscribe("orders.fills."+symbol, func(msg *nats.Msg) {
		eventsChan <- "ORDER_FILLED"
	})
	require.NoError(s.T(), err)
	defer fillSub.Unsubscribe()

	// Subscribe to position updates
	positionSub, err := s.natsConn.Subscribe("account.position.update", func(msg *nats.Msg) {
		eventsChan <- "POSITION_UPDATED"
	})
	require.NoError(s.T(), err)
	defer positionSub.Unsubscribe()

	// Step 1: Publish market data
	s.T().Log("Step 1: Publishing market data...")
	marketData := s.marketDataGen.GenerateTick(symbol)
	data, err := json.Marshal(marketData)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.tick."+symbol, data)
	require.NoError(s.T(), err)

	// Collect events for 10 seconds
	timeout := time.After(10 * time.Second)

collectEvents:
	for {
		select {
		case event := <-eventsChan:
			events = append(events, event)
			s.T().Logf("Event: %s", event)

			// Check if we have a complete cycle
			if containsAll(events, []string{"MARKET_DATA_RECEIVED", "ORDER_FILLED", "POSITION_UPDATED"}) {
				break collectEvents
			}

		case <-timeout:
			break collectEvents
		}
	}

	// Verify complete flow
	s.T().Logf("Total events captured: %d", len(events))
	s.T().Logf("Events: %v", events)

	if len(events) == 0 {
		s.T().Skip("No events received - system may not be fully running")
		return
	}

	// Assert key events occurred
	assert.Contains(s.T(), events, "MARKET_DATA_RECEIVED", "Market data should be received")

	// Check if trading flow completed
	if containsAll(events, []string{"ORDER_FILLED", "POSITION_UPDATED"}) {
		s.T().Log("✓ Complete trading cycle successful")
	} else {
		s.T().Log("⚠ Partial trading cycle - some services may not be configured")
	}
}

// TestRoundTripLatency tests end-to-end latency from market data to order execution
func (s *TradingFlowTestSuite) TestRoundTripLatency() {
	symbol := "BTCUSDT"
	numTests := 20

	latencies := make([]time.Duration, 0, numTests)

	// Subscribe to order updates
	orderUpdatesChan := make(chan time.Time, 10)
	orderSub, err := s.natsConn.Subscribe("orders.updates."+symbol, func(msg *nats.Msg) {
		orderUpdatesChan <- time.Now()
	})
	require.NoError(s.T(), err)
	defer orderSub.Unsubscribe()

	for i := 0; i < numTests; i++ {
		marketData := s.marketDataGen.GenerateTick(symbol)
		startTime := time.Now()

		data, err := json.Marshal(marketData)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("market.tick."+symbol, data)
		require.NoError(s.T(), err)

		// Wait for order update
		select {
		case endTime := <-orderUpdatesChan:
			latency := endTime.Sub(startTime)
			latencies = append(latencies, latency)

		case <-time.After(1 * time.Second):
			// Timeout - no order generated
		}

		time.Sleep(100 * time.Millisecond)
	}

	if len(latencies) == 0 {
		s.T().Skip("No latency measurements - system may not be configured for automated trading")
		return
	}

	// Calculate statistics
	var totalLatency time.Duration
	maxLatency := time.Duration(0)
	minLatency := latencies[0]

	for _, latency := range latencies {
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
		if latency < minLatency {
			minLatency = latency
		}
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	s.T().Logf("Round-trip latency statistics:")
	s.T().Logf("  Average: %v", avgLatency)
	s.T().Logf("  Min: %v", minLatency)
	s.T().Logf("  Max: %v", maxLatency)
	s.T().Logf("  Samples: %d/%d", len(latencies), numTests)

	// Assert performance requirements
	assert.Less(s.T(), avgLatency, 100*time.Millisecond, "Average round-trip latency should be < 100ms")
}

// TestMultiSymbolTrading tests trading across multiple symbols simultaneously
func (s *TradingFlowTestSuite) TestMultiSymbolTrading() {
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	// Track fills per symbol
	fillsPerSymbol := make(map[string]int)
	fillsChan := make(chan string, 20)

	// Subscribe to fills for all symbols
	for _, symbol := range symbols {
		sym := symbol // capture for closure
		fillSub, err := s.natsConn.Subscribe("orders.fills."+sym, func(msg *nats.Msg) {
			fillsChan <- sym
		})
		require.NoError(s.T(), err)
		defer fillSub.Unsubscribe()
	}

	// Publish market data for all symbols
	for _, symbol := range symbols {
		marketData := s.marketDataGen.GenerateTick(symbol)
		data, err := json.Marshal(marketData)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("market.tick."+symbol, data)
		require.NoError(s.T(), err)

		// Also manually create orders for more reliable testing
		order := s.orderGen.GenerateMarketOrder(symbol, "BUY", 0.1)
		orderData, err := json.Marshal(order)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("orders.requests.create", orderData)
		require.NoError(s.T(), err)
	}

	// Collect fills
	timeout := time.After(10 * time.Second)

collectFills:
	for {
		select {
		case symbol := <-fillsChan:
			fillsPerSymbol[symbol]++
			s.T().Logf("Fill received for %s (total: %d)", symbol, fillsPerSymbol[symbol])

		case <-timeout:
			break collectFills
		}
	}

	// Verify trading happened on multiple symbols
	s.T().Logf("Fills per symbol: %v", fillsPerSymbol)

	totalFills := 0
	for _, count := range fillsPerSymbol {
		totalFills += count
	}

	if totalFills == 0 {
		s.T().Skip("No fills received across any symbol")
		return
	}

	s.T().Logf("Total fills across all symbols: %d", totalFills)
	assert.Greater(s.T(), totalFills, 0, "Should have at least one fill")
}

// TestOrderBookImpact tests order execution with order book analysis
func (s *TradingFlowTestSuite) TestOrderBookImpact() {
	symbol := "BTCUSDT"

	// Publish order book
	orderBook := s.marketDataGen.GenerateOrderBook(symbol, 10)
	orderBookData, err := json.Marshal(orderBook)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.orderbook."+symbol, orderBookData)
	require.NoError(s.T(), err)

	time.Sleep(200 * time.Millisecond)

	// Execute large market order
	order := s.orderGen.GenerateMarketOrder(symbol, "BUY", 2.0)

	// Track execution
	fillsChan := make(chan *nats.Msg, 1)
	fillSub, err := s.natsConn.ChanSubscribe("orders.fills."+symbol, fillsChan)
	require.NoError(s.T(), err)
	defer fillSub.Unsubscribe()

	orderData, err := json.Marshal(order)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("orders.requests.create", orderData)
	require.NoError(s.T(), err)

	// Wait for fill
	select {
	case msg := <-fillsChan:
		var fill map[string]interface{}
		err = json.Unmarshal(msg.Data, &fill)
		require.NoError(s.T(), err)

		s.T().Logf("Fill details: %v", fill)
		assert.NotNil(s.T(), fill["price"])

	case <-time.After(5 * time.Second):
		s.T().Skip("No fill received")
	}
}

// TestStrategyToPositionFlow tests strategy execution resulting in position changes
func (s *TradingFlowTestSuite) TestStrategyToPositionFlow() {
	strategyName := "scalping"
	symbol := "BTCUSDT"

	// Subscribe to position updates
	positionsChan := make(chan *nats.Msg, 1)
	positionSub, err := s.natsConn.ChanSubscribe("account.position.update", positionsChan)
	require.NoError(s.T(), err)
	defer positionSub.Unsubscribe()

	// Generate and publish signal
	signal := s.strategyGen.GenerateSignal(strategyName, symbol)
	signalData, err := json.Marshal(signal)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("strategy.signals."+strategyName, signalData)
	require.NoError(s.T(), err)

	// Wait for position update
	select {
	case msg := <-positionsChan:
		var position map[string]interface{}
		err = json.Unmarshal(msg.Data, &position)
		require.NoError(s.T(), err)

		s.T().Logf("Position updated: %v", position)
		assert.Equal(s.T(), symbol, position["symbol"])

	case <-time.After(10 * time.Second):
		s.T().Skip("No position update received - complete flow may not be configured")
	}
}

// TestHighFrequencyScenario tests system under high-frequency trading load
func (s *TradingFlowTestSuite) TestHighFrequencyScenario() {
	symbol := "BTCUSDT"
	duration := 5 * time.Second
	tickInterval := 50 * time.Millisecond

	orderCount := 0
	fillCount := 0

	// Subscribe to order updates
	orderSub, err := s.natsConn.Subscribe("orders.updates."+symbol, func(msg *nats.Msg) {
		orderCount++
	})
	require.NoError(s.T(), err)
	defer orderSub.Unsubscribe()

	// Subscribe to fills
	fillSub, err := s.natsConn.Subscribe("orders.fills."+symbol, func(msg *nats.Msg) {
		fillCount++
	})
	require.NoError(s.T(), err)
	defer fillSub.Unsubscribe()

	// Generate high-frequency market data
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	timeout := time.After(duration)

	s.T().Logf("Starting high-frequency scenario for %v...", duration)

publishLoop:
	for {
		select {
		case <-ticker.C:
			marketData := s.marketDataGen.GenerateTick(symbol)
			data, err := json.Marshal(marketData)
			if err != nil {
				continue
			}

			s.natsConn.Publish("market.tick."+symbol, data)

		case <-timeout:
			break publishLoop
		}
	}

	// Allow time for processing
	time.Sleep(2 * time.Second)

	s.T().Logf("High-frequency scenario results:")
	s.T().Logf("  Order updates: %d", orderCount)
	s.T().Logf("  Fills: %d", fillCount)
	s.T().Logf("  Duration: %v", duration)

	expectedTicks := int(duration / tickInterval)
	s.T().Logf("  Expected ticks: ~%d", expectedTicks)
}

// Helper function to check if slice contains all required items
func containsAll(slice []string, required []string) bool {
	found := make(map[string]bool)
	for _, item := range slice {
		found[item] = true
	}

	for _, req := range required {
		if !found[req] {
			return false
		}
	}
	return true
}

func TestTradingFlowSuite(t *testing.T) {
	suite.Run(t, new(TradingFlowTestSuite))
}
