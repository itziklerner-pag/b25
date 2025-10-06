#!/bin/bash
set -e

# Account Monitor Service Uninstall Script
# This script removes the account-monitor service

SERVICE_NAME="account-monitor"
INSTALL_DIR="/opt/${SERVICE_NAME}"
SYSTEMD_SERVICE="${SERVICE_NAME}.service"

echo "========================================"
echo "Account Monitor Service Uninstall"
echo "========================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "ERROR: This script must be run as root (use sudo)"
    exit 1
fi

# Stop the service if running
if systemctl is-active --quiet "${SYSTEMD_SERVICE}"; then
    echo ""
    echo "Stopping ${SERVICE_NAME} service..."
    systemctl stop "${SYSTEMD_SERVICE}"
    echo "✓ Service stopped"
fi

# Disable the service
if systemctl is-enabled --quiet "${SYSTEMD_SERVICE}" 2>/dev/null; then
    echo ""
    echo "Disabling ${SERVICE_NAME} service..."
    systemctl disable "${SYSTEMD_SERVICE}"
    echo "✓ Service disabled"
fi

# Remove systemd service file
if [ -f "/etc/systemd/system/${SYSTEMD_SERVICE}" ]; then
    echo ""
    echo "Removing systemd service file..."
    rm -f "/etc/systemd/system/${SYSTEMD_SERVICE}"
    echo "✓ Service file removed"
fi

# Reload systemd
echo ""
echo "Reloading systemd..."
systemctl daemon-reload
systemctl reset-failed

# Remove installation directory
if [ -d "$INSTALL_DIR" ]; then
    echo ""
    echo "Removing installation directory..."
    echo "WARNING: This will remove all files in ${INSTALL_DIR}"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$INSTALL_DIR"
        echo "✓ Installation directory removed"
    else
        echo "⚠ Installation directory preserved at ${INSTALL_DIR}"
    fi
fi

echo ""
echo "========================================"
echo "Uninstall Complete!"
echo "========================================"
echo ""
echo "The ${SERVICE_NAME} service has been removed."
echo ""
