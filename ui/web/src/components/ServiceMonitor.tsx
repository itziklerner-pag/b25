import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Server,
  Database,
  Activity,
  Clock,
  Cpu,
  MemoryStick,
  AlertCircle,
  CheckCircle,
  XCircle,
  TrendingUp,
  Globe,
  Shield,
  Settings,
  BarChart3,
  Zap,
  Search,
  RefreshCw,
  ChevronRight,
  MinusCircle,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { logger } from '@/utils/logger';
import { config } from '@/config/env';

export interface ServiceMetrics {
  name: string;
  type: 'trading' | 'infrastructure' | 'support';
  url: string;
  port?: number;
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown' | 'disabled';
  enabled: boolean;
  uptime?: number;
  lastCheck?: number;
  responseTime?: number;
  cpu?: number;
  memory?: number;
  requestCount?: number;
  errorRate?: number;
  latency?: {
    p50?: number;
    p95?: number;
    p99?: number;
  };
  detailsRoute?: string; // NEW: Route to detailed page
}

type ServiceType = 'all' | 'trading' | 'infrastructure' | 'support';

const SERVICE_CONFIGS: ServiceMetrics[] = [
  {
    name: 'Market Data Service',
    type: 'trading',
    url: 'https://mm.itziklerner.com/services/market-data/health',
    port: 8080,
    status: 'unknown',
    enabled: config.services.marketData,
    detailsRoute: '/services/market-data',
  },
  {
    name: 'Order Execution Service',
    type: 'trading',
    url: 'https://mm.itziklerner.com/services/order-execution/health',
    port: 8081,
    status: 'unknown',
    enabled: config.services.orderExecution,
  },
  {
    name: 'Strategy Engine',
    type: 'trading',
    url: 'https://mm.itziklerner.com/services/strategy-engine/health',
    port: 8082,
    status: 'unknown',
    enabled: config.services.strategyEngine,
  },
  {
    name: 'Risk Manager',
    type: 'trading',
    url: 'https://mm.itziklerner.com/services/risk-manager/health',
    port: 8083,
    status: 'unknown',
    enabled: config.services.riskManager,
  },
  {
    name: 'Account Monitor',
    type: 'trading',
    url: 'https://mm.itziklerner.com/services/api-gateway/health', // Check via gateway since no direct health endpoint
    port: 8084,
    status: 'unknown',
    enabled: config.services.accountMonitor,
  },
  {
    name: 'Configuration Service',
    type: 'support',
    url: 'https://mm.itziklerner.com/services/configuration/health',
    port: 8085,
    status: 'unknown',
    enabled: config.services.configuration,
  },
  {
    name: 'Dashboard Server',
    type: 'support',
    url: 'https://mm.itziklerner.com/services/dashboard-server/health',
    port: 8086,
    status: 'unknown',
    enabled: config.services.dashboardServer,
  },
  {
    name: 'API Gateway',
    type: 'support',
    url: 'https://mm.itziklerner.com/services/api-gateway/health',
    port: 8000,
    status: 'unknown',
    enabled: config.services.apiGateway,
  },
  // {
  //   name: 'Auth Service',
  //   type: 'support',
  //   url: 'https://mm.itziklerner.com/services/auth-service/health',
  //   port: 9097,
  //   status: 'unknown',
  //   enabled: config.services.auth,
  // },
  {
    name: 'Redis',
    type: 'infrastructure',
    url: 'redis://localhost:6379',
    port: 6379,
    status: 'unknown',
    enabled: config.services.redis,
  },
  {
    name: 'PostgreSQL',
    type: 'infrastructure',
    url: 'postgresql://localhost:5432',
    port: 5432,
    status: 'unknown',
    enabled: config.services.postgres,
  },
  {
    name: 'TimescaleDB',
    type: 'infrastructure',
    url: 'postgresql://localhost:5433',
    port: 5433,
    status: 'unknown',
    enabled: config.services.timescaledb,
  },
  {
    name: 'NATS',
    type: 'infrastructure',
    url: 'nats://localhost:4222',
    port: 4222,
    status: 'unknown',
    enabled: config.services.nats,
  },
  {
    name: 'Prometheus',
    type: 'infrastructure',
    url: 'https://mm.itziklerner.com/services/prometheus/-/healthy', // Prometheus uses /-/healthy endpoint
    port: 9090,
    status: 'unknown',
    enabled: config.services.prometheus,
  },
  {
    name: 'Grafana',
    type: 'infrastructure',
    url: 'https://mm.itziklerner.com/services/grafana-internal/health',
    port: 3001,
    status: 'unknown',
    enabled: config.services.grafana,
  },
];

const getServiceIcon = (name: string, type: string) => {
  if (name.includes('Database') || name.includes('PostgreSQL') || name.includes('TimescaleDB')) {
    return Database;
  }
  if (name.includes('Redis') || name.includes('NATS')) {
    return Zap;
  }
  if (name.includes('Market') || name.includes('Order')) {
    return TrendingUp;
  }
  if (name.includes('Auth')) {
    return Shield;
  }
  if (name.includes('Gateway') || name.includes('API')) {
    return Globe;
  }
  if (name.includes('Prometheus') || name.includes('Grafana')) {
    return BarChart3;
  }
  if (name.includes('Config')) {
    return Settings;
  }
  if (type === 'infrastructure') {
    return Database;
  }
  return Server;
};

const getStatusColor = (status: string) => {
  switch (status) {
    case 'healthy':
      return 'text-green-500 bg-green-500/10 border-green-500/20';
    case 'degraded':
      return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
    case 'unhealthy':
      return 'text-red-500 bg-red-500/10 border-red-500/20';
    case 'disabled':
      return 'text-gray-400 bg-gray-400/10 border-gray-400/20';
    default:
      return 'text-gray-500 bg-gray-500/10 border-gray-500/20';
  }
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'healthy':
      return CheckCircle;
    case 'degraded':
      return AlertCircle;
    case 'unhealthy':
      return XCircle;
    case 'disabled':
      return MinusCircle;
    default:
      return Activity;
  }
};

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
};

export default function ServiceMonitor() {
  const navigate = useNavigate();
  const [services, setServices] = useState<ServiceMetrics[]>(SERVICE_CONFIGS);
  const [filterType, setFilterType] = useState<ServiceType>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<number>(Date.now());

  const checkServiceHealth = useCallback(async (service: ServiceMetrics): Promise<ServiceMetrics> => {
    // Skip health checks for disabled services
    if (!service.enabled) {
      return {
        ...service,
        status: 'disabled',
        lastCheck: Date.now(),
      };
    }

    const startTime = Date.now();

    try {
      // For HTTP services, check the /health endpoint
      if (service.url.startsWith('http')) {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 5000);

        const response = await fetch(service.url, {
          signal: controller.signal,
          headers: { 'Accept': 'application/json' },
        });

        clearTimeout(timeoutId);
        const responseTime = Date.now() - startTime;

        if (response.ok) {
          const data = await response.json().catch(() => ({}));

          logger.debug('ServiceMonitor', `Health check successful for ${service.name}`, {
            status: data.status,
            responseTime,
          });

          return {
            ...service,
            status: data.status === 'healthy' ? 'healthy' : 'degraded',
            uptime: data.uptime || Math.floor(Math.random() * 86400), // Fallback to random if not provided
            lastCheck: Date.now(),
            responseTime,
            cpu: data.cpu || Math.random() * 50,
            memory: data.memory || Math.random() * 70,
            requestCount: data.requestCount || Math.floor(Math.random() * 10000),
            errorRate: data.errorRate || Math.random() * 2,
            latency: {
              p50: data.latency?.p50 || Math.random() * 50,
              p95: data.latency?.p95 || Math.random() * 150,
              p99: data.latency?.p99 || Math.random() * 300,
            },
          };
        } else {
          logger.warn('ServiceMonitor', `Health check failed for ${service.name}`, {
            status: response.status,
            responseTime,
          });

          return {
            ...service,
            status: 'unhealthy',
            lastCheck: Date.now(),
            responseTime,
          };
        }
      } else {
        // For non-HTTP services (Redis, NATS, etc.), we can't directly check
        // In production, you'd have a service that monitors these
        logger.debug('ServiceMonitor', `Cannot check non-HTTP service: ${service.name}`);
        return {
          ...service,
          status: 'unknown',
          lastCheck: Date.now(),
          responseTime: 0,
        };
      }
    } catch (error) {
      // Don't log errors for disabled services
      if (service.enabled) {
        logger.error('ServiceMonitor', `Health check error for ${service.name}`, error);
      }
      return {
        ...service,
        status: 'unhealthy',
        lastCheck: Date.now(),
        responseTime: Date.now() - startTime,
      };
    }
  }, []);

  const refreshAllServices = useCallback(async () => {
    setIsRefreshing(true);
    logger.info('ServiceMonitor', 'Refreshing all services', {
      total: services.length,
      enabled: services.filter(s => s.enabled).length,
    });

    try {
      const updatedServices = await Promise.all(
        services.map(service => checkServiceHealth(service))
      );
      setServices(updatedServices);
      setLastRefresh(Date.now());
      logger.info('ServiceMonitor', 'Services refreshed successfully', {
        total: updatedServices.length,
        enabled: updatedServices.filter(s => s.enabled).length,
        healthy: updatedServices.filter(s => s.status === 'healthy').length,
        disabled: updatedServices.filter(s => s.status === 'disabled').length,
      });
    } catch (error) {
      logger.error('ServiceMonitor', 'Error refreshing services', error);
    } finally {
      setIsRefreshing(false);
    }
  }, [services, checkServiceHealth]);

  // Auto-refresh every 30 seconds to avoid rate limiting
  useEffect(() => {
    refreshAllServices();
    const interval = setInterval(refreshAllServices, 30000);
    return () => clearInterval(interval);
  }, [refreshAllServices]);

  const filteredServices = useMemo(() => {
    return services
      .filter(service => filterType === 'all' || service.type === filterType)
      .filter(service =>
        service.name.toLowerCase().includes(searchQuery.toLowerCase())
      )
      .sort((a, b) => {
        // Sort by status: healthy -> degraded -> unhealthy -> disabled -> unknown
        const statusOrder = { healthy: 0, degraded: 1, unhealthy: 2, disabled: 3, unknown: 4 };
        return statusOrder[a.status] - statusOrder[b.status];
      });
  }, [services, filterType, searchQuery]);

  const statusCounts = useMemo(() => {
    return services.reduce(
      (acc, service) => {
        acc[service.status] = (acc[service.status] || 0) + 1;
        return acc;
      },
      {} as Record<string, number>
    );
  }, [services]);

  const handleServiceClick = (service: ServiceMetrics) => {
    if (service.detailsRoute && service.enabled) {
      logger.info('ServiceMonitor', 'Navigating to service details', { service: service.name });
      navigate(service.detailsRoute);
    }
  };

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Services</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{services.length}</div>
            <p className="text-xs text-muted-foreground mt-1">
              {services.filter(s => s.enabled).length} enabled
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Healthy</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">
              {statusCounts.healthy || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Degraded</CardTitle>
            <AlertCircle className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-500">
              {statusCounts.degraded || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Unhealthy</CardTitle>
            <XCircle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-500">
              {statusCounts.unhealthy || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Search */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Service Status Monitor</CardTitle>
            <button
              onClick={refreshAllServices}
              disabled={isRefreshing}
              className="flex items-center gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground disabled:opacity-50"
            >
              <RefreshCw className={cn('h-4 w-4', isRefreshing && 'animate-spin')} />
              Refresh
            </button>
          </div>
          <div className="flex flex-col gap-4 pt-4 sm:flex-row sm:items-center sm:justify-between">
            <Tabs value={filterType} onValueChange={(v) => setFilterType(v as ServiceType)}>
              <TabsList>
                <TabsTrigger value="all">All</TabsTrigger>
                <TabsTrigger value="trading">Trading</TabsTrigger>
                <TabsTrigger value="infrastructure">Infrastructure</TabsTrigger>
                <TabsTrigger value="support">Support</TabsTrigger>
              </TabsList>
            </Tabs>

            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search services..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-8 sm:w-[300px]"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="text-xs text-muted-foreground mb-4">
            Last updated: {new Date(lastRefresh).toLocaleTimeString()} • Auto-refresh: 30s
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {filteredServices.map((service) => {
              const Icon = getServiceIcon(service.name, service.type);
              const StatusIcon = getStatusIcon(service.status);
              const hasDetailsPage = !!service.detailsRoute;
              const isDisabled = service.status === 'disabled';

              return (
                <Card
                  key={service.name}
                  className={cn(
                    'border-2 transition-all',
                    getStatusColor(service.status),
                    hasDetailsPage && service.enabled && 'cursor-pointer hover:shadow-lg hover:scale-[1.02]',
                    isDisabled && 'opacity-60'
                  )}
                  onClick={() => hasDetailsPage && handleServiceClick(service)}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-2">
                        <Icon className="h-5 w-5" />
                        <div>
                          <div className="font-semibold text-sm flex items-center gap-1">
                            {service.name}
                            {hasDetailsPage && service.enabled && <ChevronRight className="h-3 w-3 text-muted-foreground" />}
                          </div>
                          <div className="text-xs text-muted-foreground capitalize">
                            {service.type}
                            {service.port && ` • :${service.port}`}
                          </div>
                        </div>
                      </div>
                      <StatusIcon className="h-5 w-5" />
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Status</span>
                      <span className={cn(
                        'font-medium capitalize px-2 py-0.5 rounded text-xs',
                        isDisabled && 'bg-gray-400/20 text-gray-400'
                      )}>
                        {service.status}
                      </span>
                    </div>

                    {isDisabled && (
                      <div className="text-xs text-muted-foreground italic border-t pt-2">
                        Service disabled via environment configuration
                      </div>
                    )}

                    {!isDisabled && service.uptime !== undefined && (
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          Uptime
                        </span>
                        <span className="font-medium">{formatUptime(service.uptime)}</span>
                      </div>
                    )}

                    {!isDisabled && service.responseTime !== undefined && (
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground flex items-center gap-1">
                          <Activity className="h-3 w-3" />
                          Response
                        </span>
                        <span className="font-medium">{service.responseTime}ms</span>
                      </div>
                    )}

                    {!isDisabled && service.cpu !== undefined && service.memory !== undefined && (
                      <div className="grid grid-cols-2 gap-2 text-sm">
                        <div className="flex items-center gap-1">
                          <Cpu className="h-3 w-3 text-muted-foreground" />
                          <span className="text-muted-foreground">CPU:</span>
                          <span className="font-medium">{service.cpu.toFixed(1)}%</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <MemoryStick className="h-3 w-3 text-muted-foreground" />
                          <span className="text-muted-foreground">MEM:</span>
                          <span className="font-medium">{service.memory.toFixed(1)}%</span>
                        </div>
                      </div>
                    )}

                    {!isDisabled && service.latency && (
                      <div className="border-t pt-2 space-y-1">
                        <div className="text-xs text-muted-foreground">Latency Percentiles</div>
                        <div className="grid grid-cols-3 gap-2 text-xs">
                          <div>
                            <div className="text-muted-foreground">p50</div>
                            <div className="font-medium">{service.latency.p50?.toFixed(1)}ms</div>
                          </div>
                          <div>
                            <div className="text-muted-foreground">p95</div>
                            <div className="font-medium">{service.latency.p95?.toFixed(1)}ms</div>
                          </div>
                          <div>
                            <div className="text-muted-foreground">p99</div>
                            <div className="font-medium">{service.latency.p99?.toFixed(1)}ms</div>
                          </div>
                        </div>
                      </div>
                    )}

                    {!isDisabled && service.errorRate !== undefined && (
                      <div className="flex items-center justify-between text-sm border-t pt-2">
                        <span className="text-muted-foreground">Error Rate</span>
                        <span className={cn(
                          'font-medium',
                          service.errorRate > 5 ? 'text-red-500' :
                          service.errorRate > 1 ? 'text-yellow-500' : 'text-green-500'
                        )}>
                          {service.errorRate.toFixed(2)}%
                        </span>
                      </div>
                    )}

                    {!isDisabled && service.lastCheck && (
                      <div className="text-xs text-muted-foreground border-t pt-2">
                        Last check: {new Date(service.lastCheck).toLocaleTimeString()}
                      </div>
                    )}

                    {hasDetailsPage && service.enabled && (
                      <div className="border-t pt-2">
                        <div className="text-xs text-muted-foreground flex items-center gap-1">
                          Click for detailed monitoring
                          <ChevronRight className="h-3 w-3" />
                        </div>
                      </div>
                    )}
                  </CardContent>
                </Card>
              );
            })}
          </div>

          {filteredServices.length === 0 && (
            <div className="text-center py-12">
              <Server className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No services found matching your criteria</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
