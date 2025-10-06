import { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useTradingStore } from '@/store/trading';
import {
  Activity,
  ArrowLeft,
  ArrowDown,
  ArrowUp,
  CheckCircle,
  XCircle,
  TrendingUp,
  Zap,
  Clock,
  Wifi,
  WifiOff,
  Radio,
} from 'lucide-react';
import { cn, formatNumber, formatTimestamp } from '@/lib/utils';
import ReactECharts from 'echarts-for-react';

interface PriceData {
  symbol: string;
  lastPrice: number;
  priceChange24h: number;
  bid?: number;
  ask?: number;
  volume24h?: number;
  high24h?: number;
  low24h?: number;
  lastUpdate: number;
  sparklineData: number[];
}

interface UpdateMetrics {
  totalUpdates: number;
  lastUpdateTime: number;
  updatesPerSecond: number;
}

const SYMBOLS = ['BTCUSDT', 'ETHUSDT', 'SOLUSDT', 'BNBUSDT'];
const WEBSOCKET_URL = 'wss://mm.itziklerner.com/ws';

export default function MarketDataServicePage() {
  const navigate = useNavigate();

  // Subscribe to WebSocket data from Zustand store
  const marketData = useTradingStore((state) => state.marketData);
  const lastUpdate = useTradingStore((state) => state.lastUpdate);
  const status = useTradingStore((state) => state.status);
  const latency = useTradingStore((state) => state.latency);

  // Local state for sparkline data and metrics
  const [priceHistory, setPriceHistory] = useState<{ [symbol: string]: number[] }>({});
  const [updateMetrics, setUpdateMetrics] = useState<{ [symbol: string]: UpdateMetrics }>({});
  const [lastPriceUpdate, setLastPriceUpdate] = useState<{ [symbol: string]: number }>({});

  // Track price history for sparklines (WebSocket-driven)
  useEffect(() => {
    SYMBOLS.forEach((symbol) => {
      const data = marketData.get(symbol);
      if (data?.last_price) {
        const currentTime = Date.now();

        setPriceHistory((prev) => {
          const history = prev[symbol] || [];
          const lastPrice = history[history.length - 1];

          // Only add if price changed
          if (lastPrice !== data.last_price) {
            const newHistory = [...history, data.last_price].slice(-50); // Keep last 50 points
            return { ...prev, [symbol]: newHistory };
          }
          return prev;
        });

        // Update metrics
        setUpdateMetrics((prev) => {
          const metrics = prev[symbol] || { totalUpdates: 0, lastUpdateTime: currentTime, updatesPerSecond: 0 };
          const timeSinceLastUpdate = currentTime - metrics.lastUpdateTime;
          const newTotal = metrics.totalUpdates + 1;

          // Calculate updates per second (rolling average)
          const updatesPerSecond = timeSinceLastUpdate > 0
            ? Math.round((1000 / timeSinceLastUpdate) * 10) / 10
            : metrics.updatesPerSecond;

          return {
            ...prev,
            [symbol]: {
              totalUpdates: newTotal,
              lastUpdateTime: currentTime,
              updatesPerSecond: Math.max(0, Math.min(updatesPerSecond, 100)), // Cap at 100 updates/sec
            }
          };
        });

        setLastPriceUpdate((prev) => ({ ...prev, [symbol]: currentTime }));
      }
    });
  }, [marketData, lastUpdate]);

  const getPriceData = (symbol: string): PriceData => {
    const data = marketData.get(symbol);
    return {
      symbol,
      lastPrice: data?.last_price || 0,
      priceChange24h: data?.price_change_24h || 0,
      bid: data?.bid,
      ask: data?.ask,
      volume24h: data?.volume_24h,
      high24h: data?.high_24h,
      low24h: data?.low_24h,
      lastUpdate: data?.timestamp || 0,
      sparklineData: priceHistory[symbol] || [],
    };
  };

  const getSparklineOption = (data: number[]) => {
    if (data.length < 2) return null;

    const isPositive = data[data.length - 1] >= data[0];

    return {
      grid: { top: 0, bottom: 0, left: 0, right: 0 },
      xAxis: { type: 'category', show: false, data: data.map((_, i) => i) },
      yAxis: { type: 'value', show: false },
      series: [
        {
          type: 'line',
          data,
          smooth: true,
          symbol: 'none',
          lineStyle: { color: isPositive ? '#10b981' : '#ef4444', width: 2 },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: isPositive ? 'rgba(16, 185, 129, 0.4)' : 'rgba(239, 68, 68, 0.4)' },
                { offset: 1, color: isPositive ? 'rgba(16, 185, 129, 0.05)' : 'rgba(239, 68, 68, 0.05)' },
              ],
            },
          },
        },
      ],
    };
  };

  const getSecondsSinceUpdate = (timestamp: number): number => {
    if (!timestamp) return 0;
    return Math.floor((Date.now() - timestamp) / 1000);
  };

  const isConnected = status === 'connected';
  const symbolCount = marketData.size;
  const totalUpdates = useMemo(() => {
    return Object.values(updateMetrics).reduce((sum, m) => sum + m.totalUpdates, 0);
  }, [updateMetrics]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="outline" size="icon" onClick={() => navigate('/system')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold">Market Data Service</h1>
            <p className="text-muted-foreground">Real-time WebSocket market data feed</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isConnected ? (
            <Badge variant="success" className="text-sm px-3 py-1 gap-2">
              <Radio className="h-3 w-3 animate-pulse" />
              LIVE
            </Badge>
          ) : (
            <Badge variant="destructive" className="text-sm px-3 py-1">
              DISCONNECTED
            </Badge>
          )}
        </div>
      </div>

      {/* WebSocket Status Overview */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">WebSocket Status</CardTitle>
            {isConnected ? (
              <CheckCircle className="h-4 w-4 text-green-500" />
            ) : (
              <XCircle className="h-4 w-4 text-red-500" />
            )}
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold capitalize">{status}</div>
            <p className="text-xs text-muted-foreground">
              {isConnected ? 'Receiving live data' : 'No connection'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Connection Latency</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{latency > 0 ? `${latency}ms` : '-'}</div>
            <p className="text-xs text-muted-foreground">Round-trip time</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Last Update</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {getSecondsSinceUpdate(lastUpdate)}s ago
            </div>
            <p className="text-xs text-muted-foreground">{formatTimestamp(lastUpdate)}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Symbols Tracked</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{symbolCount}</div>
            <p className="text-xs text-muted-foreground">Active price feeds</p>
          </CardContent>
        </Card>
      </div>

      {/* Live Price Feed */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Activity className={cn("h-5 w-5", isConnected && "animate-pulse text-green-500")} />
              Live Price Feed
            </CardTitle>
            <div className="flex items-center gap-3">
              {isConnected && (
                <Badge variant="outline" className="font-mono text-xs">
                  <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse mr-2" />
                  {totalUpdates.toLocaleString()} updates
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {SYMBOLS.map((symbol) => {
              const priceData = getPriceData(symbol);
              const isPositive = (priceData.priceChange24h || 0) >= 0;
              const sparklineOption = getSparklineOption(priceData.sparklineData);
              const metrics = updateMetrics[symbol];
              const secondsSinceUpdate = getSecondsSinceUpdate(lastPriceUpdate[symbol] || 0);
              const isRecentlyUpdated = secondsSinceUpdate < 2;

              return (
                <Card
                  key={symbol}
                  className={cn(
                    'border-2 transition-all duration-300',
                    isPositive ? 'border-green-500/20' : 'border-red-500/20',
                    isRecentlyUpdated && 'ring-2 ring-blue-500/50 shadow-lg'
                  )}
                >
                  <CardContent className="pt-6">
                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="font-bold text-lg">{symbol.replace('USDT', '/USDT')}</span>
                        <div className="flex items-center gap-2">
                          {isConnected && (
                            <div className={cn(
                              "h-2 w-2 rounded-full transition-all",
                              isRecentlyUpdated ? "bg-green-500 animate-pulse" : "bg-gray-400"
                            )} />
                          )}
                          {isPositive ? (
                            <ArrowUp className="h-5 w-5 text-green-500" />
                          ) : (
                            <ArrowDown className="h-5 w-5 text-red-500" />
                          )}
                        </div>
                      </div>

                      <div>
                        <div className="text-3xl font-bold">
                          ${formatNumber(priceData.lastPrice, priceData.lastPrice > 1000 ? 2 : 4)}
                        </div>
                        <div className={cn('text-sm font-medium', isPositive ? 'text-green-500' : 'text-red-500')}>
                          {isPositive ? '+' : ''}
                          {(priceData.priceChange24h || 0).toFixed(2)}%
                        </div>
                      </div>

                      {sparklineOption && (
                        <div className="h-16 -mx-2">
                          <ReactECharts
                            option={sparklineOption}
                            style={{ height: '100%' }}
                            opts={{ renderer: 'canvas' }}
                          />
                        </div>
                      )}

                      {priceData.bid && priceData.ask && (
                        <div className="grid grid-cols-2 gap-2 text-xs border-t pt-2">
                          <div>
                            <div className="text-muted-foreground">Bid</div>
                            <div className="font-medium">${formatNumber(priceData.bid, priceData.bid > 1000 ? 2 : 4)}</div>
                          </div>
                          <div>
                            <div className="text-muted-foreground">Ask</div>
                            <div className="font-medium">${formatNumber(priceData.ask, priceData.ask > 1000 ? 2 : 4)}</div>
                          </div>
                        </div>
                      )}

                      {priceData.high24h && priceData.low24h && (
                        <div className="grid grid-cols-2 gap-2 text-xs border-t pt-2">
                          <div>
                            <div className="text-muted-foreground">24h High</div>
                            <div className="font-medium text-green-600">${formatNumber(priceData.high24h, priceData.high24h > 1000 ? 2 : 4)}</div>
                          </div>
                          <div>
                            <div className="text-muted-foreground">24h Low</div>
                            <div className="font-medium text-red-600">${formatNumber(priceData.low24h, priceData.low24h > 1000 ? 2 : 4)}</div>
                          </div>
                        </div>
                      )}

                      {priceData.volume24h && (
                        <div className="text-xs text-muted-foreground border-t pt-2">
                          24h Vol: {formatNumber(priceData.volume24h, 0)}
                        </div>
                      )}

                      {metrics && (
                        <div className="text-xs text-muted-foreground border-t pt-2 flex items-center justify-between">
                          <span>{metrics.totalUpdates} updates</span>
                          {metrics.updatesPerSecond > 0 && (
                            <span className="text-green-600 font-medium">
                              {metrics.updatesPerSecond.toFixed(1)} u/s
                            </span>
                          )}
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* WebSocket Connection Info */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {isConnected ? (
              <Wifi className="h-5 w-5 text-green-500" />
            ) : (
              <WifiOff className="h-5 w-5 text-red-500" />
            )}
            WebSocket Connection
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-3 rounded-lg border bg-muted/50">
              <div className="flex items-center gap-3">
                <div className={cn(
                  "h-3 w-3 rounded-full",
                  isConnected ? "bg-green-500 animate-pulse" : "bg-red-500"
                )} />
                <div>
                  <div className="font-medium">WebSocket Server</div>
                  <div className="text-xs text-muted-foreground font-mono">{WEBSOCKET_URL}</div>
                </div>
              </div>
              <Badge variant={isConnected ? 'success' : 'destructive'}>
                {isConnected ? 'Connected' : 'Disconnected'}
              </Badge>
            </div>

            <div className="grid gap-3 md:grid-cols-3">
              <div className="p-3 rounded-lg border">
                <div className="text-sm text-muted-foreground">Connection Status</div>
                <div className="text-lg font-bold capitalize mt-1">{status}</div>
              </div>
              <div className="p-3 rounded-lg border">
                <div className="text-sm text-muted-foreground">Latency</div>
                <div className="text-lg font-bold mt-1">
                  {latency > 0 ? `${latency}ms` : '-'}
                </div>
              </div>
              <div className="p-3 rounded-lg border">
                <div className="text-sm text-muted-foreground">Active Streams</div>
                <div className="text-lg font-bold mt-1">{symbolCount} symbols</div>
              </div>
            </div>

            <div className="text-sm text-muted-foreground p-3 rounded-lg bg-blue-500/10 border border-blue-500/20">
              <strong className="text-blue-600">Real-time Data:</strong> This page displays live market data
              received via WebSocket. No polling or HTTP requests are used. All updates are
              pushed from the server in real-time.
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Data Feed Statistics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Data Feed Statistics
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {SYMBOLS.map((symbol) => {
              const metrics = updateMetrics[symbol];
              const priceData = getPriceData(symbol);
              const secondsSinceUpdate = getSecondsSinceUpdate(lastPriceUpdate[symbol] || 0);

              return (
                <div key={symbol} className="flex items-center justify-between p-3 rounded-lg border">
                  <div className="flex items-center gap-3">
                    <div className={cn(
                      "h-2 w-2 rounded-full",
                      secondsSinceUpdate < 2 ? "bg-green-500 animate-pulse" : "bg-gray-400"
                    )} />
                    <div>
                      <div className="font-medium">{symbol.replace('USDT', '/USDT')}</div>
                      <div className="text-xs text-muted-foreground">
                        {priceData.lastUpdate ? formatTimestamp(priceData.lastUpdate) : 'No data'}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    {metrics ? (
                      <>
                        <div className="text-lg font-bold">{metrics.totalUpdates.toLocaleString()}</div>
                        <div className="text-xs text-muted-foreground">
                          {metrics.updatesPerSecond > 0
                            ? `${metrics.updatesPerSecond.toFixed(1)} updates/sec`
                            : 'updates received'}
                        </div>
                      </>
                    ) : (
                      <div className="text-sm text-muted-foreground">Waiting for data...</div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Service Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Data Source Configuration</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Data Source</span>
                <span className="font-medium">WebSocket Only</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Exchange</span>
                <span className="font-medium">Binance Futures</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Update Method</span>
                <span className="font-medium">Real-time Push</span>
              </div>
            </div>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Polling</span>
                <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-500/20">
                  Disabled
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">HTTP Requests</span>
                <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-500/20">
                  None
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Sparkline Points</span>
                <span className="font-mono font-medium">50</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
