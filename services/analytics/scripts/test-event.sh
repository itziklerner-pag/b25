#!/bin/bash
# Test script to send sample events to the analytics service

set -e

API_URL="${API_URL:-http://localhost:9097}"

echo "Sending test events to analytics service..."

# Test 1: Order placed event
echo "1. Sending order.placed event..."
curl -X POST "$API_URL/api/v1/events" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "order.placed",
    "user_id": "test-user-123",
    "session_id": "session-456",
    "properties": {
      "symbol": "BTCUSDT",
      "side": "BUY",
      "price": 50000,
      "quantity": 0.1,
      "order_type": "LIMIT"
    }
  }'
echo ""

# Test 2: Order filled event
echo "2. Sending order.filled event..."
curl -X POST "$API_URL/api/v1/events" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "order.filled",
    "user_id": "test-user-123",
    "session_id": "session-456",
    "properties": {
      "symbol": "BTCUSDT",
      "side": "BUY",
      "price": 50000,
      "quantity": 0.1,
      "commission": 5.0,
      "is_maker": true
    }
  }'
echo ""

# Test 3: Strategy started event
echo "3. Sending strategy.started event..."
curl -X POST "$API_URL/api/v1/events" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "strategy.started",
    "user_id": "test-user-123",
    "properties": {
      "strategy_id": "momentum-v1",
      "strategy_name": "Momentum Strategy",
      "symbols": ["BTCUSDT", "ETHUSDT"]
    }
  }'
echo ""

# Test 4: Query events
echo "4. Querying recent events..."
curl "$API_URL/api/v1/events?limit=10"
echo ""

# Test 5: Get dashboard metrics
echo "5. Getting dashboard metrics..."
curl "$API_URL/api/v1/dashboard/metrics"
echo ""

# Test 6: Health check
echo "6. Checking service health..."
curl "$API_URL/health"
echo ""

echo "Test completed!"
