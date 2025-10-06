import promClient from 'prom-client';

// Create a Registry
export const register = new promClient.Registry();

// Add default metrics (process metrics like CPU, memory)
promClient.collectDefaultMetrics({ register });

// HTTP Request Duration Histogram
export const httpRequestDuration = new promClient.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route', 'status_code'],
  buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5],
  registers: [register],
});

// HTTP Request Counter
export const httpRequestCounter = new promClient.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status_code'],
  registers: [register],
});

// Auth Operations Counter
export const authOperationsCounter = new promClient.Counter({
  name: 'auth_operations_total',
  help: 'Total number of authentication operations',
  labelNames: ['operation', 'status'],
  registers: [register],
});

// Active Users Gauge
export const activeUsersGauge = new promClient.Gauge({
  name: 'auth_active_users',
  help: 'Number of users with valid refresh tokens',
  registers: [register],
});

// Token Operations Counter
export const tokenOperationsCounter = new promClient.Counter({
  name: 'auth_token_operations_total',
  help: 'Total number of token operations',
  labelNames: ['operation', 'status'],
  registers: [register],
});

// Database Query Duration Histogram
export const dbQueryDuration = new promClient.Histogram({
  name: 'db_query_duration_seconds',
  help: 'Duration of database queries in seconds',
  labelNames: ['operation'],
  buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1],
  registers: [register],
});

// Error Counter
export const errorCounter = new promClient.Counter({
  name: 'auth_errors_total',
  help: 'Total number of errors',
  labelNames: ['error_type', 'endpoint'],
  registers: [register],
});

export default {
  register,
  httpRequestDuration,
  httpRequestCounter,
  authOperationsCounter,
  activeUsersGauge,
  tokenOperationsCounter,
  dbQueryDuration,
  errorCounter,
};
