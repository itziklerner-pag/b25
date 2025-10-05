#!/bin/bash
echo "🔍 B25 SYSTEM HEALTH REPORT"
echo "================================="
echo ""

cd /home/mm/dev/b25

echo "1. INFRASTRUCTURE (Docker Containers):"
docker compose -f docker-compose.simple.yml ps --format "  {{.Name}}: {{.State}}"

echo ""
echo "2. APPLICATION SERVICES:"
ps aux | grep -E "(market-data-service|bin/service|node src/server)" | grep -v grep | awk '{print "  PID "$2": "$11" "$12" "$13}' | head -10

echo ""
echo "3. HEALTH ENDPOINT CHECKS:"
curl -s http://localhost:8080/health | jq -r '"  Market Data (8080): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Market Data (8080): Failed"
curl -s http://localhost:8081/health | jq -r '"  Order Execution (8081): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Order Execution (8081): Failed"
curl -s http://localhost:8082/health | jq -r '"  Strategy Engine (8082): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Strategy Engine (8082): Failed"
curl -s http://localhost:8083/health | jq -r '"  Risk Manager (8083): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Risk Manager (8083): Failed"
curl -s http://localhost:8084/health | jq -r '"  Account Monitor (8084): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Account Monitor (8084): Failed"
curl -s http://localhost:8085/health | jq -r '"  Configuration (8085): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Configuration (8085): Failed"
curl -s http://localhost:8086/health | jq -r '"  Dashboard Server (8086): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ Dashboard Server (8086): Failed"
curl -s http://localhost:8000/health | jq -r '"  API Gateway (8000): \(.status // "ERROR")"' 2>/dev/null || echo "  ❌ API Gateway (8000): Failed"
curl -s http://localhost:9097/health | jq -r '"  Auth Service (9097): \(.data.status // "ERROR")"' 2>/dev/null || echo "  ❌ Auth Service (9097): Failed"

echo ""
echo "4. LOG FILE SIZES:"
ls -lh logs/*.log 2>/dev/null | awk '{print "  "$9": "$5}' | tail -10

echo ""
echo "================================="
echo "Report complete at $(date)"
