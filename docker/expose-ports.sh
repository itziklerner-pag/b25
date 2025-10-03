#!/bin/bash
# This script updates docker-compose to expose ports publicly
# WARNING: Only use this behind a firewall or VPN!

cd /home/mm/dev/b25

# Backup original
cp docker/docker-compose.dev.yml docker/docker-compose.dev.yml.backup

# Replace port bindings to expose publicly
sed -i 's/"127.0.0.1:/"/g' docker/docker-compose.dev.yml
sed -i 's/localhost:/0.0.0.0:/g' docker/docker-compose.dev.yml

echo "✅ Ports exposed publicly!"
echo "⚠️  WARNING: Services are now accessible from the internet!"
echo ""
echo "Access with:"
echo "  - Web Dashboard: http://66.94.120.149:3000"
echo "  - Grafana: http://66.94.120.149:3001"
echo "  - Prometheus: http://66.94.120.149:9090"
echo ""
echo "To revert: cp docker/docker-compose.dev.yml.backup docker/docker-compose.dev.yml"
