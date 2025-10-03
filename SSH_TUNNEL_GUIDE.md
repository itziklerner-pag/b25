# SSH Tunnel Guide for Remote VPS Access

## 🚇 Quick Start

### On Your LOCAL Machine (laptop/desktop):

```bash
# Download the tunnel script from VPS
scp <username>@<vps-ip>:/home/mm/dev/b25/ssh-tunnel.sh ~/ssh-tunnel.sh
chmod +x ~/ssh-tunnel.sh

# Run the tunnel
~/ssh-tunnel.sh <username>@<vps-ip>
```

### Example:
```bash
# If your VPS is at 192.168.1.100 with user 'root'
scp root@192.168.1.100:/home/mm/dev/b25/ssh-tunnel.sh ~/ssh-tunnel.sh
chmod +x ~/ssh-tunnel.sh
~/ssh-tunnel.sh root@192.168.1.100
```

---

## 📋 Manual SSH Command (Copy-Paste)

If you prefer to run the command directly without the script:

```bash
ssh -N \
  -L 3000:localhost:3000 \
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
  <username>@<vps-ip>
```

**Replace `<username>@<vps-ip>` with your actual VPS credentials**

Example:
```bash
ssh -N -L 3000:localhost:3000 ... root@192.168.1.100
```

---

## 🌐 Access URLs (After Tunnel is Running)

### Main Applications
- **Web Dashboard:** http://localhost:3000
- **Grafana:** http://localhost:3001 (admin / BqDocPUqSRa8lffzfuleLw==)
- **Prometheus:** http://localhost:9090
- **Alertmanager:** http://localhost:9093
- **API Gateway:** http://localhost:8000

### Service APIs (HTTP)
- **Market Data:** http://localhost:8080
- **Order Execution:** http://localhost:8081
- **Strategy Engine:** http://localhost:8082
- **Risk Manager:** http://localhost:8083
- **Account Monitor:** http://localhost:8084
- **Configuration:** http://localhost:8085
- **Dashboard Server:** http://localhost:8086

### Metrics Endpoints (Prometheus)
- **Market Data Metrics:** http://localhost:9100/metrics
- **Order Execution Metrics:** http://localhost:9101/metrics
- **Strategy Engine Metrics:** http://localhost:9102/metrics
- **Risk Manager Metrics:** http://localhost:9103/metrics
- **Account Monitor Metrics:** http://localhost:9104/metrics
- **Configuration Metrics:** http://localhost:9105/metrics
- **Dashboard Server Metrics:** http://localhost:9106/metrics

### gRPC Endpoints
- **Market Data gRPC:** localhost:50051
- **Order Execution gRPC:** localhost:50052
- **Strategy Engine gRPC:** localhost:50053
- **Risk Manager gRPC:** localhost:50054
- **Account Monitor gRPC:** localhost:50055
- **Configuration gRPC:** localhost:50056

---

## 🔧 Alternative: Port-by-Port Tunnels

If you only need specific services:

### Essential (Web + Monitoring)
```bash
ssh -N -L 3000:localhost:3000 -L 3001:localhost:3001 -L 9090:localhost:9090 <user>@<vps-ip>
```

### Trading APIs Only
```bash
ssh -N -L 8080:localhost:8080 -L 8081:localhost:8081 -L 8082:localhost:8082 <user>@<vps-ip>
```

### Metrics Only
```bash
ssh -N -L 9100:localhost:9100 -L 9101:localhost:9101 -L 9102:localhost:9102 <user>@<vps-ip>
```

---

## 🔒 SSH Config (Optional - One-time Setup)

Add to `~/.ssh/config` on your LOCAL machine:

```
Host b25-vps
    HostName <vps-ip>
    User <username>
    LocalForward 3000 localhost:3000
    LocalForward 3001 localhost:3001
    LocalForward 8000 localhost:8000
    LocalForward 8080 localhost:8080
    LocalForward 8081 localhost:8081
    LocalForward 8082 localhost:8082
    LocalForward 8083 localhost:8083
    LocalForward 8084 localhost:8084
    LocalForward 8085 localhost:8085
    LocalForward 8086 localhost:8086
    LocalForward 9090 localhost:9090
    LocalForward 9093 localhost:9093
    LocalForward 9100 localhost:9100
    LocalForward 9101 localhost:9101
    LocalForward 9102 localhost:9102
    LocalForward 9103 localhost:9103
    LocalForward 9104 localhost:9104
    LocalForward 9105 localhost:9105
    LocalForward 9106 localhost:9106
    LocalForward 50051 localhost:50051
    LocalForward 50052 localhost:50052
    LocalForward 50053 localhost:50053
    LocalForward 50054 localhost:50054
    LocalForward 50055 localhost:50055
    LocalForward 50056 localhost:50056
```

Then simply run:
```bash
ssh -N b25-vps
```

---

## 🐛 Troubleshooting

### Tunnel won't start
```bash
# Check if ports are already in use
lsof -i :3000
lsof -i :8080

# Kill existing processes if needed
kill -9 <PID>
```

### Connection drops
```bash
# Add keep-alive to SSH command
ssh -N -o ServerAliveInterval=60 -o ServerAliveCountMax=3 -L 3000:localhost:3000 ... <user>@<vps-ip>
```

### Can't access services
1. Verify tunnel is running (check terminal - it should hang, not exit)
2. Check VPS services are running: `docker-compose ps`
3. Verify no firewall blocking local ports

---

## 💡 Pro Tips

1. **Run tunnel in background:**
   ```bash
   ssh -fN -L 3000:localhost:3000 ... <user>@<vps-ip>
   ```

2. **Auto-restart on disconnect:**
   ```bash
   while true; do
       ssh -N -L 3000:localhost:3000 ... <user>@<vps-ip>
       echo "Reconnecting in 5 seconds..."
       sleep 5
   done
   ```

3. **Use tmux/screen on VPS:**
   ```bash
   # On VPS
   tmux new -s b25
   ./scripts/dev-start.sh
   # Ctrl+B, then D to detach
   ```

---

## 📝 Port Reference

| Service | HTTP | Metrics | gRPC |
|---------|------|---------|------|
| Market Data | 8080 | 9100 | 50051 |
| Order Execution | 8081 | 9101 | 50052 |
| Strategy Engine | 8082 | 9102 | 50053 |
| Risk Manager | 8083 | 9103 | 50054 |
| Account Monitor | 8084 | 9104 | 50055 |
| Configuration | 8085 | 9105 | 50056 |
| Dashboard Server | 8086 | 9106 | - |
| API Gateway | 8000 | - | - |
| Web Dashboard | 3000 | - | - |
| Grafana | 3001 | - | - |
| Prometheus | 9090 | - | - |
| Alertmanager | 9093 | - | - |

---

**Keep the SSH tunnel running while accessing the services!**
