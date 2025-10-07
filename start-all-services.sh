#!/bin/bash
# Quick start all services (for services that don't need special env vars)

cd /home/mm/dev/b25

echo "Starting all B25 services..."
echo ""

# Configuration service (needs PostgreSQL)
echo "Starting configuration service..."
cd services/configuration
if [ -f "./bin/configuration" ]; then
    nohup ./bin/configuration > logs/configuration.log 2>&1 &
    echo "✓ configuration started (check logs/configuration.log)"
else
    echo "✗ configuration binary not found (run: make build)"
fi
cd ../..

# API Gateway
echo "Starting api-gateway..."
cd services/api-gateway  
if [ -f "./bin/api-gateway" ]; then
    nohup ./bin/api-gateway > logs/api-gateway.log 2>&1 &
    echo "✓ api-gateway started (check logs/api-gateway.log)"
else
    echo "✗ api-gateway binary not found"
fi
cd ../..

echo ""
echo "Waiting 5 seconds for services to start..."
sleep 5

echo ""
echo "Service status:"
./check-all-services.sh

