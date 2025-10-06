#!/bin/bash

# Service Rebuild Script
# Rebuilds all services with CORS header fixes

set -e  # Exit on error

echo "=========================================="
echo "Rebuilding All Services with CORS Fixes"
echo "=========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Base directory
BASE_DIR="/home/mm/dev/b25/services"

# Function to build Go service
build_go_service() {
    local service=$1
    echo -e "${YELLOW}Building ${service}...${NC}"
    cd "${BASE_DIR}/${service}"
    if [ -f "cmd/server/main.go" ]; then
        go build -o "bin/${service}" ./cmd/server/main.go
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ ${service} built successfully${NC}"
        else
            echo -e "${RED}✗ ${service} build failed${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ main.go not found for ${service}${NC}"
        return 1
    fi
    echo ""
}

# Function to build Rust service
build_rust_service() {
    local service=$1
    echo -e "${YELLOW}Building ${service} (Rust)...${NC}"
    cd "${BASE_DIR}/${service}"
    cargo build --release
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ ${service} built successfully${NC}"
    else
        echo -e "${RED}✗ ${service} build failed${NC}"
        return 1
    fi
    echo ""
}

# Build Go services
echo "Building Go Services..."
echo "----------------------------------------"
build_go_service "order-execution"
build_go_service "strategy-engine"
build_go_service "account-monitor"
build_go_service "configuration"
build_go_service "risk-manager"
build_go_service "dashboard-server"
build_go_service "api-gateway"

# Build Rust service
echo "Building Rust Services..."
echo "----------------------------------------"
build_rust_service "market-data"

echo "=========================================="
echo "Build Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Stop all running services"
echo "2. Start services with the new binaries"
echo "3. Run CORS test script: ./test-cors-headers.sh"
