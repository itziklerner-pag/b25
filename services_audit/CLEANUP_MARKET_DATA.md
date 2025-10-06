# Market Data Service - Process Cleanup Report

**Date:** 2025-10-06
**Issue:** Multiple instances running, one consuming 94% CPU

---

## Problem Summary

### Current State (3 instances running)

| PID | Type | Start | CPU | Memory | Port 8080 | Status |
|-----|------|-------|-----|--------|-----------|--------|
| 34056 | **Release** | 01:27 (5.5h ago) | **94.3%** ⚠️ | 29MB | ❌ No | **RUNAWAY** |
| 46880 | Debug | 03:34 (3.7h ago) | 6.9% | 25MB | ✅ Yes | Working |
| 50071 | Debug | 04:38 (2.7h ago) | 6.4% | 25MB | ❌ No | Unknown |

### Issues Identified

1. **🔴 CRITICAL: Runaway Process**
   - PID 34056 (release build) consuming 94% CPU
   - Has burned 5.5 hours of CPU time
   - NOT serving requests (port 8080 not bound)
   - Likely stuck in infinite loop or thrashing

2. **🟡 Redundant Instances**
   - Two processes not serving HTTP (34056, 50071)
   - Only PID 46880 is actually handling /health requests
   - Wasting resources (~60MB total)

3. **⚠️ Debug Build in Production**
   - The working instance (46880) is a debug build
   - Debug builds are slower and use more memory
   - Should use optimized release build instead

---

## Recommended Action Plan

### Immediate (Do Now)

**Step 1: Kill the runaway release build**
```bash
kill 34056

# If it doesn't stop:
kill -9 34056
```
**Impact:** Frees 94% CPU immediately

**Step 2: Kill redundant debug build**
```bash
kill 50071

# If needed:
kill -9 50071
```
**Impact:** Frees ~25MB memory, reduces confusion

**Step 3: Verify only one instance remains**
```bash
ps aux | grep market-data-service | grep -v grep
# Should show only PID 46880
```

### Short-term (Within 1 hour)

**Step 4: Build and start optimized release version**
```bash
cd /home/mm/dev/b25/services/market-data

# Build release (with optimizations)
cargo build --release

# Kill debug build
kill 46880

# Start release build in background
nohup ./target/release/market-data-service > /tmp/market-data.log 2>&1 &

# Verify
ps aux | grep market-data-service
curl http://localhost:8080/health
```

**Step 5: Monitor for 5 minutes**
```bash
# Check CPU usage
watch -n 1 'ps aux | grep market-data-service | grep -v grep'

# Should show <10% CPU (normal is 5-7%)
```

### Long-term (This Week)

**Step 6: Set up proper service management**

Create systemd service to prevent multiple instances:
```bash
sudo nano /etc/systemd/system/market-data.service
```

```ini
[Unit]
Description=Market Data Service
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=mm
WorkingDirectory=/home/mm/dev/b25/services/market-data
ExecStart=/home/mm/dev/b25/services/market-data/target/release/market-data-service
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

# Resource limits
CPUQuota=50%
MemoryLimit=512M

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable market-data
sudo systemctl start market-data
sudo systemctl status market-data
```

---

## Why This Happened

### Possible Causes

1. **Manual Testing**
   - You or someone ran the service multiple times for testing
   - Forgot to kill previous instances before starting new ones

2. **Build Script Issues**
   - Build/test scripts may have left processes running
   - No cleanup after cargo run/test

3. **No Process Management**
   - Services started manually (not via systemd)
   - No automatic cleanup of old instances
   - No resource limits enforced

4. **Runaway Process Root Cause**
   - Possible infinite reconnection loop (Binance connection issue?)
   - Memory leak causing thrashing
   - Bug in older build (release was started 5.5h ago)

---

## Prevention Measures

### Immediate Prevention

**1. Always check before starting:**
```bash
# Before running service:
ps aux | grep market-data-service | grep -v grep

# Kill existing instances:
pkill -f market-data-service
```

**2. Use systemd** (prevents multiple instances automatically)

**3. Monitor CPU usage:**
```bash
# Set up alert if CPU > 50% for 5 minutes
# (Add to monitoring system)
```

### Long-term Prevention

**1. Resource Limits**
- Add CPU limits (systemd CPUQuota=50%)
- Add memory limits (systemd MemoryLimit=512M)
- Auto-restart if limits exceeded

**2. Health Monitoring**
- Monitor CPU usage in Prometheus
- Alert if CPU > 30% sustained
- Auto-restart if unhealthy

**3. Docker Containers**
- Run in Docker with `--cpus=1.0` limit
- Docker prevents multiple instances on same port
- Better isolation and resource control

---

## Cleanup Commands (Copy-Paste Ready)

### Quick Cleanup (30 seconds)
```bash
# Kill all market-data instances
pkill -f market-data-service

# Wait 2 seconds
sleep 2

# Verify all killed
ps aux | grep market-data-service | grep -v grep
# Should return nothing

# Start fresh release build
cd /home/mm/dev/b25/services/market-data
nohup ./target/release/market-data-service > /tmp/market-data.log 2>&1 &

# Verify single instance
ps aux | grep market-data-service | grep -v grep

# Test health
curl http://localhost:8080/health
```

### Safe Cleanup (2 minutes - recommended)
```bash
# 1. Kill runaway process first
kill 34056
sleep 2

# 2. Kill redundant debug
kill 50071
sleep 2

# 3. Gracefully stop working debug
kill 46880  # Sends SIGTERM (graceful)
sleep 5

# 4. Verify all stopped
ps aux | grep market-data-service | grep -v grep

# 5. Start release build
cd /home/mm/dev/b25/services/market-data
./target/release/market-data-service &

# 6. Verify
sleep 2
curl http://localhost:8080/health
docker exec b25-redis redis-cli GET "market_data:BTCUSDT"
```

---

## Post-Cleanup Verification

### Checklist

- [ ] Only 1 instance running
- [ ] CPU usage < 10%
- [ ] Memory usage ~25-30MB
- [ ] Port 8080 responding
- [ ] Redis data updating
- [ ] Health endpoint returns healthy
- [ ] No error logs

### Verification Commands

```bash
# 1. Process count
ps aux | grep market-data-service | grep -v grep | wc -l
# Expected: 1

# 2. CPU usage
ps aux | grep market-data-service | grep -v grep | awk '{print $3}'
# Expected: < 10.0

# 3. Memory
ps aux | grep market-data-service | grep -v grep | awk '{print $6}'
# Expected: ~25000-30000 (KB)

# 4. Health
curl -s http://localhost:8080/health | jq .
# Expected: {"service":"market-data","status":"healthy","version":"0.1.0"}

# 5. Data flowing
docker exec b25-redis redis-cli GET "market_data:BTCUSDT" | jq .last_price
# Expected: Current BTC price

# 6. Update frequency
timeout 5 docker exec b25-redis redis-cli --csv SUBSCRIBE "market_data:BTCUSDT" | wc -l
# Expected: 25-50 lines (5-10 updates/sec)
```

---

## Decision: Debug vs Release Build

### Debug Build Characteristics
- **Pros:**
  - Better error messages
  - Easier debugging with gdb/lldb
  - More detailed logs
- **Cons:**
  - **30-50% slower** than release
  - **2x larger binary** (150MB vs 8MB)
  - **Higher memory usage**
  - No optimizations

### Release Build Characteristics
- **Pros:**
  - **Optimized** (opt-level=3, LTO enabled)
  - **Faster** (30-50% performance improvement)
  - **Smaller** (8MB binary, stripped symbols)
  - **Lower CPU/memory** usage
- **Cons:**
  - Harder to debug
  - Less detailed panic messages

### Recommendation: **USE RELEASE BUILD**

For production/development runtime, always use release:
```bash
cargo build --release
./target/release/market-data-service
```

Only use debug for:
- Active development with breakpoints
- Investigating crashes
- Running tests

---

## Root Cause Analysis: Why 94% CPU?

### Investigation Needed

The runaway process (PID 34056) needs investigation:

**Hypothesis 1: Infinite Reconnection Loop**
- Binance connection failed
- Exponential backoff bug
- Keeps trying to reconnect without delay

**Hypothesis 2: Memory Thrashing**
- Memory leak filled RAM
- Constant garbage collection
- Swapping to disk

**Hypothesis 3: Deadlock/Livelock**
- Thread deadlock in Tokio runtime
- Spinning on lock acquisition
- CPU-bound infinite loop

### How to Investigate (if it happens again)

```bash
# 1. Check logs first
journalctl -u market-data -n 1000 | tail -100

# 2. Sample CPU usage
top -p 34056 -b -n 10 -d 1

# 3. Thread analysis
ps -T -p 34056

# 4. System calls (requires sudo)
sudo strace -c -p 34056 -f

# 5. Memory usage
pmap 34056 | tail

# 6. Stack trace (if compiled with debug symbols)
gdb -p 34056 -batch -ex "thread apply all bt"
```

---

## Summary

**Current State:**
- ❌ 3 instances running (should be 1)
- ❌ Runaway process at 94% CPU
- ❌ Debug build serving production traffic
- ✅ Data is flowing (despite the mess)

**Immediate Action Required:**
1. Kill PIDs 34056 and 50071
2. Replace debug build with release build
3. Verify single instance with normal CPU usage

**Estimated Time:** 2-5 minutes
**Risk:** Low (service will restart quickly)
**Benefit:** 94% CPU reduction + better performance

---

**Ready to proceed with cleanup?**
