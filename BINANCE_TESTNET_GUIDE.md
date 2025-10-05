# Binance Futures Testnet API Keys - How to Get Them

## ✅ NO SPECIAL REQUIREMENTS

Binance Futures Testnet API keys are:
- ✅ **FREE** - No payment needed
- ✅ **NO KYC** - No identity verification required
- ✅ **NO VIP TIER** - Available to everyone
- ✅ **INSTANT** - Generated immediately

---

## 📝 HOW TO GET TESTNET API KEYS

### Step 1: Create Testnet Account
1. Go to: **https://testnet.binancefuture.com/**
2. Click "Register" (top right)
3. Enter email + password
4. Verify email
5. Login

### Step 2: Get Testnet Funds
1. After login, you'll see your testnet account
2. Testnet gives you **FREE test USDT** (not real money)
3. You can request more testnet funds anytime (usually 1000-10000 USDT)

### Step 3: Generate API Keys
1. Click on your profile/account icon
2. Go to **"API Management"** or **"API Keys"**
3. Click **"Create API"**
4. Give it a label (e.g., "B25 Trading Bot")
5. **IMPORTANT:** Enable these permissions:
   - ✅ **Enable Futures** (MUST have this!)
   - ✅ **Enable Reading**
   - ✅ **Enable Spot & Margin Trading** (optional)
6. **IP Restriction:**
   - Option A: Leave blank (no restriction) - RECOMMENDED for testing
   - Option B: Add your VPS IP: `2605:a141:2183:7862::1`
   - Option C: Add VPN IP after connecting
7. Click **Create**
8. **SAVE** the API Key and Secret Key immediately (shown only once!)

### Step 4: Update B25 System
1. Copy your new API keys
2. Edit `.env` file:
   ```bash
   nano /home/mm/dev/b25/.env
   ```
3. Update lines:
   ```
   EXCHANGE_API_KEY=your_new_api_key_here
   EXCHANGE_SECRET_KEY=your_new_secret_key_here
   ```
4. Restart services:
   ```bash
   ./stop-all-services.sh
   ./run-all-services.sh
   ```

---

## ⚠️ COMMON MISTAKES

### 1. Using Real Binance API Keys
- ❌ Don't use keys from **binance.com**
- ✅ Use keys from **testnet.binancefuture.com**

### 2. Not Enabling Futures Permission
- The API key MUST have "Enable Futures" checked
- Without this, you get "Invalid API-key" error

### 3. IP Restriction
- If you set IP restriction, VPN IP won't work (different IP)
- Recommendation: Disable IP restriction for testnet

### 4. Wrong Testnet URL
- Spot testnet: testnet.binance.vision (wrong for futures)
- Futures testnet: testnet.binancefuture.com (correct!)

---

## 🧪 VERIFY YOUR API KEYS

After generating new keys, test them:

```bash
cd /home/mm/dev/b25
./test-binance-api.sh
```

You should see:
```
✅ Ping successful
✅ API authentication successful!
Balance: 10000 USDT
```

---

## 🔧 CURRENT ISSUE ANALYSIS

Your current API keys return: `Invalid API-key, IP, or permissions for action`

**Possible causes:**
1. Keys are from wrong testnet (spot vs futures)
2. Keys don't have Futures permission enabled
3. Keys are expired or revoked
4. IP whitelist doesn't include your IP

**Most likely:** Keys don't have Futures trading permission enabled.

---

## ✅ RECOMMENDED NEXT STEPS

1. Go to https://testnet.binancefuture.com/
2. Login (or register if you don't have account)
3. Generate NEW API keys with:
   - ✅ Enable Futures checked
   - ✅ Enable Reading checked
   - ✅ NO IP restriction (leave blank)
4. Copy the keys
5. Update `/home/mm/dev/b25/.env`
6. Run `./test-binance-api.sh`
7. If test passes, restart services

---

## 📊 WHAT WORKS NOW (Without Account API)

Even with current API key issue, you have:
- ✅ Market Data Service - Receiving live orderbook/trades
- ✅ Strategy Engine - Analyzing market (3 strategies)
- ✅ Order Execution - Ready to place orders (simulation mode)
- ✅ Dashboard - Web UI working
- ✅ Monitoring - Grafana/Prometheus active

Only missing:
- ❌ Account balance queries via API
- ❌ Position tracking via WebSocket user data stream

---

**Bottom Line:** Get new testnet API keys with proper permissions and the system will be 100% functional.
