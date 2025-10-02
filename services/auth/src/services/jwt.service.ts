import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import { jwtConfig } from '../config';
import { TokenPayload, AuthTokens } from '../types';
import logger from '../utils/logger';

export class JWTService {
  /**
   * Generate access and refresh tokens for a user
   */
  public generateTokens(userId: string, email: string): AuthTokens {
    const payload: TokenPayload = {
      userId,
      email,
    };

    const accessToken = jwt.sign(
      payload,
      jwtConfig.accessTokenSecret,
      { expiresIn: jwtConfig.accessTokenExpiry } as jwt.SignOptions
    );

    const refreshToken = jwt.sign(
      payload,
      jwtConfig.refreshTokenSecret,
      { expiresIn: jwtConfig.refreshTokenExpiry } as jwt.SignOptions
    );

    // Calculate expiry time in seconds
    const decoded = jwt.decode(accessToken) as any;
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
  public verifyAccessToken(token: string): TokenPayload {
    try {
      const decoded = jwt.verify(
        token,
        jwtConfig.accessTokenSecret
      ) as TokenPayload;
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
  public verifyRefreshToken(token: string): TokenPayload {
    try {
      const decoded = jwt.verify(
        token,
        jwtConfig.refreshTokenSecret
      ) as TokenPayload;
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
  public hashToken(token: string): string {
    return crypto.createHash('sha256').update(token).digest('hex');
  }

  /**
   * Decode token without verification (for debugging)
   */
  public decodeToken(token: string): TokenPayload | null {
    try {
      return jwt.decode(token) as TokenPayload;
    } catch (error) {
      logger.error('Error decoding token', { error });
      return null;
    }
  }
}

export const jwtService = new JWTService();
export default jwtService;
