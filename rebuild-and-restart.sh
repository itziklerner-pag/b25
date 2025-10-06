#!/bin/bash
echo "🔨 Rebuilding and Restarting All Services..."
cd /home/mm/dev/b25

# Stop all services
echo "Stopping services..."
./stop-all-services.sh 2>/dev/null || killall -9 service market-data-service node 2>/dev/null

sleep 3

# Rebuild Go services
echo "Rebuilding Go services..."
for service in order-execution strategy-engine account-monitor dashboard-server risk-manager configuration api-gateway; do
    echo "  Building $service..."
    cd services/$service
    go build -o bin/service ./cmd/server 2>&1 | grep -i error || echo "    ✅ $service"
    cd ../..
done

# Rebuild Rust services
echo "Rebuilding Rust service..."
cd services/market-data
cargo build --release 2>&1 | grep -E "(Compiling|Finished|error)" | tail -5
cd ../..

# Restart all
echo ""
echo "Starting all services..."
./run-all-services.sh

sleep 10

echo ""
echo "✅ Rebuild and restart complete!"
echo ""
echo "Checking services..."
ps aux | grep -E "(market-data|bin/service|node)" | grep -v grep | wc -l
echo "services running"
