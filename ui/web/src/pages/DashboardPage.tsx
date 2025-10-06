import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useTradingStore } from '@/store/trading';
import { formatCurrency, formatPercent } from '@/lib/utils';
import { TrendingUp, TrendingDown, DollarSign, Activity, Percent } from 'lucide-react';
import { MarketPrices } from '@/components/MarketPrices';
import { logger } from '@/utils/logger';

export default function DashboardPage() {
  const account = useTradingStore((state) => state.account);
  const positions = useTradingStore((state) => Array.from(state.positions.values()));
  const orders = useTradingStore((state) => Array.from(state.orders.values()));
  // Subscribe to lastUpdate to ensure re-renders when data changes
  const lastUpdate = useTradingStore((state) => state.lastUpdate);

  // Debug log
  logger.trace('DashboardPage', 'Rendering', {
    positionsCount: positions.length,
    ordersCount: orders.length,
    lastUpdate: new Date(lastUpdate).toISOString(),
  });

  const stats = [
    {
      title: 'Balance',
      value: formatCurrency(account?.balance || 0),
      icon: DollarSign,
      trend: null,
    },
    {
      title: 'Equity',
      value: formatCurrency(account?.equity || 0),
      icon: Activity,
      trend: null,
    },
    {
      title: 'Unrealized P&L',
      value: formatCurrency(account?.unrealizedPnl || 0),
      icon: account && account.unrealizedPnl >= 0 ? TrendingUp : TrendingDown,
      trend: account?.unrealizedPnl || 0,
    },
    {
      title: 'Margin Ratio',
      value: formatPercent(account?.marginRatio || 0),
      icon: Percent,
      trend: null,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Card key={stat.title}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
              <stat.icon
                className={`h-4 w-4 ${
                  stat.trend !== null
                    ? stat.trend >= 0
                      ? 'text-green-500'
                      : 'text-red-500'
                    : 'text-muted-foreground'
                }`}
              />
            </CardHeader>
            <CardContent>
              <div
                className={`text-2xl font-bold ${
                  stat.trend !== null
                    ? stat.trend >= 0
                      ? 'text-green-500'
                      : 'text-red-500'
                    : ''
                }`}
              >
                {stat.value}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Market Prices - NEW */}
        <MarketPrices />

        {/* Positions */}
        <Card>
          <CardHeader>
            <CardTitle>Open Positions ({positions.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {positions.length === 0 ? (
              <p className="text-sm text-muted-foreground">No open positions</p>
            ) : (
              <div className="space-y-2">
                {positions.slice(0, 5).map((position) => (
                  <div
                    key={position.symbol}
                    className="flex items-center justify-between rounded-lg border p-3"
                  >
                    <div>
                      <div className="font-medium">{position.symbol}</div>
                      <div className="text-sm text-muted-foreground">
                        {position.side} {position.size} @ {formatCurrency(position.entryPrice)}
                      </div>
                    </div>
                    <div className="text-right">
                      <div
                        className={`font-medium ${
                          position.unrealizedPnl >= 0 ? 'text-green-500' : 'text-red-500'
                        }`}
                      >
                        {formatCurrency(position.unrealizedPnl)}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {position.leverage}x leverage
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Recent Orders */}
        <Card>
          <CardHeader>
            <CardTitle>Active Orders ({orders.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {orders.length === 0 ? (
              <p className="text-sm text-muted-foreground">No active orders</p>
            ) : (
              <div className="space-y-2">
                {orders.slice(0, 5).map((order) => (
                  <div key={order.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div>
                      <div className="font-medium">{order.symbol}</div>
                      <div className="text-sm text-muted-foreground">
                        {order.side} {order.type}
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-medium">
                        {order.price ? formatCurrency(order.price) : 'Market'}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {order.filledQuantity}/{order.quantity}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
