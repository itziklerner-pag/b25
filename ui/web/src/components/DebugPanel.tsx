import { useState, useEffect } from 'react';
import { useTradingStore } from '@/store/trading';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Bug, X, Trash2 } from 'lucide-react';
import { logger, type LogLevelString } from '@/utils/logger';

export function DebugPanel() {
  const [isVisible, setIsVisible] = useState(false);
  const [position, setPosition] = useState({ x: 20, y: 20 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [logLevel, setLogLevel] = useState<LogLevelString>(logger.getLevelString());

  // Subscribe to all relevant state
  const debugInfo = useTradingStore((state) => state.debugInfo);
  const lastUpdate = useTradingStore((state) => state.lastUpdate);
  const marketData = useTradingStore((state) => state.marketData);
  const status = useTradingStore((state) => state.status);
  const latency = useTradingStore((state) => state.latency);

  // Get BTC price for quick reference
  const btcData = marketData.get('BTCUSDT');

  // Keyboard shortcut to toggle (Ctrl+D or Cmd+D)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'd') {
        e.preventDefault();
        setIsVisible((prev) => !prev);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Update log level state when it changes
  useEffect(() => {
    const interval = setInterval(() => {
      const currentLevel = logger.getLevelString();
      if (currentLevel !== logLevel) {
        setLogLevel(currentLevel);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [logLevel]);

  // Handle dragging
  const handleMouseDown = (e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('.drag-handle')) {
      setIsDragging(true);
      setDragOffset({
        x: e.clientX - position.x,
        y: e.clientY - position.y,
      });
    }
  };

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (isDragging) {
        setPosition({
          x: e.clientX - dragOffset.x,
          y: e.clientY - dragOffset.y,
        });
      }
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
    }

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, dragOffset]);

  const handleLogLevelChange = (level: LogLevelString) => {
    logger.setLevel(level);
    setLogLevel(level);
    logger.info('DebugPanel', `Log level changed to ${level}`);
  };

  const handleClearConsole = () => {
    logger.clear();
    logger.info('DebugPanel', 'Console cleared');
  };

  const logLevels: LogLevelString[] = ['ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'];

  if (!isVisible) {
    return (
      <button
        onClick={() => setIsVisible(true)}
        className="fixed bottom-4 right-4 z-50 flex items-center gap-2 rounded-full bg-purple-600 px-4 py-2 text-white shadow-lg hover:bg-purple-700 transition-colors"
        title="Show Debug Panel (Ctrl+D)"
      >
        <Bug className="h-4 w-4" />
        Debug
      </button>
    );
  }

  return (
    <Card
      className="fixed z-50 w-96 shadow-2xl border-2 border-purple-500 cursor-move select-none max-h-[90vh] overflow-hidden flex flex-col"
      style={{
        left: `${position.x}px`,
        top: `${position.y}px`,
      }}
      onMouseDown={handleMouseDown}
    >
      <CardHeader className="drag-handle cursor-move bg-purple-600 text-white pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm flex items-center gap-2">
            <Bug className="h-4 w-4" />
            Debug Panel
          </CardTitle>
          <button
            onClick={() => setIsVisible(false)}
            className="hover:bg-purple-700 rounded p-1 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="text-xs opacity-80">Press Ctrl+D to toggle</div>
      </CardHeader>
      <CardContent className="pt-4 cursor-auto space-y-3 text-sm overflow-y-auto flex-1">
        {/* Log Level Control */}
        <div className="border-b pb-3">
          <div className="font-semibold mb-2 flex items-center justify-between">
            <span>Log Level</span>
            <Button
              size="sm"
              variant="outline"
              onClick={handleClearConsole}
              className="h-7 px-2 text-xs"
            >
              <Trash2 className="h-3 w-3 mr-1" />
              Clear Console
            </Button>
          </div>
          <div className="flex gap-1">
            {logLevels.map((level) => (
              <button
                key={level}
                onClick={() => handleLogLevelChange(level)}
                className={`flex-1 px-2 py-1 rounded text-xs font-medium transition-colors ${
                  logLevel === level
                    ? level === 'ERROR'
                      ? 'bg-red-500 text-white'
                      : level === 'WARN'
                      ? 'bg-amber-500 text-white'
                      : level === 'INFO'
                      ? 'bg-blue-500 text-white'
                      : level === 'DEBUG'
                      ? 'bg-violet-500 text-white'
                      : 'bg-gray-500 text-white'
                    : 'bg-muted text-muted-foreground hover:bg-muted/80'
                }`}
              >
                {level}
              </button>
            ))}
          </div>
          <div className="mt-2 text-xs text-muted-foreground">
            Current: <span className="font-mono font-semibold">{logLevel}</span>
            {' • '}
            Logs{' '}
            {logLevel === 'ERROR'
              ? 'only errors'
              : logLevel === 'WARN'
              ? 'errors & warnings'
              : logLevel === 'INFO'
              ? 'errors, warnings & info'
              : logLevel === 'DEBUG'
              ? 'all except trace'
              : 'everything'}
          </div>
        </div>

        {/* WebSocket Status */}
        <div className="border-b pb-2">
          <div className="font-semibold mb-1">WebSocket</div>
          <div className="grid grid-cols-2 gap-1 text-xs">
            <div className="text-muted-foreground">Status:</div>
            <div className="font-mono">
              <span
                className={`px-2 py-0.5 rounded ${
                  status === 'connected'
                    ? 'bg-green-500/20 text-green-500'
                    : 'bg-red-500/20 text-red-500'
                }`}
              >
                {status}
              </span>
            </div>
            <div className="text-muted-foreground">Latency:</div>
            <div className="font-mono">{latency}ms</div>
          </div>
        </div>

        {/* Store Updates */}
        <div className="border-b pb-2">
          <div className="font-semibold mb-1">Store Updates</div>
          <div className="grid grid-cols-2 gap-1 text-xs">
            <div className="text-muted-foreground">Update Count:</div>
            <div className="font-mono text-blue-500">{debugInfo.updateCount}</div>

            <div className="text-muted-foreground">Last WS Msg:</div>
            <div className="font-mono">
              {debugInfo.lastWsMessage
                ? new Date(debugInfo.lastWsMessage).toLocaleTimeString()
                : 'Never'}
            </div>

            <div className="text-muted-foreground">Last Store Update:</div>
            <div className="font-mono">
              {debugInfo.lastStoreUpdate
                ? new Date(debugInfo.lastStoreUpdate).toLocaleTimeString()
                : 'Never'}
            </div>

            <div className="text-muted-foreground">Store lastUpdate:</div>
            <div className="font-mono">
              {new Date(lastUpdate).toLocaleTimeString()}
            </div>
          </div>
        </div>

        {/* Market Data */}
        <div className="border-b pb-2">
          <div className="font-semibold mb-1">Market Data</div>
          <div className="grid grid-cols-2 gap-1 text-xs">
            <div className="text-muted-foreground">Map Size:</div>
            <div className="font-mono">{marketData.size}</div>

            <div className="text-muted-foreground">Last Symbol:</div>
            <div className="font-mono">{debugInfo.lastSymbol || 'None'}</div>

            <div className="text-muted-foreground">Last Price:</div>
            <div className="font-mono">
              ${debugInfo.lastPrice ? debugInfo.lastPrice.toFixed(2) : '0.00'}
            </div>
          </div>
        </div>

        {/* BTC Data */}
        {btcData && (
          <div className="border-b pb-2">
            <div className="font-semibold mb-1">BTC/USDT</div>
            <div className="grid grid-cols-2 gap-1 text-xs">
              <div className="text-muted-foreground">Price:</div>
              <div className="font-mono text-green-500">
                ${btcData.last_price?.toFixed(2) || '0.00'}
              </div>

              <div className="text-muted-foreground">24h Change:</div>
              <div
                className={`font-mono ${
                  (btcData.price_change_24h || 0) >= 0
                    ? 'text-green-500'
                    : 'text-red-500'
                }`}
              >
                {((btcData.price_change_24h || 0) * 100).toFixed(2)}%
              </div>

              <div className="text-muted-foreground">Volume:</div>
              <div className="font-mono">
                ${(btcData.volume_24h || 0).toLocaleString()}
              </div>

              <div className="text-muted-foreground">Timestamp:</div>
              <div className="font-mono text-xs">
                {btcData.timestamp
                  ? new Date(btcData.timestamp).toLocaleTimeString()
                  : 'N/A'}
              </div>
            </div>
          </div>
        )}

        {/* Current Time */}
        <div className="text-xs text-muted-foreground text-center pt-1">
          Current: {new Date().toLocaleTimeString()}
        </div>

        {/* All Symbols */}
        {marketData.size > 0 && (
          <div className="pt-2">
            <div className="font-semibold mb-1 text-xs">All Symbols ({marketData.size})</div>
            <div className="space-y-1 max-h-32 overflow-y-auto text-xs">
              {Array.from(marketData.entries()).map(([symbol, data]) => (
                <div
                  key={symbol}
                  className="flex justify-between items-center border rounded px-2 py-1"
                >
                  <span className="font-mono">{symbol}</span>
                  <span className="font-mono text-green-500">
                    ${data.last_price?.toFixed(2) || '0.00'}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
