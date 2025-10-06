#!/bin/bash
set -e

# Account Monitor Service Deployment Script
# This script builds and deploys the account-monitor service

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="account-monitor"
INSTALL_DIR="/opt/${SERVICE_NAME}"
BINARY_NAME="${SERVICE_NAME}"
SYSTEMD_SERVICE="${SERVICE_NAME}.service"

echo "========================================"
echo "Account Monitor Service Deployment"
echo "========================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "ERROR: This script must be run as root (use sudo)"
    exit 1
fi

# Check environment variables
echo ""
echo "Checking required environment variables..."
MISSING_VARS=()

if [ -z "$BINANCE_API_KEY" ]; then
    MISSING_VARS+=("BINANCE_API_KEY")
fi

if [ -z "$BINANCE_SECRET_KEY" ]; then
    MISSING_VARS+=("BINANCE_SECRET_KEY")
fi

if [ -z "$POSTGRES_PASSWORD" ]; then
    MISSING_VARS+=("POSTGRES_PASSWORD")
fi

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
    echo "ERROR: Missing required environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        echo "  - $var"
    done
    echo ""
    echo "Please set these variables before running the deployment:"
    echo "  export BINANCE_API_KEY='your_key_here'"
    echo "  export BINANCE_SECRET_KEY='your_secret_here'"
    echo "  export POSTGRES_PASSWORD='your_password_here'"
    echo ""
    echo "Or create a .env file and source it:"
    echo "  cp .env.example .env"
    echo "  # Edit .env with your credentials"
    echo "  source .env"
    exit 1
fi

echo "✓ All required environment variables are set"

# Build the service
echo ""
echo "Building ${SERVICE_NAME}..."
cd "$SCRIPT_DIR"

# Ensure dependencies are up to date
go mod download

# Build binary
CGO_ENABLED=1 go build -o "${BINARY_NAME}" ./cmd/server

echo "✓ Build complete"

# Create installation directory
echo ""
echo "Creating installation directory..."
mkdir -p "$INSTALL_DIR"

# Copy binary
echo "Installing binary to ${INSTALL_DIR}..."
cp "${BINARY_NAME}" "${INSTALL_DIR}/"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Copy config
echo "Installing configuration..."
cp config.yaml "${INSTALL_DIR}/"

# Create environment file for systemd
echo ""
echo "Creating environment file..."
cat > "${INSTALL_DIR}/.env" << EOF
BINANCE_API_KEY=${BINANCE_API_KEY}
BINANCE_SECRET_KEY=${BINANCE_SECRET_KEY}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
EOF
chmod 600 "${INSTALL_DIR}/.env"

echo "✓ Environment file created (permissions: 600)"

# Create systemd service file
echo ""
echo "Creating systemd service..."
cat > "/etc/systemd/system/${SYSTEMD_SERVICE}" << 'EOF'
[Unit]
Description=Account Monitor Service
Documentation=https://github.com/yourusername/b25/services/account-monitor
After=network.target postgresql.service redis.service nats.service
Wants=postgresql.service redis.service nats.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/account-monitor
EnvironmentFile=/opt/account-monitor/.env
ExecStart=/opt/account-monitor/account-monitor
Restart=always
RestartSec=10s

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=account-monitor

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo ""
echo "Reloading systemd..."
systemctl daemon-reload

# Enable service
echo "Enabling ${SERVICE_NAME} service..."
systemctl enable "${SYSTEMD_SERVICE}"

# Start or restart service
if systemctl is-active --quiet "${SYSTEMD_SERVICE}"; then
    echo ""
    echo "Restarting ${SERVICE_NAME} service..."
    systemctl restart "${SYSTEMD_SERVICE}"
else
    echo ""
    echo "Starting ${SERVICE_NAME} service..."
    systemctl start "${SYSTEMD_SERVICE}"
fi

# Wait for service to start
sleep 3

# Check status
echo ""
echo "Checking service status..."
if systemctl is-active --quiet "${SYSTEMD_SERVICE}"; then
    echo "✓ Service is running"

    # Test health endpoint
    echo ""
    echo "Testing health endpoint..."
    if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
        echo "✓ Health check passed"
    else
        echo "⚠ Health check failed - service may still be starting up"
    fi
else
    echo "✗ Service failed to start"
    echo ""
    echo "View logs with: journalctl -u ${SYSTEMD_SERVICE} -n 50"
    exit 1
fi

echo ""
echo "========================================"
echo "Deployment Complete!"
echo "========================================"
echo ""
echo "Service Management Commands:"
echo "  Status:  systemctl status ${SYSTEMD_SERVICE}"
echo "  Logs:    journalctl -u ${SYSTEMD_SERVICE} -f"
echo "  Stop:    systemctl stop ${SYSTEMD_SERVICE}"
echo "  Start:   systemctl start ${SYSTEMD_SERVICE}"
echo "  Restart: systemctl restart ${SYSTEMD_SERVICE}"
echo ""
echo "Endpoints:"
echo "  Health:     http://localhost:8080/health"
echo "  Ready:      http://localhost:8080/ready"
echo "  Metrics:    http://localhost:9093/metrics"
echo "  API:        http://localhost:8080/api/*"
echo "  WebSocket:  ws://localhost:8080/ws"
echo "  gRPC:       localhost:50051"
echo ""
