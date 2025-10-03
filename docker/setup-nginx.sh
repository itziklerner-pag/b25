#!/bin/bash
# Setup Nginx reverse proxy

# Install nginx
sudo apt update
sudo apt install -y nginx

# Copy config
sudo cp /home/mm/dev/b25/docker/nginx-proxy.conf /etc/nginx/sites-available/b25
sudo ln -sf /etc/nginx/sites-available/b25 /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test and reload
sudo nginx -t && sudo systemctl reload nginx

echo "✅ Nginx configured!"
echo "Access: http://66.94.120.149"
