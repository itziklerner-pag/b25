/**
 * Express Application Configuration
 */
import express, { Application } from 'express';
import helmet from 'helmet';
import cors from 'cors';
import compression from 'compression';
import rateLimit from 'express-rate-limit';
import config from './config';
import routes from './routes';
import { errorHandler, notFoundHandler } from './middleware/error.middleware';
import { requestId, requestLogger } from './middleware/request-logger.middleware';
import { metricsMiddleware } from './middleware/metrics.middleware';

export function createApp(): Application {
  const app = express();

  // Trust proxy for correct IP addresses
  app.set('trust proxy', 1);

  // Security middleware
  app.use(helmet());

  // CORS configuration
  app.use(
    cors({
      origin: config.cors.origins,
      credentials: true,
    })
  );

  // Compression
  app.use(compression());

  // Body parsing
  app.use(express.json({ limit: '10mb' }));
  app.use(express.urlencoded({ extended: true, limit: '10mb' }));

  // Request ID and logging
  app.use(requestId);
  app.use(requestLogger);

  // Metrics collection
  app.use(metricsMiddleware);

  // Rate limiting
  const limiter = rateLimit({
    windowMs: config.api.rateLimit.windowMs,
    max: config.api.rateLimit.maxRequests,
    standardHeaders: true,
    legacyHeaders: false,
    message: {
      success: false,
      error: {
        code: 'RATE_LIMIT_EXCEEDED',
        message: 'Too many requests, please try again later',
      },
    },
  });

  app.use(`/api/${config.api.version}`, limiter);

  // API routes
  app.use(`/api/${config.api.version}`, routes);

  // Root endpoint
  app.get('/', (req, res) => {
    res.json({
      success: true,
      data: {
        service: 'user-profile',
        version: process.env.npm_package_version || '1.0.0',
        apiVersion: config.api.version,
        documentation: `/api/${config.api.version}/docs`,
      },
    });
  });

  // 404 handler
  app.use(notFoundHandler);

  // Global error handler (must be last)
  app.use(errorHandler);

  return app;
}

export default createApp;
