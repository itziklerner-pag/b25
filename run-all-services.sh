#!/bin/bash
# Run all B25 trading services

cd /home/mm/dev/b25

# Load environment
export $(cat .env | grep -v '^#' | xargs)

# Create log directory
mkdir -p logs

echo "🚀 Starting all B25 trading services..."
echo ""

# Start Market Data Service
echo "Starting Market Data Service on port 8080..."
cd /home/mm/dev/b25/services/market-data
RUST_LOG=info ./target/debug/market-data-service > ../../logs/market-data.log 2>&1 &
echo $! > ../../logs/market-data.pid
cd /home/mm/dev/b25

# Start Order Execution Service
echo "Starting Order Execution Service on port 8081..."
cd /home/mm/dev/b25/services/order-execution
./bin/service > ../../logs/order-execution.log 2>&1 &
echo $! > ../../logs/order-execution.pid
cd /home/mm/dev/b25

# Start Strategy Engine Service
echo "Starting Strategy Engine Service on port 8082..."
cd /home/mm/dev/b25/services/strategy-engine
./bin/service > ../../logs/strategy-engine.log 2>&1 &
echo $! > ../../logs/strategy-engine.pid
cd /home/mm/dev/b25

# Start Risk Manager Service
echo "Starting Risk Manager Service on port 8083..."
cd /home/mm/dev/b25/services/risk-manager
./bin/service > ../../logs/risk-manager.log 2>&1 &
echo $! > ../../logs/risk-manager.pid
cd /home/mm/dev/b25

# Start Account Monitor Service
echo "Starting Account Monitor Service on port 8084..."
cd /home/mm/dev/b25/services/account-monitor
./bin/service > ../../logs/account-monitor.log 2>&1 &
echo $! > ../../logs/account-monitor.pid
cd /home/mm/dev/b25

# Start Configuration Service
echo "Starting Configuration Service on port 8085..."
cd /home/mm/dev/b25/services/configuration
./bin/service > ../../logs/configuration.log 2>&1 &
echo $! > ../../logs/configuration.pid
cd /home/mm/dev/b25

# Start Dashboard Server Service
echo "Starting Dashboard Server Service on port 8086..."
cd /home/mm/dev/b25/services/dashboard-server
DASHBOARD_PORT=8086 DASHBOARD_REDIS_URL=localhost:6379 ./bin/service > ../../logs/dashboard-server.log 2>&1 &
echo $! > ../../logs/dashboard-server.pid
cd /home/mm/dev/b25

# Start API Gateway Service
echo "Starting API Gateway Service on port 8000..."
cd /home/mm/dev/b25/services/api-gateway
./bin/service > ../../logs/api-gateway.log 2>&1 &
echo $! > ../../logs/api-gateway.pid
cd /home/mm/dev/b25

# Start Auth Service
echo "Starting Auth Service on port 8001..."
cd /home/mm/dev/b25/services/auth
npm start > ../../logs/auth.log 2>&1 &
echo $! > ../../logs/auth.pid
cd /home/mm/dev/b25

echo ""
echo "✅ All services started!"
echo ""
echo "📊 Check status:"
echo "  ps aux | grep service"
echo ""
echo "📝 View logs:"
echo "  tail -f logs/*.log"
echo ""
echo "🛑 Stop all services:"
echo "  ./stop-all-services.sh"
