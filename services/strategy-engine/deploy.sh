#!/bin/bash

# Strategy Engine Deployment Script
# This script builds and deploys the strategy engine service

set -e

SERVICE_NAME="strategy-engine"
SERVICE_USER="strategy"
SERVICE_DIR="/opt/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"
LOG_DIR="/var/log/${SERVICE_NAME}"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Strategy Engine Deployment Script${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    exit 1
fi

# Function to print status
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: Create service user if not exists
print_status "Creating service user..."
if id "$SERVICE_USER" &>/dev/null; then
    print_warning "User $SERVICE_USER already exists"
else
    useradd -r -s /bin/false -d "$SERVICE_DIR" "$SERVICE_USER"
    print_status "User $SERVICE_USER created"
fi

# Step 2: Create directories
print_status "Creating directories..."
mkdir -p "$SERVICE_DIR"/{bin,plugins/go,plugins/python}
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"

# Step 3: Build the service
print_status "Building the service..."
if [ -f "go.mod" ]; then
    # Install dependencies
    go mod download

    # Build the binary
    CGO_ENABLED=1 go build -o "${SERVICE_DIR}/bin/${SERVICE_NAME}" ./cmd/server

    # Build plugins if any
    if [ -d "plugins/go" ]; then
        cd plugins/go
        for plugin in *.go; do
            if [ -f "$plugin" ] && [ "$plugin" != "*.go" ]; then
                name="${plugin%.go}"
                print_status "Building plugin: $name"
                go build -buildmode=plugin -o "${SERVICE_DIR}/plugins/go/${name}.so" "$plugin" || print_warning "Failed to build $plugin"
            fi
        done
        cd ../..
    fi
else
    print_error "go.mod not found. Are you in the correct directory?"
    exit 1
fi

# Step 4: Copy configuration
print_status "Copying configuration..."
if [ -f "config.example.yaml" ]; then
    if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
        cp config.example.yaml "${CONFIG_DIR}/config.yaml"
        print_status "Configuration copied to ${CONFIG_DIR}/config.yaml"
        print_warning "Please edit ${CONFIG_DIR}/config.yaml with your settings"
    else
        print_warning "Configuration already exists at ${CONFIG_DIR}/config.yaml"
        print_warning "Backup created at ${CONFIG_DIR}/config.yaml.backup"
        cp "${CONFIG_DIR}/config.yaml" "${CONFIG_DIR}/config.yaml.backup"
        cp config.example.yaml "${CONFIG_DIR}/config.yaml.new"
        print_warning "New config available at ${CONFIG_DIR}/config.yaml.new"
    fi
else
    print_error "config.example.yaml not found"
    exit 1
fi

# Step 5: Copy plugins
print_status "Copying plugins..."
if [ -d "plugins" ]; then
    cp -r plugins/* "${SERVICE_DIR}/plugins/" 2>/dev/null || true
fi

# Step 6: Set permissions
print_status "Setting permissions..."
chown -R "$SERVICE_USER:$SERVICE_USER" "$SERVICE_DIR"
chown -R "$SERVICE_USER:$SERVICE_USER" "$LOG_DIR"
chmod +x "${SERVICE_DIR}/bin/${SERVICE_NAME}"

# Step 7: Create systemd service file
print_status "Creating systemd service file..."
cat > "${SYSTEMD_DIR}/${SERVICE_NAME}.service" << EOF
[Unit]
Description=Strategy Engine - Trading Strategy Execution Service
After=network.target redis.service nats.service
Wants=redis.service nats.service

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
WorkingDirectory=${SERVICE_DIR}

# Environment
Environment="CONFIG_PATH=${CONFIG_DIR}/config.yaml"
Environment="STRATEGY_ENGINE_API_KEY="

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096
MemoryLimit=2G
CPUQuota=200%

# Execute
ExecStart=${SERVICE_DIR}/bin/${SERVICE_NAME}
ExecReload=/bin/kill -HUP \$MAINPID

# Restart policy
Restart=on-failure
RestartSec=10s
StartLimitInterval=200s
StartLimitBurst=5

# Logging
StandardOutput=append:${LOG_DIR}/strategy-engine.log
StandardError=append:${LOG_DIR}/strategy-engine-error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${LOG_DIR}
ReadOnlyPaths=${CONFIG_DIR}

[Install]
WantedBy=multi-user.target
EOF

print_status "Systemd service file created"

# Step 8: Reload systemd
print_status "Reloading systemd daemon..."
systemctl daemon-reload

# Step 9: Enable service
print_status "Enabling service..."
systemctl enable "${SERVICE_NAME}.service"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Edit configuration: ${CONFIG_DIR}/config.yaml"
echo "2. Set API key (if needed): export STRATEGY_ENGINE_API_KEY=your-key"
echo "3. Start the service: systemctl start ${SERVICE_NAME}"
echo "4. Check status: systemctl status ${SERVICE_NAME}"
echo "5. View logs: journalctl -u ${SERVICE_NAME} -f"
echo "   or: tail -f ${LOG_DIR}/strategy-engine.log"
echo ""
echo "Health check: curl http://localhost:9092/health"
echo "Metrics: curl http://localhost:9092/metrics"
echo ""
