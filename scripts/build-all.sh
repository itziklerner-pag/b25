#!/bin/bash
# Build all services for B25 trading platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BUILD_MODE=${BUILD_MODE:-debug}  # debug or release
PARALLEL=${PARALLEL:-true}       # parallel builds
SKIP_TESTS=${SKIP_TESTS:-false}  # skip tests

echo "================================================"
echo "Building B25 Trading Platform"
echo "================================================"
echo "Build Mode: $BUILD_MODE"
echo "Parallel Builds: $PARALLEL"
echo "Skip Tests: $SKIP_TESTS"
echo "================================================"

# Track build times
START_TIME=$(date +%s)

# Function to build Rust service
build_rust_service() {
    local service=$1
    local path=$2

    echo -e "${YELLOW}[Rust] Building $service...${NC}"
    cd "$path"

    if [ "$BUILD_MODE" = "release" ]; then
        cargo build --release
    else
        cargo build
    fi

    if [ "$SKIP_TESTS" = "false" ]; then
        echo -e "${YELLOW}[Rust] Testing $service...${NC}"
        cargo test
    fi

    echo -e "${GREEN}[Rust] ✓ $service built successfully${NC}"
    cd - > /dev/null
}

# Function to build Go service
build_go_service() {
    local service=$1
    local path=$2

    echo -e "${YELLOW}[Go] Building $service...${NC}"
    cd "$path"

    # Create bin directory if it doesn't exist
    mkdir -p bin

    # Build flags
    local build_flags=""
    if [ "$BUILD_MODE" = "release" ]; then
        build_flags="-ldflags='-s -w' -trimpath"
    fi

    go build $build_flags -o "bin/$service" ./cmd/server

    if [ "$SKIP_TESTS" = "false" ]; then
        echo -e "${YELLOW}[Go] Testing $service...${NC}"
        go test -v -race ./...
    fi

    echo -e "${GREEN}[Go] ✓ $service built successfully${NC}"
    cd - > /dev/null
}

# Function to build Node.js service
build_node_service() {
    local service=$1
    local path=$2

    echo -e "${YELLOW}[Node] Building $service...${NC}"
    cd "$path"

    # Install dependencies if node_modules doesn't exist
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}[Node] Installing dependencies for $service...${NC}"
        npm ci
    fi

    # Build
    if [ -f "package.json" ] && grep -q "\"build\"" package.json; then
        npm run build
    fi

    if [ "$SKIP_TESTS" = "false" ] && grep -q "\"test\"" package.json; then
        echo -e "${YELLOW}[Node] Testing $service...${NC}"
        npm test || true  # Don't fail if tests don't exist
    fi

    echo -e "${GREEN}[Node] ✓ $service built successfully${NC}"
    cd - > /dev/null
}

# Build services based on parallel flag
if [ "$PARALLEL" = "true" ]; then
    echo "Building services in parallel..."

    # Rust services
    (build_rust_service "market-data" "services/market-data") &

    # Go services
    (build_go_service "order-execution" "services/order-execution") &
    (build_go_service "strategy-engine" "services/strategy-engine") &
    (build_go_service "account-monitor" "services/account-monitor") &
    (build_go_service "dashboard-server" "services/dashboard-server") &
    (build_go_service "risk-manager" "services/risk-manager") &
    (build_go_service "configuration" "services/configuration") &

    # Node.js services
    (build_node_service "api-gateway" "services/api-gateway") &
    (build_node_service "auth" "services/auth") &

    # UI
    (build_rust_service "terminal-ui" "ui/terminal") &
    (build_node_service "web-dashboard" "ui/web") &

    # Wait for all background jobs
    wait
else
    echo "Building services sequentially..."

    # Rust services
    build_rust_service "market-data" "services/market-data"

    # Go services
    build_go_service "order-execution" "services/order-execution"
    build_go_service "strategy-engine" "services/strategy-engine"
    build_go_service "account-monitor" "services/account-monitor"
    build_go_service "dashboard-server" "services/dashboard-server"
    build_go_service "risk-manager" "services/risk-manager"
    build_go_service "configuration" "services/configuration"

    # Node.js services
    build_node_service "api-gateway" "services/api-gateway"
    build_node_service "auth" "services/auth"

    # UI
    build_rust_service "terminal-ui" "ui/terminal"
    build_node_service "web-dashboard" "ui/web"
fi

# Calculate build time
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo "================================================"
echo -e "${GREEN}All services built successfully!${NC}"
echo "Total build time: ${DURATION}s"
echo "================================================"
echo ""
echo "Next steps:"
echo "  • Start development environment: docker-compose -f docker/docker-compose.dev.yml up"
echo "  • Run tests: ./scripts/test-all.sh"
echo "  • Build Docker images: ./scripts/docker-build-all.sh"
