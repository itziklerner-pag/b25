/**
 * Profile Input Validation Schemas
 */
import Joi from 'joi';
import { PrivacyLevel } from '../types';

export const createProfileSchema = Joi.object({
  userId: Joi.string().required().min(1).max(255).trim(),
  name: Joi.string().required().min(1).max(255).trim(),
  bio: Joi.string().max(5000).trim().allow('', null),
  avatarUrl: Joi.string().uri().allow('', null),
  preferences: Joi.object({
    theme: Joi.string().valid('light', 'dark', 'auto'),
    language: Joi.string().min(2).max(10),
    timezone: Joi.string(),
    notifications: Joi.object({
      email: Joi.boolean(),
      push: Joi.boolean(),
      sms: Joi.boolean(),
    }),
    emailDigest: Joi.string().valid('daily', 'weekly', 'never'),
  }).unknown(true),
  privacySettings: Joi.object({
    profileVisibility: Joi.string().valid(...Object.values(PrivacyLevel)),
    showEmail: Joi.boolean(),
    showBio: Joi.boolean(),
    showAvatar: Joi.boolean(),
    allowMessaging: Joi.boolean(),
    allowFollowing: Joi.boolean(),
  }),
});

export const updateProfileSchema = Joi.object({
  name: Joi.string().min(1).max(255).trim(),
  bio: Joi.string().max(5000).trim().allow('', null),
  avatarUrl: Joi.string().uri().allow('', null),
  preferences: Joi.object({
    theme: Joi.string().valid('light', 'dark', 'auto'),
    language: Joi.string().min(2).max(10),
    timezone: Joi.string(),
    notifications: Joi.object({
      email: Joi.boolean(),
      push: Joi.boolean(),
      sms: Joi.boolean(),
    }),
    emailDigest: Joi.string().valid('daily', 'weekly', 'never'),
  }).unknown(true),
  privacySettings: Joi.object({
    profileVisibility: Joi.string().valid(...Object.values(PrivacyLevel)),
    showEmail: Joi.boolean(),
    showBio: Joi.boolean(),
    showAvatar: Joi.boolean(),
    allowMessaging: Joi.boolean(),
    allowFollowing: Joi.boolean(),
  }),
}).min(1);

export const privacySettingsSchema = Joi.object({
  profileVisibility: Joi.string().valid(...Object.values(PrivacyLevel)),
  showEmail: Joi.boolean(),
  showBio: Joi.boolean(),
  showAvatar: Joi.boolean(),
  allowMessaging: Joi.boolean(),
  allowFollowing: Joi.boolean(),
}).min(1);

export const paginationSchema = Joi.object({
  page: Joi.number().integer().min(1).default(1),
  limit: Joi.number().integer().min(1).max(100).default(20),
});

export const searchSchema = Joi.object({
  q: Joi.string().required().min(1).max(255).trim(),
  page: Joi.number().integer().min(1).default(1),
  limit: Joi.number().integer().min(1).max(100).default(20),
});

export const uuidSchema = Joi.string().uuid();
