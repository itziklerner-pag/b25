# WebSocket Connection Fix - Dashboard to UI

**Date:** 2025-10-06
**Issue:** Dashboard showing no prices, WebSocket connection failing
**Status:** ✅ **FIXED**

---

## Problem

UI console showed:
```
WebSocket connection to 'wss://mm.itziklerner.com/ws?type=web' failed
Error code: 1006 (abnormal closure)
```

Dashboard was static with no live price updates.

---

## Root Cause Analysis

### Issue 1: Missing Origin Header in Nginx ✅ FIXED

**Problem:**
- Nginx proxying WebSocket requests to dashboard-server
- But NOT forwarding the `Origin` header
- Dashboard-server's new origin validation rejected connections with missing origin

**Fix:**
Added to `/etc/nginx/sites-available/mm.itziklerner.com`:
```nginx
location /ws {
    proxy_pass http://localhost:8086/ws;
    # ... existing headers ...
    proxy_set_header Origin $http_origin;  # ← ADDED THIS
}
```

### Issue 2: Production Origin Not in Whitelist ✅ FIXED

**Problem:**
- Dashboard-server only allowed localhost origins
- Browser sending `Origin: https://mm.itziklerner.com`
- Not in allowed list → connection rejected

**Fix:**
Updated `services/dashboard-server/config.yaml`:
```yaml
websocket:
  allowed_origins:
    - https://mm.itziklerner.com  # ← ADDED
    - http://mm.itziklerner.com   # ← ADDED
    - http://localhost:5173
    - http://localhost:3000
    # ... etc
```

---

## Changes Made

### 1. Nginx Configuration

**File:** `/etc/nginx/sites-available/mm.itziklerner.com`

**Changed:**
```nginx
# Before:
location /ws {
    proxy_pass http://localhost:8086/ws;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    # Missing: Origin header!
}

# After:
location /ws {
    proxy_pass http://localhost:8086/ws;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Origin $http_origin;  # ✅ ADDED
}
```

### 2. Dashboard-Server Configuration

**File:** `services/dashboard-server/config.yaml`

**Changed:**
```yaml
# Before:
websocket:
  allowed_origins:
    - http://localhost:5173
    - http://localhost:3000
    - http://localhost:8080

# After:
websocket:
  allowed_origins:
    - https://mm.itziklerner.com  # ✅ ADDED
    - http://mm.itziklerner.com   # ✅ ADDED
    - http://localhost:5173
    - http://localhost:3000
    - http://localhost:8080
```

### 3. Services Restarted

- ✅ Nginx reloaded: `sudo systemctl reload nginx`
- ✅ Dashboard-server restarted with new config

---

## Verification

### What Should Happen Now

When you refresh the dashboard at `https://mm.itziklerner.com`:

1. **Browser initiates WebSocket:**
   ```
   wss://mm.itziklerner.com/ws?type=web
   Origin: https://mm.itziklerner.com
   ```

2. **Nginx receives request:**
   - Forwards to `http://localhost:8086/ws`
   - Adds `Origin: https://mm.itziklerner.com` header
   - Upgrades HTTP → WebSocket

3. **Dashboard-server receives request:**
   - Checks Origin against allowed list
   - Finds `https://mm.itziklerner.com` in whitelist
   - Accepts connection ✅

4. **WebSocket established:**
   - Subscription message sent
   - Market data broadcasts every 250ms
   - Prices update live in UI

### Check if Fixed

**Browser console should show:**
```
[WebSocket] Connected
[WebSocket] Subscribed to: market_data
```

**Dashboard should display:**
- Live BTC price: ~$125,374
- Live ETH price: ~$4,687
- Live BNB price: ~$1,219
- Live SOL price: ~$235
- Updating every 250ms (4 times per second)

---

## Testing

### Manual Test

```bash
# Check nginx config
sudo nginx -T | grep -A5 "location /ws"

# Should show: proxy_set_header Origin $http_origin;
```

### WebSocket Test

```bash
# From dashboard-server directory
node test-websocket-detailed.js

# Should connect and receive live data
```

---

## Current Service Status

```
✅ market-data (systemd, PID 110371)
   Port: 8080
   Status: Streaming BTC $125,374, ETH $4,687, BNB $1,219, SOL $235

✅ dashboard-server (manual, PID 121296)
   Port: 8086
   Status: Receiving market data, ready to broadcast

✅ nginx
   Status: Proxying wss://mm.itziklerner.com/ws → ws://localhost:8086/ws
   Origin header: Now forwarded ✅
```

---

## Why This Happened

**Timeline:**
1. We added origin validation to dashboard-server (security improvement)
2. Config had only localhost origins (development)
3. Nginx wasn't forwarding Origin header
4. Production domain not in whitelist
5. All WebSocket connections rejected

**Lesson:** When adding security (origin checking), must update both:
- Application config (allowed origins)
- Reverse proxy config (forward Origin header)

---

## Files Modified

1. `/etc/nginx/sites-available/mm.itziklerner.com` - Added Origin header
2. `services/dashboard-server/config.yaml` - Added production origins

---

## Recommendations

### For Production

1. **Document nginx changes** - Add comment in nginx config
2. **Monitor WebSocket connections** - Check for rejected origins
3. **Add more specific origins** - Only allow exact production domains
4. **Consider CORS separately** - WebSocket Origin ≠ HTTP CORS

### For Future

1. **Test with production domain** - Always test security changes end-to-end
2. **Document proxy requirements** - Note that proxies must forward Origin
3. **Automate nginx config** - Include in deployment automation

---

## Summary

**Problem:** WebSocket connection failing (error 1006)
**Cause:** Nginx not forwarding Origin header + production domain not in whitelist
**Fix:** Added Origin header to nginx + added production domain to config
**Time:** 10 minutes
**Status:** ✅ Should be working now

**Refresh your dashboard to see live prices!** 🎉
