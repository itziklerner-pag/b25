#!/bin/bash

# Test Binance Futures Account Access
# This script verifies sub-account access and checks futures balance

API_KEY="${BINANCE_API_KEY}"
SECRET_KEY="${BINANCE_SECRET_KEY}"
BASE_URL="https://fapi.binance.com"  # Futures API

if [ -z "$API_KEY" ] || [ -z "$SECRET_KEY" ]; then
    echo "Error: BINANCE_API_KEY and BINANCE_SECRET_KEY must be set"
    exit 1
fi

echo "========================================"
echo "Binance Futures Account Test"
echo "========================================"
echo ""

# Generate timestamp
TIMESTAMP=$(date +%s000)

# Function to create signature
create_signature() {
    local query_string="$1"
    echo -n "${query_string}" | openssl dgst -sha256 -hmac "${SECRET_KEY}" | awk '{print $2}'
}

echo "1. Testing Account Information (Account Name/ID)"
echo "------------------------------------------------"
QUERY_STRING="timestamp=${TIMESTAMP}"
SIGNATURE=$(create_signature "$QUERY_STRING")
ACCOUNT_URL="${BASE_URL}/fapi/v2/account?${QUERY_STRING}&signature=${SIGNATURE}"

echo "Endpoint: /fapi/v2/account"
echo ""
ACCOUNT_RESPONSE=$(curl -s -H "X-MBX-APIKEY: ${API_KEY}" "${ACCOUNT_URL}")
echo "$ACCOUNT_RESPONSE" | jq '.' 2>/dev/null || echo "$ACCOUNT_RESPONSE"
echo ""

# Extract account info
if echo "$ACCOUNT_RESPONSE" | jq -e '.totalWalletBalance' > /dev/null 2>&1; then
    TOTAL_BALANCE=$(echo "$ACCOUNT_RESPONSE" | jq -r '.totalWalletBalance')
    AVAILABLE_BALANCE=$(echo "$ACCOUNT_RESPONSE" | jq -r '.availableBalance')
    echo "✓ Total Wallet Balance: $TOTAL_BALANCE USDT"
    echo "✓ Available Balance: $AVAILABLE_BALANCE USDT"
    echo ""
fi

echo ""
echo "2. Testing Futures Balance"
echo "------------------------------------------------"
TIMESTAMP=$(date +%s000)
QUERY_STRING="timestamp=${TIMESTAMP}"
SIGNATURE=$(create_signature "$QUERY_STRING")
BALANCE_URL="${BASE_URL}/fapi/v2/balance?${QUERY_STRING}&signature=${SIGNATURE}"

echo "Endpoint: /fapi/v2/balance"
echo ""
BALANCE_RESPONSE=$(curl -s -H "X-MBX-APIKEY: ${API_KEY}" "${BALANCE_URL}")
echo "$BALANCE_RESPONSE" | jq '.' 2>/dev/null || echo "$BALANCE_RESPONSE"
echo ""

# Show USDT balance
if echo "$BALANCE_RESPONSE" | jq -e '.[0]' > /dev/null 2>&1; then
    echo "USDT Balance Details:"
    echo "$BALANCE_RESPONSE" | jq '.[] | select(.asset == "USDT")'
    echo ""
fi

echo ""
echo "3. Testing Positions"
echo "------------------------------------------------"
TIMESTAMP=$(date +%s000)
QUERY_STRING="timestamp=${TIMESTAMP}"
SIGNATURE=$(create_signature "$QUERY_STRING")
POSITION_URL="${BASE_URL}/fapi/v2/positionRisk?${QUERY_STRING}&signature=${SIGNATURE}"

echo "Endpoint: /fapi/v2/positionRisk"
echo ""
POSITION_RESPONSE=$(curl -s -H "X-MBX-APIKEY: ${API_KEY}" "${POSITION_URL}")
echo "$POSITION_RESPONSE" | jq '.[] | select(.positionAmt != "0")' 2>/dev/null || echo "$POSITION_RESPONSE" | jq '.[0:3]' 2>/dev/null || echo "$POSITION_RESPONSE"
echo ""

echo ""
echo "4. Testing API Key Permissions"
echo "------------------------------------------------"
TIMESTAMP=$(date +%s000)
QUERY_STRING="timestamp=${TIMESTAMP}"
SIGNATURE=$(create_signature "$QUERY_STRING")
API_PERMS_URL="${BASE_URL}/fapi/v1/apiTradingStatus?${QUERY_STRING}&signature=${SIGNATURE}"

echo "Endpoint: /fapi/v1/apiTradingStatus"
echo ""
API_PERMS_RESPONSE=$(curl -s -H "X-MBX-APIKEY: ${API_KEY}" "${API_PERMS_URL}")
echo "$API_PERMS_RESPONSE" | jq '.' 2>/dev/null || echo "$API_PERMS_RESPONSE"
echo ""

echo "========================================"
echo "Test Complete"
echo "========================================"
