#!/bin/bash
# SSH Tunnel Script for B25 HFT Trading System
# Run this on your LOCAL machine to access the remote VPS services

# Usage: ./ssh-tunnel.sh <username>@<vps-ip>
# Example: ./ssh-tunnel.sh root@192.168.1.100

if [ -z "$1" ]; then
    echo "Usage: $0 <username>@<vps-ip>"
    echo "Example: $0 root@192.168.1.100"
    exit 1
fi

VPS="$1"

echo "🚇 Creating SSH tunnels to $VPS..."
echo ""
echo "This will forward the following ports:"
echo "  - Web Dashboard: http://localhost:3000"
echo "  - Grafana: http://localhost:3001"
echo "  - Prometheus: http://localhost:9090"
echo "  - API Gateway: http://localhost:8000"
echo "  - All service APIs and metrics"
echo ""
echo "Press Ctrl+C to close all tunnels"
echo ""

ssh -N -L 3000:localhost:3000 \
       -L 3001:localhost:3001 \
       -L 8000:localhost:8000 \
       -L 8080:localhost:8080 \
       -L 8081:localhost:8081 \
       -L 8082:localhost:8082 \
       -L 8083:localhost:8083 \
       -L 8084:localhost:8084 \
       -L 8085:localhost:8085 \
       -L 8086:localhost:8086 \
       -L 9090:localhost:9090 \
       -L 9093:localhost:9093 \
       -L 9100:localhost:9100 \
       -L 9101:localhost:9101 \
       -L 9102:localhost:9102 \
       -L 9103:localhost:9103 \
       -L 9104:localhost:9104 \
       -L 9105:localhost:9105 \
       -L 9106:localhost:9106 \
       -L 50051:localhost:50051 \
       -L 50052:localhost:50052 \
       -L 50053:localhost:50053 \
       -L 50054:localhost:50054 \
       -L 50055:localhost:50055 \
       -L 50056:localhost:50056 \
       "$VPS"
