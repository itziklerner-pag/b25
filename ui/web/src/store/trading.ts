import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type {
  OrderBook,
  Position,
  Order,
  Trade,
  AccountState,
  SystemHealth,
  OrderRequest,
  ConnectionStatus,
} from '@/types';

interface TradingStore {
  // WebSocket connection state
  ws: WebSocket | null;
  status: ConnectionStatus;
  latency: number;
  lastUpdate: number;

  // Trading data
  orderbooks: Map<string, OrderBook>;
  positions: Map<string, Position>;
  orders: Map<string, Order>;
  trades: Trade[];
  account: AccountState | null;
  systemHealth: SystemHealth[];

  // UI state
  selectedSymbol: string;

  // Actions
  setWs: (ws: WebSocket | null) => void;
  setStatus: (status: ConnectionStatus) => void;
  setLatency: (latency: number) => void;
  updateOrderBook: (symbol: string, data: OrderBook) => void;
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
      positions: new Map(),
      orders: new Map(),
      trades: [],
      account: null,
      systemHealth: [],
      selectedSymbol: 'BTCUSDT',

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

      updatePosition: (position) => {
        set((state) => {
          const newPositions = new Map(state.positions);
          newPositions.set(position.symbol, position);
          return { positions: newPositions, lastUpdate: Date.now() };
        });
      },

      updateOrder: (order) => {
        set((state) => {
          const newOrders = new Map(state.orders);
          newOrders.set(order.id, order);
          return { orders: newOrders, lastUpdate: Date.now() };
        });
      },

      removeOrder: (orderId) => {
        set((state) => {
          const newOrders = new Map(state.orders);
          newOrders.delete(orderId);
          return { orders: newOrders, lastUpdate: Date.now() };
        });
      },

      addTrade: (trade) => {
        set((state) => ({
          trades: [trade, ...state.trades].slice(0, 100), // Keep last 100 trades
          lastUpdate: Date.now(),
        }));
      },

      updateAccount: (account) => {
        set({ account, lastUpdate: Date.now() });
      },

      updateSystemHealth: (health) => {
        set({ systemHealth: health });
      },

      setSelectedSymbol: (symbol) => {
        set({ selectedSymbol: symbol });
      },

      clearData: () => {
        set({
          orderbooks: new Map(),
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
        }
      },
    }),
    { name: 'TradingStore' }
  )
);
