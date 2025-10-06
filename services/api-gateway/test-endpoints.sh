#!/bin/bash

# API Gateway Endpoint Testing Script
# This script tests all major endpoints of the API Gateway

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GATEWAY_URL="${GATEWAY_URL:-http://localhost:8000}"
API_KEY="${API_KEY:-test-admin-key-123}"

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test function
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local expected_code=$4
    local headers=$5
    local data=$6

    log_info "Testing: $name"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            ${headers} \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${GATEWAY_URL}${endpoint}")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            ${headers} \
            "${GATEWAY_URL}${endpoint}")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "$expected_code" ]; then
        log_success "$name (HTTP $http_code)"
        if [ -n "$body" ] && command -v jq &> /dev/null; then
            echo "$body" | jq -C '.' 2>/dev/null || echo "$body"
        else
            echo "$body"
        fi
    else
        log_fail "$name (Expected: $expected_code, Got: $http_code)"
        echo "Response: $body"
    fi
    echo ""
}

echo ""
echo "========================================="
echo "API Gateway Endpoint Testing"
echo "========================================="
echo "Gateway URL: $GATEWAY_URL"
echo "API Key: ${API_KEY:0:10}..."
echo ""

# Test 1: Health Endpoints
echo "========================================="
echo "1. Health Endpoints"
echo "========================================="

test_endpoint \
    "Basic Health Check" \
    "GET" \
    "/health" \
    200

test_endpoint \
    "Liveness Probe" \
    "GET" \
    "/health/liveness" \
    200

test_endpoint \
    "Readiness Probe" \
    "GET" \
    "/health/readiness" \
    200

# Test 2: Version and Metrics
echo "========================================="
echo "2. Version and Metrics"
echo "========================================="

test_endpoint \
    "Version Information" \
    "GET" \
    "/version" \
    200

log_info "Testing: Metrics Endpoint"
metrics=$(curl -s "${GATEWAY_URL}/metrics")
if echo "$metrics" | grep -q "api_gateway"; then
    log_success "Metrics Endpoint (contains api_gateway metrics)"
    echo "$metrics" | grep "api_gateway" | head -5
else
    log_fail "Metrics Endpoint (no api_gateway metrics found)"
fi
echo ""

# Test 3: Authentication
echo "========================================="
echo "3. Authentication Tests"
echo "========================================="

test_endpoint \
    "Protected Endpoint Without Auth" \
    "GET" \
    "/api/v1/account/balance" \
    401

test_endpoint \
    "Protected Endpoint With Invalid API Key" \
    "GET" \
    "/api/v1/account/balance" \
    401 \
    "-H 'X-API-Key: invalid-key'"

test_endpoint \
    "Protected Endpoint With Valid API Key" \
    "GET" \
    "/api/v1/account/balance" \
    503 \
    "-H 'X-API-Key: ${API_KEY}'"

# Test 4: CORS Headers
echo "========================================="
echo "4. CORS Headers"
echo "========================================="

log_info "Testing: CORS Preflight Request"
cors_response=$(curl -s -i -X OPTIONS \
    -H "Origin: http://localhost:3000" \
    -H "Access-Control-Request-Method: POST" \
    "${GATEWAY_URL}/api/v1/orders")

if echo "$cors_response" | grep -q "Access-Control-Allow-Origin"; then
    log_success "CORS Headers Present"
    echo "$cors_response" | grep "Access-Control"
else
    log_fail "CORS Headers Missing"
fi
echo ""

# Test 5: Security Headers
echo "========================================="
echo "5. Security Headers"
echo "========================================="

log_info "Testing: Security Headers"
security_headers=$(curl -s -i "${GATEWAY_URL}/health" | grep -E "X-Content-Type-Options|X-Frame-Options|X-XSS-Protection|Content-Security-Policy")

if [ -n "$security_headers" ]; then
    log_success "Security Headers Present"
    echo "$security_headers"
else
    log_fail "Security Headers Missing"
fi
echo ""

# Test 6: Rate Limiting
echo "========================================="
echo "6. Rate Limiting"
echo "========================================="

log_info "Testing: Rate Limit Headers"
rate_limit_response=$(curl -s -i "${GATEWAY_URL}/health" | grep -E "X-Ratelimit")

if [ -n "$rate_limit_response" ]; then
    log_success "Rate Limit Headers Present"
    echo "$rate_limit_response"
else
    log_warn "Rate Limit Headers Not Found (might be disabled)"
fi
echo ""

# Test 7: Request ID
echo "========================================="
echo "7. Request ID Tracking"
echo "========================================="

log_info "Testing: Request ID Header"
request_id=$(curl -s -i "${GATEWAY_URL}/health" | grep -i "X-Request-Id" | awk '{print $2}')

if [ -n "$request_id" ]; then
    log_success "Request ID Generated: ${request_id}"
else
    log_fail "Request ID Not Found"
fi
echo ""

# Test 8: API Endpoints (with valid auth)
echo "========================================="
echo "8. API Endpoints (Backend Proxy)"
echo "========================================="

test_endpoint \
    "Market Data - Symbols" \
    "GET" \
    "/api/v1/market-data/symbols" \
    503 \
    "-H 'X-API-Key: ${API_KEY}'"

test_endpoint \
    "Account - Balance" \
    "GET" \
    "/api/v1/account/balance" \
    503 \
    "-H 'X-API-Key: ${API_KEY}'"

test_endpoint \
    "Account - Positions" \
    "GET" \
    "/api/v1/account/positions" \
    503 \
    "-H 'X-API-Key: ${API_KEY}'"

# Test 9: Invalid Endpoints
echo "========================================="
echo "9. Invalid Endpoint Handling"
echo "========================================="

test_endpoint \
    "Non-existent Endpoint" \
    "GET" \
    "/api/v1/nonexistent" \
    404

# Summary
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    log_success "All tests passed!"
    exit 0
else
    log_fail "Some tests failed"
    exit 1
fi
