import { useEffect, useRef, useCallback } from 'react';
import { useTradingStore } from '@/store/trading';
import type { WebSocketMessage } from '@/types';
import { logger } from '@/utils/logger';

interface UseWebSocketOptions {
  url: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

export function useWebSocket({
  url,
  reconnectInterval = 3000,
  maxReconnectAttempts = 10,
  heartbeatInterval = 30000,
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const heartbeatRef = useRef<number>();
  const reconnectTimeoutRef = useRef<number>();

  const setWs = useTradingStore((state) => state.setWs);
  const setStatus = useTradingStore((state) => state.setStatus);
  const setLatency = useTradingStore((state) => state.setLatency);
  const updateOrderBook = useTradingStore((state) => state.updateOrderBook);
  const updateMarketData = useTradingStore((state) => state.updateMarketData);
  const updatePosition = useTradingStore((state) => state.updatePosition);
  const updateOrder = useTradingStore((state) => state.updateOrder);
  const removeOrder = useTradingStore((state) => state.removeOrder);
  const addTrade = useTradingStore((state) => state.addTrade);
  const updateAccount = useTradingStore((state) => state.updateAccount);
  const updateSystemHealth = useTradingStore((state) => state.updateSystemHealth);
  const clearData = useTradingStore((state) => state.clearData);

  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);

        logger.debug('WebSocket', 'Received message', {
          type: message.type,
          channel: message.channel,
          timestamp: new Date().toISOString(),
        });

        // Handle pong for latency measurement
        if (message.type === 'pong' && message.timestamp) {
          const now = Date.now();
          setLatency(now - message.timestamp);
          logger.trace('WebSocket', 'Latency measured', { latency: now - message.timestamp });
          return;
        }

        // Route message to appropriate store action
        switch (message.type) {
          case 'snapshot':
          case 'full_state':
            // Handle initial snapshot with all data
            if (message.data) {
              logger.info('WebSocket', 'Processing full_state/snapshot', {
                hasMarketData: !!message.data.market_data,
                hasAccount: !!message.data.account,
                hasOrders: !!message.data.orders,
                hasPositions: !!message.data.positions,
              });

              // Update market data (prices)
              if (message.data.market_data) {
                logger.debug('WebSocket', 'Received market_data', message.data.market_data);
                Object.entries(message.data.market_data).forEach(([symbol, data]: [string, any]) => {
                  // Check if this is market data (has last_price) or full orderbook (has bids/asks)
                  if (data.last_price !== undefined) {
                    // This is market price data
                    logger.debug('WebSocket', `Updating market data for ${symbol}`, data);
                    updateMarketData(symbol, data);
                  } else if (data.bids && data.asks) {
                    // This is full orderbook data
                    logger.debug('WebSocket', `Updating orderbook for ${symbol}`);
                    updateOrderBook(symbol, { ...data, symbol, timestamp: Date.now() });
                  }
                });
              }
              // Update account
              if (message.data.account) {
                updateAccount(message.data.account);
              }
              // Update orders
              if (message.data.orders && Array.isArray(message.data.orders)) {
                message.data.orders.forEach((order: any) => updateOrder(order));
              }
              // Update positions
              if (message.data.positions) {
                Object.values(message.data.positions).forEach((pos: any) => updatePosition(pos));
              }
              logger.info('WebSocket', 'Processed full state snapshot');
            }
            break;

          case 'update':
          case 'incremental':
            // Handle incremental updates - uses 'changes' field, not 'data'!
            const updateData = message.changes || message.data;
            if (updateData) {
              logger.debug('WebSocket', 'Received incremental update', {
                hasChanges: !!message.changes,
                hasData: !!message.data,
                hasMarketData: !!(updateData.market_data || updateData.MarketData),
                rawUpdate: updateData,
              });

              // Process market data updates (check both snake_case and PascalCase)
              const marketDataUpdates = updateData.market_data || updateData.MarketData;
              if (marketDataUpdates) {
                Object.entries(marketDataUpdates).forEach(([symbol, data]: [string, any]) => {
                  // Check if this is market data or orderbook
                  if (data.last_price !== undefined || data.LastPrice !== undefined) {
                    logger.debug('WebSocket', `Incremental update for ${symbol}`, data);
                    updateMarketData(symbol, data);
                  } else if (data.bids && data.asks) {
                    updateOrderBook(symbol, { ...data, symbol, timestamp: Date.now() });
                  }
                });
              }

              const accountUpdate = updateData.account || updateData.Account;
              if (accountUpdate) {
                updateAccount(accountUpdate);
              }
            }
            break;

          case 'orderbook':
            if (message.data) {
              updateOrderBook(message.data.symbol, message.data);
            }
            break;

          case 'position':
            if (message.data) {
              updatePosition(message.data);
            }
            break;

          case 'order':
            if (message.data) {
              if (message.data.status === 'FILLED' || message.data.status === 'CANCELED') {
                removeOrder(message.data.id);
              } else {
                updateOrder(message.data);
              }
            }
            break;

          case 'trade':
            if (message.data) {
              addTrade(message.data);
            }
            break;

          case 'account':
            if (message.data) {
              updateAccount(message.data);
            }
            break;

          case 'system_health':
            if (message.data) {
              updateSystemHealth(message.data);
            }
            break;

          default:
            logger.warn('WebSocket', 'Unknown message type', { type: message.type });
        }
      } catch (error) {
        logger.error('WebSocket', 'Error parsing message', error);
      }
    },
    [
      setLatency,
      updateOrderBook,
      updateMarketData,
      updatePosition,
      updateOrder,
      removeOrder,
      addTrade,
      updateAccount,
      updateSystemHealth,
    ]
  );

  const sendHeartbeat = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          type: 'ping',
          timestamp: Date.now(),
        })
      );
      logger.trace('WebSocket', 'Heartbeat sent');
    }
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    logger.info('WebSocket', 'Connecting to', { url });
    setStatus('connecting');

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        logger.info('WebSocket', 'Connected successfully');
        setStatus('connected');
        setWs(ws);
        reconnectAttemptsRef.current = 0;

        // Start heartbeat
        heartbeatRef.current = window.setInterval(sendHeartbeat, heartbeatInterval);

        // Subscribe to all channels including market_data and strategies
        ws.send(
          JSON.stringify({
            type: 'subscribe',
            channels: ['market_data', 'account', 'positions', 'orders', 'trades', 'strategies', 'system_health'],
          })
        );
        logger.info('WebSocket', 'Sent subscription request');
      };

      ws.onclose = (event) => {
        logger.warn('WebSocket', 'Disconnected', { code: event.code, reason: event.reason });
        setStatus('disconnected');
        setWs(null);

        // Clear heartbeat
        if (heartbeatRef.current) {
          clearInterval(heartbeatRef.current);
        }

        // Clear data on disconnect
        clearData();

        // Attempt reconnection
        if (reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          const delay = reconnectInterval * Math.pow(2, reconnectAttemptsRef.current - 1);
          logger.info('WebSocket', `Reconnecting in ${delay}ms`, {
            attempt: reconnectAttemptsRef.current,
            maxAttempts: maxReconnectAttempts,
          });

          reconnectTimeoutRef.current = window.setTimeout(() => {
            connect();
          }, delay);
        } else {
          logger.error('WebSocket', 'Max reconnection attempts reached');
          setStatus('error');
        }
      };

      ws.onerror = (error) => {
        logger.error('WebSocket', 'Connection error', error);
        setStatus('error');
      };

      ws.onmessage = handleMessage;

      wsRef.current = ws;
    } catch (error) {
      logger.error('WebSocket', 'Failed to create connection', error);
      setStatus('error');
    }
  }, [
    url,
    reconnectInterval,
    maxReconnectAttempts,
    heartbeatInterval,
    handleMessage,
    sendHeartbeat,
    setStatus,
    setWs,
    clearData,
  ]);

  const disconnect = useCallback(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current);
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close(1000, 'Client disconnecting');
      wsRef.current = null;
    }
    setWs(null);
    setStatus('disconnected');
    logger.info('WebSocket', 'Disconnected by client');
  }, [setWs, setStatus]);

  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    status: useTradingStore((state) => state.status),
    latency: useTradingStore((state) => state.latency),
    reconnect: connect,
    disconnect,
  };
}
