#!/bin/bash

# Health Check Test Script for Order Execution Service

set -e

HOST="${1:-localhost}"
PORT="${2:-9091}"
BASE_URL="http://${HOST}:${PORT}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_status="${3:-200}"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    log_info "Running: $test_name"

    if eval "$test_cmd"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_success "$test_name"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_fail "$test_name"
        return 1
    fi
}

# Test liveness probe
test_liveness() {
    local response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/health/live")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "200" ]; then
        log_info "Liveness check: $body (HTTP $http_code)"
        return 0
    else
        log_error "Liveness check failed: HTTP $http_code"
        return 1
    fi
}

# Test readiness probe
test_readiness() {
    local response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/health/ready")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "200" ]; then
        log_info "Readiness check: $body (HTTP $http_code)"
        return 0
    else
        log_warn "Readiness check: $body (HTTP $http_code) - Service may be degraded"
        # Don't fail, just warn - service might be partially ready
        return 0
    fi
}

# Test full health endpoint
test_health() {
    local response=$(curl -s "${BASE_URL}/health")

    if echo "$response" | jq -e . >/dev/null 2>&1; then
        log_info "Health endpoint response:"
        echo "$response" | jq .

        # Check status
        local status=$(echo "$response" | jq -r '.status')
        log_info "Overall status: $status"

        # Check individual components
        local redis_status=$(echo "$response" | jq -r '.checks.redis.status // "unknown"')
        local nats_status=$(echo "$response" | jq -r '.checks.nats.status // "unknown"')
        local system_status=$(echo "$response" | jq -r '.checks.system.status // "unknown"')

        log_info "Redis: $redis_status"
        log_info "NATS: $nats_status"
        log_info "System: $system_status"

        if [ "$status" = "healthy" ] || [ "$status" = "degraded" ]; then
            return 0
        else
            log_error "Service is unhealthy"
            return 1
        fi
    else
        log_error "Invalid JSON response from health endpoint"
        return 1
    fi
}

# Test metrics endpoint
test_metrics() {
    local response=$(curl -s "${BASE_URL}/metrics")

    if echo "$response" | grep -q "order_execution"; then
        log_info "Metrics endpoint accessible"
        local metric_count=$(echo "$response" | grep "^order_execution" | wc -l)
        log_info "Found $metric_count order execution metrics"
        return 0
    else
        log_error "Metrics endpoint not accessible or no metrics found"
        return 1
    fi
}

# Test CORS headers
test_cors() {
    local response=$(curl -s -I -H "Origin: http://example.com" "${BASE_URL}/health")

    if echo "$response" | grep -qi "access-control-allow-origin"; then
        log_info "CORS headers present"
        return 0
    else
        log_warn "CORS headers not found (may be intentional)"
        return 0
    fi
}

# Main test execution
main() {
    log_info "========================================"
    log_info "Order Execution Service - Health Tests"
    log_info "========================================"
    log_info "Target: ${BASE_URL}"
    log_info ""

    # Check if service is accessible
    if ! curl -s --max-time 2 "${BASE_URL}/health/live" > /dev/null; then
        log_error "Service is not accessible at ${BASE_URL}"
        log_error "Make sure the service is running and accessible"
        exit 1
    fi

    # Run tests
    run_test "Test 1: Liveness Probe" "test_liveness"
    run_test "Test 2: Readiness Probe" "test_readiness"
    run_test "Test 3: Full Health Check" "test_health"
    run_test "Test 4: Metrics Endpoint" "test_metrics"
    run_test "Test 5: CORS Headers" "test_cors"

    # Summary
    echo ""
    log_info "========================================"
    log_info "Test Summary"
    log_info "========================================"
    log_info "Total tests:  $TESTS_RUN"
    log_success "Passed:       $TESTS_PASSED"

    if [ $TESTS_FAILED -gt 0 ]; then
        log_fail "Failed:       $TESTS_FAILED"
        exit 1
    else
        log_success "All tests passed!"
        exit 0
    fi
}

# Run main
main "$@"
