#!/bin/bash

# Auth Service Uninstall Script
# Removes the authentication service and optionally cleans up data

set -e

SERVICE_NAME="b25-auth"
SERVICE_DIR="/opt/b25/auth"
SERVICE_USER="b25"

echo "========================================="
echo "B25 AUTH SERVICE UNINSTALL"
echo "========================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: sudo ./uninstall.sh"
    exit 1
fi

# Confirmation
echo "⚠ WARNING: This will remove the auth service"
read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled"
    exit 0
fi

echo ""
echo "[1/5] Stopping service..."

# Stop and disable service
if systemctl is-active --quiet $SERVICE_NAME; then
    systemctl stop $SERVICE_NAME
    echo "✓ Service stopped"
else
    echo "✓ Service was not running"
fi

if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
    systemctl disable $SERVICE_NAME
    echo "✓ Service disabled"
fi

echo ""
echo "[2/5] Removing systemd service..."

# Remove systemd service file
if [ -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload
    echo "✓ Systemd service removed"
else
    echo "✓ Systemd service file not found"
fi

echo ""
echo "[3/5] Removing service files..."

# Remove service directory
if [ -d "$SERVICE_DIR" ]; then
    # Ask about backing up .env
    if [ -f "$SERVICE_DIR/.env" ]; then
        read -p "Do you want to backup .env file? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            BACKUP_FILE="/tmp/b25-auth-env-backup-$(date +%Y%m%d-%H%M%S).env"
            cp "$SERVICE_DIR/.env" "$BACKUP_FILE"
            chmod 600 "$BACKUP_FILE"
            echo "✓ .env backed up to: $BACKUP_FILE"
        fi
    fi

    rm -rf "$SERVICE_DIR"
    echo "✓ Service directory removed"
else
    echo "✓ Service directory not found"
fi

echo ""
echo "[4/5] Checking for database..."

# Ask about database
echo ""
read -p "Do you want to drop the database? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "Database name [b25_auth]: " DB_NAME
    DB_NAME=${DB_NAME:-b25_auth}

    echo "⚠ This will permanently delete all authentication data!"
    read -p "Type the database name to confirm: " CONFIRM_DB

    if [ "$CONFIRM_DB" = "$DB_NAME" ]; then
        # Try to drop database (may require postgres user)
        sudo -u postgres psql -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null && \
            echo "✓ Database dropped" || \
            echo "⚠ Failed to drop database - you may need to do this manually"
    else
        echo "Database name mismatch - skipping database deletion"
    fi
else
    echo "✓ Database preserved"
fi

echo ""
echo "[5/5] Cleaning up..."

# Clean up logs
if [ -d "/var/log/b25" ]; then
    rm -f /var/log/b25/auth*.log
    echo "✓ Log files removed"
fi

# Optionally remove user
echo ""
read -p "Do you want to remove the service user ($SERVICE_USER)? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if id "$SERVICE_USER" &>/dev/null; then
        userdel "$SERVICE_USER" 2>/dev/null && \
            echo "✓ User removed" || \
            echo "⚠ Failed to remove user"
    fi
else
    echo "✓ User preserved"
fi

echo ""
echo "========================================="
echo "UNINSTALL COMPLETE"
echo "========================================="
echo ""
echo "The auth service has been removed."
echo ""
if [ -n "$BACKUP_FILE" ]; then
    echo "⚠ BACKUP: Your .env file was backed up to:"
    echo "  $BACKUP_FILE"
    echo "  Remember to securely delete this file when no longer needed."
    echo ""
fi
