#!/bin/bash
# Test all services

set -e

echo "🧪 Testing all services..."

# Rust services
echo "Testing market-data..."
cd services/market-data && cargo test && cd ../..

echo "Testing terminal-ui..."
cd ui/terminal && cargo test && cd ../..

# Go services
for service in order-execution strategy-engine account-monitor dashboard-server risk-manager configuration; do
    echo "Testing $service..."
    cd services/$service && go test ./... && cd ../..
done

# Web dashboard
echo "Testing web-dashboard..."
cd ui/web && npm test && cd ../..

echo "✅ All tests passed!"
