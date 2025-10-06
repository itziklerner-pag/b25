import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type {
  OrderBook,
  MarketData,
  Position,
  Order,
  Trade,
  AccountState,
  SystemHealth,
  OrderRequest,
  ConnectionStatus,
} from '@/types';
import { logger } from '@/utils/logger';

interface TradingStore {
  // WebSocket connection state
  ws: WebSocket | null;
  status: ConnectionStatus;
  latency: number;
  lastUpdate: number;

  // Trading data
  orderbooks: Map<string, OrderBook>;
  marketData: Map<string, MarketData>;  // NEW: Separate market data storage
  positions: Map<string, Position>;
  orders: Map<string, Order>;
  trades: Trade[];
  account: AccountState | null;
  systemHealth: SystemHealth[];

  // UI state
  selectedSymbol: string;

  // Debug state
  debugInfo: {
    lastWsMessage: number;
    lastStoreUpdate: number;
    updateCount: number;
    lastSymbol: string;
    lastPrice: number;
  };

  // Actions
  setWs: (ws: WebSocket | null) => void;
  setStatus: (status: ConnectionStatus) => void;
  setLatency: (latency: number) => void;
  updateOrderBook: (symbol: string, data: OrderBook) => void;
  updateMarketData: (symbol: string, data: MarketData) => void;  // NEW
  updatePosition: (position: Position) => void;
  updateOrder: (order: Order) => void;
  removeOrder: (orderId: string) => void;
  addTrade: (trade: Trade) => void;
  updateAccount: (account: AccountState) => void;
  updateSystemHealth: (health: SystemHealth[]) => void;
  setSelectedSymbol: (symbol: string) => void;
  clearData: () => void;

  // WebSocket message handlers
  sendOrder: (order: OrderRequest) => void;
  cancelOrder: (orderId: string) => void;
  closePosition: (symbol: string) => void;
  subscribe: (channel: string) => void;
  unsubscribe: (channel: string) => void;
}

export const useTradingStore = create<TradingStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      ws: null,
      status: 'disconnected',
      latency: 0,
      lastUpdate: Date.now(),
      orderbooks: new Map(),
      marketData: new Map(),  // NEW
      positions: new Map(),
      orders: new Map(),
      trades: [],
      account: null,
      systemHealth: [],
      selectedSymbol: 'BTCUSDT',
      debugInfo: {
        lastWsMessage: 0,
        lastStoreUpdate: 0,
        updateCount: 0,
        lastSymbol: '',
        lastPrice: 0,
      },

      // Connection actions
      setWs: (ws) => set({ ws }),
      setStatus: (status) => set({ status }),
      setLatency: (latency) => set({ latency }),

      // Data update actions
      updateOrderBook: (symbol, data) => {
        set((state) => {
          const newOrderbooks = new Map(state.orderbooks);
          newOrderbooks.set(symbol, data);
          return { orderbooks: newOrderbooks, lastUpdate: Date.now() };
        });
      },

      updateMarketData: (symbol, data) => {
        const timestamp = Date.now();

        // CRITICAL FIX: Use set with a function to ensure we're working with fresh state
        set((state) => {
          // Create completely new Map instance (not just copy)
          const newMarketData = new Map(state.marketData);
          newMarketData.set(symbol, { ...data, timestamp });

          // Update debug info
          const newDebugInfo = {
            lastWsMessage: timestamp,
            lastStoreUpdate: timestamp,
            updateCount: state.debugInfo.updateCount + 1,
            lastSymbol: symbol,
            lastPrice: data.last_price || 0,
          };

          logger.debug('Store', `Updated market data for ${symbol}`, {
            last_price: data.last_price,
            price_change_24h: data.price_change_24h,
            timestamp: new Date(timestamp).toISOString(),
            mapSize: newMarketData.size,
            updateCount: newDebugInfo.updateCount,
          });

          // Return COMPLETELY NEW STATE OBJECT with all new references
          return {
            marketData: newMarketData,
            lastUpdate: timestamp,
            debugInfo: newDebugInfo,
          };
        });
      },

      updatePosition: (position) => {
        set((state) => {
          const newPositions = new Map(state.positions);
          newPositions.set(position.symbol, position);
          logger.debug('Store', 'Updated position', { symbol: position.symbol, size: position.size });
          return { positions: newPositions, lastUpdate: Date.now() };
        });
      },

      updateOrder: (order) => {
        set((state) => {
          const newOrders = new Map(state.orders);
          newOrders.set(order.id, order);
          logger.debug('Store', 'Updated order', { id: order.id, status: order.status });
          return { orders: newOrders, lastUpdate: Date.now() };
        });
      },

      removeOrder: (orderId) => {
        set((state) => {
          const newOrders = new Map(state.orders);
          newOrders.delete(orderId);
          logger.debug('Store', 'Removed order', { id: orderId });
          return { orders: newOrders, lastUpdate: Date.now() };
        });
      },

      addTrade: (trade) => {
        set((state) => {
          logger.debug('Store', 'Added trade', { symbol: trade.symbol, side: trade.side });
          return {
            trades: [trade, ...state.trades].slice(0, 100), // Keep last 100 trades
            lastUpdate: Date.now(),
          };
        });
      },

      updateAccount: (account) => {
        logger.debug('Store', 'Updated account', { balance: account.balance, equity: account.equity });
        set({ account, lastUpdate: Date.now() });
      },

      updateSystemHealth: (health) => {
        logger.debug('Store', 'Updated system health', { services: health.length });
        set({ systemHealth: health });
      },

      setSelectedSymbol: (symbol) => {
        logger.info('Store', 'Selected symbol changed', { symbol });
        set({ selectedSymbol: symbol });
      },

      clearData: () => {
        logger.info('Store', 'Clearing all data');
        set({
          orderbooks: new Map(),
          marketData: new Map(),  // NEW
          positions: new Map(),
          orders: new Map(),
          trades: [],
          account: null,
        });
      },

      // WebSocket actions
      sendOrder: (order) => {
        const { ws, status } = get();
        if (ws && status === 'connected') {
          ws.send(
            JSON.stringify({
              type: 'order',
              action: 'create',
              data: order,
            })
          );
          logger.info('Store', 'Sent order', { symbol: order.symbol, side: order.side, type: order.type });
        } else {
          logger.warn('Store', 'Cannot send order - not connected');
        }
      },

      cancelOrder: (orderId) => {
        const { ws, status } = get();
        if (ws && status === 'connected') {
          ws.send(
            JSON.stringify({
              type: 'order',
              action: 'cancel',
              data: { orderId },
            })
          );
          logger.info('Store', 'Sent cancel order', { orderId });
        } else {
          logger.warn('Store', 'Cannot cancel order - not connected');
        }
      },

      closePosition: (symbol) => {
        const { ws, status } = get();
        if (ws && status === 'connected') {
          ws.send(
            JSON.stringify({
              type: 'position',
              action: 'close',
              data: { symbol },
            })
          );
          logger.info('Store', 'Sent close position', { symbol });
        } else {
          logger.warn('Store', 'Cannot close position - not connected');
        }
      },

      subscribe: (channel) => {
        const { ws, status } = get();
        if (ws && status === 'connected') {
          ws.send(
            JSON.stringify({
              type: 'subscribe',
              channel,
            })
          );
          logger.info('Store', 'Subscribed to channel', { channel });
        }
      },

      unsubscribe: (channel) => {
        const { ws, status } = get();
        if (ws && status === 'connected') {
          ws.send(
            JSON.stringify({
              type: 'unsubscribe',
              channel,
            })
          );
          logger.info('Store', 'Unsubscribed from channel', { channel });
        }
      },
    }),
    { name: 'TradingStore' }
  )
);
