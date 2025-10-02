import { Router } from 'express';
import { authService } from '../services/auth.service.js';
import {
  validateRegistration,
  validateLogin,
  validateRefreshToken,
} from '../middleware/validation.js';
import { authenticate } from '../middleware/auth.js';
import { successResponse } from '../utils/response.js';
import logger from '../utils/logger.js';

const router = Router();

/**
 * POST /auth/register
 * Register a new user
 */
router.post(
  '/register',
  validateRegistration,
  async (req, res, next) => {
    try {
      const { email, password } = req.body;
      const tokens = await authService.register({ email, password });

      logger.info('User registration successful', { email });

      successResponse(res, tokens, 201);
    } catch (error) {
      next(error);
    }
  }
);

/**
 * POST /auth/login
 * Login user
 */
router.post(
  '/login',
  validateLogin,
  async (req, res, next) => {
    try {
      const { email, password } = req.body;
      const tokens = await authService.login({ email, password });

      logger.info('User login successful', { email });

      successResponse(res, tokens, 200);
    } catch (error) {
      next(error);
    }
  }
);

/**
 * POST /auth/refresh
 * Refresh access token
 */
router.post(
  '/refresh',
  validateRefreshToken,
  async (req, res, next) => {
    try {
      const { refreshToken } = req.body;
      const tokens = await authService.refreshToken(refreshToken);

      logger.info('Token refresh successful');

      successResponse(res, tokens, 200);
    } catch (error) {
      next(error);
    }
  }
);

/**
 * POST /auth/logout
 * Logout user (revoke refresh token)
 */
router.post(
  '/logout',
  validateRefreshToken,
  async (req, res, next) => {
    try {
      const { refreshToken } = req.body;
      await authService.logout(refreshToken);

      logger.info('User logout successful');

      successResponse(res, { message: 'Logged out successfully' }, 200);
    } catch (error) {
      next(error);
    }
  }
);

/**
 * GET /auth/verify
 * Verify access token and get user info
 */
router.get(
  '/verify',
  authenticate,
  async (req, res, next) => {
    try {
      // User is already attached by authenticate middleware
      successResponse(res, req.user, 200);
    } catch (error) {
      next(error);
    }
  }
);

export default router;
