#!/bin/bash

###############################################################################
# B25 Trading System - Nginx SSL Setup Script
# Domain: mm.itziklerner.com
# Purpose: Install Nginx, configure reverse proxy with SSL termination
###############################################################################

set -e

DOMAIN="mm.itziklerner.com"
EMAIL="admin@itziklerner.com"
B25_ROOT="/home/mm/dev/b25"
NGINX_SITES_AVAILABLE="/etc/nginx/sites-available"
NGINX_SITES_ENABLED="/etc/nginx/sites-enabled"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}B25 Trading System - Nginx SSL Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo bash $0"
    exit 1
fi

echo -e "${YELLOW}Step 1: Installing Nginx...${NC}"
if ! command -v nginx &> /dev/null; then
    apt-get update
    apt-get install -y nginx
    echo -e "${GREEN}Nginx installed successfully${NC}"
else
    echo -e "${GREEN}Nginx already installed${NC}"
fi

echo ""
echo -e "${YELLOW}Step 2: Installing Certbot for Let's Encrypt...${NC}"
if ! command -v certbot &> /dev/null; then
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
    echo -e "${GREEN}Certbot installed successfully${NC}"
else
    echo -e "${GREEN}Certbot already installed${NC}"
fi

echo ""
echo -e "${YELLOW}Step 3: Creating temporary Nginx config for certificate verification...${NC}"

# Create temporary config without SSL for initial certificate request
cat > "$NGINX_SITES_AVAILABLE/$DOMAIN" <<'EOF'
# B25 Trading System - HTTP only (for certificate verification)
server {
    listen 80;
    listen [::]:80;
    server_name mm.itziklerner.com;

    # Allow Let's Encrypt validation
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

# Test and reload Nginx
nginx -t
systemctl reload nginx
systemctl enable nginx

echo -e "${GREEN}Temporary Nginx config created and loaded${NC}"

echo ""
echo -e "${YELLOW}Step 4: Obtaining SSL certificate from Let's Encrypt...${NC}"

# Check if certificate already exists
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo -e "${GREEN}Certificate already exists for $DOMAIN${NC}"
    echo -e "${YELLOW}Renewing certificate...${NC}"
    certbot renew --nginx --quiet
else
    # Obtain new certificate
    certbot certonly --nginx \
        --non-interactive \
        --agree-tos \
        --email "$EMAIL" \
        -d "$DOMAIN"
fi

echo -e "${GREEN}SSL certificate obtained successfully${NC}"

echo ""
echo -e "${YELLOW}Step 5: Creating production Nginx config with SSL and reverse proxy...${NC}"

# Create production config with SSL and reverse proxy
cat > "$NGINX_SITES_AVAILABLE/$DOMAIN" <<'EOF'
# B25 Trading System - Production Nginx Configuration
# Domain: mm.itziklerner.com

# HTTP - Redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name mm.itziklerner.com;

    # Allow Let's Encrypt validation
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
    ssl_certificate /etc/letsencrypt/live/mm.itziklerner.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mm.itziklerner.com/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/mm.itziklerner.com/chain.pem;

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
echo -e "${YELLOW}Step 6: Testing Nginx configuration...${NC}"
if nginx -t; then
    echo -e "${GREEN}Nginx configuration test passed${NC}"
else
    echo -e "${RED}Nginx configuration test failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 7: Reloading Nginx...${NC}"
systemctl reload nginx
echo -e "${GREEN}Nginx reloaded successfully${NC}"

echo ""
echo -e "${YELLOW}Step 8: Setting up SSL auto-renewal...${NC}"

# Certbot auto-renewal is set up by default via systemd timer
# Let's verify it's enabled
systemctl enable certbot.timer
systemctl start certbot.timer

echo -e "${GREEN}SSL auto-renewal configured${NC}"

# Create a renewal hook to reload Nginx after certificate renewal
mkdir -p /etc/letsencrypt/renewal-hooks/deploy
cat > /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh <<'HOOK_EOF'
#!/bin/bash
systemctl reload nginx
HOOK_EOF
chmod +x /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh

echo ""
echo -e "${YELLOW}Step 9: Configuring firewall (UFW)...${NC}"
if command -v ufw &> /dev/null; then
    ufw allow 'Nginx Full'
    ufw delete allow 'Nginx HTTP' 2>/dev/null || true
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
certbot certificates

echo ""
echo -e "${GREEN}Next Steps:${NC}"
echo -e "1. Update web dashboard .env file to use domain URLs"
echo -e "2. Restart your B25 services"
echo -e "3. Verify access to:"
echo -e "   - ${BLUE}https://mm.itziklerner.com${NC} (Web Dashboard)"
echo -e "   - ${BLUE}wss://mm.itziklerner.com/ws${NC} (WebSocket)"
echo -e "   - ${BLUE}https://mm.itziklerner.com/api${NC} (API Gateway)"
echo -e "   - ${BLUE}https://mm.itziklerner.com/grafana${NC} (Grafana)"
echo ""
echo -e "${YELLOW}Certificate auto-renewal status:${NC}"
systemctl status certbot.timer --no-pager

echo ""
echo -e "${GREEN}Setup script completed!${NC}"
