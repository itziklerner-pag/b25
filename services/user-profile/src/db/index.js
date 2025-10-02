/**
 * Database Connection Pool Management
 */
import { Pool } from 'pg';
import config from '../config/index.js';
import logger from '../utils/logger.js';

class Database {
  constructor() {
    this.pool = new Pool({
      host: config.database.host,
      port: config.database.port,
      database: config.database.name,
      user: config.database.user,
      password: config.database.password,
      min: config.database.poolMin,
      max: config.database.poolMax,
      idleTimeoutMillis: 30000,
      connectionTimeoutMillis: 5000,
    });

    this.isConnected = false;

    // Handle pool errors
    this.pool.on('error', (err) => {
      logger.error('Unexpected database pool error', { error: err.message });
    });

    // Handle pool connection events
    this.pool.on('connect', () => {
      this.isConnected = true;
      logger.debug('Database pool connection established');
    });

    this.pool.on('remove', () => {
      logger.debug('Database pool connection removed');
    });
  }

  /**
   * Test database connection
   */
  async connect() {
    try {
      const client = await this.pool.connect();
      await client.query('SELECT NOW()');
      client.release();
      this.isConnected = true;
      logger.info('Database connection successful', {
        host: config.database.host,
        database: config.database.name,
      });
    } catch (error) {
      this.isConnected = false;
      logger.error('Database connection failed', { error });
      throw error;
    }
  }

  /**
   * Execute a query with parameters
   */
  async query(text, params) {
    const start = Date.now();
    try {
      const result = await this.pool.query(text, params);
      const duration = Date.now() - start;
      logger.debug('Query executed', {
        text: text.substring(0, 100),
        duration: `${duration}ms`,
        rows: result.rowCount,
      });
      return result;
    } catch (error) {
      logger.error('Query execution failed', {
        text: text.substring(0, 100),
        params,
        error,
      });
      throw error;
    }
  }

  /**
   * Get a client from the pool for transactions
   */
  async getClient() {
    return await this.pool.connect();
  }

  /**
   * Execute a transaction
   */
  async transaction(callback) {
    const client = await this.getClient();
    try {
      await client.query('BEGIN');
      const result = await callback(client);
      await client.query('COMMIT');
      return result;
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Check if database is connected
   */
  isHealthy() {
    return this.isConnected;
  }

  /**
   * Close all connections in the pool
   */
  async close() {
    try {
      await this.pool.end();
      this.isConnected = false;
      logger.info('Database pool closed');
    } catch (error) {
      logger.error('Error closing database pool', { error });
      throw error;
    }
  }

  /**
   * Get pool statistics
   */
  getStats() {
    return {
      totalCount: this.pool.totalCount,
      idleCount: this.pool.idleCount,
      waitingCount: this.pool.waitingCount,
    };
  }
}

// Export singleton instance
export const db = new Database();
export default db;
