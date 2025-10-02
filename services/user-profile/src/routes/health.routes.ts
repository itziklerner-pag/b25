/**
 * Health Check Routes
 */
import { Router } from 'express';
import HealthController from '../controllers/health.controller';
import { asyncHandler } from '../middleware/error.middleware';

const router = Router();

/**
 * @route   GET /health
 * @desc    Basic health check
 * @access  Public
 */
router.get('/', asyncHandler(HealthController.healthCheck));

/**
 * @route   GET /health/detailed
 * @desc    Detailed health check with dependencies
 * @access  Public
 */
router.get('/detailed', asyncHandler(HealthController.detailedHealthCheck));

/**
 * @route   GET /health/ready
 * @desc    Readiness probe for k8s
 * @access  Public
 */
router.get('/ready', asyncHandler(HealthController.readiness));

/**
 * @route   GET /health/live
 * @desc    Liveness probe for k8s
 * @access  Public
 */
router.get('/live', asyncHandler(HealthController.liveness));

export default router;
