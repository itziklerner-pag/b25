#!/bin/bash

###############################################################################
# B25 Trading System - Nginx SSL Setup with acme.sh
# Domain: mm.itziklerner.com
# Purpose: Install Nginx, configure reverse proxy with SSL using acme.sh
###############################################################################

set -e

DOMAIN="mm.itziklerner.com"
EMAIL="admin@itziklerner.com"
B25_ROOT="/home/mm/dev/b25"
NGINX_SITES_AVAILABLE="/etc/nginx/sites-available"
NGINX_SITES_ENABLED="/etc/nginx/sites-enabled"
ACME_HOME="/root/.acme.sh"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}B25 Trading System - Nginx SSL Setup${NC}"
echo -e "${BLUE}Using acme.sh for SSL certificates${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo bash $0"
    exit 1
fi

# Check if Caddy is running and stop it
echo -e "${YELLOW}Step 0: Checking for Caddy...${NC}"
if systemctl is-active --quiet caddy; then
    echo -e "${YELLOW}Caddy is running, stopping it...${NC}"
    systemctl stop caddy
    systemctl disable caddy
    echo -e "${GREEN}Caddy stopped and disabled${NC}"
else
    echo -e "${GREEN}Caddy is not running${NC}"
fi

echo ""
echo -e "${YELLOW}Step 1: Installing acme.sh...${NC}"
if [ ! -f "$ACME_HOME/acme.sh" ]; then
    curl https://get.acme.sh | sh -s email="$EMAIL"
    echo -e "${GREEN}acme.sh installed successfully${NC}"
else
    echo -e "${GREEN}acme.sh already installed${NC}"
fi

# Source acme.sh to make it available
export LE_WORKING_DIR="$ACME_HOME"
alias acme.sh="$ACME_HOME/acme.sh"

echo ""
echo -e "${YELLOW}Step 2: Creating temporary Nginx config...${NC}"

# Create temporary config for ACME challenge
cat > "$NGINX_SITES_AVAILABLE/$DOMAIN" <<'EOF'
# B25 Trading System - HTTP only (for certificate verification)
server {
    listen 80;
    listen [::]:80;
    server_name mm.itziklerner.com;

    # Allow ACME challenge
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    # Temporary response for all other requests
    location / {
        return 200 "B25 Trading System - SSL Setup in Progress\n";
        add_header Content-Type text/plain;
    }
}
EOF

# Enable the site
ln -sf "$NGINX_SITES_AVAILABLE/$DOMAIN" "$NGINX_SITES_ENABLED/$DOMAIN"

# Remove default site if exists
if [ -f "$NGINX_SITES_ENABLED/default" ]; then
    rm -f "$NGINX_SITES_ENABLED/default"
fi

# Test and start Nginx
nginx -t
systemctl enable nginx
systemctl restart nginx

echo -e "${GREEN}Temporary Nginx config created and loaded${NC}"

echo ""
echo -e "${YELLOW}Step 3: Obtaining SSL certificate using acme.sh...${NC}"

# Create webroot directory
mkdir -p /var/www/html/.well-known/acme-challenge

# Issue certificate using webroot mode
if [ ! -d "$ACME_HOME/${DOMAIN}_ecc" ] && [ ! -d "$ACME_HOME/$DOMAIN" ]; then
    $ACME_HOME/acme.sh --issue -d "$DOMAIN" -w /var/www/html --keylength ec-256
    echo -e "${GREEN}Certificate issued successfully${NC}"
else
    echo -e "${GREEN}Certificate already exists${NC}"
    # Try to renew
    $ACME_HOME/acme.sh --renew -d "$DOMAIN" --force || echo -e "${YELLOW}Renewal not needed or failed${NC}"
fi

# Install certificate to /etc/nginx/ssl/
mkdir -p /etc/nginx/ssl
$ACME_HOME/acme.sh --install-cert -d "$DOMAIN" --ecc \
    --key-file       /etc/nginx/ssl/${DOMAIN}.key \
    --fullchain-file /etc/nginx/ssl/${DOMAIN}.crt \
    --reloadcmd      "systemctl reload nginx"

echo -e "${GREEN}SSL certificate installed${NC}"

echo ""
echo -e "${YELLOW}Step 4: Creating production Nginx config with SSL...${NC}"

# Create production config with SSL and reverse proxy
cat > "$NGINX_SITES_AVAILABLE/$DOMAIN" <<'EOF'
# B25 Trading System - Production Nginx Configuration
# Domain: mm.itziklerner.com

# HTTP - Redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name mm.itziklerner.com;

    # Allow ACME challenge
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    # Redirect all other traffic to HTTPS
    location / {
        return 301 https://$server_name$request_uri;
    }
}

# HTTPS - Main configuration
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name mm.itziklerner.com;

    # SSL Certificate Configuration
    ssl_certificate /etc/nginx/ssl/mm.itziklerner.com.crt;
    ssl_certificate_key /etc/nginx/ssl/mm.itziklerner.com.key;

    # SSL Security Settings (Mozilla Intermediate Configuration)
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_session_tickets off;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Logging
    access_log /var/log/nginx/b25-access.log;
    error_log /var/log/nginx/b25-error.log;

    # Client settings
    client_max_body_size 10M;
    client_body_timeout 30s;
    client_header_timeout 30s;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # Root location - Web Dashboard (React/Vite)
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 90s;
        proxy_connect_timeout 90s;
        proxy_send_timeout 90s;
    }

    # API Gateway
    location /api {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 90s;
        proxy_connect_timeout 90s;
        proxy_send_timeout 90s;

        # CORS headers (if needed)
        add_header Access-Control-Allow-Origin * always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept, Authorization" always;

        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            return 204;
        }
    }

    # WebSocket - Dashboard Server
    location /ws {
        proxy_pass http://localhost:8086/ws;
        proxy_http_version 1.1;

        # WebSocket upgrade headers
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket timeout settings (extended for long-lived connections)
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_connect_timeout 90s;

        # Disable buffering for WebSocket
        proxy_buffering off;
    }

    # Grafana Dashboard
    location /grafana/ {
        proxy_pass http://localhost:3001/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 90s;
        proxy_connect_timeout 90s;
        proxy_send_timeout 90s;
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }

    # Nginx status (restricted to localhost)
    location /nginx_status {
        stub_status on;
        access_log off;
        allow 127.0.0.1;
        deny all;
    }
}
EOF

echo -e "${GREEN}Production Nginx config created${NC}"

echo ""
echo -e "${YELLOW}Step 5: Testing Nginx configuration...${NC}"
if nginx -t; then
    echo -e "${GREEN}Nginx configuration test passed${NC}"
else
    echo -e "${RED}Nginx configuration test failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 6: Reloading Nginx...${NC}"
systemctl reload nginx
echo -e "${GREEN}Nginx reloaded successfully${NC}"

echo ""
echo -e "${YELLOW}Step 7: Setting up SSL auto-renewal...${NC}"
# acme.sh automatically sets up a cron job for renewal
echo -e "${GREEN}SSL auto-renewal is configured via acme.sh cron job${NC}"

# Show cron jobs
crontab -l | grep acme.sh || echo "Cron job will be added on next renewal"

echo ""
echo -e "${YELLOW}Step 8: Configuring firewall (UFW)...${NC}"
if command -v ufw &> /dev/null; then
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw allow 'Nginx Full' 2>/dev/null || true
    echo -e "${GREEN}Firewall configured${NC}"
else
    echo -e "${YELLOW}UFW not installed, skipping firewall configuration${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Setup completed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}SSL Certificate Information:${NC}"
$ACME_HOME/acme.sh --info -d "$DOMAIN" --ecc || echo "Certificate info not available"

echo ""
echo -e "${GREEN}Certificate files location:${NC}"
echo -e "  Key:       /etc/nginx/ssl/${DOMAIN}.key"
echo -e "  Fullchain: /etc/nginx/ssl/${DOMAIN}.crt"

echo ""
echo -e "${GREEN}Next Steps:${NC}"
echo -e "1. Update web dashboard .env file: bash /home/mm/dev/b25/update-env-for-domain.sh"
echo -e "2. Rebuild and restart services"
echo -e "3. Run verification: bash /home/mm/dev/b25/verify-ssl-setup.sh"
echo ""
echo -e "${YELLOW}Access URLs:${NC}"
echo -e "  Web Dashboard:  ${BLUE}https://mm.itziklerner.com${NC}"
echo -e "  WebSocket:      ${BLUE}wss://mm.itziklerner.com/ws${NC}"
echo -e "  API Gateway:    ${BLUE}https://mm.itziklerner.com/api${NC}"
echo -e "  Grafana:        ${BLUE}https://mm.itziklerner.com/grafana${NC}"
echo ""
echo -e "${GREEN}Setup script completed!${NC}"
