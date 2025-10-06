#!/bin/bash

# Strategy Engine Signal Generation Test
# Publishes market data and monitors signal generation

SERVICE_URL="${SERVICE_URL:-http://localhost:9092}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Strategy Engine Signal Generation Test"
echo "=========================================="

# Function to get metric value
get_metric() {
    local metric_name=$1
    curl -s "$SERVICE_URL/metrics" | grep "$metric_name" | head -1 | awk '{print $2}'
}

# Get initial signal count
initial_signals=$(get_metric "strategy_engine_strategy_signals_total")
echo "Initial signals generated: $initial_signals"
echo ""

# Publish market data to trigger strategies
echo "Publishing market data to trigger strategies..."
for i in {1..10}; do
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    price=$((50000 + RANDOM % 1000))

    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" PUBLISH "market:btcusdt" "{
      \"symbol\": \"BTCUSDT\",
      \"timestamp\": \"$timestamp\",
      \"sequence\": $i,
      \"last_price\": $price,
      \"bid_price\": $((price - 5)),
      \"ask_price\": $((price + 5)),
      \"bid_size\": 10.0,
      \"ask_size\": 10.0,
      \"volume\": 1000000.0,
      \"type\": \"tick\"
    }" >/dev/null

    echo -n "."
    sleep 0.2
done

echo ""
echo ""

# Wait for processing
echo "Waiting for signal processing..."
sleep 2

# Get final signal count
final_signals=$(get_metric "strategy_engine_strategy_signals_total")
echo "Final signals generated: $final_signals"

# Calculate new signals
new_signals=$((final_signals - initial_signals))
echo ""
if [ "$new_signals" -gt 0 ]; then
    echo -e "${GREEN}✓ Generated $new_signals new signals${NC}"
else
    echo -e "${YELLOW}⚠ No new signals generated${NC}"
fi

# Check signal queue
queue_size=$(get_metric "strategy_engine_signal_queue_size")
echo "Signal queue size: $queue_size"

# Check for dropped signals
dropped=$(get_metric "strategy_engine_signals_dropped_total")
if [ -n "$dropped" ] && [ "$dropped" != "0" ]; then
    echo -e "${YELLOW}⚠ Dropped signals: $dropped${NC}"
fi

echo ""
echo "=========================================="
echo "Signal Generation Test Complete"
echo "=========================================="
