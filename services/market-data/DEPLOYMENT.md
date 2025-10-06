# Market Data Service - Deployment Guide

This guide explains how to deploy the market-data service on a new server with the same configuration as production.

---

## Quick Start

### One-Command Deployment

```bash
cd /path/to/b25/services/market-data
./deploy.sh
```

This automated script will:
1. Check all dependencies
2. Build the release binary
3. Create/validate configuration
4. Set up systemd service
5. Start and verify the service

---

## Prerequisites

### Required Software

- **Rust** (1.75+): `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
- **Docker**: For Redis container
- **jq**: JSON processor (`sudo apt-get install jq`)
- **curl**: HTTP client (usually pre-installed)

### Required Services

- **Redis**: Running on port 6379
  ```bash
  docker-compose -f ../../docker-compose.simple.yml up -d redis
  ```

- **Internet**: Access to `fstream.binance.com` for WebSocket data

### System Requirements

- **CPU**: 1 core minimum (service uses ~5-7%)
- **Memory**: 64MB minimum (service uses ~6-13MB)
- **Disk**: 50MB for binary + dependencies
- **Network**: Stable internet for WebSocket connection

---

## Manual Deployment (Step-by-Step)

If you prefer manual deployment or need to customize:

### Step 1: Clone Repository

```bash
git clone https://github.com/your-org/b25.git
cd b25/services/market-data
```

### Step 2: Configuration

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit configuration
vim config.yaml

# Key settings:
# - health_port: 8080 (must be available)
# - redis_url: redis://localhost:6379
# - symbols: List of trading pairs
```

### Step 3: Build Release Binary

```bash
cargo build --release

# Binary will be at: target/release/market-data-service
# Size: ~3.6MB (optimized)
```

### Step 4: Test Manually (Optional)

```bash
# Run service manually to test
./target/release/market-data-service

# In another terminal, check health
curl http://localhost:8080/health

# Check data in Redis
docker exec b25-redis redis-cli GET "market_data:BTCUSDT"

# Stop with Ctrl+C
```

### Step 5: Install Systemd Service

```bash
# Copy service file template
sudo cp market-data.service /etc/systemd/system/

# Or create it:
sudo nano /etc/systemd/system/market-data.service
# (Paste contents from market-data.service)

# Reload systemd
sudo systemctl daemon-reload

# Enable on boot
sudo systemctl enable market-data

# Start service
sudo systemctl start market-data
```

### Step 6: Verify

```bash
# Check status
sudo systemctl status market-data

# View logs
sudo journalctl -u market-data -f

# Test health endpoint
curl http://localhost:8080/health

# Check data
docker exec b25-redis redis-cli GET "market_data:BTCUSDT" | jq
```

---

## Deployment Script Details

### What deploy.sh Does

**Pre-deployment Checks:**
- ✓ Verifies Rust, Docker, jq, curl installed
- ✓ Checks Redis container is running
- ✓ Tests internet connectivity to Binance
- ✓ Stops any existing service instances

**Build Process:**
- ✓ Builds optimized release binary
- ✓ Verifies binary exists and reports size

**Configuration:**
- ✓ Creates config.yaml from example if missing
- ✓ Validates required config fields
- ✓ Adds missing fields (e.g., exchange_rest_url)

**Systemd Setup:**
- ✓ Generates systemd service file
- ✓ Installs to /etc/systemd/system/
- ✓ Enables service for boot
- ✓ Configures resource limits (CPU 50%, Memory 512M)
- ✓ Sets up logging to systemd journal

**Startup & Verification:**
- ✓ Starts service via systemd
- ✓ Waits for startup
- ✓ Checks service is active
- ✓ Verifies process running
- ✓ Tests health endpoint
- ✓ Confirms data flowing to Redis
- ✓ Reports resource usage

### Script Output

```
================================
Market Data Service Deployment
================================

================================
Step 1: Pre-deployment Checks
================================
ℹ Checking required dependencies...
✓ cargo is installed
✓ docker is installed
✓ jq is installed
✓ curl is installed
✓ Redis is running
✓ Internet connectivity OK

================================
Step 2: Build Service
================================
ℹ Building release binary (this may take a few minutes)...
✓ Build successful (binary size: 3.6M)

================================
Step 3: Configuration
================================
✓ config.yaml exists
✓ Configuration validated

================================
Step 4: Systemd Service Setup
================================
✓ Systemd service installed
✓ Service enabled

================================
Step 5: Start Service
================================
✓ Service started successfully

================================
Step 6: Verification
================================
✓ Service is active
✓ Process running (PID: 60188)
✓ Health endpoint responding on port 8080
✓ Data flowing to Redis (BTC: $123542.95)
✓ CPU: 5.7% | Memory: 0.2%

================================
Deployment Summary
================================

✓ Deployment successful!

Service Details:
  • Name: market-data
  • Binary: /home/mm/dev/b25/services/market-data/target/release/market-data-service
  • Config: /home/mm/dev/b25/services/market-data/config.yaml
  • Systemd: /etc/systemd/system/market-data.service
  • User: mm

Management Commands:
  • Status: sudo systemctl status market-data
  • Logs: sudo journalctl -u market-data -f
  • Restart: sudo systemctl restart market-data
  • Stop: sudo systemctl stop market-data

Health Check:
  • curl http://localhost:8080/health

✓ Deployment complete!
```

---

## Configuration Options

### Environment Variables

The deploy script supports these environment variables:

```bash
# Deploy as different user
DEPLOY_USER=market-data ./deploy.sh

# Custom Rust flags
RUSTFLAGS="-C target-cpu=native" ./deploy.sh
```

### Config File Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `symbols` | BTCUSDT, ETHUSDT, BNBUSDT, SOLUSDT | Trading pairs to track |
| `exchange_ws_url` | wss://fstream.binance.com/stream | Binance WebSocket URL |
| `exchange_rest_url` | https://fapi.binance.com | Binance REST API |
| `redis_url` | redis://localhost:6379 | Redis connection string |
| `order_book_depth` | 20 | Price levels per side |
| `health_port` | 8080 | HTTP server port |
| `reconnect_delay_ms` | 1000 | Initial reconnect delay |
| `max_reconnect_delay_ms` | 60000 | Max reconnect delay |

### Systemd Resource Limits

Configured in `market-data.service`:

- **CPUQuota**: 50% (max half of one CPU core)
- **MemoryLimit**: 512M (hard limit, OOM kill if exceeded)
- **MemoryHigh**: 256M (soft limit, throttle if exceeded)
- **TasksMax**: 100 (max threads/processes)

---

## Multi-Server Deployment

### Deploy to Multiple Servers

```bash
# Deploy to production
ssh prod-server1 'cd /opt/b25/services/market-data && git pull && ./deploy.sh'

# Deploy to staging
ssh staging-server 'cd /opt/b25/services/market-data && git pull && ./deploy.sh'

# Deploy to backup
ssh backup-server 'cd /opt/b25/services/market-data && git pull && ./deploy.sh'
```

### Ansible Playbook (Future)

```yaml
# playbook.yml
- hosts: trading_servers
  tasks:
    - name: Clone repository
      git:
        repo: https://github.com/your-org/b25.git
        dest: /opt/b25

    - name: Run deployment script
      command: ./deploy.sh
      args:
        chdir: /opt/b25/services/market-data
```

---

## Troubleshooting

### Deployment Fails: "cargo not found"

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

### Deployment Fails: "Redis not running"

```bash
# Start Redis container
cd /path/to/b25
docker-compose -f docker-compose.simple.yml up -d redis

# Or start manually
docker run -d --name b25-redis -p 6379:6379 redis:7-alpine
```

### Deployment Fails: "Port 8080 in use"

```bash
# Find what's using port 8080
sudo lsof -i :8080

# Option 1: Change health_port in config.yaml
vim config.yaml
# Change health_port to 8081 or another available port

# Option 2: Stop the other service
sudo systemctl stop other-service
```

### Service Won't Start

```bash
# Check logs
sudo journalctl -u market-data -n 50

# Common issues:
# 1. Config file missing or invalid
ls -la config.yaml
cat config.yaml

# 2. Binary permissions
chmod +x target/release/market-data-service

# 3. Redis not accessible
docker exec b25-redis redis-cli PING
```

### No Data in Redis

```bash
# Check service logs
sudo journalctl -u market-data -f

# Check Binance connectivity
curl -I https://fstream.binance.com

# Verify Redis connection
docker exec b25-redis redis-cli PING

# Wait 30 seconds for initial data
# First WebSocket messages can take time
```

---

## Updating

### Update to Latest Code

```bash
cd /path/to/b25
git pull
cd services/market-data
./deploy.sh
```

The deployment script automatically:
1. Stops the old service
2. Rebuilds with latest code
3. Restarts with new binary
4. Verifies everything works

### Rollback

```bash
# If update fails, rollback to previous version
git log --oneline -5
git checkout <previous-commit-hash>
./deploy.sh
```

---

## Uninstalling

### Remove Service

```bash
# Stop service
sudo systemctl stop market-data

# Disable from boot
sudo systemctl disable market-data

# Remove systemd file
sudo rm /etc/systemd/system/market-data.service
sudo systemctl daemon-reload

# Remove binary and build artifacts
cd /path/to/b25/services/market-data
cargo clean
```

### Or Use Uninstall Script

```bash
./uninstall.sh
```

---

## Production Checklist

Before deploying to production:

- [ ] Test deployment on staging server first
- [ ] Verify config.yaml is correct (especially ports)
- [ ] Ensure Redis is running and accessible
- [ ] Confirm internet access to Binance
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure alerting (CPU > 30%, Memory > 400MB)
- [ ] Test failover (systemd auto-restart)
- [ ] Document server IP/hostname
- [ ] Set up log rotation
- [ ] Test manual restart
- [ ] Verify backup/redundancy plan

---

## Files Included

```
market-data/
├── deploy.sh                   # Main deployment script
├── uninstall.sh               # Uninstall script
├── config.yaml                # Runtime configuration (gitignored)
├── config.example.yaml        # Configuration template
├── market-data.service        # Systemd service file template
├── DEPLOYMENT.md              # This file
├── README.md                  # Service documentation
├── Cargo.toml                 # Rust dependencies
├── src/                       # Source code
└── target/                    # Build artifacts (gitignored)
```

---

## Support

### Get Help

- **Logs**: `sudo journalctl -u market-data -f`
- **Status**: `sudo systemctl status market-data`
- **Health**: `curl http://localhost:8080/health`
- **Metrics**: `curl http://localhost:8080/metrics`

### Common Commands

```bash
# Restart service
sudo systemctl restart market-data

# View logs (live)
sudo journalctl -u market-data -f

# View logs (last 100 lines)
sudo journalctl -u market-data -n 100

# Check resource usage
systemctl status market-data | grep -E "Memory|CPU"

# Test connectivity
curl http://localhost:8080/health
docker exec b25-redis redis-cli GET "market_data:BTCUSDT" | jq
```

---

## Next Steps

After successful deployment:

1. **Set up monitoring**: Configure Prometheus to scrape `/metrics`
2. **Configure alerts**: Alert on high CPU, memory, or service down
3. **Test failover**: Kill the process, verify auto-restart
4. **Document**: Record server details in your ops documentation
5. **Deploy other services**: Use similar pattern for other microservices

---

**Deployment Guide Version:** 1.0
**Last Updated:** 2025-10-06
