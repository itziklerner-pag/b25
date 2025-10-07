interface Config {
  websocketUrl: string;
  apiUrl: string;
  authServiceUrl: string;
  environment: 'development' | 'production';
  enableDevTools: boolean;
  services: {
    [key: string]: boolean;
  };
}

export const config: Config = {
  websocketUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8086/ws?type=web',
  apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:8000/api',
  authServiceUrl: import.meta.env.VITE_AUTH_URL || 'http://localhost:3001',
  environment: (import.meta.env.MODE as 'development' | 'production') || 'development',
  enableDevTools: import.meta.env.DEV || false,
  services: {
    marketData: import.meta.env.VITE_SERVICE_MARKET_DATA_ENABLED === 'true',
    orderExecution: import.meta.env.VITE_SERVICE_ORDER_EXECUTION_ENABLED === 'true',
    strategyEngine: import.meta.env.VITE_SERVICE_STRATEGY_ENGINE_ENABLED === 'true',
    riskManager: import.meta.env.VITE_SERVICE_RISK_MANAGER_ENABLED === 'true',
    accountMonitor: import.meta.env.VITE_SERVICE_ACCOUNT_MONITOR_ENABLED === 'true',
    dashboardServer: import.meta.env.VITE_SERVICE_DASHBOARD_SERVER_ENABLED === 'true',
    apiGateway: import.meta.env.VITE_SERVICE_API_GATEWAY_ENABLED === 'true',
    auth: import.meta.env.VITE_SERVICE_AUTH_ENABLED === 'true',
    configuration: import.meta.env.VITE_SERVICE_CONFIGURATION_ENABLED === 'true',
    analytics: import.meta.env.VITE_SERVICE_ANALYTICS_ENABLED === 'true',
    prometheus: import.meta.env.VITE_SERVICE_PROMETHEUS_ENABLED === 'true',
    grafana: import.meta.env.VITE_SERVICE_GRAFANA_ENABLED === 'true',
    redis: import.meta.env.VITE_SERVICE_REDIS_ENABLED === 'true',
    postgres: import.meta.env.VITE_SERVICE_POSTGRES_ENABLED === 'true',
    timescaledb: import.meta.env.VITE_SERVICE_TIMESCALEDB_ENABLED === 'true',
    nats: import.meta.env.VITE_SERVICE_NATS_ENABLED === 'true',
  },
};
