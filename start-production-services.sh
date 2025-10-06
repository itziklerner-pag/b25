#!/bin/bash

###############################################################################
# B25 Trading System - Production Services Startup
# Starts all services for production deployment with HTTPS
###############################################################################

set -e

B25_ROOT="/home/mm/dev/b25"
WEB_DIR="$B25_ROOT/ui/web"
API_GATEWAY_DIR="$B25_ROOT/services/api-gateway"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}B25 Trading System - Production Start${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check if port is in use
check_port() {
    local port=$1
    if nc -z localhost "$port" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to wait for service
wait_for_service() {
    local name=$1
    local port=$2
    local max_attempts=30
    local attempt=0

    echo -e "${YELLOW}Waiting for $name on port $port...${NC}"

    while [ $attempt -lt $max_attempts ]; do
        if check_port "$port"; then
            echo -e "${GREEN}✓ $name is ready${NC}"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done

    echo -e "${RED}✗ $name failed to start${NC}"
    return 1
}

echo -e "${YELLOW}Step 1: Checking Nginx...${NC}"
if systemctl is-active --quiet nginx; then
    echo -e "${GREEN}✓ Nginx is running${NC}"
else
    echo -e "${YELLOW}Starting Nginx...${NC}"
    sudo systemctl start nginx
    echo -e "${GREEN}✓ Nginx started${NC}"
fi

echo ""
echo -e "${YELLOW}Step 2: Building Web Dashboard for Production...${NC}"
if [ -d "$WEB_DIR" ]; then
    cd "$WEB_DIR"

    # Check if .env is configured for production
    if grep -q "mm.itziklerner.com" .env; then
        echo -e "${GREEN}✓ Environment configured for production${NC}"
    else
        echo -e "${YELLOW}⚠ Warning: .env may not be configured for production${NC}"
    fi

    # Build the application
    echo -e "${YELLOW}Building application...${NC}"
    npm run build
    echo -e "${GREEN}✓ Build complete${NC}"

    # Kill existing dev server if running
    pkill -f "vite.*3000" || true
    sleep 2

    # Start production preview server in background
    echo -e "${YELLOW}Starting production server...${NC}"
    nohup npm run preview > "$B25_ROOT/logs/web-dashboard.log" 2>&1 &

    wait_for_service "Web Dashboard" 3000
else
    echo -e "${RED}✗ Web directory not found: $WEB_DIR${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 3: Starting API Gateway...${NC}"
if [ -d "$API_GATEWAY_DIR" ]; then
    cd "$API_GATEWAY_DIR"

    # Check if dependencies are installed
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}Installing dependencies...${NC}"
        npm install
    fi

    # Kill existing API Gateway if running
    pkill -f "node.*8000" || true
    sleep 2

    # Start API Gateway in background
    echo -e "${YELLOW}Starting API Gateway...${NC}"
    nohup npm start > "$B25_ROOT/logs/api-gateway.log" 2>&1 &

    wait_for_service "API Gateway" 8000
else
    echo -e "${YELLOW}⚠ API Gateway directory not found, skipping${NC}"
fi

echo ""
echo -e "${YELLOW}Step 4: Checking Dashboard Server (WebSocket)...${NC}"
if check_port 8086; then
    echo -e "${GREEN}✓ Dashboard Server is running${NC}"
else
    echo -e "${YELLOW}⚠ Dashboard Server is not running${NC}"
    echo -e "  Please start it manually from: $B25_ROOT/services/dashboard-server"
fi

echo ""
echo -e "${YELLOW}Step 5: Checking Grafana...${NC}"
if check_port 3001; then
    echo -e "${GREEN}✓ Grafana is running${NC}"
else
    echo -e "${YELLOW}⚠ Grafana is not running${NC}"
    echo -e "  Please start it manually if needed"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Service Status Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check all services
services=(
    "Nginx:443"
    "Web Dashboard:3000"
    "API Gateway:8000"
    "Dashboard Server:8086"
    "Grafana:3001"
)

for service_info in "${services[@]}"; do
    IFS=':' read -r name port <<< "$service_info"
    if [ "$name" = "Nginx" ]; then
        if systemctl is-active --quiet nginx; then
            echo -e "${GREEN}✓${NC} $name - Running"
        else
            echo -e "${RED}✗${NC} $name - Not Running"
        fi
    else
        if check_port "$port"; then
            echo -e "${GREEN}✓${NC} $name (port $port) - Running"
        else
            echo -e "${RED}✗${NC} $name (port $port) - Not Running"
        fi
    fi
done

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Production URLs${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Web Dashboard:${NC}  https://mm.itziklerner.com"
echo -e "${GREEN}API Gateway:${NC}    https://mm.itziklerner.com/api"
echo -e "${GREEN}WebSocket:${NC}      wss://mm.itziklerner.com/ws"
echo -e "${GREEN}Grafana:${NC}        https://mm.itziklerner.com/grafana"
echo -e "${GREEN}Health Check:${NC}   https://mm.itziklerner.com/health"
echo ""
echo -e "${YELLOW}Logs:${NC}"
echo -e "  Web Dashboard: $B25_ROOT/logs/web-dashboard.log"
echo -e "  API Gateway:   $B25_ROOT/logs/api-gateway.log"
echo -e "  Nginx Access:  /var/log/nginx/b25-access.log"
echo -e "  Nginx Error:   /var/log/nginx/b25-error.log"
echo ""
echo -e "${GREEN}All services started successfully!${NC}"
echo ""
