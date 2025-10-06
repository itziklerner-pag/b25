#!/bin/bash

# Analytics Service Uninstall Script
# Description: Safely removes the analytics service and cleans up resources

set -e

SERVICE_NAME="analytics"
SERVICE_DIR="/opt/b25/${SERVICE_NAME}"
CONFIG_DIR="/etc/b25/${SERVICE_NAME}"
LOG_DIR="/var/log/b25/${SERVICE_NAME}"
USER="b25-analytics"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Analytics Service Uninstall${NC}"
echo "======================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    exit 1
fi

# Confirmation prompt
read -p "Are you sure you want to uninstall the Analytics service? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled"
    exit 0
fi

# Ask about data backup
read -p "Do you want to backup configuration and data? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    BACKUP_DIR="/tmp/b25-analytics-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"

    [ -d "$CONFIG_DIR" ] && cp -r "$CONFIG_DIR" "$BACKUP_DIR/"
    [ -d "$LOG_DIR" ] && cp -r "$LOG_DIR" "$BACKUP_DIR/"

    echo -e "${GREEN}Backup created at: $BACKUP_DIR${NC}"
fi

# Step 1: Stop the service
echo -e "${YELLOW}[1/7] Stopping service...${NC}"
if systemctl is-active --quiet b25-analytics; then
    systemctl stop b25-analytics
    echo "Service stopped"
else
    echo "Service not running"
fi

# Step 2: Disable the service
echo -e "${YELLOW}[2/7] Disabling service...${NC}"
if systemctl is-enabled --quiet b25-analytics 2>/dev/null; then
    systemctl disable b25-analytics
    echo "Service disabled"
else
    echo "Service not enabled"
fi

# Step 3: Remove systemd service file
echo -e "${YELLOW}[3/7] Removing systemd service...${NC}"
if [ -f "/etc/systemd/system/b25-analytics.service" ]; then
    rm /etc/systemd/system/b25-analytics.service
    systemctl daemon-reload
    echo "Systemd service removed"
else
    echo "Systemd service file not found"
fi

# Step 4: Remove service directories
echo -e "${YELLOW}[4/7] Removing service files...${NC}"
[ -d "$SERVICE_DIR" ] && rm -rf "$SERVICE_DIR" && echo "Removed $SERVICE_DIR"
[ -d "$CONFIG_DIR" ] && rm -rf "$CONFIG_DIR" && echo "Removed $CONFIG_DIR"

# Step 5: Ask about log removal
read -p "Remove logs from $LOG_DIR? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    [ -d "$LOG_DIR" ] && rm -rf "$LOG_DIR" && echo "Removed $LOG_DIR"
else
    echo "Logs preserved in $LOG_DIR"
fi

# Step 6: Ask about user removal
read -p "Remove service user '$USER'? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if id "$USER" &>/dev/null; then
        userdel "$USER"
        echo "Removed user $USER"
    fi
else
    echo "User $USER preserved"
fi

# Step 7: Ask about database cleanup
echo -e "${YELLOW}[7/7] Database cleanup...${NC}"
read -p "Remove analytics database tables? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}To manually remove database:${NC}"
    echo "  psql -U postgres -c 'DROP DATABASE IF EXISTS analytics;'"
    echo "  psql -U postgres -c 'DROP USER IF EXISTS analytics_user;'"
else
    echo "Database preserved"
fi

echo ""
echo -e "${GREEN}======================================"
echo "Uninstall Complete"
echo "======================================${NC}"
echo ""
echo -e "${YELLOW}The following may still exist:${NC}"
echo "- Database: analytics"
echo "- Redis keys: dashboard:*, counter:events:*, etc."
echo "- Kafka consumer group: analytics-consumer-group"
echo ""
echo -e "${GREEN}Uninstall completed successfully!${NC}"
