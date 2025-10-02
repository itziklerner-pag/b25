/**
 * User Profile Service - Type Definitions
 */

export enum PrivacyLevel {
  PUBLIC = 'public',
  FRIENDS = 'friends',
  PRIVATE = 'private'
}

export interface UserPreferences {
  theme?: 'light' | 'dark' | 'auto';
  language?: string;
  timezone?: string;
  notifications?: {
    email?: boolean;
    push?: boolean;
    sms?: boolean;
  };
  emailDigest?: 'daily' | 'weekly' | 'never';
  [key: string]: any; // Allow additional custom preferences
}

export interface PrivacySettings {
  profileVisibility: PrivacyLevel;
  showEmail: boolean;
  showBio: boolean;
  showAvatar: boolean;
  allowMessaging: boolean;
  allowFollowing: boolean;
}

export interface UserProfile {
  id: string;
  userId: string;
  name: string;
  bio: string | null;
  avatarUrl: string | null;
  preferences: UserPreferences;
  privacySettings: PrivacySettings;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateProfileInput {
  userId: string;
  name: string;
  bio?: string;
  avatarUrl?: string;
  preferences?: UserPreferences;
  privacySettings?: Partial<PrivacySettings>;
}

export interface UpdateProfileInput {
  name?: string;
  bio?: string;
  avatarUrl?: string;
  preferences?: UserPreferences;
  privacySettings?: Partial<PrivacySettings>;
}

export interface ProfileQueryOptions {
  includePrivate?: boolean;
  requesterId?: string;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  meta?: {
    timestamp: string;
    requestId?: string;
  };
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface AuthPayload {
  userId: string;
  email?: string;
  role?: string;
  iat?: number;
  exp?: number;
}
