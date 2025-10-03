#!/bin/bash
# Stop all B25 services

cd /home/mm/dev/b25

echo "🛑 Stopping all B25 services..."

if [ -d logs ]; then
    for pidfile in logs/*.pid; do
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile")
            echo "Stopping $(basename $pidfile .pid) (PID: $pid)..."
            kill $pid 2>/dev/null || true
            rm "$pidfile"
        fi
    done
fi

echo "✅ All services stopped!"
