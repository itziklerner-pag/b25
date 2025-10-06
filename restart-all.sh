#!/bin/bash
echo "🔄 Restarting All B25 Services..."
cd /home/mm/dev/b25

# Stop all
echo "Stopping services..."
for pidfile in logs/*.pid; do
    if [ -f "$pidfile" ]; then
        pid=$(cat "$pidfile")
        kill $pid 2>/dev/null && echo "  Stopped $(basename $pidfile .pid)"
        rm "$pidfile"
    fi
done

sleep 3

# Start infrastructure
echo ""
echo "Starting infrastructure..."
docker compose -f docker-compose.simple.yml up -d

sleep 5

# Start all services
echo ""
echo "Starting trading services..."
./run-all-services.sh

sleep 10

echo ""
echo "✅ All services restarted!"
echo ""
echo "Check status: ./sanity-check.sh"
