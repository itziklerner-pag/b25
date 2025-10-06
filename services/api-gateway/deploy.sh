#!/bin/bash

# API Gateway Service Deployment Script
# This script builds, installs, and configures the API Gateway service as a systemd service

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="api-gateway"
SERVICE_USER="apigateway"
SERVICE_GROUP="apigateway"
INSTALL_DIR="/opt/api-gateway"
CONFIG_DIR="/etc/api-gateway"
LOG_DIR="/var/log/api-gateway"
DATA_DIR="/var/lib/api-gateway"
BINARY_NAME="api-gateway"

# Default values
BUILD_BINARY=true
INSTALL_SERVICE=true
START_SERVICE=true
ENABLE_SERVICE=true

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-build)
            BUILD_BINARY=false
            shift
            ;;
        --no-install)
            INSTALL_SERVICE=false
            shift
            ;;
        --no-start)
            START_SERVICE=false
            shift
            ;;
        --no-enable)
            ENABLE_SERVICE=false
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --no-build      Skip building the binary"
            echo "  --no-install    Skip installing the service"
            echo "  --no-start      Skip starting the service"
            echo "  --no-enable     Skip enabling the service at boot"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

log_info "Starting API Gateway deployment..."

# Step 1: Build the binary
if [ "$BUILD_BINARY" = true ]; then
    log_info "Building API Gateway binary..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Build the binary
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -X main.Version=1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o bin/${BINARY_NAME} \
        ./cmd/server

    if [ $? -eq 0 ]; then
        log_info "Binary built successfully"
    else
        log_error "Failed to build binary"
        exit 1
    fi
else
    log_info "Skipping binary build"
fi

# Step 2: Create service user and group
log_info "Creating service user and group..."
if ! id -u ${SERVICE_USER} &> /dev/null; then
    useradd --system --no-create-home --shell /bin/false ${SERVICE_USER}
    log_info "Created user: ${SERVICE_USER}"
else
    log_info "User ${SERVICE_USER} already exists"
fi

# Step 3: Create directories
log_info "Creating directories..."
mkdir -p ${INSTALL_DIR}
mkdir -p ${CONFIG_DIR}
mkdir -p ${LOG_DIR}
mkdir -p ${DATA_DIR}

# Step 4: Copy binary
log_info "Installing binary..."
cp bin/${BINARY_NAME} ${INSTALL_DIR}/${BINARY_NAME}
chmod +x ${INSTALL_DIR}/${BINARY_NAME}

# Step 5: Copy configuration
log_info "Installing configuration..."
if [ -f "config.yaml" ]; then
    cp config.yaml ${CONFIG_DIR}/config.yaml
elif [ -f "config.example.yaml" ]; then
    cp config.example.yaml ${CONFIG_DIR}/config.yaml
    log_warn "Using example config. Please update ${CONFIG_DIR}/config.yaml with production values"
else
    log_error "No configuration file found"
    exit 1
fi

# Step 6: Set permissions
log_info "Setting permissions..."
chown -R ${SERVICE_USER}:${SERVICE_GROUP} ${INSTALL_DIR}
chown -R ${SERVICE_USER}:${SERVICE_GROUP} ${CONFIG_DIR}
chown -R ${SERVICE_USER}:${SERVICE_GROUP} ${LOG_DIR}
chown -R ${SERVICE_USER}:${SERVICE_GROUP} ${DATA_DIR}
chmod 600 ${CONFIG_DIR}/config.yaml  # Protect config file with secrets

# Step 7: Install systemd service
if [ "$INSTALL_SERVICE" = true ]; then
    log_info "Installing systemd service..."

    cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=API Gateway Service
Documentation=https://github.com/b25/api-gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_GROUP}

# Working directory
WorkingDirectory=${INSTALL_DIR}

# Environment
Environment="CONFIG_PATH=${CONFIG_DIR}/config.yaml"
Environment="LOG_LEVEL=info"

# Binary execution
ExecStart=${INSTALL_DIR}/${BINARY_NAME}

# Restart policy
Restart=on-failure
RestartSec=5s
StartLimitInterval=60s
StartLimitBurst=3

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Memory limits (adjust based on your needs)
MemoryLimit=2G
MemoryHigh=1.5G

# CPU limits
CPUQuota=200%

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${LOG_DIR} ${DATA_DIR}
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true
RestrictSUIDSGID=true
LockPersonality=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF

    log_info "Systemd service installed"

    # Reload systemd
    systemctl daemon-reload
    log_info "Systemd daemon reloaded"
else
    log_info "Skipping systemd service installation"
fi

# Step 8: Enable service
if [ "$ENABLE_SERVICE" = true ] && [ "$INSTALL_SERVICE" = true ]; then
    log_info "Enabling service at boot..."
    systemctl enable ${SERVICE_NAME}
    log_info "Service enabled"
else
    log_info "Skipping service enable"
fi

# Step 9: Start service
if [ "$START_SERVICE" = true ] && [ "$INSTALL_SERVICE" = true ]; then
    log_info "Starting service..."

    # Stop if already running
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "Service is already running, restarting..."
        systemctl restart ${SERVICE_NAME}
    else
        systemctl start ${SERVICE_NAME}
    fi

    # Wait a moment for service to start
    sleep 2

    # Check status
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "Service started successfully"

        # Show service status
        systemctl status ${SERVICE_NAME} --no-pager
    else
        log_error "Service failed to start"
        log_error "Check logs with: journalctl -u ${SERVICE_NAME} -n 50"
        exit 1
    fi
else
    log_info "Skipping service start"
fi

# Step 10: Display summary
echo ""
log_info "========================================="
log_info "API Gateway Deployment Complete"
log_info "========================================="
echo ""
echo "Service Name:     ${SERVICE_NAME}"
echo "Install Dir:      ${INSTALL_DIR}"
echo "Config Dir:       ${CONFIG_DIR}"
echo "Log Dir:          ${LOG_DIR}"
echo "Data Dir:         ${DATA_DIR}"
echo ""
echo "Useful Commands:"
echo "  Start:          sudo systemctl start ${SERVICE_NAME}"
echo "  Stop:           sudo systemctl stop ${SERVICE_NAME}"
echo "  Restart:        sudo systemctl restart ${SERVICE_NAME}"
echo "  Status:         sudo systemctl status ${SERVICE_NAME}"
echo "  Logs:           sudo journalctl -u ${SERVICE_NAME} -f"
echo "  Config:         ${CONFIG_DIR}/config.yaml"
echo ""
log_warn "IMPORTANT: Update ${CONFIG_DIR}/config.yaml with production values"
log_warn "  - Change JWT secret"
log_warn "  - Update API keys"
log_warn "  - Configure backend service URLs"
log_warn "  - Set up TLS certificates"
echo ""
