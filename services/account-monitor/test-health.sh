#!/bin/bash

# Account Monitor - Health Endpoint Test Script

BASE_URL="${BASE_URL:-http://localhost:8080}"
METRICS_URL="${METRICS_URL:-http://localhost:9093}"

echo "========================================"
echo "Account Monitor Health Tests"
echo "========================================"
echo ""
echo "Testing endpoints at: $BASE_URL"
echo "Testing metrics at: $METRICS_URL"
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

        # Show response body for some tests
        if [[ "$test_name" == *"Health"* ]] || [[ "$test_name" == *"Ready"* ]]; then
            echo "  Response: $body" | head -c 100
            echo ""
        fi
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

# Test 1: Health endpoint
echo "Test 1: Health Check"
run_test "GET /health" "${BASE_URL}/health"
echo ""

# Test 2: Ready endpoint
echo "Test 2: Readiness Check"
run_test "GET /ready" "${BASE_URL}/ready"
echo ""

# Test 3: Metrics endpoint
echo "Test 3: Prometheus Metrics"
run_test "GET /metrics" "${METRICS_URL}/metrics"
echo ""

# Test 4: Verify metrics contain expected values
echo "Test 4: Metrics Content Validation"
echo -n "Checking for account_monitor metrics... "
metrics=$(curl -s "${METRICS_URL}/metrics" 2>/dev/null)
if echo "$metrics" | grep -q "account_"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))

    # Count metrics
    metric_count=$(echo "$metrics" | grep -c "^account_" || true)
    echo "  Found $metric_count account metrics"
else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 5: CORS headers on health endpoint
echo "Test 5: CORS Headers"
echo -n "Checking CORS headers... "
cors_header=$(curl -s -I "${BASE_URL}/health" 2>/dev/null | grep -i "access-control-allow-origin" || true)
if [ -n "$cors_header" ]; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "  Header: $cors_header"
else
    echo -e "${YELLOW}⚠ WARN${NC} (CORS headers not found)"
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

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
