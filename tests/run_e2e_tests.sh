#!/bin/bash

# E2E Test Runner for B25 Trading System
# Runs end-to-end tests including performance benchmarks

set -e

echo "========================================="
echo "B25 End-to-End Test Suite"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DOCKER_COMPOSE_FILE="$SCRIPT_DIR/testutil/docker/docker-compose.test.yml"

# Parse command line arguments
RUN_BENCHMARKS=0
VERBOSE=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --benchmark)
            RUN_BENCHMARKS=1
            shift
            ;;
        --verbose)
            VERBOSE=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--benchmark] [--verbose]"
            exit 1
            ;;
    esac
done

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Docker is not running${NC}"
    exit 1
fi

# Function to check service health
check_service() {
    local service=$1
    local max_attempts=30
    local attempt=1

    echo -n "Waiting for $service to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f "$DOCKER_COMPOSE_FILE" ps | grep -q "$service.*healthy"; then
            echo -e " ${GREEN}OK${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
        ((attempt++))
    done

    echo -e " ${RED}FAILED${NC}"
    return 1
}

# Start test infrastructure
echo -e "\n${YELLOW}Starting test infrastructure...${NC}"
docker-compose -f "$DOCKER_COMPOSE_FILE" up -d

# Wait for services
echo -e "\n${YELLOW}Checking service health...${NC}"
check_service "redis-test" || { echo "Redis failed to start"; exit 1; }
check_service "nats-test" || { echo "NATS failed to start"; exit 1; }
check_service "postgres-test" || { echo "PostgreSQL failed to start"; exit 1; }
check_service "mock-exchange" || { echo "Mock Exchange failed to start"; exit 1; }

# Set environment variables
export REDIS_ADDR="localhost:6380"
export NATS_ADDR="nats://localhost:4223"
export POSTGRES_ADDR="localhost:5433"
export POSTGRES_USER="testuser"
export POSTGRES_PASSWORD="testpass"
export POSTGRES_DB="b25_test"

cd "$SCRIPT_DIR/e2e"

# Run E2E tests
echo -e "\n${YELLOW}Running E2E tests...${NC}"

TEST_FLAGS="-v -timeout 15m"
if [ $VERBOSE -eq 1 ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if go test $TEST_FLAGS -run "^Test" ./...; then
    echo -e "\n${GREEN}✓ All E2E tests passed!${NC}"
    E2E_RESULT=0
else
    echo -e "\n${RED}✗ Some E2E tests failed${NC}"
    E2E_RESULT=1
fi

# Run benchmarks if requested
if [ $RUN_BENCHMARKS -eq 1 ]; then
    echo -e "\n${BLUE}Running performance benchmarks...${NC}"

    if go test -v -timeout 20m -run "TestLatencyBenchmark" ./...; then
        echo -e "\n${GREEN}✓ Benchmarks completed!${NC}"
        BENCH_RESULT=0
    else
        echo -e "\n${RED}✗ Benchmark failures${NC}"
        BENCH_RESULT=1
    fi
else
    echo -e "\n${BLUE}Skipping benchmarks (use --benchmark to run)${NC}"
    BENCH_RESULT=0
fi

# Cleanup
if [ "${KEEP_RUNNING:-0}" != "1" ]; then
    echo -e "\n${YELLOW}Cleaning up test infrastructure...${NC}"
    docker-compose -f "$DOCKER_COMPOSE_FILE" down -v
    echo -e "${GREEN}Cleanup complete${NC}"
else
    echo -e "\n${YELLOW}Test infrastructure kept running (KEEP_RUNNING=1)${NC}"
    echo "To stop: docker-compose -f $DOCKER_COMPOSE_FILE down -v"
fi

# Exit with combined result
if [ $E2E_RESULT -eq 0 ] && [ $BENCH_RESULT -eq 0 ]; then
    exit 0
else
    exit 1
fi
