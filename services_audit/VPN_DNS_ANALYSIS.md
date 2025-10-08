# VPN DNS Issue - Analysis

**Date:** 2025-10-06
**Issue:** DNS stops working when VPN activated
**Status:** 🔍 **ANALYZED**

---

## Problem Summary

When activating the WireGuard VPN (`wg-quick up wg0`):
- ✅ SSH still works (split-tunnel working correctly)
- ❌ DNS resolution stops working
- ❌ HTTPS requests hang/timeout
- ⚠️ VPN interface crashes or shuts down

---

## Root Cause Analysis

### 1. VPN Config Has DNS Servers

**VPN Config (`peer48-binance-only.conf`):**
```
DNS = 161.97.189.51,161.97.189.52
```

**What happens during VPN activation:**
```bash
[#] resolvconf -a wg0 -m 0 -x
```

This command tries to modify system DNS resolution to use VPN's DNS servers.

### 2. DNS Configuration Conflict

**System uses:** systemd-resolved (127.0.0.53)
**VPN tries to set:** 161.97.189.51, 161.97.189.52

**Conflict:**
- systemd-resolved is managing DNS
- resolvconf tries to override it
- Systems fight each other
- DNS resolution breaks

### 3. VPN DNS Servers Unreachable

**VPN DNS servers:** 161.97.189.51, 161.97.189.52

**Problem:**
- These might only be accessible THROUGH the VPN tunnel
- But split-tunnel only routes Binance IPs (76.223.0.0/16, 54.76.0.0/16, 99.81.0.0/16)
- DNS server IPs (161.97.189.0/24) NOT in AllowedIPs
- **DNS servers unreachable** → All DNS fails

### 4. VPN Crashes After Activation

**Evidence:**
```bash
# Right after activation:
sudo wg show  # Shows VPN running

# Shortly after:
ip link show wg0  # Device "wg0" does not exist
```

**Why it crashes:**
- DNS resolution fails
- VPN can't reach its own DNS servers
- VPN keepalive or peer communication fails
- WireGuard shuts down the interface

---

## Technical Details

### Routing Analysis

**AllowedIPs in VPN config:**
```
76.223.0.0/16   # Binance range 1
54.76.0.0/16    # Binance range 2
99.81.0.0/16    # Binance range 3
```

**VPN DNS servers:**
```
161.97.189.51   # NOT in AllowedIPs
161.97.189.52   # NOT in AllowedIPs
```

**Result:**
- Packets to Binance IPs → VPN tunnel ✅
- Packets to DNS servers → Normal route ❌ (but DNS servers expect VPN tunnel)
- **DNS servers unreachable** = DNS failure

### systemd-resolved Behavior

**Current config:**
```
/etc/resolv.conf → symlink to /run/systemd/resolve/stub-resolv.conf
nameserver 127.0.0.53  # systemd-resolved local stub
```

**When resolvconf runs:**
- Tries to add wg0 DNS servers to systemd-resolved
- But DNS servers unreachable
- systemd-resolved marks them as failed
- Might affect overall DNS resolution

---

## Why DNS Appears to Work Now

**Current test results:**
- ✅ `nslookup google.com` - Works
- ✅ `nslookup fapi.binance.com` - Works
- ✅ `ping 8.8.8.8` - Works
- ❌ `curl https://fapi.binance.com` - Hangs/timeout

**Reason:**
- VPN crashed and wg0 interface removed
- System reverted to normal DNS (209.126.70.51, 209.126.70.52)
- DNS working again because VPN is down
- But Binance API still geo-blocked (no VPN active)

---

## The Circular Problem

**Catch-22 situation:**

1. **Need VPN for Binance access** → But VPN has DNS servers
2. **VPN sets DNS servers** → But DNS servers not in AllowedIPs
3. **DNS servers unreachable** → VPN can't communicate
4. **VPN crashes** → Back to no Binance access

---

## Why Market-Data Works Without VPN

**From code analysis (`websocket.rs:112-113`):**
```rust
// Skip REST snapshot fetch (geo-blocked) - build orderbook from WebSocket
info!("Building orderbook for {} from WebSocket updates (REST API geo-blocked)");
```

**Binance WebSocket for Public Data:**
- ❌ REST API: Geo-blocked (HTTP 451)
- ✅ **WebSocket: NOT geo-blocked** (works without VPN!)

**Why it's different:**
- Public market data WebSocket: `wss://fstream.binance.com/stream` → NOT geo-restricted
- Private authenticated endpoints: REST API + User Data Stream → GEO-RESTRICTED

---

## Why Account-Monitor Needs VPN

**Account-monitor tries to access:**
1. **Binance Futures REST API** - Account balance, positions
   - Endpoint: `https://fapi.binance.com/fapi/v2/account`
   - Status: **GEO-BLOCKED** ❌

2. **Binance User Data Stream WebSocket** - Real-time account updates
   - Endpoint: `wss://fstream.binance.com/ws/{listenKey}`
   - Status: **GEO-BLOCKED** ❌

Both require authentication + both are geo-restricted.

---

## Solutions (Not Implemented Yet)

### Option 1: Fix VPN DNS Issue ⭐ RECOMMENDED

**Problem:** DNS servers not routable through split-tunnel

**Fix:** Remove DNS from VPN config or add DNS IPs to AllowedIPs
```
# Remove this line:
DNS = 161.97.189.51,161.97.189.52

# Or add DNS servers to AllowedIPs:
AllowedIPs = 76.223.0.0/16, 54.76.0.0/16, 99.81.0.0/16, 161.97.189.0/24
```

**Result:**
- VPN stays up
- Uses system DNS (not VPN DNS)
- SSH stays alive
- Binance accessible

### Option 2: Use Binance Testnet ⭐ EASIER

**No geo-restrictions on testnet:**
- `https://testnet.binancefuture.com` - Works from anywhere
- No VPN needed
- Free test funds
- Perfect for development

**Update account-monitor config:**
```yaml
exchange:
  rest_url: "https://testnet.binancefuture.com"
  ws_url: "wss://stream.binancefuture.com/ws"
```

### Option 3: Full-Tunnel VPN with Systemd Integration

**Use systemd to manage VPN:**
- Proper dependency management
- Auto-restart on failure
- Better DNS integration
- But loses split-tunnel benefit (might need work on DNS)

---

## Current State Summary

**VPN Status:** ❌ Down (crashed after brief activation)
**SSH:** ✅ Working (split-tunnel protected it)
**DNS:** ✅ Working now (VPN down, reverted to normal)
**Binance Access:** ❌ Geo-blocked (no active VPN)
**Market-Data:** ✅ Working (WebSocket not geo-blocked)
**Account-Monitor:** ❌ Failing (needs authenticated endpoints)

---

## Answers to Your Question

**Q: When running VPN, SSH still works but DNS stops working - why?**

**A: The VPN config has DNS servers (161.97.189.51/52) that:**
1. Are NOT included in the split-tunnel AllowedIPs routes
2. Can't be reached through normal routing (expect VPN tunnel)
3. Can't be reached through VPN tunnel (not in AllowedIPs)
4. Become unreachable → DNS fails
5. VPN crashes due to unreachable DNS
6. System DNS confused/broken until VPN fully removed

**The fix:** Either remove DNS line from VPN config, or use system DNS instead of VPN DNS.

---

**VPN Infrastructure:** ✅ Exists and partially working
**DNS Issue:** 🔴 Critical blocker preventing VPN stability
**Recommendation:** Remove `DNS = ...` line from VPN config to fix

