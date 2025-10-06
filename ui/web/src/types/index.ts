export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface OrderBook {
  symbol: string;
  bids: PriceLevel[];
  asks: PriceLevel[];
  timestamp: number;
}

export interface PriceLevel {
  price: number;
  volume: number;
}

// Market data from WebSocket
export interface MarketData {
  last_price: number;
  bid?: number;
  ask?: number;
  high_24h?: number;
  low_24h?: number;
  volume_24h?: number;
  price_change_24h?: number;
  timestamp?: number;
}

export interface Position {
  symbol: string;
  side: 'LONG' | 'SHORT';
  size: number;
  entryPrice: number;
  markPrice: number;
  unrealizedPnl: number;
  leverage: number;
  liquidationPrice: number;
  margin: number;
  timestamp: number;
}

export interface Order {
  id: string;
  symbol: string;
  side: 'BUY' | 'SELL';
  type: 'LIMIT' | 'MARKET' | 'STOP_LIMIT' | 'STOP_MARKET';
  status: 'NEW' | 'PARTIALLY_FILLED' | 'FILLED' | 'CANCELED' | 'REJECTED';
  price?: number;
  quantity: number;
  filledQuantity: number;
  avgFillPrice?: number;
  timestamp: number;
}

export interface Trade {
  id: string;
  symbol: string;
  side: 'BUY' | 'SELL';
  price: number;
  quantity: number;
  timestamp: number;
}

export interface AccountState {
  balance: number;
  equity: number;
  availableBalance: number;
  unrealizedPnl: number;
  realizedPnl: number;
  marginRatio: number;
  totalPnl24h: number;
  timestamp: number;
}

export interface SystemHealth {
  service: string;
  status: 'healthy' | 'degraded' | 'down';
  latency: number;
  lastCheck: number;
  uptime: number;
  errorRate: number;
}

export interface OrderRequest {
  symbol: string;
  side: 'BUY' | 'SELL';
  type: 'LIMIT' | 'MARKET' | 'STOP_LIMIT' | 'STOP_MARKET';
  price?: number;
  quantity: number;
  stopPrice?: number;
  timeInForce?: 'GTC' | 'IOC' | 'FOK';
}

export interface WebSocketMessage {
  type: string;
  data?: any;
  changes?: any;
  timestamp?: number;
  channel?: string;
}
