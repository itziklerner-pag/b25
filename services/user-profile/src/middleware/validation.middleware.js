/**
 * Validation Middleware
 */

export const validate = (schema, property = 'body') => {
  return (req, res, next) => {
    const { error, value } = schema.validate(req[property], {
      abortEarly: false,
      stripUnknown: true,
    });

    if (error) {
      const response = {
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
