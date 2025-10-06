#!/bin/bash
set -e  # Exit on any error

# Configuration Service - Deployment Script
# This script automates the complete deployment of the configuration service
# including building, migrations, systemd setup, and verification.

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="configuration"
SERVICE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_USER="${DEPLOY_USER:-$USER}"
BINARY_PATH="${SERVICE_DIR}/bin/configuration-service"
CONFIG_PATH="${SERVICE_DIR}/config.yaml"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"
SERVICE_PORT=8085

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 is not installed"
        return 1
    fi
    print_success "$1 is installed"
    return 0
}

# Main deployment steps
print_header "Configuration Service Deployment"

echo ""
print_header "Step 1: Pre-deployment Checks"

# Check required commands
print_info "Checking required dependencies..."
MISSING_DEPS=0

if ! check_command go; then MISSING_DEPS=1; fi
if ! check_command docker; then MISSING_DEPS=1; fi
if ! check_command jq; then MISSING_DEPS=1; fi
if ! check_command curl; then MISSING_DEPS=1; fi

if [ $MISSING_DEPS -eq 1 ]; then
    print_error "Missing required dependencies. Please install them first."
    exit 1
fi

# Check Docker services
print_info "Checking Docker services..."

# Check PostgreSQL
if ! docker ps | grep -q "b25-postgres"; then
    print_warning "PostgreSQL container not running. Starting it..."
    if [ -f "${SERVICE_DIR}/../../docker-compose.simple.yml" ]; then
        docker-compose -f "${SERVICE_DIR}/../../docker-compose.simple.yml" up -d postgres
        sleep 5
    else
        print_error "PostgreSQL is required but docker-compose.simple.yml not found"
        exit 1
    fi
fi
print_success "PostgreSQL is running"

# Check NATS
if ! docker ps | grep -q "b25-nats"; then
    print_warning "NATS container not running. Starting it..."
    if [ -f "${SERVICE_DIR}/../../docker-compose.simple.yml" ]; then
        docker-compose -f "${SERVICE_DIR}/../../docker-compose.simple.yml" up -d nats
        sleep 3
    else
        print_error "NATS is required but docker-compose.simple.yml not found"
        exit 1
    fi
fi
print_success "NATS is running"

# Verify database connectivity
print_info "Checking database connectivity..."
if docker exec b25-postgres psql -U b25 -d b25_config -c "SELECT 1" &> /dev/null; then
    print_success "Database connection OK"
else
    print_error "Cannot connect to database"
    exit 1
fi

echo ""
print_header "Step 2: Build Service"

cd "$SERVICE_DIR"

# Stop existing service if running
print_info "Stopping existing service..."
if systemctl is-active --quiet ${SERVICE_NAME}.service 2>/dev/null; then
    sudo systemctl stop ${SERVICE_NAME}.service
    print_success "Service stopped"
else
    print_info "Service was not running"
fi

# Build the service
print_info "Building Go binary..."
make build
if [ ! -f "$BINARY_PATH" ]; then
    print_error "Build failed - binary not found at $BINARY_PATH"
    exit 1
fi
print_success "Build completed successfully"

echo ""
print_header "Step 3: Database Migrations"

print_info "Running database migrations..."
if docker exec -i b25-postgres psql -U b25 -d b25_config < migrations/000001_init_schema.up.sql &> /dev/null; then
    print_success "Migrations completed (or already applied)"
else
    print_warning "Migration may have already been applied"
fi

# Verify tables exist
print_info "Verifying database schema..."
TABLE_COUNT=$(docker exec b25-postgres psql -U b25 -d b25_config -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('configurations', 'configuration_versions', 'audit_logs')" | tr -d ' ')
if [ "$TABLE_COUNT" -eq 3 ]; then
    print_success "All tables exist"
else
    print_error "Expected 3 tables, found $TABLE_COUNT"
    exit 1
fi

echo ""
print_header "Step 4: Configuration"

# Check if config.yaml exists
if [ ! -f "$CONFIG_PATH" ]; then
    print_warning "config.yaml not found, using config.example.yaml"
    cp config.example.yaml config.yaml
fi
print_success "Configuration file ready"

# Set API key if not already set
if [ -z "$CONFIG_API_KEY" ]; then
    print_warning "CONFIG_API_KEY not set. Service will run without authentication."
    print_info "To enable authentication, set: export CONFIG_API_KEY=your-secret-key"
else
    print_success "API key configured"
fi

echo ""
print_header "Step 5: Systemd Service Setup"

print_info "Creating systemd service file..."
sudo tee $SYSTEMD_SERVICE_PATH > /dev/null << EOF
[Unit]
Description=Configuration Service - Centralized configuration management
Documentation=https://github.com/b25/services/configuration
After=network.target docker.service
Requires=docker.service
PartOf=b25.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$SERVICE_DIR
ExecStart=$BINARY_PATH
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# Environment
Environment="GIN_MODE=release"
$([ -n "$CONFIG_API_KEY" ] && echo "Environment=\"CONFIG_API_KEY=$CONFIG_API_KEY\"")

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$SERVICE_DIR
ReadOnlyPaths=/etc

# Resource limits
LimitNOFILE=65536
MemoryLimit=512M
CPUQuota=200%

[Install]
WantedBy=multi-user.target
WantedBy=b25.target
EOF

print_success "Systemd service file created"

print_info "Reloading systemd daemon..."
sudo systemctl daemon-reload
print_success "Systemd daemon reloaded"

echo ""
print_header "Step 6: Start Service"

print_info "Starting service..."
sudo systemctl start ${SERVICE_NAME}.service
sleep 2

if systemctl is-active --quiet ${SERVICE_NAME}.service; then
    print_success "Service started successfully"
else
    print_error "Service failed to start"
    print_info "Check logs with: journalctl -u ${SERVICE_NAME}.service -n 50"
    exit 1
fi

print_info "Enabling service to start on boot..."
sudo systemctl enable ${SERVICE_NAME}.service
print_success "Service enabled"

echo ""
print_header "Step 7: Verification"

# Wait for service to be ready
print_info "Waiting for service to be ready..."
MAX_RETRIES=10
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    sleep 1
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    print_error "Service did not become ready in time"
    exit 1
fi

# Test health endpoint
print_info "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:${SERVICE_PORT}/health)
if echo "$HEALTH_RESPONSE" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
    print_success "Health check passed"
else
    print_error "Health check failed: $HEALTH_RESPONSE"
    exit 1
fi

# Test readiness endpoint
print_info "Testing readiness endpoint..."
READY_RESPONSE=$(curl -s http://localhost:${SERVICE_PORT}/ready)
if echo "$READY_RESPONSE" | jq -e '.status == "ready"' > /dev/null 2>&1; then
    print_success "Readiness check passed"

    # Show check details
    DB_STATUS=$(echo "$READY_RESPONSE" | jq -r '.checks.database')
    NATS_STATUS=$(echo "$READY_RESPONSE" | jq -r '.checks.nats')
    print_info "  Database: $DB_STATUS"
    print_info "  NATS: $NATS_STATUS"
else
    print_error "Readiness check failed: $READY_RESPONSE"
    exit 1
fi

# Test API endpoint
print_info "Testing API endpoint..."
API_RESPONSE=$(curl -s http://localhost:${SERVICE_PORT}/api/v1/configurations)
if echo "$API_RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    CONFIG_COUNT=$(echo "$API_RESPONSE" | jq -r '.total')
    print_success "API endpoint working (found $CONFIG_COUNT configurations)"
else
    print_error "API endpoint failed: $API_RESPONSE"
    exit 1
fi

# Show service status
echo ""
print_info "Service status:"
sudo systemctl status ${SERVICE_NAME}.service --no-pager -l

echo ""
print_header "Deployment Complete!"
print_success "Configuration service is running on port $SERVICE_PORT"
print_info ""
print_info "Useful commands:"
print_info "  Check status:  sudo systemctl status ${SERVICE_NAME}.service"
print_info "  View logs:     journalctl -u ${SERVICE_NAME}.service -f"
print_info "  Restart:       sudo systemctl restart ${SERVICE_NAME}.service"
print_info "  Stop:          sudo systemctl stop ${SERVICE_NAME}.service"
print_info "  Uninstall:     ./uninstall.sh"
print_info ""
print_info "API Endpoints:"
print_info "  Health:        http://localhost:${SERVICE_PORT}/health"
print_info "  Ready:         http://localhost:${SERVICE_PORT}/ready"
print_info "  Metrics:       http://localhost:${SERVICE_PORT}/metrics"
print_info "  API:           http://localhost:${SERVICE_PORT}/api/v1/configurations"
print_info ""
if [ -z "$CONFIG_API_KEY" ]; then
    print_warning "Authentication is DISABLED. Set CONFIG_API_KEY to enable it."
else
    print_info "Authentication is ENABLED. Use header: X-API-Key: <key>"
fi
