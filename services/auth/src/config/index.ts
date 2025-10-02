import dotenv from 'dotenv';
import { DatabaseConfig, JWTConfig, ServerConfig } from '../types';

// Load environment variables
dotenv.config();

function getEnv(key: string, defaultValue?: string): string {
  const value = process.env[key];
  if (!value && !defaultValue) {
    throw new Error(`Environment variable ${key} is required but not set`);
  }
  return value || defaultValue!;
}

function getEnvNumber(key: string, defaultValue: number): number {
  const value = process.env[key];
  return value ? parseInt(value, 10) : defaultValue;
}

export const databaseConfig: DatabaseConfig = {
  host: getEnv('DB_HOST', 'localhost'),
  port: getEnvNumber('DB_PORT', 5432),
  database: getEnv('DB_NAME', 'b25_auth'),
  user: getEnv('DB_USER', 'postgres'),
  password: getEnv('DB_PASSWORD', 'postgres'),
  max: getEnvNumber('DB_POOL_MAX', 20),
  idleTimeoutMillis: getEnvNumber('DB_IDLE_TIMEOUT', 30000),
  connectionTimeoutMillis: getEnvNumber('DB_CONNECTION_TIMEOUT', 2000),
};

export const jwtConfig: JWTConfig = {
  accessTokenSecret: getEnv('JWT_ACCESS_SECRET'),
  refreshTokenSecret: getEnv('JWT_REFRESH_SECRET'),
  accessTokenExpiry: getEnv('JWT_ACCESS_EXPIRY', '15m'),
  refreshTokenExpiry: getEnv('JWT_REFRESH_EXPIRY', '7d'),
};

export const serverConfig: ServerConfig = {
  port: getEnvNumber('PORT', 9097),
  nodeEnv: getEnv('NODE_ENV', 'development'),
  corsOrigins: getEnv('CORS_ORIGINS', 'http://localhost:3000').split(','),
  rateLimitWindowMs: getEnvNumber('RATE_LIMIT_WINDOW_MS', 900000), // 15 minutes
  rateLimitMaxRequests: getEnvNumber('RATE_LIMIT_MAX_REQUESTS', 100),
};

export const config = {
  database: databaseConfig,
  jwt: jwtConfig,
  server: serverConfig,
};

export default config;
