/**
 * Logging Utility
 */
import winston from 'winston';
import config from '../config';

const { combine, timestamp, json, printf, colorize, errors } = winston.format;

// Custom format for console output
const consoleFormat = printf(({ level, message, timestamp, ...metadata }) => {
  let msg = `${timestamp} [${level}]: ${message}`;
  if (Object.keys(metadata).length > 0) {
    msg += ` ${JSON.stringify(metadata)}`;
  }
  return msg;
});

// Create logger instance
const logger = winston.createLogger({
  level: config.logging.level,
  format: combine(
    errors({ stack: true }),
    timestamp({ format: 'YYYY-MM-DD HH:mm:ss' })
  ),
  defaultMeta: { service: 'user-profile' },
  transports: [],
});

// Add transports based on environment
if (config.env === 'production') {
  // Production: JSON format for log aggregation
  logger.add(
    new winston.transports.Console({
      format: combine(json()),
    })
  );

  // Optional: Add file transport for production
  logger.add(
    new winston.transports.File({
      filename: 'logs/error.log',
      level: 'error',
      format: combine(json()),
    })
  );

  logger.add(
    new winston.transports.File({
      filename: 'logs/combined.log',
      format: combine(json()),
    })
  );
} else {
  // Development: Human-readable format
  logger.add(
    new winston.transports.Console({
      format: combine(colorize(), consoleFormat),
    })
  );
}

export default logger;
