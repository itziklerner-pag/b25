#!/bin/bash
# Setup WireGuard VPN for Binance access

set -e

echo "🔒 Installing WireGuard VPN..."
echo ""

# Install WireGuard
echo "1️⃣ Installing WireGuard..."
sudo apt update
sudo apt install -y wireguard wireguard-tools resolvconf

# Copy VPN config
echo ""
echo "2️⃣ Copying VPN configuration..."
sudo cp /home/mm/dev/b25/peer48.conf /etc/wireguard/wg0.conf
sudo chmod 600 /etc/wireguard/wg0.conf

# Start VPN
echo ""
echo "3️⃣ Starting VPN connection..."
sudo wg-quick up wg0

# Verify connection
echo ""
echo "4️⃣ Verifying VPN connection..."
echo "Local IP before VPN: $(curl -s ifconfig.me)"
sleep 2
echo "IP through VPN: $(curl -s --interface wg0 ifconfig.me 2>/dev/null || echo 'Testing...')"

# Show WireGuard status
echo ""
echo "5️⃣ WireGuard Status:"
sudo wg show

echo ""
echo "✅ VPN Setup Complete!"
echo ""
echo "To enable VPN on boot:"
echo "  sudo systemctl enable wg-quick@wg0"
echo ""
echo "To stop VPN:"
echo "  sudo wg-quick down wg0"
echo ""
echo "Now restart services to use VPN:"
echo "  ./stop-all-services.sh && ./run-all-services.sh"
