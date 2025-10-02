import { userRepository } from '../database/repositories/user.repository.js';
import { tokenRepository } from '../database/repositories/token.repository.js';
import { hashPassword, comparePassword, validatePasswordStrength } from '../utils/password.js';
import { jwtService } from './jwt.service.js';
import logger from '../utils/logger.js';

export class AuthService {
  /**
   * Register a new user
   */
  async register(data) {
    const { email, password } = data;

    // Validate password strength
    const passwordValidation = validatePasswordStrength(password);
    if (!passwordValidation.valid) {
      throw new Error(`VALIDATION_ERROR: ${passwordValidation.errors.join(', ')}`);
    }

    // Check if user already exists
    const existingUser = await userRepository.findByEmail(email);
    if (existingUser) {
      throw new Error('DUPLICATE_USER');
    }

    // Hash password
    const passwordHash = await hashPassword(password);

    // Create user
    const user = await userRepository.create(email, passwordHash);

    // Generate tokens
    const tokens = jwtService.generateTokens(user.id, user.email);

    // Store refresh token
    const tokenHash = jwtService.hashToken(tokens.refreshToken);
    const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000); // 7 days
    await tokenRepository.create(user.id, tokenHash, expiresAt);

    logger.info('User registered successfully', { userId: user.id, email });

    return tokens;
  }

  /**
   * Login user
   */
  async login(data) {
    const { email, password } = data;

    // Find user
    const user = await userRepository.findByEmail(email);
    if (!user) {
      throw new Error('INVALID_CREDENTIALS');
    }

    // Check if user is active
    if (!user.is_active) {
      throw new Error('USER_INACTIVE');
    }

    // Verify password
    const isPasswordValid = await comparePassword(password, user.password_hash);
    if (!isPasswordValid) {
      throw new Error('INVALID_CREDENTIALS');
    }

    // Update last login
    await userRepository.updateLastLogin(user.id);

    // Generate tokens
    const tokens = jwtService.generateTokens(user.id, user.email);

    // Store refresh token
    const tokenHash = jwtService.hashToken(tokens.refreshToken);
    const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000); // 7 days
    await tokenRepository.create(user.id, tokenHash, expiresAt);

    logger.info('User logged in successfully', { userId: user.id, email });

    return tokens;
  }

  /**
   * Refresh access token using refresh token
   */
  async refreshToken(refreshToken) {
    // Verify refresh token
    let payload;
    try {
      payload = jwtService.verifyRefreshToken(refreshToken);
    } catch (error) {
      throw new Error(error.message || 'INVALID_TOKEN');
    }

    // Check if token exists and is valid in database
    const tokenHash = jwtService.hashToken(refreshToken);
    const isValid = await tokenRepository.isValid(tokenHash);

    if (!isValid) {
      throw new Error('TOKEN_REVOKED');
    }

    // Get user
    const user = await userRepository.findById(payload.userId);
    if (!user) {
      throw new Error('USER_NOT_FOUND');
    }

    if (!user.is_active) {
      throw new Error('USER_INACTIVE');
    }

    // Revoke old refresh token
    await tokenRepository.revoke(tokenHash);

    // Generate new tokens
    const tokens = jwtService.generateTokens(user.id, user.email);

    // Store new refresh token
    const newTokenHash = jwtService.hashToken(tokens.refreshToken);
    const expiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000); // 7 days
    await tokenRepository.create(user.id, newTokenHash, expiresAt);

    logger.info('Token refreshed successfully', { userId: user.id });

    return tokens;
  }

  /**
   * Logout user by revoking refresh token
   */
  async logout(refreshToken) {
    const tokenHash = jwtService.hashToken(refreshToken);
    await tokenRepository.revoke(tokenHash);
    logger.info('User logged out successfully');
  }

  /**
   * Verify access token and return user info
   */
  async verifyToken(accessToken) {
    try {
      const payload = jwtService.verifyAccessToken(accessToken);

      // Verify user still exists and is active
      const user = await userRepository.findById(payload.userId);
      if (!user) {
        throw new Error('USER_NOT_FOUND');
      }

      if (!user.is_active) {
        throw new Error('USER_INACTIVE');
      }

      return {
        userId: payload.userId,
        email: payload.email,
      };
    } catch (error) {
      throw new Error(error.message || 'INVALID_TOKEN');
    }
  }
}

export const authService = new AuthService();
export default authService;
