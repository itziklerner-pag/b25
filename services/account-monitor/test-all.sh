#!/bin/bash

# Account Monitor - Complete Test Suite
# Runs all test scripts

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================"
echo "Account Monitor - Complete Test Suite"
echo "========================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
SUITE_PASSED=0
SUITE_FAILED=0

# Run health tests
echo ""
echo "========== Running Health Tests =========="
if bash "${SCRIPT_DIR}/test-health.sh"; then
    SUITE_PASSED=$((SUITE_PASSED + 1))
else
    SUITE_FAILED=$((SUITE_FAILED + 1))
fi

# Run API tests
echo ""
echo "========== Running API Tests =========="
if bash "${SCRIPT_DIR}/test-api.sh"; then
    SUITE_PASSED=$((SUITE_PASSED + 1))
else
    SUITE_FAILED=$((SUITE_FAILED + 1))
fi

# Run fill event tests
echo ""
echo "========== Running Fill Event Tests =========="
if bash "${SCRIPT_DIR}/test-fill-events.sh"; then
    SUITE_PASSED=$((SUITE_PASSED + 1))
else
    SUITE_FAILED=$((SUITE_FAILED + 1))
fi

# Final summary
echo ""
echo "========================================"
echo "Complete Test Suite Summary"
echo "========================================"
echo -e "Test Suites Passed: ${GREEN}${SUITE_PASSED}${NC}"
echo -e "Test Suites Failed: ${RED}${SUITE_FAILED}${NC}"
echo "Total Test Suites:  $((SUITE_PASSED + SUITE_FAILED))"
echo ""

if [ $SUITE_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All test suites passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some test suites failed!${NC}"
    exit 1
fi
