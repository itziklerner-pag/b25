#!/bin/bash

# Risk Manager gRPC Endpoint Tests

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

GRPC_PORT=50052

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Risk Manager gRPC Tests${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}grpcurl is not installed${NC}"
    echo "Install with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

# Test 1: List Services
echo -e "${YELLOW}Test 1: List Available Services${NC}"
grpcurl -plaintext localhost:$GRPC_PORT list
echo ""

# Test 2: Get Risk Metrics
echo -e "${YELLOW}Test 2: Get Risk Metrics${NC}"
grpcurl -plaintext -d '{"account_id": "test_account"}' \
    localhost:$GRPC_PORT \
    risk_manager.RiskManager/GetRiskMetrics
echo ""

# Test 3: Check Order Risk (Buy Order)
echo -e "${YELLOW}Test 3: Check Order Risk - Buy Order${NC}"
grpcurl -plaintext -d '{
    "account_id": "test_account",
    "order_id": "test_order_001",
    "symbol": "BTCUSDT",
    "side": "buy",
    "quantity": 0.1,
    "price": 50000,
    "order_type": "limit",
    "strategy_id": "test_strategy"
}' localhost:$GRPC_PORT risk_manager.RiskManager/CheckOrder
echo ""

# Test 4: Get Emergency Stop Status
echo -e "${YELLOW}Test 4: Get Emergency Stop Status${NC}"
grpcurl -plaintext -d '{}' \
    localhost:$GRPC_PORT \
    risk_manager.RiskManager/GetEmergencyStopStatus
echo ""

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}gRPC Tests Complete${NC}"
echo -e "${YELLOW}========================================${NC}"
