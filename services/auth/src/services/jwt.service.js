import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import { jwtConfig } from '../config/index.js';
import logger from '../utils/logger.js';

export class JWTService {
  /**
   * Generate access and refresh tokens for a user
   */
  generateTokens(userId, email) {
    const payload = {
      userId,
      email,
    };

    const accessToken = jwt.sign(
      payload,
      jwtConfig.accessTokenSecret,
      { expiresIn: jwtConfig.accessTokenExpiry }
    );

    const refreshToken = jwt.sign(
      payload,
      jwtConfig.refreshTokenSecret,
      { expiresIn: jwtConfig.refreshTokenExpiry }
    );

    // Calculate expiry time in seconds
    const decoded = jwt.decode(accessToken);
    const expiresIn = decoded.exp - decoded.iat;

    logger.info('Generated tokens', { userId, email });

    return {
      accessToken,
      refreshToken,
      expiresIn,
    };
  }

  /**
   * Verify and decode access token
   */
  verifyAccessToken(token) {
    try {
      const decoded = jwt.verify(
        token,
        jwtConfig.accessTokenSecret
      );
      return decoded;
    } catch (error) {
      if (error instanceof jwt.TokenExpiredError) {
        throw new Error('TOKEN_EXPIRED');
      }
      if (error instanceof jwt.JsonWebTokenError) {
        throw new Error('INVALID_TOKEN');
      }
      throw error;
    }
  }

  /**
   * Verify and decode refresh token
   */
  verifyRefreshToken(token) {
    try {
      const decoded = jwt.verify(
        token,
        jwtConfig.refreshTokenSecret
      );
      return decoded;
    } catch (error) {
      if (error instanceof jwt.TokenExpiredError) {
        throw new Error('TOKEN_EXPIRED');
      }
      if (error instanceof jwt.JsonWebTokenError) {
        throw new Error('INVALID_TOKEN');
      }
      throw error;
    }
  }

  /**
   * Hash a token for storage (refresh tokens)
   */
  hashToken(token) {
    return crypto.createHash('sha256').update(token).digest('hex');
  }

  /**
   * Decode token without verification (for debugging)
   */
  decodeToken(token) {
    try {
      return jwt.decode(token);
    } catch (error) {
      logger.error('Error decoding token', { error });
      return null;
    }
  }
}

export const jwtService = new JWTService();
export default jwtService;
