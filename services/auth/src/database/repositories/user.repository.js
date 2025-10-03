import db from '../pool.js';
import logger from '../../utils/logger.js';

export class UserRepository {
  /**
   * Create a new user
   */
  async create(email, passwordHash) {
    const query = `
      INSERT INTO users (email, password_hash)
      VALUES ($1, $2)
      RETURNING id, email, password_hash, created_at, updated_at, last_login, is_active
    `;

    try {
      const result = await db.query(query, [email, passwordHash]);
      return result.rows[0];
    } catch (error) {
      if (error.code === '23505') {
        // Unique violation
        throw new Error('DUPLICATE_USER');
      }
      logger.error('Error creating user', { error, email });
      throw error;
    }
  }

  /**
   * Find user by email
   */
  async findByEmail(email) {
    const query = `
      SELECT id, email, password_hash, created_at, updated_at, last_login, is_active
      FROM users
      WHERE email = $1
    `;

    try {
      const result = await db.query(query, [email]);
      return result.rows[0] || null;
    } catch (error) {
      logger.error('Error finding user by email', { error, email });
      throw error;
    }
  }

  /**
   * Find user by ID
   */
  async findById(id) {
    const query = `
      SELECT id, email, password_hash, created_at, updated_at, last_login, is_active
      FROM users
      WHERE id = $1
    `;

    try {
      const result = await db.query(query, [id]);
      return result.rows[0] || null;
    } catch (error) {
      logger.error('Error finding user by ID', { error, id });
      throw error;
    }
  }

  /**
   * Update last login timestamp
   */
  async updateLastLogin(userId) {
    const query = `
      UPDATE users
      SET last_login = CURRENT_TIMESTAMP
      WHERE id = $1
    `;

    try {
      await db.query(query, [userId]);
    } catch (error) {
      logger.error('Error updating last login', { error, userId });
      throw error;
    }
  }

  /**
   * Deactivate user account
   */
  async deactivate(userId) {
    const query = `
      UPDATE users
      SET is_active = FALSE
      WHERE id = $1
    `;

    try {
      await db.query(query, [userId]);
    } catch (error) {
      logger.error('Error deactivating user', { error, userId });
      throw error;
    }
  }

  /**
   * Check if user exists by email
   */
  async exists(email) {
    const query = `
      SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
    `;

    try {
      const result = await db.query(query, [email]);
      return result.rows[0].exists;
    } catch (error) {
      logger.error('Error checking user existence', { error, email });
      throw error;
    }
  }
}

export const userRepository = new UserRepository();
export default userRepository;
