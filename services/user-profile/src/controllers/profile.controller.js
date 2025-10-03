/**
 * Profile Controller - Business Logic
 */
import ProfileModel from '../models/profile.model.js';
import { AppError } from '../middleware/error.middleware.js';
import logger from '../utils/logger.js';

export class ProfileController {
  /**
   * Create a new profile
   */
  static async create(req, res) {
    const input = req.body;

    // Check if profile already exists
    const existingProfile = await ProfileModel.findByUserId(input.userId);
    if (existingProfile) {
      throw new AppError(409, 'PROFILE_EXISTS', 'Profile already exists for this user');
    }

    const profile = await ProfileModel.create(input);

    const response = {
      success: true,
      data: profile,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.status(201).json(response);
  }

  /**
   * Get profile by ID
   */
  static async getById(req, res) {
    const { id } = req.params;
    const requesterId = req.user?.userId;

    const profile = await ProfileModel.findById(id, { requesterId });

    if (!profile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    const response = {
      success: true,
      data: profile,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * Get profile by user ID
   */
  static async getByUserId(req, res) {
    const { userId } = req.params;
    const requesterId = req.user?.userId;

    const profile = await ProfileModel.findByUserId(userId, { requesterId });

    if (!profile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    const response = {
      success: true,
      data: profile,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * List all profiles with pagination
   */
  static async list(req, res) {
    const { page = 1, limit = 20 } = req.query;
    const requesterId = req.user?.userId;

    const { profiles, total } = await ProfileModel.findAll(
      Number(page),
      Number(limit),
      { requesterId }
    );

    const totalPages = Math.ceil(total / Number(limit));

    const response = {
      success: true,
      data: profiles,
      pagination: {
        page: Number(page),
        limit: Number(limit),
        total,
        totalPages,
      },
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * Update profile
   */
  static async update(req, res) {
    const { id } = req.params;
    const input = req.body;

    // Verify ownership
    const existingProfile = await ProfileModel.findById(id, { includePrivate: true });
    if (!existingProfile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    if (req.user && existingProfile.userId !== req.user.userId) {
      throw new AppError(403, 'FORBIDDEN', 'You do not have permission to update this profile');
    }

    const updatedProfile = await ProfileModel.update(id, input);

    if (!updatedProfile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    const response = {
      success: true,
      data: updatedProfile,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * Delete profile
   */
  static async delete(req, res) {
    const { id } = req.params;

    // Verify ownership
    const existingProfile = await ProfileModel.findById(id, { includePrivate: true });
    if (!existingProfile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    if (req.user && existingProfile.userId !== req.user.userId) {
      throw new AppError(403, 'FORBIDDEN', 'You do not have permission to delete this profile');
    }

    const deleted = await ProfileModel.delete(id);

    if (!deleted) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    const response = {
      success: true,
      data: { message: 'Profile deleted successfully' },
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * Update privacy settings
   */
  static async updatePrivacySettings(req, res) {
    const { id } = req.params;
    const settings = req.body;

    // Verify ownership
    const existingProfile = await ProfileModel.findById(id, { includePrivate: true });
    if (!existingProfile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    if (req.user && existingProfile.userId !== req.user.userId) {
      throw new AppError(403, 'FORBIDDEN', 'You do not have permission to update this profile');
    }

    const updatedProfile = await ProfileModel.updatePrivacySettings(id, settings);

    if (!updatedProfile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    const response = {
      success: true,
      data: updatedProfile,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * Search profiles
   */
  static async search(req, res) {
    const { q, page = 1, limit = 20 } = req.query;
    const requesterId = req.user?.userId;

    if (!q || typeof q !== 'string') {
      throw new AppError(400, 'INVALID_QUERY', 'Search query is required');
    }

    const { profiles, total } = await ProfileModel.search(
      q,
      Number(page),
      Number(limit),
      { requesterId }
    );

    const totalPages = Math.ceil(total / Number(limit));

    const response = {
      success: true,
      data: profiles,
      pagination: {
        page: Number(page),
        limit: Number(limit),
        total,
        totalPages,
      },
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }

  /**
   * Get current user's profile
   */
  static async getMyProfile(req, res) {
    if (!req.user) {
      throw new AppError(401, 'UNAUTHORIZED', 'Authentication required');
    }

    const profile = await ProfileModel.findByUserId(req.user.userId, {
      includePrivate: true,
      requesterId: req.user.userId,
    });

    if (!profile) {
      throw new AppError(404, 'PROFILE_NOT_FOUND', 'Profile not found');
    }

    const response = {
      success: true,
      data: profile,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: req.headers['x-request-id'],
      },
    };

    res.json(response);
  }
}

export default ProfileController;
