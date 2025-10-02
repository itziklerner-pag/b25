import { createApp } from './app';
import { serverConfig } from './config';
import { checkDatabaseConnection, runMigrations } from './database/migrations';
import logger from './utils/logger';
import db from './database/pool';

async function startServer(): Promise<void> {
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

    // Graceful shutdown
    const shutdown = async (signal: string) => {
      logger.info(`Received ${signal}, starting graceful shutdown...`);

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
