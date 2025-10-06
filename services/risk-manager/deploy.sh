#!/bin/bash
set -e

# Risk Manager Service Deployment Script
# This script builds and deploys the Risk Manager service

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="risk-manager"
BINARY_NAME="risk-manager"
INSTALL_DIR="/opt/b25/${SERVICE_NAME}"
CONFIG_DIR="/etc/b25/${SERVICE_NAME}"
LOG_DIR="/var/log/b25/${SERVICE_NAME}"
DATA_DIR="/var/lib/b25/${SERVICE_NAME}"
SERVICE_USER="b25-risk"
SERVICE_GROUP="b25"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Risk Manager Service Deployment${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Create service user if it doesn't exist
if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
    echo -e "${YELLOW}Creating service user: $SERVICE_USER${NC}"
    useradd -r -s /bin/false -g "$SERVICE_GROUP" "$SERVICE_USER" 2>/dev/null || \
    (groupadd "$SERVICE_GROUP" 2>/dev/null; useradd -r -s /bin/false -g "$SERVICE_GROUP" "$SERVICE_USER")
fi

# Create directories
echo -e "${YELLOW}Creating directories...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$INSTALL_DIR/migrations"

# Build the service
echo -e "${YELLOW}Building Risk Manager service...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/server

if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Copy migrations
echo -e "${YELLOW}Copying migrations...${NC}"
if [ -d "./migrations" ]; then
    cp -r ./migrations/* "$INSTALL_DIR/migrations/"
fi

# Copy or create config file
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    echo -e "${YELLOW}Creating default config file...${NC}"
    cat > "$CONFIG_DIR/config.yaml" << 'EOF'
server:
  port: 8083
  mode: production
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 15s

database:
  host: localhost
  port: 5432
  user: b25
  password: changeme
  database: b25_config
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  max_retries: 3
  pool_size: 10

nats:
  url: nats://localhost:4222
  max_reconnect: 10
  reconnect_wait: 2s
  alert_subject: risk.alerts
  emergency_topic: risk.emergency

grpc:
  port: 50052
  max_connection_idle: 5m
  max_connection_age: 30m
  keep_alive_interval: 30s
  keep_alive_timeout: 10s
  auth_enabled: false
  api_key: ""

risk:
  monitor_interval: 1s
  cache_ttl: 100ms
  policy_cache_ttl: 1s
  max_leverage: 10.0
  max_drawdown_percent: 0.20
  emergency_threshold: 0.25
  alert_window: 5m
  account_monitor_url: localhost:50053
  market_data_redis_db: 1

logging:
  level: info
  format: json

metrics:
  enabled: true
  port: 8083
EOF
    echo -e "${YELLOW}Default config created. Please edit $CONFIG_DIR/config.yaml${NC}"
else
    echo -e "${GREEN}Config file already exists at $CONFIG_DIR/config.yaml${NC}"
fi

# Set permissions
echo -e "${YELLOW}Setting permissions...${NC}"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR"
chmod 750 "$INSTALL_DIR"
chmod 750 "$CONFIG_DIR"
chmod 750 "$LOG_DIR"
chmod 750 "$DATA_DIR"
chmod 755 "$INSTALL_DIR/$BINARY_NAME"

# Install systemd service
echo -e "${YELLOW}Installing systemd service...${NC}"
cat > /etc/systemd/system/b25-risk-manager.service << EOF
[Unit]
Description=B25 Risk Manager Service
Documentation=https://github.com/yourorg/b25
After=network.target postgresql.service redis.service nats.service
Wants=postgresql.service redis.service nats.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$LOG_DIR $DATA_DIR
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096
MemoryLimit=512M
CPUQuota=200%

# Environment
Environment="RISK_CONFIG_PATH=$CONFIG_DIR/config.yaml"

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=b25-risk-manager

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo -e "${YELLOW}Reloading systemd...${NC}"
systemctl daemon-reload

# Enable service
echo -e "${YELLOW}Enabling service...${NC}"
systemctl enable b25-risk-manager

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Edit configuration: $CONFIG_DIR/config.yaml"
echo "2. Run database migrations (if needed)"
echo "3. Start service: sudo systemctl start b25-risk-manager"
echo "4. Check status: sudo systemctl status b25-risk-manager"
echo "5. View logs: sudo journalctl -u b25-risk-manager -f"
echo ""
echo -e "${YELLOW}Service endpoints:${NC}"
echo "  gRPC:   localhost:50052"
echo "  HTTP:   localhost:8083"
echo "  Health: http://localhost:8083/health"
echo "  Metrics: http://localhost:8083/metrics"
