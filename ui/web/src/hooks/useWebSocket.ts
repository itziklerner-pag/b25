import { useEffect, useRef, useCallback } from 'react';
import { useTradingStore } from '@/store/trading';
import type { WebSocketMessage } from '@/types';

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

        // Handle pong for latency measurement
        if (message.type === 'pong' && message.timestamp) {
          const now = Date.now();
          setLatency(now - message.timestamp);
          return;
        }

        // Route message to appropriate store action
        switch (message.type) {
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
            console.warn('Unknown message type:', message.type);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    },
    [
      setLatency,
      updateOrderBook,
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
    }
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    console.log('Connecting to WebSocket:', url);
    setStatus('connecting');

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        console.log('WebSocket connected');
        setStatus('connected');
        setWs(ws);
        reconnectAttemptsRef.current = 0;

        // Start heartbeat
        heartbeatRef.current = window.setInterval(sendHeartbeat, heartbeatInterval);

        // Subscribe to initial channels
        ws.send(
          JSON.stringify({
            type: 'subscribe',
            channels: ['account', 'positions', 'orders', 'trades', 'system_health'],
          })
        );
      };

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
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
          console.log(
            `Reconnecting in ${delay}ms... (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`
          );

          reconnectTimeoutRef.current = window.setTimeout(() => {
            connect();
          }, delay);
        } else {
          console.error('Max reconnection attempts reached');
          setStatus('error');
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setStatus('error');
      };

      ws.onmessage = handleMessage;

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
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
