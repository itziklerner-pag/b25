#!/bin/bash

###############################################################################
# B25 Trading System - Update Environment Files for Domain
# Purpose: Update .env files to use mm.itziklerner.com domain
###############################################################################

set -e

DOMAIN="mm.itziklerner.com"
B25_ROOT="/home/mm/dev/b25"
WEB_ENV="$B25_ROOT/ui/web/.env"

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Updating Environment Files for Domain${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Backup existing .env file
if [ -f "$WEB_ENV" ]; then
    BACKUP_FILE="${WEB_ENV}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$WEB_ENV" "$BACKUP_FILE"
    echo -e "${GREEN}Backed up existing .env to: $BACKUP_FILE${NC}"
fi

# Create new .env file with domain URLs
cat > "$WEB_ENV" <<EOF
# B25 Trading System - Web Dashboard Environment Variables
# Updated for domain: $DOMAIN

# WebSocket URL - Dashboard Server
VITE_WS_URL=wss://$DOMAIN/ws?type=web

# API Gateway URL
VITE_API_URL=https://$DOMAIN/api

# Auth Service URL
VITE_AUTH_URL=https://$DOMAIN/api/auth

# Environment
NODE_ENV=production
EOF

echo -e "${GREEN}Updated web dashboard .env file${NC}"
echo ""
echo -e "${YELLOW}New configuration:${NC}"
cat "$WEB_ENV"
echo ""

# Check if there's a production env file
if [ -f "$B25_ROOT/.env" ]; then
    echo -e "${YELLOW}Found root .env file, you may want to update it as well${NC}"
fi

echo ""
echo -e "${GREEN}Environment files updated successfully!${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Rebuild web dashboard: cd $B25_ROOT/ui/web && npm run build"
echo -e "2. Restart services: cd $B25_ROOT && ./restart-all.sh"
echo ""
