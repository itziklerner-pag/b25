# VPN Routing Issue - Root Cause Analysis

**Date:** 2025-10-06
**Issue:** VPN up, DNS working, but Binance still geo-blocked
**Status:** 🔍 **ROOT CAUSE IDENTIFIED**

---

## Problem Summary

✅ **VPN Status:** UP and stable (wg0 interface active)
✅ **SSH:** Working perfectly
✅ **DNS:** Working perfectly (fixed by removing DNS line)
❌ **Binance Access:** Still geo-blocked

**Why:** Binance traffic is NOT being routed through VPN!

---

## Root Cause

### VPN AllowedIPs vs Actual Binance IPs

**VPN Config (AllowedIPs):**
```
76.223.0.0/16   # Range 1
54.76.0.0/16    # Range 2
99.81.0.0/16    # Range 3
```

**Actual Binance IP:**
```bash
$ dig +short fapi.binance.com
d2ukl3c6tymv7q.cloudfront.net.
3.169.165.46                    # ← NOT in AllowedIPs!
```

**Routing check:**
```bash
$ ip route get 3.169.165.46
3.169.165.46 via 66.94.120.1 dev eth0  # ← Going through eth0, NOT wg0!
```

**Result:** Binance traffic goes through normal internet connection, not VPN tunnel.

---

## Why the IP Ranges Are Wrong

### Binance Uses CloudFront CDN

**Binance domain:**
```
fapi.binance.com → d2ukl3c6tymv7q.cloudfront.net → 3.169.165.46
```

**CloudFront IP Ranges:**
- CloudFront is Amazon's CDN with **400+ IP ranges globally**
- IPs in ranges: `3.x.x.x`, `13.x.x.x`, `15.x.x.x`, `18.x.x.x`, `52.x.x.x`, `54.x.x.x`, `99.x.x.x`, `205.x.x.x`, etc.
- **Dynamic and constantly changing**

**VPN Config IP Ranges:**
- `76.223.0.0/16` - Old/incorrect range
- `54.76.0.0/16` - Partially AWS, but not where Binance resolves
- `99.81.0.0/16` - Partially AWS, but not where Binance resolves

**Mismatch:**
- Current Binance IP: `3.169.165.46`
- VPN routes: Don't include `3.0.0.0/8`
- **Traffic misses VPN tunnel**

---

## Why This Worked in the Past

**Hypothesis:**
- Binance previously resolved to IPs in those specific ranges
- CloudFront changed their routing
- Binance CDN now uses different edge locations
- VPN config became outdated

**Or:**
- Config was created for specific Binance endpoints
- `fapi.binance.com` (Futures API) uses different IPs than spot market
- Different services use different CloudFront distributions

---

## Solutions

### Option 1: Add Full AWS/CloudFront Ranges 🔴 NOT RECOMMENDED

**Pros:** Would route all Binance traffic through VPN
**Cons:**
- Would route LOTS of other AWS services through VPN (too broad)
- Could break other services using AWS
- Performance impact
- Security risk

### Option 2: Add Specific IP Range ⭐ QUICK FIX

**Add the 3.0.0.0/8 range:**
```
AllowedIPs = 76.223.0.0/16, 54.76.0.0/16, 99.81.0.0/16, 3.0.0.0/8
```

**Pros:**
- Catches current Binance IPs
- Relatively specific
- Simple to implement

**Cons:**
- Still routes some non-Binance AWS traffic through VPN
- May need updates if Binance changes CDN

### Option 3: Dynamic IP Detection 🔧 COMPLEX

**Automatically detect Binance IPs:**
- Resolve `fapi.binance.com` dynamically
- Add resolved IPs to VPN routing
- Update routing table when IPs change

**Pros:** Always up-to-date
**Cons:** Complex, requires scripting

### Option 4: Use Binance Testnet ⭐⭐ BEST SOLUTION

**Switch to testnet:**
- URL: `https://testnet.binancefuture.com`
- **NO geo-restrictions** (works from anywhere)
- No VPN needed
- Free test funds
- Perfect for development

**Pros:**
- No VPN complexity
- No geo-blocking
- No IP range maintenance
- Simpler setup

**Cons:**
- Not real trading (test environment only)
- Separate from production account

### Option 5: Full-Tunnel VPN with DNS Fix 🔴 KILLS SSH

**Route all traffic through VPN:**
```
AllowedIPs = 0.0.0.0/0
```

**Pros:** Definitely routes Binance
**Cons:**
- SSH session would die
- All traffic through VPN
- Performance impact

---

## Why Market-Data Works Without This

**Market-data uses WebSocket for public data:**
```
wss://fstream.binance.com/stream
```

**WebSocket endpoint:**
- Resolves to different IPs (WebSocket servers)
- Possibly in the configured ranges, OR
- **Public WebSocket not geo-blocked at all** (no VPN needed)

**Account-monitor needs:**
- REST API: `https://fapi.binance.com/fapi/v2/account`
- Private WebSocket: Authenticated user data stream
- **Both are geo-blocked and need VPN**

---

## Current VPN Effectiveness

**What's routed through VPN:**
- IPs in 76.223.0.0/16 ✅
- IPs in 54.76.0.0/16 ✅
- IPs in 99.81.0.0/16 ✅

**What Binance resolves to:**
- 3.169.165.46 (CloudFront) ❌ NOT ROUTED

**Result:** 0% of Binance traffic actually using VPN

---

## Recommendations

**Immediate (5 minutes):**
1. Add `3.0.0.0/8` to AllowedIPs
2. Restart VPN: `sudo wg-quick down wg0 && sudo wg-quick up wg0`
3. Test Binance access
4. Restart account-monitor

**Better (15 minutes):**
1. Use Binance Testnet instead
2. Update account-monitor config
3. No VPN needed
4. Simpler long-term

**Best (30 minutes):**
1. Use Binance Testnet for development
2. Use proper production VPN for real trading
3. Separate configs for dev/prod

---

## Analysis Complete

**VPN DNS Issue:** ✅ **FIXED** (removed DNS line)
**VPN Stability:** ✅ **WORKING** (interface stays up)
**Binance Routing:** ❌ **NOT WORKING** (wrong IP ranges)

**The VPN itself is working perfectly. The problem is the IP ranges are outdated and don't match where Binance currently resolves.**

**Want me to:**
1. Add 3.0.0.0/8 to VPN routes (quick fix)
2. Switch to Binance Testnet (better solution)
3. Both
