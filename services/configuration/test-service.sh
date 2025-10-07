#!/bin/bash

# Configuration Service - Test Script
# Tests all major endpoints and functionality

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8085"
API_KEY="${CONFIG_API_KEY:-}"

print_test() {
    echo -e "${BLUE}TEST: $1${NC}"
}

print_pass() {
    echo -e "${GREEN}✓ PASS: $1${NC}"
}

print_fail() {
    echo -e "${RED}✗ FAIL: $1${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

FAILED_TESTS=0

# Build curl command with optional API key
curl_cmd() {
    if [ -n "$API_KEY" ]; then
        curl -s -H "X-API-Key: $API_KEY" "$@"
    else
        curl -s "$@"
    fi
}

echo "================================"
echo "Configuration Service Test Suite"
echo "================================"
echo ""

# Test 1: Health Check
print_test "Health endpoint"
RESPONSE=$(curl -s ${BASE_URL}/health)
if echo "$RESPONSE" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
    print_pass "Health check returned healthy status"
else
    print_fail "Health check failed: $RESPONSE"
fi
echo ""

# Test 2: Readiness Check
print_test "Readiness endpoint"
RESPONSE=$(curl -s ${BASE_URL}/ready)
if echo "$RESPONSE" | jq -e '.status == "ready"' > /dev/null 2>&1; then
    DB_STATUS=$(echo "$RESPONSE" | jq -r '.checks.database')
    NATS_STATUS=$(echo "$RESPONSE" | jq -r '.checks.nats')
    print_pass "Readiness check passed (DB: $DB_STATUS, NATS: $NATS_STATUS)"
else
    print_fail "Readiness check failed: $RESPONSE"
fi
echo ""

# Test 3: List Configurations
print_test "List configurations"
RESPONSE=$(curl_cmd ${BASE_URL}/api/v1/configurations)
if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    COUNT=$(echo "$RESPONSE" | jq -r '.total')
    print_pass "Listed $COUNT configurations"
else
    print_fail "List configurations failed: $RESPONSE"
fi
echo ""

# Test 4: Get Configuration by Key
print_test "Get configuration by key"
RESPONSE=$(curl_cmd ${BASE_URL}/api/v1/configurations/key/default_strategy)
if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    KEY=$(echo "$RESPONSE" | jq -r '.data.key')
    VERSION=$(echo "$RESPONSE" | jq -r '.data.version')
    print_pass "Retrieved configuration '$KEY' (version $VERSION)"
else
    print_fail "Get by key failed: $RESPONSE"
fi
echo ""

# Test 5: Create Configuration
print_test "Create new configuration"
CREATE_DATA='{
  "key": "test_config_'$(date +%s)'",
  "type": "system",
  "value": {"name": "test", "value": true, "type": "boolean"},
  "format": "json",
  "description": "Test configuration",
  "created_by": "test_script"
}'
RESPONSE=$(curl_cmd -X POST ${BASE_URL}/api/v1/configurations \
  -H "Content-Type: application/json" \
  -d "$CREATE_DATA")

if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    CONFIG_ID=$(echo "$RESPONSE" | jq -r '.data.id')
    CONFIG_KEY=$(echo "$RESPONSE" | jq -r '.data.key')
    print_pass "Created configuration '$CONFIG_KEY' (ID: $CONFIG_ID)"
else
    print_fail "Create configuration failed: $RESPONSE"
    CONFIG_ID=""
fi
echo ""

# Test 6: Update Configuration (only if create succeeded)
if [ -n "$CONFIG_ID" ]; then
    print_test "Update configuration"
    UPDATE_DATA='{
      "value": {"name": "test", "value": false, "type": "boolean"},
      "format": "json",
      "description": "Updated test configuration",
      "updated_by": "test_script",
      "change_reason": "Testing update functionality"
    }'
    RESPONSE=$(curl_cmd -X PUT ${BASE_URL}/api/v1/configurations/${CONFIG_ID} \
      -H "Content-Type: application/json" \
      -d "$UPDATE_DATA")

    if echo "$RESPONSE" | jq -e '.success == true and .data.version == 2' > /dev/null 2>&1; then
        print_pass "Updated configuration (version 2)"
    else
        print_fail "Update configuration failed: $RESPONSE"
    fi
    echo ""

    # Test 7: Get Version History
    print_test "Get version history"
    RESPONSE=$(curl_cmd ${BASE_URL}/api/v1/configurations/${CONFIG_ID}/versions)
    if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
        VERSION_COUNT=$(echo "$RESPONSE" | jq '.data | length')
        print_pass "Retrieved $VERSION_COUNT versions"
    else
        print_fail "Get version history failed: $RESPONSE"
    fi
    echo ""

    # Test 8: Rollback Configuration
    print_test "Rollback configuration"
    ROLLBACK_DATA='{
      "version": 1,
      "rolled_back_by": "test_script",
      "reason": "Testing rollback"
    }'
    RESPONSE=$(curl_cmd -X POST ${BASE_URL}/api/v1/configurations/${CONFIG_ID}/rollback \
      -H "Content-Type: application/json" \
      -d "$ROLLBACK_DATA")

    if echo "$RESPONSE" | jq -e '.success == true and .data.version == 3' > /dev/null 2>&1; then
        print_pass "Rolled back configuration (version 3)"
    else
        print_fail "Rollback failed: $RESPONSE"
    fi
    echo ""

    # Test 9: Get Audit Logs
    print_test "Get audit logs"
    RESPONSE=$(curl_cmd ${BASE_URL}/api/v1/configurations/${CONFIG_ID}/audit-logs)
    if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
        AUDIT_COUNT=$(echo "$RESPONSE" | jq '.data | length')
        print_pass "Retrieved $AUDIT_COUNT audit log entries"
    else
        print_fail "Get audit logs failed: $RESPONSE"
    fi
    echo ""

    # Test 10: Deactivate Configuration
    print_test "Deactivate configuration"
    RESPONSE=$(curl_cmd -X POST ${BASE_URL}/api/v1/configurations/${CONFIG_ID}/deactivate)
    if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
        print_pass "Deactivated configuration"
    else
        print_fail "Deactivate failed: $RESPONSE"
    fi
    echo ""

    # Test 11: Activate Configuration
    print_test "Activate configuration"
    RESPONSE=$(curl_cmd -X POST ${BASE_URL}/api/v1/configurations/${CONFIG_ID}/activate)
    if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
        print_pass "Activated configuration"
    else
        print_fail "Activate failed: $RESPONSE"
    fi
    echo ""

    # Test 12: Delete Configuration
    print_test "Delete configuration"
    RESPONSE=$(curl_cmd -X DELETE ${BASE_URL}/api/v1/configurations/${CONFIG_ID})
    if echo "$RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
        print_pass "Deleted configuration"
    else
        print_fail "Delete failed: $RESPONSE"
    fi
    echo ""
fi

# Test 13: Metrics Endpoint
print_test "Metrics endpoint"
RESPONSE=$(curl -s ${BASE_URL}/metrics)
if echo "$RESPONSE" | grep -q "go_goroutines"; then
    print_pass "Metrics endpoint working"
else
    print_fail "Metrics endpoint failed"
fi
echo ""

# Summary
echo "================================"
echo "Test Summary"
echo "================================"
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}$FAILED_TESTS test(s) failed${NC}"
    exit 1
fi
