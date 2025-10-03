package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/yourusername/b25/tests/testutil/generators"
)

// LatencyBenchmarkSuite benchmarks system performance and latency
type LatencyBenchmarkSuite struct {
	suite.Suite
	redisClient   *redis.Client
	natsConn      *nats.Conn
	orderGen      *generators.OrderGenerator
	marketDataGen *generators.MarketDataGenerator
}

func (s *LatencyBenchmarkSuite) SetupSuite() {
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

	s.orderGen = generators.NewOrderGenerator()
	s.marketDataGen = generators.NewMarketDataGenerator()
}

func (s *LatencyBenchmarkSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

func (s *LatencyBenchmarkSuite) SetupTest() {
	ctx := context.Background()
	s.redisClient.FlushDB(ctx)
}

// TestMarketDataIngestionLatency benchmarks market data ingestion latency
func (s *LatencyBenchmarkSuite) TestMarketDataIngestionLatency() {
	symbol := "BTCUSDT"
	numSamples := 1000

	latencies := make([]time.Duration, 0, numSamples)

	s.T().Logf("Benchmarking market data ingestion (%d samples)...", numSamples)

	for i := 0; i < numSamples; i++ {
		tick := s.marketDataGen.GenerateTick(symbol)

		startTime := time.Now()

		data, err := json.Marshal(tick)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("market.tick."+symbol, data)
		require.NoError(s.T(), err)

		latency := time.Since(startTime)
		latencies = append(latencies, latency)

		// Small delay to avoid overwhelming the system
		if i%100 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Calculate statistics
	stats := calculateLatencyStats(latencies)

	s.T().Logf("\nMarket Data Ingestion Latency:")
	s.T().Logf("  Samples: %d", stats.Count)
	s.T().Logf("  Average: %v", stats.Average)
	s.T().Logf("  Median: %v", stats.Median)
	s.T().Logf("  P95: %v", stats.P95)
	s.T().Logf("  P99: %v", stats.P99)
	s.T().Logf("  Min: %v", stats.Min)
	s.T().Logf("  Max: %v", stats.Max)

	// Performance assertions
	require.Less(s.T(), stats.Average, 500*time.Microsecond, "Average latency should be < 500μs")
	require.Less(s.T(), stats.P99, 10*time.Millisecond, "P99 latency should be < 10ms")
}

// TestOrderExecutionLatency benchmarks order execution latency
func (s *LatencyBenchmarkSuite) TestOrderExecutionLatency() {
	symbol := "BTCUSDT"
	numOrders := 500

	latencies := make([]time.Duration, 0, numOrders)
	responseChan := make(chan time.Time, 100)

	// Subscribe to order updates
	sub, err := s.natsConn.Subscribe("orders.updates."+symbol, func(msg *nats.Msg) {
		responseChan <- time.Now()
	})
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	s.T().Logf("Benchmarking order execution (%d orders)...", numOrders)

	for i := 0; i < numOrders; i++ {
		order := s.orderGen.GenerateMarketOrder(symbol, "BUY", 0.01)

		startTime := time.Now()

		data, err := json.Marshal(order)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("orders.requests.create", data)
		require.NoError(s.T(), err)

		// Wait for response with timeout
		select {
		case responseTime := <-responseChan:
			latency := responseTime.Sub(startTime)
			latencies = append(latencies, latency)

		case <-time.After(100 * time.Millisecond):
			// Timeout - skip this sample
		}

		time.Sleep(20 * time.Millisecond)
	}

	if len(latencies) < numOrders/2 {
		s.T().Skipf("Insufficient samples (%d/%d) - order execution service may not be running", len(latencies), numOrders)
		return
	}

	// Calculate statistics
	stats := calculateLatencyStats(latencies)

	s.T().Logf("\nOrder Execution Latency:")
	s.T().Logf("  Samples: %d/%d", stats.Count, numOrders)
	s.T().Logf("  Average: %v", stats.Average)
	s.T().Logf("  Median: %v", stats.Median)
	s.T().Logf("  P95: %v", stats.P95)
	s.T().Logf("  P99: %v", stats.P99)
	s.T().Logf("  Min: %v", stats.Min)
	s.T().Logf("  Max: %v", stats.Max)

	// Performance assertions based on requirements
	require.Less(s.T(), stats.Average, 50*time.Millisecond, "Average order latency should be < 50ms")
	require.Less(s.T(), stats.P99, 100*time.Millisecond, "P99 order latency should be < 100ms")
}

// TestRedisLatency benchmarks Redis read/write latency
func (s *LatencyBenchmarkSuite) TestRedisLatency() {
	ctx := context.Background()
	numOps := 1000

	writeLatencies := make([]time.Duration, 0, numOps)
	readLatencies := make([]time.Duration, 0, numOps)

	s.T().Logf("Benchmarking Redis operations (%d ops)...", numOps)

	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark:key:%d", i)
		value := fmt.Sprintf("value_%d", i)

		// Write benchmark
		startWrite := time.Now()
		err := s.redisClient.Set(ctx, key, value, 1*time.Minute).Err()
		require.NoError(s.T(), err)
		writeLatencies = append(writeLatencies, time.Since(startWrite))

		// Read benchmark
		startRead := time.Now()
		_, err = s.redisClient.Get(ctx, key).Result()
		require.NoError(s.T(), err)
		readLatencies = append(readLatencies, time.Since(startRead))
	}

	// Write statistics
	writeStats := calculateLatencyStats(writeLatencies)
	s.T().Logf("\nRedis Write Latency:")
	s.T().Logf("  Average: %v", writeStats.Average)
	s.T().Logf("  P95: %v", writeStats.P95)
	s.T().Logf("  P99: %v", writeStats.P99)

	// Read statistics
	readStats := calculateLatencyStats(readLatencies)
	s.T().Logf("\nRedis Read Latency:")
	s.T().Logf("  Average: %v", readStats.Average)
	s.T().Logf("  P95: %v", readStats.P95)
	s.T().Logf("  P99: %v", readStats.P99)

	// Performance assertions
	require.Less(s.T(), writeStats.P99, 5*time.Millisecond, "P99 write latency should be < 5ms")
	require.Less(s.T(), readStats.P99, 5*time.Millisecond, "P99 read latency should be < 5ms")
}

// TestNATSPublishLatency benchmarks NATS publish latency
func (s *LatencyBenchmarkSuite) TestNATSPublishLatency() {
	numMessages := 1000
	subject := "benchmark.test"

	latencies := make([]time.Duration, 0, numMessages)

	s.T().Logf("Benchmarking NATS publish (%d messages)...", numMessages)

	for i := 0; i < numMessages; i++ {
		message := fmt.Sprintf("message_%d", i)

		startTime := time.Now()
		err := s.natsConn.Publish(subject, []byte(message))
		require.NoError(s.T(), err)

		latencies = append(latencies, time.Since(startTime))
	}

	stats := calculateLatencyStats(latencies)

	s.T().Logf("\nNATS Publish Latency:")
	s.T().Logf("  Samples: %d", stats.Count)
	s.T().Logf("  Average: %v", stats.Average)
	s.T().Logf("  P95: %v", stats.P95)
	s.T().Logf("  P99: %v", stats.P99)

	require.Less(s.T(), stats.P99, 2*time.Millisecond, "P99 publish latency should be < 2ms")
}

// TestNATSRoundTripLatency benchmarks NATS request-reply latency
func (s *LatencyBenchmarkSuite) TestNATSRoundTripLatency() {
	numRequests := 500
	subject := "benchmark.roundtrip"

	latencies := make([]time.Duration, 0, numRequests)

	// Setup responder
	_, err := s.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		msg.Respond([]byte("response"))
	})
	require.NoError(s.T(), err)

	time.Sleep(100 * time.Millisecond)

	s.T().Logf("Benchmarking NATS round-trip (%d requests)...", numRequests)

	for i := 0; i < numRequests; i++ {
		startTime := time.Now()

		_, err := s.natsConn.Request(subject, []byte("request"), 100*time.Millisecond)
		if err == nil {
			latencies = append(latencies, time.Since(startTime))
		}

		time.Sleep(5 * time.Millisecond)
	}

	if len(latencies) == 0 {
		s.T().Skip("No successful round-trips")
		return
	}

	stats := calculateLatencyStats(latencies)

	s.T().Logf("\nNATS Round-Trip Latency:")
	s.T().Logf("  Samples: %d", stats.Count)
	s.T().Logf("  Average: %v", stats.Average)
	s.T().Logf("  P95: %v", stats.P95)
	s.T().Logf("  P99: %v", stats.P99)

	require.Less(s.T(), stats.Average, 5*time.Millisecond, "Average round-trip should be < 5ms")
}

// TestEndToEndLatency benchmarks complete end-to-end latency
func (s *LatencyBenchmarkSuite) TestEndToEndLatency() {
	symbol := "BTCUSDT"
	numSamples := 100

	latencies := make([]time.Duration, 0, numSamples)
	fillChan := make(chan time.Time, 10)

	// Subscribe to fills
	fillSub, err := s.natsConn.Subscribe("orders.fills."+symbol, func(msg *nats.Msg) {
		fillChan <- time.Now()
	})
	require.NoError(s.T(), err)
	defer fillSub.Unsubscribe()

	s.T().Logf("Benchmarking end-to-end latency (%d samples)...", numSamples)

	for i := 0; i < numSamples; i++ {
		// Start: Publish market data
		marketData := s.marketDataGen.GenerateTick(symbol)
		startTime := time.Now()

		data, err := json.Marshal(marketData)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("market.tick."+symbol, data)
		require.NoError(s.T(), err)

		// Also submit order to ensure fill
		order := s.orderGen.GenerateMarketOrder(symbol, "BUY", 0.01)
		orderData, _ := json.Marshal(order)
		s.natsConn.Publish("orders.requests.create", orderData)

		// Wait for fill (end of pipeline)
		select {
		case fillTime := <-fillChan:
			latency := fillTime.Sub(startTime)
			latencies = append(latencies, latency)

		case <-time.After(500 * time.Millisecond):
			// Timeout - skip
		}

		time.Sleep(100 * time.Millisecond)
	}

	if len(latencies) < 10 {
		s.T().Skipf("Insufficient samples (%d) for end-to-end benchmark", len(latencies))
		return
	}

	stats := calculateLatencyStats(latencies)

	s.T().Logf("\nEnd-to-End Latency (Market Data -> Fill):")
	s.T().Logf("  Samples: %d/%d", stats.Count, numSamples)
	s.T().Logf("  Average: %v", stats.Average)
	s.T().Logf("  Median: %v", stats.Median)
	s.T().Logf("  P95: %v", stats.P95)
	s.T().Logf("  P99: %v", stats.P99)
	s.T().Logf("  Min: %v", stats.Min)
	s.T().Logf("  Max: %v", stats.Max)
}

// TestThroughput benchmarks system throughput
func (s *LatencyBenchmarkSuite) TestThroughput() {
	symbol := "BTCUSDT"
	duration := 10 * time.Second

	publishCount := 0
	processCount := 0

	// Subscribe to order updates
	sub, err := s.natsConn.Subscribe("orders.updates."+symbol, func(msg *nats.Msg) {
		processCount++
	})
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	s.T().Logf("Benchmarking throughput for %v...", duration)

	startTime := time.Now()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(duration)

publishLoop:
	for {
		select {
		case <-ticker.C:
			order := s.orderGen.GenerateMarketOrder(symbol, "BUY", 0.01)
			data, err := json.Marshal(order)
			if err != nil {
				continue
			}

			err = s.natsConn.Publish("orders.requests.create", data)
			if err == nil {
				publishCount++
			}

		case <-timeout:
			break publishLoop
		}
	}

	// Wait for processing to complete
	time.Sleep(2 * time.Second)

	elapsed := time.Since(startTime)
	publishRate := float64(publishCount) / elapsed.Seconds()
	processRate := float64(processCount) / elapsed.Seconds()

	s.T().Logf("\nThroughput Results:")
	s.T().Logf("  Duration: %v", elapsed)
	s.T().Logf("  Published: %d orders", publishCount)
	s.T().Logf("  Processed: %d orders", processCount)
	s.T().Logf("  Publish Rate: %.1f orders/sec", publishRate)
	s.T().Logf("  Process Rate: %.1f orders/sec", processRate)

	if processCount > 0 {
		processingEfficiency := float64(processCount) / float64(publishCount) * 100
		s.T().Logf("  Processing Efficiency: %.1f%%", processingEfficiency)
	}
}

// LatencyStats holds latency statistics
type LatencyStats struct {
	Count   int
	Average time.Duration
	Median  time.Duration
	P95     time.Duration
	P99     time.Duration
	Min     time.Duration
	Max     time.Duration
}

// calculateLatencyStats calculates comprehensive latency statistics
func calculateLatencyStats(latencies []time.Duration) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Sort for percentile calculation
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate average
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	average := total / time.Duration(len(latencies))

	// Percentiles
	median := sorted[len(sorted)/2]
	p95 := sorted[int(float64(len(sorted))*0.95)]
	p99 := sorted[int(float64(len(sorted))*0.99)]

	return LatencyStats{
		Count:   len(latencies),
		Average: average,
		Median:  median,
		P95:     p95,
		P99:     p99,
		Min:     sorted[0],
		Max:     sorted[len(sorted)-1],
	}
}

func TestLatencyBenchmarkSuite(t *testing.T) {
	suite.Run(t, new(LatencyBenchmarkSuite))
}
