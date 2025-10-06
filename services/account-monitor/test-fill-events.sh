#!/bin/bash

# Account Monitor - Fill Event Processing Test Script
# Tests position tracking and P&L calculation

BASE_URL="${BASE_URL:-http://localhost:8080}"
NATS_URL="${NATS_URL:-localhost:4222}"

echo "========================================"
echo "Account Monitor Fill Event Tests"
echo "========================================"
echo ""
echo "Testing API at: $BASE_URL"
echo "Testing NATS at: $NATS_URL"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if nats CLI is installed
if ! command -v nats &> /dev/null; then
    echo -e "${RED}ERROR: nats CLI not found${NC}"
    echo "Install with: go install github.com/nats-io/natscli/nats@latest"
    exit 1
fi

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function for API calls
get_api() {
    local endpoint="$1"
    curl -s "${BASE_URL}${endpoint}" 2>/dev/null
}

# Helper function to publish fill events
publish_fill() {
    local fill_json="$1"
    echo "$fill_json" | nats pub trading.fills --server="$NATS_URL" 2>/dev/null
    sleep 1  # Give service time to process
}

echo "Test 1: Opening a LONG position"
echo "================================"
echo "Publishing BUY fill for BTCUSDT..."
FILL_BUY=$(cat <<EOF
{
  "id": "fill-test-001",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": "0.001",
  "price": "50000.00",
  "fee": "0.05",
  "fee_currency": "USDT",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)
publish_fill "$FILL_BUY"

# Check position was created
echo -n "Checking if position was created... "
positions=$(get_api "/api/positions")
if echo "$positions" | grep -q "BTCUSDT"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))

    # Verify position quantity
    if command -v jq &> /dev/null; then
        qty=$(echo "$positions" | jq -r '.BTCUSDT.quantity // empty' 2>/dev/null)
        if [ "$qty" == "0.001" ]; then
            echo "  Position quantity: $qty ✓"
        else
            echo -e "  ${YELLOW}Position quantity mismatch: expected 0.001, got $qty${NC}"
        fi
    fi
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

echo "Test 2: Adding to the position"
echo "==============================="
echo "Publishing another BUY fill for BTCUSDT..."
FILL_ADD=$(cat <<EOF
{
  "id": "fill-test-002",
  "symbol": "BTCUSDT",
  "side": "BUY",
  "quantity": "0.001",
  "price": "51000.00",
  "fee": "0.051",
  "fee_currency": "USDT",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)
publish_fill "$FILL_ADD"

# Check position was updated
echo -n "Checking if position was updated... "
positions=$(get_api "/api/positions")
if command -v jq &> /dev/null; then
    qty=$(echo "$positions" | jq -r '.BTCUSDT.quantity // empty' 2>/dev/null)
    if [ "$qty" == "0.002" ]; then
        echo -e "${GREEN}✓ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "  New quantity: $qty"
    else
        echo -e "${RED}✗ FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "  Expected 0.002, got $qty"
    fi
else
    echo -e "${YELLOW}⚠ SKIP${NC} (jq not installed)"
fi
echo ""

echo "Test 3: Closing part of the position (realize P&L)"
echo "==================================================="
echo "Publishing SELL fill for BTCUSDT..."
FILL_CLOSE=$(cat <<EOF
{
  "id": "fill-test-003",
  "symbol": "BTCUSDT",
  "side": "SELL",
  "quantity": "0.001",
  "price": "52000.00",
  "fee": "0.052",
  "fee_currency": "USDT",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)
publish_fill "$FILL_CLOSE"

# Check P&L was realized
echo -n "Checking realized P&L... "
pnl=$(get_api "/api/pnl")
if command -v jq &> /dev/null; then
    realized_pnl=$(echo "$pnl" | jq -r '.realized_pnl // empty' 2>/dev/null)
    if [ -n "$realized_pnl" ] && [ "$realized_pnl" != "0" ]; then
        echo -e "${GREEN}✓ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "  Realized P&L: $realized_pnl"

        # Show P&L details
        echo "$pnl" | jq '{realized_pnl, unrealized_pnl, total_fees, net_pnl, total_trades}' 2>/dev/null
    else
        echo -e "${RED}✗ FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "  No realized P&L found"
    fi
else
    echo -e "${YELLOW}⚠ SKIP${NC} (jq not installed)"
fi
echo ""

echo "Test 4: Closing remaining position"
echo "==================================="
echo "Publishing SELL fill to close position..."
FILL_FINAL=$(cat <<EOF
{
  "id": "fill-test-004",
  "symbol": "BTCUSDT",
  "side": "SELL",
  "quantity": "0.001",
  "price": "52500.00",
  "fee": "0.0525",
  "fee_currency": "USDT",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)
publish_fill "$FILL_FINAL"

# Check position is closed
echo -n "Checking if position is closed... "
positions=$(get_api "/api/positions")
if command -v jq &> /dev/null; then
    qty=$(echo "$positions" | jq -r '.BTCUSDT.quantity // "0"' 2>/dev/null)
    if [ "$qty" == "0" ] || [ "$qty" == "0.000" ]; then
        echo -e "${GREEN}✓ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "  Position closed"
    else
        echo -e "${RED}✗ FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "  Position still open with qty: $qty"
    fi
else
    echo -e "${YELLOW}⚠ SKIP${NC} (jq not installed)"
fi
echo ""

# Test 5: Test SHORT position
echo "Test 5: Opening a SHORT position"
echo "================================="
echo "Publishing SELL fill for ETHUSDT..."
FILL_SHORT=$(cat <<EOF
{
  "id": "fill-test-005",
  "symbol": "ETHUSDT",
  "side": "SELL",
  "quantity": "0.01",
  "price": "3000.00",
  "fee": "0.03",
  "fee_currency": "USDT",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)
publish_fill "$FILL_SHORT"

# Check position was created
echo -n "Checking if SHORT position was created... "
positions=$(get_api "/api/positions")
if echo "$positions" | grep -q "ETHUSDT"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))

    if command -v jq &> /dev/null; then
        qty=$(echo "$positions" | jq -r '.ETHUSDT.quantity // empty' 2>/dev/null)
        echo "  Position quantity: $qty (negative = SHORT)"
    fi
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Summary
echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

echo "Final State:"
echo "------------"
echo "Positions:"
get_api "/api/positions" | (command -v jq &> /dev/null && jq '.' || cat)
echo ""
echo "P&L:"
get_api "/api/pnl" | (command -v jq &> /dev/null && jq '.' || cat)
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
