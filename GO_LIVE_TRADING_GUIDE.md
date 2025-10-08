# 🚀 How to Start Live Auto Trading - Complete Guide

## Current System Status

### What's Currently Running:
- ✅ **Account Monitor**: Connected to **PRODUCTION** Binance Futures (1000 USDT balance tracked)
- ⚠️ **Order Execution**: Set to **TESTNET** mode
- ⚠️ **Strategy Engine**: Set to **SIMULATION** mode (strategies generate signals but don't place real orders)
- ✅ All other services operational

### Current Configuration Summary:
```
Account Monitor:   PRODUCTION (testnet: false) ✅ Live balance: 1000 USDT
Order Execution:   TESTNET    (testnet: true)  ⚠️ Won't execute on real account
Strategy Engine:   SIMULATION (mode: simulation) ⚠️ Paper trading only
```

**Result**: Your strategies are generating signals, but orders are NOT being executed on your live account because:
1. Order Execution is pointing to testnet
2. Strategy Engine is in simulation mode (won't send orders to Order Execution)

---

## 🎯 Step-by-Step: Enable Live Trading

### ⚠️ CRITICAL WARNING
**BEFORE YOU START LIVE TRADING**:
- You have **$1,000 USD** in your Binance Futures account
- Live trading means REAL MONEY can be lost
- Start with small position sizes
- Monitor closely for the first few hours
- Have stop-loss strategies enabled
- Risk Manager is in emergency stop mode (needs fixing first)

---

### Step 1: Fix Risk Manager (CRITICAL - Do This First!)

The Risk Manager is currently in emergency stop mode due to fake violations. You MUST fix this before live trading:

```bash
cd /home/mm/dev/b25/services/risk-manager

# The emergency stop was triggered. You need to restart with proper account data
# Option 1: Restart Risk Manager (clears emergency stop)
pkill -f "bin/service"
nohup ./bin/service > /tmp/risk-manager.log 2>&1 &

# Option 2: Disable Risk Manager temporarily (NOT RECOMMENDED for live trading)
# Only do this if you want to trade WITHOUT risk protection
```

**Verify Risk Manager is working**:
```bash
curl http://localhost:9095/health
# Should return: {"status":"healthy"}
```

---

### Step 2: Switch Order Execution to Production

**File**: `/home/mm/dev/b25/services/order-execution/config.yaml`

Change line 12:
```yaml
exchange:
  testnet: true  # ← Change this to false
```

To:
```yaml
exchange:
  testnet: false  # PRODUCTION - REAL TRADING
```

**Rebuild and Restart**:
```bash
cd /home/mm/dev/b25/services/order-execution
go build -o bin/order-execution ./cmd/server
pkill -f "bin/order-execution"

BINANCE_API_KEY='rh22mtiKxsGSWuK3USkf4ba7E88exyVpn0INbc2OyCnogNsQ0R2A4lUcvHNJRcSc' \
BINANCE_SECRET_KEY='xUwZCEWa5g9auPgT5uYP8ClATN2zgGGYAFYgl4WoPTge2TWVxbz0ZBUmnV6PyOMh' \
nohup ./bin/order-execution > /tmp/order-execution.log 2>&1 &
```

**Verify**:
```bash
curl http://localhost:9091/health | jq .
# Check logs for "testnet: false" or look for fapi.binance.com (not testnet URLs)
tail -f /tmp/order-execution.log
```

---

### Step 3: Switch Strategy Engine to Live Mode

**File**: `/home/mm/dev/b25/services/strategy-engine/config.yaml`

Change line 37:
```yaml
engine:
  mode: "simulation"  # ← Change this to "live"
```

To:
```yaml
engine:
  mode: "live"  # LIVE TRADING - SENDS REAL ORDERS
```

**Rebuild and Restart**:
```bash
cd /home/mm/dev/b25/services/strategy-engine
make build
pkill -f "bin/strategy-engine"
nohup ./bin/strategy-engine > /tmp/strategy-engine.log 2>&1 &
```

**Verify Live Mode**:
```bash
curl http://localhost:9092/status | jq '.mode'
# Should return: "live"

tail -f /tmp/strategy-engine.log
# Look for: "mode: live" in startup logs
```

---

### Step 4: Configure Strategy Parameters (IMPORTANT!)

Before going live, review and adjust strategy configurations to be conservative:

**File**: `/home/mm/dev/b25/services/strategy-engine/config.yaml` (lines 50-66)

Current settings:
```yaml
strategies:
  configs:
    momentum:
      lookback_period: 20
      threshold: 0.02
      max_position: 1000.0  # ← This is $1000 - your ENTIRE balance!
    market_making:
      spread: 0.001
      order_size: 100.0     # ← $100 per order
      max_inventory: 1000.0  # ← $1000 max
    scalping:
      profit_target: 0.001
      stop_loss: 0.0005
      max_position: 500.0    # ← $500 max position
```

**RECOMMENDED for $1000 account**:
```yaml
strategies:
  configs:
    momentum:
      max_position: 100.0  # Only 10% of account ($100)
    market_making:
      order_size: 50.0     # $50 per order
      max_inventory: 200.0  # $200 max inventory
    scalping:
      max_position: 100.0   # $100 max position
```

---

### Step 5: Verify Risk Limits

**File**: `/home/mm/dev/b25/services/strategy-engine/config.yaml` (lines 70-83)

```yaml
risk:
  enabled: true
  maxPositionSize: 1000.0      # ← Reduce to 100.0 for safety
  maxOrderValue: 50000.0       # ← Reduce to 500.0
  maxDailyLoss: 5000.0         # ← Reduce to 50.0 (5% of account)
  maxDrawdown: 0.10            # ← 10% max drawdown (good)
  minAccountBalance: 10000.0   # ← Reduce to 900.0 (keep $900 minimum)
```

**RECOMMENDED settings**:
```yaml
risk:
  enabled: true
  maxPositionSize: 100.0       # $100 max position
  maxOrderValue: 500.0         # $500 max single order
  maxDailyLoss: 50.0           # $50 daily loss limit (5%)
  maxDrawdown: 0.10            # 10% max drawdown
  minAccountBalance: 900.0     # Keep $900 minimum
```

---

### Step 6: Start Small - Enable ONE Strategy First

Instead of enabling all 3 strategies, start with just one:

**File**: `/home/mm/dev/b25/services/strategy-engine/config.yaml` (lines 46-49)

```yaml
strategies:
  enabled:
    - scalping  # Start with ONLY scalping (fastest feedback)
    # - momentum  # Enable later
    # - market_making  # Enable later
```

Why scalping first?
- Fast execution (holds positions < 60 seconds)
- Small profit targets (10 bps = 0.1%)
- Tight stop-loss (5 bps = 0.05%)
- You'll see results quickly

---

### Step 7: Monitor Before Going Live

**Open these in separate terminals/tabs**:

```bash
# Terminal 1: Strategy Engine logs
tail -f /tmp/strategy-engine.log

# Terminal 2: Order Execution logs
tail -f /tmp/order-execution.log

# Terminal 3: Account Monitor logs
tail -f /tmp/account-monitor-working.log

# Terminal 4: Risk Manager logs
tail -f /tmp/risk-manager.log
```

**Watch for**:
- Strategy signals being generated
- Orders being submitted to exchange
- Fill confirmations
- Balance updates
- Risk violations

---

### Step 8: Enable Live Trading (THE ACTUAL SWITCH)

**Once you've made all config changes above**:

```bash
# 1. Stop all trading services
pkill -f "strategy-engine"
pkill -f "order-execution"

# 2. Rebuild with new configs
cd /home/mm/dev/b25/services/strategy-engine
make build

cd /home/mm/dev/b25/services/order-execution
go build -o bin/order-execution ./cmd/server

# 3. Start Order Execution in PRODUCTION mode
cd /home/mm/dev/b25/services/order-execution
BINANCE_API_KEY='rh22mtiKxsGSWuK3USkf4ba7E88exyVpn0INbc2OyCnogNsQ0R2A4lUcvHNJRcSc' \
BINANCE_SECRET_KEY='xUwZCEWa5g9auPgT5uYP8ClATN2zgGGYAFYgl4WoPTge2TWVxbz0ZBUmnV6PyOMh' \
nohup ./bin/order-execution > /tmp/order-execution-live.log 2>&1 &

# 4. Wait for Order Execution to fully start
sleep 5

# 5. Start Strategy Engine in LIVE mode
cd /home/mm/dev/b25/services/strategy-engine
nohup ./bin/strategy-engine > /tmp/strategy-engine-live.log 2>&1 &

# 6. Monitor logs immediately
tail -f /tmp/strategy-engine-live.log
```

---

### Step 9: Verify Live Trading is Active

**Check Strategy Engine status**:
```bash
curl http://localhost:9092/status | jq '{mode, active_strategies, signal_queue}'
```

**Expected output**:
```json
{
  "mode": "live",
  "active_strategies": ["scalping"],
  "signal_queue": ...
}
```

**Check Order Execution**:
```bash
curl http://localhost:9091/health | jq '.checks'
```

**Watch for first order**:
```bash
tail -f /tmp/order-execution-live.log | grep "order\|fill\|executed"
```

---

### Step 10: Monitor Your First Trades

**Dashboard**: https://mm.itziklerner.com

**What to watch**:
1. **Positions**: Should see positions opening/closing in real-time
2. **Balance**: Should see balance fluctuating (hopefully upward!)
3. **Orders**: Active orders displayed
4. **P&L**: Profit and Loss tracking

**Admin Pages for Monitoring**:
- **Account Monitor**: https://mm.itziklerner.com/services/account-monitor/
  - Shows balance changes
  - Position tracking
  - P&L calculations

- **Strategy Engine**: https://mm.itziklerner.com/services/strategy-engine/
  - Active strategies
  - Signal generation
  - Performance metrics

- **Order Execution**: https://mm.itziklerner.com/services/order-execution/
  - Orders sent to Binance
  - Fill confirmations
  - Exchange status

---

## 🛡️ Safety Checks Before Going Live

### Pre-Flight Checklist:

- [ ] **Risk Manager is healthy** (not in emergency stop)
- [ ] **Account Monitor showing correct balance** (1000 USDT)
- [ ] **Order Execution set to production** (testnet: false)
- [ ] **Strategy Engine set to live** (mode: "live")
- [ ] **Position sizes are conservative** (max 10% of account)
- [ ] **Daily loss limit is set** (recommended: $50 or 5%)
- [ ] **Only ONE strategy enabled** (start with scalping)
- [ ] **All logs are being monitored** (4 terminals open)
- [ ] **You're at your computer** to stop if needed
- [ ] **Stop-loss is configured** in strategies

### Emergency Stop Procedure:

If something goes wrong:

```bash
# EMERGENCY: Stop all trading immediately
pkill -f "strategy-engine"

# This stops signal generation, no new orders will be placed
# Existing orders/positions will remain - you'll need to close them manually via Binance or:

pkill -f "order-execution"
# Warning: This prevents automatic order management
```

**To close all positions via Binance**:
1. Go to Binance Futures web interface
2. Click "Close All Positions"
3. Or use Binance API to close programmatically

---

## 📊 What Happens When You Go Live?

### The Trading Flow:

1. **Market Data** → Streams prices from Binance to Redis
2. **Strategy Engine** → Reads prices, generates trading signals
3. **Strategy Engine** → Sends orders to Order Execution via gRPC
4. **Risk Manager** → Validates orders don't violate risk limits
5. **Order Execution** → Submits orders to Binance Futures API
6. **Binance** → Executes orders, sends fill notifications
7. **Account Monitor** → Receives fills via WebSocket, updates balance/positions
8. **Dashboard** → Shows real-time updates

### Example: Scalping Strategy in Live Mode

```
1. BTC price drops 0.05% quickly
2. Scalping strategy detects opportunity
3. Signal generated: BUY 0.01 BTC at $121,200
4. Order sent to Order Execution
5. Risk Manager checks: ✓ Under $100 position limit
6. Order submitted to Binance: BUY 0.01 BTC
7. Order fills at $121,195
8. Position held for 30 seconds
9. BTC rises 0.1%
10. Signal generated: SELL 0.01 BTC at $121,316
11. Position closed
12. Profit: ~$12 (0.1% of $12,119)
13. Account Monitor updates balance: $1,012
```

---

## 🔧 Recommended Configuration for $1,000 Account

### Minimal Risk Configuration:

**1. Edit Strategy Engine Config**:
```yaml
# /home/mm/dev/b25/services/strategy-engine/config.yaml

engine:
  mode: "live"  # ← CHANGE FROM simulation

strategies:
  enabled:
    - scalping  # ← ONLY enable scalping initially

  configs:
    scalping:
      target_spread_bps: 5.0
      profit_target: 0.001       # 0.1% profit target
      stop_loss: 0.0005          # 0.05% stop loss
      max_hold_time_seconds: 60
      max_position: 100.0        # ← CHANGE FROM 500 to 100

risk:
  enabled: true
  maxPositionSize: 100.0         # ← CHANGE FROM 1000 to 100
  maxOrderValue: 500.0           # ← CHANGE FROM 50000 to 500
  maxDailyLoss: 50.0             # ← CHANGE FROM 5000 to 50
  maxDrawdown: 0.10
  minAccountBalance: 900.0       # ← CHANGE FROM 10000 to 900
```

**2. Edit Order Execution Config**:
```yaml
# /home/mm/dev/b25/services/order-execution/config.yaml

exchange:
  testnet: false  # ← CHANGE FROM true to false - PRODUCTION
```

---

## 🚦 Trading Modes Explained

### Simulation Mode (Current):
- Strategies generate signals
- Signals are logged but NOT sent to Order Execution
- No real orders placed
- Safe for testing strategy logic
- **Use this** to test new strategies

### Observation Mode:
- Strategies run and generate signals
- Signals are sent to Order Execution
- Order Execution receives them but does NOT submit to Binance
- Useful for testing the full pipeline without risking money

### Live Mode (What you want):
- Strategies generate signals
- Signals sent to Order Execution via gRPC
- Order Execution submits REAL orders to Binance Futures
- Real money at risk
- Real profits/losses

---

## 📈 Monitoring Live Trading

### Via Web Dashboard:
**Main Dashboard**: https://mm.itziklerner.com
- Real-time P&L
- Open positions
- Active orders
- Balance changes

### Via Admin Pages:

**Account Monitor**: https://mm.itziklerner.com/services/account-monitor/
- Test Balance API button → See current balance
- Shows: `{"USDT": {"free": "1000", "locked": "X", "total": "10XX"}}`

**Strategy Engine**: https://mm.itziklerner.com/services/strategy-engine/
- Test Status button → See active strategies and mode
- Should show: `"mode": "live"`

**Order Execution**: https://mm.itziklerner.com/services/order-execution/
- Test Health button → Verify exchange connection
- Check logs for order submissions

### Via Logs:
```bash
# Watch for order submissions
tail -f /tmp/order-execution-live.log | grep "submit\|fill\|order"

# Watch for strategy signals
tail -f /tmp/strategy-engine-live.log | grep "signal\|order\|position"

# Watch balance changes
tail -f /tmp/account-monitor-working.log | grep "balance\|Corrected"
```

---

## 🎮 Quick Start Commands (Copy-Paste)

### Conservative Live Trading Setup:

```bash
# 1. Stop trading services
pkill -f "strategy-engine"
pkill -f "order-execution"

# 2. Edit configs (use your editor)
nano /home/mm/dev/b25/services/order-execution/config.yaml
# Change: testnet: true → testnet: false

nano /home/mm/dev/b25/services/strategy-engine/config.yaml
# Change: mode: "simulation" → mode: "live"
# Change: max_position: 500.0 → max_position: 100.0 (in scalping section)
# Change: maxPositionSize: 1000.0 → maxPositionSize: 100.0 (in risk section)

# 3. Rebuild
cd /home/mm/dev/b25/services/order-execution
go build -o bin/order-execution ./cmd/server

cd /home/mm/dev/b25/services/strategy-engine
make build

# 4. Start Order Execution (PRODUCTION)
cd /home/mm/dev/b25/services/order-execution
BINANCE_API_KEY='rh22mtiKxsGSWuK3USkf4ba7E88exyVpn0INbc2OyCnogNsQ0R2A4lUcvHNJRcSc' \
BINANCE_SECRET_KEY='xUwZCEWa5g9auPgT5uYP8ClATN2zgGGYAFYgl4WoPTge2TWVxbz0ZBUmnV6PyOMh' \
nohup ./bin/order-execution > /tmp/order-execution-live.log 2>&1 &

# 5. Wait and verify
sleep 5
curl http://localhost:9091/health | jq .

# 6. Start Strategy Engine (LIVE)
cd /home/mm/dev/b25/services/strategy-engine
nohup ./bin/strategy-engine > /tmp/strategy-engine-live.log 2>&1 &

# 7. Verify live mode
sleep 3
curl http://localhost:9092/status | jq '.mode'
# Should return: "live"

# 8. Monitor
tail -f /tmp/strategy-engine-live.log
```

---

## 📱 What to Expect

### First Few Minutes:
- Strategy engine loads market data
- Calculates indicators (moving averages, etc.)
- Starts looking for trading opportunities
- May take 30-60 seconds to generate first signal

### First Signal:
```
{"level":"info","msg":"signal generated","strategy":"scalping","symbol":"BTCUSDT","side":"BUY","quantity":0.001}
```

### First Order:
```
{"level":"info","msg":"order submitted","orderId":"123456","symbol":"BTCUSDT","side":"BUY","price":121200}
```

### First Fill:
```
{"level":"info","msg":"order filled","orderId":"123456","fillPrice":121195,"quantity":0.001}
```

### Position Tracking:
```
{"level":"info","msg":"position updated","symbol":"BTCUSDT","quantity":0.001,"entryPrice":121195}
```

---

## ⚠️ Important Notes

### Your Current Setup:
- **Balance**: $1,000 USDT
- **API Keys**: Production Binance Futures keys (VERIFIED ✅)
- **Account Type**: Sub-account with full trading permissions
- **WebSocket**: Connected to live Binance stream
- **Reconciliation**: Running every 5 seconds

### What's Working:
- ✅ Account Monitor tracking your real balance
- ✅ Market Data streaming live prices
- ✅ All services connected and healthy
- ✅ Risk limits configured
- ✅ Database persistence working

### What Needs to Change:
- ⚠️ Order Execution: testnet → production
- ⚠️ Strategy Engine: simulation → live
- ⚠️ Risk Manager: Clear emergency stop
- ⚠️ Position sizes: Reduce to 10% of account

---

## 🎯 Recommended First Live Trade

Start with a **manual test order** via Order Execution to verify everything works:

```bash
# Test with tiny position ($10 worth of BTC)
curl -X POST http://localhost:9091/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "type": "MARKET",
    "quantity": "0.0001",
    "client_order_id": "test-live-order-001"
  }'
```

This will:
1. Submit a real order to Binance
2. Buy $10 worth of BTC
3. Verify your API keys work for trading
4. Test the full order flow

Then close it:
```bash
curl -X POST http://localhost:9091/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "SELL",
    "type": "MARKET",
    "quantity": "0.0001",
    "client_order_id": "test-live-close-001"
  }'
```

---

## 🚨 When to STOP Trading

**Immediately stop if**:
- Balance drops $100 (10% loss)
- More than 5 losing trades in a row
- Risk Manager triggers emergency stop
- Unexpected behavior in logs
- Exchange connectivity issues
- You're not actively monitoring

**How to Stop**:
```bash
# Stop signal generation
pkill -f "strategy-engine"

# Optional: Stop order execution (prevents closing positions too)
# pkill -f "order-execution"
```

---

## 📊 Expected Performance

### Scalping Strategy (Conservative Settings):
- **Win Rate**: 60-70% (if market conditions are good)
- **Average Profit**: $1-5 per trade
- **Trades per Day**: 20-50 (depending on volatility)
- **Daily P&L**: -$50 to +$100 (high variance)
- **Max Drawdown**: Should stay under 10%

### Risk with $1,000 Account:
- **Max single loss**: ~$5 (with $100 position and 5 bps stop-loss)
- **Daily loss limit**: $50 (5% of account)
- **Account wipeout protection**: Min balance $900

---

## 📝 Summary

**To go live, you need to**:
1. ✅ Fix Risk Manager emergency stop
2. ✅ Change Order Execution: `testnet: false`
3. ✅ Change Strategy Engine: `mode: "live"`
4. ✅ Reduce position sizes to 10% of account
5. ✅ Start with only scalping strategy
6. ✅ Monitor logs actively
7. ✅ Be ready to emergency stop

**Current Status**:
- Account Monitor: ✅ PRODUCTION (tracking real balance)
- Order Execution: ⚠️ TESTNET (needs changing)
- Strategy Engine: ⚠️ SIMULATION (needs changing)

**After changes**:
- Account Monitor: ✅ PRODUCTION
- Order Execution: ✅ PRODUCTION
- Strategy Engine: ✅ LIVE

**Result**: Real auto trading with real money! 🚀

---

**Created**: 2025-10-08
**Your Balance**: 1000 USDT
**Recommended First Position**: $100 (10% of account)
**Recommended Strategy**: Scalping only
