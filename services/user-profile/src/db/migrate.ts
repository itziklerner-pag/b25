/**
 * Database Migration Script
 */
import { readFileSync } from 'fs';
import { join } from 'path';
import db from './index';
import logger from '../utils/logger';

async function runMigration() {
  try {
    logger.info('Starting database migration...');

    // Connect to database
    await db.connect();

    // Read schema file
    const schemaPath = join(__dirname, 'schema.sql');
    const schema = readFileSync(schemaPath, 'utf-8');

    // Execute schema
    await db.query(schema);

    logger.info('Database migration completed successfully');
    process.exit(0);
  } catch (error) {
    logger.error('Database migration failed', { error });
    process.exit(1);
  }
}

// Run migration
runMigration();
