interface Config {
  websocketUrl: string;
  apiUrl: string;
  authServiceUrl: string;
  environment: 'development' | 'production';
  enableDevTools: boolean;
}

export const config: Config = {
  websocketUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws',
  apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  authServiceUrl: import.meta.env.VITE_AUTH_URL || 'http://localhost:3001',
  environment: (import.meta.env.MODE as 'development' | 'production') || 'development',
  enableDevTools: import.meta.env.DEV || false,
};
