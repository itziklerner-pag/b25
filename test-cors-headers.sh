#!/bin/bash

# CORS Headers Test Script
# Tests all service health endpoints for proper CORS headers

echo "=========================================="
echo "Testing CORS Headers on Health Endpoints"
echo "=========================================="
echo ""

# Define services and their ports
declare -A services=(
    ["api-gateway"]="8080"
    ["order-execution"]="8081"
    ["strategy-engine"]="8082"
    ["account-monitor"]="8083"
    ["configuration"]="8084"
    ["risk-manager"]="8085"
    ["dashboard-server"]="8086"
    ["market-data"]="8087"
)

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_cors() {
    local service=$1
    local port=$2
    local url="http://localhost:${port}/health"
    
    echo "Testing: ${service} (${url})"
    echo "----------------------------------------"
    
    # Test GET request with Origin header
    response=$(curl -s -H "Origin: http://localhost:3000" -i "${url}" 2>&1)
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}✗ Service not running or unreachable${NC}"
        echo ""
        return 1
    fi
    
    # Check for CORS headers
    has_origin=$(echo "$response" | grep -i "Access-Control-Allow-Origin")
    has_methods=$(echo "$response" | grep -i "Access-Control-Allow-Methods")
    has_headers=$(echo "$response" | grep -i "Access-Control-Allow-Headers")
    
    if [ -n "$has_origin" ] && [ -n "$has_methods" ] && [ -n "$has_headers" ]; then
        echo -e "${GREEN}✓ CORS headers present${NC}"
        echo "  - ${has_origin}"
        echo "  - ${has_methods}"
        echo "  - ${has_headers}"
    else
        echo -e "${RED}✗ Missing CORS headers${NC}"
        [ -z "$has_origin" ] && echo "  - Missing: Access-Control-Allow-Origin"
        [ -z "$has_methods" ] && echo "  - Missing: Access-Control-Allow-Methods"
        [ -z "$has_headers" ] && echo "  - Missing: Access-Control-Allow-Headers"
    fi
    
    # Test OPTIONS preflight request
    echo ""
    echo "Testing OPTIONS preflight..."
    options_response=$(curl -s -X OPTIONS -H "Origin: http://localhost:3000" -i "${url}" 2>&1)
    
    if echo "$options_response" | grep -q "200 OK"; then
        echo -e "${GREEN}✓ OPTIONS request handled correctly${NC}"
    else
        echo -e "${YELLOW}⚠ OPTIONS request may not be handled${NC}"
    fi
    
    echo ""
}

# Run tests for all services
for service in "${!services[@]}"; do
    test_cors "$service" "${services[$service]}"
done

echo "=========================================="
echo "Test Complete"
echo "=========================================="
