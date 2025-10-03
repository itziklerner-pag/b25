import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useTradingStore } from '@/store/trading';
import { formatCurrency } from '@/lib/utils';

export default function AnalyticsPage() {
  const account = useTradingStore((state) => state.account);
  const positions = useTradingStore((state) => Array.from(state.positions.values()));

  const totalUnrealizedPnl = positions.reduce((sum, p) => sum + p.unrealizedPnl, 0);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Total P&L (24h)</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className={`text-3xl font-bold ${
                (account?.totalPnl24h || 0) >= 0 ? 'text-green-500' : 'text-red-500'
              }`}
            >
              {formatCurrency(account?.totalPnl24h || 0)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Unrealized P&L</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className={`text-3xl font-bold ${
                totalUnrealizedPnl >= 0 ? 'text-green-500' : 'text-red-500'
              }`}
            >
              {formatCurrency(totalUnrealizedPnl)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Realized P&L</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className={`text-3xl font-bold ${
                (account?.realizedPnl || 0) >= 0 ? 'text-green-500' : 'text-red-500'
              }`}
            >
              {formatCurrency(account?.realizedPnl || 0)}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>P&L by Position</CardTitle>
        </CardHeader>
        <CardContent>
          {positions.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No positions to analyze</p>
          ) : (
            <div className="space-y-3">
              {positions.map((position) => (
                <div key={position.symbol} className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <div className="font-medium">{position.symbol}</div>
                    <div className="text-sm text-muted-foreground">
                      {position.side} {position.size} @ {formatCurrency(position.entryPrice)}
                    </div>
                  </div>
                  <div className="text-right">
                    <div
                      className={`text-lg font-bold ${
                        position.unrealizedPnl >= 0 ? 'text-green-500' : 'text-red-500'
                      }`}
                    >
                      {formatCurrency(position.unrealizedPnl)}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {((position.unrealizedPnl / (position.entryPrice * position.size)) * 100).toFixed(2)}%
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
