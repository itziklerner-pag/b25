#!/bin/bash

# API Gateway Authentication Testing Script

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8000}"

echo "========================================="
echo "API Gateway Authentication Tests"
echo "========================================="
echo ""

# Test 1: No authentication
echo "1. Testing endpoint without authentication (should fail)"
response=$(curl -s -w "\n%{http_code}" "${GATEWAY_URL}/api/v1/account/balance")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 401 ]; then
    echo -e "${GREEN}✓${NC} Correctly rejected (401 Unauthorized)"
else
    echo -e "${RED}✗${NC} Expected 401, got $code"
fi
echo ""

# Test 2: Invalid API key
echo "2. Testing with invalid API key (should fail)"
response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: invalid-key" "${GATEWAY_URL}/api/v1/account/balance")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 401 ]; then
    echo -e "${GREEN}✓${NC} Correctly rejected (401 Unauthorized)"
else
    echo -e "${RED}✗${NC} Expected 401, got $code"
fi
echo ""

# Test 3: Valid admin API key
echo "3. Testing with valid admin API key (should pass auth)"
response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: test-admin-key-123" "${GATEWAY_URL}/api/v1/account/balance")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 503 ] || [ "$code" -eq 502 ] || [ "$code" -eq 200 ]; then
    echo -e "${GREEN}✓${NC} Authentication passed (got $code - backend may be down)"
else
    echo -e "${RED}✗${NC} Expected 503/502/200, got $code"
fi
echo ""

# Test 4: Valid operator API key
echo "4. Testing with valid operator API key (should pass auth)"
response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: test-operator-key-456" "${GATEWAY_URL}/api/v1/account/balance")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 503 ] || [ "$code" -eq 502 ] || [ "$code" -eq 200 ]; then
    echo -e "${GREEN}✓${NC} Authentication passed (got $code - backend may be down)"
else
    echo -e "${RED}✗${NC} Expected 503/502/200, got $code"
fi
echo ""

# Test 5: Valid viewer API key
echo "5. Testing with valid viewer API key (should pass auth)"
response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: test-viewer-key-789" "${GATEWAY_URL}/api/v1/account/balance")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 503 ] || [ "$code" -eq 502 ] || [ "$code" -eq 200 ]; then
    echo -e "${GREEN}✓${NC} Authentication passed (got $code - backend may be down)"
else
    echo -e "${RED}✗${NC} Expected 503/502/200, got $code"
fi
echo ""

# Test 6: Role-based access (operator-only endpoint with viewer key)
echo "6. Testing RBAC - viewer accessing operator endpoint (should fail)"
response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: test-viewer-key-789" "${GATEWAY_URL}/api/v1/orders")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 403 ]; then
    echo -e "${GREEN}✓${NC} Correctly rejected (403 Forbidden)"
else
    echo -e "${YELLOW}⚠${NC} Expected 403, got $code (RBAC might not be enforced)"
fi
echo ""

# Test 7: Role-based access (operator endpoint with operator key)
echo "7. Testing RBAC - operator accessing operator endpoint (should pass)"
response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: test-operator-key-456" "${GATEWAY_URL}/api/v1/orders")
code=$(echo "$response" | tail -n1)
if [ "$code" -eq 503 ] || [ "$code" -eq 502 ] || [ "$code" -eq 200 ]; then
    echo -e "${GREEN}✓${NC} Authentication and authorization passed (got $code)"
else
    echo -e "${RED}✗${NC} Expected 503/502/200, got $code"
fi
echo ""

echo "========================================="
echo "Authentication Tests Complete"
echo "========================================="
