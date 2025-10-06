#!/bin/bash

# Analytics Service Deployment Script
# Description: Builds and deploys the analytics service with all dependencies

set -e

SERVICE_NAME="analytics"
SERVICE_DIR="/opt/b25/${SERVICE_NAME}"
CONFIG_DIR="/etc/b25/${SERVICE_NAME}"
LOG_DIR="/var/log/b25/${SERVICE_NAME}"
BINARY_NAME="${SERVICE_NAME}-server"
USER="b25-analytics"
GROUP="b25"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Analytics Service Deployment${NC}"
echo "======================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    exit 1
fi

# Step 1: Create service user if it doesn't exist
echo -e "${YELLOW}[1/10] Checking service user...${NC}"
if ! id "$USER" &>/dev/null; then
    useradd -r -s /bin/false -d "$SERVICE_DIR" -c "B25 Analytics Service" "$USER"
    echo "Created user: $USER"
else
    echo "User $USER already exists"
fi

# Step 2: Create required directories
echo -e "${YELLOW}[2/10] Creating directories...${NC}"
mkdir -p "$SERVICE_DIR"/{bin,config,scripts,migrations}
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"
mkdir -p /var/run/b25

echo "Created directories"

# Step 3: Build the service
echo -e "${YELLOW}[3/10] Building service...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

go build -o "$SERVICE_DIR/bin/$BINARY_NAME" ./cmd/server
echo "Built binary: $SERVICE_DIR/bin/$BINARY_NAME"

# Step 4: Copy configuration files
echo -e "${YELLOW}[4/10] Copying configuration...${NC}"
if [ -f "config.yaml" ]; then
    cp config.yaml "$CONFIG_DIR/config.yaml"
    echo "Copied config.yaml"
elif [ -f "config.example.yaml" ]; then
    cp config.example.yaml "$CONFIG_DIR/config.yaml"
    echo "Copied config.example.yaml as config.yaml"
    echo -e "${YELLOW}WARNING: Please edit $CONFIG_DIR/config.yaml with your configuration${NC}"
else
    echo -e "${RED}Error: No configuration file found${NC}"
    exit 1
fi

# Step 5: Copy migrations
echo -e "${YELLOW}[5/10] Copying database migrations...${NC}"
if [ -d "migrations" ]; then
    cp migrations/*.sql "$SERVICE_DIR/migrations/" 2>/dev/null || true
    echo "Copied migration files"
fi

# Step 6: Copy helper scripts
echo -e "${YELLOW}[6/10] Copying helper scripts...${NC}"
if [ -d "scripts" ]; then
    cp scripts/*.sh "$SERVICE_DIR/scripts/" 2>/dev/null || true
    chmod +x "$SERVICE_DIR/scripts/"*.sh 2>/dev/null || true
    echo "Copied helper scripts"
fi

# Step 7: Set permissions
echo -e "${YELLOW}[7/10] Setting permissions...${NC}"
chown -R "$USER":"$GROUP" "$SERVICE_DIR"
chown -R "$USER":"$GROUP" "$CONFIG_DIR"
chown -R "$USER":"$GROUP" "$LOG_DIR"
chmod 750 "$SERVICE_DIR/bin/$BINARY_NAME"
chmod 640 "$CONFIG_DIR/config.yaml"

echo "Permissions set"

# Step 8: Install systemd service
echo -e "${YELLOW}[8/10] Installing systemd service...${NC}"
cat > /etc/systemd/system/b25-analytics.service <<EOF
[Unit]
Description=B25 Analytics Service
Documentation=https://github.com/b25/analytics
After=network.target postgresql.service redis.service kafka.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$SERVICE_DIR
ExecStart=$SERVICE_DIR/bin/$BINARY_NAME -config $CONFIG_DIR/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=b25-analytics

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$LOG_DIR /var/run/b25
ProtectKernelTunables=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true

# Resource limits
LimitNOFILE=65536
MemoryMax=2G
CPUQuota=200%

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
echo "Systemd service installed"

# Step 9: Enable service
echo -e "${YELLOW}[9/10] Enabling service...${NC}"
systemctl enable b25-analytics.service
echo "Service enabled"

# Step 10: Display status and next steps
echo -e "${YELLOW}[10/10] Deployment complete!${NC}"
echo ""
echo -e "${GREEN}======================================"
echo "Deployment Summary"
echo "======================================${NC}"
echo "Service binary:    $SERVICE_DIR/bin/$BINARY_NAME"
echo "Configuration:     $CONFIG_DIR/config.yaml"
echo "Migrations:        $SERVICE_DIR/migrations/"
echo "Logs:              $LOG_DIR"
echo "User:              $USER"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Edit configuration: sudo nano $CONFIG_DIR/config.yaml"
echo "2. Run migrations: psql -U postgres -d analytics -f $SERVICE_DIR/migrations/001_initial_schema.sql"
echo "3. Start service: sudo systemctl start b25-analytics"
echo "4. Check status: sudo systemctl status b25-analytics"
echo "5. View logs: sudo journalctl -u b25-analytics -f"
echo ""
echo -e "${GREEN}Service URLs:${NC}"
echo "API:               http://localhost:9097"
echo "Prometheus:        http://localhost:9098/metrics"
echo "Health:            http://localhost:9097/health"
echo ""
echo -e "${GREEN}Deployment completed successfully!${NC}"
