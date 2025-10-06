import { httpRequestDuration, httpRequestCounter } from '../utils/metrics.js';

/**
 * Middleware to track HTTP request metrics
 */
export function metricsMiddleware(req, res, next) {
  const start = Date.now();

  // Capture response finish event
  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000;
    const route = req.route?.path || req.path;
    const method = req.method;
    const statusCode = res.statusCode;

    // Record metrics
    httpRequestDuration.observe(
      {
        method,
        route,
        status_code: statusCode,
      },
      duration
    );

    httpRequestCounter.inc({
      method,
      route,
      status_code: statusCode,
    });
  });

  next();
}

export default metricsMiddleware;
