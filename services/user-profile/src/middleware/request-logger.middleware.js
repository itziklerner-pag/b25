/**
 * Request Logging Middleware
 */
import { v4 as uuidv4 } from 'uuid';
import logger from '../utils/logger.js';

/**
 * Add request ID to each request
 */
export const requestId = (req, res, next) => {
  const id = req.headers['x-request-id'] || uuidv4();
  req.headers['x-request-id'] = id;
  res.setHeader('X-Request-ID', id);
  next();
};

/**
 * Log incoming requests and responses
 */
export const requestLogger = (req, res, next) => {
  const start = Date.now();

  // Log request
  logger.info('Incoming request', {
    requestId: req.headers['x-request-id'],
    method: req.method,
    path: req.path,
    query: req.query,
    ip: req.ip,
    userAgent: req.headers['user-agent'],
  });

  // Log response
  res.on('finish', () => {
    const duration = Date.now() - start;
    logger.info('Request completed', {
      requestId: req.headers['x-request-id'],
      method: req.method,
      path: req.path,
      statusCode: res.statusCode,
      duration: `${duration}ms`,
    });
  });

  next();
};
