#!/bin/bash
set -e  # Exit on any error

# Market Data Service - Deployment Script
# This script automates the complete deployment of the market-data service
# including building, configuration, systemd setup, and verification.

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="market-data"
SERVICE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_USER="${DEPLOY_USER:-$USER}"
BINARY_PATH="${SERVICE_DIR}/target/release/market-data-service"
CONFIG_PATH="${SERVICE_DIR}/config.yaml"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

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
print_header "Market Data Service Deployment"

echo ""
print_header "Step 1: Pre-deployment Checks"

# Check required commands
print_info "Checking required dependencies..."
MISSING_DEPS=0

if ! check_command cargo; then MISSING_DEPS=1; fi
if ! check_command docker; then MISSING_DEPS=1; fi
if ! check_command jq; then MISSING_DEPS=1; fi
if ! check_command curl; then MISSING_DEPS=1; fi

if [ $MISSING_DEPS -eq 1 ]; then
    print_error "Missing required dependencies. Please install them first."
    exit 1
fi

# Check Docker services
print_info "Checking Docker services..."
if ! docker ps | grep -q "b25-redis"; then
    print_warning "Redis container not running. Starting it..."
    if [ -f "${SERVICE_DIR}/../../docker-compose.simple.yml" ]; then
        docker-compose -f "${SERVICE_DIR}/../../docker-compose.simple.yml" up -d redis
        sleep 3
    else
        print_error "Redis is required but docker-compose.simple.yml not found"
        exit 1
    fi
fi
print_success "Redis is running"

# Check internet connectivity (needed for Binance)
print_info "Checking internet connectivity..."
if ! curl -s --max-time 5 https://fstream.binance.com > /dev/null; then
    print_error "Cannot reach Binance API. Check internet connection."
    exit 1
fi
print_success "Internet connectivity OK"

echo ""
print_header "Step 2: Build Service"

cd "$SERVICE_DIR"

# Stop existing service if running
print_info "Stopping existing service (if running)..."
if systemctl is-active --quiet ${SERVICE_NAME} 2>/dev/null; then
    sudo systemctl stop ${SERVICE_NAME}
    print_success "Stopped existing systemd service"
elif pgrep -f market-data-service > /dev/null; then
    pkill -f market-data-service
    sleep 2
    print_success "Stopped manual service instance"
fi

# Build release binary
print_info "Building release binary (this may take a few minutes)..."
cargo build --release

if [ ! -f "$BINARY_PATH" ]; then
    print_error "Build failed - binary not found at $BINARY_PATH"
    exit 1
fi

BINARY_SIZE=$(du -h "$BINARY_PATH" | cut -f1)
print_success "Build successful (binary size: $BINARY_SIZE)"

echo ""
print_header "Step 3: Configuration"

# Check if config.yaml exists, create from example if not
if [ ! -f "$CONFIG_PATH" ]; then
    if [ -f "${SERVICE_DIR}/config.example.yaml" ]; then
        print_info "Creating config.yaml from config.example.yaml..."
        cp "${SERVICE_DIR}/config.example.yaml" "$CONFIG_PATH"
        print_success "Created config.yaml"
    else
        print_error "No config.yaml or config.example.yaml found"
        exit 1
    fi
else
    print_success "config.yaml exists"
fi

# Validate config has required fields
print_info "Validating configuration..."
if ! grep -q "health_port:" "$CONFIG_PATH"; then
    print_error "config.yaml missing health_port field"
    exit 1
fi

if ! grep -q "exchange_rest_url:" "$CONFIG_PATH"; then
    print_warning "config.yaml missing exchange_rest_url field, adding it..."
    echo "" >> "$CONFIG_PATH"
    echo "# Exchange REST API URL (for snapshots)" >> "$CONFIG_PATH"
    echo 'exchange_rest_url: "https://fapi.binance.com"' >> "$CONFIG_PATH"
    print_success "Added exchange_rest_url to config"
fi

print_success "Configuration validated"

echo ""
print_header "Step 4: Systemd Service Setup"

# Create systemd service file
print_info "Creating systemd service file..."
cat > /tmp/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Market Data Service - Real-time market data ingestion and distribution
Documentation=https://github.com/your-org/b25
After=network.target docker.service redis.service
Requires=docker.service
Wants=redis.service

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
WorkingDirectory=${SERVICE_DIR}

# Main service
ExecStart=${BINARY_PATH}

# Restart policy
Restart=on-failure
RestartSec=5s
StartLimitIntervalSec=120s
StartLimitBurst=5

# Resource limits (prevent runaway processes)
CPUQuota=50%
MemoryLimit=512M
MemoryHigh=256M
TasksMax=100

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/tmp

# Environment
Environment="RUST_LOG=info"
Environment="RUST_BACKTRACE=1"

[Install]
WantedBy=multi-user.target
EOF

# Install systemd service
print_info "Installing systemd service..."
sudo cp /tmp/${SERVICE_NAME}.service "$SYSTEMD_SERVICE_PATH"
sudo systemctl daemon-reload
print_success "Systemd service installed"

# Enable service
print_info "Enabling service to start on boot..."
sudo systemctl enable ${SERVICE_NAME}
print_success "Service enabled"

echo ""
print_header "Step 5: Start Service"

print_info "Starting ${SERVICE_NAME} service..."
sudo systemctl start ${SERVICE_NAME}

# Wait for service to start
sleep 3

# Check if service started successfully
if systemctl is-active --quiet ${SERVICE_NAME}; then
    print_success "Service started successfully"
else
    print_error "Service failed to start"
    print_info "Checking logs..."
    sudo journalctl -u ${SERVICE_NAME} -n 20 --no-pager
    exit 1
fi

echo ""
print_header "Step 6: Verification"

# Check 1: Service Status
print_info "Checking service status..."
if systemctl is-active --quiet ${SERVICE_NAME}; then
    print_success "Service is active"
else
    print_error "Service is not active"
    exit 1
fi

# Check 2: Process Running
print_info "Checking process..."
if pgrep -f market-data-service > /dev/null; then
    PID=$(pgrep -f market-data-service)
    print_success "Process running (PID: $PID)"
else
    print_error "Process not found"
    exit 1
fi

# Check 3: Health Endpoint
print_info "Checking health endpoint..."
HEALTH_PORT=$(grep "health_port:" "$CONFIG_PATH" | awk '{print $2}')
sleep 2  # Give health server time to start

if curl -s --max-time 5 "http://localhost:${HEALTH_PORT}/health" | grep -q "healthy"; then
    print_success "Health endpoint responding on port ${HEALTH_PORT}"
else
    print_warning "Health endpoint not responding yet (may take a few more seconds)"
fi

# Check 4: Redis Data
print_info "Checking data flow to Redis (waiting 5 seconds for data)..."
sleep 5

if docker exec b25-redis redis-cli GET "market_data:BTCUSDT" | grep -q "last_price"; then
    PRICE=$(docker exec b25-redis redis-cli GET "market_data:BTCUSDT" | jq -r '.last_price')
    print_success "Data flowing to Redis (BTC: \$$PRICE)"
else
    print_warning "No data in Redis yet (may take up to 30 seconds for first data)"
fi

# Check 5: Resource Usage
print_info "Checking resource usage..."
CPU=$(ps aux | grep market-data-service | grep -v grep | awk '{print $3}')
MEM=$(ps aux | grep market-data-service | grep -v grep | awk '{print $4}')
print_success "CPU: ${CPU}% | Memory: ${MEM}%"

echo ""
print_header "Deployment Summary"

echo ""
echo -e "${GREEN}✓ Deployment successful!${NC}"
echo ""
echo "Service Details:"
echo "  • Name: ${SERVICE_NAME}"
echo "  • Binary: ${BINARY_PATH}"
echo "  • Config: ${CONFIG_PATH}"
echo "  • Systemd: ${SYSTEMD_SERVICE_PATH}"
echo "  • User: ${SERVICE_USER}"
echo ""
echo "Management Commands:"
echo "  • Status: sudo systemctl status ${SERVICE_NAME}"
echo "  • Logs: sudo journalctl -u ${SERVICE_NAME} -f"
echo "  • Restart: sudo systemctl restart ${SERVICE_NAME}"
echo "  • Stop: sudo systemctl stop ${SERVICE_NAME}"
echo ""
echo "Health Check:"
echo "  • curl http://localhost:${HEALTH_PORT}/health"
echo ""
echo "Next Steps:"
echo "  1. Monitor logs for a few minutes: sudo journalctl -u ${SERVICE_NAME} -f"
echo "  2. Verify data in Redis: docker exec b25-redis redis-cli GET market_data:BTCUSDT"
echo "  3. Set up monitoring/alerting (Prometheus/Grafana)"
echo ""

# Save deployment info
DEPLOY_INFO_FILE="${SERVICE_DIR}/deployment-info.txt"
cat > "$DEPLOY_INFO_FILE" <<EOF
Deployment Information
======================
Date: $(date)
User: ${SERVICE_USER}
Host: $(hostname)
Service: ${SERVICE_NAME}
Binary: ${BINARY_PATH}
Binary Size: ${BINARY_SIZE}
Config: ${CONFIG_PATH}
Systemd: ${SYSTEMD_SERVICE_PATH}
PID: ${PID}
CPU: ${CPU}%
Memory: ${MEM}%
Health Port: ${HEALTH_PORT}

Service Status:
$(systemctl status ${SERVICE_NAME} --no-pager || true)
EOF

print_success "Deployment info saved to ${DEPLOY_INFO_FILE}"

echo ""
print_success "Deployment complete!"
exit 0
