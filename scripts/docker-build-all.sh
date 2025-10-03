#!/bin/bash
# Build all Docker images for B25 trading platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REGISTRY=${DOCKER_REGISTRY:-"ghcr.io/yourorg"}
VERSION=${VERSION:-$(git rev-parse --short HEAD 2>/dev/null || echo "dev")}
PLATFORM=${PLATFORM:-"linux/amd64"}  # linux/amd64,linux/arm64 for multi-arch
PUSH=${PUSH:-false}
PARALLEL=${PARALLEL:-true}

echo "================================================"
echo "Building B25 Docker Images"
echo "================================================"
echo "Registry: $REGISTRY"
echo "Version: $VERSION"
echo "Platform: $PLATFORM"
echo "Push: $PUSH"
echo "Parallel: $PARALLEL"
echo "================================================"

# Track build times
START_TIME=$(date +%s)

# Function to build Docker image
build_image() {
    local service=$1
    local context=$2
    local image_name="$REGISTRY/b25-$service:$VERSION"
    local latest_tag="$REGISTRY/b25-$service:latest"

    echo -e "${YELLOW}Building $service...${NC}"

    # Build args
    local build_args="--build-arg VERSION=$VERSION --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

    # Build command
    local build_cmd="docker buildx build --platform $PLATFORM $build_args -t $image_name -t $latest_tag"

    if [ "$PUSH" = "true" ]; then
        build_cmd="$build_cmd --push"
    else
        build_cmd="$build_cmd --load"
    fi

    build_cmd="$build_cmd $context"

    # Execute build
    if eval $build_cmd; then
        echo -e "${GREEN}✓ $service built successfully${NC}"
        echo "  Image: $image_name"
        return 0
    else
        echo -e "${RED}✗ Failed to build $service${NC}"
        return 1
    fi
}

# Services to build
declare -A SERVICES=(
    ["market-data"]="services/market-data"
    ["order-execution"]="services/order-execution"
    ["strategy-engine"]="services/strategy-engine"
    ["account-monitor"]="services/account-monitor"
    ["dashboard-server"]="services/dashboard-server"
    ["risk-manager"]="services/risk-manager"
    ["configuration"]="services/configuration"
    ["api-gateway"]="services/api-gateway"
    ["auth"]="services/auth"
    ["web-dashboard"]="ui/web"
)

# Setup Docker Buildx
if ! docker buildx inspect b25-builder &> /dev/null; then
    echo -e "${BLUE}Creating Docker buildx builder...${NC}"
    docker buildx create --name b25-builder --use
else
    echo -e "${BLUE}Using existing Docker buildx builder...${NC}"
    docker buildx use b25-builder
fi

# Build images
if [ "$PARALLEL" = "true" ]; then
    echo "Building images in parallel..."

    for service in "${!SERVICES[@]}"; do
        (build_image "$service" "${SERVICES[$service]}") &
    done

    # Wait for all builds
    wait

    # Check if any builds failed
    if [ $? -ne 0 ]; then
        echo -e "${RED}One or more builds failed${NC}"
        exit 1
    fi
else
    echo "Building images sequentially..."

    for service in "${!SERVICES[@]}"; do
        build_image "$service" "${SERVICES[$service]}"
    done
fi

# Calculate build time
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo "================================================"
echo -e "${GREEN}All images built successfully!${NC}"
echo "Total build time: ${DURATION}s"
echo "================================================"
echo ""
echo "Images tagged as:"
echo "  • $REGISTRY/b25-*:$VERSION"
echo "  • $REGISTRY/b25-*:latest"
echo ""

if [ "$PUSH" = "true" ]; then
    echo "Images pushed to registry: $REGISTRY"
else
    echo "To push images, run: PUSH=true $0"
fi

echo ""
echo "Next steps:"
echo "  • Start services: docker-compose -f docker/docker-compose.prod.yml up -d"
echo "  • View images: docker images | grep b25"
