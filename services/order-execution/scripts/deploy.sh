#!/bin/bash

# Order Execution Service Deployment Script
# This script deploys the order-execution service with proper credential handling

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVICE_NAME="order-execution"
SERVICE_USER="appuser"
INSTALL_DIR="/opt/b25/${SERVICE_NAME}"
CONFIG_DIR="/etc/b25/${SERVICE_NAME}"
LOG_DIR="/var/log/b25/${SERVICE_NAME}"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
if [ "$EUID" -ne 0 ]; then
    log_error "This script must be run as root"
    exit 1
fi

# Function to check dependencies
check_dependencies() {
    log_info "Checking dependencies..."

    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Check Redis
    if ! systemctl is-active --quiet redis; then
        log_warn "Redis is not running. Starting Redis..."
        systemctl start redis || log_error "Failed to start Redis"
    fi

    # Check NATS (optional)
    if ! systemctl is-active --quiet nats; then
        log_warn "NATS is not running. Service will work but events won't be published."
    fi

    log_info "Dependencies check complete"
}

# Function to check environment variables
check_credentials() {
    log_info "Checking credentials..."

    if [ -z "$BINANCE_API_KEY" ] || [ -z "$BINANCE_SECRET_KEY" ]; then
        log_error "Binance API credentials not set in environment"
        log_error "Please set BINANCE_API_KEY and BINANCE_SECRET_KEY"
        log_error "Or create /etc/b25/${SERVICE_NAME}/.env file"
        exit 1
    fi

    log_info "Credentials found"
}

# Function to create user
create_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        log_info "User $SERVICE_USER already exists"
    else
        log_info "Creating user $SERVICE_USER..."
        useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
    fi
}

# Function to create directories
create_directories() {
    log_info "Creating directories..."

    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"

    chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
    chown -R "$SERVICE_USER:$SERVICE_USER" "$LOG_DIR"

    log_info "Directories created"
}

# Function to build the service
build_service() {
    log_info "Building service..."

    cd "$PROJECT_ROOT"

    # Download dependencies
    go mod download
    go mod tidy

    # Build
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
        -ldflags '-extldflags "-static" -s -w' \
        -o "$INSTALL_DIR/${SERVICE_NAME}" ./cmd/server

    chmod +x "$INSTALL_DIR/${SERVICE_NAME}"
    chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/${SERVICE_NAME}"

    log_info "Build complete"
}

# Function to install config
install_config() {
    log_info "Installing configuration..."

    # Copy config file
    cp "$PROJECT_ROOT/config.yaml" "$CONFIG_DIR/config.yaml"

    # Create .env file if it doesn't exist
    if [ ! -f "$CONFIG_DIR/.env" ]; then
        log_info "Creating .env file..."
        cat > "$CONFIG_DIR/.env" << EOF
# Order Execution Service Environment Variables
BINANCE_API_KEY=${BINANCE_API_KEY:-}
BINANCE_SECRET_KEY=${BINANCE_SECRET_KEY:-}
BINANCE_TESTNET=${BINANCE_TESTNET:-true}

REDIS_ADDRESS=${REDIS_ADDRESS:-localhost:6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-}
NATS_ADDRESS=${NATS_ADDRESS:-nats://localhost:4222}

LOG_LEVEL=${LOG_LEVEL:-info}
LOG_FORMAT=${LOG_FORMAT:-json}
EOF
        chmod 600 "$CONFIG_DIR/.env"
        chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/.env"
    else
        log_info ".env file already exists, skipping creation"
    fi

    chmod 644 "$CONFIG_DIR/config.yaml"
    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/config.yaml"

    log_info "Configuration installed"
}

# Function to install systemd service
install_systemd_service() {
    log_info "Installing systemd service..."

    cat > "$SYSTEMD_DIR/${SERVICE_NAME}.service" << EOF
[Unit]
Description=Order Execution Service
After=network.target redis.service nats.service
Wants=redis.service nats.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$CONFIG_DIR/.env
ExecStart=$INSTALL_DIR/${SERVICE_NAME}
Restart=always
RestartSec=5
StandardOutput=append:$LOG_DIR/service.log
StandardError=append:$LOG_DIR/error.log

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$LOG_DIR

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    log_info "Systemd service installed"
}

# Function to start service
start_service() {
    log_info "Starting service..."

    systemctl enable "${SERVICE_NAME}.service"
    systemctl restart "${SERVICE_NAME}.service"

    sleep 2

    if systemctl is-active --quiet "${SERVICE_NAME}.service"; then
        log_info "Service started successfully"
    else
        log_error "Service failed to start"
        journalctl -u "${SERVICE_NAME}.service" -n 50 --no-pager
        exit 1
    fi
}

# Function to verify deployment
verify_deployment() {
    log_info "Verifying deployment..."

    # Check health endpoint
    sleep 3

    if curl -f -s http://localhost:9091/health/live > /dev/null; then
        log_info "Health check passed"
    else
        log_warn "Health check failed - service may still be starting"
    fi

    # Show status
    systemctl status "${SERVICE_NAME}.service" --no-pager
}

# Main deployment flow
main() {
    log_info "Starting deployment of ${SERVICE_NAME}..."

    check_dependencies
    check_credentials
    create_user
    create_directories
    build_service
    install_config
    install_systemd_service
    start_service
    verify_deployment

    log_info "Deployment complete!"
    log_info ""
    log_info "Service status: systemctl status ${SERVICE_NAME}"
    log_info "Service logs:   journalctl -u ${SERVICE_NAME} -f"
    log_info "Health check:   curl http://localhost:9091/health"
}

# Run main
main "$@"
