#!/bin/bash

# Strategy Engine Health Check Script

SERVICE_URL="${SERVICE_URL:-http://localhost:9092}"
API_KEY="${STRATEGY_ENGINE_API_KEY:-}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Strategy Engine Health Check"
echo "=========================================="
echo "Service URL: $SERVICE_URL"
echo ""

# Function to make authenticated request
make_request() {
    local endpoint=$1
    local headers=""

    if [ -n "$API_KEY" ]; then
        headers="-H 'X-API-Key: $API_KEY'"
    fi

    curl -s -f $headers "$SERVICE_URL$endpoint"
}

# Test 1: Health endpoint
echo -n "1. Health check... "
if response=$(make_request "/health"); then
    echo -e "${GREEN}PASS${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}FAIL${NC}"
    echo "   Error: Health endpoint not responding"
    exit 1
fi

# Test 2: Readiness endpoint
echo -n "2. Readiness check... "
if response=$(make_request "/ready"); then
    echo -e "${GREEN}PASS${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}FAIL${NC}"
    echo "   Error: Readiness endpoint not responding"
    exit 1
fi

# Test 3: Status endpoint
echo -n "3. Status check... "
if response=$(make_request "/status"); then
    echo -e "${GREEN}PASS${NC}"
    echo "   Response: $response"

    # Parse and validate status
    if echo "$response" | grep -q '"mode"'; then
        mode=$(echo "$response" | grep -o '"mode":"[^"]*"' | cut -d'"' -f4)
        strategies=$(echo "$response" | grep -o '"active_strategies":[0-9]*' | cut -d':' -f2)
        queue_size=$(echo "$response" | grep -o '"signal_queue_size":[0-9]*' | cut -d':' -f2)

        echo "   Mode: $mode"
        echo "   Active Strategies: $strategies"
        echo "   Queue Size: $queue_size"

        if [ "$strategies" -lt 1 ]; then
            echo -e "   ${YELLOW}WARNING: No active strategies${NC}"
        fi
    fi
else
    echo -e "${RED}FAIL${NC}"
    echo "   Error: Status endpoint not responding"
fi

# Test 4: Metrics endpoint
echo -n "4. Metrics check... "
if response=$(curl -s -f "$SERVICE_URL/metrics"); then
    echo -e "${GREEN}PASS${NC}"

    # Check for key metrics
    if echo "$response" | grep -q "strategy_engine_active_strategies"; then
        echo "   ✓ Strategy metrics present"
    fi
    if echo "$response" | grep -q "strategy_engine_market_data_received_total"; then
        echo "   ✓ Market data metrics present"
    fi
    if echo "$response" | grep -q "strategy_engine_signal_queue_size"; then
        echo "   ✓ Signal queue metrics present"
    fi
else
    echo -e "${YELLOW}WARN${NC}"
    echo "   Error: Metrics endpoint not responding"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}Health Check Complete${NC}"
echo "=========================================="
