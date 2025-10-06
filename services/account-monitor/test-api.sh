#!/bin/bash

# Account Monitor - API Test Script

BASE_URL="${BASE_URL:-http://localhost:8080}"
NATS_URL="${NATS_URL:-nats://localhost:4222}"

echo "========================================"
echo "Account Monitor API Tests"
echo "========================================"
echo ""
echo "Testing API at: $BASE_URL"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run tests
run_test() {
    local test_name="$1"
    local url="$2"
    local expected_code="${3:-200}"

    echo -n "Testing ${test_name}... "

    response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))

        # Pretty print JSON if jq is available
        if command -v jq &> /dev/null; then
            echo "$body" | jq '.' 2>/dev/null | head -20
        else
            echo "$body" | head -c 200
        fi
        echo ""
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected HTTP $expected_code, got $http_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        if [ -n "$body" ]; then
            echo "  Response: $body" | head -c 200
            echo ""
        fi
        return 1
    fi
}

# Test 1: Get all positions
echo "Test 1: GET /api/positions"
run_test "Positions" "${BASE_URL}/api/positions"

# Test 2: Get balances
echo "Test 2: GET /api/balance"
run_test "Balances" "${BASE_URL}/api/balance"

# Test 3: Get P&L
echo "Test 3: GET /api/pnl"
run_test "P&L Report" "${BASE_URL}/api/pnl"

# Test 4: Get account state
echo "Test 4: GET /api/account"
run_test "Account State" "${BASE_URL}/api/account"

# Test 5: Get alerts
echo "Test 5: GET /api/alerts"
run_test "Alerts" "${BASE_URL}/api/alerts"

# Summary
echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
