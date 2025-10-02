/**
 * Health Check Controller
 */
import { Request, Response } from 'express';
import db from '../db';
import { ApiResponse } from '../types';

export class HealthController {
  /**
   * Basic health check
   */
  static async healthCheck(req: Request, res: Response): Promise<void> {
    const response: ApiResponse = {
      success: true,
      data: {
        status: 'healthy',
        service: 'user-profile',
        version: process.env.npm_package_version || '1.0.0',
        timestamp: new Date().toISOString(),
      },
      meta: {
        timestamp: new Date().toISOString(),
      },
    };

    res.json(response);
  }

  /**
   * Detailed health check with dependencies
   */
  static async detailedHealthCheck(req: Request, res: Response): Promise<void> {
    const checks: any = {
      database: { status: 'unknown' },
      memory: { status: 'unknown' },
    };

    // Check database
    try {
      await db.query('SELECT 1');
      const stats = db.getStats();
      checks.database = {
        status: 'healthy',
        connections: stats,
      };
    } catch (error: any) {
      checks.database = {
        status: 'unhealthy',
        error: error.message,
      };
    }

    // Check memory usage
    const memUsage = process.memoryUsage();
    checks.memory = {
      status: 'healthy',
      usage: {
        rss: `${Math.round(memUsage.rss / 1024 / 1024)}MB`,
        heapTotal: `${Math.round(memUsage.heapTotal / 1024 / 1024)}MB`,
        heapUsed: `${Math.round(memUsage.heapUsed / 1024 / 1024)}MB`,
        external: `${Math.round(memUsage.external / 1024 / 1024)}MB`,
      },
    };

    const allHealthy = Object.values(checks).every(
      (check: any) => check.status === 'healthy'
    );

    const response: ApiResponse = {
      success: allHealthy,
      data: {
        status: allHealthy ? 'healthy' : 'degraded',
        service: 'user-profile',
        version: process.env.npm_package_version || '1.0.0',
        timestamp: new Date().toISOString(),
        checks,
      },
      meta: {
        timestamp: new Date().toISOString(),
      },
    };

    res.status(allHealthy ? 200 : 503).json(response);
  }

  /**
   * Readiness probe
   */
  static async readiness(req: Request, res: Response): Promise<void> {
    try {
      await db.query('SELECT 1');
      res.status(200).json({ ready: true });
    } catch (error) {
      res.status(503).json({ ready: false });
    }
  }

  /**
   * Liveness probe
   */
  static async liveness(req: Request, res: Response): Promise<void> {
    res.status(200).json({ alive: true });
  }
}

export default HealthController;
