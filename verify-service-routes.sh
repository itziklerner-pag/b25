#!/bin/bash
# Verify all B25 service routes through Nginx

echo "=========================================="
echo "B25 Service Routes Verification"
echo "=========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service endpoints
services=(
    "market-data:8080"
    "order-execution:8081"
    "strategy-engine:8082"
    "risk-manager:8083"
    "account-monitor:8084"
    "configuration:8085"
    "dashboard-server:8086"
    "api-gateway:8000"
    "auth-service:9097"
    "prometheus:9090"
    "grafana-internal:3001"
)

healthy=0
unhealthy=0
degraded=0

for service_port in "${services[@]}"; do
    IFS=':' read -r service port <<< "$service_port"
    url="https://mm.itziklerner.com/services/${service}/health"

    echo -n "Testing $service (port $port)... "

    # Make request with timeout
    response=$(curl -s -w "\n%{http_code}" --max-time 5 "$url" 2>&1)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}HEALTHY${NC}"
        echo "  Response: $body"
        ((healthy++))
    elif [ "$http_code" = "404" ]; then
        echo -e "${RED}NOT FOUND (404)${NC}"
        ((unhealthy++))
    elif [ "$http_code" = "502" ] || [ "$http_code" = "503" ]; then
        echo -e "${RED}SERVICE UNAVAILABLE ($http_code)${NC}"
        ((unhealthy++))
    elif [ -z "$http_code" ]; then
        echo -e "${RED}TIMEOUT/ERROR${NC}"
        ((unhealthy++))
    else
        echo -e "${YELLOW}DEGRADED (HTTP $http_code)${NC}"
        echo "  Response: $body"
        ((degraded++))
    fi
    echo ""
done

echo "=========================================="
echo "Summary:"
echo -e "  ${GREEN}Healthy: $healthy${NC}"
echo -e "  ${YELLOW}Degraded: $degraded${NC}"
echo -e "  ${RED}Unhealthy: $unhealthy${NC}"
echo "=========================================="

# Check if ServiceMonitor.tsx is using correct URLs
echo ""
echo "Verifying ServiceMonitor.tsx configuration..."
if grep -q "https://mm.itziklerner.com/services/market-data/health" /home/mm/dev/b25/ui/web/src/components/ServiceMonitor.tsx; then
    echo -e "${GREEN}✓ ServiceMonitor.tsx is using correct URLs${NC}"
else
    echo -e "${RED}✗ ServiceMonitor.tsx is NOT using correct URLs${NC}"
fi

# Check Nginx config
echo ""
echo "Verifying Nginx configuration..."
if sudo nginx -t 2>&1 | grep -q "test is successful"; then
    echo -e "${GREEN}✓ Nginx configuration is valid${NC}"
else
    echo -e "${RED}✗ Nginx configuration has errors${NC}"
fi

echo ""
echo "=========================================="
echo "Next Steps:"
echo "1. Open https://mm.itziklerner.com/system"
echo "2. All healthy services should show green"
echo "3. Services should auto-refresh every 10 seconds"
echo "=========================================="
