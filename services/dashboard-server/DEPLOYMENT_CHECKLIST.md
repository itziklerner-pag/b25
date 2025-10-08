# Dashboard Server - Deployment Checklist

## Build & Deploy Steps

### 1. Build the Service
```bash
cd /home/mm/dev/b25/services/dashboard-server
make build
```

This creates: `/home/mm/dev/b25/services/dashboard-server/bin/dashboard-server`

### 2. Verify Build
```bash
./bin/dashboard-server --help
# Or just run to check it starts
./bin/dashboard-server
```

### 3. Stop Running Service (if needed)
```bash
# Find the process
ps aux | grep dashboard-server

# Kill it
sudo pkill -f dashboard-server

# Or use systemd if configured
sudo systemctl stop dashboard-server
```

### 4. Deploy New Binary
```bash
# If using systemd, copy to proper location
sudo cp bin/dashboard-server /usr/local/bin/dashboard-server

# Or update symlink if that's your setup
sudo ln -sf /home/mm/dev/b25/services/dashboard-server/bin/dashboard-server /usr/local/bin/dashboard-server
```

### 5. Restart Service
```bash
# If using systemd
sudo systemctl start dashboard-server
sudo systemctl status dashboard-server

# Or run directly
./bin/dashboard-server
```

### 6. Verify Admin Page
```bash
# Test health endpoint
curl http://localhost:8086/health

# Test service info API
curl http://localhost:8086/api/service-info

# Open in browser
xdg-open http://localhost:8086/
```

### 7. Test via Nginx Proxy
```bash
# If nginx is configured with /services/dashboard-server/ route
curl http://localhost/services/dashboard-server/health
curl http://localhost/services/dashboard-server/api/service-info

# Open in browser
xdg-open http://localhost/services/dashboard-server/
```

## Quick Commands

### One-Line Build & Test
```bash
cd /home/mm/dev/b25/services/dashboard-server && make build && ./bin/dashboard-server
```

### Check If Service Is Running
```bash
curl -s http://localhost:8086/health | jq .
```

### View Service Logs (if using systemd)
```bash
sudo journalctl -u dashboard-server -f
```

### Test All Endpoints
```bash
# Health
curl http://localhost:8086/health

# Service Info (admin API)
curl http://localhost:8086/api/service-info | jq .

# Debug (full state)
curl http://localhost:8086/debug | jq .

# History
curl "http://localhost:8086/api/v1/history?type=market_data&limit=10" | jq .

# Metrics
curl http://localhost:8086/metrics
```

## Configuration

### Config File Location
- `/home/mm/dev/b25/services/dashboard-server/config.yaml`
- `/etc/dashboard-server/config.yaml`

### Key Settings
```yaml
server:
  port: 8086
  log_level: info

redis:
  url: localhost:6379

backend_services:
  order_execution:
    url: localhost:50051
  strategy_engine:
    url: http://localhost:8082
  account_monitor:
    url: localhost:50055

websocket:
  allowed_origins:
    - http://localhost:5173
    - http://localhost:3000
    - http://localhost:8080

security:
  api_key: ""  # Optional
```

## Nginx Configuration

If using nginx proxy, ensure this location block exists:

```nginx
location /services/dashboard-server/ {
    proxy_pass http://localhost:8086/;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket support
    proxy_read_timeout 86400;
    proxy_send_timeout 86400;
}
```

Then reload nginx:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Troubleshooting

### Build Fails
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

### Port Already in Use
```bash
# Find what's using port 8086
sudo lsof -i :8086

# Kill it
sudo kill -9 <PID>
```

### Can't Connect to Backend Services
```bash
# Test each backend service
curl http://localhost:8082/health  # Strategy Engine
curl http://localhost:6379         # Redis (should fail with non-HTTP)

# Check gRPC services
grpcurl -plaintext localhost:50051 list  # Order Execution
grpcurl -plaintext localhost:50055 list  # Account Monitor
```

### Admin Page Not Loading
1. Check service is running: `curl http://localhost:8086/health`
2. Check nginx config if using proxy
3. Check browser console for errors
4. Verify CORS headers in response

## Files Changed

### New Files
- `/home/mm/dev/b25/services/dashboard-server/internal/admin/admin.go` - Admin handler
- `/home/mm/dev/b25/services/dashboard-server/ADMIN_PAGE.md` - Documentation
- `/home/mm/dev/b25/services/dashboard-server/DEPLOYMENT_CHECKLIST.md` - This file

### Modified Files
- `/home/mm/dev/b25/services/dashboard-server/cmd/server/main.go` - Added admin routes
- `/home/mm/dev/b25/services/dashboard-server/internal/server/server.go` - Added GetClientCount() method

## Rollback Plan

If issues occur, revert changes:

```bash
cd /home/mm/dev/b25/services/dashboard-server
git diff HEAD cmd/server/main.go
git diff HEAD internal/server/server.go
git checkout HEAD -- cmd/server/main.go internal/server/server.go
rm -rf internal/admin/
make clean
make build
```

## Success Criteria

✓ Service builds without errors
✓ Service starts and responds to /health
✓ Admin page loads at http://localhost:8086/
✓ Service info API returns valid JSON
✓ All 4 metric cards display data
✓ Backend services show connection status
✓ Test buttons work and show results
✓ Auto-refresh updates every 5 seconds
✓ WebSocket test connects and receives data

## Next Steps After Deployment

1. Monitor service logs for errors
2. Test WebSocket connections from clients
3. Verify metrics in /metrics endpoint
4. Check Prometheus scraping if configured
5. Test under load with multiple WebSocket clients
6. Consider adding authentication for production
7. Set up alerting for service health
