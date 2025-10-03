#!/bin/bash

# Integration Test Runner for B25 Trading System
# Runs all integration tests with proper setup and teardown

set -e

echo "========================================="
echo "B25 Integration Test Suite"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DOCKER_COMPOSE_FILE="$SCRIPT_DIR/testutil/docker/docker-compose.test.yml"

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

# Wait for services to be healthy
echo -e "\n${YELLOW}Checking service health...${NC}"
check_service "redis-test" || { echo "Redis failed to start"; exit 1; }
check_service "nats-test" || { echo "NATS failed to start"; exit 1; }
check_service "postgres-test" || { echo "PostgreSQL failed to start"; exit 1; }
check_service "mock-exchange" || { echo "Mock Exchange failed to start"; exit 1; }

# Set environment variables for tests
export REDIS_ADDR="localhost:6380"
export NATS_ADDR="nats://localhost:4223"
export POSTGRES_ADDR="localhost:5433"
export POSTGRES_USER="testuser"
export POSTGRES_PASSWORD="testpass"
export POSTGRES_DB="b25_test"

# Run integration tests
echo -e "\n${YELLOW}Running integration tests...${NC}"
cd "$SCRIPT_DIR/integration"

if go test -v -timeout 10m ./...; then
    echo -e "\n${GREEN}✓ All integration tests passed!${NC}"
    TEST_RESULT=0
else
    echo -e "\n${RED}✗ Some integration tests failed${NC}"
    TEST_RESULT=1
fi

# Cleanup (optional - comment out to keep infrastructure running)
if [ "${KEEP_RUNNING:-0}" != "1" ]; then
    echo -e "\n${YELLOW}Cleaning up test infrastructure...${NC}"
    docker-compose -f "$DOCKER_COMPOSE_FILE" down -v
    echo -e "${GREEN}Cleanup complete${NC}"
else
    echo -e "\n${YELLOW}Test infrastructure kept running (KEEP_RUNNING=1)${NC}"
    echo "To stop: docker-compose -f $DOCKER_COMPOSE_FILE down -v"
fi

exit $TEST_RESULT
