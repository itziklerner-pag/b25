#!/bin/bash

# Configuration Service - Uninstall Script
# This script removes the configuration service from the system

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVICE_NAME="configuration"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_header "Configuration Service Uninstall"

# Confirm uninstall
echo ""
print_warning "This will remove the configuration service from systemd."
print_info "The database and source code will NOT be deleted."
read -p "Are you sure you want to continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "Uninstall cancelled"
    exit 0
fi

echo ""
print_info "Stopping service..."
if systemctl is-active --quiet ${SERVICE_NAME}.service 2>/dev/null; then
    sudo systemctl stop ${SERVICE_NAME}.service
    print_success "Service stopped"
else
    print_info "Service was not running"
fi

print_info "Disabling service..."
if systemctl is-enabled --quiet ${SERVICE_NAME}.service 2>/dev/null; then
    sudo systemctl disable ${SERVICE_NAME}.service
    print_success "Service disabled"
else
    print_info "Service was not enabled"
fi

print_info "Removing systemd service file..."
if [ -f "$SYSTEMD_SERVICE_PATH" ]; then
    sudo rm "$SYSTEMD_SERVICE_PATH"
    print_success "Service file removed"
else
    print_info "Service file not found"
fi

print_info "Reloading systemd daemon..."
sudo systemctl daemon-reload
sudo systemctl reset-failed
print_success "Systemd daemon reloaded"

echo ""
print_header "Uninstall Complete"
print_success "Configuration service has been removed"
print_info ""
print_info "Note: The following were NOT removed:"
print_info "  - Source code in $(dirname "${BASH_SOURCE[0]}")"
print_info "  - Database (b25_config)"
print_info "  - Docker containers (postgres, nats)"
print_info ""
print_info "To remove the database, run:"
print_info "  docker exec b25-postgres psql -U b25 -c 'DROP DATABASE b25_config;'"
