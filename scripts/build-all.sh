#!/bin/bash
# Build all services

set -e

echo "🔨 Building all services..."

# Rust services
echo "Building market-data..."
cd services/market-data && cargo build --release && cd ../..

echo "Building terminal-ui..."
cd ui/terminal && cargo build --release && cd ../..

# Go services
for service in order-execution strategy-engine account-monitor dashboard-server risk-manager configuration; do
    echo "Building $service..."
    cd services/$service && go build -o bin/$service ./cmd/server && cd ../..
done

# Web dashboard
echo "Building web-dashboard..."
cd ui/web && npm run build && cd ../..

echo "✅ All services built!"
