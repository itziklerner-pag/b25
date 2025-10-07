#!/bin/bash
set -e

# Deploy All B25 Services
# Deploys all services in dependency order

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}B25 Trading System Deployment${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Deployment order (dependencies first)
SERVICES=(
    "market-data"           # Foundation - provides data to all
    "configuration"         # Foundation - provides config
    "dashboard-server"      # UI aggregation
    "auth"                  # Authentication
    "api-gateway"           # API routing
    "account-monitor"       # Account tracking
    "risk-manager"          # Risk management
    "order-execution"       # Order execution
    "strategy-engine"       # Trading strategies
    "analytics"             # Analytics
)

FAILED_SERVICES=()
SUCCESSFUL_SERVICES=()

for service in "${SERVICES[@]}"; do
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Deploying: $service${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    SERVICE_DIR="services/$service"

    if [ ! -d "$SERVICE_DIR" ]; then
        echo -e "${YELLOW}⚠ Service directory not found: $SERVICE_DIR${NC}"
        FAILED_SERVICES+=("$service (not found)")
        continue
    fi

    if [ ! -f "$SERVICE_DIR/deploy.sh" ]; then
        echo -e "${YELLOW}⚠ No deploy.sh found for $service${NC}"
        FAILED_SERVICES+=("$service (no deploy.sh)")
        continue
    fi

    # Deploy
    cd "$SERVICE_DIR"

    if ./deploy.sh; then
        echo -e "${GREEN}✓ $service deployed successfully${NC}"
        SUCCESSFUL_SERVICES+=("$service")
    else
        echo -e "${RED}✗ $service deployment failed${NC}"
        FAILED_SERVICES+=("$service")
    fi

    cd - > /dev/null

    # Brief pause between services
    sleep 2
done

echo ""
echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Deployment Summary${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

echo "Successful deployments: ${#SUCCESSFUL_SERVICES[@]}"
for service in "${SUCCESSFUL_SERVICES[@]}"; do
    echo -e "  ${GREEN}✓${NC} $service"
done

echo ""
echo "Failed deployments: ${#FAILED_SERVICES[@]}"
for service in "${FAILED_SERVICES[@]}"; do
    echo -e "  ${RED}✗${NC} $service"
done

echo ""

if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
    echo -e "${GREEN}🎉 All services deployed successfully!${NC}"
    echo ""
    echo "Check service status:"
    echo "  ./check-all-services.sh"
    echo ""
    exit 0
else
    echo -e "${YELLOW}⚠ Some services failed to deploy${NC}"
    echo "Check logs for details"
    exit 1
fi
