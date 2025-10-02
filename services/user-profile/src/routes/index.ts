/**
 * Routes Index
 */
import { Router } from 'express';
import profileRoutes from './profile.routes';
import healthRoutes from './health.routes';
import metricsRoutes from './metrics.routes';

const router = Router();

// API routes
router.use('/profiles', profileRoutes);

// System routes
router.use('/health', healthRoutes);
router.use('/metrics', metricsRoutes);

export default router;
