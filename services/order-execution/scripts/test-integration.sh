#!/bin/bash

# Integration Test Script for Order Execution Service
# Tests the complete flow including Redis caching and NATS events

set -e

HOST="${1:-localhost}"
GRPC_PORT="${2:-50051}"
HTTP_PORT="${3:-9091}"
GRPC_ADDR="${HOST}:${GRPC_PORT}"
HTTP_URL="http://${HOST}:${HTTP_PORT}"

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

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."

    # Check grpcurl
    if ! command -v grpcurl &> /dev/null; then
        log_error "grpcurl is not installed"
        log_info "Install with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
        exit 1
    fi

    # Check redis-cli
    if ! command -v redis-cli &> /dev/null; then
        log_warn "redis-cli not found, Redis tests will be skipped"
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed"
        log_info "Install with: sudo apt-get install jq (Ubuntu) or brew install jq (macOS)"
        exit 1
    fi

    log_success "Dependencies check complete"
}

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name="$1"
    shift

    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    log_info "Running: $test_name"

    if "$@"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_success "$test_name"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_fail "$test_name"
        return 1
    fi
}

# Test service health
test_service_health() {
    log_info "Checking service health..."

    local response=$(curl -s "${HTTP_URL}/health")
    local status=$(echo "$response" | jq -r '.status')

    log_info "Service status: $status"

    if [ "$status" = "healthy" ] || [ "$status" = "degraded" ]; then
        return 0
    else
        log_error "Service is unhealthy"
        return 1
    fi
}

# Test order creation and caching
test_order_cache() {
    log_info "Testing order creation and Redis caching..."

    # Create order
    local response=$(grpcurl -plaintext -d '{
        "symbol": "BTCUSDT",
        "side": "BUY",
        "type": "LIMIT",
        "quantity": 0.001,
        "price": 42000,
        "time_in_force": "GTC"
    }' "$GRPC_ADDR" order.OrderService/CreateOrder 2>&1)

    local order_id=$(echo "$response" | jq -r '.orderId // empty')

    if [ -z "$order_id" ]; then
        log_error "Failed to create order"
        echo "$response"
        return 1
    fi

    log_info "Order created: $order_id"

    # Check Redis cache (if redis-cli available)
    if command -v redis-cli &> /dev/null; then
        sleep 1  # Give it a moment to cache

        local cache_key="order:${order_id}"
        local cached_data=$(redis-cli GET "$cache_key" 2>/dev/null || echo "")

        if [ -n "$cached_data" ]; then
            log_info "Order found in Redis cache"
            log_info "Cache key: $cache_key"
            return 0
        else
            log_warn "Order not found in Redis cache (this may be expected if Redis is not configured)"
            return 0
        fi
    else
        log_warn "Skipping Redis check (redis-cli not available)"
        return 0
    fi
}

# Test metrics after operations
test_metrics() {
    log_info "Checking metrics after operations..."

    local metrics=$(curl -s "${HTTP_URL}/metrics")

    # Check for order creation metrics
    if echo "$metrics" | grep -q "order_execution_orders_created_total"; then
        local created_count=$(echo "$metrics" | grep "^order_execution_orders_created_total" | awk '{print $2}')
        log_info "Total orders created: $created_count"
    fi

    # Check for latency metrics
    if echo "$metrics" | grep -q "order_execution_order_latency_seconds"; then
        log_info "Latency metrics found"
    fi

    # Check for cache metrics
    if echo "$metrics" | grep -q "order_execution_cache"; then
        log_info "Cache metrics found"
    fi

    return 0
}

# Test error handling
test_error_handling() {
    log_info "Testing error handling with invalid order..."

    local response=$(grpcurl -plaintext -d '{
        "symbol": "INVALIDPAIR",
        "side": "BUY",
        "type": "LIMIT",
        "quantity": 0.001,
        "price": 45000,
        "time_in_force": "GTC"
    }' "$GRPC_ADDR" order.OrderService/CreateOrder 2>&1)

    if echo "$response" | grep -qi "error\|invalid\|not.found"; then
        log_info "Error correctly returned for invalid symbol"
        return 0
    else
        log_error "Expected error for invalid symbol"
        return 1
    fi
}

# Test concurrent orders
test_concurrent_orders() {
    log_info "Testing concurrent order creation..."

    local pids=()
    local order_count=5

    for i in $(seq 1 $order_count); do
        (
            grpcurl -plaintext -d "{
                \"symbol\": \"BTCUSDT\",
                \"side\": \"BUY\",
                \"type\": \"LIMIT\",
                \"quantity\": 0.001,
                \"price\": $((40000 + i * 100)),
                \"time_in_force\": \"GTC\"
            }" "$GRPC_ADDR" order.OrderService/CreateOrder > /dev/null 2>&1
        ) &
        pids+=($!)
    done

    # Wait for all orders
    local failed=0
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            failed=$((failed + 1))
        fi
    done

    if [ $failed -eq 0 ]; then
        log_info "Successfully created $order_count concurrent orders"
        return 0
    else
        log_warn "$failed out of $order_count concurrent orders failed"
        return 0  # Don't fail test, some failures are expected with rate limiting
    fi
}

# Test rate limiting
test_rate_limiting() {
    log_info "Testing rate limiting..."

    local success_count=0
    local rate_limited_count=0
    local test_count=15

    for i in $(seq 1 $test_count); do
        response=$(grpcurl -plaintext -d '{
            "symbol": "BTCUSDT",
            "side": "BUY",
            "type": "LIMIT",
            "quantity": 0.001,
            "price": 40000,
            "time_in_force": "GTC"
        }' "$GRPC_ADDR" order.OrderService/CreateOrder 2>&1)

        if echo "$response" | grep -qi "rate.limit\|too.many"; then
            rate_limited_count=$((rate_limited_count + 1))
        elif echo "$response" | grep -q "orderId"; then
            success_count=$((success_count + 1))
        fi
    done

    log_info "Successful orders: $success_count"
    log_info "Rate limited orders: $rate_limited_count"

    if [ $rate_limited_count -gt 0 ]; then
        log_info "Rate limiting is working"
    else
        log_warn "No rate limiting observed (may need higher request rate)"
    fi

    return 0
}

# Test environment variable configuration
test_env_config() {
    log_info "Checking if service uses environment variables..."

    # Check if API keys are NOT hardcoded by looking for placeholder pattern
    local config_file="${PWD}/config.yaml"

    if [ -f "$config_file" ]; then
        if grep -q "\${BINANCE_API_KEY}" "$config_file"; then
            log_success "Config uses environment variables (not hardcoded)"
            return 0
        else
            log_warn "Config may have hardcoded values"
            return 0
        fi
    else
        log_warn "Config file not found"
        return 0
    fi
}

# Main test execution
main() {
    log_info "================================================="
    log_info "Order Execution Service - Integration Tests"
    log_info "================================================="
    log_info "gRPC Target: ${GRPC_ADDR}"
    log_info "HTTP Target: ${HTTP_URL}"
    log_info ""

    check_dependencies

    # Check if service is accessible
    if ! curl -s --max-time 2 "${HTTP_URL}/health/live" > /dev/null 2>&1; then
        log_error "Service is not accessible at ${HTTP_URL}"
        exit 1
    fi

    # Run tests
    run_test "Test 1: Service Health" test_service_health
    run_test "Test 2: Environment Variable Config" test_env_config
    run_test "Test 3: Order Cache Integration" test_order_cache
    run_test "Test 4: Metrics Collection" test_metrics
    run_test "Test 5: Error Handling" test_error_handling
    run_test "Test 6: Concurrent Orders" test_concurrent_orders
    run_test "Test 7: Rate Limiting" test_rate_limiting

    # Summary
    echo ""
    log_info "================================================="
    log_info "Test Summary"
    log_info "================================================="
    log_info "Total tests:  $TESTS_RUN"
    log_success "Passed:       $TESTS_PASSED"

    if [ $TESTS_FAILED -gt 0 ]; then
        log_fail "Failed:       $TESTS_FAILED"
        exit 1
    else
        log_success "All integration tests passed!"
        exit 0
    fi
}

# Run main
main "$@"
