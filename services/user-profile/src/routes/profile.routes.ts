/**
 * Profile Routes
 */
import { Router } from 'express';
import ProfileController from '../controllers/profile.controller';
import { authenticate, optionalAuth } from '../middleware/auth.middleware';
import { validate } from '../middleware/validation.middleware';
import { asyncHandler } from '../middleware/error.middleware';
import {
  createProfileSchema,
  updateProfileSchema,
  privacySettingsSchema,
  paginationSchema,
  searchSchema,
  uuidSchema,
} from '../validators/profile.validator';

const router = Router();

/**
 * @route   POST /api/v1/profiles
 * @desc    Create a new profile
 * @access  Authenticated
 */
router.post(
  '/',
  authenticate,
  validate(createProfileSchema, 'body'),
  asyncHandler(ProfileController.create)
);

/**
 * @route   GET /api/v1/profiles
 * @desc    List all profiles with pagination
 * @access  Public (with optional auth for privacy filtering)
 */
router.get(
  '/',
  optionalAuth,
  validate(paginationSchema, 'query'),
  asyncHandler(ProfileController.list)
);

/**
 * @route   GET /api/v1/profiles/search
 * @desc    Search profiles by name or bio
 * @access  Public (with optional auth for privacy filtering)
 */
router.get(
  '/search',
  optionalAuth,
  validate(searchSchema, 'query'),
  asyncHandler(ProfileController.search)
);

/**
 * @route   GET /api/v1/profiles/me
 * @desc    Get current user's profile
 * @access  Authenticated
 */
router.get(
  '/me',
  authenticate,
  asyncHandler(ProfileController.getMyProfile)
);

/**
 * @route   GET /api/v1/profiles/:id
 * @desc    Get profile by ID
 * @access  Public (with optional auth for privacy filtering)
 */
router.get(
  '/:id',
  optionalAuth,
  validate(uuidSchema, 'params'),
  asyncHandler(ProfileController.getById)
);

/**
 * @route   GET /api/v1/profiles/user/:userId
 * @desc    Get profile by user ID
 * @access  Public (with optional auth for privacy filtering)
 */
router.get(
  '/user/:userId',
  optionalAuth,
  asyncHandler(ProfileController.getByUserId)
);

/**
 * @route   PUT /api/v1/profiles/:id
 * @desc    Update profile
 * @access  Authenticated (owner only)
 */
router.put(
  '/:id',
  authenticate,
  validate(uuidSchema, 'params'),
  validate(updateProfileSchema, 'body'),
  asyncHandler(ProfileController.update)
);

/**
 * @route   PATCH /api/v1/profiles/:id/privacy
 * @desc    Update privacy settings
 * @access  Authenticated (owner only)
 */
router.patch(
  '/:id/privacy',
  authenticate,
  validate(uuidSchema, 'params'),
  validate(privacySettingsSchema, 'body'),
  asyncHandler(ProfileController.updatePrivacySettings)
);

/**
 * @route   DELETE /api/v1/profiles/:id
 * @desc    Delete profile
 * @access  Authenticated (owner only)
 */
router.delete(
  '/:id',
  authenticate,
  validate(uuidSchema, 'params'),
  asyncHandler(ProfileController.delete)
);

export default router;
