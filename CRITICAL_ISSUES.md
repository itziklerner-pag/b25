# Critical Issues Found

## 1. Static Prices (Not Updating)

**Problem:** Prices show initially but don't update
**Cause:** Unknown - need to debug WebSocket message flow in production
**Check:** Browser console on https://mm.itziklerner.com

## 2. ServiceMonitor Shows "Degraded"

**Problem:** Market Data Service shows degraded status
**Cause:** ServiceMonitor tries to fetch `/proxy/market-data/health` which:
- Works in Vite dev server (has proxy)
- Fails in production build (no Vite proxy)
- Nginx needs to proxy `/proxy/*` routes

**Fix Needed:** Add Nginx location for `/proxy/market-data/*`

## 3. WebSocket Data Flow

**Need to verify:**
1. Is Dashboard Server sending incremental updates?
   - Check: `tail -f logs/dashboard-server.log | grep Broadcasting`
   
2. Is browser receiving updates?
   - Check browser console for `[WebSocket] Incremental update`
   
3. Is store being updated?
   - Check browser console for `[Store] Updated market data`
   
4. Are components re-rendering?
   - Check browser console for `[MarketPrices] Component rendering`

## Quick Test

Open browser console at https://mm.itziklerner.com and look for:
- WebSocket connection logs
- Update message logs
- Store update logs
- Component render logs

If you don't see logs, the logging was stripped in production build.
