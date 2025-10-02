/**
 * Metrics Routes
 */
import { Router, Request, Response } from 'express';
import { register } from '../middleware/metrics.middleware';

const router = Router();

/**
 * @route   GET /metrics
 * @desc    Prometheus metrics endpoint
 * @access  Public
 */
router.get('/', async (req: Request, res: Response) => {
  res.set('Content-Type', register.contentType);
  const metrics = await register.metrics();
  res.end(metrics);
});

export default router;
