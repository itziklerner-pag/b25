#!/bin/bash

# Risk Manager Health Check Test Script

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

HTTP_PORT=8083
GRPC_PORT=50052

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Risk Manager Health Check Tests${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Test 1: HTTP Health Endpoint
echo -e "${YELLOW}Test 1: HTTP Health Endpoint${NC}"
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/health)
if [ "$HEALTH_RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ HTTP health endpoint is responding${NC}"
else
    echo -e "${RED}✗ HTTP health endpoint failed (HTTP $HEALTH_RESPONSE)${NC}"
fi
echo ""

# Test 2: Metrics Endpoint
echo -e "${YELLOW}Test 2: Prometheus Metrics Endpoint${NC}"
METRICS_RESPONSE=$(curl -s http://localhost:$HTTP_PORT/metrics | head -n 5)
if [ -n "$METRICS_RESPONSE" ]; then
    echo -e "${GREEN}✓ Metrics endpoint is responding${NC}"
    echo "Sample metrics:"
    curl -s http://localhost:$HTTP_PORT/metrics | grep "risk_" | head -n 5
else
    echo -e "${RED}✗ Metrics endpoint failed${NC}"
fi
echo ""

# Test 3: gRPC Health Check (requires grpcurl)
echo -e "${YELLOW}Test 3: gRPC Health Check${NC}"
if command -v grpcurl &> /dev/null; then
    GRPC_HEALTH=$(grpcurl -plaintext localhost:$GRPC_PORT grpc.health.v1.Health/Check 2>&1)
    if echo "$GRPC_HEALTH" | grep -q "SERVING"; then
        echo -e "${GREEN}✓ gRPC health check passed${NC}"
    else
        echo -e "${RED}✗ gRPC health check failed${NC}"
        echo "$GRPC_HEALTH"
    fi
else
    echo -e "${YELLOW}⚠ grpcurl not installed, skipping gRPC health check${NC}"
    echo "  Install with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi
echo ""

# Test 4: Check Service Logs
echo -e "${YELLOW}Test 4: Recent Service Logs${NC}"
if systemctl is-active --quiet b25-risk-manager; then
    echo -e "${GREEN}✓ Service is running${NC}"
    echo "Recent logs:"
    journalctl -u b25-risk-manager -n 10 --no-pager
else
    echo -e "${RED}✗ Service is not running${NC}"
fi
echo ""

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Health Check Complete${NC}"
echo -e "${YELLOW}========================================${NC}"
