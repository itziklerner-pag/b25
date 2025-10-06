#!/bin/bash

# Risk Manager Integration Tests
# Tests the service with real dependencies

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Risk Manager Integration Tests${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

FAILED=0

# Test 1: PostgreSQL Connection
echo -e "${YELLOW}Test 1: PostgreSQL Connection${NC}"
if journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "database connection established"; then
    echo -e "${GREEN}✓ PostgreSQL connection successful${NC}"
else
    echo -e "${RED}✗ PostgreSQL connection failed${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

# Test 2: Redis Connection
echo -e "${YELLOW}Test 2: Redis Connection${NC}"
if journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "Redis connection established"; then
    echo -e "${GREEN}✓ Redis connection successful${NC}"
else
    echo -e "${RED}✗ Redis connection failed${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

# Test 3: NATS Connection
echo -e "${YELLOW}Test 3: NATS Connection${NC}"
if journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "NATS connection established"; then
    echo -e "${GREEN}✓ NATS connection successful${NC}"
else
    echo -e "${RED}✗ NATS connection failed${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

# Test 4: Account Monitor Connection
echo -e "${YELLOW}Test 4: Account Monitor Connection${NC}"
if journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "Account Monitor client connected"; then
    echo -e "${GREEN}✓ Account Monitor connected${NC}"
elif journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "using mock data"; then
    echo -e "${YELLOW}⚠ Using mock data (Account Monitor not available)${NC}"
else
    echo -e "${RED}✗ Account Monitor status unknown${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

# Test 5: Policy Loading
echo -e "${YELLOW}Test 5: Policy Loading${NC}"
if journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "policies loaded"; then
    echo -e "${GREEN}✓ Policies loaded from database${NC}"
else
    echo -e "${YELLOW}⚠ Using default policies${NC}"
fi
echo ""

# Test 6: Risk Monitor Running
echo -e "${YELLOW}Test 6: Risk Monitor Status${NC}"
if journalctl -u b25-risk-manager -n 100 --no-pager | grep -q "risk monitor started"; then
    echo -e "${GREEN}✓ Risk monitor is running${NC}"
else
    echo -e "${RED}✗ Risk monitor not started${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

# Test 7: Check for Errors
echo -e "${YELLOW}Test 7: Error Check${NC}"
ERROR_COUNT=$(journalctl -u b25-risk-manager -n 100 --no-pager | grep -i "error" | grep -v "error\": 0" | wc -l)
if [ "$ERROR_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✓ No errors in recent logs${NC}"
else
    echo -e "${RED}✗ Found $ERROR_COUNT errors in recent logs${NC}"
    echo "Recent errors:"
    journalctl -u b25-risk-manager -n 100 --no-pager | grep -i "error" | tail -n 5
    FAILED=$((FAILED + 1))
fi
echo ""

# Summary
echo -e "${YELLOW}========================================${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All Integration Tests Passed!${NC}"
else
    echo -e "${RED}$FAILED Integration Test(s) Failed${NC}"
fi
echo -e "${YELLOW}========================================${NC}"

exit $FAILED
