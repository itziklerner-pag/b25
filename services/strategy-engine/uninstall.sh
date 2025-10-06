#!/bin/bash

# Strategy Engine Uninstall Script
# This script removes the strategy engine service

set -e

SERVICE_NAME="strategy-engine"
SERVICE_USER="strategy"
SERVICE_DIR="/opt/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"
LOG_DIR="/var/log/${SERVICE_NAME}"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}========================================${NC}"
echo -e "${RED}Strategy Engine Uninstall Script${NC}"
echo -e "${RED}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    exit 1
fi

# Function to print status
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Confirm uninstall
echo -e "${YELLOW}This will remove the Strategy Engine service and all its files.${NC}"
read -p "Are you sure you want to continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

# Ask about configuration and logs
read -p "Remove configuration files? (yes/no): " -r
REMOVE_CONFIG=$REPLY

read -p "Remove log files? (yes/no): " -r
REMOVE_LOGS=$REPLY

# Step 1: Stop the service
print_status "Stopping service..."
if systemctl is-active --quiet "${SERVICE_NAME}"; then
    systemctl stop "${SERVICE_NAME}"
    print_status "Service stopped"
else
    print_warning "Service is not running"
fi

# Step 2: Disable the service
print_status "Disabling service..."
if systemctl is-enabled --quiet "${SERVICE_NAME}" 2>/dev/null; then
    systemctl disable "${SERVICE_NAME}"
    print_status "Service disabled"
else
    print_warning "Service is not enabled"
fi

# Step 3: Remove systemd service file
print_status "Removing systemd service file..."
if [ -f "${SYSTEMD_DIR}/${SERVICE_NAME}.service" ]; then
    rm "${SYSTEMD_DIR}/${SERVICE_NAME}.service"
    print_status "Systemd service file removed"
fi

# Step 4: Reload systemd
print_status "Reloading systemd daemon..."
systemctl daemon-reload
systemctl reset-failed 2>/dev/null || true

# Step 5: Remove service directory
print_status "Removing service directory..."
if [ -d "$SERVICE_DIR" ]; then
    rm -rf "$SERVICE_DIR"
    print_status "Service directory removed: $SERVICE_DIR"
fi

# Step 6: Remove configuration (if requested)
if [[ $REMOVE_CONFIG =~ ^[Yy][Ee][Ss]$ ]]; then
    print_status "Removing configuration..."
    if [ -d "$CONFIG_DIR" ]; then
        # Backup config before removing
        if [ -f "${CONFIG_DIR}/config.yaml" ]; then
            cp "${CONFIG_DIR}/config.yaml" "/tmp/${SERVICE_NAME}-config-$(date +%Y%m%d-%H%M%S).yaml.backup"
            print_status "Config backed up to /tmp/"
        fi
        rm -rf "$CONFIG_DIR"
        print_status "Configuration removed: $CONFIG_DIR"
    fi
else
    print_warning "Configuration preserved at: $CONFIG_DIR"
fi

# Step 7: Remove logs (if requested)
if [[ $REMOVE_LOGS =~ ^[Yy][Ee][Ss]$ ]]; then
    print_status "Removing logs..."
    if [ -d "$LOG_DIR" ]; then
        rm -rf "$LOG_DIR"
        print_status "Logs removed: $LOG_DIR"
    fi
else
    print_warning "Logs preserved at: $LOG_DIR"
fi

# Step 8: Remove service user
read -p "Remove service user '$SERVICE_USER'? (yes/no): " -r
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    if id "$SERVICE_USER" &>/dev/null; then
        userdel "$SERVICE_USER" 2>/dev/null || print_warning "Could not remove user $SERVICE_USER"
        print_status "User $SERVICE_USER removed"
    fi
else
    print_warning "User $SERVICE_USER preserved"
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Uninstall Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
if [[ ! $REMOVE_CONFIG =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Configuration preserved at: $CONFIG_DIR"
fi
if [[ ! $REMOVE_LOGS =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Logs preserved at: $LOG_DIR"
fi
echo ""
