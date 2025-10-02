import { Request, Response, NextFunction } from 'express';
import { errorResponse, ErrorCodes } from '../utils/response';
import logger from '../utils/logger';

/**
 * Global error handling middleware
 */
export function errorHandler(
  error: Error,
  req: Request,
  res: Response,
  _next: NextFunction
): void {
  logger.error('Unhandled error', {
    error: error.message,
    stack: error.stack,
    path: req.path,
    method: req.method,
  });

  // Map specific errors to HTTP status codes
  const errorMap: Record<string, { code: string; status: number; message: string }> = {
    DUPLICATE_USER: {
      code: ErrorCodes.DUPLICATE_USER,
      status: 409,
      message: 'User with this email already exists',
    },
    USER_NOT_FOUND: {
      code: ErrorCodes.USER_NOT_FOUND,
      status: 404,
      message: 'User not found',
    },
    INVALID_CREDENTIALS: {
      code: ErrorCodes.INVALID_CREDENTIALS,
      status: 401,
      message: 'Invalid email or password',
    },
    USER_INACTIVE: {
      code: ErrorCodes.AUTHORIZATION_ERROR,
      status: 403,
      message: 'User account is inactive',
    },
    INVALID_TOKEN: {
      code: ErrorCodes.INVALID_TOKEN,
      status: 401,
      message: 'Invalid token',
    },
    TOKEN_EXPIRED: {
      code: ErrorCodes.TOKEN_EXPIRED,
      status: 401,
      message: 'Token has expired',
    },
    TOKEN_REVOKED: {
      code: ErrorCodes.TOKEN_REVOKED,
      status: 401,
      message: 'Token has been revoked',
    },
  };

  const errorInfo = errorMap[error.message] || {
    code: ErrorCodes.INTERNAL_ERROR,
    status: 500,
    message: 'Internal server error',
  };

  errorResponse(res, errorInfo.code, errorInfo.message, errorInfo.status);
}

/**
 * 404 handler for undefined routes
 */
export function notFoundHandler(
  req: Request,
  res: Response,
  _next: NextFunction
): void {
  errorResponse(
    res,
    'NOT_FOUND',
    `Route ${req.method} ${req.path} not found`,
    404
  );
}
