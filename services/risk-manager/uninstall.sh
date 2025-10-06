#!/bin/bash
set -e

# Risk Manager Service Uninstall Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="b25-risk-manager"
INSTALL_DIR="/opt/b25/risk-manager"
CONFIG_DIR="/etc/b25/risk-manager"
LOG_DIR="/var/log/b25/risk-manager"
DATA_DIR="/var/lib/b25/risk-manager"
SERVICE_USER="b25-risk"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Risk Manager Service Uninstall${NC}"
echo -e "${YELLOW}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Ask for confirmation
read -p "Are you sure you want to uninstall Risk Manager service? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Uninstall cancelled${NC}"
    exit 0
fi

# Stop and disable service
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo -e "${YELLOW}Stopping service...${NC}"
    systemctl stop "$SERVICE_NAME"
fi

if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo -e "${YELLOW}Disabling service...${NC}"
    systemctl disable "$SERVICE_NAME"
fi

# Remove systemd service file
if [ -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
    echo -e "${YELLOW}Removing systemd service file...${NC}"
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload
fi

# Ask about data removal
read -p "Do you want to remove configuration and data? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Remove directories
    echo -e "${YELLOW}Removing directories...${NC}"
    [ -d "$INSTALL_DIR" ] && rm -rf "$INSTALL_DIR"
    [ -d "$CONFIG_DIR" ] && rm -rf "$CONFIG_DIR"
    [ -d "$LOG_DIR" ] && rm -rf "$LOG_DIR"
    [ -d "$DATA_DIR" ] && rm -rf "$DATA_DIR"

    echo -e "${GREEN}All data removed${NC}"
else
    # Only remove binary
    echo -e "${YELLOW}Removing binary only (keeping config and data)...${NC}"
    [ -d "$INSTALL_DIR" ] && rm -rf "$INSTALL_DIR"

    echo -e "${YELLOW}Configuration preserved at: $CONFIG_DIR${NC}"
    echo -e "${YELLOW}Data preserved at: $DATA_DIR${NC}"
    echo -e "${YELLOW}Logs preserved at: $LOG_DIR${NC}"
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Uninstall Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Note: Service user '$SERVICE_USER' was not removed${NC}"
echo "To remove manually: sudo userdel $SERVICE_USER"
