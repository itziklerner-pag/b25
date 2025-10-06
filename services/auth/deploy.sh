#!/bin/bash

# Auth Service Deployment Script
# Deploys the authentication service with proper security configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="b25-auth"
SERVICE_USER="${SERVICE_USER:-b25}"
SERVICE_DIR="/opt/b25/auth"

echo "========================================="
echo "B25 AUTH SERVICE DEPLOYMENT"
echo "========================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: sudo ./deploy.sh"
    exit 1
fi

echo "[1/8] Checking prerequisites..."

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "Error: Node.js is not installed"
    echo "Please install Node.js 20+ first"
    exit 1
fi

NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "Error: Node.js version must be 18 or higher (found: $(node -v))"
    exit 1
fi

# Check PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "Error: PostgreSQL client is not installed"
    exit 1
fi

echo "✓ Node.js $(node -v) found"
echo "✓ PostgreSQL client found"
echo ""

echo "[2/8] Creating service user and directories..."

# Create service user if not exists
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /bin/bash -d "$SERVICE_DIR" -m "$SERVICE_USER"
    echo "✓ Created user: $SERVICE_USER"
else
    echo "✓ User already exists: $SERVICE_USER"
fi

# Create directories
mkdir -p "$SERVICE_DIR"
mkdir -p /var/log/b25
chown -R "$SERVICE_USER:$SERVICE_USER" /var/log/b25

echo "✓ Directories created"
echo ""

echo "[3/8] Copying service files..."

# Copy service files
cp -r "$SCRIPT_DIR"/* "$SERVICE_DIR/"
chown -R "$SERVICE_USER:$SERVICE_USER" "$SERVICE_DIR"

echo "✓ Files copied to $SERVICE_DIR"
echo ""

echo "[4/8] Installing dependencies..."

cd "$SERVICE_DIR"
sudo -u "$SERVICE_USER" npm install --production

echo "✓ Dependencies installed"
echo ""

echo "[5/8] Generating JWT secrets..."

# Check if .env exists
if [ -f "$SERVICE_DIR/.env" ]; then
    echo "⚠ .env file already exists"
    read -p "Do you want to regenerate JWT secrets? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping JWT secret generation"
    else
        REGENERATE_SECRETS=true
    fi
else
    REGENERATE_SECRETS=true
fi

if [ "$REGENERATE_SECRETS" = true ]; then
    # Copy .env.example if .env doesn't exist
    if [ ! -f "$SERVICE_DIR/.env" ]; then
        cp "$SERVICE_DIR/.env.example" "$SERVICE_DIR/.env"
    fi

    # Generate strong secrets
    JWT_ACCESS_SECRET=$(openssl rand -base64 64 | tr -d '\n')
    JWT_REFRESH_SECRET=$(openssl rand -base64 64 | tr -d '\n')

    # Update .env file
    sed -i "s|JWT_ACCESS_SECRET=.*|JWT_ACCESS_SECRET=$JWT_ACCESS_SECRET|" "$SERVICE_DIR/.env"
    sed -i "s|JWT_REFRESH_SECRET=.*|JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET|" "$SERVICE_DIR/.env"

    echo "✓ JWT secrets generated and updated in .env"
    echo "⚠ IMPORTANT: Keep .env file secure - it contains secrets!"
fi

# Set secure permissions on .env
chmod 600 "$SERVICE_DIR/.env"
chown "$SERVICE_USER:$SERVICE_USER" "$SERVICE_DIR/.env"

echo ""

echo "[6/8] Configuring database..."

# Prompt for database credentials
echo "Database configuration:"
read -p "Database host [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "Database port [5432]: " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "Database name [b25_auth]: " DB_NAME
DB_NAME=${DB_NAME:-b25_auth}

read -p "Database user [b25]: " DB_USER
DB_USER=${DB_USER:-b25}

read -sp "Database password: " DB_PASSWORD
echo

# Update database credentials in .env
sed -i "s|DB_HOST=.*|DB_HOST=$DB_HOST|" "$SERVICE_DIR/.env"
sed -i "s|DB_PORT=.*|DB_PORT=$DB_PORT|" "$SERVICE_DIR/.env"
sed -i "s|DB_NAME=.*|DB_NAME=$DB_NAME|" "$SERVICE_DIR/.env"
sed -i "s|DB_USER=.*|DB_USER=$DB_USER|" "$SERVICE_DIR/.env"
sed -i "s|DB_PASSWORD=.*|DB_PASSWORD=$DB_PASSWORD|" "$SERVICE_DIR/.env"

echo "✓ Database configuration updated"
echo ""

echo "[7/8] Installing systemd service..."

# Create systemd service file
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=B25 Authentication Service
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$SERVICE_DIR
Environment=NODE_ENV=production
ExecStart=/usr/bin/node $SERVICE_DIR/src/server.js
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/b25
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "✓ Systemd service installed"
echo ""

echo "[8/8] Starting service..."

# Enable and start service
systemctl enable $SERVICE_NAME
systemctl restart $SERVICE_NAME

# Wait for service to start
sleep 2

# Check if service is running
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "✓ Service started successfully"

    # Test health endpoint
    if command -v curl &> /dev/null; then
        echo ""
        echo "Testing health endpoint..."
        HEALTH=$(curl -s http://localhost:9097/health | jq -r '.data.status' 2>/dev/null || echo "error")
        if [ "$HEALTH" = "healthy" ]; then
            echo "✓ Health check passed"
        else
            echo "⚠ Health check failed - check logs: journalctl -u $SERVICE_NAME -f"
        fi
    fi
else
    echo "✗ Service failed to start"
    echo "Check logs: journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi

echo ""
echo "========================================="
echo "DEPLOYMENT COMPLETE"
echo "========================================="
echo ""
echo "Service Status:"
echo "  Name: $SERVICE_NAME"
echo "  Status: $(systemctl is-active $SERVICE_NAME)"
echo "  Port: 9097"
echo "  Location: $SERVICE_DIR"
echo ""
echo "Useful Commands:"
echo "  Status:  systemctl status $SERVICE_NAME"
echo "  Start:   systemctl start $SERVICE_NAME"
echo "  Stop:    systemctl stop $SERVICE_NAME"
echo "  Restart: systemctl restart $SERVICE_NAME"
echo "  Logs:    journalctl -u $SERVICE_NAME -f"
echo ""
echo "Health Check:"
echo "  curl http://localhost:9097/health"
echo ""
echo "⚠ SECURITY REMINDER:"
echo "  - Keep $SERVICE_DIR/.env file secure"
echo "  - Never commit .env to version control"
echo "  - Regularly rotate JWT secrets"
echo "  - Monitor logs for suspicious activity"
echo ""
