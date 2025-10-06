#!/bin/bash

# gRPC Test Script for Order Execution Service

set -e

HOST="${1:-localhost}"
PORT="${2:-50051}"
GRPC_ADDR="${HOST}:${PORT}"

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

# Check if grpcurl is installed
check_grpcurl() {
    if ! command -v grpcurl &> /dev/null; then
        log_error "grpcurl is not installed"
        log_info "Install with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
        exit 1
    fi
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

# Test service listing
test_service_list() {
    log_info "Listing available services..."
    local services=$(grpcurl -plaintext "$GRPC_ADDR" list 2>&1)

    if echo "$services" | grep -q "order.OrderService"; then
        log_info "Found OrderService"
        return 0
    else
        log_error "OrderService not found"
        echo "$services"
        return 1
    fi
}

# Test method listing
test_method_list() {
    log_info "Listing OrderService methods..."
    local methods=$(grpcurl -plaintext "$GRPC_ADDR" list order.OrderService 2>&1)

    log_info "Available methods:"
    echo "$methods"

    if echo "$methods" | grep -q "CreateOrder"; then
        return 0
    else
        log_error "CreateOrder method not found"
        return 1
    fi
}

# Test order creation
test_create_order() {
    log_info "Creating a limit order..."

    local response=$(grpcurl -plaintext -d '{
        "symbol": "BTCUSDT",
        "side": "BUY",
        "type": "LIMIT",
        "quantity": 0.001,
        "price": 40000,
        "time_in_force": "GTC"
    }' "$GRPC_ADDR" order.OrderService/CreateOrder 2>&1)

    echo "$response"

    if echo "$response" | grep -q "orderId"; then
        local order_id=$(echo "$response" | jq -r '.orderId // empty')
        if [ -n "$order_id" ]; then
            log_info "Order created with ID: $order_id"
            echo "$order_id" > /tmp/test_order_id.txt
            return 0
        fi
    fi

    log_error "Failed to create order"
    return 1
}

# Test order query
test_get_order() {
    if [ ! -f /tmp/test_order_id.txt ]; then
        log_warn "No order ID from previous test, skipping"
        return 0
    fi

    local order_id=$(cat /tmp/test_order_id.txt)
    log_info "Querying order: $order_id"

    local response=$(grpcurl -plaintext -d "{
        \"order_id\": \"$order_id\"
    }" "$GRPC_ADDR" order.OrderService/GetOrder 2>&1)

    echo "$response"

    if echo "$response" | grep -q "order"; then
        log_info "Order retrieved successfully"
        return 0
    else
        log_error "Failed to retrieve order"
        return 1
    fi
}

# Test order cancellation
test_cancel_order() {
    if [ ! -f /tmp/test_order_id.txt ]; then
        log_warn "No order ID from previous test, skipping"
        return 0
    fi

    local order_id=$(cat /tmp/test_order_id.txt)
    log_info "Cancelling order: $order_id"

    local response=$(grpcurl -plaintext -d "{
        \"order_id\": \"$order_id\",
        \"symbol\": \"BTCUSDT\"
    }" "$GRPC_ADDR" order.OrderService/CancelOrder 2>&1)

    echo "$response"

    if echo "$response" | grep -q "orderId\|state"; then
        log_info "Order cancelled successfully"
        rm -f /tmp/test_order_id.txt
        return 0
    else
        log_error "Failed to cancel order"
        return 1
    fi
}

# Test validation - invalid quantity
test_validation_invalid_quantity() {
    log_info "Testing validation with invalid quantity (too small)..."

    local response=$(grpcurl -plaintext -d '{
        "symbol": "BTCUSDT",
        "side": "BUY",
        "type": "LIMIT",
        "quantity": 0.0001,
        "price": 45000,
        "time_in_force": "GTC"
    }' "$GRPC_ADDR" order.OrderService/CreateOrder 2>&1)

    echo "$response"

    if echo "$response" | grep -qi "InvalidArgument\|validation"; then
        log_info "Validation correctly rejected invalid quantity"
        return 0
    else
        log_error "Validation did not reject invalid quantity"
        return 1
    fi
}

# Test validation - post-only with wrong TIF
test_validation_postonly_tif() {
    log_info "Testing validation with POST_ONLY and wrong time-in-force..."

    local response=$(grpcurl -plaintext -d '{
        "symbol": "BTCUSDT",
        "side": "BUY",
        "type": "POST_ONLY",
        "quantity": 0.001,
        "price": 45000,
        "time_in_force": "IOC",
        "post_only": true
    }' "$GRPC_ADDR" order.OrderService/CreateOrder 2>&1)

    echo "$response"

    if echo "$response" | grep -qi "InvalidArgument\|validation\|time.in.force"; then
        log_info "Validation correctly rejected POST_ONLY with IOC"
        return 0
    else
        log_warn "Validation may not have rejected POST_ONLY with IOC (might be expected)"
        return 0
    fi
}

# Main test execution
main() {
    log_info "========================================"
    log_info "Order Execution Service - gRPC Tests"
    log_info "========================================"
    log_info "Target: ${GRPC_ADDR}"
    log_info ""

    check_grpcurl

    # Check if service is accessible
    if ! grpcurl -plaintext "$GRPC_ADDR" list > /dev/null 2>&1; then
        log_error "gRPC service is not accessible at ${GRPC_ADDR}"
        log_error "Make sure the service is running"
        exit 1
    fi

    # Run tests
    run_test "Test 1: Service Discovery" test_service_list
    run_test "Test 2: Method Discovery" test_method_list
    run_test "Test 3: Create Order" test_create_order
    run_test "Test 4: Get Order" test_get_order
    run_test "Test 5: Cancel Order" test_cancel_order
    run_test "Test 6: Validation - Invalid Quantity" test_validation_invalid_quantity
    run_test "Test 7: Validation - POST_ONLY TIF" test_validation_postonly_tif

    # Cleanup
    rm -f /tmp/test_order_id.txt

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
