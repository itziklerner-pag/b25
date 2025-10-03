# B25 System - Quick Reference Card

## 🚀 START THE SYSTEM
```bash
cd /home/mm/dev/b25
./run-all-services.sh
```

## 🛑 STOP THE SYSTEM
```bash
cd /home/mm/dev/b25
./stop-all-services.sh
```

## 🔍 CHECK STATUS
```bash
cd /home/mm/dev/b25
./sanity-check.sh
```

## 🌐 ACCESS FROM LOCAL MACHINE

### 1. Run SSH Tunnel (on LOCAL machine):
```bash
~/tunnel.sh
```

### 2. Open in Browser:
- **Trading Dashboard:** http://localhost:3000
- **Grafana:** http://localhost:3001 (admin / BqDocPUqSRa8lffzfuleLw==)
- **Prometheus:** http://localhost:9090

## 📊 ALL SERVICE PORTS

| Service | HTTP | gRPC | Metrics |
|---------|------|------|---------|
| Market Data | 8080 | 50051 | 9100 |
| Order Execution | 8081 | 50052 | 9101 |
| Strategy Engine | 8082 | 50053 | 9102 |
| Risk Manager | 8083 | 50054 | 9103 |
| Account Monitor | 8084 | 50055 | 9104 |
| Configuration | 8085 | 50056 | 9105 |
| Dashboard Server | 8086 | - | 9106 |
| API Gateway | 8000 | - | - |
| Auth Service | 9097 | - | - |
| Web Dashboard | 3000 | - | - |

## 📝 VIEW LOGS
```bash
# All logs
tail -f /home/mm/dev/b25/logs/*.log

# Specific service
tail -f /home/mm/dev/b25/logs/strategy-engine.log
```

## 🏥 HEALTH CHECKS
```bash
curl localhost:8080/health  # Market Data
curl localhost:8081/health  # Order Execution
curl localhost:8082/health  # Strategy Engine
curl localhost:8086/health  # Dashboard Server
```

## ⚠️ CURRENT MODE: SIMULATION
- Strategies are active and analyzing market
- Orders validated but NOT sent to exchange
- Safe for testing - no real money at risk

## 🔄 TO ENABLE LIVE TRADING
```bash
nano /home/mm/dev/b25/services/strategy-engine/config.yaml
# Change: execution_mode: live
# Restart: kill $(cat logs/strategy-engine.pid) && cd services/strategy-engine && ./bin/service &
```

## 📍 KEY FILES
- **Start:** `/home/mm/dev/b25/run-all-services.sh`
- **Stop:** `/home/mm/dev/b25/stop-all-services.sh`
- **Status:** `/home/mm/dev/b25/sanity-check.sh`
- **SSH Tunnel:** `/home/mm/dev/b25/tunnel.sh`
- **Config:** `/home/mm/dev/b25/.env`
- **Logs:** `/home/mm/dev/b25/logs/`

## ✅ SYSTEM STATUS
- 16/16 services running
- 5/5 health checks passing
- Binance testnet connected
- 3 trading strategies active
- Web dashboard operational

**STATUS: FULLY OPERATIONAL** ✅
