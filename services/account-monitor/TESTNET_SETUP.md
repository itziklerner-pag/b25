# Binance Testnet Setup for Account Monitor

**Service:** account-monitor
**Status:** ✅ Testnet already configured in code
**Action needed:** Get testnet API keys

---

## Quick Start

### Step 1: Get Binance Testnet API Keys

**Go to:** https://testnet.binancefuture.com/

**Steps:**
1. Click "Log In" (top right)
2. Sign in with your email/Google account (free registration)
3. You'll get **10,000 USDT in test funds** automatically
4. Go to API Management:
   - Click your email (top right)
   - Select "API Management"
5. Create API Key:
   - Click "Create API"
   - Label: "B25 Account Monitor"
   - Permissions needed:
     - ✅ Enable Reading
     - ✅ Enable Futures (if available)
     - ❌ No trading permissions needed (just monitoring)
   - Click "Create"
6. **SAVE YOUR KEYS IMMEDIATELY:**
   - API Key: `xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
   - Secret Key: `yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy`
   - You can only see the secret ONCE!

### Step 2: Update .env File

```bash
cd /home/mm/dev/b25/services/account-monitor

# Edit .env
vim .env

# Replace placeholders:
BINANCE_API_KEY=your_actual_testnet_api_key_here
BINANCE_SECRET_KEY=your_actual_testnet_secret_key_here
POSTGRES_PASSWORD=L9JYNAeS3qdtqa6CrExpMA==
```

### Step 3: Restart Service

```bash
# Kill existing instance
pkill -f account-monitor

# Start with new credentials
export POSTGRES_PASSWORD="L9JYNAeS3qdtqa6CrExpMA=="
export BINANCE_API_KEY="your_testnet_key"
export BINANCE_SECRET_KEY="your_testnet_secret"

./bin/account-monitor > logs/account-monitor.log 2>&1 &

# Check logs
tail -f logs/account-monitor.log
```

---

## Configuration Already Set

**✅ account-monitor is already configured for testnet!**

From `config.yaml`:
```yaml
exchange:
  name: binance
  testnet: true          # ← Already enabled!
```

**What this does (from code):**
```go
// REST API
baseURL := "https://api.binance.com"
if cfg.Testnet {
    baseURL = "https://testnet.binance.vision"  // ✅ Testnet
}

// WebSocket
wsURL := "wss://stream.binance.com:9443/ws/"
if w.cfg.Testnet {
    wsURL = "wss://testnet.binance.vision/ws/"  // ✅ Testnet
}
```

**No VPN needed** - Testnet has no geo-restrictions!

---

## Testnet Benefits

**✅ No Geo-Restrictions:**
- Works from any location
- No VPN needed
- No IP range configuration

**✅ Free Test Funds:**
- 10,000 USDT provided automatically
- Can reset funds anytime
- Perfect for development

**✅ Safe Testing:**
- No real money at risk
- Can test all features
- Realistic trading environment

**✅ No Rate Limits:**
- More generous API rate limits
- Better for development

---

## After Getting API Keys

**Expected behavior:**
```json
// GET /health
{
  "status": "healthy",
  "checks": {
    "postgres": {"status": "healthy"},
    "redis": {"status": "healthy"},
    "nats": {"status": "healthy"},
    "exchange": {"status": "healthy"}  // ← Should be healthy with testnet
  }
}

// GET /api/account
{
  "totalWalletBalance": "10000.00000000",
  "availableBalance": "10000.00000000",
  "totalUnrealizedProfit": "0.00000000"
}
```

---

## Troubleshooting

**If still getting errors:**

1. **Check API key permissions:**
   - Go to testnet.binancefuture.com
   - API Management
   - Verify "Enable Reading" is checked

2. **Check API key format:**
   - Should be long alphanumeric strings
   - No spaces or quotes
   - Copy-paste directly from testnet website

3. **Test API key manually:**
   ```bash
   API_KEY="your_key"
   SECRET="your_secret"
   TIMESTAMP=$(date +%s000)

   curl "https://testnet.binancefuture.com/fapi/v1/time"
   # Should return: {"serverTime": 1234567890}
   ```

---

## Current .env File

Located at: `/home/mm/dev/b25/services/account-monitor/.env`

**Current values:**
```bash
BINANCE_API_KEY=your_test_api_key        # ← Replace with real testnet key
BINANCE_SECRET_KEY=your_test_secret_key  # ← Replace with real testnet secret
POSTGRES_PASSWORD=L9JYNAeS3qdtqa6CrExpMA==  # ✅ Correct
```

---

## Next Steps

1. Get testnet API keys from: https://testnet.binancefuture.com/
2. Update `.env` with real keys
3. Restart account-monitor
4. Check health: `curl http://localhost:9093/health`
5. Enable in UI: Set `VITE_SERVICE_ACCOUNT_MONITOR_ENABLED=true` in `ui/web/.env`
6. Rebuild UI: `cd ui/web && npm run build`

---

**The service is ready - just needs your testnet API keys!**
