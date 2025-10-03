#!/bin/bash
# Start B25 development environment

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
COMPOSE_FILE="docker/docker-compose.dev.yml"
SERVICES=${1:-all}  # Can specify specific services or "all"

echo "================================================"
echo "Starting B25 Development Environment"
echo "================================================"

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Warning: .env file not found${NC}"
    echo "Copying .env.example to .env..."
    cp .env.example .env
    echo -e "${YELLOW}Please update .env with your configuration${NC}"
fi

# Start services
if [ "$SERVICES" = "all" ]; then
    echo -e "${BLUE}Starting all services...${NC}"
    docker-compose -f $COMPOSE_FILE up -d
else
    echo -e "${BLUE}Starting services: $SERVICES${NC}"
    docker-compose -f $COMPOSE_FILE up -d $SERVICES
fi

echo ""
echo "================================================"
echo -e "${GREEN}Development environment started!${NC}"
echo "================================================"
echo ""
echo "Service URLs:"
echo "  • API Gateway:        http://localhost:8000"
echo "  • Auth Service:       http://localhost:8001"
echo "  • Market Data:        http://localhost:8080"
echo "  • Order Execution:    http://localhost:8081"
echo "  • Strategy Engine:    http://localhost:8082"
echo "  • Risk Manager:       http://localhost:8083"
echo "  • Account Monitor:    http://localhost:8084"
echo "  • Configuration:      http://localhost:8085"
echo "  • Dashboard Server:   http://localhost:8086"
echo "  • Web Dashboard:      http://localhost:3000"
echo ""
echo "Monitoring:"
echo "  • Grafana:            http://localhost:3001 (admin/admin)"
echo "  • Prometheus:         http://localhost:9090"
echo "  • Alertmanager:       http://localhost:9093"
echo "  • NATS Monitor:       http://localhost:8222"
echo ""
echo "Infrastructure:"
echo "  • Redis:              localhost:6379"
echo "  • PostgreSQL:         localhost:5432"
echo "  • TimescaleDB:        localhost:5433"
echo "  • NATS:               localhost:4222"
echo ""
echo "Commands:"
echo "  • View logs:          docker-compose -f $COMPOSE_FILE logs -f [service]"
echo "  • Stop services:      ./scripts/dev-stop.sh"
echo "  • Restart service:    docker-compose -f $COMPOSE_FILE restart [service]"
echo "  • Shell into service: docker exec -it b25-[service] sh"
echo ""
