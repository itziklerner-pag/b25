#!/bin/bash

###############################################################################
# B25 Trading System - SSL Setup Verification Script
# Purpose: Verify Nginx, SSL, and reverse proxy configuration
###############################################################################

DOMAIN="mm.itziklerner.com"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}B25 Trading System - SSL Verification${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check service
check_service() {
    local name=$1
    local port=$2

    if nc -z localhost "$port" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $name (port $port) is running"
        return 0
    else
        echo -e "${RED}✗${NC} $name (port $port) is NOT running"
        return 1
    fi
}

# Function to test HTTP endpoint
test_http() {
    local url=$1
    local expected=$2
    local name=$3

    echo -e "\n${YELLOW}Testing: $name${NC}"
    echo -e "URL: $url"

    response=$(curl -s -o /dev/null -w "%{http_code}" -k "$url" 2>/dev/null)

    if [ "$response" = "$expected" ]; then
        echo -e "${GREEN}✓${NC} HTTP $response (expected $expected)"
        return 0
    else
        echo -e "${RED}✗${NC} HTTP $response (expected $expected)"
        return 1
    fi
}

# Function to test WebSocket
test_websocket() {
    local url=$1
    local name=$2

    echo -e "\n${YELLOW}Testing: $name${NC}"
    echo -e "URL: $url"

    # Use websocat if available, otherwise use curl for basic connectivity test
    if command -v websocat &> /dev/null; then
        timeout 5 websocat -n1 "$url" 2>/dev/null && \
            echo -e "${GREEN}✓${NC} WebSocket connection successful" || \
            echo -e "${RED}✗${NC} WebSocket connection failed"
    else
        # Basic TCP connection test
        local host=$(echo "$url" | sed -E 's|wss?://([^/]+).*|\1|')
        local port=443
        if nc -z "$host" "$port" 2>/dev/null; then
            echo -e "${GREEN}✓${NC} TCP connection to $host:$port successful"
        else
            echo -e "${RED}✗${NC} TCP connection to $host:$port failed"
        fi
    fi
}

echo -e "${BLUE}1. Checking Nginx Status${NC}"
if systemctl is-active --quiet nginx; then
    echo -e "${GREEN}✓${NC} Nginx is running"
    nginx -v 2>&1 | head -1
else
    echo -e "${RED}✗${NC} Nginx is NOT running"
fi

echo -e "\n${BLUE}2. Checking SSL Certificate${NC}"
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo -e "${GREEN}✓${NC} SSL certificate exists"

    # Show certificate details
    echo ""
    sudo certbot certificates 2>/dev/null | grep -A 10 "$DOMAIN" || true

    # Check certificate expiry
    expiry=$(sudo openssl x509 -in /etc/letsencrypt/live/$DOMAIN/cert.pem -noout -enddate 2>/dev/null | cut -d= -f2)
    echo -e "\n${YELLOW}Certificate expires:${NC} $expiry"
else
    echo -e "${RED}✗${NC} SSL certificate NOT found"
fi

echo -e "\n${BLUE}3. Checking Nginx Configuration${NC}"
if sudo nginx -t 2>&1 | grep -q "successful"; then
    echo -e "${GREEN}✓${NC} Nginx configuration is valid"
else
    echo -e "${RED}✗${NC} Nginx configuration has errors"
    sudo nginx -t
fi

echo -e "\n${BLUE}4. Checking Backend Services${NC}"
check_service "Web Dashboard" 3000
check_service "API Gateway" 8000
check_service "Dashboard Server (WebSocket)" 8086
check_service "Grafana" 3001

echo -e "\n${BLUE}5. Testing HTTP to HTTPS Redirect${NC}"
test_http "http://$DOMAIN/health" "301" "HTTP Redirect"

echo -e "\n${BLUE}6. Testing HTTPS Endpoints${NC}"
test_http "https://$DOMAIN/health" "200" "Health Check"
test_http "https://$DOMAIN/" "200" "Web Dashboard"
test_http "https://$DOMAIN/api" "200" "API Gateway"

echo -e "\n${BLUE}7. Testing WebSocket${NC}"
test_websocket "wss://$DOMAIN/ws?type=web" "Dashboard WebSocket"

echo -e "\n${BLUE}8. Checking DNS Resolution${NC}"
dns_ip=$(host "$DOMAIN" | grep "has address" | awk '{print $4}')
server_ip=$(curl -4 -s ifconfig.me)

if [ "$dns_ip" = "$server_ip" ]; then
    echo -e "${GREEN}✓${NC} DNS resolves correctly"
    echo -e "Domain: $DOMAIN → $dns_ip"
    echo -e "Server: $server_ip"
else
    echo -e "${YELLOW}⚠${NC} DNS mismatch"
    echo -e "Domain resolves to: $dns_ip"
    echo -e "Server IP: $server_ip"
fi

echo -e "\n${BLUE}9. Checking Firewall Status${NC}"
if command -v ufw &> /dev/null; then
    if sudo ufw status | grep -q "Status: active"; then
        echo -e "${GREEN}✓${NC} UFW is active"
        sudo ufw status | grep -E "(80|443|Nginx)"
    else
        echo -e "${YELLOW}⚠${NC} UFW is not active"
    fi
else
    echo -e "${YELLOW}⚠${NC} UFW not installed"
fi

echo -e "\n${BLUE}10. Checking SSL Auto-Renewal${NC}"
if systemctl is-active --quiet certbot.timer; then
    echo -e "${GREEN}✓${NC} Certbot auto-renewal timer is active"
    sudo systemctl status certbot.timer --no-pager | grep -E "(Active|Trigger)"
else
    echo -e "${RED}✗${NC} Certbot auto-renewal timer is NOT active"
fi

echo -e "\n${BLUE}11. Testing SSL Security${NC}"
echo -e "Checking SSL configuration with OpenSSL..."
echo | timeout 5 openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" 2>/dev/null | \
    grep -E "(Protocol|Cipher)" | head -5

echo -e "\n${BLUE}12. Nginx Logs (Last 10 Lines)${NC}"
echo -e "${YELLOW}Access Log:${NC}"
sudo tail -5 /var/log/nginx/b25-access.log 2>/dev/null || echo "No access log yet"
echo -e "\n${YELLOW}Error Log:${NC}"
sudo tail -5 /var/log/nginx/b25-error.log 2>/dev/null || echo "No errors logged"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${GREEN}Verification Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Access URLs:${NC}"
echo -e "Web Dashboard:  ${BLUE}https://$DOMAIN${NC}"
echo -e "WebSocket:      ${BLUE}wss://$DOMAIN/ws${NC}"
echo -e "API Gateway:    ${BLUE}https://$DOMAIN/api${NC}"
echo -e "Grafana:        ${BLUE}https://$DOMAIN/grafana${NC}"
echo ""
