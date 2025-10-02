/**
 * Configuration Management
 */
import dotenv from 'dotenv';
import path from 'path';

// Load environment variables
dotenv.config({ path: path.resolve(process.cwd(), '.env') });

const config = {
  env: process.env.NODE_ENV || 'development',

  server: {
    port: parseInt(process.env.PORT || '9100', 10),
    host: process.env.HOST || '0.0.0.0',
  },

  database: {
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT || '5432', 10),
    name: process.env.DB_NAME || 'b25_user_profiles',
    user: process.env.DB_USER || 'postgres',
    password: process.env.DB_PASSWORD || 'postgres',
    poolMin: parseInt(process.env.DB_POOL_MIN || '2', 10),
    poolMax: parseInt(process.env.DB_POOL_MAX || '10', 10),
  },

  auth: {
    jwtSecret: process.env.JWT_SECRET || 'dev-secret-change-in-production',
    jwtExpiry: process.env.JWT_EXPIRY || '24h',
    bcryptRounds: parseInt(process.env.BCRYPT_ROUNDS || '12', 10),
  },

  api: {
    version: process.env.API_VERSION || 'v1',
    rateLimit: {
      windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS || '900000', 10),
      maxRequests: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS || '100', 10),
    },
  },

  upload: {
    maxFileSize: parseInt(process.env.MAX_FILE_SIZE || '5242880', 10),
    allowedFileTypes: (process.env.ALLOWED_FILE_TYPES || 'image/jpeg,image/png,image/gif,image/webp').split(','),
  },

  logging: {
    level: process.env.LOG_LEVEL || 'info',
    format: process.env.LOG_FORMAT || 'json',
  },

  cors: {
    origins: (process.env.CORS_ORIGIN || 'http://localhost:3000').split(','),
  },

  metrics: {
    enabled: process.env.METRICS_ENABLED === 'true',
    port: parseInt(process.env.METRICS_PORT || '9101', 10),
  },
};

// Validate required configuration
if (config.env === 'production' && config.auth.jwtSecret === 'dev-secret-change-in-production') {
  throw new Error('JWT_SECRET must be set in production environment');
}

export default config;
