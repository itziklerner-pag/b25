#!/bin/bash

# API Gateway Service Uninstall Script
# This script removes the API Gateway service, files, and user

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="api-gateway"
SERVICE_USER="apigateway"
SERVICE_GROUP="apigateway"
INSTALL_DIR="/opt/api-gateway"
CONFIG_DIR="/etc/api-gateway"
LOG_DIR="/var/log/api-gateway"
DATA_DIR="/var/lib/api-gateway"

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

# Confirm uninstall
echo ""
log_warn "This will completely remove the API Gateway service and all associated files"
read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    log_info "Uninstall cancelled"
    exit 0
fi

log_info "Starting API Gateway uninstallation..."

# Step 1: Stop the service
if systemctl is-active --quiet ${SERVICE_NAME}; then
    log_info "Stopping service..."
    systemctl stop ${SERVICE_NAME}
    log_info "Service stopped"
else
    log_info "Service is not running"
fi

# Step 2: Disable the service
if systemctl is-enabled --quiet ${SERVICE_NAME} 2>/dev/null; then
    log_info "Disabling service..."
    systemctl disable ${SERVICE_NAME}
    log_info "Service disabled"
else
    log_info "Service is not enabled"
fi

# Step 3: Remove systemd service file
if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    log_info "Removing systemd service file..."
    rm -f /etc/systemd/system/${SERVICE_NAME}.service
    systemctl daemon-reload
    log_info "Systemd service file removed"
else
    log_info "Systemd service file not found"
fi

# Step 4: Remove directories
log_info "Removing directories..."

if [ -d "${INSTALL_DIR}" ]; then
    rm -rf ${INSTALL_DIR}
    log_info "Removed ${INSTALL_DIR}"
fi

if [ -d "${CONFIG_DIR}" ]; then
    rm -rf ${CONFIG_DIR}
    log_info "Removed ${CONFIG_DIR}"
fi

if [ -d "${LOG_DIR}" ]; then
    rm -rf ${LOG_DIR}
    log_info "Removed ${LOG_DIR}"
fi

if [ -d "${DATA_DIR}" ]; then
    rm -rf ${DATA_DIR}
    log_info "Removed ${DATA_DIR}"
fi

# Step 5: Remove user and group
if id -u ${SERVICE_USER} &> /dev/null; then
    log_info "Removing service user..."
    userdel ${SERVICE_USER}
    log_info "User ${SERVICE_USER} removed"
else
    log_info "User ${SERVICE_USER} does not exist"
fi

# Step 6: Clean up any remaining files
log_info "Cleaning up..."
systemctl reset-failed ${SERVICE_NAME} 2>/dev/null || true

echo ""
log_info "========================================="
log_info "API Gateway Uninstall Complete"
log_info "========================================="
echo ""
log_info "The API Gateway service has been completely removed from the system"
echo ""
