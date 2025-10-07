#!/bin/bash

# Check status of all B25 services

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}B25 Services Health Check${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Check systemd services
SYSTEMD_SERVICES=(
    "market-data:8080"
)

# Check manual processes
MANUAL_SERVICES=(
    "dashboard-server:8086"
    "configuration:8085"
    "strategy-engine:9092"
    "risk-manager:9095"
    "order-execution:9091"
    "account-monitor:9093"
    "api-gateway:8000"
    "auth:9097"
    "analytics:9097"
)

total=0
running=0

# Check systemd services
for service_port in "${SYSTEMD_SERVICES[@]}"; do
    IFS=':' read -r service port <<< "$service_port"
    total=$((total + 1))

    printf "%-20s " "$service"

    if sudo systemctl is-active --quiet "$service" 2>/dev/null; then
        # Check health endpoint
        if curl -sf "http://localhost:$port/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Running (systemd + healthy)${NC}"
            running=$((running + 1))
        else
            echo -e "${YELLOW}⚠ Running (systemd, health unavailable)${NC}"
            running=$((running + 1))
        fi
    else
        echo -e "${RED}✗ Not running${NC}"
    fi
done

# Check manual processes
for service_port in "${MANUAL_SERVICES[@]}"; do
    IFS=':' read -r service port <<< "$service_port"
    total=$((total + 1))

    printf "%-20s " "$service"

    if pgrep -f "$service" > /dev/null 2>&1; then
        # Check health endpoint
        if curl -sf "http://localhost:$port/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Running (manual + healthy)${NC}"
            running=$((running + 1))
        else
            echo -e "${YELLOW}⚠ Running (manual, health unavailable)${NC}"
            running=$((running + 1))
        fi
    else
        echo -e "${RED}✗ Not running${NC}"
    fi
done

echo ""
echo -e "${BLUE}================================${NC}"
echo "Services: $running/$total running"
echo -e "${BLUE}================================${NC}"
echo ""

if [ $running -eq $total ]; then
    echo -e "${GREEN}✓ All services operational!${NC}"
    exit 0
elif [ $running -gt 0 ]; then
    echo -e "${YELLOW}⚠ Partial deployment${NC}"
    exit 1
else
    echo -e "${RED}✗ No services running${NC}"
    exit 2
fi
