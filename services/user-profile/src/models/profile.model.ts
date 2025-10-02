/**
 * User Profile Data Model and Repository
 */
import { PoolClient } from 'pg';
import db from '../db';
import {
  UserProfile,
  CreateProfileInput,
  UpdateProfileInput,
  PrivacySettings,
  PrivacyLevel,
  ProfileQueryOptions,
} from '../types';
import logger from '../utils/logger';

export class ProfileModel {
  /**
   * Create a new user profile
   */
  static async create(input: CreateProfileInput): Promise<UserProfile> {
    const defaultPrivacySettings: PrivacySettings = {
      profileVisibility: PrivacyLevel.PUBLIC,
      showEmail: false,
      showBio: true,
      showAvatar: true,
      allowMessaging: true,
      allowFollowing: true,
      ...input.privacySettings,
    };

    const query = `
      INSERT INTO user_profiles (
        user_id, name, bio, avatar_url, preferences, privacy_settings
      ) VALUES ($1, $2, $3, $4, $5, $6)
      RETURNING *
    `;

    const values = [
      input.userId,
      input.name,
      input.bio || null,
      input.avatarUrl || null,
      JSON.stringify(input.preferences || {}),
      JSON.stringify(defaultPrivacySettings),
    ];

    try {
      const result = await db.query<UserProfile>(query, values);
      return this.mapDbRowToProfile(result.rows[0]);
    } catch (error: any) {
      if (error.code === '23505') {
        // Unique constraint violation
        throw new Error(`Profile already exists for user ${input.userId}`);
      }
      logger.error('Error creating profile', { error, input });
      throw error;
    }
  }

  /**
   * Find profile by ID
   */
  static async findById(
    id: string,
    options: ProfileQueryOptions = {}
  ): Promise<UserProfile | null> {
    const query = `
      SELECT * FROM user_profiles WHERE id = $1
    `;

    const result = await db.query<UserProfile>(query, [id]);
    if (result.rows.length === 0) {
      return null;
    }

    const profile = this.mapDbRowToProfile(result.rows[0]);
    return this.applyPrivacyFilter(profile, options);
  }

  /**
   * Find profile by user ID
   */
  static async findByUserId(
    userId: string,
    options: ProfileQueryOptions = {}
  ): Promise<UserProfile | null> {
    const query = `
      SELECT * FROM user_profiles WHERE user_id = $1
    `;

    const result = await db.query<UserProfile>(query, [userId]);
    if (result.rows.length === 0) {
      return null;
    }

    const profile = this.mapDbRowToProfile(result.rows[0]);
    return this.applyPrivacyFilter(profile, options);
  }

  /**
   * Find all profiles with pagination
   */
  static async findAll(
    page: number = 1,
    limit: number = 20,
    options: ProfileQueryOptions = {}
  ): Promise<{ profiles: UserProfile[]; total: number }> {
    const offset = (page - 1) * limit;

    const countQuery = `SELECT COUNT(*) FROM user_profiles`;
    const dataQuery = `
      SELECT * FROM user_profiles
      ORDER BY created_at DESC
      LIMIT $1 OFFSET $2
    `;

    const [countResult, dataResult] = await Promise.all([
      db.query(countQuery),
      db.query<UserProfile>(dataQuery, [limit, offset]),
    ]);

    const total = parseInt(countResult.rows[0].count, 10);
    const profiles = dataResult.rows.map((row) => {
      const profile = this.mapDbRowToProfile(row);
      return this.applyPrivacyFilter(profile, options);
    });

    return { profiles, total };
  }

  /**
   * Update profile
   */
  static async update(
    id: string,
    input: UpdateProfileInput
  ): Promise<UserProfile | null> {
    const updates: string[] = [];
    const values: any[] = [];
    let paramIndex = 1;

    if (input.name !== undefined) {
      updates.push(`name = $${paramIndex++}`);
      values.push(input.name);
    }

    if (input.bio !== undefined) {
      updates.push(`bio = $${paramIndex++}`);
      values.push(input.bio);
    }

    if (input.avatarUrl !== undefined) {
      updates.push(`avatar_url = $${paramIndex++}`);
      values.push(input.avatarUrl);
    }

    if (input.preferences !== undefined) {
      updates.push(`preferences = $${paramIndex++}`);
      values.push(JSON.stringify(input.preferences));
    }

    if (input.privacySettings !== undefined) {
      // Merge with existing privacy settings
      const currentProfile = await this.findById(id);
      if (!currentProfile) {
        return null;
      }
      const mergedSettings = {
        ...currentProfile.privacySettings,
        ...input.privacySettings,
      };
      updates.push(`privacy_settings = $${paramIndex++}`);
      values.push(JSON.stringify(mergedSettings));
    }

    if (updates.length === 0) {
      return this.findById(id);
    }

    values.push(id);
    const query = `
      UPDATE user_profiles
      SET ${updates.join(', ')}
      WHERE id = $${paramIndex}
      RETURNING *
    `;

    const result = await db.query<UserProfile>(query, values);
    if (result.rows.length === 0) {
      return null;
    }

    return this.mapDbRowToProfile(result.rows[0]);
  }

  /**
   * Delete profile
   */
  static async delete(id: string): Promise<boolean> {
    const query = `DELETE FROM user_profiles WHERE id = $1`;
    const result = await db.query(query, [id]);
    return (result.rowCount ?? 0) > 0;
  }

  /**
   * Search profiles by name or bio
   */
  static async search(
    searchTerm: string,
    page: number = 1,
    limit: number = 20,
    options: ProfileQueryOptions = {}
  ): Promise<{ profiles: UserProfile[]; total: number }> {
    const offset = (page - 1) * limit;

    const countQuery = `
      SELECT COUNT(*) FROM user_profiles
      WHERE to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(bio, ''))
      @@ plainto_tsquery('english', $1)
    `;

    const dataQuery = `
      SELECT * FROM user_profiles
      WHERE to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(bio, ''))
      @@ plainto_tsquery('english', $1)
      ORDER BY created_at DESC
      LIMIT $2 OFFSET $3
    `;

    const [countResult, dataResult] = await Promise.all([
      db.query(countQuery, [searchTerm]),
      db.query<UserProfile>(dataQuery, [searchTerm, limit, offset]),
    ]);

    const total = parseInt(countResult.rows[0].count, 10);
    const profiles = dataResult.rows.map((row) => {
      const profile = this.mapDbRowToProfile(row);
      return this.applyPrivacyFilter(profile, options);
    });

    return { profiles, total };
  }

  /**
   * Map database row to UserProfile type
   */
  private static mapDbRowToProfile(row: any): UserProfile {
    return {
      id: row.id,
      userId: row.user_id,
      name: row.name,
      bio: row.bio,
      avatarUrl: row.avatar_url,
      preferences: row.preferences || {},
      privacySettings: row.privacy_settings || {},
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  }

  /**
   * Apply privacy filters based on requester
   */
  private static applyPrivacyFilter(
    profile: UserProfile,
    options: ProfileQueryOptions
  ): UserProfile {
    const { includePrivate = false, requesterId } = options;

    // If requester is the owner or has admin access, return full profile
    if (includePrivate || requesterId === profile.userId) {
      return profile;
    }

    // Apply privacy settings
    const filtered = { ...profile };

    if (!profile.privacySettings.showBio) {
      filtered.bio = null;
    }

    if (!profile.privacySettings.showAvatar) {
      filtered.avatarUrl = null;
    }

    // Hide preferences if profile is private
    if (profile.privacySettings.profileVisibility === PrivacyLevel.PRIVATE) {
      filtered.preferences = {};
    }

    return filtered;
  }

  /**
   * Update profile privacy settings
   */
  static async updatePrivacySettings(
    id: string,
    settings: Partial<PrivacySettings>
  ): Promise<UserProfile | null> {
    const currentProfile = await this.findById(id, { includePrivate: true });
    if (!currentProfile) {
      return null;
    }

    const mergedSettings = {
      ...currentProfile.privacySettings,
      ...settings,
    };

    const query = `
      UPDATE user_profiles
      SET privacy_settings = $1
      WHERE id = $2
      RETURNING *
    `;

    const result = await db.query<UserProfile>(query, [
      JSON.stringify(mergedSettings),
      id,
    ]);

    if (result.rows.length === 0) {
      return null;
    }

    return this.mapDbRowToProfile(result.rows[0]);
  }
}

export default ProfileModel;
