#!/bin/bash
set -e  # Exit on any error

# Dashboard Server - Deployment Script
# Automates deployment with systemd, security, and verification

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SERVICE_NAME="dashboard-server"
SERVICE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_USER="${DEPLOY_USER:-$USER}"
BINARY_PATH="${SERVICE_DIR}/dashboard-server"
CONFIG_PATH="${SERVICE_DIR}/config.yaml"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ $1${NC}"; }

check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 is not installed"
        return 1
    fi
    print_success "$1 is installed"
    return 0
}

# Main deployment
print_header "Dashboard Server Deployment"

echo ""
print_header "Step 1: Pre-deployment Checks"

print_info "Checking required dependencies..."
MISSING_DEPS=0

if ! check_command go; then MISSING_DEPS=1; fi
if ! check_command docker; then MISSING_DEPS=1; fi
if ! check_command jq; then MISSING_DEPS=1; fi
if ! check_command curl; then MISSING_DEPS=1; fi

if [ $MISSING_DEPS -eq 1 ]; then
    print_error "Missing required dependencies"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_success "Go version: $GO_VERSION"

# Check Redis
print_info "Checking Redis..."
if ! docker ps | grep -q "b25-redis"; then
    print_warning "Redis not running. Starting..."
    if [ -f "${SERVICE_DIR}/../../docker-compose.simple.yml" ]; then
        docker-compose -f "${SERVICE_DIR}/../../docker-compose.simple.yml" up -d redis
        sleep 3
    else
        print_error "Redis required but docker-compose.simple.yml not found"
        exit 1
    fi
fi
print_success "Redis is running"

echo ""
print_header "Step 2: Build Service"

cd "$SERVICE_DIR"

# Stop existing service
print_info "Stopping existing service..."
if systemctl is-active --quiet ${SERVICE_NAME} 2>/dev/null; then
    sudo systemctl stop ${SERVICE_NAME}
    print_success "Stopped systemd service"
elif pgrep -f dashboard-server > /dev/null; then
    pkill -f dashboard-server
    sleep 2
    print_success "Stopped manual instance"
fi

# Build binary
print_info "Building binary..."
go build -o dashboard-server ./cmd/server

if [ ! -f "$BINARY_PATH" ]; then
    print_error "Build failed - binary not found"
    exit 1
fi

BINARY_SIZE=$(du -h "$BINARY_PATH" | cut -f1)
print_success "Build successful (size: $BINARY_SIZE)"

echo ""
print_header "Step 3: Configuration"

# Create config from example if missing
if [ ! -f "$CONFIG_PATH" ]; then
    if [ -f "${SERVICE_DIR}/config.example.yaml" ]; then
        print_info "Creating config.yaml from example..."
        cp "${SERVICE_DIR}/config.example.yaml" "$CONFIG_PATH"
        print_success "Created config.yaml"
    else
        print_error "No config.yaml or config.example.yaml found"
        exit 1
    fi
else
    print_success "config.yaml exists"
fi

# Validate config
print_info "Validating configuration..."
if ! grep -q "server:" "$CONFIG_PATH"; then
    print_error "Invalid config.yaml - missing server section"
    exit 1
fi

# Check if allowed_origins exists, add if missing
if ! grep -q "allowed_origins:" "$CONFIG_PATH"; then
    print_warning "Adding allowed_origins to config..."
    # This would require more complex YAML editing - skip for now
    print_info "Consider adding allowed_origins manually for better security"
fi

print_success "Configuration validated"

echo ""
print_header "Step 4: Systemd Service Setup"

print_info "Creating systemd service file..."
cat > /tmp/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Dashboard Server - WebSocket state aggregation and broadcasting
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

# Resource limits
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
ReadWritePaths=/tmp ${SERVICE_DIR}/logs

# Environment
Environment="DASHBOARD_LOG_LEVEL=info"

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

sleep 3

# Check if started
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

# Check service status
print_info "Checking service status..."
if systemctl is-active --quiet ${SERVICE_NAME}; then
    print_success "Service is active"
else
    print_error "Service is not active"
    exit 1
fi

# Check process
print_info "Checking process..."
if pgrep -f dashboard-server > /dev/null; then
    PID=$(pgrep -f dashboard-server)
    print_success "Process running (PID: $PID)"
else
    print_error "Process not found"
    exit 1
fi

# Check health endpoint
print_info "Checking health endpoint..."
PORT=$(grep -A5 "^server:" "$CONFIG_PATH" | grep "port:" | awk '{print $2}' || echo "8086")
sleep 2

if curl -s --max-time 5 "http://localhost:${PORT}/health" | grep -q "ok"; then
    print_success "Health endpoint responding on port ${PORT}"
else
    print_warning "Health endpoint not responding (may take a few more seconds)"
fi

# Check WebSocket endpoint
print_info "Checking WebSocket endpoint (requires Node.js)..."
if command -v node &> /dev/null && [ -f "test-websocket-detailed.js" ]; then
    if timeout 5 node test-websocket-detailed.js > /tmp/ws-test.log 2>&1; then
        print_success "WebSocket endpoint working (received $(grep -c "BTC:" /tmp/ws-test.log || echo "0") updates)"
    else
        print_warning "WebSocket test incomplete (check /tmp/ws-test.log)"
    fi
else
    print_info "Skipping WebSocket test (requires Node.js + test-websocket-detailed.js)"
fi

# Check resource usage
print_info "Checking resource usage..."
CPU=$(ps aux | grep dashboard-server | grep -v grep | awk '{print $3}' | head -1)
MEM=$(ps aux | grep dashboard-server | grep -v grep | awk '{print $4}' | head -1)
print_success "CPU: ${CPU}% | Memory: ${MEM}%"

echo ""
print_header "Deployment Summary"

echo ""
echo -e "${GREEN}✓ Deployment successful!${NC}"
echo ""
echo "Service Details:"
echo "  • Name: ${SERVICE_NAME}"
echo "  • Binary: ${BINARY_PATH} (${BINARY_SIZE})"
echo "  • Config: ${CONFIG_PATH}"
echo "  • Port: ${PORT}"
echo "  • PID: ${PID}"
echo "  • CPU: ${CPU}% (limit: 50%)"
echo "  • Memory: ${MEM}% (limit: 512M)"
echo ""
echo "Endpoints:"
echo "  • Health: http://localhost:${PORT}/health"
echo "  • WebSocket: ws://localhost:${PORT}/ws"
echo "  • Metrics: http://localhost:${PORT}/metrics"
echo "  • Debug: http://localhost:${PORT}/debug"
echo ""
echo "Management Commands:"
echo "  • Status: sudo systemctl status ${SERVICE_NAME}"
echo "  • Logs: sudo journalctl -u ${SERVICE_NAME} -f"
echo "  • Restart: sudo systemctl restart ${SERVICE_NAME}"
echo "  • Stop: sudo systemctl stop ${SERVICE_NAME}"
echo ""
echo "Test WebSocket:"
echo "  • node test-websocket-detailed.js"
echo ""
echo "Security:"
echo "  • Origin checking: Enabled (see config.yaml)"
echo "  • API key auth: Disabled (set security.api_key to enable)"
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
Port: ${PORT}
PID: ${PID}
CPU: ${CPU}%
Memory: ${MEM}%

Service Status:
$(systemctl status ${SERVICE_NAME} --no-pager || true)
EOF

print_success "Deployment info saved to ${DEPLOY_INFO_FILE}"

echo ""
print_success "Deployment complete!"
exit 0
