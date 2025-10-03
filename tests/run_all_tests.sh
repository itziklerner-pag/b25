#!/bin/bash

# Complete Test Runner for B25 Trading System
# Runs all tests: integration, e2e, and benchmarks

set -e

echo "========================================="
echo "B25 Complete Test Suite"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Track results
INTEGRATION_RESULT=0
E2E_RESULT=0
BENCHMARK_RESULT=0

# Run integration tests
echo -e "\n${BLUE}=========================================${NC}"
echo -e "${BLUE}Running Integration Tests${NC}"
echo -e "${BLUE}=========================================${NC}\n"

if KEEP_RUNNING=1 bash "$SCRIPT_DIR/run_integration_tests.sh"; then
    echo -e "${GREEN}✓ Integration tests passed${NC}"
    INTEGRATION_RESULT=0
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    INTEGRATION_RESULT=1
fi

# Run E2E tests
echo -e "\n${BLUE}=========================================${NC}"
echo -e "${BLUE}Running End-to-End Tests${NC}"
echo -e "${BLUE}=========================================${NC}\n"

if KEEP_RUNNING=1 bash "$SCRIPT_DIR/run_e2e_tests.sh"; then
    echo -e "${GREEN}✓ E2E tests passed${NC}"
    E2E_RESULT=0
else
    echo -e "${RED}✗ E2E tests failed${NC}"
    E2E_RESULT=1
fi

# Run benchmarks
echo -e "\n${BLUE}=========================================${NC}"
echo -e "${BLUE}Running Performance Benchmarks${NC}"
echo -e "${BLUE}=========================================${NC}\n"

if KEEP_RUNNING=1 bash "$SCRIPT_DIR/run_e2e_tests.sh" --benchmark; then
    echo -e "${GREEN}✓ Benchmarks completed${NC}"
    BENCHMARK_RESULT=0
else
    echo -e "${RED}✗ Benchmark failures${NC}"
    BENCHMARK_RESULT=1
fi

# Cleanup
echo -e "\n${YELLOW}Cleaning up test infrastructure...${NC}"
docker-compose -f "$SCRIPT_DIR/testutil/docker/docker-compose.test.yml" down -v
echo -e "${GREEN}Cleanup complete${NC}"

# Summary
echo -e "\n${BLUE}=========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}=========================================${NC}"

if [ $INTEGRATION_RESULT -eq 0 ]; then
    echo -e "Integration Tests: ${GREEN}PASSED${NC}"
else
    echo -e "Integration Tests: ${RED}FAILED${NC}"
fi

if [ $E2E_RESULT -eq 0 ]; then
    echo -e "E2E Tests: ${GREEN}PASSED${NC}"
else
    echo -e "E2E Tests: ${RED}FAILED${NC}"
fi

if [ $BENCHMARK_RESULT -eq 0 ]; then
    echo -e "Benchmarks: ${GREEN}PASSED${NC}"
else
    echo -e "Benchmarks: ${RED}FAILED${NC}"
fi

# Exit with failure if any test suite failed
if [ $INTEGRATION_RESULT -eq 0 ] && [ $E2E_RESULT -eq 0 ] && [ $BENCHMARK_RESULT -eq 0 ]; then
    echo -e "\n${GREEN}=========================================${NC}"
    echo -e "${GREEN}ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}=========================================${NC}\n"
    exit 0
else
    echo -e "\n${RED}=========================================${NC}"
    echo -e "${RED}SOME TESTS FAILED${NC}"
    echo -e "${RED}=========================================${NC}\n"
    exit 1
fi
