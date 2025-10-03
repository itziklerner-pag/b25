import { useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useTradingStore } from '@/store/trading';
import { formatCurrency, formatNumber } from '@/lib/utils';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';

export default function OrderBookPage() {
  const selectedSymbol = useTradingStore((state) => state.selectedSymbol);
  const orderbook = useTradingStore((state) => state.orderbooks.get(selectedSymbol));
  const trades = useTradingStore((state) => state.trades.slice(0, 20));

  const chartOption: EChartsOption = useMemo(() => {
    if (!orderbook) return {};

    const bids = orderbook.bids.slice(0, 20).map((level, i) => ({
      price: level.price,
      cumulative: orderbook.bids
        .slice(0, i + 1)
        .reduce((sum, l) => sum + l.volume, 0),
    })).reverse();

    const asks = orderbook.asks.slice(0, 20).map((level, i) => ({
      price: level.price,
      cumulative: orderbook.asks
        .slice(0, i + 1)
        .reduce((sum, l) => sum + l.volume, 0),
    }));

    const allPrices = [...bids.map((b) => b.price), ...asks.map((a) => a.price)];

    return {
      grid: { left: '3%', right: '3%', bottom: '10%', top: '10%', containLabel: true },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
      },
      xAxis: {
        type: 'category',
        data: allPrices,
        axisLabel: { formatter: (value: string | number) => typeof value === 'number' ? value.toFixed(2) : value },
      },
      yAxis: {
        type: 'value',
        axisLabel: { formatter: (value: number) => value.toFixed(2) },
      },
      series: [
        {
          name: 'Bids',
          type: 'line',
          data: bids.map((b) => b.cumulative),
          itemStyle: { color: '#10b981' },
          areaStyle: { color: 'rgba(16, 185, 129, 0.2)' },
          smooth: false,
          step: 'end',
        },
        {
          name: 'Asks',
          type: 'line',
          data: [...Array(bids.length).fill(0), ...asks.map((a) => a.cumulative)],
          itemStyle: { color: '#ef4444' },
          areaStyle: { color: 'rgba(239, 68, 68, 0.2)' },
          smooth: false,
          step: 'start',
        },
      ],
    };
  }, [orderbook]);

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Order Book Depth - {selectedSymbol}</CardTitle>
          </CardHeader>
          <CardContent>
            {orderbook ? (
              <ReactECharts option={chartOption} style={{ height: 400 }} />
            ) : (
              <div className="flex h-96 items-center justify-center text-muted-foreground">
                Loading order book...
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Order Book</CardTitle>
          </CardHeader>
          <CardContent>
            {orderbook ? (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="mb-2 font-medium text-green-500">Bids</h3>
                  <div className="space-y-1">
                    {orderbook.bids.slice(0, 10).map((level, i) => (
                      <div key={i} className="flex justify-between text-sm">
                        <span className="text-green-500">{formatCurrency(level.price)}</span>
                        <span className="text-muted-foreground">{formatNumber(level.volume, 4)}</span>
                      </div>
                    ))}
                  </div>
                </div>
                <div>
                  <h3 className="mb-2 font-medium text-red-500">Asks</h3>
                  <div className="space-y-1">
                    {orderbook.asks.slice(0, 10).map((level, i) => (
                      <div key={i} className="flex justify-between text-sm">
                        <span className="text-red-500">{formatCurrency(level.price)}</span>
                        <span className="text-muted-foreground">{formatNumber(level.volume, 4)}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            ) : (
              <p className="text-center text-sm text-muted-foreground">No order book data</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Trades</CardTitle>
        </CardHeader>
        <CardContent>
          {trades.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No recent trades</p>
          ) : (
            <div className="space-y-2">
              {trades.map((trade) => (
                <div key={trade.id} className="flex items-center justify-between border-b pb-2">
                  <div>
                    <div className="font-medium">{trade.symbol}</div>
                    <div className="text-sm text-muted-foreground">
                      {new Date(trade.timestamp).toLocaleTimeString()}
                    </div>
                  </div>
                  <div className="text-right">
                    <div
                      className={`font-medium ${
                        trade.side === 'BUY' ? 'text-green-500' : 'text-red-500'
                      }`}
                    >
                      {formatCurrency(trade.price)}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {formatNumber(trade.quantity, 4)}
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
