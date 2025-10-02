/**
 * Validation Middleware
 */
import { Request, Response, NextFunction } from 'express';
import { Schema } from 'joi';
import { ApiResponse } from '../types';

export const validate = (schema: Schema, property: 'body' | 'query' | 'params' = 'body') => {
  return (req: Request, res: Response, next: NextFunction): void => {
    const { error, value } = schema.validate(req[property], {
      abortEarly: false,
      stripUnknown: true,
    });

    if (error) {
      const response: ApiResponse = {
        success: false,
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Invalid input data',
          details: error.details.map((detail) => ({
            field: detail.path.join('.'),
            message: detail.message,
          })),
        },
        meta: {
          timestamp: new Date().toISOString(),
        },
      };

      res.status(400).json(response);
      return;
    }

    // Replace request property with validated value
    req[property] = value;
    next();
  };
};
