#!/bin/bash
# Clean SSH tunnel script - checks local ports first
# Run this on your LOCAL machine

echo "🧹 Checking for port conflicts on local machine..."
echo ""

PORTS="3000 3001 8000 8080 8081 8082 8083 8084 8085 8086 9090 9093"

for port in $PORTS; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        echo "⚠️  Port $port is in use locally"
        echo "   To free it: lsof -ti:$port | xargs kill -9"
    fi
done

echo ""
echo "💡 Tip: Close any local apps using these ports, then run:"
echo "   ~/tunnel.sh"
echo ""
echo "Or kill all SSH tunnels:"
echo "   killall ssh"
