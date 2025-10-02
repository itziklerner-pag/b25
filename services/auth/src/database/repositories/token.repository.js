import db from '../pool.js';
import logger from '../../utils/logger.js';

export class TokenRepository {
  /**
   * Store a refresh token
   */
  async create(userId, tokenHash, expiresAt) {
    const query = `
      INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
      VALUES ($1, $2, $3)
      RETURNING id, user_id, token_hash, expires_at, created_at, revoked
    `;

    try {
      const result = await db.query(query, [userId, tokenHash, expiresAt]);
      return result.rows[0];
    } catch (error) {
      logger.error('Error creating refresh token', { error, userId });
      throw error;
    }
  }

  /**
   * Find refresh token by hash
   */
  async findByHash(tokenHash) {
    const query = `
      SELECT id, user_id, token_hash, expires_at, created_at, revoked
      FROM refresh_tokens
      WHERE token_hash = $1
    `;

    try {
      const result = await db.query(query, [tokenHash]);
      return result.rows[0] || null;
    } catch (error) {
      logger.error('Error finding refresh token', { error });
      throw error;
    }
  }

  /**
   * Revoke a refresh token
   */
  async revoke(tokenHash) {
    const query = `
      UPDATE refresh_tokens
      SET revoked = TRUE
      WHERE token_hash = $1
    `;

    try {
      await db.query(query, [tokenHash]);
    } catch (error) {
      logger.error('Error revoking refresh token', { error });
      throw error;
    }
  }

  /**
   * Revoke all tokens for a user
   */
  async revokeAllForUser(userId) {
    const query = `
      UPDATE refresh_tokens
      SET revoked = TRUE
      WHERE user_id = $1
    `;

    try {
      await db.query(query, [userId]);
    } catch (error) {
      logger.error('Error revoking all user tokens', { error, userId });
      throw error;
    }
  }

  /**
   * Check if token is valid (not expired or revoked)
   */
  async isValid(tokenHash) {
    const query = `
      SELECT EXISTS(
        SELECT 1 FROM refresh_tokens
        WHERE token_hash = $1
        AND expires_at > CURRENT_TIMESTAMP
        AND revoked = FALSE
      )
    `;

    try {
      const result = await db.query(query, [tokenHash]);
      return result.rows[0].exists;
    } catch (error) {
      logger.error('Error checking token validity', { error });
      throw error;
    }
  }

  /**
   * Delete expired tokens
   */
  async cleanupExpired() {
    const query = `
      DELETE FROM refresh_tokens
      WHERE expires_at < CURRENT_TIMESTAMP
      OR revoked = TRUE
      RETURNING id
    `;

    try {
      const result = await db.query(query);
      const deletedCount = result.rowCount || 0;
      logger.info('Cleaned up expired tokens', { count: deletedCount });
      return deletedCount;
    } catch (error) {
      logger.error('Error cleaning up expired tokens', { error });
      throw error;
    }
  }
}

export const tokenRepository = new TokenRepository();
export default tokenRepository;
