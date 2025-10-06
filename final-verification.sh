#!/bin/bash
echo "🔍 B25 SYSTEM - COMPREHENSIVE FINAL VERIFICATION"
echo "=================================================="
echo ""

cd /home/mm/dev/b25

# Test infrastructure
echo "1. INFRASTRUCTURE SERVICES:"
docker compose -f docker-compose.simple.yml ps --format "  {{.Name}}: {{.State}}"

echo ""
echo "2. APPLICATION SERVICES:"
ps aux | grep -E "(market-data|bin/service)" | grep -v grep | awk '{print "  PID "$2": "$11}' | head -10

echo ""
echo "3. DOMAIN & SSL:"
echo "  Testing https://mm.itziklerner.com..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://mm.itziklerner.com 2>/dev/null)
echo "  HTTP Status: $STATUS"
if [ "$STATUS" = "200" ]; then
  echo "  ✅ Domain accessible via HTTPS!"
else
  echo "  ❌ Domain not accessible (HTTP $STATUS)"
fi

echo ""
echo "4. WEBSOCKET:"
echo "  Testing wss://mm.itziklerner.com/ws..."
echo "  (Check browser console for WebSocket connection)"

echo ""
echo "5. HEALTH ENDPOINTS (with CORS):"
for port in 8080 8081 8082 8083 8084 8085 8086; do
  CORS=$(curl -s -H "Origin: https://mm.itziklerner.com" http://localhost:$port/health -I 2>/dev/null | grep -i "access-control-allow-origin")
  if [ -n "$CORS" ]; then
    echo "  ✅ Port $port: CORS enabled"
  else
    echo "  ⚠️  Port $port: No CORS header"
  fi
done

echo ""
echo "6. LIVE DATA FLOW:"
echo "  Checking if Dashboard Server is broadcasting..."
tail -20 logs/dashboard-server.log | grep "Broadcasting to clients" | tail -3

echo ""
echo "7. MARKET DATA:"
echo "  Latest BTC price in Redis:"
docker exec b25-redis redis-cli GET market_data:BTCUSDT 2>/dev/null | jq -r '.last_price' 2>/dev/null || echo "  Error reading price"

echo ""
echo "8. SERVICE COUNT:"
RUNNING=$(ps aux | grep -E "(market-data|bin/service)" | grep -v grep | wc -l)
echo "  $RUNNING services running (expected: 9-10)"

echo ""
echo "=================================================="
echo "✅ Verification Complete"
echo ""
echo "Access your system at: https://mm.itziklerner.com"
