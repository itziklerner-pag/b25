/**
 * Metrics Routes
 */
import { Router } from 'express';
import { register } from '../middleware/metrics.middleware.js';

const router = Router();

/**
 * @route   GET /metrics
 * @desc    Prometheus metrics endpoint
 * @access  Public
 */
router.get('/', async (req, res) => {
  res.set('Content-Type', register.contentType);
  const metrics = await register.metrics();
  res.end(metrics);
});

export default router;
