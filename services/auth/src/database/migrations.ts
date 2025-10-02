import { readFileSync } from 'fs';
import { join } from 'path';
import db from './pool';
import logger from '../utils/logger';

export async function runMigrations(): Promise<void> {
  try {
    logger.info('Running database migrations...');

    const schemaPath = join(__dirname, 'schema.sql');
    const schema = readFileSync(schemaPath, 'utf-8');

    await db.query(schema);

    logger.info('Database migrations completed successfully');
  } catch (error) {
    logger.error('Failed to run database migrations', { error });
    throw error;
  }
}

export async function checkDatabaseConnection(): Promise<void> {
  try {
    const isHealthy = await db.healthCheck();
    if (!isHealthy) {
      throw new Error('Database health check failed');
    }
    logger.info('Database connection verified');
  } catch (error) {
    logger.error('Database connection check failed', { error });
    throw error;
  }
}
