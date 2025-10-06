import { useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useTradingStore } from '@/store/trading';
import { formatCurrency, formatPercent } from '@/lib/utils';
import { TrendingUp, TrendingDown } from 'lucide-react';
import { logger } from '@/utils/logger';

export function MarketPrices() {
  // CRITICAL FIX: Subscribe to both marketData and lastUpdate to force re-renders
  // Using separate subscriptions ensures Zustand detects changes
  const marketData = useTradingStore((state) => state.marketData);
  const lastUpdate = useTradingStore((state) => state.lastUpdate);
  const debugInfo = useTradingStore((state) => state.debugInfo);

  // Convert Map to Array for rendering
  // IMPORTANT: This must depend on BOTH marketData AND lastUpdate
  const marketPairs = useMemo(() => {
    const pairs = Array.from(marketData.entries()).map(([symbol, data]) => ({
      symbol,
      ...data,
    }));

    logger.trace('MarketPrices', 'Converted marketData to array', {
      mapSize: marketData.size,
      pairsCount: pairs.length,
      lastUpdate: new Date(lastUpdate).toISOString(),
      updateCount: debugInfo.updateCount,
      pairs: pairs.map(p => ({ symbol: p.symbol, price: p.last_price })),
    });

    return pairs;
  }, [marketData, lastUpdate, debugInfo.updateCount]); // Triple dependency to force updates

  // Debug: Log when component renders
  logger.trace('MarketPrices', 'Component rendering', {
    marketPairsCount: marketPairs.length,
    lastUpdate: new Date(lastUpdate).toISOString(),
    updateCount: debugInfo.updateCount,
    timestamp: new Date().toISOString(),
  });

  if (marketPairs.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Market Prices</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Waiting for market data...</p>
          <div className="mt-2 text-xs text-muted-foreground">
            Updates: {debugInfo.updateCount} | Last: {new Date(lastUpdate).toLocaleTimeString()}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Market Prices</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {marketPairs.map((market) => {
            const priceChange = market.price_change_24h || 0;
            const isPositive = priceChange >= 0;

            return (
              <div
                key={market.symbol}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div>
                  <div className="font-medium">{market.symbol}</div>
                  <div className="text-sm text-muted-foreground">
                    {market.volume_24h && `Vol: ${formatCurrency(market.volume_24h)}`}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-xl font-bold">
                    {formatCurrency(market.last_price)}
                  </div>
                  {priceChange !== 0 && (
                    <div
                      className={`flex items-center justify-end gap-1 text-sm ${
                        isPositive ? 'text-green-500' : 'text-red-500'
                      }`}
                    >
                      {isPositive ? (
                        <TrendingUp className="h-3 w-3" />
                      ) : (
                        <TrendingDown className="h-3 w-3" />
                      )}
                      <span>{formatPercent(Math.abs(priceChange))}</span>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
        <div className="mt-4 text-xs text-muted-foreground">
          Last update: {new Date(lastUpdate).toLocaleTimeString()} | Updates: {debugInfo.updateCount}
        </div>
      </CardContent>
    </Card>
  );
}
