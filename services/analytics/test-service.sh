#!/bin/bash

# Analytics Service Test Script
# Description: Comprehensive testing of analytics service endpoints

set -e

# Configuration
API_URL="${API_URL:-http://localhost:9097}"
METRICS_URL="${METRICS_URL:-http://localhost:9098}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
print_header() {
    echo -e "\n${BLUE}======================================"
    echo "$1"
    echo -e "======================================${NC}"
}

test_endpoint() {
    TESTS_RUN=$((TESTS_RUN + 1))
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local expected_code="$5"

    echo -e "${YELLOW}Testing: $name${NC}"

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "Content-Type: application/json" -d "$data" "$url")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} - HTTP $http_code"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} - Expected HTTP $expected_code, got $http_code"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "$body"
        return 1
    fi
}

# Start tests
print_header "Analytics Service Test Suite"

# Test 1: Health Check
print_header "1. Health & Status Tests"
test_endpoint "Health Check" "GET" "$API_URL/health" "" "200"
test_endpoint "Healthz" "GET" "$API_URL/healthz" "" "200"
test_endpoint "Ready Check" "GET" "$API_URL/ready" "" "200"

# Test 2: Prometheus Metrics
print_header "2. Prometheus Metrics"
echo -e "${YELLOW}Testing: Prometheus Metrics Endpoint${NC}"
response=$(curl -s "$METRICS_URL/metrics")
if echo "$response" | grep -q "analytics_events_ingested_total"; then
    echo -e "${GREEN}✓ PASS${NC} - Metrics endpoint working"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "Sample metrics:"
    echo "$response" | grep "analytics_" | head -5
else
    echo -e "${RED}✗ FAIL${NC} - Metrics not found"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Test 3: Event Tracking
print_header "3. Event Tracking Tests"
EVENT_DATA='{
  "event_type": "order.placed",
  "user_id": "test-user-001",
  "session_id": "test-session-001",
  "properties": {
    "symbol": "BTCUSDT",
    "side": "BUY",
    "price": 50000,
    "quantity": 0.1
  }
}'
test_endpoint "Track Order Event" "POST" "$API_URL/api/v1/events" "$EVENT_DATA" "201"

# Test 4: Event Query
print_header "4. Event Query Tests"
test_endpoint "Get Recent Events" "GET" "$API_URL/api/v1/events?limit=10" "" "200"
test_endpoint "Get Events by Type" "GET" "$API_URL/api/v1/events?event_type=order.placed&limit=5" "" "200"

# Test 5: Event Statistics
print_header "5. Event Statistics"
test_endpoint "Event Stats" "GET" "$API_URL/api/v1/events/stats" "" "200"

# Test 6: Dashboard Metrics
print_header "6. Dashboard Metrics"
test_endpoint "Dashboard Metrics" "GET" "$API_URL/api/v1/dashboard/metrics" "" "200"

# Test 7: Ingestion Metrics
print_header "7. Ingestion Metrics"
test_endpoint "Ingestion Metrics" "GET" "$API_URL/api/v1/internal/ingestion-metrics" "" "200"

# Test 8: Custom Event Definition
print_header "8. Custom Event Tests"
CUSTOM_EVENT='{
  "name": "test.custom.event",
  "display_name": "Test Custom Event",
  "description": "A test custom event",
  "schema": {
    "type": "object",
    "properties": {
      "test_value": {"type": "string"}
    }
  },
  "is_active": true
}'
test_endpoint "Create Custom Event" "POST" "$API_URL/api/v1/custom-events" "$CUSTOM_EVENT" "201"
test_endpoint "Get Custom Event" "GET" "$API_URL/api/v1/custom-events/test.custom.event" "" "200"

# Test 9: Rate Limiting (if enabled)
print_header "9. Rate Limiting Test"
echo -e "${YELLOW}Testing: Rate Limiting (sending 10 rapid requests)${NC}"
RATE_LIMIT_EXCEEDED=0
for i in {1..10}; do
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health")
    if [ "$http_code" = "429" ]; then
        RATE_LIMIT_EXCEEDED=1
        break
    fi
done

if [ "$RATE_LIMIT_EXCEEDED" = "1" ]; then
    echo -e "${GREEN}✓ PASS${NC} - Rate limiting is working"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${YELLOW}⚠ INFO${NC} - Rate limiting not triggered or disabled"
fi
TESTS_RUN=$((TESTS_RUN + 1))

# Test 10: Load Test (Optional)
print_header "10. Light Load Test"
echo -e "${YELLOW}Sending 100 events...${NC}"
START_TIME=$(date +%s)
for i in {1..100}; do
    curl -s -X POST "$API_URL/api/v1/events" \
        -H "Content-Type: application/json" \
        -d "{
            \"event_type\": \"test.load\",
            \"user_id\": \"load-test-$i\",
            \"properties\": {\"index\": $i}
        }" > /dev/null &
done
wait
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo -e "${GREEN}✓ Completed${NC} - Sent 100 events in ${DURATION}s"
echo -e "  Throughput: ~$((100 / DURATION)) events/sec"

# Final Report
print_header "Test Summary"
echo -e "Total Tests: $TESTS_RUN"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed! ✗${NC}"
    exit 1
fi
