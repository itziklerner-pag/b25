#!/bin/bash
# Deploy B25 to production

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
COMPOSE_FILE="docker/docker-compose.prod.yml"
VERSION=${VERSION:-$(git rev-parse --short HEAD 2>/dev/null || echo "latest")}
ENV_FILE=${ENV_FILE:-.env.production}

echo "================================================"
echo "Deploying B25 Trading Platform to Production"
echo "================================================"
echo "Version: $VERSION"
echo "Env File: $ENV_FILE"
echo "================================================"

# Pre-deployment checks
echo -e "${BLUE}Running pre-deployment checks...${NC}"

# Check if .env.production exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: $ENV_FILE not found${NC}"
    echo "Please create $ENV_FILE with production configuration"
    exit 1
fi

# Check required environment variables
REQUIRED_VARS=(
    "POSTGRES_PASSWORD"
    "TIMESCALEDB_PASSWORD"
    "REDIS_PASSWORD"
    "JWT_SECRET"
)

source "$ENV_FILE"

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo -e "${RED}Error: Required variable $var is not set in $ENV_FILE${NC}"
        exit 1
    fi
done

echo -e "${GREEN}✓ Environment checks passed${NC}"

# Pull latest images
echo -e "${BLUE}Pulling latest images...${NC}"
docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE pull

# Create backup of current deployment
echo -e "${BLUE}Creating backup of current deployment...${NC}"
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup volumes
docker-compose -f $COMPOSE_FILE ps -q | while read container_id; do
    if [ ! -z "$container_id" ]; then
        container_name=$(docker inspect --format='{{.Name}}' $container_id | sed 's/\///')
        echo "Backing up $container_name..."
        docker commit $container_id "$BACKUP_DIR/$container_name"
    fi
done

echo -e "${GREEN}✓ Backup created at $BACKUP_DIR${NC}"

# Deploy
echo -e "${BLUE}Deploying services...${NC}"

# Start infrastructure first
echo "Starting infrastructure services..."
docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d \
    redis postgres timescaledb nats

# Wait for infrastructure to be healthy
echo "Waiting for infrastructure to be ready..."
sleep 30

# Start observability services
echo "Starting observability services..."
docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d \
    prometheus grafana alertmanager

# Start core trading services
echo "Starting core trading services..."
docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d \
    configuration \
    market-data \
    order-execution \
    strategy-engine \
    risk-manager \
    account-monitor \
    dashboard-server

# Wait for core services
echo "Waiting for core services to be ready..."
sleep 20

# Start API and UI services
echo "Starting API and UI services..."
docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d \
    api-gateway \
    auth \
    web-dashboard \
    nginx

echo ""
echo "================================================"
echo -e "${GREEN}Deployment complete!${NC}"
echo "================================================"

# Health checks
echo ""
echo "Running health checks..."
sleep 10

SERVICES=(
    "http://localhost:8000/health|API Gateway"
    "http://localhost:8001/health|Auth Service"
    "http://localhost:8080/health|Market Data"
    "http://localhost:8081/health|Order Execution"
    "http://localhost:8082/health|Strategy Engine"
    "http://localhost:8083/health|Risk Manager"
    "http://localhost:8084/health|Account Monitor"
    "http://localhost:8085/health|Configuration"
    "http://localhost:8086/health|Dashboard Server"
)

FAILED=0
for service in "${SERVICES[@]}"; do
    IFS='|' read -r url name <<< "$service"
    if curl -f -s "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $name${NC}"
    else
        echo -e "${RED}✗ $name${NC}"
        FAILED=$((FAILED + 1))
    fi
done

echo ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All health checks passed!${NC}"
else
    echo -e "${YELLOW}Warning: $FAILED service(s) failed health check${NC}"
    echo "Check logs with: docker-compose -f $COMPOSE_FILE logs"
fi

echo ""
echo "Deployment Summary:"
echo "  • Version:        $VERSION"
echo "  • Backup:         $BACKUP_DIR"
echo "  • Services:       $(docker-compose -f $COMPOSE_FILE ps --services | wc -l)"
echo ""
echo "To rollback, run: ./scripts/rollback.sh $BACKUP_DIR"
echo "To view logs, run: docker-compose -f $COMPOSE_FILE logs -f"
