#!/bin/bash
# Test Binance API with your credentials

source .env

echo "Testing Binance Testnet API..."
echo ""

# Test 1: Simple ping
echo "1. Testing basic connectivity (ping):"
PING_RESULT=$(curl -s "https://testnet.binancefuture.com/fapi/v1/ping")
echo "   Result: $PING_RESULT"
if [ "$PING_RESULT" = "{}" ]; then
    echo "   ✅ Ping successful"
else
    echo "   ❌ Ping failed"
fi

echo ""
echo "2. Testing server time:"
TIME_RESULT=$(curl -s "https://testnet.binancefuture.com/fapi/v1/time" | jq -r .serverTime)
echo "   Server time: $TIME_RESULT"

echo ""
echo "3. Testing authenticated endpoint (account info):"
TIMESTAMP=$(date +%s000)
QUERY="timestamp=$TIMESTAMP"
SIGNATURE=$(echo -n "$QUERY" | openssl dgst -sha256 -hmac "$EXCHANGE_SECRET_KEY" | awk '{print $2}')

ACCOUNT=$(curl -s -H "X-MBX-APIKEY: $EXCHANGE_API_KEY" \
  "https://testnet.binancefuture.com/fapi/v2/account?$QUERY&signature=$SIGNATURE")

echo "$ACCOUNT" | jq . 2>/dev/null || echo "   Response: $ACCOUNT"

if echo "$ACCOUNT" | grep -q "totalWalletBalance"; then
    echo ""
    echo "   ✅ API authentication successful!"
    echo "   Balance: $(echo $ACCOUNT | jq -r .totalWalletBalance) USDT"
else
    echo ""
    echo "   ❌ API authentication failed"
    echo "   Check your API keys in .env file"
fi
