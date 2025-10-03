import { Router } from 'express';
import db from '../database/pool.js';
import { successResponse } from '../utils/response.js';

const router = Router();

/**
 * GET /health
 * Health check endpoint
 */
router.get('/', async (_req, res) => {
  const startTime = Date.now();

  try {
    // Check database connection
    const dbHealthy = await db.healthCheck();

    const health = {
      status: dbHealthy ? 'healthy' : 'unhealthy',
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
      responseTime: Date.now() - startTime,
      database: dbHealthy ? 'connected' : 'disconnected',
      service: 'auth-service',
      version: '1.0.0',
    };

    const statusCode = dbHealthy ? 200 : 503;
    successResponse(res, health, statusCode);
  } catch (error) {
    const health = {
      status: 'unhealthy',
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
      responseTime: Date.now() - startTime,
      database: 'error',
      service: 'auth-service',
      version: '1.0.0',
    };

    successResponse(res, health, 503);
  }
});

/**
 * GET /health/ready
 * Readiness probe
 */
router.get('/ready', async (_req, res) => {
  try {
    const dbHealthy = await db.healthCheck();

    if (dbHealthy) {
      successResponse(res, { ready: true }, 200);
    } else {
      successResponse(res, { ready: false, reason: 'Database not ready' }, 503);
    }
  } catch (error) {
    successResponse(res, { ready: false, reason: 'Health check failed' }, 503);
  }
});

/**
 * GET /health/live
 * Liveness probe
 */
router.get('/live', (_req, res) => {
  successResponse(res, { alive: true }, 200);
});

export default router;
