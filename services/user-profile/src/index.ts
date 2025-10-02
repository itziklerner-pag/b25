/**
 * User Profile Service - Main Entry Point
 */
import config from './config';
import db from './db';
import logger from './utils/logger';
import createApp from './app';

async function startServer() {
  try {
    // Connect to database
    logger.info('Connecting to database...');
    await db.connect();
    logger.info('Database connected successfully');

    // Create Express app
    const app = createApp();

    // Start HTTP server
    const server = app.listen(config.server.port, config.server.host, () => {
      logger.info('User Profile Service started', {
        env: config.env,
        host: config.server.host,
        port: config.server.port,
        apiVersion: config.api.version,
      });
    });

    // Graceful shutdown
    const gracefulShutdown = async (signal: string) => {
      logger.info(`${signal} received, shutting down gracefully...`);

      server.close(async () => {
        logger.info('HTTP server closed');

        try {
          await db.close();
          logger.info('Database connections closed');
          process.exit(0);
        } catch (error) {
          logger.error('Error during shutdown', { error });
          process.exit(1);
        }
      });

      // Force shutdown after 10 seconds
      setTimeout(() => {
        logger.error('Forced shutdown after timeout');
        process.exit(1);
      }, 10000);
    };

    // Handle shutdown signals
    process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
    process.on('SIGINT', () => gracefulShutdown('SIGINT'));

    // Handle uncaught exceptions
    process.on('uncaughtException', (error) => {
      logger.error('Uncaught exception', { error });
      process.exit(1);
    });

    // Handle unhandled promise rejections
    process.on('unhandledRejection', (reason, promise) => {
      logger.error('Unhandled rejection', { reason, promise });
      process.exit(1);
    });
  } catch (error) {
    logger.error('Failed to start server', { error });
    process.exit(1);
  }
}

// Start the server
startServer();
