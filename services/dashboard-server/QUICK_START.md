# Dashboard Server - Quick Start Guide

## Prerequisites

```bash
# Install Go 1.21 or higher
go version  # Should show go1.21 or higher

# Start Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Verify Redis is running
redis-cli ping  # Should return "PONG"
```

## Running Locally

### Option 1: Using Go

```bash
cd /home/mm/dev/b25/services/dashboard-server

# Download dependencies
go mod download

# Run the server
go run cmd/server/main.go

# Server starts on http://localhost:8080
```

### Option 2: Using Makefile

```bash
cd /home/mm/dev/b25/services/dashboard-server

# Show available commands
make help

# Build and run
make deps
make build
./bin/dashboard-server
```

### Option 3: Using Docker

```bash
cd /home/mm/dev/b25/services/dashboard-server

# Build Docker image
docker build -t dashboard-server:latest .

# Run container
docker run -d --name dashboard-server \
  -p 8080:8080 \
  -e DASHBOARD_REDIS_URL=host.docker.internal:6379 \
  dashboard-server:latest

# View logs
docker logs -f dashboard-server
```

## Testing the Service

### 1. Health Check

```bash
curl http://localhost:8080/health
# Expected: {"status":"ok","service":"dashboard-server"}
```

### 2. Metrics

```bash
curl http://localhost:8080/metrics
# Expected: Prometheus metrics output
```

### 3. WebSocket Connection (using wscat)

```bash
# Install wscat if needed
npm install -g wscat

# Connect as TUI client with JSON format
wscat -c "ws://localhost:8080/ws?type=tui&format=json"

# Once connected, send subscription message
{"type":"subscribe","channels":["market_data","orders","account"]}

# You should receive a snapshot message with demo data
```

### 4. WebSocket Connection (using curl)

```bash
# Upgrade to WebSocket
curl -i -N -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: $(openssl rand -base64 16)" \
  http://localhost:8080/ws?type=web&format=json
```

### 5. Testing with Go Client

Create a file `test-client.go`:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8080/ws?type=tui&format=json"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer ws.Close()

	// Subscribe to channels
	subscribeMsg := `{"type":"subscribe","channels":["market_data","account"]}`
	if err := ws.WriteMessage(websocket.TextMessage, []byte(subscribeMsg)); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	fmt.Println("Connected! Receiving messages...")

	// Read messages for 10 seconds
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			fmt.Println("Test complete!")
			return
		default:
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			fmt.Printf("Received: %s\n", string(message))
		}
	}
}
```

Run it:

```bash
go run test-client.go
```

## Environment Variables

```bash
# Port to listen on (default: 8080)
export DASHBOARD_PORT=8080

# Log level: debug, info, warn, error (default: info)
export DASHBOARD_LOG_LEVEL=info

# Redis connection string (default: localhost:6379)
export DASHBOARD_REDIS_URL=localhost:6379
```

## Verifying It Works

You should see:

1. **In terminal output**:
```
{"level":"info","time":1696348800,"message":"Starting Dashboard Server Service","version":"1.0.0"}
{"level":"info","message":"State aggregator started"}
{"level":"info","message":"Broadcaster started"}
{"level":"info","message":"Dashboard Server started successfully","port":8080}
```

2. **Health endpoint returns OK**:
```bash
curl http://localhost:8080/health
# {"status":"ok","service":"dashboard-server"}
```

3. **Metrics endpoint returns data**:
```bash
curl http://localhost:8080/metrics | grep dashboard_
# dashboard_connected_clients{client_type="TUI"} 0
# dashboard_active_connections 0
```

4. **WebSocket client receives data**:
   - Connection successful
   - Receives snapshot message with demo data
   - Receives periodic updates

## Common Issues

### Issue: "Connection refused" when connecting to WebSocket

**Solution**: Ensure the server is running and listening on port 8080

```bash
# Check if server is running
curl http://localhost:8080/health

# Check if port is in use
lsof -i :8080  # Linux/Mac
netstat -ano | findstr :8080  # Windows
```

### Issue: "connection error: dial tcp [::1]:6379: connect: connection refused"

**Solution**: Redis is not running

```bash
# Start Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Verify Redis is running
redis-cli ping
```

### Issue: "go: command not found"

**Solution**: Install Go

```bash
# Ubuntu/Debian
sudo apt install golang-go

# Mac
brew install go

# Verify installation
go version
```

### Issue: High CPU usage

**Solution**: This is expected during development with demo data updates. In production, updates come from actual backend services.

## Next Steps

1. ✅ Start the server
2. ✅ Verify health check works
3. ✅ Connect a WebSocket client
4. ✅ Subscribe to channels
5. ✅ Receive demo data
6. 📝 Integrate with backend services (Market Data, Orders, etc.)
7. 📝 Connect TUI or Web UI
8. 📝 Deploy to staging environment

## Development Workflow

```bash
# 1. Make code changes
vim internal/server/server.go

# 2. Run tests
make test

# 3. Run locally
make run

# 4. Test with client
wscat -c "ws://localhost:8080/ws?type=tui&format=json"

# 5. Build for production
make build

# 6. Build Docker image
make docker-build
```

## Useful Commands

```bash
# Show all make targets
make help

# Run with debug logging
DASHBOARD_LOG_LEVEL=debug go run cmd/server/main.go

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Build optimized binary
go build -ldflags="-s -w" -o bin/dashboard-server ./cmd/server

# Check for race conditions
go run -race cmd/server/main.go

# Profile CPU usage
go run cmd/server/main.go &
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8081 cpu.prof
```

## Architecture Overview

```
┌─────────────────────────────────────────┐
│   Dashboard Server (Port 8080)          │
│                                         │
│   WebSocket (/ws)                       │
│   ├─ TUI clients (100ms updates)        │
│   └─ Web clients (250ms updates)        │
│                                         │
│   REST API                              │
│   ├─ GET /health                        │
│   ├─ GET /metrics                       │
│   └─ GET /api/v1/history                │
│                                         │
│   Components                            │
│   ├─ Aggregator (state cache)           │
│   ├─ Broadcaster (push updates)         │
│   └─ Metrics (Prometheus)               │
│                                         │
└────────────┬────────────────────────────┘
             │
     ┌───────▼───────┐
     │  Redis Cache  │
     │   + Pub/Sub   │
     └───────────────┘
```

## Support

- **Documentation**: README.md
- **Full Spec**: docs/service-plans/05-dashboard-server-service.md
- **Issues**: GitHub Issues

---

**Quick Start v1.0** - Last updated: 2025-10-03
