/**
 * Authentication and Authorization Middleware
 */
import jwt from 'jsonwebtoken';
import config from '../config/index.js';

/**
 * Verify JWT token and attach user to request
 */
export const authenticate = (req, res, next) => {
  try {
    const authHeader = req.headers.authorization;

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      const response = {
        success: false,
        error: {
          code: 'UNAUTHORIZED',
          message: 'Missing or invalid authorization header',
        },
        meta: {
          timestamp: new Date().toISOString(),
        },
      };
      res.status(401).json(response);
      return;
    }

    const token = authHeader.substring(7); // Remove 'Bearer ' prefix

    const decoded = jwt.verify(token, config.auth.jwtSecret);
    req.user = decoded;

    next();
  } catch (error) {
    const response = {
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: error.name === 'TokenExpiredError' ? 'Token expired' : 'Invalid token',
      },
      meta: {
        timestamp: new Date().toISOString(),
      },
    };
    res.status(401).json(response);
  }
};

/**
 * Optional authentication - attaches user if token is valid but doesn't fail if missing
 */
export const optionalAuth = (req, res, next) => {
  try {
    const authHeader = req.headers.authorization;

    if (authHeader && authHeader.startsWith('Bearer ')) {
      const token = authHeader.substring(7);
      const decoded = jwt.verify(token, config.auth.jwtSecret);
      req.user = decoded;
    }

    next();
  } catch (error) {
    // Ignore authentication errors for optional auth
    next();
  }
};

/**
 * Check if authenticated user owns the resource
 */
export const requireOwnership = (userIdParam = 'userId') => {
  return (req, res, next) => {
    if (!req.user) {
      const response = {
        success: false,
        error: {
          code: 'UNAUTHORIZED',
          message: 'Authentication required',
        },
        meta: {
          timestamp: new Date().toISOString(),
        },
      };
      res.status(401).json(response);
      return;
    }

    const resourceUserId = req.params[userIdParam] || req.body.userId;

    if (req.user.userId !== resourceUserId) {
      const response = {
        success: false,
        error: {
          code: 'FORBIDDEN',
          message: 'You do not have permission to access this resource',
        },
        meta: {
          timestamp: new Date().toISOString(),
        },
      };
      res.status(403).json(response);
      return;
    }

    next();
  };
};

/**
 * Check if user has specific role
 */
export const requireRole = (roles) => {
  return (req, res, next) => {
    if (!req.user) {
      const response = {
        success: false,
        error: {
          code: 'UNAUTHORIZED',
          message: 'Authentication required',
        },
        meta: {
          timestamp: new Date().toISOString(),
        },
      };
      res.status(401).json(response);
      return;
    }

    if (!req.user.role || !roles.includes(req.user.role)) {
      const response = {
        success: false,
        error: {
          code: 'FORBIDDEN',
          message: 'Insufficient permissions',
        },
        meta: {
          timestamp: new Date().toISOString(),
        },
      };
      res.status(403).json(response);
      return;
    }

    next();
  };
};
