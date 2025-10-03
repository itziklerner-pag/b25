import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Toaster } from 'sonner';
import { ThemeProvider } from '@/components/ThemeProvider';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import Layout from '@/components/Layout';
import DashboardPage from '@/pages/DashboardPage';
import PositionsPage from '@/pages/PositionsPage';
import OrdersPage from '@/pages/OrdersPage';
import OrderBookPage from '@/pages/OrderBookPage';
import AnalyticsPage from '@/pages/AnalyticsPage';
import TradingPage from '@/pages/TradingPage';
import SystemPage from '@/pages/SystemPage';
import LoginPage from '@/pages/LoginPage';
import { useWebSocket } from '@/hooks/useWebSocket';
import { config } from '@/config/env';

function App() {
  const { status } = useWebSocket({
    url: config.websocketUrl,
    reconnectInterval: 3000,
    maxReconnectAttempts: 10,
  });

  useEffect(() => {
    console.log('B25 Web Dashboard initialized');
  }, []);

  return (
    <ThemeProvider defaultTheme="dark" storageKey="b25-theme">
      <ErrorBoundary>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<Layout />}>
              <Route index element={<DashboardPage />} />
              <Route path="positions" element={<PositionsPage />} />
              <Route path="orders" element={<OrdersPage />} />
              <Route path="orderbook" element={<OrderBookPage />} />
              <Route path="analytics" element={<AnalyticsPage />} />
              <Route path="trade" element={<TradingPage />} />
              <Route path="system" element={<SystemPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
        <Toaster position="top-right" richColors />
      </ErrorBoundary>
    </ThemeProvider>
  );
}

export default App;
