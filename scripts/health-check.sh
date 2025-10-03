#!/bin/bash
# Health check script for B25 trading platform

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "================================================"
echo "B25 Trading Platform - Health Check"
echo "================================================"
echo ""

# Services to check
declare -A SERVICES=(
    ["API Gateway"]="http://localhost:8000/health"
    ["Auth Service"]="http://localhost:8001/health"
    ["Market Data"]="http://localhost:8080/health"
    ["Order Execution"]="http://localhost:8081/health"
    ["Strategy Engine"]="http://localhost:8082/health"
    ["Risk Manager"]="http://localhost:8083/health"
    ["Account Monitor"]="http://localhost:8084/health"
    ["Configuration"]="http://localhost:8085/health"
    ["Dashboard Server"]="http://localhost:8086/health"
)

# Infrastructure services
declare -A INFRA=(
    ["Redis"]="6379"
    ["PostgreSQL"]="5432"
    ["TimescaleDB"]="5433"
    ["NATS"]="http://localhost:8222/healthz"
)

TOTAL=0
HEALTHY=0
UNHEALTHY=0

echo -e "${BLUE}Application Services:${NC}"
for name in "${!SERVICES[@]}"; do
    url="${SERVICES[$name]}"
    TOTAL=$((TOTAL + 1))

    if curl -f -s -m 5 "$url" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} $name"
        HEALTHY=$((HEALTHY + 1))
    else
        echo -e "  ${RED}✗${NC} $name (${url})"
        UNHEALTHY=$((UNHEALTHY + 1))
    fi
done

echo ""
echo -e "${BLUE}Infrastructure Services:${NC}"
for name in "${!INFRA[@]}"; do
    check="${INFRA[$name]}"
    TOTAL=$((TOTAL + 1))

    if [[ $check == http* ]]; then
        # HTTP check
        if curl -f -s -m 5 "$check" > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $name"
            HEALTHY=$((HEALTHY + 1))
        else
            echo -e "  ${RED}✗${NC} $name"
            UNHEALTHY=$((UNHEALTHY + 1))
        fi
    else
        # Port check
        if nc -z localhost "$check" 2>/dev/null; then
            echo -e "  ${GREEN}✓${NC} $name"
            HEALTHY=$((HEALTHY + 1))
        else
            echo -e "  ${RED}✗${NC} $name (port $check)"
            UNHEALTHY=$((UNHEALTHY + 1))
        fi
    fi
done

echo ""
echo "================================================"
echo "Health Check Summary:"
echo "  Total Services:   $TOTAL"
echo -e "  Healthy:          ${GREEN}$HEALTHY${NC}"
echo -e "  Unhealthy:        ${RED}$UNHEALTHY${NC}"
echo "================================================"

if [ $UNHEALTHY -eq 0 ]; then
    echo -e "${GREEN}All services are healthy!${NC}"
    exit 0
else
    echo -e "${YELLOW}Some services are unhealthy. Check logs for details.${NC}"
    echo ""
    echo "View logs with:"
    echo "  docker-compose -f docker/docker-compose.dev.yml logs -f"
    exit 1
fi
