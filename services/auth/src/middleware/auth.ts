import { Request, Response, NextFunction } from 'express';
import { jwtService } from '../services/jwt.service';
import { errorResponse, ErrorCodes } from '../utils/response';
import logger from '../utils/logger';

// Extend Express Request type to include user
declare global {
  namespace Express {
    interface Request {
      user?: {
        userId: string;
        email: string;
      };
    }
  }
}

/**
 * Middleware to authenticate requests using JWT
 */
export function authenticate(
  req: Request,
  res: Response,
  next: NextFunction
): void {
  try {
    // Get token from header
    const authHeader = req.headers.authorization;

    if (!authHeader) {
      errorResponse(
        res,
        ErrorCodes.AUTHENTICATION_ERROR,
        'No authorization header provided',
        401
      );
      return;
    }

    // Check if Bearer token
    const parts = authHeader.split(' ');
    if (parts.length !== 2 || parts[0] !== 'Bearer') {
      errorResponse(
        res,
        ErrorCodes.AUTHENTICATION_ERROR,
        'Invalid authorization header format. Use: Bearer <token>',
        401
      );
      return;
    }

    const token = parts[1];

    // Verify token
    try {
      const payload = jwtService.verifyAccessToken(token);
      req.user = {
        userId: payload.userId,
        email: payload.email,
      };
      next();
    } catch (error: any) {
      if (error.message === 'TOKEN_EXPIRED') {
        errorResponse(
          res,
          ErrorCodes.TOKEN_EXPIRED,
          'Access token has expired',
          401
        );
        return;
      }

      errorResponse(
        res,
        ErrorCodes.INVALID_TOKEN,
        'Invalid access token',
        401
      );
      return;
    }
  } catch (error) {
    logger.error('Authentication error', { error });
    errorResponse(
      res,
      ErrorCodes.INTERNAL_ERROR,
      'Internal authentication error',
      500
    );
  }
}

/**
 * Optional authentication - attach user if token is valid, but don't fail
 */
export function optionalAuthenticate(
  req: Request,
  _res: Response,
  next: NextFunction
): void {
  const authHeader = req.headers.authorization;

  if (!authHeader) {
    next();
    return;
  }

  const parts = authHeader.split(' ');
  if (parts.length !== 2 || parts[0] !== 'Bearer') {
    next();
    return;
  }

  const token = parts[1];

  try {
    const payload = jwtService.verifyAccessToken(token);
    req.user = {
      userId: payload.userId,
      email: payload.email,
    };
  } catch (error) {
    // Silently fail for optional authentication
  }

  next();
}
