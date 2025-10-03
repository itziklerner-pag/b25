#!/bin/bash
# Stop B25 development environment

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

COMPOSE_FILE="docker/docker-compose.dev.yml"
REMOVE_VOLUMES=${REMOVE_VOLUMES:-false}

echo "================================================"
echo "Stopping B25 Development Environment"
echo "================================================"

if [ "$REMOVE_VOLUMES" = "true" ]; then
    echo -e "${YELLOW}Stopping services and removing volumes...${NC}"
    docker-compose -f $COMPOSE_FILE down -v
    echo -e "${RED}All data has been removed!${NC}"
else
    echo "Stopping services..."
    docker-compose -f $COMPOSE_FILE down
fi

echo ""
echo -e "${GREEN}Development environment stopped${NC}"
echo ""
echo "To remove all data, run: REMOVE_VOLUMES=true $0"
echo "To start again, run: ./scripts/dev-start.sh"
