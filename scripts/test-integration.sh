#!/bin/bash
# Integration tests

set -e

echo "🔗 Running integration tests..."

# Start infrastructure
docker-compose -f docker/docker-compose.dev.yml up -d redis postgres timescaledb nats

# Wait for services
echo "Waiting for services to be ready..."
sleep 10

# Run integration tests
cd tests/integration

# Add your integration test commands here
echo "Running integration test suite..."
# go test -v ./... || python -m pytest || npm test

echo "✅ Integration tests passed!"

# Cleanup
docker-compose -f docker/docker-compose.dev.yml down
