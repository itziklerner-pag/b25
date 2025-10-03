#!/bin/bash
echo "🚀 Starting B25 Trading System..."
echo ""

cd /home/mm/dev/b25

# Start infrastructure first
echo "1️⃣ Starting infrastructure (Redis, PostgreSQL, TimescaleDB, NATS)..."
docker compose -f docker/docker-compose.dev.yml up -d redis postgres timescaledb nats

echo ""
echo "⏳ Waiting for infrastructure to be ready (30 seconds)..."
sleep 30

# Start observability
echo ""
echo "2️⃣ Starting observability (Prometheus, Grafana, Alertmanager)..."
docker compose -f docker/docker-compose.dev.yml up -d prometheus grafana alertmanager

echo ""
echo "⏳ Waiting for observability to be ready (10 seconds)..."
sleep 10

# Start all other services
echo ""
echo "3️⃣ Starting all trading services..."
docker compose -f docker/docker-compose.dev.yml up -d

echo ""
echo "✅ All services started!"
echo ""
echo "📊 Checking status..."
docker compose -f docker/docker-compose.dev.yml ps
echo ""
echo "🌐 Access points:"
echo "  - Web Dashboard: http://localhost:3000"
echo "  - Grafana: http://localhost:3001"
echo "  - Prometheus: http://localhost:9090"
echo ""
echo "📝 View logs: docker compose -f docker/docker-compose.dev.yml logs -f"
