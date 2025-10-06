#!/bin/bash

# Quick Test Script for Local Development
# Tests the service without external dependencies (PostgreSQL, Redis, Kafka)

set -e

API_URL="${API_URL:-http://localhost:9097}"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Quick Analytics Service Test${NC}"
echo "======================================"

# Check if service is running
echo -e "\n${YELLOW}1. Checking if service is running...${NC}"
if curl -s "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Service is running${NC}"
else
    echo -e "${RED}✗ Service is not running at $API_URL${NC}"
    echo "Start the service with: go run ./cmd/server/main.go -config config.yaml"
    exit 1
fi

# Health check
echo -e "\n${YELLOW}2. Health Check...${NC}"
HEALTH=$(curl -s "$API_URL/health")
echo "$HEALTH" | jq '.' || echo "$HEALTH"

# Prometheus metrics
echo -e "\n${YELLOW}3. Prometheus Metrics...${NC}"
METRICS=$(curl -s "${METRICS_URL:-http://localhost:9098}/metrics")
if echo "$METRICS" | grep -q "analytics_"; then
    echo -e "${GREEN}✓ Metrics available${NC}"
    echo "$METRICS" | grep "analytics_" | head -5
else
    echo -e "${RED}✗ No metrics found${NC}"
fi

# Test event tracking
echo -e "\n${YELLOW}4. Test Event Tracking...${NC}"
RESPONSE=$(curl -s -X POST "$API_URL/api/v1/events" \
    -H "Content-Type: application/json" \
    -d '{
        "event_type": "test.event",
        "user_id": "test-user",
        "properties": {"test": true}
    }')
echo "$RESPONSE" | jq '.' || echo "$RESPONSE"

# Get events
echo -e "\n${YELLOW}5. Query Events...${NC}"
EVENTS=$(curl -s "$API_URL/api/v1/events?limit=5")
echo "$EVENTS" | jq '.' || echo "$EVENTS"

# Ingestion metrics
echo -e "\n${YELLOW}6. Ingestion Metrics...${NC}"
INGESTION=$(curl -s "$API_URL/api/v1/internal/ingestion-metrics")
echo "$INGESTION" | jq '.' || echo "$INGESTION"

echo -e "\n${GREEN}======================================"
echo "Quick test completed!"
echo "======================================${NC}"
