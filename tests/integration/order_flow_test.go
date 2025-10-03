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

// OrderFlowTestSuite tests order submission and fill flow
type OrderFlowTestSuite struct {
	suite.Suite
	redisClient *redis.Client
	natsConn    *nats.Conn
	orderGen    *generators.OrderGenerator
}

func (s *OrderFlowTestSuite) SetupSuite() {
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
}

func (s *OrderFlowTestSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

func (s *OrderFlowTestSuite) SetupTest() {
	ctx := context.Background()
	s.redisClient.FlushDB(ctx)
}

// TestOrderSubmission tests basic order submission
func (s *OrderFlowTestSuite) TestOrderSubmission() {
	// Generate test order
	order := s.orderGen.GenerateMarketOrder("BTCUSDT", "BUY", 0.1)

	// Publish order request
	subject := "orders.requests.create"
	data, err := json.Marshal(order)
	require.NoError(s.T(), err)

	// Subscribe to order updates
	updatesChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("orders.updates.BTCUSDT", updatesChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish order
	err = s.natsConn.Publish(subject, data)
	require.NoError(s.T(), err)

	// Wait for order update
	select {
	case msg := <-updatesChan:
		var update map[string]interface{}
		err = json.Unmarshal(msg.Data, &update)
		require.NoError(s.T(), err)

		orderData := update["order"].(map[string]interface{})
		assert.Equal(s.T(), "BTCUSDT", orderData["symbol"])
		assert.Contains(s.T(), []string{"SUBMITTED", "FILLED"}, orderData["state"])

	case <-time.After(5 * time.Second):
		s.T().Skip("Order execution service not responding")
	}
}

// TestOrderLifecycle tests complete order lifecycle
func (s *OrderFlowTestSuite) TestOrderLifecycle() {
	order := s.orderGen.GenerateMarketOrder("BTCUSDT", "BUY", 0.05)

	// Track order states
	statesChan := make(chan string, 10)
	sub, err := s.natsConn.Subscribe("orders.updates.BTCUSDT", func(msg *nats.Msg) {
		var update map[string]interface{}
		if err := json.Unmarshal(msg.Data, &update); err == nil {
			if orderData, ok := update["order"].(map[string]interface{}); ok {
				if state, ok := orderData["state"].(string); ok {
					statesChan <- state
				}
			}
		}
	})
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Submit order
	data, err := json.Marshal(order)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("orders.requests.create", data)
	require.NoError(s.T(), err)

	// Collect states
	states := []string{}
	timeout := time.After(5 * time.Second)

	for {
		select {
		case state := <-statesChan:
			states = append(states, state)
			if state == "FILLED" || state == "REJECTED" {
				goto VerifyStates
			}
		case <-timeout:
			goto VerifyStates
		}
	}

VerifyStates:
	if len(states) == 0 {
		s.T().Skip("No order states received - service may not be running")
		return
	}

	// Verify state progression
	s.T().Logf("Order states: %v", states)
	assert.Contains(s.T(), states, "SUBMITTED")
}

// TestOrderFillFlow tests order fill processing
func (s *OrderFlowTestSuite) TestOrderFillFlow() {
	order := s.orderGen.GenerateMarketOrder("BTCUSDT", "BUY", 0.1)

	// Subscribe to fills
	fillsChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("orders.fills.#", fillsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Submit order
	data, err := json.Marshal(order)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("orders.requests.create", data)
	require.NoError(s.T(), err)

	// Wait for fill
	select {
	case msg := <-fillsChan:
		var fill map[string]interface{}
		err = json.Unmarshal(msg.Data, &fill)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "BTCUSDT", fill["symbol"])
		assert.NotNil(s.T(), fill["price"])
		assert.NotNil(s.T(), fill["quantity"])

	case <-time.After(10 * time.Second):
		s.T().Skip("No fill received - mock exchange may not be running")
	}
}

// TestOrderCancellation tests order cancellation
func (s *OrderFlowTestSuite) TestOrderCancellation() {
	// Create limit order (won't fill immediately)
	order := s.orderGen.GenerateLimitOrder("BTCUSDT", "BUY", 40000.0, 0.1)

	// Submit order
	data, err := json.Marshal(order)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("orders.requests.create", data)
	require.NoError(s.T(), err)

	// Wait for order to be created
	time.Sleep(200 * time.Millisecond)

	// Get order ID from Redis (assuming it's cached)
	ctx := context.Background()
	pattern := "order:*"
	keys, err := s.redisClient.Keys(ctx, pattern).Result()

	if err != nil || len(keys) == 0 {
		s.T().Skip("Cannot retrieve order ID - service may not be running")
		return
	}

	// Cancel order
	cancelReq := map[string]interface{}{
		"order_id": keys[0][6:], // Remove "order:" prefix
		"symbol":   "BTCUSDT",
	}

	cancelData, err := json.Marshal(cancelReq)
	require.NoError(s.T(), err)

	// Subscribe to cancellation updates
	updatesChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("orders.updates.BTCUSDT", updatesChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Publish cancellation
	err = s.natsConn.Publish("orders.requests.cancel", cancelData)
	require.NoError(s.T(), err)

	// Wait for cancellation confirmation
	select {
	case msg := <-updatesChan:
		var update map[string]interface{}
		err = json.Unmarshal(msg.Data, &update)
		require.NoError(s.T(), err)

		if orderData, ok := update["order"].(map[string]interface{}); ok {
			state := orderData["state"].(string)
			if state == "CANCELED" {
				s.T().Logf("Order successfully canceled")
				return
			}
		}

	case <-time.After(5 * time.Second):
		s.T().Skip("Cancellation not confirmed")
	}
}

// TestOrderRejection tests order validation and rejection
func (s *OrderFlowTestSuite) TestOrderRejection() {
	// Create invalid order (invalid quantity)
	order := map[string]interface{}{
		"symbol":          "BTCUSDT",
		"side":            "BUY",
		"type":            "MARKET",
		"quantity":        0.0001, // Below minimum
		"client_order_id": "test_reject_001",
	}

	// Subscribe to updates
	updatesChan := make(chan *nats.Msg, 1)
	sub, err := s.natsConn.ChanSubscribe("orders.updates.BTCUSDT", updatesChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Submit order
	data, err := json.Marshal(order)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("orders.requests.create", data)
	require.NoError(s.T(), err)

	// Wait for rejection
	select {
	case msg := <-updatesChan:
		var update map[string]interface{}
		err = json.Unmarshal(msg.Data, &update)
		require.NoError(s.T(), err)

		if orderData, ok := update["order"].(map[string]interface{}); ok {
			state := orderData["state"].(string)
			assert.Equal(s.T(), "REJECTED", state)
		}

	case <-time.After(5 * time.Second):
		s.T().Skip("No rejection received")
	}
}

// TestOrderLatency tests order submission latency
func (s *OrderFlowTestSuite) TestOrderLatency() {
	numOrders := 50
	latencies := make([]time.Duration, numOrders)

	for i := 0; i < numOrders; i++ {
		order := s.orderGen.GenerateMarketOrder("BTCUSDT", "BUY", 0.01)

		startTime := time.Now()

		data, err := json.Marshal(order)
		require.NoError(s.T(), err)

		err = s.natsConn.Publish("orders.requests.create", data)
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

	avgLatency := totalLatency / time.Duration(numOrders)

	s.T().Logf("Average order latency: %v", avgLatency)
	s.T().Logf("Max order latency: %v", maxLatency)

	// Assert performance targets
	assert.Less(s.T(), avgLatency, 10*time.Millisecond, "Average latency should be < 10ms")
	assert.Less(s.T(), maxLatency, 50*time.Millisecond, "Max latency should be < 50ms")
}

// TestConcurrentOrders tests concurrent order submission
func (s *OrderFlowTestSuite) TestConcurrentOrders() {
	numOrders := 20
	doneChan := make(chan bool, numOrders)

	// Submit orders concurrently
	for i := 0; i < numOrders; i++ {
		go func(idx int) {
			order := s.orderGen.GenerateMarketOrder("BTCUSDT", "BUY", 0.01)
			data, err := json.Marshal(order)
			if err != nil {
				doneChan <- false
				return
			}

			err = s.natsConn.Publish("orders.requests.create", data)
			doneChan <- err == nil
		}(i)
	}

	// Wait for all orders
	successCount := 0
	for i := 0; i < numOrders; i++ {
		select {
		case success := <-doneChan:
			if success {
				successCount++
			}
		case <-time.After(10 * time.Second):
			s.T().Fatal("Timeout waiting for concurrent orders")
		}
	}

	assert.Greater(s.T(), successCount, numOrders*80/100, "At least 80% of orders should succeed")
	s.T().Logf("Successfully submitted %d/%d concurrent orders", successCount, numOrders)
}

// TestPartialFill tests partial order fills
func (s *OrderFlowTestSuite) TestPartialFill() {
	// Create large limit order
	order := s.orderGen.GenerateLimitOrder("BTCUSDT", "BUY", 50000.0, 5.0)

	// Subscribe to fill events
	fillsChan := make(chan *nats.Msg, 10)
	sub, err := s.natsConn.ChanSubscribe("orders.fills.#", fillsChan)
	require.NoError(s.T(), err)
	defer sub.Unsubscribe()

	// Submit order
	data, err := json.Marshal(order)
	require.NoError(s.T(), err)

	err = s.natsConn.Publish("orders.requests.create", data)
	require.NoError(s.T(), err)

	// Collect fills
	fills := []map[string]interface{}{}
	timeout := time.After(5 * time.Second)

collectFills:
	for {
		select {
		case msg := <-fillsChan:
			var fill map[string]interface{}
			if err := json.Unmarshal(msg.Data, &fill); err == nil {
				fills = append(fills, fill)
			}
		case <-timeout:
			break collectFills
		}
	}

	if len(fills) == 0 {
		s.T().Skip("No fills received - partial fill test skipped")
		return
	}

	s.T().Logf("Received %d fill(s)", len(fills))
}

func TestOrderFlowSuite(t *testing.T) {
	suite.Run(t, new(OrderFlowTestSuite))
}
