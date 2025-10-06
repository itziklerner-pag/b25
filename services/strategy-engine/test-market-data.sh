#!/bin/bash

# Strategy Engine Market Data Test Script
# Publishes test market data to Redis

REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Strategy Engine Market Data Test"
echo "=========================================="
echo "Redis: $REDIS_HOST:$REDIS_PORT"
echo ""

# Check if redis-cli is available
if ! command -v redis-cli &> /dev/null; then
    echo "Error: redis-cli not found. Please install redis-tools."
    exit 1
fi

# Test Redis connection
echo -n "Testing Redis connection... "
if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" PING &>/dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo "FAILED"
    echo "Error: Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
    exit 1
fi

# Function to publish market data
publish_market_data() {
    local symbol=$1
    local price=$2
    local channel="market:$(echo $symbol | tr '[:upper:]' '[:lower:]')"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local bid_price=$(echo "$price - 5" | bc)
    local ask_price=$(echo "$price + 5" | bc)

    local data=$(cat <<EOF
{
  "symbol": "$symbol",
  "timestamp": "$timestamp",
  "sequence": $RANDOM,
  "last_price": $price,
  "bid_price": $bid_price,
  "ask_price": $ask_price,
  "bid_size": 10.0,
  "ask_size": 10.0,
  "volume": 1000000.0,
  "volume_quote": 50000000000.0,
  "type": "tick"
}
EOF
)

    echo -n "Publishing to $channel... "
    if result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" PUBLISH "$channel" "$data"); then
        echo -e "${GREEN}OK${NC} (subscribers: $result)"
    else
        echo "FAILED"
    fi
}

# Publish test data for different symbols
echo ""
echo "Publishing test market data..."
echo ""

publish_market_data "BTCUSDT" 50000.0
sleep 0.5

publish_market_data "ETHUSDT" 3000.0
sleep 0.5

publish_market_data "SOLUSDT" 100.0
sleep 0.5

echo ""
echo "=========================================="
echo -e "${GREEN}Market Data Test Complete${NC}"
echo "=========================================="
echo ""
echo "Check strategy engine logs for processing confirmation:"
echo "  journalctl -u strategy-engine -f"
echo "  or"
echo "  tail -f /var/log/strategy-engine/strategy-engine.log"
echo ""
