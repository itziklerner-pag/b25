#!/bin/bash
echo "Testing Dashboard Server data..."
echo ""

echo "1. Redis has market data:"
docker exec b25-redis redis-cli --raw GET "market_data:BTCUSDT" | jq -c '{symbol, last_price, bid_price, ask_price}'

echo ""
echo "2. Dashboard Server health:"
curl -s http://localhost:8086/health | jq .

echo ""
echo "3. Strategy Engine status:"
curl -s http://localhost:8082/status 2>/dev/null | jq . || echo "No /status endpoint"

echo ""
echo "4. Checking Dashboard Server state in Redis cache:"
docker exec b25-redis redis-cli --scan --pattern "dashboard:*" | head -5

echo ""
echo "Done"
