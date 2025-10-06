import dotenv from 'dotenv';

// Load environment variables
dotenv.config();

function getEnv(key, defaultValue) {
  const value = process.env[key];
  if (!value && !defaultValue) {
    throw new Error(`Environment variable ${key} is required but not set`);
  }
  return value || defaultValue;
}

function getEnvNumber(key, defaultValue) {
  const value = process.env[key];
  return value ? parseInt(value, 10) : defaultValue;
}

export const databaseConfig = {
  host: getEnv('DB_HOST', 'localhost'),
  port: getEnvNumber('DB_PORT', 5432),
  database: getEnv('DB_NAME', 'b25_auth'),
  user: getEnv('DB_USER', 'postgres'),
  password: getEnv('DB_PASSWORD', 'postgres'),
  max: getEnvNumber('DB_POOL_MAX', 20),
  idleTimeoutMillis: getEnvNumber('DB_IDLE_TIMEOUT', 30000),
  connectionTimeoutMillis: getEnvNumber('DB_CONNECTION_TIMEOUT', 2000),
};

// Validate JWT secrets are not placeholders
function validateJwtSecret(secret, name) {
  if (!secret) {
    throw new Error(`${name} is required but not set`);
  }
  if (secret.includes('change-this') || secret.includes('your-super-secret')) {
    throw new Error(
      `${name} contains placeholder value. Generate a strong secret using: openssl rand -base64 64`
    );
  }
  if (secret.length < 32) {
    throw new Error(`${name} must be at least 32 characters long for security`);
  }
}

const accessSecret = getEnv('JWT_ACCESS_SECRET');
const refreshSecret = getEnv('JWT_REFRESH_SECRET');

// Validate secrets in production
if (process.env.NODE_ENV === 'production') {
  validateJwtSecret(accessSecret, 'JWT_ACCESS_SECRET');
  validateJwtSecret(refreshSecret, 'JWT_REFRESH_SECRET');
}

export const jwtConfig = {
  accessTokenSecret: accessSecret,
  refreshTokenSecret: refreshSecret,
  accessTokenExpiry: getEnv('JWT_ACCESS_EXPIRY', '15m'),
  refreshTokenExpiry: getEnv('JWT_REFRESH_EXPIRY', '7d'),
};

export const serverConfig = {
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
