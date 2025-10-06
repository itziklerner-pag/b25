#!/bin/bash

# Auth Service API Test Script
# Tests all authentication endpoints

set -e

BASE_URL="${BASE_URL:-http://localhost:9097}"
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="TestPass123@"

echo "========================================="
echo "AUTH SERVICE API TESTS"
echo "========================================="
echo "Base URL: $BASE_URL"
echo "Test Email: $TEST_EMAIL"
echo ""

# Test 1: Health Check
echo "[1/7] Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
HEALTH_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.data.status')
if [ "$HEALTH_STATUS" = "healthy" ]; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed"
    echo "$HEALTH_RESPONSE" | jq .
    exit 1
fi
echo ""

# Test 2: User Registration
echo "[2/7] Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
REGISTER_SUCCESS=$(echo "$REGISTER_RESPONSE" | jq -r '.success')
if [ "$REGISTER_SUCCESS" = "true" ]; then
    ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.accessToken')
    REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.refreshToken')
    echo "✓ Registration successful"
    echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "  Refresh Token: ${REFRESH_TOKEN:0:50}..."
else
    echo "✗ Registration failed"
    echo "$REGISTER_RESPONSE" | jq .
    exit 1
fi
echo ""

# Test 3: Login
echo "[3/7] Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
LOGIN_SUCCESS=$(echo "$LOGIN_RESPONSE" | jq -r '.success')
if [ "$LOGIN_SUCCESS" = "true" ]; then
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.accessToken')
    REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.refreshToken')
    echo "✓ Login successful"
else
    echo "✗ Login failed"
    echo "$LOGIN_RESPONSE" | jq .
    exit 1
fi
echo ""

# Test 4: Token Verification
echo "[4/7] Testing token verification..."
VERIFY_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/verify" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
VERIFY_SUCCESS=$(echo "$VERIFY_RESPONSE" | jq -r '.success')
if [ "$VERIFY_SUCCESS" = "true" ]; then
    USER_EMAIL=$(echo "$VERIFY_RESPONSE" | jq -r '.data.email')
    echo "✓ Token verification successful"
    echo "  Email: $USER_EMAIL"
else
    echo "✗ Token verification failed"
    echo "$VERIFY_RESPONSE" | jq .
    exit 1
fi
echo ""

# Test 5: Token Refresh
echo "[5/7] Testing token refresh..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}")
REFRESH_SUCCESS=$(echo "$REFRESH_RESPONSE" | jq -r '.success')
if [ "$REFRESH_SUCCESS" = "true" ]; then
    NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.accessToken')
    NEW_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.refreshToken')
    echo "✓ Token refresh successful"
    echo "  New Access Token: ${NEW_ACCESS_TOKEN:0:50}..."
    # Update tokens for logout test
    ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
    REFRESH_TOKEN="$NEW_REFRESH_TOKEN"
else
    echo "✗ Token refresh failed"
    echo "$REFRESH_RESPONSE" | jq .
    exit 1
fi
echo ""

# Test 6: Logout
echo "[6/7] Testing logout..."
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
    -H "Content-Type: application/json" \
    -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}")
LOGOUT_SUCCESS=$(echo "$LOGOUT_RESPONSE" | jq -r '.success')
if [ "$LOGOUT_SUCCESS" = "true" ]; then
    echo "✓ Logout successful"
else
    echo "✗ Logout failed"
    echo "$LOGOUT_RESPONSE" | jq .
    exit 1
fi
echo ""

# Test 7: Verify token after logout (should fail)
echo "[7/7] Testing token verification after logout..."
VERIFY_AFTER_LOGOUT=$(curl -s -X POST "$BASE_URL/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}")
VERIFY_AFTER_SUCCESS=$(echo "$VERIFY_AFTER_LOGOUT" | jq -r '.success')
if [ "$VERIFY_AFTER_SUCCESS" = "false" ]; then
    echo "✓ Token correctly invalidated after logout"
else
    echo "✗ Token should be invalid after logout"
    exit 1
fi
echo ""

echo "========================================="
echo "ALL TESTS PASSED ✓"
echo "========================================="
