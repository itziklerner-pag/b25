import { useMemo } from 'react';
import { Wifi, WifiOff, AlertCircle, Activity } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTradingStore } from '@/store/trading';
import type { ConnectionStatus as ConnectionStatusType } from '@/types';

export function ConnectionStatus() {
  const status = useTradingStore((state) => state.status);
  const latency = useTradingStore((state) => state.latency);

  const { icon: Icon, color, text, bgColor } = useMemo(() => {
    switch (status) {
      case 'connected':
        return {
          icon: Wifi,
          color: 'text-green-500',
          bgColor: 'bg-green-500/10',
          text: `Connected (${latency}ms)`,
        };
      case 'connecting':
        return {
          icon: Wifi,
          color: 'text-yellow-500 animate-pulse',
          bgColor: 'bg-yellow-500/10',
          text: 'Connecting...',
        };
      case 'disconnected':
        return {
          icon: WifiOff,
          color: 'text-gray-500',
          bgColor: 'bg-gray-500/10',
          text: 'Disconnected',
        };
      case 'error':
        return {
          icon: AlertCircle,
          color: 'text-red-500',
          bgColor: 'bg-red-500/10',
          text: 'Connection Error',
        };
    }
  }, [status, latency]);

  return (
    <div className={cn('flex items-center gap-2 rounded-full px-3 py-1.5', bgColor)}>
      <Icon className={cn('h-4 w-4', color)} />
      <span className="text-sm font-medium">{text}</span>
      {status === 'connected' && (
        <div className="flex items-center gap-1">
          <Activity className={cn('h-3 w-3', latency < 100 ? 'text-green-500' : 'text-yellow-500')} />
        </div>
      )}
    </div>
  );
}
