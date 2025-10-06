#!/bin/bash

# Order Execution Service Uninstall Script
# This script safely removes the order-execution service

set -e

SERVICE_NAME="order-execution"
SERVICE_USER="appuser"
INSTALL_DIR="/opt/b25/${SERVICE_NAME}"
CONFIG_DIR="/etc/b25/${SERVICE_NAME}"
LOG_DIR="/var/log/b25/${SERVICE_NAME}"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
if [ "$EUID" -ne 0 ]; then
    log_error "This script must be run as root"
    exit 1
fi

# Confirmation prompt
confirm_uninstall() {
    log_warn "This will remove the ${SERVICE_NAME} service and all its data"
    read -p "Are you sure you want to continue? (yes/no): " confirmation

    if [ "$confirmation" != "yes" ]; then
        log_info "Uninstall cancelled"
        exit 0
    fi
}

# Stop and disable service
stop_service() {
    log_info "Stopping service..."

    if systemctl is-active --quiet "${SERVICE_NAME}.service"; then
        systemctl stop "${SERVICE_NAME}.service"
        log_info "Service stopped"
    else
        log_info "Service is not running"
    fi

    if systemctl is-enabled --quiet "${SERVICE_NAME}.service" 2>/dev/null; then
        systemctl disable "${SERVICE_NAME}.service"
        log_info "Service disabled"
    fi
}

# Remove systemd service file
remove_systemd_service() {
    log_info "Removing systemd service..."

    if [ -f "$SYSTEMD_DIR/${SERVICE_NAME}.service" ]; then
        rm -f "$SYSTEMD_DIR/${SERVICE_NAME}.service"
        systemctl daemon-reload
        log_info "Systemd service removed"
    else
        log_info "Systemd service file not found"
    fi
}

# Remove installation directory
remove_install_dir() {
    log_info "Removing installation directory..."

    if [ -d "$INSTALL_DIR" ]; then
        rm -rf "$INSTALL_DIR"
        log_info "Installation directory removed"
    else
        log_info "Installation directory not found"
    fi
}

# Remove or backup configuration
remove_config() {
    log_info "Handling configuration..."

    if [ -d "$CONFIG_DIR" ]; then
        # Ask if user wants to backup config
        read -p "Do you want to backup configuration before removal? (yes/no): " backup_choice

        if [ "$backup_choice" = "yes" ]; then
            BACKUP_DIR="$HOME/${SERVICE_NAME}_config_backup_$(date +%Y%m%d_%H%M%S)"
            mkdir -p "$BACKUP_DIR"
            cp -r "$CONFIG_DIR" "$BACKUP_DIR/"
            log_info "Configuration backed up to: $BACKUP_DIR"
        fi

        rm -rf "$CONFIG_DIR"
        log_info "Configuration directory removed"
    else
        log_info "Configuration directory not found"
    fi
}

# Remove or backup logs
remove_logs() {
    log_info "Handling logs..."

    if [ -d "$LOG_DIR" ]; then
        # Ask if user wants to backup logs
        read -p "Do you want to backup logs before removal? (yes/no): " backup_choice

        if [ "$backup_choice" = "yes" ]; then
            BACKUP_DIR="$HOME/${SERVICE_NAME}_logs_backup_$(date +%Y%m%d_%H%M%S)"
            mkdir -p "$BACKUP_DIR"
            cp -r "$LOG_DIR" "$BACKUP_DIR/"
            log_info "Logs backed up to: $BACKUP_DIR"
        fi

        rm -rf "$LOG_DIR"
        log_info "Log directory removed"
    else
        log_info "Log directory not found"
    fi
}

# Remove user (optional)
remove_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        read -p "Do you want to remove the system user '$SERVICE_USER'? (yes/no): " remove_user_choice

        if [ "$remove_user_choice" = "yes" ]; then
            userdel "$SERVICE_USER" 2>/dev/null || log_warn "Failed to remove user $SERVICE_USER"
            log_info "User $SERVICE_USER removed"
        else
            log_info "User $SERVICE_USER kept"
        fi
    else
        log_info "User $SERVICE_USER not found"
    fi
}

# Main uninstall flow
main() {
    log_info "Starting uninstallation of ${SERVICE_NAME}..."

    confirm_uninstall
    stop_service
    remove_systemd_service
    remove_install_dir
    remove_config
    remove_logs
    remove_user

    log_info "Uninstallation complete!"
    log_info ""
    log_info "To reinstall, run: ./scripts/deploy.sh"
}

# Run main
main "$@"
