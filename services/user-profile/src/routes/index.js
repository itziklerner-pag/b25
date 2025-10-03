/**
 * Routes Index
 */
import { Router } from 'express';
import profileRoutes from './profile.routes.js';
import healthRoutes from './health.routes.js';
import metricsRoutes from './metrics.routes.js';

const router = Router();

// API routes
router.use('/profiles', profileRoutes);

// System routes
router.use('/health', healthRoutes);
router.use('/metrics', metricsRoutes);

export default router;
