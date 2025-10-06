#!/bin/bash
set -e

# Dashboard Server - Uninstall Script

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SERVICE_NAME="dashboard-server"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Dashboard Server Uninstall${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Confirm
read -p "Are you sure you want to uninstall ${SERVICE_NAME}? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Uninstall cancelled${NC}"
    exit 0
fi

# Stop service
echo -e "${BLUE}Stopping service...${NC}"
if systemctl is-active --quiet ${SERVICE_NAME} 2>/dev/null; then
    sudo systemctl stop ${SERVICE_NAME}
    echo -e "${GREEN}✓ Service stopped${NC}"
else
    echo -e "${YELLOW}⚠ Service not running${NC}"
fi

# Disable service
echo -e "${BLUE}Disabling service...${NC}"
if systemctl is-enabled --quiet ${SERVICE_NAME} 2>/dev/null; then
    sudo systemctl disable ${SERVICE_NAME}
    echo -e "${GREEN}✓ Service disabled${NC}"
else
    echo -e "${YELLOW}⚠ Service not enabled${NC}"
fi

# Remove systemd file
echo -e "${BLUE}Removing systemd service file...${NC}"
if [ -f "$SYSTEMD_SERVICE_PATH" ]; then
    sudo rm "$SYSTEMD_SERVICE_PATH"
    sudo systemctl daemon-reload
    echo -e "${GREEN}✓ Systemd service removed${NC}"
else
    echo -e "${YELLOW}⚠ Systemd service file not found${NC}"
fi

# Ask about binary
echo ""
read -p "Remove binary (dashboard-server)? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f "dashboard-server" ]; then
        rm dashboard-server
        echo -e "${GREEN}✓ Binary removed${NC}"
    fi
fi

# Ask about config
echo ""
read -p "Remove config.yaml? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f "config.yaml" ]; then
        rm config.yaml
        echo -e "${GREEN}✓ config.yaml removed${NC}"
    fi
fi

# Ask about logs
echo ""
read -p "Remove logs directory? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "logs" ]; then
        rm -rf logs
        echo -e "${GREEN}✓ Logs removed${NC}"
    fi
fi

echo ""
echo -e "${GREEN}✓ Uninstall complete!${NC}"
echo ""
