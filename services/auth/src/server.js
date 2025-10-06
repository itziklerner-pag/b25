import { createApp } from './app.js';
import { serverConfig } from './config/index.js';
import { checkDatabaseConnection, runMigrations } from './database/migrations.js';
import logger from './utils/logger.js';
import db from './database/pool.js';
import tokenRepository from './database/repositories/token.repository.js';

async function startServer() {
  try {
    logger.info('Starting authentication service...');

    // Check database connection
    await checkDatabaseConnection();

    // Run migrations
    await runMigrations();

    // Create Express app
    const app = createApp();

    // Start server
    const server = app.listen(serverConfig.port, () => {
      logger.info('Authentication service started', {
        port: serverConfig.port,
        nodeEnv: serverConfig.nodeEnv,
      });
    });

    // Token cleanup job - runs daily
    const cleanupInterval = setInterval(async () => {
      try {
        const count = await tokenRepository.cleanupExpired();
        logger.info('Token cleanup job completed', { deletedTokens: count });
      } catch (error) {
        logger.error('Token cleanup job failed', { error: error.message });
      }
    }, 24 * 60 * 60 * 1000); // Run every 24 hours

    // Graceful shutdown
    const shutdown = async (signal) => {
      logger.info(`Received ${signal}, starting graceful shutdown...`);

      // Clear cleanup interval
      clearInterval(cleanupInterval);

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

    process.on('SIGTERM', () => shutdown('SIGTERM'));
    process.on('SIGINT', () => shutdown('SIGINT'));

    // Handle uncaught errors
    process.on('uncaughtException', (error) => {
      logger.error('Uncaught exception', { error: error.message, stack: error.stack });
      shutdown('uncaughtException');
    });

    process.on('unhandledRejection', (reason, promise) => {
      logger.error('Unhandled rejection', { reason, promise });
      shutdown('unhandledRejection');
    });
  } catch (error) {
    logger.error('Failed to start server', { error });
    process.exit(1);
  }
}

// Start the server
startServer();
