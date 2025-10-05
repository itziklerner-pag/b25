#!/bin/bash
echo "🔍 VPN Connection Test"
echo "======================"
echo ""

# Check WireGuard status
echo "1. WireGuard Interface Status:"
sudo wg show wg0 2>/dev/null || echo "  ❌ VPN not running"

echo ""
echo "2. Network Interface Check:"
ip addr show wg0 2>/dev/null | grep "inet " || echo "  ❌ No wg0 interface"

echo ""
echo "3. Public IP Test:"
echo "  Without VPN routing: $(curl -s --max-time 3 ifconfig.me)"
echo "  Testing through VPN interface..."
VPN_IP=$(curl -s --max-time 5 --interface wg0 ifconfig.me 2>/dev/null || echo "FAILED")
echo "  Through VPN: $VPN_IP"

echo ""
echo "4. Binance Testnet API Test:"
echo "  Testing Binance Futures Testnet ping..."
BINANCE_RESPONSE=$(curl -s --max-time 5 "https://testnet.binancefuture.com/fapi/v1/ping" 2>/dev/null)
if [ "$BINANCE_RESPONSE" = "{}" ]; then
    echo "  ✅ Binance API accessible!"
else
    echo "  ❌ Binance API failed: $BINANCE_RESPONSE"
fi

echo ""
echo "5. Binance Account Endpoint Test (with API key):"
# Load API keys
export $(cat .env | grep EXCHANGE_API_KEY | xargs)
TIMESTAMP=$(date +%s000)
QUERY_STRING="timestamp=$TIMESTAMP"
SIGNATURE=$(echo -n "$QUERY_STRING" | openssl dgst -sha256 -hmac "$EXCHANGE_SECRET_KEY" | awk '{print $2}')
ACCOUNT_URL="https://testnet.binancefuture.com/fapi/v2/account?$QUERY_STRING&signature=$SIGNATURE"

ACCOUNT_RESPONSE=$(curl -s --max-time 5 -H "X-MBX-APIKEY: $EXCHANGE_API_KEY" "$ACCOUNT_URL")
if echo "$ACCOUNT_RESPONSE" | grep -q "totalWalletBalance"; then
    echo "  ✅ Account API accessible! Balance data received."
else
    echo "  Status: $ACCOUNT_RESPONSE" | head -c 200
fi

echo ""
echo "======================"
echo "Test complete"
