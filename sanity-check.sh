#!/bin/bash
echo "🔍 B25 System Sanity Check"
echo "=========================="
echo ""

cd /home/mm/dev/b25

# Check infrastructure
echo "1. Infrastructure Services:"
docker compose -f docker-compose.simple.yml ps --format "  {{.Name}}: {{.State}}" | grep -v "^$"

echo ""
echo "2. Trading Services:"
ps aux | grep -E "(market-data-service|bin/service|node src/server)" | grep -v grep | awk '{print "  "$2" "$11" "$12" "$13}' | head -10

echo ""
echo "3. Health Checks:"
for port in 8080 8081 8082 8086 9097; do
  response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$port/health 2>/dev/null)
  if [ "$response" = "200" ]; then
    echo "  ✅ Port $port: OK"
  else
    echo "  ❌ Port $port: Failed ($response)"
  fi
done

echo ""
echo "4. Log Files:"
ls -lh logs/*.log | awk '{print "  "$9" ("$5")"}'

echo ""
echo "5. Process Count:"
echo "  Total services: $(ps aux | grep -E '(market-data-service|bin/service|node src/server)' | grep -v grep | wc -l)"

echo ""
echo "=========================="
echo "✅ Sanity Check Complete"
