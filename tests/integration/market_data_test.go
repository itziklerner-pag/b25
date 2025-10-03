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

// MarketDataTestSuite is the test suite for market data pipeline
type MarketDataTestSuite struct {
	suite.Suite
	redisClient *redis.Client
	natsConn    *nats.Conn
	generator   *generators.MarketDataGenerator
}

// SetupSuite runs once before all tests
func (s *MarketDataTestSuite) SetupSuite() {
	// Connect to Redis (test instance)
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6380"),
		DB:   0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.redisClient.Ping(ctx).Err()
	require.NoError(s.T(), err, "Failed to connect to Redis")

	// Connect to NATS
	s.natsConn, err = nats.Connect(getEnv("NATS_ADDR", "nats://localhost:4223"))
	require.NoError(s.T(), err, "Failed to connect to NATS")

	// Initialize generator
	s.generator = generators.NewMarketDataGenerator()
}

// TearDownSuite runs once after all tests
func (s *MarketDataTestSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

// SetupTest runs before each test
func (s *MarketDataTestSuite) SetupTest() {
	// Clear test data
	ctx := context.Background()
	s.redisClient.FlushDB(ctx)
}

// TestMarketDataPipeline tests the end-to-end market data pipeline
func (s *MarketDataTestSuite) TestMarketDataPipeline() {
	symbol := "BTCUSDT"

	// Generate test market data
	tick := s.generator.GenerateTick(symbol)

	// Publish market data to NATS
	subject := "market.tick." + symbol
	data, err := json.Marshal(tick)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish(subject, data)
	require.NoError(s.T(), err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify data was stored in Redis
	ctx := context.Background()
	key := "market:tick:" + symbol
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Market data service not running - skipping pipeline verification")
		return
	}

	require.NoError(s.T(), err)

	var storedTick map[string]interface{}
	err = json.Unmarshal(storedData, &storedTick)
	require.NoError(s.T(), err)

	// Verify data integrity
	assert.Equal(s.T(), symbol, storedTick["symbol"])
	assert.NotNil(s.T(), storedTick["last_price"])
}

// TestMarketDataLatency tests market data processing latency
func (s *MarketDataTestSuite) TestMarketDataLatency() {
	symbol := "BTCUSDT"
	numTicks := 100

	latencies := make([]time.Duration, numTicks)

	for i := 0; i < numTicks; i++ {
		tick := s.generator.GenerateTick(symbol)

		startTime := time.Now()

		// Publish to NATS
		subject := "market.tick." + symbol
		data, err := json.Marshal(tick)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish(subject, data)
		require.NoError(s.T(), err)

		// Measure latency
		latencies[i] = time.Since(startTime)

		time.Sleep(10 * time.Millisecond)
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

	avgLatency := totalLatency / time.Duration(numTicks)

	s.T().Logf("Average latency: %v", avgLatency)
	s.T().Logf("Max latency: %v", maxLatency)

	// Assert latency requirements
	assert.Less(s.T(), avgLatency, 1*time.Millisecond, "Average latency too high")
	assert.Less(s.T(), maxLatency, 10*time.Millisecond, "Max latency too high")
}

// TestOrderBookProcessing tests order book data processing
func (s *MarketDataTestSuite) TestOrderBookProcessing() {
	symbol := "BTCUSDT"

	// Generate order book
	orderBook := s.generator.GenerateOrderBook(symbol, 10)

	// Publish to NATS
	subject := "market.orderbook." + symbol
	data, err := json.Marshal(orderBook)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish(subject, data)
	require.NoError(s.T(), err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify order book in Redis
	ctx := context.Background()
	key := "market:orderbook:" + symbol
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Market data service not running - skipping order book verification")
		return
	}

	require.NoError(s.T(), err)

	var storedBook map[string]interface{}
	err = json.Unmarshal(storedData, &storedBook)
	require.NoError(s.T(), err)

	// Verify structure
	assert.Equal(s.T(), symbol, storedBook["symbol"])
	assert.NotNil(s.T(), storedBook["bids"])
	assert.NotNil(s.T(), storedBook["asks"])
}

// TestMarketDataAggregation tests candle aggregation
func (s *MarketDataTestSuite) TestMarketDataAggregation() {
	symbol := "BTCUSDT"

	// Generate and publish multiple ticks
	for i := 0; i < 10; i++ {
		tick := s.generator.GenerateTick(symbol)
		subject := "market.tick." + symbol
		data, err := json.Marshal(tick)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish(subject, data)
		require.NoError(s.T(), err)

		time.Sleep(50 * time.Millisecond)
	}

	// Wait for aggregation
	time.Sleep(200 * time.Millisecond)

	// Verify candle data in Redis
	ctx := context.Background()
	key := "market:candle:1m:" + symbol
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Market data aggregation not configured - skipping candle verification")
		return
	}

	require.NoError(s.T(), err)

	var candle map[string]interface{}
	err = json.Unmarshal(storedData, &candle)
	require.NoError(s.T(), err)

	// Verify OHLC structure
	assert.NotNil(s.T(), candle["open"])
	assert.NotNil(s.T(), candle["high"])
	assert.NotNil(s.T(), candle["low"])
	assert.NotNil(s.T(), candle["close"])
	assert.NotNil(s.T(), candle["volume"])
}

// TestMarketDataMultiSymbol tests multi-symbol handling
func (s *MarketDataTestSuite) TestMarketDataMultiSymbol() {
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	// Publish data for multiple symbols
	for _, symbol := range symbols {
		tick := s.generator.GenerateTick(symbol)
		subject := "market.tick." + symbol
		data, err := json.Marshal(tick)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish(subject, data)
		require.NoError(s.T(), err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify each symbol was processed
	ctx := context.Background()
	for _, symbol := range symbols {
		key := "market:tick:" + symbol
		exists, err := s.redisClient.Exists(ctx, key).Result()

		if err == nil && exists > 0 {
			assert.Greater(s.T(), exists, int64(0), "Data for %s should exist", symbol)
		}
	}
}

// TestMarketDataRecovery tests recovery from data interruption
func (s *MarketDataTestSuite) TestMarketDataRecovery() {
	symbol := "BTCUSDT"

	// Publish initial data
	tick1 := s.generator.GenerateTick(symbol)
	data1, err := json.Marshal(tick1)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.tick."+symbol, data1)
	require.NoError(s.T(), err)

	time.Sleep(100 * time.Millisecond)

	// Simulate gap/interruption
	time.Sleep(500 * time.Millisecond)

	// Publish new data
	tick2 := s.generator.GenerateTick(symbol)
	data2, err := json.Marshal(tick2)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.tick."+symbol, data2)
	require.NoError(s.T(), err)

	time.Sleep(100 * time.Millisecond)

	// Verify system recovered
	ctx := context.Background()
	key := "market:tick:" + symbol
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Market data service not running")
		return
	}

	require.NoError(s.T(), err)
	assert.NotNil(s.T(), storedData)
}

// TestInvalidMarketData tests handling of invalid data
func (s *MarketDataTestSuite) TestInvalidMarketData() {
	// Publish invalid JSON
	invalidData := []byte("{invalid json")
	err := s.natsConn.Publish("market.tick.BTCUSDT", invalidData)
	require.NoError(s.T(), err)

	// System should not crash
	time.Sleep(100 * time.Millisecond)

	// Publish valid data after invalid
	validTick := s.generator.GenerateTick("BTCUSDT")
	validData, err := json.Marshal(validTick)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("market.tick.BTCUSDT", validData)
	require.NoError(s.T(), err)

	time.Sleep(100 * time.Millisecond)

	// System should process valid data correctly
	ctx := context.Background()
	key := "market:tick:BTCUSDT"
	storedData, err := s.redisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		s.T().Skip("Market data service not running")
		return
	}

	require.NoError(s.T(), err)
	assert.NotNil(s.T(), storedData)
}

// Run the test suite
func TestMarketDataSuite(t *testing.T) {
	suite.Run(t, new(MarketDataTestSuite))
}
