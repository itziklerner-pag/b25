#!/bin/bash
# Test what Dashboard Server is actually sending

echo "Installing websocat if needed..."
which websocat || (curl -L https://github.com/vi/websocat/releases/download/v1.12.0/websocat.x86_64-unknown-linux-musl -o /tmp/websocat && chmod +x /tmp/websocat && sudo mv /tmp/websocat /usr/local/bin/)

echo ""
echo "Connecting to Dashboard Server WebSocket..."
echo "This will show the actual JSON messages being sent..."
echo ""

timeout 5 websocat "ws://localhost:8086/ws?type=web" 2>/dev/null | head -1 | jq . || echo "Install websocat manually: sudo apt install websocat"
