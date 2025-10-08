# 🎚️ Leverage Configuration Guide

## Current Leverage Settings

**Verified from your Binance account** (2025-10-08):

```json
{
  "symbol": "BTCUSDT",
  "leverage": "20",        ← 20x leverage
  "marginType": "CROSS"    ← Cross margin (uses full account)
}
{
  "symbol": "ETHUSDT",
  "leverage": "20",        ← 20x leverage
  "marginType": "CROSS"
}
{
  "symbol": "SOLUSDT",
  "leverage": "20",        ← 20x leverage
  "marginType": "CROSS"
}
```

### What This Means:

**20x Leverage**:
- You can control **$20,000** worth of crypto with **$1,000** margin
- **5% price move** against you = **100% loss** (liquidation)
- **1% price move** = **20% gain or loss**

**Cross Margin**:
- Uses your **entire 1000 USDT** account as collateral
- All positions share the same margin pool
- If one position liquidates, it can take your whole account

---

## ⚠️ CRITICAL: Your Leverage is TOO HIGH for a $1000 Account!

### Why 20x is Risky:

**Example with BTCUSDT at $121,000**:

```
Your Balance: $1,000
Leverage: 20x
Position Size: 0.01 BTC ($1,210 notional)
Margin Used: $60.50

Scenario 1: BTC drops 5% to $114,950
Your Loss: $60.50 (100% of margin used) → LIQUIDATED ❌

Scenario 2: BTC drops 1% to $119,790
Your Loss: $12.10 (20% of margin) → Still open but bleeding

Scenario 3: BTC rises 1% to $122,210
Your Profit: $12.10 (20% gain) ✅
```

**With 20x leverage on $1000 account**:
- One bad 5% move = account wiped
- Crypto regularly moves 5% in hours
- Sleep for 8 hours = wake up liquidated

---

## 🎯 Recommended Leverage for $1000 Account

### Conservative (Recommended):

**3x - 5x Leverage**:
```
Balance: $1,000
Leverage: 5x
Max Position: 0.04 BTC ($4,850 notional)
Margin for Max: $970
Liquidation: ~20% price move (much safer!)

Daily volatility (BTC): ~2-3%
Your risk with 5x: 10-15% swing on bad day
Still risky but manageable
```

### Moderate:

**10x Leverage**:
```
Balance: $1,000
Leverage: 10x
Max Position: 0.08 BTC ($9,700 notional)
Liquidation: ~10% price move
Risk: High but experienced traders can manage
```

### Aggressive (Current):

**20x Leverage**:
```
Balance: $1,000
Leverage: 20x
Max Position: 0.16 BTC ($19,400 notional)
Liquidation: ~5% price move
Risk: VERY HIGH - not recommended for beginners
```

---

## 🔧 How to Change Leverage

### Method 1: Via Binance Web Interface (Easiest)

1. Go to: https://www.binance.com/en/futures/BTCUSDT
2. Log in to your account
3. Find leverage slider (usually top right)
4. **Per symbol**:
   - BTCUSDT: Set to 5x
   - ETHUSDT: Set to 5x
   - SOLUSDT: Set to 5x
5. Click "Confirm"

**Changes apply immediately** - your system will use the new leverage on next trade.

---

### Method 2: Via Binance API (Programmatic)

Add this script to your order-execution service:

**File**: `/home/mm/dev/b25/services/order-execution/scripts/set-leverage.sh`

```bash
#!/bin/bash

API_KEY="${BINANCE_API_KEY}"
SECRET_KEY="${BINANCE_SECRET_KEY}"
BASE_URL="https://fapi.binance.com"

set_leverage() {
    SYMBOL=$1
    LEVERAGE=$2

    TIMESTAMP=$(date +%s000)
    QUERY_STRING="symbol=${SYMBOL}&leverage=${LEVERAGE}&timestamp=${TIMESTAMP}"
    SIGNATURE=$(echo -n "${QUERY_STRING}" | openssl dgst -sha256 -hmac "${SECRET_KEY}" | awk '{print $2}')

    curl -X POST -H "X-MBX-APIKEY: ${API_KEY}" \
        "${BASE_URL}/fapi/v1/leverage?${QUERY_STRING}&signature=${SIGNATURE}"
    echo ""
}

echo "Setting leverage to 5x for all symbols..."
set_leverage "BTCUSDT" 5
set_leverage "ETHUSDT" 5
set_leverage "SOLUSDT" 5
echo "Leverage updated!"
```

**Usage**:
```bash
cd /home/mm/dev/b25/services/order-execution
chmod +x scripts/set-leverage.sh

BINANCE_API_KEY='rh22mtiKxsGSWuK3USkf4ba7E88exyVpn0INbc2OyCnogNsQ0R2A4lUcvHNJRcSc' \
BINANCE_SECRET_KEY='xUwZCEWa5g9auPgT5uYP8ClATN2zgGGYAFYgl4WoPTge2TWVxbz0ZBUmnV6PyOMh' \
./scripts/set-leverage.sh
```

---

### Method 3: Add Leverage Change to Order Execution Service

**Add this method to** `/home/mm/dev/b25/services/order-execution/internal/exchange/binance.go`:

```go
// ChangeLeverage changes leverage for a symbol
func (c *BinanceClient) ChangeLeverage(ctx context.Context, symbol string, leverage int) error {
    endpoint := "/fapi/v1/leverage"

    params := url.Values{}
    params.Add("symbol", symbol)
    params.Add("leverage", fmt.Sprintf("%d", leverage))
    params.Add("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))

    signature := c.sign(params.Encode())
    params.Add("signature", signature)

    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint+"?"+params.Encode(), nil)
    if err != nil {
        return err
    }

    req.Header.Set("X-MBX-APIKEY", c.apiKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to change leverage: %s", string(body))
    }

    c.logger.Info("Leverage changed successfully",
        zap.String("symbol", symbol),
        zap.Int("leverage", leverage),
    )

    return nil
}
```

Then call on startup or via API endpoint.

---

## 📊 Leverage Impact Calculator

### Position Size vs Leverage

**With $1000 USDT balance**:

| Leverage | Max Notional | BTC Amount | 1% Move P&L | 5% Move P&L | Liquidation |
|----------|--------------|------------|-------------|-------------|-------------|
| **1x**   | $1,000       | 0.008 BTC  | $10         | $50         | N/A         |
| **3x**   | $3,000       | 0.025 BTC  | $30         | $150        | ~33% move   |
| **5x**   | $5,000       | 0.041 BTC  | $50         | $250        | ~20% move   |
| **10x**  | $10,000      | 0.083 BTC  | $100        | $500        | ~10% move   |
| **20x**  | $20,000      | 0.165 BTC  | $200        | $1,000      | ~5% move    |

### Recommended Leverage by Experience:

**Beginner** (< 6 months trading):
- Leverage: **1x - 3x**
- Why: Learn without risking account wipeout
- Position size: $100 - $300

**Intermediate** (6-12 months):
- Leverage: **3x - 5x**
- Why: Balance between capital efficiency and safety
- Position size: $300 - $500

**Advanced** (> 1 year, proven strategy):
- Leverage: **5x - 10x**
- Why: Maximize returns with managed risk
- Position size: $500 - $1,000

**Expert** (Professional trader, fully automated):
- Leverage: **10x - 20x**
- Why: High capital efficiency
- Requires: Stop-losses, risk management, monitoring
- ⚠️ Only if you know what you're doing!

---

## 🛠️ How to Configure Leverage in Your System

### Option A: Change in Binance (Affects All Trading)

**Quick change via web**:
1. https://www.binance.com/en/futures/BTCUSDT
2. Adjust leverage slider
3. System automatically uses new leverage

**Pros**:
- Instant change
- Applies to all trading (manual + automated)
- No code changes needed

**Cons**:
- Must do per symbol
- Not version controlled

---

### Option B: Add Leverage API to Order Execution

**Create this file**: `/home/mm/dev/b25/services/order-execution/internal/exchange/leverage.go`

```go
package exchange

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

// SetLeverage changes the leverage for a symbol
func (c *BinanceClient) SetLeverage(ctx context.Context, symbol string, leverage int) error {
    endpoint := "/fapi/v1/leverage"

    // Get fresh server time
    serverTime, err := c.getFreshServerTime(ctx)
    if err != nil {
        serverTime = time.Now().UnixMilli()
    }

    params := url.Values{}
    params.Add("symbol", symbol)
    params.Add("leverage", fmt.Sprintf("%d", leverage))
    params.Add("timestamp", fmt.Sprintf("%d", serverTime))
    params.Add("recvWindow", "10000")

    signature := c.sign(params.Encode())
    params.Add("signature", signature)

    reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())

    req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
    if err != nil {
        return err
    }

    req.Header.Set("X-MBX-APIKEY", c.config.APIKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to set leverage: %s", string(body))
    }

    c.logger.Info("Leverage set successfully",
        zap.String("symbol", symbol),
        zap.Int("leverage", leverage),
    )

    return nil
}

// InitializeLeverageSettings sets leverage for all configured symbols on startup
func (c *BinanceClient) InitializeLeverageSettings(ctx context.Context, defaultLeverage int) error {
    symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT"}

    for _, symbol := range symbols {
        if err := c.SetLeverage(ctx, symbol, defaultLeverage); err != nil {
            c.logger.Warn("Failed to set leverage",
                zap.String("symbol", symbol),
                zap.Error(err),
            )
            continue
        }
    }

    return nil
}
```

**Then add to config.yaml**:
```yaml
exchange:
  api_key: "${BINANCE_API_KEY}"
  secret_key: "${BINANCE_SECRET_KEY}"
  testnet: false
  default_leverage: 5  # ← Add this line
```

**Call on startup** in `/home/mm/dev/b25/services/order-execution/cmd/server/main.go`:
```go
// After creating binance client:
if cfg.Exchange.DefaultLeverage > 0 {
    if err := binanceClient.InitializeLeverageSettings(ctx, cfg.Exchange.DefaultLeverage); err != nil {
        logger.Warn("Failed to initialize leverage settings", zap.Error(err))
    }
}
```

---

### Option C: Configure Per-Strategy Leverage

**Add to Strategy Engine config**:
```yaml
strategies:
  configs:
    scalping:
      leverage: 5  # ← Add leverage per strategy
      max_position: 100.0

    momentum:
      leverage: 3  # ← Conservative for momentum
      max_position: 150.0

    market_making:
      leverage: 2  # ← Very conservative for MM
      max_inventory: 200.0
```

**Then modify Strategy Engine to send leverage with each order**.

---

## 📐 Leverage vs Position Size Relationship

### Understanding the Math:

**Formula**:
```
Margin Required = (Position Notional Value) / Leverage

Position Notional = Quantity × Price
Margin = Position Notional / Leverage
```

**Example with BTCUSDT @ $121,000**:

**Want to buy 0.01 BTC ($1,210 notional)**:

```
At 1x leverage:  Margin = $1,210 / 1  = $1,210 (can't afford!)
At 5x leverage:  Margin = $1,210 / 5  = $242   (affordable)
At 10x leverage: Margin = $1,210 / 10 = $121   (even cheaper)
At 20x leverage: Margin = $1,210 / 20 = $60.50 (very cheap, very risky)
```

### Effective Leverage from Position Size:

Your strategy configs control position size, which determines effective leverage:

```yaml
# If max_position: 500.0 with $1000 account
Effective Leverage = $500 / $1000 = 0.5x (very safe)

# If max_position: 2000.0 with $1000 account
Effective Leverage = $2000 / $1000 = 2x (moderate)

# If max_position: 5000.0 with $1000 account
Effective Leverage = $5000 / $1000 = 5x (aggressive)
```

**Even if Binance leverage is 20x, if your strategy only uses $100 positions, your effective leverage is 0.1x!**

---

## 🎮 Quick Leverage Change Commands

### Change to 5x (Recommended):

```bash
# Create script
cat > /tmp/set-leverage-5x.sh << 'EOFSCRIPT'
#!/bin/bash
API_KEY="rh22mtiKxsGSWuK3USkf4ba7E88exyVpn0INbc2OyCnogNsQ0R2A4lUcvHNJRcSc"
SECRET_KEY="xUwZCEWa5g9auPgT5uYP8ClATN2zgGGYAFYgl4WoPTge2TWVxbz0ZBUmnV6PyOMh"
BASE_URL="https://fapi.binance.com"

set_leverage() {
    SYMBOL=$1
    LEVERAGE=$2
    TIMESTAMP=$(date +%s000)
    QUERY_STRING="symbol=${SYMBOL}&leverage=${LEVERAGE}&timestamp=${TIMESTAMP}"
    SIGNATURE=$(echo -n "${QUERY_STRING}" | openssl dgst -sha256 -hmac "${SECRET_KEY}" | awk '{print $2}')

    echo "Setting ${SYMBOL} leverage to ${LEVERAGE}x..."
    curl -s -X POST -H "X-MBX-APIKEY: ${API_KEY}" \
        "${BASE_URL}/fapi/v1/leverage?${QUERY_STRING}&signature=${SIGNATURE}" | jq .
}

set_leverage "BTCUSDT" 5
set_leverage "ETHUSDT" 5
set_leverage "SOLUSDT" 5
EOFSCRIPT

chmod +x /tmp/set-leverage-5x.sh
/tmp/set-leverage-5x.sh
```

### Change to 3x (Very Conservative):

```bash
# Same script, change the numbers to 3:
set_leverage "BTCUSDT" 3
set_leverage "ETHUSDT" 3
set_leverage "SOLUSDT" 3
```

### Verify New Leverage:

```bash
/tmp/get-leverage.sh
```

---

## 💡 System Leverage Flow

### Where Leverage is Set:

```
1. Binance Account Settings (per symbol)
   └─> Stored on Binance servers
       └─> Applied to ALL orders for that symbol
           └─> Your trading system uses this leverage automatically
```

**Your system does NOT set leverage**:
- Order Execution submits orders WITHOUT leverage parameter
- Binance uses the leverage you've set for that symbol
- This is standard practice (leverage is account-level, not order-level)

### How Your System Uses Leverage:

```
Strategy generates signal: "Buy 0.01 BTC"
    ↓
Order Execution receives signal
    ↓
Submits order to Binance: "BUY 0.01 BTCUSDT LIMIT 121200"
    ↓
Binance applies YOUR account's leverage setting (20x)
    ↓
Position opened with 20x leverage
    ↓
Margin Required = ($1,210 / 20) = $60.50
    ↓
Account Monitor sees position and margin usage
```

---

## 🎯 Recommended Configuration for $1000 Account

### Complete Safe Setup:

**1. Change Binance Leverage** (via web or script above):
```
BTCUSDT: 20x → 5x
ETHUSDT: 20x → 5x
SOLUSDT: 20x → 5x
```

**2. Adjust Strategy Position Sizes**:
```yaml
# /home/mm/dev/b25/services/strategy-engine/config.yaml

strategies:
  configs:
    scalping:
      max_position: 100.0  # $100 max (10% of account)

    momentum:
      max_position: 150.0  # $150 max (15% of account)

    market_making:
      order_size: 50.0     # $50 per order
      max_inventory: 200.0  # $200 total inventory
```

**3. Set Risk Limits**:
```yaml
risk:
  enabled: true
  maxPositionSize: 100.0   # Max $100 per position
  maxOrderValue: 500.0     # Max $500 total order value
  maxDailyLoss: 50.0       # Stop after $50 daily loss
  maxDrawdown: 0.10        # Stop at 10% drawdown
  minAccountBalance: 900.0  # Stop if balance < $900
```

**Result**:
- Effective leverage: ~1x - 2x (very safe)
- Account leverage: 5x (safety buffer)
- Max loss per trade: ~$5 (with stop-loss)
- Max daily loss: $50 (5% of account)
- Liquidation risk: Very low (~20% move needed)

---

## 🔍 Check Your Current Leverage

**Run this anytime**:
```bash
cd /home/mm/dev/b25/services/account-monitor
./test-futures-account.sh | grep -A 3 "BTCUSDT\|ETHUSDT\|SOLUSDT" | grep leverage
```

**Or check via API**:
```bash
curl "https://mm.itziklerner.com/services/account-monitor/api/positions" | jq '.[] | {symbol, leverage}'
```

**Or in admin page**:
https://mm.itziklerner.com/services/account-monitor/
- Click "Test Positions" button
- Look for "leverage" field in response

---

## 🎲 Leverage Scenarios

### Scenario 1: Current Setup (20x) - RISKY

```
Account: $1,000
Leverage: 20x
Strategy opens: 0.01 BTC @ $121,000

Position Notional: $1,210
Margin Used: $60.50
Available: $939.50

BTC moves to $115,000 (-5%):
Loss: $60.50 (100% of margin)
Result: LIQUIDATED ❌
New Balance: ~$939 (lost $61 + fees)
```

### Scenario 2: Recommended Setup (5x) - SAFER

```
Account: $1,000
Leverage: 5x
Strategy opens: 0.01 BTC @ $121,000

Position Notional: $1,210
Margin Used: $242
Available: $758

BTC moves to $115,000 (-5%):
Loss: $60.50 (25% of margin)
Result: Still Open ✅
Unrealized Loss: -$60.50
Can close or wait for recovery
```

### Scenario 3: Conservative (3x) - SAFEST

```
Account: $1,000
Leverage: 3x
Strategy opens: 0.008 BTC @ $121,000

Position Notional: $968
Margin Used: $323
Available: $677

BTC moves to $115,000 (-5%):
Loss: $48.40 (15% of margin)
Result: Comfortably Open ✅
Unrealized Loss: -$48.40
Plenty of buffer before liquidation
```

---

## 📋 Summary

### Your Current Leverage: **20x** ⚠️

**Symbols**:
- BTCUSDT: 20x (Cross Margin)
- ETHUSDT: 20x (Cross Margin)
- SOLUSDT: 20x (Cross Margin)

### How to Change:

**Easiest**: Binance web interface (5 minutes)
**Automated**: Use the script above (2 minutes)
**Programmatic**: Add method to Order Execution service

### Recommended for $1000 Account:

**Leverage**: 3x - 5x (not 20x!)
**Why**:
- 20x = 5% move wipes your account
- 5x = 20% move needed for liquidation
- Much safer for learning

### How Your System Uses It:

Your automated trading system:
1. ✅ Does NOT set leverage (uses Binance account settings)
2. ✅ Submits orders based on strategy signals
3. ✅ Binance applies YOUR leverage setting (currently 20x)
4. ✅ Position opened with that leverage
5. ✅ Account Monitor tracks everything

**Bottom Line**: Your system is ready for USD-M Futures trading. Change leverage to 5x for safety, then follow the GO_LIVE_TRADING_GUIDE.md! 🚀

---

**Files Created**:
- `/home/mm/dev/b25/LEVERAGE_CONFIGURATION_GUIDE.md` - This file
- `/home/mm/dev/b25/USD_M_FUTURES_VERIFICATION.md` - USD-M readiness
- `/home/mm/dev/b25/GO_LIVE_TRADING_GUIDE.md` - How to go live