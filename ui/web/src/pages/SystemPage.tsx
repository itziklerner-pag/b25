import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useTradingStore } from '@/store/trading';
import { Activity, AlertCircle, CheckCircle } from 'lucide-react';
import { formatTimestamp } from '@/lib/utils';

export default function SystemPage() {
  const status = useTradingStore((state) => state.status);
  const latency = useTradingStore((state) => state.latency);
  const lastUpdate = useTradingStore((state) => state.lastUpdate);
  const systemHealth = useTradingStore((state) => state.systemHealth);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Connection Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              {status === 'connected' ? (
                <CheckCircle className="h-5 w-5 text-green-500" />
              ) : (
                <AlertCircle className="h-5 w-5 text-red-500" />
              )}
              <span className="text-lg font-medium capitalize">{status}</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>WebSocket Latency</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <Activity
                className={`h-5 w-5 ${latency < 100 ? 'text-green-500' : latency < 500 ? 'text-yellow-500' : 'text-red-500'}`}
              />
              <span className="text-lg font-medium">{latency}ms</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Last Update</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-medium">{formatTimestamp(lastUpdate)}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>System Health</CardTitle>
        </CardHeader>
        <CardContent>
          {systemHealth.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No health data available</p>
          ) : (
            <div className="space-y-3">
              {systemHealth.map((service) => (
                <div key={service.service} className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <div className="font-medium">{service.service}</div>
                    <div className="text-sm text-muted-foreground">
                      Uptime: {Math.floor(service.uptime / 60)}h {service.uptime % 60}m
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-2">
                      {service.status === 'healthy' ? (
                        <CheckCircle className="h-5 w-5 text-green-500" />
                      ) : service.status === 'degraded' ? (
                        <AlertCircle className="h-5 w-5 text-yellow-500" />
                      ) : (
                        <AlertCircle className="h-5 w-5 text-red-500" />
                      )}
                      <span className="capitalize">{service.status}</span>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      Latency: {service.latency}ms
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
